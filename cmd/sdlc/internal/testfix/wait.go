package testfix

import (
	"bytes"
	"strings"
	"sync"
	"time"
)

// HangCeiling bounds a wait for an event that a correct program produces
// promptly. It catches a hang, not a slow machine: with several slots' suites
// running at once (load ~50), spawning a binary or running git is many times
// slower, and a deadline of a few seconds turns contention into failures
// (#325). Assert ordering or concurrency with event evidence (see Watch); this
// ceiling only keeps a broken run from stalling until go test's own timeout.
const HangCeiling = 2 * time.Minute

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
