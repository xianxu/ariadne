package gatestate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	oldmeasurement "github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate/testdata/measurementv1"
	"go.yaml.in/yaml/v3"
)

// issue586Branch is the shape of #586's issue branch name, 112 characters.
var issue586Branch = ("000586-resume-an-affected-check-attempt-after-a-registered-flake-running-the-groups-it" + strings.Repeat("-x", 20))[:112]

// windowShaped is a version 4 branch measurement shaped like #586's close
// window (#593): n main-session-like sources in one worktree, each with an
// ownership-only sliver then its branch interval, every field at the length
// the real snapshot carried (a 36-character session id, a 52-character cwd,
// offset times).
func windowShaped(n int) MeasurementInputs {
	const layout = "2006-01-02T15:04:05-07:00"
	zone := time.FixedZone("", -7*3600)
	t0 := time.Date(2026, 10, 23, 22, 52, 14, 0, zone)
	at := func(d time.Duration) string { return t0.Add(d).Format(layout) }
	m := validMeasurementInputs()
	m.Version, m.Method = MeasurementCompactBranchInputsVersion, MeasurementMethodBranchV1
	m.Since, m.Until = at(0), at(4*time.Hour)
	m.Engagements = []MeasurementEngagement{{Start: at(0), End: at(4 * time.Hour), Began: at(-time.Hour), Roots: []string{"/Users/xianxu/workspace/worktree/ariadne-slot2/ariadne"}}}
	m.Cuts = nil
	m.Sources = nil
	record := MeasurementBranchEvidence{Root: "/Users/xianxu/workspace/worktree/ariadne-slot2/ariadne", Path: "/Users/xianxu/workspace/ariadne/.git/worktrees/ariadne1/logs/HEAD", Head: strings.Repeat("f", 40), Branch: issue586Branch, Digest: strings.Repeat("e", 64), Bytes: 393720}
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("0191d03b-74fe-4185-8091-%012d", i)
		start, end := at(time.Duration(i)*time.Second), at(3*time.Hour+time.Duration(i)*time.Second)
		window := []MeasurementInterval{{Start: start, End: end}}
		m.Sources = append(m.Sources, MeasurementSource{Harness: "claude", ID: id, Slot: "ariadne:2", CWD: "/Users/xianxu/workspace/worktree/ariadne-slot2/ariadne", Digest: strings.Repeat("d", 64), Started: "2026-10-18T22:59:31.062Z", Bytes: 393720, Windows: window, DispatchWindows: window})
		sliver := at(time.Duration(i)*time.Second + 10*time.Second)
		record.Intervals = append(record.Intervals,
			MeasurementBranchInterval{Source: "claude/" + id, Start: start, End: sliver, Branch: "main-slot2", Kind: "ownership-only"},
			MeasurementBranchInterval{Source: "claude/" + id, Start: sliver, End: end, Branch: issue586Branch, Kind: "branch"})
	}
	m.Branches = []MeasurementBranchEvidence{record}
	return m
}

func roundTrip(t *testing.T, m MeasurementInputs, useYAML bool) MeasurementInputs {
	t.Helper()
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
	if err != nil {
		t.Fatalf("decode: %v\n%s", err, data)
	}
	return again
}

// #593: compaction is lossless: every interval comes back as written,
// including a branch that differs from its record's, an empty branch, a source
// the snapshot does not list, and unknown members at each level.
func TestCompactBranchesRoundTrip(t *testing.T) {
	m := windowShaped(2)
	r := &m.Branches[0]
	r.extra = measurementUnknown(`{"future_capture":{"n":9007199254740993}}`)
	r.Intervals[0].extra = measurementUnknown(`{"future_interval":true}`)
	r.Intervals = append(r.Intervals, MeasurementBranchInterval{Source: "codex/unlisted", Start: m.Since, End: m.Until, Branch: "", Kind: "ownership-only"})
	m.extra = measurementUnknown(`{"future_inputs":[1,2]}`)
	if got := expandOrFail(t, m.Sources, compactBranches(m.Sources, m.Branches)); !reflect.DeepEqual(got, m.Branches) {
		t.Fatalf("expand(compact) changed evidence:\n%+v\n%+v", got, m.Branches)
	}
	for _, useYAML := range []bool{false, true} {
		if again := roundTrip(t, m, useYAML); !reflect.DeepEqual(again, m) {
			t.Fatalf("yaml=%v round trip changed the snapshot:\n%+v\n%+v", useYAML, again.Branches, m.Branches)
		}
	}
	raw, _ := json.Marshal(m)
	if strings.Count(string(raw), issue586Branch) != 1 || strings.Contains(string(raw), `"source":"claude/`) {
		t.Fatalf("v4 bytes repeat the branch or listed sources:\n%s", raw)
	}
}

func expandOrFail(t *testing.T, sources []MeasurementSource, wire []measurementBranchEvidenceV4) []MeasurementBranchEvidence {
	t.Helper()
	out, err := expandBranches(sources, wire)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// #593: a hand-edited or truncated ledger cannot point outside its sources.
func TestExpandBranchesRefusesBadIndex(t *testing.T) {
	m := windowShaped(1)
	for name, edit := range map[string]func(*measurementBranchIntervalV4){
		"past the end": func(c *measurementBranchIntervalV4) { i := 1; c.SourceIndex = &i },
		"negative":     func(c *measurementBranchIntervalV4) { i := -1; c.SourceIndex = &i },
		"both":         func(c *measurementBranchIntervalV4) { c.Source = "claude/x" },
		"neither":      func(c *measurementBranchIntervalV4) { c.SourceIndex = nil },
	} {
		wire := compactBranches(m.Sources, m.Branches)
		edit(&wire[0].Intervals[0])
		if _, err := expandBranches(m.Sources, wire); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// #593: the window that refused #586's close measures under version 4; the
// same evidence as version 3 still exceeds its 64 KiB bound.
func TestMeasurementV4Issue586Shape(t *testing.T) {
	m := windowShaped(75)
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(m)
	if len(raw) > SnapshotLimitBytes {
		t.Errorf("v4 #586 shape is %d bytes; compaction should bring it under %d", len(raw), SnapshotLimitBytes)
	}
	for _, useYAML := range []bool{false, true} {
		if again := roundTrip(t, m, useYAML); !reflect.DeepEqual(again, m) {
			t.Fatalf("yaml=%v round trip changed the snapshot", useYAML)
		}
	}
	v3 := m
	v3.Version = MeasurementBranchInputsVersion
	if err := v3.Validate(); err == nil || !strings.Contains(err.Error(), "over the version 3 bound") {
		t.Fatalf("the fixture no longer reproduces #586's v3 refusal: %v", err)
	}
}

// #593: at the source bound, the byte bound does not bind.
func TestMeasurementV4At128Sources(t *testing.T) {
	m := windowShaped(SnapshotLimitSources)
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	if again := roundTrip(t, m, true); !reflect.DeepEqual(again, m) {
		t.Fatal("yaml round trip at the source bound changed the snapshot")
	}
}

// #593: a snapshot past its bound refuses with its size, bound and counts,
// never the codec's generic serialization failure.
func TestMeasurementOversizeCountedRefusal(t *testing.T) {
	m := windowShaped(SnapshotLimitSources)
	m.Sources[0].CWD = "/" + strings.Repeat("x", SnapshotLimitBytesV4)
	m.Engagements[0].Roots = append(m.Engagements[0].Roots, m.Sources[0].CWD)
	err := m.Validate()
	if err == nil || strings.Contains(err.Error(), "does not serialize") || !strings.Contains(err.Error(), "over the version 4 bound of 131072") || !strings.Contains(err.Error(), "128 sources, 1 branch records, 256 branch intervals") {
		t.Fatalf("oversize refusal: %v", err)
	}
}

// #593: version 3 snapshots replay as they were written: same bytes out, and
// its 64 KiB bound still holds on decode.
func TestMeasurementV3ReplaysUnchanged(t *testing.T) {
	m := branchMeasurement()
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var again MeasurementInputs
	if err := json.Unmarshal(raw, &again); err != nil {
		t.Fatal(err)
	}
	if again2, _ := json.Marshal(again); string(again2) != string(raw) || !strings.Contains(string(raw), `"source":"codex/main"`) {
		t.Fatalf("v3 bytes changed:\n%s\n%s", raw, again2)
	}
	big := windowShaped(75)
	big.Version = MeasurementBranchInputsVersion
	w, _ := big.compactWire()
	raw, err = encodeMeasurementObject(w, big.extra)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &again); err == nil || !strings.Contains(err.Error(), "exceeds 65536 bytes") {
		t.Fatalf("a v3 snapshot over 64 KiB decoded: %v", err)
	}
}

// #593: a reader that predates version 4 refuses it by version, never reads
// its compact intervals as sourceless evidence.
func TestMeasurementV4RefusedByOlderReader(t *testing.T) {
	raw, err := json.Marshal(windowShaped(2))
	if err != nil {
		t.Fatal(err)
	}
	var old oldmeasurement.MeasurementInputs
	if err := json.Unmarshal(raw, &old); err == nil {
		if err := old.Validate(); err == nil {
			t.Fatal("the frozen v1 reader accepted a version 4 snapshot")
		}
	}
	// A pre-#593 reader decodes the compact interval as its v3 shape: no source.
	var v3 []MeasurementBranchEvidence
	var probe struct {
		Branches json.RawMessage `json:"branches"`
	}
	_ = json.Unmarshal(raw, &probe)
	if err := json.Unmarshal(probe.Branches, &v3); err != nil {
		t.Fatal(err)
	}
	if v3[0].Intervals[0].Source != "" {
		t.Fatal("compact interval still carries its source string")
	}
	bad := MeasurementInputs{Version: MeasurementBranchInputsVersion, Method: MeasurementMethodBranchV1, Branches: v3}
	if err := bad.validateBranchEvidence(time.Time{}, time.Now()); err == nil {
		t.Fatal("v3 validation accepted a compact interval")
	}
}

// #593: unknown members at every level survive a v4 rewrite.
func TestMeasurementV4PreservesUnknownMembers(t *testing.T) {
	m := windowShaped(2)
	m.extra = measurementUnknown(`{"future_inputs":true}`)
	m.Branches[0].extra = measurementUnknown(`{"future_capture":1}`)
	m.Branches[0].Intervals[1].extra = measurementUnknown(`{"future_interval":"x"}`)
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var again MeasurementInputs
	if err := json.Unmarshal(raw, &again); err != nil {
		t.Fatal(err)
	}
	if raw2, _ := json.Marshal(again); string(raw2) != string(raw) {
		t.Fatalf("rewrite changed bytes:\n%s\n%s", raw, raw2)
	}
	for _, member := range []string{`"future_inputs":true`, `"future_capture":1`, `"future_interval":"x"`} {
		if !strings.Contains(string(raw), member) {
			t.Fatalf("lost %s", member)
		}
	}
}

// #593 close review: intervals per source have no count bound, so a window
// inside the source bound can pass the encoder's structural cap. That too is
// the counted refusal, never the late MarshalJSON failure.
func TestMeasurementPastStructuralCapCountedRefusal(t *testing.T) {
	m := windowShaped(SnapshotLimitSources)
	since, _ := time.Parse(time.RFC3339, m.Since)
	r := &m.Branches[0]
	for i := 0; i < 4000; i++ {
		start := since.Add(time.Duration(3*i) * time.Second)
		r.Intervals = append(r.Intervals, MeasurementBranchInterval{Source: "codex/unlisted-0191d03b-74fe-4185-8091", Start: start.Format(time.RFC3339), End: start.Add(2 * time.Second).Format(time.RFC3339), Branch: "main-slot2", Kind: "ownership-only"})
	}
	err := m.Validate()
	if err == nil || strings.Contains(err.Error(), "does not serialize") || !strings.Contains(err.Error(), "bytes, over the version 4 bound of 131072") || !strings.Contains(err.Error(), "at least ") || !strings.Contains(err.Error(), "4256 branch intervals") {
		t.Fatalf("past the structural cap: %v", err)
	}
}
