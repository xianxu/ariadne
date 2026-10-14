package estimate

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// LedgerRow is one estimate↔actual calibration data point, appended at `sdlc
// close`. WindowTrusted records whether Actual came from a `started:`-windowed
// measurement (#116) or the legacy first-commit-parent window — untrusted rows
// are excluded from drift stats (the #68 posture: a truncated actual must not
// masquerade as a clean data point).
type LedgerRow struct {
	// RawID identifies the exact original TSV bytes, excluding the line feed.
	RawID      string
	Correction *CorrectionRecord

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

	// ── #231: which flow the issue closed under (indices 20–22) ─────────────────
	// FlowKind is "full" or "quick" (empty on a row written before #231);
	// FlowProvenance "inferred" or "operator" (empty when the issue had no flow
	// record); FlowUpgraded marks a quick issue the close moved to full. Quick and
	// upgraded rows carry no estimate, so they are not estimate↔actual data points
	// (see ExcludedFromCalibration) — but their hours are real hours.
	FlowKind       string
	FlowProvenance string
	FlowUpgraded   bool

	// ── #234: the actual split at the boundary reviews (indices 23–27) ─────────
	// BuildHours runs to each boundary's first review dispatch, ConvergeHours from
	// there to the boundary's last #N commit; on a measured row they sum to Actual.
	// BoundaryRounds counts the boundary-review rounds; ReviewRecipe names the
	// recipes they ran ("+"-joined when mixed). SplitSource is the evidence for the
	// cuts: dispatch | finalize | no-review | manual (typed --actual, no split), or
	// "" on a row written before #234 — whose Actual still counts every earlier
	// boundary's verification tail and is not comparable with a split row.
	BuildHours     float64
	ConvergeHours  float64
	BoundaryRounds int
	SplitSource    string
	ReviewRecipe   string

	// ── #202: findings settled by implementer dispute (indices 28–29) ──────────
	// GateDisputed sits beside GateAddressed/GateWithdrawn for the plan gate;
	// BoundaryDisputed counts the boundary review's, where disputes mostly land. A
	// high count against the gate's findings is the abuse signal: a dispute is the
	// gate's escape hatch, and its use must show up in history.
	GateDisputed     int
	BoundaryDisputed int
}

// HasSplit reports a row whose actual was measured and split at its boundary
// reviews (#234): the only rows est-vs-build and converge factors read.
func (r LedgerRow) HasSplit() bool {
	return r.SplitSource != "" && r.SplitSource != "manual"
}

// ExcludedFromCalibration reports a row that is no estimate↔actual data point:
// a quick-flow close (no estimate at all) or one upgraded from quick (it started
// without one). Drift, the project forecast's calibration and any future
// est/actual reader skip it (#231).
//
// A row with no estimate is excluded whatever its flow (#234 E3): a full-flow
// close that skipped the estimate gate wrote est 0, ratio 0 and passed as a data
// point — and as drift's "latest" row, its empty model switched drift off.
func (r LedgerRow) ExcludedFromCalibration() bool {
	return r.FlowKind == "quick" || r.FlowUpgraded || r.Estimate == 0
}

// BuildRatio is estimate/build (#234), the ratio drift reads: 0 when build is 0.
func (r LedgerRow) BuildRatio() float64 {
	if r.BuildHours == 0 {
		return 0
	}
	return r.Estimate / r.BuildHours
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
//
// Growth (ARCH-FUNERAL): one row per close, about 155 bytes at 30 columns, so roughly
// 150 KB per thousand closes (#234 added five cells, ~25 bytes; #202 two, ~4). Nothing removes rows: the
// ledger is the calibration history. Its reader is ParseRows over the whole file at each
// close, linear and well under a millisecond at that size; a split by year would be the
// remedy if it ever is not.
const ledgerHeader = "issue\testimate\test_design\test_impl\tactual\tratio\tmodel\tmode\twindow_trusted\tdate" +
	"\tchurn_prod\tchurn_test\tchurn_atlas\tchurn_workshop\trework" +
	"\tgate_rounds\tgate_forced\tgate_addressed\tgate_withdrawn\tgate_open" +
	"\tflow_kind\tflow_provenance\tflow_upgraded" +
	"\tbuild_hours\tconverge_hours\tboundary_rounds\tsplit_source\treview_recipe" +
	"\tgate_disputed\tboundary_disputed\tcorrection"

// legacyCols is the column count before #187's append — the boundary between "row from an
// older binary" and "row carrying cost metrics".
const legacyCols = 10

// costCols is the column count through #187's block (indices 10–19): a row this
// wide carries the cost metrics, whether or not it also carries #231's flow block.
// Each appended block is read by its OWN width — reading by the full header width
// would zero every older row's metrics the day a new block is appended.
const costCols = 20

// flowCols is the width through #231's flow block (indices 20–22); splitCols
// through #234's build/converge block (23–27).
const (
	flowCols    = 23
	splitCols   = 28
	disputeCols = 30 // #202's block (28–29)
)

// Header returns the ledger's TSV header line (written when the file is created).
func Header() string { return ledgerHeader }

// FormatRow renders one tab-separated ledger line in a stable column order.
func FormatRow(r LedgerRow) string {
	line := strings.Join([]string{
		r.Issue,
		ftoa(r.Estimate), ftoa(r.EstDesign), ftoa(r.EstImpl), ftoa(r.Actual), ratioCell(r),
		r.Model, Dash(r.Mode), yesno(r.WindowTrusted), r.Date,
		itoa(r.ChurnProd), itoa(r.ChurnTest), itoa(r.ChurnAtlas), itoa(r.ChurnWorkshop), ftoa(r.Rework),
		itoa(r.GateRounds), itoa(r.GateForced), itoa(r.GateAddressed), itoa(r.GateWithdrawn), itoa(r.GateOpen),
		Dash(r.FlowKind), Dash(r.FlowProvenance), yesno(r.FlowUpgraded),
		ftoa(r.BuildHours), ftoa(r.ConvergeHours), itoa(r.BoundaryRounds), Dash(r.SplitSource), Dash(r.ReviewRecipe),
		itoa(r.GateDisputed), itoa(r.BoundaryDisputed),
	}, "\t")
	if r.Correction != nil {
		payload, _ := json.Marshal(r.Correction)
		line += "\t" + string(payload)
	}
	return line
}

// ratioCell is the ratio column: "-" for a row with no estimate, which is no
// calibration point (#234 E3). ParseRows recomputes the ratio, never reads it.
func ratioCell(r LedgerRow) string {
	if r.Estimate == 0 {
		return "-"
	}
	return ftoa(r.Ratio())
}

// ParseRows parses ledger lines into rows, skipping the header, blanks, and
// `#`-comments. The ratio column is recomputed from estimate/actual, not read.
func ParseRows(text string) []LedgerRow {
	rows, _ := ParseRowsChecked(text)
	return rows
}

// ParseRowsChecked preserves legacy parsing while refusing malformed correction evidence.
func ParseRowsChecked(text string) ([]LedgerRow, error) {
	var rows []LedgerRow
	for lineNumber, line := range strings.Split(text, "\n") {
		rawID := fmt.Sprintf("%x", sha256.Sum256([]byte(line)))
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "issue\t") || strings.HasPrefix(line, "#") {
			continue
		}
		c := strings.Split(line, "\t")
		var correction *CorrectionRecord
		if len(c) > disputeCols && c[disputeCols] != "" && c[disputeCols] != "-" {
			var err error
			if len(c) != disputeCols+1 {
				return nil, fmt.Errorf("ledger line %d: extra correction columns", lineNumber+1)
			}
			correction, err = parseCorrection(c[disputeCols])
			if err != nil {
				return nil, fmt.Errorf("ledger line %d: %w", lineNumber+1, err)
			}
		}
		if len(c) < legacyCols {
			continue
		}
		est, ok := parseLedgerFloat(c[1])
		if !ok {
			if correction != nil {
				return nil, fmt.Errorf("ledger line %d: invalid correction sample number", lineNumber+1)
			}
			continue
		}
		estDesign, ok := parseLedgerFloat(c[2])
		if !ok {
			if correction != nil {
				return nil, fmt.Errorf("ledger line %d: invalid correction sample number", lineNumber+1)
			}
			continue
		}
		estImpl, ok := parseLedgerFloat(c[3])
		if !ok {
			if correction != nil {
				return nil, fmt.Errorf("ledger line %d: invalid correction sample number", lineNumber+1)
			}
			continue
		}
		actual, ok := parseLedgerFloat(c[4])
		if !ok {
			if correction != nil {
				return nil, fmt.Errorf("ledger line %d: invalid correction sample number", lineNumber+1)
			}
			continue
		}
		row := LedgerRow{
			RawID: rawID, Correction: correction,
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
		if len(c) >= costCols {
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
		if len(c) >= flowCols { // #231's flow block
			row.FlowKind = undash(c[20])
			row.FlowProvenance = undash(c[21])
			row.FlowUpgraded = c[22] == "yes"
		}
		if len(c) >= splitCols { // #234's build/converge block
			row.BuildHours, _ = parseLedgerFloat(c[23])
			row.ConvergeHours, _ = parseLedgerFloat(c[24])
			row.BoundaryRounds = atoiOrZero(c[25])
			row.SplitSource = undash(c[26])
			row.ReviewRecipe = undash(c[27])
		}
		if len(c) >= disputeCols { // #202's dispute block
			row.GateDisputed = atoiOrZero(c[28])
			row.BoundaryDisputed = atoiOrZero(c[29])
		}
		rows = append(rows, row)
	}
	return rows, nil
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

// parseLedgerFloat is the parse boundary for every ledger number: a non-finite
// value (a hand-edited Inf or NaN, which ParseFloat accepts) reads as absent,
// so no reader has to re-check it.
func parseLedgerFloat(s string) (float64, bool) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(v, 0) || math.IsNaN(v) {
		return 0, false
	}
	return v, true
}

// Dash renders an empty ledger cell as "-" (the ledger's absent marker).
func Dash(s string) string {
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

// UpgradeHeader rewrites a ledger's header line to the CURRENT column set, returning the
// updated text and whether it changed.
//
// It exists because appending columns (#187) only protects the reader. The writer emits a
// header solely when creating the file, so every pre-existing ledger in the fleet kept its
// 10-column header while newly appended rows carried 20 — the data present but unlabeled,
// which is worse than absent for a file whose whole audience is humans and scripts reading
// columns by name.
//
// Upgrades ONLY when the existing header is a genuine PREFIX of the current one, i.e. a
// strictly older version of this schema. A header with unknown or reordered columns is left
// untouched: that is either a foreign file or a newer binary's, and rewriting it would be
// this function inventing a schema migration it cannot verify.
func UpgradeHeader(text string) (string, bool) {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if !strings.HasPrefix(trimmed, "issue\t") {
			continue
		}
		if trimmed == ledgerHeader {
			return text, false
		}
		have := strings.Split(trimmed, "\t")
		want := strings.Split(ledgerHeader, "\t")
		if len(have) >= len(want) {
			return text, false // same width or wider — not ours to migrate
		}
		for j, col := range have {
			if col != want[j] {
				return text, false // diverges, so it is not an older version of THIS schema
			}
		}
		lines[i] = ledgerHeader
		return strings.Join(lines, "\n"), true
	}
	return text, false
}

// NewestPerIssue returns, for each distinct issue in an ALREADY-FILTERED slice, the LAST row for
// that issue — its newest measurement, the ledger being append-only. Result order follows the
// position of each issue's newest row.
//
// It exists because the ledger is written per CLOSE, not per issue (ariadne#192). Re-closing a
// done issue is legal, and each re-close measures a LONGER cumulative window of the same work —
// so the rows are PARTIAL SUMS, not repeated observations. ariadne#167 accumulated seven rows
// summing 14.57h for work that measured 2.71h. Summing them inflated the blessed throughput
// baseline by 1.41x, and every roadmap forecast derived from it.
//
// Two contract details, both load-bearing:
//
//   - It takes a PRE-FILTERED slice. driftSample filters (trusted, actual>0, same model) BEFORE
//     deduping, and that order matters: dedupe-then-filter would let an untrusted newest row mask
//     an older trusted one and drop the issue from the sample entirely.
//   - A blank Issue is keyed POSITIONALLY, never collapsed. Rows carrying no issue id cannot be
//     known to name the same issue, and drift_test.go's fixtures are exactly that shape — three
//     blank-issue rows are three distinct observations there. Collapsing them would drop the
//     drift sample below its threshold and silently disable the check.
//
// Direction-agnostic by construction (last-wins over the given order), because driftSample walks
// backward taking first-seen while ConvergeFactors and CalibrateForecast keep the last-seen — the
// same rule under opposite traversal.
func NewestPerIssue(rows []LedgerRow) []LedgerRow {
	newest := map[string]int{} // key -> index of its newest row
	for i, r := range rows {
		key := r.Issue
		if key == "" {
			key = "@row:" + strconv.Itoa(i)
		}
		newest[key] = i
	}
	// Sorting the surviving indices both restores input order and makes the result deterministic
	// despite the random map iteration — no separate first-seen order needs tracking.
	idx := make([]int, 0, len(newest))
	for _, i := range newest {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	out := make([]LedgerRow, 0, len(idx))
	for _, i := range idx {
		out = append(out, rows[i])
	}
	return out
}
