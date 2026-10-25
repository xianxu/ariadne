package testfix

import (
	"bytes"
	"strings"
	"sync"

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

// AwaitCondition is testgate.AwaitCondition, the one poll-until-event helper
// with required workers (#622), shared with cmd/weave and cmd/tl-lead.
var AwaitCondition = testgate.AwaitCondition
