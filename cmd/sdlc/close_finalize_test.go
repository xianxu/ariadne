package main

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// TestCloseVerdictOutcome pins the finalize policy (#139), derived from the #147
// verdict model: finalizing → finalize, blocking → rework, everything else → halt.
func TestCloseVerdictOutcome(t *testing.T) {
	cases := []struct {
		v    judge.Verdict
		want closeOutcome
	}{
		{judge.VerdictShip, closeFinalize},
		{judge.VerdictFixThenShip, closeFinalize},
		{judge.VerdictRework, closeRework},
		{judge.VerdictUnknown, closeHalt},
		{judge.VerdictNotRun, closeHalt},
	}
	for _, c := range cases {
		if got := closeVerdictOutcome(c.v); got != c.want {
			t.Errorf("closeVerdictOutcome(%q) = %v, want %v", c.v, got, c.want)
		}
	}
}

func closeFlagsFor(issuesDir string) *closeFlags {
	return &closeFlags{Issue: 69, Actual: "1", Verified: "tests pass", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
}

// #139: a REWORK boundary review must NOT finalize — the issue stays `working`,
// no close log line, no actual_hours, a non-nil error, and no "flipped → done".
func TestRunCloseWithReview_REWORK_DoesNotFinalize(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nNeeds rework.\n")

	var stderr strings.Builder
	err := runCloseWithReview(io.Discard, &stderr, closeFlagsFor(issuesDir))
	if err == nil {
		t.Fatal("REWORK must return a non-nil error (close not finalized)")
	}
	got := readIssue(t, issuesDir)
	if strings.Contains(got, "status: done") {
		t.Error("REWORK must NOT flip the issue to status: done")
	}
	if strings.Contains(got, "closed —") {
		t.Error("REWORK must NOT append a closed log line")
	}
	if strings.Contains(got, "actual_hours: 1") {
		t.Error("REWORK must NOT write actual_hours")
	}
	if strings.Contains(stderr.String(), "flipped") {
		t.Errorf("REWORK must NOT print 'flipped → done':\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "REWORK") {
		t.Error("REWORK should tell the operator to fix + re-run")
	}
}

// #139 I3: a judge DISPATCH ERROR (Run returns err, not just unparseable output)
// halts — close is not finalized, issue stays working, and there is no false
// "close succeeded" message (the pre-#139 line, now removed).
func TestRunCloseWithReview_DispatchError_Halts(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, errors.New("boom: agent not found")
	}

	var stderr strings.Builder
	err := runCloseWithReview(io.Discard, &stderr, closeFlagsFor(issuesDir))
	if err == nil {
		t.Fatal("a judge dispatch error must halt (non-nil error)")
	}
	got := readIssue(t, issuesDir)
	if strings.Contains(got, "status: done") {
		t.Error("dispatch error must NOT finalize the close")
	}
	if strings.Contains(got, "closed —") {
		t.Error("dispatch error must NOT append a closed log line")
	}
	if strings.Contains(stderr.String(), "close succeeded") {
		t.Errorf("dispatch error must not claim 'close succeeded':\n%s", stderr.String())
	}
}

// #139: an unexpected verdict (unknown — no schema-valid verdict) halts for a
// human; it does not finalize.
func TestRunCloseWithReview_Unknown_Halts(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "I reviewed the diff and have some thoughts, but no clear call.\n")

	var stderr strings.Builder
	err := runCloseWithReview(io.Discard, &stderr, closeFlagsFor(issuesDir))
	if err == nil {
		t.Fatal("an unknown verdict must return a non-nil error (halt)")
	}
	if strings.Contains(readIssue(t, issuesDir), "status: done") {
		t.Error("an unknown verdict must NOT finalize the close")
	}
	if !strings.Contains(stderr.String(), "UNEXPECTED") || !strings.Contains(stderr.String(), "consult a human") {
		t.Errorf("halt should tell the operator to stop + consult a human:\n%s", stderr.String())
	}
}

// #139: after a REWORK, the issue is still `working`, so a rerun after fixing the
// findings finalizes cleanly with exactly one close line and NO --no-reclose-guard.
func TestRunCloseWithReview_RerunAfterREWORK(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	f := closeFlagsFor(issuesDir)

	stubJudge(t, "VERDICT: REWORK\n\nfix it")
	if err := runCloseWithReview(io.Discard, io.Discard, f); err == nil {
		t.Fatal("first close (REWORK) should error")
	}
	// Rerun with SHIP — note f carries NO NoReclose flag.
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood")
	if err := runCloseWithReview(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("rerun (SHIP) should finalize cleanly (no --no-reclose-guard), got: %v", err)
	}
	got := readIssue(t, issuesDir)
	if !strings.Contains(got, "status: done") {
		t.Error("rerun should finalize → done")
	}
	if n := strings.Count(got, "closed — tests pass"); n != 1 {
		t.Errorf("expected exactly one closed log line, got %d:\n%s", n, got)
	}
	if !strings.Contains(got, "review verdict: SHIP") {
		t.Error("finalized close should annotate the verdict")
	}
}

// #139: milestone-close folds into the same two-phase — a REWORK milestone review
// leaves the milestone unwritten (no "closed M1" log line), non-nil error.
func TestRunMilestoneClose_REWORK_DoesNotFinalize(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: REWORK\n\nnot yet")

	f := &milestoneCloseFlags{Issue: 69, Milestone: "M1", Actual: "1", Verified: "tests pass",
		NoAtlas: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
	if err := runMilestoneClose(io.Discard, io.Discard, f); err == nil {
		t.Fatal("milestone REWORK must return a non-nil error")
	}
	if strings.Contains(readIssue(t, issuesDir), "closed M1") {
		t.Error("milestone REWORK must NOT append a 'closed M1' log line")
	}
}

// #139: a milestone SHIP finalizes (writes + annotates the closed-M1 line).
func TestRunMilestoneClose_SHIP_Finalizes(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood slice")

	f := &milestoneCloseFlags{Issue: 69, Milestone: "M1", Actual: "1", Verified: "tests pass",
		NoAtlas: true, IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
	if err := runMilestoneClose(io.Discard, io.Discard, f); err != nil {
		t.Fatalf("milestone SHIP should finalize, got: %v", err)
	}
	if got := readIssue(t, issuesDir); !strings.Contains(got, "closed M1 — tests pass; review verdict: SHIP") {
		t.Errorf("milestone SHIP should write + annotate the closed-M1 line:\n%s", got)
	}
}
