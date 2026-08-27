package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/gitx"
)

// Vantage is the canonical identity of the Git checkout containing a caller.
// RepoIdentity is Git's shared common directory; PrimaryRoot is the main
// checkout; WorktreeRoot is the checkout containing the caller; and FleetRoot
// is the parent directory from which sibling repositories are discovered.
type Vantage struct {
	RepoIdentity string
	PrimaryRoot  string
	WorktreeRoot string
	FleetRoot    string
}

// NormalizeVantage resolves a caller directory to stable Git and fleet paths.
// Git's relative paths, including --git-common-dir from nested directories,
// are relative to the directory in which Git ran. Every identity is then
// absolute and symlink-resolved before it is compared or returned.
func NormalizeVantage(git GitReader, dir string) (Vantage, error) {
	if git == nil {
		return Vantage{}, fmt.Errorf("normalize fleet vantage: nil Git reader")
	}
	commandDir, err := canonicalPath(dir)
	if err != nil {
		return Vantage{}, fmt.Errorf("normalize fleet vantage directory %q: %w", dir, err)
	}

	worktreeOutput, err := git.GitInDir(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Vantage{}, gitPathError(dir, "rev-parse --show-toplevel", err, worktreeOutput)
	}
	containingRoot, err := canonicalGitOutputPath(commandDir, worktreeOutput)
	if err != nil {
		return Vantage{}, fmt.Errorf("normalize containing worktree from %q: %w", dir, err)
	}

	porcelain, err := git.GitInDir(dir, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return Vantage{}, gitPathError(dir, "worktree list --porcelain -z", err, porcelain)
	}
	worktrees, err := gitx.ParseWorktrees(porcelain)
	if err != nil {
		return Vantage{}, fmt.Errorf("parse git worktree list from %q: %w", dir, err)
	}
	if len(worktrees) == 0 {
		return Vantage{}, fmt.Errorf("git worktree list from %q contains no worktrees", dir)
	}

	primaryRoot, err := canonicalListedWorktreePath(commandDir, worktrees[0].Path)
	if err != nil {
		return Vantage{}, fmt.Errorf("canonicalize primary worktree %q: %w", worktrees[0].Path, err)
	}
	if worktrees[0].Bare {
		return Vantage{}, fmt.Errorf("git worktree list from %q identifies a bare primary worktree", dir)
	}

	worktreeRoot := ""
	for _, worktree := range worktrees {
		candidate, err := canonicalListedWorktreePath(commandDir, worktree.Path)
		if err != nil {
			continue
		}
		if candidate == containingRoot {
			worktreeRoot = candidate
			break
		}
	}
	if worktreeRoot == "" {
		return Vantage{}, fmt.Errorf("containing worktree %q is absent from git worktree list", containingRoot)
	}

	commonOutput, err := git.GitInDir(dir, "rev-parse", "--git-common-dir")
	if err != nil {
		return Vantage{}, gitPathError(dir, "rev-parse --git-common-dir", err, commonOutput)
	}
	commonDir, err := canonicalGitOutputPath(commandDir, commonOutput)
	if err != nil {
		return Vantage{}, fmt.Errorf("normalize Git common directory from %q: %w", dir, err)
	}

	return Vantage{
		RepoIdentity: commonDir,
		PrimaryRoot:  primaryRoot,
		WorktreeRoot: worktreeRoot,
		FleetRoot:    filepath.Dir(primaryRoot),
	}, nil
}

func gitPathError(dir, command string, err error, output []byte) error {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("git -C %q %s: %w", dir, command, err)
	}
	return fmt.Errorf("git -C %q %s: %w: %s", dir, command, err, message)
}

// canonicalGitOutputPath resolves a Git LF-terminated command response to a
// stable absolute identity. Only Git's final LF is removed: every preceding
// byte, including a path-owned CR, remains part of the path.
func canonicalGitOutputPath(commandDir string, output []byte) (string, error) {
	path := string(output)
	if strings.HasSuffix(path, "\n") {
		path = strings.TrimSuffix(path, "\n")
	}
	return canonicalGitPath(commandDir, path)
}

// canonicalListedWorktreePath resolves a path parsed from NUL-delimited
// worktree porcelain. Its bytes are already separated from Git's record
// delimiters, so whitespace is meaningful and must remain untouched.
func canonicalListedWorktreePath(commandDir, path string) (string, error) {
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
	return canonicalPath(path)
}

func canonicalPath(path string) (string, error) {
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

// CanonicalProspectivePath resolves every existing path component through
// symlinks while preserving a not-yet-created suffix. The second result is the
// deepest existing directory and is suitable as Git's command directory when
// Requested itself does not exist yet.
func CanonicalProspectivePath(path string) (requested, containingDir string, err error) {
	abs, err := absolutePathSpelling(path)
	if err != nil {
		return "", "", fmt.Errorf("canonicalize prospective path %q: %w", path, err)
	}
	abs = filepath.FromSlash(abs)
	volume := filepath.VolumeName(abs)
	remainder := strings.TrimPrefix(abs, volume)
	separator := string(filepath.Separator)
	if !strings.HasPrefix(remainder, separator) {
		return "", "", fmt.Errorf("canonicalize prospective path %q: absolute path is not rooted", path)
	}
	current := volume + separator
	currentIsDir := true
	components := strings.Split(strings.TrimLeft(remainder, separator), separator)
	missing := make([]string, 0)
	for _, component := range components {
		switch component {
		case "", ".":
			if !currentIsDir {
				return "", "", fmt.Errorf("canonicalize prospective path %q: directory syntax follows a non-directory component", path)
			}
			continue
		case "..":
			if len(missing) != 0 {
				return "", "", fmt.Errorf("canonicalize prospective path %q: parent traversal follows a missing component", path)
			}
			if !currentIsDir {
				return "", "", fmt.Errorf("canonicalize prospective path %q: parent traversal follows a non-directory component", path)
			}
			current = filepath.Dir(current)
			continue
		}
		if !currentIsDir {
			return "", "", fmt.Errorf("canonicalize prospective path %q: component follows a non-directory path", path)
		}
		if len(missing) != 0 {
			missing = append(missing, component)
			continue
		}

		candidate := filepath.Join(current, component)
		info, statErr := os.Lstat(candidate)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				missing = append(missing, component)
				continue
			}
			return "", "", fmt.Errorf("inspect prospective path component %q: %w", candidate, statErr)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			resolved, resolveErr := filepath.EvalSymlinks(candidate)
			if resolveErr != nil {
				return "", "", fmt.Errorf("resolve prospective path symlink %q: %w", candidate, resolveErr)
			}
			current = filepath.Clean(resolved)
			resolvedInfo, resolveErr := os.Stat(current)
			if resolveErr != nil {
				return "", "", fmt.Errorf("inspect resolved prospective path %q: %w", current, resolveErr)
			}
			currentIsDir = resolvedInfo.IsDir()
			continue
		}
		current = candidate
		currentIsDir = info.IsDir()
	}

	if len(missing) != 0 {
		info, statErr := os.Stat(current)
		if statErr != nil {
			return "", "", fmt.Errorf("inspect prospective path ancestor %q: %w", current, statErr)
		}
		if !info.IsDir() {
			return "", "", fmt.Errorf("prospective path ancestor %q is not a directory", current)
		}
		requested = current
		for _, component := range missing {
			requested = filepath.Join(requested, component)
		}
		return filepath.Clean(requested), filepath.Clean(current), nil
	}
	info, statErr := os.Stat(current)
	if statErr != nil {
		return "", "", fmt.Errorf("inspect prospective path %q: %w", current, statErr)
	}
	current = filepath.Clean(current)
	if info.IsDir() {
		return current, current, nil
	}
	return current, filepath.Dir(current), nil
}

func absolutePathSpelling(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if path == "" {
		return workingDir, nil
	}
	return workingDir + string(filepath.Separator) + path, nil
}
