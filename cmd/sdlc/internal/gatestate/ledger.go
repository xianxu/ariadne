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
	// Boundary names which review boundary this round belongs to — "M1", "M2", "" for
	// the whole-issue close, or BoundaryAll for a round that belongs to all of them
	// (ariadne#194 D1). Empty on every plan-quality round, which has one boundary by
	// construction; omitempty keeps those ledgers byte-identical to their pre-#194 form.
	//
	// It exists because Decide reads len(l.Rounds) against the round cap and
	// OpenFindings spans the whole ledger. ONE boundary ledger per issue is what lets
	// a finding FAMILY be seen recurring across milestones; scoping the cap and the
	// open set per boundary via FilterBoundary is what keeps that from silently
	// demoting the whole-issue close's first-round findings.
	Boundary string `yaml:"boundary,omitempty"`
	// NoCap marks a round that did NOT consume a review cycle, so it does not count
	// toward the round cap (ariadne#194 M2 review). Three kinds qualify: the plan-gate
	// SEED round (no reviewer ran), a dispatch that never started, and a round persisted
	// before a non-review refusal.
	//
	// The cap exists to bound a judge that keeps churning. Letting a machine-sleep
	// interruption or a bookkeeping round eat the budget punishes the operator for
	// something no reviewer did — observed live on ariadne#194, where two reviews killed
	// by host sleep put a boundary at 2 of 3 rounds having received zero review content.
	//
	// It is deliberately the NEGATIVE spelling: rounds written before this field existed
	// default to false and therefore keep counting, so no historical ledger changes
	// meaning. A round where the reviewer genuinely ran and emitted no fence still counts
	// — that is the case #187's persist-the-protocol-error rule exists to bound.
	NoCap bool `yaml:"no_cap,omitempty"`
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

// BoundaryAll marks a round that belongs to EVERY boundary rather than one (#194 D5).
// Its use is the plan-gate seed round: those findings were deferred to "the boundary
// review" generically, not to whichever milestone happened to close first, so scoping
// them to one boundary would hide them everywhere else — a regression against
// code-review.md's pre-#194 instruction that every boundary reviewer read the plan-gate
// ledger.
const BoundaryAll = "*"

// FilterBoundary returns a view of l holding only the rounds belonging to `boundary`
// (plus every BoundaryAll round). Pure: it is a caller-side transform, so Decide and
// OpenFindings keep the signatures plan-quality already depends on rather than growing
// a boundary parameter that one of the two gates would always pass empty (ARCH-PURE).
//
// FamilyCounts deliberately takes the UNFILTERED ledger — a family recurring across
// milestones is precisely the signal #194 exists to surface.
func FilterBoundary(l Ledger, boundary string) Ledger {
	out := l
	out.Rounds = nil
	for _, r := range l.Rounds {
		if r.Boundary == boundary || r.Boundary == BoundaryAll {
			out.Rounds = append(out.Rounds, r)
		}
	}
	return out
}

// CountedRounds is the number of rounds that consumed a review cycle — the figure the
// round cap is about. See Round.NoCap for which rounds are excluded and why. Pure.
func CountedRounds(l Ledger) int {
	n := 0
	for _, r := range l.Rounds {
		if !r.NoCap {
			n++
		}
	}
	return n
}

// Ledger is the accumulated state of ONE gate on ONE issue across every invocation.
type Ledger struct {
	Gate     string  `yaml:"gate"` // e.g. "plan-quality"
	IssueNum int     `yaml:"issue"`
	IDPrefix string  `yaml:"id_prefix"` // e.g. "PQ" — IDs are <prefix>-<n>
	Rounds   []Round `yaml:"rounds,omitempty"`
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
	// highest, not `max` — Decide in this package spells out `roundCap` for the same reason,
	// and one convention per package beats two.
	highest := 0
	for _, r := range l.Rounds {
		for _, f := range r.New {
			if n, err := strconv.Atoi(trimIDPrefix(f.ID, l.IDPrefix)); err == nil && n > highest {
				highest = n
			}
		}
	}
	return highest + 1
}

// AssignIDs stamps binary-assigned stable IDs onto a RoundReport's new findings and
// returns the durable Round. The judge emits `id: new`; assigning IDs here (not in the
// prompt) means the judge never has to invent a globally-unique identifier — it only has
// to REFER to the ones we handed it.
func AssignIDs(l Ledger, rr RoundReport, n int, timestamp, agent string) Round {
	return AssignIDsAt(l, rr, n, timestamp, agent, "")
}

// AssignIDsAt is AssignIDs with the round's boundary stamped on (#194 D1). AssignIDs
// stays as the one-boundary spelling plan-quality uses, so its call sites need no edit.
func AssignIDsAt(l Ledger, rr RoundReport, n int, timestamp, agent, boundary string) Round {
	seq := nextIDSeq(l)
	out := Round{N: n, Timestamp: timestamp, Agent: agent, Boundary: boundary}
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
//
// REJECTION IS PER-DISPOSITION, not per-round (ariadne#194 M2 review). Failing the whole
// round on the first bad id meant one typo nullified every VALID disposal beside it — at
// a gate whose entire purpose is disposal, that turns a formatting slip into findings
// that stay open for another full review cycle. The bad ones are dropped and named; the
// good ones apply. The error still surfaces, so the caller records a protocol error.
func ApplyChecked(l Ledger, r Round) (Ledger, error) {
	known := map[string]bool{}
	for _, prev := range l.Rounds {
		for _, f := range prev.New {
			known[f.ID] = true
		}
	}
	m := vocab.Finding()
	var kept []Disposition
	var rejected []string
	for _, d := range r.Dispositions {
		switch {
		case !known[d.ID]:
			rejected = append(rejected, fmt.Sprintf("%s (unknown finding)", d.ID))
		case !m.IsDisposition(d.State):
			rejected = append(rejected, fmt.Sprintf("%s (unmodeled disposition %q)", d.ID, d.State))
		default:
			kept = append(kept, d)
		}
	}
	r.Dispositions = kept
	if len(rejected) > 0 {
		return Apply(l, r), fmt.Errorf("round %d: dropped %d invalid disposition(s): %s",
			r.N, len(rejected), strings.Join(rejected, ", "))
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
// The tally is keyed off the MODEL, not a switch on disposition literals. A switch would
// be the exact posture finding.cue argues against: adding `obsolete` to
// `dispositions.closing` would leave OpenFindings correct (it derives from Closes) while
// this silently counted the finding as open — under-reporting settled findings in the one
// metric meant to answer "did the gate's findings get acted on, or worked around?".
func DispositionCounts(l Ledger) (counts map[string]int, open int) {
	// Last disposition wins, matching closedSet.
	state := map[string]string{}
	for _, r := range l.Rounds {
		for _, d := range r.Dispositions {
			state[d.ID] = d.State
		}
	}
	m := vocab.Finding()
	counts = map[string]int{}
	// Seed only the CLOSING dispositions. An open one (`not-addressed`) can never be
	// incremented here — those findings are counted in `open` — so seeding it produced a
	// permanent zero bucket that told a caller iterating the map that a real category had
	// no members. The buckets are exactly "the ways a finding got settled".
	for _, d := range m.AllDispositions() {
		if m.Closes(d) {
			counts[d] = 0
		}
	}
	for _, r := range l.Rounds {
		for _, f := range r.New {
			s := state[f.ID]
			if s == "" || !m.Closes(s) {
				open++
				continue
			}
			counts[s]++
		}
	}
	return counts, open
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
