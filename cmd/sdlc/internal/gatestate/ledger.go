// Package gatestate is the durable memory of an SDLC gate: the findings a fresh-context
// judge raised, the stable IDs the binary assigned them, how later rounds disposed of
// them, and the pure decision of whether the gate may be crossed.
//
// It exists because a stateless gate cannot converge (ariadne#187). `sdlc change-code`
// re-dispatched a brand-new plan reviewer on every invocation; with no memory of its own
// prior findings it re-derived an absolute bar each round and surfaced the next-deepest
// layer of a plan that kept improving — five rejections for one 126-line change. A gate
// that remembers says "you addressed my three findings, ship."
//
// Everything here is PURE (ARCH-PURE): no filesystem, no clock, no subprocess. The
// timestamp and agent name are captured at the IO boundary (cmd/sdlc/planreview.go) and
// passed in. That is what lets the whole convergence policy be tested on in-memory
// strings with no mocks.
//
// The package is deliberately gate-agnostic — Ledger.Gate and Ledger.IDPrefix are data,
// not constants — so ariadne#183's milestone-close `--fixed-to-ship` consumes the same
// notion of gate state rather than inventing a second one (ARCH-DRY).
package gatestate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/xianxu/ariadne/pkg/vocab"
)

// Finding is one judgment the gate raised, carrying the binary-assigned stable ID that
// lets a LATER round refer to it. Severity is validated against the `finding` model at
// parse time, so by the time a Finding exists here its severity is modeled.
type Finding struct {
	ID       string `yaml:"id"`
	Severity string `yaml:"severity"`
	Title    string `yaml:"title"`
	Detail   string `yaml:"detail,omitempty"`
	Round    int    `yaml:"round"` // the round that first raised it
}

// Disposition is a later round's verdict on an EARLIER finding.
type Disposition struct {
	ID    string `yaml:"id"`
	State string `yaml:"disposition"`
	Note  string `yaml:"note,omitempty"`
	Round int    `yaml:"round,omitempty"` // the round that disposed it
}

// RoundReport is what the judge emitted this round, BEFORE the binary assigns IDs.
type RoundReport struct {
	Dispositions []Disposition `yaml:"dispose,omitempty"`
	New          []Finding     `yaml:"findings,omitempty"`
}

// Round is one gate invocation's durable record: what it disposed, what it raised, whether
// the gate blocked, and whether the operator forced past it.
type Round struct {
	N            int           `yaml:"n"`
	Timestamp    string        `yaml:"timestamp"`
	Agent        string        `yaml:"agent"`
	Dispositions []Disposition `yaml:"dispose,omitempty"`
	New          []Finding     `yaml:"findings,omitempty"`
	// Forced carries the --force rationale, set ONLY when this gate actually blocked.
	// --force is a GLOBAL bypass, so stamping it unconditionally would mark a plan-gate
	// round "forced" when the operator forced past a structural failure — over-reporting
	// overrides in the one number meant to answer "which gates earn their cost".
	Forced  string `yaml:"forced,omitempty"`
	Blocked bool   `yaml:"blocked"`
	// ProtocolError is set when the judge emitted no valid findings block. Such a round
	// carries no findings but is still PERSISTED: if it were dropped, len(Rounds) would
	// stay 0 forever for an agent CLI that never emits the fence, the round cap could
	// never fire, and the close-time gate_rounds metric would report 0 for precisely the
	// most expensive sessions.
	ProtocolError string `yaml:"protocol_error,omitempty"`
}

// Ledger is the accumulated state of ONE gate on ONE issue across every invocation.
type Ledger struct {
	Gate     string  `yaml:"gate"` // e.g. "plan-quality"
	IssueNum int     `yaml:"issue"`
	IDPrefix string  `yaml:"id_prefix"` // e.g. "PQ" — IDs are <prefix>-<n>
	Rounds   []Round `yaml:"rounds"`
	// ContentHash is sha256(issue+plan) as of the last PASSING round — the pass-through
	// key. Without it, moving the estimate gates below plan-quality (#187 B1) would make
	// every estimate-gate failure cost a fresh multi-minute judge dispatch on the retry,
	// where today it costs milliseconds. #183's --fixed-to-ship is the same mechanism at
	// the close boundary, which is why this lives on the shared Ledger (ARCH-DRY).
	ContentHash string `yaml:"content_hash,omitempty"`
}

// trimIDPrefix strips "<prefix>-" from an ID, returning the remainder (the sequence
// number, for a well-formed ID).
func trimIDPrefix(id, prefix string) string {
	return strings.TrimPrefix(id, prefix+"-")
}

// nextIDSeq returns one past the highest ID sequence number ever assigned in l. IDs are
// never reused, even after a finding is withdrawn: a stable ID that changed meaning
// between rounds would let a later round dispose the wrong finding.
//
// Non-conforming IDs are skipped rather than erroring — they cannot have come from
// AssignIDs, and refusing to issue new IDs because an old one looks odd would wedge the
// gate shut.
func nextIDSeq(l Ledger) int {
	max := 0
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if n, err := strconv.Atoi(trimIDPrefix(f.ID, l.IDPrefix)); err == nil && n > max {
				max = n
			}
		}
	}
	return max + 1
}

// AssignIDs stamps binary-assigned stable IDs onto a RoundReport's new findings and
// returns the durable Round. The judge emits `id: new`; assigning IDs here (not in the
// prompt) means the judge never has to invent a globally-unique identifier — it only has
// to REFER to the ones we handed it.
func AssignIDs(l Ledger, rr RoundReport, n int, timestamp, agent string) Round {
	seq := nextIDSeq(l)
	out := Round{N: n, Timestamp: timestamp, Agent: agent}
	for _, d := range rr.Dispositions {
		d.Round = n
		out.Dispositions = append(out.Dispositions, d)
	}
	for _, f := range rr.New {
		f.ID = fmt.Sprintf("%s-%d", l.IDPrefix, seq)
		f.Round = n
		seq++
		out.New = append(out.New, f)
	}
	return out
}

// Apply appends a round to the ledger. Use ApplyChecked for an agent-sourced round.
func Apply(l Ledger, r Round) Ledger {
	l.Rounds = append(append([]Round{}, l.Rounds...), r)
	return l
}

// ApplyChecked is Apply plus the protocol validation an agent-sourced round needs: every
// disposition must name a finding raised in an EARLIER round, and carry a modeled
// disposition. A judge disposing an ID we never issued is a genuine protocol error to
// surface, not a value to guess (the agent-binary-handoff-schema target).
func ApplyChecked(l Ledger, r Round) (Ledger, error) {
	known := map[string]bool{}
	for _, prev := range l.Rounds {
		for _, f := range prev.New {
			known[f.ID] = true
		}
	}
	m := vocab.Finding()
	for _, d := range r.Dispositions {
		if !known[d.ID] {
			return l, fmt.Errorf("round %d disposes unknown finding %q", r.N, d.ID)
		}
		if !m.IsDisposition(d.State) {
			return l, fmt.Errorf("round %d: unmodeled disposition %q for %s", r.N, d.State, d.ID)
		}
	}
	return Apply(l, r), nil
}

// closedSet computes, for every finding ID ever disposed, whether it is currently settled.
// Closed-ness comes from the MODEL's closing/open partition, never a switch on literals: a
// disposition added to finding.cue must not be able to reach here as an unhandled case
// that silently leaves a finding open forever.
//
// Later rounds override earlier ones, so a finding disposed `addressed` and then
// re-disposed `not-addressed` is open again.
func closedSet(l Ledger) map[string]bool {
	m := vocab.Finding()
	closed := map[string]bool{}
	for _, r := range l.Rounds {
		for _, d := range r.Dispositions {
			closed[d.ID] = m.Closes(d.State)
		}
	}
	return closed
}

// OpenFindings returns every finding never settled by a closing disposition, in the order
// they were first raised. A `not-addressed` disposition leaves the finding OPEN — that is
// the whole point: a judge saying "still not addressed" must keep blocking.
func OpenFindings(l Ledger) []Finding {
	closed := closedSet(l)
	var out []Finding
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if !closed[f.ID] {
				out = append(out, f)
			}
		}
	}
	return out
}

// DispositionCounts tallies how the gate's findings were resolved: settled as addressed,
// retracted as withdrawn, and still open. This is the "finding disposition" the close-time
// report emits — the number that answers "did the gate's findings get acted on, or worked
// around?", distinct from the accepted-vs-forced count.
func DispositionCounts(l Ledger) (addressed, withdrawn, open int) {
	// Last disposition wins, matching closedSet.
	state := map[string]string{}
	for _, r := range l.Rounds {
		for _, d := range r.Dispositions {
			state[d.ID] = d.State
		}
	}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			switch state[f.ID] {
			case "addressed":
				addressed++
			case "withdrawn":
				withdrawn++
			default:
				open++
			}
		}
	}
	return addressed, withdrawn, open
}

// ContentHash is the pass-through key: sha256 of the issue + plan text the gate reviewed.
// The separator cannot appear in either input's normal content, so it prevents a
// boundary-shifting collision between (issue+plan) splits.
func ContentHash(issueContent, planContent string) string {
	sum := sha256.Sum256([]byte(issueContent + "\x00--gatestate--\x00" + planContent))
	return hex.EncodeToString(sum[:])
}

// PassesUnchanged reports whether the gate may skip dispatch entirely: there is at least
// one round, the most recent one did NOT block, and the content is byte-identical to what
// that round passed. Pure — the caller supplies the hash.
//
// The "did not block" clause is load-bearing: a short-circuit after a blocking round would
// let a refused plan through unchanged.
func PassesUnchanged(l Ledger, hash string) bool {
	if len(l.Rounds) == 0 || l.ContentHash == "" || hash != l.ContentHash {
		return false
	}
	return !l.Rounds[len(l.Rounds)-1].Blocked
}
