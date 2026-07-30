package estimate

import (
	"strings"
	"testing"
)

// A calibration-ledger fixture in the REAL 10-column shape (issue estimate
// est_design est_impl actual ratio model mode window_trusted date). SpanThroughput
// consumes estimate.ParseRows output — no new parser.
const ledgerFixture = `# comment line, skipped
issue	estimate	est_design	est_impl	actual	ratio	model	mode	window_trusted	date
a#1	1.00	0.00	0.00	4.00	0.25	m	-	yes	2026-06-01
a#2	1.00	0.00	0.00	6.00	0.17	m	-	no	2026-06-08
a#3	1.00	0.00	0.00	10.00	0.10	m	-	yes	2026-06-28
a#4	1.00	0.00	0.00	99.00	0.01	m	-	yes	2026-05-01
a#5	1.00	0.00	0.00	2.00	0.50	m	-	yes	not-a-date`

func TestSpanThroughput_28DaySpan(t *testing.T) {
	rows := ParseRows(ledgerFixture)
	// span 2026-06-01..2026-06-28 inclusive = 28 days = 4.0 weeks.
	// in-span rows: a#1 (4), a#2 (6, untrusted but counted), a#3 (10) = 20h.
	// a#4 is 2026-05-01 (out), a#5 has a bad date (skipped+counted).
	m, err := SpanThroughput(rows, "2026-06-01", "2026-06-28")
	if err != nil {
		t.Fatal(err)
	}
	if m.Days != 28 {
		t.Errorf("Days = %d, want 28", m.Days)
	}
	if got := m.HoursPerWeek; got < 4.99 || got > 5.01 {
		t.Errorf("HoursPerWeek = %.4f, want 5.00 (20h / 4wk)", got)
	}
	// Issues, not rows: the fixture has no re-closes, so they coincide here — asserting both
	// keeps that coincidence honest rather than implicit.
	if m.Issues != 3 {
		t.Errorf("Issues = %d, want 3", m.Issues)
	}
	if m.RowsScanned != 3 {
		t.Errorf("RowsScanned = %d, want 3 (no re-closes in this fixture)", m.RowsScanned)
	}
	if m.UntrustedRows != 1 {
		t.Errorf("UntrustedRows = %d, want 1 (a#2)", m.UntrustedRows)
	}
	// a#5's actual parses fine (2.00) but its DATE is unparsable → Skipped.
	// (a bad ACTUAL never reaches us — ParseRows drops it upstream.)
	if m.Skipped != 1 {
		t.Errorf("Skipped = %d, want 1 (a#5 bad date)", m.Skipped)
	}
}

func TestSpanThroughput_PartialWeek(t *testing.T) {
	rows := ParseRows(ledgerFixture)
	// 2026-06-01..2026-06-27 = 27 days; in-span trusted+untrusted actuals
	// (a#1 4, a#2 6) = 10h (a#3 is the 28th, excluded). 10 / (27/7) = 2.5926.
	m, err := SpanThroughput(rows, "2026-06-01", "2026-06-27")
	if err != nil {
		t.Fatal(err)
	}
	if m.Days != 27 {
		t.Errorf("Days = %d, want 27", m.Days)
	}
	want := 10.0 / (27.0 / 7.0)
	if got := m.HoursPerWeek; got < want-0.001 || got > want+0.001 {
		t.Errorf("HoursPerWeek = %.4f, want %.4f", got, want)
	}
}

func TestSpanThroughput_EmptySpan(t *testing.T) {
	rows := ParseRows(ledgerFixture)
	// A span with no in-range rows must error (nothing to measure).
	if _, err := SpanThroughput(rows, "2026-01-01", "2026-01-31"); err == nil {
		t.Fatal("want error for a span containing no ledger rows")
	}
}

func TestSpanThroughput_BadSpanBounds(t *testing.T) {
	rows := ParseRows(ledgerFixture)
	if _, err := SpanThroughput(rows, "bad", "2026-06-28"); err == nil {
		t.Error("want error for an unparsable from-date")
	}
	if _, err := SpanThroughput(rows, "2026-06-28", "2026-06-01"); err == nil {
		t.Error("want error when to < from")
	}
}

func TestBaselineTSV_RoundTrip(t *testing.T) {
	rows := []ThroughputBaseline{
		{BlessedDate: "2026-07-01", SpanStart: "2026-06-01", SpanEnd: "2026-06-28", HoursPerWeek: 5.0, Rows: 3, Ceiling: 2},
		{BlessedDate: "2026-07-19", SpanStart: "2026-06-22", SpanEnd: "2026-07-19", HoursPerWeek: 110.0, Rows: 40, Ceiling: 3},
	}
	var b strings.Builder
	b.WriteString("# throughput baseline\n")
	b.WriteString(BaselineHeader() + "\n")
	for _, r := range rows {
		b.WriteString(RenderBaselineRow(r) + "\n")
	}
	parsed, err := ParseBaselineTSV(b.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != 2 {
		t.Fatalf("parsed %d rows, want 2", len(parsed))
	}
	// Last row is the current baseline.
	cur := parsed[len(parsed)-1]
	if cur.HoursPerWeek != 110.0 || cur.Ceiling != 3 || cur.SpanStart != "2026-06-22" {
		t.Errorf("current baseline wrong: %+v", cur)
	}
}

func TestParseBaselineTSV_BadFloat(t *testing.T) {
	text := BaselineHeader() + "\n2026-07-01\t2026-06-01\t2026-06-28\tNOPE\t3\t2\n"
	if _, err := ParseBaselineTSV(text); err == nil {
		t.Error("want error on an unparsable hours_per_week")
	}
}

func TestParseBaselineTSV_Empty(t *testing.T) {
	text := "# only comments\n" + BaselineHeader() + "\n"
	rows, err := ParseBaselineTSV(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Errorf("want 0 rows from a header-only file, got %d", len(rows))
	}
}

// A re-closed issue in the REAL ariadne#167 shape: seven closes, each measuring a LONGER
// cumulative window of the same work. Summed naively they contribute 14.57h for work that
// took 2.71h — 5.4x (ariadne#192).
const reCloseFixture = `issue	estimate	est_design	est_impl	actual	ratio	model	mode	window_trusted	date
a#167	1.08	0.00	0.00	1.43	0.76	m	-	yes	2026-06-05
a#167	1.08	0.00	0.00	1.90	0.57	m	-	yes	2026-06-06
a#167	1.08	0.00	0.00	2.71	0.40	m	-	yes	2026-06-07
a#9	1.00	0.00	0.00	1.00	1.00	m	-	yes	2026-06-08`

// The defect: SpanThroughput must count a re-closed issue ONCE, at its newest measurement.
// The ledger is written per CLOSE, not per issue, and re-closing is legal.
func TestSpanThroughput_CountsReClosedIssueOnce(t *testing.T) {
	rows := ParseRows(reCloseFixture)
	// 2026-06-05..2026-06-11 = 7 days = exactly 1 week.
	// Correct: a#167 newest (2.71) + a#9 (1.00) = 3.71h → 3.71 h/wk over 2 issues.
	// Buggy:   1.43+1.90+2.71+1.00 = 7.04h → 7.04 h/wk over "4 rows".
	m, err := SpanThroughput(rows, "2026-06-05", "2026-06-11")
	if err != nil {
		t.Fatal(err)
	}
	if got := m.HoursPerWeek; got < 3.70 || got > 3.72 {
		t.Errorf("HoursPerWeek = %.4f, want 3.71 — a re-closed issue's partial sums were counted as separate work", got)
	}
	if m.Issues != 2 {
		t.Errorf("Issues = %d, want 2 (a#167, a#9)", m.Issues)
	}
	if m.RowsScanned != 4 {
		t.Errorf("RowsScanned = %d, want 4 — the gap between this and Issues IS the duplication", m.RowsScanned)
	}
}

// The equivalence the fix asserts: measuring a span with a re-closed issue equals measuring it
// with only that issue's last row present. Stated as its own test because it is the property,
// independent of the specific numbers above.
func TestSpanThroughput_ReCloseEqualsLastRowAlone(t *testing.T) {
	full := ParseRows(reCloseFixture)
	lastOnly := ParseRows(`issue	estimate	est_design	est_impl	actual	ratio	model	mode	window_trusted	date
a#167	1.08	0.00	0.00	2.71	0.40	m	-	yes	2026-06-07
a#9	1.00	0.00	0.00	1.00	1.00	m	-	yes	2026-06-08`)
	a, err := SpanThroughput(full, "2026-06-05", "2026-06-11")
	if err != nil {
		t.Fatal(err)
	}
	b, err := SpanThroughput(lastOnly, "2026-06-05", "2026-06-11")
	if err != nil {
		t.Fatal(err)
	}
	if a.HoursPerWeek != b.HoursPerWeek || a.Issues != b.Issues {
		t.Errorf("re-closes changed the measure: %+v vs last-row-only %+v", a, b)
	}
}

// UntrustedRows must be counted over the DEDUPED set, or the warning at
// projectthroughput.go:103 prints mixed denominators ("12 of 8 rows").
func TestSpanThroughput_UntrustedCountedOverDedupedSet(t *testing.T) {
	// a#7's newest row is TRUSTED, so the issue must not count as untrusted even though an
	// earlier close of it was not.
	rows := ParseRows(`issue	estimate	est_design	est_impl	actual	ratio	model	mode	window_trusted	date
a#7	1.00	0.00	0.00	1.00	1.00	m	-	no	2026-06-05
a#7	1.00	0.00	0.00	2.00	0.50	m	-	yes	2026-06-06`)
	m, err := SpanThroughput(rows, "2026-06-05", "2026-06-11")
	if err != nil {
		t.Fatal(err)
	}
	if m.Issues != 1 {
		t.Fatalf("Issues = %d, want 1", m.Issues)
	}
	if m.UntrustedRows > m.Issues {
		t.Errorf("UntrustedRows %d > Issues %d — mixed denominators", m.UntrustedRows, m.Issues)
	}
	if m.UntrustedRows != 0 {
		t.Errorf("UntrustedRows = %d, want 0 — the issue's NEWEST measurement is trusted", m.UntrustedRows)
	}
}
