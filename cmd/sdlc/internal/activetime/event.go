// Package activetime is the native Go port of the former active-time-v3.py
// (removed in #110; see git history for the Python original) — the
// segment-anchored per-issue dev-hour attribution behind `sdlc actual` (#68,
// #110). It collapses what was a python3 subprocess + stdout-regex + script-
// resolution into an in-process engine: a pure core (gap-truncated active
// minutes, segment construction, the commit-weight/mention split) behind a thin
// IO seam (transcript .jsonl event loading + a git-log window loader).
//
// Provenance comments below name the Python functions each Go function ports;
// the source lives in git history, not the tree.
//
// Semantics are preserved 1:1 with the Python so measured actuals don't shift;
// the only intentional divergences are cosmetic (the unattributed bucket renders
// as "unattributed", not the Python's "##unattributed").
package activetime

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
}

// TaskSpan is one synchronous subagent execution: the interval between an
// assistant `tool_use` dispatch (name "Agent") and its matching `tool_result`
// return, both timestamped in the operator's main transcript. The subagent runs
// in its own transcript (outside the dirs we read), so this gap shows as one big
// inter-event gap; it is active project work, not idle, and must count in full
// rather than truncate at the 15-min cap (#118 — measure ship wall-clock).
type TaskSpan struct {
	Start, End time.Time
}

// rawLine is one decoded transcript JSONL record. Content is left raw because it
// is polymorphic (string for a plain user turn, array of blocks otherwise).
type rawLine struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   struct {
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// contentBlock is one element of a content array. Content (the nested key) is
// raw because some block shapes carry a string under "content".
type contentBlock struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	Content json.RawMessage `json:"content"`
}

// walkSessionEvents parses one session .jsonl file into the events we attribute.
// Mirrors active-time-v3.py walk_session_events exactly, including the load-
// bearing user/assistant asymmetry: a user turn is dropped on empty text (and a
// pure tool_result is dropped), but an assistant turn (when includeAssistant) is
// ALWAYS emitted — its timestamp counts toward active time even with no text.
func walkSessionEvents(path string, pat *regexp.Regexp, includeAssistant bool) ([]Event, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var events []Event
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
		ts, err := parseISO(d.Timestamp)
		if err != nil {
			continue
		}
		events = append(events, Event{Time: ts, Mentions: parseEventMentions(text, pat)})
	}
	return events, nil
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
// Python: ts < since skip, ts > until skip), sorted by time. Mirrors load_events.
func loadEvents(dirs []string, pat *regexp.Regexp, includeAssistant bool, sinceISO, untilISO string) ([]Event, error) {
	var since, until time.Time
	haveSince, haveUntil := sinceISO != "", untilISO != ""
	if haveSince {
		t, err := parseISO(sinceISO)
		if err != nil {
			return nil, err
		}
		since = t
	}
	if haveUntil {
		t, err := parseISO(untilISO)
		if err != nil {
			return nil, err
		}
		until = t
	}
	var events []Event
	for _, d := range dirs {
		dir := expandUser(d)
		fi, err := os.Stat(dir)
		if err != nil || !fi.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		for _, f := range files {
			evs, err := walkSessionEvents(f, pat, includeAssistant)
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
				events = append(events, e)
			}
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Time.Before(events[j].Time) })
	return events, nil
}
