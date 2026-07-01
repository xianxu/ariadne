//go:build darwin || linux

package main

import (
	"syscall"
	"unsafe"
)

// isTerminal reports whether fd refers to a terminal, using the same ioctl
// probe golang.org/x/term uses internally (TIOCGETA on darwin, TCGETS on
// linux): fetching the terminal attributes succeeds only for an actual tty.
// This is the correct test — distinct from "is a character device", which is
// also true for /dev/null and would misclassify an agent's redirected stdin as
// interactive (#141). Stdlib-only, so the sdlc binary keeps its cobra-only
// dependency set (the per-OS request constant lives in tty_{darwin,linux}.go).
func isTerminal(fd uintptr) bool {
	var t syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios,
		uintptr(unsafe.Pointer(&t)), 0, 0, 0)
	return errno == 0
}
