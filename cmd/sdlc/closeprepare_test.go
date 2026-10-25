package main

import (
	"context"
	"fmt"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// The incident happened before reviewer dispatch: measuring transcripts while
// holding the repository lock serializes every worktree's bookkeeping.
func TestClosePreparationMeasuresWithoutRepoLock(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		t.Run(verb, func(t *testing.T) {
			issues := closeRepo(t, 69)
			lock := newObservedRepoLock()
			restore := stubRepoLockAcquire(t, lock.acquire)
			defer restore()
			original := computeActualForCloseFn
			t.Cleanup(func() { computeActualForCloseFn = original })
			calls := 0
			computeActualForCloseFn = func(context.Context, string, string, string, ...string) actualResult {
				calls++
				if held := lock.held(); held != 0 {
					t.Errorf("measurement monopolizes repository lock: held=%d events=%v", held, lock.events())
				}
				return actualResult{Status: actualMeasured, Hours: 1, Window: "private fixture"}
			}
			stubJudge(t, "VERDICT: SHIP (confidence: high)\n"+emptyBoundaryFindings)
			args := []string{verb, "--issue", "69", "--verified", "private preparation fixture", "--no-atlas", "--no-project", "--issues-dir", issues, "--brain-dir", "../nonexistent-brain"}
			if verb == "milestone-close" {
				args = append(args, "--milestone", "M1")
			}
			if _, _, err := executeSDLCTestCommand(args...); err != nil {
				t.Fatalf("%s: %v", verb, err)
			}
			if calls != 1 {
				t.Fatalf("measurement ran %d times, want one", calls)
			}
		})
	}
}

func TestCloseReviewRefusesChangedPublicationTarget(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		t.Run(verb, func(t *testing.T) {
			plan := "- [x] implement\n"
			if verb == "milestone-close" {
				plan = "- [ ] M1 — implement\n"
			}
			r, cardPath, detailPath := closeReadyWithPlan(t, 617, plan)
			mirror := filepath.Join(t.TempDir(), "mirror.git")
			r.git("clone", "--bare", "-q", r.origin, mirror)
			before := r.card(cardPath)
			original := judge.Run
			t.Cleanup(func() { judge.Run = original })
			judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
				r.git("remote", "set-url", "origin", mirror)
				return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n" + emptyBoundaryFindings)}, nil
			}
			args := []string{verb, "--issue", "617", "--no-actual", "--no-atlas", "--no-project", "--verified", "private destination mutation"}
			if verb == "milestone-close" {
				args = append(args, "--milestone", "M1")
			}
			_, stderr, err := executeSDLCTestCommand(args...)
			if err == nil {
				t.Fatalf("changed publication target finalized: %s", stderr)
			}
			r.git("remote", "set-url", "origin", r.origin)
			if got := r.card(cardPath); got != before {
				t.Fatal("original card overwritten")
			}
			if raw, err := os.ReadFile(filepath.Join(r.root, detailPath)); err != nil || strings.Contains(string(raw), "closed —") {
				t.Fatalf("stale target wrote close: %q %v", raw, err)
			}
		})
	}
}

func TestClosePreparationPreservesChangedInputs(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		for _, changed := range []string{"issue", "plan", "head"} {
			t.Run(verb+"/"+changed, func(t *testing.T) {
				issues := closeRepo(t, 69)
				plans := "workshop/plans"
				if err := os.MkdirAll(plans, 0755); err != nil {
					t.Fatal(err)
				}
				plan := filepath.Join(plans, "000069-x-plan.md")
				if err := os.WriteFile(plan, []byte("Original plan\n"), 0644); err != nil {
					t.Fatal(err)
				}
				issue := filepath.Join(issues, "000069-x.md")
				original := computeActualForCloseFn
				t.Cleanup(func() { computeActualForCloseFn = original })
				computeActualForCloseFn = func(context.Context, string, string, string, ...string) actualResult {
					switch changed {
					case "head":
						git(t, "", "commit", "--allow-empty", "-qm", "concurrent work")
					case "plan":
						if err := os.WriteFile(plan, []byte("Newer plan\n"), 0644); err != nil {
							t.Fatal(err)
						}
					case "issue":
						f, err := os.OpenFile(issue, os.O_APPEND|os.O_WRONLY, 0644)
						if err != nil {
							t.Fatal(err)
						}
						_, err = f.WriteString("\nNewer issue\n")
						_ = f.Close()
						if err != nil {
							t.Fatal(err)
						}
					}
					return actualResult{Status: actualMeasured, Hours: 1, Window: "private preparation"}
				}
				calls, _ := stubJudge(t, "VERDICT: SHIP (confidence: high)\n"+emptyBoundaryFindings)
				args := []string{verb, "--issue", "69", "--verified", "stale fixture", "--no-atlas", "--no-project", "--plans-dir", plans, "--issues-dir", issues, "--brain-dir", "../nonexistent-brain"}
				if verb == "milestone-close" {
					args = append(args, "--milestone", "M1")
				}
				_, stderr, err := executeSDLCTestCommand(args...)
				if err == nil || *calls != 0 {
					t.Fatalf("stale %s admitted: err=%v calls=%d\n%s", changed, err, *calls, stderr)
				}
				if changed != "head" {
					path := issue
					if changed == "plan" {
						path = plan
					}
					data, err := os.ReadFile(path)
					if err != nil || !strings.Contains(string(data), fmt.Sprintf("Newer %s", changed)) {
						t.Fatalf("newer %s lost: %q %v", changed, data, err)
					}
				}
			})
		}
	}
}

func TestCloseBranchPublicationDoesNotHoldRepoLock(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		t.Run(verb, func(t *testing.T) {
			plan := "- [x] do it\n"
			if verb == "milestone-close" {
				plan = "- [ ] M1 — do it\n"
			}
			r, _, _ := closeReadyWithPlan(t, 617, plan)
			common := r.git("rev-parse", "--git-common-dir")
			if !filepath.IsAbs(common) {
				common = filepath.Join(r.root, common)
			}
			marker := filepath.Join(t.TempDir(), "branch-push")
			hooks := t.TempDir()
			r.git("config", "core.hooksPath", hooks)
			// Only issue-branch publication is optional. The tracker's own
			// conditional publication remains an authoritative transaction.
			hook := "#!/bin/sh\nwhile read -r localref localsha remoteref remotesha; do\ncase \"$remoteref\" in *000617*)\nif [ -d '" + filepath.Join(common, "sdlc.lock") + "' ]; then echo locked > '" + marker + "'; else echo unlocked > '" + marker + "'; fi;;\nesac\ndone\n"
			if err := os.WriteFile(filepath.Join(hooks, "pre-push"), []byte(hook), 0755); err != nil {
				t.Fatal(err)
			}
			stubJudge(t, "VERDICT: SHIP (confidence: high)\n"+emptyBoundaryFindings)
			args := []string{verb, "--issue", "617", "--verified", "private branch push fixture", "--no-actual", "--no-atlas", "--no-project"}
			if verb == "milestone-close" {
				args = append(args, "--milestone", "M1")
			}
			_, stderr, err := executeSDLCTestCommand(args...)
			if err != nil {
				t.Fatalf("close: %v\n%s", err, stderr)
			}
			state, err := os.ReadFile(marker)
			if err != nil || strings.TrimSpace(string(state)) != "unlocked" {
				t.Fatalf("issue branch push lock state %q: %v\n%s", state, err, stderr)
			}
		})
	}
}

// A deferred publication owns a concrete commit, not whichever tip exists
// when its warning-only network phase eventually gets CPU.
func TestCloseBranchPublicationPinsFinalizedHead(t *testing.T) {
	r, _, _ := startedHere(t)
	original := remoteTip(t, r, s09Branch)
	r.git("commit", "--allow-empty", "-qm", "finalized close")
	pinned := r.git("rev-parse", "HEAD")
	f := &closeFlags{effects: &closeEffects{}}
	var diagnostics strings.Builder
	pushCloseBranch(f, boundaryEnv(t), &diagnostics, "fixture")
	if got := remoteTip(t, r, s09Branch); got != original {
		t.Fatalf("publication ran before the unlocked tail: %s", got)
	}
	r.git("commit", "--allow-empty", "-qm", "later unrelated work")
	f.effects.finish()
	if got := remoteTip(t, r, s09Branch); got != pinned {
		t.Fatalf("published %s, want finalized %s: %s", got, pinned, &diagnostics)
	}
}

func TestCloseBranchPublicationPreservesNewerAuthority(t *testing.T) {
	for _, changed := range []string{"remote", "tracking"} {
		t.Run(changed, func(t *testing.T) {
			r, _, _ := startedHere(t)
			r.git("commit", "--allow-empty", "-qm", "finalized close")
			pinned := r.git("rev-parse", "HEAD")
			f := &closeFlags{effects: &closeEffects{}}
			var diagnostics strings.Builder
			pushCloseBranch(f, boundaryEnv(t), &diagnostics, "fixture")
			r.git("commit", "--allow-empty", "-qm", "newer actor work")
			newer := r.git("rev-parse", "HEAD")
			if changed == "remote" {
				r.git("push", "-q", "origin", "refs/heads/"+s09Branch)
			} else {
				r.git("update-ref", "refs/remotes/origin/"+s09Branch, newer)
			}
			f.effects.finish()
			want := pinned
			if changed == "remote" {
				want = newer
			}
			if got := remoteTip(t, r, s09Branch); got != want {
				t.Fatalf("remote=%s, want %s; %s", got, want, &diagnostics)
			}
			if got := r.git("rev-parse", "refs/remotes/origin/"+s09Branch); got != newer {
				t.Fatalf("newer tracking metadata overwritten: %s, want %s; %s", got, newer, &diagnostics)
			}
			message := "was not pushed"
			if changed == "tracking" {
				message = "push completed; tracking metadata was not updated"
			}
			if !strings.Contains(diagnostics.String(), message) {
				t.Fatalf("expected visible recovery %q: %s", message, &diagnostics)
			}
		})
	}
}

// BR-1: helper-level comparison tests cannot prove that the commands carry
// measurement authority into BOTH admission and post-review finalization.
func TestCloseCommandsRefuseChangedAdoptions(t *testing.T) {
	for _, verb := range []string{"close", "milestone-close"} {
		for _, phase := range []string{"preparation", "review"} {
			for _, change := range []string{"new", "changed"} {
				t.Run(verb+"/"+phase+"/"+change, func(t *testing.T) {
					plan := "- [x] implement\n"
					if verb == "milestone-close" {
						plan = "- [ ] M1 — implement\n"
					}
					r, ownCard, ownDetail := closeReadyWithPlan(t, 617, plan)
					otherPath, otherCard, _, _ := seededIssue(t, "000649", "competing-adoption")
					// Use the existing card setter for either initial state; it also handles
					// scaffold formats whose empty hours spelling differs.
					setOther := func(hours string) {
						t.Helper()
						if err := r.tracker.Update(otherPath, "private concurrent adoption", func(raw []byte) ([]byte, error) {
							if len(raw) == 0 {
								raw = []byte(otherCard)
							}
							return issue.SetCardField(raw, "actual_hours", hours)
						}); err != nil {
							t.Fatal(err)
						}
					}
					if change == "changed" {
						setOther("0.25")
					} else {
						setOther("N/A")
					}
					beforeCard := r.card(ownCard)
					beforeDetail, err := os.ReadFile(filepath.Join(r.root, ownDetail))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.MkdirAll(filepath.Join(r.root, "workshop/plans"), 0755); err != nil {
						t.Fatal(err)
					}
					beforePlans := reviewPlanFiles(t, filepath.Join(r.root, "workshop/plans"))
					beforeHead := r.git("rev-parse", "HEAD")
					engine := computeActualForCloseFn
					t.Cleanup(func() { computeActualForCloseFn = engine })
					computeActualForCloseFn = func(ctx context.Context, _, _, _ string, _ ...string) actualResult {
						records, err := loadIssueRecordsAt(withIssueRecordsScope(ctx), r.root, filepath.Join(r.root, "workshop/issues"), tracker.Fresh)
						if err != nil {
							t.Fatal(err)
						}
						since := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
						candidates, err := adoptedExclusionCandidates(records.All(), "617", since, since.Add(time.Hour))
						if err != nil {
							t.Fatal(err)
						}
						saved := captureAdoptionReadSet("617", filepath.Join(r.root, "workshop/issues"), since, since.Add(time.Hour), candidates)
						if phase == "preparation" {
							setOther("0.75")
						}
						return actualResult{Status: actualMeasured, Hours: 1, Window: "private authority schedule", Adoptions: saved}
					}
					originalJudge := judge.Run
					t.Cleanup(func() { judge.Run = originalJudge })
					calls := 0
					judge.Run = func(context.Context, func(int), string, ...string) (judge.ProcessOutput, error) {
						calls++
						if phase == "review" {
							setOther("0.75")
						}
						return judge.ProcessOutput{Stdout: []byte("VERDICT: SHIP (confidence: high)\n" + emptyBoundaryFindings)}, nil
					}
					args := []string{verb, "--issue", "617", "--verified", "private adoption schedule", "--no-atlas", "--no-project", "--brain-dir", "../nonexistent-brain"}
					if verb == "milestone-close" {
						args = append(args, "--milestone", "M1")
					}
					_, stderr, err := executeSDLCTestCommand(args...)
					if err == nil || !strings.Contains(err.Error(), "649") || !strings.Contains(err.Error(), "stale") {
						t.Fatalf("changed adoption finalized or wrong refusal: %v\n%s", err, stderr)
					}
					wantCalls := 0
					if phase == "review" {
						wantCalls = 1
					}
					if calls != wantCalls {
						t.Fatalf("review calls=%d, want%d", calls, wantCalls)
					}
					afterDetail, err := os.ReadFile(filepath.Join(r.root, ownDetail))
					if err != nil || string(afterDetail) != string(beforeDetail) {
						t.Fatalf("stale close changed issue: %v", err)
					}
					if r.card(ownCard) != beforeCard || r.git("rev-parse", "HEAD") != beforeHead {
						t.Fatal("stale close published card or evidence")
					}
					if after := reviewPlanFiles(t, filepath.Join(r.root, "workshop/plans")); !reflect.DeepEqual(after, beforePlans) {
						t.Fatalf("stale close wrote sidecar/ledger: before%v after%v", beforePlans, after)
					}
				})
			}
		}
	}
}
