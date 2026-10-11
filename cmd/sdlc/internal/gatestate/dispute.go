package gatestate

import (
	"fmt"
	"sort"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// Disputed is the implementer's closing disposition (ariadne#202). Named here because the
// dispute stream writes it; TestDisputedIsTheModelsImplementerDisposition pins it to
// finding.cue's `authority.implementer`, so the literal cannot drift from the model.
const Disputed = "disputed"

// Dispute is the IMPLEMENTER's answer to a finding: "considered, and declined, with
// reasons and a counter-measurement" (ariadne#202). Before it, the only closing exit an
// implementer had was a fix, so a wrong finding and a correct-but-not-worth-it one both
// left by the round cap and the next judge inherited no counter-evidence.
//
// It is a separate stream on the Ledger, not a Round: a Round is a reviewer invocation,
// and len(Rounds) feeds the round cap, id numbering, PassesUnchanged and the convergence
// line — none of which a dispute may move. The two streams meet in exactly one place,
// timeline, which orders them.
type Dispute struct {
	ID        string `yaml:"id"`
	Rationale string `yaml:"rationale"`
	// Evidence is the counter-measurement: the command re-run and what it showed, or the
	// pinned invariant the fix would break. Mandatory — disputing costs evidence, fixing
	// costs a diff.
	Evidence string `yaml:"evidence"`
	// After is the highest round N in the ledger when the dispute was recorded: it takes
	// effect after that round's dispositions and before the next round's, so the next
	// judge's answer overrides it and an earlier judgment does not.
	After     int    `yaml:"after"`
	Timestamp string `yaml:"timestamp"`
	By        string `yaml:"by,omitempty"` // the workspace that recorded it
}

// disposition is the dispute as a timeline entry.
func (d Dispute) disposition() Disposition {
	return Disposition{ID: d.ID, State: Disputed, Note: d.Rationale, Round: d.After}
}

// timeline returns every disposition in the order it took effect: each round's
// dispositions in ledger order, with every dispute whose After precedes that round's N
// applied just before it, and the rest after the last round. It is the ONE place the
// judge's and the implementer's streams are ordered; closedSet, DispositionCounts and the
// prompt all read it. Pure.
func timeline(l Ledger) []Disposition {
	pending := append([]Dispute{}, l.Disputes...)
	sort.SliceStable(pending, func(i, j int) bool { return pending[i].After < pending[j].After })
	var out []Disposition
	i := 0
	for _, r := range l.Rounds {
		for i < len(pending) && pending[i].After < r.N {
			out = append(out, pending[i].disposition())
			i++
		}
		out = append(out, r.Dispositions...)
	}
	for ; i < len(pending); i++ {
		out = append(out, pending[i].disposition())
	}
	return out
}

// latestStates maps each disposed finding id to its current disposition: last on the
// timeline wins. Pure.
func latestStates(l Ledger) map[string]string {
	state := map[string]string{}
	for _, d := range timeline(l) {
		state[d.ID] = d.State
	}
	return state
}

// settles reports whether disposition `state` settles a finding of severity `sev`. A
// closing disposition settles, except an implementer one on an appeal-only severity: that
// is an appeal, put to the next judge, and the finding keeps blocking until the judge
// answers. Read from the model, never a switch on literals. Pure.
func settles(m *vocab.FindingModel, state, sev string) bool {
	if !m.Closes(state) {
		return false
	}
	return !(m.ImplementerMay(state) && m.IsAppealOnly(sev))
}

// findingByID indexes every finding ever raised in l. Pure.
func findingByID(l Ledger) map[string]Finding {
	out := map[string]Finding{}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			out[f.ID] = f
		}
	}
	return out
}

// LiveDispute is a finding whose current state is a dispute the judge has not yet
// answered, with the dispute that put it there.
type LiveDispute struct {
	Finding Finding
	Dispute Dispute
}

// LiveDisputes returns, in first-raised order, every finding whose latest disposition is
// a dispute — what the next judge must answer, and what the gate line counts. Pure.
func LiveDisputes(l Ledger) []LiveDispute {
	state := latestStates(l)
	last := map[string]Dispute{}
	for _, d := range l.Disputes {
		if prev, ok := last[d.ID]; !ok || d.After >= prev.After {
			last[d.ID] = d
		}
	}
	var out []LiveDispute
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if state[f.ID] == Disputed {
				out = append(out, LiveDispute{Finding: f, Dispute: last[f.ID]})
			}
		}
	}
	return out
}

// AdmitDispute is the dispute verb's pure admission check: which finding `id` names, or
// why it cannot be disputed now. Refusals, each with its own message:
//
//   - an id this ledger never issued;
//   - a finding whose latest state is already a dispute — the next review round answers
//     it, and a second dispute before then would let the implementer argue unopposed;
//   - a finding the judge already settled — there is nothing left to decline.
//
// An undisposed finding, or one the judge re-raised `not-addressed`, is admitted.
func AdmitDispute(l Ledger, id string) (Finding, error) {
	f, ok := findingByID(l)[id]
	if !ok {
		return Finding{}, fmt.Errorf("no finding %s in the %s ledger", id, l.Gate)
	}
	switch state := latestStates(l)[id]; {
	case state == Disputed:
		after := 0
		for _, d := range l.Disputes {
			if d.ID == id && d.After > after {
				after = d.After
			}
		}
		return Finding{}, fmt.Errorf("%s is already disputed (after round %d); the next review round answers it", id, after)
	case state != "" && vocab.Finding().Closes(state):
		return Finding{}, fmt.Errorf("%s is already %s by the judge; there is nothing to dispute", id, state)
	}
	return f, nil
}

// HighestRound is the largest round N in l — the After a dispute recorded now carries.
// Pure.
func HighestRound(l Ledger) int {
	n := 0
	for _, r := range l.Rounds {
		if r.N > n {
			n = r.N
		}
	}
	return n
}
