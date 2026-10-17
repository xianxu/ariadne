package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/tracker"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/transcripts"
)

// selectOwnedQueueAttempts admits a queue wait only when a single pinned
// engagement owns its slot for the entire original wait. A partial measurement
// window clips the selected interval, never the ownership proof.
func selectOwnedQueueAttempts(attempts []SuiteAttempt, engagements []tracker.EngagementWindow, repository string, since, until time.Time) ([]gatestate.MeasurementQueueWait, []string, error) {
	var waits []gatestate.MeasurementQueueWait
	var warnings []string
	seen := map[string]bool{}
	for _, a := range attempts {
		if a.Repo != strings.TrimPrefix(repository, "github.com/") || a.Check != "suite" || a.MainBaseline {
			continue
		}
		if seen[a.ID] {
			return nil, nil, fmt.Errorf("duplicate suite attempt %s", a.ID)
		}
		seen[a.ID] = true
		if a.QueueSeconds == nil {
			// Acquisition before the window proves the wait is irrelevant. An
			// acquisition after it does not: the missing duration could have
			// queued across the window's end.
			if !a.StartedAt.Before(since) {
				warnings = append(warnings, fmt.Sprintf("suite attempt %s has unknown queue provenance", a.ID))
			}
			continue
		}
		if a.QueuedAt == nil || a.AcquiredAt == nil {
			// StartedAt follows acquisition and an unrecorded amount of setup.
			// Reconstructing a queue start from it cannot prove non-overlap,
			// even when the apparent wait starts after release or measurement.
			if a.StartedAt.Before(since) {
				continue
			}
			relevant := false
			for _, e := range engagements {
				if a.StartedAt.Before(e.Start) {
					continue
				}
				for _, owner := range e.Owners {
					if a.Slot == "" || owner.Workspace == a.Slot {
						relevant = true
						break
					}
				}
				if relevant {
					break
				}
			}
			if relevant {
				warnings = append(warnings, fmt.Sprintf("legacy suite attempt %s lacks exact queue endpoints; queue idle is unknown and this sample is not calibration evidence", a.ID))
			}
			continue
		}
		start, end := *a.QueuedAt, *a.AcquiredAt
		if !end.After(since) || !start.Before(until) || !end.After(start) {
			continue
		}
		if a.Slot == "" {
			for _, e := range engagements {
				if start.Before(e.End) && end.After(e.Start) && len(e.Owners) > 0 {
					warnings = append(warnings, fmt.Sprintf("suite attempt %s has unknown slot provenance", a.ID))
					break
				}
			}
			continue
		}
		owners, overlaps := 0, 0
		for _, e := range engagements {
			for _, owner := range e.Owners {
				if owner.Workspace != a.Slot || !start.Before(e.End) || !end.After(e.Start) {
					continue
				}
				overlaps++
				if !start.Before(e.Start) && !end.After(e.End) {
					owners++
				}
			}
		}
		if overlaps == 0 {
			continue
		} // other slots and unowned gaps are not this issue's waits
		if owners != 1 {
			return nil, nil, fmt.Errorf("suite attempt %s has %d matching pinned slot engagements", a.ID, owners)
		}
		selectedStart, selectedEnd := start, end
		if selectedStart.Before(since) {
			selectedStart = since
		}
		if selectedEnd.After(until) {
			selectedEnd = until
		}
		if !selectedEnd.After(selectedStart) {
			continue
		}
		canonical, err := json.Marshal(a)
		if err != nil {
			return nil, nil, err
		}
		digest := sha256.Sum256(canonical)
		q := gatestate.MeasurementQueueWait{AttemptID: a.ID, Slot: a.Slot, RecordDigest: fmt.Sprintf("%x", digest), Start: selectedStart.UTC().Format(time.RFC3339Nano), End: selectedEnd.UTC().Format(time.RFC3339Nano), QueueSeconds: selectedEnd.Sub(selectedStart).Seconds(), RawSeconds: *a.QueueSeconds, Exact: true}
		if q.QueueSeconds > q.RawSeconds && q.QueueSeconds-q.RawSeconds <= 1e-6 {
			q.RawSeconds = q.QueueSeconds
		}
		if err := q.Validate(); err != nil {
			return nil, nil, err
		}
		waits = append(waits, q)
	}
	sort.Slice(waits, func(i, j int) bool {
		return waits[i].Start < waits[j].Start || waits[i].Start == waits[j].Start && waits[i].AttemptID < waits[j].AttemptID
	})
	return waits, warnings, nil
}

// ownedQueueSourceSlots derives source identity from the same pinned owner
// windows used to admit transcripts. Ambiguous cross-slot sessions stay unbound.
func ownedQueueSourceSlots(sessions []transcripts.SessionEvidence, engagements []tracker.EngagementWindow) map[string]string {
	out := map[string]string{}
	for _, s := range sessions {
		slots := map[string]bool{}
		for _, e := range engagements {
			if s.Started.After(e.End) {
				continue
			}
			for _, owner := range e.Owners {
				if filepath.Clean(owner.Worktree) == filepath.Clean(s.CWD) && owner.Workspace != "" {
					slots[owner.Workspace] = true
				}
			}
		}
		if len(slots) == 1 {
			for slot := range slots {
				out[s.File] = slot
			}
		}
	}
	return out
}
