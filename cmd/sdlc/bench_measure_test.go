package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

func benchMeasureFixture(t *testing.T, configure ...func(*benchRuntimeState, *benchRetirement)) (*benchSession, *benchCleanupRuntime, benchRuntimeState) {
	t.Helper()
	s, err := prepareFixture(t, preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	since := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	makeActor := func(role, slot, root string, profile bench.AgentProfile) benchRuntimeActor {
		native := role + "-native"
		source := filepath.Join(t.TempDir(), native+".jsonl")
		body := fmt.Sprintf("{\"type\":\"session_meta\",\"timestamp\":\"2026-10-23T10:00:00Z\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", native, root) +
			fmt.Sprintf("{\"type\":\"turn_context\",\"timestamp\":\"2026-10-23T10:00:01Z\",\"payload\":{\"model\":%q,\"effort\":%q,\"turn_id\":\"turn\"}}\n", profile.Model, profile.Effort) +
			fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-23T10:00:02Z\",\"payload\":{\"thread_id\":%q,\"turn_id\":\"turn\",\"response_id\":%q,\"usage\":{\"input_tokens\":10,\"output_tokens\":2,\"total_tokens\":12}}}\n", native, role+"-response")
		if err := os.WriteFile(source, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return benchRuntimeActor{Identity: bench.RuntimeIdentity{Slot: slot, Session: role + "-session", Generation: role + "-generation", Root: root, Agent: profile.Agent, Model: profile.Model, Effort: profile.Effort}, NativeSession: native, Transcripts: []string{source}, PID: 42, ProcessStart: "synthetic", Scope: "synthetic", Tag: "synthetic"}
	}
	runtime := benchRuntimeState{Version: 1, RunID: r.ID, Since: since.Format(time.RFC3339Nano), Deadline: since.Add(time.Hour).Format(time.RFC3339Nano), Provision: json.RawMessage(`{"synthetic":true}`), Hub: makeActor("tl", "repo:0", filepath.Join(s.root, "public", "hub"), r.Task.TL), Worker: makeActor("worker", "repo:1", filepath.Join(s.root, "public", "work"), r.Task.Worker)}
	for i, step := range []struct {
		kind bench.EventKind
		op   string
	}{{bench.BeginDispatch, "dispatch"}, {bench.DispatchConfirmed, "dispatch"}, {bench.Stop, "stop"}, {bench.CancelConfirmed, "stop"}} {
		if _, _, err := s.apply(bench.Event{ID: fmt.Sprint("measure-", i), RunID: r.ID, OperationID: step.op, Kind: step.kind, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, Reason: "synthetic fixture"}); err != nil {
			t.Fatal(err)
		}
	}
	proof := benchRetirement{Version: 1, RunID: r.ID, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, ConfirmedBy: "fixture operator", ConfirmedAt: since.Add(time.Minute).Format(time.RFC3339Nano), Evidence: "synthetic retirement", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	for _, change := range configure {
		change(&runtime, &proof)
	}
	if err := s.putData("runtime", runtime); err != nil {
		t.Fatal(err)
	}
	if err := s.putData("retirement", proof); err != nil {
		t.Fatal(err)
	}
	return s, &benchCleanupRuntime{}, runtime
}

func TestMeasureBenchRunCapturesOnceAndPersistsImmutableMetrics(t *testing.T) {
	s, couch, runtime := benchMeasureFixture(t)
	got, err := measureBenchRun(context.Background(), s, couch.deps())
	if err != nil {
		t.Fatal(err)
	}
	if got.Worker.InputTokens == nil || *got.Worker.InputTokens != 10 || got.TL.InputTokens == nil || *got.TL.InputTokens != 10 || got.Reviewer.InputTokens != nil || got.ActiveHours != nil || got.GrossNanos != int64(time.Minute) {
		t.Fatalf("measurements %+v", got)
	}
	if got.Autonomous || !strings.Contains(strings.Join(got.ValidityIssues, " "), "autonomy") {
		t.Fatalf("autonomy guessed from task label: %+v", got)
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"captured_sources", "collection", "measurements"} {
		if len(r.Data[key]) == 0 {
			t.Fatalf("missing immutable %s", key)
		}
	}
	var sources []benchCapturedSource
	if err := json.Unmarshal(r.Data["captured_sources"], &sources); err != nil || len(sources) != 2 {
		t.Fatalf("capture pins %+v %v", sources, err)
	}
	for _, actor := range []benchRuntimeActor{runtime.Hub, runtime.Worker} {
		for _, source := range actor.Transcripts {
			if err := os.WriteFile(source, []byte("live source changed\n"), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	again, err := measureBenchRun(context.Background(), s, couch.deps())
	if err != nil || !reflect.DeepEqual(again, got) {
		t.Fatalf("reused evidence changed: %+v %v", again, err)
	}
	for _, source := range sources {
		if source.File == source.OriginalFile || !strings.HasPrefix(source.File, filepath.Join(s.root, "observer")) {
			t.Fatalf("capture not private: %+v", source)
		}
	}

}

func TestMeasureBenchRunRefusesUnretiredAndAmbiguousSources(t *testing.T) {
	s, couch, runtime := benchMeasureFixture(t)
	couch.actors = []benchActorSnapshot{{Binding: runtime.Worker.binding(), Known: true, Supported: true}}
	if _, err := measureBenchRun(context.Background(), s, couch.deps()); err == nil {
		t.Fatal("live actor measured")
	}
	r, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Data["captured_sources"]) != 0 || len(r.Data["measurements"]) != 0 {
		t.Fatal("effects preceded retirement verification")
	}
	ambiguous, quiet, _ := benchMeasureFixture(t, func(runtime *benchRuntimeState, _ *benchRetirement) {
		runtime.Worker.Transcripts = append(runtime.Worker.Transcripts, filepath.Join(t.TempDir(), "another-copy.jsonl"))
	})
	if _, err := measureBenchRun(context.Background(), ambiguous, quiet.deps()); err == nil {
		t.Fatal("ambiguous same-session source selection guessed")
	}

}

func TestMeasureBenchRunUsesOnlyExactAutonomyAttestation(t *testing.T) {
	for _, exact := range []bool{true, false} {
		t.Run(fmt.Sprint(exact), func(t *testing.T) {
			s, couch, _ := benchMeasureFixture(t, func(runtime *benchRuntimeState, proof *benchRetirement) {
				proof.Autonomy = &benchAutonomyAttestation{Since: runtime.Since, Until: proof.ConfirmedAt, NoOperatorInput: true, Evidence: "synthetic explicit no-input attestation"}
				if !exact {
					proof.Autonomy.Since = "2026-10-23T09:59:59Z"
				}
			})
			got, err := measureBenchRun(context.Background(), s, couch.deps())
			if err != nil {
				t.Fatal(err)
			}
			if got.Autonomous != exact || strings.Contains(strings.Join(got.ValidityIssues, " "), "autonomy:") == exact {
				t.Fatalf("attestation coverage %+v", got)
			}
		})
	}
}

func TestMeasureBenchUndispatchedRootOnlyIsUnknown(t *testing.T) {
	s, err := prepareFixture(t, preparedFixtureTask(t))
	if err != nil {
		t.Fatal(err)
	}
	r, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := benchRetirementSubject(r, state)
	if err != nil {
		t.Fatal(err)
	}
	since, err := time.Parse(time.RFC3339Nano, r.CreatedAt)
	if err != nil {
		t.Fatal(err)
	}
	retirement := benchRetirement{Version: 1, RunID: r.ID, Hub: runtime.Hub.Identity, Worker: runtime.Worker.Identity, ConfirmedBy: "fixture", ConfirmedAt: since.Add(time.Minute).Format(time.RFC3339Nano), Evidence: "prepared roots never dispatched", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	if err := s.putData("retirement", retirement); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.apply(bench.Event{ID: "stop-before-dispatch", RunID: r.ID, Kind: bench.Stop, OperationID: "stop", Reason: "synthetic root-only retirement"}); err != nil {
		t.Fatal(err)
	}
	couch := &benchCleanupRuntime{}
	got, err := measureBenchRun(context.Background(), s, couch.deps())
	if err != nil {
		t.Fatal(err)
	}
	if got.Validity != "unverified" || got.Worker.InputTokens != nil || got.TL.InputTokens != nil || got.Reviewer.InputTokens != nil || got.ActiveHours != nil || got.GateRefusals != nil || got.ReviewRounds != nil || got.Autonomous || got.Leaks.Status != "unknown" {
		t.Fatalf("undispatched metrics fabricated: %+v", got)
	}
}
