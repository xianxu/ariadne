package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// computeOwnedActualFor is the shared execution path for preview, adoption and
// correction. The resolver supplies pinned authority; the engine only measures.
func computeOwnedActualFor(ctx context.Context, root, id, milestone string, inputs *ownedActualInputs) actualResult {
	m := inputs.Snapshot
	allSources := append([]gatestate.MeasurementSource(nil), m.Sources...)
	if !inputs.Replay {
		m.Protected = nil
		for _, credit := range inputs.Protected {
			m.Protected = append(m.Protected, gatestate.MeasurementCredit{Issue: credit.Issue, Slot: credit.Slot, Start: credit.Start.Format(time.RFC3339Nano), End: credit.End.Format(time.RFC3339Nano)})
		}
	}
	res := actualResult{Issue: id, CodeHead: m.CodeHead, Window: m.Since + " → " + m.Until, Warnings: actualExclusionWarnings(inputs.Warnings)}
	fail := func(err error) actualResult { res.Status, res.Detail = actualError, err.Error(); return res }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	if err := validateProtectedSlotIdentities(inputs); err != nil {
		return fail(err)
	}
	peers, err := activetime.WindowIssues(root, m.Since, m.Until, inputs.Scope, inputs.TrackerRefs...)
	if err != nil {
		return fail(err)
	}
	if (inputs.LegacyPrefix == nil) != (m.LegacyPrefix == nil) {
		return fail(fmt.Errorf("legacy prefix runtime and retained snapshot differ"))
	}
	res.Peers = peers
	options := activetime.Options{Context: ctx, Files: inputs.Files, GitRepo: root, ExtraRefs: inputs.TrackerRefs, Scope: inputs.Scope,
		SinceISO: m.Since, UntilISO: m.Until, Issues: peers, CommitWeight: m.CommitWeight, ThresholdMin: m.GapMinutes,
		IncludeAssistant: m.IncludeAssistant, ActivityWindows: inputs.Windows, DispatchWindows: inputs.DispatchWindows, SourcePrefixes: inputs.Prefixes, SourceSlots: inputs.SourceSlots,
		ProtectedIntervals: inputs.Protected}
	if inputs.LegacyPrefix != nil {
		options.SinceISO = inputs.LegacyPrefix.Snapshot.Until
	}
	for _, q := range m.QueueWaits {
		start, err := time.Parse(time.RFC3339Nano, q.Start)
		if err != nil {
			return fail(err)
		}
		end, err := time.Parse(time.RFC3339Nano, q.End)
		if err != nil {
			return fail(err)
		}
		options.QueueIntervals = append(options.QueueIntervals, activetime.QueueInterval{Start: start, End: end, Slot: q.Slot, AttemptID: q.AttemptID})
	}
	out, err := activetime.Compute(options)
	if err != nil {
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	ownedOut := out
	if inputs.LegacyPrefix != nil {
		res.Warnings = append(res.Warnings, actualExclusionWarnings(inputs.LegacyPrefix.Warnings)...)
		prefix, pinned, err := measureLegacyPrefix(ctx, root, id, inputs.LegacyPrefix)
		if err != nil {
			return fail(err)
		}
		cutover, err := time.Parse(time.RFC3339Nano, pinned.Until)
		if err != nil {
			return fail(err)
		}
		out, err = joinLegacyActualResults(prefix, out, cutover)
		if err != nil {
			return fail(err)
		}
		m.LegacyPrefix = pinned
		ownedSince := m.Engagements[0].Start
		claimAt, err := time.Parse(time.RFC3339Nano, ownedSince)
		if err != nil {
			return fail(err)
		}
		if claimAt.After(cutover) {
			res.Warnings = append(res.Warnings, fmt.Sprintf("work between cutover %s and claim %s is unowned and not counted", pinned.Until, ownedSince))
		}
		res.Warnings = append(res.Warnings, fmt.Sprintf("legacy repository-plus-brain heuristic: %s → %s; tracker ownership: %s → %s", pinned.Since, pinned.Until, ownedSince, m.Until))
	}
	res.Status, res.Hours = statusFromResult(out, id)
	res.Segments = out.Segments
	res.Exclusions = out.Exclusions
	if res.Status != actualMeasured {
		res.Detail = "no activity from ownership-proven transcript sources"
		return res
	}
	rounds := measurementRounds(m.Cuts)
	tl, err := actualTimeline(out, milestone, rounds, m.Since, m.Until)
	if err != nil {
		return fail(err)
	}
	res.PhaseDiagnostic = actualPhaseDiagnostic(tl, milestone)
	split := splitFromTimeline(out.Segments, id, tl)
	res.Hours, res.Split = split.Build+split.Converge, &split
	for _, phase := range tl.Phases {
		if phase.Kind != activetime.Tail {
			res.CountedPhases = append(res.CountedPhases, phase)
		}
	}
	if milestone != "" {
		res.Window += " (" + milestone + "'s phases)"
	}
	if !inputs.Replay {
		contributing := map[string]bool{}
		for _, file := range ownedOut.SourceIDs {
			contributing[file] = true
		}
		for _, file := range ownedOut.QueueBindingSourceIDs {
			contributing[file] = true // queue replay still needs this prefix when its activity is zero
		}
		// A positive legacy prefix can make the combined result measurable
		// while the owned suffix has exactly observed zero elapsed activity.
		// Keep those inspected bytes so replay can distinguish zero from a
		// missing transcript population.
		if inputs.LegacyPrefix != nil && ownedOut.TotalActive == 0 && len(ownedOut.SourceIDs) == 0 {
			for _, file := range inputs.Files {
				contributing[file] = true
			}
		}
		m.Sources = retainedMeasurementSources(inputs, contributing)
		if inputs.LegacyPrefix != nil && len(m.Sources) != len(contributing) {
			return fail(fmt.Errorf("owned suffix source evidence is incomplete"))
		}

		if len(inputs.ExcludedBrain) > 0 {
			excluded := options
			excluded.Files = nil
			excluded.QueueIntervals = nil // excluded sources are a separate diagnostic population
			excluded.ProtectedIntervals = nil
			excluded.SourceSlots = nil
			excluded.ActivityWindows = nil
			excluded.DispatchWindows = nil
			if len(m.Engagements) > 0 {
				excluded.ActivityWindows = make(map[string][]activetime.ActivityWindow)
			}
			excluded.SourcePrefixes = make(map[string]activetime.SourcePrefix)
			for _, s := range inputs.ExcludedBrain {
				excluded.Files = append(excluded.Files, s.File)
				excluded.SourcePrefixes[s.File] = activetime.SourcePrefix{Bytes: s.Length, Digest: s.Digest}
				for _, interval := range m.Engagements {
					start, _ := time.Parse(time.RFC3339, interval.Start)
					end, _ := time.Parse(time.RFC3339, interval.End)
					excluded.ActivityWindows[s.File] = append(excluded.ActivityWindows[s.File], activetime.ActivityWindow{Start: start, End: end})
				}
			}
			excludedOut, err := activetime.Compute(excluded)
			if err != nil {
				return fail(fmt.Errorf("excluded brain provenance: %w", err))
			}
			m.ExcludedBrainSessions = len(excludedOut.SourceIDs)
			m.ExcludedBrainMinutes = excludedOut.TotalActive
		}
	}
	if !inputs.Replay {
		var coverageNotes []string
		m, coverageNotes = attachCarWaitCoverage(m, inputs.CarWaits, inputs.CarWaitNotes, allSources)
		res.Warnings = append(res.Warnings, carWaitCoverageWarnings(nil, coverageNotes)...)
	}
	if err := m.Validate(); err != nil {
		return fail(fmt.Errorf("measurement evidence incomplete: %w", err))
	}
	res.Inputs = &m
	res.Warnings = append(res.Warnings, carWaitCoverageWarnings(m.CarWaits, m.CarWaitNotes)...)
	res.Warnings = append(res.Warnings, fmt.Sprintf("provenance: %s; %d ownership intervals; %d admitted sessions; unbound brain excluded: %d sessions, %.2f minutes; code %s; tracker %s", m.Method, len(m.Engagements), len(m.Sources), m.ExcludedBrainSessions, m.ExcludedBrainMinutes, m.CodeHead, m.TrackerHead))
	if len(m.QueueWaits) > 0 {
		res.Warnings = append(res.Warnings, fmt.Sprintf("suite queue: %.2f raw minutes, %.2f excluded activity minutes across %d recorded waits", out.QueueAttribution.RawMinutes, out.QueueAttribution.ExcludedMinutes, len(m.QueueWaits)))
	}
	for _, warning := range out.Warnings {
		res.Warnings = append(res.Warnings, formatAttributionWarning(warning))
	}
	for _, exclusion := range out.Exclusions {
		res.Warnings = append(res.Warnings, fmt.Sprintf("excluded %.2f minutes from %s during %s → %s: already adopted by issue #%s in slot %s", exclusion.Minutes, exclusion.Source, exclusion.Start.Format(time.RFC3339Nano), exclusion.End.Format(time.RFC3339Nano), exclusion.Issue, exclusion.Slot))
	}
	return res
}

// The activity engine receives strings for persisted evidence compatibility.
// Admit only canonical values before it compares source and protected slots;
// old measurements without protection may still have absent source pins.
func validateProtectedSlotIdentities(inputs *ownedActualInputs) error {
	if len(inputs.Protected) == 0 {
		return nil
	}
	for _, credit := range inputs.Protected {
		if !parseAdoptedSlot(credit.Slot).known {
			return fmt.Errorf("protected issue %s has no canonical pinned slot", credit.Issue)
		}
	}
	for _, source := range inputs.Files {
		if !parseAdoptedSlot(inputs.SourceSlots[source]).known {
			return fmt.Errorf("source %q has no canonical pinned slot for protected intervals", source)
		}
	}
	return nil
}

func measurementRounds(cuts []gatestate.MeasurementCut) []activetime.BoundaryRound {
	var rounds []activetime.BoundaryRound
	for _, cut := range cuts {
		finalized, _ := time.Parse(time.RFC3339, cut.Finalized)
		dispatched, _ := time.Parse(time.RFC3339, cut.Dispatched)
		rounds = append(rounds, activetime.BoundaryRound{Boundary: cut.Boundary, Finalized: finalized, Dispatched: dispatched, Finalizing: cut.Finalizing})
	}
	return rounds
}

// These are warnings, never inferred corrections. Tenfold divergence and a
// half-hour difference catch the two motivating samples without tiny-task noise.
func actualEstimateWarning(actual, estimate float64) string {
	if estimate <= 0 || actual < 0 || math.IsNaN(actual) || math.IsNaN(estimate) || math.IsInf(actual, 0) || math.IsInf(estimate, 0) || math.Abs(actual-estimate) < .5 {
		return ""
	}
	ratio := actual / estimate
	if ratio < .1 || ratio > 10 {
		return fmt.Sprintf("measured actual %.2fh is %.2f× estimate %.2fh; inspect ownership/source provenance before adopting this calibration sample", actual, ratio, estimate)
	}
	return ""
}

// Discovery sees historical sessions too. Keep exclusion counts and examples
// visible without flooding the command's adoption/refusal diagnostics.
func actualExclusionWarnings(warnings []string) []string {
	var out, examples []string
	count := 0
	for _, warning := range warnings {
		if strings.HasPrefix(warning, "excluded ") && strings.HasSuffix(warning, ": no verified ownership provenance") {
			count++
			if len(examples) < 3 {
				if len(warning) > 160 {
					warning = warning[:157] + "..."
				}
				examples = append(examples, warning)
			}
		} else {
			out = append(out, warning)
		}
	}
	if count > 0 {
		out = append(out, fmt.Sprintf("%d sessions excluded: no verified ownership provenance; examples: %s", count, strings.Join(examples, "; ")))
	}
	return out
}
