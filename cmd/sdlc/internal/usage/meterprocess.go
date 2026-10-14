package usage

import (
	"errors"
	"syscall"
)

func meterProcessAlive(pid int) bool {
	e := syscall.Kill(pid, 0)
	return e == nil || errors.Is(e, syscall.EPERM)
}
