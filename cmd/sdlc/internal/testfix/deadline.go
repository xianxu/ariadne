package testfix

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// cueDeadline is a context that expires, as a deadline, only when the test
// says so (#434).
type cueDeadline struct {
	context.Context
	done chan struct{}
	once sync.Once
}

func (c *cueDeadline) Done() <-chan struct{} { return c.done }

func (c *cueDeadline) Err() error {
	select {
	case <-c.done:
		return context.DeadlineExceeded
	default:
		return nil
	}
}

// DeadlineOnCue returns a context, carrying parent's values, whose deadline
// passes when expire is called, not on a clock. Hand it to code that derives
// its own WithTimeout and call expire once the fixture has signalled
// readiness: the code sees a real deadline (DeadlineExceeded), and a slow
// machine can no longer make it fire before the fixture is ready.
func DeadlineOnCue(parent context.Context) (ctx context.Context, expire func()) {
	c := &cueDeadline{Context: parent, done: make(chan struct{})}
	return c, func() { c.once.Do(func() { close(c.done) }) }
}

// AwaitPIDFile waits for a fixture to write its pid to path: the readiness
// signal a deadline or cancellation test cues on. It gives up when stop
// reports true (the process under test ended first; nil never stops) or after
// ceiling, returning 0 either way.
func AwaitPIDFile(path string, ceiling time.Duration, stop func() bool) int {
	until := time.Now().Add(ceiling)
	for time.Now().Before(until) {
		raw, _ := os.ReadFile(path)
		if pid, _ := strconv.Atoi(strings.TrimSpace(string(raw))); pid > 0 {
			return pid
		}
		if stop != nil && stop() {
			return 0
		}
		time.Sleep(5 * time.Millisecond)
	}
	return 0
}
