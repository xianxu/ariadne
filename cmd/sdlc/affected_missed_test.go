package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
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

// The CLI over a real clone's ledger: list prunes what main's mapping covers (and lists without pruning when the
// mapping is unavailable), a read-only list never rewrites the file, and dismiss removes one entry while a repeat
// refuses.
func TestAffectedMissedListAndDismiss(t *testing.T) {
	root := t.TempDir()
	if out, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	store, err := missedStoreFor(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, m := range []affected.Miss{{Test: "p.TestA", Pkg: "p"}, {Test: "q.TestB", Pkg: "q"}} {
		if _, err := store.record(ctx, []affected.Miss{m}, missEvidence(1, "x/"+m.Pkg)); err != nil {
			t.Fatal(err)
		}
	}
	unavailable := func(context.Context) (func(affected.MissedEntry) bool, func(), error) {
		return nil, nil, errors.New("no graph")
	}
	before, _ := os.ReadFile(store.path())
	var out bytes.Buffer
	if err := runMissedList(ctx, &out, root, false, unavailable); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "nothing pruned: main's mapping is unavailable (no graph)") || strings.Count(out.String(), "paths: x/") != 2 {
		t.Fatalf("list without coverage:\n%s", out.String())
	}
	if after, _ := os.ReadFile(store.path()); string(after) != string(before) {
		t.Fatal("a read-only list rewrote the ledger")
	}
	coversQ := func(context.Context) (func(affected.MissedEntry) bool, func(), error) {
		return func(e affected.MissedEntry) bool { return e.Package == "q" }, func() {}, nil
	}
	out.Reset()
	if err := runMissedList(ctx, &out, root, true, coversQ); err != nil {
		t.Fatal(err)
	}
	var entries []affected.MissedEntry
	if err := json.Unmarshal(out.Bytes(), &entries); err != nil || len(entries) != 1 || entries[0].Package != "p" {
		t.Fatalf("pruned list: %v\n%s", err, out.String())
	}
	if l, _ := store.read(); len(l.Entries) != 1 {
		t.Fatalf("the covered entry is still in the ledger: %+v", l.Entries)
	}
	out.Reset()
	if err := runMissedDismiss(ctx, &out, root, entries[0].ID, "an environment failure"); err != nil || !strings.Contains(out.String(), "dismissed "+entries[0].ID) {
		t.Fatalf("dismiss: %v %s", err, out.String())
	}
	if err := runMissedDismiss(ctx, &out, root, entries[0].ID, "again"); err == nil || !strings.Contains(err.Error(), "no missed mapping") {
		t.Fatalf("a repeat dismiss: %v", err)
	}
	if err := runMissedDismiss(ctx, &out, root, entries[0].ID, " "); err == nil {
		t.Fatal("dismiss without a reason")
	}
	out.Reset()
	if err := runMissedList(ctx, &out, root, false, coversQ); err != nil || !strings.Contains(out.String(), "no open missed mappings") {
		t.Fatalf("empty list: %v %s", err, out.String())
	}
}
