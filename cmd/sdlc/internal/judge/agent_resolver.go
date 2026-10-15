package judge

import (
	"fmt"
	"os"

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
	ReviewerSame     ReviewerMode = "same"     // the main agent reviews itself (default)
	ReviewerOther    ReviewerMode = "other"    // claude and codex review each other
	ReviewerExplicit ReviewerMode = "explicit" // Pairs names each main agent's reviewer
)

// ReviewerPolicy is a repo's reviewer choice. The zero value is `same`.
type ReviewerPolicy struct {
	Mode  ReviewerMode
	Pairs map[AgentCLI]AgentCLI
}

// Reviewer maps the main agent to its reviewer. An agent `explicit` doesn't
// pair reviews itself.
func (p ReviewerPolicy) Reviewer(main AgentCLI) AgentCLI {
	switch p.Mode {
	case ReviewerOther:
		if main == AgentClaude {
			return AgentCodex
		}
		return AgentClaude
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
	return nil
}

func knownAgent(s string) (AgentCLI, bool) {
	if !agentid.Known(s) {
		return "", false
	}
	return AgentCLI(s), true
}
