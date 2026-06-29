---
id: 000129
status: working
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-29
estimate_hours:
started: 2026-06-29T10:16:30-07:00
---

# Default sdlc judges to current agent

## Problem

`sdlc` judge dispatch supports multiple agent CLIs (`claude`, `codex`,
`gemini`) and several gates expose `--agent`, but the implicit default still
falls back to Claude when `AGENT_CMD` is unset. In pair/codex sessions the main
agent is already exposed as `PAIR_AGENT=codex`, so closing gates and plan-quality
judges can dispatch the wrong stack unless the operator remembers to pass
`--agent codex`.

The fresh-context review property should mean "new context for the same agent
stack by default", not "always Claude unless manually overridden".

Alternatively, we may have a configuration to drive such selection strategy, maybe:

- "same": use same coding agent as main
- "other": use different one, so codex main agent would use claude; claude main would use codex subagent. we do need to think generic case if we have N how to configure, maybe the next:
- "explicit": which is a string of: codex:claude,claude:codex etc. basically codex:claude means if codex is main, use claude as subagent.

The prose should be driven by the above configuration.

## Spec

Centralize judge-agent default resolution so every `sdlc` subagent review uses
the same precedence:

1. explicit `--agent`
2. `AGENT_CMD` for existing scripts and operator overrides
3. `PAIR_AGENT` when running under pair-managed sessions
4. conservative environment auto-detection for known stacks (`CODEX_*`,
   Claude/Gemini-specific signals)
5. existing Claude fallback when nothing identifies the caller

Apply the resolver to `sdlc judge`, `change-code` plan/estimate quality judges,
`close` boundary review, `milestone-close`, and preflight judges. Keep the
agent-specific command adapters (`claude -p`, `codex exec`, `gemini -p`) intact;
the change is only default selection, not a generic one-size command line.

Update help text and tests so the documented default is no longer simply
`$AGENT_CMD or claude`.

## Done when

- In a `PAIR_AGENT=codex` environment with no `AGENT_CMD`, `sdlc judge --dry-run`
  and close-boundary review dry runs show `codex exec`.
- Explicit `--agent claude` still wins over `PAIR_AGENT=codex`.
- Existing `AGENT_CMD=gemini` behavior is preserved.
- Unit tests cover the resolver precedence and at least one close/judge call
  path that previously fell back to Claude.

## Plan

- [ ] Add a pure shared resolver for judge-agent defaults.
- [ ] Route judge, change-code, close/milestone-close, and preflight dispatch through it.
- [ ] Update help text and tests for the new default contract.

## Log

### 2026-06-25

- Created from investigation of closing gates: `PAIR_AGENT=codex` is present in
  Codex pair sessions, but current `sdlc` call sites consult `AGENT_CMD` and then
  coerce empty agent values to Claude.
