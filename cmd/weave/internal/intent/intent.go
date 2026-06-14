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

// Intent is one parsed manifest entry. Target defaults to Source when the
// manifest row omits the third column.
type Intent struct {
	Kind   Kind
	Source string
	Target string
}
