// merge.go — `sdlc merge` subcommand. Ports the `merge:` Make target.
//
// The longest + most safety-conscious script in the lift table. Runs from a
// feature branch (refuses main), guards every irreversible step, merges via a
// GitHub PR (server-side, so CI gates it), then cleans up. Two topologies
// (#51), detected automatically:
//   - in-place: the primary checkout sitting on a feature branch → switch it
//     back to main, pull, delete the branch (no worktree).
//   - worktree: a linked worktree → archive in the main worktree, remove the
//     worktree, delete the branch.
//
// Sequence:
//
//  1. branch != main / non-empty
//  2. no uncommitted tracked changes (untracked files warn, don't block — #78)
//  3. upstream configured
//  4. branch not ahead of upstream
//  5. pre-merge judges (plan + specs + lessons, skippable with --no-judge) —
//     read-only reviewers; they report, they do not edit the tree (#62 M2)
//  6. resolve topology (in-place vs worktree)
//  7. show unmerged commits (informational)
//  8. not-done issue warn (vs main)
//  9. interactive confirmation (skippable with --yes)
//     9b. re-assert no tracked dirt before the irreversible merge — refuse if a
//     judge/hook dirtied a tracked file since step 2 (#62 M1; never cross dirty)
//  10. gh pr merge (server-side), OR resume an already-merged PR if a prior run
//     was interrupted (#62 M3) → in-place: switch main; both: pull main
//  11. archive done/wontfix/punt issues into history/ (in the main checkout)
//  12. cleanup — in-place: branch delete; worktree: worktree remove + branch delete + .goto
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// mergeFlags holds the parsed flag values for the merge subcommand.
type mergeFlags struct {
	Yes        bool
	NoJudge    bool
	DryRun     bool
	IssuesDir  string
	HistoryDir string
}

// mergeRunner is the package-level runner for merge (test seam). Type
// lives in runner.go.
var mergeRunner gitRunner = execGitRunner{}

// mergePrompter is a tiny indirection over stdin so tests can drive the
// confirmation prompts deterministically. Production wraps os.Stdin.
var mergePrompter prompter = stdinPrompter{}

// runPreflightJudgesFn is the package-level seam for merge's step-5 pre-merge
// judges. Production points at runPreflightJudges. Tests swap it for a stub
// "judge" — most usefully one that DIRTIES the worktree, to prove step 9b
// re-checks cleanliness after the judges ran and refuses before the
// irreversible merge (#62 M1 / #63). It runs after step 2's clean check and
// before 9b's re-check, which is exactly the window a real dirtying hook
// would occupy.
var runPreflightJudgesFn = runPreflightJudges

// prompter abstracts the "read a line, return trimmed text" surface.
type prompter interface {
	Ask(question string, w io.Writer) string
}

type stdinPrompter struct{}

func (stdinPrompter) Ask(question string, w io.Writer) string {
	fmt.Fprint(w, question)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

// NewMergeCmd returns the cobra command for `sdlc merge`.
func NewMergeCmd() *cobra.Command {
	f := mergeFlags{}
	cmd := &cobra.Command{
		Use:           "merge",
		Short:         "Merge the current branch (in-place or worktree) via GitHub PR, archive done issues, clean up",
		Long:          "Placeholder — replaced by helptext.MustGet(\"merge\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMerge(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	}
	cmd.Flags().BoolVar(&f.Yes, "yes", false, "skip the final irreversible-merge confirmation AND not-done warn")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip pre-merge judges (emergency-only)")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be operations; do not merge or clean up")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory for archived issues")
	return cmd
}

// runMerge dispatches the merge workflow.
// worktreeDirty returns the trimmed `git status --porcelain` output ("" =
// clean) via the runner, or an error if git status itself fails. Checked at the
// start of merge AND — per #62 — re-checked immediately before the irreversible
// `gh pr merge`: a pre-merge judge/hook can dirty the tree after the initial
// check, and the post-merge `git switch main` then refuses, stranding the merge
// (remote merged, local stuck). Re-asserting here converts that into a clean
// pre-merge refusal.
func worktreeDirty(r gitRunner) (string, error) {
	out, err := r.Git("status", "--porcelain")
	if err != nil {
		return "", fmt.Errorf("git status: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

// dirtyAssessment splits a `git status --porcelain` result into the classes that
// matter to the merge guard: Blocking (tracked code modifications — staged,
// modified, renamed, deleted, unmerged), Untracked (`??` non-tracker lines), and
// Tracker (workshop issue/history markdown — never blocks, #82 M2).
type dirtyAssessment struct {
	Blocking  []string // tracked code changes that MUST block the merge
	Untracked []string // untracked files — surfaced, but not a blocker (#78)
	Tracker   []string // workshop/issues|history/*.md — tracker state, never blocks (#82 M2)
}

// Refuse reports whether the tree state must block the merge. The single source
// of the merge-block decision (#78): both the step-2 check and the step-9b
// re-assert ask Refuse() rather than re-deriving `len(...) > 0` inline, so the
// two guards can't diverge (the Spec's elevated invariant). Tracker files are
// out of Blocking by construction (#82 M2), so they can't flip this.
func (d dirtyAssessment) Refuse() bool { return len(d.Blocking) > 0 }

// assessDirty classifies porcelain output. Pure (mirrors decideMergeAction's
// extracted-decision pattern). Only tracked CODE modifications block a merge: a
// dirty tracked file makes the post-merge `git switch main` / `git pull` refuse,
// stranding the server-side merge. Two classes are non-blocking: untracked `??`
// files carry across the branch switch untouched (#78), and tracker files
// (workshop issue/history markdown) are append-only shared state synced to main
// out-of-band (#82 M1) — a dirty issue file is never code contention, whether
// tracked-modified or untracked (#82 M2). Both are surfaced as warnings, not
// refusals.
func assessDirty(porcelain, issuesDir, historyDir string) dirtyAssessment {
	var d dirtyAssessment
	for _, line := range strings.Split(porcelain, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Tracker classification is by PATH and comes first, so a dirty issue
		// file lands in Tracker regardless of its tracked/untracked status code.
		path, dest := porcelainPaths(line)
		if isTrackerPath(path, issuesDir, historyDir) || isTrackerPath(dest, issuesDir, historyDir) {
			d.Tracker = append(d.Tracker, line)
			continue
		}
		if strings.HasPrefix(line, "??") {
			d.Untracked = append(d.Untracked, line)
		} else {
			d.Blocking = append(d.Blocking, line)
		}
	}
	return d
}

// isTrackerPath reports whether p is a tracker file — a workshop issue or its
// archived history copy. Reuses push.go's issue/history path predicates so the
// "what counts as tracker state" definition has one source of truth (#82 M2).
func isTrackerPath(p, issuesDir, historyDir string) bool {
	if p == "" {
		return false
	}
	return isIssuePath(p, issuesDir) || isHistoryPath(p, historyDir)
}

// porcelainPaths extracts the path (and rename/copy dest) from a `git status
// --porcelain` line. It splits on whitespace rather than slicing fixed status
// columns, because worktreeDirty whole-trims its output — stripping the leading
// status space off the first line (" M f" → "M f") and shifting any column
// parse. Field-splitting is immune: fields[0] is the status code, the path is
// the next field (last field for an `orig -> dest` rename). (Quoted paths with
// embedded spaces aren't handled — same limitation as parsePorcelainStatus.)
func porcelainPaths(line string) (path, dest string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return "", ""
	}
	for i, f := range fields {
		if f == "->" && i+1 < len(fields) {
			return fields[1], fields[len(fields)-1]
		}
	}
	return fields[1], ""
}

// mergeAction is what step 10 should do, given the PR state.
type mergeAction int

const (
	actionMergeOpen mergeAction = iota // an open PR exists → merge it (irreversible)
	actionResume                       // no open PR, but a merged one → resume cleanup (#62 M3)
	actionNoPR                         // neither → create-PR / abandon path
)

// decideMergeAction picks the step-10 action from the open-PR number and whether
// a merged PR already exists for the branch. Pure + testable; the irreversible
// PRMerge stays in runMerge. Resume (merged-but-not-open) is how a re-run
// recovers an interrupted merge (#62 M3).
func decideMergeAction(openPRNumber string, mergedExists bool) mergeAction {
	switch {
	case openPRNumber != "":
		return actionMergeOpen
	case mergedExists:
		return actionResume
	default:
		return actionNoPR
	}
}

func runMerge(stdout, stderr io.Writer, f *mergeFlags) error {
	// ── 1. Refuse if main / empty branch ────────────────────────────────────
	branch := gitx.Capture("branch", "--show-current")
	if branch == "" || branch == "main" {
		die(stderr, fmt.Sprintf("sdlc merge must be run from a worktree branch (current: %s)", valueOr(branch, "(detached)")))
	}
	cinfo(stderr, fmt.Sprintf("Branch: %s", branch))

	// ── 2. No uncommitted TRACKED changes ───────────────────────────────────
	// Untracked files don't block: they survive `git switch main`, so they can't
	// strand the post-merge cleanup the way a dirty tracked file does (#78). They
	// are surfaced as a warning so the operator still sees them.
	dirty, err := worktreeDirty(mergeRunner)
	if err != nil {
		die(stderr, err.Error())
	}
	assessment := assessDirty(dirty, f.IssuesDir, f.HistoryDir)
	if assessment.Refuse() {
		fmt.Fprintf(stderr, "  %s[x]%s Uncommitted tracked changes found — cannot merge\n", ansiRed, ansiReset)
		fmt.Fprintln(stderr, strings.Join(assessment.Blocking, "\n"))
		die(stderr, "commit or stash these tracked changes before merging")
	}
	if len(assessment.Untracked) > 0 {
		cwarn(stderr, fmt.Sprintf("%d untracked file(s) present — not blocking the merge (they survive the branch switch):", len(assessment.Untracked)))
		fmt.Fprintln(stderr, strings.Join(assessment.Untracked, "\n"))
	}
	if len(assessment.Tracker) > 0 {
		cwarn(stderr, fmt.Sprintf("%d tracker file(s) dirty — not blocking the merge (tracker state syncs to main out-of-band, #82):", len(assessment.Tracker)))
		fmt.Fprintln(stderr, strings.Join(assessment.Tracker, "\n"))
	}
	cok(stderr, "No uncommitted tracked changes")

	// Fail fast — before the slow pre-merge judges — if the confirmation
	// prompts (steps 8-9) can't be answered. They read os.Stdin, and in a
	// non-tty agent/background context a bare scan blocks forever; convert
	// that hang into a clear next-action.
	if mergeNeedsTTY(f.Yes, f.DryRun, isTTY(os.Stdin)) {
		die(stderr, "sdlc merge needs interactive confirmation, but stdin is not a terminal.\n"+
			"  Re-run with --yes to merge non-interactively (skips the confirm + not-done prompts),\n"+
			"  or run it from a terminal.")
	}

	// ── 3. Branch pushed + HEAD synced ──────────────────────────────────────
	// Key off the remote-tracking ref (origin/<branch>), NOT @{u}: a push
	// updates origin/<branch> even when it can't write the local
	// upstream-tracking config (a sandbox that blocks .git/config, or a plain
	// `git push` without -u), so requiring @{u} spuriously refuses a branch
	// that is genuinely pushed (and has an open PR).
	remoteRef := "origin/" + branch
	if gitx.Capture("rev-parse", "--verify", "--quiet", remoteRef) == "" {
		fmt.Fprintf(stderr, "  %s[x]%s %s is not on origin\n", ansiRed, ansiReset, branch)
		die(stderr, fmt.Sprintf("push the branch first (e.g. sdlc pr, or git push origin %s)", branch))
	}

	// ── 4. No local commits missing from the remote branch ──────────────────
	aheadStr := gitx.Capture("rev-list", "--count", remoteRef+"..HEAD")
	ahead, _ := strconv.Atoi(aheadStr)
	if ahead > 0 {
		fmt.Fprintf(stderr, "  %s[x]%s %d local commit(s) not yet on %s\n",
			ansiRed, ansiReset, ahead, remoteRef)
		die(stderr, "push your branch before merging — run `git push`, then re-run `sdlc merge`. "+
			"(merge is server-side: a fix you committed for a failed pre-merge gate must reach origin first.)")
	}
	cok(stderr, fmt.Sprintf("Branch pushed; HEAD synced with %s", remoteRef))

	// ── 5. Pre-merge judges ─────────────────────────────────────────────────
	if !f.NoJudge {
		preOpts := preflightOptions{
			Categories: []judge.Category{judge.Plan, judge.Specs, judge.Lessons},
			IssuesDir:  f.IssuesDir,
			HistoryDir: f.HistoryDir,
			DryRun:     f.DryRun,
			Stdout:     stdout,
			Stderr:     stderr,
		}
		if err := runPreflightJudgesFn(preOpts); err != nil {
			die(stderr, fmt.Sprintf("pre-merge judges failed: %v\n"+
				"  → fix the finding, commit, `git push`, then re-run `sdlc merge` "+
				"(the fix must reach origin — merge is server-side).", err))
		}
	} else {
		cwarn(stderr, "--no-judge: skipping pre-merge judges")
	}

	// ── 6. Resolve merge topology: in-place vs worktree ─────────────────────
	// In-place = the primary checkout sitting on a feature branch (main is
	// reached by switching here). Worktree = a linked worktree (main lives in
	// a separate dir). A linked worktree's git-dir is under .git/worktrees/.
	wtPath, _ := gitx.RepoTopLevel()
	gitDir := gitx.Capture("rev-parse", "--git-dir")
	inPlace := isInPlaceCheckout(gitDir)
	var mainPath string
	if inPlace {
		mainPath = wtPath // same checkout; we switch it back to main post-merge
		cok(stderr, "In-place branch (no worktree) — will merge, then switch this checkout back to main")
	} else {
		mp, ferr := findMainWorktree(mergeRunner)
		if ferr != nil {
			die(stderr, fmt.Sprintf("find main worktree: %v", ferr))
		}
		mainPath = mp
		cok(stderr, fmt.Sprintf("Main worktree: %s", mainPath))
	}

	repo, err := detectRepo()
	if err != nil {
		die(stderr, err.Error())
	}

	// ── 7. Show unmerged commits ────────────────────────────────────────────
	unmergedOut, _ := mergeRunner.Git("log", "main..HEAD", "--oneline")
	unmerged := strings.TrimRight(string(unmergedOut), "\n")
	if unmerged != "" {
		cok(stderr, "Unmerged local commits found:")
		for _, line := range strings.Split(unmerged, "\n") {
			fmt.Fprintf(stderr, "       %s\n", line)
		}
	} else {
		cok(stderr, "No unmerged local commits (branch is clean)")
	}

	// ── 8. Not-done issue warn (vs main) ────────────────────────────────────
	notDone, _ := touchedIssuesNotDone("main", f.IssuesDir, mergeRunner)
	if len(notDone) > 0 && !f.Yes && !f.DryRun {
		fmt.Fprintf(stderr, "  %s[!]%s Touched issue files that are NOT done:\n", ansiYellow, ansiReset)
		for _, p := range notDone {
			fmt.Fprintf(stderr, "       %s\n", p)
		}
		ans := mergePrompter.Ask("Continue anyway? [y/N] ", stderr)
		if ans != "y" && ans != "Y" {
			die(stderr, "aborted by operator")
		}
	} else if len(notDone) > 0 && f.Yes {
		cwarn(stderr, fmt.Sprintf("--yes: continuing past %d not-done issue(s)", len(notDone)))
	}

	// ── 9. Interactive confirmation ─────────────────────────────────────────
	if !f.Yes && !f.DryRun {
		ans := mergePrompter.Ask("Final confirmation: proceed with irreversible merge/cleanup actions? [y/N] ", stderr)
		if ans != "y" && ans != "Y" {
			die(stderr, "aborted by operator")
		}
	}

	if f.DryRun {
		cinfo(stderr, "dry-run — skipping merge / archive / cleanup")
		fmt.Fprintf(stdout, "Would: gh pr merge ... (or offer to create) for %s\n", branch)
		if inPlace {
			fmt.Fprintf(stdout, "Would: git switch main && git pull (in %s)\n", wtPath)
			fmt.Fprintf(stdout, "Would: archive done issues under %s/%s/\n", wtPath, f.HistoryDir)
			fmt.Fprintf(stdout, "Would: git branch -D %s\n", branch)
		} else {
			fmt.Fprintf(stdout, "Would: archive done issues under %s/%s/\n", mainPath, f.HistoryDir)
			fmt.Fprintf(stdout, "Would: git worktree remove %s\n", wtPath)
			fmt.Fprintf(stdout, "Would: git branch -D %s\n", branch)
		}
		return nil
	}

	// ── 9b. Re-assert clean tree before the irreversible merge (#62 M1) ──────
	// The step-2 check ran before the pre-merge judges; a judge/hook may have
	// dirtied the tree since. Refuse here rather than merge-then-strand: a dirty
	// tree breaks both `gh pr merge`'s downstream `git switch main` and the
	// resume cleanup. (With read-only judges (#62 M2) the tree stays clean; this
	// is the defense-in-depth that makes the irreversible boundary safe.)
	// Only tracked changes refuse here, via the same assessDirty.Refuse()
	// decision as step 2 (#78) — untracked files were already surfaced above and
	// don't strand the merge. A pre-merge judge/hook that dirties a tracked file
	// is the real risk this re-assert defends against.
	if redirty, derr := worktreeDirty(mergeRunner); derr != nil {
		die(stderr, derr.Error())
	} else if reassess := assessDirty(redirty, f.IssuesDir, f.HistoryDir); reassess.Refuse() {
		fmt.Fprintf(stderr, "  %s[x]%s Tracked files dirtied after the initial check (likely a pre-merge judge/hook):\n", ansiRed, ansiReset)
		fmt.Fprintln(stderr, strings.Join(reassess.Blocking, "\n"))
		die(stderr, "review + commit (or discard) these changes, then re-run `sdlc merge` — refusing before the irreversible merge")
	}

	// ── 10. Find PR → merge, or resume an already-merged one (#62 M3) ────────
	prNumber, _ := ghClient.PRListForBranch(repo, branch)
	mergedExists := false
	if prNumber == "" {
		mergedExists, _ = ghClient.PRMergedForBranch(repo, branch)
	}
	merged := false
	switch decideMergeAction(prNumber, mergedExists) {
	case actionMergeOpen:
		cok(stderr, fmt.Sprintf("Open PR found: #%s", prNumber))
		cinfo(stderr, fmt.Sprintf("Merging PR #%s (%s) into main via GitHub...", prNumber, branch))
		if err := ghClient.PRMerge(repo, branch); err != nil {
			die(stderr, err.Error())
		}
		merged = true
	case actionResume:
		// Interrupted prior run: the PR merged server-side (irreversible) but the
		// local cleanup never finished. Resume it idempotently instead of erroring
		// on "no open PR" — re-running `sdlc merge` just completes the cleanup (#62 M3).
		cwarn(stderr, fmt.Sprintf("No open PR, but a MERGED PR exists for %s — resuming post-merge cleanup", branch))
		merged = true
	}

	if merged {
		// Post-merge local steps — run for both a fresh merge and a resume, and
		// idempotent either way. In-place: the merge is server-side, so switch
		// this checkout back to main (a no-op if a prior run already did) before
		// pulling the merged result.
		if inPlace {
			cinfo(stderr, "Switching to main...")
			if out, gerr := mergeRunner.Git("switch", "main"); gerr != nil {
				die(stderr, fmt.Sprintf("git switch main: %v\n%s", gerr, out))
			}
		}
		cinfo(stderr, "Pulling main...")
		if out, gerr := mergeRunner.GitInDir(mainPath, "pull"); gerr != nil {
			die(stderr, fmt.Sprintf("git -C %s pull: %v\n%s", mainPath, gerr, out))
		}
	} else {
		cwarn(stderr, fmt.Sprintf("No open or merged PR for branch %s", branch))
		if inPlace {
			die(stderr, "no PR for this branch — run `sdlc pr` first, then re-run `sdlc merge` (or `git switch main` to abandon the branch)")
		}
		if unmerged != "" {
			ans := mergePrompter.Ask("Would you like to create a pull request first? [Y/n] ", stderr)
			if ans != "n" && ans != "N" {
				die(stderr, "run `sdlc pr` to create a PR, then re-run `sdlc merge`")
			}
			ans2 := mergePrompter.Ask("Remove worktree without merging? [y/N] ", stderr)
			if ans2 != "y" && ans2 != "Y" {
				die(stderr, "aborted by operator")
			}
		}
		// Worktree path: falls through to archive + worktree removal regardless
		// — the shell does the same. If no unmerged, we silently proceed.
	}

	// ── 11. Archive done issues in MAIN worktree ────────────────────────────
	moves, err := archiveDoneIssuesInDir(stderr, repo, mainPath, f.IssuesDir, f.HistoryDir)
	if err != nil {
		die(stderr, err.Error())
	}
	if len(moves) > 0 {
		cinfo(stderr, "Committing archived history in main...")
		if out, gerr := mergeRunner.GitInDir(mainPath, archiveAddArgs(moves)...); gerr != nil {
			die(stderr, fmt.Sprintf("git -C %s add: %v\n%s", mainPath, gerr, out))
		}
		if out, gerr := mergeRunner.GitInDir(mainPath, "commit", "-m", "archive completed issues to history"); gerr != nil {
			die(stderr, fmt.Sprintf("git -C %s commit: %v\n%s", mainPath, gerr, out))
		}
		if out, gerr := mergeRunner.GitInDir(mainPath, "push"); gerr != nil {
			die(stderr, fmt.Sprintf("git -C %s push: %v\n%s", mainPath, gerr, out))
		}
	}

	// ── 12. Cleanup ─────────────────────────────────────────────────────────
	if inPlace {
		// Already switched to main + pulled above; just delete the merged branch.
		cinfo(stderr, fmt.Sprintf("Deleting merged branch %s...", branch))
		if out, gerr := mergeRunner.Git("branch", "-D", branch); gerr != nil {
			cwarn(stderr, fmt.Sprintf("git branch -D %s: %v\n%s", branch, gerr, out))
		}
		cok(stderr, "Done. You are on main.")
		return nil
	}

	cinfo(stderr, fmt.Sprintf("Removing worktree at %s...", wtPath))
	// Run worktree remove + branch delete from the MAIN worktree, since
	// removing the current worktree from within itself is undefined.
	// Best-effort (matches shell `|| true`).
	if out, gerr := mergeRunner.GitInDir(mainPath, "worktree", "remove", wtPath); gerr != nil {
		cwarn(stderr, fmt.Sprintf("git worktree remove %s: %v\n%s", wtPath, gerr, out))
	}
	if out, gerr := mergeRunner.GitInDir(mainPath, "branch", "-D", branch); gerr != nil {
		cwarn(stderr, fmt.Sprintf("git branch -D %s: %v\n%s", branch, gerr, out))
	}
	// .goto in the soon-to-be-removed worktree points back to main, so
	// `g` after re-creating the dir lands the operator in main.
	gotoPath := filepath.Join(wtPath, ".goto")
	if err := os.WriteFile(gotoPath, []byte(mainPath), 0o644); err != nil {
		cwarn(stderr, fmt.Sprintf(".goto write failed: %v", err))
	}
	cok(stderr, "Done. Run: g (to cd back to main)")
	return nil
}

// mergeNeedsTTY reports whether merge's confirmation prompts require an
// interactive terminal that isn't present — i.e. a bare stdin scan would
// block. True → refuse fast with a --yes hint instead of hanging. Pure so
// the decision is unit-testable without a real tty.
func mergeNeedsTTY(yes, dryRun, stdinIsTTY bool) bool {
	return !yes && !dryRun && !stdinIsTTY
}

// isInPlaceCheckout reports whether `git rev-parse --git-dir` indicates the
// primary working tree (in-place: a bare ".git") rather than a linked worktree
// (whose git-dir lives under ".git/worktrees/<name>"). Drives the in-place vs
// worktree merge topology (#51).
func isInPlaceCheckout(gitDir string) bool {
	return !strings.Contains(gitDir, "/worktrees/")
}

// archiveDoneIssuesInDir is the merge-side equivalent of push.go's
// archiveDoneIssues, but it scans + mutates inside the main worktree
// at mainPath (so the archive commit lands on main, not on the feature
// branch).
func archiveDoneIssuesInDir(stderr io.Writer, repo, mainPath, issuesDir, historyDir string) ([]preparedArchiveMove, error) {
	issuesFull := filepath.Join(mainPath, issuesDir)
	historyFull := filepath.Join(mainPath, historyDir)
	matches, _ := filepath.Glob(filepath.Join(issuesFull, "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"))
	sort.Strings(matches)
	var moves []preparedArchiveMove
	cinfo(stderr, fmt.Sprintf("Archiving completed issues to %s/...", historyDir))
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		fm, _, perr := issue.Parse(string(data))
		if perr != nil {
			continue
		}
		st, _ := issue.GetField(fm, "status")
		if !vocab.Issue().IsTerminal(st) {
			continue
		}
		// Merge target's shell DOES NOT call gh issue close — only push:
		// closes GH issues. We mirror that. (Rationale: PR merge itself
		// closes the linked GH issue via the "Fixes #N" body, so a second
		// `gh issue close` would be redundant.) Repo param kept in
		// signature for API symmetry with push's archive helper.
		_ = repo
		if err := os.MkdirAll(historyFull, 0o755); err != nil {
			return moves, fmt.Errorf("mkdir %s: %v", historyFull, err)
		}
		base := filepath.Base(p)
		dest := filepath.Join(historyFull, base)
		fmt.Fprintf(stderr, "  Moving %s to %s/\n", base, historyDir)
		if err := os.Rename(p, dest); err != nil {
			return moves, fmt.Errorf("mv %s → %s: %v", p, dest, err)
		}
		// Record paths relative to mainPath: GitInDir(mainPath, "add", …)
		// resolves them from the main worktree root, so an absolute path here
		// would silently miss the staged move.
		moves = append(moves, preparedArchiveMove{
			IssuePath:   filepath.Join(issuesDir, base),
			HistoryPath: filepath.Join(historyDir, base),
		})
	}
	return moves, nil
}
