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
// recorded in the ledger and reported to the operator — and the BOUNDARY REVIEW reads the
// ledger (see code-review.md's plan-gate carry-forward), which is what makes the demotion
// safe rather than a silent loss.
//
// `roundCap` is spelled out rather than `cap` so it doesn't shadow the builtin.
func Decide(l Ledger, roundCap int) Decision {
	if roundCap <= 0 {
		roundCap = DefaultRoundCap
	}
	m := vocab.Finding()
	d := Decision{Rounds: len(l.Rounds), CapReached: len(l.Rounds) > roundCap}

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
	}
	d.Reason = b.String()
	return d
}
