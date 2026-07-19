package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestCloseCommands_IssueChangedDuringBoundaryReview_DoesNotFinalize(t *testing.T) {
	cases := []struct {
		name      string
		args      func(string) []string
		forbidden []string
		wantErr   string
		wantStays string
	}{
		{
			name: "close",
			args: func(issuesDir string) []string {
				return []string{"close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain"}
			},
			forbidden: []string{"status: codecomplete", "closed — tests pass", "actual_hours: 1"},
			wantErr:   "boundary review stale",
			wantStays: "status: working",
		},
		{
			name: "milestone-close",
			args: func(issuesDir string) []string {
				return []string{"milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain"}
			},
			forbidden: []string{"closed M1 — tests pass"},
			wantErr:   "boundary review stale",
			wantStays: "status: working",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			issuesDir := closeRepo(t, 69)
			started := make(chan struct{})
			releaseReview := make(chan struct{})
			orig := judge.Run
			t.Cleanup(func() { judge.Run = orig })
			judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
				close(started)
				<-releaseReview
				return []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n"), nil
			}

			done := make(chan struct {
				stdout string
				err    error
			}, 1)
			go func() {
				stdout, _, err := executeSDLCTestCommand(tc.args(issuesDir)...)
				done <- struct {
					stdout string
					err    error
				}{stdout: stdout, err: err}
			}()

			waitForSignal(t, started, "boundary review to start")
			issuePath := filepath.Join(issuesDir, "000069-x.md")
			f, err := os.OpenFile(issuePath, os.O_APPEND|os.O_WRONLY, 0)
			if err != nil {
				t.Fatalf("open issue for concurrent edit: %v", err)
			}
			if _, err := f.WriteString("\nconcurrent operator note\n"); err != nil {
				_ = f.Close()
				t.Fatalf("write concurrent edit: %v", err)
			}
			if err := f.Close(); err != nil {
				t.Fatalf("close concurrent edit: %v", err)
			}
			close(releaseReview)

			var got struct {
				stdout string
				err    error
			}
			select {
			case got = <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for stale close command")
			}
			if got.err == nil || !strings.Contains(got.err.Error(), tc.wantErr) {
				t.Fatalf("%s should return stale-review error, got %v", tc.name, got.err)
			}
			if !strings.Contains(got.stdout, "Review-Verdict: SHIP") {
				t.Fatalf("%s should emit review trailer without finalizing:\n%s", tc.name, got.stdout)
			}
			text := readIssue(t, issuesDir)
			if !strings.Contains(text, tc.wantStays) {
				t.Fatalf("%s should leave issue working:\n%s", tc.name, text)
			}
			if !strings.Contains(text, "concurrent operator note") {
				t.Fatalf("%s should preserve concurrent edit:\n%s", tc.name, text)
			}
			for _, forbidden := range tc.forbidden {
				if strings.Contains(text, forbidden) {
					t.Fatalf("%s finalized stale state; found %q:\n%s", tc.name, forbidden, text)
				}
			}
		})
	}
}

func TestCloseCommand_HEADChangedDuringBoundaryReview_DoesNotFinalize(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	started := make(chan struct{})
	releaseReview := make(chan struct{})
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		close(started)
		<-releaseReview
		return []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n"), nil
	}

	done := make(chan struct {
		stdout string
		err    error
	}, 1)
	go func() {
		stdout, _, err := executeSDLCTestCommand("close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain")
		done <- struct {
			stdout string
			err    error
		}{stdout: stdout, err: err}
	}()

	waitForSignal(t, started, "boundary review to start")
	if err := os.WriteFile("other.txt", []byte("new head\n"), 0o644); err != nil {
		t.Fatalf("write concurrent file: %v", err)
	}
	git(t, "", "add", "other.txt")
	git(t, "", "commit", "-q", "-m", "concurrent #69 side change")
	close(releaseReview)

	var got struct {
		stdout string
		err    error
	}
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for stale HEAD close command")
	}
	if got.err == nil || !strings.Contains(got.err.Error(), "boundary review stale") {
		t.Fatalf("close should return stale-review error, got %v", got.err)
	}
	if !strings.Contains(got.stdout, "Review-Verdict: SHIP") {
		t.Fatalf("close should emit review trailer without finalizing:\n%s", got.stdout)
	}
	text := readIssue(t, issuesDir)
	for _, forbidden := range []string{"status: codecomplete", "closed — tests pass", "actual_hours: 1"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("close finalized stale HEAD; found %q:\n%s", forbidden, text)
		}
	}
}

func TestCloseCommand_ProjectChangedDuringBoundaryReview_DoesNotFinalize(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	brainDir := t.TempDir()
	// The project lives in the repo's own workshop/projects (the fleet home
	// DiscoverByIssueRef scans — the repo is a sibling under its own parent),
	// not brain/data/project. brainDir remains only for the calibration ledger.
	repoRoot, _ := os.Getwd()
	projectDir := filepath.Join(repoRoot, "workshop", "projects")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir project dir: %v", err)
	}
	projectPath := filepath.Join(projectDir, "roadmap.md")
	projectText := "# roadmap\n\n- [ ] ship close fix [" + repoIdentity() + "#69]\n"
	if err := os.WriteFile(projectPath, []byte(projectText), 0o644); err != nil {
		t.Fatalf("write project file: %v", err)
	}

	started := make(chan struct{})
	releaseReview := make(chan struct{})
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		close(started)
		<-releaseReview
		return []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n"), nil
	}

	done := make(chan struct {
		stdout string
		err    error
	}, 1)
	go func() {
		stdout, _, err := executeSDLCTestCommand("close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", brainDir)
		done <- struct {
			stdout string
			err    error
		}{stdout: stdout, err: err}
	}()

	waitForSignal(t, started, "boundary review to start")
	concurrentText := projectText + "\noperator updated project scope\n"
	if err := os.WriteFile(projectPath, []byte(concurrentText), 0o644); err != nil {
		t.Fatalf("write concurrent project edit: %v", err)
	}
	close(releaseReview)

	var got struct {
		stdout string
		err    error
	}
	select {
	case got = <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for stale project close command")
	}
	if got.err == nil || !strings.Contains(got.err.Error(), "boundary review stale") {
		t.Fatalf("close should return stale-review error, got %v", got.err)
	}
	if !strings.Contains(got.stdout, "Review-Verdict: SHIP") {
		t.Fatalf("close should emit review trailer without finalizing:\n%s", got.stdout)
	}
	gotProject, err := os.ReadFile(projectPath)
	if err != nil {
		t.Fatalf("read project file: %v", err)
	}
	if string(gotProject) != concurrentText {
		t.Fatalf("close overwrote concurrent project edit:\n%s", gotProject)
	}
	text := readIssue(t, issuesDir)
	for _, forbidden := range []string{"status: codecomplete", "closed — tests pass", "actual_hours: 1"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("close finalized stale project state; found %q:\n%s", forbidden, text)
		}
	}
}

// #160 Q4: the lessons reminder moved from the publish gate to `sdlc close` — a
// finalizing whole-issue close emits it (agent engaged, findings fresh); a
// non-finalizing (REWORK) close does not.
func TestRunCloseWithReview_EmitsLessonsReminder(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood")
	var stdout strings.Builder
	if err := runCloseWithReview(&stdout, io.Discard, closeFlagsFor(issuesDir)); err != nil {
		t.Fatalf("SHIP close should finalize: %v", err)
	}
	if !strings.Contains(stdout.String(), judge.LessonsReminder) {
		t.Error("finalizing whole-issue close should emit the lessons reminder (#160 Q4)")
	}

	issuesDir2 := closeRepo(t, 69)
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nnope")
	var stdout2 strings.Builder
	_ = runCloseWithReview(&stdout2, io.Discard, closeFlagsFor(issuesDir2))
	if strings.Contains(stdout2.String(), judge.LessonsReminder) {
		t.Error("a non-finalizing (REWORK) close must NOT emit the lessons reminder")
	}
}

// #171: a whole-issue close updates EVERY project across the fleet that
// references the issue — multiple matches are legitimate membership, not
// ambiguity (the old FindByIssueRef refused on >1; DiscoverByIssueRef ticks all).
func TestRunClose_UpdatesAllMatchingProjects(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood")
	repoRoot, _ := os.Getwd()
	projectsDir := filepath.Join(repoRoot, "workshop", "projects")
	if err := os.MkdirAll(projectsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := repoIdentity()
	for _, name := range []string{"alpha", "beta"} {
		body := "# " + name + "\n\n- [ ] ship it [" + id + "#69]\n"
		if err := os.WriteFile(filepath.Join(projectsDir, name+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := runCloseWithReview(io.Discard, io.Discard, closeFlagsFor(issuesDir)); err != nil {
		t.Fatalf("SHIP close should finalize: %v", err)
	}
	for _, name := range []string{"alpha", "beta"} {
		data, err := os.ReadFile(filepath.Join(projectsDir, name+".md"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "- [x] ship it ["+id+"#69]") {
			t.Errorf("project %s.md was not ticked by the all-match close:\n%s", name, data)
		}
	}
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
	if strings.Contains(got, "status: codecomplete") {
		t.Error("REWORK must NOT flip the issue to status: codecomplete")
	}
	if strings.Contains(got, "closed —") {
		t.Error("REWORK must NOT append a closed log line")
	}
	if strings.Contains(got, "actual_hours: 1") {
		t.Error("REWORK must NOT write actual_hours")
	}
	if strings.Contains(stderr.String(), "flipped") {
		t.Errorf("REWORK must NOT print 'flipped → codecomplete':\n%s", stderr.String())
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
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) ([]byte, error) {
		return nil, errors.New("boom: agent not found")
	}

	var stderr strings.Builder
	err := runCloseWithReview(io.Discard, &stderr, closeFlagsFor(issuesDir))
	if err == nil {
		t.Fatal("a judge dispatch error must halt (non-nil error)")
	}
	got := readIssue(t, issuesDir)
	if strings.Contains(got, "status: codecomplete") {
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
	if strings.Contains(readIssue(t, issuesDir), "status: codecomplete") {
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
	if !strings.Contains(got, "status: codecomplete") {
		t.Error("rerun should finalize → codecomplete (#160)")
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
	var stdout strings.Builder
	if err := runMilestoneClose(&stdout, io.Discard, f); err != nil {
		t.Fatalf("milestone SHIP should finalize, got: %v", err)
	}
	if got := readIssue(t, issuesDir); !strings.Contains(got, "closed M1 — tests pass; review verdict: SHIP") {
		t.Errorf("milestone SHIP should write + annotate the closed-M1 line:\n%s", got)
	}
	// #160 Q4: the lessons ping fires ONLY at the whole-issue close boundary, never
	// at milestone-close — the `f.Milestone == ""` guard in reviewThenFinalize is
	// the only thing enforcing that, so pin it (M2 boundary-review Important #1).
	if strings.Contains(stdout.String(), judge.LessonsReminder) {
		t.Error("milestone-close must NOT emit the lessons reminder (Q4 — whole-issue close only)")
	}
}
