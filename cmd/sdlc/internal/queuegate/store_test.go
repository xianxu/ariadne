package queuegate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	return &Store{Root: t.TempDir()}
}

func enrollTest(t *testing.T, s *Store, id string, managed bool) *Lease {
	t.Helper()
	l, err := s.Enroll(context.Background(), Operation{SchemaVersion: 1, ID: id, Workspace: "repo:3", Kinds: []string{"suite"}, Phase: "waiting", PID: os.Getpid(), PIDStarted: "test-incarnation", Managed: managed, StartedAt: time.Now().UTC(), BootID: "test-boot", Root: "/fixture"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l
}

func observeTest(t *testing.T, s *Store) Snapshot {
	t.Helper()
	x, err := s.Snapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return x
}

func TestPauseAdmissionAtomic(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "atomic", true)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		_, err := l.TryAdmit(context.Background(), func() (bool, error) {
			close(entered)
			<-release
			return true, nil
		})
		done <- err
	}()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	err := s.Pause(ctx, "drain", "reason", false)
	cancel()
	close(release)
	if runErr := <-done; runErr != nil {
		t.Fatal(runErr)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("pause crossed admission transaction: %v", err)
	}
	if err := s.Pause(context.Background(), "drain", "reason", false); err != nil {
		t.Fatal(err)
	}
	x := observeTest(t, s)
	if x.State != "paused" || !x.Complete || len(x.Active) != 1 || x.Active[0].ID != "atomic" {
		t.Fatalf("admission missing from paused snapshot: %+v", x)
	}
}

func TestPauseWaitingExclusionAndOwnerResume(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "waiting", true)
	for _, owner := range []string{"a", "b", "a"} {
		if err := s.Pause(context.Background(), owner, owner, false); err != nil {
			t.Fatal(err)
		}
	}
	called := false
	_, err := l.TryAdmit(context.Background(), func() (bool, error) { called = true; return true, nil })
	var paused *PausedError
	if !errors.As(err, &paused) || called {
		t.Fatalf("paused admission ran resource callback: called=%v err=%v", called, err)
	}
	x := observeTest(t, s)
	if len(x.Active) != 0 || len(x.PauseOwners) != 2 {
		t.Fatalf("waiting became active or owners lost: %+v", x)
	}
	if err := s.Resume(context.Background(), "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := l.TryAdmit(context.Background(), nil); !errors.As(err, &paused) {
		t.Fatalf("removing a released b: %v", err)
	}
	if err := s.Resume(context.Background(), "b"); err != nil {
		t.Fatal(err)
	}
	if ok, err := l.TryAdmit(context.Background(), nil); !ok || err != nil {
		t.Fatalf("last owner resume: ok=%v %v", ok, err)
	}
}

func TestFastPauseRetainsInterruptedNotFailed(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "managed", true)
	if _, err := l.TryAdmit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Pause(context.Background(), "drain", "fast", true); err != nil {
		t.Fatal(err)
	}
	if x := observeTest(t, s); len(x.Active) != 1 || x.Active[0].State != "interrupting" {
		t.Fatalf("not interrupting: %+v", x)
	}
	if requested, err := l.CancellationRequested(context.Background()); !requested || err != nil {
		t.Fatalf("intent not durable: %v %v", requested, err)
	}
	if err := s.Resume(context.Background(), "drain"); err != nil {
		t.Fatal(err)
	}
	if requested, _ := l.CancellationRequested(context.Background()); !requested {
		t.Fatal("resume undid cancellation")
	}
	if err := l.Retire(context.Background(), "drain cancellation"); err != nil {
		t.Fatal(err)
	}
	if x := observeTest(t, s); !x.Complete || len(x.Active) != 0 {
		t.Fatalf("managed retirement: %+v", x)
	}
	rows, err := os.ReadDir(filepath.Join(s.Root, "interrupted"))
	if err != nil || len(rows) != 1 {
		t.Fatalf("no distinct interruption evidence: %v %v", rows, err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, "failed")); !os.IsNotExist(err) {
		t.Fatalf("interruption entered failure family: %v", err)
	}
}

func TestUnmanagedRetirementStaysUnknown(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "unmanaged", false)
	if _, err := l.TryAdmit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := s.Pause(context.Background(), "drain", "fast", true); err != nil {
		t.Fatal(err)
	}
	if err := l.Retire(context.Background(), "make exited; descendants unproved"); err != nil {
		t.Fatal(err)
	}
	x := observeTest(t, s)
	if x.Complete || len(x.Active) != 1 || x.Active[0].State != "unknown" || x.Detail == "" {
		t.Fatalf("unmanaged death became quiescence: %+v", x)
	}
	if err := s.Resume(context.Background(), "drain"); err != nil {
		t.Fatal(err)
	}
	if x := observeTest(t, s); x.Complete || len(x.Active) != 1 {
		t.Fatalf("resume erased unresolved descendants: %+v", x)
	}
}

func TestQueueUnknownHeldLease(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "unreadable", true)
	if err := l.file.Truncate(0); err != nil {
		t.Fatal(err)
	}
	if x := observeTest(t, s); x.Complete || x.Detail == "" || len(x.Active) != 1 || x.Active[0].State != "unknown" {
		t.Fatalf("held malformed lease treated absent: %+v", x)
	}
}

func TestQueueUnknownFilenameStillNamesOneActivity(t *testing.T) {
	s := testStore(t)
	_ = enrollTest(t, s, "renamed", true)
	if err := os.Rename(filepath.Join(s.Root, "operations", "renamed.json"), filepath.Join(s.Root, "operations", "untrusted 💥.json")); err != nil {
		t.Fatal(err)
	}
	x := observeTest(t, s)
	if x.Complete || len(x.Active) != 1 || x.Active[0].State != "unknown" {
		t.Fatalf("noncanonical filename hid unknown holder: %+v", x)
	}
}

func TestDeadLeaseRetirementKeepsOneUnknown(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "owner-died", true)
	if _, err := l.TryAdmit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		x := observeTest(t, s)
		if x.Complete || len(x.Active) != 1 || x.Active[0].ID != "owner-died" || x.Active[0].State != "unknown" {
			t.Fatalf("snapshot %d lost or duplicated unknown owner: %+v", i, x)
		}
		if strings.Contains(x.Detail, "duplicate") {
			t.Fatalf("snapshot duplicated one retired owner: %s", x.Detail)
		}
	}
}

func TestQueueLinearScan(t *testing.T) {
	for _, n := range []int{8, 16} {
		s := testStore(t)
		for i := 0; i < n; i++ {
			l := enrollTest(t, s, fmt.Sprintf("linear-%d", i), true)
			if _, err := l.TryAdmit(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
		}
		reads := 0
		s.onLeaseRead = func(string) { reads++ }
		x := observeTest(t, s)
		if !x.Complete || len(x.Active) != n {
			t.Fatalf("bad fixture snapshot: %+v", x)
		}
		if reads != n {
			t.Fatalf("%d leases caused %d reads; require one pass", n, reads)
		}
	}
}

func TestQueueMalformedControlCannotResumeAdmission(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "bad-control", true)
	if err := os.WriteFile(filepath.Join(s.Root, "pause.json"), []byte(`{"schema_version":1,"pause_owners":null}`), 0o600); err != nil {
		t.Fatal(err)
	}
	called := false
	if ok, err := l.TryAdmit(context.Background(), func() (bool, error) { called = true; return true, nil }); ok || err == nil || called {
		t.Fatalf("malformed control admitted work: ok=%v called=%v err=%v", ok, called, err)
	}
	if x := observeTest(t, s); x.Complete || x.Detail == "" {
		t.Fatalf("malformed control became absence: %+v", x)
	}
}

func TestQueuePriorBootResolvesUnmanagedUncertainty(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "old-boot", false)
	if _, err := l.TryAdmit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := l.Retire(context.Background(), "owner died without descendant proof"); err != nil {
		t.Fatal(err)
	}
	s.BootID = func() (string, error) { return "", os.ErrPermission }
	if x := observeTest(t, s); x.Complete {
		t.Fatal("unavailable boot identity erased uncertainty")
	}
	s.BootID = func() (string, error) { return "test-boot", nil }
	if x := observeTest(t, s); x.Complete {
		t.Fatal("same boot erased uncertainty")
	}
	s.BootID = func() (string, error) { return "definitely-new-boot", nil }
	if x := observeTest(t, s); !x.Complete || len(x.Active) != 0 {
		t.Fatalf("prior-boot operation survived reconciliation: %+v", x)
	}
}

func TestQueueAbandonedEnrollmentHasNoAdmittedWork(t *testing.T) {
	s := testStore(t)
	if err := os.MkdirAll(filepath.Join(s.Root, "operations"), 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(s.Root, "operations", ".enroll-abandoned")
	if err := os.WriteFile(path, []byte(`{"partial":`), 0o600); err != nil {
		t.Fatal(err)
	}
	if x := observeTest(t, s); !x.Complete || len(x.Active) != 0 {
		t.Fatalf("unpublished enrollment became active uncertainty: %+v", x)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("abandoned unpublished file not swept: %v", err)
	}
}
