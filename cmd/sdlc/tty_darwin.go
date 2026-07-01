//go:build darwin

package main

import "syscall"

// ioctlReadTermios is the darwin request that reads terminal attributes; the
// call succeeds only for a real tty (see isTerminal in tty_unix.go).
const ioctlReadTermios = syscall.TIOCGETA
