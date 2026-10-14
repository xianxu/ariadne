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
