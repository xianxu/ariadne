package gatestate

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/xianxu/ariadne/pkg/workspace"
)

const (
	SnapshotLimitCarWaits     = 16
	SnapshotLimitCarWaitBytes = 8 * 1024
)

// MeasurementCarWait retains descriptive train-wait coverage. It makes no
// assertion that a source was idle and does not alter measurement arithmetic.
type MeasurementCarWait struct {
	extra        measurementUnknown `json:"-" yaml:"-"`
	Ticket       int64              `json:"ticket" yaml:"ticket"`
	Trains       []int64            `json:"trains" yaml:"trains"`
	Slot         string             `json:"slot" yaml:"slot"`
	Head         string             `json:"head" yaml:"head"`
	RecordDigest string             `json:"record_digest" yaml:"record_digest"`
	Start        string             `json:"start" yaml:"start"`
	End          string             `json:"end" yaml:"end"`
	Sources      []string           `json:"sources" yaml:"sources"`
}

// Validate checks the retained invocation identity and clipped interval shape.
// Membership in pinned measurement sources is checked by the parent snapshot.
func (w MeasurementCarWait) Validate() error {
	address, err := workspace.ParseAddress(w.Slot)
	if w.Ticket <= 0 || address.Repo == "" || err != nil || address.String() != w.Slot || !measurementHex(w.Head, 40, 64) || !measurementHex(w.RecordDigest, 64) {
		return fmt.Errorf("measurement car wait ticket, slot, head or digest is invalid")
	}
	if len(w.Trains) == 0 {
		return fmt.Errorf("measurement car wait trains must not be empty")
	}
	trains := map[int64]bool{}
	for _, train := range w.Trains {
		if train <= 0 || train == w.Ticket || trains[train] {
			return fmt.Errorf("measurement car wait trains must be positive, distinct and not self")
		}
		trains[train] = true
	}
	start, err := measurementTime("car wait start", w.Start)
	if err != nil {
		return err
	}
	end, err := measurementTime("car wait end", w.End)
	if err != nil {
		return err
	}
	if !end.After(start) {
		return fmt.Errorf("measurement car wait end must follow start")
	}
	if len(w.Sources) == 0 {
		return fmt.Errorf("measurement car wait sources must not be empty")
	}
	sources := map[string]bool{}
	for _, ref := range w.Sources {
		harness, id, ok := strings.Cut(ref, "/")
		if !ok || harness == "" || id == "" || strings.ContainsAny(ref, "\x00\r\n\t ") || sources[ref] {
			return fmt.Errorf("measurement car wait sources must be canonical distinct harness/ID identities")
		}
		sources[ref] = true
	}
	return nil
}

// ValidateCarWaitCoverage bounds optional diagnostics independently of the
// accounting snapshot. Fresh callers validate a candidate copy and retain the
// original snapshot if optional coverage cannot fit.
func (m MeasurementInputs) ValidateCarWaitCoverage() error {
	if len(m.CarWaits) > SnapshotLimitCarWaits {
		return fmt.Errorf("measurement car waits exceed snapshot limit %d", SnapshotLimitCarWaits)
	}
	if len(m.CarWaits) == 0 && len(m.CarWaitNotes) == 0 {
		return nil
	}
	data, err := json.Marshal(struct {
		Waits []MeasurementCarWait `json:"car_waits,omitempty"`
		Notes []string             `json:"car_wait_notes,omitempty"`
	}{m.CarWaits, m.CarWaitNotes})
	if err != nil {
		return fmt.Errorf("measurement car coverage does not serialize: %w", err)
	}
	if len(data) > SnapshotLimitCarWaitBytes {
		return fmt.Errorf("measurement car coverage exceeds %d bytes", SnapshotLimitCarWaitBytes)
	}
	for _, note := range m.CarWaitNotes {
		if strings.TrimSpace(note) == "" || note != strings.TrimSpace(note) {
			return fmt.Errorf("measurement car wait note must be nonempty and trimmed")
		}
	}
	since, err := measurementTime("since", m.Since)
	if err != nil {
		return err
	}
	until, err := measurementTime("until", m.Until)
	if err != nil {
		return err
	}
	pinned := map[string][]MeasurementSource{}
	for _, s := range m.Sources {
		pinned[s.Harness+"/"+s.ID] = append(pinned[s.Harness+"/"+s.ID], s)
	}
	type intervalKey struct {
		ticket     int64
		start, end time.Time
	}
	seen := map[intervalKey]bool{}
	for i, w := range m.CarWaits {
		if err := w.Validate(); err != nil {
			return fmt.Errorf("measurement car_waits[%d]: %w", i, err)
		}
		start, _ := time.Parse(time.RFC3339Nano, w.Start)
		end, _ := time.Parse(time.RFC3339Nano, w.End)
		key := intervalKey{w.Ticket, start.UTC(), end.UTC()}
		if seen[key] {
			return fmt.Errorf("measurement duplicate car wait ticket/interval")
		}
		seen[key] = true
		if start.Before(since) || end.After(until) {
			return fmt.Errorf("measurement car_waits[%d] outside sample window", i)
		}
		for _, ref := range w.Sources {
			sources := pinned[ref]
			if len(sources) != 1 {
				return fmt.Errorf("measurement car_waits[%d] source %q is not uniquely pinned", i, ref)
			}
			s := sources[0]
			if s.Slot != w.Slot {
				return fmt.Errorf("measurement car_waits[%d] source slot does not match waiter slot", i)
			}
			if strings.Contains(s.Harness, "/") || s.Harness == "" || s.ID == "" {
				return fmt.Errorf("measurement car_waits[%d] source identity is ambiguous", i)
			}
			contained := false
			for _, window := range s.Windows {
				a, errA := measurementTime("source window start", window.Start)
				b, errB := measurementTime("source window end", window.End)
				if errA == nil && errB == nil && !start.Before(a) && !end.After(b) {
					contained = true
					break
				}
			}
			if !contained {
				return fmt.Errorf("measurement car_waits[%d] outside pinned source window", i)
			}
		}
	}
	return nil
}
