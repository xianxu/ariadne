package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecGit is the production GitReader: git run in dir, combined output,
// bounded by Context when set. GitReader owns path framing: trailing
// whitespace can belong to a path, so output is returned untrimmed.
type ExecGit struct{ Context context.Context }

func (g ExecGit) GitInDir(dir string, args ...string) ([]byte, error) {
	ctx := g.Context
	if ctx == nil {
		ctx = context.Background()
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("git %v in %q: %w", args, dir, err)
	}
	return out, nil
}

// ErrBarePrimary marks a repository whose primary worktree is bare: it has
// no primary checkout.
var ErrBarePrimary = errors.New("bare primary worktree")

// NoAnchorError is an expected absence of a dependency anchor (#230): dir is
// not a linked worktree with a usable primary checkout. Any other error from
// DependencyAnchor is a failure to find out.
type NoAnchorError struct{ Dir, Reason string }

func (e *NoAnchorError) Error() string { return e.Dir + " " + e.Reason }

// DependencyAnchor is a layer's fallback anchor for relative dependencies
// (#230): its repository's primary checkout, when dir is exactly the root of
// a linked worktree. A *NoAnchorError says why there is none; any other error
// means Git could not answer. A numbered environment provisions dependencies
// inside itself (#296), so a vantage whose primary lies outside its
// environment has no anchor.
func DependencyAnchor(git GitReader, dir string) (string, error) {
	none := func(reason string) (string, error) { return "", &NoAnchorError{Dir: dir, Reason: reason} }
	// Without a .git entry, Git would discover an enclosing repository and
	// lend its primary to a layer that is not a checkout at all. A .git
	// directory is a repository's own checkout: linked worktrees always carry
	// a .git file, so only that shape needs Git.
	info, err := os.Lstat(filepath.Join(dir, ".git"))
	switch {
	case errors.Is(err, os.ErrNotExist):
		return none("is not a Git checkout root")
	case err != nil:
		return "", err
	case info.IsDir():
		return none("is its repository's primary checkout")
	}
	v, err := NormalizeVantage(git, dir)
	if errors.Is(err, ErrBarePrimary) {
		return none("belongs to a repository whose primary worktree is bare, so it has no primary checkout")
	} else if err != nil {
		return "", err
	}
	root, err := CanonicalPath(dir)
	if err != nil {
		return "", err
	}
	switch {
	case v.WorktreeRoot != root:
		return none("is not a Git checkout root")
	case v.PrimaryRoot == root:
		return none("is its repository's primary checkout")
	case v.EnvironmentHost != nil && !within(v.EnvironmentRoot, v.PrimaryRoot):
		return none("is in numbered environment " + v.EnvironmentRoot + ", which provisions its own dependencies")
	}
	return v.PrimaryRoot, nil
}

// within reports whether path is root or lies beneath it.
func within(root, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}
