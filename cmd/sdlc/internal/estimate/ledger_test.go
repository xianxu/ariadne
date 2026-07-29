package estimate

import (
	"strings"
	"testing"
)

func TestFormatRow_Stable(t *testing.T) {
	r := LedgerRow{
		Issue: "ariadne#117", Estimate: 3.4, EstDesign: 0.91, EstImpl: 2.5,
		Actual: 1.7, Model: "estimate-logic-v2", Mode: "supervised",
		WindowTrusted: false, Date: "2026-06-17",
	}
	got := FormatRow(r)
	cols := strings.Split(got, "\t")
	// Derived from Header() rather than a literal: the count grew from 10 to 20 in #187,
	// and an assertion that restates the number just has to be edited alongside it
	// without ever having caught anything. What matters is that row and header agree.
	if want := len(strings.Split(Header(), "\t")); len(cols) != want {
		t.Fatalf("FormatRow produced %d columns, header has %d: %q", len(cols), want, got)
	}
	if cols[0] != "ariadne#117" || cols[6] != "estimate-logic-v2" || cols[8] != "no" {
		t.Errorf("unexpected columns: %v", cols)
	}
	if cols[5] != "2.00" { // ratio 3.4/1.7
		t.Errorf("ratio col = %q, want 2.00", cols[5])
	}
}

// Appending columns must not break the reader for PRE-EXISTING 10-column rows. Every
// ledger in the fleet is full of them, and projectthroughput_test.go seeds them as
// fixtures — a presence check that read the new columns unconditionally would drop the
// entire calibration history rather than just the metrics it lacks.
func TestParseRowsAcceptsLegacyTenColumnRows(t *testing.T) {
	legacy := Header() + "\nariadne#1\t2\t1\t1\t3\t0.67\tm\t-\tyes\t2026-01-01\n"
	rows := ParseRows(legacy)
	if len(rows) != 1 || rows[0].Issue != "ariadne#1" {
		t.Fatalf("legacy row lost: %+v", rows)
	}
	if rows[0].Actual != 3 || rows[0].Date != "2026-01-01" {
		t.Errorf("legacy row misread: %+v", rows[0])
	}
	// The #187 metrics are simply absent, not garbage.
	if rows[0].ChurnProd != 0 || rows[0].GateRounds != 0 || rows[0].Rework != 0 {
		t.Errorf("legacy row should carry zero churn/gate metrics: %+v", rows[0])
	}
}

func TestFormatRowCarriesChurnColumns(t *testing.T) {
	r := LedgerRow{Issue: "ariadne#187", Estimate: 4, Actual: 5,
		ChurnProd: 554, ChurnTest: 300, ChurnAtlas: 20, ChurnWorkshop: 778,
		Rework: 2.4, GateRounds: 6, GateForced: 1,
		GateAddressed: 2, GateWithdrawn: 1, GateOpen: 3}
	cols := strings.Split(FormatRow(r), "\t")
	if len(cols) != len(strings.Split(Header(), "\t")) {
		t.Fatalf("row has %d columns, header has %d", len(cols), len(strings.Split(Header(), "\t")))
	}
	// Positional, because ParseRows indexes positionally: these assertions are what makes
	// "append, never reorder" enforceable rather than a comment.
	for i, want := range map[int]string{
		10: "554", 11: "300", 12: "20", 13: "778", 14: "2.40",
		15: "6", 16: "1", 17: "2", 18: "1", 19: "3",
	} {
		if cols[i] != want {
			t.Errorf("col %d = %q, want %q (header: %s)", i, cols[i], want,
				strings.Split(Header(), "\t")[i])
		}
	}
}

// The full round trip through the new columns — a metric that formats but does not parse
// back is a column of write-only noise.
func TestRoundTripChurnColumns(t *testing.T) {
	in := LedgerRow{Issue: "ariadne#187", Estimate: 8.45, Actual: 4.2, Date: "2026-07-29",
		ChurnProd: 554, ChurnTest: 300, ChurnAtlas: 20, ChurnWorkshop: 778,
		Rework: 2.4, GateRounds: 6, GateForced: 1,
		GateAddressed: 2, GateWithdrawn: 1, GateOpen: 3}
	rows := ParseRows(Header() + "\n" + FormatRow(in) + "\n")
	if len(rows) != 1 {
		t.Fatalf("ParseRows returned %d rows, want 1", len(rows))
	}
	got := rows[0]
	if got.ChurnProd != 554 || got.ChurnTest != 300 || got.ChurnAtlas != 20 || got.ChurnWorkshop != 778 {
		t.Errorf("churn columns lost: %+v", got)
	}
	if got.Rework != 2.4 || got.GateRounds != 6 || got.GateForced != 1 {
		t.Errorf("rework/round columns lost: %+v", got)
	}
	if got.GateAddressed != 2 || got.GateWithdrawn != 1 || got.GateOpen != 3 {
		t.Errorf("disposition columns lost: %+v", got)
	}
}

// A 19-column row is malformed, not legacy: reading it as if the last column were there
// would panic on c[19]. The presence check is `>= 20` for exactly this reason, so pin it.
func TestParseRowsSkipsNineteenColumnRow(t *testing.T) {
	cols := make([]string, 19)
	for i := range cols {
		cols[i] = "0"
	}
	cols[0] = "ariadne#9"
	text := Header() + "\n" + strings.Join(cols, "\t") + "\n"
	rows := ParseRows(text)
	// It parses as a legacy row (the first 10 columns are well-formed) and simply carries
	// no metrics — what must NOT happen is a panic or a half-read metric block.
	if len(rows) != 1 {
		t.Fatalf("want the row read as legacy, got %d rows", len(rows))
	}
	if rows[0].GateOpen != 0 || rows[0].ChurnProd != 0 {
		t.Errorf("a 19-column row must contribute no metrics: %+v", rows[0])
	}
}

func TestFormatRow_EmptyModeDashed(t *testing.T) {
	got := FormatRow(LedgerRow{Issue: "x", Actual: 1, Estimate: 1})
	if !strings.Contains(got, "\t-\t") {
		t.Errorf("empty mode should render as '-': %q", got)
	}
}

func TestRoundTrip(t *testing.T) {
	rows := []LedgerRow{
		{Issue: "ariadne#110", Estimate: 5, EstDesign: 0.4, EstImpl: 4.3, Actual: 0.89, Model: "estimate-logic-v2", WindowTrusted: false, Date: "2026-06-16"},
		{Issue: "ariadne#117", Estimate: 3.4, EstDesign: 0.91, EstImpl: 2.5, Actual: 1.7, Model: "estimate-logic-v2", Mode: "supervised", WindowTrusted: true, Date: "2026-06-17"},
	}
	var sb strings.Builder
	sb.WriteString(Header() + "\n")
	for _, r := range rows {
		sb.WriteString(FormatRow(r) + "\n")
	}
	got := ParseRows(sb.String())
	if len(got) != 2 {
		t.Fatalf("ParseRows returned %d rows, want 2", len(got))
	}
	if got[0].Issue != "ariadne#110" || got[0].WindowTrusted {
		t.Errorf("row0 mismatch: %+v", got[0])
	}
	if got[1].Mode != "supervised" || !got[1].WindowTrusted || got[1].Actual != 1.7 {
		t.Errorf("row1 mismatch: %+v", got[1])
	}
}

func TestParseRows_SkipsHeaderAndComments(t *testing.T) {
	text := Header() + "\n# a comment\n\n" +
		FormatRow(LedgerRow{Issue: "x", Estimate: 1, Actual: 1, Model: "estimate-logic-v2", Date: "2026-06-17"}) + "\n"
	if got := ParseRows(text); len(got) != 1 {
		t.Fatalf("expected 1 data row, got %d", len(got))
	}
}

func TestParseRows_SkipsRaggedRows(t *testing.T) {
	text := Header() + "\nonly\ttwo\tcols\n" +
		FormatRow(LedgerRow{Issue: "x", Estimate: 1, Actual: 1, Model: "estimate-logic-v2", Date: "2026-06-17"}) + "\n"
	if got := ParseRows(text); len(got) != 1 {
		t.Fatalf("ragged (<10 col) row should be skipped; want 1 data row, got %d", len(got))
	}
}

func TestParseRows_SkipsNotApplicableActualRows(t *testing.T) {
	text := Header() + "\n" +
		"ariadne#135\t2.00\t0.20\t1.80\tN/A\t-\testimate-logic-v2\t-\tyes\t2026-06-26\n" +
		FormatRow(LedgerRow{Issue: "x", Estimate: 1, Actual: 1, Model: "estimate-logic-v2", Date: "2026-06-26"}) + "\n"
	if got := ParseRows(text); len(got) != 1 {
		t.Fatalf("N/A actual row should be skipped; got %d rows", len(got))
	}
}
