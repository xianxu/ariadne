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
}

// Ratio is estimate/actual (0 when actual is 0, to avoid div-by-zero).
func (r LedgerRow) Ratio() float64 {
	if r.Actual == 0 {
		return 0
	}
	return r.Estimate / r.Actual
}

const ledgerHeader = "issue\testimate\test_design\test_impl\tactual\tratio\tmodel\tmode\twindow_trusted\tdate"

// Header returns the ledger's TSV header line (written when the file is created).
func Header() string { return ledgerHeader }

// FormatRow renders one tab-separated ledger line in a stable column order.
func FormatRow(r LedgerRow) string {
	return strings.Join([]string{
		r.Issue,
		ftoa(r.Estimate), ftoa(r.EstDesign), ftoa(r.EstImpl), ftoa(r.Actual), ftoa(r.Ratio()),
		r.Model, dash(r.Mode), yesno(r.WindowTrusted), r.Date,
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
		if len(c) < 10 {
			continue
		}
		rows = append(rows, LedgerRow{
			Issue:         c[0],
			Estimate:      atof(c[1]),
			EstDesign:     atof(c[2]),
			EstImpl:       atof(c[3]),
			Actual:        atof(c[4]),
			Model:         c[6],
			Mode:          undash(c[7]),
			WindowTrusted: c[8] == "yes",
			Date:          c[9],
		})
	}
	return rows
}

func ftoa(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
func atof(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }

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
