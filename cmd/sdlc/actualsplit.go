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
// under the plans dir, else the archived one under workshop/history/plans (an
// issue measured after it landed). found is false when the issue has none; a
// ledger that exists but cannot be read is an error, never "no rounds".
func issueBoundaryLedger(repoTop string, id int) (l gatestate.Ledger, found bool, err error) {
	live := plansDir()
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
		if r.NoCap {
			continue
		}
		fin, err := time.Parse(time.RFC3339, r.Timestamp)
		if err != nil {
			continue
		}
		disp, _ := time.Parse(time.RFC3339, r.Dispatched)
		out = append(out, activetime.BoundaryRound{
			Boundary: r.Boundary, Dispatched: disp, Finalized: fin, Finalizing: r.Reviewed != "",
		})
	}
	return out
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
