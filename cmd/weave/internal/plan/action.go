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
// Touch), ToolDep (M3: intent.Tool lowers to a ToolDep, applied via the injected
// GoModEditor seam), and MergeSettings (M4: intent.Merge lowers to a
// MergeSettings, applied by reading base + optional local and running the pure
// mergeSettings — merge-settings.sh's port). (Skill has no Action — it feeds the
// M3 SkillIndex, not a file-op slot.)
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
// settings.local.json off disk, runs the pure mergeSettings (the
// merge-settings.sh port — deep dict merge, $merge_keys array union, $remove
// filter, meta-key strip), and writes the result to Target.
type MergeSettings struct {
	Source string // the layer's base settings (e.g. .claude/settings.ariadne.json), repo-relative
	Target string // the merged output (e.g. .claude/settings.json), repo-relative
}

// ToolDep is the lowering of an intent.Tool — declaring the tool owner as a
// substrate dep (derivative) or adding a `go mod edit -tool` directive (owner
// self-walk). This unit emits one ToolDep per `tool` intent (see Plan); Apply
// realizes it via the injected GoModEditor, the one exec seam ported from
// ensure_go_tool_dependency (ARCH-PURE keeps that exec out of the planner).
type ToolDep struct {
	Owner string // absolute owner path
	Path  string // tool path within the owner module (e.g. cmd/sdlc)
}

func (Symlink) isAction()       {}
func (WriteFile) isAction()     {}
func (Mkdir) isAction()         {}
func (Touch) isAction()         {}
func (MergeSettings) isAction() {}
func (ToolDep) isAction()       {}
