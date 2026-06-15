// Package intent models a layer's base.manifest as typed intents and parses
// the manifest text into them. An Intent is one manifest entry: a Kind plus a
// Source and a Target (Target defaults to Source). Pure — no IO (ARCH-PURE);
// the manifest file is read at the boundary and its content passed in.
//
// The Kind set is a hybrid (see workshop/plans/000095-weave-plan.md Core
// concepts): the file-op verbs ported from setup.sh's walk_manifest
// (Symlink|Seed|Scaffold|Touch|Merge|Tool — the dominant case in the live
// base.manifest) plus the new semantic verbs Prose|Skill that weave adds.
package intent

// Kind is the typed manifest verb. The ported file-op kinds lower
// near-identity to filesystem Actions; the semantic kinds (Prose, Skill)
// compose across layers (e.g. all layers' Prose → one AGENTS.md).
type Kind int

const (
	// Symlink — create a symlink from the target repo to the upstream path.
	Symlink Kind = iota
	// Seed — content-tracking real-file copy (NOT a symlink) into the target,
	// for first-run entrypoints that must work before substrate is present.
	Seed
	// Scaffold — create an empty directory structure in the target repo.
	Scaffold
	// Touch — create an empty file if missing (no source content needed).
	Touch
	// Merge — JSON settings cascade (settings.<layer>.json under
	// settings.local.json). Lowering deferred to M4 (the settings backend).
	Merge
	// Tool — declare the tool's owner as a substrate dependency / add a
	// `go mod edit -tool` directive. Lowering deferred (the one exec seam).
	Tool
	// Prose — a prose fragment (e.g. AGENTS.local.md) concatenated
	// foundation-first into the composed AGENTS.md. New in weave; replaces
	// the @AGENTS.local.md @-import.
	Prose
	// Skill — a skill directory aggregated into the agent-agnostic skill
	// index. New in weave; serving deferred to M3.
	Skill
)

// Visibility is the composition-algebra axis (workshop/targets/weave-composition-
// algebra.md): an artifact a layer declares is either Export (it flows down to
// every consumer) or Internal (it stays with the declaring repo). Compiling a
// repo R over the layer chain ⟨L₀…Lₙ=R⟩ selects 𝒜(R) = {every Lᵢ's exports} ⊎
// {Lₙ's internals only} — an ancestor's internal artifacts NEVER reach R. The
// type picks the compose operator; the visibility picks the operands.
type Visibility int

const (
	// Export is the DEFAULT (zero value): an artifact that flows down to every
	// consumer. A manifest row with no leading visibility token is an export, so
	// every pre-visibility row is unchanged.
	Export Visibility = iota
	// Internal — an artifact that stays with the declaring repo; it is selected
	// only on that repo's own self-walk (when it is the leaf Lₙ), never in a
	// derivative. (ariadne's AGENTS.local.md is internal: in a derivative it must
	// not leak — the #95 bug.)
	Internal
)

// Selected reports whether an artifact of visibility v, declared by a layer that
// is (isLeaf) or is not the leaf Lₙ, participates in the selected multiset 𝒜(R):
//
//	𝒜(R) = {every layer's exports} ⊎ {the leaf's internals only}
//
// i.e. an EXPORT from any layer, OR ANY visibility from the leaf. Equivalently,
// the ONLY thing excluded is an ANCESTOR's internal. This is the single source of
// truth for the visibility-axis selection (ARCH-DRY) — both the planner (which
// composes 𝒜(R)) and the completeness guard (which must not flag an ancestor's
// internal as under-produced, since it was deliberately excluded, not dropped)
// call it. See workshop/targets/weave-composition-algebra.md.
func Selected(v Visibility, isLeaf bool) bool {
	return v == Export || isLeaf
}

// Intent is one parsed manifest entry. Target defaults to Source when the
// manifest row omits the third column. Visibility defaults to Export (the zero
// value) when the row omits a leading export|internal token.
type Intent struct {
	Kind       Kind
	Visibility Visibility
	Source     string
	Target     string
}
