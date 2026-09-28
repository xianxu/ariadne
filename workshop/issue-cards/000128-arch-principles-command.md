---
id: '000128'
status: done
started: 2026-06-25T11:25:20-07:00
created: 2026-06-25
updated: 2026-06-25
estimate_hours: 1.66
actual_hours: 0.13
---

# sdlc arch-principles command — single-source the ARCH-* narrative; replace AGENTS.base.md abstract with a command, keep the start-plan push

## Problem

`AGENTS.base.md` "Core Design Principles" hand-maintains a one-line abstract of each
`ARCH-*` marker (DRY/PURE/PURPOSE) plus a file-path pointer to the registry. That abstract
is a **parallel restatement** of `architecture.md`'s `principle:` lines, bound only by a
*presence-check* drift test (`TestArchitecture_NarrativeInSyncWithAgentsMd` asserts each
marker string appears) — the prose itself can drift unguarded. This is an `ARCH-PURPOSE`
smell (#126) on #126's own narrative: a consumer that *restates* the model instead of
*deriving* from it.

Two delivery facts frame the fix:
- The registry is already **pushed** at the gates: `sdlc start-plan` prints the `at-plan`
  lens to the main thread; the plan-quality + boundary-review judges embed it inline. The
  judge prompts are **fresh-context subagents** — they MUST keep the inline embed
  (`architecture.go`: "a marker alone would be a dangling pointer in a fresh-context
  subagent — the definitions must be co-present"). That embed is non-negotiable.
- The gap pure-injection leaves: **non-gate work** (§7 autonomous bug-fixing, trivial
  fixes, Q&A, in-session iteration) never runs `start-plan`, so it never sees ARCH-* —
  today its only source is the static abstract.
