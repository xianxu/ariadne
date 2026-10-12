package testfix

import (
	"errors"
	"fmt"
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
// for every git the test binary runs and every repository it creates, and
// returns the cleanup for TestMain to run after m.Run. A fetch or push may
// leave one running after the command returns, still writing into a temp
// repository when t.TempDir's cleanup removes it ("directory not empty";
// #284, #286, #302).
//
// It takes two channels because neither reaches every git. The GIT_CONFIG_*
// environment covers git run in any repository, but git strips it from the far
// end of a local transport (`unset GIT_CONFIG_COUNT …; git-receive-pack`), so a
// push into a test's bare origin would still start `git maintenance run --auto
// --detach` there (#302). GIT_TEMPLATE_DIR's config is copied into every
// repository `git init` and `git clone` create, and repository config survives
// the strip. A test that sets either variable itself overrides this for its own
// duration.
func QuietBackgroundGit() (cleanup func()) {
	for k, v := range map[string]string{
		"GIT_CONFIG_COUNT":   "2",
		"GIT_CONFIG_KEY_0":   "gc.auto",
		"GIT_CONFIG_VALUE_0": "0",
		"GIT_CONFIG_KEY_1":   "maintenance.auto",
		"GIT_CONFIG_VALUE_1": "false",
	} {
		_ = os.Setenv(k, v)
	}
	// hooks/ and info/ stand in for the default template's: tests write hooks
	// and info/exclude into the repositories they create.
	dir, err := os.MkdirTemp("", "testfix-git-template-")
	if err == nil {
		err = errors.Join(os.Mkdir(filepath.Join(dir, "hooks"), 0o755), os.Mkdir(filepath.Join(dir, "info"), 0o755),
			os.WriteFile(filepath.Join(dir, "config"), []byte(quietTemplateConfig), 0o644))
	}
	if err != nil {
		panic(fmt.Sprintf("testfix: quiet git template: %v", err))
	}
	_ = os.Setenv("GIT_TEMPLATE_DIR", dir)
	return func() { _ = os.RemoveAll(dir) }
}

const quietTemplateConfig = "[gc]\n\tauto = 0\n[maintenance]\n\tauto = false\n[receive]\n\tautogc = false\n"
