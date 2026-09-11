package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// claimRunnerStub extends captureRunner so we can stub responses per
// invocation. The claim subcommand makes several distinct git calls
// (diff HEAD, diff --cached, ls-files, worktree list, branch). We
// dispatch by the first 1-2 args.
type claimRunnerStub struct {
	captureRunner
	// Per-query stub outputs. Key is the first arg (e.g. "diff",
	// "ls-files", "worktree"). Value is the stdout. Errors are nil.
	responses map[string][]byte
	// gitInDirResponses keyed similarly for GitInDir calls.
	gitInDirResponses map[string][]byte
}

func (s *claimRunnerStub) Git(args ...string) ([]byte, error) {
	s.gitCalls = append(s.gitCalls, append([]string{}, args...))
	if len(args) == 0 {
		return nil, nil
	}
	// Compose a key from the longest prefix that has a response and fall back to
	// shorter ones. The window is wider than the 3 args it started at because two
	// distinct calls can share a 3-arg prefix and differ only in their pathspec
	// (`diff --cached --name-only -- <dir>/` vs `-- <file>`), and a stub that
	// can't tell them apart makes the code under test look like it did nothing.
	for n := min(len(args), 8); n > 0; n-- {
		key := strings.Join(args[:n], " ")
		if v, ok := s.responses[key]; ok {
			return v, nil
		}
	}
	return nil, nil
}

func (s *claimRunnerStub) GitInDir(dir string, args ...string) ([]byte, error) {
	s.gitInDirCalls = append(s.gitInDirCalls, gitInDirCall{Dir: dir, Args: append([]string{}, args...)})
	for n := min(len(args), 8); n > 0; n-- {
		key := strings.Join(args[:n], " ")
		if v, ok := s.gitInDirResponses[key]; ok {
			return v, nil
		}
	}
	// Fall back to the plain Git responses: real GitInDir(dir, args) IS Git(args)
	// run in dir, so a stub that answers one and not the other models a git that
	// does not exist (#207 BR-16 moved the sync onto GitInDir).
	return s.Git(args...)
}

func TestChangedIssueFiles_DedupesAndSorts(t *testing.T) {
	r := &claimRunnerStub{
		responses: map[string][]byte{
			// NUL-delimited, because the queries pass -z: a stub that answers in
			// git's newline form models a command the code no longer runs (#207 BR-9).
			"diff --name-only -z HEAD":     []byte("workshop/issues/000002-b.md\x00workshop/issues/000001-a.md\x00"),
			"diff --cached --name-only -z": []byte("workshop/issues/000001-a.md\x00"),
			"ls-files -z --others":         []byte("workshop/issues/000003-c.md\x00"),
		},
	}
	got, err := changedIssueFiles(&claimFlags{IssuesDir: "workshop/issues"}, r, syncPaths{Root: "."})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"workshop/issues/000001-a.md",
		"workshop/issues/000002-b.md",
		"workshop/issues/000003-c.md",
	}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] got %q want %q", i, got[i], want[i])
		}
	}
}

func TestChangedIssueFiles_FilterByIssue(t *testing.T) {
	r := &claimRunnerStub{
		responses: map[string][]byte{
			"diff --name-only -z HEAD": []byte("workshop/issues/000001-a.md\x00workshop/issues/000031-target.md\x00"),
		},
	}
	got, err := changedIssueFiles(&claimFlags{IssuesDir: "workshop/issues", Issue: 31}, r, syncPaths{Root: "."})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "workshop/issues/000031-target.md" {
		t.Fatalf("got %v want [workshop/issues/000031-target.md]", got)
	}
}

func TestFindMainWorktree_Parses(t *testing.T) {
	out := worktreePorcelainZ(
		[]string{"worktree /repo/main", "HEAD abc123", "branch refs/heads/main"},
		[]string{"worktree /repo/feature", "HEAD def456", "branch refs/heads/feature-x"},
	)
	r := &claimRunnerStub{
		responses: map[string][]byte{"worktree list": []byte(out)},
	}
	got, err := findMainWorktree(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/repo/main" {
		t.Errorf("got %q want /repo/main", got)
	}
	if !gitCalled(r.gitCalls, "worktree", "list", "--porcelain", "-z") {
		t.Errorf("findMainWorktree must request NUL-delimited porcelain: %v", r.gitCalls)
	}
}

func TestFindMainWorktree_NoMain(t *testing.T) {
	out := worktreePorcelainZ([]string{"worktree /repo/feature", "HEAD def456", "branch refs/heads/feature-x"})
	r := &claimRunnerStub{
		responses: map[string][]byte{"worktree list": []byte(out)},
	}
	_, err := findMainWorktree(r)
	if err == nil {
		t.Fatal("expected error when no main worktree")
	}
	if !strings.Contains(err.Error(), "main") {
		t.Errorf("error message should mention main: %q", err.Error())
	}
}

// ── startOnClaim: the folded-in open→working start flip ──────────────────────

// writeIssue is a small fixture helper for the start-flip tests.
func writeIssue(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestStartOnClaim_FlipsOpenToWorking: an open issue with an estimate is
// flipped to working as part of the claim.
func TestStartOnClaim_FlipsOpenToWorking(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000031-foo.md")
	writeIssue(t, dir, "000031-foo.md", "---\nid: 000031\nstatus: open\nestimate_hours: 2\n---\n# Foo\n")

	var stdout, stderr bytes.Buffer
	f := &claimFlags{Issue: 31, IssuesDir: dir}
	if err := startOnClaim(&stdout, &stderr, f); err != nil {
		t.Fatalf("startOnClaim err: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "status: working") {
		t.Errorf("issue not flipped to working:\n%s", got)
	}
}

// TestStartOnClaim_OpenWithoutEstimateStillFlips: #113 made claim a cheap
// lock — an open issue with NO estimate_hours must still flip to working
// (the estimate gate moved to `sdlc change-code`).
func TestStartOnClaim_OpenWithoutEstimateStillFlips(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000031-foo.md")
	writeIssue(t, dir, "000031-foo.md", "---\nid: 000031\nstatus: open\n---\n# Foo\n")

	var stdout, stderr bytes.Buffer
	f := &claimFlags{Issue: 31, IssuesDir: dir}
	if err := startOnClaim(&stdout, &stderr, f); err != nil {
		t.Fatalf("claim with no estimate should flip cleanly now, got: %v", err)
	}
	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), "status: working") {
		t.Errorf("issue not flipped to working without an estimate:\n%s", got)
	}
}

// TestStartOnClaim_LeavesNonOpenUntouched: claim must not clobber a status
// the operator set on purpose (here, blocked).
func TestStartOnClaim_LeavesNonOpenUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000031-foo.md")
	original := "---\nid: 000031\nstatus: blocked\nestimate_hours: 2\n---\n# Foo\n"
	writeIssue(t, dir, "000031-foo.md", original)

	var stdout, stderr bytes.Buffer
	f := &claimFlags{Issue: 31, IssuesDir: dir}
	if err := startOnClaim(&stdout, &stderr, f); err != nil {
		t.Fatalf("startOnClaim err: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != original {
		t.Errorf("non-open issue was mutated:\n--- got ---\n%s\n--- want ---\n%s", got, original)
	}
}

// TestStartOnClaim_DryRunDoesNotWrite: --dry-run reports but writes nothing.
func TestStartOnClaim_DryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "000031-foo.md")
	original := "---\nid: 000031\nstatus: open\nestimate_hours: 2\n---\n# Foo\n"
	writeIssue(t, dir, "000031-foo.md", original)

	var stdout, stderr bytes.Buffer
	f := &claimFlags{Issue: 31, IssuesDir: dir, DryRun: true}
	if err := startOnClaim(&stdout, &stderr, f); err != nil {
		t.Fatalf("startOnClaim err: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != original {
		t.Errorf("dry-run mutated file:\n%s", got)
	}
	if !strings.Contains(stderr.String(), "would flip") {
		t.Errorf("dry-run stderr missing notice: %q", stderr.String())
	}
}
