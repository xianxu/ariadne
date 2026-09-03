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

// deferredState is what the guard needs to know about architecture-deferred.md:
// which ARCH-* markers are documented-but-not-gated, and how many sections the
// file has. The markers are DERIVED with the same scanner the registry uses
// (markersIn), never hardcoded — that is what makes the deferred file's own
// claim true: "move a section into architecture.md to activate it — that is the
// whole activation step." A guard asserting the literal "ARCH-AUTHORITY" would
// go RED on activation, so activating an entry would mean editing the guard too,
// and the file's instruction would be a lie.
//
// Sections are counted independently of markers, and that separation is the
// whole point — see deferredVerdict.
type deferredState struct {
	Markers  []string
	Sections int
}

// parseDeferred reads the file's shape. PURE (takes content, not a path) so the
// three states below are table-testable without mutating the embedded file —
// which matters because the committed tree can only ever exhibit ONE of them.
func parseDeferred(content string) deferredState {
	d := deferredState{Markers: markersIn(content)}
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			d.Sections++
		}
	}
	return d
}

// deferredVerdict classifies the file, and is the SINGLE source of what each
// state means. An empty marker set makes the disjointness check vacuously true,
// so "empty" has to be explained rather than waved through — and there are two
// very different reasons for it.
//
// The verdict → action mapping, enforced in exactly one place
// (TestDeferredPrinciplesReachNoGate) and documented nowhere else:
//
//	guard   — entries present                  → ENFORCE disjointness
//	broken  — sections but no markers parsed   → FAIL (the guard is disarmed)
//	retired — no sections, no markers          → SKIP (every entry was activated)
//
// `retired` must SKIP and never fail. Failing it would mean that activating the
// last deferred entry reds the suite — which is exactly the "activation is a
// pure move" contract this whole design exists to keep. An earlier cut added a
// separate test asserting the committed file is always `guard`; that duplicated
// the `broken` failure and its only unique effect was to break the contract on
// `retired`, so it is deliberately absent. The skip is visible in test output,
// which is the right amount of noise for a state that should be rare and
// intentional.
type deferredVerdict string

const (
	deferredGuard   deferredVerdict = "guard"   // entries present — enforce
	deferredBroken  deferredVerdict = "broken"  // sections, no markers — headings stopped parsing
	deferredRetired deferredVerdict = "retired" // nothing left — every entry was activated
)

func (d deferredState) Verdict() deferredVerdict {
	switch {
	case len(d.Markers) > 0:
		return deferredGuard
	case d.Sections > 0:
		return deferredBroken
	default:
		return deferredRetired
	}
}

func deferredFileState(t *testing.T) deferredState {
	t.Helper()
	b, err := deferredFS.ReadFile("architecture-deferred.md")
	if err != nil {
		t.Fatalf("read architecture-deferred.md: %v", err)
	}
	return parseDeferred(string(b))
}

// TestDeferredVerdict covers all three classifications against synthetic content.
// The committed tree exhibits only `guard`, so without this the other two
// branches ship unexecuted — and they are the branches that decide whether an
// empty forbidden set is a legitimate retirement or a silently disarmed guard.
func TestDeferredVerdict(t *testing.T) {
	for _, tc := range []struct {
		name    string
		content string
		want    deferredVerdict
	}{
		{"one entry", "# Deferred\n\n## ARCH-AUTHORITY — x\n\n- **principle:** y\n", deferredGuard},
		{"several entries", "# D\n\n## ARCH-AUTHORITY — a\n\n## ARCH-LATER — b\n", deferredGuard},
		{"heading stopped parsing", "# D\n\n## ARCH_AUTHORITY — underscore, not a marker\n", deferredBroken},
		{"section with no marker at all", "# D\n\n## Some other heading\n\ntext\n", deferredBroken},
		{"all activated — prose only", "# Deferred\n\nNothing is deferred right now.\n", deferredRetired},
		{"empty file", "", deferredRetired},
		// A marker mentioned in prose still counts: the guard's job is to keep the
		// TOKEN out of gate text, so a token anywhere in this file is in scope.
		{"marker in prose only", "# D\n\nARCH-AUTHORITY is deferred.\n", deferredGuard},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseDeferred(tc.content).Verdict(); got != tc.want {
				t.Errorf("Verdict() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestDeferredPrinciplesReachNoGate is the guard that makes "documented but not
// gated" a CHECKED property rather than a convention.
//
// Both halves are derived, because both are the kind of list that goes stale:
// the forbidden markers come from architecture-deferred.md, and the gate-facing
// text is every prompt walked out of promptFS plus the three render helpers. A
// prompt added next year is covered without anyone remembering this issue.
func TestDeferredPrinciplesReachNoGate(t *testing.T) {
	state := deferredFileState(t)
	deferred := state.Markers

	// An empty marker set makes every assertion below vacuously true, so it is
	// classified rather than waved through. Renaming or deleting the file is
	// already a BUILD error (//go:embed on a missing path), which leaves one
	// real disarm vector: the entries are still there but stopped parsing.
	switch state.Verdict() {
	case deferredBroken:
		t.Fatalf("architecture-deferred.md has %d section(s) but parses no ARCH-* markers — "+
			"the guard below would pass vacuously. A heading probably stopped matching "+
			"%s; fix the heading, don't delete the check.", state.Sections, archMarkerRE)
	case deferredRetired:
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
