//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestNetSignalProcess(t *testing.T) {
	if os.Getenv("SDLC_NET_SIGNAL_ROLE") != "cli" {
		return
	}
	r := buildRoot()
	code := executeCLI(r, []string{"net", "check", "--host", "example.com"})
	if !errors.Is(r.Context().Err(), context.Canceled) {
		os.Exit(93)
	}
	os.Exit(code)
}

func TestNetCheckCLISignalsJoinNativeUtilities(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			dir := t.TempDir()
			ready := filepath.Join(dir, "ready")
			if err := os.Mkdir(ready, 0700); err != nil {
				t.Fatal(err)
			}
			utility := "#!/bin/sh\ntrap '' INT TERM\nsleep 3600 &\nprintf '%s %s\\n' \"$$\" \"$!\" > \"$SDLC_NET_READY/$$\"\nwait\n"
			for _, name := range []string{"dig", "dscacheutil", "getent"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(utility), 0700); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestNetSignalProcess$")
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "SDLC_NET_SIGNAL_ROLE=cli", "SDLC_NET_READY="+ready, "GORACE=atexit_sleep_ms=0")
			cmd.WaitDelay = time.Second
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { defer close(done); done <- cmd.Wait() }()
			finished := false
			var pids []int
			t.Cleanup(func() {
				for _, pid := range pids {
					_ = syscall.Kill(-pid, syscall.SIGKILL)
					_ = syscall.Kill(pid, syscall.SIGKILL)
				}
				if !finished {
					_ = cmd.Process.Kill()
					<-done
				}
			})
			testDeadline, hasDeadline := t.Deadline()
			budget := landingCompletionBudget(testDeadline, hasDeadline, time.Now())
			deadline := time.Now().Add(budget)
			for len(pids) < 4 { // both resolver paths have a real utility and child ready.
				entries, _ := os.ReadDir(ready)
				pids = nil
				for _, e := range entries {
					data, _ := os.ReadFile(filepath.Join(ready, e.Name()))
					fields := strings.Fields(string(data))
					if len(fields) != 2 {
						continue
					}
					for _, f := range fields {
						n, err := strconv.Atoi(f)
						if err == nil && n > 0 {
							pids = append(pids, n)
						}
					}
				}
				select {
				case err := <-done:
					finished = true
					t.Fatalf("CLI exited before signal: %v\n%s", err, out.String())
				default:
				}
				if time.Now().After(deadline) {
					t.Fatal("native probes did not become ready")
				}
				if len(pids) < 4 {
					time.Sleep(time.Millisecond)
				}
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				finished = true
				if err == nil || cmd.ProcessState.ExitCode() != 1 {
					t.Fatalf("signal lost cancellation owner: %v; state=%v\n%s", err, cmd.ProcessState, out.String())
				}
			case <-time.After(budget):
				t.Fatal("CLI did not join cancelled probes")
			}
			for _, pid := range pids {
				if err := testfix.AwaitExit(pid, budget); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
