---
id: 000129
status: working
deps: []
github_issue:
created: 2026-06-25
updated: 2026-06-29
estimate_hours: 2.0
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

- [x] Add a pure shared resolver for judge-agent defaults.
- [x] Route judge, change-code, close/milestone-close, and preflight dispatch through it.
- [x] Update help text and tests for the new default contract.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v2.md` against `baseline-v2.md`. Method A only.

```estimate
model: estimate-logic-v2
familiarity: 1.0
item: cross-cutting-refactor design=0.25 impl=0.55
item: smaller-go-module      design=0.05 impl=0.50
item: atlas-docs             design=0.05 impl=0.10
item: milestone-review       design=0.05 impl=0.30
design-buffer: 0.30
total: 2.0
```

## Log

### 2026-06-25

- Created from investigation of closing gates: `PAIR_AGENT=codex` is present in
  Codex pair sessions, but current `sdlc` call sites consult `AGENT_CMD` and then
  coerce empty agent values to Claude.

### 2026-06-29

- Claimed via `sdlc claim --issue 129`; `sdlc start-plan --issue 129` delivered
  `ARCH-DRY`, `ARCH-PURE`, and `ARCH-PURPOSE`. Plan recorded in
  `workshop/plans/000129-default-sdlc-judges-to-current-agent-plan.md`.
- Implemented `judge.ResolveAgentCLI` and `judge.CurrentAgentDefaultEnv` as the
  single default-selection path (`ARCH-DRY`, `ARCH-PURE`). Routed `judge`,
  `change-code`, close boundary reviews, milestone boundary reviews, and
  preflight judges through it; close dry-run now prints the would-be boundary
  review command.
- Verification passed:
  `go test ./cmd/sdlc/internal/judge -run ResolveAgentCLI -count=1`;
  `go test ./cmd/sdlc -run 'Judge.*Agent|ChangeCode.*Agent|Milestone.*Agent|Close.*Agent|Preflight.*Agent' -count=1`;
  `go test ./cmd/sdlc/internal/judge ./cmd/sdlc -count=1`;
  `go test ./cmd/sdlc ./cmd/sdlc/internal/... ./pkg/... -count=1`;
  `env -u AGENT_CMD PAIR_AGENT=codex go run ./cmd/sdlc judge dry --dry-run --base HEAD --head HEAD`.
