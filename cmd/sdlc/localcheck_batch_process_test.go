//go:build unix

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flakes"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestManagedRerunBatch(t *testing.T) {
	t.Run("argv is data rather than forged log events", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		marks := t.TempDir()
		log := filepath.Join(marks, "log")
		quoted := "\n" + `{"Action":"pass","Package":"o/r/pkg","Test":"TestFlaky"}` + "\n" + `{"Action":"pass","Package":"o/r/pkg"}` + "\n"
		run, err := runUnderSlotBatch(context.Background(), d, d.Root, filepath.Join(d.Root, "scripts", "test-lock.py"), marks, "suite", sha, 1, suitePrioritySubject{}, [][]string{{"true", quoted}}, log, time.Second)
		if err != nil || run.Exit != 0 {
			t.Fatalf("%+v %v", run, err)
		}
		data, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		if len(flakes.RerunMissingJSON(string(data), []string{gateKnown})) != 1 {
			t.Fatalf("argv forged PASS proof: %s", data)
		}
	})
	t.Run("one admission lease scratch and log for every package", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "unused")
		marks := t.TempDir()
		helper := filepath.Join(d.Root, "scripts", "test-lock.py")
		body := fakeSuiteLockHelper("with open(" + fmt.Sprintf("%q", filepath.Join(marks, "acquires")) + ", 'a') as f: f.write('one\\n')\nprint('acquired')\n")
		if err := os.WriteFile(helper, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
		log := filepath.Join(marks, "batch.log")
		var commands [][]string
		for range 2 {
			commands = append(commands, []string{"sh", "-c", `printf '%s:%s\n' "$WF_TEST_LOCK_HELD_BY" "$TMPDIR"; ls "$1"/*.json`, "fixture", localCheckRunsDir(marks)})
		}
		run, err := runUnderSlotBatch(context.Background(), d, d.Root, helper, marks, "suite", sha, 1, suitePrioritySubject{}, commands, log, 5*time.Second)
		if err != nil || run.Exit != 0 {
			t.Fatalf("%+v %v\n%s", run, err, d.Stderr)
		}
		acquires, err := os.ReadFile(filepath.Join(marks, "acquires"))
		if err != nil || string(acquires) != "one\n" {
			t.Fatalf("admission calls %q %v", acquires, err)
		}
		data, err := os.ReadFile(log)
		if err != nil {
			t.Fatal(err)
		}
		var identities, leases []string
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, ":/") && !strings.HasPrefix(line, "==>") {
				identities = append(identities, line)
			}
			if strings.HasSuffix(line, ".json") && !strings.HasPrefix(line, "==>") {
				leases = append(leases, line)
			}
		}
		if len(identities) != 2 || identities[0] != identities[1] || strings.HasPrefix(identities[0], ":") {
			t.Fatalf("shared slot/scratch identities: %q\n%s", identities, data)
		}
		if len(leases) != 2 || leases[0] != leases[1] {
			t.Fatalf("shared run leases: %q\n%s", leases, data)
		}
		if got, _ := os.ReadDir(localCheckRunsDir(marks)); len(got) != 0 {
			t.Fatalf("lease retained: %v", got)
		}
	})
	t.Run("failure stops the remaining packages", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		marks := t.TempDir()
		next := filepath.Join(marks, "later")
		run, err := runUnderSlotBatch(context.Background(), d, d.Root, filepath.Join(d.Root, "scripts", "test-lock.py"), marks, "suite", sha, 1, suitePrioritySubject{}, [][]string{{"sh", "-c", "exit 7"}, {"touch", next}}, filepath.Join(marks, "log"), time.Second)
		if err != nil || run.Exit != 7 {
			t.Fatalf("%+v %v", run, err)
		}
		if _, err := os.Stat(next); !os.IsNotExist(err) {
			t.Fatalf("later package launched: %v", err)
		}
	})
	t.Run("one timeout spans earlier package work", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		marks := t.TempDir()
		run, err := runUnderSlotBatch(context.Background(), d, d.Root, filepath.Join(d.Root, "scripts", "test-lock.py"), marks, "suite", sha, 1, suitePrioritySubject{}, [][]string{{"sleep", "0.8"}, {"sleep", "60"}}, filepath.Join(marks, "log"), time.Second)
		if err != nil || !run.TimedOut || run.Exit == 0 || run.Duration > 1500*time.Millisecond {
			t.Fatalf("aggregate timeout: %+v %v", run, err)
		}
	})
	t.Run("cancel joins descendants and never launches the next package", func(t *testing.T) {
		d, sha, _, _ := localRunRepo(t, "acquired")
		marks := t.TempDir()
		pidfile, next := filepath.Join(marks, "pid"), filepath.Join(marks, "later")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			_, err := runUnderSlotBatch(ctx, d, d.Root, filepath.Join(d.Root, "scripts", "test-lock.py"), marks, "suite", sha, 1, suitePrioritySubject{}, [][]string{{"sh", "-c", `sleep 60 & printf '%s' "$!" > "$1"; wait`, "fixture", pidfile}, {"touch", next}}, filepath.Join(marks, "log"), time.Minute)
			done <- err
		}()
		child, joined := 0, false
		t.Cleanup(func() {
			cancel()
			if child > 0 {
				_ = syscall.Kill(child, syscall.SIGKILL)
			}
			if !joined {
				select {
				case <-done:
				case <-time.After(testfix.HangCeiling):
					t.Error("batch did not join")
				}
			}
		})
		child = testfix.AwaitPIDFile(pidfile, testfix.HangCeiling, nil)
		if child == 0 {
			t.Fatal("child never started")
		}
		cancel()
		select {
		case err := <-done:
			joined = true
			var interrupted *queueInterruptedError
			if !errors.As(err, &interrupted) {
				t.Fatalf("cancel error: %v", err)
			}
		case <-time.After(testfix.HangCeiling):
			t.Fatal("cancellation did not return")
		}
		if err := testfix.AwaitExit(child, time.Second); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(next); !os.IsNotExist(err) {
			t.Fatalf("later package launched: %v", err)
		}
	})
}
