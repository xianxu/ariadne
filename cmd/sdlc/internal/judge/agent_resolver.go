package judge

import (
	"os"

	"github.com/xianxu/ariadne/cmd/sdlc/internal/agentid"
	"github.com/xianxu/ariadne/pkg/reviewpolicy"
)

// AgentDefaultEnv is the explicit environment slice used to derive a default
// judge agent; agentid owns it so commit stamping (#367) shares it.
type AgentDefaultEnv = agentid.Env

// CurrentAgentDefaultEnv reads the process environment at the IO boundary.
func CurrentAgentDefaultEnv() AgentDefaultEnv { return agentid.EnvFrom(os.Getenv) }

// MainAgent is the agent driving this session, from its environment signals.
func MainAgent(env AgentDefaultEnv) AgentCLI { return AgentCLI(env.Main()) }

// The reviewer policy (#310, #409) is pkg/reviewpolicy's, shared with
// doc-review (#441); judge re-exports it under its historical names.
type (
	ReviewerMode   = reviewpolicy.Mode
	ReviewerPolicy = reviewpolicy.Policy
)

const (
	ReviewerSame        = reviewpolicy.Same
	ReviewerOther       = reviewpolicy.Other
	ReviewerExplicit    = reviewpolicy.Explicit
	DefaultReviewerMode = reviewpolicy.DefaultMode
)
