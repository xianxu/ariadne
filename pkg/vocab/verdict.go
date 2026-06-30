package vocab

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate sh -c "vocabulary export --noun verdict > verdict.json"

//go:embed verdict.json
var verdictJSON []byte

// VerdictModel is the read-only, parsed `verdict` noun: the boundary-review verdict
// tokens by semantic category + per-token semantics. Derived from
// construct/vocabulary/verdict.cue at generate time; never hand-edited (ariadne#147).
// The single Go read of the verdict vocabulary — the prompt's emitted set, the
// parser's accepted set, and the close finalize-policy (#139) all derive from here.
type VerdictModel struct {
	Categories map[string][]string `json:"categories"` // "finalizing"/"blocking"/"internal" → tokens
	When       map[string]string   `json:"when"`       // token → one-line semantics
}

var verdictModel = mustLoadVerdict()

func mustLoadVerdict() *VerdictModel {
	var m VerdictModel
	if err := json.Unmarshal(verdictJSON, &m); err != nil {
		panic(fmt.Sprintf("vocab: corrupt embedded verdict.json (run `make vocab-embed`): %v", err))
	}
	return &m
}

// Verdict returns the embedded `verdict` model.
func Verdict() *VerdictModel { return verdictModel }

func (m *VerdictModel) inCategory(cat, t string) bool {
	for _, v := range m.Categories[cat] {
		if v == t {
			return true
		}
	}
	return false
}

// IsEmitted reports whether t is a token a reviewer may emit (finalizing or
// blocking) — the set the structured verdict block is validated against.
func (m *VerdictModel) IsEmitted(t string) bool {
	return m.inCategory("finalizing", t) || m.inCategory("blocking", t)
}

// IsFinalizing reports whether t is an acceptable verdict the gate may finalize on
// (SHIP / FIX-THEN-SHIP).
func (m *VerdictModel) IsFinalizing(t string) bool { return m.inCategory("finalizing", t) }

// IsBlocking reports whether t blocks the boundary — rework + re-run (REWORK).
func (m *VerdictModel) IsBlocking(t string) bool { return m.inCategory("blocking", t) }

// Emitted returns the reviewer-emittable tokens (finalizing then blocking) — the
// set the prompt renders and the parser validates against.
func (m *VerdictModel) Emitted() []string {
	out := make([]string, 0, len(m.Categories["finalizing"])+len(m.Categories["blocking"]))
	out = append(out, m.Categories["finalizing"]...)
	out = append(out, m.Categories["blocking"]...)
	return out
}
