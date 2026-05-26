package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Pure-helper tests ────────────────────────────────────────────────────────

func TestExtractFirstTitle(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"# Hello\nbody", "Hello"},
		{"---\nid: 1\n---\n\n# Title here\n", "Title here"},
		{"no title at all\n## Subhead", ""},
		{"  leading space # not-title\n# Real Title", "Real Title"},
		{"# Trimmed   \n", "Trimmed"},
		{"", ""},
	}
	for _, c := range cases {
		t.Run(c.want, func(t *testing.T) {
			if got := extractFirstTitle(c.in); got != c.want {
				t.Errorf("extractFirstTitle(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestIsTerminalStatus(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"done", true},
		{"wontfix", true},
		{"punt", true},
		{"working", false},
		{"open", false},
		{"blocked", false},
		{"", false},
		{"DONE", false}, // case-sensitive — matches shell
	}
	for _, c := range cases {
		if got := isTerminalStatus(c.s); got != c.want {
			t.Errorf("isTerminalStatus(%q) = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestSplitNonEmptyLines(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"\n\n\n", nil},
		{"a\nb\nc", []string{"a", "b", "c"}},
		{"  a  \n\n  b \n", []string{"a", "b"}},
	}
	for _, c := range cases {
		got := splitNonEmptyLines(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("splitNonEmptyLines(%q) = %v, want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("[%d] %q vs %q", i, got[i], c.want[i])
			}
		}
	}
}

// ── buildPushCommitMessage ───────────────────────────────────────────────────

// pushTestRunner stubs only what buildPushCommitMessage uses: `git diff
// --quiet -- <path>` (twice — unstaged + cached). We mark a file as "dirty"
// by making its diff return an error (non-zero exit).
type pushTestRunner struct {
	captureRunner
	dirty map[string]bool // file path → has changes
}

func (r *pushTestRunner) Git(args ...string) ([]byte, error) {
	r.gitCalls = append(r.gitCalls, append([]string{}, args...))
	// "diff --quiet [--cached] -- <path>" → exit 1 iff path is dirty.
	if len(args) >= 2 && args[0] == "diff" {
		for _, a := range args {
			if r.dirty[a] {
				return nil, &fakeExitErr{}
			}
		}
		return nil, nil
	}
	return nil, nil
}

type fakeExitErr struct{}

func (fakeExitErr) Error() string { return "exit status 1" }

func TestBuildPushCommitMessage_NoChanges(t *testing.T) {
	tmp := t.TempDir()
	r := &pushTestRunner{}
	got := buildPushCommitMessage(tmp, r)
	if got != "auto-commit before push" {
		t.Errorf("expected fallback message, got %q", got)
	}
}

func TestBuildPushCommitMessage_SingleIssue(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "000031-target.md")
	if err := os.WriteFile(path, []byte("---\nid: 31\n---\n\n# Target title here\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &pushTestRunner{dirty: map[string]bool{path: true}}
	got := buildPushCommitMessage(tmp, r)
	if got != "Target title here" {
		t.Errorf("got %q, want %q", got, "Target title here")
	}
}

func TestBuildPushCommitMessage_MultipleIssues(t *testing.T) {
	tmp := t.TempDir()
	a := filepath.Join(tmp, "000031-a.md")
	b := filepath.Join(tmp, "000032-b.md")
	if err := os.WriteFile(a, []byte("# First title\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("# Second title\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &pushTestRunner{dirty: map[string]bool{a: true, b: true}}
	got := buildPushCommitMessage(tmp, r)
	want := "First title\nSecond title"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildPushCommitMessage_OnlyDirtyOnesContribute(t *testing.T) {
	tmp := t.TempDir()
	clean := filepath.Join(tmp, "000010-clean.md")
	dirty := filepath.Join(tmp, "000020-dirty.md")
	for _, p := range []string{clean, dirty} {
		base := filepath.Base(p)
		if err := os.WriteFile(p, []byte("# Title for "+base+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	r := &pushTestRunner{dirty: map[string]bool{dirty: true}}
	got := buildPushCommitMessage(tmp, r)
	if got != "Title for 000020-dirty.md" {
		t.Errorf("got %q, expected only dirty file's title", got)
	}
}

// ── touchedIssuesNotDone ─────────────────────────────────────────────────────

// notDoneRunner stubs `git diff --name-only` for the touched-issues query.
type notDoneRunner struct {
	captureRunner
	touched []byte
}

func (r *notDoneRunner) Git(args ...string) ([]byte, error) {
	r.gitCalls = append(r.gitCalls, append([]string{}, args...))
	if len(args) >= 2 && args[0] == "diff" && args[1] == "--name-only" {
		return r.touched, nil
	}
	return nil, nil
}

func TestTouchedIssuesNotDone(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	issuesDir := "workshop/issues"
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	mkIssue := func(name, status string) {
		p := filepath.Join(issuesDir, name)
		content := "---\nid: 0\nstatus: " + status + "\n---\n\n# X\n"
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mkIssue("000001-working.md", "working")
	mkIssue("000002-done.md", "done")
	mkIssue("000003-open.md", "open")

	r := &notDoneRunner{touched: []byte("workshop/issues/000001-working.md\nworkshop/issues/000002-done.md\nworkshop/issues/000003-open.md\n")}
	notDone, err := touchedIssuesNotDone("origin/main", issuesDir, r)
	if err != nil {
		t.Fatal(err)
	}
	// Expect 000001 (working) and 000003 (open).
	if len(notDone) != 2 {
		t.Fatalf("got %d not-done; want 2: %v", len(notDone), notDone)
	}
	if !strings.Contains(notDone[0], "000001") || !strings.Contains(notDone[1], "000003") {
		t.Errorf("entries: %v", notDone)
	}
}

// ── archiveDoneIssues ────────────────────────────────────────────────────────

// ghCallStub embeds stubGH (which provides PRCreate/PRListForBranch/PRMerge
// no-ops) and overrides IssueClose to record what was closed. Pointer
// receiver on IssueClose so the append survives the assignment.
type ghCallStub struct {
	stubGH
	closed []string // issueNum values that IssueClose was called with
}

func (g *ghCallStub) IssueClose(repo, issueNum, comment string) error {
	g.closed = append(g.closed, issueNum)
	return nil
}

func TestArchiveDoneIssues_MovesAndClosesGH(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "workshop", "issues")
	historyDir := filepath.Join(tmp, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	mk := func(name, status, gh string) {
		p := filepath.Join(issuesDir, name)
		body := "---\nid: 0\nstatus: " + status + "\ngithub_issue: " + gh + "\n---\n\n# T\n"
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("000001-done.md", "done", "100")
	mk("000002-wontfix.md", "wontfix", "")  // wontfix has no GH close
	mk("000003-punt.md", "punt", "200")     // punt has no GH close even with gh number
	mk("000004-working.md", "working", "300") // working stays put

	prev := ghClient
	stub := &ghCallStub{}
	ghClient = stub
	defer func() { ghClient = prev }()

	var stderr bytes.Buffer
	moved, err := archiveDoneIssues(&stderr, "owner/repo", issuesDir, historyDir)
	if err != nil {
		t.Fatal(err)
	}
	if moved != 3 {
		t.Errorf("moved = %d, want 3", moved)
	}
	// Only the done issue with a github_issue should have been closed.
	if len(stub.closed) != 1 || stub.closed[0] != "100" {
		t.Errorf("closed = %v, want [100]", stub.closed)
	}
	// Working file stays put.
	if _, err := os.Stat(filepath.Join(issuesDir, "000004-working.md")); err != nil {
		t.Errorf("working issue should still be in issues/: %v", err)
	}
	// Done file moved.
	if _, err := os.Stat(filepath.Join(historyDir, "000001-done.md")); err != nil {
		t.Errorf("done issue should be in history/: %v", err)
	}
}

func TestArchiveDoneIssues_NoneToArchive(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "issues")
	historyDir := filepath.Join(tmp, "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(issuesDir, "000010-working.md")
	_ = os.WriteFile(p, []byte("---\nstatus: working\n---\n\n# x\n"), 0o644)

	var stderr bytes.Buffer
	moved, err := archiveDoneIssues(&stderr, "owner/repo", issuesDir, historyDir)
	if err != nil {
		t.Fatal(err)
	}
	if moved != 0 {
		t.Errorf("moved = %d, want 0", moved)
	}
}

// ── Edge cases via runPush refusals ──────────────────────────────────────────

// runPush calls die() (which exits) when refusal conditions are hit.
// We can't drive runPush end-to-end without process exit, so the high-
// level coverage lives in build-tag-free smoke checks above. The
// refusal-on-branch test instead exercises the runner-stub path: ensure
// the early git ls-files isn't called when branch != main (the die is
// the first thing).
//
// Since die() calls os.Exit, we don't test it here — close.go and the
// other verbs have the same posture; the integration smoke check at the
// `make sdlc-build && sdlc push --dry-run` level is the cross-cutting
// path. The unit-test surface focuses on the pure helpers above.

// Confirm the cobra command is registered and has the expected flags.
func TestPushCmd_Registered(t *testing.T) {
	cmd := NewPushCmd()
	for _, flag := range []string{"yes", "no-judge", "dry-run", "issues-dir", "history-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("push command missing flag: --%s", flag)
		}
	}
}

// Ensure that runPush's dry-run path with --no-judge writes nothing
// alarming when the directory state is "no untracked, no dirty, no
// touched issues". This is the closest we can get to a smoke test in
// pure Go (no subprocess) without spinning up a real git repo.
func TestRunPush_DryRun_NoOpEnvironment(t *testing.T) {
	t.Skip("requires a real git repo on `main` with origin set; smoke-tested manually via Makefile.workflow")
}

// silence unused-import warnings in cases the file shrinks
var _ io.Writer = (*bytes.Buffer)(nil)
