//go:build unix

package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestFleetMergeIntentsCLISignalsJoinAllGitPhases(t *testing.T) {
	isolateFleetIntegrationGit(t)
	binary := buildFleetE2EBinary(t)
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"vantage", "collection", "intents"} {
		for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
			t.Run(phase+"/"+sig.String(), func(t *testing.T) {
				base := canonicalRoot(t.TempDir())
				root := newFleetE2ERepo(t, base, "sample", "")
				line, err := landingLineDir(execGitRunner{}, root)
				if err != nil {
					t.Fatal(err)
				}
				for _, branch := range []string{"000624-first", "000624-second"} {
					runFleetIntegrationGit(t, root, "branch", branch)
					hold, err := landingline.OpenIntents(line).Hold(landingline.Intent{Branch: branch, Head: git(t, root, "rev-parse", "HEAD"), Root: root, PID: 4242, StartedAt: time.Now().UTC()})
					if err != nil {
						t.Fatal(err)
					}
					if err := hold.Release(true); err != nil {
						t.Fatal(err)
					}
				}
				bin := t.TempDir()
				parentFile, childFile, trace := filepath.Join(bin, "parent"), filepath.Join(bin, "child"), filepath.Join(bin, "trace")
				pattern := "*"
				if phase == "collection" {
					pattern = "*'worktree list'*"
				} else if phase == "intents" {
					pattern = "*refs/heads/*"
				}
				// Child deliberately retains both output pipes. Only the real
				// command's group cleanup may make CLI Wait complete normally.
				script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> " + quoteHelperShell(trace) + "\ncase \"$*\" in " + pattern + ") sleep 60 & printf '%s' \"$!\" > " + quoteHelperShell(childFile) + "; printf '%s' \"$$\" > " + quoteHelperShell(parentFile) + "; wait;; *) exec " + quoteHelperShell(realGit) + " \"$@\";; esac\n"
				if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0755); err != nil {
					t.Fatal(err)
				}
				cmd := exec.CommandContext(t.Context(), binary, "fleet", "inventory", "--json", "--path", root)
				cmd.Dir = root
				cmd.Env = append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "WF_SDLC_QUEUE_ROOT="+filepath.Join(bin, "queue"), "COUCH_TREE=")
				cmd.WaitDelay = time.Second
				var stdout, stderr bytes.Buffer
				cmd.Stdout, cmd.Stderr = &stdout, &stderr
				if err := cmd.Start(); err != nil {
					t.Fatal(err)
				}
				done := make(chan error, 1)
				go func() { done <- cmd.Wait() }()
				finished := false
				t.Cleanup(func() {
					if !finished {
						_ = cmd.Process.Kill()
						<-done
					}
				})
				parent, child := awaitLandingGitChild(t, parentFile), awaitLandingGitChild(t, childFile)
				before, err := os.ReadFile(trace)
				if err != nil {
					t.Fatal(err)
				}
				if err := cmd.Process.Signal(sig); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-done:
					finished = true
					if err == nil || cmd.ProcessState.ExitCode() != 1 {
						t.Errorf("native interruption bypassed cancellation owner: %v, %v: %s", err, cmd.ProcessState, &stderr)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("CLI did not join interrupted Git group and output pipes")
				}
				for _, pid := range []int{parent, child} {
					if err := testfix.AwaitExit(pid, time.Second); err != nil {
						t.Errorf("CLI left Git or its pipe-holding child alive: %v", err)
					}
				}
				after, err := os.ReadFile(trace)
				if err != nil || !bytes.Equal(before, after) || stdout.Len() != 0 {
					t.Fatalf("interruption started later probes or published JSON: before=%q after=%q stdout=%q err=%v", before, after, &stdout, err)
				}
			})
		}
	}
}

func TestFleetMergeIntentsCancelsActiveGitProbe(t *testing.T) {
	isolateFleetIntegrationGit(t)
	base := canonicalRoot(t.TempDir())
	root := newFleetE2ERepo(t, base, "sample", "")
	line, err := landingLineDir(execGitRunner{}, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, branch := range []string{"000624-first", "000624-second"} {
		runFleetIntegrationGit(t, root, "branch", branch)
		hold, err := landingline.OpenIntents(line).Hold(landingline.Intent{Branch: branch, Head: git(t, root, "rev-parse", "HEAD"), Root: root, PID: 4242, StartedAt: time.Now().UTC()})
		if err != nil {
			t.Fatal(err)
		}
		if err := hold.Release(true); err != nil {
			t.Fatal(err)
		}
	}
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	ready, probes := filepath.Join(bin, "ready"), filepath.Join(bin, "probes")
	script := "#!/bin/sh\ncase \"$*\" in *refs/heads/*) printf 'probe\\n' >> " + quoteHelperShell(probes) + "; printf '%s' \"$$\" > " + quoteHelperShell(ready) + "; exec sleep 60;; esac\nexec " + quoteHelperShell(realGit) + " \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := observeFleetMergeIntents(ctx, root, ""); done <- err }()
	child := awaitLandingGitChild(t, ready)
	defer func() { _ = syscall.Kill(child, syscall.SIGKILL) }()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("active cancellation lost: %v", err)
		}
	case <-time.After(3 * time.Second):
		_ = syscall.Kill(child, syscall.SIGKILL)
		// Release any later unsafe probe too, so the RED control cannot leak.
		_ = os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\nexit 1\n"), 0755)
		<-done
		t.Fatal("caller cancellation did not interrupt the active production Git probe")
	}
	raw, err := os.ReadFile(probes)
	if err != nil || strings.Count(string(raw), "probe\n") != 1 {
		t.Fatalf("cancelled observer started further probes: %q %v", raw, err)
	}
}
