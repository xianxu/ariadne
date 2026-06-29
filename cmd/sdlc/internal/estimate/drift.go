package estimate

import "fmt"

// DriftVerdict inspects the latest n unique rows for the latest model revision:
// if they all miss in the same direction by >2× (estimate/actual > 2, or < 0.5),
// it warns. Untrusted rows are excluded — a truncated actual isn't a real data
// point (the #68 posture applied to calibration). Returns (false, "") when
// there's no systematic drift or fewer than n trusted rows for the latest model.
// Pure.
func DriftVerdict(rows []LedgerRow, n int) (warn bool, msg string) {
	if n <= 0 {
		return false, ""
	}
	trusted := driftSample(rows, n)
	if len(trusted) < n {
		return false, ""
	}
	over, under := 0, 0
	for _, r := range trusted {
		switch ratio := r.Ratio(); {
		case ratio > 2.0:
			over++
		case ratio < 0.5:
			under++
		}
	}
	switch {
	case over == n:
		return true, fmt.Sprintf("calibration drift: last %d trusted estimates all >2× OVER actual — estimates running high; add a Model-Revision note", n)
	case under == n:
		return true, fmt.Sprintf("calibration drift: last %d trusted estimates all >2× UNDER actual — estimates running low; add a Model-Revision note", n)
	}
	return false, ""
}

func driftSample(rows []LedgerRow, n int) []LedgerRow {
	latest := latestTrustedActual(rows)
	if latest < 0 || !KnownModel(rows[latest].Model) {
		return nil
	}
	model := rows[latest].Model
	seen := map[string]bool{}
	var out []LedgerRow
	for i := latest; i >= 0 && len(out) < n; i-- {
		r := rows[i]
		if !r.WindowTrusted || r.Actual <= 0 || r.Model != model {
			continue
		}
		key := r.Issue
		if key == "" {
			key = fmt.Sprintf("@row:%d", i)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

func latestTrustedActual(rows []LedgerRow) int {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].WindowTrusted && rows[i].Actual > 0 {
			return i
		}
	}
	return -1
}
