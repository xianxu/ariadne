// inside.go — the one containment predicate: does this path stay inside that root?
package gitx

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Escapes reports whether a relative path leaves the root it is relative to.
//
// COMPONENT-WISE, deliberately (#213 BR-24). `strings.HasPrefix(rel, "..")` also
// matches a legitimate directory named `..config` or `..git-keep`, and catches
// nothing the component test misses — so it is strictly a false-positive
// generator. It had been copied into four call sites (issue-id dirs, review
// directories, and both of `migrate`'s), which is why the correct form already
// living in one of them never reached the other three.
func Escapes(rel string) bool {
	return rel == ".." || rel == "" || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// InsideRoot converts path — absolute, or relative to root — into a
// slash-separated path inside root, refusing anything that escapes.
//
// `root` must already be symlink-resolved: resolving it here per call would
// re-resolve it once per directory, and EvalSymlinks fails on a path that does
// not exist yet, which is the normal state of workshop/history in a fresh repo.
// A RELATIVE path joins onto root rather than the process cwd — "workshop/issues"
// names a place in the repository, not a place relative to wherever the operator
// happens to be standing (#213 BR-25).
func InsideRoot(root, path string) (string, error) {
	abs := path
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(root, abs)
	} else if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("%s is not expressible relative to %s: %w", path, root, err)
	}
	if Escapes(rel) {
		return "", fmt.Errorf("%s is outside %s", path, root)
	}
	return filepath.ToSlash(rel), nil
}
