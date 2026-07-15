package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
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

func TestCloseCommandsReleaseLockDuringBoundaryReview(t *testing.T) {
	cases := []struct {
		name string
		args func(string) []string
	}{
		{
			name: "close",
			args: func(issuesDir string) []string {
				return []string{"close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain"}
			},
		},
		{
			name: "milestone-close",
			args: func(issuesDir string) []string {
				return []string{"milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain"}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issuesDir := closeRepo(t, 69)
			lock := newObservedRepoLock()
			restore := stubRepoLockAcquire(t, lock.acquire)
			defer restore()

			started := make(chan struct{})
			releaseReview := make(chan struct{})
			orig := judge.Run
			t.Cleanup(func() { judge.Run = orig })
			judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
				close(started)
				<-releaseReview
				return []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n"), nil
			}

			done := make(chan error, 1)
			go func() {
				_, _, err := executeSDLCTestCommand(tc.args(issuesDir)...)
				done <- err
			}()

			waitForSignal(t, started, "boundary review to start")
			if held := lock.held(); held != 0 {
				close(releaseReview)
				t.Fatalf("repo lock held during %s boundary review: held=%d events=%v", tc.name, held, lock.events())
			}
			close(releaseReview)
			if err := waitForErr(t, done, tc.name+" command"); err != nil {
				t.Fatalf("%s command returned error: %v", tc.name, err)
			}
			if got := lock.acquireCount(); got < 2 {
				t.Fatalf("%s should acquire for compute and finalization, got %d events=%v", tc.name, got, lock.events())
			}
		})
	}
}

type observedRepoLock struct {
	mu       sync.Mutex
	heldNow  int
	acquired int
	eventLog []string
}

func newObservedRepoLock() *observedRepoLock {
	return &observedRepoLock{}
}

func (l *observedRepoLock) acquire(cmd *cobra.Command) (func() error, error) {
	l.mu.Lock()
	l.heldNow++
	l.acquired++
	l.eventLog = append(l.eventLog, "acquire "+cmd.CommandPath())
	l.mu.Unlock()
	return func() error {
		l.mu.Lock()
		defer l.mu.Unlock()
		l.heldNow--
		l.eventLog = append(l.eventLog, "release "+cmd.CommandPath())
		return nil
	}, nil
}

func (l *observedRepoLock) held() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.heldNow
}

func (l *observedRepoLock) acquireCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.acquired
}

func (l *observedRepoLock) events() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.eventLog...)
}

func waitForSignal(t *testing.T, ch <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", label)
	}
}

func waitForErr(t *testing.T, ch <-chan error, label string) error {
	t.Helper()
	select {
	case err := <-ch:
		return err
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for %s", label)
		return nil
	}
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

// #69 guard: a milestone passed to runCloseWithReview must be REFUSED (#146) — the
// no-review `close --milestone` path was removed, so runCloseWithReview redirects a
// milestone to `sdlc milestone-close` and dispatches nothing. The "exactly one
// review per boundary" invariant this used to guard is now structural —
// milestone-close computes + reviews via computeClose, never runCloseWithReview.
func TestRunCloseWithReview_MilestoneRefuses(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	calls, _ := stubJudge(t, "VERDICT: SHIP\n")

	f := &closeFlags{
		Issue: 69, Milestone: "M1", Actual: "1", Verified: "slice done", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	err := runCloseWithReview(io.Discard, io.Discard, f)
	if err == nil {
		t.Fatal("expected refusal for a milestone passed to runCloseWithReview")
	}
	if !strings.Contains(err.Error(), "milestone-close") {
		t.Errorf("refusal should redirect to milestone-close; got: %v", err)
	}
	if *calls != 0 {
		t.Fatalf("refused close must not dispatch a review, got %d dispatch(es)", *calls)
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

// rewriteIssuePlan replaces the fixture issue's body with the given ## Plan
// rows and commits the edit with a NEUTRAL subject (no `#69 Mx` anchor, so
// the rewrite itself never fakes milestone-close evidence). Shared by the
// #175 trailing/midstream gate tests.
func rewriteIssuePlan(t *testing.T, issuesDir, planRows string) {
	t.Helper()
	issuePath := filepath.Join(issuesDir, "000069-x.md")
	content := "---\nid: 000069\nstatus: working\nestimate_hours: 1\n---\n\n" +
		"# x\n\n## Spec\n\nThing.\n\n## Plan\n\n" + planRows + "\n\n## Log\n"
	if err := os.WriteFile(issuePath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", issuePath},
		{"commit", "-q", "-m", "#69: plan update"},
	} {
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
}

// #175: single-pass work with legacy Mx tags closes WITHOUT --no-verdict —
// the trailing misses are covered by the issue-close boundary review, which
// this close dispatches (window branch-point→HEAD).
func TestClose_TrailingUnclosedMilestones_AcceptedByCloseReview(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	rewriteIssuePlan(t, issuesDir, "- [x] **M1 — all the work**")
	calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nLooks good.\n")

	var stdout, stderr strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, &stderr, f); err != nil {
		t.Fatalf("runCloseWithReview should accept trailing Mx misses: %v\nstderr:\n%s", err, stderr.String())
	}
	if *calls != 1 {
		t.Fatalf("the issue-close review must actually run (the acceptance premise), got %d dispatches", *calls)
	}
	for _, want := range []string{"M1", "issue-close boundary review"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("acceptance info line missing %q:\n%s", want, stderr.String())
		}
	}
	// close finalizes: codecomplete (#160 — close never writes `done`) + the
	// verdict-annotated log line.
	got := readIssue(t, issuesDir)
	for _, want := range []string{"status: codecomplete", "review verdict: SHIP"} {
		if !strings.Contains(got, want) {
			t.Errorf("finalized issue missing %q:\n%s", want, got)
		}
	}
}

// #175: a midstream miss (M1 unreviewed while M2 closed WITH a trailer)
// still refuses, BEFORE the review dispatch, and the refusal cites the §3
// recovery — the boundary into M2 was genuinely crossed unreviewed.
func TestClose_MidstreamMissingVerdict_Refuses(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	rewriteIssuePlan(t, issuesDir, "- [x] **M1 — first**\n- [x] **M2 — second**")
	issuePath := filepath.Join(issuesDir, "000069-x.md")
	// M2's close commit carries the trailer; M1 never got one. The commit
	// must touch the issue file (the probe scopes to it) — an innocuous
	// whitespace append keeps the file's real content intact.
	fh, err := os.OpenFile(issuePath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fh.WriteString("\n"); err != nil {
		t.Fatal(err)
	}
	if err := fh.Close(); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", issuePath},
		{"commit", "-q", "-m", "#69 M2: close — tick milestone",
			"-m", "Body.\n\nReview-Verdict: SHIP\nReview-Window: abc1234..HEAD"},
	} {
		cmd := exec.Command("git", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v — %s", args, err, out)
		}
	}
	calls, _ := stubJudge(t, "VERDICT: SHIP\n")

	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	msg, died := expectDie(t, func() {
		_ = runCloseWithReview(io.Discard, io.Discard, f)
	})
	if !died {
		t.Fatal("midstream missing verdict should refuse the close")
	}
	for _, want := range []string{"M1", "AGENTS.md §3", "plain checkboxes"} {
		if !strings.Contains(msg, want) {
			t.Errorf("midstream refusal missing %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "milestones M1, M2") || strings.Contains(msg, "M2 lack") {
		t.Errorf("refusal must name only the midstream miss (M2 has evidence):\n%s", msg)
	}
	if *calls != 0 {
		t.Errorf("refusal must fire at the gate, before review dispatch; got %d dispatches", *calls)
	}
}

// #175: a trailing miss with --no-judge refuses — the issue-close review
// that justifies the acceptance is exactly what --no-judge skips.
func TestClose_TrailingMissingVerdict_NoJudgeRefuses(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	rewriteIssuePlan(t, issuesDir, "- [x] **M1 — all the work**")
	calls, _ := stubJudge(t, "VERDICT: SHIP\n")

	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true, NoJudge: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	msg, died := expectDie(t, func() {
		_ = runCloseWithReview(io.Discard, io.Discard, f)
	})
	if !died {
		t.Fatal("trailing miss + --no-judge should refuse the close")
	}
	for _, want := range []string{"M1", "--no-judge", "Or pass --no-verdict (or --force); record"} {
		if !strings.Contains(msg, want) {
			t.Errorf("needs-judge refusal missing %q:\n%s", want, msg)
		}
	}
	if *calls != 0 {
		t.Errorf("refusal must dispatch nothing; got %d", *calls)
	}
}

// #174 leg A: a FIX-THEN-SHIP close finalizes AND states the protocol —
// fix now, bundle into one commit, don't re-run close.
func TestClose_FixThenShip_EmitsProtocol(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: FIX-THEN-SHIP (confidence: high)\n\nFinding: nit.\n")

	var stdout, stderr strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, &stderr, f); err != nil {
		t.Fatalf("FIX-THEN-SHIP is finalizing; close should succeed: %v", err)
	}
	for _, want := range []string{"ONE commit", "Do NOT re-run"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("protocol block missing %q:\n%s", want, stderr.String())
		}
	}
	got := readIssue(t, issuesDir)
	for _, want := range []string{"status: codecomplete", "review verdict: FIX-THEN-SHIP"} {
		if !strings.Contains(got, want) {
			t.Errorf("finalized issue missing %q:\n%s", want, got)
		}
	}
}

// #174 leg A negative: a SHIP close emits NO protocol block — it is
// verdict-conditional.
func TestClose_Ship_NoProtocolBlock(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\nLooks good.\n")

	var stdout, stderr strings.Builder
	f := &closeFlags{
		Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain",
	}
	if err := runCloseWithReview(&stdout, &stderr, f); err != nil {
		t.Fatalf("runCloseWithReview: %v", err)
	}
	for _, forbidden := range []string{"FIX-THEN-SHIP", "ONE commit"} {
		if strings.Contains(stderr.String(), forbidden) {
			t.Errorf("SHIP close must not emit the protocol block (%q found):\n%s", forbidden, stderr.String())
		}
	}
}
