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
