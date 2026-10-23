//go:build unix

// Package flock wraps flock(2): an exclusive lock on an open file that the
// kernel releases when its holder exits, however it exits.
package flock

import (
	"os"
	"syscall"
)

// Exclusive takes an exclusive flock on f; nonBlocking fails at once when
// another open file description holds it.
func Exclusive(f *os.File, nonBlocking bool) error {
	how := syscall.LOCK_EX
	if nonBlocking {
		how |= syscall.LOCK_NB
	}
	return syscall.Flock(int(f.Fd()), how)
}

// Shared takes a shared flock on f: readers hold it together, and it fails
// (nonBlocking) only while an exclusive holder has the file.
func Shared(f *os.File, nonBlocking bool) error {
	how := syscall.LOCK_SH
	if nonBlocking {
		how |= syscall.LOCK_NB
	}
	return syscall.Flock(int(f.Fd()), how)
}
