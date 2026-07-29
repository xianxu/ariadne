// churnreport.go — the close-time COST report (#187 D1–D3): where a window's lines
// landed, how many times they were written, and what the plan gate charged to let them
// through. `sdlc close` prints it and the calibration ledger records it, so "which gates
// earn their cost" becomes a query over history rather than a recollection.
//
// Everything here degrades. A cost report is diagnostic: no failure in it may cost a close
// its completion, so each measurement warns and zeroes rather than returning an error
// upward. That is the `appendCalibrationRow` precedent (a missing ledger must never break
// `sdlc close`) applied to churn and gate state.
package main

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/churn"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
	"github.com/xianxu/ariadne/pkg/vocab"
)

// churnForWindow measures churn over [baseLong, HEAD] — the SAME window the boundary
// review and the atlas gate use. Callers pass boundaryWindowBase's result rather than
// re-deriving it (ARCH-DRY), which is what makes the reported churn provably cover the
// commits that were actually reviewed.
//
// Two git reads, and they are deliberately different queries:
//
//   - `diff --numstat base..HEAD` — the NET result: what survived to be reviewed.
//   - `log --numstat --format= base..HEAD` — every commit's insertions, so a file
//     rewritten five times counts five times. That difference IS the rework signal.
//
// An EMPTY base is not an error: boundaryWindowBase returns "" on a docs-only window
// with no `#N` commit, and a close there must still succeed. It yields a zero report.
//
// A BAD base is an error, and reporting it is why this uses gitx.RunGit rather than
// gitx.Capture. Capture flattens any failure to "" (internal/gitx/window.go:50-56, whose
// own doc warns against exactly this use), so a bogus SHA would render
// `churn: prod 0 / test 0 / …` — indistinguishable from a genuinely empty window, in the
// one number introduced to answer "which gates earn their cost". The error goes up; the
// CALLER warns and zeroes.
func churnForWindow(baseLong string) (churn.Report, error) {
	if baseLong == "" {
		return churn.Report{}, nil
	}
	span := baseLong + "..HEAD"

	finalOut, err := gitx.RunGit("diff", "--numstat", span)
	if err != nil {
		return churn.Report{}, fmt.Errorf("git diff --numstat %s: %w", span, err)
	}
	commitOut, err := gitx.RunGit("log", "--numstat", "--format=", span)
	if err != nil {
		return churn.Report{}, fmt.Errorf("git log --numstat %s: %w", span, err)
	}

	final := churn.ParseNumstat(string(finalOut))
	commitTotal := churn.TotalInsertions(churn.ParseNumstat(string(commitOut)))
	return churn.Summarize(final, commitTotal), nil
}

// Closing dispositions that have a dedicated ledger column. The TSV is positional and
// append-only, so unlike gatestate's logic these names CANNOT be model-derived — a column
// name is a schema commitment. What can be enforced is that schema and model stay in step:
// TestLedgerCoversEveryClosingDisposition fails the moment finding.cue adds a closing
// disposition with no column, rather than letting those findings quietly vanish from the
// one metric meant to answer "did the gate's findings get acted on, or worked around?".
const (
	dispAddressed = "addressed"
	dispWithdrawn = "withdrawn"
)

// closeCostMetrics is the cost picture of one close window: the churn split and the plan
// gate's round-trips. Zero values mean "not measured" as much as they mean "measured
// zero" — the distinction lives in the warning closeMetrics prints when it degrades.
type closeCostMetrics struct {
	Churn     churn.Report
	Rounds    int
	Forced    int
	Addressed int
	Withdrawn int
	Open      int
}

// closeMetrics gathers the cost report. It NEVER returns an error: every measurement that
// fails warns and leaves its values at zero, because none of this is worth failing a close
// over.
//
// The window is boundaryWindowBase's — the same one the boundary review and the atlas gate
// use (ARCH-DRY). Note what that excludes: the close's own edits are not committed yet, so
// they fall outside base..HEAD. That is deliberate rather than a rounding error — the
// report describes the window that was REVIEWED, and diverging from the review's window to
// chase a few final lines would break the one property that makes these numbers comparable
// across issues.
func closeMetrics(stderr io.Writer, f *closeFlags, res closeResult) closeCostMetrics {
	var m closeCostMetrics

	base := boundaryWindowBase(res.issueStr, f.Milestone, res.issuePath)
	if r, err := churnForWindow(base); err != nil {
		cwarn(stderr, fmt.Sprintf("churn not measured (%v)", err))
	} else {
		m.Churn = r
	}

	// An ABSENT sidecar is the normal case for any issue that never ran the stateful gate
	// (every issue closed before #187, and any closed with --no-plan-quality): the reader
	// yields an empty ledger with no error, so the gate columns are honestly zero. A
	// sidecar that exists but does not PARSE is different, and it warns.
	issueFile := filepath.Base(res.issuePath)
	l, err := readPlanGateLedger(f.plansDir(), issueFile, f.Issue)
	if err != nil {
		cwarn(stderr, fmt.Sprintf("plan-gate metrics not measured (%v)", err))
		return m
	}
	m.Rounds = len(l.Rounds)
	for _, r := range l.Rounds {
		if r.Forced != "" {
			m.Forced++
		}
	}
	counts, open := gatestate.DispositionCounts(l)
	m.Addressed, m.Withdrawn, m.Open = counts[dispAddressed], counts[dispWithdrawn], open
	return m
}

// ChurnLine is the D1/D2 operator line: where the lines landed, and how many were written
// to land them.
func (m closeCostMetrics) ChurnLine() string {
	return fmt.Sprintf("churn: prod %d / test %d / atlas %d / workshop %d (final %d, rework %.1f×)",
		m.Churn.Final.CodeProd, m.Churn.Final.CodeTest, m.Churn.Final.Atlas,
		m.Churn.Final.Workshop, m.Churn.FinalTotal, m.Churn.Rework)
}

// GateLine is the D3 operator line. It reports BOTH senses of "disposition" that #187's
// spec conflated — accepted-vs-forced (did the operator work around the gate) and
// addressed/withdrawn/open (did the findings get acted on) — because both are free from
// the ledger and picking one would leave the other unanswerable.
func (m closeCostMetrics) GateLine() string {
	return fmt.Sprintf("plan gate: %d round(s), %d forced; findings %d addressed / %d withdrawn / %d still open",
		m.Rounds, m.Forced, m.Addressed, m.Withdrawn, m.Open)
}

// modelClosingDispositions returns the model's closing dispositions — the set the ledger
// schema must cover. Used by the coverage guard; kept here beside the constants it checks.
func modelClosingDispositions() []string {
	m := vocab.Finding()
	var out []string
	for _, d := range m.AllDispositions() {
		if m.Closes(d) {
			out = append(out, d)
		}
	}
	return out
}
