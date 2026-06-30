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

// #143: archivePlanArtifacts moves exactly the id-prefixed plan + sidecars,
// leaves unrelated ids, and records paths that stage cleanly via archiveAddArgs.
func TestArchivePlanArtifacts(t *testing.T) {
	tmp := t.TempDir()
	plans := filepath.Join(tmp, "plans")
	history := filepath.Join(tmp, "history")
	mkArtifact(t, filepath.Join(plans, "000143-x-plan.md"), "the plan")
	mkArtifact(t, filepath.Join(plans, "000143-x-close-review.md"), "the review")
	mkArtifact(t, filepath.Join(plans, "000999-y-plan.md"), "unrelated")

	moves, err := archivePlanArtifacts("000143-x.md", plans, history, "plans", "history")
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

func TestArchivePlanArtifacts_NoPlanIsNoOp(t *testing.T) {
	tmp := t.TempDir()
	plans := filepath.Join(tmp, "plans")
	if err := os.MkdirAll(plans, 0o755); err != nil {
		t.Fatal(err)
	}
	moves, err := archivePlanArtifacts("000143-x.md", plans, filepath.Join(tmp, "history"), "plans", "history")
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
