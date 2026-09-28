---
id: 000170
status: open
created: 2026-07-13
updated: 2026-07-13
estimate_hours:
github_issue:
---

# audit ariadne stack for opportunities to simplify

## Problem

The stack grew organically to ~100 markdown artifacts (67 SKILL.md + 22 helptext
+ 7 judge prompts + 5 AGENTS-chain) + 5 `cmd/` binaries + the introspect
knowledge. Much of the raw *count* is on-demand (helptext fires at `--help`,
judge prompts at a gate, skill bodies on trigger) — the real cost is **always-on
context** (constitution + all 721 lines of `lessons.md` + 67 skill-trigger
lines) and **conceptual load**. Audit holistically, then simplify.

Prereqs: **#169** (fresh introspect run) and **#172** (sdlc painpoint audit) —
each feeds this one with data. Run **#169 → #172 → #170**.
