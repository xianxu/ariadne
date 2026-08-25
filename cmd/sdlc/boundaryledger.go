// boundaryledger.go — the IO shell for the BOUNDARY REVIEW's gate ledger (#194).
//
// #187 gave the plan-quality gate memory: findings with binary-assigned stable ids that
// a later round must dispose of before raising anything new. Its own comment predicted
// this file — "the close-boundary gate will declare its own pair" — and that is all this
// is. The read/write bodies live in planreview.go and are shared, since the two gates
// differ by nothing but the triple below.
//
// TWO ARTIFACTS, TWO CONSUMERS. The #136 review sidecar (`-close-review.md`,
// `-m2-review.md`) persists the reviewer's PROSE, for a human or a resuming agent. This
// ledger persists ADDRESSABLE findings, for the next round's prompt. Prose cannot
// answer "has BR-2 been disposed of?", which is why the boundary review kept renumbering
// C1/C2/I1 every round. Neither artifact replaces the other.
//
// ONE LEDGER PER ISSUE, scoped per boundary at the call site — see gatestate.FilterBoundary
// (#194 D1) for why the file is issue-wide but the round cap is not.
package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// boundaryGateSuffix names the boundary-review gate's ledger file.
//
// Deliberately NOT `*-review.md`: construct/vocabulary/verdict.cue declares
// `discovery.glob: "*-review.md"`, and that glob asserts "this document carries a
// boundary verdict". A gate ledger carries findings and no verdict. The trap is sharper
// here than it was for the plan gate, because this gate's PROSE sidecar legitimately IS
// `-close-review.md` — the two files sit side by side for the same boundary.
const boundaryGateSuffix = "close-gate"

// boundaryGateKind identifies the boundary-review gate within gatestate.
var boundaryGateKind = gateLedgerKind{
	Gate:     "boundary-review",
	IDPrefix: "BR",
	Suffix:   boundaryGateSuffix,
}

func boundaryGatePath(plansDir, issueFileName string) string {
	return boundaryGateKind.path(plansDir, issueFileName)
}

func readBoundaryGateLedger(plansDir, issueFileName string, issueNum int) (gatestate.Ledger, error) {
	return readGateLedger(boundaryGateKind, plansDir, issueFileName, issueNum)
}

func writeBoundaryGateLedger(plansDir, issueFileName string, l gatestate.Ledger, repo string) error {
	return writeGateLedger(boundaryGateKind, plansDir, issueFileName, l, repo)
}

// boundaryPriorFindings renders the prior-round block this boundary's reviewer sees —
// the mechanism that turns the memoryless boundary review into a converging one, which
// plan-quality has had since #187 and this gate did not inherit.
//
// Scoped to THIS boundary via FilterBoundary (#194 D1): the ledger is issue-wide so a
// finding family can be seen recurring across milestones, but the findings a reviewer is
// asked to dispose of are its own boundary's, plus anything seeded at BoundaryAll.
//
// A read failure is NOT fatal — the review is still worth running memoryless, and
// refusing to review because a ledger is unreadable would be a worse trade. It warns
// loudly, because a silently-empty block would tell the reviewer "nothing was ever
// raised" and invite it to re-raise everything under new ids.
func boundaryPriorFindings(stderr io.Writer, p boundaryReviewParams) string {
	if p.PlansDir == "" || p.IssueNum <= 0 {
		return ""
	}
	issuePath, err := issueFilePath(p.IssuesDir, p.IssueNum)
	if err != nil {
		return ""
	}
	l, err := readBoundaryGateLedger(p.PlansDir, filepath.Base(issuePath), p.IssueNum)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("boundary gate ledger unreadable — this review runs WITHOUT prior-round memory: %v", err))
		return ""
	}
	// Scoped for what must be disposed, FULL for family counts: a family recurring across
	// milestones is the signal one issue-wide ledger exists to preserve (BR-20).
	return gatestate.RenderPriorFindingsScoped(openScopeFor(l, p.Milestone), l)
}

// seedFromPlanGate returns a BoundaryAll round carrying the plan gate's still-open
// findings, re-issued under this gate's id prefix (#194 D2).
//
// It replaces code-review.md's "Plan-gate carry-forward" instruction, which told the
// reviewer to read `<stem>-plan-gate.md` off disk and re-raise its open findings itself.
// Two carry-forward channels would have put two id namespaces (`PQ-*` and `BR-*`) into
// one output fence with no rule for a disposition naming an id this ledger never issued.
// Seeding folds them into ONE namespace the reviewer can actually dispose of.
//
// BoundaryAll, not the current boundary: those findings were deferred to "the boundary
// review" generically, so scoping them to whichever milestone happened to run first
// would hide them from every later one — a regression against the behavior replaced.
func seedFromPlanGate(stderr io.Writer, bl gatestate.Ledger, plansDir, issueFileName string, issueNum int, timestamp string) (gatestate.Round, bool) {
	pl, err := readPlanGateLedger(plansDir, issueFileName, issueNum)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate ledger unreadable — its deferred findings are NOT carried into this review: %v", err))
		return gatestate.Round{}, false
	}
	open := gatestate.OpenFindings(pl)
	if len(open) == 0 {
		return gatestate.Round{}, false
	}
	// Ids come from AssignIDs, not a hand-rolled index (#194 M2 review): minting
	// `BR-<i+1>` is correct only on an empty ledger, and nothing pinned that
	// precondition. Routing through the ledger's own allocator makes it correct
	// regardless, and keeps one id-minting site (ARCH-DRY).
	rr := gatestate.RoundReport{}
	for _, f := range open {
		rr.New = append(rr.New, gatestate.Finding{
			ID:       "new",
			Severity: f.Severity,
			Title:    f.Title,
			Family:   f.Family, // carry the rule identity across gates, or escalation cannot fire on it

			Detail: strings.TrimSpace(f.Detail + "\n(carried from plan-quality " + f.ID + ", deferred to the boundary review)"),
		})
	}
	seed := gatestate.AssignIDsAt(bl, rr, 1, timestamp, "sdlc", gatestate.BoundaryAll)
	seed.NoCap = true // bookkeeping, not a review cycle
	return seed, true
}

// persistBoundaryRound applies this review's findings to the issue's boundary ledger and
// returns the gate decision for the CURRENT boundary. Called at finalize time, under the
// repo transaction lock — the ledger is a repo write and dispatch runs with the lock
// released.
//
// Seeding (D2) happens on the ledger's first round, so a plan-gate finding deferred to
// "the boundary review" is disposed of through the same fence as everything else.
//
// A protocol miss is PERSISTED, not dropped: otherwise len(Rounds) stays 0 forever for a
// reviewer CLI that never emits the fence, the round cap can never fire, and the prompt
// keeps announcing "no prior rounds" on invocation six.
func persistBoundaryRound(stderr io.Writer, p boundaryReviewParams, review reviewResult, timestamp string) gatestate.Decision {
	if p.PlansDir == "" || p.IssueNum <= 0 {
		return gatestate.Decision{}
	}
	issuePath, err := issueFilePath(p.IssuesDir, p.IssueNum)
	if err != nil {
		return blockOnLedgerFailure(stderr, fmt.Sprintf("cannot resolve the issue file: %v", err))
	}
	issueFileName := filepath.Base(issuePath)
	l, err := readBoundaryGateLedger(p.PlansDir, issueFileName, p.IssueNum)
	if err != nil {
		return blockOnLedgerFailure(stderr, err.Error())
	}
	if len(l.Rounds) == 0 {
		if seed, ok := seedFromPlanGate(stderr, l, p.PlansDir, issueFileName, p.IssueNum, timestamp); ok {
			l = gatestate.Apply(l, seed)
			cinfo(stderr, fmt.Sprintf("boundary gate: carried %d deferred plan-gate finding(s) into this issue's ledger (#194)", len(seed.New)))
		}
	}

	n := len(l.Rounds) + 1
	if review.Round == nil {
		cwarn(stderr, "boundary review: no valid ```findings block — this round carries NO findings, so the gate cannot converge on it")
		// A review that never STARTED did not consume a cycle; one that ran and emitted
		// no fence did (that is the case #187's persist rule exists to bound).
		neverRan := strings.HasPrefix(review.ProtocolError, "review did not run:")
		l = gatestate.Apply(l, gatestate.Round{
			N: n, Timestamp: timestamp, Agent: review.Agent, Boundary: p.Milestone,
			ProtocolError: review.ProtocolError, Blocked: true, NoCap: neverRan,
		})
	} else {
		round := gatestate.AssignIDsAt(l, *review.Round, n, timestamp, review.Agent, p.Milestone)
		applied, aerr := gatestate.ApplyChecked(l, round)
		l = applied // ApplyChecked applies the round either way, keeping the valid dispositions
		if aerr != nil {
			cwarn(stderr, "boundary review: "+aerr.Error())
			l.Rounds[len(l.Rounds)-1].ProtocolError = aerr.Error()
		}
	}

	// Cap per boundary; open findings from the WHOLE issue at the final boundary (BR-37).
	d := gatestate.DecideScoped(gatestate.FilterBoundary(l, p.Milestone), openScopeFor(l, p.Milestone),
		roundCapFromEnvVar("WF_BOUNDARY_ROUND_CAP"))
	// Stamp the outcome onto the round BEFORE writing (mirrors changecode.go:536-537).
	// Without this the one durable record of "did this gate refuse" says `passed` for a
	// round that refused, and PassesUnchanged — which #183's --fixed-to-ship pass-through
	// will read at exactly this gate — reads that field.
	return stampAndPersist(stderr, gatePersist{
		Label: "boundary gate",
		Write: func(out gatestate.Ledger) error {
			return writeBoundaryGateLedger(p.PlansDir, issueFileName, out, repoIdentity())
		},
		Extra: func(d gatestate.Decision) {
			// #194 M3: the convergence signal — the line missing when tools#1 ran four
			// rounds with no way to tell whether round five would find more. Capping on
			// finding COUNT is arbitrary; capping when families stop repeating is not.
			cinfo(stderr, "boundary gate: "+gatestate.ConvergenceLine(l, len(l.Rounds)))
			// A demotion means something DIFFERENT here than at the plan gate, which is
			// why this is Extra rather than shared. There, a demoted finding is deferred
			// to the boundary review, which picks it up — that is what makes the cap
			// safe. Here there IS no later gate, so a demoted finding ships having
			// blocked nothing and the operator has to be told.
			for _, fnd := range d.Demoted {
				cwarn(stderr, fmt.Sprintf("boundary gate: [%s] %s demoted past the round cap and will NOT block — "+
					"no later gate picks it up: %s", fnd.ID, fnd.Severity, fnd.Title))
			}
		},
	}, l, d, p.ForcedRationale)
}

// blockOnLedgerFailure is the fail-closed answer to an unusable boundary ledger.
//
// Returning an empty Decision would mean Block:false — this round's findings dropped,
// the corrupt file left unwritten, and the close FINALIZING. That is precisely the
// behavior readGateLedger refuses at a finer grain ("a silent reset is worse than the
// status quo because it would look like it worked"), and the plan gate halts on it
// (changecode.go). Doing the opposite one level up would let a corrupt file turn the
// gate off silently — the single worst failure mode a gate has.
func blockOnLedgerFailure(stderr io.Writer, reason string) gatestate.Decision {
	cwarn(stderr, "boundary gate ledger unusable — refusing to finalize rather than close without it: "+reason)
	cwarn(stderr, "fix or delete the ledger file, then re-run; do NOT let the gate silently forget")
	return gatestate.Decision{Block: true, Reason: "boundary gate ledger unusable: " + reason}
}

// openScopeFor returns the ledger view whose OPEN FINDINGS this boundary must dispose of.
//
// A milestone sees its own. The WHOLE-ISSUE close (milestone "") sees everything
// (ariadne#194 M3 review BR-37): it is the last gate before publish, and a finding left
// undisposed at a milestone has no other path to disposal — its boundary has closed, so
// nothing will ever look at it again. Measured when this was found: 15 open findings,
// three of them Important, invisible to the close that was about to ship them.
//
// This is NOT the same as dropping the filter. The round cap still scopes per boundary
// (see DecideScoped) — unfiltered, this issue's 8 counted rounds against a cap of 3 would
// demote every Important on the close's first round.
func openScopeFor(l gatestate.Ledger, milestone string) gatestate.Ledger {
	if milestone == "" {
		return l
	}
	return gatestate.FilterBoundary(l, milestone)
}
