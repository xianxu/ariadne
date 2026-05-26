package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Simple Title", "simple-title"},
		{"with-Existing-Dashes", "with-existing-dashes"},
		{"  leading whitespace", "leading-whitespace"},
		{"Trailing punctuation!", "trailing-punctuation"},
		{"--multiple--dashes--", "multiple-dashes"},
		{"Caps  AND   Spaces", "caps-and-spaces"},
		{"Symbols / & special?", "symbols-special"},
		{"Numbers 42 keep", "numbers-42-keep"},
		{"UPPER", "upper"},
		{"unicode café", "unicode-caf"}, // accent stripped, matches sed [^a-z0-9]
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			if got := slugify(c.in); got != c.want {
				t.Errorf("slugify(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestNextIssueID(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	history := filepath.Join(dir, "history")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatal(err)
	}

	// Highest is 000031 in issues; history has lower numbers.
	for _, name := range []string{"000005-old.md", "000010-older.md"} {
		if err := os.WriteFile(filepath.Join(history, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"000020-a.md", "000031-b.md", "not-an-issue.md"} {
		if err := os.WriteFile(filepath.Join(issues, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := nextIssueID(issues, history)
	if err != nil {
		t.Fatal(err)
	}
	if got != "000032" {
		t.Errorf("nextIssueID = %q, want 000032", got)
	}
}

func TestNextIssueID_HighestInHistory(t *testing.T) {
	dir := t.TempDir()
	issues := filepath.Join(dir, "issues")
	history := filepath.Join(dir, "history")
	if err := os.MkdirAll(issues, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(history, 0o755); err != nil {
		t.Fatal(err)
	}
	// Max is in history (closed issue archived).
	if err := os.WriteFile(filepath.Join(history, "000099-done.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(issues, "000050-active.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := nextIssueID(issues, history)
	if err != nil {
		t.Fatal(err)
	}
	if got != "000100" {
		t.Errorf("nextIssueID = %q, want 000100", got)
	}
}

func TestNextIssueID_MissingDirs(t *testing.T) {
	dir := t.TempDir()
	got, err := nextIssueID(filepath.Join(dir, "nope"), filepath.Join(dir, "nope2"))
	if err != nil {
		t.Fatal(err)
	}
	if got != "000001" {
		t.Errorf("nextIssueID = %q, want 000001 for empty dirs", got)
	}
}

func TestRenderFetchedIssue_Shape(t *testing.T) {
	out := renderFetchedIssue("000032", "555", "Add Foo Feature", "Body paragraph.\n\nMore detail.", "2026-05-25")
	for _, want := range []string{
		"id: 000032",
		"status: open",
		"deps: []",
		"github_issue: 555",
		"created: 2026-05-25",
		"updated: 2026-05-25",
		"# Add Foo Feature",
		"Body paragraph.",
		"## Done when",
		"\n-\n",
		"## Plan",
		"\n- [ ]\n",
		"## Log",
		"### 2026-05-25",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered issue missing %q\n--- output ---\n%s", want, out)
		}
	}
	// Frontmatter fence
	if !strings.HasPrefix(out, "---\n") {
		t.Errorf("rendered issue should open with frontmatter fence; got prefix %q", out[:20])
	}
}

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
