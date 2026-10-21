package main

import "strings"

// A runner owns its own capacity; inherited suite-held authority cannot be
// delegated across that boundary, even when launched by a merge process.
func withoutSuiteAuthority(env []string) []string {
	out := make([]string, 0, len(env))
	for _, s := range env {
		if !strings.HasPrefix(s, "WF_TEST_LOCK_HELD_BY=") {
			out = append(out, s)
		}
	}
	return out
}

// runnerCleanupUnconfirmed keeps recovery separate from a completed failure.
type runnerCleanupUnconfirmed struct{ Detail string }

func (e *runnerCleanupUnconfirmed) Error() string { return "runner cleanup unconfirmed: " + e.Detail }
