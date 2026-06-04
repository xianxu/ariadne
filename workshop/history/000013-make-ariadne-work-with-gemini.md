---
id: 000013
status: done
deps: []
created: 2026-04-28
updated: 2026-06-03
---

# make ariadne work with gemini

likely it's about issue #11, to make .claude's skill system (https://agentskills.io/home) working in gemini. may work out of box already.

## Done when

- A Google agent CLI can follow the ariadne constitution + run the SDLC workflow.

## Spec

Target shifted: **`gemini` is deprecated**; Google's current agent CLI is
**Antigravity** ("agy"). The goal — ariadne works with a Google agent — is met
there, and via the same agent-agnostic path as [[#11]]: the agent reads the
shared `AGENTS.md` (no claude-specific anything required) and shells out to the
`sdlc` binary for the workflow. The original guess ("may work out of box
already") held — no ariadne-side change was needed.

## Plan

- [x] Validated: Antigravity follows the AGENTS.md constitution + drives the
  `sdlc` workflow. No code change — it works via the binary-first / shared-
  AGENTS.md architecture (gemini-specific work moot — deprecated).

## Log

### 2026-06-03 — done (target shifted gemini → Antigravity)
- 2026-06-03: closed — ariadne works under Antigravity (Google new agentic CLI; gemini deprecated) — agent reads shared AGENTS.md + runs sdlc binary, no ariadne-side change (the agent-agnostic architecture from #11 carried it). Validation close: no implementation to attribute (--no-actual); doc-only (--no-judge, --no-atlas).; review verdict: not-run
- Operator: works under **Antigravity** (Google's new agentic CLI; `gemini`
  deprecated). No ariadne-side implementation — agent-agnosticism via shared
  `AGENTS.md` + the `sdlc` binary (see #11) carried it. Closing as done against
  the *intent* (Google-agent compatibility), not the deprecated literal target.
- Note: `dispatch.go` still has a `gemini` AgentCLI branch; harmless, but an
  `antigravity`/`agy` branch could be added if/when the judge dispatches to it
  (not needed for the agent to *run* ariadne — only for ariadne to dispatch
  judges *to* that agent). Left for a fresh issue if it becomes live.

### 2026-04-28

