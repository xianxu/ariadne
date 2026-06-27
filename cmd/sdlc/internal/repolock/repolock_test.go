package repolock

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

	long := Metadata{PID: 23456, Hostname: "host-a", Command: "sdlc change-code"}
	line = HolderLine(long)
	for _, want := range []string{"pid 23456", "sdlc change-code", "review/ship transaction"} {
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
		{"sdlc change-code", true},
		{"sdlc close", true},
		{"sdlc milestone-close", true},
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
			name: "stale by age",
			meta: Metadata{PID: 100, Hostname: "host-a", StartedAt: now.Add(-maxAge - time.Second)},
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
	for _, want := range []string{"waiting for sdlc repo lock held by", "pid 555", "sdlc change-code", "review/ship transaction"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stderr missing %q:\n%s", want, out)
		}
	}
	if sleeps == 0 {
		t.Fatal("Acquire did not wait")
	}
	if !strings.Contains(err.Error(), "timed out waiting") || !strings.Contains(err.Error(), "review/ship transaction") {
		t.Fatalf("timeout error missing guidance: %v", err)
	}
}

func TestAcquireReportsStaleLockWithoutRemovingIt(t *testing.T) {
	gitDir := filepath.Join(t.TempDir(), ".git")
	lockDir := filepath.Join(gitDir, "sdlc.lock")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	holder := Metadata{PID: 555, Hostname: "host-a", Command: "sdlc issue new", StartedAt: time.Now()}
	if err := os.WriteFile(filepath.Join(lockDir, "meta.json"), Encode(holder), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Acquire(context.Background(), Options{
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
	if err == nil {
		t.Fatal("Acquire should report stale lock")
	}
	if !strings.Contains(err.Error(), "stale sdlc repo lock") || !strings.Contains(err.Error(), ".git/sdlc.lock") {
		t.Fatalf("stale error missing recovery details: %v", err)
	}
	if _, statErr := os.Stat(lockDir); statErr != nil {
		t.Fatalf("stale lock should not be removed automatically: %v", statErr)
	}
}
