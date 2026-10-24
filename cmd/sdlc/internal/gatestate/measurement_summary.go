package gatestate

import (
	"fmt"
	"math"
	"sort"
	"time"

	"go.yaml.in/yaml/v3"
)

const (
	// MeasurementSummaryInputsVersion is a summarised snapshot (#603). Older
	// readers refuse the version, the safe direction: ignoring the missing
	// source list would read an absent population as evidence.
	MeasurementSummaryInputsVersion = 5
	MeasurementMethodSummaryV1      = "active-time-v3-summary-v1"
	// MaxSummaryCredits bounds a summary's credited intervals on the writer.
	MaxSummaryCredits = 2048
	// SnapshotLimitBytesSummary is a summary's byte bound: MaxSummaryCredits
	// second-precision credits plus the header fit under it.
	SnapshotLimitBytesSummary = 256 * 1024
)

// MeasurementSummary is the retained form of a measurement whose full
// snapshot stays over its bound after trimming to contributing sources (#603).
// The hours are exact; only replay is lost. Credited keeps the intervals a
// later overlapping issue must not count again, so exclusion needs no replay.
type MeasurementSummary struct {
	extra    measurementUnknown  `json:"-" yaml:"-"`
	Sources  int                 `json:"sources" yaml:"sources"`
	Bytes    int                 `json:"bytes" yaml:"bytes"`
	Hours    float64             `json:"hours" yaml:"hours"`
	Digest   string              `json:"digest" yaml:"digest"`
	Reason   string              `json:"reason" yaml:"reason"`
	Credited []MeasurementCredit `json:"credited,omitempty" yaml:"credited,omitempty"`
}

// IsSummaryMethod reports a summarised, non-replayable snapshot.
func IsSummaryMethod(method string) bool { return method == MeasurementMethodSummaryV1 }

// validateSummary checks a summarised snapshot: a header without per-source
// evidence, plus the summary itself.
func (m MeasurementInputs) validateSummary() error {
	if m.Version != MeasurementSummaryInputsVersion || m.Summary == nil {
		return fmt.Errorf("measurement method %s requires version %d and a summary", m.Method, MeasurementSummaryInputsVersion)
	}
	if len(m.Sources) != 0 || len(m.Branches) != 0 || len(m.QueueWaits) != 0 || len(m.Engagements) != 0 || len(m.CarWaits) != 0 || len(m.Protected) != 0 || m.LegacyPrefix != nil {
		return fmt.Errorf("measurement summary must not retain per-source evidence")
	}
	if !measurementHex(m.CodeHead, 40, 64) || m.TrackerHead != "" && !measurementHex(m.TrackerHead, 40, 64) {
		return fmt.Errorf("measurement summary heads must be Git OIDs")
	}
	since, err := measurementTime("since", m.Since)
	if err != nil {
		return err
	}
	until, err := measurementTime("until", m.Until)
	if err != nil {
		return err
	}
	if until.Before(since) {
		return fmt.Errorf("measurement until precedes since")
	}
	s := m.Summary
	if s.Sources <= 0 || s.Bytes < 0 || s.Hours < 0 || math.IsNaN(s.Hours) || math.IsInf(s.Hours, 0) || !measurementHex(s.Digest, 64) || s.Reason == "" {
		return fmt.Errorf("measurement summary counts, hours, digest or reason are invalid")
	}
	if len(s.Credited) > MaxSummaryCredits {
		return fmt.Errorf("measurement summary credits exceed %d", MaxSummaryCredits)
	}
	if err := validateCredits("summary.credited", s.Credited, since, until); err != nil {
		return err
	}
	return m.ValidateUnattributed()
}

// validateCredits checks credited intervals: canonical issue and slot, ordered,
// inside the sample. Protected credits and summary credits share it.
func validateCredits(field string, credits []MeasurementCredit, since, until time.Time) error {
	for i, credit := range credits {
		if err := credit.validate(fmt.Sprintf("%s[%d]", field, i), since, until); err != nil {
			return err
		}
	}
	return nil
}

// CoalesceCredits bounds credits to limit intervals (#603). It merges each
// (issue, slot) group's touching intervals, then repeatedly closes the
// smallest gap inside a group until the total fits. Coalescing only widens, so
// the result covers the input; widened is the added cover, per group.
func CoalesceCredits(credits []MeasurementCredit, limit int) (out []MeasurementCredit, widened []MeasurementCredit) {
	type span struct{ start, end time.Time }
	type group struct {
		issue, slot string
		spans       []span
	}
	index := map[[2]string]int{}
	var groups []group
	for _, c := range credits {
		start, e1 := time.Parse(time.RFC3339Nano, c.Start)
		end, e2 := time.Parse(time.RFC3339Nano, c.End)
		if e1 != nil || e2 != nil || !end.After(start) {
			continue
		}
		// Second precision, widened outward, keeps a credit compact and covering.
		start, end = start.Truncate(time.Second), end.Add(time.Second-time.Nanosecond).Truncate(time.Second)
		key := [2]string{c.Issue, c.Slot}
		i, ok := index[key]
		if !ok {
			i = len(groups)
			index[key] = i
			groups = append(groups, group{issue: c.Issue, slot: c.Slot})
		}
		groups[i].spans = append(groups[i].spans, span{start, end})
	}
	total := 0
	for i := range groups {
		s := groups[i].spans
		sort.Slice(s, func(a, b int) bool { return s[a].start.Before(s[b].start) })
		var merged []span
		for _, x := range s {
			if n := len(merged) - 1; n >= 0 && !x.start.After(merged[n].end) {
				if x.end.After(merged[n].end) {
					merged[n].end = x.end
				}
				continue
			}
			merged = append(merged, x)
		}
		groups[i].spans = merged
		total += len(merged)
	}
	for total > limit && limit > 0 {
		gi, si := -1, -1
		var best time.Duration
		for g := range groups {
			for k := 1; k < len(groups[g].spans); k++ {
				gap := groups[g].spans[k].start.Sub(groups[g].spans[k-1].end)
				if gi < 0 || gap < best {
					gi, si, best = g, k, gap
				}
			}
		}
		if gi < 0 {
			break // one span per group already; more groups than the limit
		}
		s := groups[gi].spans
		widened = append(widened, MeasurementCredit{Issue: groups[gi].issue, Slot: groups[gi].slot, Start: s[si-1].end.Format(time.RFC3339), End: s[si].start.Format(time.RFC3339)})
		s[si-1].end = s[si].end
		groups[gi].spans = append(s[:si], s[si+1:]...)
		total--
	}
	for _, g := range groups {
		for _, x := range g.spans {
			out = append(out, MeasurementCredit{Issue: g.issue, Slot: g.slot, Start: x.start.Format(time.RFC3339), End: x.end.Format(time.RFC3339)})
		}
	}
	return out, widened
}

type measurementSummaryWire MeasurementSummary

func (s *MeasurementSummary) UnmarshalJSON(data []byte) error {
	var v measurementSummaryWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*s = MeasurementSummary(v)
	return nil
}
func (s MeasurementSummary) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementSummaryWire(s), s.extra)
}
func (s *MeasurementSummary) UnmarshalYAML(n *yaml.Node) error { return decodeMeasurementYAML(n, s) }
func (s MeasurementSummary) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementSummaryWire(s), s.extra)
}
