package vocab

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:generate sh -c "vocabulary export --noun finding > finding.json"

//go:embed finding.json
var findingJSON []byte

// FindingModel is the read-only, parsed `finding` noun: severities by blocking/advisory
// category, the dispositions by closing/open category, and per-token semantics. Derived
// from construct/vocabulary/finding.cue at generate time; never hand-edited (ariadne#187).
//
// The single Go read of the finding vocabulary — the plan-quality prompt's emitted set,
// the findings-block parser's accepted set, and the gate decision's blocking test all
// derive from here, so none of them can drift from the others.
type FindingModel struct {
	Categories   map[string][]string `json:"categories"`   // "blocking"/"advisory" → severities
	HardBlocking []string            `json:"hardBlocking"` // still blocks past the round cap
	Dispositions map[string][]string `json:"dispositions"` // "closing"/"open" → dispositions
	When         map[string]string   `json:"when"`         // severity → one-line semantics
	WhenDisposed map[string]string   `json:"whenDisposed"` // disposition → one-line semantics
}

var findingModel = mustLoadFinding()

func mustLoadFinding() *FindingModel {
	var m FindingModel
	if err := json.Unmarshal(findingJSON, &m); err != nil {
		panic(fmt.Sprintf("vocab: corrupt embedded finding.json (run `make vocab-embed`): %v", err))
	}
	return &m
}

// Finding returns the embedded `finding` model.
func Finding() *FindingModel { return findingModel }

// findingCategoryOrder is the finding noun's severity ordering: blocking first, then
// advisory — the order the prompt renders and the ledger groups by.
var findingCategoryOrder = []string{"blocking", "advisory"}

// Severities returns every severity, blocking first then advisory.
func (m *FindingModel) Severities() []string {
	return allStatuses(m.Categories, findingCategoryOrder)
}

// IsSeverity reports whether s is a modeled severity — the parser's accepted set.
func (m *FindingModel) IsSeverity(s string) bool { return contains(m.Severities(), s) }

// Blocks reports whether an UNDISPOSED finding at severity s refuses the gate.
func (m *FindingModel) Blocks(s string) bool { return inCat(m.Categories, "blocking", s) }

// BlocksPastCap reports whether s still refuses the gate once the round cap is reached.
// The rest of the blocking set is demoted there — recorded and reported, not blocking.
func (m *FindingModel) BlocksPastCap(s string) bool { return contains(m.HardBlocking, s) }

// dispositionCategoryOrder mirrors findingCategoryOrder for the disposition partition.
var dispositionCategoryOrder = []string{"closing", "open"}

// AllDispositions returns every modeled disposition, closing first then open.
func (m *FindingModel) AllDispositions() []string {
	return allStatuses(m.Dispositions, dispositionCategoryOrder)
}

// IsDisposition reports whether d is a modeled disposition.
func (m *FindingModel) IsDisposition(d string) bool { return contains(m.AllDispositions(), d) }

// Closes reports whether disposition d SETTLES a finding, i.e. stops it blocking. Derived
// from the model's closing/open partition, so a disposition added to finding.cue can never
// reach the open-set computation as an unhandled case that silently wedges a finding open.
func (m *FindingModel) Closes(d string) bool { return inCat(m.Dispositions, "closing", d) }

// RenderBlockInstruction renders the structured findings-handoff instruction for the
// plan-quality prompt — the fenced ```findings block template plus the per-severity and
// per-disposition gloss — entirely from the model, so the prompt's accepted set never
// drifts from finding.cue (ariadne#187, the agent-binary-handoff-schema target).
//
// Mirrors VerdictModel.RenderBlockInstruction: the model renders its own handoff
// instruction, so the prompt layer holds no severity or disposition names at all.
func (m *FindingModel) RenderBlockInstruction() string {
	var b strings.Builder
	b.WriteString("Emit your findings as this fenced block — the machine-read handoff the\n")
	b.WriteString("binary parses. `dispose:` first (every prior finding), then `findings:`\n")
	b.WriteString("(anything newly raised). Use `id: new` for a new finding — the binary\n")
	b.WriteString("assigns the stable id. Omit a key entirely when it has no entries.\n\n")
	// title/detail/note are shown as BLOCK SCALARS (`|`), not plain scalars. This is not
	// cosmetic: in a YAML plain scalar a ` #` begins a comment, so a finding whose text
	// contains "issue #187" or "## Estimate" is SILENTLY TRUNCATED at that point — and the
	// truncated text is what the next round is shown as its own prior finding. Observed
	// live on ariadne#187 round 1, where a detail lost its second half. Block scalars are
	// immune to `#` and `:` alike.
	b.WriteString("```findings\ndispose:\n  - id: <a prior finding's id>\n    disposition: <" +
		strings.Join(m.AllDispositions(), " | ") + ">\n    note: |\n      <optional, one line>\nfindings:\n" +
		"  - id: new\n    severity: <" + strings.Join(m.Severities(), " | ") + ">\n" +
		"    family: <slug>\n" +
		"    title: |\n      <one line>\n    detail: |\n      <a sentence or two, optional>\n```\n\n")
	// family (#194 M3): the slug names the underlying RULE, not the symptom, and is what
	// lets the gate say "this is the 3rd instance — state the rule" instead of watching
	// the same missing rule get patched case by case. The instruction has to be explicit:
	// a reviewer that omits it, or coins a synonym, silently defeats the escalation.
	b.WriteString("`family` is a short slug naming the underlying RULE a finding is an instance of,\n")
	b.WriteString("not its symptom — `block-opener-rule`, not `bracket-depth-bug`. If the prior-round\n")
	b.WriteString("block above lists families already in play, REUSE the matching slug verbatim;\n")
	b.WriteString("coin a new one only when the finding genuinely belongs to no existing family.\n\n")
	// Observed on ariadne#194 itself: a finding slugged for its symptom did not match the
	// next instance of the same rule, so the escalation silently failed to fire — the
	// mechanism looked implemented and did nothing. Naming this failure mode in the
	// instruction is the only place it can be caught, since the binary cannot tell a
	// symptom-slug from a rule-slug.
	b.WriteString("Slug the RULE, because a symptom-slug will not match the next instance and the\n")
	b.WriteString("escalation will silently fail to fire. Ask: \"if this recurs in a different file\n")
	b.WriteString("with different symptoms, would I still reach for this slug?\" If not, it names the\n")
	b.WriteString("symptom. `boundary-scope-strands-reads` survives that test; `family-counts-filtered`\n")
	b.WriteString("does not — it describes one site, and the same rule broke a second read elsewhere.\n\n")
	b.WriteString("Use the `|` block form for title, detail and note exactly as shown, and indent\n")
	b.WriteString("their text by six spaces. In plain YAML a ` #` starts a comment, so an\n")
	b.WriteString("unquoted `## Estimate` or `issue #187` would silently truncate your finding.\n\n")

	width := 0
	for _, t := range append(append([]string{}, m.Severities()...), m.AllDispositions()...) {
		if len(t) > width {
			width = len(t)
		}
	}
	for _, s := range m.Severities() {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, s, m.When[s])
	}
	b.WriteString("\n")
	for _, d := range m.AllDispositions() {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, d, m.WhenDisposed[d])
	}
	return strings.TrimRight(b.String(), "\n")
}
