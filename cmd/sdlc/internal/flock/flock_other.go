//go:build !unix

// Package flock wraps flock(2): an exclusive lock on an open file that the
// kernel releases when its holder exits, however it exits.
package flock

import (
	"errors"
	"os"
)

func Exclusive(*os.File, bool) error {
	return errors.New("flock is unix only")
}
