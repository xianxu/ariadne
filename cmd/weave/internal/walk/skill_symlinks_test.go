package walk

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// LowerSkillSymlinks is the IO-seam lowering that absorbs sync-local-skills.sh:
// each `skill <source-dir>` intent → one plan.Symlink per discovered skill,
// .claude/skills/<prefix><dir> (local, prefixed) or .claude/skills/<dir>
// (adapted, bare). Exercised against a real OSFS rooted at t.TempDir (the seam
// end-to-end, no mocks). The naming is cross-checked against the bash hook so
// the live .claude/skills output stays byte-identical.

// skillRows is the canonical pair of manifest intents that drive the lowering:
// local (prefixed) then adapted (bare), matching the hook's two sync_skills
// calls and base.manifest's `symlink construct/local` / `symlink
// construct/adapted` source rows.
func skillRows() []intent.Intent {
	return []intent.Intent{
		{Kind: intent.Skill, Source: localSkillRel, Target: localSkillRel},
		{Kind: intent.Skill, Source: adaptedSkillRel, Target: adaptedSkillRel},
	}
}

func TestLowerSkillSymlinks_PrefixAndBare(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "construct"), 0o755); err != nil {
		t.Fatal(err)
	}
	// A configured prefix (the hook reads construct/config.json's localPrefix).
	if err := os.WriteFile(filepath.Join(root, "construct", "config.json"),
		[]byte(`{"localPrefix": "zz-"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(root, "construct", "local", "skillA"),
		"skillA", "a local skill", "BODY A")
	writeSkill(t, filepath.Join(root, "construct", "adapted", "skillB"),
		"skillB", "an adapted skill", "BODY B")

	layers := []layer.Layer{{Name: "repo", Path: root, Intents: skillRows()}}

	got, err := LowerSkillSymlinks(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}

	// Local skill is prefixed; adapted keeps its bare dir name. The Src is the
	// upstream skill DIR (not the SKILL.md) — plan.Apply turns each into the
	// hook's relative `../../construct/<local|adapted>/<dir>` target.
	want := []plan.Symlink{
		{Src: filepath.Join(root, "construct", "local", "skillA"), Dst: ".claude/skills/zz-skillA"},
		{Src: filepath.Join(root, "construct", "adapted", "skillB"), Dst: ".claude/skills/skillB"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LowerSkillSymlinks =\n %#v\nwant\n %#v", got, want)
	}
}

func TestLowerSkillSymlinks_DefaultPrefixWhenNoConfig(t *testing.T) {
	root := t.TempDir()
	// No construct/config.json ⇒ default prefix xx- (sync-local-skills.sh:19).
	writeSkill(t, filepath.Join(root, "construct", "local", "fix"),
		"fix", "fix markers", "FIX BODY")

	layers := []layer.Layer{{Name: "repo", Path: root, Intents: skillRows()}}

	got, err := LowerSkillSymlinks(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	want := []plan.Symlink{
		{Src: filepath.Join(root, "construct", "local", "fix"), Dst: ".claude/skills/xx-fix"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LowerSkillSymlinks =\n %#v\nwant\n %#v", got, want)
	}
}

func TestLowerSkillSymlinks_DirWithoutSKILLmdSkipped(t *testing.T) {
	root := t.TempDir()
	// A dir under construct/local with no SKILL.md is NOT a skill — skip it
	// (reusing scanSkillDir's discovery; never matters on live data, where every
	// skill dir ships a SKILL.md).
	if err := os.MkdirAll(filepath.Join(root, "construct", "local", "notaskill"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeSkill(t, filepath.Join(root, "construct", "local", "real"),
		"real", "a real skill", "BODY")

	layers := []layer.Layer{{Name: "repo", Path: root, Intents: skillRows()}}

	got, err := LowerSkillSymlinks(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	want := []plan.Symlink{
		{Src: filepath.Join(root, "construct", "local", "real"), Dst: ".claude/skills/xx-real"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LowerSkillSymlinks =\n %#v\nwant\n %#v", got, want)
	}
}

func TestLowerSkillSymlinks_AcrossLayersFoundationFirst(t *testing.T) {
	// Two layers, foundation-first: base ships a local skill, derived ships an
	// adapted skill. Symlinks come out in layer order (base's first), each layer
	// scanned in manifest-row order. Each layer's Src is rooted at its OWN path.
	parent := t.TempDir()
	base := filepath.Join(parent, "base")
	derived := filepath.Join(parent, "derived")
	writeSkill(t, filepath.Join(base, "construct", "local", "sdlc"),
		"sdlc", "SDLC gates", "BASE SDLC")
	writeSkill(t, filepath.Join(derived, "construct", "adapted", "superpowers-brainstorming"),
		"superpowers-brainstorming", "brainstorm", "BRAINSTORM")

	layers := []layer.Layer{
		{Name: "base", Path: base, Intents: skillRows()},
		{Name: "derived", Path: derived, Intents: skillRows()},
	}

	got, err := LowerSkillSymlinks(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	want := []plan.Symlink{
		{Src: filepath.Join(base, "construct", "local", "sdlc"), Dst: ".claude/skills/xx-sdlc"},
		{Src: filepath.Join(derived, "construct", "adapted", "superpowers-brainstorming"), Dst: ".claude/skills/superpowers-brainstorming"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LowerSkillSymlinks =\n %#v\nwant\n %#v", got, want)
	}
}

func TestLowerSkillSymlinks_AppliesToHookIdenticalTargets(t *testing.T) {
	// End-to-end through plan.Apply: the emitted Symlinks must realize the SAME
	// relative link targets the bash hook writes (`../../construct/<dir>/<skill>`),
	// proving byte-identical .claude/skills output. This is the parity check the
	// config migration's golden-diff will rely on.
	root := t.TempDir()
	writeSkill(t, filepath.Join(root, "construct", "local", "sdlc"),
		"sdlc", "SDLC gates", "BODY")
	writeSkill(t, filepath.Join(root, "construct", "adapted", "superpowers-brainstorming"),
		"superpowers-brainstorming", "brainstorm", "BODY")

	layers := []layer.Layer{{Name: "repo", Path: root, Intents: skillRows()}}
	links, err := LowerSkillSymlinks(weavefs.OSFS{}, layers)
	if err != nil {
		t.Fatal(err)
	}
	actions := make([]plan.Action, len(links))
	for i, l := range links {
		actions[i] = l
	}
	if err := plan.Apply(weavefs.OSFS{}, root, actions); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	// Read back each link target and compare to the hook's `target=` string.
	cases := map[string]string{
		filepath.Join(root, ".claude", "skills", "xx-sdlc"):                   "../../construct/local/sdlc",
		filepath.Join(root, ".claude", "skills", "superpowers-brainstorming"): "../../construct/adapted/superpowers-brainstorming",
	}
	for link, wantTarget := range cases {
		fi, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("expected symlink %s: %v", link, err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("%s is not a symlink", link)
		}
		got, err := os.Readlink(link)
		if err != nil {
			t.Fatal(err)
		}
		if got != wantTarget {
			t.Errorf("readlink %s = %q, want %q (hook-identical)", link, got, wantTarget)
		}
	}
}
