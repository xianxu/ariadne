package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecGit is the production GitReader: git run in dir, combined output.
type ExecGit struct{}

func (ExecGit) GitInDir(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v in %q: %w", args, dir, err)
	}
	return out, nil
}

// DependencyAnchor is a layer's fallback anchor for relative dependencies
// (#230): its repository's primary checkout, when dir is exactly the root of
// a linked worktree. The error names why there is none. A numbered
// environment provisions dependencies inside itself (#296), so a vantage
// whose primary lies outside its environment has no anchor.
func DependencyAnchor(git GitReader, dir string) (string, error) {
	// Without a .git entry, Git would discover an enclosing repository and
	// lend its primary to a layer that is not a checkout at all.
	if _, err := os.Lstat(filepath.Join(dir, ".git")); err != nil {
		return "", fmt.Errorf("%s is not a Git checkout root", dir)
	}
	v, err := NormalizeVantage(git, dir)
	if err != nil {
		return "", err
	}
	root, err := CanonicalPath(dir)
	if err != nil {
		return "", err
	}
	switch {
	case v.WorktreeRoot != root:
		return "", fmt.Errorf("%s is not a Git checkout root", dir)
	case v.PrimaryRoot == root:
		return "", fmt.Errorf("%s is its repository's primary checkout", dir)
	case v.EnvironmentHost != nil && !within(v.EnvironmentRoot, v.PrimaryRoot):
		return "", fmt.Errorf("%s is in numbered environment %s, which provisions its own dependencies", dir, v.EnvironmentRoot)
	}
	return v.PrimaryRoot, nil
}

// within reports whether path is root or lies beneath it.
func within(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
