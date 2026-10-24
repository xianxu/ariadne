package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func TestBenchControlPublicProposalRequiresExactInvoker(t *testing.T) {
	s, f, d, _ := benchObserverFixture(t)
	_ = runBenchObserver(context.Background(), s, f, d)
	r, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := loadBenchRuntime(r, state)
	if err != nil {
		t.Fatal(err)
	}
	h := runtime.Hub
	t.Chdir(h.Identity.Root)
	for key, value := range map[string]string{"COUCH_TREE": h.Identity.Root, "COUCH_THREAD_SCOPE": h.Scope, "COUCH_THREAD_TAG": h.Tag, "PAIR_LAUNCH_NONCE": h.Identity.Generation, "PAIR_SESSION_NAME": h.Identity.Session} {
		t.Setenv(key, value)
	}
	if _, err := recordBenchProposal(context.Background(), "decision", r.ID, "ruling", "original requirement clarified", "", d.Couch); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PAIR_LAUNCH_NONCE", "replacement")
	if _, err := recordBenchProposal(context.Background(), "finish", r.ID, "", "", strings.Repeat("a", 40), d.Couch); err == nil {
		t.Fatal("replacement invoker accepted")
	}
	public, err := loadBenchContext(context.Background(), runtime.Worker.Identity.Root)
	if err != nil {
		t.Fatal(err)
	}
	events, err := readBenchAudit(context.Background(), public)
	if err != nil || len(events) != 1 || events[0].Session != h.Identity.Session || events[0].Generation != h.Identity.Generation {
		t.Fatalf("proposal binding %+v %v", events, err)
	}
}

func TestBenchControlRetirementRequiresEvidenceAndVacantRoots(t *testing.T) {
	s, f, d, _ := benchObserverFixture(t)
	_ = runBenchObserver(context.Background(), s, f, d)
	r, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := loadBenchRuntime(r, state)
	if err != nil {
		t.Fatal(err)
	}
	proof := benchRetirement{Version: 1, RunID: r.ID, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, ConfirmedBy: "fixture operator", ConfirmedAt: time.Now().UTC().Format(time.RFC3339Nano), Evidence: "synthetic: exact actors parked and all descendants verified gone", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	p := filepath.Join(t.TempDir(), "retired.json")
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := acceptBenchRetirement(context.Background(), s, p, d.Couch); err == nil {
		t.Fatal("live actors accepted as retired")
	}
	empty := benchCouchDeps{Command: func(context.Context, ...string) ([]byte, error) {
		return []byte(`{"Code":"ok","Error":"","Actors":[]}`), nil
	}}
	if err := acceptBenchRetirement(context.Background(), s, p, empty); err != nil {
		t.Fatal(err)
	}
	_, state, err = s.read()
	if err != nil || !state.RuntimeStopped() || state.Phase() != bench.PhaseStopped {
		t.Fatalf("retirement not confirmed: %s %v", state.Phase(), err)
	}
	if err := verifyBenchRetirement(context.Background(), s, empty); err != nil {
		t.Fatal(err)
	}
	if err := acceptBenchRetirement(context.Background(), s, p, empty); err != nil {
		t.Fatal("same proof not idempotent", err)
	}
}

func TestBenchControlCanRetirePreparedUndispatchedRun(t *testing.T) {
	s, err := prepareFixture(t, preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	proof := benchRetirement{Version: 1, RunID: r.ID, Hub: bench.RuntimeIdentity{Root: filepath.Join(s.root, "public", "hub")}, Worker: bench.RuntimeIdentity{Root: filepath.Join(s.root, "public", "work")}, ConfirmedBy: "fixture operator", ConfirmedAt: time.Now().UTC().Format(time.RFC3339Nano), Evidence: "synthetic: no dispatch; any provisioned processes independently retired", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	p := filepath.Join(t.TempDir(), "retired.json")
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, raw, 0600); err != nil {
		t.Fatal(err)
	}
	d := benchCouchDeps{Command: func(context.Context, ...string) ([]byte, error) {
		return []byte(`{"Code":"ok","Error":"","Actors":[]}`), nil
	}}
	if err := acceptBenchRetirement(context.Background(), s, p, d); err != nil {
		t.Fatal(err)
	}
	_, state, err := s.read()
	if err != nil || !state.RuntimeStopped() || state.Phase() != bench.PhaseStopped {
		t.Fatalf("prepared run not terminal: %s %v", state.Phase(), err)
	}
	if err := verifyBenchRetirement(context.Background(), s, d); err != nil {
		t.Fatal(err)
	}
}
