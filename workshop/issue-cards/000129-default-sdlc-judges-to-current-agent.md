---
id: '000129'
status: done
started: 2026-06-29T10:16:30-07:00
created: 2026-06-25
updated: 2026-06-29
estimate_hours: 2.0
actual_hours: 0.30
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
