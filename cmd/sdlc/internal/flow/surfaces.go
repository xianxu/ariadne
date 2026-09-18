package flow

import (
	"fmt"
	"path"
	"strings"
)

// DeclarationPath is where a repo declares its shared surfaces (#231): one
// repo-owned file, like .sdlc/fleet.json, not woven from the base layer.
const DeclarationPath = ".sdlc/shared-surfaces"

// Surfaces is a repo's declared shared surfaces: path patterns whose change
// takes a diff out of the quick flow whatever its size — a keybinding registry,
// an option schema, a cross-module seam.
type Surfaces struct{ patterns []string }

// ParseSurfaces reads a declaration: one pattern per line, blank lines and
// `#` comments skipped. A pattern is a path.Match glob (`*` does not cross a
// `/`), or a directory with a trailing `/`, meaning anything under it. A
// malformed pattern is an error naming its line; the caller treats that as a
// crossing, so a broken declaration fails toward the full flow.
func ParseSurfaces(text string) (Surfaces, error) {
	var s Surfaces
	for i, line := range strings.Split(text, "\n") {
		p := strings.TrimSpace(line)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		p = strings.TrimPrefix(strings.TrimPrefix(p, "./"), "/")
		// The directory form is a literal prefix, so a glob inside it would be
		// accepted and then never match anything — a guard that guards nothing.
		// Refuse it here, where the error becomes a crossing (fail toward full).
		if strings.HasSuffix(p, "/") && strings.ContainsAny(p, `*?[\`) {
			return Surfaces{}, fmt.Errorf("%s line %d: %q: a directory entry (trailing /) is a literal prefix and cannot hold a glob", DeclarationPath, i+1, p)
		}
		if _, err := path.Match(strings.TrimSuffix(p, "/"), ""); err != nil {
			return Surfaces{}, fmt.Errorf("%s line %d: %q: %v", DeclarationPath, i+1, p, err)
		}
		s.patterns = append(s.patterns, p)
	}
	return s, nil
}

// Match reports whether a repo-relative path is a shared surface. The
// declaration file always is: a branch that edits its own shell away is
// touching a shared surface by doing so.
func (s Surfaces) Match(p string) bool {
	if p == DeclarationPath {
		return true
	}
	for _, pat := range s.patterns {
		if strings.HasSuffix(pat, "/") {
			if strings.HasPrefix(p, pat) {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pat, p); ok {
			return true
		}
	}
	return false
}

// Union is every pattern of both. Close reads the declaration as committed at
// the window base and at HEAD and takes the union, so a branch can add a
// surface but cannot remove one from its own shell, and an uncommitted edit is
// never read (#231 PQ-5).
func Union(a, b Surfaces) Surfaces {
	return Surfaces{patterns: append(append([]string{}, a.patterns...), b.patterns...)}
}
