package issue

import (
	"path/filepath"
	"strings"
)

// FilenamePattern is the canonical workshop issue filename grammar.
const FilenamePattern = "[0-9][0-9][0-9][0-9][0-9][0-9]-*.md"

// ParseFilename extracts the six-digit ID and slug from an issue filename.
// Paths are accepted for compatibility with existing sdlc callers; only the
// final path component participates in the filename grammar.
func ParseFilename(name string) (id, slug string, ok bool) {
	base := filepath.Base(name)
	matched, _ := filepath.Match(FilenamePattern, base)
	if !matched {
		return "", "", false
	}
	return base[:6], strings.TrimSuffix(base[7:], ".md"), true
}
