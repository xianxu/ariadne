package activetime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEventMentions(t *testing.T) {
	pat := newMentionScope("ariadne", []string{"8", "10"})
	got := parseEventMentions("#8 and #8 then #10 but not #9", pat)
	if got["8"] != 2 || got["10"] != 1 {
		t.Fatalf("want {8:2,10:1}, got %v", got)
	}
	if _, ok := got["9"]; ok {
		t.Fatalf("untracked #9 should be ignored: %v", got)
	}
	// nil pattern → no mentions.
	if m := parseEventMentions("#8", mentionScope{}); len(m) != 0 {
		t.Fatalf("nil pattern should match nothing, got %v", m)
	}
}

// loadEvents must faithfully reproduce active-time-v3.py's per-line handling:
// user string content counts; pure tool_result is dropped; empty user text is
// dropped; an assistant text line counts (with --include-assistant); AND an
// EMPTY-text assistant line is still emitted (its timestamp feeds active time).
func TestLoadEventsShapes(t *testing.T) {
	dir := t.TempDir()
	// One file with a mix of line shapes, plus an out-of-window line.
	lines := []string{
		// user, string content, mentions #8 → counted
		`{"timestamp":"2026-01-01T00:00:00Z","type":"user","message":{"content":"working on #8"}}`,
		// user, pure tool_result (no text parts) → dropped
		`{"timestamp":"2026-01-01T00:01:00Z","type":"user","message":{"content":[{"type":"tool_result","content":"out"}]}}`,
		// user, empty text → dropped
		`{"timestamp":"2026-01-01T00:02:00Z","type":"user","message":{"content":"   "}}`,
		// user, list content with a text block mentioning #10 → counted
		`{"timestamp":"2026-01-01T00:03:00Z","type":"user","message":{"content":[{"type":"text","text":"see #10"}]}}`,
		// assistant, text block mentioning #8 → counted (include-assistant)
		`{"timestamp":"2026-01-01T00:04:00Z","type":"assistant","message":{"content":[{"type":"text","text":"on #8"}]}}`,
		// assistant, NO text (tool_use only) → STILL emitted, empty mentions
		`{"timestamp":"2026-01-01T00:05:00Z","type":"assistant","message":{"content":[{"type":"tool_use","name":"x"}]}}`,
		// malformed JSON → skipped
		`{not json`,
		// out of window → filtered
		`{"timestamp":"2030-01-01T00:00:00Z","type":"user","message":{"content":"future #8"}}`,
	}
	writeJSONL(t, filepath.Join(dir, "session.jsonl"), lines)

	pat := newMentionScope("ariadne", []string{"8", "10"})

	// With include-assistant.
	evs, _, err := loadEvents([]string{dir}, pat, true,
		"2026-01-01T00:00:00Z", "2026-01-01T23:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	// Expected emitted: user#8, user#10, asst#8, asst(empty) = 4 events.
	if len(evs) != 4 {
		t.Fatalf("want 4 events, got %d: %+v", len(evs), evs)
	}
	// Sorted by time.
	for i := 1; i < len(evs); i++ {
		if evs[i].Time.Before(evs[i-1].Time) {
			t.Fatalf("events not sorted: %+v", evs)
		}
	}
	// The empty-text assistant event (last, 00:05) must be present with no mentions.
	last := evs[len(evs)-1]
	if last.Time != tm("2026-01-01T00:05:00Z") {
		t.Fatalf("want empty-assistant event last at 00:05, got %v", last.Time)
	}
	if len(last.Mentions) != 0 {
		t.Fatalf("empty-assistant event should have no mentions, got %v", last.Mentions)
	}
	// Mention totals across the stream.
	total := map[string]int{}
	for _, e := range evs {
		for k, v := range e.Mentions {
			total[k] += v
		}
	}
	if total["8"] != 2 || total["10"] != 1 {
		t.Fatalf("want totals {8:2,10:1}, got %v", total)
	}

	// Without include-assistant: only the two user events remain.
	evs2, _, err := loadEvents([]string{dir}, pat, false,
		"2026-01-01T00:00:00Z", "2026-01-01T23:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs2) != 2 {
		t.Fatalf("without assistant want 2 events, got %d: %+v", len(evs2), evs2)
	}
}

func TestLoadEventsMissingDirSkipped(t *testing.T) {
	evs, _, err := loadEvents([]string{"/no/such/dir"}, newMentionScope("ariadne", []string{"8"}), true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 0 {
		t.Fatalf("missing dir should yield no events, got %+v", evs)
	}
}

// A synchronous Agent dispatch (assistant tool_use) followed by its tool_result
// return produces exactly one TaskSpan; a non-Agent tool call (Bash) produces
// none. The return is a pure tool_result (dropped as an Event) but still closes
// the span.
func TestLoadEventsTaskSpans(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		`{"timestamp":"2026-01-01T00:30:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"A1","content":"done"}]}}`,
		`{"timestamp":"2026-01-01T00:31:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"B1","name":"Bash","input":{}}]}}`,
		`{"timestamp":"2026-01-01T00:32:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"B1","content":"ok"}]}}`,
	})
	_, spans, err := loadEvents([]string{dir}, mentionScope{}, true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("want exactly 1 Agent span, got %d: %+v", len(spans), spans)
	}
	if !spans[0].Start.Equal(tm("2026-01-01T00:00:00Z")) || !spans[0].End.Equal(tm("2026-01-01T00:30:00Z")) {
		t.Fatalf("span should be [00:00,00:30], got %+v", spans[0])
	}
}

func TestLoadEventsCodexTranscriptShape(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "codex.jsonl"), []string{
		`{"timestamp":"2026-06-26T16:55:08.631Z","type":"session_meta","payload":{"cwd":"/w/repo"}}`,
		`{"timestamp":"2026-06-26T16:55:08.635Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"work on #144"}]}}`,
		`{"timestamp":"2026-06-26T16:55:14.321Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"working on #144"}]}}`,
		`{"timestamp":"2026-06-26T16:55:14.322Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","arguments":"{\"cmd\":\"git status\"}"}}`,
	})

	evs, _, err := loadEvents([]string{dir}, newMentionScope("ariadne", []string{"144"}), true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 3 {
		t.Fatalf("want 3 Codex events, got %d: %+v", len(evs), evs)
	}
	total := map[string]int{}
	for _, e := range evs {
		for k, v := range e.Mentions {
			total[k] += v
		}
	}
	if total["144"] != 2 {
		t.Fatalf("want two #144 mentions from user+assistant messages, got %v", total)
	}

	evs, _, err = loadEvents([]string{dir}, newMentionScope("ariadne", []string{"144"}), false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || evs[0].Mentions["144"] != 1 {
		t.Fatalf("without assistant want only the Codex user message, got %+v", evs)
	}
}

func TestLoadEventsDanglingDispatchNoSpan(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		// no matching tool_result → no span.
	})
	_, spans, err := loadEvents([]string{dir}, mentionScope{}, true, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 0 {
		t.Fatalf("dangling dispatch should yield no span, got %+v", spans)
	}
}

// A span whose return lands after `until` is clamped to the window, not counted
// past it.
func TestLoadEventsSpanClampedToWindow(t *testing.T) {
	dir := t.TempDir()
	writeJSONL(t, filepath.Join(dir, "s.jsonl"), []string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"assistant","message":{"content":[{"type":"tool_use","id":"A1","name":"Agent","input":{}}]}}`,
		`{"timestamp":"2026-01-01T01:00:00Z","type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"A1","content":"done"}]}}`,
	})
	_, spans, err := loadEvents([]string{dir}, mentionScope{}, true, "", "2026-01-01T00:20:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 || !spans[0].End.Equal(tm("2026-01-01T00:20:00Z")) {
		t.Fatalf("span End should clamp to until 00:20, got %+v", spans)
	}
}

func writeJSONL(t *testing.T, path string, lines []string) {
	t.Helper()
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
