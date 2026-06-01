package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// stubGH stubs ghClient for tests so we don't shell out to `gh`.
// Implements the full ghCaller interface; methods other than the one
// under test are no-ops returning zero values.
type stubGH struct {
	title, body string
	err         error
}

func (s stubGH) TitleAndBody(repo, issueNum string) (string, string, error) {
	return s.title, s.body, s.err
}

func (s stubGH) IssueClose(repo, issueNum, comment string) error           { return nil }
func (s stubGH) PRCreate(repo, base, head, body string) (string, error)    { return "", nil }
func (s stubGH) PRListForBranch(repo, headRef string) (string, error)      { return "", nil }
func (s stubGH) PRMerge(repo, branch string) error                         { return nil }

// TestRunFetch_DryRun exercises the dry-run path end-to-end with stubbed
// gh and a temp workspace. Skips the real gh invocation; verifies the
// rendered output flows through stdout and no file is written.
func TestRunFetch_DryRun(t *testing.T) {
	// Swap in stub gh, restore at test end.
	prev := ghClient
	ghClient = stubGH{title: "My GH Title", body: "Body line one.\nLine two.", err: nil}
	defer func() { ghClient = prev }()

	// Provide a minimal git remote so detectRepo succeeds. We override
	// detectRepo by running inside a git temp dir that has origin set.
	tmp := t.TempDir()
	gitInit(t, tmp, "git@github.com:xianxu/ariadne.git")
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	issuesDir := filepath.Join(tmp, "workshop", "issues")
	historyDir := filepath.Join(tmp, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	f := &fetchFlags{
		GitHubIssue: 42,
		IssuesDir:   issuesDir,
		HistoryDir:  historyDir,
		DryRun:      true,
	}
	if err := runFetch(&stdout, &stderr, f); err != nil {
		t.Fatalf("runFetch dry-run returned err: %v", err)
	}
	// No file created.
	entries, _ := os.ReadDir(issuesDir)
	if len(entries) != 0 {
		t.Errorf("dry-run wrote files: %v", entries)
	}
	if !strings.Contains(stdout.String(), "Would create:") {
		t.Errorf("dry-run stdout missing 'Would create:' — got:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "# My GH Title") {
		t.Errorf("dry-run stdout missing title — got:\n%s", stdout.String())
	}
}

func TestRunFetch_CreatesFileWithNextID(t *testing.T) {
	prev := ghClient
	ghClient = stubGH{title: "Fix the thing", body: "context"}
	defer func() { ghClient = prev }()

	tmp := t.TempDir()
	gitInit(t, tmp, "https://github.com/xianxu/ariadne")
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	issuesDir := filepath.Join(tmp, "workshop", "issues")
	historyDir := filepath.Join(tmp, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Existing 7 → next should be 8.
	if err := os.WriteFile(filepath.Join(issuesDir, "000007-prev.md"), []byte("---\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	f := &fetchFlags{
		GitHubIssue: 99,
		IssuesDir:   issuesDir,
		HistoryDir:  historyDir,
	}
	if err := runFetch(&stdout, &stderr, f); err != nil {
		t.Fatalf("runFetch returned err: %v", err)
	}
	want := filepath.Join(issuesDir, "000008-fix-the-thing.md")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected file at %s, got: %v", want, err)
	}
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"id: 000008",
		"github_issue: 99",
		"# Fix the thing",
	} {
		if !strings.Contains(string(got), want) {
			t.Errorf("file missing %q", want)
		}
	}
}

// TestFetchAlias_ThroughTree exercises the folded `fetch` alias end-to-end
// via the real command tree (buildRoot). `fetch --github-issue N` is the
// thing that changed in M2 (it now delegates to runIssueNew), so prove the
// GH body lands under ## Problem in the canonical template — not just that
// the command is hidden+deprecated.
func TestFetchAlias_ThroughTree(t *testing.T) {
	prev := ghClient
	ghClient = stubGH{title: "Folded Fetch", body: "GH body text."}
	defer func() { ghClient = prev }()

	tmp := t.TempDir()
	gitInit(t, tmp, "git@github.com:xianxu/ariadne.git")
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	issuesDir := filepath.Join(tmp, "workshop", "issues")
	historyDir := filepath.Join(tmp, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	root := buildRoot()
	root.SetArgs([]string{"fetch", "--github-issue", "7", "--issues-dir", issuesDir, "--history-dir", historyDir})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute fetch alias: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(issuesDir, "000001-folded-fetch.md"))
	if err != nil {
		t.Fatalf("expected created file: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "github_issue: 7") {
		t.Errorf("fold lost github_issue:\n%s", body)
	}
	probIdx := strings.Index(body, "## Problem")
	specIdx := strings.Index(body, "## Spec")
	ghIdx := strings.Index(body, "GH body text.")
	if probIdx < 0 || ghIdx < probIdx || ghIdx > specIdx {
		t.Errorf("GH body should sit under ## Problem in the canonical template:\n%s", body)
	}
}

func TestDetectRepo_OriginShapes(t *testing.T) {
	cases := []struct {
		url, want string
	}{
		{"git@github.com:xianxu/ariadne.git\n", "xianxu/ariadne"},
		{"https://github.com/xianxu/ariadne.git\n", "xianxu/ariadne"},
		{"https://github.com/xianxu/ariadne\n", "xianxu/ariadne"},
	}
	for _, c := range cases {
		t.Run(c.url, func(t *testing.T) {
			m := originRE.FindStringSubmatch(c.url)
			if m == nil {
				t.Fatalf("originRE did not match %q", c.url)
			}
			if m[1] != c.want {
				t.Errorf("got %q want %q", m[1], c.want)
			}
		})
	}
}

// gitInit creates a minimal git repo at dir with origin set to remoteURL.
// Returns silently — test fails on git errors via t.Fatal.
func gitInit(t *testing.T, dir, remoteURL string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", remoteURL},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
}
