package gatestate

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// DefaultRoundCap is the round after which only hard-blocking findings (Critical) refuse
// the gate — ariadne#187 A3. Three rounds is where the pair#127 postmortem showed the
// reviewer had stopped finding substance and started descending severity levels on a plan
// that kept improving.
const DefaultRoundCap = 3

// Decision is the gate's answer: block or not, the reason to print, and the counts the
// close-time metrics read back.
type Decision struct {
	Block        bool
	Reason       string
	OpenBlocking []Finding // open findings that still block (post-cap: hard-blocking only)
	Demoted      []Finding // open blocking findings the round cap demoted to advisory
	OpenMinor    int
	Rounds       int
	CapReached   bool
	// Disputed are the findings settled by an implementer dispute rather than a fix
	// (ariadne#202), and Raised the findings in scope — the pair the operator line reports,
	// so a gate whose findings were mostly disputed away reads as such.
	Disputed []Finding
	Raised   int
}

// Decide is the gate's pure decision over the accumulated ledger.
//
// Block iff some finding is still OPEN at a blocking severity. This is the mechanic that
// makes the gate converge: a judge that raises a fresh Minor every round can no longer
// cost a round-trip, and a judge that has disposed its own prior findings sees the gate
// open. Compare the pre-#187 behavior, where the decision was read off the LLM's verdict
// token and every fresh reviewer re-derived an absolute bar.
//
// Past `roundCap` rounds, only hard-blocking severities refuse. Findings demoted there are
// recorded in the ledger and reported to the operator — and the BOUNDARY REVIEW picks them
// up, which is what makes the demotion safe rather than a silent loss. Since ariadne#194
// that pickup is a SEED into the boundary gate's own ledger (seedFromPlanGate), not an
// instruction telling the reviewer to read this file itself: one carry-forward mechanism,
// one id namespace the reviewer can actually dispose of.
//
// `roundCap` is spelled out rather than `cap` so it doesn't shadow the builtin.
func Decide(l Ledger, roundCap int) Decision { return DecideScoped(l, l, roundCap) }

// DecideScoped separates the two questions Decide answers, because they want different
// scopes (ariadne#194 M3 review BR-37):
//
//   - `capScope` bounds the ROUND CAP — per boundary, or an issue with three milestones
//     arrives at the whole-issue close already past the cap and demotes every Important
//     on its first round.
//   - `openScope` supplies the OPEN FINDINGS — the whole issue at the final boundary,
//     because the whole-issue close is the last gate before publish and a finding left
//     undisposed at a milestone otherwise strands: its boundary has closed, so nothing
//     will ever look at it again.
//
// The rule this encodes: scope per boundary only what the round cap needs; every other
// read wants the full issue.
func DecideScoped(capScope, openScope Ledger, roundCap int) Decision {
	l := openScope
	if roundCap <= 0 {
		roundCap = DefaultRoundCap
	}
	m := vocab.Finding()
	// CountedRounds, not len(l.Rounds): a seed round and a dispatch that never started
	// invoked no reviewer, and the cap is about reviewer invocations (ariadne#194 M2).
	counted := CountedRounds(capScope)
	d := Decision{Rounds: counted, CapReached: counted > roundCap}

	appealed := map[string]bool{}
	open := map[string]bool{}
	for _, f := range OpenFindings(l) {
		open[f.ID] = true
	}
	for _, ld := range LiveDisputes(l) {
		if open[ld.Finding.ID] {
			appealed[ld.Finding.ID] = true
		} else {
			d.Disputed = append(d.Disputed, ld.Finding)
		}
	}
	d.Raised = len(findingByID(l))

	for _, f := range OpenFindings(l) {
		switch {
		case !m.Blocks(f.Severity):
			d.OpenMinor++
		case d.CapReached && !m.BlocksPastCap(f.Severity):
			d.Demoted = append(d.Demoted, f)
		default:
			d.OpenBlocking = append(d.OpenBlocking, f)
		}
	}

	if len(d.OpenBlocking) == 0 {
		d.Reason = fmt.Sprintf("no open blocking findings after %d round(s)", d.Rounds)
		if len(d.Demoted) > 0 {
			d.Reason += fmt.Sprintf("; %d finding(s) recorded but not blocking (round cap %d reached)", len(d.Demoted), roundCap)
		}
		if d.OpenMinor > 0 {
			d.Reason += fmt.Sprintf("; %d advisory finding(s) recorded for the close review", d.OpenMinor)
		}
		return d
	}

	d.Block = true
	var b strings.Builder
	fmt.Fprintf(&b, "%d open blocking finding(s):", len(d.OpenBlocking))
	for _, f := range d.OpenBlocking {
		fmt.Fprintf(&b, "\n  [%s] %s — %s", f.ID, f.Severity, f.Title)
		if appealed[f.ID] {
			b.WriteString(" (disputed — appeal-only, so it blocks until the judge answers)")
		}
	}

	d.Reason = b.String()
	return d
}

// DisputeLine is the operator-facing count of findings settled by dispute rather than a
// fix, or "" when there are none (ariadne#202). Loud by design: every structure needs a
// bypass, and using it must be announced. Kept out of Reason so the one reporter,
// stampAndPersist, prints it once and at warning level whether or not the gate passed.
// Pure.
func (d Decision) DisputeLine() string {
	if len(d.Disputed) == 0 {
		return ""
	}
	return fmt.Sprintf("%d of %d finding(s) settled by implementer DISPUTE, not a fix: %s",
		len(d.Disputed), d.Raised, strings.Join(idsOf(d.Disputed), ", "))
}

// idsOf lists finding ids. Pure.
func idsOf(fs []Finding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.ID)
	}
	return out
}
