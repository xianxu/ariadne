package gitx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// remoteCommand is a git command that contacts the remote.
func remoteCommand(args []string) bool {
	return len(args) > 0 && (args[0] == "fetch" || args[0] == "ls-remote" || args[0] == "push")
}

// slowLocal makes every local git command take d, as on a loaded machine where
// process spawns are slow (#360), leaving remote commands at their own speed.
func slowLocal(t *testing.T, d time.Duration) {
	t.Helper()
	original := runGitInContext
	t.Cleanup(func() { runGitInContext = original })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if !remoteCommand(args) {
			time.Sleep(d)
		}
		return original(ctx, dir, env, args...)
	}
}

// #360: the remote budget is not spent by local work. Local commands that
// together take far longer than the budget still complete; the old shape, one
// deadline over the whole operation, fails the same read.
func TestRemoteBudgetDoesNotCountLocalWork(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	slowLocal(t, 150*time.Millisecond)

	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	tf.BoundRemote(300 * time.Millisecond)
	start := time.Now()
	got, err := tf.Read("note.md")
	if err != nil || string(got) != "seed\n" {
		t.Fatalf("bounded remote read: %q, %v", got, err)
	}
	if elapsed := time.Since(start); elapsed < 300*time.Millisecond {
		t.Fatalf("local work took %v; the fixture must exceed the budget to prove anything", elapsed)
	}

	whole, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	old, err := NewTrunkFileContext(whole, repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := old.Read("note.md"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("a whole-operation deadline should fail on slow local work, got %v", err)
	}
}

// hangRemote makes every remote command block until its context ends.
func hangRemote(t *testing.T, delay time.Duration) {
	t.Helper()
	original := runGitInContext
	t.Cleanup(func() { runGitInContext = original })
	runGitInContext = func(ctx context.Context, dir string, env []string, args ...string) ([]byte, []byte, error) {
		if remoteCommand(args) {
			select {
			case <-ctx.Done():
				return nil, nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		return original(ctx, dir, env, args...)
	}
}

// A remote that never answers is bounded by the budget and named as such.
func TestRemoteBudgetBoundsAHangingRemote(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	hangRemote(t, time.Hour)
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	tf.BoundRemote(200 * time.Millisecond)
	start := time.Now()
	_, err = tf.Read("note.md")
	if err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("hanging remote: %v, want an error naming the deadline", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("hanging remote held the read for %v with a 200ms budget", elapsed)
	}
}

// The budget is one budget for all of a trunk's remote commands, from the first:
// a hanging remote cannot multiply it by the number of remote calls.
func TestRemoteBudgetIsSharedAcrossRemoteCommands(t *testing.T) {
	repo, _ := trunkFixture(t, "seed\n")
	hangRemote(t, 150*time.Millisecond)
	tf, err := NewTrunkFileContext(context.Background(), repo, "origin", "main")
	if err != nil {
		t.Fatal(err)
	}
	tf.BoundRemote(200 * time.Millisecond)
	if _, err := tf.RemoteExists(); err != nil {
		t.Fatalf("first remote command, within the budget: %v", err)
	}
	if _, err := tf.Read("note.md"); err == nil || !strings.Contains(err.Error(), "deadline") {
		t.Fatalf("second remote command past the shared budget: %v, want the deadline", err)
	}
}
