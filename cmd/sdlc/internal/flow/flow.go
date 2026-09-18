// Package flow owns the issue's SDLC flow record (#231): which gate set an issue
// runs — the full flow, or the quick flow for small diffs — and who decided.
//
// The record lives in issue frontmatter on ONE line, because the issue
// frontmatter helpers (internal/issue.GetField/SetField) are line-based:
//
//	flow: {kind: quick, provenance: inferred, spec: "1a2b3c4d", done: "5e6f7a8b"}
//
// Everything here is PURE (ARCH-PURE): change-code, start-plan and close are the
// IO shells that read and write the record; the decisions are tested on strings.
package flow

import (
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/issue"
)

// Kind is which gate set the issue runs.
type Kind string

const (
	// Full is today's flow: durable plan, plan-quality, the estimate gates, and
	// the full boundary review. An issue with no record reads as Full.
	Full Kind = "full"
	// Quick skips the plan, plan-quality and estimate gates and keeps one close
	// gate with the small-diff review, inside the hard shell (limits.go).
	Quick Kind = "quick"
)

// Provenance is who decided the kind.
type Provenance string

const (
	// Inferred: a gate decided from the issue's shape or its diff.
	Inferred Provenance = "inferred"
	// Operator: the operator pinned it (`sdlc change-code --flow`). Gates never
	// re-infer an operator pin, except through the hard shell.
	Operator Provenance = "operator"
)

// Field is the frontmatter key the record lives under.
const Field = "flow"

// Flow is one issue's flow record. Its fields are UNEXPORTED, so outside this
// package a Flow can only come from Parse/FromFrontmatter/Recorded (reading a
// record) or Decide/WithContract (making a decision): the compiler, not a source
// scan, enforces that no caller decides a flow on its own (#231 BR-14, BR-16).
// The zero Flow reads as Full, the stricter flow.
type Flow struct {
	kind       Kind
	provenance Provenance
	spec, done string // the contract hashes on the quick flow (see ContractHashes)
}

// Kind is the gate set; the zero Flow reads as Full.
func (f Flow) Kind() Kind {
	if f.kind == "" {
		return Full
	}
	return f.kind
}

// Provenance is who decided the kind; empty for an unrecorded (zero) Flow.
func (f Flow) Provenance() Provenance { return f.provenance }

// Spec is the Spec+Revisions contract hash recorded at change-code, or "".
func (f Flow) Spec() string { return f.spec }

// Done is the Done-when contract hash recorded at change-code, or "".
func (f Flow) Done() string { return f.done }

// hashRE is the shape of a contract hash: the 8-hex prefix of a sha256.
var hashRE = regexp.MustCompile(`^[0-9a-f]{8}$`)

// Reason names which of Parse's reject branches refused a record — one per
// branch, so the shared corpus can prove every branch is pinned by at least one
// row (#231 BR-17): a branch the corpus never exercises is a branch the two
// readers could silently disagree on.
type Reason string

const (
	ReasonYAML          Reason = "yaml"           // not parseable as YAML at all
	ReasonNotMap        Reason = "not-map"        // parses, but not a one-line map
	ReasonDuplicateKey  Reason = "duplicate-key"  // a key repeated
	ReasonNotString     Reason = "not-string"     // a value YAML types as other than a string
	ReasonUnknownKey    Reason = "unknown-key"    // a key the record does not have
	ReasonBadKind       Reason = "bad-kind"       // kind missing or outside its set
	ReasonBadProvenance Reason = "bad-provenance" // provenance missing or outside its set
	ReasonBadHash       Reason = "bad-hash"       // a present hash that is not 8 lowercase hex
)

// Reasons is every reject branch, for the corpus coverage check.
var Reasons = []Reason{ReasonYAML, ReasonNotMap, ReasonDuplicateKey, ReasonNotString,
	ReasonUnknownKey, ReasonBadKind, ReasonBadProvenance, ReasonBadHash}

// ParseError is Parse's refusal, tagged with the branch that refused.
type ParseError struct {
	Reason Reason
	msg    string
}

func (e *ParseError) Error() string { return e.msg }

func reject(r Reason, format string, a ...any) error {
	return &ParseError{Reason: r, msg: fmt.Sprintf(format, a...)}
}

// Format renders the one-line record. The hashes are always QUOTED: an
// unquoted all-digit hash is an int to both cue and yaml.v3, so the same bytes
// would not parse as the string the record holds (#231 PQ-9). (Measured: other
// unquoted hex shapes, `1a2b3c4d` and even `12e45678`, read as strings under
// both readers — the shared corpus in construct/vocabulary/testdata pins that.)
func Format(f Flow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "{kind: %s, provenance: %s", f.Kind(), f.provenance)
	if f.spec != "" {
		fmt.Fprintf(&b, ", spec: %q", f.spec)
	}
	if f.done != "" {
		fmt.Fprintf(&b, ", done: %q", f.done)
	}
	b.WriteString("}")
	return b.String()
}

// Parse reads a record value (the text after `flow:`). Frontmatter is
// hand-editable, so this is a trust boundary (ARCH-SECURE): every malformed
// shape is an error — a non-map, an unknown or duplicate key, a value outside
// its closed set, or a hash that is not a quoted 8-hex string. Callers resolve
// an error to Full, the stricter flow; nothing here defaults silently.
func Parse(value string) (Flow, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(value), &doc); err != nil {
		return Flow{}, reject(ReasonYAML, "flow record %q does not parse: %v", value, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return Flow{}, reject(ReasonNotMap, "flow record %q is not a one-line map like %s", value, Format(Flow{kind: Quick, provenance: Inferred}))
	}
	var f Flow
	m := doc.Content[0]
	seen := map[string]bool{}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i].Value, m.Content[i+1]
		if seen[k] {
			return Flow{}, reject(ReasonDuplicateKey, "flow record %q repeats %q", value, k)
		}
		seen[k] = true
		if v.Kind != yaml.ScalarNode || v.ShortTag() != "!!str" {
			return Flow{}, reject(ReasonNotString, "flow record %q: %s must be a string (quote it)", value, k)
		}
		switch k {
		case "kind":
			f.kind = Kind(v.Value)
		case "provenance":
			f.provenance = Provenance(v.Value)
		case "spec":
			f.spec = v.Value
		case "done":
			f.done = v.Value
		default:
			return Flow{}, reject(ReasonUnknownKey, "flow record %q has unknown key %q", value, k)
		}
	}
	if f.kind != Full && f.kind != Quick {
		return Flow{}, reject(ReasonBadKind, "flow record %q: kind must be %s or %s", value, Full, Quick)
	}
	if f.provenance != Inferred && f.provenance != Operator {
		return Flow{}, reject(ReasonBadProvenance, "flow record %q: provenance must be %s or %s", value, Inferred, Operator)
	}
	// A hash key that is PRESENT must hold a hash — `spec: ""` is rejected, as the
	// cue model rejects it; absent and empty are not the same record (#231 BR-12).
	for key, h := range map[string]string{"spec": f.spec, "done": f.done} {
		if seen[key] && !hashRE.MatchString(h) {
			return Flow{}, reject(ReasonBadHash, "flow record %q: contract hash %s %q is not 8 lowercase hex", value, key, h)
		}
	}
	return f, nil
}

// FromFrontmatter reads the record from an issue's frontmatter. With no
// `flow:` line it returns Full and recorded=false: every issue filed before
// #231 runs the full flow. A present but malformed line returns recorded=true
// and the parse error, with Full as the value, so a caller that only warns
// still lands on the stricter flow.
func FromFrontmatter(fm string) (f Flow, recorded bool, err error) {
	v, ok := issue.GetField(fm, Field)
	if !ok {
		return Flow{kind: Full}, false, nil
	}
	f, err = Parse(v)
	if err != nil {
		return Flow{kind: Full, provenance: Inferred}, true, err
	}
	return f, true, nil
}

// Recorded reads the record as a decision input: nil when there is none, the
// record when it is valid, and full/inferred when it is malformed — the stricter
// flow, which Decide then never downgrades — together with the parse error for
// the caller to report. The one place a malformed record is resolved.
func Recorded(fm string) (*Flow, error) {
	f, recorded, err := FromFrontmatter(fm)
	if !recorded {
		return nil, nil
	}
	return &f, err
}

// DecideInput is what change-code knows when it infers the flow. Recorded comes
// from Recorded.
type DecideInput struct {
	Recorded      *Flow  // the record already on the issue, nil if none
	Pin           string // --flow value: "", "quick" or "full"
	HasMilestones bool   // the Plan has Mx rows
	HasPlan       bool   // a durable plan exists (the file plan-quality would judge)
}

// Rule names which of Decide's rules fired — returned rather than re-derived by
// callers, so the reason a flow is reported with cannot drift from the order the
// rules are actually checked in. Opaque, with unexported values: outside this
// package a Rule can only come from Decide (#231 BR-16).
type Rule struct{ text string }

func (r Rule) String() string { return r.text }

var (
	rulePinned         = Rule{"pinned with --flow"}
	ruleMilestones     = Rule{"the Plan has Mx milestones"}
	ruleOperatorStands = Rule{"the operator's pin stands"}
	ruleNoDowngrade    = Rule{"already full — gates never downgrade"}
	rulePlan           = Rule{"a durable plan exists"}
	ruleNoPlan         = Rule{"no Mx milestones and no durable plan"}
)

// Decide computes the flow change-code records, and the rule that decided it.
// The rules, in order:
//
//  1. An unknown pin, or a quick pin on a Plan with Mx rows, is an error: a pin
//     that cannot be honoured says so while the operator is there.
//  2. A pin sets {pin, operator}.
//  3. Mx rows cross the shell, so they make the flow {full, inferred} whatever
//     was recorded — unless it is already full, which stays as it is.
//  4. A recorded operator flow stands; a recorded full stands (no downgrade).
//  5. Otherwise infer: a durable plan means full, its absence means quick.
//
// The invariant (ARCH-ORDER): only an operator pin produces quick from full, and
// nothing produces quick while Mx rows exist. The contract hashes are not
// Decide's business; change-code adds them with WithContract.
func Decide(in DecideInput) (Flow, Rule, error) {
	switch in.Pin {
	case "":
	case string(Quick):
		if in.HasMilestones {
			return Flow{}, Rule{}, fmt.Errorf("--flow quick: the Plan has Mx milestone rows, and the quick flow has a single boundary — drop the milestones or keep the full flow")
		}
		return Flow{kind: Quick, provenance: Operator}, rulePinned, nil
	case string(Full):
		return Flow{kind: Full, provenance: Operator}, rulePinned, nil
	default:
		return Flow{}, Rule{}, fmt.Errorf("--flow %q: want %s or %s", in.Pin, Quick, Full)
	}
	r := in.Recorded
	if in.HasMilestones && (r == nil || r.Kind() != Full) {
		return Flow{kind: Full, provenance: Inferred}, ruleMilestones, nil
	}
	if r != nil && r.provenance == Operator {
		return Flow{kind: r.Kind(), provenance: r.provenance}, ruleOperatorStands, nil
	}
	if r != nil && r.Kind() == Full {
		return Flow{kind: r.Kind(), provenance: r.provenance}, ruleNoDowngrade, nil
	}
	if in.HasPlan {
		return Flow{kind: Full, provenance: Inferred}, rulePlan, nil
	}
	return Flow{kind: Quick, provenance: Inferred}, ruleNoPlan, nil
}
