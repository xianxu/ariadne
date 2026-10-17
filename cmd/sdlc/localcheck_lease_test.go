package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/processgroup"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestClassifyRunLease(t *testing.T) {
	ours := runLease{Root: "/slot2", Check: "suite"}
	for _, tc := range []struct {
		name  string
		rec   runLease
		held  bool
		check string
		want  leaseVerdict
	}{
		{"unheld is stale, even ours", ours, false, "", leaseStale},
		{"unheld foreign is stale too", runLease{Root: "/slot3", Check: "suite"}, false, "", leaseStale},
		{"held by another checkout", runLease{Root: "/slot3", Check: "suite"}, true, "", leaseForeign},
		{"another checkout, same name", runLease{Root: "/slot3", Check: "suite"}, true, "suite", leaseForeign},
		{"ours, another check named", ours, true, "e2e", leaseOtherCheck},
		{"ours, no name", ours, true, "", leaseTarget},
		{"ours, named", ours, true, "suite", leaseTarget},
	} {
		if got := classifyRunLease(tc.rec, tc.held, "/slot2", tc.check); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}

func leaseFiles(t *testing.T, dir string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	return paths
}

// A lease is never visible unlocked: concurrent cancels from another checkout
// sweep stale leases throughout, and not one live lease disappears.
func TestHoldRunLeaseLockedBeforeVisible(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "runs")
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_, _ = cancelLocalChecks(dir, "/elsewhere", "", &bytes.Buffer{}, time.Second)
				}
			}
		}()
	}
	for i := 0; i < 300; i++ {
		release, err := holdRunLease(dir, runLease{Root: "/slot2", Check: "suite", PID: os.Getpid()})
		if err != nil {
			t.Fatal(err)
		}
		if got := leaseFiles(t, dir); len(got) != 1 {
			close(stop)
			wg.Wait()
			t.Fatalf("iteration %d: live lease swept by a concurrent cancel (leases: %v)", i, got)
		}
		release()
	}
	close(stop)
	wg.Wait()
	if got := leaseFiles(t, dir); len(got) != 0 {
		t.Errorf("released leases remain: %v", got)
	}
	if tmps, _ := filepath.Glob(filepath.Join(dir, ".lease-*")); len(tmps) != 0 {
		t.Errorf("temporary lease names remain: %v", tmps)
	}
}

func TestCancelSweepsStaleLeaseWithoutSignal(t *testing.T) {
	dir := t.TempDir()
	victim := exec.Command("sleep", "30")
	if err := victim.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = victim.Process.Kill(); _ = victim.Wait() }()
	started, err := processStartTime(victim.Process.Pid)
	if err != nil || started == "" {
		t.Fatalf("start time: %q %v", started, err)
	}
	// Ours, naming a live process, but nobody holds the lock: the run is gone.
	raw, _ := json.Marshal(runLease{Root: canonicalRoot(dir), Check: "suite", PID: victim.Process.Pid, PIDStarted: started})
	stale := filepath.Join(dir, "stale.json")
	if err := os.WriteFile(stale, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	n, err := cancelLocalChecks(dir, dir, "", &out, time.Second)
	if err != nil || n != 0 {
		t.Fatalf("cancel: %d %v\n%s", n, err, out.String())
	}
	if _, err := os.Stat(stale); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("stale lease not swept: %v", err)
	}
	if err := victim.Process.Signal(syscall.Signal(0)); err != nil {
		t.Errorf("a stale lease's pid was signalled: %v", err)
	}
}

// A held lease whose pid now belongs to another process (start time differs)
// is not signalled; with the matching start time it is.
func TestCancelChecksProcessStartTime(t *testing.T) {
	dir := t.TempDir()
	victim := exec.Command("sleep", "30")
	if err := victim.Start(); err != nil {
		t.Fatal(err)
	}
	exited := make(chan struct{})
	go func() { _ = victim.Wait(); close(exited) }()
	defer func() { _ = victim.Process.Kill(); <-exited }()
	started, err := processStartTime(victim.Process.Pid)
	if err != nil || started == "" {
		t.Fatalf("start time: %q %v", started, err)
	}
	root := canonicalRoot(dir)
	release, err := holdRunLease(filepath.Join(dir, "runs"), runLease{Root: root, Check: "suite", PID: victim.Process.Pid, PIDStarted: "0.000000"})
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if n, _ := cancelLocalChecks(filepath.Join(dir, "runs"), dir, "", &out, 200*time.Millisecond); n != 0 || !strings.Contains(out.String(), "no longer that run") {
		t.Errorf("reused pid: stopped %d\n%s", n, out.String())
	}
	select {
	case <-exited:
		t.Fatal("a reused pid was signalled")
	case <-time.After(100 * time.Millisecond):
	}
	release()
	release, err = holdRunLease(filepath.Join(dir, "runs"), runLease{Root: root, Check: "suite", PID: victim.Process.Pid, PIDStarted: started})
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	// The lease stays held by this test, so cancel reports the run as not unwound.
	if _, err := cancelLocalChecks(filepath.Join(dir, "runs"), dir, "", &bytes.Buffer{}, 200*time.Millisecond); err == nil || !strings.Contains(err.Error(), "still running") {
		t.Errorf("held lease after signal: %v", err)
	}
	select {
	case <-exited:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("matching pid was not signalled")
	}
}

// TestLocalCheckLeaseRunChild is one slot's `local-check run` for
// TestLocalCheckCancelOnlyOwnRun: a real runLocalCheck from SDLC_LEASE_ROOT whose
// check records its pid and sleeps.
func TestLocalCheckLeaseRunChild(t *testing.T) {
	root := os.Getenv("SDLC_LEASE_ROOT")
	if root == "" {
		return
	}
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput()
		return string(out), err
	}
	d := localRunDeps{GH: newCIFakeGH(), Repo: "o/r", Root: root, Git: git, Stderr: os.Stderr, minute: time.Minute}
	check := localCheck{Name: "suite", Command: []string{"sh", "-c", "echo $$ > \"$SDLC_LEASE_CHECK_PID\"; exec sleep 60"}, TimeoutMinutes: 1, Weight: 1}
	_, err := runLocalCheck(context.Background(), d, check, os.Getenv("SDLC_LEASE_SHA"))
	fmt.Fprintln(os.Stderr, "child run returned:", err)
	if err != nil {
		os.Exit(3)
	}
	os.Exit(0)
}

// Two slots' runs coexist in one clone's shared lease directory; cancel from
// one checkout stops only its own run, and the other keeps running.
func TestLocalCheckCancelOnlyOwnRun(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "skipped")
	rootA := d.Root
	rootB := filepath.Join(t.TempDir(), "slotB")
	if out, err := d.Git("worktree", "add", "-q", "--detach", rootB, sha); err != nil {
		t.Fatalf("worktree: %s %v", out, err)
	}
	logDir, err := localCheckLogDir(d)
	if err != nil {
		t.Fatal(err)
	}
	runs := localCheckRunsDir(logDir)
	type child struct {
		cmd     *exec.Cmd
		done    chan error
		pidFile string
		out     *bytes.Buffer
	}
	start := func(root string) *child {
		c := &child{done: make(chan error, 1), pidFile: filepath.Join(t.TempDir(), "check.pid"), out: &bytes.Buffer{}}
		c.cmd = exec.Command(os.Args[0], "-test.run=^TestLocalCheckLeaseRunChild$")
		processgroup.Configure(c.cmd)
		c.cmd.Dir = root
		c.cmd.WaitDelay = time.Second
		c.cmd.Env = append(os.Environ(), "SDLC_LEASE_ROOT="+root, "SDLC_LEASE_SHA="+sha, "SDLC_LEASE_CHECK_PID="+c.pidFile)
		c.cmd.Stdout, c.cmd.Stderr = c.out, c.out
		if err := c.cmd.Start(); err != nil {
			t.Fatal(err)
		}
		go func() { c.done <- c.cmd.Wait() }()
		t.Cleanup(func() {
			if raw, _ := os.ReadFile(c.pidFile); len(raw) > 0 {
				if pid, _ := strconv.Atoi(strings.TrimSpace(string(raw))); pid > 0 {
					_ = syscall.Kill(pid, syscall.SIGKILL)
				}
			}
			_ = processgroup.Terminate(c.cmd, true)
			select {
			case <-c.done:
			case <-time.After(testfix.HangCeiling):
			}
		})
		return c
	}
	a, b := start(rootA), start(rootB)
	checkPID := func(c *child) int {
		raw, _ := os.ReadFile(c.pidFile)
		pid, _ := strconv.Atoi(strings.TrimSpace(string(raw)))
		return pid
	}
	deadline := time.Now().Add(testfix.HangCeiling)
	for len(leaseFiles(t, runs)) < 2 || checkPID(a) == 0 || checkPID(b) == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("runs did not start:\nA: %s\nB: %s", a.out.String(), b.out.String())
		}
		time.Sleep(20 * time.Millisecond)
	}
	checkA, checkB := checkPID(a), checkPID(b)

	var out bytes.Buffer
	n, err := cancelLocalChecks(runs, rootA, "", &out, testfix.HangCeiling)
	if err != nil || n != 1 {
		t.Fatalf("cancel from A: stopped %d, %v\n%s", n, err, out.String())
	}
	if !strings.Contains(out.String(), "1 run(s) of other checkouts left running") {
		t.Errorf("cancel did not report B as another checkout's run:\n%s", out.String())
	}
	select {
	case err := <-a.done:
		if err == nil || !strings.Contains(a.out.String(), "interrupted") {
			t.Errorf("A did not stop on the managed interruption path: %v\n%s", err, a.out.String())
		}
	case <-time.After(testfix.HangCeiling):
		t.Fatal("A's run kept running after its cancel")
	}
	if err := testfix.AwaitExit(checkA, time.Second); err != nil {
		t.Errorf("A's check survived its run: %v", err)
	}
	select {
	case err := <-b.done:
		t.Fatalf("cancel from A stopped B: %v\n%s", err, b.out.String())
	default:
	}
	if err := syscall.Kill(checkB, 0); err != nil {
		t.Errorf("B's check is not running: %v", err)
	}
	if got := leaseFiles(t, runs); len(got) != 1 {
		t.Errorf("want only B's lease, got %v", got)
	}

	// B cancels its own run the same way.
	out.Reset()
	if n, err := cancelLocalChecks(runs, rootB, "", &out, testfix.HangCeiling); err != nil || n != 1 {
		t.Fatalf("cancel from B: stopped %d, %v\n%s", n, err, out.String())
	}
	select {
	case <-b.done:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("B's run kept running after its own cancel")
	}
	if got := leaseFiles(t, runs); len(got) != 0 {
		t.Errorf("leases remain after both runs: %v", got)
	}
}
