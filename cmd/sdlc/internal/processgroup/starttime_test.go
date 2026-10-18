//go:build darwin || linux

package processgroup

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func TestStartTime(t *testing.T) {
	self, err := StartTime(os.Getpid())
	if err != nil || self == "" {
		t.Fatalf("own start time: %q %v", self, err)
	}
	if again, _ := StartTime(os.Getpid()); again != self {
		t.Errorf("start time is not stable: %q then %q", self, again)
	}
	child := exec.Command("true")
	if err := child.Run(); err != nil {
		t.Fatal(err)
	}
	// Reaped, so the pid names no process (or, rarely, a newer one).
	if got, err := StartTime(child.Process.Pid); err == nil && got == self {
		t.Errorf("a different process shares the start time %q", got)
	} else if err != nil && !errors.Is(err, ErrNoProcess) {
		t.Errorf("absent pid: %v", err)
	}
}
