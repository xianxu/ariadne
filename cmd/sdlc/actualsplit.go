package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/activetime"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// actualSplit is a measured actual divided at its boundary reviews (#234):
// Build up to each boundary's first review dispatch, Converge from there to the
// boundary's last #N commit. Hours are Build+Converge; Tail is the verification
// after a finalized boundary's last commit, which no actual counts. Source names
// the evidence for the cuts (activetime.Split*).
type actualSplit struct {
	Build, Converge, Tail float64 // hours
	Source                string
}

// issueBoundaryLedger reads the issue's boundary gate ledger: the live one
// under plans (the caller's plans dir, the same one the close persists rounds
// to; "" for the default), else the archived one under workshop/history/plans
// (an issue measured after it landed). found is false when the issue has none;
// a ledger that exists but cannot be read is an error, never "no rounds".
func issueBoundaryLedger(repoTop, plans string, id int) (l gatestate.Ledger, found bool, err error) {
	live := plans
	if live == "" {
		live = plansDir()
	}
	if !filepath.IsAbs(live) {
		live = filepath.Join(repoTop, live)
	}
	for _, dir := range []string{live, filepath.Join(repoTop, "workshop", "history", "plans")} {
		matches, gerr := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%06d-*-%s.md", id, boundaryGateSuffix)))
		if gerr != nil || len(matches) == 0 {
			continue
		}
		issueFileName := strings.TrimSuffix(filepath.Base(matches[0]), "-"+boundaryGateSuffix+".md") + ".md"
		l, err = readBoundaryGateLedger(dir, issueFileName, id)
		return l, err == nil, err
	}
	return gatestate.Ledger{}, false, nil
}

// splitRounds maps the ledger's boundary rounds to the split's view. Rounds
// that consumed no review (the plan-gate seed, a dispatch that never started)
// carry no cut, and neither does a round whose finalize time does not parse. A
// dispatch stamp that does not parse reads as missing, so the split falls back
// to the finalize time and says so in its source.
func splitRounds(l gatestate.Ledger) []activetime.BoundaryRound {
	var out []activetime.BoundaryRound
	for _, r := range l.Rounds {
		fin, disp, ok := cutTimes(r)
		if !ok {
			continue
		}
		out = append(out, activetime.BoundaryRound{
			Boundary: r.Boundary, Dispatched: disp, Finalized: fin, Finalizing: r.Reviewed != "",
		})
	}
	return out
}

// cutTimes is the ONE rule for which rounds can cut a split and with what
// times, shared by the split and the row's metadata so they cannot disagree;
// fresh Plan selection can then omit retired unfinalized boundaries while metadata
// retains consumed review history. Unusable: a round that consumed no
// review, one that belongs to every boundary (gatestate.BoundaryAll: no cut
// of its own), or one whose finalize time does not parse. An unparseable
// dispatch stamp reads as zero (missing).
func cutTimes(r gatestate.Round) (fin, disp time.Time, ok bool) {
	if r.NoCap || r.Boundary == gatestate.BoundaryAll {
		return fin, disp, false
	}
	fin, err := time.Parse(time.RFC3339, r.Timestamp)
	if err != nil {
		return fin, disp, false
	}
	disp, _ = time.Parse(time.RFC3339, r.Dispatched)
	return fin, disp, true
}

// splitActual divides a measured result for issue at its boundary reviews.
// milestone "" measures the whole issue; "Mx" keeps only Mx's own phases, the
// increment milestone-close records (E1).
func splitActual(out activetime.Result, issue, milestone string, rounds []activetime.BoundaryRound, sinceISO, untilISO string) (actualSplit, error) {
	tl, err := actualTimeline(out, milestone, rounds, sinceISO, untilISO)
	if err != nil {
		return actualSplit{}, err
	}
	return splitFromTimeline(out.Segments, issue, tl), nil
}

func splitFromTimeline(segments []activetime.Segment, issue string, tl activetime.Timeline) actualSplit {
	s := activetime.Split(segments, issue, tl)
	return actualSplit{Build: s.Build / 60, Converge: s.Converge / 60, Tail: s.Tail / 60, Source: tl.Source}
}

func actualTimeline(out activetime.Result, milestone string, rounds []activetime.BoundaryRound, sinceISO, untilISO string) (activetime.Timeline, error) {
	since, err := time.Parse(time.RFC3339, sinceISO)
	if err != nil {
		return activetime.Timeline{}, fmt.Errorf("window start %q: %w", sinceISO, err)
	}
	until, err := time.Parse(time.RFC3339, untilISO)
	if err != nil {
		return activetime.Timeline{}, fmt.Errorf("window end %q: %w", untilISO, err)
	}
	tl := activetime.BuildTimeline(rounds, out.IssueCommitTimes, milestone, since, until)
	if milestone != "" {
		tl = tl.For(milestone)
	}
	return tl, nil
}

// splitMeta is what a calibration row records about the boundary reviews
// (#234): how many rounds ran, which recipes, and whether every one carries a
// dispatch stamp.
type splitMeta struct {
	Rounds  int
	Recipe  string
	Stamped bool
}

// rowSplitMeta reads it from the boundary gate ledger AS PERSISTED when the row
// is written (PQ-1): the measurement runs before this close dispatches its
// review, so only the ledger has the finalizing round. It counts usable consumed
// rounds (cutTimes), including history retired from fresh measurement cuts.
// An empty recipe on a boundary round is the full review (pre-#231).
func rowSplitMeta(l gatestate.Ledger) splitMeta {
	meta := splitMeta{Stamped: true}
	seen := map[string]bool{}
	var recipes []string
	for _, r := range l.Rounds {
		_, disp, ok := cutTimes(r)
		if !ok {
			continue
		}
		meta.Rounds++
		if disp.IsZero() {
			meta.Stamped = false
		}
		recipe := r.Recipe
		if recipe == "" {
			recipe = "milestone-review"
		}
		if !seen[recipe] {
			seen[recipe] = true
			recipes = append(recipes, recipe)
		}
	}
	meta.Recipe = strings.Join(recipes, "+")
	return meta
}

// rowSplitSource is the row's split_source: manual for a typed --actual (no
// measurement, so no split), no-review when no boundary review ran, finalize
// when some cut had to fall back to a round's finalize time, else dispatch.
func rowSplitSource(measured *actualSplit, meta splitMeta) string {
	switch {
	case measured == nil:
		return "manual"
	case meta.Rounds == 0:
		return activetime.SplitNoReview
	case !meta.Stamped || measured.Source == activetime.SplitFinalize:
		return activetime.SplitFinalize
	default:
		return activetime.SplitDispatch
	}
}

// actualPhaseDiagnostic describes the exact selected timeline used by Split.
func actualPhaseDiagnostic(tl activetime.Timeline, milestone string) string {
	label := "issue"
	if milestone != "" {
		label = milestone
	}
	if len(tl.Phases) == 0 {
		return fmt.Sprintf("effective %s phases: empty; no selected phase interval (inspect retained executable open boundaries)", label)
	}
	return fmt.Sprintf("effective %s phases: %s → %s (%d phase(s); tail remains uncounted)", label, tl.Phases[0].Start.Format(time.RFC3339Nano), tl.Phases[len(tl.Phases)-1].End.Format(time.RFC3339Nano), len(tl.Phases))
}

// selectFreshSplitRounds changes only newly derived cuts. Finalized history
// always wins over current Plan membership; saved snapshots never call this.
func selectFreshSplitRounds(rounds []activetime.BoundaryRound, milestones []string) ([]activetime.BoundaryRound, []string) {
	keep := map[string]bool{"": true}
	for _, m := range milestones {
		keep[m] = true
	}
	for _, r := range rounds {
		if r.Finalizing {
			keep[r.Boundary] = true
		}
	}
	var selected []activetime.BoundaryRound
	var retired []string
	seen := map[string]bool{}
	for _, r := range rounds {
		if keep[r.Boundary] {
			selected = append(selected, r)
		} else if !seen[r.Boundary] {
			retired = append(retired, r.Boundary)
			seen[r.Boundary] = true
		}
	}
	return selected, retired
}
