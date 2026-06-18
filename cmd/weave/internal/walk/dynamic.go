package walk

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// dynamic.go is the dynamic-skill SELECTION side of the #111/#115 generate stage:
// it names which package dirs carry an executable `.dynamic-skill` that weave
// should exec at compile time, and where each one materializes. The actual exec is
// the injected weavefs.Runner (the generate stage in cmd/weave); this file is
// selection-only.
//
// ALL-LAYERS, VISIBLE-SET (#115 M3): unlike the retired leaf-only DynamicSkillDirs,
// DynamicSkills scans EVERY resolved layer's `skill <dir>` intents — so an
// ANCESTOR-owned marker (e.g. ariadne's datatype) is selected when compiling a
// derivative. The byte-pristine guarantee no longer rests on leaf-only selection;
// it rests on LEAF-ROOTED OUTPUT: the generate stage runs the (possibly
// ancestor-owned) marker with cwd = the compiling repo's root and a repo-relative
// --output, so materialization lands in the COMPILING repo's tree, never the
// ancestor's. A dynamic skill is included only if VISIBLE to the compiling repo R
// via intent.Selected (an ancestor's INTERNAL dynamic skill is excluded, exactly
// like every other artifact — ARCH-DRY: the SAME visibility rule the planner uses).
//
// ARCH-DRY: the scan set reuses the SAME `skill <dir>` intents GatherSkills
// consumes — no second hardcoded dir list (the skill-system "one discovery").
// construct/adapted is excluded (foreign-origin superpowers; out of scope, #111).

// dynamicSkillRel is the tracked, language-neutral marker filename. Its presence
// (with an executable bit) in a skill package dir makes that package a dynamic
// skill weave regenerates at compile time.
const dynamicSkillRel = ".dynamic-skill"

// GeneratedRel is the repo-relative root of the per-repo dynamic-skill
// materialization tree: a dynamic skill <dir>'s body renders to
// <repo>/construct/generated/<dir>/SKILL.md (gitignored, regenerated every
// compile). This is the SINGLE SOURCE for that convention — the write-path
// (DynamicSkills.OutputRel), the read-path (scanSkillDir's BodyPath), the
// gitignore entry (plan.GeneratedRuntimeGitignoreEntries), and the prune scope
// (plan.PruneGenerated) all derive from it. They MUST agree or the materialized
// body and the discovered body diverge (generation writes one place, discovery
// reads another → silent empty description + dangling symlink — M3-review #1).
const GeneratedRel = "construct/generated"

// GeneratedSkillDir is the repo-relative materialization dir for a dynamic-skill
// package <dir>: construct/generated/<dir>.
func GeneratedSkillDir(dir string) string { return filepath.Join(GeneratedRel, dir) }

// GeneratedSkillBody is the absolute materialized SKILL.md path for a dynamic-skill
// package <dir> in the repo rooted at root.
func GeneratedSkillBody(root, dir string) string {
	return filepath.Join(root, GeneratedRel, dir, "SKILL.md")
}

// dynamicMarker reports whether pkgDir carries an EXECUTABLE `.dynamic-skill`
// marker (returning its absolute path). This is the SINGLE predicate the
// generate-stage selection (DynamicSkills) and discovery (scanSkillDir) must agree
// on (M3-review #2); a missing marker, a dir by that name, or a non-executable
// marker is not one. The Stat is the IO edge; the keep decision is pure.
func dynamicMarker(fs weavefs.FS, pkgDir string) (markerPath string, ok bool) {
	markerPath = filepath.Join(pkgDir, dynamicSkillRel)
	fi, err := fs.Stat(markerPath)
	if err != nil || fi.IsDir() || !isExecutable(fi.Mode()) {
		return "", false
	}
	return markerPath, true
}

// DynamicSkill is one selected dynamic skill across the layer graph: its prefixed
// skill Name, its bare package Dir, the absolute path to the (possibly
// ancestor-owned) marker the generate stage execs, and the repo-relative output
// dir (construct/generated/<Dir>) the compiling repo materializes into. Dir is the
// dedup key (a name declared in multiple layers materializes once — most-leafward
// wins); OutputRel/Name use the bare Dir / the declaring layer's prefix.
type DynamicSkill struct {
	Name       string // prefixed skill name (e.g. "xx-datatype") — the discovery name
	Dir        string // bare package dir name (e.g. "datatype") — the output-path key
	MarkerPath string // absolute path to the owner's executable .dynamic-skill
	OutputRel  string // repo-relative materialization dir, "construct/generated/<Dir>"
}

// DynamicSkills returns the dynamic skills VISIBLE to the compiling repo R (the
// leaf), scanning ALL resolved layers' `skill <dir>` intents for package dirs that
// carry an EXECUTABLE `.dynamic-skill`. A dynamic skill is included only when
// intent.Selected(in.Visibility, i==leafIdx) holds — so an ancestor's INTERNAL
// dynamic skill is excluded (the same visibility filter the planner applies,
// ARCH-DRY). construct/adapted is excluded (foreign origin). MarkerPath is the
// owner's marker (where it physically lives), but OutputRel is repo-relative so the
// generate stage materializes under R's root. Deduped by Dir (most-leafward wins —
// a derivative's re-declaration shadows an ancestor's) and sorted by Dir for
// determinism. Listing/stat go through the injected FS seam (the IO edge); the
// keep/exclude/executable/visibility decisions are pure.
func DynamicSkills(fs weavefs.FS, layers []layer.Layer) ([]DynamicSkill, error) {
	if len(layers) == 0 {
		return nil, nil
	}
	leafIdx := len(layers) - 1

	byDir := map[string]DynamicSkill{}
	for i, l := range layers {
		prefix := skillPrefix(fs, l.Path)
		for _, in := range l.Intents {
			if in.Kind != intent.Skill {
				continue
			}
			if in.Source == adaptedSkillRel { // adapted = foreign origin, excluded (#111)
				continue
			}
			// VISIBILITY filter (#115 M3): an ancestor's INTERNAL dynamic skill is not
			// visible to R, so it's excluded — the SAME intent.Selected rule the planner
			// uses for every artifact (ARCH-DRY).
			if !intent.Selected(in.Visibility, i == leafIdx) {
				continue
			}
			rowPrefix := prefix // adapted already skipped above
			sourceDir := filepath.Join(l.Path, in.Source)
			dirents, err := fs.ReadDir(sourceDir)
			if err != nil {
				continue // absent / unreadable skill dir ⇒ no dynamic skills here
			}
			for _, de := range dirents {
				if !de.IsDir() {
					continue
				}
				pkgDir := filepath.Join(sourceDir, de.Name())
				markerPath, ok := dynamicMarker(fs, pkgDir)
				if !ok {
					continue // no executable .dynamic-skill ⇒ not a dynamic skill
				}
				dir := de.Name()
				// Dedup by Dir, most-leafward wins: layers iterate foundation-first, so
				// a later (more-leafward) layer's row overwrites an ancestor's entry for
				// the same bare dir (the prefix + marker of the winning layer).
				byDir[dir] = DynamicSkill{
					Name:       rowPrefix + dir,
					Dir:        dir,
					MarkerPath: markerPath,
					OutputRel:  GeneratedSkillDir(dir),
				}
			}
		}
	}

	out := make([]DynamicSkill, 0, len(byDir))
	for _, ds := range byDir {
		out = append(out, ds)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out, nil
}

// isExecutable reports whether any executable bit is set on mode (owner, group,
// or other). Pure — the mode is OBSERVED at the IO edge (fs.Stat) and the bit
// test stays here, so this file's selection logic is unit-testable.
func isExecutable(mode os.FileMode) bool {
	return mode&0o111 != 0
}
