//go:build unix

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestAffectedGraphPreparationCancellationAndPipeDrain(t *testing.T) {
	for _, mode := range []string{"cancel", "exited-parent"} {
		t.Run(mode, func(t *testing.T) {
			d, gh, change, _, _ := affectedFixture(t)
			sha := change(map[string]string{"leaf/leaf.go": "package leaf\nfunc F() int { return 2 }\n"})
			pidfile := filepath.Join(t.TempDir(), "child.pid")
			script := `sleep 60 & printf '%s' "$!" > "$1"; wait`
			if mode == "exited-parent" {
				script = `sleep 60 & printf '%s' "$!" > "$1"; exit 0`
			}
			cfg := affectedFixtureConfig(t)
			cfg.GraphPrepare = &affected.Preparation{Command: []string{"sh", "-c", script, "fixture", pidfile}, TimeoutMinutes: 30}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() {
				_, err := runAffectedCheck(ctx, d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha, affectedOrdinary)
				done <- err
			}()
			finished := false
			child := 0
			t.Cleanup(func() {
				cancel()
				if child > 0 {
					_ = syscall.Kill(child, syscall.SIGKILL)
				}
				if !finished {
					select {
					case <-done:
					case <-time.After(testfix.HangCeiling):
						t.Error("runner failed to stop")
					}
				}
			})
			child = testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
			if child == 0 {
				t.Fatal("preparation descendant did not start")
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				finished = true
				if mode == "cancel" && err == nil {
					t.Fatal("cancelled command reported success")
				}
				if mode == "exited-parent" && err != nil {
					t.Fatal(err)
				}
			case <-time.After(testfix.HangCeiling):
				t.Fatal("preparation exceeded cancellation/pipe bound")
			}
			if gh.statuses[sha]["sdlc-local/affected"] == "success" {
				t.Fatalf("interrupted preparation vouched for affected: %v", gh.statuses)
			}
			if mode == "cancel" && gh.statuses[sha]["sdlc-local/suite"] == "success" {
				t.Fatalf("cancelled run vouched for suite: %v", gh.statuses)
			}
			if mode == "exited-parent" && gh.statuses[sha]["sdlc-local/suite"] != "success" {
				t.Fatalf("pipe failure lost fallback: %v", gh.statuses)
			}
			if err := testfix.AwaitExit(child, time.Second); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestAffectedGraphPreparationSignalProcess(t *testing.T) {
	if os.Getenv("SDLC_PREP_SIGNAL_CHILD") != "1" {
		return
	}
	if bin := os.Getenv("SDLC_QUEUE_BIN"); bin != "" {
		_ = os.Setenv("PATH", testfix.PrependPath(bin, os.Getenv("PATH")))
	}
	os.Args = []string{"sdlc", "local-check", "run", "--name", "affected"}
	main()
	os.Exit(0)
}

func TestAffectedGraphPreparationCLISignals(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(sig.String(), func(t *testing.T) {
			d, _, _, _ := localRunRepo(t, "acquired")
			temp := t.TempDir()
			prepPIDFile, childPIDFile := filepath.Join(temp, "prep.pid"), filepath.Join(temp, "child.pid")
			treeFile, ghCalls := filepath.Join(temp, "tree"), filepath.Join(temp, "gh-calls")
			script := `printf '%s' "$$" > "$PREP_PID"; pwd > "$PREP_TREE"; sleep 60 & printf '%s' "$!" > "$PREP_CHILD"; wait`
			raw, _ := json.Marshal(map[string]any{"version": 1, "command": []string{"true"}, "graph_prepare": map[string]any{"command": []string{"sh", "-c", script}}})
			writeRepoFile(t, d.Root, affectedConfigPath, string(raw))
			wireLocalCheckSignalCLI(t, d.Root)
			git(t, d.Root, "switch", "-qc", "feature")
			writeRepoFile(t, d.Root, "change.go", "package fixture\n")
			git(t, d.Root, "add", "-A")
			git(t, d.Root, "commit", "-qm", "candidate")
			git(t, d.Root, "push", "-q", "origin", "feature")
			bin := os.Getenv("SDLC_QUEUE_BIN")
			if err := os.WriteFile(filepath.Join(bin, "gh"), []byte("#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$PREP_GH_CALLS\"\nexit 0\n"), 0700); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(os.Args[0], "-test.run=^TestAffectedGraphPreparationSignalProcess$")
			processgroup.Configure(cmd)
			cmd.Dir = d.Root
			cmd.WaitDelay = time.Second
			cmd.Env = append(os.Environ(), "SDLC_PREP_SIGNAL_CHILD=1", "PREP_PID="+prepPIDFile, "PREP_CHILD="+childPIDFile, "PREP_TREE="+treeFile, "PREP_GH_CALLS="+ghCalls)
			var output bytes.Buffer
			cmd.Stdout, cmd.Stderr = &output, &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			finished, prep, child := false, 0, 0
			t.Cleanup(func() {
				if prep > 0 {
					_ = syscall.Kill(-prep, syscall.SIGKILL)
				}
				if child > 0 {
					_ = syscall.Kill(child, syscall.SIGKILL)
				}
				_ = processgroup.Terminate(cmd, true)
				if !finished {
					select {
					case <-done:
					case <-time.After(testfix.HangCeiling):
						t.Error("CLI failed to stop")
					}
				}
			})
			prep = testfix.AwaitPIDFile(prepPIDFile, testfix.HangCeiling, nil)
			child = testfix.AwaitPIDFile(childPIDFile, testfix.HangCeiling, nil)
			if prep == 0 || child == 0 {
				t.Fatal("preparation did not start")
			}
			rawTree, err := os.ReadFile(treeFile)
			if err != nil {
				t.Fatal(err)
			}
			tree := strings.TrimSpace(string(rawTree))
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				finished = true
				if err == nil {
					t.Error("signal reported successful check")
				}
			case <-time.After(testfix.HangCeiling):
				t.Fatal("CLI exceeded shutdown bound")
			}
			for _, pid := range []int{prep, child} {
				if err := testfix.AwaitExit(pid, time.Second); err != nil {
					t.Errorf("stranded preparation: %v", err)
				}
			}
			if exists(tree) {
				t.Errorf("graph checkout survived signal: %s", tree)
			}
			calls, _ := os.ReadFile(ghCalls)
			if strings.Contains(string(calls), "statuses/") {
				t.Errorf("interrupted preparation posted status: %s", calls)
			}
		})
	}
}
