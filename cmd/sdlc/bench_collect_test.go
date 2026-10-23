package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/bench"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/reviewusage"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/usage"
)

func benchCollectionRun(t *testing.T, worker bench.RuntimeIdentity, candidates ...string) bench.Run {
	t.Helper()
	hub := worker
	hub.Root = t.TempDir()
	hub.Slot = "repo:0"
	hub.Session = "hub-session"
	run, err := bench.NewRun("collection-run")
	if err != nil {
		t.Fatal(err)
	}
	type runStep struct {
		kind bench.EventKind
		op   string
	}
	steps := []runStep{{bench.BeginPrepare, "prepare"}, {bench.Prepared, "prepare"}, {bench.BeginDispatch, "dispatch"}, {bench.DispatchConfirmed, "dispatch"}}
	if len(candidates) > 0 {
		steps = append(steps, runStep{bench.ProposeWouldLand, "proof"}, runStep{bench.ProofAccepted, "proof"})
	}
	steps = append(steps, runStep{bench.Stop, "stop"}, runStep{bench.CancelConfirmed, "stop"})
	for i, step := range steps {
		event := bench.Event{ID: fmt.Sprint(i), RunID: run.ID(), Kind: step.kind, OperationID: step.op, Hub: hub, Worker: worker, Reason: "synthetic fixture"}
		if step.kind == bench.ProofAccepted {
			candidate := candidates[0]
			event.Facts = bench.WouldLandFacts{Candidate: candidate, ObservedHead: candidate, PushedHead: candidate, PrivateOrigin: "file:/bench/origin.git", ExpectedOrigin: "file:/bench/origin.git", SuiteHead: candidate, SuiteReceiptID: "synthetic-suite", SuiteLogDigest: strings.Repeat("a", 64), SuitePassed: true, FullSuite: true, CompletionHead: candidate, ReviewedHead: candidate, CompletionEvidence: "synthetic-completion", CompletionToken: "synthetic-token", ReviewID: "synthetic-review", ReviewVerdict: "SHIP", CompletionBound: true}
		}
		run, _, err = bench.Step(run, event)
		if err != nil {
			t.Fatal(err)
		}
	}
	return run
}

func TestBenchCollectUsageIdentityAndUnknownFocusedTime(t *testing.T) {
	input := benchCaptureInput(t, "")
	since := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	body := `{"type":"session_meta","timestamp":"2026-10-23T10:00:00Z","payload":{"id":"native-session","cwd":"` + input.Runtime.Root + `"}}` + "\n" +
		`{"type":"turn_context","timestamp":"2026-10-23T10:00:01Z","payload":{"model":"gpt-test","turn_id":"turn-1"}}` + "\n" +
		`{"type":"token_usage_record","timestamp":"2026-10-23T10:00:02Z","payload":{"thread_id":"native-session","turn_id":"turn-1","response_id":"response-1","usage":{"input_tokens":10,"output_tokens":2,"cached_input_tokens":3,"total_tokens":12}}}` + "\n"
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
	if err != nil {
		t.Fatal(err)
	}
	task := benchFixtureTask(t)
	profile := bench.AgentProfile{Agent: input.Runtime.Agent, Model: input.Runtime.Model, Effort: input.Runtime.Effort}
	task.Worker, task.TL, task.Reviewer = profile, profile, profile
	args := benchCollectionInput{Task: task, Run: benchCollectionRun(t, input.Runtime), Sources: captured, Since: since, Until: since.Add(time.Minute), WorkerRepo: input.Runtime.Root}
	got, err := collectBenchMeasurements(context.Background(), args)
	if err != nil || !got.Valid || got.Measurements.Worker.InputTokens == nil || *got.Measurements.Worker.InputTokens != 10 || *got.Measurements.Worker.OutputTokens != 2 || *got.Measurements.Worker.CachedTokens != 3 {
		t.Fatalf("collection %+v %v", got, err)
	}
	if got.Measurements.ActiveHours != nil || got.Measurements.TL.InputTokens != nil || got.Measurements.Reviewer.InputTokens != nil {
		t.Fatal("unknown dimensions became zero")
	}
	if got.Sources[0].Effort != "" || got.Sources[0].EffortEvidence != "trusted-provisioning-unverified" {
		t.Fatalf("requested effort borrowed as evidence: %+v", got.Sources[0])
	}
	if got.Measurements.Validity != "unverified" {
		t.Fatalf("unknown effort/roles claimed valid: %+v", got.Measurements)
	}
	if got.Visits.NormalizedRecords != 3 || got.Visits.UsageRecords != 3 || got.Visits.UsageResponses != 1 || len(got.Digest) != 64 {
		t.Fatalf("unaccounted passes: %+v", got)
	}
	// The captured native model differs from a newly nominated descriptor.
	args.Task.Worker.Model = "other-model"
	got, err = collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Valid {
		t.Fatalf("profile drift admitted: %+v %v", got, err)
	}
	if got.Measurements.Validity != "invalid" || len(got.Measurements.ValidityIssues) == 0 {
		t.Fatalf("invalid profile not retained in typed evidence: %+v", got.Measurements)
	}
}

func TestBenchNormalizeGitHubToolDiagnostic(t *testing.T) {
	event := normalizeBenchTool("mcp__github__get_file_contents", "tool-1", "/bench/work", json.RawMessage(`{"owner":"example","repo":"source"}`))
	got := bench.ScanLeakEvents(bench.LeakScope{PreparedRoots: []string{"/bench/work"}, Complete: true}, []bench.ToolEvent{event})
	if got.Status != "possibly-leaked" {
		t.Fatalf("GitHub tool access disappeared: %+v", got)
	}
}

func TestBenchCollectProfileChecksOnlyAttributedRunWindow(t *testing.T) {
	input := benchCaptureInput(t, "")
	body := fmt.Sprintf("{\"type\":\"session_meta\",\"timestamp\":\"2026-10-23T09:00:00Z\",\"payload\":{\"id\":\"native-session\",\"cwd\":%q}}\n", input.Runtime.Root) +
		`{"type":"turn_context","timestamp":"2026-10-23T09:58:00Z","payload":{"turn_id":"warmup","model":"old-model","effort":"high"}}` + "\n" +
		`{"type":"token_usage_record","timestamp":"2026-10-23T09:59:00Z","payload":{"thread_id":"native-session","turn_id":"warmup","response_id":"warmup-response","usage":{"input_tokens":7,"output_tokens":1,"total_tokens":8}}}` + "\n" +
		`{"type":"turn_context","timestamp":"2026-10-23T10:00:01Z","payload":{"turn_id":"run","model":"gpt-test","effort":"medium"}}` + "\n" +
		`{"type":"token_usage_record","timestamp":"2026-10-23T10:00:02Z","payload":{"thread_id":"native-session","turn_id":"run","response_id":"run-response","usage":{"input_tokens":10,"output_tokens":2,"total_tokens":12}}}` + "\n" +
		`{"type":"turn_context","timestamp":"2026-10-23T10:01:00Z","payload":{"turn_id":"after","model":"post-run-model","effort":"low"}}` + "\n"
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
	if err != nil {
		t.Fatal(err)
	}
	task := benchFixtureTask(t)
	p := bench.AgentProfile{Agent: "codex", Model: "gpt-test", Effort: "medium"}
	task.Worker, task.TL, task.Reviewer = p, p, p
	since := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	got, err := collectBenchMeasurements(context.Background(), benchCollectionInput{Task: task, Run: benchCollectionRun(t, input.Runtime), Sources: captured, Since: since, Until: since.Add(time.Minute), WorkerRepo: input.Runtime.Root})
	if err != nil || !got.Valid || got.Measurements.Worker.InputTokens == nil || *got.Measurements.Worker.InputTokens != 10 || got.Sources[0].Model != "gpt-test" || got.Sources[0].Effort != "medium" {
		t.Fatalf("warmup/post-run identity poisoned run: %+v %v", got, err)
	}
}

func FuzzNormalizeBenchTool(f *testing.F) {
	for _, raw := range []string{`{}`, `{"cmd":"git fetch origin"}`, `{"file_path":"/host/source/fix.go"}`, `{"cmd":null}`, `{"cmd":"a","cmd":"b"}`, `"opaque"`} {
		f.Add(raw)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		if len(raw) > benchCaptureLineLimit {
			t.Skip()
		}
		for _, name := range []string{"Read", "exec_command", "unknown_tool"} {
			got := normalizeBenchTool(name, "fixture", "/bench/work", json.RawMessage(raw))
			if got.Kind == "read" && len(got.Paths) == 0 {
				t.Fatalf("read without operands: %+v", got)
			}
			if !json.Valid([]byte(raw)) && got.Unsupported == "" {
				t.Fatalf("malformed tool declared supported: %+v", got)
			}
		}
	})
}

func benchCaptureInput(t *testing.T, body string) benchSourceInput {
	t.Helper()
	root := t.TempDir()
	file := filepath.Join(root, "source.jsonl")
	if err := os.WriteFile(file, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	return benchSourceInput{Role: "worker", Runtime: bench.RuntimeIdentity{Slot: "repo:1", Session: "runtime-session", Generation: "generation", Root: root, Agent: "codex", Model: "gpt-test", Effort: "medium"}, NativeSession: "native-session", File: file}
}

func TestBenchCaptureExactPrefixIdentityAndBounds(t *testing.T) {
	body := "{\"type\":\"session_meta\",\"payload\":{\"id\":\"native-session\"}}\n"
	input := benchCaptureInput(t, body)
	dir := filepath.Join(t.TempDir(), "captures")
	captured, err := captureBenchSources(context.Background(), dir, input.Runtime.Root, []benchSourceInput{input})
	if err != nil || len(captured) != 1 {
		t.Fatalf("capture %+v %v", captured, err)
	}
	got := captured[0]
	if got.Length != int64(len(body)) || len(got.Digest) != 64 || !got.Complete || got.Device == 0 || got.Inode == 0 {
		t.Fatalf("pin %+v", got)
	}
	if err := os.WriteFile(input.File, []byte("changed live source\n"), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(got.File)
	if err != nil || string(raw) != body {
		t.Fatalf("immutable captured prefix %q %v", raw, err)
	}
	if _, err := captureBenchSources(context.Background(), dir, input.Runtime.Root, []benchSourceInput{input}); err == nil {
		t.Fatal("existing capture directory adopted")
	}
	oversized := benchCaptureInput(t, strings.Repeat("x", benchCaptureLineLimit+1)+"\n")
	if _, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), oversized.Runtime.Root, []benchSourceInput{oversized}); err == nil {
		t.Fatal("oversized line accepted")
	}
	link := filepath.Join(t.TempDir(), "source.jsonl")
	if err := os.Symlink(input.File, link); err != nil {
		t.Fatal(err)
	}
	input.File = link
	if _, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input}); err == nil {
		t.Fatal("source symlink followed")
	}
}

func TestBenchNormalizeSyntheticBoundClaudeCodex(t *testing.T) {
	// SYNTHETIC protocol fixtures, not recordings of real agent sessions.
	for _, agent := range []string{"codex", "claude"} {
		input := benchCaptureInput(t, "")
		input.Runtime.Agent = agent
		var body string
		if agent == "codex" {
			body = `{"type":"session_meta","payload":{"id":"native-session","cwd":"` + input.Runtime.Root + `"}}` + "\n" +
				`{"type":"turn_context","payload":{"model":"gpt-test","effort":"medium"}}` + "\n" +
				`{"type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"tool-1","arguments":"{\"cmd\":\"cat /host/source/solution.go\"}"}}` + "\n"
		} else {
			body = `{"type":"user","sessionId":"native-session","cwd":"` + input.Runtime.Root + `","message":{"content":"work"}}` + "\n" +
				`{"type":"assistant","sessionId":"native-session","effort":"medium","message":{"model":"gpt-test","content":[{"type":"tool_use","id":"tool-1","name":"Read","input":{"file_path":"/host/source/solution.go"}}]}}` + "\n"
		}
		if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		sources, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
		if err != nil {
			t.Fatal(err)
		}
		got, err := normalizeBenchCapturedSource(context.Background(), sources[0])
		if err != nil || !got.Valid || got.Model != "gpt-test" || got.Effort != "medium" || len(got.Tools) != 1 {
			t.Fatalf("%s: %+v %v", agent, got, err)
		}
		scope := bench.LeakScope{PreparedRoots: []string{input.Runtime.Root}, ProtectedRoots: []string{"/host/source"}, Complete: got.Complete}
		if leaks := bench.ScanLeakEvents(scope, got.Tools); leaks.Status != "possibly-leaked" {
			t.Fatalf("%s lost tool evidence: %+v", agent, leaks)
		}
		// First executing identity cannot be replaced by a later copied header.
		wrong := strings.Replace(body, "native-session", "foreign-session", 1)
		if err := os.WriteFile(input.File, []byte(wrong), 0600); err != nil {
			t.Fatal(err)
		}
		foreign, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
		if err != nil {
			t.Fatal(err)
		}
		got, err = normalizeBenchCapturedSource(context.Background(), foreign[0])
		if err != nil || got.Valid {
			t.Fatalf("foreign source admitted: %+v %v", got, err)
		}
	}
}

func TestBenchCaptureAndNormalizeLinearVisits(t *testing.T) {
	for _, n := range []int{100, 200} {
		input := benchCaptureInput(t, "")
		header := `{"type":"session_meta","payload":{"id":"native-session","cwd":"` + input.Runtime.Root + `"}}` + "\n"
		body := header + strings.Repeat(`{"type":"response_item","payload":{"type":"function_call","name":"unknown_tool","arguments":"{}"}}`+"\n", n)
		if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		sources, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
		if err != nil {
			t.Fatal(err)
		}
		got, err := normalizeBenchCapturedSource(context.Background(), sources[0])
		if err != nil {
			t.Fatal(err)
		}
		if got.Records != n+1 || len(got.Tools) != n || sources[0].Length != int64(len(body)) {
			t.Fatalf("n=%d source=%+v normalized=%+v", n, sources[0], got)
		}
		if got.Complete {
			t.Fatal("unsupported calls reported complete coverage")
		}
	}
}

func TestBenchNormalizeUnknownNativeToolRemainsUnknown(t *testing.T) {
	input := benchCaptureInput(t, "")
	body := `{"type":"session_meta","payload":{"id":"native-session","cwd":"` + input.Runtime.Root + `"}}` + "\n" +
		`{"type":"response_item","payload":{"type":"web_search_call","action":{"query":"hidden solution"}}}` + "\n"
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	sources, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
	if err != nil {
		t.Fatal(err)
	}
	got, err := normalizeBenchCapturedSource(context.Background(), sources[0])
	if err != nil || got.Complete || len(got.Tools) != 1 || got.Tools[0].Kind != "opaque" {
		t.Fatalf("unsupported native tool disappeared: %+v %v", got, err)
	}
}

func TestBenchClaudeInitialSessionPrecedesWorkspaceRecord(t *testing.T) {
	input := benchCaptureInput(t, "")
	input.Runtime.Agent = "claude"
	body := `{"type":"queue-operation","sessionId":"native-session","operation":"enqueue"}` + "\n" +
		fmt.Sprintf("{\"type\":\"user\",\"sessionId\":\"native-session\",\"cwd\":%q,\"message\":{\"content\":\"work\"}}\n", input.Runtime.Root)
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
	if err != nil {
		t.Fatal(err)
	}
	got, err := normalizeBenchCapturedSource(context.Background(), captured[0])
	if err != nil || !got.Valid {
		t.Fatalf("native metadata prelude rejected: %+v %v", got, err)
	}
}

func TestBenchCaptureRefusesFIFOWithoutBlocking(t *testing.T) {
	input := benchCaptureInput(t, "")
	fifo := filepath.Join(t.TempDir(), "fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Fatal(err)
	}
	input.File = fifo
	done := make(chan error, 1)
	captureDir := filepath.Join(t.TempDir(), "captures")
	go func() {
		_, err := captureBenchSources(context.Background(), captureDir, input.Runtime.Root, []benchSourceInput{input})
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("FIFO source accepted")
		}
	case <-time.After(250 * time.Millisecond):
		// Release a broken blocking implementation so the test leaves no waiter.
		writer, err := os.OpenFile(fifo, os.O_WRONLY|syscall.O_NONBLOCK, 0)
		if err == nil {
			writer.Close()
		}
		<-done
		t.Fatal("nonregular source open blocked before type validation")
	}
}

func TestBenchCollectFixedOwnerPassVisits(t *testing.T) {
	for _, n := range []int{100, 200} {
		input := benchCaptureInput(t, "")
		var body strings.Builder
		fmt.Fprintf(&body, "{\"type\":\"session_meta\",\"timestamp\":\"2026-10-23T10:00:00Z\",\"payload\":{\"id\":\"native-session\",\"cwd\":%q}}\n", input.Runtime.Root)
		body.WriteString(`{"type":"turn_context","timestamp":"2026-10-23T10:00:01Z","payload":{"model":"gpt-test","effort":"medium","turn_id":"turn-1"}}` + "\n")
		for i := 0; i < n; i++ {
			fmt.Fprintf(&body, "{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-23T10:00:02Z\",\"payload\":{\"thread_id\":\"native-session\",\"turn_id\":\"turn-1\",\"response_id\":\"response-%d\",\"usage\":{\"input_tokens\":1,\"output_tokens\":1,\"total_tokens\":2}}}\n", i)
			arguments := fmt.Sprintf(`{"path":%q}`, filepath.Join(input.Runtime.Root, "a.go"))
			fmt.Fprintf(&body, "{\"type\":\"response_item\",\"payload\":{\"type\":\"function_call\",\"name\":\"read_file\",\"call_id\":\"tool-%d\",\"arguments\":%q}}\n", i, arguments)
		}
		if err := os.WriteFile(input.File, []byte(body.String()), 0600); err != nil {
			t.Fatal(err)
		}
		captured, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
		if err != nil {
			t.Fatal(err)
		}
		task := benchFixtureTask(t)
		p := bench.AgentProfile{Agent: "codex", Model: "gpt-test", Effort: "medium"}
		task.Worker, task.TL, task.Reviewer = p, p, p
		since := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
		got, err := collectBenchMeasurements(context.Background(), benchCollectionInput{Task: task, Run: benchCollectionRun(t, input.Runtime), Sources: captured, Since: since, Until: since.Add(time.Minute), WorkerRepo: input.Runtime.Root})
		if err != nil {
			t.Fatal(err)
		}
		if got.Visits.CapturedBytes != int64(body.Len()) || got.Visits.NormalizedRecords != 2*n+2 || got.Visits.UsageRecords != int64(2*n+2) || got.Visits.UsageResponses != n || got.Visits.AggregatedResponses != n || got.Visits.LeakEvents != n || got.Measurements.Leaks.Coverage.Operands != n {
			t.Fatalf("n=%d fixed passes: %+v leaks=%+v", n, got.Visits, got.Measurements.Leaks)
		}
	}
}

func TestBenchCollectFocusedOwnerReplaysOnlyCapturedBytes(t *testing.T) {
	// SYNTHETIC retained ownership inputs, exercised through the real engine.
	ownedFixtureAt(t, 0)
	root := testfix.Repo(t, testfix.InitialCommit())
	ownedFixtureAt(t, 30)
	testfix.Git(t, root, "commit", "--allow-empty", "-qm", "#1: measured fixture")
	head := strings.TrimSpace(testfix.Capture(t, root, "rev-parse", "HEAD"))
	since := time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)
	until := since.Add(30 * time.Minute)
	input := benchCaptureInput(t, "")
	input.Runtime.Agent = "claude"
	input.Runtime.Root = root
	body := fmt.Sprintf("{\"timestamp\":\"2026-10-14T00:00:00Z\",\"type\":\"system\",\"sessionId\":\"native-session\",\"cwd\":%q}\n", root) +
		`{"timestamp":"2026-10-14T00:01:00Z","type":"assistant","sessionId":"native-session","effort":"medium","message":{"model":"gpt-test","content":[]}}` + "\n" +
		`{"timestamp":"2026-10-14T00:05:00Z","type":"user","sessionId":"native-session","message":{"content":"#1"}}` + "\n" +
		`{"timestamp":"2026-10-14T00:15:00Z","type":"user","sessionId":"native-session","message":{"content":"#1"}}` + "\n"
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	captured, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), input.Runtime.Root, []benchSourceInput{input})
	if err != nil {
		t.Fatal(err)
	}
	source := captured[0]
	measurementSource := gatestate.MeasurementSource{Harness: "claude", ID: input.NativeSession, CWD: root, Digest: source.Digest, Bytes: source.Length, Started: since.Format(time.RFC3339Nano), Windows: []gatestate.MeasurementInterval{{Start: since.Format(time.RFC3339Nano), End: until.Format(time.RFC3339Nano)}}}
	snapshot := gatestate.MeasurementInputs{Version: 1, Method: gatestate.MeasurementMethodEngagementV1, Repository: "repo", CodeHead: head, TrackerHead: head, Since: since.Format(time.RFC3339Nano), Until: until.Format(time.RFC3339Nano), GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true, Engagements: []gatestate.MeasurementEngagement{{Start: since.Format(time.RFC3339Nano), End: until.Format(time.RFC3339Nano), Began: since.Format(time.RFC3339Nano), Roots: []string{root}}}, Sources: []gatestate.MeasurementSource{measurementSource}}
	focused := &ownedActualInputs{Snapshot: snapshot, Files: []string{input.File}, Prefixes: map[string]activetime.SourcePrefix{input.File: {Bytes: source.Length, Digest: source.Digest}}, Windows: map[string][]activetime.ActivityWindow{input.File: {{Start: since, End: until, Began: since}}}, SourceSlots: map[string]string{input.File: input.Runtime.Slot}, Scope: activetime.Scope{Head: head, Issue: "1"}, Sessions: []transcripts.SessionEvidence{{Harness: "claude", ID: input.NativeSession, CWD: root, File: input.File, Digest: source.Digest, Length: source.Length, Started: since}}, Replay: true}
	baseline := computeOwnedActualFor(context.Background(), root, "1", "", focused)
	if baseline.Status != actualMeasured || baseline.Hours <= 0 {
		t.Fatalf("native owner fixture: %+v", baseline)
	}
	if err := os.WriteFile(input.File, []byte("live transcript now unavailable\n"), 0600); err != nil {
		t.Fatal(err)
	}
	task := benchFixtureTask(t)
	p := bench.AgentProfile{Agent: "claude", Model: "gpt-test", Effort: "medium"}
	task.Worker, task.TL, task.Reviewer = p, p, p
	args := benchCollectionInput{Task: task, Run: benchCollectionRun(t, input.Runtime), Sources: captured, Since: since, Until: until, WorkerRepo: root, Focused: focused}
	got, err := collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Measurements.ActiveHours == nil || *got.Measurements.ActiveHours != baseline.Hours || got.Visits.ActivityPrefixes != 1 {
		t.Fatalf("captured owner replay: %+v %v", got, err)
	}
	args.Focused = &ownedActualInputs{}
	got, err = collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Measurements.ActiveHours != nil {
		t.Fatalf("invalid retained authority fabricated hours: %+v %v", got, err)
	}
}

func TestBenchReviewerCaptureRequiresExactLocalAccountingBeforeOpen(t *testing.T) {
	b, head := benchLocalCheckFixture(t)
	input := benchCaptureInput(t, "")
	input.Role = "reviewer"
	input.ReviewAttemptID = "review-one"
	input.NativeSession = "11111111-1111-4111-8111-111111111111"
	// A nonexistent source proves no file-open fallback happens before binding.
	input.File = filepath.Join(t.TempDir(), "missing.jsonl")
	excluded, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), b.WorkRoot, []benchSourceInput{input})
	if err != nil || len(excluded) != 1 || excluded[0].File != "" || excluded[0].Length != 0 || excluded[0].Complete || excluded[0].Issue == "" {
		t.Fatalf("unbound reviewer capture: %+v %v", excluded, err)
	}
	record := reviewusage.Record{Version: reviewusage.Version, Intent: reviewusage.Intent{ID: input.ReviewAttemptID, Repository: b.OriginIdentity, IssueID: "000001", Recipe: "boundary-review", Round: 1, Attempt: 1, Agent: input.Runtime.Agent, Root: input.Runtime.Root, ReviewedHEAD: head, Slot: input.Runtime.Slot, RequestedAt: "2026-10-23T10:00:00Z"}, Phase: reviewusage.Finished, PID: 42, StartedAt: "2026-10-23T10:00:01Z", FinishedAt: "2026-10-23T10:00:03Z", SessionID: input.NativeSession, Outcome: "returned"}
	store, err := reviewusage.NewStore(b.WorkRoot)
	if err != nil {
		t.Fatal(err)
	}
	requested, err := reviewusage.NewRequested(record.Intent)
	if err != nil {
		t.Fatal(err)
	}
	oid, err := store.Save(context.Background(), requested, "")
	if err != nil {
		t.Fatal(err)
	}
	started, err := reviewusage.Step(requested, reviewusage.Event{Kind: reviewusage.Start, Intent: record.Intent, PID: record.PID, At: record.StartedAt})
	if err != nil {
		t.Fatal(err)
	}
	oid, err = store.Save(context.Background(), started, oid)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Save(context.Background(), record, oid); err != nil {
		t.Fatal(err)
	}
	body := fmt.Sprintf("{\"type\":\"session_meta\",\"payload\":{\"id\":%q,\"cwd\":%q}}\n", input.NativeSession, input.Runtime.Root) +
		`{"type":"turn_context","timestamp":"2026-10-23T10:00:01Z","payload":{"turn_id":"review-turn","model":"gpt-test","effort":"medium"}}` + "\n" +
		fmt.Sprintf("{\"type\":\"token_usage_record\",\"timestamp\":\"2026-10-23T10:00:02Z\",\"payload\":{\"thread_id\":%q,\"turn_id\":\"review-turn\",\"response_id\":\"review-response\",\"usage\":{\"input_tokens\":3,\"output_tokens\":1,\"total_tokens\":4}}}\n", input.NativeSession)
	if err := os.WriteFile(input.File, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	sources, err := captureBenchSources(context.Background(), filepath.Join(t.TempDir(), "captures"), b.WorkRoot, []benchSourceInput{input})
	if err != nil || len(sources) != 1 || !sources[0].Complete || sources[0].Length == 0 {
		t.Fatalf("bound reviewer not captured: %+v %v", sources, err)
	}
	if _, err := validateBenchReviewer(context.Background(), b.WorkRoot, input, "000001", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := validateBenchReviewer(context.Background(), b.WorkRoot, input, "000002", nil); err == nil {
		t.Fatal("foreign issue reviewer bound")
	}
	since := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	late := usage.Window{Since: since.Add(time.Hour), Until: since.Add(2 * time.Hour)}
	if _, err := validateBenchReviewer(context.Background(), b.WorkRoot, input, "000001", &late); err == nil {
		t.Fatal("out-of-window reviewer bound")
	}
	worker := input.Runtime
	worker.Root = b.WorkRoot
	worker.Session = "worker-runtime"
	task := benchFixtureTask(t)
	profile := bench.AgentProfile{Agent: "codex", Model: "gpt-test", Effort: "medium"}
	task.Worker, task.TL, task.Reviewer = profile, profile, profile
	args := benchCollectionInput{Task: task, Run: benchCollectionRun(t, worker, strings.Repeat("f", 40)), Sources: sources, Since: since, Until: since.Add(time.Minute), WorkerRepo: b.WorkRoot}
	got, err := collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Measurements.Reviewer.InputTokens == nil || *got.Measurements.Reviewer.InputTokens != 3 {
		t.Fatalf("earlier review head lost from owned run cost: %+v %v", got, err)
	}
	args.Task.ID = "000002"
	got, err = collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Measurements.Reviewer.InputTokens != nil {
		t.Fatalf("foreign issue reviewer cost included: %+v %v", got, err)
	}
	args.Task.ID = "000001"
	args.Since, args.Until = late.Since, late.Until
	got, err = collectBenchMeasurements(context.Background(), args)
	if err != nil || got.Measurements.Reviewer.InputTokens != nil {
		t.Fatalf("out-of-window reviewer cost included: %+v %v", got, err)
	}

	input.NativeSession = "22222222-2222-4222-8222-222222222222"
	if _, err := validateBenchReviewer(context.Background(), b.WorkRoot, input, "000001", nil); err == nil {
		t.Fatal("foreign reviewer session bound")
	}
}
