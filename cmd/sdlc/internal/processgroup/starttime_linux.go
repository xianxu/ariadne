package processgroup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// StartTime is pid's start time (clock ticks since boot, /proc/<pid>/stat
// field 22), or ErrNoProcess. Two processes that reuse one pid never share it.
func StartTime(pid int) (string, error) {
	raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if errors.Is(err, fs.ErrNotExist) {
		return "", ErrNoProcess
	}
	if err != nil {
		return "", err
	}
	// Fields resume after the parenthesised command name, which may hold spaces;
	// the first of them is field 3 (state), so field 22 is index 19.
	fields := strings.Fields(string(raw[strings.LastIndexByte(string(raw), ')')+1:]))
	if len(fields) < 20 {
		return "", fmt.Errorf("/proc/%d/stat: %d fields after the command", pid, len(fields))
	}
	return fields[19], nil
}
