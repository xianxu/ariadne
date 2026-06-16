package plan

import (
	"reflect"
	"testing"

	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
)

// SkillSymlinks is PURE: selected entries → .claude/skills/<name> Symlinks whose
// Src is the skill dir (Dir(BodyPath)). No IO — the discovery (GatherSkills) +
// selection (SelectVisible) already happened upstream.
func TestSkillSymlinks(t *testing.T) {
	got := SkillSymlinks([]skill.Entry{
		{Name: "xx-fix", BodyPath: "/ws/ariadne/construct/local/fix/SKILL.md"},
		{Name: "nous-tools", BodyPath: "/ws/nous/construct/local/tools/SKILL.md"},
		{Name: "superpowers-brainstorming", BodyPath: "/ws/ariadne/construct/adapted/superpowers-brainstorming/SKILL.md"},
	})
	want := []Symlink{
		{Src: "/ws/ariadne/construct/local/fix", Dst: ".claude/skills/xx-fix"},
		{Src: "/ws/nous/construct/local/tools", Dst: ".claude/skills/nous-tools"},
		{Src: "/ws/ariadne/construct/adapted/superpowers-brainstorming", Dst: ".claude/skills/superpowers-brainstorming"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SkillSymlinks =\n %#v\nwant\n %#v", got, want)
	}
}

func TestSkillSymlinks_Empty(t *testing.T) {
	if got := SkillSymlinks(nil); len(got) != 0 {
		t.Fatalf("SkillSymlinks(nil) = %v, want empty", got)
	}
}
