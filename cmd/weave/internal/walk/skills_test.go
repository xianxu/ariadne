package walk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// skillRows builds `export skill <dir>` intents — the manifest declarations the
// intent-driven GatherSkills reads (#104). Visibility defaults to Export; the
// visibility tests set it explicitly.
func skillRows(sources ...string) []intent.Intent {
	rows := make([]intent.Intent, len(sources))
	for i, s := range sources {
		rows[i] = intent.Intent{Kind: intent.Skill, Source: s}
	}
	return rows
}

// writeConfig writes a layer's construct/config.json localPrefix — pins a
// predictable prefix where a test asserts specific names (otherwise skillPrefix
// defaults to the repo-name basename, #104 M2).
func writeConfig(t *testing.T, root, prefix string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "construct", "config.json"),
		[]byte(`{"localPrefix": "`+prefix+`"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

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
		{Name: "base", Path: base, Intents: skillRows("construct/local", "construct/adapted")},
		{Name: "derived", Path: derived, Intents: skillRows("construct/local")},
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

func TestGatherSkills_RepoNamePrefixWhenNoConfig(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	// No construct/config.json ⇒ default prefix = the layer's REPO NAME (#104 M2):
	// basename("repo") + "-" → "repo-".
	writeSkill(t, filepath.Join(root, "construct", "local", "fix"),
		"fix", "fix markers", "FIX BODY")

	entries, err := GatherSkills(weavefs.OSFS{},
		[]layer.Layer{{Name: "repo", Path: root, Intents: skillRows("construct/local")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "repo-fix" {
		t.Fatalf("got %#v, want one repo-fix entry under the repo-name default", entries)
	}
}

func TestGatherSkills_LayerWithoutSkillDirs(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "bare")
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Declares `skill construct/local` but the dir is absent — scanSkillDir treats
	// an absent source dir as no skills (not an error).
	entries, err := GatherSkills(weavefs.OSFS{},
		[]layer.Layer{{Name: "bare", Path: root, Intents: skillRows("construct/local")}})
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
	writeConfig(t, root, "xx-")
	// A dir under construct/local with no SKILL.md is NOT a skill — skip it.
	if err := os.MkdirAll(filepath.Join(root, "construct", "local", "notaskill"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(root, "construct", "local", "real"),
		"xx-real", "a real skill", "BODY")

	entries, err := GatherSkills(weavefs.OSFS{},
		[]layer.Layer{{Name: "repo", Path: root, Intents: skillRows("construct/local")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "xx-real" {
		t.Fatalf("got %#v, want only the SKILL.md-bearing dir", entries)
	}
}

func TestGatherSkills_IntentDrivenVisibilityAndNonStandardDir(t *testing.T) {
	// The v2 capability: GatherSkills reads `skill <dir>` INTENTS (not hardcoded
	// dirs), so a skill in a NON-construct/local dir is found, and the row's
	// visibility + the layer index are stamped on each entry — so SelectVisible
	// can drop an ancestor's internal skill. construct/adapted stays bare;
	// other dirs (here construct/priv) get the prefix.
	root := t.TempDir()
	writeConfig(t, root, "xx-")
	writeSkill(t, filepath.Join(root, "construct", "mine", "tool"), "tool", "exported tool", "T")
	writeSkill(t, filepath.Join(root, "construct", "priv", "secret"), "secret", "private", "S")
	layers := []layer.Layer{{Name: "repo", Path: root, Intents: []intent.Intent{
		{Kind: intent.Skill, Visibility: intent.Export, Source: "construct/mine"},
		{Kind: intent.Skill, Visibility: intent.Internal, Source: "construct/priv"},
	}}}

	entries, err := GatherSkills(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]skill.Entry{}
	for _, e := range entries {
		got[e.Name] = e
	}
	mine, ok := got["xx-tool"] // prefixed (not construct/adapted), from construct/mine
	if !ok {
		t.Fatalf("xx-tool not gathered from the non-standard dir; got %v", names2(entries))
	}
	if mine.Visibility != intent.Export || mine.LayerIndex != 0 {
		t.Errorf("xx-tool visibility/layer = %v/%d, want Export/0", mine.Visibility, mine.LayerIndex)
	}
	priv, ok := got["xx-secret"]
	if !ok {
		t.Fatalf("xx-secret not gathered; got %v", names2(entries))
	}
	if priv.Visibility != intent.Internal {
		t.Errorf("xx-secret visibility = %v, want Internal (from the internal row)", priv.Visibility)
	}
}

func names2(es []skill.Entry) []string {
	out := make([]string, 0, len(es))
	for _, e := range es {
		out = append(out, e.Name)
	}
	return out
}
