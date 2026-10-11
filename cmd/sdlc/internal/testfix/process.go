package testfix

import (
	"fmt"
	"time"
)

// AwaitExit waits until pid no longer executes: exited, whether reaped or a
// zombie awaiting its reaper. Cancellation tests use it to prove a descendant
// did not survive shutdown. Delivering SIGKILL to a group is asynchronous, so
// a probe at the instant the runner returns can still see the victim running;
// the ceiling is only a bound for a process that truly survived.
//
// It asks the kernel directly instead of running ps: agent sandboxes (macOS
// Seatbelt) deny exec of /bin/ps, which failed those tests in every sandboxed
// slot or, where the error was ignored, silently passed them (#330).
func AwaitExit(pid int, within time.Duration) error {
	deadline := time.Now().Add(within)
	for {
		running, err := processRunning(pid)
		if err != nil {
			return fmt.Errorf("inspect pid %d: %w", pid, err)
		}
		if !running {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("pid %d still running after %v", pid, within)
		}
		time.Sleep(5 * time.Millisecond)
	}
}
