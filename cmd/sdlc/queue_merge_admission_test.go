package main

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestQueueMergePauseCoversNormalAndResume(t *testing.T) {
	for _, name := range []string{"normal", "resume", "legacy"} {
		t.Run(name, func(t *testing.T) {
			roots, _, gh := landingFixture(t, 1)
			workspaceRoot := roots[1]
			if name == "legacy" {
				git(t, roots[1], "switch", "main-slot1")
				workspaceRoot = filepath.Join(t.TempDir(), "ordinary-worktree")
				git(t, roots[1], "worktree", "add", workspaceRoot, landingTestBranch)
				t.Chdir(workspaceRoot)
			}
			resume := name == "resume"
			if resume {
				if err := gh.LandingMerge(context.Background(), gh.pr.Repo, gh.pr.Number, gh.pr.HeadOID); err != nil {
					t.Fatal(err)
				}
			}
			root := t.TempDir()
			t.Setenv("WF_SDLC_QUEUE_ROOT", root)
			store := &queuegate.Store{Root: root}
			if err := store.Pause(context.Background(), "drain-test", "fixture hold", false); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f := landingFlags()
			f.Context = ctx
			done := mergeAsync(f)
			deadline := time.After(10 * time.Second)
			observed := false
			for !observed {
				select {
				case err := <-done:
					t.Fatalf("merge exited before publishing its paused waiting lease: %v", err)
				case <-deadline:
					cancel()
					_ = finished(t, done)
					t.Fatal("merge never published its waiting lease")
				case <-time.After(5 * time.Millisecond):
					paths, _ := filepath.Glob(filepath.Join(root, "operations", "*.json"))
					for _, path := range paths {
						data, readErr := os.ReadFile(path)
						if readErr != nil {
							continue
						}
						op, decodeErr := queuegate.DecodeOperation(data)
						if decodeErr == nil && op.Phase == "waiting" {
							observed = true
							break
						}
					}
				}
			}
			cancel()
			err := finished(t, done)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("paused merge returned %v instead of waiting for resume", err)
			}
			if got := git(t, workspaceRoot, "branch", "--show-current"); got != landingTestBranch {
				t.Fatalf("cleanup ran while paused: branch %q", got)
			}
			if !resume && gh.merges != 0 {
				t.Fatalf("merged while paused: %d", gh.merges)
			}
			snap, err := store.Snapshot(context.Background())
			if err != nil || len(snap.Active) != 0 {
				t.Fatalf("ended waiter stayed active: %+v %v", snap, err)
			}
		})
	}
}

func TestQueueMergeDryRunDoesNotEnroll(t *testing.T) {
	_, _, _ = landingFixture(t, 1)
	root := filepath.Join(t.TempDir(), "queue")
	t.Setenv("WF_SDLC_QUEUE_ROOT", root)
	f := landingFlags()
	f.DryRun = true
	if err := runMerge(io.Discard, io.Discard, f); err != nil {
		t.Fatal(err)
	}
	// A dry run does not need a writable queue or publish a lease.
	entries, err := filepath.Glob(filepath.Join(root, "operations", "*.json"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("dryrun enrolled: %v %v", entries, err)
	}
}

// GitHub's post-admission read is the first point shared by ordinary landing
// and already-merged cleanup. It must see an active lifetime binding.
type queueActiveMergeGH struct {
	*landingFakeGH
	queueRoot string
	calls     int
}

func (g *queueActiveMergeGH) LandingPRs(ctx context.Context, repo, branch string) ([]landingPR, error) {
	g.calls++
	paths, err := filepath.Glob(filepath.Join(g.queueRoot, "operations", "*.json"))
	if err != nil || len(paths) != 1 {
		g.t.Fatalf("PR read %d lacks lifetime lease: %v %v", g.calls, paths, err)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		g.t.Fatal(err)
	}
	op, err := queuegate.DecodeOperation(data)
	if err != nil {
		g.t.Fatal(err)
	}
	expected := "running"
	if g.calls == 1 {
		expected = "waiting"
	}
	if op.Phase != expected {
		g.t.Fatalf("PR read %d phase=%q, want %q", g.calls, op.Phase, expected)
	}
	if expected == "running" && g.pr.State == "OPEN" {
		common := git(g.t, op.Root, "rev-parse", "--path-format=absolute", "--git-common-dir")
		raw, err := os.ReadFile(filepath.Join(common, "sdlc", "landing", "head.lock"))
		if err != nil {
			g.t.Fatal(err)
		}
		fields, err := queueLegacyFields(raw)
		if err != nil || !queueLegacyBound(fields, "landing", queueLegacyIndexKnown([]queuegate.Operation{op}, nil), processStartTime) {
			g.t.Fatalf("new landing appeared as unbound legacy activity: %s %v", raw, err)
		}
	}
	return g.landingFakeGH.LandingPRs(ctx, repo, branch)
}
func TestQueueMergeActiveLeaseCoversNormalAndResumeCleanup(t *testing.T) {
	for _, resume := range []bool{false, true} {
		name := "normal"
		if resume {
			name = "resume"
		}
		t.Run(name, func(t *testing.T) {
			_, _, gh := landingFixture(t, 1)
			if resume {
				if err := gh.LandingMerge(context.Background(), gh.pr.Repo, gh.pr.Number, gh.pr.HeadOID); err != nil {
					t.Fatal(err)
				}
			}
			root := t.TempDir()
			t.Setenv("WF_SDLC_QUEUE_ROOT", root)
			observed := &queueActiveMergeGH{landingFakeGH: gh, queueRoot: root}
			ghClient = observed
			if err := runMerge(io.Discard, io.Discard, landingFlags()); err != nil {
				t.Fatal(err)
			}
			if observed.calls < 2 {
				t.Fatalf("post-admission PR read not reached: %d", observed.calls)
			}
			snapshot, err := (&queuegate.Store{Root: root}).Snapshot(context.Background())
			if err != nil || len(snapshot.Active) != 0 {
				t.Fatalf("cleanup did not retire lease: %+v %v", snapshot, err)
			}
		})
	}
}
