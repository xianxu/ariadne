package fleet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xianxu/ariadne/pkg/workspace"
)

type Vantage = workspace.Vantage

func NormalizeVantage(git GitReader, dir string) (Vantage, error) {
	return workspace.NormalizeVantage(git, dir)
}
func canonicalPath(path string) (string, error) { return workspace.CanonicalPath(path) }
func canonicalGitOutputPath(dir string, output []byte) (string, error) {
	return workspace.CanonicalGitOutputPath(dir, output)
}
func canonicalListedWorktreePath(dir, path string) (string, error) {
	return workspace.CanonicalListedWorktreePath(dir, path)
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
