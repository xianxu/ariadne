package flakes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const evidenceInvocation = "0123456789abcdef0123456789abcdef"

func eventFor(unit, kind string) UnitEvent {
	return UnitEvent{Version: 1, Invocation: evidenceInvocation, Unit: unit, Event: kind}
}
func terminalFor(unit string, exit int) UnitEvent {
	e := eventFor(unit, "finish")
	e.Exit = &exit
	e.Complete = true
	return e
}
func evidenceLog(events ...UnitEvent) string {
	var b strings.Builder
	for _, e := range events {
		data, _ := json.Marshal(e)
		b.Write(data)
		b.WriteByte('\n')
	}
	return b.String()
}

func TestUnitEvidenceCompletionOracle(t *testing.T) {
	for _, kind := range []string{"script:a.sh", "lua:a.lua"} {
		t.Run(kind, func(t *testing.T) {
			start, end := eventFor(kind, "start"), terminalFor(kind, 0)
			c := eventFor(kind, "case")
			c.Case = "suite exact case"
			c.Result = "pass"
			events := []UnitEvent{start}
			required := kind
			if strings.HasPrefix(kind, "lua:") {
				events = append(events, c)
				required += "#" + c.Case
			}
			events = append(events, end)
			for mask := 0; mask < 1<<len(events); mask++ {
				var subset []UnitEvent
				for i, e := range events {
					if mask&(1<<i) != 0 {
						subset = append(subset, e)
					}
				}
				facts := ParseUnitEvidence(evidenceLog(subset...))
				got := len(RequiredUnitProof(facts, []string{required})) == 0
				want := mask == (1<<len(events))-1
				if got != want {
					t.Fatalf("mask %b proof=%v want=%v facts=%+v", mask, got, want, facts)
				}
			}
			facts := ParseUnitEvidence(evidenceLog(events...))
			if !facts.Units[kind].Passed || !facts.Seen || facts.Visits != len(events) {
				t.Fatalf("%+v", facts)
			}
			for i := range events {
				repeated := append([]UnitEvent{}, events...)
				repeated = append(repeated, events[i])
				if got := ParseUnitEvidence(evidenceLog(repeated...)); len(RequiredUnitProof(got, []string{required})) == 0 {
					t.Fatalf("repetition %d granted proof", i)
				}
			}
		})
	}
}

func TestUnitEvidenceFailuresAndInconsistency(t *testing.T) {
	for _, tc := range []struct {
		name         string
		exit, errors int
		complete     bool
		results      []string
		failures     int
		bad          bool
	}{
		{"lua failure", 1, 0, true, []string{"fail", "fail", "pass"}, 2, false},
		{"inner timeout", 124, 0, true, []string{"fail"}, 1, false},
		{"missing summary", 1, 0, false, []string{"fail"}, 0, true},
		{"native error", 1, 1, true, []string{"fail"}, 0, true},
		{"zero with failure", 0, 0, true, []string{"fail"}, 0, true},
		{"nonzero without case", 1, 0, true, nil, 0, true},
		{"nonzero only passes", 1, 0, true, []string{"pass"}, 0, true},
		{"empty success", 0, 0, true, nil, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			events := []UnitEvent{eventFor("lua:a.lua", "start")}
			for i, result := range tc.results {
				e := eventFor("lua:a.lua", "case")
				e.Case = fmt.Sprintf("case %d", i)
				e.Result = result
				events = append(events, e)
			}
			end := terminalFor("lua:a.lua", tc.exit)
			end.Complete = tc.complete
			end.Errors = tc.errors
			events = append(events, end)
			f := ParseUnitEvidence(evidenceLog(events...))
			if len(f.Failures) != tc.failures || (len(f.Unattributed) > 0) != tc.bad {
				t.Fatalf("%+v", f)
			}
			if len(RequiredUnitProof(f, []string{"lua:a.lua#absent"})) != 1 {
				t.Fatal("absent case passed")
			}
		})
	}
	for _, exit := range []int{1, 124, 255} {
		f := ParseUnitEvidence(evidenceLog(eventFor("script:a.sh", "start"), terminalFor("script:a.sh", exit)))
		if !reflect.DeepEqual(f.Failures, []string{"script:a.sh"}) || len(f.Unattributed) != 0 {
			t.Fatalf("exit %d: %+v", exit, f)
		}
	}
}

func TestUnitEvidenceStrictFrames(t *testing.T) {
	good := strings.TrimSpace(evidenceLog(eventFor("script:a.sh", "start")))
	for _, line := range []string{"noise", " " + good, "\t" + good, `{"Action":"output","Output":` + fmt.Sprintf("%q", good) + `}`, `{"future":{"sdlc_unit":1}}`} {
		if _, seen, err := DecodeUnitEvent(line); seen || err != nil {
			t.Fatalf("diagnostic recognized: %s %v", line, err)
		}
	}
	for _, line := range []string{
		strings.Replace(good, `"sdlc_unit":1`, `"sdlc_unit":2`, 1),
		strings.Replace(good, `"sdlc_unit":1`, `"sdlc_unit":null`, 1),
		strings.Replace(good, `"sdlc_unit":1`, `"sdlc_unit":"1"`, 1),
		strings.Replace(good, `"sdlc_unit":1`, `"SDLC_UNIT":1`, 1),
		strings.Replace(good, `"complete":false`, `"Complete":false`, 1),
		strings.Replace(good, `"errors":0`, `"errors":null`, 1),
		strings.Replace(good, `"errors":0`, `"errors":0,"errors":0`, 1),
		strings.Replace(good, `"errors":0`, `"errors":-1`, 1),
		strings.Replace(good, `"errors":0`, `"errors":10001`, 1),
		strings.Replace(good, `"errors":0`, `"errors":1`, 1),
		strings.Replace(good, `"complete":false`, `"complete":true`, 1),
		strings.Replace(good, `"event":"start"`, `"event":"start","exit":0`, 1),
		strings.Replace(good, `"event":"start"`, `"event":"case","case":"x","result":"pass"`, 1),
		strings.Replace(good, `"event":"start"`, `"event":"finish"`, 1),
		`{"sdlc_unit":`, `{"sdlc_unit":1,`, good + ` {}`,
	} {
		if _, seen, err := DecodeUnitEvent(line); !seen || err == nil {
			t.Fatalf("accepted %s (%v,%v)", line, seen, err)
		}
	}
	additive := strings.Replace(good, `"errors":0`, `"errors":0,"future":{"big":1e400}`, 1)
	if _, seen, err := DecodeUnitEvent(additive); !seen || err != nil {
		t.Fatalf("future rejected: %v", err)
	}
	f := ParseUnitEvidence(evidenceLog(eventFor("script:a.sh", "start"), terminalFor("script:a.sh", 0)) + `{"sdlc_unit":`)
	if len(RequiredUnitProof(f, []string{"script:a.sh"})) == 0 {
		t.Fatal("malformed control did not revoke proof")
	}
}

func TestUnitEvidenceStickyAndIndependent(t *testing.T) {
	start, end := eventFor("lua:a.lua", "start"), terminalFor("lua:a.lua", 0)
	c := eventFor("lua:a.lua", "case")
	c.Case = "a"
	c.Result = "pass"
	for _, events := range [][]UnitEvent{{c, start, end}, {end, start, c}, {start, c, end, c}, {start, start, c, end}, {start, c, c, end}} {
		f := ParseUnitEvidence(evidenceLog(events...))
		if len(f.Unattributed) == 0 || len(RequiredUnitProof(f, []string{"lua:a.lua#a"})) == 0 {
			t.Fatalf("unsafe proof %+v", f)
		}
	}
	wrong := c
	wrong.Invocation = strings.Repeat("f", 32)
	f := ParseUnitEvidence(evidenceLog(start, wrong, c, end))
	if len(f.Unattributed) == 0 {
		t.Fatal("invocation mismatch accepted")
	}
	model := NewUnitEvidence()
	for _, e := range []UnitEvent{start, c, end} {
		if err := model.Apply(e); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := model.Finish()
	snapshot.Units["lua:a.lua"].Cases["a"] = "fail"
	delete(snapshot.Units, "lua:a.lua")
	if len(RequiredUnitProof(model.Finish(), []string{"lua:a.lua#a"})) != 0 {
		t.Fatal("snapshot mutated private state")
	}
	invalid := eventFor("lua:b.lua", "case")
	_ = model.Apply(invalid)
	if len(RequiredUnitProof(model.Finish(), []string{"lua:a.lua#a"})) != 1 {
		t.Fatal("unattributed sibling did not block")
	}
}

func TestDecodeUnitSummaryStrict(t *testing.T) {
	good := `{"version":1,"unit":"lua:a.lua","complete":true,"cases":[{"name":"suite a","result":"pass"},{"name":"suite b","result":"fail"}],"errors":0}`
	cases, n, err := DecodeUnitSummary([]byte(good), "lua:a.lua")
	if err != nil || n != 0 || len(cases) != 2 || cases[1].Result != "fail" {
		t.Fatalf("%+v %d %v", cases, n, err)
	}
	for _, bad := range []string{strings.Replace(good, `"version":1`, `"version":null`, 1), strings.Replace(good, `"version":1`, `"Version":1`, 1), strings.Replace(good, `"version":1`, `"version":1,"version":1`, 1), strings.Replace(good, `"complete":true`, `"complete":false`, 1), strings.Replace(good, `"errors":0`, `"errors":null`, 1), strings.Replace(good, `"errors":0`, `"errors":-1`, 1), strings.Replace(good, `"cases":[`, `"cases":null,"ignore":[`, 1), strings.Replace(good, `"name":"suite a"`, `"name":null`, 1), strings.Replace(good, `"name":"suite a"`, `"Name":"suite a"`, 1), strings.Replace(good, `"suite b"`, `"suite a"`, 1), strings.Replace(good, `"pass"`, `"skip"`, 1), strings.Replace(good, `"suite a"`, `"bad\nlabel"`, 1), strings.Replace(good, `"unit":"lua:a.lua"`, `"unit":"lua:b.lua"`, 1)} {
		if _, _, err := DecodeUnitSummary([]byte(bad), "lua:a.lua"); err == nil {
			t.Fatalf("accepted %s", bad)
		}
	}
	for _, good := range []string{strings.Replace(good, `"errors":0`, `"errors":2`, 1), strings.Replace(good, `"errors":0`, `"errors":0,"future":{"big":1e400}`, 1), `{"version":1,"unit":"lua:a.lua","complete":true,"cases":[],"errors":0}`} {
		if _, _, err := DecodeUnitSummary([]byte(good), "lua:a.lua"); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUnitEvidenceBoundedWork(t *testing.T) {
	for _, n := range []int{8, 16, 10000} {
		events := make([]UnitEvent, 0, n*2)
		tests := make([]string, 0, n)
		for i := 0; i < n; i++ {
			unit := fmt.Sprintf("script:%d.sh", i)
			events = append(events, eventFor(unit, "start"), terminalFor(unit, 0))
			tests = append(tests, unit)
		}
		f := ParseUnitEvidence(evidenceLog(events...))
		if f.Visits != 2*n || len(RequiredUnitProof(f, tests)) != 0 {
			t.Fatalf("n=%d visits=%d units=%d bad=%v", n, f.Visits, len(f.Units), f.Unattributed)
		}
	}
	f := ParseUnitEvidence(strings.Repeat("x", (1<<20)+1))
	if len(f.Unattributed) == 0 {
		t.Fatal("oversized line accepted")
	}
	f = ParseUnitEvidence(strings.Repeat("x\n", (64<<20)/2+1))
	if len(f.Unattributed) == 0 {
		t.Fatal("oversized channel accepted")
	}
	events := []UnitEvent{eventFor("lua:a.lua", "start")}
	for i := 0; i < 10001; i++ {
		e := eventFor("lua:a.lua", "case")
		e.Case = fmt.Sprint(i)
		e.Result = "pass"
		events = append(events, e)
	}
	events = append(events, terminalFor("lua:a.lua", 0))
	f = ParseUnitEvidence(evidenceLog(events...))
	if len(f.Unattributed) == 0 {
		t.Fatal("too many cases accepted")
	}
}

func FuzzDecodeUnitEvidence(f *testing.F) {
	f.Add(`{"sdlc_unit":1}`)
	f.Add(`{"Action":"output","Output":"{\"sdlc_unit\":1}"}`)
	f.Fuzz(func(t *testing.T, s string) {
		_, recognized, err := DecodeUnitEvent(s)
		facts := ParseUnitEvidence(s)
		if recognized && err != nil && len(facts.Unattributed) == 0 {
			t.Fatal("malformed recognized event lost")
		}
		if len(RequiredUnitProof(facts, []string{"script:absent.sh", "lua:absent.lua#absent"})) != 2 {
			t.Fatal("invented proof")
		}
	})
}

func TestUnitEvidenceTruncatedReservedKeyFailsClosed(t *testing.T) {
	for _, s := range []string{`{"sdlc_unit"`, `{"sdlc_unit"}`, `{"future":false,"sdlc_unit"`} {
		if _, recognized, err := DecodeUnitEvent(s); !recognized || err == nil {
			t.Fatalf("lost truncated frame %q: recognized=%v err=%v", s, recognized, err)
		}
	}
	for _, s := range []string{`{"Output":"sdlc_unit"}`, `{"Output":["sdlc_unit"]}`} {
		if _, recognized, _ := DecodeUnitEvent(s); recognized {
			t.Fatalf("diagnostic value recognized: %s", s)
		}
	}
}

func TestUnitEvidenceInvalidFirstFrameSticky(t *testing.T) {
	model := NewUnitEvidence()
	bad := eventFor("script:a.sh", "start")
	bad.Version = 2
	if err := model.Apply(bad); err == nil {
		t.Fatal("invalid version accepted")
	}
	_ = model.Apply(eventFor("script:a.sh", "start"))
	_ = model.Apply(terminalFor("script:a.sh", 0))
	if model.Finish().Units["script:a.sh"].Passed {
		t.Fatal("invalid first frame recovered into pass")
	}
}

func TestUnitEvidenceExactCaseAndCountedCases(t *testing.T) {
	for _, n := range []int{8, 16} {
		events := []UnitEvent{eventFor("lua:a.lua", "start")}
		var tests []string
		for i := 0; i < n; i++ {
			e := eventFor("lua:a.lua", "case")
			e.Case = fmt.Sprintf("case %d", i)
			e.Result = "pass"
			events = append(events, e)
			tests = append(tests, e.Unit+"#"+e.Case)
		}
		events = append(events, terminalFor("lua:a.lua", 0))
		facts := ParseUnitEvidence(evidenceLog(events...))
		if facts.Visits != n+2 || len(RequiredUnitProof(facts, tests)) != 0 {
			t.Fatalf("%+v", facts)
		}
		missing := RequiredUnitProof(facts, []string{"lua:a.lua#case", "lua:a.lua#case 0 child", "lua:a.lua#CASE 0", "pkg.TestGo"})
		if len(missing) != 3 {
			t.Fatalf("broadened exact identities: %v", missing)
		}
	}
	c := eventFor("lua:a.lua", "case")
	c.Case = "pending"
	c.Result = "skip"
	facts := ParseUnitEvidence(evidenceLog(eventFor(c.Unit, "start"), c, terminalFor(c.Unit, 0)))
	if len(RequiredUnitProof(facts, []string{c.Unit + "#pending"})) != 1 {
		t.Fatal("skip counted as pass")
	}
}

func FuzzDecodeUnitSummary(f *testing.F) {
	f.Add([]byte(`{"version":1,"unit":"lua:a.lua","complete":true,"cases":[],"errors":0}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		cases, errors, err := DecodeUnitSummary(data, "lua:a.lua")
		if err != nil {
			return
		}
		if len(cases) > 10000 || errors < 0 || errors > 10000 {
			t.Fatal("unbounded summary")
		}
		seen := map[string]bool{}
		for _, c := range cases {
			if _, err := ParseIdentity("lua:a.lua#" + c.Name); err != nil || seen[c.Name] || (c.Result != "pass" && c.Result != "fail") {
				t.Fatal("invalid accepted case")
			}
			seen[c.Name] = true
		}
	})
}

func TestUnitEvidencePermutationOracle(t *testing.T) {
	start := eventFor("lua:a.lua", "start")
	c := eventFor(start.Unit, "case")
	c.Case = "exact"
	c.Result = "pass"
	end := terminalFor(start.Unit, 0)
	events := []UnitEvent{start, c, end}
	for _, order := range [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		facts := ParseUnitEvidence(evidenceLog(events[order[0]], events[order[1]], events[order[2]]))
		got := len(RequiredUnitProof(facts, []string{start.Unit + "#exact"})) == 0
		if got != (order == [3]int{0, 1, 2}) {
			t.Fatalf("order %v proof=%v", order, got)
		}
	}
	for _, events := range [][]UnitEvent{{eventFor("script:a.sh", "start"), terminalFor("script:a.sh", 1), terminalFor("script:a.sh", 0)}, {eventFor("script:a.sh", "start"), terminalFor("script:a.sh", 0), terminalFor("script:a.sh", 1)}} {
		facts := ParseUnitEvidence(evidenceLog(events...))
		if facts.Units["script:a.sh"].Passed || len(facts.Unattributed) == 0 {
			t.Fatal("contradictory terminal recovered pass")
		}
	}
}

func TestUnitSummaryBounds(t *testing.T) {
	summary := func(cases string) []byte {
		return []byte(`{"version":1,"unit":"lua:a.lua","complete":true,"errors":0,"cases":` + cases + `}`)
	}
	cases := make([]CaseResult, 10001)
	for i := range cases {
		cases[i] = CaseResult{Name: fmt.Sprint(i), Result: "pass"}
	}
	for _, n := range []int{10000, 10001} {
		data, _ := json.Marshal(cases[:n])
		got, _, err := DecodeUnitSummary(summary(string(data)), "lua:a.lua")
		if (err == nil) != (n == 10000) || (n == 10000 && len(got) != n) {
			t.Fatalf("n=%d count=%d err=%v", n, len(got), err)
		}
	}
	for _, label := range []string{strings.Repeat("x", 2048), strings.Repeat("x", 2049)} {
		data, _ := json.Marshal([]CaseResult{{Name: label, Result: "pass"}})
		_, _, err := DecodeUnitSummary(summary(string(data)), "lua:a.lua")
		if (err == nil) != (len(label) == 2048) {
			t.Fatalf("label %d err=%v", len(label), err)
		}
	}
	if _, _, err := DecodeUnitSummary(bytesForOversizedSummary(), "lua:a.lua"); err == nil {
		t.Fatal("oversized summary accepted")
	}
}

func bytesForOversizedSummary() []byte { return []byte(strings.Repeat(" ", (16<<20)+1)) }

func TestUnitEvidenceOuterLogBudgetIndependentOfSummary(t *testing.T) {
	// Supported mixed outer logs can exceed a private Lua producer's 16 MiB.
	diagnostics := strings.Repeat("ordinary Go diagnostics\n", (17<<20)/24+1)
	log := diagnostics + evidenceLog(eventFor("script:a.sh", "start"), terminalFor("script:a.sh", 0))
	facts := ParseUnitEvidence(log)
	if len(RequiredUnitProof(facts, []string{"script:a.sh"})) != 0 {
		t.Fatalf("supported outer log rejected: %v", facts.Unattributed)
	}
}
