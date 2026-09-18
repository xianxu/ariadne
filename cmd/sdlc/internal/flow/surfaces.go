package flow

import (
	"fmt"
	"path"
	"strings"
)

// DeclarationPath is where a repo declares its shared surfaces (#231): one
// repo-owned file, like .sdlc/fleet.json, not woven from the base layer.
const DeclarationPath = ".sdlc/shared-surfaces"

// SurfaceForms is the declaration grammar in the words every surface prints —
// rendered into `sdlc close --help` ({{SURFACE_FORMS}}), pointed at by the
// declaration's own header and the atlas, and pinned example by example against
// ParseSurfaces/Match by TestSurfaceFormsDescribesTheParser (#231 BR-31). The
// one statement of the grammar; nothing else restates it.
const SurfaceForms = `one pattern per line; blank lines and # comments are skipped.
    pkg/vocab       a literal path: that path, and everything under it
    pkg/vocab/      a directory entry: everything under it
    cmd/*.go        a path.Match glob: exactly the paths it matches — * does
                    not cross /, and a glob never covers a subtree
  Refused (so a declaration cannot guard nothing): **, a leading !, a glob in a
  directory entry, and any non-canonical path (leading / or ./, ../, //).`

// Surfaces is a repo's declared shared surfaces: path patterns whose change
// takes a diff out of the quick flow whatever its size — a keybinding registry,
// an option schema, a cross-module seam.
type Surfaces struct{ patterns []string }

// ParseSurfaces reads a declaration: one pattern per line, blank lines and `#`
// comments skipped. Every form it ACCEPTS takes effect in Match — a form that
// parsed but never matched would be a guard guarding nothing (#231 BR-21,
// BR-26) — so the forms are few and each has one meaning:
//
//   - a literal path, `pkg/vocab` or `AGENTS.base.md`: that path, and anything
//     under it if it is a directory;
//   - a directory with a trailing `/`, `pkg/vocab/`: anything under it;
//   - a path.Match glob, `construct/vocabulary/*.cue` (`*` does not cross `/`).
//
// Refused, as an error naming the line: `**` (path.Match has no recursive
// glob), a leading `!` (there is no negation), a glob inside a directory entry,
// and a malformed glob. The caller treats the error as a crossing, so a broken
// declaration fails toward the full flow.
func ParseSurfaces(text string) (Surfaces, error) {
	var s Surfaces
	for i, line := range strings.Split(text, "\n") {
		p := strings.TrimSpace(line)
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		refuse := func(why string) (Surfaces, error) {
			return Surfaces{}, fmt.Errorf("%s line %d: %q: %s", DeclarationPath, i+1, p, why)
		}
		// Only CANONICAL patterns: repo-relative, no `.`/`..`/`//` segments, not
		// empty — the forms a path from git can ever equal, so an accepted pattern
		// can always match something (#231 BR-33).
		if body := strings.TrimSuffix(p, "/"); body == "" || body == "." || path.Clean(body) != body || strings.HasPrefix(body, "/") || strings.HasPrefix(body, "../") {
			return refuse("not a canonical repo-relative path (no leading /, no ./, ../, // or trailing /. segments)")
		}
		switch {
		case strings.HasPrefix(p, "!"):
			return refuse("there is no negation — declare only what IS a shared surface")
		case strings.Contains(p, "**"):
			return refuse("`**` is not supported — use a directory entry (`dir/`) for a subtree")
		case strings.HasSuffix(p, "/") && isGlob(p):
			return refuse("a directory entry (trailing /) is a literal prefix and cannot hold a glob")
		}
		if _, err := path.Match(strings.TrimSuffix(p, "/"), ""); err != nil {
			return refuse(err.Error())
		}
		s.patterns = append(s.patterns, p)
	}
	return s, nil
}

func isGlob(p string) bool { return strings.ContainsAny(p, `*?[\`) }

// Match reports whether a repo-relative path is a shared surface. The
// declaration file always is: a branch that edits its own shell away is
// touching a shared surface by doing so.
func (s Surfaces) Match(p string) bool {
	if p == DeclarationPath {
		return true
	}
	for _, pat := range s.patterns {
		switch {
		case strings.HasSuffix(pat, "/"):
			if strings.HasPrefix(p, pat) {
				return true
			}
		case !isGlob(pat):
			if p == pat || strings.HasPrefix(p, pat+"/") {
				return true
			}
		default:
			if ok, _ := path.Match(pat, p); ok {
				return true
			}
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
