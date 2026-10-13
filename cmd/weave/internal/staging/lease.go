package staging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var ErrInUse = errors.New("stage still has producer leases")

// Lease is acquired BEFORE starting a payload process and inherited by its
// descendants. Close, never LOCK_UN: duplicated descriptors share the lock and
// it must survive until the last inherited copy closes. Marker code is trusted
// to preserve this descriptor and remain in its assigned process group.
func Lease(stage string) (*os.File, error) { return lock(stage, syscall.LOCK_SH) }

// Exclusive proves no inherited writer remains. The caller holds this proof
// through cleanup and closes it afterward. PID reuse cannot fake this proof.
func Exclusive(stage string) (*os.File, error) { return lock(stage, syscall.LOCK_EX) }

// Settle bounds how long a dead or signalled producer may take to release its
// inherited lease. An exiting process closes its descriptors in order, so a
// parent waiting on its output pipe (fd 1/2) can resume before the lease (fd 3)
// is released (#343); a live producer keeps it past the bound.
const Settle = time.Second

// AwaitExclusive is Exclusive retried until within elapses, for a lease whose
// holders are expected to be exiting. ErrInUse after the bound means a holder
// is still alive.
func AwaitExclusive(stage string, within time.Duration) (*os.File, error) {
	deadline := time.Now().Add(within)
	for {
		lease, err := Exclusive(stage)
		if !errors.Is(err, ErrInUse) || time.Now().After(deadline) {
			return lease, err
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func lock(stage string, kind int) (*os.File, error) {
	path := filepath.Join(stage, "lease")
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("invalid producer lease %s", path)
	}
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	if err = syscall.Flock(int(f.Fd()), kind|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, ErrInUse
		}
		return nil, err
	}
	return f, nil
}
