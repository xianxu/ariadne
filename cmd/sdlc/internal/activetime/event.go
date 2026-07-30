// Package activetime is the native Go active-time-v3 engine behind `sdlc actual`
// (#68, #110, #92). It collapses what was a python3 subprocess + stdout-regex +
// script-resolution into an in-process engine: a pure core (gap-truncated active
// runs, global commit-boundary attribution, warning computation) behind a thin
// IO seam (transcript .jsonl event loading + a git-log window loader).
//
// Provenance comments below name the Python functions each Go function ports;
// the source lives in git history, not the tree.
//
// Event decoding still follows the Python script closely. Attribution no longer
// preserves the Python segment loop: #92 moved it to source-aware activity runs
// claimed by nearby issue commit boundaries.
package activetime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Event is one transcript line we attribute: a human user turn, or — when
// IncludeAssistant is set — an assistant text turn. Mentions holds the count of
// tracked-issue #N refs found in this single event's text.
//
// Asymmetry worth remembering (mirrors active-time-v3.py): a user event is
// dropped when its text is empty/whitespace, but an assistant event is ALWAYS
// emitted (even with no text and no mentions) — its timestamp is part of the
// active-time stream. See walkSessionEvents.
type Event struct {
	Time     time.Time
	Mentions map[string]int
	Source   string
}

// TaskSpan is one synchronous subagent execution: the interval between an
// assistant `tool_use` dispatch (name "Agent") and its matching `tool_result`
// return, both timestamped in the operator's main transcript. The subagent runs
// in its own transcript (outside the dirs we read), so this gap shows as one big
// inter-event gap; it is active project work, not idle, and must count in full
// rather than truncate at the 15-min cap (#118 — measure ship wall-clock).
type TaskSpan struct {
	Start, End time.Time
	Source     string
}

// rawLine is one decoded transcript JSONL record. Content is left raw because it
// is polymorphic (string for a plain user turn, array of blocks otherwise).
type rawLine struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// contentBlock is one element of a content array. Content (the nested key) is
// raw because some block shapes carry a string under "content".
type contentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Content   json.RawMessage `json:"content"`
	Name      string          `json:"name"`        // tool_use: tool name (e.g. "Agent")
	ID        string          `json:"id"`          // tool_use: id matched by a later tool_result
	ToolUseID string          `json:"tool_use_id"` // tool_result: id of the dispatch it answers
}

type codexPayload struct {
	Type      string          `json:"type"`
	Role      string          `json:"role"`
	Content   json.RawMessage `json:"content"`
	Name      string          `json:"name"`
	Arguments string          `json:"arguments"`
	Input     json.RawMessage `json:"input"`
}

// walkSessionEvents parses one session .jsonl file into the events we attribute.
// Mirrors active-time-v3.py walk_session_events exactly, including the load-
// bearing user/assistant asymmetry: a user turn is dropped on empty text (and a
// pure tool_result is dropped), but an assistant turn (when includeAssistant) is
// ALWAYS emitted — its timestamp counts toward active time even with no text.
func walkSessionEvents(path string, sc mentionScope, includeAssistant bool) ([]Event, []TaskSpan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var events []Event
	var spans []TaskSpan
	pending := map[string]time.Time{} // Agent dispatch id → dispatch time
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var d rawLine
		if json.Unmarshal([]byte(line), &d) != nil {
			continue // malformed line — skip (Python's except: continue)
		}
		if d.Timestamp == "" {
			continue
		}
		// ts is parsed once, up front: a line with an unparseable timestamp is
		// dropped (same outcome as before, where the parse failed after text
		// extraction) and contributes to neither events nor spans.
		ts, err := parseISO(d.Timestamp)
		if err != nil {
			continue
		}
		if d.Type == "response_item" {
			ev, ok := codexEvent(d.Payload, ts, sc, includeAssistant)
			if ok {
				events = append(events, ev)
			}
			continue
		}
		// Structural span tracking — independent of includeAssistant and of the
		// text-skip below: a pure tool_result is dropped as an Event but still
		// closes a span. decodeBlocks reports false for non-array content (which
		// carries no tool blocks).
		if blocks, ok := decodeBlocks(d.Message.Content); ok {
			for _, blk := range blocks {
				switch {
				case d.Type == "assistant" && blk.Type == "tool_use" && blk.Name == "Agent" && blk.ID != "":
					pending[blk.ID] = ts
				case d.Type == "user" && blk.Type == "tool_result" && blk.ToolUseID != "":
					if start, ok := pending[blk.ToolUseID]; ok {
						spans = append(spans, TaskSpan{Start: start, End: ts})
						delete(pending, blk.ToolUseID)
					}
				}
			}
		}
		// Event emission — unchanged logic (the user/assistant asymmetry is intact).
		var text string
		switch {
		case d.Type == "user":
			t, skip := userText(d.Message.Content)
			if skip {
				continue // pure tool_result, non-string/list content, or empty text
			}
			text = t
		case d.Type == "assistant" && includeAssistant:
			text = assistantText(d.Message.Content)
			// No empty-text skip here — assistant events always emit.
		default:
			continue
		}
		events = append(events, Event{Time: ts, Mentions: parseEventMentions(text, sc)})
	}
	return events, spans, nil
}

func codexEvent(payload json.RawMessage, ts time.Time, sc mentionScope, includeAssistant bool) (Event, bool) {
	var p codexPayload
	if json.Unmarshal(payload, &p) != nil {
		return Event{}, false
	}
	switch p.Type {
	case "message":
		text := codexMessageText(p.Content)
		switch p.Role {
		case "user":
			if strings.TrimSpace(text) == "" {
				return Event{}, false
			}
			return Event{Time: ts, Mentions: parseEventMentions(text, sc)}, true
		case "assistant":
			if !includeAssistant {
				return Event{}, false
			}
			return Event{Time: ts, Mentions: parseEventMentions(text, sc)}, true
		default:
			return Event{}, false
		}
	case "function_call", "custom_tool_call":
		if !includeAssistant {
			return Event{}, false
		}
		text := p.Arguments
		if text == "" && len(p.Input) > 0 {
			text = string(p.Input)
		}
		return Event{Time: ts, Mentions: parseEventMentions(text, sc)}, true
	default:
		return Event{}, false
	}
}

func codexMessageText(content json.RawMessage) string {
	var s string
	if json.Unmarshal(content, &s) == nil {
		return s
	}
	blocks, ok := decodeBlocks(content)
	if !ok {
		return ""
	}
	var parts []string
	for _, blk := range blocks {
		switch blk.Type {
		case "input_text", "output_text", "text":
			parts = append(parts, blk.Text)
		default:
			var cs string
			if len(blk.Content) > 0 && string(blk.Content) != "null" && json.Unmarshal(blk.Content, &cs) == nil {
				parts = append(parts, cs)
			}
		}
	}
	return strings.Join(parts, "\n")
}

// userText extracts the human-typed text from a user message's content and
// reports whether the event should be skipped (pure tool_result with no text,
// content that is neither string nor list, or empty/whitespace text). Mirrors
// the `t == "user"` branch of walk_session_events.
func userText(content json.RawMessage) (text string, skip bool) {
	var s string
	if json.Unmarshal(content, &s) == nil {
		// string content used directly.
	} else {
		blocks, ok := decodeBlocks(content)
		if !ok {
			return "", true // neither string nor list → skip
		}
		sawToolResult := false
		var parts []string
		for _, blk := range blocks {
			switch {
			case blk.Type == "tool_result":
				sawToolResult = true
			case blk.Type == "text":
				parts = append(parts, blk.Text)
			default:
				// A block carrying a string under "content" contributes it
				// (Python's `"content" in blk and isinstance(blk["content"], str)`).
				// Guard against explicit JSON null — unmarshaling "null" into a
				// string is a no-op (no error) and would append "", whereas
				// Python's isinstance(None, str) skips it.
				var cs string
				if len(blk.Content) > 0 && string(blk.Content) != "null" && json.Unmarshal(blk.Content, &cs) == nil {
					parts = append(parts, cs)
				}
			}
		}
		if sawToolResult && len(parts) == 0 {
			return "", true // pure tool result, not human typing
		}
		s = strings.Join(parts, "\n")
	}
	if strings.TrimSpace(s) == "" {
		return "", true
	}
	return s, false
}

// assistantText joins the text blocks of an assistant message (empty string when
// content is not a block array). Mirrors the `t == "assistant"` branch.
func assistantText(content json.RawMessage) string {
	blocks, ok := decodeBlocks(content)
	if !ok {
		return ""
	}
	var parts []string
	for _, blk := range blocks {
		if blk.Type == "text" {
			parts = append(parts, blk.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// decodeBlocks parses content as an array of blocks, skipping non-object
// elements (Python's `if isinstance(blk, dict)` guard). Reports false when
// content is not a JSON array at all.
func decodeBlocks(content json.RawMessage) ([]contentBlock, bool) {
	var raw []json.RawMessage
	if json.Unmarshal(content, &raw) != nil {
		return nil, false
	}
	blocks := make([]contentBlock, 0, len(raw))
	for _, rb := range raw {
		var blk contentBlock
		if json.Unmarshal(rb, &blk) != nil {
			continue // non-object element — skip
		}
		blocks = append(blocks, blk)
	}
	return blocks, true
}

// eventTimesAndMentions returns the timestamps of events (in order) and the
// summed tracked-issue mention counts across them. Shared by buildSegments (per
// segment) and Compute's no-commits whole-window fallback.
func eventTimesAndMentions(events []Event) ([]time.Time, map[string]int) {
	times := make([]time.Time, len(events))
	mentions := map[string]int{}
	for i, e := range events {
		times[i] = e.Time
		for iss, n := range e.Mentions {
			mentions[iss] += n
		}
	}
	return times, mentions
}

// loadEvents loads all attributed events across every session file under each
// dir, filtered to [sinceISO, untilISO] (inclusive bounds, mirroring the
// Python: ts < since skip, ts > until skip), sorted by time. It also collects
// the subagent TaskSpans (#118), clamped to the same window so all measured
// active time lies within it. Mirrors load_events (events) + adds the span pass.
func loadEvents(dirs []string, sc mentionScope, includeAssistant bool, sinceISO, untilISO string) ([]Event, []TaskSpan, error) {
	return loadEventsWithFiles(dirs, nil, sc, includeAssistant, sinceISO, untilISO)
}

func loadEventsWithFiles(dirs, files []string, sc mentionScope, includeAssistant bool, sinceISO, untilISO string) ([]Event, []TaskSpan, error) {
	var since, until time.Time
	haveSince, haveUntil := sinceISO != "", untilISO != ""
	if haveSince {
		t, err := parseISO(sinceISO)
		if err != nil {
			return nil, nil, err
		}
		since = t
	}
	if haveUntil {
		t, err := parseISO(untilISO)
		if err != nil {
			return nil, nil, err
		}
		until = t
	}
	var events []Event
	var spans []TaskSpan
	for _, d := range dirs {
		dir := expandUser(d)
		fi, err := os.Stat(dir)
		if err != nil || !fi.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		for _, f := range files {
			evs, sps, err := walkSessionEvents(f, sc, includeAssistant)
			if err != nil {
				// Skip a session file that became unreadable after globbing —
				// more robust than Python's unguarded open() (which would crash),
				// an intentional divergence from the otherwise 1:1 port.
				continue
			}
			for _, e := range evs {
				if haveSince && e.Time.Before(since) {
					continue
				}
				if haveUntil && e.Time.After(until) {
					continue
				}
				e.Source = f
				events = append(events, e)
			}
			for _, s := range sps {
				// Clamp to the window (not just filter by Start) so a span still
				// returning at the close instant is counted only up to `until`.
				if haveSince && s.Start.Before(since) {
					s.Start = since
				}
				if haveUntil && s.End.After(until) {
					s.End = until
				}
				if s.End.After(s.Start) {
					s.Source = f
					spans = append(spans, s)
				}
			}
		}
	}
	for _, f := range files {
		evs, sps, err := walkSessionEvents(expandUser(f), sc, includeAssistant)
		if err != nil {
			continue
		}
		for _, e := range evs {
			if haveSince && e.Time.Before(since) {
				continue
			}
			if haveUntil && e.Time.After(until) {
				continue
			}
			e.Source = f
			events = append(events, e)
		}
		for _, sp := range sps {
			if haveSince && sp.End.Before(since) {
				continue
			}
			if haveUntil && sp.Start.After(until) {
				continue
			}
			if haveSince && sp.Start.Before(since) {
				sp.Start = since
			}
			if haveUntil && sp.End.After(until) {
				sp.End = until
			}
			if sp.End.After(sp.Start) {
				sp.Source = f
				spans = append(spans, sp)
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Time.Before(events[j].Time) })
	sort.Slice(spans, func(i, j int) bool { return spans[i].Start.Before(spans[j].Start) })
	return events, spans, nil
}
