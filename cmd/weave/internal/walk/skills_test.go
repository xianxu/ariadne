package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// GatherSkills is the IO seam that ports sync-local-skills.sh's discovery:
// per layer, scan construct/local/*/ (prefixed) + construct/adapted/*/ (bare),
// parse each SKILL.md's frontmatter description. Exercised against a real OSFS
// rooted at t.TempDir (the seam end-to-end, no mocks).

func writeSkill(t *testing.T, dir, name, desc, body string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: " + desc + "\n---\n\n" + body + "\n"
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// buildSkillFixture lays a base layer (with a local skill `sdlc` + an adapted
// skill `superpowers-brainstorming`) and a derived layer (a local skill
// `issues`) as sibling dirs, returning their paths. Each layer ships its own
// construct/config.json with the xx- prefix.
func buildSkillFixture(t *testing.T) (base, derived string) {
	t.Helper()
	parent := t.TempDir()
	base = filepath.Join(parent, "base")
	derived = filepath.Join(parent, "derived")

	for _, root := range []string{base, derived} {
		if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "construct", "config.json"),
			[]byte(`{"localPrefix": "xx-"}`+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeSkill(t, filepath.Join(base, "construct", "local", "sdlc"),
		"sdlc", "SDLC checkpoint gates", "BASE SDLC BODY")
	writeSkill(t, filepath.Join(base, "construct", "adapted", "superpowers-brainstorming"),
		"superpowers-brainstorming", "Brainstorm before building", "BRAINSTORM BODY")
	writeSkill(t, filepath.Join(derived, "construct", "local", "issues"),
		"xx-issues", "Issue files in workshop/issues", "ISSUES BODY")

	return base, derived
}

func TestGatherSkills_AcrossLayers(t *testing.T) {
	base, derived := buildSkillFixture(t)
	layers := []layer.Layer{
		{Name: "base", Path: base},
		{Name: "derived", Path: derived},
	}

	entries, err := GatherSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}

	// Foundation-first: base's skills before derived's. Within a layer, local
	// then adapted, each alphabetical (ReadDir is sorted). Names are
	// discovery-derived: <prefix><dir> for local, bare <dir> for adapted.
	got := map[string]skill.Entry{}
	var names []string
	for _, e := range entries {
		got[e.Name] = e
		names = append(names, e.Name)
	}

	wantOrder := []string{"xx-sdlc", "superpowers-brainstorming", "xx-issues"}
	if len(names) != len(wantOrder) {
		t.Fatalf("got %d entries %v, want %d %v", len(names), names, len(wantOrder), wantOrder)
	}
	for i, w := range wantOrder {
		if names[i] != w {
			t.Errorf("entry[%d] name = %q, want %q (order %v)", i, names[i], w, names)
		}
	}

	// Description comes from frontmatter; the namespaced name is dir-derived
	// (NOT the frontmatter name — sdlc's dir is `sdlc`, prefixed to `xx-sdlc`).
	if d := got["xx-sdlc"].Description; d != "SDLC checkpoint gates" {
		t.Errorf("xx-sdlc description = %q, want frontmatter desc", d)
	}
	// BodyPath points at the actual SKILL.md.
	wantBody := filepath.Join(base, "construct", "local", "sdlc", "SKILL.md")
	if got["xx-sdlc"].BodyPath != wantBody {
		t.Errorf("xx-sdlc BodyPath = %q, want %q", got["xx-sdlc"].BodyPath, wantBody)
	}
	// Adapted skill keeps its bare dir name (no prefix).
	if _, ok := got["superpowers-brainstorming"]; !ok {
		t.Error("adapted skill should be named by bare dir (no prefix)")
	}
}

func TestGatherSkills_DefaultPrefixWhenNoConfig(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	// No construct/config.json ⇒ default prefix xx- (sync-local-skills.sh:19).
	writeSkill(t, filepath.Join(root, "construct", "local", "fix"),
		"xx-fix", "fix markers", "FIX BODY")

	entries, err := GatherSkills(weavefs.OSFS{}, []layer.Layer{{Name: "repo", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "xx-fix" {
		t.Fatalf("got %#v, want one xx-fix entry under the default prefix", entries)
	}
}

func TestGatherSkills_LayerWithoutSkillDirs(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "bare")
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	// No construct/local or construct/adapted at all — must not error.
	entries, err := GatherSkills(weavefs.OSFS{}, []layer.Layer{{Name: "bare", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("bare layer yielded %#v, want no skills", entries)
	}
}

func TestGatherSkills_DirWithoutSKILLmdSkipped(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	// A dir under construct/local with no SKILL.md is NOT a skill — skip it.
	if err := os.MkdirAll(filepath.Join(root, "construct", "local", "notaskill"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(root, "construct", "local", "real"),
		"xx-real", "a real skill", "BODY")

	entries, err := GatherSkills(weavefs.OSFS{}, []layer.Layer{{Name: "repo", Path: root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "xx-real" {
		t.Fatalf("got %#v, want only the SKILL.md-bearing dir", entries)
	}
}
