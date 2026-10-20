//go:build !darwin

package scratchdisk

import "syscall"

// TestScratch is a macOS facility; other platforms retain override fixtures.
func fileSystemName(_ *syscall.Statfs_t) string { return "" }
