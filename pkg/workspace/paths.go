package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GitReader interface {
	GitInDir(string, ...string) ([]byte, error)
}

// Vantage is the canonical identity of the Git checkout containing a caller.
// RepoIdentity is Git's shared common directory; PrimaryRoot is the main
// checkout; WorktreeRoot is the checkout containing the caller; and FleetRoot
// is the parent directory from which sibling repositories are discovered.
type Vantage struct {
	RepoIdentity    string
	PrimaryRoot     string
	WorktreeRoot    string
	FleetRoot       string
	EnvironmentRoot string
	EnvironmentHost *EnvironmentHost
}

// NormalizeVantage resolves a caller directory to stable Git and fleet paths.
// Git's relative paths, including --git-common-dir from nested directories,
// are relative to the directory in which Git ran. Every identity is then
// absolute and symlink-resolved before it is compared or returned.
func NormalizeVantage(git GitReader, dir string) (Vantage, error) {
	v, _, err := loadVantage(git, dir)
	return v, err
}

func loadVantage(git GitReader, dir string) (Vantage, []Worktree, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Vantage{}, nil, err
	}
	var observed *Environment
	for p := abs; filepath.Dir(p) != p; p = filepath.Dir(p) {
		candidate, e := environmentCandidate(p)
		if e != nil {
			return Vantage{}, nil, e
		}
		if candidate != nil {
			observed = candidate
			info, e := os.Lstat(p)
			if e != nil {
				return Vantage{}, nil, e
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return Vantage{}, nil, fmt.Errorf("redirected environment checkout %q", p)
			}
		}
	}

	v, trees, err := loadTopology(git, dir)
	if err != nil {
		return v, trees, err
	}
	if observed != nil {
		root, e := CanonicalPath(observed.Root)
		if e != nil {
			return Vantage{}, nil, e
		}
		// A flat legacy worktree may contain subdirectories; preserve that ordinary
		// identity, but an enclosing repository cannot lend Git authority to a slot.
		if v.WorktreeRoot != root && !strings.HasPrefix(v.WorktreeRoot, root+string(filepath.Separator)) {
			return Vantage{}, nil, fmt.Errorf("numbered environment candidate has no independent checkout")
		}
	}
	env, err := environmentForVantage(git, v)
	if err != nil {
		return Vantage{}, nil, err
	}
	if env != nil {
		v.EnvironmentRoot = env.Root
		v.EnvironmentHost = &env.Host
		v.FleetRoot = filepath.Dir(env.Host.PrimaryRoot)
	}
	return v, trees, nil
}

func loadTopology(git GitReader, dir string) (Vantage, []Worktree, error) {
	if git == nil {
		return Vantage{}, nil, fmt.Errorf("normalize fleet vantage: nil Git reader")
	}
	commandDir, err := CanonicalPath(dir)
	if err != nil {
		return Vantage{}, nil, fmt.Errorf("normalize fleet vantage directory %q: %w", dir, err)
	}

	worktreeOutput, err := git.GitInDir(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Vantage{}, nil, gitPathError(dir, "rev-parse --show-toplevel", err, worktreeOutput)
	}
	containingRoot, err := CanonicalGitOutputPath(commandDir, worktreeOutput)
	if err != nil {
		return Vantage{}, nil, fmt.Errorf("normalize containing worktree from %q: %w", dir, err)
	}

	porcelain, err := git.GitInDir(dir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return Vantage{}, nil, gitPathError(dir, "worktree list --porcelain -z", err, porcelain)
	}
	worktrees, err := ParseWorktrees(porcelain)
	if err != nil {
		return Vantage{}, nil, fmt.Errorf("parse git worktree list from %q: %w", dir, err)
	}
	if len(worktrees) == 0 {
		return Vantage{}, nil, fmt.Errorf("git worktree list from %q contains no worktrees", dir)
	}

	primaryRoot, err := CanonicalListedWorktreePath(commandDir, worktrees[0].Path)
	if err != nil {
		return Vantage{}, nil, fmt.Errorf("canonicalize primary worktree %q: %w", worktrees[0].Path, err)
	}
	if worktrees[0].Bare {
		return Vantage{}, nil, fmt.Errorf("git worktree list from %q identifies a bare primary worktree", dir)
	}

	worktreeRoot := ""
	for _, worktree := range worktrees {
		candidate, err := CanonicalListedWorktreePath(commandDir, worktree.Path)
		if err != nil {
			continue
		}
		if candidate == containingRoot {
			worktreeRoot = candidate
			break
		}
	}
	if worktreeRoot == "" {
		return Vantage{}, nil, fmt.Errorf("containing worktree %q is absent from git worktree list", containingRoot)
	}

	commonOutput, err := git.GitInDir(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return Vantage{}, nil, gitPathError(dir, "rev-parse --git-common-dir", err, commonOutput)
	}
	commonDir, err := CanonicalGitOutputPath(commandDir, commonOutput)
	if err != nil {
		return Vantage{}, nil, fmt.Errorf("normalize Git common directory from %q: %w", dir, err)
	}

	return Vantage{
		RepoIdentity:    commonDir,
		PrimaryRoot:     primaryRoot,
		WorktreeRoot:    worktreeRoot,
		FleetRoot:       filepath.Dir(primaryRoot),
		EnvironmentRoot: filepath.Dir(primaryRoot),
	}, worktrees, nil
}

func gitPathError(dir, command string, err error, output []byte) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("git -C %q %s: %w", dir, command, err)
	}
	return fmt.Errorf("git -C %q %s: %w: %s", dir, command, err, message)
}

// CanonicalGitOutputPath resolves a Git LF-terminated command response to a
// stable absolute identity. Only Git's final LF is removed: every preceding
// byte, including a path-owned CR, remains part of the path.
func CanonicalGitOutputPath(commandDir string, output []byte) (string, error) {
	path := string(output)
	if strings.HasSuffix(path, "\n") {
		path = strings.TrimSuffix(path, "\n")
	}
	return canonicalGitPath(commandDir, path)
}

// CanonicalListedWorktreePath resolves a path parsed from NUL-delimited
// worktree porcelain. Its bytes are already separated from Git's record
// delimiters, so whitespace is meaningful and must remain untouched.
func CanonicalListedWorktreePath(commandDir, path string) (string, error) {
	return canonicalGitPath(commandDir, path)
}

// canonicalGitPath resolves a Git path to a stable absolute identity. Git's
// relative paths are anchored at its command directory, which can differ from
// a worktree root for a nested caller.
func canonicalGitPath(commandDir, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("Git returned an empty path")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(commandDir, path)
	}
	return CanonicalPath(path)
}

func CanonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return filepath.Clean(resolved), nil
}

// FeatureWorktreePath keeps independent clone worktrees within their owner
// environment. Branch validation by Git is separate; this enforces containment.
func FeatureWorktreePath(id Identity, branch string) (string, error) {
	if branch == "" || branch == "." || filepath.IsAbs(branch) || filepath.Clean(branch) != branch || branch == ".." || strings.HasPrefix(branch, ".."+string(filepath.Separator)) || strings.Contains(branch, "\\") {
		return "", fmt.Errorf("invalid feature branch path %q", branch)
	}
	if !validRepoName(id.Repo) {
		return "", fmt.Errorf("invalid repository name %q", id.Repo)
	}
	if id.EnvironmentHost != nil && id.RepoIdentity != id.EnvironmentHost.RepoIdentity {
		return filepath.Join(id.EnvironmentRoot, ".worktrees", id.Repo, branch), nil
	}
	return filepath.Join(id.FleetRoot, "worktree", id.Repo, branch), nil
}
