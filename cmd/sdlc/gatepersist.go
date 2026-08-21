// gatepersist.go — the ONE persist tail both gate ledgers share (ariadne#194 close
// review BR-43).
//
// The plan gate (#187) and the boundary gate (#194) each end a round the same way: stamp
// the decision onto the round, write the ledger, report what the gate decided. The
// boundary gate began as a copy of that tail, and the copy diverged FIVE times before
// this file existed — each caught separately, by a different review round:
//
//	Blocked never stamped               (M2 review I2)
//	Forced never stamped                (close review BR-17)
//	Forced stamped unconditionally      (close review BR-42)
//	ApplyChecked's round hand-rebuilt   (M2 review BR-10, fixed twice)
//	Decide's Reason never printed       (close review BR-43)
//
// The fifth is why this is a file rather than a sixth patch: `d.Reason` carries "N
// advisory finding(s) recorded for the close review", and at the boundary gate — the last
// read before publish — those findings shipped unannounced. Extracting the tail is the
// fix the reviews recommended four times; patching the fifth divergence would have
// invited a sixth.
package main

import (
	"fmt"
	"io"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// gatePersist is what a gate must supply to end a round: how to name itself in operator
// output, and how to write its own ledger.
type gatePersist struct {
	Label string // "plan-quality" / "boundary gate" — leads every reported line
	Write func(gatestate.Ledger) error
	// Extra runs after the decision is stamped and before the write, for the reporting a
	// gate does NOT share. The boundary gate's demotion warnings live here: a demotion
	// means something different at a gate with no successor, which is a real difference
	// rather than drift.
	Extra func(gatestate.Decision)
}

// stampAndPersist applies a Decision to the ledger's last round, writes it, and reports.
// Returns the Decision unchanged so callers can branch on it.
//
// Reporting the decision is not decoration: `Decide` folds the round cap, demotions and
// the advisory count into `Reason`, and a gate that drops it leaves the operator with no
// statement of what it concluded — which is how four advisory findings shipped
// unannounced from the boundary gate.
func stampAndPersist(stderr io.Writer, g gatePersist, l gatestate.Ledger, d gatestate.Decision, forced string) gatestate.Decision {
	if n := len(l.Rounds); n > 0 {
		l.Rounds[n-1].Blocked = d.Block
		// Only when the gate actually blocked — Round.Forced's documented contract, and
		// --force is a GLOBAL bypass, so an unconditional stamp records a waiver here for
		// a refusal that happened at some other gate.
		l.Rounds[n-1].Forced = forcedRationale(forced, d.Block)
	}
	if g.Extra != nil {
		g.Extra(d)
	}
	if err := g.Write(l); err != nil {
		cwarn(stderr, fmt.Sprintf("%s ledger not persisted: %v", g.Label, err))
	}
	if d.Block {
		cwarn(stderr, g.Label+": "+d.Reason)
		return d
	}
	cok(stderr, g.Label+": "+d.Reason)
	return d
}
