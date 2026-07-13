package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// errStub is a sentinel git failure for probe/runner tests.
var errStub = errors.New("git failed")

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
		if got := vocab.Issue().IsTerminal(c.s); got != c.want {
			t.Errorf("vocab.Issue().IsTerminal(%q) = %v, want %v", c.s, got, c.want)
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

type archiveRecoveryRunner struct {
	captureRunner
	status []byte
}

func (r *archiveRecoveryRunner) Git(args ...string) ([]byte, error) {
	r.gitCalls = append(r.gitCalls, append([]string{}, args...))
	if len(args) >= 3 && args[0] == "status" && args[1] == "--porcelain" && args[2] == "--untracked-files=all" {
		return r.status, nil
	}
	return nil, nil
}

func callsJoined(calls [][]string) string {
	var lines []string
	for _, c := range calls {
		lines = append(lines, strings.Join(c, " "))
	}
	return strings.Join(lines, "\n")
}

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

// ── interrupted archive recovery ────────────────────────────────────────────

func writeArchiveCandidate(t *testing.T, path, status string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nid: 0\nstatus: " + status + "\n---\n\n# T\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestPreparedArchiveMovesDetectsUnstagedMove(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	writeArchiveCandidate(t, "workshop/history/000036-done.md", "done")

	status := " D workshop/issues/000036-done.md\n?? workshop/history/000036-done.md\n"
	moves, other, err := preparedArchiveMoves(status, "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil {
		t.Fatal(err)
	}
	if len(other) != 0 {
		t.Fatalf("other = %v, want none", other)
	}
	if len(moves) != 1 || moves[0].IssuePath != "workshop/issues/000036-done.md" || moves[0].HistoryPath != "workshop/history/000036-done.md" {
		t.Fatalf("moves = %#v", moves)
	}
}

func TestPreparedArchiveMovesRejectsNonTerminalHistoryFile(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	writeArchiveCandidate(t, "workshop/history/000036-open.md", "open")

	status := " D workshop/issues/000036-open.md\n?? workshop/history/000036-open.md\n"
	moves, other, err := preparedArchiveMoves(status, "workshop/issues", "workshop/history", "workshop/plans")
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 0 {
		t.Fatalf("moves = %#v, want none", moves)
	}
	if len(other) != 2 {
		t.Fatalf("other = %v, want both halves refused", other)
	}
}

func TestRecoverInterruptedArchiveCommitsAndPushes(t *testing.T) {
	tmp := t.TempDir()
	cwd, _ := os.Getwd()
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	writeArchiveCandidate(t, "workshop/history/000036-done.md", "done")

	prev := pushRunner
	r := &archiveRecoveryRunner{
		status: []byte(" D workshop/issues/000036-done.md\n?? workshop/history/000036-done.md\n"),
	}
	pushRunner = r
	defer func() { pushRunner = prev }()

	var stdout, stderr bytes.Buffer
	recovered, err := recoverInterruptedArchive(&stdout, &stderr, &pushFlags{
		IssuesDir:  "workshop/issues",
		HistoryDir: "workshop/history",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !recovered {
		t.Fatal("expected recovery")
	}
	got := callsJoined(r.gitCalls)
	for _, want := range []string{
		"status --porcelain --untracked-files=all",
		// Precise add of the exact prepared move — not the broad `add <dir>/`
		// that swept unrelated untracked WIP onto main (#80).
		"add -- workshop/issues/000036-done.md workshop/history/000036-done.md",
		"commit -m archive completed issues to history",
		"push",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("git calls missing %q:\n%s", want, got)
		}
	}
}

// ── touchedIssuesNotDone ─────────────────────────────────────────────────────

// notDoneRunner stubs `git diff --name-only` for the touched-issues query.
type notDoneRunner struct {
	captureRunner
	touched    []byte
	touchedErr error
}

func (r *notDoneRunner) Git(args ...string) ([]byte, error) {
	r.gitCalls = append(r.gitCalls, append([]string{}, args...))
	if len(args) >= 2 && args[0] == "diff" && args[1] == "--name-only" {
		return r.touched, r.touchedErr
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
	// #160: codecomplete is the normal pre-publish state — the publish gate is about
	// to flip it to done — so it must NOT be flagged "not done" (else every merge/push
	// would trip the "Continue anyway?" prompt). This pins the one-token carve-out.
	mkIssue("000004-cc.md", "codecomplete")
	missingStatus := filepath.Join(issuesDir, "000005-missing.md")
	if err := os.WriteFile(missingStatus, []byte("---\nid: 5\n---\n\n# X\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	r := &notDoneRunner{touched: []byte("workshop/issues/000005-missing.md\nworkshop/issues/000001-working.md\nworkshop/issues/000002-done.md\nworkshop/issues/000003-open.md\nworkshop/issues/000004-cc.md\n")}
	notDone, err := touchedIssuesNotDone("origin/main", issuesDir, r)
	if err != nil {
		t.Fatal(err)
	}
	// Expect missing, 000001 (working), and 000003 (open), in git order;
	// NOT 000002 (done) or 000004 (codecomplete).
	if len(notDone) != 3 {
		t.Fatalf("got %d not-done; want 3: %v", len(notDone), notDone)
	}
	if got, want := notDone[0], "workshop/issues/000005-missing.md (status: unset)"; got != want {
		t.Errorf("missing-status entry = %q, want %q", got, want)
	}
	if !strings.Contains(notDone[1], "000001") || !strings.Contains(notDone[2], "000003") {
		t.Errorf("entries: %v", notDone)
	}
	for _, e := range notDone {
		if strings.Contains(e, "000004") {
			t.Errorf("codecomplete issue must NOT be flagged not-done (#160): %v", notDone)
		}
	}
}

func TestTouchedIssuesNotDonePreservesGitOutputOnFailure(t *testing.T) {
	cause := errors.New("exit status 128")
	r := &notDoneRunner{touched: []byte("fatal: bad revision\n"), touchedErr: cause}
	_, err := touchedIssuesNotDone("origin/main", "workshop/issues", r)
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "git diff origin/main..HEAD: exit status 128\nfatal: bad revision\n"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
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
	mk("000002-wontfix.md", "wontfix", "")    // wontfix has no GH close
	mk("000003-punt.md", "punt", "200")       // punt has no GH close even with gh number
	mk("000004-working.md", "working", "300") // working stays put

	prev := ghClient
	stub := &ghCallStub{}
	ghClient = stub
	defer func() { ghClient = prev }()

	var stderr bytes.Buffer
	moves, err := archiveDoneIssues(&stderr, "owner/repo", issuesDir, historyDir, filepath.Join(issuesDir, "..", "plans"))
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 3 {
		t.Errorf("moved = %d, want 3", len(moves))
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

// #160: the push publish sequence — step 6.5 flip (codecomplete → done) then step 7
// archive — must land a codecomplete issue in history/ as done. Mirrors merge's
// TestRunMerge_CodecompleteFlippedToDoneAndArchived for the direct-to-main path.
func TestPushPublishSequence_CodecompleteFlippedThenArchived(t *testing.T) {
	tmp := t.TempDir()
	issuesDir := filepath.Join(tmp, "workshop", "issues")
	historyDir := filepath.Join(tmp, "workshop", "history")
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cc := "---\nid: 160\nstatus: codecomplete\nactual_hours: 1\n---\n\n# cc\n"
	if err := os.WriteFile(filepath.Join(issuesDir, "000160-cc.md"), []byte(cc), 0o644); err != nil {
		t.Fatal(err)
	}

	// Step 6.5: flip codecomplete → done.
	flipped, err := publishCodecompleteIssues(issuesDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(flipped) != 1 {
		t.Fatalf("flipped %d, want 1", len(flipped))
	}
	// Step 7: archive (now terminal).
	var stderr bytes.Buffer
	moves, err := archiveDoneIssues(&stderr, "", issuesDir, historyDir, filepath.Join(issuesDir, "..", "plans"))
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 1 {
		t.Fatalf("archived %d, want 1", len(moves))
	}
	data, err := os.ReadFile(filepath.Join(historyDir, "000160-cc.md"))
	if err != nil {
		t.Fatalf("codecomplete issue should be archived to history: %v", err)
	}
	if !strings.Contains(string(data), "status: done") {
		t.Errorf("archived issue should be flipped codecomplete → done:\n%s", data)
	}
}

// archiveAddArgs must stage exactly the moved paths (src deletion + history
// addition), behind a `--` separator, and never a directory — the broad
// `git add <dir>/` is what swept unrelated untracked WIP onto main (#80).
func TestArchiveAddArgs(t *testing.T) {
	cases := []struct {
		name  string
		moves []preparedArchiveMove
		want  []string
	}{
		{
			name:  "empty",
			moves: nil,
			want:  []string{"add", "--"},
		},
		{
			name:  "one move",
			moves: []preparedArchiveMove{{IssuePath: "workshop/issues/000001-done.md", HistoryPath: "workshop/history/000001-done.md"}},
			want:  []string{"add", "--", "workshop/issues/000001-done.md", "workshop/history/000001-done.md"},
		},
		{
			name: "two moves",
			moves: []preparedArchiveMove{
				{IssuePath: "workshop/issues/000001-done.md", HistoryPath: "workshop/history/000001-done.md"},
				{IssuePath: "workshop/issues/000002-punt.md", HistoryPath: "workshop/history/000002-punt.md"},
			},
			want: []string{"add", "--",
				"workshop/issues/000001-done.md", "workshop/history/000001-done.md",
				"workshop/issues/000002-punt.md", "workshop/history/000002-punt.md"},
		},
		{
			// #154: an untracked source (vanished at the rename) must NOT be staged —
			// `git add <it>` would fail "pathspec did not match". Only its dest.
			name:  "untracked source stages dest only",
			moves: []preparedArchiveMove{{IssuePath: "workshop/plans/000154-x-close-review.md", HistoryPath: "workshop/history/000154-x-close-review.md", SourceUntracked: true}},
			want:  []string{"add", "--", "workshop/history/000154-x-close-review.md"},
		},
		{
			// A tracked issue/plan move alongside an untracked sidecar: the tracked
			// one stages both halves, the untracked one stages dest only.
			name: "mixed tracked and untracked",
			moves: []preparedArchiveMove{
				{IssuePath: "workshop/issues/000154-x.md", HistoryPath: "workshop/history/000154-x.md"},
				{IssuePath: "workshop/plans/000154-x-close-review.md", HistoryPath: "workshop/history/000154-x-close-review.md", SourceUntracked: true},
			},
			want: []string{"add", "--",
				"workshop/issues/000154-x.md", "workshop/history/000154-x.md",
				"workshop/history/000154-x-close-review.md"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := archiveAddArgs(tc.moves)
			if len(got) != len(tc.want) {
				t.Fatalf("archiveAddArgs = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("arg[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
			// No arg is ever a bare directory path (the #80 hazard).
			for _, a := range got {
				if strings.HasSuffix(a, "/") {
					t.Errorf("arg %q is a directory — broad add reintroduced (#80)", a)
				}
			}
		})
	}
}

// #154: gitSrcUntracked classifies a source as untracked ONLY when `git ls-files`
// cleanly reports no index entry (empty output, no error). A tracked path (git
// echoes it) and any git error both classify as tracked — the conservative
// direction that preserves the pre-#154 "stage the source deletion" behavior.
func TestGitSrcUntracked(t *testing.T) {
	cases := []struct {
		name string
		out  []byte
		err  error
		want bool // true = untracked
	}{
		{"empty output → untracked", []byte(""), nil, true},
		{"whitespace-only output → untracked", []byte("\n"), nil, true},
		{"path echoed → tracked", []byte("workshop/plans/000154-x-plan.md\n"), nil, false},
		{"git error → tracked (conservative)", nil, errStub, false},
		{"error with stale output → tracked", []byte("workshop/plans/000154-x-plan.md\n"), errStub, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotArgs []string
			probe := gitSrcUntracked(func(args ...string) ([]byte, error) {
				gotArgs = args
				return tc.out, tc.err
			})
			if got := probe("workshop/plans/000154-x-plan.md"); got != tc.want {
				t.Errorf("gitSrcUntracked = %v, want %v", got, tc.want)
			}
			// Always an index-scoped ls-files, `--`-guarded, on the exact path.
			want := []string{"ls-files", "--", "workshop/plans/000154-x-plan.md"}
			if strings.Join(gotArgs, " ") != strings.Join(want, " ") {
				t.Errorf("git args = %v, want %v", gotArgs, want)
			}
		})
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
	moves, err := archiveDoneIssues(&stderr, "owner/repo", issuesDir, historyDir, filepath.Join(issuesDir, "..", "plans"))
	if err != nil {
		t.Fatal(err)
	}
	if len(moves) != 0 {
		t.Errorf("moved = %d, want 0", len(moves))
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
