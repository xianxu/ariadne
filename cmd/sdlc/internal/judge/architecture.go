package judge

import (
	_ "embed"
	"fmt"
	"regexp"
)

// archMarkerRE matches an ARCH-<NAME> marker (e.g. ARCH-DRY, ARCH-SHIM). The
// name is [A-Z][A-Z-]* so it stops at the surrounding prose/punctuation.
var archMarkerRE = regexp.MustCompile(`ARCH-([A-Z][A-Z-]*)`)

// ArchitectureMarkers returns the ARCH-* marker names in registry order, deduped
// (e.g. ["ARCH-DRY", "ARCH-PURE"]). It is the single extraction site (ARCH-DRY):
// both the {{ARCH_STAR}} substitution in the code-review prompt and the
// AGENTS.md narrative-drift test consume it, so adding ARCH-SHIM (#71) flows into
// the review checklist and the drift guard with no other edits.
func ArchitectureMarkers() []string {
	var markers []string
	seen := map[string]bool{}
	for _, m := range archMarkerRE.FindAllStringSubmatch(ArchitectureRegistry, -1) {
		full := m[0]
		if seen[full] {
			continue
		}
		seen[full] = true
		markers = append(markers, full)
	}
	return markers
}

// ArchitectureRegistry is the embedded architecture.md — the single source of
// the ARCH-* architectural principles (#75). It is delivered verbatim into the
// planning, plan-quality, and code-review prompts; each prompt then directs the
// agent to the relevant lens (`at-plan` vs `at-review`). One file, embedded into
// each fresh context (a marker alone would be a dangling pointer in a
// fresh-context subagent — the definitions must be co-present).
//
//go:embed architecture.md
var ArchitectureRegistry string

// ArchitectureBlock renders the registry under a lens-specific header for
// embedding in a prompt (or delivering to the main thread via `sdlc start-plan`).
// lens is "at-plan" or "at-review" (advisory — the agent applies the named lens
// of each ARCH-* entry, all of which are in the block).
func ArchitectureBlock(lens string) string {
	return fmt.Sprintf("ARCHITECTURE PRINCIPLES — work through each of the %d "+
		"entries below explicitly, applying its `%s` lens; cite the marker "+
		"(e.g. ARCH-DRY) in any finding.\n\n%s",
		len(ArchitectureMarkers()), lens, ArchitectureRegistry)
}
