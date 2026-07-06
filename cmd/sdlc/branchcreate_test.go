package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// ── pure deciders (#156) ─────────────────────────────────────────────────────

func TestDecideInPlaceBranch(t *testing.T) {
	cases := []struct {
		name    string
		current string
		target  string
		exists  bool
		want    inPlaceAction
	}{
		{"already on target", "feat", "feat", true, inPlaceOnTarget},
		{"exists not checked out", "main", "feat", true, inPlaceSwitch},
		{"brand new branch", "main", "feat", false, inPlaceCreate},
		{"detached head, absent", "", "feat", false, inPlaceCreate},
		{"detached head, exists", "", "feat", true, inPlaceSwitch},
	}
	for _, tc := range cases {
		if got := decideInPlaceBranch(tc.current, tc.target, tc.exists); got != tc.want {
			t.Errorf("%s: decideInPlaceBranch(%q,%q,%v) = %d, want %d", tc.name, tc.current, tc.target, tc.exists, got, tc.want)
		}
	}
}

func TestDecideWorktreeBranch(t *testing.T) {
	cases := []struct {
		name         string
		branchExists bool
		wtFound      bool
		want         worktreeAction
	}{
		{"already in a worktree → reuse (wins over exists)", true, true, worktreeReuse},
		{"branch exists, no worktree → add existing", true, false, worktreeAddExisting},
		{"brand new → add -b", false, false, worktreeAddNew},
		{"wtFound wins even if branchExists false (shouldn't happen, defensive)", false, true, worktreeReuse},
	}
	for _, tc := range cases {
		if got := decideWorktreeBranch(tc.branchExists, tc.wtFound); got != tc.want {
			t.Errorf("%s: decideWorktreeBranch(%v,%v) = %d, want %d", tc.name, tc.branchExists, tc.wtFound, got, tc.want)
		}
	}
}

// ── porcelain parser + branch filter (#156, single-source grammar) ───────────

func TestParseWorktrees(t *testing.T) {
	porcelain := "worktree /repo\nHEAD aaa\nbranch refs/heads/main\n\n" +
		"worktree /repo/wt/feat\nHEAD bbb\nbranch refs/heads/000156-feat\n\n" +
		"worktree /repo/wt/detached\nHEAD ccc\ndetached\n"
	got := parseWorktrees(porcelain)
	if len(got) != 3 {
		t.Fatalf("parseWorktrees returned %d entries, want 3: %+v", len(got), got)
	}
	want := []WorktreeState{
		{Path: "/repo", Branch: "main"},
		{Path: "/repo/wt/feat", Branch: "000156-feat"},
		{Path: "/repo/wt/detached", Branch: "(detached)"},
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("entry[%d] = %+v, want %+v", i, got[i], w)
		}
	}
	if parseWorktrees("") != nil {
		t.Errorf("empty porcelain must yield nil")
	}
}

func TestWorktreeForBranch(t *testing.T) {
	porcelain := "worktree /repo\nHEAD aaa\nbranch refs/heads/main\n\n" +
		"worktree /repo/wt/feat\nHEAD bbb\nbranch refs/heads/000156-feat\n"
	if p, ok := worktreeForBranch(porcelain, "000156-feat"); !ok || p != "/repo/wt/feat" {
		t.Errorf("worktreeForBranch(feat) = %q,%v, want /repo/wt/feat,true", p, ok)
	}
	if p, ok := worktreeForBranch(porcelain, "nope"); ok || p != "" {
		t.Errorf("worktreeForBranch(nope) = %q,%v, want \"\",false", p, ok)
	}
}

// ── IO-shell wiring: the exact git command per state (the original bug is here) ──

// gitCalled reports whether the runner issued a call whose args exactly equal want.
func gitCalled(calls [][]string, want ...string) bool {
	for _, c := range calls {
		if len(c) != len(want) {
			continue
		}
		match := true
		for i := range c {
			if c[i] != want[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// anyGitCallStartsWith reports whether some call begins with the given prefix.
func anyGitCallStartsWith(calls [][]string, prefix ...string) bool {
	for _, c := range calls {
		if len(c) < len(prefix) {
			continue
		}
		match := true
		for i := range prefix {
			if c[i] != prefix[i] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestCreateInPlaceBranch_AlreadyOnTarget(t *testing.T) {
	r := &captureRunner{currentBranch: "000156-x", branchExists: true}
	var out, errb bytes.Buffer
	if _, err := createInPlaceBranch(&out, &errb, "000156-x", r); err != nil {
		t.Fatal(err)
	}
	if anyGitCallStartsWith(r.gitCalls, "checkout") {
		t.Errorf("already-on-target must not run any checkout: %v", r.gitCalls)
	}
	if !strings.Contains(errb.String(), "Already on branch 000156-x") {
		t.Errorf("expected already-on-branch info, got: %q", errb.String())
	}
}

func TestCreateInPlaceBranch_ExistsNotCheckedOut(t *testing.T) {
	r := &captureRunner{currentBranch: "main", branchExists: true}
	var out, errb bytes.Buffer
	if _, err := createInPlaceBranch(&out, &errb, "000156-x", r); err != nil {
		t.Fatal(err)
	}
	if !gitCalled(r.gitCalls, "checkout", "000156-x") {
		t.Errorf("branch-exists must switch via `checkout <name>` (no -b): %v", r.gitCalls)
	}
	if gitCalled(r.gitCalls, "checkout", "-b", "000156-x") {
		t.Errorf("branch-exists must NOT use -b: %v", r.gitCalls)
	}
}

func TestCreateInPlaceBranch_CreatesWhenAbsent(t *testing.T) {
	r := &captureRunner{currentBranch: "main", branchExists: false}
	var out, errb bytes.Buffer
	if _, err := createInPlaceBranch(&out, &errb, "000156-x", r); err != nil {
		t.Fatal(err)
	}
	if !gitCalled(r.gitCalls, "checkout", "-b", "000156-x") {
		t.Errorf("absent branch must be created via `checkout -b <name>`: %v", r.gitCalls)
	}
}

func TestCreateWorktreeBranch_ReusesExisting(t *testing.T) {
	// The branch is already checked out in a worktree → reuse it, never add.
	r := &captureRunner{
		branchExists:      true,
		worktreePorcelain: "worktree /existing/wt/000156-x\nHEAD abc\nbranch refs/heads/000156-x\n",
	}
	var out, errb bytes.Buffer
	wtPath, err := createWorktreeBranch(&out, &errb, "000156-x", r)
	if err != nil {
		t.Fatal(err)
	}
	if anyGitCallStartsWith(r.gitCalls, "worktree", "add") {
		t.Errorf("reuse must NOT run `worktree add`: %v", r.gitCalls)
	}
	if wtPath != "/existing/wt/000156-x" {
		t.Errorf("reuse must return the existing worktree path, got %q", wtPath)
	}
	// .goto is rewritten to the reused worktree.
	foundGoto := false
	for _, w := range r.writes {
		if strings.HasSuffix(w.Path, ".goto") && w.Data == "/existing/wt/000156-x" {
			foundGoto = true
		}
	}
	if !foundGoto {
		t.Errorf(".goto must point at the reused worktree, writes: %+v", r.writes)
	}
}

func TestCreateWorktreeBranch_AddsExistingBranchWithoutDashB(t *testing.T) {
	// Branch exists but isn't in any worktree → `worktree add <path> <name>` (no -b).
	r := &captureRunner{branchExists: true, worktreePorcelain: "worktree /repo\nHEAD abc\nbranch refs/heads/main\n"}
	var out, errb bytes.Buffer
	if _, err := createWorktreeBranch(&out, &errb, "000156-x", r); err != nil {
		t.Fatal(err)
	}
	if anyGitCallStartsWith(r.gitCalls, "worktree", "add", "-b") {
		t.Errorf("existing branch must NOT use -b: %v", r.gitCalls)
	}
	if !anyGitCallStartsWith(r.gitCalls, "worktree", "add") {
		t.Errorf("existing branch must still run `worktree add`: %v", r.gitCalls)
	}
	// The name is the last arg of the add call (checking out the existing branch).
	for _, c := range r.gitCalls {
		if len(c) >= 2 && c[0] == "worktree" && c[1] == "add" {
			if c[len(c)-1] != "000156-x" {
				t.Errorf("worktree add existing must end with the branch name: %v", c)
			}
		}
	}
}

func TestCreateWorktreeBranch_AddsNewWithDashB(t *testing.T) {
	// Brand-new branch → today's `worktree add -b <name> <path> HEAD`.
	r := &captureRunner{branchExists: false, worktreePorcelain: "worktree /repo\nHEAD abc\nbranch refs/heads/main\n"}
	var out, errb bytes.Buffer
	if _, err := createWorktreeBranch(&out, &errb, "000156-x", r); err != nil {
		t.Fatal(err)
	}
	if !anyGitCallStartsWith(r.gitCalls, "worktree", "add", "-b", "000156-x") {
		t.Errorf("new branch must use `worktree add -b <name> …`: %v", r.gitCalls)
	}
}

// TestCreateInPlaceBranch_RealRepo_IdempotentRerun reproduces the exact #156 bug
// end-to-end against a REAL git repo (the captureRunner tests use synthetic probe
// output; this proves the real `rev-parse`/`show-ref` output parses correctly).
// The second call — re-running while already on the branch — is the milestone
// re-run that used to die "fatal: a branch named '…' already exists".
func TestCreateInPlaceBranch_RealRepo_IdempotentRerun(t *testing.T) {
	hermeticRepo(t) // real repo on main, chdir'd in
	run := execGitRunner{}
	if err := os.WriteFile("f.txt", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := run.Git("add", "-A"); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	if out, err := run.Git("commit", "-m", "init"); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}

	name := "000156-real"
	var out, errb bytes.Buffer

	// 1. First run creates the branch (checkout -b).
	if _, err := createInPlaceBranch(&out, &errb, name, run); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if cur := currentBranch(run); cur != name {
		t.Fatalf("after create, current branch = %q, want %q", cur, name)
	}

	// 2. Re-run while ON the branch — THE #156 BUG. Old code: exit-128
	//    "branch already exists". New code: idempotent "already on branch".
	errb.Reset()
	if _, err := createInPlaceBranch(&out, &errb, name, run); err != nil {
		t.Fatalf("re-run on the existing branch must be idempotent (#156), got: %v", err)
	}
	if !strings.Contains(errb.String(), "Already on branch") {
		t.Errorf("re-run should report already-on-branch, got: %q", errb.String())
	}

	// 3. Switch away, then re-run — must switch back (checkout, no -b), not error.
	if out2, err := run.Git("checkout", "main"); err != nil {
		t.Fatalf("git checkout main: %v\n%s", err, out2)
	}
	errb.Reset()
	if _, err := createInPlaceBranch(&out, &errb, name, run); err != nil {
		t.Fatalf("re-run from main onto an existing branch must switch, not error: %v", err)
	}
	if cur := currentBranch(run); cur != name {
		t.Errorf("after switch, current branch = %q, want %q", cur, name)
	}
	if !strings.Contains(errb.String(), "Switched to existing branch") {
		t.Errorf("expected switch message, got: %q", errb.String())
	}
}
