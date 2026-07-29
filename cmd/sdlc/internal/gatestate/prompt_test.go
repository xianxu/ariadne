package gatestate

import (
	"strings"
	"testing"
)

// The prompt block must (a) list every OPEN finding with its id and severity so the judge
// can dispose them, and (b) list disposed ids so the judge does not re-raise them — the
// A2/A3 contract, and the entire mechanism by which the gate converges.
func TestRenderPriorFindings(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam", "Important/absorb", "Minor/naming")),
		round(2, dispose("PQ-1", "addressed"), nil),
	)
	out := RenderPriorFindings(l)

	for _, want := range []string{
		"2 prior round",
		"OPEN FINDINGS",
		"PQ-2", "Important", "absorb",
		"PQ-3", "Minor", "naming",
		"ALREADY DISPOSED",
		"PQ-1", "addressed", "seam",
		"not-addressed",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("prior-findings block missing %q:\n%s", want, out)
		}
	}

	// The disposed finding must NOT appear in the open list.
	openSection := out[strings.Index(out, "OPEN FINDINGS"):strings.Index(out, "ALREADY DISPOSED")]
	if strings.Contains(openSection, "PQ-1") {
		t.Errorf("an addressed finding must not be listed as open:\n%s", openSection)
	}
}

// Round 1 has no prior state — the block must say so explicitly rather than render empty,
// so the judge knows it is the FIRST reviewer, not one whose history was silently dropped.
func TestRenderPriorFindingsEmpty(t *testing.T) {
	out := RenderPriorFindings(Ledger{Gate: "plan-quality", IDPrefix: "PQ"})
	if !strings.Contains(out, "FIRST round") {
		t.Errorf("empty ledger should announce a first round, got %q", out)
	}
	if strings.Contains(out, "OPEN FINDINGS") {
		t.Error("empty ledger must not render an open-findings list")
	}
}

// Everything disposed: the judge must be told the slate is clear, not handed a bare header
// it might read as "findings were dropped".
func TestRenderPriorFindingsAllDisposed(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam")),
		round(2, dispose("PQ-1", "addressed"), nil),
	)
	out := RenderPriorFindings(l)
	if !strings.Contains(out, "(none — every prior finding has been disposed)") {
		t.Errorf("all-disposed ledger should say so:\n%s", out)
	}
}

// A finding re-disposed `not-addressed` is open again and must be back in the OPEN list,
// not stranded in the disposed one.
func TestRenderPriorFindingsReopened(t *testing.T) {
	l := ledgerWith(
		round(1, nil, findings("Critical/seam")),
		round(2, dispose("PQ-1", "addressed"), nil),
		round(3, dispose("PQ-1", "not-addressed"), nil),
	)
	out := RenderPriorFindings(l)
	openSection := out[strings.Index(out, "OPEN FINDINGS"):]
	if !strings.Contains(openSection, "PQ-1") {
		t.Errorf("a re-opened finding must be listed as open:\n%s", out)
	}
	if strings.Contains(out, "ALREADY DISPOSED") {
		t.Errorf("nothing is settled, so there must be no disposed section:\n%s", out)
	}
}

// Detail text is indented under its finding so a multi-line detail can't be misread as a
// new list entry by the judge.
func TestRenderPriorFindingsIndentsDetail(t *testing.T) {
	l := ledgerWith(round(1, nil, findings("Critical/seam")))
	l.Rounds[0].New[0].Detail = "line one\nline two"
	out := RenderPriorFindings(l)
	if !strings.Contains(out, "      line one\n      line two") {
		t.Errorf("multi-line detail should be indented consistently:\n%s", out)
	}
}
