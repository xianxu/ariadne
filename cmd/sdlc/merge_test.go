package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── #62 M1: worktree clean re-check before the irreversible merge ────────────
func TestWorktreeDirty(t *testing.T) {
	cases := []struct {
		name      string
		porcelain string
		wantDirty bool
	}{
		{"clean", "", false},
		{"clean-whitespace-only", "  \n\n", false},
		{"dirty-modified", " M atlas/x.md\n", true},
		{"dirty-untracked", "?? deps\n", true},
	}
	for _, tc := range cases {
		r := &claimRunnerStub{responses: map[string][]byte{
			"status --porcelain": []byte(tc.porcelain),
		}}
		dirty, err := worktreeDirty(r)
		if err != nil {
			t.Fatalf("%s: unexpected err: %v", tc.name, err)
		}
		if (dirty != "") != tc.wantDirty {
			t.Errorf("%s: dirty=%v (%q), want dirty=%v", tc.name, dirty != "", dirty, tc.wantDirty)
		}
	}
}

// ── #78: untracked files don't block; #82 M2: tracker files never block ──────
func TestAssessDirty(t *testing.T) {
	cases := []struct {
		name          string
		porcelain     string
		wantRefuse    bool
		wantBlocking  int
		wantUntracked int
		wantTracker   int
	}{
		{"clean", "", false, 0, 0, 0},
		{"whitespace-only", "  \n\n", false, 0, 0, 0},
		{"untracked-only proceeds", "?? deps\n?? construct/local/x\n", false, 0, 2, 0},
		{"modified refuses", " M atlas/x.md\n", true, 1, 0, 0},
		{"staged refuses", "A  cmd/sdlc/new.go\n", true, 1, 0, 0},
		{"deleted refuses", " D old.txt\n", true, 1, 0, 0},
		{"mixed refuses but still surfaces untracked", " M x.go\n?? y.tmp\n", true, 1, 1, 0},
		// #82 M2 — tracker files are never blocking, tracked-modified OR untracked.
		{"dirty tracked issue file proceeds", " M workshop/issues/000082-x.md\n", false, 0, 0, 1},
		{"staged issue file proceeds", "A  workshop/issues/000082-x.md\n", false, 0, 0, 1},
		{"untracked issue file proceeds", "?? workshop/issues/000888-wip.md\n", false, 0, 0, 1},
		{"dirty history file proceeds", " M workshop/history/000080-done.md\n", false, 0, 0, 1},
		{"renamed into history proceeds", "R  workshop/issues/000080-x.md -> workshop/history/000080-x.md\n", false, 0, 0, 1},
		// Regression: worktreeDirty whole-trims its output, so the FIRST porcelain
		// line loses its leading status space ("M workshop/..." not " M workshop/
		// ..."). Column-slicing would mis-read the path and bucket it as Blocking;
		// field-splitting must still see it as Tracker.
		{"leading-space-trimmed issue line proceeds", "M workshop/issues/000082-x.md\n", false, 0, 0, 1},
		// The crux both-directions case: a dirty issue file must NOT rescue a
		// dirty CODE file — code still blocks, issue file is bucketed to Tracker.
		{"dirty code still blocks despite dirty issue file", " M cmd/sdlc/merge.go\n M workshop/issues/000082-x.md\n", true, 1, 0, 1},
		// A markdown file OUTSIDE the tracker dirs is still code-class (blocks).
		{"non-tracker markdown still blocks", " M atlas/workflow/x.md\n", true, 1, 0, 0},
	}
	for _, tc := range cases {
		d := assessDirty(tc.porcelain, "workshop/issues", "workshop/history")
		if d.Refuse() != tc.wantRefuse {
			t.Errorf("%s: Refuse()=%v, want %v (assessment=%+v)", tc.name, d.Refuse(), tc.wantRefuse, d)
		}
		if len(d.Blocking) != tc.wantBlocking {
			t.Errorf("%s: len(Blocking)=%d, want %d (%q)", tc.name, len(d.Blocking), tc.wantBlocking, d.Blocking)
		}
		if len(d.Untracked) != tc.wantUntracked {
			t.Errorf("%s: len(Untracked)=%d, want %d (%q)", tc.name, len(d.Untracked), tc.wantUntracked, d.Untracked)
		}
		if len(d.Tracker) != tc.wantTracker {
			t.Errorf("%s: len(Tracker)=%d, want %d (%q)", tc.name, len(d.Tracker), tc.wantTracker, d.Tracker)
		}
	}
}

// TestPorcelainPaths pins the field-based extractor that #82 M2 relies on —
// robust to whether the leading status space was trimmed (worktreeDirty does).
func TestPorcelainPaths(t *testing.T) {
	cases := []struct {
		line, path, dest string
	}{
		{" M workshop/issues/000082-x.md", "workshop/issues/000082-x.md", ""},
		{"M workshop/issues/000082-x.md", "workshop/issues/000082-x.md", ""}, // leading space trimmed
		{"?? workshop/issues/000888-wip.md", "workshop/issues/000888-wip.md", ""},
		{"A  cmd/sdlc/new.go", "cmd/sdlc/new.go", ""},
		{"R  workshop/issues/000080-x.md -> workshop/history/000080-x.md", "workshop/issues/000080-x.md", "workshop/history/000080-x.md"},
		{"", "", ""},
		{"   ", "", ""},
	}
	for _, tc := range cases {
		p, d := porcelainPaths(tc.line)
		if p != tc.path || d != tc.dest {
			t.Errorf("porcelainPaths(%q) = (%q,%q), want (%q,%q)", tc.line, p, d, tc.path, tc.dest)
		}
	}
}

// ── #62 M3: merge-vs-resume decision ─────────────────────────────────────────
func TestDecideMergeAction(t *testing.T) {
	cases := []struct {
		name         string
		openPR       string
		mergedExists bool
		want         mergeAction
	}{
		{"open PR → merge it", "42", false, actionMergeOpen},
		{"open PR present, merged irrelevant", "42", true, actionMergeOpen},
		{"no open PR but merged exists → resume", "", true, actionResume},
		{"no PR at all → no-PR path", "", false, actionNoPR},
	}
	for _, tc := range cases {
		if got := decideMergeAction(tc.openPR, tc.mergedExists); got != tc.want {
			t.Errorf("%s: decideMergeAction(%q,%v) = %d, want %d",
				tc.name, tc.openPR, tc.mergedExists, got, tc.want)
		}
	}
}

// ── merge command flag wiring ────────────────────────────────────────────────

func TestMergeCmd_Registered(t *testing.T) {
	cmd := NewMergeCmd()
	for _, flag := range []string{"yes", "no-judge", "dry-run", "issues-dir", "history-dir"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("merge command missing flag: --%s", flag)
		}
	}
}

// ── in-place vs worktree topology (#51) ──────────────────────────────────────

func TestIsInPlaceCheckout(t *testing.T) {
	cases := []struct {
		name   string
		gitDir string
		want   bool
	}{
		{"primary relative", ".git", true},
		{"primary absolute", "/Users/x/repo/.git", true},
		{"linked worktree absolute", "/Users/x/repo/.git/worktrees/000051-foo", false},
		{"linked worktree relative", ".git/worktrees/feature", false},
	}
	for _, c := range cases {
		if got := isInPlaceCheckout(c.gitDir); got != c.want {
			t.Errorf("%s: isInPlaceCheckout(%q) = %v, want %v", c.name, c.gitDir, got, c.want)
		}
	}
}

// ── archiveDoneIssuesInDir ───────────────────────────────────────────────────

func TestArchiveDoneIssuesInDir_MovesAndDoesNotCloseGH(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := "workshop/issues"
	historyDir := "workshop/history"
	fullIssues := filepath.Join(tmp, issuesDir)
	if err := os.MkdirAll(fullIssues, 0o755); err != nil {
		t.Fatal(err)
	}
	mk := func(name, status, gh string) {
		p := filepath.Join(fullIssues, name)
		body := "---\nid: 0\nstatus: " + status + "\n"
		if gh != "" {
			body += "github_issue: " + gh + "\n"
		}
		body += "---\n\n# T\n"
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("000001-done.md", "done", "100")
	mk("000002-working.md", "working", "200")

	// Track that IssueClose is NOT called (merge ships through PR which
	// closes via "Fixes #N" body — calling gh issue close would be a bug).
	stub := &ghCallStub{}
	prev := ghClient
	ghClient = stub
	defer func() { ghClient = prev }()

	var stderr stringWriter
	moves, err := archiveDoneIssuesInDir(&stderr, "owner/repo", tmp, issuesDir, historyDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 1 {
		t.Errorf("moved = %d, want 1", len(moves))
	}
	// Returned paths are mainPath-relative (so GitInDir resolves them) — never
	// absolute, or a precise `git add` from the main worktree would silently miss.
	if len(moves) == 1 {
		if got, want := moves[0].IssuePath, filepath.Join(issuesDir, "000001-done.md"); got != want {
			t.Errorf("IssuePath = %q, want relative %q", got, want)
		}
		if got, want := moves[0].HistoryPath, filepath.Join(historyDir, "000001-done.md"); got != want {
			t.Errorf("HistoryPath = %q, want relative %q", got, want)
		}
	}
	if len(stub.closed) != 0 {
		t.Errorf("merge must NOT call gh issue close (PR merge does it via Fixes); got closed = %v", stub.closed)
	}
	if _, err := os.Stat(filepath.Join(tmp, historyDir, "000001-done.md")); err != nil {
		t.Errorf("expected file in history/: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, issuesDir, "000002-working.md")); err != nil {
		t.Errorf("working file should remain in issues/: %v", err)
	}
}

func TestArchiveDoneIssuesInDir_EmptyTree(t *testing.T) {
	tmp := t.TempDir()
	fullIssues := filepath.Join(tmp, "workshop", "issues")
	if err := os.MkdirAll(fullIssues, 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr stringWriter
	moves, err := archiveDoneIssuesInDir(&stderr, "owner/repo", tmp, "workshop/issues", "workshop/history")
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 0 {
		t.Errorf("moved = %d, want 0", len(moves))
	}
}

// ── stdinPrompter / prompter substitution ────────────────────────────────────

// fakePrompter returns canned answers in order for each Ask call. Useful
// for end-to-end tests of merge's confirmation flow; we don't drive
// runMerge directly (it calls die() which exits), but the prompter
// indirection itself is unit-tested here.
type fakePrompter struct {
	answers []string
	asked   []string
}

func (p *fakePrompter) Ask(question string, w io.Writer) string {
	p.asked = append(p.asked, question)
	if len(p.answers) == 0 {
		return ""
	}
	ans := p.answers[0]
	p.answers = p.answers[1:]
	return ans
}

// TestMergeNeedsTTY pins the non-interactive guard: without --yes and
// without a tty, merge must refuse (rather than block on stdin) — the
// stall bug. --yes or --dry-run (no prompts) or a real tty all proceed.
func TestMergeNeedsTTY(t *testing.T) {
	cases := []struct {
		name                    string
		yes, dryRun, stdinIsTTY bool
		want                    bool
	}{
		{"non-tty, no --yes → refuse", false, false, false, true},
		{"--yes overrides non-tty", true, false, false, false},
		{"--dry-run has no prompts", false, true, false, false},
		{"real terminal answers prompts", false, false, true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := mergeNeedsTTY(c.yes, c.dryRun, c.stdinIsTTY); got != c.want {
				t.Errorf("mergeNeedsTTY(yes=%v dryRun=%v tty=%v) = %v, want %v",
					c.yes, c.dryRun, c.stdinIsTTY, got, c.want)
			}
		})
	}
}

func TestFakePrompter_Order(t *testing.T) {
	p := &fakePrompter{answers: []string{"y", "n"}}
	var w stringWriter
	if got := p.Ask("first?", &w); got != "y" {
		t.Errorf("first answer = %q, want y", got)
	}
	if got := p.Ask("second?", &w); got != "n" {
		t.Errorf("second answer = %q, want n", got)
	}
	if got := p.Ask("third?", &w); got != "" {
		t.Errorf("exhausted prompter should return \"\", got %q", got)
	}
	if len(p.asked) != 3 {
		t.Errorf("asked %d times, want 3", len(p.asked))
	}
}

// stringWriter is a tiny io.Writer that accumulates output without
// requiring bytes.Buffer (avoids the bytes import in this file).
type stringWriter struct{ b strings.Builder }

func (s *stringWriter) Write(p []byte) (int, error) {
	return s.b.Write(p)
}
func (s *stringWriter) String() string { return s.b.String() }

// ── findMainWorktree — exercised through merge ───────────────────────────────

// We reuse the lock_test's findMainWorktree coverage via the runner stub;
// merge.go calls the same helper. A redundant test here would just copy
// lock_test.go. We confirm the runner indirection works for merge by
// constructing the stub directly.

func TestMergeUsesFindMainWorktree_ViaStub(t *testing.T) {
	r := &claimRunnerStub{
		responses: map[string][]byte{"worktree list": []byte(
			"worktree /repo/main\nHEAD abc\nbranch refs/heads/main\n\n" +
				"worktree /repo/feat\nHEAD def\nbranch refs/heads/feature\n",
		)},
	}
	got, err := findMainWorktree(r)
	if err != nil {
		t.Fatal(err)
	}
	if got != "/repo/main" {
		t.Errorf("got %q, want /repo/main", got)
	}
}
