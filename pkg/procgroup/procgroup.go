//go:build unix

// Package procgroup signals a process group and says when nothing in it can
// run any more, the one place weave and sdlc read the kernel's answer (#361).
package procgroup

import (
	"errors"
	"os"
	"syscall"
)

// Kill sends sig to process group pgid. It returns os.ErrProcessDone when no
// member can run again:
//   - ESRCH: the group has no members.
//   - EPERM: on macOS (XNU), killpg skips members that are zombies or already
//     exiting and reports EPERM when it signalled none of them (measured,
//     #361). Our groups hold only same-user processes we started, so EPERM
//     there means the group is dying, not that a live member refused us.
//
// Done is not released: an exiting process still holds its descriptors for a
// moment, so a caller that needs proof of release waits for it on the resource
// itself (weave's producer lease), never on this answer.
func Kill(pgid int, sig syscall.Signal) error {
	err := syscall.Kill(-pgid, sig)
	if errors.Is(err, syscall.ESRCH) || errors.Is(err, syscall.EPERM) {
		return os.ErrProcessDone
	}
	return err
}
