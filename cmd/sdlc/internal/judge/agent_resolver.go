package judge

import (
	"fmt"
	"os"
	"strings"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
)

// AgentDefaultEnv is the explicit environment slice used to derive a default
// judge agent; agentid owns it so commit stamping (#367) shares it.
type AgentDefaultEnv = agentid.Env

// CurrentAgentDefaultEnv reads the process environment at the IO boundary.
func CurrentAgentDefaultEnv() AgentDefaultEnv { return agentid.EnvFrom(os.Getenv) }

// ResolveAgentCLI selects the judge agent: the main agent's own CLI (the `same`
// reviewer policy). Explicit --agent and AGENT_CMD are operator/script
// overrides, so unknown values are returned as-is for BuildArgs to reject
// instead of being silently hidden by auto-detection.
func ResolveAgentCLI(explicit string, explicitSet bool, env AgentDefaultEnv) AgentCLI {
	return ResolveReviewerCLI(explicit, explicitSet, env, ReviewerPolicy{})
}

// ResolveReviewerCLI selects the reviewer under a repo's policy (#310). The
// overrides stay above the policy; otherwise the policy maps the main agent.
func ResolveReviewerCLI(explicit string, explicitSet bool, env AgentDefaultEnv, policy ReviewerPolicy) AgentCLI {
	if explicitSet {
		return AgentCLI(explicit)
	}
	if env.AgentCmd != "" {
		return AgentCLI(env.AgentCmd)
	}
	return policy.Reviewer(MainAgent(env))
}

// MainAgent is the agent driving this session, from its environment signals.
func MainAgent(env AgentDefaultEnv) AgentCLI { return AgentCLI(env.Main()) }

// ReviewerMode is how a repo picks its reviewer from the main agent (#310,
// reviving #129's modes).
type ReviewerMode string

const (
	ReviewerSame     ReviewerMode = "same"     // the main agent reviews itself
	ReviewerOther    ReviewerMode = "other"    // a family absent from the authors (#409)
	ReviewerExplicit ReviewerMode = "explicit" // Pairs names each main agent's reviewer
)

// DefaultReviewerMode is the mode of a repo with no policy. Changing the
// fleet's default is the operator's call (#409); it is this one constant.
const DefaultReviewerMode = ReviewerSame

// ReviewerPolicy is a repo's reviewer choice. The zero value is
// DefaultReviewerMode with the default orders and each CLI's default model.
type ReviewerPolicy struct {
	Mode   ReviewerMode
	Pairs  map[AgentCLI]AgentCLI
	Orders map[ChangeKind][]AgentCLI // preference per kind under `other`; nil keeps defaultOrders
	Models map[AgentCLI]string       // the model to ask each reviewer for; absent is the CLI's default
}

func (p ReviewerPolicy) mode() ReviewerMode {
	if p.Mode == "" {
		return DefaultReviewerMode
	}
	return p.Mode
}

func (p ReviewerPolicy) order(kind ChangeKind) []AgentCLI {
	if o, ok := p.Orders[kind]; ok {
		return o
	}
	if o, ok := defaultOrders[kind]; ok {
		return o
	}
	return defaultOrders[KindBounded]
}

// Reviewer maps the main agent to one reviewer, ignoring availability: under
// `other` the first agent of another family for a bounded change. An agent
// `explicit` doesn't pair reviews itself.
func (p ReviewerPolicy) Reviewer(main AgentCLI) AgentCLI {
	switch p.mode() {
	case ReviewerOther:
		// Never the main agent's own family: Validate keeps two families in
		// every order, and the default orders hold all three besides.
		for _, q := range []ReviewerPolicy{p, {Mode: ReviewerOther}} {
			if plan, _ := ReviewPlan(q, main, []Family{FamilyOf(string(main))}, KindBounded, nil); len(plan) > 0 {
				return plan[0].Agent
			}
		}
	case ReviewerExplicit:
		if r, ok := p.Pairs[main]; ok {
			return r
		}
	}
	return main
}

// Validate refuses a mode or agent name the resolver doesn't know, so a typo
// in a repo's config fails loudly instead of quietly reviewing with `same`.
func (p ReviewerPolicy) Validate() error {
	switch p.Mode {
	case "", ReviewerSame, ReviewerOther:
		if len(p.Pairs) > 0 {
			return fmt.Errorf("reviewer_pairs applies only to boundary_reviewer %q", ReviewerExplicit)
		}
	case ReviewerExplicit:
		if len(p.Pairs) == 0 {
			return fmt.Errorf("boundary_reviewer %q needs reviewer_pairs", ReviewerExplicit)
		}
		for main, rev := range p.Pairs {
			for _, a := range []AgentCLI{main, rev} {
				if _, ok := knownAgent(string(a)); !ok {
					return fmt.Errorf("reviewer_pairs: unknown agent %q (supported: claude, codex, gemini)", a)
				}
			}
		}
	default:
		return fmt.Errorf("unknown boundary_reviewer %q (same | other | explicit)", p.Mode)
	}
	for kind, order := range p.Orders {
		if kind != KindDesignHeavy && kind != KindBounded {
			return fmt.Errorf("reviewer_order: unknown change kind %q (%s | %s)", kind, KindDesignHeavy, KindBounded)
		}
		seen, families := map[AgentCLI]bool{}, map[Family]bool{}
		for _, a := range order {
			if _, ok := knownAgent(string(a)); !ok {
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
		if _, ok := knownAgent(string(a)); !ok {
			return fmt.Errorf("reviewer_models: unknown agent %q (supported: claude, codex, gemini)", a)
		}
		if strings.TrimSpace(model) == "" || strings.ContainsAny(model, " \t\n") || strings.HasPrefix(model, "-") {
			return fmt.Errorf("reviewer_models: %s names no usable model %q", a, model)
		}
	}
	return nil
}

func knownAgent(s string) (AgentCLI, bool) {
	if !agentid.Known(s) {
		return "", false
	}
	return AgentCLI(s), true
}
