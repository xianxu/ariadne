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

// cutTimes is the one rule for which rounds can cut a split and with what
// times, shared by the split and the row's metadata so they cannot disagree:
// a round that consumed no review, or whose finalize time does not parse, is
// unusable; an unparseable dispatch stamp reads as zero (missing).
func cutTimes(r gatestate.Round) (fin, disp time.Time, ok bool) {
	if r.NoCap {
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
	since, err := time.Parse(time.RFC3339, sinceISO)
	if err != nil {
		return actualSplit{}, fmt.Errorf("window start %q: %w", sinceISO, err)
	}
	until, err := time.Parse(time.RFC3339, untilISO)
	if err != nil {
		return actualSplit{}, fmt.Errorf("window end %q: %w", untilISO, err)
	}
	tl := activetime.BuildTimeline(rounds, out.IssueCommitTimes, milestone, since, until)
	if milestone != "" {
		tl = tl.For(milestone)
	}
	s := activetime.Split(out.Segments, issue, tl)
	return actualSplit{Build: s.Build / 60, Converge: s.Converge / 60, Tail: s.Tail / 60, Source: tl.Source}, nil
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
// review, so only the ledger has the finalizing round. It counts exactly the
// rounds the split can use (cutTimes).
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
