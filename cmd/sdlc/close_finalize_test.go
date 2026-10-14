package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
	project "github.com/xianxu/ariadne/cmd/sdlc/internal/project"
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
			judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
				close(started)
				<-releaseReview
				return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
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
			case <-time.After(hangGuard):
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
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		close(started)
		<-releaseReview
		return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
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
	case <-time.After(hangGuard):
		t.Fatal("timeout waiting for stale HEAD close command")
	}
	if got.err == nil || !strings.Contains(got.err.Error(), "boundary review stale") {
		t.Fatalf("close should return stale-review error, got %v", got.err)
	}
	// #194: a code delta still refuses — but by NAMING what the review did not cover,
	// not with a bare HEAD-identity comparison the operator cannot act on.
	if !strings.Contains(got.err.Error(), "concurrent #69 side change") {
		t.Fatalf("refusal must name the unreviewed commit, got: %v", got.err)
	}
	if strings.Contains(got.err.Error(), "HEAD changed from") {
		t.Fatalf("refusal should report commits, not bare HEAD identity: %v", got.err)
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
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		close(started)
		<-releaseReview
		return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
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
	case <-time.After(hangGuard):
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

func TestCloseCommand_PlanChangedDuringBoundaryReview_DoesNotFinalize(t *testing.T) {
	commands := map[string]func(string, string) []string{
		"close": func(issuesDir, plansDir string) []string {
			return []string{"close", "--issue", "69", "--actual", "1", "--verified", "tests pass",
				"--no-atlas", "--issues-dir", issuesDir, "--plans-dir", plansDir,
				"--brain-dir", "../nonexistent-brain"}
		},
		"milestone-close": func(issuesDir, plansDir string) []string {
			return []string{"milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1",
				"--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir,
				"--plans-dir", plansDir, "--brain-dir", "../nonexistent-brain"}
		},
	}
	mutations := map[string]func(*testing.T, string){
		"modified": func(t *testing.T, path string) {
			if err := os.WriteFile(path, []byte("# Plan\n\nmodified\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"created": func(t *testing.T, path string) {
			if err := os.WriteFile(path, []byte("# Plan\n\ncreated\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
		"deleted": func(t *testing.T, path string) {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		},
		"replaced": func(t *testing.T, path string) {
			replacement := path + ".replacement"
			if err := os.WriteFile(replacement, []byte("# Plan\n\nreplacement\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.Rename(replacement, path); err != nil {
				t.Fatal(err)
			}
		},
	}
	for commandName, args := range commands {
		for mutationName, mutate := range mutations {
			t.Run(commandName+"/"+mutationName, func(t *testing.T) {
				issuesDir := closeRepo(t, 69)
				plansDir := filepath.Join("workshop", "plans")
				if err := os.MkdirAll(plansDir, 0o755); err != nil {
					t.Fatal(err)
				}
				planPath := filepath.Join(plansDir, "000069-x-plan.md")
				if mutationName != "created" {
					if err := os.WriteFile(planPath, []byte("# Plan\n\noriginal\n"), 0o644); err != nil {
						t.Fatal(err)
					}
				}

				started := make(chan struct{})
				releaseReview := make(chan struct{})
				orig := judge.Run
				t.Cleanup(func() { judge.Run = orig })
				judge.Run = func(ctx context.Context, onStart func(pid int), agent string, args ...string) (judge.ProcessOutput, error) {
					close(started)
					<-releaseReview
					return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
				}

				done := make(chan error, 1)
				go func() {
					_, _, err := executeSDLCTestCommand(args(issuesDir, plansDir)...)
					done <- err
				}()

				waitForSignal(t, started, "boundary review to start")
				mutate(t, planPath)
				close(releaseReview)

				select {
				case err := <-done:
					if err == nil || !strings.Contains(err.Error(), "boundary review stale") {
						t.Fatalf("plan %s during %s must invalidate the review, got %v", mutationName, commandName, err)
					}
				case <-time.After(hangGuard):
					t.Fatal("timeout waiting for stale close command")
				}
				if text := readIssue(t, issuesDir); !strings.Contains(text, "status: working") {
					t.Fatalf("stale plan review finalized issue:\n%s", text)
				}
			})
		}
	}
}

// #160 Q4: the lessons reminder moved from the publish gate to `sdlc close` — a
// finalizing whole-issue close emits it (agent engaged, findings fresh); a
// non-finalizing (REWORK) close does not.
func TestRunCloseWithReview_EmitsLessonsReminder(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood"+emptyBoundaryFindings)
	var stdout strings.Builder
	if err := runCloseWithReview(&stdout, io.Discard, closeFlagsFor(issuesDir)); err != nil {
		t.Fatalf("SHIP close should finalize: %v", err)
	}
	if !strings.Contains(stdout.String(), judge.LessonsReminder) {
		t.Error("finalizing whole-issue close should emit the lessons reminder (#160 Q4)")
	}

	issuesDir2 := closeRepo(t, 69)
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nnope"+reworkBoundaryFindings)
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
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood"+emptyBoundaryFindings)
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
	stubJudge(t, "VERDICT: REWORK (confidence: high)\n\nNeeds rework.\n"+reworkBoundaryFindings)

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
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		return judge.ProcessOutput{}, errors.New("boom: agent not found")
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
// protocol refusal after one retry; it does not finalize.
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
	if !strings.Contains(stderr.String(), "protocol failed after two attempts") || !strings.Contains(stderr.String(), "missing verdict") {
		t.Errorf("halt should report the exhausted response-protocol retry:\n%s", stderr.String())
	}
}

// #139: after a REWORK, the issue is still `working`, so a rerun after fixing the
// findings finalizes cleanly with exactly one close line and NO --no-reclose-guard.
func TestRunCloseWithReview_RerunAfterREWORK(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	f := closeFlagsFor(issuesDir)

	stubJudge(t, "VERDICT: REWORK\n\nfix it"+reworkBoundaryFindings)
	if err := runCloseWithReview(io.Discard, io.Discard, f); err == nil {
		t.Fatal("first close (REWORK) should error")
	}
	// Rerun with SHIP — note f carries NO NoReclose flag.
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood"+addressedBoundaryFindings)
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
	stubJudge(t, "VERDICT: REWORK\n\nnot yet"+reworkBoundaryFindings)

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
	stubJudge(t, "VERDICT: SHIP (confidence: high)\n\ngood slice"+emptyBoundaryFindings)

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

// #194: the interleaving that motivated the issue. A DOC-ONLY commit landing during the
// ~20-minute review is a delta, not an invalidation — the reviewed code surface is
// untouched, so the close finalizes and says what it decided. This is the bookkeeping a
// close itself invites (lessons.md, atlas, plan ticks), which previously discarded the
// review that was already most of the way through.
func TestCloseCommand_DocOnlyCommitDuringBoundaryReview_Finalizes(t *testing.T) {
	issuesDir := closeRepo(t, 69)
	started := make(chan struct{})
	releaseReview := make(chan struct{})
	orig := judge.Run
	t.Cleanup(func() { judge.Run = orig })
	judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
		close(started)
		<-releaseReview
		return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
	}

	done := make(chan struct {
		stdout, stderr string
		err            error
	}, 1)
	go func() {
		stdout, stderr, err := executeSDLCTestCommand("close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas", "--issues-dir", issuesDir, "--brain-dir", "../nonexistent-brain")
		done <- struct {
			stdout, stderr string
			err            error
		}{stdout, stderr, err}
	}()

	waitForSignal(t, started, "boundary review to start")
	if err := os.MkdirAll("workshop", 0o755); err != nil {
		t.Fatalf("mkdir workshop: %v", err)
	}
	if err := os.WriteFile("workshop/lessons.md", []byte("- learned a thing\n"), 0o644); err != nil {
		t.Fatalf("write lessons: %v", err)
	}
	git(t, "", "add", "workshop/lessons.md")
	git(t, "", "commit", "-q", "-m", "#69: lessons from the review")
	close(releaseReview)

	var got struct {
		stdout, stderr string
		err            error
	}
	select {
	case got = <-done:
	case <-time.After(hangGuard):
		t.Fatal("timeout waiting for close command")
	}
	if got.err != nil {
		t.Fatalf("doc-only delta must finalize, got error: %v\nstderr:\n%s", got.err, got.stderr)
	}
	text := readIssue(t, issuesDir)
	for _, want := range []string{"status: codecomplete", "closed — tests pass"} {
		if !strings.Contains(text, want) {
			t.Fatalf("close did not finalize; missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(got.stderr, "doc-only commit(s) arrived since") {
		t.Fatalf("expected the anchored pass line on stderr:\n%s", got.stderr)
	}
}

// TestCloseCommands_TLOwnedProjectUntouched pins #326: close and milestone-close
// leave a `tl:` project byte-identical (no tick; milestone-close doesn't demand
// its detail block), print the skip line, and a TL edit during the review
// doesn't void SHIP, because the file isn't in the snapshot. A sibling project
// without `tl:` is still ticked, so the filter is per project.
func TestCloseCommands_TLOwnedProjectUntouched(t *testing.T) {
	commands := map[string][]string{
		"close":           {"close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas"},
		"milestone-close": {"milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1", "--verified", "tests pass", "--no-atlas"},
	}
	for name, args := range commands {
		for _, concurrentEdit := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/edit=%v", name, concurrentEdit), func(t *testing.T) {
				issuesDir := closeRepo(t, 69)
				brainDir := t.TempDir()
				repoRoot, _ := os.Getwd()
				projectDir := filepath.Join(repoRoot, "workshop", "projects")
				if err := os.MkdirAll(projectDir, 0o755); err != nil {
					t.Fatal(err)
				}
				ref := repoIdentity() + "#69"
				tlPath := filepath.Join(projectDir, "tl-run.md")
				tlText := "---\ntype: project\nname: tl-run\nstatus: executing\ntl: xian-xu@ops:0\n---\n# tl-run\n\n- [ ] ship [" + ref + "]\n- [ ] M1 [" + ref + " M1]\n"
				plainPath := filepath.Join(projectDir, "roadmap.md")
				// The plain project carries its M1 detail block, so no --no-project
				// waiver is needed: the TL project must not trip that gate itself.
				anchor := project.AnchorFor(repoIdentity(), "69", "M1")
				plainText := "# roadmap\n\n- [ ] ship [" + ref + "]\n- [ ] M1 [" + ref + " M1]\n\n## details\n\n<a id=\"" + anchor + "\"></a>\n### M1\n\nprose\n"
				for p, text := range map[string]string{tlPath: tlText, plainPath: plainText} {
					if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
						t.Fatal(err)
					}
				}

				started := make(chan struct{})
				releaseReview := make(chan struct{})
				orig := judge.Run
				t.Cleanup(func() { judge.Run = orig })
				judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
					close(started)
					<-releaseReview
					return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
				}
				done := make(chan error, 1)
				var stderr string
				go func() {
					var err error
					_, stderr, err = executeSDLCTestCommand(append(args, "--issues-dir", issuesDir, "--brain-dir", brainDir)...)
					done <- err
				}()
				waitForSignal(t, started, "boundary review to start")
				wantTL := tlText
				if concurrentEdit {
					wantTL = tlText + "\nTL ticked the row itself\n"
					if err := os.WriteFile(tlPath, []byte(wantTL), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				close(releaseReview)
				select {
				case err := <-done:
					if err != nil {
						t.Fatalf("%s refused: %v\n%s", name, err, stderr)
					}
				case <-time.After(hangGuard):
					t.Fatal("timeout waiting for close")
				}

				if got, _ := os.ReadFile(tlPath); string(got) != wantTL {
					t.Fatalf("TL-owned project changed:\n%s", got)
				}
				if !strings.Contains(stderr, "tl-run.md is TL-owned (tl: xian-xu@ops:0)") {
					t.Fatalf("missing skip line:\n%s", stderr)
				}
				if got, _ := os.ReadFile(plainPath); !strings.Contains(string(got), "- [x]") {
					t.Fatalf("project without tl: was not ticked:\n%s", got)
				}
				want := "closed — tests pass"
				if name == "milestone-close" {
					want = "closed M1 — tests pass"
				}
				if text := readIssue(t, issuesDir); !strings.Contains(text, want) {
					t.Fatalf("%s did not finalize; want %q:\n%s", name, want, text)
				}
			})
		}
	}
}

// #329: a project (without tl:) that carries the issue only as a train car has
// no row to tick, so close and milestone-close leave it byte-identical, demand
// no detail block for it, and say why.
func TestCloseCommands_TrainOnlyProjectUntouched(t *testing.T) {
	commands := map[string][]string{
		"close":           {"close", "--issue", "69", "--actual", "1", "--verified", "tests pass", "--no-atlas"},
		"milestone-close": {"milestone-close", "--issue", "69", "--milestone", "M1", "--actual", "1", "--verified", "tests pass", "--no-atlas"},
	}
	for name, args := range commands {
		t.Run(name, func(t *testing.T) {
			issuesDir := closeRepo(t, 69)
			brainDir := t.TempDir()
			repoRoot, _ := os.Getwd()
			projectDir := filepath.Join(repoRoot, "workshop", "projects")
			if err := os.MkdirAll(projectDir, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(projectDir, "trains-only.md")
			text := "---\ntype: project\nname: trains-only\nstatus: executing\n---\n# trains-only\n\n## Trains\n\n| Train | Cars | Slot |\n|---|---|---|\n| Core | " + repoIdentity() + "#69 | :1 |\n"
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				t.Fatal(err)
			}
			orig := judge.Run
			t.Cleanup(func() { judge.Run = orig })
			judge.Run = func(ctx context.Context, onStart func(pid int), name string, args ...string) (judge.ProcessOutput, error) {
				return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n\nLooks good.\n" + emptyBoundaryFindings)}, nil
			}
			_, stderr, err := executeSDLCTestCommand(append(args, "--issues-dir", issuesDir, "--brain-dir", brainDir)...)
			if err != nil {
				t.Fatalf("%s refused: %v\n%s", name, err, stderr)
			}
			if got, _ := os.ReadFile(path); string(got) != text {
				t.Fatalf("train-only project changed:\n%s", got)
			}
			if !strings.Contains(stderr, "trains-only.md carries") || !strings.Contains(stderr, "only as a train car (Core)") {
				t.Fatalf("missing train-only info line:\n%s", stderr)
			}
		})
	}
}
