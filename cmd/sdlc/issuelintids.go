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
	"context"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
)

type issueLintIDsFlags struct {
	Context    context.Context
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
  CARDLESS      (issue tracker repositories) refused (exit 1): details the
                range adds must belong to a card with the same id and slug —
                ids are allocated on the tracker, so a hand-made file collides
  PRE-EXISTING  reported; renumbering is operator work, and blocking every
                merge until it is done would be worse than the bug

Exit codes: 0 clean, 1 collisions introduced, 2 THE CHECK COULD NOT RUN.
A degraded read exits 2 rather than 0, so a required status check cannot go
green on a check that never looked.

Read-only.`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			f.Context = cmd.Context()
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

// lintCouldNotRun is the exit code for a check that did not get to look (#213
// BR-23).
//
// "Skipped" used to warn and exit 0 — and this verb is what CI shells to, so a
// read failure was a GREEN required status check. Green is the one answer a
// check that did not run must never give: it is indistinguishable from "looked
// and found nothing", which is precisely the confusion this whole issue is
// about. 2 rather than 1 keeps "could not check" separable from "found a
// collision", so CI can treat them differently.
const lintCouldNotRun = 2

func runIssueLintIDs(stdout, stderr io.Writer, f *issueLintIDsFlags) error {
	r := claimRunner

	degraded := func(err error) error {
		cwarn(stderr, fmt.Sprintf("id lint COULD NOT RUN: %v", err))
		cwarn(stderr, "      exiting 2 — a check that did not look must not report clean")
		exitWithCode(lintCouldNotRun)
		return nil
	}

	dirs, err := resolveIDDirs(f.IssuesDir, f.HistoryDir)
	if err != nil {
		return degraded(err)
	}
	head, err := refIDSpace(f.Head, dirs, r)
	if err != nil {
		return degraded(err)
	}

	var baseSpace map[int][]string
	if f.Base != "" {
		b, berr := refIDSpace(f.Base, dirs, r)
		if berr != nil {
			return degraded(berr)
		}
		baseSpace = b
	}
	for _, d := range classifyDuplicates(head, baseSpace, f.Base != "") {
		cwarn(stderr, fmt.Sprintf("%s #%06d: %s", d.Label, d.ID, strings.Join(d.Paths, ", ")))
	}

	if f.Base == "" {
		cok(stderr, "id lint: duplicates within "+f.Head+" reported above, if any (no range given)")
		return nil
	}

	clashes, err := introducedIDClashes(baseSpace, f.Base, f.Trunk, head, dirs, r)
	if err != nil {
		return degraded(err)
	}
	cardless, err := cardlessAdditions(commandContext(f.Context), f.Head, baseSpace, head, dirs, r)
	if err != nil {
		return degraded(err)
	}
	if len(cardless) > 0 {
		fmt.Fprintf(stderr, "%sthis range adds %d details file(s) with no matching issue card:%s\n%s\n",
			ansiRed, len(cardless), ansiReset, strings.Join(cardless, "\n"))
		fmt.Fprintln(stderr, "  In an issue tracker repository ids are allocated on cards; a hand-made details")
		fmt.Fprintln(stderr, "  file can collide with the next allocated id. File it with `sdlc issue new`.")
		exitWithCode(1)
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
//
// Takes the base SPACE, not the base ref name (#213 BR-29): one ref's id space
// is read once per command and passed down, never re-derived by a callee. The
// caller already read it for the duplicate labelling, and re-deriving it cost a
// second rev-parse plus one ls-tree per id directory — and, worse, admitted the
// possibility of two reads of "the base" disagreeing within one invocation.
func introducedIDClashes(baseByID map[int][]string, base, trunk string, head map[int][]string, dirs idDirs, r gitRunner) ([]string, error) {
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

// cardlessAdditions is the tracker-era half of the id check (#252): in a
// repository whose head carries the cutover marker, every details file the
// range adds must belong to a card at the same id and slug. The tracker is read
// fresh; an unreadable tracker is a check that could not run, never clean.
func cardlessAdditions(ctx context.Context, headRef string, base, head map[int][]string, dirs idDirs, r gitRunner) ([]string, error) {
	if len(dirs.Rel) == 0 {
		return nil, nil
	}
	if _, err := r.Git("cat-file", "-e", "--end-of-options", headRef+":"+tracker.CutoverMarkerPath); err != nil {
		return nil, nil // not a tracker repository at head
	}
	repo, err := tracker.RepositoryForCheckout(ctx, dirs.Top)
	if err != nil {
		return nil, err
	}
	if repo == nil {
		return nil, fmt.Errorf("%s marks an issue tracker repository, but no issue tracker is reachable", tracker.CutoverMarkerPath)
	}
	snap, err := repo.Snapshot()
	if err != nil {
		return nil, err
	}
	cards := map[string]string{}
	for _, c := range snap.Records() {
		cards[c.ID] = c.Path
	}
	return detailsWithoutCards(addedDetails(base, head, dirs.Rel[0]), cards), nil
}

// addedDetails lists the details paths head has under issuesDir that base does
// not. Pure.
func addedDetails(base, head map[int][]string, issuesDir string) []string {
	had := map[string]bool{}
	for _, paths := range base {
		for _, p := range paths {
			had[p] = true
		}
	}
	var added []string
	for _, paths := range head {
		for _, p := range paths {
			if !had[p] && path.Dir(p) == issuesDir {
				added = append(added, p)
			}
		}
	}
	sort.Strings(added)
	return added
}

// detailsWithoutCards returns the added details whose id has no card, or whose
// card sits at another slug. Pure.
func detailsWithoutCards(added []string, cards map[string]string) []string {
	var missing []string
	for _, p := range added {
		id, slug, ok := issue.ParseFilename(path.Base(p))
		if !ok {
			continue
		}
		if cards[id] != tracker.CardPath(id, slug) {
			missing = append(missing, p)
		}
	}
	return missing
}
