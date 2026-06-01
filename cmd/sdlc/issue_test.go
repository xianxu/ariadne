package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
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

// TestRunIssueNew_SlugTargetDeps: --slug, --target, and --deps flow
// through to the written file.
func TestRunIssueNew_SlugTargetDeps(t *testing.T) {
	issues, history := newTestDirs(t)
	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: issues, HistoryDir: history, Slug: "custom-slug", Target: "my-target", Deps: []string{"repo#1", "repo#2"}}
	if err := runIssueNew(&stdout, &stderr, f, []string{"Ignored Title"}); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}
	want := filepath.Join(issues, "000001-custom-slug.md")
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected file at %s: %v", want, err)
	}
	body := string(data)
	if !strings.Contains(body, "target: my-target") {
		t.Errorf("target not written:\n%s", body)
	}
	if !strings.Contains(body, "deps: [repo#1, repo#2]") {
		t.Errorf("deps not written:\n%s", body)
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

// ── issue list / show ────────────────────────────────────────────────────────

func writeIssueFile(t *testing.T, dir, id, status, title string) {
	t.Helper()
	body := fmt.Sprintf("---\nid: %s\nstatus: %s\n---\n\n# %s\n\n## Problem\n\nprose body here\n", id, status, title)
	if err := os.WriteFile(filepath.Join(dir, id+"-x.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestRunIssueList_SortsAndFilters: list reuses listIssues (sorted by ID)
// and --status filters.
func TestRunIssueList_SortsAndFilters(t *testing.T) {
	issues, _ := newTestDirs(t)
	writeIssueFile(t, issues, "000003", "working", "Third")
	writeIssueFile(t, issues, "000001", "open", "First")
	writeIssueFile(t, issues, "000002", "open", "Second")

	var stdout, stderr bytes.Buffer
	if err := runIssueList(&stdout, &stderr, &issueListFlags{IssuesDir: issues}); err != nil {
		t.Fatalf("runIssueList err: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d:\n%s", len(lines), stdout.String())
	}
	if !strings.HasPrefix(lines[0], "000001") || !strings.HasPrefix(lines[2], "000003") {
		t.Errorf("not sorted by ID:\n%s", stdout.String())
	}

	var so, se bytes.Buffer
	if err := runIssueList(&so, &se, &issueListFlags{IssuesDir: issues, Status: "working"}); err != nil {
		t.Fatalf("runIssueList filter err: %v", err)
	}
	if !strings.Contains(so.String(), "000003") || strings.Contains(so.String(), "000001") {
		t.Errorf("--status working filter wrong:\n%s", so.String())
	}
}

// TestRunIssueShow_HeadersNotBodies: show prints frontmatter + section
// headers but not the section prose; accepts both "5" and "000005".
func TestRunIssueShow_HeadersNotBodies(t *testing.T) {
	issues, _ := newTestDirs(t)
	writeIssueFile(t, issues, "000005", "open", "My Title")

	for _, arg := range []string{"5", "000005"} {
		var stdout, stderr bytes.Buffer
		if err := runIssueShow(&stdout, &stderr, &issueShowFlags{IssuesDir: issues}, arg); err != nil {
			t.Fatalf("runIssueShow(%q) err: %v", arg, err)
		}
		out := stdout.String()
		for _, want := range []string{"000005-x.md", "id: 000005", "# My Title", "## Problem"} {
			if !strings.Contains(out, want) {
				t.Errorf("show(%q) missing %q:\n%s", arg, want, out)
			}
		}
		if strings.Contains(out, "prose body here") {
			t.Errorf("show(%q) leaked section body:\n%s", arg, out)
		}
	}
}

// ── back-compat aliases (#56 M2) ─────────────────────────────────────────────

// TestSetStatusAlias_BothPathsMutate: the flat `sdlc set-status` and the
// grouped `sdlc issue set-status` both resolve to the same handler and
// mutate the issue identically — the back-compat promise, exercised
// through the real command tree (buildRoot).
func TestSetStatusAlias_BothPathsMutate(t *testing.T) {
	issues, _ := newTestDirs(t)
	writeOpen := func() {
		if err := os.WriteFile(filepath.Join(issues, "000001-x.md"), []byte("---\nid: 000001\nstatus: open\n---\n\n# X\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	statusOf := func() string {
		data, _ := os.ReadFile(filepath.Join(issues, "000001-x.md"))
		fm, _, _ := issue.Parse(string(data))
		s, _ := issue.GetField(fm, "status")
		return s
	}
	run := func(args ...string) {
		root := buildRoot()
		root.SetArgs(args)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		if err := root.Execute(); err != nil {
			t.Fatalf("execute %v: %v", args, err)
		}
	}

	writeOpen()
	run("issue", "set-status", "blocked", "--issue", "1", "--issues-dir", issues)
	if got := statusOf(); got != "blocked" {
		t.Errorf("grouped `issue set-status` left status %q, want blocked", got)
	}

	writeOpen()
	run("set-status", "blocked", "--issue", "1", "--issues-dir", issues)
	if got := statusOf(); got != "blocked" {
		t.Errorf("flat `set-status` alias left status %q, want blocked", got)
	}
}

// TestCommandTree_AliasShape: fetch + set-status flat aliases are hidden +
// deprecated, and the grouped commands resolve.
func TestCommandTree_AliasShape(t *testing.T) {
	root := buildRoot()
	find := func(args ...string) *cobra.Command {
		c, _, err := root.Find(args)
		if err != nil {
			t.Fatalf("Find %v: %v", args, err)
		}
		return c
	}
	for _, grouped := range [][]string{{"issue", "new"}, {"issue", "set-status"}, {"issue", "list"}, {"issue", "show"}} {
		if c := find(grouped...); c.Name() != grouped[len(grouped)-1] {
			t.Errorf("%v resolved to %q", grouped, c.Name())
		}
	}
	for _, name := range []string{"set-status", "fetch"} {
		c := find(name)
		if !c.Hidden || c.Deprecated == "" {
			t.Errorf("flat %q should be hidden + deprecated: hidden=%v deprecated=%q", name, c.Hidden, c.Deprecated)
		}
	}
}
