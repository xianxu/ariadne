package bench

import (
	"encoding/json"
	"strings"
	"testing"
)

func resultPtr[T any](v T) *T { return &v }
func resultMeasurements() Measurements {
	return Measurements{GrossNanos: 100, Autonomous: true, Worker: RoleUsage{Coverage: "unknown"}, TL: RoleUsage{Coverage: "unknown"}, Reviewer: RoleUsage{Coverage: "unknown"}, Leaks: LeakReport{Status: "unknown", Coverage: LeakCoverage{Basis: "missing transcript"}}}
}
func resultRun(t *testing.T, scored bool, unavailable ...bool) Run {
	t.Helper()
	task := validTask()
	h := RuntimeIdentity{Slot: "hub:0", Session: "hs", Generation: "1", Root: "/run/hub", Agent: task.TL.Agent, Model: task.TL.Model, Effort: task.TL.Effort}
	w := RuntimeIdentity{Slot: "work:0", Session: "ws", Generation: "1", Root: "/run/work", Agent: task.Worker.Agent, Model: task.Worker.Model, Effort: task.Worker.Effort}
	r, _ := NewRun("result-1")
	i := 0
	step := func(kind EventKind, op string, f WouldLandFacts) {
		i++
		next, _, err := Step(r, Event{ID: strings.Repeat("e", i), RunID: r.ID(), Kind: kind, OperationID: op, Hub: h, Worker: w, Facts: f})
		if err != nil {
			t.Fatal(err)
		}
		r = next
	}
	step(BeginPrepare, "prepare", WouldLandFacts{})
	step(Prepared, "prepare", WouldLandFacts{})
	step(BeginDispatch, "dispatch", WouldLandFacts{})
	step(DispatchConfirmed, "dispatch", WouldLandFacts{})
	if scored {
		f := admittedFacts()
		f.Candidate = strings.Repeat("f", 40)
		f.ObservedHead = f.Candidate
		f.PushedHead = f.Candidate
		f.SuiteHead = f.Candidate
		f.CompletionHead = f.Candidate
		f.ReviewedHead = f.Candidate
		step(ProposeWouldLand, "proof", WouldLandFacts{})
		step(ProofAccepted, "proof", f)
	}
	step(BeginCancel, "cancel", WouldLandFacts{})
	step(CancelConfirmed, "cancel", WouldLandFacts{})
	if scored {
		step(BeginScore, "score", WouldLandFacts{})
		kind := ScoreCompleted
		if len(unavailable) > 0 && unavailable[0] {
			kind = ScoreUnavailable
		}
		step(kind, "score", WouldLandFacts{})
	}
	return r
}
func resultAcceptance() Acceptance {
	return Acceptance{Status: "command", CommandPassed: resultPtr(true), ParentRed: true, LandingGreen: true, PatchSHA256: strings.Repeat("d", 64)}
}

func TestBenchResultBuildsProvenRowAndUnknownIsNotZero(t *testing.T) {
	row, err := BuildRow(validTask(), resultRun(t, true), resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), true)
	if err != nil {
		t.Fatal(err)
	}
	if row.RunID != "result-1" || row.Outcome != "would-land" || row.Candidate != strings.Repeat("f", 40) || !row.Synthetic || row.OriginalHours != 0.14 || row.Worker.Model != "gpt-6.1-sol" {
		t.Fatalf("wrong identity/outcome: %+v", row)
	}
	if row.Measurements.Worker.InputTokens != nil || row.Measurements.ActiveHours != nil || row.Measurements.GateRefusals != nil {
		t.Fatal("unknown measurement fabricated as zero")
	}
	if row.Acceptance.Total != nil {
		t.Fatal("exit result fabricated case counts")
	}
	raw, err := EncodeLedgerRow(row)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"input_tokens":null`) {
		t.Fatalf("unknown coverage missing: %s", raw)
	}
}

func TestBenchResultRejectsFalseAuthority(t *testing.T) {
	r, _ := NewRun("new")
	if _, err := BuildRow(validTask(), r, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("nonterminal row accepted")
	}
	cases := map[string]func(*Measurements, *Acceptance){"missing parent red": func(_ *Measurements, a *Acceptance) { a.ParentRed = false }, "missing landing green": func(_ *Measurements, a *Acceptance) { a.LandingGreen = false }, "wrong patch": func(_ *Measurements, a *Acceptance) { a.PatchSHA256 = strings.Repeat("c", 64) }, "command has cases": func(_ *Measurements, a *Acceptance) { a.Total = resultPtr(1) }, "inconsistent cases": func(_ *Measurements, a *Acceptance) {
		a.Status = "cases"
		a.Total = resultPtr(2)
		a.Passed = resultPtr(1)
		a.Failed = resultPtr(0)
	}, "zero cases": func(_ *Measurements, a *Acceptance) {
		a.Status = "cases"
		a.Total = resultPtr(0)
		a.Passed = resultPtr(0)
		a.Failed = resultPtr(0)
	}, "negative usage": func(m *Measurements, _ *Acceptance) { m.Worker.InputTokens = resultPtr(int64(-1)) }, "bad coverage": func(m *Measurements, _ *Acceptance) { m.Worker.Coverage = "complete" }, "wait exceeds gross": func(m *Measurements, _ *Acceptance) { m.OperatorWaitNanos = 101 }, "autonomous wait": func(m *Measurements, _ *Acceptance) { m.OperatorWaitNanos = 1 }, "duplicate decision": func(m *Measurements, _ *Acceptance) {
		m.Decisions = []TLDecision{{ID: "d", Kind: "approve", Reason: "ok"}, {ID: "d", Kind: "stop", Reason: "stop"}}
	}, "credits": func(m *Measurements, _ *Acceptance) { m.Worker.Credits = resultPtr(json.Number("-0.1")) }}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			m, a := resultMeasurements(), resultAcceptance()
			mutate(&m, &a)
			if _, err := BuildRow(validTask(), resultRun(t, true), m, a, strings.Repeat("e", 64), false); err == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
	stopped := resultRun(t, false)
	if _, err := BuildRow(validTask(), stopped, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("stopped run scored without candidate")
	}
	if _, err := BuildRow(validTask(), stopped, resultMeasurements(), Acceptance{Status: "unavailable", Reason: "run stopped"}, strings.Repeat("e", 64), false); err != nil {
		t.Fatal(err)
	}
}

func TestBenchResultNestedUnknownRoundTripAndStrictJSON(t *testing.T) {
	m := resultMeasurements()
	m.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"huge":1e400}`)}
	m.Worker.Unknown = map[string]json.RawMessage{"future_role": json.RawMessage(`{"a":true}`)}
	m.Worker.Coverage = "partial"
	m.Worker.SourceIDs = []string{"transcript"}
	m.Worker.Credits = resultPtr(json.Number("0.0000000000000000001234"))
	m.Leaks.Unknown = map[string]json.RawMessage{"future_leak": json.RawMessage(`{"x":1}`)}
	m.Leaks.Coverage.UnknownFields = map[string]json.RawMessage{"future_coverage": json.RawMessage(`{"y":2}`)}
	row, err := BuildRow(validTask(), resultRun(t, true), m, resultAcceptance(), strings.Repeat("e", 64), true)
	if err != nil {
		t.Fatal(err)
	}
	row.Unknown = map[string]json.RawMessage{"future_row": json.RawMessage(`{"z":3}`)}
	raw, err := EncodeLedgerRow(row)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeLedgerRow(raw)
	if err != nil {
		t.Fatal(err)
	}
	again, err := EncodeLedgerRow(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(again) {
		t.Fatalf("unknown fields lost:\n%s\n%s", raw, again)
	}
	bad := []string{strings.Replace(string(raw), `"version":1`, `"version":1,"version":1`, 1), strings.Replace(string(raw), `"gross_nanos":100`, `"gross_nanos":"100"`, 1), strings.Replace(string(raw), `"run_id":"result-1"`, `"RUN_ID":"result-1"`, 1), strings.Replace(string(raw), `"version":1,`, "", 1), string(raw) + `{}`, "null"}
	for _, b := range bad {
		if _, err := DecodeLedgerRow([]byte(b)); err == nil {
			t.Fatalf("strict decoder accepted %s", b)
		}
	}
	got.Unknown["large"] = json.RawMessage(`"` + strings.Repeat("x", LedgerRowLimit) + `"`)
	if _, err := EncodeLedgerRow(got); err == nil {
		t.Fatal("oversized row accepted")
	}
}

func TestBenchResultCasesAndFrozenCallerEvidence(t *testing.T) {
	a := resultAcceptance()
	a.Status = "cases"
	a.Passed = resultPtr(2)
	a.Failed = resultPtr(1)
	a.Total = resultPtr(3)
	a.CommandPassed = resultPtr(false)
	m := resultMeasurements()
	m.GateRefusals = resultPtr(1)
	m.GateRefusalIDs = []string{"gate-1"}
	m.Decisions = []TLDecision{{ID: "decision-1", Kind: "approve", Reason: "bounded fix", Session: "hs", Generation: "1"}}
	m.Leaks.Status = "possibly-leaked"
	m.Leaks.Coverage.Events = 1
	m.Leaks.Coverage.Supported = 1
	m.Leaks.Findings = []LeakFinding{{EventIndex: 0, Kind: "protected-read", Evidence: "source", Unknown: map[string]json.RawMessage{"future_finding": json.RawMessage(`{"x":2}`)}}}
	row, err := BuildRow(validTask(), resultRun(t, true), m, a, strings.Repeat("e", 64), false)
	if err != nil {
		t.Fatal(err)
	}
	*a.Passed = 99
	m.Decisions[0].Reason = "changed"
	m.GateRefusalIDs[0] = "changed"
	m.Leaks.Findings[0].Unknown["future_finding"] = json.RawMessage(`false`)
	if *row.Acceptance.Passed != 2 || row.Measurements.Decisions[0].Reason != "bounded fix" || row.Measurements.GateRefusalIDs[0] != "gate-1" || string(row.Measurements.Leaks.Findings[0].Unknown["future_finding"]) != `{"x":2}` {
		t.Fatal("derived evidence aliases mutable caller state")
	}
	raw, err := EncodeLedgerRow(row)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeLedgerRow(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Measurements.Leaks.Findings[0].Unknown["future_finding"] == nil {
		t.Fatal("nested leak finding extension lost")
	}
}

func TestBenchResultPhaseProfilesAndNullableSchema(t *testing.T) {
	r := resultRun(t, true)
	task := validTask()
	task.Worker.Model = "another-model"
	if _, err := BuildRow(task, r, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("changed worker profile accepted")
	}
	// A same-head gate proof without confirmed runtime shutdown cannot authorize
	// recording, even when the caller already has plausible acceptance data.
	unquiet := r
	unquiet.quiet = false
	if _, err := BuildRow(validTask(), unquiet, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("unconfirmed runtime accepted")
	}
	if _, err := BuildRow(validTask(), r, resultMeasurements(), Acceptance{Status: "unavailable", Reason: "tool missing"}, strings.Repeat("e", 64), false); err == nil {
		t.Fatal("scored phase became unavailable")
	}
	row, err := BuildRow(validTask(), r, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := EncodeLedgerRow(row)
	for name, bad := range map[string]string{
		"missing nullable":      strings.Replace(string(raw), `"input_tokens":null,`, "", 1),
		"quoted credits":        strings.Replace(string(raw), `"credits":null`, `"credits":"0.2"`, 1),
		"nested duplicate":      strings.Replace(string(raw), `"coverage":"unknown"`, `"coverage":"unknown","COVERAGE":"unknown"`, 1),
		"negative zero credits": strings.Replace(string(raw), `"credits":null`, `"credits":-0.0`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeLedgerRow([]byte(bad)); err == nil {
				t.Fatal("invalid nullable/number schema accepted")
			}
		})
	}
}

func TestBenchResultRecordedUnavailableCannotAcquireScore(t *testing.T) {
	r := resultRun(t, true, true)
	h, w := r.Bindings()
	for _, kind := range []EventKind{BeginRecord, RecordConfirmed} {
		next, _, err := Step(r, Event{ID: string(kind), RunID: r.ID(), Kind: kind, OperationID: "record", Hub: h, Worker: w})
		if err != nil {
			t.Fatal(err)
		}
		r = next
	}
	if _, err := BuildRow(validTask(), r, resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("recorded unavailable outcome acquired invented score")
	}
	if _, err := BuildRow(validTask(), r, resultMeasurements(), Acceptance{Status: "unavailable", Reason: "test tool missing"}, strings.Repeat("e", 64), false); err != nil {
		t.Fatal(err)
	}
}

func TestBenchResultValidityInvalidatesAdmittedAndStoppedRuns(t *testing.T) {
	m := resultMeasurements()
	m.Validity = "invalid"
	m.ValidityIssues = []string{"worker/session-ws: actual model differs from provisioned profile"}
	unavailable := Acceptance{Status: "unavailable", Reason: "invalid profile"}
	for _, candidate := range []bool{false, true} {
		r := resultRun(t, candidate, true)
		row, err := BuildRow(validTask(), r, m, unavailable, strings.Repeat("e", 64), false)
		if err != nil {
			t.Fatal(err)
		}
		if row.Outcome != "invalid" || row.Candidate != r.Candidate() || row.Measurements.ValidityIssues[0] != m.ValidityIssues[0] {
			t.Fatalf("invalid result concealed: %+v", row)
		}
	}
	if _, err := BuildRow(validTask(), resultRun(t, true), m, resultAcceptance(), strings.Repeat("e", 64), false); err == nil {
		t.Fatal("invalid run claims a scored acceptance")
	}
	if _, err := BuildRow(validTask(), resultRun(t, true), m, unavailable, strings.Repeat("e", 64), false); err == nil {
		t.Fatal("invalid diagnostic erased an accepted score outcome")
	}
	task := validTask()
	task.Worker.Model = "different-provisioned-binding"
	if _, err := BuildRow(task, resultRun(t, true, true), m, unavailable, strings.Repeat("e", 64), false); err == nil {
		t.Fatal("validity bypassed immutable runtime bindings")
	}
}

func TestBenchResultValidityCompatibilityAndStrictAdditiveFields(t *testing.T) {
	legacy, err := BuildRow(validTask(), resultRun(t, true), resultMeasurements(), resultAcceptance(), strings.Repeat("e", 64), false)
	if err != nil {
		t.Fatal(err)
	}
	old, err := EncodeLedgerRow(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(old), `"validity"`) || strings.Contains(string(old), `"validity_issues"`) {
		t.Fatal("omitted legacy fields normalized into new record bytes")
	}
	decoded, err := DecodeLedgerRow(old)
	if err != nil || decoded.Outcome != "would-land" || decoded.Measurements.Validity != "" {
		t.Fatalf("legacy meaning changed: %+v %v", decoded, err)
	}
	again, err := EncodeLedgerRow(decoded)
	if err != nil || string(again) != string(old) {
		t.Fatal("legacy canonical bytes changed")
	}
	for _, validity := range []string{"valid", "unverified", ""} {
		m := resultMeasurements()
		m.Validity = validity
		row, err := BuildRow(validTask(), resultRun(t, true), m, resultAcceptance(), strings.Repeat("e", 64), false)
		if err != nil || row.Outcome != "would-land" {
			t.Fatalf("validity %q reclassified old outcome: %v", validity, err)
		}
	}
	for name, mutate := range map[string]func(*Measurements){
		"unknown enum":           func(m *Measurements) { m.Validity = "unknown" },
		"invalid without issues": func(m *Measurements) { m.Validity = "invalid" },
		"valid with issues":      func(m *Measurements) { m.Validity = "valid"; m.ValidityIssues = []string{"worker/session: mismatch"} },
		"unlabelled issue":       func(m *Measurements) { m.Validity = "unverified"; m.ValidityIssues = []string{"missing source label"} },
		"empty reason":           func(m *Measurements) { m.ValidityIssues = []string{"worker/session: "} },
		"oversized issue":        func(m *Measurements) { m.ValidityIssues = []string{"worker/session: " + strings.Repeat("x", 4096)} },
		"extension shadows omitted validity": func(m *Measurements) {
			m.Unknown = map[string]json.RawMessage{"validity": json.RawMessage(`"invalid"`)}
		},
	} {
		t.Run(name, func(t *testing.T) {
			m := resultMeasurements()
			mutate(&m)
			if _, err := BuildRow(validTask(), resultRun(t, true, true), m, Acceptance{Status: "unavailable", Reason: "not scored"}, strings.Repeat("e", 64), false); err == nil {
				t.Fatal("invalid diagnostic schema accepted")
			}
		})
	}
	raw := strings.Replace(string(old), `"autonomous":true`, `"autonomous":true,"validity":"unverified","validity_issues":["worker/session: transcript model missing"]`, 1)
	withFuture, err := DecodeLedgerRow([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if withFuture.Measurements.Validity != "unverified" || len(withFuture.Measurements.ValidityIssues) != 1 {
		t.Fatal("new fields lost")
	}
	withFuture.Measurements.Unknown = map[string]json.RawMessage{"future_validity_evidence": json.RawMessage(`{"proof":{"n":1e400}}`)}
	newRaw, err := EncodeLedgerRow(withFuture)
	if err != nil {
		t.Fatal(err)
	}
	redecoded, err := DecodeLedgerRow(newRaw)
	if err != nil || redecoded.Measurements.Unknown["future_validity_evidence"] == nil {
		t.Fatal("future evidence lost", err)
	}
	for _, bad := range []string{strings.Replace(raw, `"validity":"unverified"`, `"validity":null`, 1), strings.Replace(raw, `"validity":"unverified"`, `"VALIDITY":"unverified"`, 1), strings.Replace(raw, `"validity":"unverified"`, `"validity":true`, 1), strings.Replace(raw, `"validity":"unverified"`, `"validity":"unverified","validity":"valid"`, 1)} {
		if _, err := DecodeLedgerRow([]byte(bad)); err == nil {
			t.Fatal("malformed additive validity accepted")
		}
	}
}
