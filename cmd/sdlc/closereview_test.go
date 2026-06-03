package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// closeRepo builds a minimal temp git repo with one no-milestone issue file
// committed under a bare `#<issue>` subject (the §12 convention — not zero-
// padded, which is what resolveReviewWindow matches), and chdir's into it. The
// issue is status:working with an all-checked ## Plan so runClose's gates pass
// when given Actual+Verified+NoAtlas. Returns issuesDir; restores cwd on cleanup.
func closeRepo(t *testing.T, issueNum int) string {
	t.Helper()
	padded := fmt.Sprintf("%06d", issueNum)
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	runGit("init", "-q", "-b", "main")
	runGit("config", "user.email", "test@example.com")
	runGit("config", "user.name", "Test")
	runGit("config", "commit.gpgsign", "false")

	// An initial commit so the #<issue> commit has a parent (baseLong = firstSHA^).
	if err := os.WriteFile("README", []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", "README")
	runGit("commit", "-q", "-m", "init")

	issuesDir := "workshop/issues"
	if err := os.MkdirAll(issuesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	issuePath := filepath.Join(issuesDir, padded+"-x.md")
	content := "---\nid: " + padded + "\nstatus: working\nestimate_hours: 1\n---\n\n" +
		"# x\n\n## Spec\n\nThing.\n\n## Plan\n\n- [x] do it\n\n## Log\n"
	if err := os.WriteFile(issuePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit("add", ".")
	runGit("commit", "-q", "-m", fmt.Sprintf("#%d: implement the thing", issueNum))
	return issuesDir
}

// stubJudge swaps judge.Run for a recorder; returns a pointer to the call count
// and the last prompt seen. Restores on cleanup.
func stubJudge(t *testing.T, output string) (*int, *string) {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	calls := 0
	var lastPrompt string
	judge.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		calls++
		if len(args) > 0 {
			lastPrompt = args[len(args)-1] // BuildArgs puts the prompt last
		}
		return []byte(output), nil
	}
	return &calls, &lastPrompt
}

// #69 (load-bearing invariant): a standalone full-issue close auto-dispatches
// exactly one boundary review on the whole-issue window and emits its trailer.
func TestRunCloseWithReview_IssueClose_Dispatches(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	calls, lastPrompt := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nLooks good.\n")

	var stdout strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, io.Discard, f); err != nil {
		t.Fatalf("runCloseWithReview: %v", err)
	}
	if *calls != 1 {
		t.Fatalf("expected exactly 1 review dispatch, got %d", *calls)
	}
	if !strings.Contains(*lastPrompt, "ariadne#69") {
		t.Errorf("dispatched prompt missing issue ref ariadne#69")
	}
	out := stdout.String()
	for _, want := range []string{"── close trailers", "Review-Verdict: SHIP", "..HEAD"} {
		if !strings.Contains(out, want) {
			t.Errorf("close stdout missing %q:\n%s", want, out)
		}
	}
	// The verdict is also mirrored into the close log line (#69 M2 review I1).
	if got := readIssue(t, issuesDir); !strings.Contains(got, "closed — tests pass; review verdict: SHIP") {
		t.Errorf("issue ## Log line missing the verdict annotation:\n%s", got)
	}
}

func readIssue(t *testing.T, issuesDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(issuesDir, "000069-x.md"))
	if err != nil {
		t.Fatalf("read issue: %v", err)
	}
	return string(data)
}

// #69 guard: a milestone close routed through runClose (as milestone-close does)
// must NOT dispatch the whole-issue review — that's milestone-close's own job.
// This is the structural invariant that keeps it "exactly one review per
// boundary" on the multi-milestone path.
func TestRunCloseWithReview_MilestoneClose_DoesNotDispatch(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	calls, _ := stubJudge(t, "VERDICT: SHIP\n")

	f := &closeFlags{
		Issue: 69, Milestone: "M1", Actual: "1", Verified: "slice done", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("runCloseWithReview (milestone): %v", err)
	}
	if *calls != 0 {
		t.Fatalf("milestone close must not dispatch the issue review, got %d dispatch(es)", *calls)
	}
}

// #69: --no-judge on a full-issue close skips the dispatch but still records the
// boundary (not-run trailer), per the #67 per-gate-bypass convention.
func TestRunCloseWithReview_NoJudge_Skips(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	calls, _ := stubJudge(t, "VERDICT: SHIP\n")

	var stdout strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true, NoJudge: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, io.Discard, f); err != nil {
		t.Fatalf("runCloseWithReview: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("--no-judge must skip dispatch, got %d", *calls)
	}
	if !strings.Contains(stdout.String(), "Review-Verdict: not-run") {
		t.Errorf("--no-judge close should still emit a not-run trailer:\n%s", stdout.String())
	}
	// I1: the not-run verdict is mirrored into the log line too (parity with
	// milestone-close), not just the trailer.
	if got := readIssue(t, issuesDir); !strings.Contains(got, "; review verdict: not-run") {
		t.Errorf("--no-judge close should still annotate the log line:\n%s", got)
	}
}
