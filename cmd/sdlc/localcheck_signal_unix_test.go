//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// Exercise the real CLI signal boundary and real queue helper without GitHub
// or the machine's shared test slots.
func TestLocalCheckSignalProcess(t *testing.T) {
	if os.Getenv("SDLC_QUEUE_SIGNAL_CHILD") != "1" {
		return
	}
	// TestMain prefers the real Git directory; restore this fixture's
	// transport shim after that setup, without replacing the handler.
	if bin := os.Getenv("SDLC_QUEUE_BIN"); bin != "" {
		_ = os.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
	}
	os.Args = []string{"sdlc", "local-check", "run", "--name", "suite"}
	main()
	os.Exit(0)
}

func TestLocalCheckSignalCancelsQueuedHelper(t *testing.T) {
	helper, err := os.ReadFile("../../scripts/test-lock.py")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		sig   syscall.Signal
		phase string
	}{{syscall.SIGINT, ""}, {syscall.SIGTERM, ""}, {syscall.SIGTERM, "cleanup"}} {
		sig := tc.sig
		name := sig.String()
		if tc.phase != "" {
			name += "/" + tc.phase
		}
		t.Run(name, func(t *testing.T) {
			d, _, _, _ := localRunRepo(t, "unused")
			temp := t.TempDir()
			lockPath, helperPIDPath := filepath.Join(temp, "lock"), filepath.Join(temp, "helper.pid")
			cleanupPIDPath := filepath.Join(temp, "cleanup.pid")
			// The cleanup phase's shim waits on this gate after reporting itself,
			// so the test, not a clock, decides when cleanup Git returns (#434).
			cleanupGate := filepath.Join(temp, "cleanup.gate")
			if err := syscall.Mkfifo(cleanupGate, 0o600); err != nil {
				t.Fatal(err)
			}
			// Opening the write end waits for the shim to open the read end,
			// so the release lands whichever side gets there first.
			releaseCleanup := func() {
				go func() {
					if f, err := os.OpenFile(cleanupGate, os.O_WRONLY, 0); err == nil {
						_ = f.Close()
					}
				}()
			}
			t.Cleanup(func() {
				if f, err := os.OpenFile(cleanupGate, os.O_WRONLY|syscall.O_NONBLOCK, 0); err == nil {
					_ = f.Close()
				}
			})
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
			wireLocalCheckSignalCLI(t, d.Root)
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
			cmd.Dir = d.Root
			cmd.WaitDelay = time.Second
			cmd.Env = append(os.Environ(), "SDLC_QUEUE_SIGNAL_CHILD=1", "SDLC_QUEUE_ROOT="+d.Root, "SDLC_QUEUE_SHA="+sha, "SDLC_QUEUE_HELPER_PID="+helperPIDPath, "SDLC_BLOCK_PHASE="+tc.phase, "SDLC_PHASE_READY="+cleanupPIDPath, "SDLC_CLEANUP_GATE="+cleanupGate)
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			finished, helperPID, ownerPID := false, 0, 0
			t.Cleanup(func() {
				if raw, _ := os.ReadFile(cleanupPIDPath); len(raw) > 0 {
					if pid, _ := strconv.Atoi(string(raw)); pid > 0 {
						_ = syscall.Kill(-pid, syscall.SIGKILL)
						_ = syscall.Kill(pid, syscall.SIGKILL)
					}
				}
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
				select {
				case err := <-done:
					finished = true
					t.Fatalf("production handler exited before queue: %v %s", err, output.String())
				default:
				}
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
			cleanupPID := 0
			if tc.phase == "cleanup" {
				// Cancellation must not skip retirement: the cancelled run is
				// still waiting on its cleanup Git when that Git reports in.
				var exited error
				cleanupPID = testfix.AwaitPIDFile(cleanupPIDPath, testfix.HangCeiling, func() bool {
					select {
					case exited = <-done:
						finished = true
						return true
					default:
						return false
					}
				})
				if finished {
					t.Fatalf("cancelled run exited without its cleanup Git: %v\n%s", exited, output.String())
				}
				if cleanupPID == 0 {
					t.Fatal("cleanup never reached its Git")
				}
				select {
				case err := <-done:
					finished = true
					t.Fatalf("cancelled run did not wait for its cleanup Git: %v", err)
				default:
				}
				releaseCleanup()
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
			if cleanupPID > 0 {
				if err := testfix.AwaitExit(cleanupPID, time.Second); err != nil {
					t.Errorf("cleanup Git outlived its run: %v", err)
				}
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

// Keep the production handler intact. Only GitHub transport and the private
// fixture's public-shaped URL are substituted; queueing/cleanup are real.
func wireLocalCheckSignalCLI(t *testing.T, root string) {
	t.Helper()
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	call := func(at string, args ...string) {
		t.Helper()
		cmd := exec.Command(realGit, args...)
		cmd.Dir = at
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s %v", args, out, err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, ".sdlc"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sdlc", "local-checks.json"), []byte(`{"version":1,"checks":[{"name":"suite","command":["true"],"timeout_minutes":1,"weight":1}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	call(root, "add", ".sdlc")
	call(root, "commit", "-qm", "production command declaration")
	remote := t.TempDir()
	call(remote, "init", "--bare", "-q")
	call(root, "remote", "add", "origin", remote)
	call(root, "push", "-q", "origin", "main")
	bin := t.TempDir()
	script := `#!/bin/sh
if [ "$SDLC_BLOCK_PHASE" = fetch ] && case " $* " in *" fetch "*) true;; *) false;; esac; then
  printf '%s' "$$" > "$SDLC_PHASE_READY"
  exec sleep 60
fi
if [ "$SDLC_BLOCK_PHASE" = cleanup ] && case " $* " in *" worktree remove --force "*) true;; *) false;; esac; then
  printf '%s' "$$" > "$SDLC_PHASE_READY"
  cat "$SDLC_CLEANUP_GATE" >/dev/null
fi
if [ "$SDLC_BLOCK_PHASE" = trailer ] && case " $* " in *" show -s --format="*) true;; *) false;; esac; then
  printf '%s' "$$" > "$SDLC_PHASE_READY"
  exec sleep 60
fi
if case " $* " in *" remote get-url origin "*) true;; *) false;; esac; then
  echo https://github.com/fixture/queue.git
  exit 0
fi
exec ` + strconv.Quote(realGit) + ` "$@"
`
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nexit 0\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("SDLC_QUEUE_BIN", bin)
	t.Cleanup(func() {
		canonical, err := filepath.EvalSymlinks(root)
		if err != nil {
			t.Error(err)
			return
		}
		cmd := exec.Command(realGit, "-C", root, "worktree", "list", "--porcelain")
		out, _ := cmd.Output()
		for _, line := range strings.Split(string(out), "\n") {
			if at, ok := strings.CutPrefix(line, "worktree "); ok && at != canonical {
				if out, err := exec.Command(realGit, "-C", root, "worktree", "remove", "--force", at).CombinedOutput(); err != nil {
					t.Errorf("fixture worktree cleanup: %s %v", out, err)
				}
			}
		}
	})
}

func TestLocalCheckSignalPreservesOutsideOwnedPhase(t *testing.T) {
	for _, phase := range []string{"fetch", "trailer"} {
		t.Run(phase, func(t *testing.T) {
			for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
				t.Run(sig.String(), func(t *testing.T) {
					d, _, _, _ := localRunRepo(t, "skipped")
					wireLocalCheckSignalCLI(t, d.Root)
					ready := filepath.Join(t.TempDir(), "fetch.pid")
					// An asynchronous suite can inherit SIGINT=SIG_IGN. The
					// production CLI correctly preserves that precondition;
					// this case tests default termination, so establish SIG_DFL
					// before exec without changing the production handler.
					cmd := exec.Command("python3", "-c",
						"import os, signal, sys; signal.signal(signal.SIGINT, signal.SIG_DFL); os.execv(sys.argv[1], sys.argv[1:])",
						os.Args[0], "-test.run=^TestLocalCheckSignalProcess$")
					processgroup.Configure(cmd)
					cmd.Dir = d.Root
					cmd.WaitDelay = time.Second
					cmd.Env = append(os.Environ(), "SDLC_QUEUE_SIGNAL_CHILD=1", "SDLC_BLOCK_PHASE="+phase, "SDLC_PHASE_READY="+ready)
					var output bytes.Buffer
					cmd.Stdout = &output
					cmd.Stderr = &output
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
								t.Error("production CLI did not finish cleanup")
							}
						}
					})
					deadline := time.Now().Add(testfix.HangCeiling)
					for {
						if raw, _ := os.ReadFile(ready); len(raw) > 0 {
							break
						}
						select {
						case err := <-done:
							finished = true
							t.Fatalf("production handler exited before blocked phase: %v %s", err, output.String())
						default:
						}
						if time.Now().After(deadline) {
							t.Fatal("production blocking phase did not start")
						}
						time.Sleep(5 * time.Millisecond)
					}
					if err := cmd.Process.Signal(sig); err != nil {
						t.Fatal(err)
					}
					select {
					case err := <-done:
						finished = true
						var exit *exec.ExitError
						if !errors.As(err, &exit) {
							t.Fatalf("signal did not terminate CLI: %v", err)
						}
						status, ok := exit.Sys().(syscall.WaitStatus)
						if !ok || !status.Signaled() || status.Signal() != sig {
							t.Fatalf("default termination changed: %v %s", err, output.String())
						}
					case <-time.After(3 * time.Second):
						t.Fatal("production CLI swallowed termination outside owned phase")
					}
				})
			}

		})
	}
}
