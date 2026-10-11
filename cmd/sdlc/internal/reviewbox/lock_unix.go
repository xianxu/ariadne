//go:build unix

package reviewbox

import (
	"os"
	"syscall"
)

// lockExclusive takes an exclusive flock on f; nonBlocking fails at once when
// another open file description holds it.
func lockExclusive(f *os.File, nonBlocking bool) error {
	how := syscall.LOCK_EX
	if nonBlocking {
		how |= syscall.LOCK_NB
	}
	return syscall.Flock(int(f.Fd()), how)
}
