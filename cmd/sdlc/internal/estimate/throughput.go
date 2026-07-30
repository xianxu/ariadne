package estimate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// throughput.go — measured focused-hours/week over an operator-blessed span
// (#182). A project's calendar forecast divides remaining issue-hours by this
// rate; both sides are ship wall-clock engineer+AI hours (#118), so the
// division needs no unit conversion. The number is MEASURED from the
// calibration ledger (one row per issue close); the operator only picks the
// span — trailing windows skew under vacations/crunch, so the representative
// span is blessed deliberately.
//
// The ledger holds one row per CLOSE, not per issue: re-closing a done issue is legal and each
// re-close measures a longer cumulative window of the same work. Every reader must therefore
// dedupe per issue (NewestPerIssue). Not doing so inflated the blessed baseline 1.41x — see
// ariadne#192.

// SpanMeasure is the result of measuring throughput over a date span.
type SpanMeasure struct {
	HoursPerWeek float64
	// Issues is the number of DISTINCT issues summed — the ledger is written per CLOSE, so a
	// re-closed issue contributes one observation, at its newest measurement (ariadne#192).
	Issues int
	// RowsScanned is how many in-span ledger rows were seen. The GAP between this and Issues
	// is the re-close duplication; reporting both makes a future recurrence visible instead of
	// silent — the old single `Rows` field read as an issue count and hid a 1.41x inflation.
	RowsScanned int
	// UntrustedRows counts issues whose NEWEST measurement is window_trusted=no (their hours
	// still count; reported so bless can warn). Counted over the DEDUPED set so it shares a
	// denominator with Issues — raw-counted it could print "12 of 8".
	UntrustedRows int
	Skipped       int // rows with an unparsable date (excluded)
	Days          int // inclusive span length
}

const isoDate = "2006-01-02"

// SpanThroughput measures focused hours/week over [from, to] (inclusive, ISO
// dates) from parsed calibration-ledger rows. Pure: no IO. It consumes the
// existing LedgerRow shape (Actual already a float, Date an ISO string) — a
// row whose Date won't parse is counted in Skipped, not summed. Untrusted rows
// ARE included in the sum (their measured hours are real; excluding them would
// under-count the rate), with their count reported so callers can warn.
// Errors on unparsable bounds, to<from, or a span containing zero rows.
func SpanThroughput(rows []LedgerRow, from, to string) (SpanMeasure, error) {
	fromT, err := time.Parse(isoDate, from)
	if err != nil {
		return SpanMeasure{}, fmt.Errorf("bad from-date %q: %w", from, err)
	}
	toT, err := time.Parse(isoDate, to)
	if err != nil {
		return SpanMeasure{}, fmt.Errorf("bad to-date %q: %w", to, err)
	}
	if toT.Before(fromT) {
		return SpanMeasure{}, fmt.Errorf("span end %s precedes start %s", to, from)
	}
	days := int(toT.Sub(fromT).Hours()/24) + 1 // inclusive

	m := SpanMeasure{Days: days}
	// FILTER to the span first (counting unparsable dates), THEN dedupe per issue, THEN sum.
	// The order matters: deduping before the span filter would let an out-of-span re-close
	// mask the in-span measurement it superseded.
	inSpan := make([]LedgerRow, 0, len(rows))
	for _, r := range rows {
		d, derr := time.Parse(isoDate, r.Date)
		if derr != nil {
			m.Skipped++
			continue
		}
		if d.Before(fromT) || d.After(toT) {
			continue
		}
		inSpan = append(inSpan, r)
	}
	m.RowsScanned = len(inSpan)
	var sum float64
	for _, r := range NewestPerIssue(inSpan) {
		sum += r.Actual
		m.Issues++
		if !r.WindowTrusted {
			m.UntrustedRows++
		}
	}
	if m.Issues == 0 {
		return SpanMeasure{}, fmt.Errorf("no calibration-ledger rows fall in %s..%s — pick a span that contains closed issues", from, to)
	}
	m.HoursPerWeek = sum / (float64(days) / 7.0)
	return m, nil
}

// ThroughputBaseline is one blessed row: a span the operator designated as
// representative, its measured rate, and the attention ceiling in force.
// The baseline TSV is append-only — the last row is the current baseline,
// earlier rows are the re-blessing history.
type ThroughputBaseline struct {
	BlessedDate  string
	SpanStart    string
	SpanEnd      string
	HoursPerWeek float64
	// Rows records the number of observations behind HoursPerWeek. Since ariadne#192 that is
	// the count of distinct ISSUES; rows blessed BEFORE #192 recorded raw ledger lines and are
	// therefore not comparable (the 2026-07-19 row's 280 is ~1.41x inflated). The column keeps
	// its name deliberately: ParseBaselineTSV reads 6 columns positionally from an append-only
	// file, so renaming it would leave the stored header describing one thing and new rows
	// meaning another.
	Rows    int
	Ceiling int
}

const baselineHeader = "blessed_date\tspan_start\tspan_end\thours_per_week\trows\tceiling"

// BaselineHeader returns the throughput-baseline TSV header (written on create).
func BaselineHeader() string { return baselineHeader }

// RenderBaselineRow renders one baseline row as a tab-separated line.
func RenderBaselineRow(b ThroughputBaseline) string {
	return strings.Join([]string{
		b.BlessedDate, b.SpanStart, b.SpanEnd,
		strconv.FormatFloat(b.HoursPerWeek, 'f', 2, 64),
		strconv.Itoa(b.Rows), strconv.Itoa(b.Ceiling),
	}, "\t")
}

// ParseBaselineTSV parses the baseline file, skipping the header, blanks, and
// `#`-comments. Rows are returned in file order (last = current). A malformed
// numeric column is an error (a corrupt baseline must not silently read as 0).
func ParseBaselineTSV(text string) ([]ThroughputBaseline, error) {
	var out []ThroughputBaseline
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "blessed_date\t") {
			continue
		}
		c := strings.Split(line, "\t")
		if len(c) < 6 {
			return nil, fmt.Errorf("baseline row has %d columns, want 6: %q", len(c), line)
		}
		hpw, err := strconv.ParseFloat(c[3], 64)
		if err != nil {
			return nil, fmt.Errorf("baseline hours_per_week %q: %w", c[3], err)
		}
		rows, err := strconv.Atoi(c[4])
		if err != nil {
			return nil, fmt.Errorf("baseline rows %q: %w", c[4], err)
		}
		ceiling, err := strconv.Atoi(c[5])
		if err != nil {
			return nil, fmt.Errorf("baseline ceiling %q: %w", c[5], err)
		}
		out = append(out, ThroughputBaseline{
			BlessedDate: c[0], SpanStart: c[1], SpanEnd: c[2],
			HoursPerWeek: hpw, Rows: rows, Ceiling: ceiling,
		})
	}
	return out, nil
}
