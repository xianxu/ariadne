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
	// #149: set-status is a mutating verb, so buildRoot().Execute() acquires the
	// repo transaction lock. Chdir into a temp git repo so that lock resolves to
	// the temp .git, not the developer's real .git/sdlc.lock (which hangs the
	// suite when a live `sdlc` holds it). --issues-dir stays an absolute temp path.
	hermeticRepo(t)
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
	// #122 M4: use a model-legal flip (open→working) — set-status now gates on the
	// lifecycle graph and open→blocked is illegal (claim first). The test's point is
	// alias-vs-grouped parity, not the specific transition.
	run("issue", "set-status", "working", "--issue", "1", "--issues-dir", issues)
	if got := statusOf(); got != "working" {
		t.Errorf("grouped `issue set-status` left status %q, want working", got)
	}

	writeOpen()
	run("set-status", "working", "--issue", "1", "--issues-dir", issues)
	if got := statusOf(); got != "working" {
		t.Errorf("flat `set-status` alias left status %q, want working", got)
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
