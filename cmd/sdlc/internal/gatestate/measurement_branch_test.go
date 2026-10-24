package gatestate

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	oldmeasurement "github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate/testdata/measurementv1"
	"go.yaml.in/yaml/v3"
)

func branchMeasurement() MeasurementInputs {
	m := validMeasurementInputs()
	m.Version = MeasurementBranchInputsVersion
	m.Method = MeasurementMethodBranchV1
	for i := range m.Sources {
		m.Sources[i].DispatchWindows = m.Sources[i].Windows
	}
	m.Sources[0].DispatchWindows = []MeasurementInterval{{Start: m.Since, End: m.Until}}
	m.Branches = []MeasurementBranchEvidence{{Root: "/old/repo", Path: "/repo/.git/worktrees/worker/logs/HEAD", Head: strings.Repeat("f", 40), Branch: "000335-b", Digest: strings.Repeat("e", 64), Bytes: 120,
		Intervals: []MeasurementBranchInterval{{Source: "codex/main", Start: m.Sources[0].Windows[0].Start, End: m.Sources[0].Windows[0].End, Branch: "000335-b", Kind: "branch"}}}}
	return m
}

func TestMeasurementBranchEvidencePreservesUnknownAndOldAccounting(t *testing.T) {
	m := branchMeasurement()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	m.Branches[0].extra = measurementUnknown(`{"future_capture":{"n":9007199254740993}}`)
	m.Branches[0].Intervals[0].extra = measurementUnknown(`{"future_interval":true}`)
	for _, useYAML := range []bool{false, true} {
		var data []byte
		var err error
		if useYAML {
			data, err = yaml.Marshal(m)
		} else {
			data, err = json.Marshal(m)
		}
		if err != nil {
			t.Fatal(err)
		}
		var again MeasurementInputs
		if useYAML {
			err = yaml.Unmarshal(data, &again)
		} else {
			err = json.Unmarshal(data, &again)
		}
		if err != nil || !reflect.DeepEqual(again, m) {
			t.Fatalf("unknown branch evidence changed: %+v %v", again, err)
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var old oldmeasurement.MeasurementInputs
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatal(err)
	}
	if err := old.Validate(); err == nil {
		t.Fatal("old reader silently accepted branch span interpretation")
	}
	if !reflect.DeepEqual(old.Sources[0].Windows[0].Start, m.Sources[0].Windows[0].Start) || old.Sources[0].Windows[0].End != m.Sources[0].Windows[0].End {
		t.Fatal("old reader widened narrowed windows")
	}
	for _, data := range []string{`{"dispatch_windows":"bad"}`, `{"dispatch_windows":[{"start":7}]}`, `{"dispatch_windows":[],"dispatch_windows":[]}`} {
		var s MeasurementSource
		if err := json.Unmarshal([]byte(data), &s); err == nil {
			t.Fatal("malformed dispatch windows accepted")
		}
	}
	for _, downgrade := range []func(*MeasurementInputs){func(m *MeasurementInputs) { m.Version = 1 }, func(m *MeasurementInputs) { m.Method = MeasurementMethodEngagementV1 }, func(m *MeasurementInputs) { m.Version = 1; m.Method = MeasurementMethodEngagementV1 }} {
		bad := m
		downgrade(&bad)
		if err := bad.Validate(); err == nil {
			t.Fatal("downgraded branch evidence accepted")
		}
	}
}

func TestMeasurementBranchEvidenceRejectsKnownCorruption(t *testing.T) {
	for name, change := range map[string]func(*MeasurementBranchEvidence){
		"root":    func(b *MeasurementBranchEvidence) { b.Root = "relative" },
		"digest":  func(b *MeasurementBranchEvidence) { b.Digest = "bad" },
		"bytes":   func(b *MeasurementBranchEvidence) { b.Bytes = -1 },
		"head":    func(b *MeasurementBranchEvidence) { b.Head = "bad" },
		"kind":    func(b *MeasurementBranchEvidence) { b.Intervals[0].Kind = "guess" },
		"range":   func(b *MeasurementBranchEvidence) { b.Intervals[0].End = b.Intervals[0].Start },
		"outside": func(b *MeasurementBranchEvidence) { b.Intervals[0].Start = "2026-10-01T00:00:00Z" },
		"source":  func(b *MeasurementBranchEvidence) { b.Intervals[0].Source = "" },
	} {
		t.Run(name, func(t *testing.T) {
			m := branchMeasurement()
			change(&m.Branches[0])
			if err := m.Validate(); err == nil {
				t.Fatal("corrupt branch evidence accepted")
			}
		})
	}
	for _, raw := range []string{`{"root":"/a","root":"/b"}`, `{"bytes":"12"}`, `{"intervals":[{"kind":7}]}`} {
		var b MeasurementBranchEvidence
		if err := json.Unmarshal([]byte(raw), &b); err == nil {
			t.Fatalf("known malformed branch evidence accepted: %s", raw)
		}
	}
}

func TestMeasurementNeutralExclusionStrictAndPreserved(t *testing.T) {
	for _, tc := range []struct{ kind, branch string }{
		{"resting-excluded", "main-slot4"}, {"stage-excluded", "stage/000849-000850"},
	} {
		t.Run(tc.kind, func(t *testing.T) { measurementNeutralExclusionStrictAndPreserved(t, tc.kind, tc.branch) })
	}
}

func measurementNeutralExclusionStrictAndPreserved(t *testing.T, kind, branch string) {
	t.Helper()
	m := branchMeasurement()
	w := &m.Branches[0].Intervals[0]
	w.Kind = kind
	w.Branch = branch
	w.extra = measurementUnknown(`{"future":{"n":9007199254740993}}`)
	for _, useYAML := range []bool{false, true} {
		var raw []byte
		var err error
		var again MeasurementInputs
		if useYAML {
			raw, err = yaml.Marshal(m)
			if err == nil {
				err = yaml.Unmarshal(raw, &again)
			}
		} else {
			raw, err = json.Marshal(m)
			if err == nil {
				err = json.Unmarshal(raw, &again)
			}
		}
		if err != nil || !reflect.DeepEqual(m, again) {
			t.Fatalf("neutral exclusion changed: %v", err)
		}
		if err := again.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var old oldmeasurement.MeasurementInputs
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatal(err)
	}
	if err := old.Validate(); err == nil {
		t.Fatal("old reader accepted neutral exclusion proof")
	}
	for _, branch := range []string{"", "feature", "000856-cast", "main-slot04", "stage/", "stage/a..b", "stage/a.lock", "refs/heads/stage/a"} {
		w.Branch = branch
		if err := m.Validate(); err == nil {
			t.Fatalf("invalid neutral exclusion accepted: %s", branch)
		}
	}
}
