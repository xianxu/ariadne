package gatestate

import (
	"fmt"
	"strings"
)

// Shared test fixtures for the gatestate package. Written once, here, because the ledger,
// decide, and render tests all build the same shapes — and the ID contract between them is
// load-bearing: the findings a fixture raises must carry the same PQ-<n> sequence
// AssignIDs would produce, or a `dispose("PQ-1", …)` elsewhere silently references nothing.
//
// ledgerWith owns the numbering, so there is no shared counter to reset and no evaluation-
// order subtlety: findings are stamped PQ-1, PQ-2, … in first-raised order across the
// whole ledger, at assembly time.

const (
	testTimestamp = "2026-07-29T10:00:00Z"
	testAgent     = "claude"
	testPrefix    = "PQ"
)

// findings parses "Severity/title" specs into Findings with NO ids — ledgerWith stamps them.
//
//	findings("Critical/seam in wrong layer", "Minor/naming")
func findings(specs ...string) []Finding {
	out := make([]Finding, 0, len(specs))
	for _, spec := range specs {
		sev, title, ok := strings.Cut(spec, "/")
		if !ok {
			panic("gatestate test: finding spec must be `Severity/title`, got " + spec)
		}
		out = append(out, Finding{Severity: sev, Title: title})
	}
	return out
}

// dispose builds Dispositions from id/state pairs:
//
//	dispose("PQ-1", "addressed", "PQ-2", "withdrawn")
func dispose(pairs ...string) []Disposition {
	if len(pairs)%2 != 0 {
		panic("gatestate test: dispose() needs id/state pairs")
	}
	out := make([]Disposition, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, Disposition{ID: pairs[i], State: pairs[i+1]})
	}
	return out
}

// round builds an UNSTAMPED round; ledgerWith assigns the finding ids.
func round(n int, disp []Disposition, fs []Finding) Round {
	r := Round{N: n, Timestamp: testTimestamp, Agent: testAgent}
	for _, d := range disp {
		d.Round = n
		r.Dispositions = append(r.Dispositions, d)
	}
	r.New = append(r.New, fs...)
	return r
}

// ledgerWith assembles a plan-quality ledger, stamping finding ids in first-raised order
// so they match what AssignIDs would have produced for the same sequence.
func ledgerWith(rs ...Round) Ledger {
	l := Ledger{Gate: "plan-quality", IssueNum: 187, IDPrefix: testPrefix}
	seq := 0
	for _, r := range rs {
		stamped := r
		stamped.New = nil
		for _, f := range r.New {
			seq++
			f.ID = fmt.Sprintf("%s-%d", testPrefix, seq)
			f.Round = r.N
			stamped.New = append(stamped.New, f)
		}
		l.Rounds = append(l.Rounds, stamped)
	}
	return l
}

// ids extracts finding IDs, for readable failure messages.
func ids(fs []Finding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.ID)
	}
	return out
}
