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
	b.WriteString("```findings\ndispose:\n  - id: <a prior finding's id>\n    disposition: <" +
		strings.Join(m.AllDispositions(), " | ") + ">\n    note: <one line, optional>\nfindings:\n" +
		"  - id: new\n    severity: <" + strings.Join(m.Severities(), " | ") + ">\n" +
		"    title: <one line>\n    detail: <a sentence or two, optional>\n```\n\n")

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
