// resolve.go — `sdlc resolve` / `sdlc open`. Read-only artifact-reference
// resolver (ariadne#144): map a symbolic ref (ariadne#11, #15 M4, pair#84) to
// the current file path(s) of the issue and its id-keyed plan/review family,
// correct after archiving and across sibling repos.
//
// Two layers (ARCH-PURE):
//   - Pure core: parseRef (the single-source ref grammar parser) + classifyFamily
//     (family classification + ordering). string→struct, no IO, unit-tested.
//   - Thin IO shell: resolveRepoDir (sibling-repo scan) + familyFiles (3-dir glob,
//     directories from vocab.Discovery — nothing hardcoded), surfaced as two cobra
//     commands. Read-only ⟹ never tagged markMutatingCommand ⟹ takes no
//     .git/sdlc.lock, which is what lets parley shell to it on a keypress.
//
// The grammar is single-sourced HERE, as the parser: parley#160 and agents shell
// to `sdlc resolve` rather than re-encoding the grammar. helptext/resolve.md
// documents it for humans; parseRef is the authority.
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ArtifactRef is a parsed symbolic artifact reference. Repo "" means the current
// repo. GitHub marks a GitHub-inbox ref (resolved to a label, not a local path —
// resolve stays read-only and offline).
type ArtifactRef struct {
	Repo      string
	ID        int
	Milestone string
	GitHub    bool
}

var milestoneRe = regexp.MustCompile(`^M\d+[a-z]?$`)

// parseRef parses the single-sourced ref grammar (see helptext/resolve.md). Pure:
// no filesystem, no repo knowledge — a bare `#id` yields Repo "" and the caller
// substitutes the current repo.
//
//	[repo]#id [Mx]        workshop ref   (repo attaches directly to '#')
//	[repo] gh#id          github ref     (a 'gh' token immediately before '#')
//	id is 1–6 digits (zero-padded to 6 at glob time); Mx is a milestone tag.
func parseRef(raw string) (ArtifactRef, error) {
	s := strings.TrimSpace(raw)
	if strings.Count(s, "#") != 1 {
		return ArtifactRef{}, fmt.Errorf("ref %q: expected exactly one '#'", raw)
	}
	i := strings.IndexByte(s, '#')
	left, right := strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])

	// right: id-digits [WS milestone]
	rf := strings.Fields(right)
	if len(rf) == 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: missing id", raw)
	}
	if len(rf[0]) < 1 || len(rf[0]) > 6 {
		return ArtifactRef{}, fmt.Errorf("ref %q: id must be 1–6 digits", raw)
	}
	id, err := strconv.Atoi(rf[0])
	if err != nil || id <= 0 {
		return ArtifactRef{}, fmt.Errorf("ref %q: bad id %q", raw, rf[0])
	}
	ref := ArtifactRef{ID: id}
	if len(rf) >= 2 {
		if !milestoneRe.MatchString(rf[1]) {
			return ArtifactRef{}, fmt.Errorf("ref %q: bad milestone %q", raw, rf[1])
		}
		ref.Milestone = rf[1]
	}
	if len(rf) > 2 {
		return ArtifactRef{}, fmt.Errorf("ref %q: trailing tokens after milestone", raw)
	}

	// left: [repo] ["gh" marker]
	switch {
	case left == "gh":
		ref.GitHub = true
	case strings.HasSuffix(left, " gh"):
		ref.GitHub = true
		ref.Repo = strings.TrimSpace(strings.TrimSuffix(left, " gh"))
	default:
		ref.Repo = left
	}
	if strings.ContainsAny(ref.Repo, " \t") {
		return ArtifactRef{}, fmt.Errorf("ref %q: malformed repo token %q", raw, ref.Repo)
	}
	return ref, nil
}
