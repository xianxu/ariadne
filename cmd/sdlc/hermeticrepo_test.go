package main

import (
	"strings"
	"testing"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/testfix"
)

// hermeticRepo inits a throwaway git repo, chdirs into it (restored on cleanup),
// and returns its path. A command-tree test that drives a MUTATING verb via
// buildRoot().Execute() acquires the repo transaction lock, whose dir is resolved
// from cwd (repoLockGitCommonDir → `git rev-parse`); without this, cwd is inside
// the REAL ariadne checkout and the test grabs the real .git/sdlc.lock (#149).
// Chdir-ing into a temp git repo makes that resolution land on the temp .git.
func hermeticRepo(t *testing.T) string {
	t.Helper()
	return testfix.Repo(t, testfix.Chdir())
}

// TestRepoLock_IsolatedFromRealRepo is #149's acceptance: inside a hermeticRepo,
// the repo-lock common dir must NOT resolve under the real checkout — so a
// command-tree test never contends on the developer's real .git/sdlc.lock (which
// is what hung `go test` when a live `sdlc` held it).
func TestRepoLock_IsolatedFromRealRepo(t *testing.T) {
	realRoot := realRepoRoot()
	if realRoot == "" {
		t.Skip("not in a git repo — nothing to isolate from")
	}
	hermeticRepo(t) // chdir into a temp git repo
	got, err := repoLockGitCommonDir()
	if err != nil {
		t.Fatalf("repoLockGitCommonDir: %v", err)
	}
	if strings.HasPrefix(got, realRoot) {
		t.Errorf("lock common dir %q is under the REAL repo %q — a command-tree test would grab the real .git/sdlc.lock (#149)", got, realRoot)
	}
}
