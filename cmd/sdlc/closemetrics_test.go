package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/estimate"
	"github.com/xianxu/ariadne/cmd/sdlc/internal/gatestate"
)

// The cost report must print on a MILESTONE close, under --no-actual, and with no brain
// dir — the three conditions that gate the calibration ledger row but must NOT gate the
// report. This is the trap Task 13 exists to avoid: emitting the lines from inside
// appendCalibrationRow would have satisfied a hand test on a normal close and silently
// produced nothing on a milestone close, under --no-actual, or in any downstream repo
// without a sibling brain/.
func TestChurnLinePrintsWhenLedgerRowIsSkipped(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*closeFlags)
	}{
		{"milestone", func(f *closeFlags) { f.Milestone = "M1"; f.Actual = "1.0" }},
		{"no-actual", func(f *closeFlags) { f.NoActual = true }},
		{"no-brain", func(f *closeFlags) { f.Actual = "1.0"; f.BrainDir = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			issuesDir := closeRepo(t, 187)
			t.Setenv("WF_CALIB_LEDGER", "") // force the brain-path branch
			f := &closeFlags{Issue: 187, Verified: "done", NoAtlas: true, IssuesDir: issuesDir,
				BrainDir: "../nonexistent-brain"}
			tc.mut(f)

			var errb bytes.Buffer
			if err := runClose(&bytes.Buffer{}, &errb, f); err != nil {
				t.Fatalf("runClose: %v", err)
			}
			out := errb.String()
			if !strings.Contains(out, "churn: prod ") {
				t.Errorf("churn line must print even when the calibration row is skipped:\n%s", out)
			}
			if !strings.Contains(out, "plan gate: ") {
				t.Errorf("gate line must print even when the calibration row is skipped:\n%s", out)
			}
		})
	}
}

// An absent plan-gate sidecar is the normal case for every issue closed before #187. It
// must read as honest zeroes with no warning and no failure — not as an error, and not as
// the "unreadable sidecar" warning, which means something quite different.
func TestCloseWithoutPlanGateSidecarStillCloses(t *testing.T) {
	issuesDir := closeRepo(t, 187)
	t.Setenv("WF_CALIB_LEDGER", "")
	f := &closeFlags{Issue: 187, Actual: "1.0", Verified: "done", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}

	var errb bytes.Buffer
	if err := runClose(&bytes.Buffer{}, &errb, f); err != nil {
		t.Fatalf("a close with no plan-gate sidecar must succeed: %v", err)
	}
	out := errb.String()
	if !strings.Contains(out, "plan gate: 0 round(s), 0 forced") {
		t.Errorf("want an honest zero gate line, got:\n%s", out)
	}
	if strings.Contains(out, "plan-gate metrics not measured") {
		t.Errorf("an ABSENT sidecar is not a measurement failure; that warning is for an unreadable one:\n%s", out)
	}
}

// A sidecar that exists but does not PARSE warns and zeroes — it does not fail the close.
// The inverse posture from `change-code`, deliberately: there, a corrupt ledger is an
// error because silently forgetting findings would re-open work the operator addressed;
// here the ledger is only being read for a diagnostic, and refusing to close over a
// mangled metrics file would hold work hostage to a number nobody is waiting on.
func TestCloseWithCorruptPlanGateSidecarWarnsAndCloses(t *testing.T) {
	issuesDir := closeRepo(t, 187)
	t.Setenv("WF_CALIB_LEDGER", "")
	if err := os.MkdirAll("workshop/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	corrupt := filepath.Join("workshop/plans", "000187-x-plan-gate.md")
	if err := os.WriteFile(corrupt, []byte("---\nnot: [valid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	f := &closeFlags{Issue: 187, Actual: "1.0", Verified: "done", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}

	var errb bytes.Buffer
	if err := runClose(&bytes.Buffer{}, &errb, f); err != nil {
		t.Fatalf("a corrupt sidecar must not fail the close: %v", err)
	}
	out := errb.String()
	if !strings.Contains(out, "plan-gate metrics not measured") {
		t.Errorf("want a warning naming the unmeasured metrics, got:\n%s", out)
	}
	if !strings.Contains(out, "plan gate: 0 round(s)") {
		t.Errorf("want the line to still print with zeroes, got:\n%s", out)
	}
}

// End to end: a real plan-gate ledger and a real commit window must reach the calibration
// ledger's appended columns. Without this the columns could format and parse perfectly and
// still be wired to nothing.
func TestCloseLedgerRowCarriesCostMetrics(t *testing.T) {
	issuesDir := closeRepo(t, 187)
	ledgerPath := filepath.Join(t.TempDir(), "l.tsv")
	t.Setenv("WF_CALIB_LEDGER", ledgerPath)

	// A window with rework: one file written twice, plus a test file.
	commitFile(t, "cmd/x.go", strings.Repeat("a\n", 20), "#187: v1")
	commitFile(t, "cmd/x.go", strings.Repeat("b\n", 20), "#187: v2")
	commitFile(t, "cmd/x_test.go", strings.Repeat("c\n", 10), "#187: tests")

	// A gate ledger: round 1 raises two findings and blocks; round 2 disposes one
	// addressed and one withdrawn, and is forced.
	if err := os.MkdirAll("workshop/plans", 0o755); err != nil {
		t.Fatal(err)
	}
	l := gatestate.Ledger{Gate: planGateKind.Gate, IssueNum: 187, IDPrefix: planGateKind.IDPrefix}
	r1 := gatestate.AssignIDs(l, gatestate.RoundReport{New: []gatestate.Finding{
		{Severity: "Critical", Title: "one"},
		{Severity: "Important", Title: "two"},
	}}, 1, "2026-07-29T00:00:00Z", "claude")
	r1.Blocked = true
	l = gatestate.Apply(l, r1)
	r2 := gatestate.AssignIDs(l, gatestate.RoundReport{Dispositions: []gatestate.Disposition{
		{ID: "PQ-1", State: "addressed"},
		{ID: "PQ-2", State: "withdrawn"},
	}}, 2, "2026-07-29T01:00:00Z", "claude")
	r2.Forced = "shipping the hotfix"
	l = gatestate.Apply(l, r2)
	if err := writePlanGateLedger("workshop/plans", "000187-x.md", l, "ariadne"); err != nil {
		t.Fatal(err)
	}

	f := &closeFlags{Issue: 187, Actual: "2.0", Verified: "done", NoAtlas: true,
		IssuesDir: issuesDir, BrainDir: "../nonexistent-brain"}
	var errb bytes.Buffer
	if err := runClose(&bytes.Buffer{}, &errb, f); err != nil {
		t.Fatalf("runClose: %v", err)
	}

	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("ledger not written: %v", err)
	}
	rows := estimate.ParseRows(string(data))
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d: %q", len(rows), string(data))
	}
	got := rows[0]
	// 20 prod lines survived (x.go was rewritten, not grown) + 10 test lines.
	if got.ChurnProd != 20 || got.ChurnTest != 10 {
		t.Errorf("churn columns = prod %d / test %d, want 20 / 10 (row %q)", got.ChurnProd, got.ChurnTest, string(data))
	}
	// The window is boundaryWindowBase's, which anchors at the PARENT of the first #187
	// commit — so closeRepo's own issue-file commit is in-window and its lines land in the
	// workshop bucket. Rather than hard-coding that fixture's line count, assert the
	// ratio's DEFINITION against the buckets: three of my four commits inserted 20+20+10,
	// the fixture commit inserted whatever the row reports as workshop, and only one of the
	// two x.go writes survives.
	if got.ChurnWorkshop == 0 {
		t.Errorf("the issue-file commit should be in-window as workshop churn: %+v", got)
	}
	wantRework := float64(20+20+10+got.ChurnWorkshop) / float64(20+10+got.ChurnWorkshop)
	if diff := got.Rework - wantRework; diff > 0.01 || diff < -0.01 {
		t.Errorf("Rework = %.2f, want %.2f (commit insertions over final insertions)", got.Rework, wantRework)
	}
	if got.Rework <= 1.0 {
		t.Errorf("Rework = %.2f, but x.go was written twice — the rewrite must show", got.Rework)
	}
	if got.GateRounds != 2 || got.GateForced != 1 {
		t.Errorf("gate_rounds/gate_forced = %d/%d, want 2/1", got.GateRounds, got.GateForced)
	}
	if got.GateAddressed != 1 || got.GateWithdrawn != 1 || got.GateOpen != 0 {
		t.Errorf("dispositions = %d addressed / %d withdrawn / %d open, want 1/1/0",
			got.GateAddressed, got.GateWithdrawn, got.GateOpen)
	}
	// And the operator saw the same numbers the ledger recorded.
	if !strings.Contains(errb.String(), "findings 1 addressed / 1 withdrawn / 0 still open") {
		t.Errorf("operator line disagrees with the ledger row:\n%s", errb.String())
	}
}

// The ledger's column NAMES are literals — a positional append-only TSV cannot derive them
// from the model the way gatestate's logic does. This is the guard that keeps the schema
// and the model in step: adding a closing disposition to finding.cue without a column
// would silently drop those findings out of gate_addressed/gate_withdrawn, under-reporting
// the one metric meant to answer "did the gate's findings get acted on?".
func TestLedgerCoversEveryClosingDisposition(t *testing.T) {
	haveColumn := map[string]bool{dispAddressed: true, dispWithdrawn: true}
	closing := modelClosingDispositions()
	if len(closing) == 0 {
		t.Fatal("the finding model declares no closing dispositions — the guard would pass vacuously")
	}
	for _, d := range closing {
		if !haveColumn[d] {
			t.Errorf("finding.cue declares closing disposition %q with no ledger column: its findings "+
				"would vanish from the counts. APPEND a gate_%s column (never insert) and map it in closeMetrics.", d, d)
		}
	}
}

// An unset PlansDir must resolve to the convention, not to the repo root. closeFlags is
// built directly by milestoneCloseFlags.closeFlags(), so a field only cobra populates is
// empty on that production path — and an empty plans dir fails silently in the worst way:
// every issue reads as "no plan-gate ledger" and sidecars land beside the Makefile.
func TestResolvePlansDirDefaults(t *testing.T) {
	t.Setenv("WF_PLANS_DIR", "")
	if got := resolvePlansDir(""); got != "workshop/plans" {
		t.Errorf("resolvePlansDir(\"\") = %q, want workshop/plans", got)
	}
	if got := resolvePlansDir("custom/plans"); got != "custom/plans" {
		t.Errorf("an explicit value must win: got %q", got)
	}
	t.Setenv("WF_PLANS_DIR", "env/plans")
	if got := resolvePlansDir(""); got != "env/plans" {
		t.Errorf("WF_PLANS_DIR must be honored: got %q", got)
	}
	if got := (&closeFlags{}).plansDir(); got != "env/plans" {
		t.Errorf("closeFlags.plansDir() must route through resolvePlansDir: got %q", got)
	}
}

// milestone-close translates its own flags into closeFlags; PlansDir must survive the
// translation or the cost report reads the wrong directory on every milestone close.
func TestMilestoneCloseFlagsPropagatePlansDir(t *testing.T) {
	mf := &milestoneCloseFlags{Issue: 187, Milestone: "M1", PlansDir: "custom/plans"}
	if got := mf.closeFlags().plansDir(); got != "custom/plans" {
		t.Errorf("closeFlags().plansDir() = %q, want custom/plans", got)
	}
}
