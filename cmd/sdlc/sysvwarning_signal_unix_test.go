//go:build unix

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Subprocess only: real signals must never target the parent test process.
func TestSysVAfterProbeSignalProcess(t *testing.T) {
	if os.Getenv("SDLC_SYSV_SIGNAL_CHILD") != "1" {
		t.Skip("owned subprocess fixture")
	}
	d, sha, _, _ := localRunRepo(t, "acquired")
	t.Setenv("WF_SDLC_QUEUE_ROOT", os.Getenv("SDLC_SYSV_SIGNAL_STATE"))
	d.sysvObserver = &fakeSysVObserver{onAfter: func() {
		if err := os.WriteFile(os.Getenv("SDLC_SYSV_SIGNAL_READY"), []byte(fmt.Sprint(os.Getpid())), 0600); err != nil {
			t.Fatal(err)
		}
		// Positive readiness lets the parent signal without a launch-time guess.
		// A late signal must have default disposition after managed retirement.
		select {}
	}}
	logDir := t.TempDir()
	_, err := runUnderResourceBatch(t.Context(), d, d.Root, filepath.Join(d.Root, "scripts/test-lock.py"), logDir, "zellij", sha, 1, suitePrioritySubject{}, [][]string{{"true"}}, filepath.Join(logDir, "run.log"), time.Minute, "")
	t.Fatalf("blocked post-run diagnostic returned: %v", err)
}

func TestSysVAfterProbeHasDefaultSignalDisposition(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			root := t.TempDir()
			ready := filepath.Join(root, "ready")
			state := filepath.Join(root, "queue")
			cmd := exec.Command("python3", "-c", "import os,signal,sys; signal.signal(signal.SIGINT,signal.SIG_DFL); os.execv(sys.argv[1],sys.argv[1:])", os.Args[0], "-test.run=^TestSysVAfterProbeSignalProcess$")
			processgroup.Configure(cmd)
			cmd.WaitDelay = time.Second
			cmd.Env = append(os.Environ(), "SDLC_SYSV_SIGNAL_CHILD=1", "SDLC_SYSV_SIGNAL_STATE="+state, "SDLC_SYSV_SIGNAL_READY="+ready, "TMPDIR="+root, "WF_TEST_SCRATCH="+root)
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			finished := false
			t.Cleanup(func() {
				_ = processgroup.Terminate(cmd, true)
				if !finished {
					select {
					case <-done:
					case <-time.After(testfix.HangCeiling):
						t.Error("owned signal fixture did not join")
					}
				}
			})
			if pid := testfix.AwaitPIDFile(ready, testfix.HangCeiling, func() bool {
				select {
				case err := <-done:
					finished = true
					t.Errorf("fixture exited before after-probe: %v", err)
					return true
				default:
					return false
				}
			}); pid == 0 {
				t.Fatal("after-probe did not become ready")
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				finished = true
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatalf("default signal did not terminate fixture: %v %s", err, &output)
				}
				status, ok := exit.Sys().(syscall.WaitStatus)
				if !ok || !status.Signaled() || status.Signal() != sig {
					t.Fatalf("signal disposition changed: %v %s", err, &output)
				}
			case <-time.After(testfix.HangCeiling):
				t.Fatal("after-probe swallowed default termination")
			}
			for _, dir := range []string{"operations", "interrupted"} {
				rows, err := os.ReadDir(filepath.Join(state, dir))
				if err != nil && !os.IsNotExist(err) {
					t.Fatal(err)
				}
				if len(rows) != 0 {
					t.Fatalf("late signal changed completed queue state: %s %v", dir, rows)
				}
			}
		})
	}
}
