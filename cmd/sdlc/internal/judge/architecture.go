package judge

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// archMarkerRE matches an ARCH-<NAME> marker (e.g. ARCH-DRY, ARCH-SHIM). The
// name is [A-Z][A-Z-]* so it stops at the surrounding prose/punctuation.
var archMarkerRE = regexp.MustCompile(`ARCH-([A-Z][A-Z-]*)`)

// ArchitectureMarkers returns the ARCH-* marker names in registry order, deduped
// (e.g. ["ARCH-DRY", "ARCH-PURE"]). It is the single extraction site (ARCH-DRY):
// both the {{ARCH_STAR}} substitution in the code-review prompt and the
// AGENTS.md narrative-drift test consume it, so adding ARCH-SHIM (#71) flows into
// the review checklist and the drift guard with no other edits.
func ArchitectureMarkers() []string { return markersIn(ArchitectureRegistry) }

// markersIn extracts ARCH-* markers from any text, in order, deduped. Pure.
//
// Split out of ArchitectureMarkers in #208 because the deferred-principle guard
// needs the identical extraction over architecture-deferred.md: two copies of
// "scan, dedupe, keep order" would let the gated and the not-gated sets be
// computed by subtly different rules, which is the one thing that must not
// happen when the whole guard is "these two sets are disjoint" (ARCH-DRY).
func markersIn(text string) []string {
	var markers []string
	seen := map[string]bool{}
	for _, m := range archMarkerRE.FindAllStringSubmatch(text, -1) {
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
// of each ARCH-* entry, all of which are in the block). It is ArchitectureBlockFor
// over every marker, which is the registry verbatim.
func ArchitectureBlock(lens string) string {
	return ArchitectureBlockFor(lens, ArchitectureMarkers())
}

// ArchitectureBlockFor renders the block for a subset of the principles: the
// header counting them, the registry's preamble, and the selected entries in
// registry order, each verbatim. Over every marker it is the whole registry byte
// for byte (the split is lossless), so the full-flow prompts are unchanged.
func ArchitectureBlockFor(lens string, markers []string) string {
	pre, secs := mustArchitectureSections()
	want := map[string]bool{}
	for _, m := range markers {
		want[m] = true
	}
	var b strings.Builder
	b.WriteString(pre)
	n := 0
	for _, sec := range secs {
		if want[sec.marker] {
			b.WriteString(sec.text)
			n++
		}
	}
	return fmt.Sprintf("ARCHITECTURE PRINCIPLES — work through each of the %d "+
		"entries below explicitly, applying its `%s` lens; cite the marker "+
		"(e.g. ARCH-DRY) in any finding.\n\n%s", n, lens, b.String())
}

// QuickMarkers returns, in registry order, the principles whose entry declares
// `quick-flow: yes` — the ones the quick flow's small-diff review checks (#231).
// This is where that set is chosen: flip an entry's field in architecture.md.
func QuickMarkers() []string {
	_, secs := mustArchitectureSections()
	var out []string
	for _, sec := range secs {
		if sec.quick {
			out = append(out, sec.marker)
		}
	}
	return out
}

// archSection is one `## ARCH-*` entry of the registry, verbatim.
type archSection struct {
	marker string
	text   string
	quick  bool
}

var quickFieldRE = regexp.MustCompile(`(?m)^- \*\*quick-flow:\*\* *(.*?) *$`)

// architectureSections splits a registry into its preamble and entries,
// losslessly: preamble + every entry's text re-joins to the input. Each entry
// must declare `- **quick-flow:** yes|no` exactly once, so a principle cannot be
// added without deciding whether the quick flow checks it. Pure. (Promoted from
// the architectureEntry test helper, which slices one entry the same way.)
func architectureSections(registry string) (string, []archSection, error) {
	starts := regexp.MustCompile(`(?m)^## ARCH-`).FindAllStringIndex(registry, -1)
	if len(starts) == 0 {
		return "", nil, fmt.Errorf("architecture registry has no `## ARCH-*` entries")
	}
	pre := registry[:starts[0][0]]
	var secs []archSection
	for i, st := range starts {
		end := len(registry)
		if i+1 < len(starts) {
			end = starts[i+1][0]
		}
		text := registry[st[0]:end]
		marker := archMarkerRE.FindString(text)
		fields := quickFieldRE.FindAllStringSubmatch(text, -1)
		if len(fields) != 1 {
			return "", nil, fmt.Errorf("%s declares quick-flow %d times, want exactly once", marker, len(fields))
		}
		var quick bool
		switch fields[0][1] {
		case "yes":
			quick = true
		case "no":
		default:
			return "", nil, fmt.Errorf("%s: quick-flow is %q, want yes or no", marker, fields[0][1])
		}
		secs = append(secs, archSection{marker: marker, text: text, quick: quick})
	}
	return pre, secs, nil
}

var (
	archOnce     sync.Once
	archPreamble string
	archSecs     []archSection
)

// mustArchitectureSections parses the embedded registry once. A malformed
// registry is a build-time bug, like a missing prompt template — it panics, and
// TestEveryPrincipleDeclaresQuickFlow keeps the embedded one well-formed.
func mustArchitectureSections() (string, []archSection) {
	archOnce.Do(func() {
		pre, secs, err := architectureSections(ArchitectureRegistry)
		if err != nil {
			panic("judge: " + err.Error())
		}
		archPreamble, archSecs = pre, secs
	})
	return archPreamble, archSecs
}
