//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Exercise the real CLI signal boundary and real queue helper without GitHub
// or the machine's shared test slots.
func TestLocalCheckSignalProcess(t *testing.T) {
	if os.Getenv("SDLC_QUEUE_SIGNAL_CHILD") != "1" {
		return
	}
	rootDir, sha := os.Getenv("SDLC_QUEUE_ROOT"), os.Getenv("SDLC_QUEUE_SHA")
	d := localRunDeps{Root: rootDir, Repo: "o/r", GH: newCIFakeGH(), Stderr: os.Stderr,
		Git: func(args ...string) (string, error) {
			out, err := exec.Command("git", append([]string{"-C", rootDir}, args...)...).CombinedOutput()
			return string(out), err
		},
	}
	root := &cobra.Command{Use: "sdlc"}
	local := NewLocalCheckCmd()
	run, _, err := local.Find([]string{"run"})
	if err != nil {
		t.Fatal(err)
	}
	run.RunE = func(cmd *cobra.Command, _ []string) error {
		_, err := runLocalCheck(cmd.Context(), d, localCheck{Name: "suite", Command: []string{"true"}, TimeoutMinutes: 30}, sha)
		return err
	}
	root.AddCommand(local)
	os.Exit(executeCLI(root, []string{"local-check", "run", "--name", "suite"}))
}

func TestLocalCheckSignalCancelsQueuedHelper(t *testing.T) {
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			d, _, _, _ := localRunRepo(t, "unused")
			temp := t.TempDir()
			lockPath, helperPIDPath := filepath.Join(temp, "lock"), filepath.Join(temp, "helper.pid")
			// Record the exact helper we created: never find a global PID to stop.
			prefix := "import os, sys\nif '--help' not in sys.argv:\n    open(os.environ['SDLC_QUEUE_HELPER_PID'], 'w').write(str(os.getpid()))\n"
			if err := os.WriteFile(filepath.Join(d.Root, "scripts", "test-lock.py"), append([]byte(prefix), helper...), 0700); err != nil {
				t.Fatal(err)
			}
			for _, args := range [][]string{{"add", "."}, {"commit", "-qm", "real helper"}} {
				if out, err := d.Git(args...); err != nil {
					t.Fatalf("git: %s %v", out, err)
				}
			}
			rawSHA, err := d.Git("rev-parse", "HEAD")
			if err != nil {
				t.Fatal(err)
			}
			sha := strings.TrimSpace(rawSHA)
			t.Setenv("WF_TEST_LOCK_FILE", lockPath)
			t.Setenv("WF_TEST_SUITES", "1")
			t.Setenv("WF_TEST_LOCK", "")
			t.Setenv("WF_TEST_LOCK_TIMEOUT", "")
			t.Setenv("WF_TEST_LOCK_HELD_BY", "")
			t.Setenv("SDLC_QUEUE_HELPER_PID", helperPIDPath)
			// An unrelated owner holds the slot throughout cancellation.
			slot, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				t.Fatal(err)
			}
			defer slot.Close()
			if err := syscall.Flock(int(slot.Fd()), syscall.LOCK_EX); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestLocalCheckSignalProcess$")
			processgroup.Configure(cmd)
			cmd.WaitDelay = time.Second
			cmd.Env = append(os.Environ(), "SDLC_QUEUE_SIGNAL_CHILD=1", "SDLC_QUEUE_ROOT="+d.Root, "SDLC_QUEUE_SHA="+sha, "SDLC_QUEUE_HELPER_PID="+helperPIDPath)
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			finished, helperPID, ownerPID := false, 0, 0
			t.Cleanup(func() {
				_ = processgroup.Terminate(cmd, true)
				for _, pid := range []int{helperPID, ownerPID} {
					if pid > 0 {
						_ = syscall.Kill(pid, syscall.SIGKILL)
					}
				}
				if !finished {
					select {
					case <-done:
					case <-time.After(testfix.HangCeiling):
						t.Error("fixture did not exit")
					}
				}
			})
			ticketDir := lockPath + ".tickets"
			deadline := time.Now().Add(testfix.HangCeiling)
			for time.Now().Before(deadline) {
				raw, _ := os.ReadFile(helperPIDPath)
				helperPID, _ = strconv.Atoi(string(raw))
				paths, _ := filepath.Glob(filepath.Join(ticketDir, "[0-9]*"))
				if len(paths) == 1 {
					raw, _ := os.ReadFile(paths[0])
					var rec struct {
						PID int `json:"pid"`
					}
					if json.Unmarshal(raw, &rec) == nil {
						ownerPID = rec.PID
					}
				}
				if helperPID > 0 && ownerPID > 0 {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			if helperPID == 0 || ownerPID == 0 {
				t.Fatal("run did not reach the queue")
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				finished = true
				if err == nil {
					t.Error("cancelled run succeeded")
				}
			case <-time.After(5 * time.Second):
				t.Fatal("queued cancellation exceeded five seconds")
			}
			if err := testfix.AwaitExit(helperPID, time.Second); err != nil {
				t.Errorf("orphan helper: %v", err)
			}
			if err := testfix.AwaitExit(ownerPID, time.Second); err != nil {
				t.Errorf("orphan holder: %v", err)
			}
			if !strings.Contains(output.String(), "context canceled") {
				t.Errorf("missing managed cancellation: %s", output.String())
			}
			if out, err := d.Git("worktree", "list", "--porcelain"); err != nil || strings.Contains(out, "sdlc-local-") {
				t.Errorf("cancelled run retained its worktree: %s %v", out, err)
			}
			// A fresh run must remain queued until the unrelated slot owner releases.
			ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
			defer cancel()
			next := make(chan error, 1)
			nextFinished := make(chan struct{})
			t.Cleanup(func() {
				cancel()
				select {
				case <-nextFinished:
				case <-time.After(testfix.HangCeiling):
					t.Error("successor cleanup did not finish")
				}
			})
			go func() {
				_, release, err := acquireSuiteSlot(ctx, d, d.Root, 1, filepath.Join(d.Root, "scripts", "test-lock.py"))
				if err == nil {
					release()
				}
				next <- err
				close(nextFinished)
			}()
			select {
			case err := <-next:
				t.Fatalf("unrelated holder lost its lock: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			if err := syscall.Flock(int(slot.Fd()), syscall.LOCK_UN); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-next:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("successor not admitted after cancellation")
			}
			// A killed helper leaves an unlocked ticket; admission prunes it normally.
			paths, _ := filepath.Glob(filepath.Join(ticketDir, "[0-9]*"))
			if len(paths) != 0 {
				t.Errorf("retained queue tickets: %v", paths)
			}
		})
	}
}
