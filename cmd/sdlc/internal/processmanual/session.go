// session.go — the DYNAMIC counterpart to the static process manual (#157). Given
// a Claude session transcript (JSONL), reconstruct which catalogued injection
// points actually FIRED, in timestamp order, segmented on the 60-min-gap /
// away_summary boundary, each matched to its M1 `Kind`.
//
// It lives beside the M1 catalog on purpose (ARCH-DRY): reconstruction matches
// against the in-process `InjectionSource` set rather than serializing across a
// shell boundary to Python. The pure core (parseEvents / classifyToolUse /
// segmentEvents / renderSessionReport) does no IO; only locateSessionJSONL + the
// file read touch the filesystem.
//
// Two hard limits are documented, not fought (verified against 68 local sessions):
//  1. agents-chain (AGENTS/CLAUDE.md) + memory are session-start SYSTEM-PROMPT
//     injections that never appear in the transcript — only their explicit
//     mid-session Reads would show, so we can assert availability, never firing.
//  2. Forked review PROMPTS aren't in the main JSONL — only their OUTPUT, streamed
//     back through the `sdlc close`/`milestone-close` Bash stdout (where we recover
//     the verdict via judge.ParseVerdict).
package processmanual

import (
	"bufio"
	"bytes"
	"encoding/json"
	"path"
	"regexp"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/judge"
)

// FiredEvent is one injection that actually fired in a session — the dynamic
// record that REFERENCES the M1 catalog (via Kind + Detail), rather than mutating
// the static InjectionSource. Its link is resolved against the catalog at render
// time (renderSessionReport), so this stays pure parse output.
type FiredEvent struct {
	Time    time.Time
	Kind    Kind
	Tool    string // "Bash" | "Skill" | "Read"
	Detail  string // verb / skill name / file basename
	Verdict string // optional — review verdict for close/milestone-close
}

// rec is the tolerant JSONL record. Only the fields we consume are named; unknown
// record types (newer sessions add bookkeeping types) unmarshal into a rec whose
// Type we don't switch on and are skipped. encoding/json matches keys
// case-insensitively (Type↔type, ID↔id, …); ToolUseID needs its tag because
// "ToolUseID" won't match the underscored "tool_use_id".
type rec struct {
	Type      string    `json:"type"`
	Subtype   string    `json:"subtype"`
	Timestamp time.Time `json:"timestamp"`
	Message   struct {
		Content []struct {
			Type      string          `json:"type"` // "tool_use" | "tool_result"
			Name      string          `json:"name"`
			ID        string          `json:"id"`
			Input     json.RawMessage `json:"input"`
			ToolUseID string          `json:"tool_use_id"` // tool_result → its tool_use's id
		} `json:"content"`
	} `json:"message"`
	ToolUseResult json.RawMessage `json:"toolUseResult"` // polymorphic: {stdout,…} | string | null
}

// parseEvents is pure over bytes. It scans the JSONL, keeps the fired injections
// (classifyToolUse), recovers close/milestone-close verdicts from the following
// tool_result's stdout (linked by tool_use_id), and reports segmentation inputs:
// allTimes (every record's timestamp — so non-injection work between fired events
// doesn't trigger a false gap split) and awaySummaryTimes.
func parseEvents(data []byte) (events []FiredEvent, allTimes []time.Time, awaySummaryTimes []time.Time, err error) {
	// A fired event plus the tool_use id it needs for verdict recovery — dropped
	// once the verdict is linked, so FiredEvent stays free of transcript plumbing.
	type pending struct {
		ev FiredEvent
		id string
	}
	var pend []pending
	stdoutByID := map[string]string{}

	sc := bufio.NewScanner(bytes.NewReader(data))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024) // transcript lines can be large
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r rec
		if uerr := json.Unmarshal(line, &r); uerr != nil {
			continue // tolerant: a malformed line is skipped, not fatal
		}
		if !r.Timestamp.IsZero() {
			allTimes = append(allTimes, r.Timestamp)
		}
		switch r.Type {
		case "assistant":
			for _, c := range r.Message.Content {
				if c.Type != "tool_use" {
					continue
				}
				kind, detail, ok := classifyToolUse(c.Name, c.Input)
				if !ok {
					continue
				}
				pend = append(pend, pending{
					ev: FiredEvent{Time: r.Timestamp, Kind: kind, Tool: c.Name, Detail: detail},
					id: c.ID,
				})
			}
		case "user":
			// The tool_result record carries BOTH its tool_use_id (in content) and the
			// top-level toolUseResult.stdout — so a single record links id → stdout.
			for _, c := range r.Message.Content {
				if c.Type == "tool_result" && c.ToolUseID != "" {
					if sd, ok := extractStdout(r.ToolUseResult); ok {
						stdoutByID[c.ToolUseID] = sd
					}
				}
			}
		case "system":
			if r.Subtype == "away_summary" && !r.Timestamp.IsZero() {
				awaySummaryTimes = append(awaySummaryTimes, r.Timestamp)
			}
		}
	}
	if serr := sc.Err(); serr != nil {
		return nil, nil, nil, serr
	}

	// Resolve verdicts after the full scan (the tool_result follows its tool_use, so
	// the map must be complete first). judge.ParseVerdict is the exact fn `sdlc close`
	// uses (ARCH-DRY): it reads the reviewer's `VERDICT:`/block output, not the trailer.
	for _, p := range pend {
		ev := p.ev
		if ev.Kind == KindSDLCPrompt && (ev.Detail == "close" || ev.Detail == "milestone-close") {
			if sd, ok := stdoutByID[p.id]; ok {
				if v := judge.ParseVerdict(sd); v != judge.VerdictUnknown {
					ev.Verdict = string(v)
				}
			}
		}
		events = append(events, ev)
	}
	return events, allTimes, awaySummaryTimes, nil
}

// extractStdout pulls .stdout from the polymorphic toolUseResult. It is a dict
// {stdout,…} for Bash; a string or null for other tools — both of which fail the
// struct unmarshal, which we swallow to stay tolerant.
func extractStdout(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var obj struct {
		Stdout string `json:"stdout"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", false // string / null / non-object → no stdout
	}
	if obj.Stdout == "" {
		return "", false
	}
	return obj.Stdout, true
}

// sdlcVerbRE matches an `sdlc <verb>` invocation anchored on a word boundary, so
// `sdlcx` and `mysdlc ` don't match. Ports the proven jq matcher from the issue's
// grounding digest. The verb is [a-z-]+ (e.g. close, milestone-close, state).
var sdlcVerbRE = regexp.MustCompile(`(^|[^a-zA-Z])sdlc ([a-z-]+)`)

// classifyToolUse is the pure match table (ports the IDEA of introspect's
// segment_text.py summarize_tool_input, not its code): the three injection-bearing
// tool calls we can see in a transcript. Anything else → ok=false.
func classifyToolUse(name string, input json.RawMessage) (Kind, string, bool) {
	switch name {
	case "Skill":
		var in struct {
			Skill string `json:"skill"`
		}
		if json.Unmarshal(input, &in) == nil && in.Skill != "" {
			return KindSkill, in.Skill, true
		}
	case "Bash":
		var in struct {
			Command string `json:"command"`
		}
		if json.Unmarshal(input, &in) != nil {
			return "", "", false
		}
		m := sdlcVerbRE.FindStringSubmatch(in.Command)
		if m == nil {
			return "", "", false
		}
		verb := m[2]
		// `sdlc <verb> --help` prints embedded help text (a distinct Kind from the
		// injected review/gate prompts the bare verb fires).
		if bytes.Contains([]byte(in.Command), []byte("--help")) {
			return KindHelpText, verb, true
		}
		return KindSDLCPrompt, verb, true
	case "Read":
		var in struct {
			FilePath string `json:"file_path"`
		}
		if json.Unmarshal(input, &in) != nil {
			return "", "", false
		}
		base := path.Base(in.FilePath)
		if base == "lessons.md" {
			return KindLessons, base, true
		}
	}
	return "", "", false
}
