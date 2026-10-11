// Package testgate bounds how many heavy Go test binaries run at once on this
// machine (#324). #319's lock covers `make test` only; agents iterate with
// targeted `go test` runs, and four slots doing that at once pushed the load past
// 7× the core count. An environment cap cannot reach every such run (it is fixed
// when a harness launches, and differs per harness), so the gate sits in the test
// binary itself: a heavy package calls Enter first in its TestMain.
//
// Enter waits while a full suite holds #319's test-suite.lock, then takes one of N
// machine-wide slots, flock(2) on test-gate.<k>.lock beside it. The descriptor is close-on-exec and held for the life of
// the process, so the kernel releases the slot however the binary exits. Each slot
// file holds its holder's record, rewritten on every acquire. Enter also caps
// -test.parallel unless the run set it, and bounds its wait by the run's
// -test.timeout (see Enter).
//
// Environment:
//
//	WF_TEST_LOCK=off          skip the gate, loudly (the same switch as #319's lock)
//	WF_TEST_LOCK_TIMEOUT=<s>  give up after s seconds of waiting
//	WF_TEST_GATE_SLOTS=<n>    concurrent heavy test binaries machine-wide (default NCPU/4)
//	WF_TEST_PARALLEL=<n>      the -test.parallel cap (default 2)
//	WF_TEST_GATE_DIR=<dir>    where the slot files live (tests)
//
// Under `make test` (WF_TEST_LOCK_HELD_BY set) the gate stands aside: the full
// suite holds #319's lock and sizes its own shards.
//
// In package-list mode `go test` drops a passing binary's output, so the wait lines
// show only with -v, on a failure, or in local-directory mode. The slot files and
// test-suite.lock name the holders of a run that looks hung.
package testgate

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const reportEvery = 30 * time.Second

// held keeps the slot's descriptor open, and so the flock, until the process exits.
var held *os.File

type config struct {
	off, bypass bool
	timeout     time.Duration
	slots       int
	parallel    string
	dir         string
	suiteLock   string        // #319's full-suite lock
	budget      time.Duration // the most this run may wait before going ahead ungated; 0 = unbounded
}

func configFromEnv() config {
	sw := strings.ToLower(os.Getenv("WF_TEST_LOCK"))
	c := config{
		off:      sw == "off" || sw == "0" || sw == "no",
		bypass:   os.Getenv("WF_TEST_LOCK_HELD_BY") != "",
		slots:    max(1, runtime.NumCPU()/4),
		parallel: "2",
		dir:      os.Getenv("WF_TEST_GATE_DIR"),
	}
	if s, err := strconv.ParseFloat(os.Getenv("WF_TEST_LOCK_TIMEOUT"), 64); err == nil && s > 0 {
		c.timeout = time.Duration(s * float64(time.Second))
	}
	if n, err := strconv.Atoi(os.Getenv("WF_TEST_GATE_SLOTS")); err == nil && n > 0 {
		c.slots = n
	}
	if p := os.Getenv("WF_TEST_PARALLEL"); p != "" {
		c.parallel = p
	}
	if c.dir == "" {
		state := os.Getenv("XDG_STATE_HOME")
		if state == "" {
			home, _ := os.UserHomeDir()
			state = filepath.Join(home, ".local", "state")
		}
		c.dir = filepath.Join(state, "ariadne")
	}
	c.suiteLock = os.Getenv("WF_TEST_LOCK_FILE") // test-lock.py's own override
	if c.suiteLock == "" {
		c.suiteLock = filepath.Join(c.dir, "test-suite.lock")
	}
	return c
}

// Enter gates this test binary; call it first in TestMain. It returns once a slot
// is held (or the gate stands aside), and exits the process on a timed-out wait.
//
// The wait spends the run's -test.timeout: cmd/go kills a test binary at that
// timeout plus a minute from its start, however long it queued. So the gate waits
// at most a quarter of it and then goes ahead ungated, and it shrinks -test.timeout
// by the time waited, so the binary's own alarm, which names the running tests,
// fires before that kill.
func Enter() {
	c := configFromEnv()
	if !c.bypass && !c.off {
		capParallel(c.parallel)
	}
	runTimeout := testTimeout()
	c.budget = runTimeout / 4
	waited, err := enter(c, os.Stderr, time.Sleep)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[test-gate] %v\n", err)
		os.Exit(1)
	}
	if runTimeout > 0 && waited > 0 {
		_ = flag.Set("test.timeout", (runTimeout - waited).String())
	}
}

// testTimeout is the run's -test.timeout, 0 when unset or disabled.
func testTimeout() time.Duration {
	f := flag.Lookup("test.timeout")
	if f == nil {
		return 0
	}
	d, _ := f.Value.(flag.Getter).Get().(time.Duration)
	return d
}

// capParallel sets -test.parallel unless the command line (or GOFLAGS) set it.
func capParallel(n string) {
	if !flag.Parsed() {
		flag.Parse()
	}
	set := false
	flag.Visit(func(f *flag.Flag) { set = set || f.Name == "test.parallel" })
	if !set && flag.Lookup("test.parallel") != nil {
		_ = flag.Set("test.parallel", n)
	}
}

// enter returns how long it waited.
func enter(c config, stderr io.Writer, sleep func(time.Duration)) (time.Duration, error) {
	switch {
	case c.bypass:
		return 0, nil
	case c.off:
		fmt.Fprintln(stderr, "[test-gate] WF_TEST_LOCK=off: NOT taking a machine-wide test slot; this run adds to the load unbounded")
		return 0, nil
	}
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return 0, fmt.Errorf("cannot create %s: %v", c.dir, err)
	}
	start := time.Now()
	var last time.Time
	for {
		f, holders, err := tryAcquire(c)
		if err != nil {
			return time.Since(start), err
		}
		if f != nil {
			held = f
			if !last.IsZero() {
				fmt.Fprintf(stderr, "[test-gate] slot acquired after %ds\n", int(time.Since(start).Seconds()))
			}
			return time.Since(start), nil
		}
		waited := time.Since(start)
		if last.IsZero() || time.Since(last) >= reportEvery {
			fmt.Fprintf(stderr, "[test-gate] waiting %ds for one of %d test slots, or a full suite to finish (%s); held by:\n%s",
				int(waited.Seconds()), c.slots, c.dir, holders)
			last = time.Now()
		}
		if c.timeout > 0 && waited >= c.timeout {
			return waited, fmt.Errorf("gave up after %ds waiting for a test slot — a lock wait, not a test failure", int(waited.Seconds()))
		}
		if c.budget > 0 && waited >= c.budget {
			fmt.Fprintf(stderr, "[test-gate] going ahead WITHOUT a slot after %ds: a quarter of this run's -test.timeout is the most it may queue\n", int(waited.Seconds()))
			return waited, nil
		}
		sleep(time.Second)
	}
}

// tryAcquire takes the first free slot without blocking. With none free, or while a
// full suite runs, it returns the holders' records, one line each, for the wait
// message.
func tryAcquire(c config) (*os.File, string, error) {
	if suite, err := suiteRunning(c.suiteLock); err != nil || suite != "" {
		return nil, suite, err
	}
	var holders strings.Builder
	for k := 0; k < c.slots; k++ {
		p := filepath.Join(c.dir, fmt.Sprintf("test-gate.%d.lock", k))
		f, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE, 0o644)
		if err != nil {
			return nil, "", fmt.Errorf("cannot open %s: %v", p, err)
		}
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return f, "", writeRecord(f)
		}
		rec, _ := io.ReadAll(f)
		f.Close()
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, "", fmt.Errorf("flock %s: %v", p, err)
		}
		fmt.Fprintf(&holders, "  - %s\n", describe(rec))
	}
	return nil, holders.String(), nil
}

// suiteRunning reports a full suite holding #319's lock, which already saturates the
// machine: a gated binary waits it out. It probes with a shared lock and drops it at
// once, so a stream of targeted runs can never starve a waiting suite.
func suiteRunning(path string) (string, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0o644)
	if err != nil {
		return "", fmt.Errorf("cannot open %s: %v", path, err)
	}
	defer f.Close()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
	switch {
	case err == nil:
		return "", syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	case errors.Is(err, syscall.EWOULDBLOCK):
		rec, _ := io.ReadAll(f)
		return "  - " + describe(rec) + "\n", nil
	default:
		return "", fmt.Errorf("flock %s: %v", path, err)
	}
}

// record is a slot's holder; a full suite's record (scripts/test-lock.py) carries
// worktree instead of dir.
type record struct {
	Repo     string `json:"repo,omitempty"` // a suite's; a slot's dir already names it
	Dir      string `json:"dir"`
	Worktree string `json:"worktree,omitempty"`
	PID      int    `json:"pid"`
	Started  string `json:"started"`
}

func writeRecord(f *os.File) error {
	dir, _ := os.Getwd() // the package directory: go test runs a binary there
	b, _ := json.Marshal(record{Dir: dir, PID: os.Getpid(), Started: time.Now().Format(time.RFC3339)})
	if err := f.Truncate(0); err != nil {
		return err
	}
	_, err := f.WriteAt(b, 0)
	return err
}

func describe(b []byte) string {
	var r record
	if json.Unmarshal(b, &r) != nil || r.PID == 0 {
		return "an unidentified holder (record not yet written)"
	}
	if r.Dir == "" {
		return fmt.Sprintf("the full %s suite (make test) in %s, make pid %d, since %s", r.Repo, r.Worktree, r.PID, r.Started)
	}
	return fmt.Sprintf("test binary in %s, pid %d, since %s", r.Dir, r.PID, r.Started)
}
