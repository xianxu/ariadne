package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestDirs makes an issues/ + history/ pair under a temp dir and
// returns both paths.
func newTestDirs(t *testing.T) (issues, history string) {
	t.Helper()
	dir := t.TempDir()
	issues = filepath.Join(dir, "issues")
	history = filepath.Join(dir, "history")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatal(err)
	}
	return issues, history
}

// TestRunIssueNew_BlankCreatesNextID: a blank `issue new` allocates the
// next ID, derives the slug, and writes a parseable canonical skeleton
// with Problem/Spec present-but-empty.
func TestRunIssueNew_BlankCreatesNextID(t *testing.T) {
	issues, history := newTestDirs(t)
	if err := os.WriteFile(filepath.Join(issues, "000007-prev.md"), []byte("---\nid: 000007\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: issues, HistoryDir: history}
	if err := runIssueNew(&stdout, &stderr, f, []string{"Lift the Issue Subsystem!"}); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}

	want := filepath.Join(issues, "000008-lift-the-issue-subsystem.md")
	if got := strings.TrimSpace(stdout.String()); got != want {
		t.Errorf("stdout path = %q, want %q", got, want)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	body := string(data)
	for _, s := range []string{"id: 000008", "status: open", "# Lift the Issue Subsystem!", "## Problem", "## Spec", "## Done when", "## Plan"} {
		if !strings.Contains(body, s) {
			t.Errorf("file missing %q\n%s", s, body)
		}
	}
	// Problem/Spec present but empty (back-to-back headers with no prose).
	if !strings.Contains(body, "## Problem\n\n## Spec") {
		t.Errorf("blank issue should have empty Problem before Spec:\n%s", body)
	}
}

// TestRunIssueNew_SlugOverrideAndTarget: --slug and --target are honored.
func TestRunIssueNew_SlugOverrideAndTarget(t *testing.T) {
	issues, history := newTestDirs(t)
	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: issues, HistoryDir: history, Slug: "custom-slug", Target: "my-target"}
	if err := runIssueNew(&stdout, &stderr, f, []string{"Ignored Title"}); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}
	want := filepath.Join(issues, "000001-custom-slug.md")
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	if !strings.Contains(string(data), "target: my-target") {
		t.Errorf("target not written:\n%s", data)
	}
}

// TestRunIssueNew_DryRunWritesNothing: --dry-run prints the body but
// creates no file.
func TestRunIssueNew_DryRunWritesNothing(t *testing.T) {
	issues, history := newTestDirs(t)
	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: issues, HistoryDir: history, DryRun: true}
	if err := runIssueNew(&stdout, &stderr, f, []string{"Some title"}); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}
	if !strings.Contains(stdout.String(), "Would create:") {
		t.Errorf("dry-run missing summary: %q", stdout.String())
	}
	entries, _ := os.ReadDir(issues)
	if len(entries) != 0 {
		t.Errorf("dry-run wrote files: %v", entries)
	}
}

// TestRunIssueNew_FromGitHubFillsProblem: --from-github takes the title
// from GitHub and seeds the body under ## Problem.
func TestRunIssueNew_FromGitHubFillsProblem(t *testing.T) {
	prev := ghClient
	ghClient = stubGH{title: "Imported From GH", body: "The GH body.\n\nDetail."}
	defer func() { ghClient = prev }()

	issues, history := newTestDirs(t)
	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: issues, HistoryDir: history, FromGitHub: 42}
	if err := runIssueNew(&stdout, &stderr, f, nil); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}
	want := filepath.Join(issues, "000001-imported-from-gh.md")
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	body := string(data)
	if !strings.Contains(body, "github_issue: 42") {
		t.Errorf("github_issue not set:\n%s", body)
	}
	probIdx := strings.Index(body, "## Problem")
	specIdx := strings.Index(body, "## Spec")
	ghIdx := strings.Index(body, "The GH body.")
	if ghIdx < probIdx || ghIdx > specIdx {
		t.Errorf("GH body should sit under ## Problem:\n%s", body)
	}
}
