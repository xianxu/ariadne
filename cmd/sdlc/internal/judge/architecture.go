package judge

import _ "embed"

// ArchitectureRegistry is the embedded architecture.md — the single source of
// the ARCH-* architectural principles (#75). It is delivered verbatim into the
// planning, plan-quality, and code-review prompts; each prompt then directs the
// agent to the relevant lens (`at-plan` vs `at-review`). One file, embedded into
// each fresh context (a marker alone would be a dangling pointer in a
// fresh-context subagent — the definitions must be co-present).
//
//go:embed architecture.md
var ArchitectureRegistry string

// architectureBlock renders the registry under a lens-specific header for
// embedding in a prompt. lens is "at-plan" or "at-review" (advisory — the agent
// applies the named lens of each ARCH-* entry, all of which are in the block).
func architectureBlock(lens string) string {
	return "ARCHITECTURE PRINCIPLES — apply each ARCH-* entry's `" + lens +
		"` lens; cite the marker (e.g. ARCH-DRY) in any finding.\n\n" +
		ArchitectureRegistry
}
