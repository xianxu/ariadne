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
//  3. issue branch published on its explicit remote ref
//  4. branch not ahead of its remote ref
//  5. pre-merge PUBLISH GATE (#160) — the deterministic reviewed-HEAD-unchanged
//     invariant (refuse unless HEAD is unchanged since the codecomplete issues'
//     `sdlc close`); NO LLM (all LLM review is close-time). Skippable --no-judge.
//  6. resolve topology (in-place vs worktree)
//  7. show unmerged commits (informational)
//  8. not-done issue warn (vs main)
//  9. interactive confirmation (skippable with --yes)
//     9b. re-assert no tracked dirt before the irreversible merge — refuse if a
//     gate/hook dirtied a tracked file since step 2 (#62 M1; never cross dirty)
//  10. gh pr merge (server-side), OR resume an already-merged PR if a prior run
//     was interrupted (#62 M3) → in-place: switch main; both: pull main
//     10.5 publish flip (#160): codecomplete → done on main (before archiving)
//  11. archive done/wontfix/punt issues into history/ (in the main checkout)
//  12. cleanup — in-place: branch delete; worktree: worktree remove + branch delete + .goto
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// mergeFlags holds the parsed flag values for the merge subcommand.
type mergeFlags struct {
	Context    context.Context
	Lock       func(func() error) error
	Branch     string
	Yes        bool
	NoJudge    bool
	NoValidate bool
	NoCI       bool
	DryRun     bool
	IssuesDir  string
	HistoryDir string
	PlansDir   string
}

func (f *mergeFlags) withLock(run func() error) error {
	if f.Lock == nil { // Direct package tests supply an unlocked runner.
		return run()
	}
	return f.Lock(run)
}

// mergeRunner is the package-level runner for merge (test seam). Type
// lives in runner.go.
var mergeRunner gitRunner = execGitRunner{}

// mergePrompter is a tiny indirection over stdin so tests can drive the
// confirmation prompts deterministically. Production wraps os.Stdin.
var mergePrompter prompter = stdinPrompter{}

// runPublishGateFn is the package-level seam for merge's step-5 pre-merge publish
// gate (#160 — the deterministic reviewed-HEAD-unchanged invariant that replaced
// the LLM judges). Production points at runPublishGate. Tests swap it for a stub —
// most usefully one that DIRTIES the worktree, to prove step 9b re-checks
// cleanliness before the irreversible merge (#62 M1 / #63); it runs after step 2's
// clean check and before 9b's re-check, the window a real dirtying hook would occupy.
var runPublishGateFn = runPublishGate

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
	cmd := markManualLockCommand(&cobra.Command{
		Use:           "merge",
		Short:         "Merge the current branch (in-place or worktree) via GitHub PR, archive done issues, clean up",
		Long:          "Placeholder — replaced by helptext.MustGet(\"merge\") in main.go.",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			guardSpineRepo(cmd.ErrOrStderr()) // #176 lifecycle guard
			f.Context = cmd.Context()
			f.Lock = func(run func() error) error { return withRequiredRepoTransactionLock(cmd, run) }
			return runMerge(cmd.OutOrStdout(), cmd.ErrOrStderr(), &f)
		},
	})
	cmd.Flags().StringVar(&f.Branch, "branch", "", "resume landing this issue branch from its workspace resting branch")
	cmd.Flags().BoolVar(&f.Yes, "yes", false, "skip the final irreversible-merge confirmation AND not-done warn — REQUIRED for non-interactive/agent runs (merge fail-fasts before the publish gate when stdin is not a terminal)")
	cmd.Flags().BoolVar(&f.NoJudge, "no-judge", false, "skip the pre-merge publish gate — #160 reviewed-HEAD-unchanged invariant (emergency-only)")
	cmd.Flags().BoolVar(&f.NoValidate, "no-validate", false, "skip the #124 instance-conformance gate (escape hatch — announced loudly)")
	cmd.Flags().BoolVar(&f.NoCI, "no-ci", false, "skip the #309 CI gate (red or pending checks on the PR head) — for a broken CI system only, announced loudly")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "print would-be operations; do not merge or clean up")
	cmd.Flags().StringVar(&f.IssuesDir, "issues-dir", envOr("WF_ISSUES_DIR", "workshop/issues"), "directory holding issue files")
	cmd.Flags().StringVar(&f.HistoryDir, "history-dir", envOr("WF_HISTORY_DIR", "workshop/history"), "directory for archived issues")
	cmd.Flags().StringVar(&f.PlansDir, "plans-dir", envOr("WF_PLANS_DIR", "workshop/plans"), "directory holding durable plans + review sidecars (archived with the issue, #143)")
	return cmd
}

// runMerge dispatches the merge workflow.
// worktreeDirty returns the parsed `git status` entries (none = clean) via the
// runner, or an error if git status itself fails. Checked at the
// start of merge AND — per #62 — re-checked immediately before the irreversible
// `gh pr merge`: a pre-merge judge/hook can dirty the tree after the initial
// check, and the post-merge `git switch main` then refuses, stranding the merge
// (remote merged, local stuck). Re-asserting here converts that into a clean
// pre-merge refusal.
func worktreeDirty(r gitRunner) ([]gitx.StatusEntry, error) {
	out, err := r.Git("status", "--porcelain=v1", "-z")
	if err != nil {
		return nil, fmt.Errorf("git status: %v\n%s", err, out)
	}
	return gitx.ParseStatusZ(out)
}

// statusLine renders a status entry as porcelain v1 prints it, for messages.
func statusLine(e gitx.StatusEntry) string {
	if e.Orig != "" {
		return e.XY + " " + e.Orig + " -> " + e.Path
	}
	return e.XY + " " + e.Path
}

// dirtyAssessment splits a `git status --porcelain` result into the classes that
// matter to the merge guard: Blocking (tracked code modifications — staged,
// modified, renamed, deleted, unmerged), Untracked (`??` non-tracker lines), and
// Tracker (workshop issue/history markdown, #82 M2).
type dirtyAssessment struct {
	Blocking     []string // tracked code changes that MUST block the merge
	Untracked    []string // untracked files — surfaced, but not a blocker (#78)
	Tracker      []string // workshop/issues|history/*.md — tracker state (#82 M2)
	TrackerPaths []string // the paths Tracker names, rename sources included (#196)
}

// assessDirty classifies parsed status entries. Pure (mirrors decideMergeAction's
// extracted-decision pattern). A dirty tracked file makes the post-merge `git
// switch main` / `git pull` refuse, stranding the server-side merge, so tracked
// CODE always blocks. Untracked `??` files survive the switch (#78). Tracker files
// are never code contention (#82 M2), but git does not know that, so they are
// bucketed apart for mergeRefusal to ask whether git would carry them (#196).
func assessDirty(entries []gitx.StatusEntry, issuesDir, historyDir string) dirtyAssessment {
	var d dirtyAssessment
	for _, e := range entries {
		line := statusLine(e)
		// Tracker classification is by PATH and comes first, so a dirty issue
		// file lands in Tracker regardless of its tracked/untracked status code.
		if isTrackerPath(e.Path, issuesDir, historyDir) || isTrackerPath(e.Orig, issuesDir, historyDir) {
			d.Tracker = append(d.Tracker, line)
			d.TrackerPaths = append(d.TrackerPaths, e.Path)
			if e.Orig != "" {
				d.TrackerPaths = append(d.TrackerPaths, e.Orig)
			}
			continue
		}
		if e.XY == "??" {
			d.Untracked = append(d.Untracked, line)
		} else {
			d.Blocking = append(d.Blocking, line)
		}
	}
	return d
}

// strandingTracker returns the dirty tracker paths git would refuse on the in-place
// switch (main vs HEAD) or the pull after it (main vs mainRef); paths are porcelain's
// root-relative literals, hence :(top,literal).
func strandingTracker(r gitRunner, paths []string, mainRef string) ([]string, error) {
	seen := map[string]bool{}
	var out []string
	for _, rng := range [][2]string{{"main", "HEAD"}, {"main", mainRef}} {
		args := []string{"diff", "--name-only", "--no-renames", "-z", rng[0], rng[1], "--"}
		for _, p := range paths {
			args = append(args, ":(top,literal)"+p)
		}
		raw, err := r.Git(args...)
		if err != nil {
			return nil, fmt.Errorf("git diff %s %s: %v\n%s", rng[0], rng[1], err, raw)
		}
		for _, p := range strings.Split(string(raw), "\x00") {
			if p != "" && !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out, nil
}

// mergeRefusal is the one merge-block decision for step 2 and 9b (#78): tracked
// code, plus in place the tracker paths that strand the switch (#196).
func mergeRefusal(r gitRunner, d dirtyAssessment, inPlace bool, mainRef string) ([]string, error) {
	refuse := append([]string(nil), d.Blocking...)
	if !inPlace || len(d.TrackerPaths) == 0 { // no paths: a bare `git diff --` would list every file
		return refuse, nil
	}
	stranding, err := strandingTracker(r, d.TrackerPaths, mainRef)
	if err != nil {
		return nil, err
	}
	for _, p := range stranding {
		refuse = append(refuse, "tracker file differs from main: "+p)
	}
	return refuse, nil
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

// mergeAction is what step 10 should do, given the PR state.
type mergeAction int

const (
	actionMergeOpen     mergeAction = iota // an open PR exists → merge it (irreversible)
	actionResume                           // no open PR, a merged one, branch fully merged → resume cleanup (#62 M3)
	actionResumeBlocked                    // merged PR, but the branch has commits not in base → refuse (#148, likely a reused branch name)
	actionNoPR                             // neither → create-PR / abandon path
)

// countUnmerged returns how many commits are on head but not in base
// (`git rev-list --count base..head`), via the injected runner so it's fakeable
// (#148). Errors on a failed git call or non-numeric output — the caller treats
// an error as "can't verify" and fails safe (refuses) rather than defaulting to 0.
func countUnmerged(r gitRunner, base, head string) (int, error) {
	out, err := r.Git("rev-list", "--count", base+".."+head)
	if err != nil {
		return 0, fmt.Errorf("git rev-list --count %s..%s: %w\n%s", base, head, err, out)
	}
	trimmed := strings.TrimSpace(string(out))
	n, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, fmt.Errorf("unexpected `rev-list --count %s..%s` output %q: %w", base, head, trimmed, err)
	}
	return n, nil
}

// decideMergeAction picks the step-10 action from the open-PR number, whether a
// merged PR already exists for the branch, and how many commits the branch has
// that are NOT in the base. Pure + testable; the irreversible PRMerge stays in
// runMerge and the count IO stays in the caller (countUnmerged). Resume
// (merged-but-not-open, fully merged) is how a re-run recovers an interrupted
// merge (#62 M3); a merged PR whose branch still carries unmerged commits is a
// reused branch name (#148) — refuse instead of silently cleaning it up.
func decideMergeAction(openPRNumber string, mergedExists bool, unmergedCount int) mergeAction {
	switch {
	case openPRNumber != "":
		return actionMergeOpen
	case mergedExists && unmergedCount > 0:
		return actionResumeBlocked
	case mergedExists:
		return actionResume
	default:
		return actionNoPR
	}
}

func runMerge(stdout, stderr io.Writer, f *mergeFlags) (err error) {
	ctx := commandContext(f.Context)
	target, err := resolveLandingTarget(boundedLandingGit(ctx, mergeRunner))
	if err != nil {
		return err
	}
	if target != nil {
		return runDurableMerge(stdout, stderr, f, *target)
	}
	if f.Branch != "" {
		return fmt.Errorf("--branch recovery requires an addressable primary or numbered slot")
	}

	// ── 1. Refuse if main / empty branch ────────────────────────────────────
	branch := gitx.Capture("branch", "--show-current")
	if branch == "" || branch == "main" {
		die(stderr, fmt.Sprintf("sdlc merge must be run from a worktree branch (current: %s)", valueOr(branch, "(detached)")))
	}
	branchHead := gitx.Capture("rev-parse", "HEAD")
	cinfo(stderr, fmt.Sprintf("Branch: %s", branch))

	// ── 2. No uncommitted TRACKED changes ───────────────────────────────────
	// Untracked files don't block: they survive `git switch main`, so they can't
	// strand the post-merge cleanup the way a dirty tracked file does (#78). They
	// are surfaced as a warning. Only an in-place checkout switches, hence inPlace (#196).
	inPlace := isInPlaceCheckout(gitx.Capture("rev-parse", "--git-dir"))
	dirty, err := worktreeDirty(mergeRunner)
	if err != nil {
		die(stderr, err.Error())
	}
	assessment := assessDirty(dirty, f.IssuesDir, f.HistoryDir)
	refusal, err := mergeRefusal(mergeRunner, assessment, inPlace, "origin/main")
	if err != nil {
		die(stderr, err.Error())
	}
	if len(refusal) > 0 {
		fmt.Fprintf(stderr, "  %s[x]%s Uncommitted tracked changes found — cannot merge\n", ansiRed, ansiReset)
		fmt.Fprintln(stderr, strings.Join(refusal, "\n"))
		die(stderr, "commit or stash these tracked changes before merging (git would refuse them on the post-merge switch to main)")
	}
	if len(assessment.Untracked) > 0 {
		cwarn(stderr, fmt.Sprintf("%d untracked file(s) present — not blocking the merge (they survive the branch switch):", len(assessment.Untracked)))
		fmt.Fprintln(stderr, strings.Join(assessment.Untracked, "\n"))
	}
	if len(assessment.Tracker) > 0 {
		why := map[bool]string{true: "they match main, so the switch carries them", false: "this linked worktree is never switched"}[inPlace]
		cwarn(stderr, fmt.Sprintf("%d tracker file(s) dirty — not blocking the merge (%s; #82, #196):", len(assessment.Tracker), why))
		fmt.Fprintln(stderr, strings.Join(assessment.Tracker, "\n"))
	}
	cok(stderr, "No uncommitted tracked changes")

	// Fail fast — before the publish gate + irreversible merge — if the confirmation
	// prompts (steps 8-9) can't be answered. They read os.Stdin, and in a
	// non-tty agent/background context a bare scan blocks forever; convert
	// that hang into a clear next-action.
	if mergeNeedsTTY(f.Yes, f.DryRun, isTTY(os.Stdin)) {
		die(stderr, "sdlc merge needs interactive confirmation, but stdin is not a terminal.\n"+
			"  Re-run with --yes to merge non-interactively (skips the confirm + not-done prompts),\n"+
			"  or run it from a terminal.")
	}

	if !f.DryRun {
		root, rootErr := gitx.RepoTopLevel()
		if rootErr != nil {
			return rootErr
		}
		queueCtx, q, startErr := startQueueOperation(ctx, root, branchHead, []string{"merge", "landing"})
		if startErr != nil {
			return startErr
		}
		ctx = queueCtx
		defer func() {
			reason := queueInterruptionReason(err)
			if finishErr := q.finish(ctx, reason); finishErr != nil {
				if err == nil {
					err = finishErr
				} else {
					fmt.Fprintf(stderr, "queue retirement: %v\n", finishErr)
				}
			}
		}()
		if err = q.waitAdmission(ctx, stderr); err != nil {
			return err
		}
		ctx = withQueueAdmitted(ctx)
		localFlags := *f
		f = &localFlags
		f.Context = ctx
	}

	repo, err := detectRepo()
	if err != nil {
		die(stderr, err.Error())
	}

	// ── 3. Branch pushed + HEAD synced ──────────────────────────────────────
	// Key off the remote-tracking ref (origin/<branch>), NOT @{u}: a push
	// updates origin/<branch> even when it can't write the local
	// upstream-tracking config (a sandbox that blocks .git/config, or a plain
	// `git push` without -u), so requiring @{u} spuriously refuses a branch
	// that is genuinely pushed (and has an open PR).
	remoteRef := "origin/" + branch
	// A completed PR may have deleted its remote branch before this command
	// could finish local archive/cleanup. Recovery is proven from main below;
	// an open PR still requires the exact pushed remote head.
	earlyPR, _ := ghClient.PRListForBranch(repo, branch)
	earlyMerged := false
	if earlyPR == "" {
		earlyMerged, _ = ghClient.PRMergedForBranch(repo, branch)
	}
	recoveringMerged := earlyPR == "" && earlyMerged
	if !recoveringMerged && gitx.Capture("rev-parse", "--verify", "--quiet", remoteRef) == "" {
		fmt.Fprintf(stderr, "  %s[x]%s %s is not on origin\n", ansiRed, ansiReset, branch)
		die(stderr, fmt.Sprintf("push the branch first (e.g. sdlc pr, or git push origin %s)", branch))
	}

	// ── 4. No local commits missing from the remote branch ──────────────────
	ahead := 0
	if !recoveringMerged {
		aheadStr := gitx.Capture("rev-list", "--count", remoteRef+"..HEAD")
		ahead, _ = strconv.Atoi(aheadStr)
	}
	if ahead > 0 {
		fmt.Fprintf(stderr, "  %s[x]%s %d local commit(s) not yet on %s\n",
			ansiRed, ansiReset, ahead, remoteRef)
		die(stderr, "push your branch before merging — run `sdlc pr`, then re-run `sdlc merge`. "+
			"(merge is server-side: a fix you committed for a failed pre-merge gate must reach origin first.)")
	}
	if !recoveringMerged {
		cok(stderr, fmt.Sprintf("Branch pushed; HEAD synced with %s", remoteRef))
	}

	// ── 4.5 Instance-conformance gate (#124) ────────────────────────────────
	// Deterministic, separate from the judges. Runs after the branch is synced to
	// origin (merge is server-side) so it checks the same tree that will merge.
	if !f.NoValidate {
		if err := validateChangedInstancesFn("", gitx.DiffBase(), "", nounGates(f.IssuesDir), stdout, stderr); err != nil {
			die(stderr, err.Error()+"\n  → fix and `sdlc pr` (merge is server-side), or --no-validate to bypass.")
		}
	} else {
		cwarn(stderr, "⚠️  --no-validate: SKIPPING the instance-conformance gate (#124) — modeled frontmatter and issue sections NOT verified before main. Escape hatch: say why in your commit/log.")
	}

	// ── 4.6 Duplicate-id gate (#213) ────────────────────────────────────────
	// The last point where an id collision is still repairable. Allocation now
	// reads the trunk, but a branch created before that fix — or with a failed
	// fetch behind it — can still carry a reused id, and nothing else would ever
	// object: the two files have different slugs, so they are different paths and
	// git merges both cleanly.
	if !f.NoValidate {
		if err := f.withLock(func() error { return refuseDuplicateIssueIDs(stderr, f.IssuesDir, f.HistoryDir, mergeRunner) }); err != nil {
			die(stderr, err.Error())
		}
	}

	// ── 5. Pre-merge publish gate (#160) — deterministic, NO LLM ─────────────
	// All LLM review is now close-time (the boundary review). The publish gate
	// enforces the reviewed-HEAD-unchanged invariant: refuse unless HEAD is
	// unchanged since the codecomplete issues' `sdlc close`. (Replaces the old
	// plan/specs/lessons pre-merge judges — #142 folded here.)
	if !f.NoJudge {
		if err := runPublishGateFn(commandContext(f.Context), gitx.DiffBase(), f.IssuesDir, stderr); err != nil {
			if f.DryRun {
				cwarn(stderr, fmt.Sprintf("dry-run: publish gate WOULD refuse: %v", err))
			} else {
				die(stderr, err.Error()+"\n  (the fix must reach origin — merge is server-side.)")
			}
		}
	} else {
		cwarn(stderr, "--no-judge: skipping the pre-merge publish gate (#160 reviewed-HEAD-unchanged invariant)")
	}
	if err := f.withLock(func() error {
		return revalidateLegacyMerge(branch, branchHead, remoteRef, inPlace, !recoveringMerged, f)
	}); err != nil {
		die(stderr, err.Error())
	}

	// ── 5.5 Tracker ownership (#252) ─────────────────────────────────────────
	// The server-side merge may squash or rebase, leaving the evidence commits
	// off main, so the closes this branch owns are selected now, from the
	// branch, and completed after the landing is pulled.
	var trackerOwned []ownedCompletion
	var trackerEnvForMerge *trackerEnv
	err = f.withLock(func() error {
		var err error
		trackerOwned, trackerEnvForMerge, err = branchOwnedCompletions(commandContext(f.Context), f.IssuesDir)
		return err
	})
	if err != nil {
		die(stderr, fmt.Sprintf("tracker completions: %v", err))
	}

	// ── 6. Resolve merge topology: in-place vs worktree ─────────────────────
	// In-place = the primary checkout sitting on a feature branch (main is
	// reached by switching here). Worktree = a linked worktree (main lives in
	// a separate dir). A linked worktree's git-dir is under .git/worktrees/.
	wtPath, _ := gitx.RepoTopLevel()
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

	// ── 7. Show unmerged commits ────────────────────────────────────────────
	base := gitx.BranchPoint()
	if base == "" {
		base = "main"
	}
	unmergedOut, _ := mergeRunner.Git("log", base+"..HEAD", "--oneline")
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
	notDone, nerr := touchedIssuesNotDone(ctx, "main", f.IssuesDir, mergeRunner)
	if nerr != nil {
		cwarn(stderr, fmt.Sprintf("could not scan touched issues: %v", nerr))
	}
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
	// The step-2 check ran before the step-5 publish gate; a gate/hook may have
	// dirtied the tree since. Refuse here rather than merge-then-strand: a dirty
	// tree breaks both `gh pr merge`'s downstream `git switch main` and the
	// resume cleanup. (With read-only judges (#62 M2) the tree stays clean; this
	// is the defense-in-depth that makes the irreversible boundary safe.)
	// Same mergeRefusal decision as step 2 (#78); in place, against a freshly
	// fetched origin/main so main's newer tracker edits refuse here (#196).
	if inPlace {
		if ferr := f.withLock(func() error {
			out, err := mergeRunner.Git("fetch", "origin", "main")
			if err != nil {
				return fmt.Errorf("git fetch origin main (to check dirty tracker files against it): %v\n%s", err, out)
			}
			return nil
		}); ferr != nil {
			die(stderr, ferr.Error())
		}
	}
	redirty, derr := worktreeDirty(mergeRunner)
	if derr != nil {
		die(stderr, derr.Error())
	}
	rerefusal, derr := mergeRefusal(mergeRunner, assessDirty(redirty, f.IssuesDir, f.HistoryDir), inPlace, "origin/main")
	if derr != nil {
		die(stderr, derr.Error())
	}
	if len(rerefusal) > 0 {
		fmt.Fprintf(stderr, "  %s[x]%s Tracked files that would strand the post-merge switch (a pre-merge hook, or main moved under a tracker file):\n", ansiRed, ansiReset)
		fmt.Fprintln(stderr, strings.Join(rerefusal, "\n"))
		die(stderr, "discard (`git restore`) or move aside these changes, then re-run `sdlc merge` — refusing before the irreversible merge (a commit now would change the reviewed HEAD)")
	}

	// ── 10. Find PR → merge, or resume an already-merged one (#62 M3) ────────
	prNumber, _ := ghClient.PRListForBranch(repo, branch)
	mergedExists := false
	if prNumber == "" {
		mergedExists, _ = ghClient.PRMergedForBranch(repo, branch)
	}
	// #148: only a MERGED PR (no open one) means "resume cleanup" — but a reused
	// branch name can carry commits that the merged PR never shipped. Verify the
	// branch is genuinely fully merged before we switch/pull/archive/delete it.
	// origin/main is stale here (the flow doesn't pull until AFTER deciding, below),
	// so fetch the base first, then count commits on the branch not in it. Any
	// error → die (fail-safe: never clean up a branch we couldn't verify).
	unmergedCount := 0
	if prNumber == "" && mergedExists {
		if ferr := f.withLock(func() error {
			out, err := mergeRunner.Git("fetch", "origin", "main")
			if err != nil {
				return fmt.Errorf("git fetch origin main (to verify %s is fully merged): %v\n%s", branch, err, out)
			}
			return nil
		}); ferr != nil {
			die(stderr, ferr.Error())
		}
		n, cerr := countUnmerged(mergeRunner, "origin/main", branchHead)
		if cerr != nil {
			die(stderr, fmt.Sprintf("couldn't verify %s is fully merged — refusing to clean up: %v", branch, cerr))
		}
		unmergedCount = n
	}
	merged := false
	switch decideMergeAction(prNumber, mergedExists, unmergedCount) {
	case actionMergeOpen:
		cok(stderr, fmt.Sprintf("Open PR found: #%s", prNumber))
		// #309: the CI gate reads the checks on the pushed head (step 4 proved
		// local = origin), and the merge is pinned to that head.
		head := gitx.Capture("rev-parse", "--verify", remoteRef)
		prNum, perr := strconv.Atoi(prNumber)
		if head == "" || perr != nil {
			die(stderr, fmt.Sprintf("CI gate: cannot pin the merge (head %q, PR %q)", head, prNumber))
		}
		gitCall := func(args ...string) (string, error) {
			out, err := mergeRunner.Git(args...)
			return string(out), err
		}
		// The gate decides from main's tree whether the guard's verdict is
		// required, so it reads a freshly fetched main, never a stale ref.
		var main string
		err := f.withLock(func() error {
			var err error
			main, err = requireLegacyMainIntegrated(gitCall, head)
			return err
		})
		if err != nil {
			die(stderr, err.Error())
		}
		root := gitx.Capture("rev-parse", "--show-toplevel")
		localDeps := localRunDeps{sysvObserver: nativeSysVObserver(), GH: ghClient, Repo: repo, Root: root, Git: gitCall, CleanupGit: localCheckCleanupGit(ctx, mergeRunner, root), PR: prNum, Stderr: stderr, FlakeRegistry: gateFlakeRegistry(root)}
		localRun := newLocalCheckRunner(localDeps)
		if err := runCIGate(ctx, ghClient, repo, prNum, main, head, gitCall, localRun, f.NoCI, stderr); err != nil {
			die(stderr, err.Error())
		}
		landingChecks := landingCheckOwner{}
		if err := landingChecks.run(ctx, gitCall, main, head, []string{branchHead}, !f.NoCI, localRun, func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
			return runLocalCheck(ctx, localDeps, c, sha)
		}); err != nil {
			die(stderr, err.Error())
		}
		if err := f.withLock(func() error {
			if err := revalidateLegacyMerge(branch, branchHead, remoteRef, inPlace, true, f); err != nil {
				return err
			}
			if _, err := requireLegacyMainIntegrated(gitCall, head); err != nil {
				return err
			}
			prHead, err := ghClient.PRHeadOID(ctx, repo, prNum)
			if err != nil {
				return err
			}
			if prHead != head {
				return fmt.Errorf("PR head moved during local checks: expected %s, got %s; retry sdlc merge", head, prHead)
			}
			cinfo(stderr, fmt.Sprintf("Merging PR #%s (%s) into main via GitHub...", prNumber, branch))
			return ghClient.PRMerge(repo, branch, head)
		}); err != nil {
			die(stderr, err.Error())
		}
		merged = true
	case actionResume:
		// Interrupted prior run: the PR merged server-side (irreversible) but the
		// local cleanup never finished. Resume it idempotently instead of erroring
		// on "no open PR" — re-running `sdlc merge` just completes the cleanup (#62 M3).
		cwarn(stderr, fmt.Sprintf("No open PR, but a MERGED PR exists for %s — resuming post-merge cleanup", branch))
		merged = true
	case actionResumeBlocked:
		// #148: the branch has commits not in main despite a merged PR — almost
		// always a reused branch name (its old work shipped via that PR; new work
		// piled on the same name). Cleaning up here would delete the branch and
		// strand the new commits on origin, main never advancing. Refuse loudly —
		// BEFORE any switch/pull/archive/delete — so nothing is lost.
		die(stderr, fmt.Sprintf(
			"branch '%s' has %d commit(s) not in main despite a merged PR — likely a reused branch name.\n"+
				"  Rename the branch to a fresh name (e.g. <issue>-<short-slug>) and run `sdlc pr`, then `sdlc merge`.\n"+
				"  Not switching / pulling / archiving / deleting — your %d commit(s) are safe on %s.",
			branch, unmergedCount, unmergedCount, remoteRef))
	}
	if !merged {
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
	}

	err = f.withLock(func() error {
		if err := revalidateLegacyMerge(branch, branchHead, remoteRef, inPlace, false, f); err != nil {
			die(stderr, err.Error())
		}
		if merged {
			// Post-merge local steps — run for both a fresh merge and a resume, and
			// idempotent either way. In-place: the merge is server-side, so switch
			// this checkout back to main (a no-op if a prior run already did) before
			// pulling the merged result.
			cinfo(stderr, "Fetching and fast-forwarding main...")
			if atMain, err := materializeMergedMain(ctx, mergeRunner, mainPath, f.IssuesDir, inPlace); err != nil {
				die(stderr, strandedMessage(branch, err.Error(), inPlace && atMain))
			}
		}

		// ── 10.5 Publish flip (#160): codecomplete → done on main ────────────────
		// The merged issues arrive on main at `codecomplete`; flip them to done (the
		// deterministic publish flip) BEFORE archiving — archiveDoneIssuesInDir keys on
		// IsTerminal, and codecomplete is active. Actuals were set at close, so the
		// done-guard holds. The flip is captured by the archive commit below (the
		// flipped files move to history).
		// Only a confirmed landing completes a card: a PR merged now, or found merged
		// on resume. An abandoned branch (removed without merging) completes nothing.
		if trackerEnvForMerge != nil && merged {
			out, lerr := mergeRunner.GitInDir(mainPath, "rev-parse", "--verify", "HEAD^{commit}")
			if lerr != nil {
				die(stderr, fmt.Sprintf("resolve the landed main commit: %v\n%s", lerr, out))
			}
			landed := strings.TrimSpace(string(out))
			for _, oc := range trackerOwned {
				if cerr := completeOnCard(trackerEnvForMerge, oc, landed); cerr != nil {
					die(stderr, fmt.Sprintf("complete #%s on its card: %v\n  retry `sdlc merge` or `sdlc issue recovery reconcile --issue %s`", issue.CLIRef(oc.ID), cerr, issue.CLIRef(oc.ID)))
				}
			}
		}
		if flipped, ferr := publishCodecompleteIssues(ctx, filepath.Join(mainPath, f.IssuesDir)); ferr != nil {
			die(stderr, fmt.Sprintf("publish flip (codecomplete → done): %v", ferr))
		} else if len(flipped) > 0 {
			cinfo(stderr, fmt.Sprintf("Published %d issue(s): codecomplete → done", len(flipped)))
		}

		// ── 11. Archive done issues in MAIN worktree ────────────────────────────
		moves, err := archiveDoneIssuesInDir(ctx, stderr, repo, mainPath, f.IssuesDir, f.HistoryDir, f.PlansDir)
		if err != nil {
			die(stderr, err.Error())
		}
		if len(moves) > 0 {
			cinfo(stderr, "Committing archived history in main...")
			if out, gerr := mergeRunner.GitInDir(mainPath, archiveAddArgs(moves)...); gerr != nil {
				die(stderr, fmt.Sprintf("git -C %s add: %v\n%s", mainPath, gerr, out))
			}
			commitArgs, cerr := archiveCommitArgs(agentid.StampCurrent(archiveCommitMessage, ctx), moves)
			if cerr != nil {
				die(stderr, cerr.Error())
			}
			if err := verifyArchivedAssetMoves(ctx, mainPath, moves); err != nil {
				die(stderr, err.Error())
			}
			if err := verifyArchivedSideMoves(mainPath, moves); err != nil {
				die(stderr, err.Error())
			}
			if out, gerr := mergeRunner.GitInDir(mainPath, commitArgs...); gerr != nil {
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
	})
	// ── 13. Landing tick (#373) — after the lock: warn-only IO on other
	// repositories' origins ─────────────────────────────────────────────────────
	if err == nil && merged {
		tickOnLanding(ctx, stdout, stderr, mainPath, branch, prNumber)
	}
	return err
}

// revalidateLegacyMerge protects the older, non-durable merge path each time
// it reacquires the lock after an unlocked check or confirmation window.
func revalidateLegacyMerge(branch, head, remoteRef string, inPlace, requireRemote bool, f *mergeFlags) error {
	if got := gitx.Capture("branch", "--show-current"); got != branch {
		return fmt.Errorf("branch moved during merge: expected %s, got %s; retry sdlc merge", branch, got)
	}
	if got := gitx.Capture("rev-parse", "HEAD"); got != head {
		return fmt.Errorf("HEAD moved during merge: expected %s, got %s; retry sdlc merge", head, got)
	}
	if requireRemote {
		if out, err := mergeRunner.Git("fetch", "origin", branch); err != nil {
			return fmt.Errorf("fetch %s: %v\n%s", branch, err, out)
		}
		if got := gitx.Capture("rev-parse", "--verify", remoteRef); got != head {
			return fmt.Errorf("remote head moved during merge: expected %s, got %s; retry sdlc merge", head, got)
		}
	}
	entries, err := worktreeDirty(mergeRunner)
	if err != nil {
		return err
	}
	refusal, err := mergeRefusal(mergeRunner, assessDirty(entries, f.IssuesDir, f.HistoryDir), inPlace, "origin/main")
	if err != nil {
		return err
	}
	if len(refusal) > 0 {
		return fmt.Errorf("worktree changed during merge: %s; discard or move aside these changes and retry — refusing before the irreversible merge", strings.Join(refusal, ", "))
	}
	return nil
}

// strandedMessage names the merged-but-stranded state and its resuming re-run (#62 M3, #196).
func strandedMessage(branch, cause string, onMain bool) string {
	back := ""
	if onMain {
		back = fmt.Sprintf("`git switch %s`, then ", branch)
	}
	return fmt.Sprintf("%s\n  The PR for %s IS MERGED on GitHub, but the local cleanup (pull, publish flip,\n"+
		"  archive, branch delete) did not run. Fix what git reports (move aside or `git restore`\n"+
		"  files it would overwrite; do not commit them, the branch is merged), then %sre-run\n"+
		"  `sdlc merge --yes`: it resumes.",
		cause, branch, back)
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
func archiveDoneIssuesInDir(ctx context.Context, stderr io.Writer, repo, mainPath, issuesDir, historyDir, plansDir string) ([]preparedArchiveMove, error) {
	issuesFull := filepath.Join(mainPath, issuesDir)
	historyFull := filepath.Join(mainPath, historyDir)
	plansFull := filepath.Join(mainPath, plansDir)
	refs, err := scanIssueFiles(ctx, "", issuesFull, nil)
	if err != nil {
		return nil, err
	}
	var moves []preparedArchiveMove
	// #181: issues subdir on BOTH legs — the rename dest (mainPath-joined)
	// and the recorded mainPath-relative path. This duplicates push.go's
	// archiveDoneIssues dest logic (pre-existing two-write-site debt); both
	// derive via ArchiveSubdir.
	issuesSubFull := vocab.ArchiveSubdir(historyFull, vocab.ArchiveIssues)
	issuesSubRec := vocab.ArchiveSubdir(historyDir, vocab.ArchiveIssues)
	cinfo(stderr, fmt.Sprintf("Archiving completed issues to %s/...", issuesSubRec))
	mirrors := newArchiveMirrors(ctx, stderr)
	for _, ref := range terminalIssueFiles(refs) {
		// Merge target's shell DOES NOT call gh issue close — only push:
		// closes GH issues. We mirror that. (Rationale: PR merge itself
		// closes the linked GH issue via the "Fixes #N" body, so a second
		// `gh issue close` would be redundant.) Repo param kept in
		// signature for API symmetry with push's archive helper.
		_ = repo
		base := filepath.Base(ref.Path)
		dest := archiveDestination(historyFull, vocab.ArchiveIssues, base)
		body, err := os.ReadFile(ref.Path)
		if err != nil {
			return moves, err
		}
		assetMoves, err := prepareAssetArchive(ctx, mainPath, filepath.Join(issuesDir, base), archiveDestination(historyDir, vocab.ArchiveIssues, base), body, gitSrcUntracked(func(a ...string) ([]byte, error) { return mergeRunner.GitInDir(mainPath, a...) }))
		if err != nil {
			return moves, err
		}
		// #424: the issue's and its plans' side folders move with them;
		// prepared (and refused) before anything moves.
		sideDocs := issueArchiveDocs(filepath.Join(issuesDir, base), archiveDestination(historyDir, vocab.ArchiveIssues, base), plansFull, plansDir, historyDir)
		sideMoves, err := prepareSideArchive(mainPath, sideDocs, gitSrcUntracked(func(a ...string) ([]byte, error) { return mergeRunner.GitInDir(mainPath, a...) }))
		if err != nil {
			return moves, err
		}
		if err := os.MkdirAll(issuesSubFull, 0o755); err != nil {
			return moves, fmt.Errorf("mkdir %s: %v", issuesSubFull, err)
		}
		fmt.Fprintf(stderr, "  Moving %s to %s/\n", base, issuesSubRec)
		if err := os.Rename(ref.Path, dest); err != nil {
			return moves, fmt.Errorf("mv %s → %s: %v", ref.Path, dest, err)
		}
		mirrors.refresh(dest)
		// Record paths relative to mainPath: GitInDir(mainPath, "add", …)
		// resolves them from the main worktree root, so an absolute path here
		// would silently miss the staged move.
		moves = append(moves, preparedArchiveMove{
			IssuePath:   filepath.Join(issuesDir, base),
			HistoryPath: archiveDestination(historyDir, vocab.ArchiveIssues, base),
		})
		if err := finishArchiveMemberMoves(mainPath, assetMoves); err != nil {
			return moves, err
		}
		moves = append(moves, assetMoves...)
		// Sweep the issue's durable plan + review sidecars to history too (#143).
		// Rename under mainPath; record mainPath-relative paths for the git add.
		// An untracked sidecar (#154) stages only its history dest — probe via
		// `git ls-files` in the main worktree.
		planMoves, perr := archivePlanArtifacts(base, plansFull, historyFull, plansDir, historyDir,
			gitSrcUntracked(func(a ...string) ([]byte, error) { return mergeRunner.GitInDir(mainPath, a...) }))
		if perr != nil {
			return moves, perr
		}
		moves = append(moves, planMoves...)
		if err := finishSideArchive(mainPath, sideDocs, sideMoves); err != nil {
			return moves, err
		}
		moves = append(moves, sideMoves...)
	}
	// #304 D4: the archived issues ended; so do their reviewed-head pins.
	unpinArchived(stderr, mainPath, moves)
	return moves, nil
}
