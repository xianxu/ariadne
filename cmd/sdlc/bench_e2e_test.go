package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
)

// Synthetic evidence exercises orchestration and storage. It never certifies a
// real agent launch, native Couch service, full suite run, or benchmark result.
func TestBenchSyntheticEndToEnd(t *testing.T) {
	ctx := context.Background()
	s, candidate, public := benchAdmissionFixture(t)
	r, state, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	sourceHead := prepareTestGit(t, r.Task.SourceRepo, "rev-parse", "HEAD")
	sourceStatus := prepareTestGit(t, r.Task.SourceRepo, "status", "--porcelain")
	sourceObjects := prepareTestGit(t, r.Task.SourceRepo, "count-objects", "-v")
	suiteCalls := 0
	facts, err := verifyBenchWouldLandWithSuite(ctx, s, candidate, func(_ context.Context, root, head, name string) (benchLocalCheckReceipt, error) {
		suiteCalls++
		if root != public.WorkRoot || head != candidate || name != "suite" {
			t.Fatal("suite evidence changed subject")
		}
		return benchLocalCheckReceipt{Version: 1, RunID: r.ID, WorkRoot: root, OriginIdentity: public.OriginIdentity, Head: head, Check: name, Record: localRunRecord{Name: name, SHA: head, LogSHA256: strings.Repeat("a", 64)}}, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if suiteCalls != 1 || facts.CompletionEvidence == candidate || facts.ReviewedHead != candidate || facts.ReviewVerdict != "SHIP" {
		t.Fatalf("native admission did not bind real close metadata: %+v", facts)
	}
	h, w := state.Bindings()
	if _, _, err := s.apply(bench.Event{ID: "e2e-proof", RunID: r.ID, Kind: bench.ProofAccepted, OperationID: "proof", Hub: h, Worker: w, Facts: facts}); err != nil {
		t.Fatal(err)
	}
	since := time.Now().UTC().Add(-10 * time.Second)
	ledger := ledgerFixturePath(t)
	runtime := benchRuntimeState{Version: 1, RunID: r.ID, Since: since.Format(time.RFC3339Nano), Deadline: since.Add(time.Hour).Format(time.RFC3339Nano), Ledger: ledger, Provision: json.RawMessage(`{"synthetic":true}`), Hub: benchE2EActor(t, h, "tl", since), Worker: benchE2EActor(t, w, "worker", since)}
	if err := s.putData("runtime", runtime); err != nil {
		t.Fatal(err)
	}
	census := &benchCleanupRuntime{}
	if err := scoreBenchRun(ctx, s, census.deps()); err == nil {
		t.Fatal("empty actor census substituted for explicit retirement")
	}
	proof := benchRetirement{Version: 1, RunID: r.ID, Hub: h, Worker: w, ConfirmedBy: "synthetic operator", ConfirmedAt: since.Add(time.Minute).Format(time.RFC3339Nano), Evidence: "synthetic fixture: exact owned generations and descendants confirmed gone", ProcessesGone: true, Survivors: []string{}, ReviewerSources: []benchSourceInput{}}
	retirementPath := filepath.Join(t.TempDir(), "retirement.json")
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(retirementPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	census.actors = []benchActorSnapshot{{Binding: runtime.Worker.binding(), Known: true, Supported: true}}
	if err := cancelBenchRun(ctx, s, retirementPath, census.deps()); err == nil {
		t.Fatal("live replacement accepted as retired")
	}
	if mustBenchRecordState(t, s).RuntimeStopped() {
		t.Fatal("refused retirement changed runtime custody")
	}
	census.actors = nil
	if err := cancelBenchRun(ctx, s, retirementPath, census.deps()); err != nil {
		t.Fatal(err)
	}
	if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseWouldLand || !state.RuntimeStopped() || state.Candidate() != candidate {
		t.Fatalf("retirement lost admitted subject: %s", state.Phase())
	}
	var roles []string
	score := benchScoringDeps{Couch: census.deps(), Evaluate: func(ctx context.Context, task bench.Task, work, head, observer string) (bench.Acceptance, error) {
		if work != public.WorkRoot || head != candidate || observer != filepath.Join(s.root, "observer") {
			t.Fatal("scorer changed admitted inputs")
		}
		return evaluateBenchAcceptanceWith(ctx, task, work, head, observer, benchScoreDeps{Execute: func(_ context.Context, c benchScoreCommand) (benchScoreRun, error) {
			if mustBenchRecordState(t, s).Phase() != bench.PhaseScoring {
				t.Fatal("execution before durable scoring authorization")
			}
			if !runnerArtifactProtected(c.Root) {
				t.Fatal("synthetic command lacks genuine native worktree lease")
			}
			if strings.Join(c.Argv, " ") != strings.Join(task.Acceptance.Argv, " ") {
				t.Fatal("hidden argv changed")
			}
			if _, err := os.Stat(filepath.Join(c.Root, "tests", "acceptance_test.go")); err != nil {
				t.Fatal("captured acceptance missing", err)
			}
			product, err := os.ReadFile(filepath.Join(c.Root, "src", "main.txt"))
			if err != nil {
				t.Fatal(err)
			}
			roles = append(roles, c.Role)
			exit := 1
			if strings.HasPrefix(string(product), "fixed") {
				exit = 0
			}
			if err := os.WriteFile(c.LogPath, []byte(fmt.Sprintf("synthetic %s exit=%d\n", c.Role, exit)), 0600); err != nil {
				t.Fatal(err)
			}
			return benchScoreRun{Exit: exit, Quiet: true}, nil
		}})
	}}
	if err := scoreBenchRunWith(ctx, s, score); err != nil {
		t.Fatal(err)
	}
	if strings.Join(roles, ",") != "parent,landing,candidate" || mustBenchRecordState(t, s).Phase() != bench.PhaseScored {
		t.Fatalf("oracle sequence %v", roles)
	}
	scored, _, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	var measurements bench.Measurements
	if err := json.Unmarshal(scored.Data["measurements"], &measurements); err != nil {
		t.Fatal(err)
	}
	if measurements.Worker.InputTokens == nil || *measurements.Worker.InputTokens != 10 || measurements.TL.OutputTokens == nil || *measurements.TL.OutputTokens != 2 || measurements.Reviewer.InputTokens != nil || measurements.ActiveHours != nil || measurements.GrossNanos != int64(time.Minute) {
		t.Fatalf("collector fabricated or lost dimensions: %+v", measurements)
	}
	var captured []benchCapturedSource
	if err := json.Unmarshal(scored.Data["captured_sources"], &captured); err != nil || len(captured) != 2 {
		t.Fatalf("captured sources %+v: %v", captured, err)
	}
	for _, source := range captured {
		if !source.Complete || !benchPathWithin(filepath.Join(s.root, "observer"), source.File) || source.File == source.OriginalFile {
			t.Fatalf("source not pinned privately: %+v", source)
		}
		data, err := os.ReadFile(source.File)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != source.Digest {
			t.Fatal("captured prefix digest mismatch")
		}
		if err := os.WriteFile(source.OriginalFile, []byte("live synthetic source changed after capture\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := scoreBenchRunWith(ctx, s, score); err != nil || len(roles) != 3 {
		t.Fatal("scored retry executed a second command", err)
	}
	recordDeps := benchRecordDeps{Couch: census.deps(), Synthetic: true}
	if err := recordBenchRunWith(ctx, s, "", recordDeps); err != nil {
		t.Fatal(err)
	}
	ledgerBefore := ledgerBytes(t, ledger)
	evidencePath := filepath.Join(s.root, "observer", "result-evidence.json")
	evidence := ledgerBytes(t, evidencePath)
	if err := recordBenchRunWith(ctx, s, "", recordDeps); err != nil || !bytes.Equal(ledgerBefore, ledgerBytes(t, ledger)) {
		t.Fatal("record retry changed durable row", err)
	}
	rows, err := readBenchLedger(ctx, ledger)
	if err != nil || len(rows) != 1 {
		t.Fatalf("ledger %+v %v", rows, err)
	}
	digest := sha256.Sum256(evidence)
	row := rows[0]
	if !row.Synthetic || row.RunID != r.ID || row.Outcome != "would-land" || row.RawDigest != hex.EncodeToString(digest[:]) || row.Acceptance.CommandPassed == nil || !*row.Acceptance.CommandPassed || !row.Acceptance.ParentRed || !row.Acceptance.LandingGreen {
		t.Fatalf("synthetic result attribution: %+v", row)
	}
	// Leave the process cwd before removing its public working tree. The fixture
	// owns restoring the original cwd through t.Chdir's registered cleanup.
	if err := os.Chdir(filepath.Join(s.root, "observer")); err != nil {
		t.Fatal(err)
	}
	if err := cleanBenchRun(ctx, s, census.deps()); err != nil {
		t.Fatal(err)
	}
	if state := mustBenchRecordState(t, s); state.Phase() != bench.PhaseCleaned {
		t.Fatal(state.Phase())
	}
	if _, err := os.Lstat(filepath.Join(s.root, "public")); !os.IsNotExist(err) {
		t.Fatal("public replay tree retained", err)
	}
	if !bytes.Equal(evidence, ledgerBytes(t, evidencePath)) || !bytes.Equal(ledgerBefore, ledgerBytes(t, ledger)) {
		t.Fatal("cleanup changed raw evidence or external ledger")
	}
	for _, source := range captured {
		if _, err := os.Stat(source.File); err != nil {
			t.Fatal("cleanup deleted captured evidence", err)
		}
	}
	if prepareTestGit(t, r.Task.SourceRepo, "rev-parse", "HEAD") != sourceHead || prepareTestGit(t, r.Task.SourceRepo, "status", "--porcelain") != sourceStatus || prepareTestGit(t, r.Task.SourceRepo, "count-objects", "-v") != sourceObjects {
		t.Fatal("replay changed original repository")
	}
	if census.calls < 5 {
		t.Fatal("lifecycle omitted fresh native census boundaries")
	}
}

func benchE2EActor(t *testing.T, identity bench.RuntimeIdentity, role string, since time.Time) benchRuntimeActor {
	t.Helper()
	native := role + "-synthetic-native"
	transcript := filepath.Join(t.TempDir(), native+".jsonl")
	stamp := func(offset time.Duration) string { return since.Add(offset).Format(time.RFC3339Nano) }
	body := fmt.Sprintf("{\"type\":\"session_meta\",\"timestamp\":%q,\"payload\":{\"id\":%q,\"cwd\":%q}}\n", stamp(0), native, identity.Root) +
		fmt.Sprintf("{\"type\":\"turn_context\",\"timestamp\":%q,\"payload\":{\"model\":%q,\"effort\":%q,\"turn_id\":\"turn\"}}\n", stamp(time.Second), identity.Model, identity.Effort) +
		fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":%q,\"payload\":{\"thread_id\":%q,\"turn_id\":\"turn\",\"response_id\":%q,\"usage\":{\"input_tokens\":10,\"output_tokens\":2,\"total_tokens\":12}}}\n", stamp(2*time.Second), native, role+"-response")
	if err := os.WriteFile(transcript, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return benchRuntimeActor{Identity: identity, NativeSession: native, Transcripts: []string{transcript}, PID: 123, ProcessStart: "synthetic", Scope: "synthetic", Tag: role}
}
