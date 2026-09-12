// reallocate.go — the one place #207 rewrites an artifact's identity.
//
// ariadne#188 exists because renumbering AFTER the fact is expensive: the id
// leaks into the branch name, commit subjects agents grep, `deps:` in sibling
// issues, and review sidecar filenames. Re-allocation is therefore offered only
// at first publication, when nothing references the id yet — and even then it
// touches exactly two things, together.
package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
)

// reallocation records one id change, so the caller can announce it and clean
// up afterwards. DATA ONLY: the local-file bookkeeping lives on publishResult,
// which is the thing that spans retry attempts (#207 BR-2, BR-3).
type reallocation struct {
	OldID, NewID     int
	OldPath, NewPath string
	Foreign          []string // the trunk paths that forced the move
}

// idFrontmatterRE matches the `id:` line INSIDE the frontmatter block.
//
// Anchored to a line start, and applied only to the leading `---` block: a
// document that merely mentions `id: 000207` in prose has no identity to move,
// and rewriting that line would edit the body while leaving the real identity
// untouched (#207 BR-14).
var idFrontmatterRE = regexp.MustCompile(`(?m)^id: *(\d{6})\s*$`)

// frontmatterSpan returns the byte range of the leading `---` block.
func frontmatterSpan(content []byte) (int, int, bool) {
	const open = "---\n"
	if !bytes.HasPrefix(content, []byte(open)) {
		return 0, 0, false
	}
	rest := content[len(open):]
	i := bytes.Index(rest, []byte("\n---"))
	if i < 0 {
		return 0, 0, false
	}
	return len(open), len(open) + i + 1, true
}

// rewriteIdentity produces the file's content and path under a new id.
//
// Filename AND frontmatter move together, and that is the whole point: leaving
// either behind produces an artifact whose name and body disagree, and the two
// are read by different consumers — `sdlc claim` matches the filename,
// `vocabulary validate-instance` reads the frontmatter. A half-rewrite is worse
// than the collision it was fixing, because it is silent.
func rewriteIdentity(oldPath string, content []byte, newID int) (string, []byte, error) {
	base := filepath.Base(oldPath)
	if len(base) < 7 || base[6] != '-' {
		return "", nil, fmt.Errorf("re-allocate: %q does not follow the NNNNNN-slug.md convention", base)
	}
	newPath := filepath.Join(filepath.Dir(oldPath), fmt.Sprintf("%06d", newID)+base[6:])

	lo, hi, ok := frontmatterSpan(content)
	var loc []int
	if ok {
		if m := idFrontmatterRE.FindSubmatchIndex(content[lo:hi]); m != nil {
			loc = []int{m[0] + lo, m[1] + lo, m[2] + lo, m[3] + lo}
		}
	}
	if loc == nil {
		// A file with no `id:` line — the grafted-fragment shape measured on
		// 2026-09-09. Renaming it alone would be a half-rewrite, so refuse rather
		// than produce an artifact whose name and body disagree.
		return "", nil, fmt.Errorf(
			"re-allocate: %s has no `id:` line in a leading frontmatter block; renaming "+
				"alone would leave its name and body disagreeing — fix the file by hand", oldPath)
	}
	out := append([]byte{}, content[:loc[2]]...)
	out = append(out, []byte(fmt.Sprintf("%06d", newID))...)
	out = append(out, content[loc[3]:]...)
	return filepath.ToSlash(newPath), out, nil
}
