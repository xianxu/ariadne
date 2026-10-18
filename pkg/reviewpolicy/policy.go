// Package reviewpolicy is who reviews whom (#409), shared by every binary that
// dispatches a reviewer (#441: sdlc's judge re-exports it, doc-review imports
// it): reviewers are planned from the vendor families of the change's authors
// and the kind of change, within a repo's #310 policy. Everything here is
// pure; callers supply availability and run the candidates.
package reviewpolicy

import (
	"fmt"
	"strings"

	"github.com/xianxu/ariadne/pkg/agentenv"
)

// Agent is a reviewer CLI name.
type Agent string

const (
	Claude Agent = agentenv.Claude
	Codex  Agent = agentenv.Codex
	Gemini Agent = agentenv.Gemini
)

// Family is a model vendor. Two agents of one family share blind spots, which
// is what cross-family review is for.
type Family string

const (
	Anthropic Family = "anthropic"
	OpenAI    Family = "openai"
	Google    Family = "google"
)

// FamilyOf is an agent's vendor, or "" for one sdlc does not know (a human,
// an unrecognized CLI): an unknown author is never guessed into a family.
func FamilyOf(agent string) Family {
	switch Agent(agent) {
	case Claude:
		return Anthropic
	case Codex:
		return OpenAI
	case Gemini:
		return Google
	}
	return ""
}

// ChangeKind is what the review judges: a design (a new mechanism, state or
// invariant) or a bounded change.
type ChangeKind string

const (
	DesignHeavy ChangeKind = "design-heavy"
	Bounded     ChangeKind = "bounded"
)

// defaultOrders is the preference among candidate reviewers per kind, before
// the authors' families are removed. Designs go first to Claude, whose reviews
// caught Codex designs over-engineering; bounded code goes first to Codex,
// whose reviews caught bugs in Claude code (#409's evidence). Claude and Codex
// lead both, so #310's `other` (claude and codex review each other) holds;
// Gemini is the fallback when neither can.
var defaultOrders = map[ChangeKind][]Agent{
	DesignHeavy: {Claude, Codex, Gemini},
	Bounded:     {Codex, Claude, Gemini},
}

// Mode is how a repo picks its reviewer from the main agent (#310).
type Mode string

const (
	Same     Mode = "same"     // the main agent reviews itself
	Other    Mode = "other"    // a family absent from the authors (#409)
	Explicit Mode = "explicit" // Pairs names each main agent's reviewer
)

// DefaultMode is the mode of a repo with no policy. Changing the fleet's
// default is the operator's call (#409); it is this one constant.
const DefaultMode = Same

// Policy is a repo's reviewer choice. The zero value is DefaultMode with the
// default orders and each CLI's default model.
type Policy struct {
	Mode   Mode
	Pairs  map[Agent]Agent
	Orders map[ChangeKind][]Agent // preference per kind under `other`; nil keeps defaultOrders
	Models map[Agent]string       // the model to ask each reviewer for; absent is the CLI's default
}

// EffectiveMode is the policy's mode, DefaultMode when unset.
func (p Policy) EffectiveMode() Mode {
	if p.Mode == "" {
		return DefaultMode
	}
	return p.Mode
}

func (p Policy) order(kind ChangeKind) []Agent {
	if o, ok := p.Orders[kind]; ok {
		return o
	}
	if o, ok := defaultOrders[kind]; ok {
		return o
	}
	return defaultOrders[Bounded]
}

// Reviewer maps the main agent to one reviewer, ignoring availability: under
// `other` the first agent of another family for a bounded change. An agent
// `explicit` doesn't pair reviews itself.
func (p Policy) Reviewer(main Agent) Agent {
	switch p.EffectiveMode() {
	case Other:
		// Never the main agent's own family: Validate keeps two families in
		// every order, and the default orders hold all three besides.
		for _, q := range []Policy{p, {Mode: Other}} {
			if plan, _ := Plan(q, main, []Family{FamilyOf(string(main))}, Bounded, nil); len(plan) > 0 {
				return plan[0].Agent
			}
		}
	case Explicit:
		if r, ok := p.Pairs[main]; ok {
			return r
		}
	}
	return main
}

// Validate refuses a mode or agent name the resolver doesn't know, so a typo
// in a repo's config fails loudly instead of quietly reviewing with `same`.
func (p Policy) Validate() error {
	switch p.Mode {
	case "", Same, Other:
		if len(p.Pairs) > 0 {
			return fmt.Errorf("reviewer_pairs applies only to boundary_reviewer %q", Explicit)
		}
	case Explicit:
		if len(p.Pairs) == 0 {
			return fmt.Errorf("boundary_reviewer %q needs reviewer_pairs", Explicit)
		}
		for main, rev := range p.Pairs {
			for _, a := range []Agent{main, rev} {
				if !agentenv.Known(string(a)) {
					return fmt.Errorf("reviewer_pairs: unknown agent %q (supported: claude, codex, gemini)", a)
				}
			}
		}
	default:
		return fmt.Errorf("unknown boundary_reviewer %q (same | other | explicit)", p.Mode)
	}
	for kind, order := range p.Orders {
		if kind != DesignHeavy && kind != Bounded {
			return fmt.Errorf("reviewer_order: unknown change kind %q (%s | %s)", kind, DesignHeavy, Bounded)
		}
		seen, families := map[Agent]bool{}, map[Family]bool{}
		for _, a := range order {
			if !agentenv.Known(string(a)) {
				return fmt.Errorf("reviewer_order: unknown agent %q (supported: claude, codex, gemini)", a)
			}
			if seen[a] {
				return fmt.Errorf("reviewer_order: %s lists %s twice", kind, a)
			}
			seen[a], families[FamilyOf(string(a))] = true, true
		}
		// One family cannot review its own authors cross-family (#409).
		if len(families) < 2 {
			return fmt.Errorf("reviewer_order: %s needs agents of at least two families, so another family can always review", kind)
		}
	}
	for a, model := range p.Models {
		if !agentenv.Known(string(a)) {
			return fmt.Errorf("reviewer_models: unknown agent %q (supported: claude, codex, gemini)", a)
		}
		if strings.TrimSpace(model) == "" || strings.ContainsAny(model, " \t\n") || strings.HasPrefix(model, "-") {
			return fmt.Errorf("reviewer_models: %s names no usable model %q", a, model)
		}
	}
	return nil
}

// Candidate is one reviewer to try, with the model to ask for ("" is the
// CLI's own default).
type Candidate struct {
	Agent  Agent
	Model  string
	Family Family
}

func (c Candidate) String() string {
	s := string(c.Agent)
	if c.Model != "" {
		s += "/" + c.Model
	}
	return fmt.Sprintf("%s (%s)", s, orUnknown(string(c.Family), "unknown family"))
}

// Skip is a candidate passed over before or instead of reviewing, and why.
type Skip struct {
	Agent  Agent
	Reason string
}

// Availability reports, before launch, whether an agent can review here, and
// why not when it cannot.
type Availability func(Agent) (reason string, ok bool)

// Plan is the ordered reviewers to try for a change by authors of the given
// families, under policy. `same` and `explicit` are the #310 floor and name one
// reviewer; the dynamic `other` lists every agent whose family is not among
// the authors, in the kind's order. Candidates available rejects are returned
// as skips; an empty plan means none can review here.
func Plan(policy Policy, main Agent, authors []Family, kind ChangeKind, available Availability) ([]Candidate, []Skip) {
	var agents []Agent
	switch policy.EffectiveMode() {
	case Other:
		excluded := map[Family]bool{}
		for _, f := range authors {
			excluded[f] = true
		}
		for _, a := range policy.order(kind) {
			if !excluded[FamilyOf(string(a))] {
				agents = append(agents, a)
			}
		}
	case Explicit:
		agents = []Agent{policy.Reviewer(main)}
	default:
		agents = []Agent{main}
	}
	var plan []Candidate
	var skips []Skip
	for _, a := range agents {
		if available != nil {
			if reason, ok := available(a); !ok {
				skips = append(skips, Skip{Agent: a, Reason: reason})
				continue
			}
		}
		plan = append(plan, Candidate{Agent: a, Model: policy.Models[a], Family: FamilyOf(string(a))})
	}
	return plan, skips
}

// AuthorFamilies is the session's family plus every family the commits' Agent
// trailer values (`<agent>/<model>/<effort>`) name; merges and agents sdlc
// does not know contribute nothing. The session's family is always included,
// so a forged trailer can narrow the candidates but never remove the closer's
// own family.
func AuthorFamilies(session Family, commits []agentenv.Commit) []Family {
	families := []Family{session}
	for _, c := range commits {
		if c.Merge {
			continue
		}
		for _, v := range c.Agents {
			agent, _, _ := strings.Cut(strings.TrimSpace(v), "/")
			if f := FamilyOf(agent); f != "" {
				families = append(families, f)
			}
		}
	}
	return families
}

func orUnknown(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}
