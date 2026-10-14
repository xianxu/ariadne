package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// computeOwnedActualFor is the shared execution path for preview, adoption and
// correction. The resolver supplies pinned authority; the engine only measures.
func computeOwnedActualFor(ctx context.Context, root, id, milestone string, inputs *ownedActualInputs) actualResult {
	m := inputs.Snapshot
	res := actualResult{Issue: id, Window: m.Since + " → " + m.Until, Warnings: actualExclusionWarnings(inputs.Warnings)}
	fail := func(err error) actualResult { res.Status, res.Detail = actualError, err.Error(); return res }
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	peers, err := activetime.WindowIssues(root, m.Since, m.Until, inputs.Scope, inputs.TrackerRefs...)
	if err != nil {
		return fail(err)
	}
	res.Peers = peers
	options := activetime.Options{Context: ctx, Files: inputs.Files, GitRepo: root, ExtraRefs: inputs.TrackerRefs, Scope: inputs.Scope,
		SinceISO: m.Since, UntilISO: m.Until, Issues: peers, CommitWeight: m.CommitWeight, ThresholdMin: m.GapMinutes,
		IncludeAssistant: m.IncludeAssistant, ActivityWindows: inputs.Windows, SourcePrefixes: inputs.Prefixes}
	out, err := activetime.Compute(options)
	if err != nil {
		return fail(err)
	}
	if err := ctx.Err(); err != nil {
		return fail(err)
	}
	res.Status, res.Hours = statusFromResult(out, id)
	if res.Status != actualMeasured {
		res.Detail = "no activity from ownership-proven transcript sources"
		return res
	}
	rounds := measurementRounds(m.Cuts)
	split, err := splitActual(out, id, milestone, rounds, m.Since, m.Until)
	if err != nil {
		return fail(err)
	}
	res.Hours, res.Split = split.Build+split.Converge, &split
	if milestone != "" {
		res.Window += " (" + milestone + "'s phases)"
	}
	if !inputs.Replay {
		contributing := map[string]bool{}
		for _, file := range out.SourceIDs {
			contributing[file] = true
		}
		m.Sources = nil
		for _, s := range inputs.Sessions {
			if !contributing[s.File] {
				continue
			}
			item := gatestate.MeasurementSource{Harness: s.Harness, ID: s.ID, CWD: s.CWD, ParentID: s.ParentID, Digest: s.Digest,
				Started: s.Started.Format(time.RFC3339Nano), Spawned: s.Spawned, Bytes: s.Length}
			for _, w := range inputs.Windows[s.File] {
				item.Windows = append(item.Windows, gatestate.MeasurementInterval{Start: w.Start.Format(time.RFC3339Nano), End: w.End.Format(time.RFC3339Nano)})
			}
			m.Sources = append(m.Sources, item)
		}
		sort.Slice(m.Sources, func(i, j int) bool {
			a, b := m.Sources[i], m.Sources[j]
			return a.Harness+":"+a.ID < b.Harness+":"+b.ID
		})
		if len(inputs.ExcludedBrain) > 0 {
			excluded := options
			excluded.Files = nil
			excluded.ActivityWindows = nil
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
	if err := m.Validate(); err != nil {
		return fail(fmt.Errorf("measurement evidence incomplete: %w", err))
	}
	res.Inputs = &m
	res.Warnings = append(res.Warnings, fmt.Sprintf("provenance: %s; %d ownership intervals; %d admitted sessions; unbound brain excluded: %d sessions, %.2f minutes; code %s; tracker %s", m.Method, len(m.Engagements), len(m.Sources), m.ExcludedBrainSessions, m.ExcludedBrainMinutes, m.CodeHead, m.TrackerHead))
	for _, warning := range out.Warnings {
		res.Warnings = append(res.Warnings, formatAttributionWarning(warning))
	}
	return res
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
