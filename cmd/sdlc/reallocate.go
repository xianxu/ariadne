// reallocate.go — the one place #207 rewrites an artifact's identity.
//
// ariadne#188 exists because renumbering AFTER the fact is expensive: the id
// leaks into the branch name, commit subjects agents grep, `deps:` in sibling
// issues, and review sidecar filenames. Re-allocation is therefore offered only
// at first publication, when nothing references the id yet — and even then it
// touches exactly two things, together.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// reallocation records an id change so the caller can announce it and clean up.
type reallocation struct {
	OldID, NewID     int
	OldPath, NewPath string
	Foreign          []string // the trunk paths that forced the move
	// written are every local path this attempt-loop created, in order. Retries
	// can produce several; all but NewPath are orphans to remove.
	written []string
}

// idFrontmatterRE matches the `id:` line in an issue's frontmatter.
//
// Anchored to the line start and to the first six-digit run, so a bare "id:"
// appearing later in prose is not rewritten. The frontmatter block is the first
// thing in the file, so the first match is the right one.
var idFrontmatterRE = regexp.MustCompile(`(?m)^id: *(\d{6})\s*$`)

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

	loc := idFrontmatterRE.FindSubmatchIndex(content)
	if loc == nil {
		// A file with no `id:` line — the grafted-fragment shape measured on
		// 2026-09-09. Renaming it alone would be a half-rewrite, so refuse rather
		// than produce an artifact whose name and body disagree.
		return "", nil, fmt.Errorf(
			"re-allocate: %s has no `id:` frontmatter line; renaming alone would leave "+
				"its name and body disagreeing — fix the file by hand", oldPath)
	}
	out := append([]byte{}, content[:loc[2]]...)
	out = append(out, []byte(fmt.Sprintf("%06d", newID))...)
	out = append(out, content[loc[3]:]...)
	return filepath.ToSlash(newPath), out, nil
}

// finish removes the old file and any orphans left by earlier attempts.
//
// Ordering: the new path is written BEFORE the push (so a crash never leaves the
// trunk ahead of the working tree), and the old one is removed only after the
// push succeeds. Retries can write several candidate paths; every one except the
// published NewPath is an orphan.
func (rc *reallocation) finish() error {
	var errs []string
	for _, p := range rc.written {
		if p == rc.NewPath {
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err.Error())
		}
	}
	if rc.OldPath != rc.NewPath {
		if err := os.Remove(rc.OldPath); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("re-allocate cleanup: %s", strings.Join(errs, "; "))
	}
	return nil
}
