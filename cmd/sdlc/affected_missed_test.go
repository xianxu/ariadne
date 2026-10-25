package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/affected"
)

func missEvidence(at int, paths ...string) affected.MissEvidence {
	return affected.MissEvidence{Paths: paths, Head: "h", Main: "m", Source: "merge", At: time.Unix(int64(at), 0).UTC()}
}

// #616: every store operation is a read-modify-write under one flock, so concurrent writers on one clone never lose
// each other's effects.
func TestMissedStoreConcurrentWriters(t *testing.T) {
	s := missedLedgerStore{dir: filepath.Join(t.TempDir(), "sdlc")}
	ctx := context.Background()
	miss := []affected.Miss{{Test: "p.TestX", Pkg: "p"}}
	if _, err := s.record(ctx, []affected.Miss{{Test: "q.TestY", Pkg: "q"}}, missEvidence(1, "b")); err != nil {
		t.Fatal(err)
	}
	other := affected.MissID("q", []string{"b"})
	var wg sync.WaitGroup
	errs := make(chan error, 3)
	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.record(ctx, miss, missEvidence(10+i, "a"))
			errs <- err
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- s.update(ctx, time.Now(), func(l *affected.MissedLedger) error { _, err := l.Dismiss(other); return err })
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	l, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Entries) != 1 || l.Entries[0].Package != "p" || l.Entries[0].Hits != 2 {
		t.Fatalf("entries = %+v", l.Entries)
	}
}

// A ledger that does not decode is refused and never overwritten; a crashed writer's old temp file is swept.
func TestMissedStoreRefusesAndSweeps(t *testing.T) {
	s := missedLedgerStore{dir: filepath.Join(t.TempDir(), "sdlc")}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte(`{"version":2,"entries":[]}`)
	if err := os.WriteFile(s.path(), bad, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := s.record(context.Background(), []affected.Miss{{Test: "p.TestX", Pkg: "p"}}, missEvidence(1, "a")); err == nil || !strings.Contains(err.Error(), "repair or remove") {
		t.Fatalf("record over a refused ledger: %v", err)
	}
	if got, _ := os.ReadFile(s.path()); string(got) != string(bad) {
		t.Fatalf("refused ledger rewritten: %s", got)
	}
	os.Remove(s.path())
	stale := filepath.Join(s.dir, missedTempPrefix+"crashed")
	fresh := filepath.Join(s.dir, missedTempPrefix+"live")
	for _, p := range []string{stale, fresh} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := time.Now().Add(-2 * missedTempMaxAge)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	if _, err := s.record(context.Background(), []affected.Miss{{Test: "p.TestX", Pkg: "p"}}, missEvidence(1, "a")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale temp kept: %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatalf("a live temp was swept: %v", err)
	}
}

// A cancelled caller never waits on the lock.
func TestMissedStoreLockHonoursContext(t *testing.T) {
	s := missedLedgerStore{dir: filepath.Join(t.TempDir(), "sdlc")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.update(ctx, time.Now(), func(*affected.MissedLedger) error { return nil }); err == nil {
		t.Fatal("a cancelled update ran")
	}
}
