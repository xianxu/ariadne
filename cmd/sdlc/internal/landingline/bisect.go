package landingline

// Bisect finds a failing train's culprit. A train of n cars has prefixes
// 1..n: prefix k is main with the first k cars merged in, and prefix 0 is
// main itself, known good (a red main is refused before any split). results
// maps a tested prefix to whether its suite passed. Bisect returns the next
// prefix to test, or done with the culprit car (1-based) and lo, the longest
// prefix known good, which is what may land.
//
// The search keeps lo (largest passing prefix below hi) and hi (smallest
// failing prefix): every suite halves hi-lo, so a train of n costs at most
// ceil(log2 n) suites after the first. A result that contradicts another (a
// pass above a failure) cannot move the bounds past what the failure proved.
func Bisect(n int, results map[int]bool) (next int, culprit int, lo int, done bool) {
	hi := n + 1
	for k, pass := range results {
		if k >= 1 && k <= n && !pass && k < hi {
			hi = k
		}
	}
	if hi > n {
		return 0, 0, 0, false // no failure to split: nothing to bisect
	}
	lo = 0
	for k, pass := range results {
		if pass && k < hi && k > lo {
			lo = k
		}
	}
	if hi-lo == 1 {
		return 0, hi, lo, true
	}
	return lo + (hi-lo)/2, 0, lo, false
}
