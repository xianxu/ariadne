package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func TestBenchSessionExclusiveRootAndImmutableFacts(t *testing.T) {
	root := filepath.Join(t.TempDir(), "session")
	task := preparedFixtureTask(t)
	s, err := createBenchSession(root, task, "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	task.Acceptance.Paths[0] = "tests/changed_test.go"
	snapshot, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Task.Acceptance.Paths[0] == "tests/changed_test.go" {
		t.Fatal("caller changed source task evidence")
	}
	if _, err = createBenchSession(root, preparedFixtureTask(t)); err == nil {
		t.Fatal("adopted existing root")
	}
	if err = s.putData("measurement", map[string]int{"b": 2, "a": 1}); err != nil {
		t.Fatal(err)
	}
	if err = s.putData("measurement", json.RawMessage(`{"a":1,"b":2}`)); err != nil {
		t.Fatal(err)
	}
	if err = s.putData("measurement", map[string]int{"a": 2}); err == nil {
		t.Fatal("overwrote observation")
	}
	if _, err = os.Stat(filepath.Join(root, "observer", "run.json")); !os.IsNotExist(err) {
		t.Fatal("created independent run record")
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	delete(r.Data, "measurement")
	r, _, _ = s.read()
	if len(r.Data["measurement"]) == 0 {
		t.Fatal("read exposed mutable facts")
	}
}

func TestBenchSessionDispatchIntentPrecedesEffectAndRefusesPriorIntent(t *testing.T) {
	s, err := createBenchSession(filepath.Join(t.TempDir(), "session"), preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, _, _ := s.read()
	for _, kind := range []bench.EventKind{bench.BeginPrepare, bench.Prepared} {
		if _, _, err = s.apply(bench.Event{ID: string(kind), RunID: r.ID, OperationID: "prepare", Kind: kind}); err != nil {
			t.Fatal(err)
		}
	}
	h := bench.RuntimeIdentity{Slot: "hub:0", Session: "h", Generation: "g", Root: filepath.Join(s.root, "hub"), Agent: "claude", Model: "opus", Effort: "high"}
	w := h
	w.Slot = "work:0"
	w.Root = filepath.Join(s.root, "work")
	if err = os.WriteFile(filepath.Join(s.root, "observer", "dispatch-intent.json"), []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	_, fx, err := s.apply(bench.Event{ID: "dispatch", RunID: r.ID, OperationID: "dispatch", Kind: bench.BeginDispatch, Hub: h, Worker: w})
	if err == nil || len(fx) != 0 {
		t.Fatal("previous dispatch intent authorized launch")
	}
	if err = os.Remove(filepath.Join(s.root, "observer", "dispatch-intent.json")); err != nil {
		t.Fatal(err)
	}
	_, fx, err = s.apply(bench.Event{ID: "dispatch", RunID: r.ID, OperationID: "dispatch", Kind: bench.BeginDispatch, Hub: h, Worker: w})
	if err != nil || len(fx) != 1 {
		t.Fatalf("dispatch authorization %v %v", fx, err)
	}
	if _, err = os.Stat(filepath.Join(s.root, "observer", "dispatch-intent.json")); err != nil {
		t.Fatal("dispatch effect lacks intent")
	}
}

func TestBenchSessionDetectsRootReplacementAndSymlinkEvidence(t *testing.T) {
	s, err := createBenchSession(filepath.Join(t.TempDir(), "session"), preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err = os.WriteFile(outside, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(s.root, "observer", "receipt.json")); err != nil {
		t.Fatal(err)
	}
	if err = s.putData("receipt", map[string]int{"x": 1}); err == nil {
		t.Fatal("followed evidence symlink")
	}
	if err = os.Rename(filepath.Join(s.root, "observer"), filepath.Join(s.root, "moved")); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(s.root, "observer"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, _, err = s.read(); err == nil {
		t.Fatal("accepted replaced observer root")
	}
}

type benchCleanupRuntime struct {
	actors  []benchActorSnapshot
	unknown bool
	calls   int
	onRead  func()
}

func (f *benchCleanupRuntime) deps() benchCouchDeps {
	return benchCouchDeps{Command: func(_ context.Context, args ...string) ([]byte, error) {
		f.calls++
		if f.onRead != nil {
			f.onRead()
		}
		if f.unknown {
			return nil, errors.New("supervisor unavailable")
		}
		if strings.Join(args, " ") != "--actors --json" {
			return nil, errors.New("unexpected runtime mutation")
		}
		actors := f.actors
		if actors == nil {
			actors = []benchActorSnapshot{}
		}
		return json.Marshal(benchActorsReply{Code: "ok", Actors: actors})
	}}
}

func benchFixtureTask(t *testing.T) bench.Task {
	t.Helper()
	root := t.TempDir()
	parent, landing := strings.Repeat("a", 40), strings.Repeat("b", 40)
	profile := bench.AgentProfile{Agent: "codex", Model: "gpt-6.1-sol", Effort: "medium"}
	return bench.Task{Version: 1, CorpusVersion: "v1", Repo: "fixture", ID: "000001", Type: "small-fix", Repeat: 1,
		SourceRepo: root, ParentOID: parent, LandingOID: landing, OriginalIssueOID: strings.Repeat("c", 40),
		Spec: "Fix the autonomous fixture.", DoneWhen: "The acceptance test passes.", OriginalHours: 0.2,
		Acceptance: bench.AcceptanceSpec{Patch: filepath.Join(root, "hidden.patch"), PatchSHA256: strings.Repeat("d", 64), Paths: []string{"tests/acceptance_test.go"}, Argv: []string{"make", "test"}, ParentOID: parent, LandingOID: landing},
		Worker:     profile, TL: profile, Reviewer: profile}
}

func TestBenchSessionRefusesUnsafePlacementAndConcurrentOwner(t *testing.T) {
	task := benchFixtureTask(t)
	if _, err := createBenchSession(filepath.Join(task.SourceRepo, "nested"), task); err == nil {
		t.Fatal("created session in source checkout")
	}
	if _, err := createBenchSession(filepath.Join(t.TempDir(), "bad-id"), task, "not-a-metis-identity"); err == nil {
		t.Fatal("accepted invalid repeat identity")
	}
	parent := t.TempDir()
	link := filepath.Join(t.TempDir(), "linked-parent")
	if err := os.Symlink(parent, link); err != nil {
		t.Fatal(err)
	}
	if _, err := createBenchSession(filepath.Join(link, "nested"), task); err == nil {
		t.Fatal("followed parent symlink")
	}
	s, err := createBenchSession(filepath.Join(t.TempDir(), "session"), task)
	if err != nil {
		t.Fatal(err)
	}
	release, err := s.acquireRun()
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if other, err := s.acquireRun(); err == nil {
		other()
		t.Fatal("two observers acquired session lifetime")
	}
}
