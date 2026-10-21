package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// A runner owns its own capacity; inherited suite-held authority cannot be
// delegated across that boundary, even when launched by a merge process.
func withoutSuiteAuthority(env []string) []string {
	out := make([]string, 0, len(env))
	for _, s := range env {
		if !strings.HasPrefix(s, "WF_TEST_LOCK_HELD_BY=") {
			out = append(out, s)
		}
	}
	return out
}

// runnerCleanupUnconfirmed keeps recovery separate from a completed failure.
type runnerCleanupUnconfirmed struct{ Detail string }

func (e *runnerCleanupUnconfirmed) Error() string { return "runner cleanup unconfirmed: " + e.Detail }

// One existing Git worktree lock protects every runner artifact. These bounded
// references live only inside existing private scratch/inflight directories.
const runnerArtifactOwnerFile = ".sdlc-runner-owner"

func readRunnerOwner(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 4097))
	if err == nil && len(raw) > 4096 {
		err = fmt.Errorf("runner owner exceeds4096 bytes")
	}
	return raw, err
}

func parseRunnerOwner(raw []byte, base string, relative bool) (string, error) {
	if len(raw) > 4096 {
		return "", fmt.Errorf("oversized runner owner")
	}
	text := strings.TrimSuffix(string(raw), "\n")
	if !strings.HasPrefix(text, "gitdir: ") {
		return "", fmt.Errorf("invalid runner owner")
	}
	dir := strings.TrimPrefix(text, "gitdir: ")
	if dir == "" || strings.ContainsAny(dir, "\x00\r\n") {
		return "", fmt.Errorf("invalid runner gitdir")
	}
	if !filepath.IsAbs(dir) {
		if !relative {
			return "", fmt.Errorf("runner owner must be absolute")
		}
		dir = filepath.Join(base, dir)
	}
	return dir, nil
}

func attachRunnerArtifactOwner(dir, wt string) error {
	raw, err := readRunnerOwner(filepath.Join(wt, ".git"))
	if err != nil {
		return err
	}
	gitdir, err := parseRunnerOwner(raw, wt, true)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, runnerArtifactOwnerFile), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	_, err = f.WriteString("gitdir: " + gitdir + "\n")
	if err == nil {
		err = f.Sync()
	}
	return errors.Join(err, f.Close())
}

func runnerArtifactProtected(root string) bool {
	return runnerArtifactProtectedWithRead(root, readRunnerOwner)
}

func runnerArtifactProtectedWithRead(root string, read func(string) ([]byte, error)) bool {
	raw, err := read(filepath.Join(root, runnerArtifactOwnerFile))
	relative := false
	if errors.Is(err, os.ErrNotExist) {
		raw, err = read(filepath.Join(root, ".git"))
		relative = true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		return true
	}
	gitdir, err := parseRunnerOwner(raw, root, relative)
	if err != nil {
		return true
	}
	info, err := os.Stat(gitdir)
	if err != nil || !info.IsDir() {
		return true
	} // unavailable authority never proves release
	_, err = os.Lstat(filepath.Join(gitdir, "locked"))
	return !errors.Is(err, os.ErrNotExist)
}

// Immediate cleanup shares one proof bit. Before launch there is no child;
// from launch intent through confirmed group completion, retirement is denied.
// In particular a panic cannot turn unwinding into proof of completion.
type runnerArtifactLifetime struct{ retirable bool }

func (l *runnerArtifactLifetime) mayRetire() bool { return l == nil || l.retirable }
