package scratchdisk

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// NormalTempDir selects a normal temporary directory with proven write access,
// keeping os.TempDir when usable. The returned path is absolute so child working
// directories cannot change its meaning. It never changes the process environment.
func NormalTempDir() (string, error) {
	// The Unix default must escape an unusable inherited TMPDIR; os.TempDir would repeat it.
	return normalTempDir(os.TempDir(), UserTempDir, "/tmp", CheckTempDir)
}

// UserTempDir returns Darwin's per-user temporary directory independently of
// TMPDIR, or an empty directory on other platforms.
func UserTempDir() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", nil
	}
	out, err := exec.Command("getconf", "DARWIN_USER_TEMP_DIR").Output()
	return strings.TrimRight(strings.TrimSpace(string(out)), "/"), err
}

func normalTempDir(current string, user func() (string, error), fallback string, probe func(string) error) (string, error) {
	seen := make(map[string]bool)
	var failures []error
	try := func(root string) string {
		if root == "" {
			return ""
		}
		absolute, err := filepath.Abs(root)
		if err != nil {
			failures = append(failures, fmt.Errorf("resolve temp root %s: %w", root, err))
			return ""
		}
		if seen[absolute] {
			return ""
		}
		seen[absolute] = true
		if err := probe(absolute); err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", absolute, err))
			return ""
		}
		return absolute
	}
	if root := try(current); root != "" {
		return root, nil
	}
	userRoot, err := user()
	if err != nil {
		failures = append(failures, fmt.Errorf("per-user temp lookup: %w", err))
	} else if root := try(userRoot); root != "" {
		return root, nil
	}
	if root := try(fallback); root != "" {
		return root, nil
	}
	return "", fmt.Errorf("no writable normal temp directory: %w", errors.Join(failures...))
}

// CheckTempDir tests actual creation and removal rather than mode bits:
// ACLs and sandbox policy can refuse a directory that appears writable.
func CheckTempDir(root string) error {
	dir, err := os.MkdirTemp(root, "sdlc-temp-probe-*")
	if err != nil {
		return err
	}
	return os.Remove(dir)
}
