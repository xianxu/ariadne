package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// joinLegacyActualResults admits two complete measurements on opposite sides
// of one immutable cutover. Clipping makes their boundary half-open even when
// both engine inputs include the event at cutover.
func joinLegacyActualResults(prefix, suffix activetime.Result, cutover time.Time) (activetime.Result, error) {
	if cutover.IsZero() {
		return activetime.Result{}, fmt.Errorf("legacy prefix cutover is missing")
	}
	out := suffix
	out.PerIssue = map[string]float64{}
	out.Segments = nil
	out.TotalActive = 0
	out.IssueCommitTimes = nil
	out.SourceIDs = nil
	for i, component := range []activetime.Result{prefix, suffix} {
		label := []string{"legacy prefix", "tracker ownership"}[i]
		if component.Status != activetime.Measured && component.Status != activetime.EmptyWindow {
			return activetime.Result{}, fmt.Errorf("%s has unavailable telemetry (status %d)", label, component.Status)
		}
		for _, segment := range component.Segments {
			if segment.Start.IsZero() || segment.End.Before(segment.Start) || math.IsNaN(segment.Active) || math.IsInf(segment.Active, 0) || segment.Active < 0 {
				return activetime.Result{}, fmt.Errorf("%s has invalid segment", label)
			}
			for _, minutes := range segment.Alloc {
				if math.IsNaN(minutes) || math.IsInf(minutes, 0) || minutes < 0 {
					return activetime.Result{}, fmt.Errorf("%s has invalid allocation", label)
				}
			}
			start, end := segment.Start, segment.End
			if i == 0 && end.After(cutover) {
				end = cutover
			}
			if i == 1 && start.Before(cutover) {
				start = cutover
			}
			if !start.Before(end) {
				continue
			}
			ratio := end.Sub(start).Seconds() / segment.End.Sub(segment.Start).Seconds()
			segment.Start, segment.End, segment.Active = start, end, segment.Active*ratio
			alloc := map[string]float64{}
			for issue, minutes := range segment.Alloc {
				alloc[issue] = minutes * ratio
			}
			segment.Alloc = alloc
			out.Segments = append(out.Segments, segment)
		}
		out.IssueCommitTimes = append(out.IssueCommitTimes, component.IssueCommitTimes...)
		out.SourceIDs = append(out.SourceIDs, component.SourceIDs...)
	}
	sort.Slice(out.Segments, func(i, j int) bool {
		a, b := out.Segments[i], out.Segments[j]
		if !a.Start.Equal(b.Start) {
			return a.Start.Before(b.Start)
		}
		if !a.End.Equal(b.End) {
			return a.End.Before(b.End)
		}
		return a.Source < b.Source
	})
	for _, segment := range out.Segments {
		out.TotalActive += segment.Active
		for issue, minutes := range segment.Alloc {
			out.PerIssue[issue] += minutes
		}
	}
	sort.Slice(out.IssueCommitTimes, func(i, j int) bool { return out.IssueCommitTimes[i].Before(out.IssueCommitTimes[j]) })
	times := out.IssueCommitTimes[:0]
	for _, at := range out.IssueCommitTimes {
		if len(times) == 0 || !times[len(times)-1].Equal(at) {
			times = append(times, at)
		}
	}
	out.IssueCommitTimes = times
	sort.Strings(out.SourceIDs)
	ids := out.SourceIDs[:0]
	for _, id := range out.SourceIDs {
		if len(ids) == 0 || ids[len(ids)-1] != id {
			ids = append(ids, id)
		}
	}
	out.SourceIDs = ids
	out.NumEvents += prefix.NumEvents
	out.NumCommits += prefix.NumCommits
	out.Warnings = append(append([]activetime.AttributionWarning(nil), prefix.Warnings...), suffix.Warnings...)
	out.Status = activetime.EmptyWindow
	if prefix.Status == activetime.Measured || suffix.Status == activetime.Measured {
		out.Status = activetime.Measured
	}
	return out, nil
}

// measureLegacyPrefix uses the existing legacy attribution engine with pinned
// code and transcript bytes. It never borrows modern ownership or queue credit.
func measureLegacyPrefix(ctx context.Context, root, id string, inputs *ownedActualInputs) (activetime.Result, *gatestate.MeasurementInputs, error) {
	m := inputs.Snapshot
	fail := func(err error) (activetime.Result, *gatestate.MeasurementInputs, error) {
		return activetime.Result{}, nil, fmt.Errorf("legacy prefix: %w", err)
	}
	if m.Method != gatestate.MeasurementMethodLegacyV1 || inputs.LegacyPrefix != nil || len(inputs.Protected) > 0 || len(m.QueueWaits) > 0 || len(m.Engagements) > 0 || m.TrackerHead != "" {
		return fail(fmt.Errorf("unexpected ownership/queue authority"))
	}
	for _, slot := range inputs.SourceSlots {
		if slot != "" {
			return fail(fmt.Errorf("unexpected ownership source slot"))
		}
	}
	if len(inputs.Files) == 0 {
		return fail(fmt.Errorf("no exact transcript evidence is available"))
	}
	for _, file := range inputs.Files {
		prefix, ok := inputs.Prefixes[file]
		if !ok || prefix.Bytes <= 0 || prefix.Digest == "" {
			return fail(fmt.Errorf("source %q has no exact retained prefix", file))
		}
	}
	if !inputs.Replay {
		m.Sources = nil
	}
	scope := inputs.Scope
	scope.Issue = id
	if err := m.Validate(); err != nil {
		return fail(err)
	}
	peers, err := activetime.WindowIssues(root, m.Since, m.Until, scope, inputs.TrackerRefs...)
	if err != nil {
		return fail(err)
	}
	out, err := activetime.Compute(activetime.Options{Context: ctx, Files: inputs.Files, GitRepo: root, ExtraRefs: inputs.TrackerRefs, Scope: scope, SinceISO: m.Since, UntilISO: m.Until, Issues: peers, CommitWeight: m.CommitWeight, ThresholdMin: m.GapMinutes, IncludeAssistant: m.IncludeAssistant, ActivityWindows: inputs.Windows, SourcePrefixes: inputs.Prefixes})
	if err != nil {
		return fail(err)
	}
	if out.Status != activetime.Measured && out.Status != activetime.EmptyWindow {
		return fail(fmt.Errorf("transcript telemetry is unavailable (status %d)", out.Status))
	}
	if !inputs.Replay {
		contributing := map[string]bool{}
		for _, file := range out.SourceIDs {
			contributing[file] = true
		}
		if out.TotalActive == 0 && len(out.SourceIDs) == 0 {
			for _, file := range inputs.Files {
				contributing[file] = true
			}
		}
		m.Sources = retainedMeasurementSources(inputs, contributing)
		if len(m.Sources) != len(contributing) {
			return fail(fmt.Errorf("contributing source evidence is incomplete"))
		}
	}
	if len(m.Sources) == 0 {
		return fail(fmt.Errorf("measurement has no retained source evidence"))
	}
	if err := m.Validate(); err != nil {
		return fail(err)
	}
	return out, &m, nil
}

func retainedMeasurementSources(inputs *ownedActualInputs, contributing map[string]bool) []gatestate.MeasurementSource {
	var sources []gatestate.MeasurementSource
	for _, s := range inputs.Sessions {
		if !contributing[s.File] {
			continue
		}
		item := gatestate.MeasurementSource{Harness: s.Harness, ID: s.ID, CWD: s.CWD, ParentID: s.ParentID, Digest: s.Digest, Started: s.Started.Format(time.RFC3339Nano), Spawned: s.Spawned, Bytes: s.Length, Slot: inputs.SourceSlots[s.File]}
		for _, w := range inputs.Windows[s.File] {
			item.Windows = append(item.Windows, gatestate.MeasurementInterval{Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano)})
		}
		for _, w := range inputs.DispatchWindows[s.File] {
			item.DispatchWindows = append(item.DispatchWindows, gatestate.MeasurementInterval{Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano)})
		}
		sources = append(sources, item)
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Harness+":"+sources[i].ID < sources[j].Harness+":"+sources[j].ID
	})
	return sources
}
