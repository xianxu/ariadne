// issuelintids.go — `sdlc issue lint-ids`, the id-collision check (#213).
//
// Exists as a VERB so CI and the operator run the same logic (ARCH-DRY): the
// merge-checks script is a four-line adapter, not a bash reimplementation of
// filename parsing and directory selection.
//
// It answers two questions that need different answers:
//
//	INTRODUCED    the range adds a file whose id already exists at --base under
//	              a different path. REFUSED — the range is where renaming is
//	              still cheap.
//	PRE-EXISTING  one tree already contains two files claiming one id. REPORTED,
//	              never refused: these predate the check, renumbering is operator
//	              work, and blocking every merge until it is done is worse than
//	              the bug. It is also the ONLY way the already-merged collisions
//	              are visible at all — a branch-vs-trunk diff sees two agreeing
//	              trees and finds nothing.
package main

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

type issueLintIDsFlags struct {
	Base       string
	Head       string
	Trunk      string
	IssuesDir  string
	HistoryDir string
}

func newIssueLintIDsCmd() *cobra.Command {
	f := issueLintIDsFlags{}
	cmd := &cobra.Command{
		Use:   "lint-ids",
		Short: "Refuse issue ids reused across a range (the collision check CI runs)",
		Long: `Check for issue-id collisions — two files claiming one id.

  sdlc issue lint-ids                             # this tree contradicting itself
  sdlc issue lint-ids --base <sha> --head <sha>   # what a range introduces

Two files with the same id but different slugs are different PATHS, so git
merges both cleanly and nothing else in the lifecycle objects — which is why
this check exists rather than relying on a conflict.

  INTRODUCED    refused (exit 1); the range is where renaming is still cheap
  PRE-EXISTING  reported; renumbering is operator work, and blocking every
                merge until it is done would be worse than the bug

Read-only.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runIssueLintIDs(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().StringVar(&f.Base, "base", "", "base ref of the range (omit to check --head alone)")
	cmd.Flags().StringVar(&f.Head, "head", "HEAD", "head ref of the range")
	cmd.Flags().StringVar(&f.Trunk, "trunk", "", "published ref the range will merge into (default: --base)")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory holding archived issues")
	return cmd
}

func runIssueLintIDs(stdout, stderr io.Writer, f *issueLintIDsFlags) error {
	r := claimRunner

	listing, err := idListing(f.Head, f.IssuesDir, f.HistoryDir, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("id lint skipped: %v", err))
		return nil
	}
	for _, c := range issue.DuplicateIDsInRef(listing) {
		cwarn(stderr, fmt.Sprintf("pre-existing duplicate id #%06d: %s", c.ID, strings.Join(c.Paths, ", ")))
	}

	if f.Base == "" {
		cok(stderr, "id lint: pre-existing duplicates reported above, if any (no range given)")
		return nil
	}

	clashes, err := introducedIDClashes(f.Base, f.Head, f.Trunk, f.IssuesDir, f.HistoryDir, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("id lint skipped: %v", err))
		return nil
	}
	if len(clashes) == 0 {
		cok(stderr, "id lint: this range introduces no reused issue ids")
		return nil
	}
	fmt.Fprintf(stderr, "%sthis range reuses %d issue id(s) that already exist at %s:%s\n%s\n",
		ansiRed, len(clashes), f.Base, ansiReset, strings.Join(clashes, "\n"))
	fmt.Fprintln(stderr, "  Two files with the same id but different slugs are different PATHS, so git")
	fmt.Fprintln(stderr, "  merges both and nothing else objects. Rename this range's file to a fresh id")
	fmt.Fprintln(stderr, "  (and its `id:` frontmatter) before merging.")
	exitWithCode(1)
	return nil
}

// introducedIDClashes returns the rendered clash reports for ids present at head
// under a different path than at base. Split from the command so the decision is
// testable without driving cobra or exiting.
func introducedIDClashes(base, head, trunk, issuesDir, historyDir string, r gitRunner) ([]string, error) {
	headByID, err := issueFilesByID(head, issuesDir, historyDir, r)
	if err != nil {
		return nil, err
	}
	baseByID, err := issueFilesByID(base, issuesDir, historyDir, r)
	if err != nil {
		return nil, err
	}
	// Trunk defaults to base when not given (a plain two-ref comparison).
	trunkByID := baseByID
	if trunk != "" && trunk != base {
		if t, terr := issueFilesByID(trunk, issuesDir, historyDir, r); terr == nil {
			trunkByID = t
		}
	}
	merged := mergedPathsFor(headByID, baseByID, trunkByID)
	var clashes []string
	for _, id := range introducedCollisions(headByID, baseByID, trunkByID) {
		clashes = append(clashes, fmt.Sprintf("  #%06d would be claimed by %d files after merge:\n      %s",
			id, len(merged[id]), strings.Join(merged[id], "\n      ")))
	}
	sort.Strings(clashes)
	return clashes, nil
}

// idListing concatenates `git ls-tree` output for one ref's id-bearing
// directories — the IO half of the within-ref scan.
func idListing(ref, issuesDir, historyDir string, r gitRunner) (string, error) {
	if out, err := r.Git("rev-parse", "--verify", "--quiet", ref); err != nil || strings.TrimSpace(string(out)) == "" {
		return "", fmt.Errorf("no %s ref", ref)
	}
	dirs, err := repoRelativeIDDirs(issuesDir, historyDir)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, dir := range dirs {
		out, err := r.Git("ls-tree", "--name-only", ref, dir+"/")
		if err != nil {
			// Same rule as the other two read sites (#213 BR-7/BR-15): a partial
			// listing parses as a clean tree, so it would report "no duplicates"
			// having seen a fraction of them. ls-tree exits 0 printing nothing
			// for a directory absent from the ref, so this is a real failure.
			return "", fmt.Errorf("ls-tree %s %s/: %v\n%s", ref, dir, err, out)
		}
		b.Write(out)
		b.WriteString("\n")
	}
	return b.String(), nil
}
