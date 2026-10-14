package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestWaitCIGate(t *testing.T) {
	for _, mode := range []string{"pending-green", "unregistered-green", "timeout", "red", "unknown", "read-error", "head-lag", "missing", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			oldNow, oldPause, oldFacts := ciWaitNow, ciWaitPause, makeTestFactsFn
			t.Cleanup(func() { ciWaitNow, ciWaitPause, makeTestFactsFn = oldNow, oldPause, oldFacts })
			now := time.Unix(0, 0)
			ciWaitNow = func() time.Time { return now }
			probes, polls := 0, 0
			makeTestFactsFn = func(func(...string) (string, error), string) (bool, int, error) { probes++; return false, 1, nil }
			g := &checksGH{checks: []prCheck{{Name: "build", Bucket: "pending"}}}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ciWaitPause = func(ctx context.Context, d time.Duration) error {
				polls++
				now = now.Add(d)
				if mode == "cancel" {
					cancel()
					return ctx.Err()
				}
				if mode == "pending-green" || mode == "unregistered-green" {
					g.checks = []prCheck{{Name: "build", Bucket: "pass"}}
				}
				return nil
			}
			switch mode {
			case "unregistered-green":
				g.checks = nil
			case "red":
				g.checks[0].Bucket = "fail"
			case "unknown":
				g.checks[0].Bucket = "other"
			case "read-error":
				g.err = errors.New("unavailable")
			case "head-lag":
				g.heads = []string{"different"}
			case "missing":
				g.checks[0].Bucket = "pass"
			}
			git := func(args ...string) (string, error) {
				if mode == "missing" && args[len(args)-1] == localGatesWorkflow {
					return localGatesWorkflow, nil
				}
				if args[len(args)-1] == ".github/workflows" {
					return ".github/workflows/ci.yml", nil
				}
				return "", nil
			}
			var out bytes.Buffer
			err := waitCIGate(ctx, g, "o/r", 3, "b", "h", git, nil, false, &out)
			switch mode {
			case "pending-green", "unregistered-green":
				if err != nil || polls != 1 {
					t.Fatalf("polls=%d err=%v", polls, err)
				}
			case "timeout":
				if err == nil || !strings.Contains(err.Error(), "build") || !strings.Contains(err.Error(), "20m") || polls != 40 {
					t.Fatalf("polls=%d err=%v", polls, err)
				}
			case "cancel":
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancel: %v", err)
				}
			default:
				if err == nil || polls != 0 {
					t.Fatalf("terminal error retried: polls=%d err=%v", polls, err)
				}
			}
			if probes != 1 {
				t.Fatalf("immutable suite facts probed %d times", probes)
			}
			if polls > 0 && !strings.Contains(out.String(), "waiting for CI on PR #3") {
				t.Fatalf("no progress: %s", out.String())
			}
		})
	}
}

func TestWaitCIGateSuiteTimeIsSeparate(t *testing.T) {
	dir := testfix.Repo(t, testfix.InitialCommit())
	writeRepoFile(t, dir, localChecksPath, `{"version":1,"checks":[{"name":"suite","command":["true"]}]}`)
	testfix.Git(t, dir, "add", ".")
	testfix.Git(t, dir, "commit", "-qm", "suite")
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		return strings.TrimSpace(string(out)), err
	}
	head, _ := git("rev-parse", "HEAD")
	g := newCIFakeGH()
	g.setHead(3, head, prCheck{Name: "build", Bucket: "pending"})
	oldNow, oldPause := ciWaitNow, ciWaitPause
	t.Cleanup(func() { ciWaitNow, ciWaitPause = oldNow, oldPause })
	now := time.Unix(0, 0)
	ciWaitNow = func() time.Time { return now }
	pauses, runs := 0, 0
	ciWaitPause = func(context.Context, time.Duration) error {
		pauses++
		now = now.Add(ciWaitInterval)
		g.setHead(3, head, prCheck{Name: "build", Bucket: "pass"})
		return nil
	}
	run := func(ctx context.Context, c localCheck, sha string) (localRunRecord, error) {
		runs++
		if _, ok := ctx.Deadline(); ok {
			t.Fatal("suite inherited polling deadline")
		}
		now = now.Add(time.Hour)
		err := g.PostStatus(ctx, "o/r", sha, localStatusContext(c.Name), "success", "passed")
		g.setHead(3, head, prCheck{Name: "late-check", Bucket: "pending"})
		return localRunRecord{Name: c.Name, SHA: sha, Exit: 0}, err
	}
	if err := waitCIGate(context.Background(), g, "o/r", 3, head, head, git, run, false, io.Discard); err != nil || runs != 1 || pauses != 2 {
		t.Fatalf("suite must not spend CI wait budget: runs=%d pauses=%d err=%v", runs, pauses, err)
	}
}

func TestLandingWaitsForPendingCI(t *testing.T) {
	_, _, gh := landingFixture(t, 1)
	gh.checks = []prCheck{{Name: "build", Bucket: "pending"}}
	oldPause := ciWaitPause
	t.Cleanup(func() { ciWaitPause = oldPause })
	pauses := 0
	ciWaitPause = func(context.Context, time.Duration) error {
		pauses++
		gh.checks[0].Bucket = "pass"
		return nil
	}
	if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil || pauses != 1 || gh.merges != 1 {
		t.Fatalf("durable wait not wired: pauses=%d merges=%d err=%v", pauses, gh.merges, err)
	}
}

// A read may honor cancellation or return a stale success despite cancellation.
// Both cases must spend the CI budget, unlike the local suite's runtime.
type budgetReadGH struct {
	checksGH
	before func(context.Context, string) error
}

func (g *budgetReadGH) PRHeadOID(ctx context.Context, repo string, number int) (string, error) {
	if err := g.before(ctx, "head"); err != nil {
		return "", err
	}
	return g.checksGH.PRHeadOID(ctx, repo, number)
}
func (g *budgetReadGH) PRChecks(ctx context.Context, repo string, number int) ([]prCheck, error) {
	if err := g.before(ctx, "checks"); err != nil {
		return nil, err
	}
	return g.checksGH.PRChecks(ctx, repo, number)
}

func TestWaitCIGateInflightBudget(t *testing.T) {
	for _, mode := range []string{"slow-head", "slow-checks", "late-green", "late-git"} {
		t.Run(mode, func(t *testing.T) {
			oldNow, oldPause, oldFacts := ciWaitNow, ciWaitPause, makeTestFactsFn
			t.Cleanup(func() { ciWaitNow, ciWaitPause, makeTestFactsFn = oldNow, oldPause, oldFacts })
			now := time.Unix(0, 0)
			ciWaitNow = func() time.Time { return now }
			makeTestFactsFn = func(func(...string) (string, error), string) (bool, int, error) { return false, 1, nil }
			polls := 0
			g := &budgetReadGH{checksGH: checksGH{checks: []prCheck{{Name: "build", Bucket: "pending"}}}}
			g.before = func(ctx context.Context, method string) error {
				if polls == 0 {
					return nil
				}
				if mode == "late-green" && method == "checks" {
					now = now.Add(time.Minute)
					return nil
				}
				if mode != "slow-"+method {
					return nil
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > time.Second {
					return errors.New("poll read lacks remaining-budget deadline")
				}
				<-ctx.Done()
				return ctx.Err()
			}
			ciWaitPause = func(context.Context, time.Duration) error {
				polls++
				now = time.Unix(0, 0).Add(ciWaitLimit - 20*time.Millisecond)
				g.checks = []prCheck{{Name: "build", Bucket: "pass"}}
				return nil
			}
			git := func(args ...string) (string, error) {
				if mode == "late-git" {
					now = now.Add(ciWaitLimit + time.Second)
				}
				if args[len(args)-1] == ".github/workflows" {
					return ".github/workflows/ci.yml", nil
				}
				return "", nil
			}
			err := waitCIGate(context.Background(), g, "o/r", 3, "b", "h", git, nil, false, io.Discard)
			if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "20m") {
				t.Fatalf("in-flight budget not enforced: %v", err)
			}
		})
	}
}

func (g *budgetReadGH) JobRun(ctx context.Context, _ string, _ int64) (jobRun, error) {
	return jobRun{}, g.before(ctx, "job")
}
func (g *budgetReadGH) CommitStatuses(ctx context.Context, _, _ string) (map[string]string, error) {
	return nil, g.before(ctx, "statuses")
}

func TestWaitCIGateAllReadsShareBudget(t *testing.T) {
	for _, method := range []string{"head", "checks", "job", "statuses"} {
		t.Run(method, func(t *testing.T) {
			oldNow := ciWaitNow
			t.Cleanup(func() { ciWaitNow = oldNow })
			now := time.Unix(0, 0)
			ciWaitNow = func() time.Time { return now }
			g := &budgetReadGH{before: func(ctx context.Context, got string) error {
				if got != method {
					t.Fatalf("read %s, want %s", got, method)
				}
				deadline, ok := ctx.Deadline()
				if !ok || time.Until(deadline) > time.Second {
					t.Fatal("missing remaining-budget deadline")
				}
				now = now.Add(time.Second)
				return nil // A late successful response must still be refused.
			}}
			b := &ciReadBudget{started: now.Add(-ciWaitLimit + 100*time.Millisecond)}
			bounded := ciBudgetGH{ghCaller: g, budget: b}
			var err error
			switch method {
			case "head":
				_, err = bounded.PRHeadOID(context.Background(), "o/r", 3)
			case "checks":
				_, err = bounded.PRChecks(context.Background(), "o/r", 3)
			case "job":
				_, err = bounded.JobRun(context.Background(), "o/r", 1)
			case "statuses":
				_, err = bounded.CommitStatuses(context.Background(), "o/r", "h")
			}
			if !errors.Is(err, context.DeadlineExceeded) || !b.expired {
				t.Fatalf("late read accepted: %v", err)
			}
		})
	}
}

func TestWaitCIGateGitContextBudget(t *testing.T) {
	oldNow, oldFacts := ciWaitNow, makeTestFactsFn
	t.Cleanup(func() { ciWaitNow, makeTestFactsFn = oldNow, oldFacts })
	now := time.Unix(0, 0)
	ciWaitNow = func() time.Time { return now }
	makeTestFactsFn = func(func(...string) (string, error), string) (bool, int, error) { return false, 1, nil }
	calls := 0
	factory := func(ctx context.Context) func(...string) (string, error) {
		return func(...string) (string, error) {
			calls++
			if _, ok := ctx.Deadline(); !ok {
				t.Fatal("git read lacks deadline")
			}
			if calls == 1 {
				now = now.Add(ciWaitLimit - 20*time.Millisecond)
				return "", nil
			}
			if d, _ := ctx.Deadline(); time.Until(d) > time.Second {
				t.Fatal("git read budget not reduced")
			}
			<-ctx.Done()
			return "", ctx.Err()
		}
	}
	fallback := func(...string) (string, error) { t.Fatal("context factory bypassed"); return "", nil }
	err := waitCIGate(context.Background(), &checksGH{}, "o/r", 3, "b", "h", fallback, nil, false, io.Discard, factory)
	if !errors.Is(err, context.DeadlineExceeded) || calls != 2 {
		t.Fatalf("git budget: calls=%d err=%v", calls, err)
	}
}

func TestWaitCIGateProbeCleanupAfterExpiry(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{false: "budget", true: "canceled"}[canceled], func(t *testing.T) {
			oldNow, oldFacts := ciWaitNow, makeTestFactsFn
			t.Cleanup(func() { ciWaitNow, makeTestFactsFn = oldNow, oldFacts })
			now := time.Unix(0, 0)
			ciWaitNow = func() time.Time { return now }
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cleanupCalls := 0
			makeTestFactsFn = func(git func(...string) (string, error), _ string) (bool, int, error) {
				if canceled {
					cancel()
				} else {
					now = now.Add(ciWaitLimit + time.Second)
				}
				for _, args := range [][]string{{"worktree", "remove", "--force", "/tmp/probe"}, {"worktree", "prune"}} {
					if _, err := git(args...); err != nil {
						t.Fatalf("cleanup refused: %v", err)
					}
				}
				return false, 1, nil
			}
			factory := func(ioCtx context.Context) func(...string) (string, error) {
				return func(args ...string) (string, error) {
					if args[0] == "worktree" {
						cleanupCalls++
						deadline, ok := ioCtx.Deadline()
						if ioCtx.Err() != nil || !ok || time.Until(deadline) > landingRecoveryTimeout {
							t.Fatalf("cleanup context not fresh and bounded: %v", ioCtx.Err())
						}
					}
					return "", nil
				}
			}
			err := waitCIGate(ctx, &checksGH{}, "o/r", 3, "b", "h", func(...string) (string, error) { return "", nil }, nil, false, io.Discard, factory)
			want := context.DeadlineExceeded
			if canceled {
				want = context.Canceled
			}
			if !errors.Is(err, want) || cleanupCalls != 2 {
				t.Fatalf("cleanup=%d err=%v", cleanupCalls, err)
			}
		})
	}
}
