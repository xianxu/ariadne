package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

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

func TestBenchStoreReplaysAndPreservesFutureFields(t *testing.T) {
	root := filepath.Join(t.TempDir(), "run")
	s, err := createBenchStore(root, benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(s.recordPath())
	if err != nil {
		t.Fatal(err)
	}
	var extension map[string]json.RawMessage
	if err := json.Unmarshal(raw, &extension); err != nil {
		t.Fatal(err)
	}
	extension["future"] = json.RawMessage(`{"nested":{"magnitude":1e400}}`)
	raw, err = json.Marshal(extension)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.recordPath(), raw, 0600); err != nil {
		t.Fatal(err)
	}
	record, state, err := s.read()
	if err != nil || state.Phase() != bench.PhaseNew {
		t.Fatalf("read: %+v %v", state, err)
	}
	e := bench.Event{ID: "begin", RunID: record.ID, OperationID: "prepare", Kind: bench.BeginPrepare}
	state, effects, err := s.apply(e)
	if err != nil || state.Phase() != bench.PhasePreparing || len(effects) != 1 || effects[0].Kind != bench.Prepare {
		t.Fatalf("transition: %+v %+v %v", state, effects, err)
	}
	if _, effects, err := s.apply(e); err != nil || len(effects) != 0 {
		t.Fatalf("repeat reauthorized a preparation effect: %+v %v", effects, err)
	}
	record, _, err = s.read()
	if err != nil || !strings.Contains(string(record.Unknown["future"]), "1e400") || len(record.Events) != 1 {
		t.Fatalf("lost extension/event identity: %+v %v", record, err)
	}
}

func TestBenchStoreRefusesCollisionAndConcurrentOwner(t *testing.T) {
	root := filepath.Join(t.TempDir(), "run")
	s, err := createBenchStore(root, benchFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := createBenchStore(root, benchFixtureTask(t)); err == nil {
		t.Fatal("adopted an existing run as a new run")
	}
	release, err := s.acquireRun()
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := s.acquireRun(); err == nil {
		t.Fatal("two observers acquired the same lifetime")
	}
}

func TestBenchStoreNeverCreatesInsideAliasedSource(t *testing.T) {
	task := benchFixtureTask(t)
	physical, err := filepath.EvalSymlinks(task.SourceRepo)
	if err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(physical, alias); err != nil {
		t.Fatal(err)
	}
	task.SourceRepo = alias
	root := filepath.Join(physical, "must-not-create")
	if _, err := createBenchStore(root, task); err == nil {
		t.Fatal("created a bench run inside the source through an alias")
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("refusal modified source: %v", err)
	}
}

func TestBenchStoreRejectsOversizedEnvelopeBeforeCreatingRun(t *testing.T) {
	task := benchFixtureTask(t)
	raw, err := bench.EncodeTask(task)
	if err != nil {
		t.Fatal(err)
	}
	task.Spec += strings.Repeat("x", bench.TaskRecordLimit-len(raw))
	if err := bench.ValidateTask(task); err != nil {
		t.Fatalf("fixture must fit the task bound: %v", err)
	}
	root := filepath.Join(t.TempDir(), "oversized")
	if _, err := createBenchStore(root, task); err == nil {
		t.Fatal("unbounded enclosing run record accepted")
	}
	if _, err := os.Lstat(root); !os.IsNotExist(err) {
		t.Fatalf("writer-bound refusal left partial state: %v", err)
	}
}
