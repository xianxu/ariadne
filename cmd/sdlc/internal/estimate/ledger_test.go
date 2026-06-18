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
	if len(cols) != 10 {
		t.Fatalf("FormatRow produced %d columns, want 10: %q", len(cols), got)
	}
	if cols[0] != "ariadne#117" || cols[6] != "estimate-logic-v2" || cols[8] != "no" {
		t.Errorf("unexpected columns: %v", cols)
	}
	if cols[5] != "2.00" { // ratio 3.4/1.7
		t.Errorf("ratio col = %q, want 2.00", cols[5])
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
