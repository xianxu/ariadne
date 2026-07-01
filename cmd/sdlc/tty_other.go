//go:build !darwin && !linux

package main

// isTerminal has no stdlib-only ioctl probe on non-unix platforms, so it
// conservatively reports false: callers (merge's confirm gate, change-code's
// --worktree=ask) then treat the context as non-interactive and require the
// explicit --yes / answer flag rather than risk a blocked prompt (#141).
func isTerminal(fd uintptr) bool {
	return false
}
