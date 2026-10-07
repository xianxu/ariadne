package issue

import (
	"path"
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

// CLIRef renders a canonical (zero-padded) issue ID as the decimal value a
// `--issue N` flag expects. A padded "000253" would be parsed as octal (171) by
// the flag library, so every next-action hint goes through this one function.
func CLIRef(id string) string {
	if n := strings.TrimLeft(id, "0"); n != "" {
		return n
	}
	return id
}

// JoinRefs renders several issue IDs through CLIRef, each with prefix, joined
// by sep: ("#", ", ") for humans, ("", ",") for a rerunnable --issue list,
// ("#", ",") for a tracker commit subject (#284).
func JoinRefs(ids []string, prefix, sep string) string {
	refs := make([]string, len(ids))
	for i, id := range ids {
		refs[i] = prefix + CLIRef(id)
	}
	return strings.Join(refs, sep)
}

// BranchName is an issue's branch: the stem of its details or card filename
// (#284: one derivation for every caller).
func BranchName(detailsOrCardPath string) string {
	return strings.TrimSuffix(path.Base(detailsOrCardPath), ".md")
}
