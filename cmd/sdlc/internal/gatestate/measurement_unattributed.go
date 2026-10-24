package gatestate

import (
	"fmt"
	"math"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Unattributed kinds name why seen minutes were not counted (#603).
const (
	UnattributedUnknownBranch     = "unknown-branch"     // concurrent claims, branch unknown, detached or ambiguous
	UnattributedSourceAttribution = "source-attribution" // one source's branch attribution failed
	UnattributedQueueUnknown      = "queue-unknown"      // a suite attempt whose queue provenance is unknown
	UnattributedSummaryWidened    = "summary-widened"    // an earlier issue's summarised credits over-protect
	SnapshotLimitUnattributed     = 32
	unattributedReasonLimit       = 400
)

// MeasurementUnattributed records minutes the measurement saw but did not
// count (#603). It is audit evidence only: the sources' counted windows never
// held these minutes, so a reader that ignores the member replays exactly the
// adopted hours. That makes it safe for an older reader to preserve and ignore.
// A bounded record has its interval; an unbounded one (a queue attempt with no
// endpoints) has none and makes the measurement an upper bound.
type MeasurementUnattributed struct {
	extra   measurementUnknown `json:"-" yaml:"-"`
	Kind    string             `json:"kind" yaml:"kind"`
	Source  string             `json:"source,omitempty" yaml:"source,omitempty"`
	Start   string             `json:"start,omitempty" yaml:"start,omitempty"`
	End     string             `json:"end,omitempty" yaml:"end,omitempty"`
	Bounded bool               `json:"bounded" yaml:"bounded"`
	Reason  string             `json:"reason" yaml:"reason"`
}

// Validate checks one record's shape; the snapshot checks it lies in sample.
func (u MeasurementUnattributed) Validate() error {
	switch u.Kind {
	case UnattributedUnknownBranch, UnattributedSourceAttribution, UnattributedQueueUnknown, UnattributedSummaryWidened:
	default:
		return fmt.Errorf("measurement unattributed kind %q is invalid", u.Kind)
	}
	if strings.TrimSpace(u.Reason) == "" || len(u.Reason) > unattributedReasonLimit || strings.ContainsAny(u.Reason, "\r\n\x00") || strings.ContainsAny(u.Source, "\r\n\x00") {
		return fmt.Errorf("measurement unattributed reason or source is invalid")
	}
	if !u.Bounded {
		if u.Start != "" || u.End != "" {
			return fmt.Errorf("measurement unbounded unattributed record carries an interval")
		}
		return nil
	}
	start, err := measurementTime("unattributed start", u.Start)
	if err != nil {
		return err
	}
	end, err := measurementTime("unattributed end", u.End)
	if err != nil {
		return err
	}
	if !end.After(start) {
		return fmt.Errorf("measurement unattributed end must follow start")
	}
	return nil
}

// ValidateUnattributed bounds the audit records independently of the
// accounting snapshot, as car-wait coverage is bounded: the records are capped,
// while unattributed_minutes always carries the whole total.
func (m MeasurementInputs) ValidateUnattributed() error {
	if len(m.Unattributed) > SnapshotLimitUnattributed {
		return fmt.Errorf("measurement unattributed records exceed snapshot limit %d", SnapshotLimitUnattributed)
	}
	if m.UnattributedMinutes < 0 || math.IsNaN(m.UnattributedMinutes) || math.IsInf(m.UnattributedMinutes, 0) {
		return fmt.Errorf("measurement unattributed_minutes must be finite and nonnegative")
	}
	if len(m.Unattributed) == 0 {
		return nil
	}
	since, err := measurementTime("since", m.Since)
	if err != nil {
		return err
	}
	until, err := measurementTime("until", m.Until)
	if err != nil {
		return err
	}
	for i, u := range m.Unattributed {
		if err := u.Validate(); err != nil {
			return fmt.Errorf("measurement unattributed[%d]: %w", i, err)
		}
		if !u.Bounded {
			continue
		}
		start, _ := measurementTime("", u.Start)
		end, _ := measurementTime("", u.End)
		if start.Before(since) || end.After(until) {
			return fmt.Errorf("measurement unattributed[%d] outside sample window", i)
		}
	}
	return nil
}

type measurementUnattributedWire MeasurementUnattributed

func (u *MeasurementUnattributed) UnmarshalJSON(data []byte) error {
	var v measurementUnattributedWire
	extra, err := decodeMeasurementObject(data, &v, v)
	if err != nil {
		return err
	}
	v.extra = extra
	*u = MeasurementUnattributed(v)
	return nil
}
func (u MeasurementUnattributed) MarshalJSON() ([]byte, error) {
	return encodeMeasurementObject(measurementUnattributedWire(u), u.extra)
}
func (u *MeasurementUnattributed) UnmarshalYAML(n *yaml.Node) error {
	return decodeMeasurementYAML(n, u)
}
func (u MeasurementUnattributed) MarshalYAML() (any, error) {
	return encodeMeasurementYAML(measurementUnattributedWire(u), u.extra)
}
