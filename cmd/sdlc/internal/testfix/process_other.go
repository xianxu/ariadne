//go:build !darwin && !linux

package testfix

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

func processRunning(pid int) (bool, error) {
	out, err := exec.Command("ps", "-o", "stat=", "-p", strconv.Itoa(pid)).Output()
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		return false, nil // no such process
	}
	if err != nil {
		return false, err
	}
	stat := strings.TrimSpace(string(out))
	return stat != "" && !strings.HasPrefix(stat, "Z"), nil
}
