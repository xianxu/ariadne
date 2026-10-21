//go:build unix

package main

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// The immediate leader exits on TERM while its child closes inherited pipes
// and continues cleanup. Parent Wait alone is deliberately insufficient.
func TestRunLocalCheckRunnerGraceAfterLeaderExit(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "unused")
	tmp := t.TempDir()
	ready := filepath.Join(tmp, "ready")
	clean := filepath.Join(tmp, "clean")
	script := `import os,signal,sys,time
pid=os.fork()
if pid:
 signal.signal(signal.SIGTERM,lambda *_:sys.exit(0))
 while True: time.sleep(.01)
else:
 def stop(*_):
  os.close(1);os.close(2)
  time.sleep(6)
  open(sys.argv[2],"w").write("clean")
  os._exit(0)
 signal.signal(signal.SIGTERM,stop)
 open(sys.argv[1],"w").write(str(os.getpid()))
 while True: time.sleep(.01)
`
	c := localCheck{Name: "vm", Resource: "runner", Command: []string{"python3", "-c", script, ready, clean}, TimeoutMinutes: 60}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := runLocalCheck(ctx, d, c, sha); done <- err }()
	if testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil) == 0 {
		t.Fatal("child did not become ready")
	}
	cancel()
	select {
	case err := <-done:
		var interrupted *queueInterruptedError
		if !errors.As(err, &interrupted) {
			t.Fatalf("not interrupted: %v", err)
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("cancel did not join")
	}
	if _, err := os.Stat(clean); err != nil {
		t.Fatalf("cleanup killed after leader exit: %v", err)
	}
}

func TestRunnerCleanupEscalatesAndJoins(t *testing.T) {
	ready := filepath.Join(t.TempDir(), "ready")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", "-c", `import os,signal,sys,time
signal.signal(signal.SIGTERM,signal.SIG_IGN)
open(sys.argv[1],"w").write(str(os.getpid()))
while True: time.sleep(.01)
`, ready)
	processgroup.Configure(cmd)
	join := configureRunnerCleanupWithin(cmd, 100*time.Millisecond)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = processgroup.Terminate(cmd, true) })
	if testfix.AwaitPIDFile(ready, testfix.HangCeiling, nil) == 0 {
		t.Fatal("child not ready")
	}
	cancel()
	if err := cmd.Wait(); err == nil {
		t.Fatal("ignoring child returned success")
	}
	if err := join(); err != nil {
		t.Fatal(err)
	}
	if err := testfix.AwaitExit(cmd.Process.Pid, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerCleanupNilTerminationIsNotCompletion(t *testing.T) {
	signals := 0
	err := finishRunnerGroup(time.Now(), 0, func() bool { return false }, func(force bool) error { signals++; return nil })
	var uncertain *runnerCleanupUnconfirmed
	if !errors.As(err, &uncertain) || signals != 1 {
		t.Fatalf("unconfirmed group accepted: %v signals%d", err, signals)
	}
}

func TestRunnerCleanupUnconfirmedPreservesRecovery(t *testing.T) {
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	d, sha, _, _ := localRunRepo(t, "unused")
	var retainedScratch string
	t.Cleanup(func() {
		if retainedScratch != "" {
			_ = os.RemoveAll(retainedScratch)
		}
	})
	d.runnerCleanup = func(cmd *exec.Cmd) func() error {
		return func() error {
			for _, v := range cmd.Env {
				if strings.HasPrefix(v, "TMPDIR=") {
					retainedScratch = strings.TrimPrefix(v, "TMPDIR=")
				}
			}
			return &runnerCleanupUnconfirmed{Detail: "fixture uncertainty"}
		}
	}
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "vm", Resource: "runner", Command: []string{"sh", "-c", "echo retained"}, TimeoutMinutes: 30}, sha)
	var uncertain *runnerCleanupUnconfirmed
	if !errors.As(err, &uncertain) {
		t.Fatalf("lost uncertainty %v", err)
	}
	if _, err := os.Stat(rec.LogPath); err != nil {
		t.Fatalf("lost recovery log %s: %v", rec.LogPath, err)
	}
	trees, err := d.Git("worktree", "list", "--porcelain")
	if err != nil {
		t.Fatal(err)
	}
	retained := false
	for _, line := range strings.Split(trees, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			wt := strings.TrimPrefix(line, "worktree ")
			if wt != d.Root {
				retained = true
				t.Cleanup(func() { _, _ = d.Git("worktree", "unlock", wt); _, _ = d.Git("worktree", "remove", "--force", wt) })
				old := time.Now().Add(-48 * time.Hour)
				if err := os.Chtimes(wt, old, old); err != nil {
					t.Fatal(err)
				}
				other, _, _, _ := localRunRepo(t, "unused")
				sweepStaleLocalWorktrees(other)
				if _, err := os.Stat(wt); err != nil {
					t.Fatalf("stale sweep erased uncertain inputs: %v", err)
				}
			}
		}
	}
	if !retained {
		t.Fatal("unconfirmed descendants lost their worktree")
	}
	root, err := queueStateRoot()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := (&queuegate.Store{Root: root}).Snapshot(context.Background())
	if err != nil || snapshot.Complete || len(snapshot.Active) != 1 || snapshot.Active[0].State != "unknown" {
		t.Fatalf("uncertainty became absence: %+v %v", snapshot, err)
	}
}
