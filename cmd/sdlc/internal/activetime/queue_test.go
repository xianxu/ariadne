package activetime

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func qt(min int) time.Time { return time.Date(2026, 1, 1, 0, min, 0, 0, time.UTC) }

func TestSubtractQueueIdleIntervals(t *testing.T) {
	runs := []ActivityRun{
		{Start: qt(0), End: qt(60), Active: 60, Source: "a", Mentions: map[string]int{"8": 1}},
		{Start: qt(0), End: qt(60), Active: 60, Source: "b"},
	}
	waits := []QueueInterval{
		{Start: qt(10), End: qt(50), Source: "a", AttemptID: "one"},
		{Start: qt(20), End: qt(40), Source: "a", AttemptID: "two"},
	}
	work := []ToolSpan{{Start: qt(25), End: qt(30), Source: "a", Kind: toolWork}}
	got, attribution := subtractQueueIdle(runs, waits, work)
	if len(got) != 4 {
		t.Fatalf("want 3 retained a pieces and unchanged b, got %+v", got)
	}
	var a, b float64
	for _, run := range got {
		if run.Source == "a" {
			a += run.Active
		} else {
			b += run.Active
		}
	}
	if a != 25 || b != 60 || attribution.RawMinutes != 40 || attribution.ExcludedMinutes != 35 || attribution.RetainedWorkMinutes != 5 {
		t.Fatalf("bad source-scoped conservation: a=%v b=%v attribution=%+v", a, b, attribution)
	}
}

func TestWorkSpansFromTranscriptNativeExecWrapper(t *testing.T) {
	record := func(min int, payload map[string]any) string {
		raw, err := json.Marshal(map[string]any{"timestamp": qt(min).Format(time.RFC3339), "type": "response_item", "payload": payload})
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	output := func(id, body string) map[string]any {
		return map[string]any{"type": "custom_tool_call_output", "call_id": id, "output": []map[string]string{{"type": "input_text", "text": "Script completed\nOutput:\n"}, {"type": "input_text", "text": body}}}
	}
	for name, pollInput := range map[string]string{
		"direct":   `text(await tools.write_stdin({session_id:7,chars:""}))`,
		"assigned": `const r=await tools.write_stdin({session_id:7,chars:"",yield_time_ms:30000});text(r)`,
	} {
		t.Run(name, func(t *testing.T) {
			lines := []string{
				record(0, map[string]any{"type": "custom_tool_call", "name": "exec", "call_id": "suite", "input": `text(await tools.exec_command({cmd:"sdlc local-check run --name suite",yield_time_ms:30000}))`}),
				record(1, output("suite", `{"session_id":7,"output":"queued"}`)),
				record(10, map[string]any{"type": "custom_tool_call", "name": "exec", "call_id": "poll", "input": pollInput}),
				record(11, output("poll", `{"session_id":7,"output":"waiting"}`)),
				record(20, map[string]any{"type": "custom_tool_call", "name": "exec", "call_id": "done", "input": pollInput}),
				record(21, output("done", `{"exit_code":0,"output":"passed"}`)),
			}
			spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
			if len(spans) != 0 {
				t.Fatalf("yielded suite must not depend on poll pairing: %+v", spans)
			}
		})
	}
}

func TestSubtractQueueIdleReportsFullRecordedWait(t *testing.T) {
	runs := []ActivityRun{{Start: qt(20), End: qt(30), Active: 10, Source: "a"}}
	waits := []QueueInterval{{Start: qt(10), End: qt(50), Source: "a", AttemptID: "one"}}
	_, stats := subtractQueueIdle(runs, waits, nil)
	if stats.RawMinutes != 40 || stats.ExcludedMinutes != 10 {
		t.Fatalf("raw wait is record duration while excluded idle intersects activity: %+v", stats)
	}
}

func TestSubtractQueueIdleOneAttemptAcrossTwoOwnedSources(t *testing.T) {
	runs := []ActivityRun{{Start: qt(0), End: qt(30), Source: "a"}, {Start: qt(0), End: qt(30), Source: "b"}}
	wait := QueueInterval{Start: qt(10), End: qt(20), Slot: "ariadne:1", AttemptID: "one"}
	bound, err := bindQueueSources(wait, runs, map[string]string{"a": "ariadne:1", "b": "ariadne:1"})
	if err != nil {
		t.Fatal(err)
	}
	got, stats := subtractQueueIdle(runs, bound, []ToolSpan{{Start: qt(12), End: qt(15), Source: "b", Kind: toolWork}})
	var a, b float64
	for _, run := range got {
		if run.Source == "a" {
			a += run.Active
		} else {
			b += run.Active
		}
	}
	if a != 20 || b != 23 || stats.RawMinutes != 10 || stats.ExcludedMinutes != 17 || stats.RetainedWorkMinutes != 3 {
		t.Fatalf("one recorded wait across owned sources: a=%v b=%v stats=%+v", a, b, stats)
	}
}

func TestQueueNativeTranscript(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "#8 queue work")
	dir := t.TempDir()
	first := filepath.Join(dir, "first.jsonl")
	second := filepath.Join(dir, "second.jsonl")
	writeJSONL(t, first, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"suite","arguments":"{\"cmd\":\"sdlc local-check run --name suite\"}"}}`,
		`{"timestamp":"2026-01-01T00:25:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"done"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":"Process running with session ID 7"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 done"}}`,
	})
	writeJSONL(t, second, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 other source"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 other source done"}}`,
	})
	base := Options{Files: []string{first, second}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z", Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 120, IncludeAssistant: true}
	old, err := Compute(base)
	if err != nil {
		t.Fatal(err)
	}
	base.QueueIntervals = []QueueInterval{{Start: qt(10), End: qt(50), Slot: "ariadne:1", AttemptID: "one"}}
	base.SourceSlots = map[string]string{first: "ariadne:1", second: "ariadne:2"}
	got, err := Compute(base)
	if err != nil {
		t.Fatal(err)
	}
	if old.TotalActive != 120 || got.TotalActive != 85 || got.QueueAttribution.ExcludedMinutes != 35 || got.QueueAttribution.RetainedWorkMinutes != 5 {
		t.Fatalf("queue attribution old=%v new=%v diagnostics=%+v", old.TotalActive, got.TotalActive, got.QueueAttribution)
	}
}

func TestQueueNativeTranscriptRefusesYieldedWorkOverlap(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "#8 queue work")
	file := filepath.Join(t.TempDir(), "session.jsonl")
	writeJSONL(t, file, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T00:25:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:26:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"Process running with session ID 8"}}`,
		`{"timestamp":"2026-01-01T00:27:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"stage","arguments":"{\"session_id\":8}"}}`,
		`{"timestamp":"2026-01-01T00:28:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"unknown process"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 done"}}`,
	})
	_, err := Compute(Options{Files: []string{file}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z", Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 120, IncludeAssistant: true,
		QueueIntervals: []QueueInterval{{Start: qt(10), End: qt(50), Slot: "ariadne:1", AttemptID: "one"}}, SourceSlots: map[string]string{file: "ariadne:1"}})
	if err == nil || !strings.Contains(err.Error(), "incomplete work") {
		t.Fatalf("yielded work must make queue attribution incomplete: %v", err)
	}
}

func TestBindQueueSources(t *testing.T) {
	wait := QueueInterval{Start: qt(10), End: qt(20), Slot: "ariadne:1", AttemptID: "one"}
	runs := []ActivityRun{{Start: qt(0), End: qt(30), Source: "a"}, {Start: qt(5), End: qt(25), Source: "b"}, {Start: qt(0), End: qt(30), Source: "foreign"}}
	got, err := bindQueueSources(wait, runs, map[string]string{"a": "ariadne:1", "b": "ariadne:1", "foreign": "ariadne:2"})
	if err != nil || len(got) != 2 || got[0].Source != "a" || got[1].Source != "b" {
		t.Fatalf("want both owned sources only: %+v, %v", got, err)
	}
	if _, err := bindQueueSources(wait, runs, map[string]string{"a": "ariadne:1", "foreign": "ariadne:2"}); err == nil {
		t.Fatal("unmapped overlapping source admitted")
	}
}

func TestWorkSpansFromTranscriptSequence(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"suite","arguments":"{\"cmd\":\"cd repo && sdlc local-check run --name suite\"}"}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":"Process running with session ID 7"}}`,
		`{"timestamp":"2026-01-01T00:03:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:06:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"done"}}`,
		`{"timestamp":"2026-01-01T00:07:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"poll","arguments":"{\"session_id\":7,\"chars\":\"\"}"}}`,
		`{"timestamp":"2026-01-01T00:08:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"poll","output":"still running"}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"end","arguments":"{\"session_id\":7,\"chars\":\"\"}"}}`,
		`{"timestamp":"2026-01-01T00:31:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"end","output":"Process exited with code 0"}}`,
	}
	spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
	var waits, works int
	for _, sp := range spans {
		switch sp.Kind {
		case toolWait:
			waits++
			if !sp.Start.Equal(qt(0)) || !sp.End.Equal(qt(31)) {
				t.Errorf("suite wait not followed through polls: %+v", sp)
			}
		case toolWork:
			works++
			if !sp.Start.Equal(qt(3)) || !sp.End.Equal(qt(6)) {
				t.Errorf("unexpected work span %+v", sp)
			}
		}
	}
	if waits != 0 || works != 1 {
		t.Fatalf("want only direct work, got %+v", spans)
	}
}

func TestWorkSpansFromTranscriptJSONSessionID(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"suite","arguments":"{\"cmd\":\"sdlc local-check run --name suite\"}"}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":{"session_id":7,"output":""}}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"poll","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T00:11:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"poll","output":{"session_id":7,"output":""}}}`,
		`{"timestamp":"2026-01-01T00:20:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"done","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T00:21:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"done","output":{"exit_code":0,"output":"ok"}}}`,
	}
	spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
	if len(spans) != 0 {
		t.Fatalf("yielded suite/polls should not supply queue evidence: %+v", spans)
	}
}

func TestWorkSpansFromTranscriptYieldedWorkNeedsCompletion(t *testing.T) {
	unknownSession := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"still running"}}`,
	}
	if spans := workSpansFromTranscript([]byte(strings.Join(unknownSession, "\n")), "a"); len(spans) != 1 || spans[0].Kind != toolUncertainWork {
		t.Fatalf("untrackable running command not marked uncertain: %+v", spans)
	}
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"Process running with session ID 7"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"poll","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T00:11:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"poll","output":"still running"}}`,
		`{"timestamp":"2026-01-01T00:20:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"done","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T00:21:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"done","output":"Process exited with code 0"}}`,
	}
	if spans := workSpansFromTranscript([]byte(strings.Join(lines[:4], "\n")), "a"); len(spans) != 1 || spans[0].Kind != toolUncertainWork {
		t.Fatalf("unfinished work not marked uncertain: %+v", spans)
	}
	spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
	if len(spans) != 1 || spans[0].Kind != toolUncertainWork {
		t.Fatalf("yielded work cannot be proven by poll: %+v", spans)
	}
}

func TestQueueEvidenceDoesNotDependOnYieldedSuitePolls(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"suite","arguments":"{\"cmd\":\"sdlc local-check run --name suite\"}"}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":"Process running with session ID 7"}}`,
		`{"timestamp":"2026-01-01T00:10:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"suite","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T00:11:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"suite","output":"unknown process"}}`,
	}
	spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
	for _, span := range spans {
		if span.Kind == toolWork {
			t.Fatalf("yielded suite or reused poll became work: %+v", spans)
		}
	}
	runs := []ActivityRun{{Start: qt(0), End: qt(20), Source: "a"}}
	if _, err := bindQueueSources(QueueInterval{Start: qt(2), End: qt(9), Slot: "ariadne:1", AttemptID: "one"}, runs, map[string]string{"a": "ariadne:1"}); err != nil {
		t.Fatalf("recorded queue should not need a terminal suite poll: %v", err)
	}
	unfinished := workSpansFromTranscript([]byte(strings.Join(lines[:2], "\n")), "a")
	if len(unfinished) != 0 {
		t.Fatalf("yielded suite must not be uncertain work: %+v", unfinished)
	}
	if _, err := bindQueueSources(QueueInterval{Start: qt(2), End: qt(9), Slot: "ariadne:1", AttemptID: "one"}, runs, map[string]string{"a": "ariadne:1"}); err != nil {
		t.Fatalf("recorded queue should bind while suite has no poll at all: %v", err)
	}
}

func TestWorkSpansFromTranscriptUnconfirmedPollKeepsYieldedWorkPending(t *testing.T) {
	for name, uncertain := range map[string]string{
		"transport error":              "transport-error",
		"transport error quoting exit": "transport-error: previous poll said Process exited with code 0",
		"empty":                        "",
		"unknown process":              "unknown process",
	} {
		t.Run(name, func(t *testing.T) {
			lines := []string{
				`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
				`{"timestamp":"2026-01-01T00:01:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"stage","output":"Process running with session ID 7"}}`,
				`{"timestamp":"2026-01-01T00:10:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"uncertain","arguments":"{\"session_id\":7}"}}`,
			}
			uncertainRecord, err := json.Marshal(map[string]any{"timestamp": "2026-01-01T00:11:00Z", "type": "response_item", "payload": map[string]any{"type": "function_call_output", "call_id": "uncertain", "output": uncertain}})
			if err != nil {
				t.Fatal(err)
			}
			lines = append(lines, string(uncertainRecord))
			if spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a"); len(spans) != 1 || spans[0].Kind != toolUncertainWork {
				t.Fatalf("unconfirmed poll lost uncertainty: %+v", spans)
			}
			lines = append(lines,
				`{"timestamp":"2026-01-01T00:20:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"done","arguments":"{\"session_id\":7}"}}`,
				`{"timestamp":"2026-01-01T00:21:00Z","type":"response_item","payload":{"type":"function_call_output","call_id":"done","output":"Process exited with code 0"}}`,
			)
			spans := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
			if len(spans) != 1 || spans[0].Kind != toolUncertainWork {
				t.Fatalf("poll completion must not establish work: %+v", spans)
			}
		})
	}
}

func TestWorkSpansFromTranscriptFailedDirectResultIsUncertain(t *testing.T) {
	for _, output := range []string{"", "transport-error", "unknown process"} {
		lines := []string{
			`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"stage","arguments":"{\"cmd\":\"sdlc stage --build\"}"}}`,
		}
		record, err := json.Marshal(map[string]any{"timestamp": "2026-01-01T00:01:00Z", "type": "response_item", "payload": map[string]any{"type": "function_call_output", "call_id": "stage", "output": output}})
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, string(record))
		got := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a")
		if len(got) != 1 || got[0].Kind != toolUncertainWork {
			t.Fatalf("failed direct result %q became work: %+v", output, got)
		}
	}
}

func TestWorkSpansFromTranscriptRejectsDuplicateAndIncomplete(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"dup","name":"Bash","input":{"command":"sdlc stage --build"}}]}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"dup","name":"Bash","input":{"command":"sdlc stage --clear"}}]}}`,
		`{"timestamp":"2026-01-01T00:02:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"dup","content":"done"}]}}`,
		`{"timestamp":"2026-01-01T00:03:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"missing","name":"Bash","input":{"command":"sdlc stage --build"}}]}}`,
	}
	if got := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a"); len(got) != 2 || got[0].Kind != toolUncertainWork || got[1].Kind != toolUncertainWork {
		t.Fatalf("ambiguous/incomplete calls need uncertainty evidence: %+v", got)
	}
}

func TestWorkSpansFromTranscriptRejectsReusedCompletedID(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"again","name":"Bash","input":{"command":"sdlc stage --build"}}]}}`,
		`{"timestamp":"2026-01-01T00:01:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"again","content":"done"}]}}`,
		`{"timestamp":"2026-01-01T00:02:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"again","name":"Bash","input":{"command":"sdlc stage --clear"}}]}}`,
	}
	if got := workSpansFromTranscript([]byte(strings.Join(lines, "\n")), "a"); len(got) != 1 || got[0].Kind != toolUncertainWork {
		t.Fatalf("reused tool ID must invalidate earlier span as uncertain: %+v", got)
	}
}

func TestClassifyToolRequiresExecutableCommand(t *testing.T) {
	for _, input := range []string{
		`{"cmd":"echo 'sdlc stage --build'"}`,
		`{"cmd":"printf 'sdlc merge'"}`,
		`{"cmd":"git log --grep 'sdlc local-check run'"}`,
	} {
		if got := classifyTool("exec_command", input); got != toolUnknown {
			t.Fatalf("quoted/prose command classified as work or wait: %q -> %v", input, got)
		}
	}
	if got := classifyTool("exec_command", `{"cmd":"cd repo && sdlc local-check run --name suite"}`); got != toolWait {
		t.Fatalf("wrapped suite should be wait, got %v", got)
	}
	if got := classifyTool("functions.exec", `{"code":"const r=await tools.exec_command({cmd:\"sdlc local-check run --name suite\"});"}`); got != toolWait {
		t.Fatalf("native JS shell wrapper should be wait, got %v", got)
	}
	if got := classifyTool("functions.exec", `{"code":"await tools.apply_patch(\"*** Begin Patch\");"}`); got != toolWork {
		t.Fatalf("native JS edit wrapper should be work, got %v", got)
	}
}
