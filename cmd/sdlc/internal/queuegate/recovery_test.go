package queuegate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func crashIntent(t *testing.T, s *Store, id string) {
	t.Helper()
	data := []byte(fmt.Sprintf(`{"schema_version":1,"id":%q,"reason":"crash after durable intent","future":{"large":1e99999}}`, id))
	if err := atomicWrite(filepath.Join(s.Root, "cancel", id+".json"), data); err != nil {
		t.Fatal(err)
	}
}

func TestQueueOrphanIntentNeedsMatchingInterruptionProof(t *testing.T) {
	for _, proof := range []string{"matching", "missing", "corrupt"} {
		t.Run(proof, func(t *testing.T) {
			s := testStore(t)
			l := enrollTest(t, s, "orphan", true)
			if _, err := l.TryAdmit(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
			if err := l.Retire(context.Background(), "cancelled"); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(s.Root, "interrupted", "orphan.json")
			if proof == "missing" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			}
			if proof == "corrupt" {
				if err := os.WriteFile(path, []byte(`{"id":"other"}`), 0600); err != nil {
					t.Fatal(err)
				}
			}
			crashIntent(t, s, "orphan") // Crash after removing the lease, before removing its intent.
			x := observeTest(t, s)
			_, err := os.Stat(s.intentPath("orphan"))
			if proof == "matching" {
				if !x.Complete || len(x.Active) != 0 || !os.IsNotExist(err) {
					t.Fatalf("proved orphan not collected: %+v %v", x, err)
				}
			} else if x.Complete || len(x.Active) != 1 || err != nil {
				t.Fatalf("unproved orphan disappeared: %+v %v", x, err)
			}
		})
	}
}

func TestQueuePauseAtRegistryCapacity(t *testing.T) {
	s := testStore(t)
	for i := 0; i < 2; i++ {
		l := enrollTest(t, s, fmt.Sprintf("active-%04d", i), true)
		if _, err := l.TryAdmit(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}
	// Build the rest directly: setup must not measure thousands of Enroll scans.
	op := Operation{SchemaVersion: 1, Workspace: "repo:3", Kinds: []string{"suite"}, Phase: "waiting", PID: os.Getpid(), PIDStarted: "fixture", Managed: true, StartedAt: time.Now().UTC(), Root: "/fixture"}
	for i := 2; i < MaxOperations; i++ {
		op.ID = fmt.Sprintf("waiting-%04d", i)
		data, err := EncodeOperation(op)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(s.Root, "operations", op.ID+".json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	reads := 0
	s.onLeaseRead = func(string) { reads++ }
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	err := s.Pause(ctx, "capacity", "capacity check", true)
	t.Logf("capacity=%d reads=%d elapsed=%s result=%v", MaxOperations, reads, time.Since(start), err)
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	if reads > MaxOperations {
		t.Fatalf("repeated scan: %d", reads)
	}
	s.onLeaseRead = nil
	control, readErr := s.control()
	if readErr != nil || len(control.PauseOwners) != 1 {
		t.Fatalf("applied hold lost: %+v %v", control, readErr)
	}
	if err == nil && reads != MaxOperations {
		t.Fatalf("successful scan skipped rows: %d", reads)
	}
}

func TestQueueCrashIntentBeforePhaseIsAuthoritative(t *testing.T) {
	for _, consumer := range []string{"owner", "snapshot"} {
		t.Run(consumer, func(t *testing.T) {
			s := testStore(t)
			l := enrollTest(t, s, "intent-window", true)
			if _, err := l.TryAdmit(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
			crashIntent(t, s, "intent-window") // Pause died before rewriting phase.
			if consumer == "owner" {
				if yes, err := l.CancellationRequested(context.Background()); !yes || err != nil {
					t.Fatalf("durable intent ignored: %v %v", yes, err)
				}
			} else {
				x := observeTest(t, s)
				if len(x.Active) != 1 || x.Active[0].State != "interrupting" {
					t.Fatalf("phase hid persisted intent: %+v", x)
				}
			}
		})
	}
}

func TestQueueInterruptedOwnerDeathRetainsEvidenceAndCleansIntent(t *testing.T) {
	s := testStore(t)
	l := enrollTest(t, s, "dead-interrupted", true)
	if _, err := l.TryAdmit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	crashIntent(t, s, "dead-interrupted")
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
	x := observeTest(t, s)
	if x.Complete || len(x.Active) != 1 || x.Active[0].State != "unknown" {
		t.Fatalf("dead owner's descendants became complete: %+v", x)
	}
	path := filepath.Join(s.Root, "interrupted", "dead-interrupted.json")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lost interruption at owner death: %v", err)
	}
	var record map[string]json.RawMessage
	if err := json.Unmarshal(before, &record); err != nil {
		t.Fatal(err)
	}
	if string(record["outcome"]) != `"interrupted"` || len(record["operation"]) == 0 {
		t.Fatalf("interruption envelope missing: %s", before)
	}
	for i := 0; i < 2; i++ {
		_ = observeTest(t, s)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(before) != string(after) {
		t.Fatalf("recovery rewrote immutable interruption: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, "cancel", "dead-interrupted.json")); !os.IsNotExist(err) {
		t.Fatalf("retired intent leaked: %v", err)
	}
	s.BootID = func() (string, error) { return "next-boot", nil }
	if x := observeTest(t, s); !x.Complete || len(x.Active) != 0 {
		t.Fatalf("prior boot unresolved: %+v", x)
	}
}

func TestQueueIntentReadRequiresExactTypedIdentity(t *testing.T) {
	for _, raw := range []string{
		`{"schema_version":1,"id":"other","reason":"bad"}`,
		`{"schema_version":1,"id":"identity","id":"other","reason":"bad"}`,
		`{"schema_version":1,"id":"identity","reason":null}`,
	} {
		t.Run(raw, func(t *testing.T) {
			s := testStore(t)
			l := enrollTest(t, s, "identity", true)
			if _, err := l.TryAdmit(context.Background(), nil); err != nil {
				t.Fatal(err)
			}
			if err := atomicWrite(filepath.Join(s.Root, "cancel", "identity.json"), []byte(raw)); err != nil {
				t.Fatal(err)
			}
			if _, err := l.CancellationRequested(context.Background()); err == nil {
				t.Fatal("untrusted intent accepted or ignored")
			}
			if x := observeTest(t, s); x.Complete {
				t.Fatalf("bad intent proved complete: %+v", x)
			}
		})
	}
}

func TestQueueSnapshotDeadlineInsideTransaction(t *testing.T) {
	s := testStore(t)
	for i := 0; i < 16; i++ {
		enrollTest(t, s, fmt.Sprintf("deadline-%d", i), true)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reads := 0
	s.onLeaseRead = func(string) { reads++; cancel() }
	_, err := s.Snapshot(ctx)
	if !errors.Is(err, context.Canceled) || reads != 1 {
		t.Fatalf("transaction ignored deadline: reads=%d err=%v", reads, err)
	}
}

func TestQueuePauseDeadlineKeepsAppliedOwnerAndPartialIntents(t *testing.T) {
	s := testStore(t)
	for i := 0; i < 16; i++ {
		l := enrollTest(t, s, fmt.Sprintf("budget-%d", i), true)
		if _, err := l.TryAdmit(context.Background(), nil); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reads := 0
	s.onLeaseRead = func(string) {
		reads++
		if reads == 2 {
			cancel()
		}
	}
	err := s.Pause(ctx, "budget-owner", "fast", true)
	if !errors.Is(err, context.Canceled) || reads > 2 {
		t.Fatalf("pause ignored transaction budget: reads=%d err=%v", reads, err)
	}
	s.onLeaseRead = nil
	x := observeTest(t, s)
	if x.State != "paused" || len(x.PauseOwners) != 1 {
		t.Fatalf("timeout undid applied hold: %+v", x)
	}
	if err := s.Pause(context.Background(), "budget-owner", "fast", true); err != nil {
		t.Fatal(err)
	}
	for _, active := range observeTest(t, s).Active {
		if active.State != "interrupting" {
			t.Fatalf("retry did not finish exact-ID publication: %+v", active)
		}
	}
}
