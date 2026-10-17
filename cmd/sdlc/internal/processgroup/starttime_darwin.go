package processgroup

import (
	"errors"
	"fmt"
	"syscall"

	"golang.org/x/sys/unix"
)

// StartTime is pid's start time from the kernel's process table, or
// ErrNoProcess. Two processes that reuse one pid never share it.
func StartTime(pid int) (string, error) {
	k, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	// An absent pid returns an empty record, which x/sys reports as EIO.
	if errors.Is(err, syscall.EIO) || (err == nil && int(k.Proc.P_pid) != pid) {
		return "", ErrNoProcess
	}
	if err != nil {
		return "", err
	}
	t := k.Proc.P_starttime
	return fmt.Sprintf("%d.%06d", t.Sec, t.Usec), nil
}
