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
func resultAcceptance() Acceptance {
	return Acceptance{Status: "command", CommandPassed: resultPtr(true), ParentRed: true, LandingGreen: true, PatchSHA256: strings.Repeat("d", 64)}
}

func TestBenchResultMeasurementsUnknownAndExtensionsRoundTrip(t *testing.T) {
	m := resultMeasurements()
	m.Worker.Credits = resultPtr(json.Number("0.0000000000000000001234"))
	m.Worker.Coverage = "partial"
	m.Worker.SourceIDs = []string{"worker-source"}
	m.Worker.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`{"x":1}`)}
	m.Unknown = map[string]json.RawMessage{"future": json.RawMessage(`true`)}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"input_tokens":null`) {
		t.Fatal("unknown token counter was lost")
	}
	var got Measurements
	if err = json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != string(again) {
		t.Fatalf("round trip changed evidence: %s\n%s", raw, again)
	}
	if got.Worker.InputTokens != nil || got.ActiveHours != nil {
		t.Fatal("unknown fabricated as zero")
	}
}

func TestBenchResultRejectsInvalidMeasurementEvidence(t *testing.T) {
	for name, change := range map[string]func(*Measurements){
		"negative usage":               func(m *Measurements) { m.Worker.InputTokens = resultPtr(int64(-1)) },
		"missing complete counters":    func(m *Measurements) { m.Worker.Coverage = "complete" },
		"wait exceeds gross":           func(m *Measurements) { m.OperatorWaitNanos = 101 },
		"autonomy contradicts wait":    func(m *Measurements) { m.OperatorWaitNanos = 1 },
		"negative credits":             func(m *Measurements) { m.Worker.Credits = resultPtr(json.Number("-0.1")) },
		"count contradicts identities": func(m *Measurements) { m.ReviewRounds = resultPtr(1) },
		"invalid without reason":       func(m *Measurements) { m.Validity = "invalid" },
		"unlabelled validity issue":    func(m *Measurements) { m.Validity = "invalid"; m.ValidityIssues = []string{"mismatch"} },
	} {
		t.Run(name, func(t *testing.T) {
			m := resultMeasurements()
			change(&m)
			raw, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			var got Measurements
			if json.Unmarshal(raw, &got) == nil {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
}

func TestBenchResultMeasurementStrictJSON(t *testing.T) {
	raw, err := json.Marshal(resultMeasurements())
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		strings.Replace(string(raw), `"active_hours":null,`, "", 1),
		strings.Replace(string(raw), `"gross_nanos":100`, `"gross_nanos":100,"gross_nanos":100`, 1),
		strings.Replace(string(raw), `"gross_nanos":100`, `"Gross_Nanos":100`, 1),
		strings.Replace(string(raw), `"autonomous":true`, `"autonomous":true,"validity":null`, 1),
		strings.Replace(string(raw), `"autonomous":true`, `"autonomous":true,"future":"`+strings.Repeat("x", ResultObjectLimit)+`"`, 1),
	} {
		var got Measurements
		if json.Unmarshal([]byte(bad), &got) == nil {
			t.Fatal("malformed measurement accepted")
		}
	}
}

func TestBenchResultAcceptanceCodecAndCaseValidation(t *testing.T) {
	a := resultAcceptance()
	raw, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	var got Acceptance
	if err = json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Total != nil {
		t.Fatal("command result invented case counts")
	}
	a.Status = "cases"
	a.CommandPassed = resultPtr(false)
	a.Passed = resultPtr(2)
	a.Failed = resultPtr(1)
	a.Total = resultPtr(3)
	if err = ValidateAcceptance(a); err != nil {
		t.Fatal(err)
	}
	*a.Failed = 0
	if ValidateAcceptance(a) == nil {
		t.Fatal("inconsistent case counts accepted")
	}
	if ValidateAcceptance(Acceptance{Status: "unavailable", Reason: "missing scorer"}) != nil {
		t.Fatal("unavailable score rejected")
	}
}
