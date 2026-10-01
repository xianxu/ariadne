// claimant.go — who is responsible for an issue and where its work belongs
// (#277). The record lives on the card, is written with the claim's status
// compare-and-swap, and is mirrored into details like every card field.
package issue

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ClaimantField is the card field holding the responsibility record.
const ClaimantField = "claimant"

// Claimant identifies a workspace. Repository, Machine and Worktree decide
// ownership; Operator, MachineName and Workspace are descriptive. Workspace is
// a slot label (`repo:N`) when the checkout has one — optional, never required:
// Ariadne works without slots or Couch.
type Claimant struct {
	Operator    string
	Machine     string // MachineFingerprint of the OS machine ID, never the raw ID
	MachineName string
	Workspace   string // optional
	Worktree    string // canonical absolute path
	Repository  string
}

// claimantKeys is the record's exact schema, in rendering order.
var claimantKeys = []struct {
	key      string
	optional bool
	get      func(*Claimant) *string
}{
	{"operator", false, func(c *Claimant) *string { return &c.Operator }},
	{"machine", false, func(c *Claimant) *string { return &c.Machine }},
	{"machine_name", false, func(c *Claimant) *string { return &c.MachineName }},
	{"workspace", true, func(c *Claimant) *string { return &c.Workspace }},
	{"worktree", false, func(c *Claimant) *string { return &c.Worktree }},
	{"repository", false, func(c *Claimant) *string { return &c.Repository }},
}

var fingerprintRE = regexp.MustCompile(`^[0-9a-f]{32}$`)

// MachineFingerprint keys the OS machine ID to Ariadne claims, so the published
// value matches exactly when the machine does without revealing or linking the
// raw identifier (the tracker may be public).
func MachineFingerprint(raw string) string {
	sum := sha256.Sum256([]byte("ariadne-claimant\x00" + raw))
	return fmt.Sprintf("%x", sum[:16])
}

// parseClaimant validates a claimant node: a block mapping of exactly the
// schema's keys to nonempty strings. Anything else fails closed.
func parseClaimant(n *yaml.Node) (Claimant, error) {
	fail := func(reason string) (Claimant, error) { return Claimant{}, invalidCard(ClaimantField, reason) }
	if n.Kind != yaml.MappingNode || n.Style&yaml.FlowStyle != 0 {
		return fail("must be a block mapping")
	}
	var c Claimant
	seen := map[string]bool{}
	for i := 0; i < len(n.Content); i += 2 {
		k, v := n.Content[i], n.Content[i+1]
		var dst *string
		for _, s := range claimantKeys {
			if s.key == k.Value {
				dst = s.get(&c)
			}
		}
		if dst == nil {
			return fail("unknown key " + k.Value)
		}
		if v.Kind != yaml.ScalarNode || v.Tag != "!!str" || strings.TrimSpace(v.Value) == "" || strings.ContainsAny(v.Value, "\r\n") {
			return fail(k.Value + " must be a nonempty one-line string")
		}
		*dst, seen[k.Value] = v.Value, true
	}
	for _, s := range claimantKeys {
		if !s.optional && !seen[s.key] {
			return fail("missing " + s.key)
		}
	}
	if !fingerprintRE.MatchString(c.Machine) {
		return fail("machine must be a fingerprint (32 lowercase hex), never a raw ID")
	}
	return c, nil
}

// CardClaimant reads a card's claimant; ok is false when it has none.
func CardClaimant(card []byte) (Claimant, bool, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return Claimant{}, false, err
	}
	n := d.fields[ClaimantField]
	if n == nil {
		return Claimant{}, false, nil
	}
	c, err := parseClaimant(n)
	return c, err == nil, err
}

// SetCardClaimant writes c as the card's claimant, replacing any earlier one
// and leaving every other byte alone.
func SetCardClaimant(card []byte, c Claimant) ([]byte, error) {
	d, err := parseCardDocument(card)
	if err != nil {
		return nil, err
	}
	block, err := c.render()
	if err != nil {
		return nil, err
	}
	fm := d.fm + "\n" + block
	if span, ok := d.spans[ClaimantField]; ok {
		fm = d.fm[:span.start] + block + d.fm[span.end:]
	}
	out := []byte(Compose(fm, d.body))
	if _, err := ParseCard(out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c Claimant) render() (string, error) {
	m := &yaml.Node{Kind: yaml.MappingNode}
	for _, s := range claimantKeys {
		v := *s.get(&c)
		if v == "" && s.optional {
			continue
		}
		m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: s.key},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v})
	}
	doc := &yaml.Node{Kind: yaml.MappingNode, Content: []*yaml.Node{{Kind: yaml.ScalarNode, Value: ClaimantField}, m}}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// Ownership is a card's claimant judged against the current workspace.
type Ownership int

const (
	OwnershipUnknown Ownership = iota // no claimant recorded (legacy or never claimed)
	OwnershipMine
	OwnershipForeign
)

func (o Ownership) String() string {
	return [...]string{"unknown", "mine", "foreign"}[o]
}

// MatchClaimant decides ownership. The same repository, machine and worktree
// make a workspace the owner; operator and slot label are descriptive, and
// matching operator or machine alone never does.
func MatchClaimant(recorded *Claimant, current Claimant) Ownership {
	switch {
	case recorded == nil:
		return OwnershipUnknown
	case recorded.Repository == current.Repository && recorded.Machine == current.Machine && recorded.Worktree == current.Worktree:
		return OwnershipMine
	default:
		return OwnershipForeign
	}
}

// RelocationAllowed lets an owner move its own work to another worktree on the
// same machine (sdlc move, or claim repairing an interrupted move): the current
// checkout is on the issue branch and the recorded worktree — on this machine,
// so locally observable — no longer holds it. Never across machines or
// repositories, and never while the old worktree still holds the branch.
func RelocationAllowed(recorded, current Claimant, onIssueBranch, oldWorktreeHoldsBranch bool) bool {
	return onIssueBranch && !oldWorktreeHoldsBranch &&
		recorded.Repository == current.Repository && recorded.Machine == current.Machine &&
		recorded.Worktree != current.Worktree
}
