package judge

import (
	"embed"
	"io/fs"
	"strings"
	"testing"
)

// deferredFS holds architecture-deferred.md — principles written down but
// deliberately NOT gated. It is a SEPARATE embed from ArchitectureRegistry on
// purpose: architecture.md is embedded whole-file and archMarkerRE scans that
// same string, so anything carrying an ARCH- marker in it is counted into the
// block header and injected into every prompt. A `gates:` field would need
// entry-level parsing and filtering for one deferred entry; a second file costs
// nothing and makes activation a cut-and-paste.
//
//go:embed architecture-deferred.md
var deferredFS embed.FS

// deferredMarkers returns the ARCH-* markers that are documented but must NOT be
// gated — DERIVED by scanning architecture-deferred.md with the same regex the
// registry uses, never hardcoded.
//
// Deriving is what makes the deferred file's own claim true: "move a section
// into architecture.md to activate it — that is the whole activation step." A
// guard asserting the literal "ARCH-AUTHORITY" would go RED on activation, so
// activating an entry would mean editing the guard too, and the file's
// instruction would be a lie. Because the set is derived, moving the section
// empties it and the guard stays green with no other edit.
func deferredMarkers(t *testing.T) (markers []string, sections int) {
	t.Helper()
	b, err := deferredFS.ReadFile("architecture-deferred.md")
	if err != nil {
		t.Fatalf("read architecture-deferred.md: %v", err)
	}
	seen := map[string]bool{}
	for _, m := range archMarkerRE.FindAllStringSubmatch(string(b), -1) {
		if !seen[m[0]] {
			seen[m[0]] = true
			markers = append(markers, m[0])
		}
	}
	// Sections are counted independently of markers, and that separation is the
	// whole point: it tells "every entry has been activated" (no sections, no
	// markers — legitimate) apart from "the entries are still here but their
	// headings stopped parsing" (sections, no markers — broken guard).
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "## ") {
			sections++
		}
	}
	return markers, sections
}

// TestDeferredPrinciplesReachNoGate is the guard that makes "documented but not
// gated" a CHECKED property rather than a convention.
//
// Both halves are derived, because both are the kind of list that goes stale:
// the forbidden markers come from architecture-deferred.md, and the gate-facing
// text is every prompt walked out of promptFS plus the three render helpers. A
// prompt added next year is covered without anyone remembering this issue.
func TestDeferredPrinciplesReachNoGate(t *testing.T) {
	deferred, sections := deferredMarkers(t)

	// A deferred file that parses no markers makes every assertion below
	// vacuously true, so the empty case has to be classified rather than
	// waved through. Renaming or deleting the file is already a BUILD error
	// (//go:embed on a missing path), which leaves one real disarm vector: the
	// entries are still in the file but their headings stopped parsing.
	switch {
	case len(deferred) == 0 && sections > 0:
		t.Fatalf("architecture-deferred.md has %d section(s) but parses no ARCH-* markers — "+
			"the guard below would pass vacuously. A heading probably stopped matching "+
			"%s; fix the heading, don't delete the check.", sections, archMarkerRE)
	case len(deferred) == 0:
		// Every entry has been activated. Nothing is deferred, so there is
		// nothing to keep out of the gates — and activation stayed a pure MOVE,
		// which is exactly what the deferred file promises. Retire this test
		// along with the file when that state is deliberate and permanent.
		t.Skip("no deferred principles remain — all have been activated")
	}
	// A marker cannot be both deferred and gated. This is the activation
	// invariant read from the other side.
	for _, m := range deferred {
		for _, gated := range ArchitectureMarkers() {
			if m == gated {
				t.Errorf("%s is in BOTH architecture-deferred.md and architecture.md. "+
					"Activation is a MOVE: delete the section from the deferred file.", m)
			}
		}
	}

	for _, g := range gateFacingTexts(t) {
		for _, m := range deferred {
			if strings.Contains(g.text, m) {
				t.Errorf("%s reaches %s — a deferred principle must not be gated. "+
					"Either it was activated (move it out of architecture-deferred.md) "+
					"or something now embeds the deferred file.", m, g.name)
			}
		}
	}
}

type gateText struct{ name, text string }

// gateFacingTexts enumerates every string this package can put in front of a
// judge or the operator, derived rather than listed: every prompts/*.md rendered
// through the real substitution path, both lenses of the architecture block, the
// code-review body, and the raw registry.
//
// Walking promptFS is the load-bearing part. A hand-written list of prompt names
// is the restatement that silently stops covering the prompt someone adds later.
func gateFacingTexts(t *testing.T) []gateText {
	t.Helper()
	out := []gateText{
		{"ArchitectureBlock(at-plan)", ArchitectureBlock("at-plan")},
		{"ArchitectureBlock(at-review)", ArchitectureBlock("at-review")},
		{"ArchitectureRegistry", ArchitectureRegistry},
		{"CodeReviewBody", CodeReviewBody(PromptInput{})},
	}
	entries, err := fs.ReadDir(promptFS, "prompts")
	if err != nil {
		t.Fatalf("read prompts dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("promptFS walked to zero prompts — the guard would cover nothing")
	}
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".md")
		// Render through the production path so a marker reaching a prompt via
		// ANY substitution token is caught, not just via {{ARCH_BLOCK}}.
		out = append(out, gateText{
			name: "prompts/" + e.Name(),
			text: BuildPrompt(Category(name), PromptInput{}),
		})
	}
	return out
}
