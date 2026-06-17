package walk

import (
	"os"
	"path/filepath"

	"github.com/xianxu/ariadne/cmd/weave/internal/intent"
	"github.com/xianxu/ariadne/cmd/weave/internal/layer"
	"github.com/xianxu/ariadne/cmd/weave/internal/weavefs"
)

// dynamic.go is the dynamic-skill SELECTION side of the #111 generate stage: it
// names which package dirs carry an executable `.dynamic-skill` that weave should
// exec at compile time. The actual exec is the injected weavefs.Runner (the
// generate stage in cmd/weave); this file is selection-only.
//
// LEAF-ONLY (the byte-pristine guarantee): it scans ONLY the leaf layer's `skill`
// intents (layers[len-1] — the repo being compiled), never an ancestor's. weave
// iterates ALL resolved layers at their real on-disk paths, so "no inheritance
// symlinks post-#104" is necessary but NOT sufficient — without leaf-scoping a
// derivative's compile would find ariadne's marker and mutate ariadne's tree.
//
// ARCH-DRY: the scan set reuses the SAME `skill <dir>` intents GatherSkills
// consumes — no second hardcoded dir list (the skill-system "one discovery").
// construct/adapted is excluded (foreign-origin superpowers; out of scope, #111).

// dynamicSkillRel is the tracked, language-neutral marker filename. Its presence
// (with an executable bit) in a skill package dir makes that package a dynamic
// skill weave regenerates at compile time.
const dynamicSkillRel = ".dynamic-skill"

// DynamicSkillDirs returns the absolute package dirs under the LEAF layer's
// `skill <dir>` intents that contain an EXECUTABLE `.dynamic-skill`, excluding the
// construct/adapted dir. The order is intent order then alphabetical within a dir
// (ReadDir is sorted). An absent skill dir contributes nothing (not an error — a
// layer need not ship dynamic skills). Listing/stat go through the injected FS
// seam (the IO edge); the keep/exclude/executable decision is pure.
func DynamicSkillDirs(fs weavefs.FS, layers []layer.Layer) ([]string, error) {
	if len(layers) == 0 {
		return nil, nil
	}
	leaf := layers[len(layers)-1] // leaf-only — never an ancestor

	var dirs []string
	for _, in := range leaf.Intents {
		if in.Kind != intent.Skill {
			continue
		}
		if in.Source == adaptedSkillRel { // adapted = foreign origin, excluded (#111)
			continue
		}
		sourceDir := filepath.Join(leaf.Path, in.Source)
		dirents, err := fs.ReadDir(sourceDir)
		if err != nil {
			continue // absent / unreadable skill dir ⇒ no dynamic skills here
		}
		for _, de := range dirents {
			if !de.IsDir() {
				continue
			}
			pkgDir := filepath.Join(sourceDir, de.Name())
			fi, serr := fs.Stat(filepath.Join(pkgDir, dynamicSkillRel))
			if serr != nil || fi.IsDir() {
				continue // no marker (or a dir by that name) ⇒ not a dynamic skill
			}
			if !isExecutable(fi.Mode()) {
				continue // a non-executable marker is ignored (it must be runnable)
			}
			dirs = append(dirs, pkgDir)
		}
	}
	return dirs, nil
}

// isExecutable reports whether any executable bit is set on mode (owner, group,
// or other). Pure — the mode is OBSERVED at the IO edge (fs.Stat) and the bit
// test stays here, so this file's selection logic is unit-testable.
func isExecutable(mode os.FileMode) bool {
	return mode&0o111 != 0
}
