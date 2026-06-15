package walk

import (
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/plan"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// skill_symlinks.go is the IO-seam lowering that absorbs sync-local-skills.sh
// (the SessionStart hook): it turns a layer's `skill <source-dir>` manifest
// intents into the .claude/skills/ symlinks every in-use harness reads. This is
// the prerequisite for retiring that hook — weave becomes the one renderer of
// .claude/skills/, intent-driven (a `skill construct/local` / `skill
// construct/adapted` manifest row) rather than a hardcoded hook scan.
//
// It ports sync-local-skills.sh's naming EXACTLY (ARCH-DRY — that hook is the
// source of truth for WHERE skills live and HOW their links are named), so the
// live .claude/skills output stays byte-identical:
//
//   - construct/local/<dir>   → .claude/skills/<prefix><dir>  (prefixed)
//   - construct/adapted/<dir> → .claude/skills/<dir>          (bare, no prefix)
//
// each link's target being the upstream skill dir — the relative target
// `../../construct/<local|adapted>/<dir>` the hook writes, reproduced because
// plan.Apply computes rel(dir(dst), src) and Src is exactly that upstream dir
// (so dst=.claude/skills/<name>, src=<root>/construct/local/<dir> ⇒
// ../../construct/local/<dir>, identical to the hook's `target=`).
//
// The prefix is the layer's construct/config.json `localPrefix` (default "xx-",
// the hook's `PREFIX="${PREFIX:-xx-}"` fallback) and is applied ONLY to the
// local source dir — exactly the hook's `sync_skills "$LOCAL_DIR" "$PREFIX"` vs
// `sync_skills "$ADAPTED_DIR" ""` split.
//
// Discovery REUSES GatherSkills' scanSkillDir (ARCH-DRY): the same dir scan +
// prefix application that builds the M3 menu. One consequence is intentional —
// a dir lacking a SKILL.md is SKIPPED here (scanSkillDir treats only
// SKILL.md-bearing dirs as skills), whereas the bash hook would symlink any
// subdirectory. This never diverges on live data (every real skill dir ships a
// SKILL.md), and reusing the one discovery path is the DRY choice.

// LowerSkillSymlinks walks the resolved layers foundation-first and lowers each
// layer's `skill <source-dir>` intents into the .claude/skills/ symlinks that
// absorb sync-local-skills.sh. For each `skill` row it scans the named source
// dir (via the shared scanSkillDir) and emits one plan.Symlink per discovered
// skill: Src = the absolute upstream skill dir, Dst = .claude/skills/<name>
// (name = the discovery-derived <prefix><dir> for local, bare <dir> for
// adapted). IO is confined to this seam (scanSkillDir reads the dir + config);
// the emitted Symlinks are applied by the existing plan.Apply (idempotency +
// relative-target computation come for free, ARCH-DRY).
//
// Layers arrive foundation-first; within a layer, intents are processed in
// manifest order and scanSkillDir returns alphabetically (ReadDir is sorted).
func LowerSkillSymlinks(fs weavefs.FS, layers []layer.Layer) ([]plan.Symlink, error) {
	var links []plan.Symlink
	for _, l := range layers {
		prefix := localPrefix(fs, l.Path)
		for _, in := range l.Intents {
			if in.Kind != intent.Skill {
				continue
			}
			sourceRel := in.Source
			sourceDir := filepath.Join(l.Path, sourceRel)
			// Faithful to the hook's split: the local source dir gets the
			// prefix; every other (adapted) source dir gets the bare name.
			rowPrefix := ""
			if sourceRel == localSkillRel {
				rowPrefix = prefix
			}
			entries, err := scanSkillDir(fs, sourceDir, rowPrefix)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				// e.BodyPath is <sourceDir>/<dir>/SKILL.md; the symlink Src is
				// the skill DIR (the hook links the dir, not the SKILL.md).
				skillDir := filepath.Dir(e.BodyPath)
				links = append(links, plan.Symlink{
					Src: skillDir,
					Dst: filepath.Join(".claude", "skills", e.Name),
				})
			}
		}
	}
	return links, nil
}
