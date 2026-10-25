package testgate

import "time"

// T is the part of testing.TB that AwaitCondition uses, so this package needs
// no testing import.
type T interface {
	Helper()
	Fatalf(format string, args ...any)
}

// AwaitCondition polls cond until it holds, bounded only by HangCeiling
// (#622). workers are the goroutines or processes whose progress produces the
// condition; one that ends first fails t at once instead of at the ceiling.
// The argument is required: pass nil, with a comment saying why, only when no
// worker can end first. A worker's result is put back on its channel before
// t fails, so a later reader (a deferred join, a completion wait) still gets
// it; each channel therefore needs room for one value.
func AwaitCondition(t T, what string, cond func() bool, workers []chan error) {
	t.Helper()
	deadline := time.Now().Add(HangCeiling)
	for !cond() {
		for i, w := range workers {
			select {
			case err := <-w:
				w <- err
				t.Fatalf("%s: worker %d ended first: %v", what, i, err)
			default:
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s: not reached within %s", what, HangCeiling)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
