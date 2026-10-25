package testfix

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xianxu/ariadne/pkg/testgate"
)

// HangCeiling is testgate.HangCeiling: the one hang bound for tests, which
// catches a hang, not a slow machine (#325).
const HangCeiling = testgate.HangCeiling

// Watch is an io.Writer that keeps what is written to it and closes Seen once
// the output contains needle. Point a command's stderr at it to fail the
// moment a diagnostic appears (such as the repo lock's wait line) instead of
// inferring it from elapsed time.
type Watch struct {
	needle string
	mu     sync.Mutex
	buf    bytes.Buffer
	seen   chan struct{}
	closed bool
}

// NewWatch returns a Watch for needle.
func NewWatch(needle string) *Watch {
	return &Watch{needle: needle, seen: make(chan struct{})}
}

func (w *Watch) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.buf.Write(p)
	if !w.closed && strings.Contains(w.buf.String(), w.needle) {
		w.closed = true
		close(w.seen)
	}
	return n, err
}

// Seen is closed once the output contains the needle.
func (w *Watch) Seen() <-chan struct{} { return w.seen }

// String returns everything written so far.
func (w *Watch) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.String()
}

// AwaitCondition polls cond until it holds, bounded only by HangCeiling
// (#622). workers are the goroutines or processes whose progress produces the
// condition; one that ends first fails t at once instead of at the ceiling.
// The argument is required: pass nil, with a comment saying why, only when no
// worker can end first. A worker's result is put back on its channel before
// t fails, so a later reader (a deferred join, a completion wait) still gets
// it; each channel therefore needs room for one value.
func AwaitCondition(t testing.TB, what string, cond func() bool, workers []chan error) {
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
