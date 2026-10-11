package testgate

import (
	"bytes"
	"flag"
	"fmt"
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
}

func runHelper(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestHelperReport$", "-test.v"}, args...)...)
	cmd.Env = append(os.Environ(), append([]string{"TESTGATE_HELPER=1", "WF_TEST_GATE_DIR=" + dir,
		"WF_TEST_GATE_SLOTS=1", "WF_TEST_LOCK=", "WF_TEST_LOCK_HELD_BY=", "WF_TEST_PARALLEL=", "WF_TEST_LOCK_FILE="}, env...)...)
	out, err := cmd.CombinedOutput()
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
	if _, err := enter(config{slots: 2, dir: dir, suiteLock: filepath.Join(dir, "test-suite.lock")}, &stderr, func(time.Duration) { t.Fatal("must not wait") }); err != nil {
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
	_, err := enter(config{slots: 1, dir: dir, suiteLock: filepath.Join(dir, "test-suite.lock")}, &stderr, func(time.Duration) { sleeps++; f.Close() })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Close(); held = nil })
	out := stderr.String()
	for _, want := range []string{"waiting 0s for one of 1 test slots, or a full suite to finish", "test binary in /w/pair/cmd, pid 4242", "slot acquired after"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stderr missing %q:\n%s", want, out)
		}
	}
	if sleeps != 1 {
		t.Fatalf("sleeps = %d, want 1", sleeps)
	}
}

// A running full suite (#319's exclusive lock, as test-lock.py's watcher holds it)
// holds every gated binary back, even with slots free; its end releases them.
func TestEnterWaitsOutAFullSuite(t *testing.T) {
	dir := t.TempDir()
	suite := filepath.Join(dir, "test-suite.lock")
	f, err := os.OpenFile(suite, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"repo": "ariadne", "worktree": "/w/ariadne-slot2", "pid": 2608, "holder": 2654, "started": "2026-10-10T18:05:47"}`)
	var stderr bytes.Buffer
	if _, err := enter(config{slots: 3, dir: dir, suiteLock: suite}, &stderr, func(time.Duration) { f.Close() }); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { held.Close(); held = nil })
	if !strings.Contains(stderr.String(), "the full ariadne suite (make test) in /w/ariadne-slot2, make pid 2608") {
		t.Fatalf("the wait must name the suite:\n%s", stderr.String())
	}
	// The probe never keeps the suite lock: a suite can still take it exclusively.
	g, _ := os.OpenFile(suite, os.O_RDWR, 0)
	defer g.Close()
	if err := syscall.Flock(int(g.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("a gated binary must not hold the suite lock: %v", err)
	}
}

// The wait spends the run's timeout budget, so past it the run goes ahead loudly
// rather than being killed by cmd/go mid-queue.
func TestEnterGoesAheadPastItsBudget(t *testing.T) {
	dir := t.TempDir()
	holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	waited, err := enter(config{slots: 1, dir: dir, suiteLock: filepath.Join(dir, "test-suite.lock"), budget: time.Nanosecond}, &stderr, func(time.Duration) {})
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
	go func() { time.Sleep(1500 * time.Millisecond); f.Close() }()
	out := runHelper(t, dir, nil, "-test.timeout=1m")
	if !strings.Contains(out, "slot acquired after") || strings.Contains(out, "timeout=1m0s") || !strings.Contains(out, "timeout=5") {
		t.Fatalf("want a queued run with a shrunk timeout (5x.xs):\n%s", out)
	}
}

func TestEnterTimesOutAsALockWait(t *testing.T) {
	dir := t.TempDir()
	holdSlot(t, dir, 0)
	var stderr bytes.Buffer
	_, err := enter(config{slots: 1, dir: dir, suiteLock: filepath.Join(dir, "test-suite.lock"), timeout: time.Nanosecond}, &stderr, func(time.Duration) {})
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
	t.Setenv("WF_TEST_LOCK_FILE", "")
	c := configFromEnv()
	if !c.off || c.timeout != 1500*time.Millisecond || c.slots != 4 || c.dir != "/state/ariadne" || c.parallel != "2" || c.suiteLock != "/state/ariadne/test-suite.lock" {
		t.Fatalf("configFromEnv = %+v", c)
	}
}
