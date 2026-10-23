package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func benchRecordFixture(t *testing.T) (*benchStore, string, *benchCleanupRuntime) {
	t.Helper()
	s, err := createBenchStore(filepath.Join(t.TempDir(), "run"), benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range []bench.EventKind{bench.BeginPrepare, bench.Prepared, bench.Stop} {
		op := "prepare"
		if kind == bench.Stop {
			op = "stop"
		}
		if _, _, err := s.apply(bench.Event{ID: []string{"prepare", "prepared", "stop"}[i], RunID: r.ID, Kind: kind, OperationID: op}); err != nil {
			t.Fatal(err)
		}
	}
	m := ledgerFixtureRow(t, r.ID).Measurements
	m.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"nested":1e400}`)}
	if err := s.putData("measurements", m); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("acceptance", bench.Acceptance{Status: "unavailable", Reason: "stopped"}); err != nil {
		t.Fatal(err)
	}
	runtime, err := benchRetirementSubject(r, mustBenchRecordState(t, s))
	if err != nil {
		t.Fatal(err)
	}
	if err := s.putData("retirement", benchRetirement{Version: 1, RunID: r.ID, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, ConfirmedBy: "fixture", ConfirmedAt: time.Now().UTC().Add(time.Second).Format(time.RFC3339Nano), Evidence: "synthetic fixture: processes independently checked", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}); err != nil {
		t.Fatal(err)
	}
	return s, ledgerFixturePath(t), &benchCleanupRuntime{}
}
func mustBenchRecordState(t *testing.T, s *benchStore) bench.Run {
	t.Helper()
	_, run, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	return run
}
func TestBenchRecordImmutableDigestAndIdempotence(t *testing.T) {
	s, path, couch := benchRecordFixture(t)
	if err := recordBenchRun(context.Background(), s, path, couch.deps()); err != nil {
		t.Fatal(err)
	}
	if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseRecorded {
		t.Fatal(state.Phase())
	}
	rows, err := readBenchLedger(context.Background(), path)
	if err != nil || len(rows) != 1 {
		t.Fatalf("%+v %v", rows, err)
	}
	raw := ledgerBytes(t, filepath.Join(s.root, "observer", "result-evidence.json"))
	digest := sha256.Sum256(raw)
	if rows[0].RawDigest != hex.EncodeToString(digest[:]) || rows[0].Synthetic || !bytes.Contains(raw, []byte("1e400")) {
		t.Fatalf("row %+v evidence %s", rows[0], raw)
	}
	before := ledgerBytes(t, path)
	if err := recordBenchRun(context.Background(), s, "", couch.deps()); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, ledgerBytes(t, path)) {
		t.Fatal("idempotent record changed ledger")
	}
}
func TestBenchRecordLostAppendReplyRetainsExactRow(t *testing.T) {
	s, path, couch := benchRecordFixture(t)
	calls := 0
	deps := benchRecordDeps{Couch: couch.deps(), Append: func(ctx context.Context, path string, row bench.LedgerRow) error {
		calls++
		if err := recordBenchRow(ctx, path, row); err != nil {
			return err
		}
		return errors.New("lost append reply")
	}}
	if err := recordBenchRunWith(context.Background(), s, path, deps); err == nil {
		t.Fatal("lost reply confirmed")
	}
	if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseRecordUnknown {
		t.Fatal(state.Phase())
	}
	evidence := ledgerBytes(t, filepath.Join(s.root, "observer", "result-evidence.json"))
	before := ledgerBytes(t, path)
	if err := recordBenchRun(context.Background(), s, path, couch.deps()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !bytes.Equal(before, ledgerBytes(t, path)) || !bytes.Equal(evidence, ledgerBytes(t, filepath.Join(s.root, "observer", "result-evidence.json"))) {
		t.Fatal("retry changed immutable result")
	}
}
func TestBenchRecordRefusesChangedFactsAndUnsafePaths(t *testing.T) {
	s, path, couch := benchRecordFixture(t)
	if err := recordBenchRunWith(context.Background(), s, path, benchRecordDeps{Couch: couch.deps(), Append: func(context.Context, string, bench.LedgerRow) error { return errors.New("unavailable") }}); err == nil {
		t.Fatal("append failure hidden")
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	r.Data["measurements"] = json.RawMessage(`{}`)
	if err := s.write(r); err != nil {
		t.Fatal(err)
	}
	if err := recordBenchRun(context.Background(), s, path, couch.deps()); err == nil {
		t.Fatal("changed evidence accepted")
	}
	other, _, f := benchRecordFixture(t)
	unsafe := filepath.Join(other.root, "observer", "ledger.jsonl")
	if err := recordBenchRun(context.Background(), other, unsafe, f.deps()); err == nil {
		t.Fatal("ledger inside run accepted")
	}
	if _, err := os.Stat(filepath.Join(other.root, "observer", "result-evidence.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unsafe path caused snapshot effects", err)
	}
}
func TestBenchRecordRequiresFreshRetirement(t *testing.T) {
	s, path, couch := benchRecordFixture(t)
	couch.unknown = true
	if err := recordBenchRun(context.Background(), s, path, couch.deps()); err == nil {
		t.Fatal("unknown census accepted")
	}
	if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseStopped {
		t.Fatal(state.Phase())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := recordBenchRun(ctx, s, path, couch.deps()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestBenchRecordRefusesExistingEvidenceReplacement(t *testing.T) {
	for _, kind := range []string{"partial", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			s, path, couch := benchRecordFixture(t)
			evidence := filepath.Join(s.root, "observer", "result-evidence.json")
			if kind == "partial" {
				if err := os.WriteFile(evidence, []byte("partial"), 0600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Symlink(path, evidence); err != nil {
					t.Fatal(err)
				}
			}
			if err := recordBenchRun(context.Background(), s, path, couch.deps()); err == nil {
				t.Fatal("existing evidence replaced")
			}
			if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseStopped {
				t.Fatal(state.Phase())
			}
			if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("ledger effect: %v", err)
			}
		})
	}
}

func TestBenchRecordUsesBoundRuntimeLedger(t *testing.T) {
	s, couch := benchCleanupFixture(t, t.TempDir(), "run", time.Now().UTC().Add(-time.Hour))
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	r.Events = r.Events[:len(r.Events)-2] // Fixture's accepted recording is beyond this test's boundary.
	if err := s.write(r); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("measurements", ledgerFixtureRow(t, r.ID).Measurements); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("acceptance", bench.Acceptance{Status: "unavailable", Reason: "stopped"}); err != nil {
		t.Fatal(err)
	}
	if err := recordBenchRun(context.Background(), s, ledgerFixturePath(t), couch.deps()); err == nil {
		t.Fatal("different explicit ledger overrode bound runtime")
	}
	if _, err := os.Lstat(filepath.Join(s.root, "observer", "result-evidence.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("mismatch caused snapshot effects: %v", err)
	}
	if err := recordBenchRun(context.Background(), s, "", couch.deps()); err != nil {
		t.Fatal(err)
	}
	var runtime benchRuntimeState
	if err := json.Unmarshal(r.Data["runtime"], &runtime); err != nil {
		t.Fatal(err)
	}
	rows, err := readBenchLedger(context.Background(), runtime.Ledger)
	if err != nil || len(rows) != 1 {
		t.Fatalf("%+v %v", rows, err)
	}
}
