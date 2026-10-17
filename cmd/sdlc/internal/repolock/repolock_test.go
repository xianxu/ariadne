package repolock

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

func TestMetadataRoundTrip(t *testing.T) {
	started := time.Date(2026, 6, 27, 12, 1, 2, 0, time.FixedZone("PDT", -7*60*60))
	meta := Metadata{
		PID:       12345,
		Hostname:  "host-a",
		CWD:       "/repo",
		Command:   "sdlc issue new",
		Args:      []string{"sdlc", "issue", "new", "Race test"},
		StartedAt: started,
	}

	got, err := Decode(Encode(meta))
	if err != nil {
		t.Fatalf("Decode(Encode(meta)) err: %v", err)
	}
	if got.PID != meta.PID || got.Hostname != meta.Hostname || got.CWD != meta.CWD || got.Command != meta.Command {
		t.Fatalf("decoded scalar fields = %+v, want %+v", got, meta)
	}
	if !got.StartedAt.Equal(started) {
		t.Fatalf("StartedAt = %s, want %s", got.StartedAt, started)
	}
	if strings.Join(got.Args, "\x00") != strings.Join(meta.Args, "\x00") {
		t.Fatalf("Args = %#v, want %#v", got.Args, meta.Args)
	}
}

func TestHolderLineIncludesPIDCommandAndReviewHint(t *testing.T) {
	quick := Metadata{PID: 12345, Hostname: "host-a", Command: "sdlc issue new"}
	line := HolderLine(quick)
	for _, want := range []string{"pid 12345", "sdlc issue new"} {
		if !strings.Contains(line, want) {
			t.Fatalf("HolderLine missing %q: %q", want, line)
		}
	}
	if strings.Contains(line, "review/ship transaction") {
		t.Fatalf("quick command should not get review/ship hint: %q", line)
	}

	long := Metadata{PID: 23456, Hostname: "host-a", Command: "sdlc merge"}
	line = HolderLine(long)
	for _, want := range []string{"pid 23456", "sdlc merge", "review/ship transaction"} {
		if !strings.Contains(line, want) {
			t.Fatalf("HolderLine missing %q: %q", want, line)
		}
	}
}

func TestIsLongRunningCommand(t *testing.T) {
	cases := []struct {
		command string
		want    bool
	}{
		{"sdlc change-code", false},
		{"sdlc close", false},
		{"sdlc milestone-close", false},
		{"sdlc merge", true},
		{"sdlc push", true},
		{"sdlc issue new", false},
		{"sdlc claim", false},
	}
	for _, c := range cases {
		t.Run(c.command, func(t *testing.T) {
			if got := IsLongRunningCommand(Metadata{Command: c.command}); got != c.want {
				t.Fatalf("IsLongRunningCommand(%q) = %v, want %v", c.command, got, c.want)
			}
		})
	}
}

func TestObserveClassifiesStaleLocks(t *testing.T) {
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	maxAge := 30 * time.Minute
	alive := func(pid int) bool { return pid == 100 }

	cases := []struct {
		name string
		meta Metadata
		want ObservationKind
	}{
		{
			name: "active same-host process",
			meta: Metadata{PID: 100, Hostname: "host-a", StartedAt: now.Add(-time.Minute)},
			want: ObservationActive,
		},
		{
			name: "stale same-host missing process",
			meta: Metadata{PID: 101, Hostname: "host-a", StartedAt: now.Add(-time.Minute)},
			want: ObservationStaleMissingProcess,
		},
		{
			name: "different host does not use local process check",
			meta: Metadata{PID: 101, Hostname: "host-b", StartedAt: now.Add(-time.Minute)},
			want: ObservationActive,
		},
		{
			name: "live same-host process overrides age",
			meta: Metadata{PID: 100, Hostname: "host-a", StartedAt: now.Add(-maxAge - time.Second)},
			want: ObservationActive,
		},
		{
			name: "stale by age when liveness is not provable",
			meta: Metadata{PID: 100, Hostname: "host-b", StartedAt: now.Add(-maxAge - time.Second)},
			want: ObservationStaleAge,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Observe(c.meta, now, "host-a", alive, maxAge)
			if got.Kind != c.want {
				t.Fatalf("Observe kind = %v, want %v (message: %s)", got.Kind, c.want, got.Message)
			}
			if c.want != ObservationActive && !strings.Contains(got.Message, ".git/sdlc.lock") {
				t.Fatalf("stale/recovery message should mention .git/sdlc.lock: %q", got.Message)
			}
		})
	}
}

func TestAcquireCreatesMetadataAndReleaseRemovesLock(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	var stderr bytes.Buffer

	lock, err := Acquire(context.Background(), Options{
		GitCommonDir:  gitDir,
		Command:       "sdlc issue new",
		Args:          []string{"sdlc", "issue", "new", "x"},
		Hostname:      "host-a",
		PID:           123,
		CWD:           "/repo",
		Now:           func() time.Time { return now },
		ProcessAlive:  func(int) bool { return true },
		Sleep:         func(context.Context, time.Duration) error { return nil },
		Stderr:        &stderr,
		WaitTimeout:   time.Minute,
		StaleDuration: time.Hour,
	})
	if err != nil {
		t.Fatalf("Acquire err: %v", err)
	}
	metaPath := filepath.Join(gitDir, "sdlc.lock", "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("metadata not written at %s: %v", metaPath, err)
	}
	meta, err := Decode(data)
	if err != nil {
		t.Fatalf("metadata decode: %v", err)
	}
	if meta.PID != 123 || meta.Command != "sdlc issue new" || !meta.StartedAt.Equal(now) {
		t.Fatalf("metadata = %+v", meta)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("Release err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(gitDir, "sdlc.lock")); !os.IsNotExist(err) {
		t.Fatalf("lock dir should be removed, stat err = %v", err)
	}
}

func TestAcquireWaitsAndReportsHolder(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	if err := os.MkdirAll(filepath.Join(gitDir, "sdlc.lock"), 0o755); err != nil {
		t.Fatal(err)
	}
	holder := Metadata{PID: 555, Hostname: "host-a", Command: "sdlc change-code", StartedAt: time.Now()}
	if err := os.WriteFile(filepath.Join(gitDir, "sdlc.lock", "meta.json"), Encode(holder), 0o644); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	var sleeps int
	now := holder.StartedAt
	_, err := Acquire(context.Background(), Options{
		GitCommonDir: gitDir,
		Command:      "sdlc issue new",
		Hostname:     "host-a",
		PID:          999,
		Now:          func() time.Time { return now },
		ProcessAlive: func(int) bool { return true },
		Sleep: func(context.Context, time.Duration) error {
			sleeps++
			now = now.Add(time.Second)
			return nil
		},
		Stderr:        &stderr,
		WaitTimeout:   2 * time.Second,
		StaleDuration: time.Hour,
	})
	if err == nil {
		t.Fatal("Acquire should time out while holder remains active")
	}
	out := stderr.String()
	for _, want := range []string{"waiting for sdlc repo lock held by", "pid 555", "sdlc change-code"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stderr missing %q:\n%s", want, out)
		}
	}
	if sleeps == 0 {
		t.Fatal("Acquire did not wait")
	}
	if !strings.Contains(err.Error(), "timed out waiting") {
		t.Fatalf("timeout error missing guidance: %v", err)
	}
}

func TestAcquireReclaimsDeadSameHostHolder(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	lockDir := filepath.Join(gitDir, "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	holder := Metadata{PID: 555, Hostname: "host-a", Command: "sdlc issue new", StartedAt: time.Now()}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), Encode(holder), 0o644); err != nil {
		t.Fatal(err)
	}

	lock, err := Acquire(context.Background(), Options{
		GitCommonDir:  gitDir,
		Command:       "sdlc claim",
		Hostname:      "host-a",
		PID:           999,
		Now:           func() time.Time { return holder.StartedAt },
		ProcessAlive:  func(int) bool { return false },
		Sleep:         func(context.Context, time.Duration) error { return nil },
		Stderr:        &bytes.Buffer{},
		WaitTimeout:   time.Second,
		StaleDuration: time.Hour,
	})
	if err != nil {
		t.Fatalf("Acquire should reclaim dead same-host holder: %v", err)
	}
	defer func() { _ = lock.Release() }()
	data, err := os.ReadFile(filepath.Join(lockDir, "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	meta, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Command != "sdlc claim" || meta.PID != 999 {
		t.Fatalf("metadata was not replaced after reclaim: %+v", meta)
	}
}

func TestAcquireWaitsForInitializingMetadata(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	lockDir := filepath.Join(gitDir, "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	var sleeps int
	lock, err := Acquire(context.Background(), Options{
		GitCommonDir: gitDir,
		Command:      "sdlc claim",
		Hostname:     "host-a",
		PID:          999,
		Now:          func() time.Time { return now },
		ProcessAlive: func(int) bool { return true },
		Sleep: func(context.Context, time.Duration) error {
			sleeps++
			now = now.Add(100 * time.Millisecond)
			if sleeps == 1 {
				if err := os.RemoveAll(lockDir); err != nil {
					return err
				}
			}
			return nil
		},
		Stderr:        &bytes.Buffer{},
		WaitTimeout:   time.Second,
		StaleDuration: time.Hour,
		PollInterval:  100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Acquire should wait through missing metadata: %v", err)
	}
	defer func() { _ = lock.Release() }()
	if sleeps == 0 {
		t.Fatal("Acquire did not wait for initializing metadata")
	}
}

func TestConcurrentAcquireSerializesRealMkdirLock(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	if err := os.MkdirAll(gitDir, 0o755); err != nil {
		t.Fatal(err)
	}
	first, err := Acquire(context.Background(), Options{
		GitCommonDir:  gitDir,
		Command:       "sdlc issue new",
		Hostname:      "host-a",
		PID:           111,
		Now:           time.Now,
		ProcessAlive:  func(int) bool { return true },
		Stderr:        &bytes.Buffer{},
		WaitTimeout:   testfix.HangCeiling, // not under test (#325)
		StaleDuration: time.Hour,
		PollInterval:  10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("first Acquire err: %v", err)
	}

	waiting := make(chan struct{}, 1)
	done := make(chan error, 1)
	go func() {
		second, err := Acquire(context.Background(), Options{
			GitCommonDir: gitDir,
			Command:      "sdlc claim",
			Hostname:     "host-a",
			PID:          222,
			Now:          time.Now,
			ProcessAlive: func(int) bool { return true },
			Sleep: func(ctx context.Context, d time.Duration) error {
				select {
				case waiting <- struct{}{}:
				default:
				}
				timer := time.NewTimer(d)
				defer timer.Stop()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-timer.C:
					return nil
				}
			},
			Stderr:        &bytes.Buffer{},
			WaitTimeout:   testfix.HangCeiling, // not under test (#325)
			StaleDuration: time.Hour,
			PollInterval:  10 * time.Millisecond,
		})
		if err != nil {
			done <- err
			return
		}
		done <- second.Release()
	}()

	select {
	case <-waiting:
	case <-time.After(testfix.HangCeiling):
		t.Fatal("second Acquire did not wait on first holder")
	}
	if err := first.Release(); err != nil {
		t.Fatalf("first Release err: %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("second Acquire/Release err: %v", err)
	}
}

func TestConcurrentAcquireReclaimsDeadHolderOnce(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	lockDir := filepath.Join(gitDir, "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	holder := Metadata{PID: 555, Hostname: "host-a", Command: "sdlc issue new", StartedAt: time.Now()}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), Encode(holder), 0o644); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	var held atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
	defer cancel()
	start := make(chan struct{})
	errs := make(chan error, 2)
	for _, pid := range []int{111, 222} {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			<-start
			lock, err := Acquire(ctx, Options{
				GitCommonDir:  gitDir,
				Command:       "sdlc claim",
				Hostname:      "host-a",
				PID:           pid,
				Now:           time.Now,
				ProcessAlive:  func(pid int) bool { return pid != 555 },
				Stderr:        &bytes.Buffer{},
				WaitTimeout:   testfix.HangCeiling, // not under test (#325)
				StaleDuration: time.Hour,
				PollInterval:  10 * time.Millisecond,
			})
			if err != nil {
				errs <- err
				return
			}
			if n := held.Add(1); n != 1 {
				err = fmt.Errorf("%d concurrent lock holders", n)
			}
			meta, readErr := ReadMetadata(lockDir)
			if readErr != nil || meta.PID != pid {
				err = fmt.Errorf("PID %d holds mismatched metadata %+v: %v", pid, meta, readErr)
			}
			held.Add(-1)
			if releaseErr := lock.Release(); releaseErr != nil {
				err = releaseErr
			}
			errs <- err
		}(pid)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(lockDir); !os.IsNotExist(err) {
		t.Fatalf("lock remains after both releases: %v", err)
	}
}

func TestAcquireDelayedDeadObservationPreservesNewLiveHolder(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	lockDir := filepath.Join(gitDir, "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), Encode(Metadata{PID: 555, Hostname: "host-a", StartedAt: time.Now()}), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), testfix.HangCeiling)
	sawDead, resume, released := make(chan struct{}), make(chan struct{}), make(chan struct{})
	waiting := make(chan struct{}, 1)
	type result struct {
		lock *Lock
		err  error
	}
	done := make(chan result, 1)
	var resumeOnce sync.Once
	var workers sync.WaitGroup
	var first, second *Lock
	t.Cleanup(func() {
		cancel()
		resumeOnce.Do(func() { close(resume) })
		if first != nil {
			_ = first.Release()
		}
		workers.Wait()
		select {
		case r := <-done:
			second = r.lock
		default:
		}
		if second != nil {
			_ = second.Release()
		}
	})
	opts := Options{GitCommonDir: gitDir, Hostname: "host-a", PID: 222, CallerHandlesSignals: true,
		ProcessAlive: func(pid int) bool { return pid != 555 }, Stderr: &bytes.Buffer{}, WaitTimeout: testfix.HangCeiling}
	delayed := opts
	delayed.ProcessAlive = func(pid int) bool {
		if pid == 555 {
			close(sawDead)
			select {
			case <-resume:
			case <-ctx.Done():
			}
		}
		return pid != 555 // Only the original holder is dead.
	}
	delayed.Sleep = func(ctx context.Context, _ time.Duration) error {
		select {
		case waiting <- struct{}{}:
		default:
		}
		select {
		case <-released:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	workers.Add(1)
	go func() {
		defer workers.Done()
		lock, err := Acquire(ctx, delayed)
		done <- result{lock, err}
	}()
	select {
	case <-sawDead:
	case <-ctx.Done():
		t.Fatal("contender did not observe the dead holder")
	}
	opts.PID = 111
	var err error
	first, err = Acquire(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	resumeOnce.Do(func() { close(resume) })
	select {
	case <-waiting:
	case r := <-done:
		second = r.lock
		t.Fatalf("delayed contender completed while live PID 111 still held the lock: %v", r.err)
	case <-ctx.Done():
		t.Fatal("delayed contender neither waited nor completed")
	}
	meta, err := ReadMetadata(lockDir)
	if err != nil || meta.PID != 111 {
		t.Fatalf("live ownership changed: %+v, %v", meta, err)
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	first = nil
	close(released)
	select {
	case r := <-done:
		second = r.lock
		if r.err != nil {
			t.Fatal(r.err)
		}
	case <-ctx.Done():
		t.Fatal("contender did not acquire after release")
	}
	meta, err = ReadMetadata(lockDir)
	if err != nil || meta.PID != 222 {
		t.Fatalf("new ownership missing: %+v, %v", meta, err)
	}
}

func TestAcquireCancelledContextDoesNotCreateLock(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	lock, err := Acquire(ctx, Options{GitCommonDir: dir, PID: os.Getpid()})
	if lock != nil {
		defer lock.Release()
	}
	if err != context.Canceled {
		t.Errorf("cancelled acquire: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sdlc.lock")); !os.IsNotExist(err) {
		t.Errorf("cancelled caller acquired lock: %v", err)
	}
}
