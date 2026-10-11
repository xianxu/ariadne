package testfix

import (
	"errors"
	"syscall"

	"golang.org/x/sys/unix"
)

// sZomb is SZOMB from <sys/proc.h>; x/sys/unix does not export it.
const sZomb = 5

func processRunning(pid int) (bool, error) {
	k, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	// An absent pid returns an empty record, which x/sys reports as EIO.
	if errors.Is(err, syscall.EIO) || (err == nil && int(k.Proc.P_pid) != pid) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return k.Proc.P_stat != sZomb, nil
}
