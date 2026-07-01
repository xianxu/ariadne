package processmanual

import (
	"strings"
	"testing"
)

// fixtureJSONL assembles one JSON object per line — the tolerant parser's input.
func fixtureJSONL(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

// Task 2: classifyToolUse is the pure match table — the three injection-bearing
// tool calls, with the sdlc matcher anchored so `sdlcx` doesn't false-positive.
func TestClassifyToolUse_table(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		wantKind   Kind
		wantDetail string
		wantOK     bool
	}{
		{"skill", `{"skill":"xx-fix"}`, KindSkill, "xx-fix", true},
		{"sdlc verb", `{"command":"sdlc close --issue 9"}`, KindSDLCPrompt, "close", true},
		{"sdlc milestone-close", `{"command":"sdlc milestone-close --issue 9 --milestone M2"}`, KindSDLCPrompt, "milestone-close", true},
		{"sdlc help", `{"command":"sdlc state --help"}`, KindHelpText, "state", true},
		{"lessons read", `{"file_path":"/repo/workshop/lessons.md"}`, KindLessons, "lessons.md", true},
		{"plain bash", `{"command":"ls -la"}`, "", "", false},
		{"sdlcx not anchored", `{"command":"echo sdlcx"}`, "", "", false},
		{"non-injection read", `{"file_path":"/repo/cmd/main.go"}`, "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name := "Bash"
			switch {
			case tc.name == "skill":
				name = "Skill"
			case tc.name == "lessons read" || tc.name == "non-injection read":
				name = "Read"
			}
			kind, detail, ok := classifyToolUse(name, []byte(tc.input))
			if kind != tc.wantKind || detail != tc.wantDetail || ok != tc.wantOK {
				t.Errorf("classifyToolUse(%s, %s) = (%q, %q, %v); want (%q, %q, %v)",
					name, tc.input, kind, detail, ok, tc.wantKind, tc.wantDetail, tc.wantOK)
			}
		})
	}
}

// Task 1: parseEvents is pure over bytes. It must recover the fired injections in
// order, link a close's review verdict from the following tool_result's stdout,
// collect away_summary boundaries, and skip unknown record types without erroring.
func TestParseEvents_firedStreamAndVerdict(t *testing.T) {
	data := fixtureJSONL(
		// assistant: a Bash `sdlc close` (id toolu_close) — its verdict lives in the
		// following user record's toolUseResult.stdout, linked by tool_use_id.
		`{"type":"assistant","timestamp":"2026-07-01T10:00:00.000Z","message":{"content":[{"type":"tool_use","id":"toolu_close","name":"Bash","input":{"command":"sdlc close --issue 9 --verified 'done'","description":"close issue"}}]}}`,
		// user tool_result: the forked review's output streamed into Bash stdout. The
		// reviewer body carries a `VERDICT: SHIP` line (what ParseVerdict reads), then
		// the `Review-Verdict:` git-trailer (which ParseVerdict alone returns unknown for).
		`{"type":"user","timestamp":"2026-07-01T10:05:00.000Z","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_close"}]},"toolUseResult":{"stdout":"running review...\nVERDICT: SHIP (confidence: high)\n\nno blocking issues.\n\nReview-Verdict: SHIP\n","stderr":"","interrupted":false}}`,
		// assistant: a Skill tool_use.
		`{"type":"assistant","timestamp":"2026-07-01T10:06:00.000Z","message":{"content":[{"type":"tool_use","id":"toolu_skill","name":"Skill","input":{"skill":"xx-fix"}}]}}`,
		// assistant: a Read of workshop/lessons.md.
		`{"type":"assistant","timestamp":"2026-07-01T10:07:00.000Z","message":{"content":[{"type":"tool_use","id":"toolu_read","name":"Read","input":{"file_path":"/repo/workshop/lessons.md"}}]}}`,
		// system away_summary: a segment boundary.
		`{"type":"system","subtype":"away_summary","timestamp":"2026-07-01T10:08:00.000Z","content":"user stepped away"}`,
		// an unknown bookkeeping type — must be skipped, not error.
		`{"type":"file-history-snapshot","timestamp":"2026-07-01T10:09:00.000Z","messageId":"x"}`,
	)

	events, allTimes, awaySummaryTimes, err := parseEvents(data)
	if err != nil {
		t.Fatalf("parseEvents returned error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("want 3 fired events, got %d: %+v", len(events), events)
	}

	// Ordered: close, skill, lessons.
	if events[0].Kind != KindSDLCPrompt || events[0].Tool != "Bash" || events[0].Detail != "close" {
		t.Errorf("event[0] = %+v; want Bash/close/KindSDLCPrompt", events[0])
	}
	if events[0].Verdict != "SHIP" {
		t.Errorf("event[0].Verdict = %q; want SHIP (recovered via tool_use_id → stdout → judge.ParseVerdict)", events[0].Verdict)
	}
	if events[1].Kind != KindSkill || events[1].Detail != "xx-fix" {
		t.Errorf("event[1] = %+v; want Skill/xx-fix/KindSkill", events[1])
	}
	if events[2].Kind != KindLessons || events[2].Detail != "lessons.md" {
		t.Errorf("event[2] = %+v; want Read/lessons.md/KindLessons", events[2])
	}

	if len(awaySummaryTimes) != 1 {
		t.Errorf("want 1 away_summary time, got %d", len(awaySummaryTimes))
	}
	if len(allTimes) != 6 {
		t.Errorf("want 6 record timestamps (incl. the unknown line), got %d", len(allTimes))
	}
}
