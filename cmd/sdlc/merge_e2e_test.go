package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ── e2e harness for runMerge (#63) ───────────────────────────────────────────
//
// runMerge resists in-process testing for two reasons: die() halts the process
// (handled by expectDie, die_test.go), and a dozen git calls run against the
// real working directory (gitx.Capture, detectRepo, RepoTopLevel, and the
// execGitRunner switch/pull/archive/branch-delete). Rather than stub all of
// those, tempRepo stands up a throwaway repo with a local bare origin and runs
// runMerge against it for real — so the cleanup (switch/pull/archive/push/
// branch-delete) is exercised end-to-end, not mocked. Only the network (gh) is
// stubbed (ghClient) and the origin-URL parse (detectRepo, which would demand a
// github.com URL we can't push a local bare origin to).
//
// Topology: a single git checkout sitting on `feature` is an *in-place*
// checkout (isInPlaceCheckout sees a plain ".git"), so these tests cover the
// in-place merge path (git switch main → pull → archive → branch -D). The
// linked-worktree topology is not exercised by this harness.

// git runs `git <args>` in dir and fails the test on error, returning stdout.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s (in %s): %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// tempRepo builds a throwaway repo on branch `feature` with a local bare
// origin, chdirs into it (restored on cleanup), and returns its path. main is
// pushed with an upstream (so the resume path's `git pull` has an origin/main
// to track) and a `status: done` issue is seeded + committed on main (so the
// archive step has something to move). The feature branch adds one innocuous
// commit and is pushed, so merge's step-3/4 push checks pass.
func tempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)

	// init + identity (local config persists for the real runners merge uses).
	git(t, "", "init", "-b", "main", dir)
	git(t, dir, "config", "user.email", "e2e@example.com")
	git(t, dir, "config", "user.name", "e2e")
	git(t, dir, "config", "commit.gpgsign", "false")

	// Seed a done issue + history dir on main so archive has work to do.
	issuesDir := filepath.Join(dir, "workshop", "issues")
	historyDir := filepath.Join(dir, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// .gitkeep so history/ is tracked (git ignores empty dirs).
	if err := os.WriteFile(filepath.Join(historyDir, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	doneIssue := "---\nid: 999\nstatus: done\n---\n\n# seeded done issue\n"
	if err := os.WriteFile(filepath.Join(issuesDir, "000999-done.md"), []byte(doneIssue), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "seed main")
	git(t, dir, "remote", "add", "origin", origin)
	git(t, dir, "push", "-u", "origin", "main")

	// Feature branch with one innocuous commit, pushed (satisfies step 3/4).
	git(t, dir, "switch", "-c", "feature")
	if err := os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "feature work")
	git(t, dir, "push", "-u", "origin", "feature")

	// Chdir so the real runners (gitx.Capture / execGitRunner / RepoTopLevel)
	// operate on this repo. Restored on cleanup.
	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })
	return dir
}

// e2eGH is a recording ghCaller for the e2e tests. It embeds stubGH for the
// no-op methods and overrides the three merge cares about, recording PRMerge
// calls so a test can assert the irreversible merge did (or did NOT) happen.
type e2eGH struct {
	stubGH
	openPR       string // PRListForBranch result
	mergedExists bool   // PRMergedForBranch result
	prMergeCalls int    // how many times PRMerge fired
}

func (g *e2eGH) PRListForBranch(repo, headRef string) (string, error) { return g.openPR, nil }
func (g *e2eGH) PRMergedForBranch(repo, headRef string) (bool, error) { return g.mergedExists, nil }
func (g *e2eGH) PRMerge(repo, branch string) error                    { g.prMergeCalls++; return nil }

// swapMergeDeps swaps the package-level seams (ghClient, detectRepo,
// runPublishGateFn) and restores them on cleanup. gate may be nil to keep the
// default; pass a stub to inject a step-5 hook (#160 the publish gate replaced the
// LLM judges, but the seam still models "a step-5 hook that can dirty the tree").
//
// These are process-global vars with no synchronization. Safe because these
// tests run serially (no t.Parallel(), and tempRepo does a process-global
// os.Chdir). Do NOT add t.Parallel() to merge e2e tests without first giving
// each its own isolated state — the swaps (and the chdir) would race.
func swapMergeDeps(t *testing.T, gh ghCaller, gate func(baseRef, issuesDir string, stderr io.Writer) error) {
	t.Helper()
	prevGH, prevDetect, prevGate, prevVal := ghClient, detectRepo, runPublishGateFn, validateChangedIssuesFn
	ghClient = gh
	detectRepo = func() (string, error) { return "test/repo", nil }
	if gate != nil {
		runPublishGateFn = gate
	}
	// Neutralize the #124 instance-conformance gate — these e2e tests exercise the
	// merge FLOW, not the gate (which has its own unit tests in validategate_test.go)
	// and would otherwise shell the `vocabulary` binary, absent in the test env.
	validateChangedIssuesFn = func(_, _, _ string, _, _ io.Writer) error { return nil }
	t.Cleanup(func() {
		ghClient = prevGH
		detectRepo = prevDetect
		runPublishGateFn = prevGate
		validateChangedIssuesFn = prevVal
	})
}

// ── #62 M1 regression: a judge dirties a TRACKED file → refuse PRE-merge ──────
//
// Step 2 checked the tree clean; a pre-merge judge then modifies a tracked file.
// Step 9b must catch it and refuse BEFORE the irreversible gh pr merge — never
// merge server-side and then strand on the local `git switch` (a dirty tracked
// file makes switch/pull refuse). Per #78 the judge dirties a TRACKED file
// (README.md), since untracked dirt no longer blocks (see the sibling test).
func TestRunMerge_DirtyAfterJudge_RefusesPreMerge(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "42"} // an open PR exists; merge would proceed if not refused
	// A "judge" that dirties a tracked file (README.md is committed by tempRepo)
	// and returns nil (success). NoJudge stays false so this fires (step 5 is
	// gated on it). A passing judge that left a tracked file dirty is the #62 hazard.
	dirtyingGate := func(_, _ string, _ io.Writer) error {
		return os.WriteFile(filepath.Join(dir, "README.md"), []byte("dirtied by step-5 hook\n"), 0o644)
	}
	swapMergeDeps(t, gh, dirtyingGate)

	f := &mergeFlags{Yes: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })

	if !died {
		t.Fatal("expected merge to refuse (die) on a tracked file dirtied by the judge")
	}
	if !strings.Contains(msg, "before the irreversible merge") {
		t.Errorf("refusal message = %q, want the 9b dirty-tree guidance", msg)
	}
	if gh.prMergeCalls != 0 {
		t.Errorf("PRMerge called %d times — must be 0 (refusal is PRE-merge)", gh.prMergeCalls)
	}
}

// ── #78: a judge leaves only an UNTRACKED file → merge PROCEEDS ───────────────
//
// The mirror of the test above. An untracked file survives `git switch main`,
// so it can't strand the post-merge cleanup the way a dirty tracked file does.
// Step 9b must NOT refuse on it: the irreversible merge fires and cleanup
// completes (this is the live #58-shipping scenario that motivated #78).
func TestRunMerge_UntrackedAfterJudge_Proceeds(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "42"}
	untrackingGate := func(_, _ string, _ io.Writer) error {
		return os.WriteFile(filepath.Join(dir, "gate-scratch.txt"), []byte("x\n"), 0o644)
	}
	swapMergeDeps(t, gh, untrackingGate)

	f := &mergeFlags{Yes: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if died {
		t.Fatalf("merge should proceed past an untracked-only dirty tree, but died: %s", msg)
	}
	if gh.prMergeCalls != 1 {
		t.Errorf("PRMerge called %d times — want 1 (untracked dirt must not block)", gh.prMergeCalls)
	}
	if got := git(t, dir, "branch", "--show-current"); got != "main" {
		t.Errorf("ended on branch %q, want main", got)
	}
	if got := git(t, dir, "branch", "--list", "feature"); got != "" {
		t.Errorf("feature branch not deleted: %q", got)
	}
	// The untracked file is still there (we never touch it) — proof it neither
	// blocked the merge nor was clobbered by the branch switch.
	if _, err := os.Stat(filepath.Join(dir, "gate-scratch.txt")); err != nil {
		t.Errorf("untracked file should survive the merge/switch: %v", err)
	}
}

// ── #82 M2: a dirty TRACKED tracker file does not block the merge ─────────────
//
// The complement of TestRunMerge_DirtyAfterJudge_RefusesPreMerge (dirty tracked
// CODE → refuse). tempRepo commits 000999-done.md on both main and feature; we
// dirty it (tracked-modified, status kept `done`) before merging. assessDirty
// buckets it as Tracker — never Blocking — so the merge proceeds: tracker state
// is append-only shared state synced to main out-of-band, not code contention.
func TestRunMerge_DirtyTrackerFile_Proceeds(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "42"}
	swapMergeDeps(t, gh, nil)

	// Dirty the committed issue file in the working tree (still status: done so
	// the archive still moves it). It's identical on main, so `git switch main`
	// carries the modification without conflict.
	issuePath := filepath.Join(dir, "workshop", "issues", "000999-done.md")
	if err := os.WriteFile(issuePath, []byte("---\nid: 999\nstatus: done\n---\n\n# seeded done issue\n\nedited in the working tree\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if died {
		t.Fatalf("merge should proceed past a dirty tracked tracker file, but died: %s", msg)
	}
	if gh.prMergeCalls != 1 {
		t.Errorf("PRMerge called %d times — want 1 (dirty tracker file must not block)", gh.prMergeCalls)
	}
	if got := git(t, dir, "branch", "--show-current"); got != "main" {
		t.Errorf("ended on branch %q, want main", got)
	}
}

// ── #80: archive stages only the moved issue, not unrelated untracked WIP ─────
//
// tempRepo seeds a done issue (000999) the archive moves. Before merge we drop
// an UNRELATED untracked issue file (000888) into workshop/issues/ — local WIP
// for an unclaimed issue. The archive commit must contain exactly the moved
// issue (src deletion + history addition) and must NOT sweep 000888, which stays
// untracked after the merge. Regression for the broad `git add <dir>/` (#80):
// before the fix, #78's untracked-tolerant guard let the merge proceed and the
// directory-wide add then committed the unrelated WIP onto main.
func TestRunMerge_ArchiveDoesNotSweepUntrackedIssue(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "42"}
	swapMergeDeps(t, gh, nil)

	// Unrelated, never-claimed WIP issue file sitting untracked in issues/.
	unrelated := filepath.Join(dir, "workshop", "issues", "000888-unrelated.md")
	if err := os.WriteFile(unrelated, []byte("---\nid: 888\nstatus: open\n---\n\n# unrelated WIP\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if died {
		t.Fatalf("merge should proceed past an untracked issue file, but died: %s", msg)
	}

	// The archive commit is HEAD on main. It must record exactly the 000999
	// move (deletion + history addition) — a wrong (absolute) path would
	// silently miss this — and must NOT include the unrelated 000888.
	files := git(t, dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
	if !strings.Contains(files, "workshop/issues/000999-done.md") {
		t.Errorf("archive commit should record the moved issue's deletion; files:\n%s", files)
	}
	if !strings.Contains(files, "workshop/history/000999-done.md") {
		t.Errorf("archive commit should record the history addition; files:\n%s", files)
	}
	if strings.Contains(files, "000888-unrelated.md") {
		t.Errorf("archive commit swept the unrelated untracked file (#80 regression); files:\n%s", files)
	}

	// And the unrelated file is still there, still untracked after the merge.
	if _, err := os.Stat(unrelated); err != nil {
		t.Errorf("unrelated WIP file should survive untouched: %v", err)
	}
	status := git(t, dir, "status", "--porcelain", "--untracked-files=all")
	if !strings.Contains(status, "000888-unrelated.md") {
		t.Errorf("unrelated file should remain untracked after merge; status:\n%s", status)
	}
}

// ── #62 M3 regression: resume an already-merged PR → finish cleanup ──────────
//
// No open PR but a merged one exists (a prior run was interrupted after the
// server-side merge). Re-running must NOT re-merge; it resumes the local
// cleanup idempotently: switch to main, pull, archive done issues, delete the
// branch — exercised for real against the temp repo + bare origin.
func TestRunMerge_ResumeMergedPR_FinishesCleanup(t *testing.T) {
	dir := tempRepo(t)
	gh := &e2eGH{openPR: "", mergedExists: true}
	swapMergeDeps(t, gh, nil)

	// NoJudge: tree stays clean (no judge to dirty it); the resume path is
	// what's under test, not the judges.
	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) })
	if died {
		t.Fatalf("resume path should not refuse, but died: %s", msg)
	}

	if gh.prMergeCalls != 0 {
		t.Errorf("PRMerge called %d times on resume — must be 0 (PR already merged)", gh.prMergeCalls)
	}
	if got := git(t, dir, "branch", "--show-current"); got != "main" {
		t.Errorf("ended on branch %q, want main", got)
	}
	if got := git(t, dir, "branch", "--list", "feature"); got != "" {
		t.Errorf("feature branch not deleted: %q", got)
	}
	// Archive ran for real: the done issue moved into history/, off issues/.
	if _, err := os.Stat(filepath.Join(dir, "workshop", "history", "000999-done.md")); err != nil {
		t.Errorf("archived issue not in history/: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "workshop", "issues", "000999-done.md")); !os.IsNotExist(err) {
		t.Errorf("done issue should have moved out of issues/ (err=%v)", err)
	}
}

// ── #160: merge flips a codecomplete issue → done and archives it ────────────
//
// The publish flip (step 10.5) runs on main after the merge, before the archive
// (which keys on IsTerminal). A codecomplete issue on main must end up archived to
// history with status: done. The publish gate is stubbed to pass (its invariant is
// unit-tested in publishgate_test.go); this pins the merge-side flip+archive wiring.
func TestRunMerge_CodecompleteFlippedToDoneAndArchived(t *testing.T) {
	dir := tempRepo(t)
	// Seed a codecomplete issue on main (the flip's target).
	git(t, dir, "switch", "main")
	cc := "---\nid: 160\nstatus: codecomplete\nactual_hours: 1\n---\n\n# cc issue\n"
	if err := os.WriteFile(filepath.Join(dir, "workshop/issues/000160-cc.md"), []byte(cc), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "seed codecomplete issue")
	git(t, dir, "push", "origin", "main")
	git(t, dir, "switch", "feature")

	gh := &e2eGH{openPR: "42"}
	// --no-judge SKIPS the publish gate, but the flip (step 10.5) is unconditional —
	// pin that codecomplete → done still happens on the emergency-bypass path.
	swapMergeDeps(t, gh, nil)

	f := &mergeFlags{Yes: true, NoJudge: true, IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if msg, died := expectDie(t, func() { runMerge(io.Discard, io.Discard, f) }); died {
		t.Fatalf("merge should succeed, died: %s", msg)
	}

	data, err := os.ReadFile(filepath.Join(dir, "workshop/history/000160-cc.md"))
	if err != nil {
		t.Fatalf("codecomplete issue should be archived to history: %v", err)
	}
	if !strings.Contains(string(data), "status: done") {
		t.Errorf("archived issue should be flipped codecomplete → done:\n%s", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "workshop/issues/000160-cc.md")); !os.IsNotExist(err) {
		t.Error("codecomplete issue should have moved out of workshop/issues/")
	}
}
