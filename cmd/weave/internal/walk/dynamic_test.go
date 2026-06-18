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

// dsByDir indexes the selected DynamicSkills by their bare Dir for assertions.
func dsByDir(ds []DynamicSkill) map[string]DynamicSkill {
	out := map[string]DynamicSkill{}
	for _, d := range ds {
		out[d.Dir] = d
	}
	return out
}

// TestDynamicSkills_AncestorSelectedExecutableExcludeAdapted is the INVERTED
// selection contract (#115 M3): unlike the retired leaf-only DynamicSkillDirs, an
// ANCESTOR-owned executable marker IS now selected (the generate stage runs it
// with cwd = the compiling repo's root, so its output still lands in the
// derivative). Given —
//   - ancestor construct/local/datatype/.dynamic-skill executable → SELECTED (all-layers)
//   - leaf construct/adapted/foo/.dynamic-skill executable        → EXCLUDED (adapted)
//   - leaf construct/local/plain has no marker                     → ignored
//   - leaf construct/local/nonexec/.dynamic-skill NOT executable   → ignored
//   - leaf construct/local/leafdyn/.dynamic-skill executable       → SELECTED (leaf)
//
// — datatype (ancestor) AND leafdyn (leaf) are selected; foo/plain/nonexec are not.
func TestDynamicSkills_AncestorSelectedExecutableExcludeAdapted(t *testing.T) {
	parent := t.TempDir()
	ancestor := filepath.Join(parent, "ancestor")
	leaf := filepath.Join(parent, "leaf")

	// Ancestor declares a `skill` dir with an executable marker — under all-layers
	// selection it IS scanned (leaf-rooted OUTPUT, not leaf-only selection, keeps
	// the ancestor tree byte-pristine).
	writeConfig(t, ancestor, "xx-")
	writeMarker(t, filepath.Join(ancestor, "construct", "local", "datatype"), 0o755)

	// Leaf layout.
	writeConfig(t, leaf, "xx-")
	writeMarker(t, filepath.Join(leaf, "construct", "local", "leafdyn"), 0o755) // executable → selected
	writeMarker(t, filepath.Join(leaf, "construct", "local", "nonexec"), 0o644) // non-exec → ignored
	if err := os.MkdirAll(filepath.Join(leaf, "construct", "local", "plain"), 0o755); err != nil {
		t.Fatal(err) // no marker at all → ignored
	}
	writeMarker(t, filepath.Join(leaf, "construct", "adapted", "foo"), 0o755) // executable but ADAPTED → excluded

	layers := []layer.Layer{
		{Name: "ancestor", Path: ancestor, Intents: skillRows("construct/local")},
		{Name: "leaf", Path: leaf, Intents: skillRows("construct/local", "construct/adapted")},
	}

	got, err := DynamicSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}

	by := dsByDir(got)
	if len(got) != 2 {
		t.Fatalf("got %d dynamic skills %v, want 2 (datatype ancestor + leafdyn leaf)", len(got), got)
	}
	dt, ok := by["datatype"]
	if !ok {
		t.Fatalf("ancestor-owned datatype not selected; got %v", got)
	}
	// Name is prefixed; OutputRel uses the bare Dir under construct/generated.
	if dt.Name != "xx-datatype" {
		t.Errorf("datatype Name = %q, want xx-datatype", dt.Name)
	}
	if dt.OutputRel != filepath.Join("construct", "generated", "datatype") {
		t.Errorf("datatype OutputRel = %q, want construct/generated/datatype", dt.OutputRel)
	}
	// MarkerPath points at the ANCESTOR's marker (where it physically lives).
	wantMarker := filepath.Join(ancestor, "construct", "local", "datatype", ".dynamic-skill")
	if dt.MarkerPath != wantMarker {
		t.Errorf("datatype MarkerPath = %q, want ancestor's marker %q", dt.MarkerPath, wantMarker)
	}
	if _, ok := by["leafdyn"]; !ok {
		t.Errorf("leaf-owned leafdyn not selected; got %v", got)
	}
	if _, ok := by["foo"]; ok {
		t.Error("adapted foo must be excluded")
	}
	// Sorted by Dir: datatype before leafdyn.
	if got[0].Dir != "datatype" || got[1].Dir != "leafdyn" {
		t.Errorf("not sorted by Dir: %v", got)
	}
}

// TestDynamicSkills_AncestorInternalExcluded: a dynamic skill declared by an
// ANCESTOR with INTERNAL visibility is NOT visible to the compiling leaf, so it is
// excluded — the same intent.Selected rule the planner applies to every artifact.
func TestDynamicSkills_AncestorInternalExcluded(t *testing.T) {
	parent := t.TempDir()
	ancestor := filepath.Join(parent, "ancestor")
	leaf := filepath.Join(parent, "leaf")
	writeConfig(t, ancestor, "xx-")
	writeConfig(t, leaf, "xx-")
	// Ancestor declares an INTERNAL skill dir carrying an executable marker.
	writeMarker(t, filepath.Join(ancestor, "construct", "priv", "secretdyn"), 0o755)

	layers := []layer.Layer{
		{Name: "ancestor", Path: ancestor, Intents: []intent.Intent{
			{Kind: intent.Skill, Visibility: intent.Internal, Source: "construct/priv"},
		}},
		{Name: "leaf", Path: leaf, Intents: skillRows("construct/local")},
	}

	got, err := DynamicSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want none (ancestor's INTERNAL dynamic skill is not visible to the leaf)", got)
	}
}

// TestDynamicSkills_LeafInternalSelected: the leaf's OWN internal dynamic skill IS
// visible (intent.Selected is true for any visibility on the leaf), the mirror of
// the ancestor-internal exclusion above.
func TestDynamicSkills_LeafInternalSelected(t *testing.T) {
	root := t.TempDir()
	writeConfig(t, root, "xx-")
	writeMarker(t, filepath.Join(root, "construct", "priv", "leafsecret"), 0o755)
	layers := []layer.Layer{
		{Name: "repo", Path: root, Intents: []intent.Intent{
			{Kind: intent.Skill, Visibility: intent.Internal, Source: "construct/priv"},
		}},
	}
	got, err := DynamicSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Dir != "leafsecret" {
		t.Fatalf("got %v, want the leaf's own internal dynamic skill leafsecret", got)
	}
}

// TestDynamicSkills_DedupByDir: the SAME bare dir declared in both an ancestor and
// the leaf materializes ONCE, most-leafward winning (the leaf's prefix + marker).
func TestDynamicSkills_DedupByDir(t *testing.T) {
	parent := t.TempDir()
	ancestor := filepath.Join(parent, "ancestor")
	leaf := filepath.Join(parent, "leaf")
	writeConfig(t, ancestor, "anc-")
	writeConfig(t, leaf, "xx-")
	writeMarker(t, filepath.Join(ancestor, "construct", "local", "datatype"), 0o755)
	writeMarker(t, filepath.Join(leaf, "construct", "local", "datatype"), 0o755)

	layers := []layer.Layer{
		{Name: "ancestor", Path: ancestor, Intents: skillRows("construct/local")},
		{Name: "leaf", Path: leaf, Intents: skillRows("construct/local")},
	}
	got, err := DynamicSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d entries %v, want 1 (deduped by Dir)", len(got), got)
	}
	// Most-leafward wins: leaf's prefix (xx-) + leaf's marker path.
	if got[0].Name != "xx-datatype" {
		t.Errorf("dedup winner Name = %q, want xx-datatype (leaf wins)", got[0].Name)
	}
	wantMarker := filepath.Join(leaf, "construct", "local", "datatype", ".dynamic-skill")
	if got[0].MarkerPath != wantMarker {
		t.Errorf("dedup winner MarkerPath = %q, want leaf's marker %q", got[0].MarkerPath, wantMarker)
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

// TestDynamicSkills_AbsentSkillDir: a leaf with an absent skill dir (or a non-skill
// intent) selects nothing, never errors.
func TestDynamicSkills_AbsentSkillDir(t *testing.T) {
	root := t.TempDir()
	layers := []layer.Layer{
		{Name: "repo", Path: root, Intents: []intent.Intent{
			{Kind: intent.Skill, Source: "construct/local"}, // dir absent
			{Kind: intent.Prose, Source: "AGENTS.local.md"}, // non-skill intent ignored
		}},
	}
	got, err := DynamicSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %v, want no dynamic skills (absent skill dir)", got)
	}
}
