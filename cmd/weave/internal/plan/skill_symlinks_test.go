package plan

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
)

// SkillSymlinks is PURE: selected entries → <destDir>/<name> Symlinks whose Src is
// the skill dir (Dir(BodyPath)). No IO — the discovery (GatherSkills) + selection
// (SelectVisible) already happened upstream. ONE renderer for every per-harness
// skill dir (Option B): the same entries lower into .claude/skills (Claude) AND
// .agents/skills (Codex/Gemini), differing only by destDir.
func TestSkillSymlinks(t *testing.T) {
	entries := []skill.Entry{
		{Name: "xx-fix", BodyPath: "/ws/ariadne/construct/local/fix/SKILL.md"},
		{Name: "nous-tools", BodyPath: "/ws/nous/construct/local/tools/SKILL.md"},
		{Name: "superpowers-brainstorming", BodyPath: "/ws/ariadne/construct/adapted/superpowers-brainstorming/SKILL.md"},
	}
	for _, dir := range []string{".claude/skills", ".agents/skills"} {
		got := SkillSymlinks(entries, dir)
		want := []Symlink{
			{Src: "/ws/ariadne/construct/local/fix", Dst: dir + "/xx-fix"},
			{Src: "/ws/nous/construct/local/tools", Dst: dir + "/nous-tools"},
			{Src: "/ws/ariadne/construct/adapted/superpowers-brainstorming", Dst: dir + "/superpowers-brainstorming"},
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("SkillSymlinks(%q) =\n %#v\nwant\n %#v", dir, got, want)
		}
	}
}

func TestSkillSymlinks_Empty(t *testing.T) {
	if got := SkillSymlinks(nil, ".claude/skills"); len(got) != 0 {
		t.Fatalf("SkillSymlinks(nil) = %v, want empty", got)
	}
}

// TestSkillSymlinks_DynamicLowersToGenerated (#115 M3): a Dynamic entry's BodyPath
// is the compiling repo's construct/generated/<dir>/SKILL.md, so SkillSymlinks —
// UNCHANGED, still Src = Dir(BodyPath) — lowers its link straight at this-repo's
// materialized copy, NOT the owner. The lowering switch collapsed into the entry's
// BodyPath (set by GatherSkills), so no branch is needed here.
func TestSkillSymlinks_DynamicLowersToGenerated(t *testing.T) {
	entries := []skill.Entry{
		{Name: "xx-datatype", Dynamic: true, BodyPath: "/ws/derived/construct/generated/datatype/SKILL.md"},
		{Name: "xx-sdlc", BodyPath: "/ws/ariadne/construct/local/sdlc/SKILL.md"},
	}
	got := SkillSymlinks(entries, ".claude/skills")
	want := []Symlink{
		{Src: "/ws/derived/construct/generated/datatype", Dst: ".claude/skills/xx-datatype"},
		{Src: "/ws/ariadne/construct/local/sdlc", Dst: ".claude/skills/xx-sdlc"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SkillSymlinks(dynamic) =\n %#v\nwant\n %#v", got, want)
	}
}
