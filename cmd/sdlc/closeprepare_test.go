package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
