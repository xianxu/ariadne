package plan

import (
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
)

// SkillSymlinks lowers selected skill entries to the claude .claude/skills/<name>
// links — PURE, derived from the SAME entries the menu (skill.Build) uses (#104),
// so the two harness renderings of a repo's skills can never diverge (one scan in
// walk.GatherSkills, two pure renderings here + in skill.Build). Each link's Src is
// the skill's source dir (the dir holding its SKILL.md, in the SOURCE LAYER — so a
// derivative's link points straight at the owning layer); plan.Apply computes the
// relative target + idempotency. This replaces the deleted IO walk.LowerSkillSymlinks.
func SkillSymlinks(entries []skill.Entry) []Symlink {
	out := make([]Symlink, 0, len(entries))
	for _, e := range entries {
		out = append(out, Symlink{
			Src: filepath.Dir(e.BodyPath),
			Dst: filepath.Join(".claude", "skills", e.Name),
		})
	}
	return out
}
