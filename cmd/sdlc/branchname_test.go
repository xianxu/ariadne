package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureRunner records git invocations + filesystem mutations so we can
// drive runStart / resolveBranchName / listUntrackedIssues without
// shelling out to git or touching the disk in surprising places.
type captureRunner struct {
	// untrackedOutput is what `git ls-files --others --exclude-standard`
	// should return (newline-separated paths).
	untrackedOutput string
	// currentBranch is what `git rev-parse --abbrev-ref HEAD` returns (#156).
	currentBranch string
	// branchExists drives `git show-ref --verify --quiet refs/heads/<name>`:
	// nil error when true (branch present), error when false (#156).
	branchExists bool
	// worktreePorcelain is what `git worktree list --porcelain` returns (#156).
	worktreePorcelain string
	// gitCalls records every Git(...) invocation in order.
	gitCalls [][]string
	// gitInDirCalls records every GitInDir(...) invocation.
	gitInDirCalls []gitInDirCall
	// mkdirs records every MkdirAll(path).
	mkdirs []string
	// writes records every WriteFile(path, data).
	writes []writeOp
}

type gitInDirCall struct {
	Dir  string
	Args []string
}

type writeOp struct {
	Path string
	Data string
}

func (c *captureRunner) Git(args ...string) ([]byte, error) {
	c.gitCalls = append(c.gitCalls, append([]string{}, args...))
	switch {
	case len(args) >= 1 && args[0] == "ls-files":
		return []byte(c.untrackedOutput), nil
	case len(args) >= 2 && args[0] == "rev-parse" && args[1] == "--abbrev-ref":
		return []byte(c.currentBranch), nil // #156 current-branch probe
	case len(args) >= 1 && args[0] == "show-ref":
		if c.branchExists { // #156 branch-existence probe (exit code)
			return nil, nil
		}
		return nil, fmt.Errorf("refs/heads not found")
	case len(args) >= 2 && args[0] == "worktree" && args[1] == "list":
		return []byte(c.worktreePorcelain), nil // #156 worktree probe
	}
	return nil, nil
}

func (c *captureRunner) GitInDir(dir string, args ...string) ([]byte, error) {
	c.gitInDirCalls = append(c.gitInDirCalls, gitInDirCall{Dir: dir, Args: append([]string{}, args...)})
	return nil, nil
}

func (c *captureRunner) MkdirAll(path string) error {
	c.mkdirs = append(c.mkdirs, path)
	return nil
}

func (c *captureRunner) WriteFile(path string, data []byte) error {
	c.writes = append(c.writes, writeOp{Path: path, Data: string(data)})
	return nil
}

func TestResolveStartName_ExplicitName(t *testing.T) {
	r := &captureRunner{}
	name, untracked, err := resolveBranchName(&nameFlags{Name: "feature-x"}, r)
	if err != nil {
		t.Fatal(err)
	}
	if name != "feature-x" {
		t.Errorf("name = %q, want feature-x", name)
	}
	if untracked != "" {
		t.Errorf("untracked = %q, want empty for --name mode", untracked)
	}
	// Should NOT have called git when --name is given.
	if len(r.gitCalls) != 0 {
		t.Errorf("git was called for --name mode: %v", r.gitCalls)
	}
}

func TestResolveStartName_BothFlagsRejected(t *testing.T) {
	r := &captureRunner{}
	_, _, err := resolveBranchName(&nameFlags{Name: "x", Issue: 42}, r)
	if err == nil {
		t.Fatal("expected error for --name + --issue together")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error %q should mention 'mutually exclusive'", err.Error())
	}
}

func TestResolveStartName_IssueFlag_ResolvesFromFile(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issuesDir, "000031-sdlc-checkpoint-binary.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &captureRunner{untrackedOutput: ""} // already tracked
	name, untracked, err := resolveBranchName(&nameFlags{Issue: 31, IssuesDir: issuesDir}, r)
	if err != nil {
		t.Fatal(err)
	}
	if name != "000031-sdlc-checkpoint-binary" {
		t.Errorf("name = %q, want 000031-sdlc-checkpoint-binary", name)
	}
	if untracked != "" {
		t.Errorf("untracked = %q, want empty (file is tracked)", untracked)
	}
}

func TestResolveStartName_IssueFlag_FlagsUntracked(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(issuesDir, "000042-foo.md")
	if err := os.WriteFile(target, []byte("---\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &captureRunner{untrackedOutput: target + "\n"}
	name, untracked, err := resolveBranchName(&nameFlags{Issue: 42, IssuesDir: issuesDir}, r)
	if err != nil {
		t.Fatal(err)
	}
	if name != "000042-foo" {
		t.Errorf("name = %q", name)
	}
	if untracked != target {
		t.Errorf("untracked = %q, want %q", untracked, target)
	}
}

func TestResolveStartName_IssueFlag_NoFile(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "issues")
	_ = os.MkdirAll(issuesDir, 0o755)
	r := &captureRunner{}
	_, _, err := resolveBranchName(&nameFlags{Issue: 99, IssuesDir: issuesDir}, r)
	if err == nil {
		t.Fatal("expected error for missing issue file")
	}
}

func TestResolveStartName_AutoDetect_Single(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "issues")
	_ = os.MkdirAll(issuesDir, 0o755)
	target := filepath.Join(issuesDir, "000050-only.md")
	r := &captureRunner{untrackedOutput: target + "\n"}
	name, untracked, err := resolveBranchName(&nameFlags{IssuesDir: issuesDir}, r)
	if err != nil {
		t.Fatal(err)
	}
	if name != "000050-only" {
		t.Errorf("name = %q", name)
	}
	if untracked != target {
		t.Errorf("untracked = %q", untracked)
	}
}

func TestResolveStartName_AutoDetect_None(t *testing.T) {
	r := &captureRunner{untrackedOutput: ""}
	_, _, err := resolveBranchName(&nameFlags{IssuesDir: "issues"}, r)
	if err == nil {
		t.Fatal("expected error for zero untracked")
	}
	if !strings.Contains(err.Error(), "no untracked issue file") {
		t.Errorf("error msg unexpected: %q", err.Error())
	}
}

func TestResolveStartName_AutoDetect_Multiple(t *testing.T) {
	r := &captureRunner{
		untrackedOutput: "issues/000050-a.md\nissues/000051-b.md\n",
	}
	_, _, err := resolveBranchName(&nameFlags{IssuesDir: "issues"}, r)
	if err == nil {
		t.Fatal("expected error for multiple untracked")
	}
	if !strings.Contains(err.Error(), "multiple") {
		t.Errorf("error msg unexpected: %q", err.Error())
	}
}

func TestResolveStartName_AutoDetect_FiltersJunk(t *testing.T) {
	// Untracked output contains a junk file + a real one. Only the
	// NNNNNN-*.md should be considered.
	r := &captureRunner{
		untrackedOutput: "issues/.DS_Store\nissues/not-an-issue.md\nissues/000077-real.md\n",
	}
	name, untracked, err := resolveBranchName(&nameFlags{IssuesDir: "issues"}, r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "000077-real" {
		t.Errorf("name = %q, want 000077-real", name)
	}
	if untracked != "issues/000077-real.md" {
		t.Errorf("untracked = %q", untracked)
	}
}

// TestListUntrackedIssues_FilterShape verifies that the filename filter
// covers the shapes that have shown up in real ls-files output: bare
// filenames, leading dirs, .DS_Store, etc.
func TestListUntrackedIssues_FilterShape(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"issues/000077-real.md\n", []string{"issues/000077-real.md"}},
		{"workshop/issues/000001-foo.md\nworkshop/issues/junk.md\n",
			[]string{"workshop/issues/000001-foo.md"}},
		// 5 digits → must not match.
		{"issues/00001-too-short.md\n", nil},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%q", c.in), func(t *testing.T) {
			r := &captureRunner{untrackedOutput: c.in}
			got, err := listUntrackedIssues("issues", r)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("got[%d]=%q want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}
