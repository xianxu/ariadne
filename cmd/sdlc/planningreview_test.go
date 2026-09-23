package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

type planningCLIResult struct {
	output string
	err    error
}

func startPlanningReviewCLI(t *testing.T, binary, pause string, prepare ...func(string)) (repo, barrier string, finish func() planningCLIResult) {
	t.Helper()
	repo, _ = syncRepo(t)
	text := "---\nid: 000206\nstatus: working\n---\n## Spec\nDesign\n## Done when\n- Safe\n## Plan\n- [ ] Implement\n## Estimate\nReview this estimate\n"
	writeSyncIssue(t, repo, filepath.Base(issuePath206), text)
	writeSyncIssue(t, repo, "000207-unrelated.md", "---\nid: 000207\nstatus: open\n---\nOther work\n")
	for _, before := range prepare {
		before(repo)
	}
	git(t, repo, "add", "--", syncIssuesDir)
	git(t, repo, "commit", "-m", "review inputs")
	barrier = t.TempDir()
	agentDir := barrier
	script := `#!/bin/sh
kind=plan
case "$*" in *"reviewing an issue's ## Estimate block"*) kind=estimate;; esac
if [ "$kind" = "$PLANNING_PAUSE" ]; then
  echo "$kind" > "$PLANNING_BARRIER/ready"
  n=0
  while [ ! -f "$PLANNING_BARRIER/release" ]; do
    n=$((n+1)); [ "$n" -lt 2000 ] || exit 1
    sleep 0.01
  done
fi
if [ -n "$PLANNING_REPLY_FILE" ]; then cat "$PLANNING_REPLY_FILE"; else
printf 'VERDICT: CLEAN\n\n\140\140\140findings\nfindings: []\n\140\140\140\n'
fi
`
	if err := os.WriteFile(filepath.Join(agentDir, "claude"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	args := []string{"change-code", "--issue", "206", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--agent", "claude", "--worktree", "no", "--force", "test cannot waive stale reviews"}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Dir = repo
	cmd.WaitDelay = 2 * time.Second
	cmd.Env = append(os.Environ(), "PATH="+agentDir+":"+os.Getenv("PATH"), "PLANNING_PAUSE="+pause, "PLANNING_BARRIER="+barrier)
	done := make(chan planningCLIResult, 1)
	go func() { out, err := cmd.CombinedOutput(); done <- planningCLIResult{string(out), err} }()
	finished := false
	t.Cleanup(func() {
		_ = os.WriteFile(filepath.Join(barrier, "release"), nil, 0644)
		cancel()
		if !finished {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("planning CLI did not reap")
			}
		}
	})
	deadline := time.Now().Add(15 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(barrier, "ready")); err == nil {
			break
		}
		select {
		case r := <-done:
			finished = true
			t.Fatalf("review did not pause: %v %s", r.err, r.output)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("review barrier timed out")
		}
		time.Sleep(10 * time.Millisecond)
	}
	finish = func() planningCLIResult {
		t.Helper()
		if err := os.WriteFile(filepath.Join(barrier, "release"), nil, 0644); err != nil {
			t.Fatal(err)
		}
		r := <-done
		finished = true
		return r
	}
	return
}

func TestPlanningReviewConcurrencySchedules(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	for _, kind := range []string{"plan", "estimate"} {
		t.Run(kind, func(t *testing.T) {
			repo, _, finish := startPlanningReviewCLI(t, binary, kind)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			other := exec.CommandContext(ctx, binary, "issue", "set-status", "working", "--issue", "207")
			other.Dir = repo
			out, err := other.CombinedOutput()
			if err != nil {
				t.Errorf("unrelated mutation could not finish during %s review: %v %s", kind, err, out)
			}
			r := finish()
			if r.err != nil {
				t.Fatalf("review failed after unrelated mutation: %v %s", r.err, r.output)
			}
		})
	}
}

func TestPlanningReviewStaleBeforePersistence(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	for _, kind := range []string{"plan", "estimate"} {
		for _, mutation := range []string{"issue", "plan-appears", "plan-edited", "plan-removed", "ledger", "branch", "head"} {
			t.Run(kind+"/"+mutation, func(t *testing.T) {
				repo, _, finish := startPlanningReviewCLI(t, binary, kind, func(repo string) {
					if mutation == "plan-edited" || mutation == "plan-removed" {
						p := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan.md")
						if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(p, []byte("original plan"), 0644); err != nil {
							t.Fatal(err)
						}
					}
				})
				ledger := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")
				beforeLedger, _ := os.ReadFile(ledger)
				switch mutation {
				case "issue":
					writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 000206\nstatus: working\n---\nchanged draft\n")
				case "plan-appears":
					p := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan.md")
					if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(p, []byte("new plan"), 0644); err != nil {
						t.Fatal(err)
					}
				case "plan-edited", "plan-removed":
					p := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan.md")
					var err error
					if mutation == "plan-removed" {
						err = os.Remove(p)
					} else {
						err = os.WriteFile(p, []byte("edited plan"), 0644)
					}
					if err != nil {
						t.Fatal(err)
					}
				case "ledger":
					if err := os.MkdirAll(filepath.Dir(ledger), 0755); err != nil {
						t.Fatal(err)
					}
					beforeLedger = []byte("peer ledger generation")
					if err := os.WriteFile(ledger, beforeLedger, 0644); err != nil {
						t.Fatal(err)
					}
				case "branch":
					git(t, repo, "switch", "-c", "different-branch")
				case "head":
					git(t, repo, "commit", "--allow-empty", "-m", "changed head")
				}
				head := git(t, repo, "rev-parse", "HEAD")
				branch := git(t, repo, "branch", "--show-current")
				r := finish()
				if r.err == nil {
					t.Errorf("stale %s review passed with force: %s", mutation, r.output)
				}
				after, _ := os.ReadFile(ledger)
				if string(after) != string(beforeLedger) {
					t.Errorf("stale review wrote ledger: %s", after)
				}
				if got := git(t, repo, "rev-parse", "HEAD"); got != head {
					t.Fatal("stale review committed")
				}
				if got := git(t, repo, "branch", "--show-current"); got != branch {
					t.Fatal("stale review created branch")
				}
				if !strings.Contains(r.output, "stale") {
					t.Errorf("missing stale diagnostic: %v %s", r.err, r.output)
				}
			})
		}
	}
}

func TestPlanningReviewInterruption(t *testing.T) {
	for _, failure := range []string{"cancel", "relock", "release"} {
		t.Run(failure, func(t *testing.T) {
			repo, _ := syncRepo(t)
			writeSyncIssue(t, repo, filepath.Base(issuePath206), "---\nid: 000206\nstatus: working\n---\n## Spec\nDesign\n## Plan\n- [ ] Implement\n")
			before, _ := os.ReadFile(filepath.Join(repo, issuePath206))
			head := git(t, repo, "rev-parse", "HEAD")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			previousAcquire, previousRun := repoLockAcquireForCommand, judge.Run
			t.Cleanup(func() { repoLockAcquireForCommand = previousAcquire; judge.Run = previousRun })
			acquires, releases, dispatches := 0, 0, 0
			held := false
			repoLockAcquireForCommand = func(*cobra.Command) (func() error, error) {
				acquires++
				if failure == "relock" && acquires == 2 {
					return nil, errors.New("injected relock failure")
				}
				held = true
				return func() error {
					releases++
					held = false
					if failure == "release" {
						return errors.New("injected release failure")
					}
					return nil
				}, nil
			}
			judge.Run = func(ctx context.Context, _ func(int), _ string, _ ...string) (judge.ProcessOutput, error) {
				dispatches++
				if held {
					t.Error("review ran while lock was held")
				}
				if failure == "cancel" {
					cancel()
					<-ctx.Done()
					return judge.ProcessOutput{}, ctx.Err()
				}
				return judge.ProcessOutput{Stdout: []byte(findingsReply("CLEAN", "findings: []\n"))}, nil
			}
			root := buildRoot()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{"change-code", "--issue", "206", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--worktree", "no", "--force", "cannot waive interruption"})
			if err := root.ExecuteContext(ctx); err == nil {
				t.Fatalf("%s bypassed interruption: %s", failure, out.String())
			}
			if failure == "release" && dispatches != 0 {
				t.Fatal("dispatched after failed unlock")
			}
			if held || releases == 0 {
				t.Fatal("command retained lock")
			}
			if got := git(t, repo, "rev-parse", "HEAD"); got != head {
				t.Fatal("interruption committed")
			}
			after, _ := os.ReadFile(filepath.Join(repo, issuePath206))
			if !bytes.Equal(before, after) {
				t.Fatal("interruption mutated issue")
			}
			if _, err := os.Stat(filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")); !os.IsNotExist(err) {
				t.Fatalf("interruption wrote ledger: %v", err)
			}
		})
	}
}

func TestPlanningReviewTimeoutCannotBeForced(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	t.Setenv("WF_REVIEW_TIMEOUT", "1s")
	repo, _, finish := startPlanningReviewCLI(t, binary, "plan")
	time.Sleep(1200 * time.Millisecond)
	r := finish()
	if r.err == nil || !strings.Contains(r.output, "deadline") {
		t.Fatalf("timeout passed with force: %v %s", r.err, r.output)
	}
	if _, err := os.Stat(filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")); !os.IsNotExist(err) {
		t.Fatalf("timeout wrote ledger: %v", err)
	}
}

func TestPlanningReviewRefreshKeepsReviewedDocuments(t *testing.T) {
	repo, _ := syncRepo(t)
	p := filepath.Join(repo, issuePath206)
	issueArtifact, err := captureReviewArtifact(p)
	if err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan.md")
	planArtifact, err := captureReviewArtifact(planPath)
	if err != nil {
		t.Fatal(err)
	}
	ledgerPath := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")
	tx := &planningReviewTransaction{cmd: &cobra.Command{}}
	if err := tx.prepare([]reviewArtifact{issueArtifact, planArtifact}, ledgerPath); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(ledgerPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ledgerPath, []byte("our accepted round"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := tx.refreshOwnLedger(); err != nil {
		t.Fatalf("our ledger refresh refused: %v", err)
	}
	if err := os.WriteFile(p, []byte("concurrent editor change"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := tx.refreshOwnLedger(); err == nil {
		t.Fatal("ledger refresh adopted editor changes to reviewed issue")
	}
	if err := os.WriteFile(p, []byte(issueArtifact.text), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte("new optional plan"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := tx.refreshOwnLedger(); err == nil {
		t.Fatal("ledger refresh adopted optional plan appearance")
	}
}

func TestPlanningReviewStaleEveryVerdict(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	for _, reply := range []string{findingsReply("FAILURE", "findings: []\n"), "VERDICT: FAILURE\ninvalid findings protocol\n", "unclassified output"} {
		t.Run(strings.SplitN(reply, "\n", 2)[0], func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "reply")
			if err := os.WriteFile(file, []byte(reply), 0644); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PLANNING_REPLY_FILE", file)
			repo, _, finish := startPlanningReviewCLI(t, binary, "plan")
			writeSyncIssue(t, repo, filepath.Base(issuePath206), "changed draft")
			r := finish()
			if r.err == nil || !strings.Contains(r.output, "stale") {
				t.Fatalf("stale response handled as verdict: %v %s", r.err, r.output)
			}
			if _, err := os.Stat(filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")); !os.IsNotExist(err) {
				t.Fatalf("stale response wrote ledger: %v", err)
			}
		})
	}
}

func TestPlanningReviewCompetingResults(t *testing.T) {
	binary := buildFleetE2EBinary(t)
	repo, first, finish := startPlanningReviewCLI(t, binary, "plan")
	second := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "change-code", "--issue", "206", "--flow", "full", "--no-structural", "--no-estimate", "--no-estimate-recon", "--agent", "claude", "--worktree", "no", "--force", "cannot waive stale")
	cmd.Dir = repo
	cmd.WaitDelay = 2 * time.Second
	cmd.Env = append(os.Environ(), "PATH="+first+":"+os.Getenv("PATH"), "PLANNING_PAUSE=plan", "PLANNING_BARRIER="+second)
	done := make(chan planningCLIResult, 1)
	go func() { out, err := cmd.CombinedOutput(); done <- planningCLIResult{string(out), err} }()
	completed := false
	defer func() {
		_ = os.WriteFile(filepath.Join(second, "release"), nil, 0644)
		cancel()
		if !completed {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("competing CLI did not reap")
			}
		}
	}()
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(second, "ready")); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("competing review did not start while first waited")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if r := finish(); r.err != nil {
		t.Fatalf("first review failed: %v %s", r.err, r.output)
	}
	ledger := filepath.Join(repo, "workshop/plans/000206-issue-sync-verb-plan-gate.md")
	firstLedger, err := os.ReadFile(ledger)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(second, "release"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	r := <-done
	completed = true
	if r.err == nil || !strings.Contains(r.output, "stale") {
		t.Fatalf("second response accepted: %v %s", r.err, r.output)
	}
	after, err := os.ReadFile(ledger)
	if err != nil || !bytes.Equal(firstLedger, after) {
		t.Fatalf("competing response changed ledger: %v %s", err, after)
	}
}
