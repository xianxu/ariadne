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

// TestRunIssueNew_AutoSyncsToMainCleanTree (#82 M1): on main, `issue new`
// scaffolds the file AND broadcasts it to origin/main via the shared sync —
// leaving the working tree clean for that file and NOT sweeping an unrelated
// untracked issue file (rides #80's filtered add). Run against a real repo +
// bare origin so the commit/push happen for real. (Serial — chdirs the process,
// like the merge e2e harness; no t.Parallel().)
func TestRunIssueNew_AutoSyncsToMainCleanTree(t *testing.T) {
	dir := t.TempDir()
	origin := filepath.Join(t.TempDir(), "origin.git")
	git(t, "", "init", "--bare", "-b", "main", origin)
	git(t, "", "init", "-b", "main", dir)
	git(t, dir, "config", "user.email", "e2e@example.com")
	git(t, dir, "config", "user.name", "e2e")
	git(t, dir, "config", "commit.gpgsign", "false")

	issuesDir := filepath.Join(dir, "workshop", "issues")
	historyDir := filepath.Join(dir, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issuesDir, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("seed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "seed main")
	git(t, dir, "remote", "add", "origin", origin)
	git(t, dir, "push", "-u", "origin", "main")

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	// An unrelated, never-claimed untracked issue file — must stay untracked.
	unrelated := filepath.Join(issuesDir, "000777-unrelated.md")
	if err := os.WriteFile(unrelated, []byte("---\nid: 000777\nstatus: open\n---\n\n# wip\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if err := runIssueNew(&stdout, &stderr, f, []string{"Auto Synced Issue"}); err != nil {
		t.Fatalf("runIssueNew err: %v", err)
	}

	// stdout is JUST the created path — the sync's "synced" marker is routed to
	// stderr so it can't pollute the path contract callers parse.
	created := strings.TrimSpace(stdout.String())
	if !strings.HasSuffix(created, "-auto-synced-issue.md") || strings.Contains(created, "\n") {
		t.Fatalf("stdout should be only the created path, got %q", created)
	}
	base := filepath.Base(created)

	// The new file is committed on HEAD (issue-sync commit) and pushed to origin.
	head := git(t, dir, "diff-tree", "--no-commit-id", "--name-only", "-r", "HEAD")
	if !strings.Contains(head, "workshop/issues/"+base) {
		t.Errorf("new issue not in HEAD commit; files:\n%s", head)
	}
	if originHead := git(t, dir, "log", "origin/main", "--oneline", "-1"); !strings.Contains(originHead, "issue-sync") {
		t.Errorf("issue-sync commit not on origin/main: %q", originHead)
	}

	// Working tree clean for the new file (committed), and the UNRELATED file
	// stays untracked — never swept by the filtered add.
	status := git(t, dir, "status", "--porcelain", "--untracked-files=all")
	if strings.Contains(status, base) {
		t.Errorf("created issue should be committed, not left dirty; status:\n%s", status)
	}
	if !strings.Contains(status, "000777-unrelated.md") {
		t.Errorf("unrelated untracked file should remain untracked; status:\n%s", status)
	}
}

// TestRunIssueNew_AutoSyncBestEffort (#82 M1): when the sync can't complete
// (here: a repo with NO origin remote, so the push fails), `issue new` must
// still create + report the file and return nil — the sync is best-effort, not
// a gate. A warning is surfaced. (Regression: an earlier cut had the sync die()
// internally, which os.Exit'd the whole command on a failed push.)
func TestRunIssueNew_AutoSyncBestEffort(t *testing.T) {
	dir := t.TempDir()
	git(t, "", "init", "-b", "main", dir)
	git(t, dir, "config", "user.email", "e2e@example.com")
	git(t, dir, "config", "user.name", "e2e")
	git(t, dir, "config", "commit.gpgsign", "false")
	issuesDir := filepath.Join(dir, "workshop", "issues")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issuesDir, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "seed")
	// NOTE: no `git remote add origin` — the push step has nowhere to go.

	prevWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prevWD) })

	var stdout, stderr bytes.Buffer
	f := &issueNewFlags{IssuesDir: "workshop/issues", HistoryDir: "workshop/history"}
	if err := runIssueNew(&stdout, &stderr, f, []string{"No Origin Issue"}); err != nil {
		t.Fatalf("issue new must not fail when sync can't push, got err: %v", err)
	}
	created := strings.TrimSpace(stdout.String())
	if !strings.HasSuffix(created, "-no-origin-issue.md") {
		t.Fatalf("stdout should be the created path, got %q", created)
	}
	if _, err := os.Stat(created); err != nil {
		t.Errorf("file should exist despite sync failure: %v", err)
	}
	if !strings.Contains(stderr.String(), "auto-sync to main did not complete") {
		t.Errorf("expected a best-effort sync warning on stderr; got:\n%s", stderr.String())
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
