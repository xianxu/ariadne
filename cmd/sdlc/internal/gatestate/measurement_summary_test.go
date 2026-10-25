package gatestate

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// #603: unattributed minutes are audit records beside the counted evidence.
func TestMeasurementUnattributedRoundTripAndValidation(t *testing.T) {
	m := validMeasurementInputs()
	m.Unattributed = []MeasurementUnattributed{
		{Kind: UnattributedUnknownBranch, Source: "codex/main", Start: "2026-10-01T01:10:00Z", End: "2026-10-01T01:20:00Z", Bounded: true, Reason: "concurrent claims have unknown, detached or ambiguous active branch"},
		{Kind: UnattributedQueueUnknown, Reason: "legacy suite attempt aabb lacks exact queue endpoints"},
	}
	m.UnattributedMinutes = 10
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseSidecar(Render(Ledger{Gate: "boundary-review", IssueNum: 603, IDPrefix: "BR", Rounds: []Round{{N: 1, ActualInputs: &m}}}, "ariadne"))
	if err != nil || !reflect.DeepEqual(parsed.Rounds[0].ActualInputs, &m) {
		t.Fatalf("unattributed records lost or changed: %v %#v", err, parsed.Rounds)
	}
	for name, mutate := range map[string]func(*MeasurementInputs){
		"unknown kind":         func(m *MeasurementInputs) { m.Unattributed[0].Kind = "guess" },
		"empty reason":         func(m *MeasurementInputs) { m.Unattributed[0].Reason = " " },
		"multi-line reason":    func(m *MeasurementInputs) { m.Unattributed[0].Reason = "a\nb" },
		"unbounded with times": func(m *MeasurementInputs) { m.Unattributed[1].Start = "2026-10-01T01:10:00Z" },
		"outside sample":       func(m *MeasurementInputs) { m.Unattributed[0].End = "2026-10-01T04:00:00Z" },
		"unordered":            func(m *MeasurementInputs) { m.Unattributed[0].End = m.Unattributed[0].Start },
		"negative minutes":     func(m *MeasurementInputs) { m.UnattributedMinutes = -1 },
		"over the record bound": func(m *MeasurementInputs) {
			m.Unattributed = make([]MeasurementUnattributed, SnapshotLimitUnattributed+1)
		},
	} {
		bad := m
		bad.Unattributed = append([]MeasurementUnattributed(nil), m.Unattributed...)
		mutate(&bad)
		if bad.Validate() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func validSummaryInputs() MeasurementInputs {
	return MeasurementInputs{Version: MeasurementSummaryInputsVersion, Method: MeasurementMethodSummaryV1, Repository: "github.com/xianxu/ariadne", CodeHead: strings.Repeat("a", 40), TrackerHead: strings.Repeat("b", 64), Since: "2026-10-01T01:00:00Z", Until: "2026-10-01T03:00:00Z", GapMinutes: 15, CommitWeight: 1,
		Summary: &MeasurementSummary{Sources: 140, Bytes: 200000, Hours: 1.5, Digest: strings.Repeat("f", 64), Reason: "measurement sources exceed snapshot limit 128",
			Credited: []MeasurementCredit{{Issue: "603", Slot: "ariadne:2", Start: "2026-10-01T01:00:00Z", End: "2026-10-01T02:30:00Z"}}}}
}

// #603: a summary keeps the header and the credits, never per-source evidence.
func TestMeasurementSummaryValidation(t *testing.T) {
	m := validSummaryInputs()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var back MeasurementInputs
	if err := json.Unmarshal(raw, &back); err != nil || !reflect.DeepEqual(back, m) {
		t.Fatalf("summary round trip: %v", err)
	}
	full := validMeasurementInputs()
	for name, mutate := range map[string]func(*MeasurementInputs){
		"wrong version":        func(m *MeasurementInputs) { m.Version = MeasurementCompactBranchInputsVersion },
		"missing summary":      func(m *MeasurementInputs) { m.Summary = nil },
		"per-source evidence":  func(m *MeasurementInputs) { m.Sources = full.Sources },
		"bad digest":           func(m *MeasurementInputs) { m.Summary.Digest = "x" },
		"credit outside":       func(m *MeasurementInputs) { m.Summary.Credited[0].End = "2026-10-01T04:00:00Z" },
		"too many credits":     func(m *MeasurementInputs) { m.Summary.Credited = make([]MeasurementCredit, MaxSummaryCredits+1) },
		"summary on full form": func(m *MeasurementInputs) { *m = validMeasurementInputs(); m.Summary = validSummaryInputs().Summary },
	} {
		bad := validSummaryInputs()
		mutate(&bad)
		if bad.Validate() == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// #603: the bounds are a typed error, so a fresh measurement can summarise
// instead of refusing, while the shape stays strict.
func TestMeasurementBoundsAreTypedAndSeparateFromShape(t *testing.T) {
	m := validMeasurementInputs()
	base := m.Sources[0]
	m.Sources = nil
	for i := 0; i <= SnapshotLimitSources; i++ {
		s := base
		s.ID = fmt.Sprintf("s%03d", i)
		m.Sources = append(m.Sources, s)
	}
	if err := m.ValidateShape(); err != nil {
		t.Fatalf("shape refused an over-cap snapshot: %v", err)
	}
	var over *OversizeError
	if err := m.ValidateBounds(); !errors.As(err, &over) || !strings.Contains(err.Error(), "sources exceed snapshot limit 128") {
		t.Fatalf("bounds = %v, want a typed source-cap error", err)
	}
	if err := m.Validate(); !errors.As(err, &over) {
		t.Fatalf("Validate = %v, want the typed bound", err)
	}
}

func coverOf(credits []MeasurementCredit) map[[2]string][][2]time.Time {
	out := map[[2]string][][2]time.Time{}
	for _, c := range credits {
		s, _ := time.Parse(time.RFC3339Nano, c.Start)
		e, _ := time.Parse(time.RFC3339Nano, c.End)
		k := [2]string{c.Issue, c.Slot}
		out[k] = append(out[k], [2]time.Time{s, e})
	}
	for k, spans := range out {
		sort.Slice(spans, func(i, j int) bool { return spans[i][0].Before(spans[j][0]) })
		var merged [][2]time.Time
		for _, x := range spans {
			if n := len(merged) - 1; n >= 0 && !x[0].After(merged[n][1]) {
				if x[1].After(merged[n][1]) {
					merged[n][1] = x[1]
				}
				continue
			}
			merged = append(merged, x)
		}
		out[k] = merged
	}
	return out
}

func coveredSeconds(cover map[[2]string][][2]time.Time) float64 {
	total := 0.0
	for _, spans := range cover {
		for _, x := range spans {
			total += x[1].Sub(x[0]).Seconds()
		}
	}
	return total
}

func covers(outer, inner map[[2]string][][2]time.Time) bool {
	for k, spans := range inner {
		for _, x := range spans {
			ok := false
			for _, y := range outer[k] {
				if !x[0].Before(y[0]) && !x[1].After(y[1]) {
					ok = true
					break
				}
			}
			if !ok {
				return false
			}
		}
	}
	return true
}

// #603 coalescer property: within the limit, covering the input, and the
// widened minutes are exactly the cover it added.
func TestCoalesceCreditsProperty(t *testing.T) {
	r := rand.New(rand.NewSource(603))
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for trial := 0; trial < 200; trial++ {
		n, limit := 1+r.Intn(300), 1+r.Intn(40)
		var in []MeasurementCredit
		for i := 0; i < n; i++ {
			start := base.Add(time.Duration(r.Intn(36000)) * time.Second)
			in = append(in, MeasurementCredit{Issue: fmt.Sprint(1 + r.Intn(3)), Slot: fmt.Sprintf("ariadne:%d", 1+r.Intn(2)), Start: start.Format(time.RFC3339), End: start.Add(time.Duration(1+r.Intn(600)) * time.Second).Format(time.RFC3339)})
		}
		out, widened := CoalesceCredits(in, limit)
		groups := len(coverOf(in))
		if len(out) > limit && len(out) > groups {
			t.Fatalf("trial %d: %d credits over limit %d", trial, len(out), limit)
		}
		inCover, outCover := coverOf(in), coverOf(out)
		if !covers(outCover, inCover) {
			t.Fatalf("trial %d: output does not cover the input", trial)
		}
		// The first pass's gaps are exactly the added cover; bounding them can
		// only widen what a later issue is told it lost, never hide any.
		_, exact := coalesceCredits(in, limit)
		added := coveredSeconds(outCover) - coveredSeconds(inCover)
		if got := coveredSeconds(coverOf(exact)); got != added {
			t.Fatalf("trial %d: widened %vs, added cover %vs", trial, got, added)
		}
		if len(widened) > limit && len(widened) > len(coverOf(widened)) || !covers(coverOf(widened), coverOf(exact)) {
			t.Fatalf("trial %d: %d widened gaps over limit %d, or they lost cover", trial, len(widened), limit)
		}
	}
}

// Within the limit the coalescer only merges; below it, each closed gap
// removes exactly one credit, at N and 2N.
func TestCoalesceCreditsWithinLimitIsMergeOnly(t *testing.T) {
	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	for _, n := range []int{500, 1000} {
		var in []MeasurementCredit
		for i := 0; i < n; i++ {
			start := base.Add(time.Duration(i) * time.Minute)
			in = append(in, MeasurementCredit{Issue: "603", Slot: "ariadne:2", Start: start.Format(time.RFC3339), End: start.Add(30 * time.Second).Format(time.RFC3339)})
		}
		out, widened := CoalesceCredits(in, MaxSummaryCredits)
		if len(out) != n || len(widened) != 0 {
			t.Fatalf("n=%d: %d out, %d widened", n, len(out), len(widened))
		}
		out, widened = CoalesceCredits(in, n/2)
		if len(out) != n/2 || len(widened) != n-n/2 {
			t.Fatalf("n=%d to %d: %d out, %d widened", n, n/2, len(out), len(widened))
		}
	}
}

// pair#866 (#603): under the source cap, over the version 4 byte bound on its
// branch intervals. Its close refused "measurement snapshot is 182432 bytes,
// over the version 4 bound of 131072 (119 sources, 1 branch records, 860
// branch intervals)"; the bound is now a typed error a fresh measurement
// summarises, while the shape stays valid.
func TestMeasurementPair866ShapeIsTypedOversize(t *testing.T) {
	m := windowShaped(119)
	record := &m.Branches[0]
	var intervals []MeasurementBranchInterval
	for _, w := range record.Intervals {
		if w.Kind != "branch" {
			intervals = append(intervals, w)
			continue
		}
		// Each source's branch span becomes six alternating slivers (seven
		// intervals per source), and the first source takes the remainder, so
		// the capture carries 860 intervals, as the long issue's did.
		start, _ := time.Parse(time.RFC3339, w.Start)
		end, _ := time.Parse(time.RFC3339, w.End)
		parts := 6
		if w.Source == record.Intervals[1].Source {
			parts += 860 - 119*7
		}
		step := end.Sub(start) / time.Duration(parts)
		for k := 0; k < parts; k++ {
			a, b := start.Add(time.Duration(k)*step), start.Add(time.Duration(k+1)*step)
			if k == parts-1 {
				b = end
			}
			kind, branch := "branch", w.Branch
			if k%2 == 1 {
				kind, branch = "ownership-only", "main-slot2"
			}
			intervals = append(intervals, MeasurementBranchInterval{Source: w.Source, Start: a.Format("2006-01-02T15:04:05-07:00"), End: b.Format("2006-01-02T15:04:05-07:00"), Branch: branch, Kind: kind})
		}
	}
	record.Intervals = intervals
	if len(record.Intervals) != 860 {
		t.Fatalf("fixture carries %d intervals, want 860", len(record.Intervals))
	}
	if err := m.ValidateShape(); err != nil {
		t.Fatalf("pair#866's shape is invalid: %v", err)
	}
	var over *OversizeError
	err := m.ValidateBounds()
	if !errors.As(err, &over) || over.Bytes <= SnapshotLimitBytesV4 || !strings.Contains(err.Error(), "over the version 4 bound of 131072 (119 sources, 1 branch records, 860 branch intervals)") {
		t.Fatalf("bounds = %v, want pair#866's typed oversize", err)
	}
}

// #603: the writer's bounds match the validator's: a window with far more than
// twice MaxSummaryCredits credits still summarises into a valid snapshot.
func TestCoalesceCreditsBoundsWidenedForTheValidator(t *testing.T) {
	base := time.Date(2026, 10, 1, 1, 0, 0, 0, time.UTC)
	var in []MeasurementCredit
	for i := 0; i < 3*MaxSummaryCredits; i++ {
		start := base.Add(time.Duration(2*i) * time.Second)
		in = append(in, MeasurementCredit{Issue: "603", Slot: "ariadne:2", Start: start.Format(time.RFC3339), End: start.Add(time.Second).Format(time.RFC3339)})
	}
	out, widened := CoalesceCredits(in, MaxSummaryCredits)
	if len(out) > MaxSummaryCredits || len(widened) > MaxSummaryCredits {
		t.Fatalf("%d credits, %d widened gaps over %d", len(out), len(widened), MaxSummaryCredits)
	}
	if _, exact := coalesceCredits(in, MaxSummaryCredits); !covers(coverOf(widened), coverOf(exact)) {
		t.Fatal("bounding the widened gaps lost cover")
	}
	m := validSummaryInputs()
	m.Since, m.Until = base.Format(time.RFC3339), base.Add(4*time.Hour).Format(time.RFC3339)
	m.Summary.Credited, m.Summary.Widened = out, widened
	if err := m.Validate(); err != nil {
		t.Fatalf("a writer-bounded summary fails validation: %v", err)
	}
}

// #603 rule: the writer's normaliser and Validate share one predicate, so
// Validate accepts every normalised record, whatever adversarial strings
// (NUL, CR/LF, invalid UTF-8, over-long, empty) its writer was handed.
func TestUnattributedNormalisedAlwaysValidates(t *testing.T) {
	r := rand.New(rand.NewSource(603))
	pieces := []string{"", " ", "\x00", "\r\n", "\n", "\xff\xfe", "é", "ascii", strings.Repeat("é", 300), strings.Repeat("x", 500), "\t", "#", ": "}
	field := func() string {
		var b strings.Builder
		for i := r.Intn(6); i > 0; i-- {
			b.WriteString(pieces[r.Intn(len(pieces))])
		}
		return b.String()
	}
	for trial := 0; trial < 2000; trial++ {
		u := MeasurementUnattributed{Kind: UnattributedQueueUnknown, Source: field(), Reason: field()}.Normalised()
		if err := u.Validate(); err != nil {
			t.Fatalf("trial %d: normalised record %q / %q fails validation: %v", trial, u.Source, u.Reason, err)
		}
	}
	if got := BoundLine("short line", UnattributedReasonLimit); got != "short line" {
		t.Fatalf("a valid short line changed: %q", got)
	}
	if got := BoundLine(strings.Repeat("é", UnattributedReasonLimit), UnattributedReasonLimit); len(got) > UnattributedReasonLimit || !utf8.ValidString(got) || !strings.HasSuffix(got, "...") {
		t.Fatalf("a cut is not rune-safe: %d bytes", len(got))
	}
}
