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

	dirs, err := resolveIDDirs(f.IssuesDir, f.HistoryDir)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("id lint skipped: %v", err))
		return nil
	}
	head, err := refIDSpace(f.Head, dirs, r)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("id lint skipped: %v", err))
		return nil
	}

	var baseSpace map[int][]string
	if f.Base != "" {
		if b, berr := refIDSpace(f.Base, dirs, r); berr == nil {
			baseSpace = b
		}
	}
	for _, d := range classifyDuplicates(head, baseSpace, f.Base != "") {
		cwarn(stderr, fmt.Sprintf("%s #%06d: %s", d.Label, d.ID, strings.Join(d.Paths, ", ")))
	}

	if f.Base == "" {
		cok(stderr, "id lint: duplicates within "+f.Head+" reported above, if any (no range given)")
		return nil
	}

	clashes, err := introducedIDClashes(f.Base, f.Trunk, head, dirs, r)
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

// labeledDuplicate is a within-ref duplicate together with who owns it.
type labeledDuplicate struct {
	issue.IDCollision
	Label string
}

// classifyDuplicates labels each duplicate WITHIN head by whether the range
// inherited it or introduced it (#213 BR-18).
//
// Every one of them used to be labelled "pre-existing". So a range that added a
// second file for a live id — the exact thing this verb refuses — was reported
// as inherited damage in the same run that refused it, and the report
// contradicted the exit code. Worse, "pre-existing" is what tells an operator to
// ignore a line.
//
// With no range there is nothing to attribute to, and the honest label claims
// nothing about origin. Pure, because it is a DECISION: the verb exits on the
// refusal path, so a label reachable only through os.Exit is a label nothing can
// test.
func classifyDuplicates(head, base map[int][]string, hasRange bool) []labeledDuplicate {
	var out []labeledDuplicate
	for _, c := range issue.DuplicatesIn(head) {
		label := "duplicate id"
		switch {
		case !hasRange:
			// no range: attribution is unavailable, so claim none
		case len(base[c.ID]) > 1:
			label = "pre-existing duplicate id"
		default:
			label = "INTRODUCED duplicate id"
		}
		out = append(out, labeledDuplicate{IDCollision: c, Label: label})
	}
	return out
}

// introducedIDClashes returns the rendered clash reports for the merge result of
// this range. Split from the command so the decision is testable without driving
// cobra or exiting.
//
// A trunk read that FAILS is an error, not a reason to substitute the base
// (#213 BR-18): the two differ precisely when the trunk moved while the branch
// was open, which is the case the gate exists for, so the silent substitution
// answered the easy question and reported it as the hard one.
func introducedIDClashes(base, trunk string, head map[int][]string, dirs idDirs, r gitRunner) ([]string, error) {
	baseByID, err := refIDSpace(base, dirs, r)
	if err != nil {
		return nil, err
	}
	// Trunk defaults to base when not given (a plain two-ref comparison).
	trunkByID := baseByID
	if trunk != "" && trunk != base {
		t, terr := refIDSpace(trunk, dirs, r)
		if terr != nil {
			return nil, terr
		}
		trunkByID = t
	}
	return renderClashes(head, baseByID, trunkByID), nil
}
