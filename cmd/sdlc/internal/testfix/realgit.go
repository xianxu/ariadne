package testfix

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PreferRealGit puts git's exec-path first on PATH, so the suite's many git
// spawns start the git binary directly. On macOS /usr/bin/git is the xcrun
// shim, which spends about 11 ms per spawn locating the real binary (#253:
// cmd/sdlc went from 1,531 s to 867 s). git already prepends this directory to
// PATH for its own child processes, so this changes the launcher, not the git.
//
// Call it from TestMain before m.Run. It is a no-op when git is absent or its
// exec-path holds no git binary. Child processes (a built sdlc binary) inherit
// the PATH, so they get the same git.
func PreferRealGit() {
	out, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		return
	}
	dir := strings.TrimSpace(string(out))
	if fi, err := os.Stat(filepath.Join(dir, "git")); err != nil || fi.IsDir() {
		return
	}
	_ = os.Setenv("PATH", PrependPath(dir, os.Getenv("PATH")))
}

// PrependPath returns path with dir as its first entry, leaving path alone
// when dir already leads it.
func PrependPath(dir, path string) string {
	if path == "" {
		return dir
	}
	sep := string(os.PathListSeparator)
	if first, _, _ := strings.Cut(path, sep); first == dir {
		return path
	}
	return dir + sep + path
}
