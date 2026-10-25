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
	for i, w := range workers {
		if cap(w) < 1 {
			t.Fatalf("%s: worker %d's channel is unbuffered; AwaitCondition needs room to put its result back", what, i)
		}
	}
	deadline := time.Now().Add(HangCeiling)
	for !cond() {
		for i, w := range workers {
			select {
			case err, open := <-w:
				if !open {
					t.Fatalf("%s: worker %d's channel was closed; send its result instead", what, i)
				}
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

// Returns runs call and fails t unless it returns within HangCeiling (#622).
// It proves a bounded operation (a cancelled or timed-out call, a killed
// child) ended by its own bound, not by a timing check: make the work the
// bound must stop outlast HangCeiling, so only a broken bound reaches it.
func Returns(t T, what string, call func()) {
	t.Helper()
	returned := make(chan struct{})
	go func() {
		defer close(returned)
		call()
	}()
	// nil: the call itself is what is awaited; nothing else can end first.
	AwaitCondition(t, what+" returned", func() bool {
		select {
		case <-returned:
			return true
		default:
			return false
		}
	}, nil)
}
