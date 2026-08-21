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
// ledger persists ADDRESSABLE findings, for the next round's prompt. A transcript cannot
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
	return gatestate.RenderPriorFindings(gatestate.FilterBoundary(l, p.Milestone))
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
func seedFromPlanGate(stderr io.Writer, plansDir, issueFileName string, issueNum int, timestamp string) (gatestate.Round, bool) {
	pl, err := readPlanGateLedger(plansDir, issueFileName, issueNum)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate ledger unreadable — its deferred findings are NOT carried into this review: %v", err))
		return gatestate.Round{}, false
	}
	open := gatestate.OpenFindings(pl)
	if len(open) == 0 {
		return gatestate.Round{}, false
	}
	seed := gatestate.Round{N: 1, Timestamp: timestamp, Agent: "sdlc", Boundary: gatestate.BoundaryAll}
	for i, f := range open {
		seed.New = append(seed.New, gatestate.Finding{
			ID:       fmt.Sprintf("%s-%d", boundaryGateKind.IDPrefix, i+1),
			Severity: f.Severity,
			Title:    f.Title,
			Detail:   strings.TrimSpace(f.Detail + "\n(carried from plan-quality " + f.ID + ", deferred to the boundary review)"),
			Round:    1,
		})
	}
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
		cwarn(stderr, fmt.Sprintf("boundary gate ledger not updated: %v", err))
		return gatestate.Decision{}
	}
	issueFileName := filepath.Base(issuePath)
	l, err := readBoundaryGateLedger(p.PlansDir, issueFileName, p.IssueNum)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("boundary gate ledger not updated: %v", err))
		return gatestate.Decision{}
	}
	if len(l.Rounds) == 0 {
		if seed, ok := seedFromPlanGate(stderr, p.PlansDir, issueFileName, p.IssueNum, timestamp); ok {
			l = gatestate.Apply(l, seed)
			cinfo(stderr, fmt.Sprintf("boundary gate: carried %d deferred plan-gate finding(s) into this issue's ledger (#194)", len(seed.New)))
		}
	}

	n := len(l.Rounds) + 1
	if review.Round == nil {
		cwarn(stderr, "boundary review: no valid ```findings block — this round carries NO findings, so the gate cannot converge on it")
		l = gatestate.Apply(l, gatestate.Round{
			N: n, Timestamp: timestamp, Agent: review.Agent, Boundary: p.Milestone,
			ProtocolError: review.ProtocolError, Blocked: true,
		})
	} else {
		round := gatestate.AssignIDsAt(l, *review.Round, n, timestamp, review.Agent, p.Milestone)
		applied, aerr := gatestate.ApplyChecked(l, round)
		if aerr != nil {
			// KEEP the findings — only the DISPOSITIONS failed validation. Discarding
			// them would tell the NEXT round that nothing was ever raised.
			cwarn(stderr, "boundary review: "+aerr.Error())
			round.Dispositions = nil
			round.ProtocolError = aerr.Error()
			l = gatestate.Apply(l, round)
		} else {
			l = applied
		}
	}

	d := gatestate.Decide(gatestate.FilterBoundary(l, p.Milestone), 0)
	if werr := writeBoundaryGateLedger(p.PlansDir, issueFileName, l, repoIdentity()); werr != nil {
		cwarn(stderr, fmt.Sprintf("boundary gate ledger not written: %v", werr))
	} else {
		cok(stderr, "boundary gate ledger: "+boundaryGatePath(p.PlansDir, issueFileName))
	}
	return d
}
