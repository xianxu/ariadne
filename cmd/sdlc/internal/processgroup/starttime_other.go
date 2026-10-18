//go:build !darwin && !linux

package processgroup

import "errors"

// StartTime is unsupported here, so callers that verify a pid refuse to act.
func StartTime(int) (string, error) {
	return "", errors.New("process start time is not supported on this platform")
}
