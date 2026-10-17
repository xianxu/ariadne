package activetime

import (
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
	got, attribution := subtractQueueIdle(runs, waits)
	if len(got) != 3 {
		t.Fatalf("want 2 retained a pieces and unchanged b, got %+v", got)
	}
	var a, b float64
	for _, run := range got {
		if run.Source == "a" {
			a += run.Active
		} else {
			b += run.Active
		}
	}
	if a != 20 || b != 60 || attribution.RawMinutes != 40 || attribution.ExcludedMinutes != 40 {
		t.Fatalf("bad source-scoped conservation: a=%v b=%v attribution=%+v", a, b, attribution)
	}
}

func TestSubtractQueueIdleReportsFullRecordedWait(t *testing.T) {
	runs := []ActivityRun{{Start: qt(20), End: qt(30), Active: 10, Source: "a"}}
	waits := []QueueInterval{{Start: qt(10), End: qt(50), Source: "a", AttemptID: "one"}}
	_, stats := subtractQueueIdle(runs, waits)
	if stats.RawMinutes != 40 || stats.ExcludedMinutes != 10 {
		t.Fatalf("raw wait is record duration while excluded idle intersects activity: %+v", stats)
	}
}

func TestQueueRawMinutesIncludesRecordedWaitWithoutActivity(t *testing.T) {
	waits := []QueueInterval{{Start: qt(10), End: qt(20), AttemptID: "one"}}
	if got := queueRawMinutes(waits); got != 10 {
		t.Fatalf("recorded queue wait lost when no activity source overlaps: %v", got)
	}
}

func TestSubtractQueueIdleOneAttemptAcrossTwoOwnedSources(t *testing.T) {
	runs := []ActivityRun{{Start: qt(0), End: qt(30), Source: "a"}, {Start: qt(0), End: qt(30), Source: "b"}}
	wait := QueueInterval{Start: qt(10), End: qt(20), Slot: "ariadne:1", AttemptID: "one"}
	bound, err := bindQueueSources(wait, runs, map[string]string{"a": "ariadne:1", "b": "ariadne:1"})
	if err != nil {
		t.Fatal(err)
	}
	got, stats := subtractQueueIdle(runs, bound)
	var a, b float64
	for _, run := range got {
		if run.Source == "a" {
			a += run.Active
		} else {
			b += run.Active
		}
	}
	if a != 20 || b != 20 || stats.RawMinutes != 10 || stats.ExcludedMinutes != 20 {
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
	_, err = Compute(base)
	if old.TotalActive != 120 || err == nil || !strings.Contains(err.Error(), "overlapping work, unmeasured") {
		t.Fatalf("a stage call inside the wait must exclude the sample; old=%v err=%v", old.TotalActive, err)
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
	if err == nil || !strings.Contains(err.Error(), "overlapping work, unmeasured") {
		t.Fatalf("yielded work must make queue attribution incomplete: %v", err)
	}
}

func TestQueueNativeTranscriptPollsDoNotRestoreWait(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "#8 queue work")
	file := filepath.Join(t.TempDir(), "session.jsonl")
	writeJSONL(t, file, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"suite","input":"text(await tools.exec_command({cmd:\"sdlc local-check run --name suite\"}))"}}`,
		`{"timestamp":"2026-01-01T00:20:00Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"poll1","input":"const r=await tools.write_stdin({session_id:7,chars:\"\"});text(r)"}}`,
		`{"timestamp":"2026-01-01T00:21:00Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"poll1","output":"unknown process"}}`,
		`{"timestamp":"2026-01-01T00:40:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"poll2","arguments":"{\"session_id\":7}"}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 done"}}`,
	})
	got, err := Compute(Options{Files: []string{file}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z", Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 120, IncludeAssistant: true,
		QueueIntervals: []QueueInterval{{Start: qt(10), End: qt(50), Slot: "ariadne:1", AttemptID: "one"}}, SourceSlots: map[string]string{file: "ariadne:1"}})
	if err != nil || got.TotalActive != 20 || got.QueueAttribution.ExcludedMinutes != 40 {
		t.Fatalf("poll-only queue must subtract all wait activity: result=%+v err=%v", got, err)
	}
}

func TestQueueToolCallsTreatInteractiveInputAsWork(t *testing.T) {
	lines := []string{
		`{"timestamp":"2026-01-01T00:10:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"empty","arguments":"{\"session_id\":7,\"chars\":\"\"}"}}`,
		`{"timestamp":"2026-01-01T00:11:00Z","type":"response_item","payload":{"type":"function_call","name":"write_stdin","call_id":"input","arguments":"{\"session_id\":7,\"chars\":\"yes\\n\"}"}}`,
		`{"timestamp":"2026-01-01T00:12:00Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"wrapped","input":"text(await tools.write_stdin({session_id:7,chars:\"yes\"}))"}}`,
	}
	calls := queueToolCallsFromTranscript([]byte(strings.Join(lines, "\n")), "a")
	if len(calls) != 3 || !calls[0].Poll || calls[1].Poll || calls[2].Poll {
		t.Fatalf("only empty-input polls may be ignored: %+v", calls)
	}
}

func TestQueueClaudeCallRefusesWithoutInspectingResult(t *testing.T) {
	repo := gitInit(t)
	gitCommit(t, repo, "2026-01-01T01:00:00Z", "#8 queue work")
	file := filepath.Join(t.TempDir(), "claude.jsonl")
	writeJSONL(t, file, []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"#8 start"}}`,
		`{"timestamp":"2026-01-01T00:25:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"edit","name":"Edit","input":{"file_path":"x"}}]}}`,
		`{"timestamp":"2026-01-01T00:26:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"edit","is_error":true,"content":"done"}]}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":"#8 done"}}`,
	})
	_, err := Compute(Options{Files: []string{file}, GitRepo: repo, SinceISO: "2026-01-01T00:00:00Z", UntilISO: "2026-01-01T01:00:00Z", Issues: []string{"8"}, CommitWeight: 1, ThresholdMin: 120, IncludeAssistant: true,
		QueueIntervals: []QueueInterval{{Start: qt(10), End: qt(50), Slot: "ariadne:1", AttemptID: "one"}}, SourceSlots: map[string]string{file: "ariadne:1"}})
	if err == nil || !strings.Contains(err.Error(), "overlapping work, unmeasured") {
		t.Fatalf("Claude call inside wait must exclude the sample regardless of result: %v", err)
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
