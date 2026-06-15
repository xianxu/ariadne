// Package plan is the pure planner: it lowers a foundation-first []layer.Layer
// into an ordered []Action — the pending filesystem operations weave will
// apply. No IO (ARCH-PURE): the planner only COMPUTES Actions; a later IO seam
// (part 2: Apply over an FS) EXECUTES them. The lowering is one switch over
// intent.Kind, ported from setup.sh's walk_manifest dispatch (ARCH-DRY).
package plan

// Action is one pending filesystem operation — a sum type over the concrete
// op structs below. It is a closed interface (the isAction marker keeps the
// set in this package), so a type switch in the IO seam handles every case.
//
// The set covers the file-ops this unit lowers (Symlink, WriteFile, Mkdir,
// Touch, Seed) and MergeSettings (M4: intent.Merge lowers to a MergeSettings,
// applied by reading base + optional local and running the pure settingsx.Merge
// — merge-settings.sh's port). (Skill has no Action — it feeds the M3
// SkillIndex, not a file-op slot.) The retired `tool` verb (#95 M5) lowers to
// no Action: Go-tool ownership is location-based (construct/dev-aliases.sh) and
// deps come from `weave link`, so weave never edits go.mod.
type Action interface{ isAction() }

// Symlink creates a symlink at Dst pointing to Src. Lowered from an
// intent.Symlink (walk_manifest: create_symlink "$upstream/$source"
// "$TARGET_DIR/$target"). Src is the absolute upstream path; Dst the
// target-relative path.
type Symlink struct {
	Src string
	Dst string
}

// WriteFile writes Content to Path (creating parents). Lowered from the
// composed prose (→ AGENTS.md) and from intent.Touch (empty Content), the
// near-identity of walk_manifest's `touch` case.
type WriteFile struct {
	Path    string
	Content string
}

// Mkdir creates an empty directory at Path. Lowered from intent.Scaffold
// (walk_manifest: create_scaffold "$TARGET_DIR/$target").
type Mkdir struct {
	Path string
}

// Seed is a content-tracking real-file COPY of an upstream Src into Dst —
// lowered from an intent.Seed (walk_manifest: create_seed "$upstream/$source"
// "$TARGET_DIR/$target"). Unlike Symlink, the result is a standalone file that
// survives a clone with no upstream beside it (first-run entrypoints like
// bootstrap.sh that must run before any substrate is present, so they
// definitionally can't be symlinks). A seed is a *flattened symlink*: its
// content is upstream-owned and carries no local edits, so it TRACKS upstream —
// created on first run, refreshed when it drifts, a silent no-op when already
// identical (the convergence #45 added). Src is the absolute upstream path; Dst
// the target-relative path. (Distinct from WriteFile, whose Content the planner
// already holds; a Seed's content lives in Src on disk and is read by the IO
// seam — keeping the planner pure.) A missing Src is non-fatal: the seam warns
// and leaves the target intact, never erroring the walk.
type Seed struct {
	Src string
	Dst string
}

// Touch ensures an EMPTY file exists at Path, create-if-missing — it does NOT
// overwrite an existing file. Lowered from intent.Touch, the faithful port of
// walk_manifest's `touch` case (`if [[ ! -f ]] then touch`, setup.sh:347). This
// is distinct from WriteFile (which writes content unconditionally): a Touch
// target like workshop/lessons.md accumulates real content over time, and weave
// must NOT clobber it (the golden-diff harness surfaced that an unconditional
// WriteFile{content:""} would destroy it). Idempotent: a no-op when the file is
// already present (with any content).
type Touch struct {
	Path string
}

// MergeSettings is the lowering of an intent.Merge — the JSON settings cascade
// (the base settings.ariadne.json deep-merged UNDER the sibling
// settings.local.json). The planner emits one per `merge` row, recording only
// the path facts (pure); Apply reads Source + the optional sibling
// settings.local.json off disk, runs the pure settingsx.Merge (the
// merge-settings.sh port — deep dict merge, $merge_keys array union, $remove
// filter, meta-key strip), and writes the result to Target.
type MergeSettings struct {
	Source string // the layer's base settings (e.g. .claude/settings.ariadne.json), repo-relative
	Target string // the merged output (e.g. .claude/settings.json), repo-relative
}

func (Symlink) isAction()       {}
func (WriteFile) isAction()     {}
func (Mkdir) isAction()         {}
func (Seed) isAction()          {}
func (Touch) isAction()         {}
func (MergeSettings) isAction() {}
