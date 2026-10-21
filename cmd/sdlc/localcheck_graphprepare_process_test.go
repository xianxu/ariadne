//go:build unix

package main

import (
	"context"
	"path/filepath"
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
				_, err := runAffectedCheck(ctx, d, productionAffectedDeps(d, "main", cfg), cfg, affectedFixtureSuite, sha)
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
