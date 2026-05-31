package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── Pure-helper tests ────────────────────────────────────────────────────────

func TestFormatFixes(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"42"}, "Fixes #42"},
		{[]string{"42", "43"}, "Fixes #42, #43"},
		{[]string{"1", "100", "10"}, "Fixes #1, #100, #10"}, // input order preserved (caller pre-sorts)
		// Cross-repo / qualified github_issue values are used verbatim (no
		// leading "#") — GitHub closes "owner/repo#3", not "#owner/repo#3".
		{[]string{"xianxu/you-decide#3"}, "Fixes xianxu/you-decide#3"},
		{[]string{"42", "owner/repo#7"}, "Fixes #42, owner/repo#7"},
		{[]string{"#9"}, "Fixes #9"}, // already-hashed → unchanged
	}
	for _, c := range cases {
		got := formatFixes(c.in)
		if got != c.want {
			t.Errorf("formatFixes(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCombineBody(t *testing.T) {
	cases := []struct {
		commits, fixes, want string
	}{
		{"", "", ""},
		{"- one", "", "- one"},
		{"", "Fixes #1", "Fixes #1"},
		{"- one\n- two", "Fixes #5", "- one\n- two\n\nFixes #5"},
		{"  - one  ", "  Fixes #5  ", "- one\n\nFixes #5"}, // trims edge whitespace
	}
	for _, c := range cases {
		got := combineBody(c.commits, c.fixes)
		if got != c.want {
			t.Errorf("combineBody(%q,%q) = %q, want %q", c.commits, c.fixes, got, c.want)
		}
	}
}

func TestCollectGitHubIssueNumbers(t *testing.T) {
	tmp := t.TempDir()
	mk := func(name, gh string) string {
		p := filepath.Join(tmp, name)
		body := "---\nid: 1\nstatus: done\n"
		if gh != "" {
			body += "github_issue: " + gh + "\n"
		}
		body += "---\n\n# T\n"
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	a := mk("a.md", "42")
	b := mk("b.md", "10")
	c := mk("c.md", "") // no github_issue → skipped
	d := mk("d.md", "42") // duplicate of a → deduped

	got := collectGitHubIssueNumbers([]string{a, b, c, d})
	// Numerically sorted: 10, 42
	if len(got) != 2 {
		t.Fatalf("got %v, want [10 42]", got)
	}
	if got[0] != "10" || got[1] != "42" {
		t.Errorf("got %v, want [10 42]", got)
	}
}

func TestCollectGitHubIssueNumbers_SkipsMissingFiles(t *testing.T) {
	got := collectGitHubIssueNumbers([]string{"/does/not/exist.md", "/also/missing.md"})
	if len(got) != 0 {
		t.Errorf("got %v, want empty", got)
	}
}

// ── pr command flag wiring ───────────────────────────────────────────────────

func TestPRCmd_Registered(t *testing.T) {
	cmd := NewPRCmd()
	for _, flag := range []string{"dry-run", "issues-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("pr command missing flag: --%s", flag)
		}
	}
}

// ── runPR integration via stubbed ghClient ───────────────────────────────────

// prTestRunner stubs git for runPR. We pre-load responses keyed by the
// first 1-3 args (same shape as claimRunnerStub).
type prTestRunner struct {
	captureRunner
	responses map[string][]byte
}

func (r *prTestRunner) Git(args ...string) ([]byte, error) {
	r.gitCalls = append(r.gitCalls, append([]string{}, args...))
	for n := min(len(args), 3); n > 0; n-- {
		key := strings.Join(args[:n], " ")
		if v, ok := r.responses[key]; ok {
			return v, nil
		}
	}
	return nil, nil
}

// recordingGH captures PRCreate args for assertions.
type recordingGH struct {
	stubGH
	prCreated struct {
		repo, base, head, body string
		called                 bool
	}
}

func (g *recordingGH) PRCreate(repo, base, head, body string) (string, error) {
	g.prCreated.repo = repo
	g.prCreated.base = base
	g.prCreated.head = head
	g.prCreated.body = body
	g.prCreated.called = true
	return "https://github.com/owner/repo/pull/123", nil
}

func TestCollectGitHubIssueNumbers_DedupesAndOrders(t *testing.T) {
	tmp := t.TempDir()
	mkIssue := func(name, gh string) string {
		p := filepath.Join(tmp, name)
		body := "---\nid: 1\ngithub_issue: " + gh + "\n---\n\n# X\n"
		_ = os.WriteFile(p, []byte(body), 0o644)
		return p
	}
	a := mkIssue("a.md", "5")
	b := mkIssue("b.md", "1")
	c := mkIssue("c.md", "10")
	d := mkIssue("d.md", "1") // dup

	got := collectGitHubIssueNumbers([]string{a, b, c, d})
	want := []string{"1", "5", "10"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] %q vs %q", i, got[i], want[i])
		}
	}
}
