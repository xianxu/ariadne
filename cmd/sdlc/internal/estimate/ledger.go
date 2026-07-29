package estimate

import (
	"strconv"
	"strings"
)

// LedgerRow is one estimate↔actual calibration data point, appended at `sdlc
// close`. WindowTrusted records whether Actual came from a `started:`-windowed
// measurement (#116) or the legacy first-commit-parent window — untrusted rows
// are excluded from drift stats (the #68 posture: a truncated actual must not
// masquerade as a clean data point).
type LedgerRow struct {
	Issue         string
	Estimate      float64
	EstDesign     float64
	EstImpl       float64
	Actual        float64
	Model         string
	Mode          string // supervised | delegated | "" (unknown)
	WindowTrusted bool
	Date          string // ISO date

	// ── #187 D1–D3: what the work COST, beside what it was estimated at ──────────
	// Ten appended columns (indices 10–19). They answer questions an estimate↔actual
	// pair cannot: where the hours went, how much was rewritten, and whether the plan
	// gate earned its round-trips. All are zero on a legacy row and on any close where
	// measurement failed — a metric that degraded must read as absent, not as a real 0
	// that happens to look the same. (It does look the same in the TSV; the honesty
	// lives in the warning `sdlc close` prints when it zeroes them.)
	ChurnProd     int
	ChurnTest     int
	ChurnAtlas    int
	ChurnWorkshop int
	// Rework is insertions-across-commits over insertions-in-the-final-diff. 1.0 means
	// nothing was rewritten.
	Rework float64
	// GateRounds/GateForced: how many plan-quality rounds this issue took, and how many
	// were bypassed with --force. Together they are the "did the gate earn its cost"
	// numerator and denominator.
	GateRounds int
	GateForced int
	// GateAddressed/GateWithdrawn/GateOpen: how the gate's findings were RESOLVED —
	// distinct from accepted-vs-forced, and the number that answers "did the findings get
	// acted on, or worked around?". Still-open findings at close are the ones demoted past
	// the round cap or filed Minor, carried into the boundary review.
	GateAddressed int
	GateWithdrawn int
	GateOpen      int
}

// Ratio is estimate/actual (0 when actual is 0, to avoid div-by-zero).
func (r LedgerRow) Ratio() float64 {
	if r.Actual == 0 {
		return 0
	}
	return r.Estimate / r.Actual
}

// ledgerHeader is the column order, and the order is a CONTRACT: columns are APPENDED,
// never reordered or inserted. ParseRows indexes positionally and live ledgers across the
// fleet are full of rows written by older binaries, so an insertion would not fail — it
// would silently re-interpret every historical row. The #187 block (churn_prod onward)
// occupies indices 10–19.
const ledgerHeader = "issue\testimate\test_design\test_impl\tactual\tratio\tmodel\tmode\twindow_trusted\tdate" +
	"\tchurn_prod\tchurn_test\tchurn_atlas\tchurn_workshop\trework" +
	"\tgate_rounds\tgate_forced\tgate_addressed\tgate_withdrawn\tgate_open"

// legacyCols is the column count before #187's append — the boundary between "row from an
// older binary" and "row carrying cost metrics".
const legacyCols = 10

// Header returns the ledger's TSV header line (written when the file is created).
func Header() string { return ledgerHeader }

// FormatRow renders one tab-separated ledger line in a stable column order.
func FormatRow(r LedgerRow) string {
	return strings.Join([]string{
		r.Issue,
		ftoa(r.Estimate), ftoa(r.EstDesign), ftoa(r.EstImpl), ftoa(r.Actual), ftoa(r.Ratio()),
		r.Model, dash(r.Mode), yesno(r.WindowTrusted), r.Date,
		itoa(r.ChurnProd), itoa(r.ChurnTest), itoa(r.ChurnAtlas), itoa(r.ChurnWorkshop), ftoa(r.Rework),
		itoa(r.GateRounds), itoa(r.GateForced), itoa(r.GateAddressed), itoa(r.GateWithdrawn), itoa(r.GateOpen),
	}, "\t")
}

// ParseRows parses ledger lines into rows, skipping the header, blanks, and
// `#`-comments. The ratio column is recomputed from estimate/actual, not read.
func ParseRows(text string) []LedgerRow {
	var rows []LedgerRow
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "issue\t") || strings.HasPrefix(line, "#") {
			continue
		}
		c := strings.Split(line, "\t")
		if len(c) < legacyCols {
			continue
		}
		est, ok := parseLedgerFloat(c[1])
		if !ok {
			continue
		}
		estDesign, ok := parseLedgerFloat(c[2])
		if !ok {
			continue
		}
		estImpl, ok := parseLedgerFloat(c[3])
		if !ok {
			continue
		}
		actual, ok := parseLedgerFloat(c[4])
		if !ok {
			continue
		}
		row := LedgerRow{
			Issue:         c[0],
			Estimate:      est,
			EstDesign:     estDesign,
			EstImpl:       estImpl,
			Actual:        actual,
			Model:         c[6],
			Mode:          undash(c[7]),
			WindowTrusted: c[8] == "yes",
			Date:          c[9],
		}
		// The #187 metrics are read only when the WHOLE appended block is present. The
		// check is the full width, not a partial one: a 19-column row would satisfy any
		// looser bound and then panic on c[19]. A short row keeps its estimate↔actual data
		// point and simply carries no metrics — losing the row entirely would throw away
		// the calibration history this ledger exists for.
		if len(c) >= len(strings.Split(ledgerHeader, "\t")) {
			row.ChurnProd = atoiOrZero(c[10])
			row.ChurnTest = atoiOrZero(c[11])
			row.ChurnAtlas = atoiOrZero(c[12])
			row.ChurnWorkshop = atoiOrZero(c[13])
			row.Rework, _ = parseLedgerFloat(c[14])
			row.GateRounds = atoiOrZero(c[15])
			row.GateForced = atoiOrZero(c[16])
			row.GateAddressed = atoiOrZero(c[17])
			row.GateWithdrawn = atoiOrZero(c[18])
			row.GateOpen = atoiOrZero(c[19])
		}
		rows = append(rows, row)
	}
	return rows
}

func ftoa(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
func itoa(v int) string     { return strconv.Itoa(v) }

// atoiOrZero reads a metric column, yielding 0 on anything unparseable. Metrics are
// diagnostic: a mangled column must not cost the row its estimate↔actual data point.
func atoiOrZero(s string) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return v
}
func parseLedgerFloat(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
func undash(s string) string {
	if s == "-" {
		return ""
	}
	return s
}
func yesno(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
