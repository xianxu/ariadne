package main

import (
	"strings"
	"testing"
)

// The collision decision is PURE — no git, no filesystem. Its whole job is to
// separate three outcomes that an earlier draft of #207 collapsed into one
// ("if the id is taken, re-allocate"), which would have renumbered every
// existing issue on every sync.
func TestDecideCollision(t *testing.T) {
	const mine = "workshop/issues/000207-sync-without-worktree.md"
	for _, tc := range []struct {
		name      string
		space     map[int][]string
		firstPub  bool
		want      collisionVerdict
		wantPaths []string
	}{
		{
			name:     "id free — publish",
			space:    map[int][]string{206: {"workshop/issues/000206-x.md"}},
			firstPub: true,
			want:     verdictPublish,
		},
		{
			name:     "our own prior publication — publish, NOT a collision",
			space:    map[int][]string{207: {mine}},
			firstPub: false,
			want:     verdictPublish,
		},
		{
			// The regression the first draft would have shipped: every `issue
			// sync` and `claim` republishes an id already on the trunk — its own.
			name:     "republish of an already-published issue never renumbers",
			space:    map[int][]string{207: {mine}},
			firstPub: true, // even asserted as a first publication
			want:     verdictPublish,
		},
		{
			name:      "different slug at our id, first publication — re-allocate",
			space:     map[int][]string{207: {"workshop/issues/000207-someone-else.md"}},
			firstPub:  true,
			want:      verdictReallocate,
			wantPaths: []string{"workshop/issues/000207-someone-else.md"},
		},
		{
			// Renumbering is safe only before anything references the id. By
			// republish time it is in the branch name, commit subjects, deps:,
			// and sidecar filenames (ariadne#188).
			name:      "different slug at our id, republication — REFUSE",
			space:     map[int][]string{207: {"workshop/issues/000207-someone-else.md"}},
			firstPub:  false,
			want:      verdictRefuse,
			wantPaths: []string{"workshop/issues/000207-someone-else.md"},
		},
		{
			name:      "our path AND a foreign one — already collided; refuse on republish",
			space:     map[int][]string{207: {mine, "workshop/issues/000207-other.md"}},
			firstPub:  false,
			want:      verdictRefuse,
			wantPaths: []string{"workshop/issues/000207-other.md"},
		},
		{
			name:      "archived duplicate counts — the ariadne#188 #179 shape",
			space:     map[int][]string{207: {"workshop/history/issues/000207-done-long-ago.md"}},
			firstPub:  true,
			want:      verdictReallocate,
			wantPaths: []string{"workshop/history/issues/000207-done-long-ago.md"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, paths := decideCollision(207, mine, tc.space, tc.firstPub)
			if got != tc.want {
				t.Errorf("verdict = %v, want %v", got, tc.want)
			}
			if len(tc.wantPaths) > 0 {
				if strings.Join(paths, ",") != strings.Join(tc.wantPaths, ",") {
					t.Errorf("conflicting paths = %v, want %v", paths, tc.wantPaths)
				}
			}
		})
	}
}

// nextFreeID walks up from a starting id past everything the trunk holds.
func TestNextFreeID(t *testing.T) {
	space := map[int][]string{207: {"a"}, 208: {"b"}, 210: {"c"}}
	if got := nextFreeID(207, space); got != 209 {
		t.Errorf("nextFreeID = %d, want 209", got)
	}
	if got := nextFreeID(211, map[int][]string{}); got != 211 {
		t.Errorf("nextFreeID on an empty space = %d, want 211", got)
	}
}

// A colliding file with NO FRONTMATTER still collides, because the id space is
// keyed on the FILENAME. Measured 2026-09-09: a grafted `000207-*` log fragment
// with no frontmatter at all blocked `sdlc claim`, and a frontmatter-keyed check
// would have missed it.
func TestDecideCollision_KeysOnFilenameNotFrontmatter(t *testing.T) {
	const mine = "workshop/issues/000207-sync-without-worktree.md"
	// refIDSpace builds this from `ls-tree --name-only`, so contents never enter.
	space := map[int][]string{207: {"workshop/issues/000207-publish-issue-files.md"}}
	if got, _ := decideCollision(207, mine, space, true); got != verdictReallocate {
		t.Errorf("verdict = %v, want re-allocate — the id space is keyed on filename", got)
	}
}
