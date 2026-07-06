package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mkArtifact(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// trackedProbe is the archivePlanArtifacts source-trackedness probe faked to
// report every source as tracked (the pre-#154 assumption). Tests exercising the
// untracked branch pass their own probe. Injecting it keeps the git IO out of the
// pure move-builder (ARCH-PURE) — these tests run on plain temp dirs, not repos.
func trackedProbe(string) bool { return false }

// #143: archivePlanArtifacts moves exactly the id-prefixed plan + sidecars,
// leaves unrelated ids, and records paths that stage cleanly via archiveAddArgs.
func TestArchivePlanArtifacts(t *testing.T) {
	tmp := t.TempDir()
	plans := filepath.Join(tmp, "plans")
	history := filepath.Join(tmp, "history")
	mkArtifact(t, filepath.Join(plans, "000143-x-plan.md"), "the plan")
	mkArtifact(t, filepath.Join(plans, "000143-x-close-review.md"), "the review")
	mkArtifact(t, filepath.Join(plans, "000999-y-plan.md"), "unrelated")

	moves, err := archivePlanArtifacts("000143-x.md", plans, history, "plans", "history", trackedProbe)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 2 {
		t.Fatalf("want 2 plan moves, got %d: %#v", len(moves), moves)
	}
	for _, name := range []string{"000143-x-plan.md", "000143-x-close-review.md"} {
		if _, err := os.Stat(filepath.Join(history, name)); err != nil {
			t.Errorf("%s should be in history: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(plans, name)); !os.IsNotExist(err) {
			t.Errorf("%s should be gone from plans (err=%v)", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(plans, "000999-y-plan.md")); err != nil {
		t.Errorf("unrelated 000999 plan must stay: %v", err)
	}

	// Staging contract: every moved plan path (both halves) appears in the
	// archive git-add args — this is what makes them committed, pinned at unit level.
	add := strings.Join(archiveAddArgs(moves), " ")
	for _, want := range []string{
		filepath.Join("plans", "000143-x-plan.md"), filepath.Join("history", "000143-x-plan.md"),
		filepath.Join("plans", "000143-x-close-review.md"), filepath.Join("history", "000143-x-close-review.md"),
	} {
		if !strings.Contains(add, want) {
			t.Errorf("archiveAddArgs missing %q:\n%s", want, add)
		}
	}
}

// #154: an untracked review sidecar (created by `sdlc close`, never committed)
// must archive without dying on `git add <vanished-source>`. The move records
// SourceUntracked=true so archiveAddArgs stages only its history dest, while the
// tracked durable plan alongside it still stages source-deletion + dest. The
// probe is faked (index-based `git ls-files` in production) — a working-tree
// os.Stat check would wrongly flag the tracked-then-renamed plan as untracked.
func TestArchivePlanArtifacts_UntrackedSidecarStagesDestOnly(t *testing.T) {
	tmp := t.TempDir()
	plans := filepath.Join(tmp, "plans")
	history := filepath.Join(tmp, "history")
	mkArtifact(t, filepath.Join(plans, "000154-x-plan.md"), "tracked durable plan")
	mkArtifact(t, filepath.Join(plans, "000154-x-close-review.md"), "untracked sidecar")

	// Fake the probe: the sidecar is untracked, the durable plan is tracked.
	untrackedSet := map[string]bool{filepath.Join("plans", "000154-x-close-review.md"): true}
	probe := func(recPath string) bool { return untrackedSet[recPath] }

	moves, err := archivePlanArtifacts("000154-x.md", plans, history, "plans", "history", probe)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 2 {
		t.Fatalf("want 2 plan moves, got %d: %#v", len(moves), moves)
	}
	// Both files physically moved to history regardless of trackedness.
	for _, name := range []string{"000154-x-plan.md", "000154-x-close-review.md"} {
		if _, err := os.Stat(filepath.Join(history, name)); err != nil {
			t.Errorf("%s should be in history: %v", name, err)
		}
	}
	// The move flags: sidecar untracked, plan tracked.
	byBase := map[string]preparedArchiveMove{}
	for _, m := range moves {
		byBase[filepath.Base(m.IssuePath)] = m
	}
	if !byBase["000154-x-close-review.md"].SourceUntracked {
		t.Errorf("untracked sidecar move must set SourceUntracked=true: %#v", byBase["000154-x-close-review.md"])
	}
	if byBase["000154-x-plan.md"].SourceUntracked {
		t.Errorf("tracked plan move must keep SourceUntracked=false: %#v", byBase["000154-x-plan.md"])
	}

	// The staging contract: the untracked sidecar's vanished plans source is NOT
	// in the add-list (that's the exit-128 pathspec failure this fixes), its
	// history dest IS; the tracked plan stages both halves.
	add := strings.Join(archiveAddArgs(moves), " ")
	sidecarSrc := filepath.Join("plans", "000154-x-close-review.md")
	if strings.Contains(add, sidecarSrc) {
		t.Errorf("untracked sidecar source %q must NOT be staged (pathspec would fail):\n%s", sidecarSrc, add)
	}
	for _, want := range []string{
		filepath.Join("history", "000154-x-close-review.md"), // sidecar dest
		filepath.Join("plans", "000154-x-plan.md"),           // tracked plan src
		filepath.Join("history", "000154-x-plan.md"),         // tracked plan dest
	} {
		if !strings.Contains(add, want) {
			t.Errorf("archiveAddArgs missing %q:\n%s", want, add)
		}
	}
}

// #154 end-to-end: in a REAL git repo, archive a done issue whose durable plan is
// committed (tracked) but whose review sidecar is untracked — exactly the state
// `sdlc close` + a FIX-THEN-SHIP fixup leaves. This drives the real `git ls-files`
// probe (not a fake) through pushRunner, then runs the real `git add`/`commit`
// that used to die "pathspec did not match" (exit 128). It must complete cleanly.
func TestArchiveDoneIssues_UntrackedSidecar_RealRepo(t *testing.T) {
	hermeticRepo(t) // real repo, chdir'd in; pushRunner (execGitRunner) runs here
	issues, history, plans := "workshop/issues", "workshop/history", "workshop/plans"
	mkArtifact(t, filepath.Join(issues, "000154-x.md"), "---\nid: 154\nstatus: done\n---\n\n# x\n")
	mkArtifact(t, filepath.Join(plans, "000154-x-plan.md"), "durable plan")
	// Commit the issue + durable plan → they are tracked at archive time.
	if out, err := pushRunner.Git("add", "--", filepath.Join(issues, "000154-x.md"), filepath.Join(plans, "000154-x-plan.md")); err != nil {
		t.Fatalf("git add seed: %v\n%s", err, out)
	}
	if out, err := pushRunner.Git("commit", "-m", "seed"); err != nil {
		t.Fatalf("git commit seed: %v\n%s", err, out)
	}
	// The review sidecar exists on disk but was never committed → untracked.
	mkArtifact(t, filepath.Join(plans, "000154-x-close-review.md"), "untracked sidecar")

	var stderr bytes.Buffer
	moves, err := archiveDoneIssues(&stderr, "", issues, history, plans)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 3 {
		t.Fatalf("want 3 moves (issue + plan + sidecar), got %d: %#v", len(moves), moves)
	}
	// The untracked sidecar move must be flagged so its vanished source is not staged.
	var sawUntrackedSidecar bool
	for _, m := range moves {
		if filepath.Base(m.IssuePath) == "000154-x-close-review.md" {
			sawUntrackedSidecar = m.SourceUntracked
		}
	}
	if !sawUntrackedSidecar {
		t.Errorf("real untracked sidecar must set SourceUntracked=true: %#v", moves)
	}

	// The real archive git-add + commit — this is what failed with exit 128 before.
	if out, gerr := pushRunner.Git(archiveAddArgs(moves)...); gerr != nil {
		t.Fatalf("archive git add died (the #154 bug): %v\n%s", gerr, out)
	}
	if out, gerr := pushRunner.Git("commit", "-m", "archive completed issues to history"); gerr != nil {
		t.Fatalf("archive commit failed: %v\n%s", gerr, out)
	}
	// Clean worktree + all three files tracked in history, none left in plans/issues.
	if out, _ := pushRunner.Git("status", "--porcelain"); strings.TrimSpace(string(out)) != "" {
		t.Errorf("worktree not clean after archive — half-archived state (#154):\n%s", out)
	}
	for _, name := range []string{"000154-x.md", "000154-x-plan.md", "000154-x-close-review.md"} {
		if out, gerr := pushRunner.Git("ls-files", "--error-unmatch", "--", filepath.Join(history, name)); gerr != nil {
			t.Errorf("%s should be tracked in history after archive: %v\n%s", name, gerr, out)
		}
	}
}

func TestArchivePlanArtifacts_NoPlanIsNoOp(t *testing.T) {
	tmp := t.TempDir()
	plans := filepath.Join(tmp, "plans")
	if err := os.MkdirAll(plans, 0o755); err != nil {
		t.Fatal(err)
	}
	moves, err := archivePlanArtifacts("000143-x.md", plans, filepath.Join(tmp, "history"), "plans", "history", trackedProbe)
	if err != nil {
		t.Fatalf("a no-plan issue must not error: %v", err)
	}
	if len(moves) != 0 {
		t.Errorf("no-plan issue should yield 0 moves, got %d", len(moves))
	}
}

// #143: the push archive loop sweeps a done issue's plan artifacts to history,
// and leaves an open issue's plan untouched.
func TestArchiveDoneIssues_SweepsPlanArtifacts(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	issues, history, plans := "workshop/issues", "workshop/history", "workshop/plans"
	mkArtifact(t, filepath.Join(issues, "000143-x.md"), "---\nid: 143\nstatus: done\n---\n\n# x\n")
	mkArtifact(t, filepath.Join(plans, "000143-x-plan.md"), "plan")
	mkArtifact(t, filepath.Join(plans, "000143-x-close-review.md"), "review")
	mkArtifact(t, filepath.Join(issues, "000144-y.md"), "---\nid: 144\nstatus: working\n---\n\n# y\n")
	mkArtifact(t, filepath.Join(plans, "000144-y-plan.md"), "plan-open")

	var stderr bytes.Buffer
	moves, err := archiveDoneIssues(&stderr, "", issues, history, plans)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 3 {
		t.Fatalf("want 3 moves (issue + plan + sidecar), got %d: %#v", len(moves), moves)
	}
	for _, name := range []string{"000143-x.md", "000143-x-plan.md", "000143-x-close-review.md"} {
		if _, err := os.Stat(filepath.Join(history, name)); err != nil {
			t.Errorf("%s should be archived to history: %v", name, err)
		}
	}
	// The open issue and its plan are untouched.
	if _, err := os.Stat(filepath.Join(issues, "000144-y.md")); err != nil {
		t.Errorf("open issue must remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plans, "000144-y-plan.md")); err != nil {
		t.Errorf("open issue's plan must remain: %v", err)
	}
}

// #143: the merge archive loop (mainPath-relative) sweeps a done issue's plan +
// review sidecar to history, records mainPath-relative move paths, and leaves an
// open issue's plan untouched.
func TestArchiveDoneIssuesInDir_SweepsPlanArtifacts(t *testing.T) {
	tmp := t.TempDir() // acts as the main worktree root
	issues, history, plans := "workshop/issues", "workshop/history", "workshop/plans"
	mkArtifact(t, filepath.Join(tmp, issues, "000143-x.md"), "---\nid: 143\nstatus: done\n---\n\n# x\n")
	mkArtifact(t, filepath.Join(tmp, plans, "000143-x-plan.md"), "plan")
	mkArtifact(t, filepath.Join(tmp, plans, "000143-x-m1-review.md"), "milestone review")
	mkArtifact(t, filepath.Join(tmp, issues, "000144-y.md"), "---\nid: 144\nstatus: working\n---\n\n# y\n")
	mkArtifact(t, filepath.Join(tmp, plans, "000144-y-plan.md"), "open plan")

	var stderr bytes.Buffer
	moves, err := archiveDoneIssuesInDir(&stderr, "owner/repo", tmp, issues, history, plans)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 3 {
		t.Fatalf("want 3 moves (issue + plan + sidecar), got %d: %#v", len(moves), moves)
	}
	for _, m := range moves {
		if filepath.IsAbs(m.IssuePath) || filepath.IsAbs(m.HistoryPath) {
			t.Errorf("merge archive paths must be mainPath-relative (GitInDir resolves them): %#v", m)
		}
	}
	for _, name := range []string{"000143-x.md", "000143-x-plan.md", "000143-x-m1-review.md"} {
		if _, err := os.Stat(filepath.Join(tmp, history, name)); err != nil {
			t.Errorf("%s should be archived to history: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(tmp, plans, "000144-y-plan.md")); err != nil {
		t.Errorf("open issue's plan must remain: %v", err)
	}
}

// #143: an interrupted-archive recovery (push) reconstructs BOTH the issue move
// and the plan move from git status — the plan move with no terminal-frontmatter
// gate (its plans-dir source is the membership proof).
func TestPreparedArchiveMoves_RecoversPlanArtifacts(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	mkArtifact(t, "workshop/history/000143-x.md", "---\nstatus: done\n---\n\n# x\n")
	mkArtifact(t, "workshop/history/000143-x-plan.md", "a durable plan, no frontmatter")

	status := " D workshop/issues/000143-x.md\n" +
		"?? workshop/history/000143-x.md\n" +
		" D workshop/plans/000143-x-plan.md\n" +
		"?? workshop/history/000143-x-plan.md\n"
	moves, other, err := preparedArchiveMoves(status, "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Fatalf("other = %v, want none (plan move must not be flagged unrelated)", other)
	}
	if len(moves) != 2 {
		t.Fatalf("want 2 moves (issue + plan), got %d: %#v", len(moves), moves)
	}
	var sawPlan bool
	for _, m := range moves {
		if strings.Contains(m.IssuePath, "000143-x-plan.md") {
			sawPlan = true
		}
	}
	if !sawPlan {
		t.Errorf("the plan move was not recovered: %#v", moves)
	}
}
