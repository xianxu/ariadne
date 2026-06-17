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
