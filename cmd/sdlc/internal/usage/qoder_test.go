package usage

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const qoderHeader = `{"type":"workspace-directories","sessionId":"session-q","directories":["/repo"]}
{"type":"runtime-config","sessionId":"session-q","timestamp":1791968400000,"model":"auto"}`

func qoderRow(id, stop, usage string) string {
	return `{"type":"assistant","sessionId":"session-q","timestamp":"2026-10-14T10:00:00Z","uuid":"event-` + id + `","message":{"id":"` + id + `","stop_reason":` + stop + `,"usage":` + usage + `}}`
}
func TestQoderTerminalAndStreamingCredits(t *testing.T) {
	raw := qoderHeader + "\n" + qoderRow("m", "null", "null") + "\n" + qoderRow("m", `"end_turn"`, `{"credits":0.123456789012345678,"original_credits":2,"billable":true,"input_tokens":0,"output_tokens":0}`)
	rr, c, e := Decode(context.Background(), "qoder", "filename", strings.NewReader(raw))
	if e != nil || len(rr) != 1 || rr[0].ID != "m" || rr[0].SourceID != "session-q" || rr[0].Model != "auto" || rr[0].Tokens != nil || rr[0].Credits == nil || rr[0].Credits.String() != "0.123456789012345678" || !c.SpendComplete() {
		t.Fatalf("rows %+v coverage %+v error %v", rr, c, e)
	}
}
func TestQoderChargeAndUnfinishedGaps(t *testing.T) {
	for _, u := range []string{`{"credits":1,"billable":false}`, `{"credits":1}`, `{"credits":null}`, `{"credits":-1}`} {
		rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(qoderHeader+"\n"+qoderRow("m", `"tool_use"`, u)))
		if e != nil || c.SpendComplete() {
			t.Fatalf("accepted charge ambiguity %s: %+v %+v %v", u, rr, c, e)
		}
	}
	rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(qoderHeader+"\n"+qoderRow("m", "null", "null")))
	if e != nil || len(rr) != 1 || !rr[0].Provisional || rr[0].Credits != nil {
		t.Fatalf("unfinished %+v %+v %v", rr, c, e)
	}
}
func TestQoderCopiedTerminalPermutation(t *testing.T) {
	a := qoderRow("m", `"end_turn"`, `{"credits":1,"billable":true}`)
	b := qoderRow("n", `"tool_use"`, `{"credits":2,"billable":true}`)
	for _, rows := range []string{a + "\n" + b + "\n" + a, b + "\n" + a + "\n" + b} {
		rr, c, e := Decode(context.Background(), "qoder", "copy", strings.NewReader(qoderHeader+"\n"+rows))
		if e != nil {
			t.Fatal(e)
		}
		rr, n, e := NormalizeResponses(rr)
		if e != nil || len(rr) != 2 || !c.SpendComplete() || !n.SpendComplete() {
			t.Fatalf("%+v %+v %+v %v", rr, c, n, e)
		}
	}
}

func TestQoderGlobalOffWindowConflictAndUnknownWorkspace(t *testing.T) {
	raw := `{"type":"runtime-config","sessionId":"session-q","timestamp":1791968400000,"model":"auto"}` + "\n" + qoderRow("m", `"end_turn"`, `{"credits":1,"billable":true}`)
	rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
	if e != nil || len(rr) != 1 || rr[0].SourceID != "session-q" || !c.SpendComplete() || c.Complete() {
		t.Fatalf("unknown workspace lost spend %+v %+v %v", rr, c, e)
	}
	w := Window{stamp("2026-10-14T10:00:00Z"), stamp("2026-10-14T11:00:00Z")}
	conflict := rr[0]
	conflict.At = w.Since.Add(-1)
	conflict.Credits = nativeCredit(t, "2")
	for _, rows := range [][]Response{{rr[0], conflict}, {conflict, rr[0]}} {
		report, e := AggregateUsage(rows, w, w, nil)
		if e != nil || report.Credits != nil || report.Coverage.Counts["conflicts"] != 1 {
			t.Fatalf("off-window conflict %+v %v", report, e)
		}
	}
}
func TestQoderBoundsCancellationAndTornTail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, e := Decode(ctx, "qoder", "f", strings.NewReader(qoderHeader)); e == nil {
		t.Fatal("accepted cancellation")
	}
	_, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(qoderHeader+"\n"+`{"type":"assistant"`))
	if e != nil || c.Counts["truncated_tails"] != 1 || c.SpendComplete() {
		t.Fatalf("tail %+v %v", c, e)
	}
	_, c, e = Decode(context.Background(), "qoder", "f", strings.NewReader(strings.Repeat("x", MaxLineBytes+1)))
	if e == nil || c.Counts["unavailable_sources"] != 1 {
		t.Fatalf("bounds %+v %v", c, e)
	}
}
func TestQoderResponseModelAndConflictingSession(t *testing.T) {
	raw := qoderHeader + "\n" + strings.Replace(qoderRow("m", `"end_turn"`, `{"credits":0,"billable":true}`), `"id":"m"`, `"id":"m","model":"reported-alias"`, 1)
	rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
	if e != nil || len(rr) != 1 || rr[0].Model != "auto" || rr[0].Credits == nil || !rr[0].Credits.IsZero() || !c.SpendComplete() {
		t.Fatalf("%+v %+v %v", rr, c, e)
	}
	raw += "\n" + `{"type":"user","sessionId":"other","timestamp":"2026-10-14T10:00:00Z"}`
	rr, c, e = Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
	if e != nil || len(rr) != 0 || c.SpendComplete() {
		t.Fatalf("conflicting session %+v %+v %v", rr, c, e)
	}
}

func FuzzQoderStreamingTerminalPermutation(f *testing.F) {
	f.Add(uint16(5), uint8(3))
	f.Fuzz(func(t *testing.T, amount uint16, copies uint8) {
		terminal := qoderRow("m", `"end_turn"`, fmt.Sprintf(`{"credits":%d,"billable":true}`, amount))
		stream := qoderRow("m", "null", "null")
		n := int(copies%8) + 1
		for _, body := range []string{stream + "\n" + strings.Repeat(terminal+"\n", n), strings.Repeat(terminal+"\n", n) + stream + "\n"} {
			rows, c, e := Decode(context.Background(), "qoder", "copy", strings.NewReader(qoderHeader+"\n"+body))
			if e != nil {
				t.Fatal(e)
			}
			rows, nc, e := NormalizeResponses(rows)
			if e != nil || len(rows) != 1 || rows[0].Provisional || rows[0].Tokens != nil || rows[0].Credits.String() != fmt.Sprint(amount) || !c.SpendComplete() || !nc.SpendComplete() {
				t.Fatalf("%+v %+v %+v %v", rows, c, nc, e)
			}
		}
	})
}

func TestQoderAmbiguousAttributionPreservesCreditDenominator(t *testing.T) {
	for _, metadata := range []string{
		`{"type":"workspace-directories","sessionId":"session-q","directories":["/repo","/other"]}`,
		`{"type":"workspace-directories","sessionId":"session-q","directories":["/other"]}`,
		`{"type":"runtime-config","sessionId":"session-q","timestamp":1791968400000,"model":"different"}`,
	} {
		raw := qoderHeader + "\n" + metadata + "\n" + qoderRow("m", `"end_turn"`, `{"credits":1.234567890123456789,"billable":true}`)
		rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
		if e != nil || len(rr) != 1 || rr[0].SourceID != "session-q" || rr[0].Credits == nil || rr[0].Credits.String() != "1.234567890123456789" || !c.SpendComplete() || c.Complete() || c.Counts["unknown_metadata"] == 0 {
			t.Fatalf("ambiguous attribution lost complete known spend: %+v %+v %v", rr, c, e)
		}
	}
}

func TestQoderResponseUsesApplicableRuntimeConfig(t *testing.T) {
	config := func(at, model string) string {
		return fmt.Sprintf(`{"type":"runtime-config","sessionId":"session-q","timestamp":%d,"model":%q}`, stamp(at).UnixMilli(), model)
	}
	row := func(id, at string) string {
		return strings.Replace(strings.Replace(qoderRow(id, `"end_turn"`, `{"credits":1,"billable":true}`), `"id":"`+id+`"`, `"id":"`+id+`","model":"auto"`, 1), "2026-10-14T10:00:00Z", at, 1)
	}
	old := config("2026-10-14T09:00:00Z", "auto")
	raw := old + "\n" + row("first", "2026-10-14T10:00:00Z") + "\n" + config("2026-10-14T10:30:00Z", "fast") + "\n" + row("second", "2026-10-14T11:00:00Z") + "\n" + old
	rr, c, e := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
	if e != nil || len(rr) != 2 || rr[0].Model != "auto" || rr[1].Model != "fast" || !c.SpendComplete() {
		t.Fatalf("config epochs %+v %+v %v", rr, c, e)
	}
	for _, prefix := range []string{"", old + "\n" + config("2026-10-14T09:00:00Z", "other") + "\n"} {
		rr, c, e = Decode(context.Background(), "qoder", "f", strings.NewReader(prefix+row("m", "2026-10-14T10:00:00Z")))
		if e != nil || len(rr) != 1 || rr[0].Model != "" || rr[0].Credits == nil || !c.SpendComplete() || c.Complete() || c.Counts["unknown_model"] == 0 || c.Counts["unknown_metadata"] == 0 {
			t.Fatalf("unknown config %+v %+v %v", rr, c, e)
		}
	}
}

func TestQoderUnsupportedTokenPlaceholdersKeepKnownCredits(t *testing.T) {
	fields := []string{"input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"}
	for _, field := range fields {
		for _, value := range []string{"1", "-1", "0.5", "null", "1e999999"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				raw := qoderHeader + "\n" + qoderRow("m", `"end_turn"`, fmt.Sprintf(`{"credits":1.25,"billable":true,%q:%s}`, field, value))
				rows, c, err := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
				if err != nil {
					t.Fatal(err)
				}
				rows, n, err := NormalizeResponses(rows)
				c.Merge(n)
				if err != nil || len(rows) != 1 || rows[0].Tokens != nil || rows[0].Credits == nil || rows[0].Credits.String() != "1.25" || c.Counts["unsupported_formats"] != 1 || c.SpendComplete() {
					t.Fatalf("future token format silently accepted: rows=%+v coverage=%+v err=%v", rows, c, err)
				}
			})
		}
	}
}
func TestQoderZeroTokenPlaceholdersAndAbsentRemainSupported(t *testing.T) {
	for _, value := range []string{"0", "0.0", "0e0"} {
		raw := qoderHeader + "\n" + qoderRow("m", `"end_turn"`, fmt.Sprintf(`{"credits":1,"billable":true,"input_tokens":%s,"output_tokens":%s,"cache_read_input_tokens":%s,"cache_creation_input_tokens":%s}`, value, value, value, value))
		rows, c, err := Decode(context.Background(), "qoder", "f", strings.NewReader(raw))
		if err != nil {
			t.Fatal(err)
		}
		rows, n, err := NormalizeResponses(rows)
		c.Merge(n)
		if err != nil || len(rows) != 1 || rows[0].Tokens != nil || rows[0].Credits == nil || !c.SpendComplete() {
			t.Fatalf("zero %s: rows=%+v coverage=%+v err=%v", value, rows, c, err)
		}
	}
}
func TestQoderInvalidTypedTokenPlaceholdersDoNotHideOtherCredits(t *testing.T) {
	for _, field := range []string{"input_tokens", "output_tokens", "cache_read_input_tokens", "cache_creation_input_tokens"} {
		for _, value := range []string{`"1"`, `true`, `{}`, `[]`} {
			t.Run(field+"/"+value, func(t *testing.T) {
				bad := qoderRow("bad", `"end_turn"`, fmt.Sprintf(`{"credits":9,"billable":true,%q:%s}`, field, value))
				good := qoderRow("good", `"end_turn"`, `{"credits":1.25,"billable":true}`)
				rows, c, err := Decode(context.Background(), "qoder", "f", strings.NewReader(qoderHeader+"\n"+bad+"\n"+good))
				if err != nil {
					t.Fatal(err)
				}
				rows, n, err := NormalizeResponses(rows)
				c.Merge(n)
				if err != nil || len(rows) != 1 || rows[0].ID != "good" || rows[0].Tokens != nil || rows[0].Credits == nil || rows[0].Credits.String() != "1.25" || c.Counts["malformed_records"] != 1 || c.SpendComplete() {
					t.Fatalf("typed invalid row: rows=%+v coverage=%+v err=%v", rows, c, err)
				}
			})
		}
	}
}
func TestQoderStreamingTokenFormatDriftRemainsVisible(t *testing.T) {
	stream := qoderRow("m", "null", `{"output_tokens":5}`)
	terminal := qoderRow("m", `"end_turn"`, `{"credits":1.25,"billable":true}`)
	rows, c, err := Decode(context.Background(), "qoder", "f", strings.NewReader(qoderHeader+"\n"+stream+"\n"+terminal))
	if err != nil || len(rows) != 1 || rows[0].Tokens != nil || rows[0].Credits == nil || c.Counts["unsupported_formats"] != 1 || c.SpendComplete() {
		t.Fatalf("stream drift hidden: rows=%+v coverage=%+v err=%v", rows, c, err)
	}
}
