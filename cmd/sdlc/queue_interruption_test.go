package main

import (
	"context"
	"errors"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/landingline"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/flock"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/queuegate"
)

func TestQueueManagedSuiteFastPauseInterrupts(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "unused")
	d.Stderr = io.Discard
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	t.Setenv("WF_TEST_SCRATCH", filepath.Join(t.TempDir(), "absent"))
	d.normalTempDir = func() (string, error) { return t.TempDir(), nil }
	helper := filepath.Join(t.TempDir(), "helper.py")
	if err := os.WriteFile(helper, []byte("print('acquired')\n"), 0600); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(t.TempDir(), "started")
	childLifetime := filepath.Join(t.TempDir(), "child-lifetime")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	log := filepath.Join(t.TempDir(), "suite.log")
	go func() {
		child := "import fcntl,pathlib,sys,time; f=open(sys.argv[1],'w'); fcntl.flock(f,fcntl.LOCK_EX); pathlib.Path(sys.argv[2]).write_text('ready'); time.sleep(30)"
		_, err := runUnderSlot(ctx, d, d.Root, helper, t.TempDir(), "suite", "head", 1, suitePrioritySubject{}, []string{"sh", "-c", "python3 -c \"$1\" \"$2\" \"$3\" & wait", "fixture", child, childLifetime, marker}, log, time.Minute)
		done <- err
	}()
	joined := false
	defer func() {
		cancel()
		if !joined {
			<-done
		}
	}()
	s := &queuegate.Store{Root: os.Getenv("WF_SDLC_QUEUE_ROOT")}
	// done keeps its result if the suite ends first, so the deferred join
	// still receives it.
	testfix.AwaitCondition(t, "the suite published a running lease", func() bool {
		x, err := s.Snapshot(context.Background())
		_, started := os.Stat(marker)
		return err == nil && len(x.Active) == 1 && started == nil
	}, []chan error{done})
	if err := s.Pause(context.Background(), "fast", "fixture drain", true); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		joined = true
		var interrupted *queueInterruptedError
		if !errors.As(err, &interrupted) {
			t.Fatalf("not interrupted: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("managed suite ignored fast pause")
	}
	x, err := s.Snapshot(context.Background())
	if err != nil || len(x.Active) != 0 {
		t.Fatalf("managed cleanup left active work: %+v %v", x, err)
	}
	childLock, err := os.OpenFile(childLifetime, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer childLock.Close()
	if err := flock.Exclusive(childLock, true); err != nil {
		t.Fatalf("managed descendant still holds its lifetime resource: %v", err)
	}
	rows, _ := filepath.Glob(filepath.Join(s.Root, "interrupted", "*.json"))
	if len(rows) != 1 {
		t.Fatalf("interruption evidence absent: %v", rows)
	}
}

func TestQueueInterruptedCheckKeepsDiagnosticsAndAllowsSameHeadRetry(t *testing.T) {
	d, sha, gh, _ := localRunRepo(t, "acquired")
	d.Stderr = io.Discard
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	t.Setenv("WF_TEST_SCRATCH", filepath.Join(t.TempDir(), "absent"))
	normal := t.TempDir()
	d.normalTempDir = func() (string, error) { return normal, nil }
	marker := filepath.Join(t.TempDir(), "started")
	check := localCheck{Name: suiteCheckName, Command: []string{"sh", "-c", "echo diagnostic; echo begun > \"$1\"; sleep 30", "fixture", marker}, TimeoutMinutes: 60}
	// rec is written before done is sent, so a receive from done orders it.
	var rec localRunRecord
	done := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		var err error
		rec, err = runLocalCheck(ctx, d, check, sha)
		done <- err
	}()
	joined := false
	defer func() {
		cancel()
		if !joined {
			<-done
		}
	}()
	testfix.AwaitCondition(t, "the suite reached its execution barrier", func() bool {
		_, err := os.Stat(marker)
		return err == nil
	}, []chan error{done})
	s := &queuegate.Store{Root: os.Getenv("WF_SDLC_QUEUE_ROOT")}
	if err := s.Pause(ctx, "fast", "fixture interruption", true); err != nil {
		t.Fatal(err)
	}
	var gotErr error
	select {
	case gotErr = <-done:
		joined = true
	case <-time.After(3 * time.Second):
		t.Fatal("suite did not interrupt")
	}
	var interrupted *queueInterruptedError
	if !errors.As(gotErr, &interrupted) || !strings.HasSuffix(rec.LogPath, ".interrupted.log") {
		t.Fatalf("diagnostic not separated: %+v %v", rec, gotErr)
	}
	log, err := os.ReadFile(rec.LogPath)
	if err != nil || !strings.Contains(string(log), "diagnostic") {
		t.Fatalf("raw diagnostic lost: %q %v", log, err)
	}
	if len(gh.statuses[sha]) != 0 || len(gh.comments[9]) != 0 {
		t.Fatalf("interruption posted a result: %+v %+v", gh.statuses, gh.comments)
	}
	logDir, _ := localCheckLogDir(d)
	rows, warnings, err := readSuiteAttempts(suiteStoreRoot(logDir), 100)
	if err != nil || len(rows) != 0 || len(warnings) != 0 {
		t.Fatalf("interruption became suite evidence: %v %v %v", rows, warnings, err)
	}
	if diagnoseSuiteFailure(context.Background(), gotErr, sha, d) != gotErr {
		t.Fatal("interruption entered baseline diagnosis")
	}
	c := &conductor{ctx: context.Background(), heads: []string{sha}}
	c.run = func(context.Context, localCheck, string) (localRunRecord, error) { return rec, gotErr }
	if event, err := c.runSuite(1, landingline.ExecInitial); event != nil || err == nil || c.baseline != nil {
		t.Fatalf("interruption blamed train: %v %v", event, err)
	}
	if err := s.Resume(context.Background(), "fast"); err != nil {
		t.Fatal(err)
	}
	check.Command = []string{"true"}
	if rec, err := runLocalCheck(context.Background(), d, check, sha); err != nil || !rec.passed() {
		t.Fatalf("same-head retry blocked: %+v %v", rec, err)
	}
}

func TestQueueAdmittedMergeDoesNotStartSuiteWhilePaused(t *testing.T) {
	d, _, _, _ := localRunRepo(t, "unused")
	d.Stderr = io.Discard
	t.Setenv("WF_SDLC_QUEUE_ROOT", t.TempDir())
	s := &queuegate.Store{Root: os.Getenv("WF_SDLC_QUEUE_ROOT")}
	if err := s.Pause(context.Background(), "graceful", "fixture", false); err != nil {
		t.Fatal(err)
	}
	_, err := runUnderSlot(withQueueAdmitted(context.Background()), d, d.Root, "missing-helper", t.TempDir(), "suite", "head", 1, suitePrioritySubject{}, []string{"false"}, filepath.Join(t.TempDir(), "log"), time.Second)
	var interrupted *queueInterruptedError
	if !errors.As(err, &interrupted) {
		t.Fatalf("paused boundary did not interrupt merge: %v", err)
	}
}
