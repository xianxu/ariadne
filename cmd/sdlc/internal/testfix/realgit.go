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

// QuietBackgroundGit turns off git's detached auto-gc and auto-maintenance
// for every git the test binary runs. A fetch or push may leave one running
// after the command returns, still writing into a temp repository when
// t.TempDir's cleanup removes it ("directory not empty"; #284, #286). A test
// that sets GIT_CONFIG_COUNT itself overrides this for its own duration.
func QuietBackgroundGit() {
	for k, v := range map[string]string{
		"GIT_CONFIG_COUNT":   "2",
		"GIT_CONFIG_KEY_0":   "gc.auto",
		"GIT_CONFIG_VALUE_0": "0",
		"GIT_CONFIG_KEY_1":   "maintenance.auto",
		"GIT_CONFIG_VALUE_1": "false",
	} {
		_ = os.Setenv(k, v)
	}
}
