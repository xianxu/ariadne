package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func queueBootID() (string, error) {
	t, err := unix.SysctlTimeval("kern.boottime")
	if err != nil {
		return "", err
	}
	// Compare semantic time fields, never possible struct padding bytes.
	return fmt.Sprintf("%d.%06d", t.Sec, t.Usec), nil
}
