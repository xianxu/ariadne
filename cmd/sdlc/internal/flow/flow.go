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

// Flow is one issue's flow record. Spec and Done are the contract hashes
// change-code records on the quick flow (see ContractHashes); empty otherwise.
type Flow struct {
	Kind       Kind
	Provenance Provenance
	Spec       string
	Done       string
}

// hashRE is the shape of a contract hash: the 8-hex prefix of a sha256.
var hashRE = regexp.MustCompile(`^[0-9a-f]{8}$`)

// Format renders the one-line record. The hashes are always QUOTED: an
// unquoted all-digit hash is an int to both cue and yaml.v3, and a hash like
// `12e45678` is a float, so the same bytes would parse to different types under
// different readers (#231 PQ-9).
func Format(f Flow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "{kind: %s, provenance: %s", f.Kind, f.Provenance)
	if f.Spec != "" {
		fmt.Fprintf(&b, ", spec: %q", f.Spec)
	}
	if f.Done != "" {
		fmt.Fprintf(&b, ", done: %q", f.Done)
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
		return Flow{}, fmt.Errorf("flow record %q does not parse: %v", value, err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) != 1 || doc.Content[0].Kind != yaml.MappingNode {
		return Flow{}, fmt.Errorf("flow record %q is not a one-line map like %s", value, Format(Flow{Kind: Quick, Provenance: Inferred}))
	}
	var f Flow
	m := doc.Content[0]
	seen := map[string]bool{}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k, v := m.Content[i].Value, m.Content[i+1]
		if seen[k] {
			return Flow{}, fmt.Errorf("flow record %q repeats %q", value, k)
		}
		seen[k] = true
		if v.Kind != yaml.ScalarNode || v.ShortTag() != "!!str" {
			return Flow{}, fmt.Errorf("flow record %q: %s must be a string (quote it)", value, k)
		}
		switch k {
		case "kind":
			f.Kind = Kind(v.Value)
		case "provenance":
			f.Provenance = Provenance(v.Value)
		case "spec":
			f.Spec = v.Value
		case "done":
			f.Done = v.Value
		default:
			return Flow{}, fmt.Errorf("flow record %q has unknown key %q", value, k)
		}
	}
	if f.Kind != Full && f.Kind != Quick {
		return Flow{}, fmt.Errorf("flow record %q: kind must be %s or %s", value, Full, Quick)
	}
	if f.Provenance != Inferred && f.Provenance != Operator {
		return Flow{}, fmt.Errorf("flow record %q: provenance must be %s or %s", value, Inferred, Operator)
	}
	for _, h := range []string{f.Spec, f.Done} {
		if h != "" && !hashRE.MatchString(h) {
			return Flow{}, fmt.Errorf("flow record %q: contract hash %q is not 8 lowercase hex", value, h)
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
		return Flow{Kind: Full}, false, nil
	}
	f, err = Parse(v)
	if err != nil {
		return Flow{Kind: Full}, true, err
	}
	return f, true, nil
}

// DecideInput is what change-code knows when it infers the flow.
type DecideInput struct {
	Recorded      *Flow  // the record already on the issue, nil if none
	Pin           string // --flow value: "", "quick" or "full"
	HasMilestones bool   // the Plan has Mx rows
	HasPlan       bool   // a durable plan exists (the file plan-quality would judge)
}

// Decide computes the flow change-code records. The rules, in order:
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
func Decide(in DecideInput) (Flow, error) {
	switch in.Pin {
	case "":
	case string(Quick):
		if in.HasMilestones {
			return Flow{}, fmt.Errorf("--flow quick: the Plan has Mx milestone rows, and the quick flow has a single boundary — drop the milestones or keep the full flow")
		}
		return Flow{Kind: Quick, Provenance: Operator}, nil
	case string(Full):
		return Flow{Kind: Full, Provenance: Operator}, nil
	default:
		return Flow{}, fmt.Errorf("--flow %q: want %s or %s", in.Pin, Quick, Full)
	}
	r := in.Recorded
	if in.HasMilestones && (r == nil || r.Kind != Full) {
		return Flow{Kind: Full, Provenance: Inferred}, nil
	}
	if r != nil && (r.Provenance == Operator || r.Kind == Full) {
		return Flow{Kind: r.Kind, Provenance: r.Provenance}, nil
	}
	if in.HasPlan {
		return Flow{Kind: Full, Provenance: Inferred}, nil
	}
	return Flow{Kind: Quick, Provenance: Inferred}, nil
}
