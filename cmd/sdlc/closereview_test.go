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
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		calls++
		if len(args) > 0 {
			lastPrompt = args[len(args)-1] // BuildArgs puts the prompt last
		}
		return []byte(output), nil
	}
	return &calls, &lastPrompt
}

func stubJudgeCommand(t *testing.T, output string) (*int, *string) {
	t.Helper()
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	calls := 0
	var lastName string
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		calls++
		lastName = name
		return []byte(output), nil
	}
	return &calls, &lastName
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
	// #137: the prompt's issue ref is derived from the live repo (the temp repo
	// here), NOT a hardcoded ariadne#69.
	wantRef := repoIdentity() + "#69"
	if !strings.Contains(*lastPrompt, wantRef) {
		t.Errorf("dispatched prompt missing derived issue ref %q", wantRef)
	}
	if strings.Contains(*lastPrompt, "ariadne#69") {
		t.Errorf("prompt must not hardcode ariadne#69 for a non-ariadne repo (#137)")
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

	// #136: the full review transcript is persisted to a durable sidecar under
	// workshop/plans/, so an agent can reopen it after scrollback loss.
	scData, err := os.ReadFile(filepath.Join("workshop/plans", "000069-x-close-review.md"))
	if err != nil {
		t.Fatalf("#136 review sidecar not written: %v", err)
	}
	for _, want := range []string{
		"# Boundary Review — " + repoIdentity() + "#69", "Looks good.", // #137: repo-derived H1
		"sdlc close --issue 69", "| verdict | SHIP |",
	} {
		if !strings.Contains(string(scData), want) {
			t.Errorf("#136 close sidecar missing %q:\n%s", want, scData)
		}
	}
	// The RESOLVED reviewer must reach the sidecar — the raw --agent flag is "" by
	// default, so an empty reviewer cell means the resolved agent wasn't threaded.
	if strings.Contains(string(scData), "| reviewer |  |") {
		t.Errorf("#136 sidecar reviewer cell is empty — resolved agent not threaded:\n%s", scData)
	}
}

// #136: the milestone-close boundary persists its review to a per-milestone
// sidecar (NNNNNN-slug-m<x>-review.md) via the same shared dispatch.
func TestDispatchBoundaryReview_WritesMilestoneSidecar(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nMilestone looks good.\n")

	res := dispatchBoundaryReview(io.Discard, io.Discard, boundaryReviewParams{
		Label:     "#69 M1",
		Base:      "HEAD",
		BaseLong:  "HEAD",
		Head:      "HEAD",
		IssuesDir: issuesDir,
		IssueNum:  69,
		Milestone: "M1",
		PlansDir:  "workshop/plans",
	})
	if filepath.Base(res.SidecarPath) != "000069-x-m1-review.md" {
		t.Fatalf("milestone sidecar path = %q, want …/000069-x-m1-review.md", res.SidecarPath)
	}
	data, err := os.ReadFile(res.SidecarPath)
	if err != nil {
		t.Fatalf("milestone sidecar not written: %v", err)
	}
	for _, want := range []string{
		"milestone M1", "Milestone looks good.",
		"sdlc milestone-close --issue 69 --milestone M1",
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("milestone sidecar missing %q:\n%s", want, data)
		}
	}
}

func TestRunCloseWithReview_AgentDefaultUsesPairAgent(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	issuesDir := closeRepo(t, 69)
	calls, lastName := stubJudgeCommand(t, "VERDICT: SHIP (confidence: high)\n\nLooks good.\n")

	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("runCloseWithReview: %v", err)
	}
	if *calls != 1 {
		t.Fatalf("expected exactly 1 review dispatch, got %d", *calls)
	}
	if *lastName != "codex" {
		t.Fatalf("close boundary review agent = %q, want codex", *lastName)
	}
}

func TestRunCloseWithReview_DryRunPrintsPairAgentCommand(t *testing.T) {
	t.Setenv("AGENT_CMD", "")
	t.Setenv("PAIR_AGENT", "codex")
	issuesDir := closeRepo(t, 69)
	calls, _ := stubJudgeCommand(t, "VERDICT: SHIP (confidence: high)\n\nLooks good.\n")

	var stdout strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true, DryRun: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, io.Discard, f); err != nil {
		t.Fatalf("runCloseWithReview: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("dry-run must not dispatch, got %d dispatch(es)", *calls)
	}
	got := stdout.String()
	if !strings.Contains(got, "codex exec") {
		t.Fatalf("close dry-run command missing codex exec:\n%s", got)
	}
	// #137: the dry-run prompt must carry the repo-derived issue ref — not "#0"
	// (the bug where the dry-run literal omitted IssueNum).
	if wantRef := repoIdentity() + "#69"; !strings.Contains(got, wantRef) {
		t.Errorf("dry-run command missing derived issue ref %q:\n%s", wantRef, got)
	}
	if strings.Contains(got, repoIdentity()+"#0") {
		t.Error("dry-run shows <repo>#0 — IssueNum not threaded into the dry-run orientation (#137)")
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
	// #136 D4: a skipped boundary writes NO sidecar (there is no review body to
	// persist; the trailer already records not-run).
	if _, err := os.Stat(filepath.Join("workshop/plans", "000069-x-close-review.md")); !os.IsNotExist(err) {
		t.Errorf("--no-judge must not write a review sidecar (stat err=%v)", err)
	}
}
