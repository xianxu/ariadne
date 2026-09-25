package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	// Both spellings reach the same card writer (#252). The fixture is a temp repo,
	// so the mutating-verb lock never touches the developer's .git (#149).
	cardPath, card, _, _ := seededIssue(t, "000001", "x")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	run := func(args ...string) {
		root := buildRoot()
		root.SetArgs(args)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Fatalf("execute %v: %v", args, err)
		}
	}
	run("issue", "set-status", "working", "--issue", "1")
	if got := r.card(cardPath); !strings.Contains(got, "status: working") {
		t.Errorf("grouped `issue set-status` left:\n%s", got)
	}
	run("set-status", "blocked", "--issue", "1")
	if got := r.card(cardPath); !strings.Contains(got, "status: blocked") {
		t.Errorf("flat `set-status` alias left:\n%s", got)
	}
}

// TestCardSetters_UpdateOnlyTheirField drives each card setter through the
// command tree and checks the card changed in exactly that field.
func TestCardSetters_UpdateOnlyTheirField(t *testing.T) {
	cardPath, card, _, _ := seededIssue(t, "000001", "x")
	r := newTrackerRepo(t, map[string]string{cardPath: card}, nil)
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"issue", "set-estimate", "--issue", "1", "--hours", "2.5"}, "estimate_hours: 2.5"},
		{[]string{"issue", "set-github", "--issue", "1", "--number", "42"}, "github_issue: 42"},
		{[]string{"issue", "set-title", "--issue", "1", "Better title"}, "# Better title"},
	} {
		before := r.card(cardPath)
		root := buildRoot()
		root.SetArgs(c.args)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Fatalf("%v: %v", c.args, err)
		}
		after := r.card(cardPath)
		if !strings.Contains(after, c.want) {
			t.Fatalf("%v: card lacks %q:\n%s", c.args, c.want, after)
		}
		if changed := diffLines(before, after); changed != 1 {
			t.Errorf("%v changed %d lines, want 1", c.args, changed)
		}
	}
	for _, bad := range [][]string{{"issue", "set-estimate", "--issue", "1", "--hours", "0"}, {"issue", "set-github", "--issue", "1"}, {"issue", "set-title", "--issue", "1", " "}} {
		root := buildRoot()
		root.SetArgs(bad)
		root.SetOut(&bytes.Buffer{})
		root.SetErr(&bytes.Buffer{})
		if err := root.ExecuteContext(context.Background()); err == nil {
			t.Errorf("%v accepted", bad)
		}
	}
}

func diffLines(a, b string) int {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	if len(al) != len(bl) {
		return -1
	}
	n := 0
	for i := range al {
		if al[i] != bl[i] {
			n++
		}
	}
	return n
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
