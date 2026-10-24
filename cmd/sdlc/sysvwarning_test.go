package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/sysvshm"
)

func TestSysVFooterDoesNotChangeOutcomeAndSurvivesReplay(t *testing.T) {
	r := sysvshm.Report{Before: 1, After: 2, Candidates: []sysvshm.Candidate{{Segment: sysvshm.Segment{ID: 7, Creator: 42, Size: 56}}}}
	var out bytes.Buffer
	writeSysVReport(&out, "suite", "head", r)
	for _, want := range []string{"new segments appeared during this run, not attributed", "id=7", "creator=42", "command=\"unknown\""} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q: %s", want, &out)
		}
	}
	rec := localRunRecord{Name: "suite", SHA: "head", Exit: 0, Tail: strings.Split(strings.TrimSpace(out.String()), "\n"), Scope: strings.Repeat("long", 80)}
	if !rec.passed() || rec.state() != "success" || !strings.HasPrefix(rec.statusDescription(), "SysV warning") {
		t.Fatalf("%+v %s", rec, rec.statusDescription())
	}
	writeSysVReport(&out, "suite", "head", sysvshm.Report{})
	rec.Tail = strings.Split(strings.TrimSpace(out.String()), "\n")
	if strings.Contains(rec.statusDescription(), "SysV warning") {
		t.Fatal("earlier child-looking footer overrode final clean footer")
	}
}

func TestSysVUnavailableNeverDisplaysAnEmptyCount(t *testing.T) {
	var out bytes.Buffer
	writeSysVReport(&out, "suite", "head", sysvshm.Report{Unavailable: "denied"})
	if !strings.Contains(out.String(), "owner-before=unknown owner-after=unknown") {
		t.Fatalf("unknown observation looked empty: %s", &out)
	}
}

type fakeSysVObserver struct {
	beforeErr, afterErr string
	onBefore            func()
	onAfter             func()
	onContext           func(context.Context)
	segments            []sysvshm.Segment
	captures            int
	tablePath           string
}

func (f *fakeSysVObserver) Snapshot(ctx context.Context) sysvshm.Snapshot {
	if f.onContext != nil {
		f.onContext(ctx)
	}
	f.captures++
	if f.captures == 1 {
		if f.onBefore != nil {
			f.onBefore()
		}
		if f.beforeErr != "" {
			return sysvshm.Snapshot{Unavailable: f.beforeErr}
		}
	} else {
		if f.onAfter != nil {
			f.onAfter()
		}
		if f.afterErr != "" {
			return sysvshm.Snapshot{Unavailable: f.afterErr}
		}
	}
	if f.tablePath != "" {
		data, err := os.ReadFile(f.tablePath)
		if err != nil {
			return sysvshm.Snapshot{Unavailable: err.Error()}
		}
		return sysvshm.ParseLinux(data, os.Geteuid())
	}
	if f.captures == 1 {
		return sysvshm.Snapshot{}
	}
	return sysvshm.Snapshot{Segments: append([]sysvshm.Segment(nil), f.segments...)}
}
func TestSysVManagedRunWarningIsHashedWithoutChangingPass(t *testing.T) {
	d, sha, _, stderr := localRunRepo(t, "acquired")
	table := filepath.Join(t.TempDir(), "shm-table")
	header := "key shmid perms size cpid lpid nattch uid gid cuid cgid atime dtime ctime rss swap\n"
	if err := os.WriteFile(table, []byte(header), 0600); err != nil {
		t.Fatal(err)
	}
	f := &fakeSysVObserver{tablePath: table}
	row := fmt.Sprintf("123 7 600 56 42 0 0 %d 20 %d 20 0 0 12345 0 0\n", os.Geteuid(), os.Geteuid())
	d.sysvObserver = f
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "zellij", Command: []string{"sh", "-c", `printf "%s" "$2" >> "$1"`, "fake-allocate", table, row}, TimeoutMinutes: 1}, sha)
	if err != nil || !rec.passed() {
		t.Fatalf("%+v %v %s", rec, err, stderr)
	}
	if f.captures != 2 {
		t.Fatalf("captures=%d", f.captures)
	}
	if !strings.HasPrefix(rec.statusDescription(), "SysV warning") || !strings.Contains(stderr.String(), "not attributed") {
		t.Fatalf("%+v %s", rec, stderr)
	}
	hash, tail, err := hashAndTail(rec.LogPath, localCheckTailLines)
	if err != nil || hash != rec.LogSHA256 || !hasSysVWarning(tail, rec.Name, sha) {
		t.Fatalf("hash=%s saved=%s err=%v tail=%v", hash, rec.LogSHA256, err, tail)
	}
}

func TestSysVReplayAndParserDoNotInventFailure(t *testing.T) {
	var out bytes.Buffer
	writeSysVReport(&out, "suite", "head", sysvshm.Report{Candidates: []sysvshm.Candidate{{Segment: sysvshm.Segment{ID: 7, Creator: 42}}}})
	raw := out.String()
	facts := parseSuiteLogFacts(raw)
	if facts.Failed || len(facts.Failures) != 0 || len(facts.Unattributed) != 0 || len(facts.Warnings) != 0 {
		t.Fatalf("footer affected failure proof: %+v", facts)
	}
	seconds := 1.0
	attempt := SuiteAttempt{ID: "saved", Repo: "o/r", SHA: "head", Check: "suite", WallSeconds: &seconds, LogSHA256: "digest", Exit: 0}
	tail := strings.Split(strings.TrimSpace(raw), "\n")
	rec, err := savedAttemptRecord(attempt, "digest", tail, "o/r", "head", "suite", "host")
	if err != nil || !strings.HasPrefix(rec.statusDescription(), "SysV warning") {
		t.Fatalf("%+v %v", rec, err)
	}
	if _, err := savedAttemptRecord(attempt, "mutated", tail, "o/r", "head", "suite", "host"); err == nil {
		t.Fatal("mutated log accepted")
	}
	rec, err = savedAttemptRecord(attempt, "digest", []string{"old log"}, "o/r", "head", "suite", "host")
	if err != nil || strings.Contains(rec.statusDescription(), "SysV") {
		t.Fatalf("old log %+v %v", rec, err)
	}
}

func TestSysVTerminalLifecycleAndCountedBatchCost(t *testing.T) {
	for _, tc := range []struct {
		name     string
		commands [][]string
		timeout  time.Duration
		cleanup  bool
	}{
		{"pass", [][]string{{"sh", "-c", "echo ok"}}, time.Second, false},
		{"failed", [][]string{{"sh", "-c", "exit 7"}}, time.Second, false},
		{"never started", [][]string{{"/missing-sysv-test-command"}}, time.Second, false},
		{"timeout", [][]string{{"sh", "-c", "sleep 30"}}, 50 * time.Millisecond, false},
		{"early batch error", [][]string{{"sh", "-c", "exit 7"}, {"sh", "-c", "echo forbidden"}}, time.Second, false},
		{"eight commands", [][]string{{"true"}, {"true"}, {"true"}, {"true"}, {"true"}, {"true"}, {"true"}, {"true"}}, time.Second, false},
		{"cleanup unknown", [][]string{{"true"}}, time.Second, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d, sha, _, _ := localRunRepo(t, "acquired")
			t.Setenv("WF_TEST_SCRATCH", t.TempDir())
			wt := d.Root
			if tc.cleanup {
				var cleanup func()
				var setupErr error
				wt, cleanup, setupErr = detachedWorktree(d.Git, sha)
				if setupErr != nil {
					t.Fatal(setupErr)
				}
				t.Cleanup(cleanup)
			}
			f := &fakeSysVObserver{}
			d.sysvObserver = f
			if tc.cleanup {
				d.StrictGroupCleanup = true
				d.runnerCleanup = func(*exec.Cmd) func() error {
					return func() error {
						if f.captures != 1 {
							t.Fatal("after-snapshot preceded terminal cleanup")
						}
						return &runnerCleanupUnconfirmed{Detail: "fixture unknown"}
					}
				}
			}
			logDir := t.TempDir()
			log := filepath.Join(logDir, "run.log")
			run, err := runUnderResourceBatch(context.Background(), d, wt, filepath.Join(d.Root, "scripts/test-lock.py"), logDir, "zellij", sha, 1, suitePrioritySubject{}, tc.commands, log, tc.timeout, "")
			if f.captures != 2 {
				t.Fatalf("captures=%d run=%+v err=%v", f.captures, run, err)
			}
			b, readErr := os.ReadFile(log)
			if readErr != nil || !strings.Contains(string(b), "state=observed") {
				t.Fatalf("log %q %v", b, readErr)
			}
			if tc.cleanup {
				var unknown *runnerCleanupUnconfirmed
				if !errors.As(err, &unknown) {
					t.Fatalf("%v", err)
				}
			}
			if tc.name == "timeout" && !run.TimedOut {
				t.Fatalf("timeout changed %+v %v", run, err)
			}
			if tc.name == "failed" && run.Exit != 7 {
				t.Fatalf("exit changed %+v %v", run, err)
			}
		})
	}
}

func TestSysVRunnerSkipsHostObservation(t *testing.T) {
	d, sha, _, _ := localRunRepo(t, "must never acquire")
	f := &fakeSysVObserver{}
	d.sysvObserver = f
	rec, err := runLocalCheck(context.Background(), d, localCheck{Name: "vm", RunScope: "landing", Resource: "runner", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
	if err != nil || !rec.passed() || f.captures != 0 {
		t.Fatalf("%+v %v captures=%d", rec, err, f.captures)
	}
}

func TestSysVUnavailableIsAdvisoryAndCancellationStillObserves(t *testing.T) {
	for _, which := range []string{"before", "after", "cancel"} {
		t.Run(which, func(t *testing.T) {
			d, sha, _, stderr := localRunRepo(t, "acquired")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			f := &fakeSysVObserver{}
			switch which {
			case "before":
				f.beforeErr = "denied"
			case "after":
				f.afterErr = "truncated"
			case "cancel":
				f.onBefore = cancel
			}
			d.sysvObserver = f
			rec, err := runLocalCheck(ctx, d, localCheck{Name: "zellij", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
			if f.captures != 2 {
				t.Fatalf("captures=%d err=%v", f.captures, err)
			}
			if which == "cancel" {
				var interrupted *queueInterruptedError
				if !errors.As(err, &interrupted) || !strings.Contains(stderr.String(), "state=observed") {
					t.Fatalf("%v %s", err, stderr)
				}
			} else if err != nil || !rec.passed() || !strings.Contains(stderr.String(), "observation unavailable") {
				t.Fatalf("%+v %v %s", rec, err, stderr)
			}
		})
	}
}

func TestSysVNoObservationBeforeAdmissionOrWritableLog(t *testing.T) {
	for _, which := range []string{"admission", "log"} {
		t.Run(which, func(t *testing.T) {
			d, sha, _, _ := localRunRepo(t, "acquired")
			f := &fakeSysVObserver{}
			d.sysvObserver = f
			helper := filepath.Join(d.Root, "scripts/test-lock.py")
			logDir := t.TempDir()
			log := filepath.Join(logDir, "run.log")
			if which == "admission" {
				if err := os.WriteFile(helper, []byte(fakeSuiteLockHelper("import sys\nsys.exit(3)\n")), 0755); err != nil {
					t.Fatal(err)
				}
			} else {
				blocked := filepath.Join(logDir, "blocked")
				if err := os.WriteFile(blocked, []byte("file"), 0600); err != nil {
					t.Fatal(err)
				}
				log = filepath.Join(blocked, "run.log")
			}
			_, err := runUnderResourceBatch(context.Background(), d, d.Root, helper, logDir, "zellij", sha, 1, suitePrioritySubject{}, [][]string{{"true"}}, log, time.Second, "")
			if err == nil || f.captures != 0 {
				t.Fatalf("%v captures=%d", err, f.captures)
			}
		})
	}
}

func TestSysVObservationPreservesCallerValuesWithIndependentBound(t *testing.T) {
	type contextKey struct{}
	for _, interrupted := range []bool{false, true} {
		t.Run(fmt.Sprint(interrupted), func(t *testing.T) {
			d, sha, _, stderr := localRunRepo(t, "acquired")
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "caller"))
			defer cancel()
			observations := 0
			f := &fakeSysVObserver{onContext: func(observed context.Context) {
				observations++
				if observed.Value(contextKey{}) != "caller" {
					t.Error("observation discarded caller values")
				}
				if err := observed.Err(); err != nil {
					t.Errorf("observation inherited cancellation: %v", err)
				}
				deadline, ok := observed.Deadline()
				if !ok || time.Until(deadline) > 2*time.Second {
					t.Error("observation lacks its independent two-second ceiling")
				}
			}}
			if interrupted {
				f.onBefore = cancel
			}
			d.sysvObserver = f
			rec, err := runLocalCheck(ctx, d, localCheck{Name: "zellij", Command: []string{"true"}, TimeoutMinutes: 1}, sha)
			if f.captures != 2 || observations != f.captures {
				t.Fatalf("observations=%d captures=%d", observations, f.captures)
			}
			if interrupted {
				var stopped *queueInterruptedError
				if !errors.As(err, &stopped) {
					t.Fatalf("%v %s", err, stderr)
				}
			} else if err != nil || !rec.passed() {
				t.Fatalf("%+v %v %s", rec, err, stderr)
			}
		})
	}
}

// The test advances only the execution deadline, without depending on process
// launch latency or sleeping under fleet load.
type sysVExecutionClock struct {
	context.Context
	expired atomic.Bool
}

func (c *sysVExecutionClock) Err() error {
	if c.expired.Load() {
		return context.DeadlineExceeded
	}
	return c.Context.Err()
}

func TestSysVDiagnosticsCannotConsumeExecutionBudget(t *testing.T) {
	for _, wantExit := range []int{0, 7} {
		t.Run(fmt.Sprint(wantExit), func(t *testing.T) {
			d, sha, _, _ := localRunRepo(t, "acquired")
			var clock *sysVExecutionClock
			d.executionTimeout = func(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) {
				child, cancel := context.WithCancel(ctx)
				clock = &sysVExecutionClock{Context: child}
				return clock, cancel
			}
			f := &fakeSysVObserver{
				onBefore: func() {
					if clock != nil {
						t.Error("execution timer began before before-observation finished")
					}
				},
				onAfter: func() {
					if clock == nil {
						t.Fatal("execution context absent")
					}
					clock.expired.Store(true)
				},
			}
			d.sysvObserver = f
			logDir := t.TempDir()
			run, err := runUnderResourceBatch(t.Context(), d, d.Root, filepath.Join(d.Root, "scripts/test-lock.py"), logDir, "zellij", sha, 1, suitePrioritySubject{}, [][]string{{"sh", "-c", fmt.Sprintf("exit %d", wantExit)}}, filepath.Join(logDir, "run.log"), time.Minute, "")
			if err != nil || run.TimedOut || run.Exit != wantExit {
				t.Fatalf("diagnostic changed completed success: %+v %v", run, err)
			}

		})
	}
}

func TestSysVCurrentProcessNameIsNotCreatorProof(t *testing.T) {
	var out bytes.Buffer
	writeSysVReport(&out, "suite", "head", sysvshm.Report{Candidates: []sysvshm.Candidate{{Segment: sysvshm.Segment{ID: 7, Creator: os.Getpid()}}}})
	if strings.Contains(out.String(), "unrelated-reused-pid") || strings.Contains(out.String(), "witnessed in run process group") || !strings.Contains(out.String(), "command=\"unknown\"") {
		t.Fatalf("current PID incarnation was mistaken for allocating creator: %s", &out)
	}
}
