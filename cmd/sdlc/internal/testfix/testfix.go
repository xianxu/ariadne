// Package testfix provides the shared git-repo fixtures for cmd/sdlc tests
// (ariadne#186). It consolidates two idioms that were copy-pasted across the
// suite: the "run git in a dir, fatal on error" runner (previously duplicated as
// git/gitIn/captureGit/inline run closures) and the "init a throwaway repo on
// main, config a test identity, gpgsign off, optionally commit" setup sequence
// (previously copy-pasted across hermeticRepo/initFleetRepo/windowRepo/closeRepo/
// publishRepo and the internal/gitx + internal/activetime sub-packages).
//
// It imports only the standard library and testing, so it introduces no import
// cycle with any cmd/sdlc package that consumes it.
package testfix

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Git runs `git <args>` in dir (dir == "" → the current working directory),
// failing the test on error, and returns the combined stdout+stderr. It is the
// one runner the fixtures and tests share.
func Git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v (dir %q): %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// Capture runs `git <args>` in dir (dir == "" → cwd) and returns stdout only,
// failing the test on error. Use it when the caller parses the output (e.g.
// `rev-parse HEAD`) and must not fold stderr into the result.
func Capture(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v (dir %q): %v", args, dir, err)
	}
	return string(out)
}

type cfg struct {
	parent, name  string
	chdir         bool
	initialCommit bool
}

// Option tunes Repo.
type Option func(*cfg)

// Chdir chdir's into the new repo and restores the previous working directory on
// test cleanup. Needed by verbs that resolve the repo from cwd (the repo
// transaction lock, gitx.CommitWindow).
func Chdir() Option { return func(c *cfg) { c.chdir = true } }

// InitialCommit writes a README and makes an initial commit, so the first
// subsequent commit has a parent (a review window's branch-start = firstSHA^).
func InitialCommit() Option { return func(c *cfg) { c.initialCommit = true } }

// At places the repo at parent/name instead of a fresh t.TempDir() — for fleet
// fixtures that build several sibling repos under one parent.
func At(parent, name string) Option {
	return func(c *cfg) { c.parent, c.name = parent, name }
}

// Repo initializes a throwaway git repo on branch main with a deterministic test
// identity and commit signing disabled, and returns its path. Options tune
// placement (At), whether to chdir in (Chdir), and whether to seed an initial
// commit (InitialCommit).
func Repo(t *testing.T, opts ...Option) string {
	t.Helper()
	var c cfg
	for _, o := range opts {
		o(&c)
	}

	var dir string
	if c.parent != "" {
		dir = filepath.Join(c.parent, c.name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	} else {
		dir = t.TempDir()
	}

	Git(t, dir, "init", "-q", "-b", "main")
	Git(t, dir, "config", "user.email", "t@t")
	Git(t, dir, "config", "user.name", "t")
	Git(t, dir, "config", "commit.gpgsign", "false")

	if c.initialCommit {
		if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		Git(t, dir, "add", "README")
		Git(t, dir, "commit", "-q", "-m", "init")
	}

	if c.chdir {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chdir(dir); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chdir(cwd) })
	}

	return dir
}
