package sysvshm

import (
	"bytes"
	"fmt"
	"golang.org/x/sys/unix"
)

// Membership, command and incarnation come from the same kernel record.
func kernelWitness(pid, pgid int) (Witness, bool) {
	k, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil || int(k.Proc.P_pid) != pid || int(k.Eproc.Pgid) != pgid {
		return Witness{}, false
	}
	t := k.Proc.P_starttime
	return Witness{Name: string(bytes.TrimRight(k.Proc.P_comm[:], "\x00")), Start: fmt.Sprintf("%d.%06d", t.Sec, t.Usec)}, true
}
