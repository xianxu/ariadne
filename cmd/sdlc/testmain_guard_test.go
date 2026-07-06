package main

import "testing"

func snap(head, branch string, porcelain []string, lock bool) repoSnapshot {
	set := map[string]bool{}
	for _, p := range porcelain {
		set[p] = true
	}
	return repoSnapshot{head: head, branch: branch, porcelain: set, lockFile: lock, resolved: true}
}

// TestSnapshotDiff pins the pure guard decision (#149/#165): reports only NEW
// mutations, ignores pre-existing untracked, and catches HEAD/branch/lock changes.
func TestSnapshotDiff(t *testing.T) {
	base := snap("abc123def", "main", []string{"?? workshop/pensive/x.md"}, false)

	if d := snapshotDiff(base, base); len(d) != 0 {
		t.Errorf("identical snapshots should diff empty, got %v", d)
	}

	// Pre-existing untracked (present in BOTH) must NOT trip it.
	after := snap("abc123def", "main", []string{"?? workshop/pensive/x.md"}, false)
	if d := snapshotDiff(base, after); len(d) != 0 {
		t.Errorf("pre-existing untracked must not trip the guard, got %v", d)
	}

	// A NEW working-tree change is flagged.
	if d := snapshotDiff(base, snap("abc123def", "main", []string{"?? workshop/pensive/x.md", "?? f"}, false)); len(d) != 1 || d[0] != "new working-tree change: ?? f" {
		t.Errorf("new change: got %v", d)
	}

	// HEAD moved.
	if d := snapshotDiff(base, snap("999fffaaa", "main", []string{"?? workshop/pensive/x.md"}, false)); len(d) != 1 || d[0] == "" || d[0][:4] != "HEAD" {
		t.Errorf("HEAD move: got %v", d)
	}

	// Branch switched.
	if d := snapshotDiff(base, snap("abc123def", "feature", []string{"?? workshop/pensive/x.md"}, false)); len(d) != 1 || d[0][:6] != "branch" {
		t.Errorf("branch switch: got %v", d)
	}

	// Leaked lock (false → true).
	if d := snapshotDiff(base, snap("abc123def", "main", []string{"?? workshop/pensive/x.md"}, true)); len(d) != 1 || d[0][:6] != "leaked" {
		t.Errorf("leaked lock: got %v", d)
	}

	// Unresolvable repo → guard skips (empty).
	if d := snapshotDiff(repoSnapshot{}, repoSnapshot{}); len(d) != 0 {
		t.Errorf("unresolved repo should skip, got %v", d)
	}
}

// TestGuardVerdict proves the guard FIRES on a passing run that mutated the repo,
// and does NOT override an already-failing run (but still surfaces the mutation).
func TestGuardVerdict(t *testing.T) {
	clean := snap("abc", "main", nil, false)
	dirty := snap("abc", "main", []string{"?? f"}, false)

	// Passing run + mutation → FAIL (exit 1) with the mutation surfaced.
	if exit, muts := guardVerdict(clean, dirty, 0); exit != 1 || len(muts) != 1 {
		t.Errorf("passing+mutation: exit=%d muts=%v, want exit=1 with 1 mutation", exit, muts)
	}
	// Passing run + clean → pass (0).
	if exit, muts := guardVerdict(clean, clean, 0); exit != 0 || len(muts) != 0 {
		t.Errorf("passing+clean: exit=%d muts=%v, want exit=0 none", exit, muts)
	}
	// Failing run + mutation → keep the failure (don't mask), still surface it.
	if exit, muts := guardVerdict(clean, dirty, 2); exit != 2 || len(muts) != 1 {
		t.Errorf("failing+mutation: exit=%d muts=%v, want exit=2 with mutation surfaced", exit, muts)
	}
}
