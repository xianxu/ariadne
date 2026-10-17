package gatestate

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func validMeasurementInputs() MeasurementInputs {
	return MeasurementInputs{Version: 1, Method: MeasurementMethodEngagementV1, Repository: "github.com/xianxu/ariadne", CodeHead: strings.Repeat("a", 40), TrackerHead: strings.Repeat("b", 64), BranchPoint: strings.Repeat("c", 40), Since: "2026-10-01T01:00:00Z", Until: "2026-10-01T03:00:00Z", BrainRoot: "/brain", GapMinutes: 15, CommitWeight: 1, IncludeAssistant: true,
		Engagements: []MeasurementEngagement{{Start: "2026-10-01T00:00:00Z", End: "2026-10-01T03:00:00Z", Began: "2026-09-30T00:00:00Z", Roots: []string{"/repo", "/old/repo"}}},
		Sources:     []MeasurementSource{{Harness: "codex", ID: "main", CWD: "/old/repo", Digest: strings.Repeat("d", 64), Started: "2026-09-30T00:00:00Z", Bytes: 123, Windows: []MeasurementInterval{{Start: "2026-10-01T01:00:00Z", End: "2026-10-01T02:00:00Z"}}}, {Harness: "codex", ID: "child", CWD: "/repo", ParentID: "main", Spawned: true, Digest: strings.Repeat("e", 64), Started: "2026-10-01T00:30:00Z", Bytes: 25, Windows: []MeasurementInterval{{Start: "2026-10-01T02:00:00Z", End: "2026-10-01T03:00:00Z"}}}},
		Cuts:        []MeasurementCut{{Boundary: "M1", Dispatched: "2026-10-01T02:00:00Z", Finalized: "2026-10-01T02:30:00Z"}, {Dispatched: "2026-10-01T03:00:00Z", Finalized: "2026-10-01T03:00:00Z", Finalizing: true}}}
}

func TestMeasurementInputsRoundTrip(t *testing.T) {
	m := validMeasurementInputs()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	l := Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []Round{{N: 1, ActualInputs: &m}}}
	rendered := Render(l, "ariadne")
	got, err := ParseSidecar(rendered)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Rounds[0].ActualInputs, &m) {
		t.Fatalf("snapshot lost or changed: %#v", got.Rounds[0].ActualInputs)
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"code_head"`) {
		t.Fatalf("JSON field names: %s", data)
	}
	var decoded MeasurementInputs
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, m) {
		t.Fatal("JSON snapshot changed")
	}
}

func TestMeasurementQueueWaitValidate(t *testing.T) {
	m := validMeasurementInputs()
	m.Method = MeasurementMethodQueueV1
	q := MeasurementQueueWait{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: "2026-10-01T01:10:00Z", End: "2026-10-01T01:11:00Z", QueueSeconds: 60, RawSeconds: 60, Exact: true}
	m.QueueWaits = []MeasurementQueueWait{q}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid queue snapshot: %v", err)
	}
	for name, mutate := range map[string]func(*MeasurementInputs){
		"identity":   func(m *MeasurementInputs) { m.QueueWaits[0].AttemptID = "" },
		"digest":     func(m *MeasurementInputs) { m.QueueWaits[0].RecordDigest = "bad" },
		"reverse":    func(m *MeasurementInputs) { m.QueueWaits[0].End = "2026-10-01T01:09:00Z" },
		"duration":   func(m *MeasurementInputs) { m.QueueWaits[0].QueueSeconds = 59 },
		"negative":   func(m *MeasurementInputs) { m.QueueWaits[0].QueueSeconds = -1 },
		"outside":    func(m *MeasurementInputs) { m.QueueWaits[0].Start = "2026-10-01T00:10:00Z" },
		"duplicate":  func(m *MeasurementInputs) { m.QueueWaits = append(m.QueueWaits, m.QueueWaits[0]) },
		"old method": func(m *MeasurementInputs) { m.Method = MeasurementMethodEngagementV1 },
	} {
		t.Run(name, func(t *testing.T) {
			bad := m
			bad.QueueWaits = append([]MeasurementQueueWait(nil), m.QueueWaits...)
			mutate(&bad)
			if err := bad.Validate(); err == nil {
				t.Fatal("accepted invalid queue evidence")
			}
		})
	}
}

func TestMeasurementInputsValidation(t *testing.T) {
	cases := []struct {
		name, field string
		mutate      func(*MeasurementInputs)
	}{
		{"version", "version", func(m *MeasurementInputs) { m.Version = 2 }},
		{"method", "method", func(m *MeasurementInputs) { m.Method = "active-time-v4" }},
		{"repository", "repository", func(m *MeasurementInputs) { m.Repository = "" }},
		{"brain", "brain_root", func(m *MeasurementInputs) { m.BrainRoot = "brain" }},
		{"code", "code_head", func(m *MeasurementInputs) { m.CodeHead = "abcd" }},
		{"tracker", "tracker_head", func(m *MeasurementInputs) { m.TrackerHead = "" }},
		{"base", "branch_point", func(m *MeasurementInputs) { m.BranchPoint = strings.Repeat("g", 40) }},
		{"since", "since", func(m *MeasurementInputs) { m.Since = "today" }},
		{"until", "until", func(m *MeasurementInputs) { m.Until = m.Engagements[0].Start }},
		{"brain count", "brain provenance", func(m *MeasurementInputs) { m.ExcludedBrainSessions = -1 }},
		{"brain minutes", "brain provenance", func(m *MeasurementInputs) { m.ExcludedBrainMinutes = math.NaN() }},
		{"gap", "gap_minutes", func(m *MeasurementInputs) { m.GapMinutes = 0 }},
		{"weight", "commit_weight", func(m *MeasurementInputs) { m.CommitWeight = -1 }},
		{"weight nan", "commit_weight", func(m *MeasurementInputs) { m.CommitWeight = math.NaN() }},
		{"weight infinity", "commit_weight", func(m *MeasurementInputs) { m.CommitWeight = math.Inf(1) }},
		{"engagements", "engagements", func(m *MeasurementInputs) { m.Engagements = nil }},
		{"engagement end", "engagements", func(m *MeasurementInputs) { m.Engagements[0].End = m.Engagements[0].Began }},
		{"engagement began", "began", func(m *MeasurementInputs) { m.Engagements[0].Began = m.Until }},
		{"roots", "roots", func(m *MeasurementInputs) { m.Engagements[0].Roots = nil }},
		{"root relative", "roots", func(m *MeasurementInputs) { m.Engagements[0].Roots[0] = "repo" }},
		{"harness", "harness", func(m *MeasurementInputs) { m.Sources[0].Harness = "" }},
		{"id", "id", func(m *MeasurementInputs) { m.Sources[0].ID = "" }},
		{"cwd", "cwd", func(m *MeasurementInputs) { m.Sources[0].CWD = "relative" }},
		{"digest", "digest", func(m *MeasurementInputs) { m.Sources[0].Digest = strings.Repeat("g", 64) }},
		{"bytes", "bytes", func(m *MeasurementInputs) { m.Sources[0].Bytes = 0 }},
		{"started", "started", func(m *MeasurementInputs) { m.Sources[0].Started = "invalid" }},
		{"started future", "started", func(m *MeasurementInputs) { m.Sources[0].Started = "2026-10-02T00:00:00Z" }},
		{"duplicate", "duplicate", func(m *MeasurementInputs) { m.Sources[1] = m.Sources[0] }},
		{"windows absent", "windows", func(m *MeasurementInputs) { m.Sources[0].Windows = nil }},
		{"window before", "windows", func(m *MeasurementInputs) { m.Sources[0].Windows[0].Start = m.Engagements[0].Start }},
		{"window after", "windows", func(m *MeasurementInputs) { m.Sources[0].Windows[0].End = "2026-10-02T00:00:00Z" }},
		{"window reverse", "windows", func(m *MeasurementInputs) {
			m.Sources[0].Windows[0].End = m.Since
			m.Sources[0].Windows[0].Start = "2026-10-01T02:00:00Z"
		}},
		{"cut", "cuts", func(m *MeasurementInputs) { m.Cuts[0].Finalized = m.Since }},
		{"cut absent finalize", "finalized", func(m *MeasurementInputs) { m.Cuts[0].Finalized = "" }},
		{"sources cap", "sources", func(m *MeasurementInputs) { m.Sources = make([]MeasurementSource, SnapshotLimitSources+1) }},
		{"bytes cap", "snapshot", func(m *MeasurementInputs) { m.Sources[0].ID = strings.Repeat("x", SnapshotLimitBytes) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := validMeasurementInputs()
			tc.mutate(&m)
			err := m.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("want %s rejection, got %v", tc.field, err)
			}
			if _, err := ParseSidecar(Render(Ledger{Rounds: []Round{{ActualInputs: &m}}}, "test")); err == nil {
				t.Fatal("invalid persisted snapshot accepted")
			}
		})
	}
}

func TestMeasurementInputsLegacyAndAbsent(t *testing.T) {
	m := validMeasurementInputs()
	m.Method = MeasurementMethodLegacyV1
	m.TrackerHead = ""
	m.Engagements = nil
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMeasurementInputs(nil); err != nil {
		t.Fatal(err)
	}
	l := Ledger{Gate: "boundary-review", IssueNum: 335, IDPrefix: "BR", Rounds: []Round{{N: 1, Timestamp: "then", Agent: "codex"}}}
	text := Render(l, "ariadne")
	if strings.Contains(text, "actual_inputs") {
		t.Fatal("absent snapshot changed legacy serialization")
	}
	parsed, err := ParseSidecar(text)
	if err != nil {
		t.Fatal(err)
	}
	if Render(parsed, "ariadne") != text {
		t.Fatal("legacy bytes changed")
	}
}

func TestMeasurementInputsOptionalPinsAndSourceCeiling(t *testing.T) {
	m := validMeasurementInputs()
	m.BranchPoint = ""
	m.BrainRoot = ""
	m.CommitWeight = 0
	m.Cuts[0].Dispatched = ""
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	source := m.Sources[0]
	m.Sources = nil
	for i := 0; i < SnapshotLimitSources; i++ {
		copy := source
		copy.ID = strconv.Itoa(i)
		m.Sources = append(m.Sources, copy)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("exact source ceiling rejected: %v", err)
	}
}

func TestMeasurementInputsMissingVersionIsRetainedError(t *testing.T) {
	m := validMeasurementInputs()
	doc := Render(Ledger{Rounds: []Round{{N: 7, ActualInputs: &m}}}, "ariadne")
	doc = strings.Replace(doc, "        version: 1\n", "", 1)
	_, err := ParseSidecar(doc)
	if err == nil || !strings.Contains(err.Error(), "round 7 actual_inputs") || !strings.Contains(err.Error(), "version") {
		t.Fatalf("missing snapshot version must remain an evidence error: %v", err)
	}
}
