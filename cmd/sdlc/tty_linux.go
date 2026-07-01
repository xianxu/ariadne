//go:build linux

package main

import "syscall"

// ioctlReadTermios is the linux request that reads terminal attributes; the
// call succeeds only for a real tty (see isTerminal in tty_unix.go).
const ioctlReadTermios = syscall.TCGETS
