package main

import (
	"strings"
	"testing"
)

// lockRunnerStub extends captureRunner so we can stub responses per
// invocation. The lock subcommand makes several distinct git calls
// (diff HEAD, diff --cached, ls-files, worktree list, branch). We
// dispatch by the first 1-2 args.
type lockRunnerStub struct {
	captureRunner
	// Per-query stub outputs. Key is the first arg (e.g. "diff",
	// "ls-files", "worktree"). Value is the stdout. Errors are nil.
	responses map[string][]byte
	// gitInDirResponses keyed similarly for GitInDir calls.
	gitInDirResponses map[string][]byte
}

func (s *lockRunnerStub) Git(args ...string) ([]byte, error) {
	s.gitCalls = append(s.gitCalls, append([]string{}, args...))
	if len(args) == 0 {
		return nil, nil
	}
	// Compose a key from arg[0..2] (most specific) and fall back.
	for n := min(len(args), 3); n > 0; n-- {
		key := strings.Join(args[:n], " ")
		if v, ok := s.responses[key]; ok {
			return v, nil
		}
	}
	return nil, nil
}

func (s *lockRunnerStub) GitInDir(dir string, args ...string) ([]byte, error) {
	s.gitInDirCalls = append(s.gitInDirCalls, gitInDirCall{Dir: dir, Args: append([]string{}, args...)})
	for n := min(len(args), 3); n > 0; n-- {
		key := strings.Join(args[:n], " ")
		if v, ok := s.gitInDirResponses[key]; ok {
			return v, nil
		}
	}
	return nil, nil
}

func TestChangedIssueFiles_DedupesAndSorts(t *testing.T) {
	r := &lockRunnerStub{
		responses: map[string][]byte{
			"diff --name-only HEAD":   []byte("workshop/issues/000002-b.md\nworkshop/issues/000001-a.md\n"),
			"diff --cached --name-only": []byte("workshop/issues/000001-a.md\n"),
			"ls-files --others":         []byte("workshop/issues/000003-c.md\n"),
		},
	}
	got, err := changedIssueFiles(&lockFlags{IssuesDir: "workshop/issues"}, r)
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
	r := &lockRunnerStub{
		responses: map[string][]byte{
			"diff --name-only HEAD": []byte("workshop/issues/000001-a.md\nworkshop/issues/000031-target.md\n"),
		},
	}
	got, err := changedIssueFiles(&lockFlags{IssuesDir: "workshop/issues", Issue: 31}, r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "workshop/issues/000031-target.md" {
		t.Fatalf("got %v want [workshop/issues/000031-target.md]", got)
	}
}

func TestFindMainWorktree_Parses(t *testing.T) {
	out := `worktree /repo/main
HEAD abc123
branch refs/heads/main

worktree /repo/feature
HEAD def456
branch refs/heads/feature-x
`
	r := &lockRunnerStub{
		responses: map[string][]byte{"worktree list": []byte(out)},
	}
	got, err := findMainWorktree(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/repo/main" {
		t.Errorf("got %q want /repo/main", got)
	}
}

func TestFindMainWorktree_NoMain(t *testing.T) {
	out := `worktree /repo/feature
HEAD def456
branch refs/heads/feature-x
`
	r := &lockRunnerStub{
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

func TestMainHasUncommittedIssueChanges_Union(t *testing.T) {
	r := &lockRunnerStub{
		gitInDirResponses: map[string][]byte{
			"diff --name-only":         []byte("workshop/issues/000001-a.md\n"),
			"diff --cached --name-only": []byte("workshop/issues/000002-b.md\n"),
		},
	}
	got, err := mainHasUncommittedIssueChanges("/main", "workshop/issues", r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %v, want 2 entries", got)
	}
	if got[0] != "workshop/issues/000001-a.md" || got[1] != "workshop/issues/000002-b.md" {
		t.Errorf("entries unexpected: %v", got)
	}
}

func TestMainHasUncommittedIssueChanges_None(t *testing.T) {
	r := &lockRunnerStub{
		gitInDirResponses: map[string][]byte{}, // empty stdout for both queries
	}
	got, err := mainHasUncommittedIssueChanges("/main", "workshop/issues", r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}
