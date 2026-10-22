package gatestate

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	oldmeasurement "github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate/testdata/measurementv1"
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

func TestMeasurementProtectedIntervalsRoundTripAndValidation(t *testing.T) {
	m := validMeasurementInputs()
	credit := MeasurementCredit{Issue: "649", Slot: "ariadne:6", Start: "2026-10-01T01:10:00Z", End: "2026-10-01T01:20:00Z"}
	m.Protected = []MeasurementCredit{credit}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	ledger := Ledger{Gate: "boundary-review", IssueNum: 636, IDPrefix: "BR", Rounds: []Round{{N: 1, ActualInputs: &m}}}
	parsed, err := ParseSidecar(Render(ledger, "ariadne"))
	if err != nil || !reflect.DeepEqual(parsed.Rounds[0].ActualInputs.Protected, m.Protected) {
		t.Fatalf("protected replay pin lost: %+v, %v", parsed, err)
	}
	for name, change := range map[string]func(*MeasurementCredit){
		"issue":   func(c *MeasurementCredit) { c.Issue = "#649" },
		"slot":    func(c *MeasurementCredit) { c.Slot = "" },
		"reverse": func(c *MeasurementCredit) { c.End = c.Start },
		"outside": func(c *MeasurementCredit) { c.Start = "2026-10-01T00:10:00Z" },
	} {
		t.Run(name, func(t *testing.T) {
			bad := m
			bad.Protected = []MeasurementCredit{credit}
			change(&bad.Protected[0])
			if err := bad.Validate(); err == nil {
				t.Fatal("accepted invalid protected interval")
			}
		})
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

func validLegacyPrefixInputs() MeasurementInputs {
	m := validMeasurementInputs()
	m.Version = MeasurementHybridInputsVersion
	m.Method = MeasurementMethodHybridV1
	prefix := validMeasurementInputs()
	prefix.Method = MeasurementMethodLegacyV1
	prefix.CodeHead = strings.Repeat("f", 40)
	prefix.TrackerHead = ""
	prefix.BranchPoint = ""
	prefix.Until = "2026-10-01T02:00:00Z"
	prefix.Engagements = nil
	prefix.Sources = prefix.Sources[:1]
	prefix.Cuts = nil
	m.Engagements[0].Start = prefix.Until
	m.Sources = m.Sources[1:]
	m.LegacyPrefix = &prefix
	m.LegacyCutover = strings.Repeat("e", 40)
	return m
}

func TestMeasurementLegacyPrefixRoundTrip(t *testing.T) {
	for _, method := range []string{MeasurementMethodHybridV1, MeasurementMethodHybridQueueV1} {
		t.Run(method, func(t *testing.T) {
			m := validLegacyPrefixInputs()
			m.Method = method
			if err := m.Validate(); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), `"legacy_prefix"`) || !strings.Contains(string(data), `"legacy_cutover"`) {
				t.Fatalf("missing prefix fields: %s", data)
			}
			var decoded MeasurementInputs
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, m) {
				t.Fatal("JSON changed either component")
			}
			if err := decoded.Validate(); err != nil {
				t.Fatal(err)
			}
			l := Ledger{Gate: "boundary-review", IssueNum: 544, IDPrefix: "BR", Rounds: []Round{{N: 1, ActualInputs: &m}}}
			rendered := Render(l, "ariadne")
			if !strings.Contains(rendered, "legacy_prefix:") || !strings.Contains(rendered, "legacy_cutover:") {
				t.Fatal("missing YAML prefix fields")
			}
			parsed, err := ParseSidecar(rendered)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(parsed.Rounds[0].ActualInputs, &m) {
				t.Fatal("YAML changed either component")
			}
		})
	}
	m := validLegacyPrefixInputs()
	m.Method = MeasurementMethodHybridQueueV1
	m.LegacyCutover = strings.Repeat("e", 64)
	m.QueueWaits = []MeasurementQueueWait{{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: m.LegacyPrefix.Until, End: "2026-10-01T02:01:00Z", QueueSeconds: 60, RawSeconds: 60, Exact: true}}
	m.Protected = []MeasurementCredit{{Issue: "1", Slot: "ariadne:1", Start: m.LegacyPrefix.Until, End: "2026-10-01T02:01:00Z"}}
	if err := m.Validate(); err != nil {
		t.Fatalf("queue/protected boundary equality or SHA256 cutover rejected: %v", err)
	}
	m = validLegacyPrefixInputs()
	m.LegacyPrefix.Since = "2026-09-30T18:00:00-07:00"
	if err := m.Validate(); err != nil {
		t.Fatalf("equivalent Since rejected: %v", err)
	}
	m.LegacyPrefix.Until = m.Until
	m.LegacyPrefix.Sources = nil
	m.Sources = nil
	m.Engagements[0].Start = m.Until
	m.Engagements[0].End = "2026-10-01T04:00:00Z"
	if err := m.Validate(); err != nil {
		t.Fatalf("prefix ending at parent Until rejected: %v", err)
	}
}

func TestMeasurementLegacyPrefixValidation(t *testing.T) {
	for name, mutate := range map[string]func(*MeasurementInputs){
		"downgraded version": func(m *MeasurementInputs) { m.Version = MeasurementInputsVersion },
		"downgraded method":  func(m *MeasurementInputs) { m.Method = MeasurementMethodEngagementV1 },
		"downgraded both": func(m *MeasurementInputs) {
			m.Version = MeasurementInputsVersion
			m.Method = MeasurementMethodEngagementV1
		},
		"cutover absent":           func(m *MeasurementInputs) { m.LegacyCutover = "" },
		"cutover short":            func(m *MeasurementInputs) { m.LegacyCutover = "abc" },
		"cutover nonhex":           func(m *MeasurementInputs) { m.LegacyCutover = strings.Repeat("z", 40) },
		"cutover without prefix":   func(m *MeasurementInputs) { m.LegacyPrefix = nil },
		"legacy parent":            func(m *MeasurementInputs) { m.Method = MeasurementMethodLegacyV1 },
		"child method":             func(m *MeasurementInputs) { m.LegacyPrefix.Method = MeasurementMethodEngagementV1 },
		"child nesting":            func(m *MeasurementInputs) { m.LegacyPrefix.LegacyPrefix = &MeasurementInputs{} },
		"child cutover":            func(m *MeasurementInputs) { m.LegacyPrefix.LegacyCutover = m.LegacyCutover },
		"child repository":         func(m *MeasurementInputs) { m.LegacyPrefix.Repository = "file:/foreign" },
		"child code head":          func(m *MeasurementInputs) { m.LegacyPrefix.CodeHead = "bad" },
		"child tracker":            func(m *MeasurementInputs) { m.LegacyPrefix.TrackerHead = m.TrackerHead },
		"child branch point":       func(m *MeasurementInputs) { m.LegacyPrefix.BranchPoint = m.BranchPoint },
		"child engagements":        func(m *MeasurementInputs) { m.LegacyPrefix.Engagements = m.Engagements },
		"child queue":              func(m *MeasurementInputs) { m.LegacyPrefix.QueueWaits = []MeasurementQueueWait{{}} },
		"child protected":          func(m *MeasurementInputs) { m.LegacyPrefix.Protected = []MeasurementCredit{{}} },
		"child cuts":               func(m *MeasurementInputs) { m.LegacyPrefix.Cuts = m.Cuts },
		"child since":              func(m *MeasurementInputs) { m.LegacyPrefix.Since = "2026-10-01T00:59:59Z" },
		"child until equal since":  func(m *MeasurementInputs) { m.LegacyPrefix.Until = m.Since },
		"child until after parent": func(m *MeasurementInputs) { m.LegacyPrefix.Until = "2026-10-01T04:00:00Z" },
		"child until malformed":    func(m *MeasurementInputs) { m.LegacyPrefix.Until = "bad" },
		"child windows":            func(m *MeasurementInputs) { m.LegacyPrefix.Sources[0].Windows[0].End = m.Until },
		"engagement overlap":       func(m *MeasurementInputs) { m.Engagements[0].Start = "2026-10-01T01:59:59Z" },
		"source overlap":           func(m *MeasurementInputs) { m.Sources[0].Windows[0].Start = "2026-10-01T01:59:59Z" },
		"queue overlap": func(m *MeasurementInputs) {
			m.Method = MeasurementMethodHybridQueueV1
			m.QueueWaits = []MeasurementQueueWait{{AttemptID: "aabb", Slot: "ariadne:1", RecordDigest: strings.Repeat("d", 64), Start: "2026-10-01T01:59:00Z", End: "2026-10-01T02:00:00Z", QueueSeconds: 60, RawSeconds: 60, Exact: true}}
		},
		"protected overlap": func(m *MeasurementInputs) {
			m.Protected = []MeasurementCredit{{Issue: "1", Slot: "ariadne:1", Start: "2026-10-01T01:59:00Z", End: "2026-10-01T02:00:00Z"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			m := validLegacyPrefixInputs()
			mutate(&m)
			if err := m.Validate(); err == nil {
				t.Fatal("accepted invalid legacy prefix")
			}
			// Recursive prefix shapes now refuse before YAML encoding so a
			// caller-constructed cycle never reaches the encoder's recursion.
			if m.LegacyPrefix != nil && (m.LegacyPrefix.LegacyPrefix != nil || m.LegacyPrefix.LegacyCutover != "") {
				if _, err := m.MarshalYAML(); err == nil {
					t.Fatal("accepted recursive prefix encoding")
				}
				return
			}
			if _, err := ParseSidecar(Render(Ledger{Rounds: []Round{{ActualInputs: &m}}}, "test")); err == nil {
				t.Fatal("accepted invalid persisted legacy prefix")
			}
		})
	}
	// Cycles must be rejected before JSON serialization or recursive validation.
	m := validLegacyPrefixInputs()
	m.LegacyPrefix.LegacyPrefix = &m
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "legacy_prefix") {
		t.Fatalf("cycle was not rejected directly: %v", err)
	}
}

func TestMeasurementLegacyPrefixAggregateLimits(t *testing.T) {
	m := validLegacyPrefixInputs()
	duplicate := func(s MeasurementSource, n int) []MeasurementSource {
		out := make([]MeasurementSource, n)
		for i := range out {
			out[i] = s
			out[i].ID = strconv.Itoa(i)
		}
		return out
	}
	m.Sources = duplicate(m.Sources[0], 64)
	m.LegacyPrefix.Sources = duplicate(m.LegacyPrefix.Sources[0], 64)
	if err := m.Validate(); err != nil {
		t.Fatalf("exact aggregate source limit: %v", err)
	}
	m.LegacyPrefix.Sources = append(m.LegacyPrefix.Sources, m.LegacyPrefix.Sources[0])
	m.LegacyPrefix.Sources[64].ID = "extra"
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "sources") {
		t.Fatalf("aggregate source cap: %v", err)
	}
	m = validLegacyPrefixInputs()
	// Each component independently fits; only their combined retained JSON is too large.
	m.Sources[0].ID = strings.Repeat("x", SnapshotLimitBytes/2)
	m.LegacyPrefix.Sources[0].ID = strings.Repeat("y", SnapshotLimitBytes/2)
	if err := m.LegacyPrefix.Validate(); err != nil {
		t.Fatalf("child alone: %v", err)
	}
	root := m
	root.LegacyPrefix = nil
	root.LegacyCutover = ""
	root.Version = MeasurementInputsVersion
	root.Method = MeasurementMethodEngagementV1
	if err := root.Validate(); err != nil {
		t.Fatalf("parent alone: %v", err)
	}
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "bytes") {
		t.Fatalf("aggregate JSON cap: %v", err)
	}
}

func TestMeasurementHybridRejectsPinnedV1Decoder(t *testing.T) {
	for _, method := range []string{MeasurementMethodHybridV1, MeasurementMethodHybridQueueV1} {
		t.Run(method, func(t *testing.T) {
			m := validLegacyPrefixInputs()
			m.Method = method
			for _, discriminator := range []string{"both", "version only", "method only"} {
				t.Run(discriminator, func(t *testing.T) {
					sample := m
					if discriminator == "version only" {
						sample.Method = MeasurementMethodEngagementV1
					}
					if discriminator == "method only" {
						sample.Version = MeasurementInputsVersion
					}
					data, err := json.Marshal(sample)
					if err != nil {
						t.Fatal(err)
					}
					var old oldmeasurement.MeasurementInputs
					if err := json.Unmarshal(data, &old); err != nil {
						t.Fatal(err)
					}
					if err := old.Validate(); err == nil {
						t.Fatalf("pinned v1 decoder silently accepted hybrid without its prefix: %s", data)
					}
				})
			}
		})
	}
	// Unknown optional fields remain accepted by old plain-v1 readers. Refusal
	// depends on schema/interpretation, not an accidental blanket strict decoder.
	for _, method := range []string{MeasurementMethodEngagementV1, MeasurementMethodQueueV1, MeasurementMethodLegacyV1} {
		plain := validMeasurementInputs()
		plain.Method = method
		data, err := json.Marshal(plain)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		fields["future_optional"] = json.RawMessage(`{"untouched":true}`)
		data, err = json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		var old oldmeasurement.MeasurementInputs
		if err := json.Unmarshal(data, &old); err != nil {
			t.Fatal(err)
		}
		if err := old.Validate(); err != nil {
			t.Fatalf("plain v1 compatibility changed for %s: %v", method, err)
		}
		if err := plain.Validate(); err != nil {
			t.Fatalf("current reader rejected plain v1 %s: %v", method, err)
		}
	}
}
