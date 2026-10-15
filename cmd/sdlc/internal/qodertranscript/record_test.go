package qodertranscript

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func parse(t *testing.T, s string) Record {
	t.Helper()
	r, e := ParseRecord([]byte(s))
	if e != nil {
		t.Fatal(e)
	}
	return r
}
func TestParseRecordWhitelist(t *testing.T) {
	r := parse(t, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26.567Z","parentUuid":"other","content":"PRIVATE","message":{"id":"R","model":"auto","stop_reason":"end_turn","content":"PRIVATE","usage":{"credits":0.9265415628571427,"billable":true,"input_tokens":0,"request_id":"PRIVATE"}}}`)
	if r.SessionID != "S" || r.Message.ID != "R" || string(r.Message.Usage["credits"]) != "0.9265415628571427" {
		t.Fatalf("bad projection: %+v", r)
	}
	b, _ := json.Marshal(r)
	if strings.Contains(string(b), "PRIVATE") {
		t.Fatal("retained unwhitelisted fields")
	}
	if len(r.Message.Usage) != 3 {
		t.Fatal("unexpected usage fields")
	}
	c := parse(t, `{"type":"runtime-config","sessionId":"S","timestamp":1791989630430,"model":"auto","reasoningEffort":null}`)
	if !c.Timestamp.Equal(time.UnixMilli(1791989630430)) || c.ReasoningEffort != nil {
		t.Fatal("bad config")
	}
}
func TestParseRecordInvalidRedacted(t *testing.T) {
	for _, s := range []string{`null`, `[]`, `{"type":"runtime-config","timestamp":"PRIVATE"}`, `{"type":"user","timestamp":1791989630430}`, `{"type":"assistant","timestamp":"PRIVATE"}`, `{"type":"user","sessionId":"` + strings.Repeat("x", MaxStringBytes+1) + `"}`, `{"type":"user","timestamp":null}`, `{"type":"user","timestamp":"2026-10-14T15:43:26Z","sessionId":"S","sessionId":"T"}`} {
		_, e := ParseRecord([]byte(s))
		if e == nil {
			t.Fatalf("accepted invalid record")
		}
		if strings.Contains(e.Error(), "PRIVATE") {
			t.Fatal("private error")
		}
	}
	if r, e := ParseRecord([]byte(`{"type":"unknown","message":{"content":"PRIVATE"}}`)); e != nil || r.Type != "unknown" || r.Message != nil {
		t.Fatal("unknown record")
	}
}
func TestSessionReplayAndChronology(t *testing.T) {
	rows := []string{`{"type":"workspace-directories","sessionId":"S","directories":["/workspace/repo"]}`, `{"type":"runtime-config","sessionId":"S","timestamp":2000,"model":"new","reasoningEffort":"high"}`, `{"type":"runtime-config","sessionId":"S","timestamp":1000,"model":"old"}`, `{"type":"user","sessionId":"S","timestamp":"2026-10-14T15:43:26Z"}`, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:44:26Z"}`}
	var want Metadata
	for rotation := 0; rotation < len(rows); rotation++ {
		var state SessionState
		for repeat := 0; repeat < 2; repeat++ {
			for j := range rows {
				if e := state.Step(parse(t, rows[(j+rotation)%len(rows)])); e != nil {
					t.Fatal(e)
				}
			}
		}
		m, e := state.Metadata()
		if e != nil {
			t.Fatal(e)
		}
		if rotation == 0 {
			want = m
		}
		if m != want {
			t.Fatal("arrival order changed metadata")
		}
		model, effort, e := state.ConfigAt(time.UnixMilli(2500))
		if e != nil || model != "new" || effort != "high" {
			t.Fatal("config rewound")
		}
		model, _, e = state.ConfigAt(time.UnixMilli(1500))
		if e != nil || model != "old" {
			t.Fatal("wrong epoch")
		}
	}
}
func TestSessionIncompleteAndPoison(t *testing.T) {
	var s SessionState
	if e := s.Step(parse(t, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26Z"}`)); e != nil {
		t.Fatal(e)
	}
	if id, e := s.Identity(); e != nil || id != "S" {
		t.Fatal("lost usable spend identity")
	}
	if _, e := s.Metadata(); !errors.Is(e, ErrIncomplete) {
		t.Fatal("missing cwd complete")
	}
	for _, bad := range []string{`{"type":"workspace-directories","sessionId":"T","directories":["/workspace/repo"]}`, `{"type":"workspace-directories","sessionId":"S","directories":["/workspace/a","/workspace/b"]}`} {
		var st SessionState
		_ = st.Step(parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/workspace/repo"]}`))
		if e := st.Step(parse(t, bad)); !errors.Is(e, ErrConflict) {
			t.Fatal("not poisoned")
		}
		_ = st.Step(parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/workspace/repo"]}`))
		if _, e := st.Metadata(); !errors.Is(e, ErrConflict) {
			t.Fatal("repaired poison")
		}
		if strings.Contains(bad, `"sessionId":"S"`) {
			if id, e := st.Identity(); e != nil || id != "S" {
				t.Fatal("workspace conflict erased response identity")
			}
		}
	}
}
func TestConfigConflictIsScoped(t *testing.T) {
	var s SessionState
	for _, row := range []string{`{"type":"runtime-config","sessionId":"S","timestamp":1000,"model":"a"}`, `{"type":"runtime-config","sessionId":"S","timestamp":2000,"model":"b"}`, `{"type":"runtime-config","sessionId":"S","timestamp":2000,"model":"c"}`} {
		_ = s.Step(parse(t, row))
	}
	if _, e := s.Identity(); e != nil {
		t.Fatal("config ambiguity poisoned identity")
	}
	if m, _, e := s.ConfigAt(time.UnixMilli(1500)); e != nil || m != "a" {
		t.Fatal("earlier epoch lost")
	}
	if _, _, e := s.ConfigAt(time.UnixMilli(2500)); !errors.Is(e, ErrConflict) {
		t.Fatal("conflicting epoch accepted")
	}
}
func FuzzParseRecord(f *testing.F) {
	for _, s := range []string{`{}`, `null`, `{"type":"assistant","timestamp":"bad"}`, `{"type":"runtime-config","timestamp":1}`, `{"type":"workspace-directories","sessionId":"S","directories":["/repo"]}`} {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		r, e := ParseRecord(b)
		if e != nil {
			if e.Error() != ErrInvalidRecord.Error() && e.Error() != ErrLimit.Error() {
				t.Fatal("unclassified diagnostic")
			}
			return
		}
		var s SessionState
		_ = s.Step(r)
		_, _ = s.Metadata()
		_, _, _ = s.ConfigAt(time.Now())
	})
}

func TestUsageProjectionRejectsPrivatePayloads(t *testing.T) {
	for _, value := range []string{`"PRIVATE"`, `{"content":"PRIVATE"}`, `["PRIVATE"]`, `true`} {
		_, e := ParseRecord([]byte(`{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26Z","message":{"usage":{"credits":` + value + `}}}`))
		if !errors.Is(e, ErrInvalidRecord) {
			t.Fatal("retained nonnumeric usage payload")
		}
	}
	for _, value := range []string{`null`, `1`, `0.9265415628571427`, `-1`, `1e100`} {
		r := parse(t, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26Z","message":{"usage":{"credits":`+value+`}}}`)
		if string(r.Message.Usage["credits"]) != value {
			t.Fatal("changed numeric evidence")
		}
	}
	for _, value := range []string{`1`, `"PRIVATE"`, `{}`} {
		_, e := ParseRecord([]byte(`{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26Z","message":{"usage":{"billable":` + value + `}}}`))
		if !errors.Is(e, ErrInvalidRecord) {
			t.Fatal("bad billable type")
		}
	}
}

func TestSessionBoundsAndReplay(t *testing.T) {
	var s SessionState
	for j := 0; j < MaxConfigEpochs; j++ {
		if e := s.Step(Record{Type: "runtime-config", SessionID: "S", Timestamp: time.UnixMilli(int64(j + 1)), Model: "auto"}); e != nil {
			t.Fatal(e)
		}
	}
	for j := 0; j < MaxConfigEpochs; j++ {
		if e := s.Step(Record{Type: "runtime-config", SessionID: "S", Timestamp: time.UnixMilli(int64(j + 1)), Model: "auto"}); e != nil {
			t.Fatal("replay consumes epoch limit")
		}
	}
	if e := s.Step(Record{Type: "runtime-config", SessionID: "S", Timestamp: time.UnixMilli(MaxConfigEpochs + 1), Model: "auto"}); !errors.Is(e, ErrLimit) {
		t.Fatal("epoch bound missing")
	}
	if _, _, e := s.ConfigAt(time.Now()); !errors.Is(e, ErrLimit) {
		t.Fatal("partial config set used after limit")
	}
	if _, e := s.Identity(); e != nil {
		t.Fatal("config limit erased usable spend identity")
	}
	if _, e := ParseRecord([]byte(strings.Repeat(" ", MaxRecordBytes+1))); !errors.Is(e, ErrLimit) {
		t.Fatal("record limit missing")
	}
}

func TestGeneratedStatePermutations(t *testing.T) {
	// Exhaust every permutation, plus replay, instead of depending on selected
	// hand-written orderings of bookkeeping and activity.
	records := []Record{
		parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/repo"]}`),
		parse(t, `{"type":"runtime-config","sessionId":"S","timestamp":1000,"model":"old"}`),
		parse(t, `{"type":"runtime-config","sessionId":"S","timestamp":2000,"model":"new"}`),
		parse(t, `{"type":"user","sessionId":"S","timestamp":"2026-10-14T15:43:26Z"}`),
		parse(t, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:44:26Z"}`),
	}
	var permute func(int)
	permute = func(i int) {
		if i < len(records) {
			for j := i; j < len(records); j++ {
				records[i], records[j] = records[j], records[i]
				permute(i + 1)
				records[i], records[j] = records[j], records[i]
			}
			return
		}
		var state SessionState
		for repeat := 0; repeat < 2; repeat++ {
			for _, r := range records {
				if e := state.Step(r); e != nil {
					t.Fatal(e)
				}
			}
		}
		m, e := state.Metadata()
		if e != nil || m.SessionID != "S" || m.CWD != "/repo" || m.StartedAt.Format(time.RFC3339) != "2026-10-14T15:43:26Z" || m.LastActivity.Sub(m.StartedAt) != time.Minute {
			t.Fatal("permutation changed metadata")
		}
		if model, _, e := state.ConfigAt(time.UnixMilli(2000)); e != nil || model != "new" {
			t.Fatal("permutation changed config")
		}
	}
	permute(0)
}

func TestConfigEffortCopyAndNoBookkeepingActivity(t *testing.T) {
	var s SessionState
	effort := "high"
	if e := s.Step(Record{Type: "runtime-config", SessionID: "S", Timestamp: time.UnixMilli(1000), Model: "auto", ReasoningEffort: &effort}); e != nil {
		t.Fatal(e)
	}
	effort = "low"
	if _, v, e := s.ConfigAt(time.UnixMilli(1000)); e != nil || v != "high" {
		t.Fatal("caller mutated stored config")
	}
	if m, e := s.Metadata(); !errors.Is(e, ErrIncomplete) || !m.StartedAt.IsZero() || !m.LastActivity.IsZero() {
		t.Fatal("bookkeeping became activity")
	}
	if _, _, e := s.ConfigAt(time.UnixMilli(999)); !errors.Is(e, ErrIncomplete) {
		t.Fatal("future config applied early")
	}
}

func TestStreamingNullStopReason(t *testing.T) {
	r := parse(t, `{"type":"assistant","sessionId":"S","timestamp":"2026-10-14T15:43:26.567Z","message":{"id":"R","model":"auto","stop_reason":null,"content":[{"type":"text","text":"PRIVATE"}]}}`)
	if r.Message == nil || r.Message.ID != "R" || r.Message.StopReason != "" || r.Message.Usage != nil {
		t.Fatal("streaming row lost")
	}
}

func TestGeneratedConflictsNeverRepair(t *testing.T) {
	good := parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/repo"]}`)
	for _, conflict := range []Record{
		parse(t, `{"type":"workspace-directories","sessionId":"T","directories":["/repo"]}`),
		parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/elsewhere"]}`),
		parse(t, `{"type":"workspace-directories","sessionId":"S","directories":["/repo","/elsewhere"]}`),
	} {
		for pos := 0; pos < 7; pos++ {
			var s SessionState
			for i := 0; i < 7; i++ {
				r := good
				if i == pos {
					r = conflict
				}
				_ = s.Step(r)
			}
			if _, e := s.Metadata(); !errors.Is(e, ErrConflict) {
				t.Fatal("conflict repaired by replay")
			}
			id, e := s.Identity()
			if conflict.SessionID == good.SessionID {
				if e != nil || id != "S" {
					t.Fatal("workspace poisoned spend identity")
				}
			} else if !errors.Is(e, ErrConflict) {
				t.Fatal("session conflict accepted")
			}
		}
	}
	for pos := 0; pos < 5; pos++ {
		var s SessionState
		for i := 0; i < 5; i++ {
			model := "a"
			if i == pos {
				model = "b"
			}
			_ = s.Step(Record{Type: "runtime-config", SessionID: "S", Timestamp: time.UnixMilli(1000), Model: model})
		}
		if _, _, e := s.ConfigAt(time.UnixMilli(1000)); !errors.Is(e, ErrConflict) {
			t.Fatal("config conflict repaired")
		}
	}
	var incomplete SessionState
	_ = incomplete.Step(Record{Type: "workspace-directories"})
	_ = incomplete.Step(good)
	if _, e := incomplete.Identity(); !errors.Is(e, ErrIncomplete) {
		t.Fatal("failed initial identity repaired")
	}
}

func TestTimestampClasses(t *testing.T) {
	for _, raw := range []string{`"1791989630430"`, `null`, `1.5`, `1e3`, `0`, `-1`, `253402300800000`, `true`, `{}`} {
		_, e := ParseRecord([]byte(`{"type":"runtime-config","sessionId":"S","timestamp":` + raw + `}`))
		if !errors.Is(e, ErrInvalidRecord) {
			t.Fatal("invalid config timestamp class")
		}
	}
	for _, raw := range []string{`1791989630430`, `null`, `"not-a-date"`, `"0001-01-01T00:00:00Z"`, `"2026-10-14"`, `true`, `{}`} {
		_, e := ParseRecord([]byte(`{"type":"user","sessionId":"S","timestamp":` + raw + `}`))
		if !errors.Is(e, ErrInvalidRecord) {
			t.Fatal("invalid activity timestamp class")
		}
	}
}
