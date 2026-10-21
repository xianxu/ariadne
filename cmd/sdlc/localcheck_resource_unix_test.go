//go:build unix

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
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
