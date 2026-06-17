package plan

import (
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/skill"
)

// SkillSymlinks lowers the selected skill entries to <destDir>/<name> symlinks —
// PURE, derived from the SAME entries every harness rendering uses (#104), so the
// renderings of a repo's skills can never diverge (one scan in walk.GatherSkills,
// pure renderings here). ONE renderer for every per-harness skill dir (Option B,
// #107): destDir is ".claude/skills" (Claude) or ".agents/skills" (Codex/Gemini);
// the Union calls it once per distinct dir with the SAME selected set, so each
// harness's dir holds the identical skills (ARCH-DRY). Each link's Src is the
// skill's source dir (the dir holding its SKILL.md, in the SOURCE LAYER — so a
// derivative's link points straight at the owning layer); plan.Apply computes the
// relative target + idempotency.
func SkillSymlinks(entries []skill.Entry, destDir string) []Symlink {
	out := make([]Symlink, 0, len(entries))
	for _, e := range entries {
		out = append(out, Symlink{
			Src: filepath.Dir(e.BodyPath),
			Dst: filepath.Join(destDir, e.Name),
		})
	}
	return out
}
