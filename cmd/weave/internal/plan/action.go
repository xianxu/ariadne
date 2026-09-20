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
// The concrete op structs declared in THIS FILE are the set — do not restate it
// in prose anywhere, here or elsewhere: a hand-written list derives from nothing
// and goes stale on every new verb (which is defect 3 of ariadne#239, one level
// down). Read the `isAction()` markers at the bottom of the file for the roster.
//
// Not every manifest verb has an Action: `skill` feeds the SkillIndex, not a
// file-op slot, and the retired `tool` verb (#95 M5) lowers to nothing at all —
// Go-tool ownership is location-based (construct/dev-aliases.sh) and deps come
// from `weave link`, so weave never edits go.mod.
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

// SeedOnce is a WRITE-ONCE real-file copy of an upstream Src into Dst — the
// ownership sibling of Seed. Seed TRACKS upstream (content-tracking, converges
// on every compile) because its content is upstream-owned; SeedOnce hands the
// slot to the REPO on first write and never touches it again, whatever it later
// contains. It exists for the root Makefile: a repo's own front door, which a
// greenfield repo should get for free but an adopting repo must keep (#239).
//
// "Present" means anything that is NOT a symlink — a regular file, a directory.
// A SYMLINK is NOT presence: it is weave's own pre-#239 `symlink Makefile`
// lowering, and materializing it is the #225 convergence. See applySeedOnce.
type SeedOnce struct {
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

// MergeSettings is the lowering of one or more intent.Merge rows sharing a
// Target — the JSON settings cascade across ordered layer sources and the
// sibling settings.local.json. The planner records only path facts (pure);
// Apply reads Sources + the optional sibling local off disk, runs the pure
// settingsx.MergeChain, and writes the result to Target.
type MergeSettings struct {
	Sources []string // ordered source settings files, usually absolute layer paths
	Target  string   // the merged output (e.g. .claude/settings.json), repo-relative
}

func (Symlink) isAction()       {}
func (WriteFile) isAction()     {}
func (Mkdir) isAction()         {}
func (Seed) isAction()          {}
func (SeedOnce) isAction()      {}
func (Touch) isAction()         {}
func (MergeSettings) isAction() {}
