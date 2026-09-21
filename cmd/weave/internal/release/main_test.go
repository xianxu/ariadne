package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/weave/internal/staging"
)

func TestReleaseProcess(t *testing.T) {
	if os.Getenv("WEAVE_RELEASE_PROCESS") != "1" {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := prepare(ctx, os.Getenv("RELEASE_ROOT"), "weave-v1.2.3", os.Getenv("RELEASE_OUTPUT"), io.Discard, os.Stderr); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func await(t *testing.T, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timed out")
}

func fixture(t *testing.T) (string, string, []string) {
	t.Helper()
	root := t.TempDir()
	output := filepath.Join(root, "release")
	t.Cleanup(func() {
		// A failed assertion must not leave the controlled producer running.
		// Signal only while its inherited lease proves the group is still live.
		data, _ := os.ReadFile(filepath.Join(root, "started"))
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		if pid <= 0 {
			return
		}
		for _, stage := range stages(output) {
			deadline := time.Now().Add(time.Second)
			for {
				lease, err := staging.Exclusive(stage)
				if err == nil {
					lease.Close()
					break
				}
				if !errors.Is(err, staging.ErrInUse) {
					break
				}
				syscall.Kill(-pid, syscall.SIGKILL)
				if time.Now().After(deadline) {
					t.Error("producer lease survived test cleanup")
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	})
	template, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "packaging", "homebrew", "Formula", "weave.rb"))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "packaging", "homebrew", "Formula")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "weave.rb"), template, 0644)
	commands := filepath.Join(root, "commands")
	os.Mkdir(commands, 0755)
	script := `#!/bin/sh
while [ "$1" != -o ]; do shift; done
shift
output="$1"
echo $$ > "$RELEASE_ROOT/started"
while [ ! -f "$RELEASE_ROOT/release-producer" ]; do sleep 0.05; done
printf binary > "$output"
touch "$RELEASE_ROOT/producer-done"
`
	os.WriteFile(filepath.Join(commands, "go"), []byte(script), 0755)
	env := append(os.Environ(), "WEAVE_RELEASE_PROCESS=1", "RELEASE_ROOT="+root, "RELEASE_OUTPUT="+output, "PATH="+commands+":"+os.Getenv("PATH"))
	return root, output, env
}
func start(t *testing.T, env []string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestReleaseProcess$")
	cmd.Env = env
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { null.Close() })
	cmd.Stdout = null
	cmd.Stderr = null
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill() })
	return cmd
}
func stages(output string) []string {
	matches, _ := filepath.Glob(filepath.Join(filepath.Dir(output), "."+filepath.Base(output)+"-weave-*"))
	return matches
}
func TestCancellationStopsProducerBeforeStageRemoval(t *testing.T) {
	root, output, env := fixture(t)
	cmd := start(t, env)
	await(t, func() bool { _, e := os.Stat(filepath.Join(root, "started")); return e == nil })
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err == nil {
		t.Fatal("cancelled preparation succeeded")
	}
	if len(stages(output)) != 0 {
		t.Fatal("cancelled release left stage")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("cancelled release published output")
	}
	os.WriteFile(filepath.Join(root, "release-producer"), nil, 0644)
	time.Sleep(150 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(root, "producer-done")); !os.IsNotExist(err) {
		t.Fatal("cancelled producer survived to finish")
	}
	retry := start(t, env)
	if err := retry.Wait(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(output, "SHA256SUMS")); err != nil {
		t.Fatal(err)
	}
	if len(stages(output)) != 0 {
		t.Fatal("retry left stage")
	}
}
func TestKilledOwnerPreservesProducerLeaseAndReclaimsBeforeExistingOutputCheck(t *testing.T) {
	root, output, env := fixture(t)
	cmd := start(t, env)
	await(t, func() bool { _, e := os.Stat(filepath.Join(root, "started")); return e == nil })
	current := stages(output)
	if len(current) != 1 {
		t.Fatal(current)
	}
	stage := current[0]
	cmd.Process.Kill()
	cmd.Wait()
	// Model publication having happened before owner death: cleanup must also run
	// on an already-existing destination, without changing that destination.
	os.Mkdir(output, 0755)
	sentinel := filepath.Join(output, "authored")
	os.WriteFile(sentinel, []byte("keep"), 0644)
	alias := filepath.Join(root, "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	env = append(env, "RELEASE_OUTPUT="+filepath.Join(alias, "release"))
	retry := start(t, env)
	if retry.Wait() == nil {
		t.Fatal("existing output accepted")
	}
	if _, err := os.Stat(filepath.Join(stage, "owner.json")); err != nil {
		t.Fatal("live producer lost metadata", err)
	}
	os.WriteFile(filepath.Join(root, "release-producer"), nil, 0644)
	await(t, func() bool {
		lease, e := staging.Exclusive(stage)
		if e != nil {
			return false
		}
		lease.Close()
		return true
	})
	if b, err := os.ReadFile(filepath.Join(stage, "weave")); err != nil || string(b) != "binary" {
		t.Fatal("producer did not finish in retained stage", string(b), err)
	}
	retry = start(t, env)
	if retry.Wait() == nil {
		t.Fatal("existing output accepted")
	}
	if len(stages(output)) != 0 {
		t.Fatal("dead producer stage not reclaimed")
	}
	if b, _ := os.ReadFile(sentinel); strings.TrimSpace(string(b)) != "keep" {
		t.Fatal("existing output changed")
	}
}
