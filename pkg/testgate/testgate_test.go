package testgate

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The helper mode re-execs this binary as a gated test binary, the way a direct
// `go test` would run a heavy package: TestMain calls Enter, then the helper test
// reports what the gate left behind.
func TestMain(m *testing.M) {
	if os.Getenv("TESTGATE_HELPER") == "1" {
		Enter()
	}
	os.Exit(m.Run())
}

func TestHelperReport(t *testing.T) {
	if os.Getenv("TESTGATE_HELPER") != "1" {
		t.Skip("helper process only")
	}
	fmt.Printf("parallel=%s held=%v timeout=%s\n", flag.Lookup("test.parallel").Value, held != nil, flag.Lookup("test.timeout").Value)
	if os.Getenv("TESTGATE_NESTED") == "1" {
		// Re-exec this binary as a test helper, inheriting the environment the
		// way cmd/weave's and judge's subprocess helpers do.
		ctx, cancel := context.WithTimeout(context.Background(), HangCeiling)
		defer cancel()
		child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestHelperReport$", "-test.v")
		child.Env = append(os.Environ(), "TESTGATE_NESTED=")
		out, err := child.CombinedOutput()
		fmt.Printf("nested err=%v\n%s", err, strings.ReplaceAll(string(out), "parallel=", "nested parallel="))
	}
}

func runHelper(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	return runHelperTee(t, dir, env, nil, args...)
}

// runHelperTee also copies the helper's output to tee as it is written.
func runHelperTee(t *testing.T, dir string, env []string, tee io.Writer, args ...string) string {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestHelperReport$", "-test.v"}, args...)...)
	cmd.Env = append(os.Environ(), append([]string{"TESTGATE_HELPER=1", "WF_TEST_GATE_DIR=" + dir,
		"WF_TEST_GATE_SLOTS=1", "WF_TEST_LOCK=", "WF_TEST_LOCK_HELD_BY=", "WF_TEST_GATE_HELD_BY=", "WF_TEST_PARALLEL="}, env...)...)
	var buf bytes.Buffer
	var sink io.Writer = &buf
	if tee != nil {
		sink = io.MultiWriter(&buf, tee)
	}
	cmd.Stdout, cmd.Stderr = sink, sink
	err := cmd.Run()
	out := buf.Bytes()
	if err != nil {
		t.Fatalf("helper: %v\n%s", err, out)
	}
	return string(out)
}

// A direct run of a gated binary takes a slot and gets the parallel cap; a cap the
// run set itself is kept.
func TestDirectRunHonorsCapAndGate(t *testing.T) {
	dir := t.TempDir()
	if out := runHelper(t, dir, nil); !strings.Contains(out, "parallel=2 held=true") {
		t.Fatalf("want the default cap and a held slot:\n%s", out)
	}
	if out := runHelper(t, dir, []string{"WF_TEST_PARALLEL=3"}); !strings.Contains(out, "parallel=3 held=true") {
		t.Fatalf("want WF_TEST_PARALLEL honored:\n%s", out)
	}
	if out := runHelper(t, dir, nil, "-test.parallel=5"); !strings.Contains(out, "parallel=5 held=true") {
		t.Fatalf("an explicit -parallel must win:\n%s", out)
	}
	// Standing aside means no slot and no cap: go's own -parallel default (GOMAXPROCS).
	if out := runHelper(t, dir, []string{"WF_TEST_LOCK_HELD_BY=123", "GOMAXPROCS=7"}); !strings.Contains(out, "parallel=7 held=false") || strings.Contains(out, "[test-gate]") {
		t.Fatalf("under make test the gate stands aside silently:\n%s", out)
	}
	if out := runHelper(t, dir, []string{"WF_TEST_LOCK=off", "GOMAXPROCS=7"}); !strings.Contains(out, "parallel=7 held=false") || !strings.Contains(out, "WF_TEST_LOCK=off: NOT taking") {
		t.Fatalf("the hatch stands aside loudly:\n%s", out)
	}
}

// A gated binary that re-execs itself as a helper must not queue its child
// behind the slot it holds: with no -test.timeout the child's wait is unbounded,
// and with every slot taken it never ends (#330: cmd/weave's
// TestCompileRecoversAfterProcessDeathDuringGeneration hung 22 minutes).
func TestNestedHelperStandsAside(t *testing.T) {
	dir := t.TempDir()
	out := runHelper(t, dir, []string{"TESTGATE_NESTED=1", "GOMAXPROCS=7"})
	if !strings.Contains(out, "parallel=2 held=true") || !strings.Contains(out, "nested err=<nil>") ||
		!strings.Contains(out, "nested parallel=7 held=false") || strings.Count(out, "[test-gate]") != 0 {
		t.Fatalf("the parent holds the slot; its helper child must stand aside silently:\n%s", out)
	}
}

// holdSlot takes slot k from a separate open file description, as another process would.
func holdSlot(t *testing.T, dir string, k int) *os.File {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(dir, fmt.Sprintf("test-gate.%d.lock", k)), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"dir":"/w/pair/cmd","pid":4242,"started":"2026-10-10T18:00:00-07:00"}`)
	t.Cleanup(func() { f.Close() })
	return f
}

func TestEnterTakesAFreeSlot(t *testing.T) {
	dir := t.TempDir()
	holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	if _, err := enter(config{slots: 2, dir: dir}, &stderr, func(time.Duration) { t.Fatal("must not wait") }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Close(); held = nil })
	b, _ := os.ReadFile(filepath.Join(dir, "test-gate.1.lock"))
	if !strings.Contains(string(b), fmt.Sprintf(`"pid":%d`, os.Getpid())) || stderr.Len() != 0 {
		t.Fatalf("slot 1 should carry this process's record, silently: %s / %q", b, stderr.String())
	}
}

func TestEnterWaitsLoudlyThenAcquires(t *testing.T) {
	dir := t.TempDir()
	f := holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	sleeps := 0
	_, err := enter(config{slots: 1, dir: dir}, &stderr, func(time.Duration) { sleeps++; f.Close() })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Close(); held = nil })
	out := stderr.String()
	for _, want := range []string{"waiting 0s for one of 1 test slots (", "test binary in /w/pair/cmd, pid 4242", "slot acquired after"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stderr missing %q:\n%s", want, out)
		}
	}
	if sleeps != 1 {
		t.Fatalf("sleeps = %d, want 1", sleeps)
	}
}

// Running full suites never hold a targeted run back (#341): with both of
// test-lock.py's suite slots held exclusively, as its watchers hold them, the gate
// takes a slot at once and leaves the suite files alone.
func TestEnterIgnoresFullSuites(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"test-suite.lock", "test-suite.lock.1"} {
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { f.Close() })
	}
	var stderr bytes.Buffer
	if _, err := enter(config{slots: 1, dir: dir}, &stderr, func(time.Duration) { t.Fatal("must not wait for a full suite") }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Close(); held = nil })
	if held == nil || stderr.Len() != 0 {
		t.Fatalf("want a gate slot taken silently, got held=%v stderr=%q", held != nil, stderr.String())
	}
}

// The wait spends the run's timeout budget, so past it the run goes ahead loudly
// rather than being killed by cmd/go mid-queue.
func TestEnterGoesAheadPastItsBudget(t *testing.T) {
	dir := t.TempDir()
	holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	waited, err := enter(config{slots: 1, dir: dir, budget: time.Nanosecond}, &stderr, func(time.Duration) {})
	if err != nil || waited <= 0 || held != nil {
		t.Fatalf("want a slotless go-ahead, got waited=%v err=%v held=%v", waited, err, held != nil)
	}
	if !strings.Contains(stderr.String(), "going ahead WITHOUT a slot") {
		t.Fatalf("the go-ahead must be loud:\n%s", stderr.String())
	}
}

// A run that queued gets its own -test.timeout shrunk by the wait.
func TestDirectRunShrinksTimeoutByTheWait(t *testing.T) {
	dir := t.TempDir()
	f := holdSlot(t, dir, 0)
	// Release on the helper's own waiting line, not after a sleep: a helper slow
	// to start under load would find the slot already free (#325).
	r, w := io.Pipe()
	queued := make(chan struct{})
	go func() {
		var seen bytes.Buffer
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			seen.Write(buf[:n])
			if strings.Contains(seen.String(), "[test-gate] waiting") {
				f.Close()
				close(queued)
				_, _ = io.Copy(io.Discard, r)
				return
			}
			if err != nil {
				return
			}
		}
	}()
	out := runHelperTee(t, dir, nil, w, "-test.timeout=1m")
	_ = w.Close()
	select {
	case <-queued:
	default:
		t.Fatalf("the helper never queued:\n%s", out)
	}
	if !strings.Contains(out, "slot acquired after") || strings.Contains(out, "timeout=1m0s") || !strings.Contains(out, "timeout=5") {
		t.Fatalf("want a queued run with a shrunk timeout (5x.xs):\n%s", out)
	}
}

func TestEnterTimesOutAsALockWait(t *testing.T) {
	dir := t.TempDir()
	holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	//lint:allow test-shrunk-deadline — holdSlot already owns the lock; tests an immediate lock-wait refusal without subprocess startup
	_, err := enter(config{slots: 1, dir: dir, timeout: time.Nanosecond}, &stderr, func(time.Duration) {})
	if err == nil || !strings.Contains(err.Error(), "a lock wait, not a test failure") {
		t.Fatalf("want the timeout refusal, got %v", err)
	}
}

func TestEnterOffIsLoudAndTakesNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "absent")
	var stderr bytes.Buffer
	if _, err := enter(config{off: true, slots: 1, dir: dir}, &stderr, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "WF_TEST_LOCK=off: NOT taking") {
		t.Fatalf("the hatch must be loud: %q", stderr.String())
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("the hatch must not create slot files")
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("WF_TEST_LOCK", "OFF")
	t.Setenv("WF_TEST_LOCK_TIMEOUT", "1.5")
	t.Setenv("WF_TEST_GATE_SLOTS", "4")
	t.Setenv("WF_TEST_GATE_DIR", "")
	t.Setenv("XDG_STATE_HOME", "/state")
	c := configFromEnv()
	if !c.off || c.timeout != 1500*time.Millisecond || c.slots != 4 || c.dir != "/state/ariadne" || c.parallel != "2" {
		t.Fatalf("configFromEnv = %+v", c)
	}
}
