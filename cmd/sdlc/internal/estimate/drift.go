package estimate

import "fmt"

// DriftVerdict inspects the last n window-trusted rows: if they all miss in the
// same direction by >2× (estimate/actual > 2, or < 0.5), it warns. Untrusted rows
// are excluded — a truncated actual isn't a real data point (the #68 posture
// applied to calibration). Returns (false, "") when there's no systematic drift or
// fewer than n trusted rows. Pure.
func DriftVerdict(rows []LedgerRow, n int) (warn bool, msg string) {
	if n <= 0 {
		return false, ""
	}
	var trusted []LedgerRow
	for _, r := range rows {
		if r.WindowTrusted && r.Actual > 0 {
			trusted = append(trusted, r)
		}
	}
	if len(trusted) < n {
		return false, ""
	}
	last := trusted[len(trusted)-n:]
	over, under := 0, 0
	for _, r := range last {
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
