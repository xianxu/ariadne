package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// writeMarker creates dir/.dynamic-skill with the given mode (so a test can lay
// down an executable vs a non-executable marker). The package dir is created if
// absent.
func writeMarker(t *testing.T, dir string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".dynamic-skill"),
		[]byte("#!/bin/sh\n"), mode); err != nil {
		t.Fatal(err)
	}
}

// TestDynamicSkillDirs_LeafOnlyExecutableExcludeAdapted is the core selection
// contract (Task 1.2): given the LEAF layer's `skill` intents and a directory
// listing where —
//   - leaf construct/local/datatype/.dynamic-skill is executable    → returned
//   - leaf construct/adapted/foo/.dynamic-skill is executable        → EXCLUDED (adapted)
//   - leaf construct/local/plain has no marker                        → ignored
//   - leaf construct/local/nonexec/.dynamic-skill is NOT executable   → ignored
//   - an ANCESTOR layer's construct/local/anc/.dynamic-skill is exec  → NOT returned (leaf-only)
//
// — only the leaf's construct/local/datatype dir is selected.
func TestDynamicSkillDirs_LeafOnlyExecutableExcludeAdapted(t *testing.T) {
	parent := t.TempDir()
	ancestor := filepath.Join(parent, "ancestor")
	leaf := filepath.Join(parent, "leaf")

	// Ancestor declares a `skill` dir with an executable marker — it must NOT be
	// scanned (leaf-only scoping is what keeps ancestor trees byte-pristine).
	writeMarker(t, filepath.Join(ancestor, "construct", "local", "anc"), 0o755)

	// Leaf layout.
	writeMarker(t, filepath.Join(leaf, "construct", "local", "datatype"), 0o755) // executable → returned
	writeMarker(t, filepath.Join(leaf, "construct", "local", "nonexec"), 0o644)  // non-exec → ignored
	if err := os.MkdirAll(filepath.Join(leaf, "construct", "local", "plain"), 0o755); err != nil {
		t.Fatal(err) // no marker at all → ignored
	}
	writeMarker(t, filepath.Join(leaf, "construct", "adapted", "foo"), 0o755) // executable but ADAPTED → excluded

	layers := []layer.Layer{
		{Name: "ancestor", Path: ancestor, Intents: skillRows("construct/local")},
		{Name: "leaf", Path: leaf, Intents: skillRows("construct/local", "construct/adapted")},
	}

	got, err := DynamicSkillDirs(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{filepath.Join(leaf, "construct", "local", "datatype")}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if got[0] != want[0] {
		t.Errorf("selected dir = %q, want %q", got[0], want[0])
	}
}

// TestIsExecutable locks the pure bit test: any of the three executable bits
// (owner/group/other) counts as executable; a plain 0644 does not.
func TestIsExecutable(t *testing.T) {
	for _, tc := range []struct {
		mode os.FileMode
		want bool
	}{
		{0o644, false},
		{0o600, false},
		{0o755, true},
		{0o744, true}, // owner-x only
		{0o711, true},
		{0o111, true},
	} {
		if got := isExecutable(tc.mode); got != tc.want {
			t.Errorf("isExecutable(%#o) = %v, want %v", tc.mode, got, tc.want)
		}
	}
}

// TestDynamicSkillDirs_NoLayers / single-layer edge: a leaf with no skill intents
// (or an absent skill dir) selects nothing, never errors.
func TestDynamicSkillDirs_AbsentSkillDir(t *testing.T) {
	root := t.TempDir()
	layers := []layer.Layer{
		{Name: "repo", Path: root, Intents: []intent.Intent{
			{Kind: intent.Skill, Source: "construct/local"}, // dir absent
			{Kind: intent.Prose, Source: "AGENTS.local.md"}, // non-skill intent ignored
		}},
	}
	got, err := DynamicSkillDirs(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want no dirs (absent skill dir)", got)
	}
}
