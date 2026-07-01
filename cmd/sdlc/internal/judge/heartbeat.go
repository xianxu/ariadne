package judge

import (
	"fmt"
	"time"
)

// heartbeatInterval is how often Dispatch emits a progress line to opts.Stderr
// while the agent subprocess is still running (#140). 30s is frequent enough to
// refute the "looks hung" reading the issue reported (multiple silent 60s
// polling windows) without being chatty. Overridable in tests.
var heartbeatInterval = 30 * time.Second

// newHeartbeatTicker returns a channel that fires every d plus a stop func. It
// is the injected clock seam: production uses a real time.Ticker; tests replace
// it to drive ticks by hand, so the heartbeat is tested without sleeping.
var newHeartbeatTicker = func(d time.Duration) (ticks <-chan time.Time, stop func()) {
	t := time.NewTicker(d)
	return t.C, t.Stop
}

// sinceStart reports elapsed time since a start instant. Injected alongside
// newHeartbeatTicker so a test can assert exact elapsed wording deterministically.
var sinceStart = time.Since

// heartbeatLine renders one progress line for a still-running agent dispatch
// (#140). It is deliberately harness-agnostic: elapsed, agent, and pid all come
// from sdlc wrapping the child process, not from anything the child emits, so it
// reads identically whether the reviewer is claude, codex, or gemini. A pid <= 0
// means the child has not been observed starting yet (the brief window before
// Run's onStart fires). Pure — the IO is the caller's.
func heartbeatLine(elapsed time.Duration, agent string, pid int) string {
	if agent == "" {
		agent = "agent"
	}
	e := elapsed.Round(time.Second)
	if pid <= 0 {
		return fmt.Sprintf("    … still working — %s elapsed via %s (pid pending)", e, agent)
	}
	return fmt.Sprintf("    … still working — %s elapsed via %s (pid %d; inspect: ps -p %d)", e, agent, pid, pid)
}
