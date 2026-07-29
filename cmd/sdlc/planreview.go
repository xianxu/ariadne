// planreview.go — the IO shell for change-code's plan-gate ledger (#187).
//
// The gate's memory lives in `workshop/plans/NNNNNN-slug-plan-gate.md`. This file is the
// ONLY place that touches the filesystem or the clock for that ledger; all parsing,
// rendering, ID assignment and gate decisions stay pure in internal/gatestate (ARCH-PURE),
// which is what lets the whole convergence policy be tested on in-memory strings.
//
// Deliberately NOT named `-plan-review.md`: construct/vocabulary/verdict.cue declares
// `discovery.glob: "*-review.md"`, and that glob asserts "this document carries a boundary
// verdict". A gate ledger carries findings and no verdict, so it stays out of that family —
// otherwise a future verdict consumer would be handed a document it cannot validate.
package main

import (
	"fmt"
	"os"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// planGateSuffix names the plan-quality gate's ledger file. Matches the `finding` noun's
// discovery glob (`*-plan-gate.md`) in construct/vocabulary/finding.cue.
const planGateSuffix = "plan-gate"

// planGateGate / planGateIDPrefix identify the plan-quality gate within the gate-agnostic
// gatestate package. #183's close-boundary gate will declare its own pair.
const (
	planGateGate     = "plan-quality"
	planGateIDPrefix = "PQ"
)

// planGatePath returns the plan-gate ledger path for an issue file.
func planGatePath(plansDir, issueFileName string) string {
	return sidecarPathFor(plansDir, issueFileName, planGateSuffix)
}

// readPlanGateLedger loads the ledger, or returns a fresh empty one when the sidecar does
// not exist yet (the normal round-1 state).
//
// A sidecar that EXISTS but does not parse is an ERROR, never an empty ledger. Silently
// resetting would erase every disposition and re-open findings the operator already
// addressed — the exact forgetting this feature exists to prevent, and worse than the
// status quo because it would look like it worked.
func readPlanGateLedger(plansDir, issueFileName string, issueNum int) (gatestate.Ledger, error) {
	empty := gatestate.Ledger{Gate: planGateGate, IssueNum: issueNum, IDPrefix: planGateIDPrefix}
	path := planGatePath(plansDir, issueFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return empty, nil
		}
		return empty, fmt.Errorf("read %s: %w", path, err)
	}
	l, perr := gatestate.ParseSidecar(string(raw))
	if perr != nil {
		return empty, fmt.Errorf("%s: %w (fix or delete it — do NOT let the gate silently forget)", path, perr)
	}
	// Identity fields are owned by the binary, not the file: repair them rather than
	// trusting a hand-edited header.
	l.Gate, l.IssueNum, l.IDPrefix = planGateGate, issueNum, planGateIDPrefix
	return l, nil
}

// writePlanGateLedger renders and atomically writes the ledger, reusing reviewsidecar.go's
// atomicWriteFile so both durable gate artifacts share one write path (ARCH-DRY).
func writePlanGateLedger(plansDir, issueFileName string, l gatestate.Ledger, repo string) error {
	return atomicWriteFile(planGatePath(plansDir, issueFileName), []byte(gatestate.Render(l, repo)))
}
