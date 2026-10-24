package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func TestBenchStepPreparePublishesNativeImmutableTemplate(t *testing.T) {
	task := preparedFixtureTask(t)
	run := t.TempDir()
	dir := filepath.Join(run, "prepare")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("METIS_RUN_DIR", run)
	t.Setenv("METIS_STEP_DIR", dir)
	t.Setenv("METIS_STEP_ID", "prepare")
	t.Chdir(dir)
	raw, err := json.Marshal(benchStepInput{Version: 1, Task: task, ToolSource: task.SourceRepo, ToolSHA256: strings.Repeat("a", 64), ToolRevision: strings.Repeat("b", 40)})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("with.json", raw, 0600); err != nil {
		t.Fatal(err)
	}
	// A stateful dependency double stands only for the build/read-set channel;
	// template preparation still validates and reads real local Git/oracle data.
	probes := 0
	d := benchStepDeps{Provenance: func(_ context.Context, f benchStepFolder, source string) error {
		probes++
		return writeBenchJSONExclusive(filepath.Join(f.Dir, "tool-provenance.json"), map[string]string{"source": source})
	}}
	if err := runBenchStepWith(context.Background(), "prepare", io.Discard, d); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"template.tar", "preparation.json", "metrics.json", "step-intent.json"} {
		if _, err := readBenchInput(filepath.Join(dir, path), benchTemplateArchiveLimit); err != nil {
			t.Fatal(path, err)
		}
	}
	if _, err := os.Lstat("runtime"); !os.IsNotExist(err) {
		t.Fatal("preparation created live roots", err)
	}
	before, err := os.ReadFile("preparation.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := runBenchStepWith(context.Background(), "prepare", io.Discard, d); err == nil {
		t.Fatal("duplicate preparation ran")
	}
	after, _ := os.ReadFile("preparation.json")
	if string(before) != string(after) || probes != 2 {
		t.Fatal("retry changed evidence")
	}
}

func TestBenchStepUpstreamRefusesTraversalAndAlias(t *testing.T) {
	run := t.TempDir()
	f := benchStepFolder{Run: run, ID: "score"}
	if err := os.Symlink(t.TempDir(), filepath.Join(run, "session")); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"../source", "session", "score", ""} {
		if _, err := benchUpstreamFolder(f, id); err == nil {
			t.Fatal("unsafe upstream admitted", id)
		}
	}
}

func TestBenchStepSessionEvidenceRejectsInvalidRuntime(t *testing.T) {
	task := benchFixtureTask(t)
	f := benchStepFolder{Run: t.TempDir(), Identity: strings.Repeat("a", 32)}
	session := filepath.Join(f.Run, "session")
	digest, err := benchTaskDigest(task)
	if err != nil {
		t.Fatal(err)
	}
	p := benchPreparation{Version: 1, TaskDigest: digest, TemplateDigest: strings.Repeat("b", 64)}
	identity := func(slot, leaf string) bench.RuntimeIdentity {
		return bench.RuntimeIdentity{Slot: slot, Session: leaf, Generation: "g", Root: filepath.Join(session, "runtime", "public", leaf), Agent: task.Worker.Agent, Model: task.Worker.Model, Effort: task.Worker.Effort}
	}
	now := time.Now().UTC()
	runtime := benchRuntimeState{Version: 1, RunID: f.Identity, Since: now.Format(time.RFC3339Nano), Hub: benchRuntimeActor{Identity: identity("hub:0", "hub")}, Worker: benchRuntimeActor{Identity: identity("work:0", "work")}}
	e := benchSessionEvidence{Version: 1, MetisRun: f.Run, Identity: f.Identity, TaskDigest: digest, TemplateDigest: p.TemplateDigest, Outcome: "stopped", Runtime: runtime, Retirement: benchRetirement{Version: 1, RunID: f.Identity, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, ProcessesGone: true, ConfirmedBy: "synthetic operator", ConfirmedAt: now.Add(time.Minute).Format(time.RFC3339Nano), Evidence: "synthetic supported retirement"}}
	if err := validateBenchSessionEvidence(f, p, e, session, task); err != nil {
		t.Fatal("valid retired stopped evidence refused", err)
	}
	for _, kind := range []string{"empty", "outside", "profile"} {
		bad := e
		switch kind {
		case "empty":
			bad.Runtime.Hub.Identity = bench.RuntimeIdentity{}
		case "outside":
			bad.Runtime.Hub.Identity.Root = filepath.Join(t.TempDir(), "other")
		case "profile":
			bad.Runtime.Hub.Identity.Model = "unapproved-model"
		}
		bad.Retirement.Hub = bad.Runtime.Hub.Identity
		if err := validateBenchSessionEvidence(f, p, bad, session, task); err == nil {
			t.Fatal("invalid runtime evidence admitted", kind)
		}
	}
}

func TestBenchStepJSONWriterRefusesOversizeBeforePublication(t *testing.T) {
	path := filepath.Join(t.TempDir(), "evidence.json")
	if err := writeBenchJSONExclusive(path, map[string]string{"payload": strings.Repeat("x", benchRecordLimit)}); err == nil {
		t.Fatal("oversize evidence published")
	}
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatal("oversize refusal wrote output", err)
	}
}

func TestBenchStepRefusesOutsideMetisBeforeEffects(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, k := range []string{"METIS_STEP_DIR", "METIS_RUN_DIR", "METIS_STEP_ID"} {
		t.Setenv(k, "")
	}
	if err := runBenchStep(context.Background(), "prepare", io.Discard); err == nil || !strings.Contains(err.Error(), "Metis") {
		t.Fatal("unbound execution admitted", err)
	}
	entries, _ := os.ReadDir(".")
	if len(entries) != 0 {
		t.Fatal("refusal wrote files")
	}
}

func TestBenchStepRejectsMovedDirectory(t *testing.T) {
	root := t.TempDir()
	step := filepath.Join(root, "prepare")
	if err := os.Mkdir(step, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("METIS_RUN_DIR", root)
	t.Setenv("METIS_STEP_DIR", step)
	t.Setenv("METIS_STEP_ID", "prepare")
	t.Chdir(root)
	if err := runBenchStep(context.Background(), "prepare", io.Discard); err == nil {
		t.Fatal("moved cwd admitted")
	}
}

func TestBenchStepRejectsDuplicateInput(t *testing.T) {
	root := t.TempDir()
	step := filepath.Join(root, "prepare")
	if err := os.Mkdir(step, 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("METIS_RUN_DIR", root)
	t.Setenv("METIS_STEP_DIR", step)
	t.Setenv("METIS_STEP_ID", "prepare")
	t.Chdir(step)
	if err := os.WriteFile("with.json", []byte(`{"version":1,"version":1}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runBenchStep(context.Background(), "prepare", io.Discard); err == nil {
		t.Fatal("duplicate input admitted")
	}
	if _, err := os.Lstat("metrics.json"); !os.IsNotExist(err) {
		t.Fatal("invalid input published metrics", err)
	}
}
