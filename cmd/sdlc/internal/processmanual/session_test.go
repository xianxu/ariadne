package processmanual

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Task 5 Test A: locateSessionJSONL resolves "current" to $CLAUDE_CODE_SESSION_ID
// (authoritative) when set + present, else the newest *.jsonl by mtime; an explicit
// path is returned as-is.
func TestLocateSessionJSONL(t *testing.T) {
	home := t.TempDir()
	repo := "/repo/ariadne"
	projDir := filepath.Join(home, ".claude", "projects", claudeProjectSlug(repo))
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	older := filepath.Join(projDir, "older.jsonl")
	newer := filepath.Join(projDir, "newer.jsonl")
	for _, p := range []string{older, newer} {
		if err := os.WriteFile(p, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	oldT := time.Now().Add(-2 * time.Hour)
	newT := time.Now().Add(-1 * time.Minute)
	if err := os.Chtimes(older, oldT, oldT); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newer, newT, newT); err != nil {
		t.Fatal(err)
	}

	// "current", env unset → newest by mtime.
	t.Setenv("CLAUDE_CODE_SESSION_ID", "")
	if got, err := locateSessionJSONL(home, repo, "current"); err != nil || got != newer {
		t.Errorf("current (no env) = (%q, %v); want %q", got, err, newer)
	}

	// "current", env set + file present → that file wins over newest-mtime (authoritative).
	t.Setenv("CLAUDE_CODE_SESSION_ID", "older")
	if got, err := locateSessionJSONL(home, repo, "current"); err != nil || got != older {
		t.Errorf("current (env=older) = (%q, %v); want %q (authoritative)", got, err, older)
	}

	// Explicit path → returned as-is (no stat; the read step surfaces a missing file).
	if got, err := locateSessionJSONL(home, repo, "/x/explicit.jsonl"); err != nil || got != "/x/explicit.jsonl" {
		t.Errorf("explicit path = (%q, %v); want as-is", got, err)
	}
}

// Task 3: segmentEvents splits the fired stream on a >60min lull in ALL activity
// (gap over allTimes) OR an away_summary boundary.
func TestSegmentEvents_gapAndAwaySummary(t *testing.T) {
	base := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	at := func(min int) time.Time { return base.Add(time.Duration(min) * time.Minute) }
	ev := func(min int) FiredEvent { return FiredEvent{Time: at(min), Kind: KindSkill} }
	shape := func(segs [][]FiredEvent) []int {
		out := make([]int, len(segs))
		for i, s := range segs {
			out[i] = len(s)
		}
		return out
	}

	// Gap: events at t, t+10, t+90 — the 80min lull before t+90 opens a new segment.
	segs := segmentEvents(
		[]FiredEvent{ev(0), ev(10), ev(90)},
		[]time.Time{at(0), at(10), at(90)},
		nil,
	)
	if got := shape(segs); len(got) != 2 || got[0] != 2 || got[1] != 1 {
		t.Fatalf("gap split: want [2 1], got %v", got)
	}

	// away_summary at t+2 (no 60min gap) still splits the two events around it.
	segs2 := segmentEvents(
		[]FiredEvent{ev(0), ev(5)},
		[]time.Time{at(0), at(2), at(5)},
		[]time.Time{at(2)},
	)
	if got := shape(segs2); len(got) != 2 || got[0] != 1 || got[1] != 1 {
		t.Fatalf("away split: want [1 1], got %v", got)
	}

	// No boundary → one segment.
	segs3 := segmentEvents([]FiredEvent{ev(0), ev(5)}, []time.Time{at(0), at(5)}, nil)
	if got := shape(segs3); len(got) != 1 || got[0] != 2 {
		t.Fatalf("no split: want [2], got %v", got)
	}
}

// fixtureJSONL assembles one JSON object per line — the tolerant parser's input.
func fixtureJSONL(lines ...string) []byte {
	return []byte(strings.Join(lines, "\n") + "\n")
}

// testVerbs is the real-verb set the classifier validates against (stands in for
// the catalog's help-text titles).
var testVerbs = map[string]bool{"close": true, "milestone-close": true, "state": true}

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
		// Regression: smoke-run false positives that the naive substring match wrongly
		// counted as fired verbs (#157 real-run finding).
		{"flag not a verb", `{"command":"grep -rn foo cmd/sdlc --include=*.go"}`, "", "", false},
		{"real verb in commit prose", `{"command":"git commit -m 'fix sdlc close bug'"}`, "", "", false},
		{"non-verb after sdlc", `{"command":"echo done; sdlc matcher notes"}`, "", "", false},
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
			kind, detail, ok := classifyToolUse(name, []byte(tc.input), testVerbs)
			if kind != tc.wantKind || detail != tc.wantDetail || ok != tc.wantOK {
				t.Errorf("classifyToolUse(%s, %s) = (%q, %q, %v); want (%q, %q, %v)",
					name, tc.input, kind, detail, ok, tc.wantKind, tc.wantDetail, tc.wantOK)
			}
		})
	}
}

// Task 4: renderSessionReport is pure markdown — segment headers, chronological
// lines, links resolved against the M1 catalog (with the help-text verb fallback),
// verdicts inline, and a header stating the two hard limits.
func TestRenderSessionReport(t *testing.T) {
	base := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	at := func(min int) time.Time { return base.Add(time.Duration(min) * time.Minute) }
	catalog := []InjectionSource{
		{Kind: KindSkill, Title: "xx-fix", Link: "construct/adapted/xx-fix/SKILL.md"},
		{Kind: KindLessons, Title: "workshop/lessons.md", Link: "workshop/lessons.md"},
		{Kind: KindHelpText, Title: "close", Link: "cmd/sdlc/helptext/close.md"},
		{Kind: KindSDLCPrompt, Title: "milestone-review", Link: "cmd/sdlc/internal/judge/prompts/milestone-review.md"},
	}
	segments := [][]FiredEvent{
		{
			{Time: at(0), Kind: KindSDLCPrompt, Detail: "close", Verdict: "SHIP"},
			{Time: at(6), Kind: KindSkill, Detail: "xx-fix"},
		},
		{
			// a verb with no help-text file → must render unlinked, not silently.
			{Time: at(90), Kind: KindSDLCPrompt, Detail: "bogus-verb"},
		},
	}
	out := renderSessionReport(segments, catalog, "")

	for _, want := range []string{"## Segment 1", "## Segment 2", "10:00:00", "10:06:00", "close", "xx-fix"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in report:\n%s", want, out)
		}
	}
	// close → help-text fallback link; skill → catalog SKILL.md link.
	if !strings.Contains(out, "cmd/sdlc/helptext/close.md") {
		t.Errorf("close should link to its help-text fallback:\n%s", out)
	}
	if !strings.Contains(out, "construct/adapted/xx-fix/SKILL.md") {
		t.Errorf("skill should link to its catalog SKILL.md:\n%s", out)
	}
	// verdict inline for the close event.
	if !strings.Contains(out, "SHIP") {
		t.Errorf("close verdict SHIP should render inline:\n%s", out)
	}
	// header must state BOTH hard limits.
	if !strings.Contains(out, "agents-chain") || !strings.Contains(out, "memory") {
		t.Errorf("header must state the agents-chain/memory limit:\n%s", out)
	}
	if !strings.Contains(strings.ToLower(out), "forked") {
		t.Errorf("header must state the forked-prompt limit:\n%s", out)
	}
	// unresolved verb: its detail renders, but NOT as a markdown link.
	if !strings.Contains(out, "bogus-verb") {
		t.Errorf("unresolved verb should still render its detail:\n%s", out)
	}
	if strings.Contains(out, "[bogus-verb]") {
		t.Errorf("unresolved verb must render UNLINKED (no [..](..)):\n%s", out)
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

	events, allTimes, awaySummaryTimes, err := parseEvents(data, testVerbs)
	if err != nil {
		t.Fatalf("parseEvents returned error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("want 3 fired events, got %d: %+v", len(events), events)
	}

	// Ordered: close, skill, lessons.
	if events[0].Kind != KindSDLCPrompt || events[0].Detail != "close" {
		t.Errorf("event[0] = %+v; want close/KindSDLCPrompt", events[0])
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

// Task 1 (boundary-review I1): a RE-CLOSE streams trailer-only stdout — no fresh
// reviewer body, just the `Review-Verdict:` git-trailer. The verdict must still be
// recovered (via judge.ParseVerdictTrailer), not dropped as ParseVerdict alone would.
func TestParseEvents_trailerOnlyVerdict(t *testing.T) {
	data := fixtureJSONL(
		`{"type":"assistant","timestamp":"2026-07-01T11:00:00.000Z","message":{"content":[{"type":"tool_use","id":"toolu_re","name":"Bash","input":{"command":"sdlc close --issue 9 --no-verified"}}]}}`,
		`{"type":"user","timestamp":"2026-07-01T11:01:00.000Z","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_re"}]},"toolUseResult":{"stdout":"Review-Verdict: FIX-THEN-SHIP\nReview-Window: abc123..HEAD\n","stderr":"","interrupted":false}}`,
	)
	events, _, _, err := parseEvents(data, testVerbs)
	if err != nil {
		t.Fatalf("parseEvents error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("want 1 fired event, got %d", len(events))
	}
	if events[0].Verdict != "FIX-THEN-SHIP" {
		t.Errorf("trailer-only verdict = %q; want FIX-THEN-SHIP (recovered via trailer fallback)", events[0].Verdict)
	}
}
