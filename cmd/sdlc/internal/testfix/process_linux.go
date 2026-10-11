package testfix

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func processRunning(pid int) (bool, error) {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// The state follows the parenthesised command name, which may contain spaces.
	rest := string(raw[strings.LastIndexByte(string(raw), ')')+1:])
	state := strings.TrimSpace(rest)
	return state != "" && state[0] != 'Z' && state[0] != 'X', nil
}
