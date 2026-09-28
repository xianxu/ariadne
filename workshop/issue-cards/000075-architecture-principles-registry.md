---
id: '000075'
status: done
created: 2026-06-03
updated: 2026-06-03
estimate_hours: 3
actual_hours: 2.75
---

# architecture.md — a referenced architectural-principles registry delivered to planning + reviews

## Problem

Agents are strong at **tactics** (clean function, handled edge case) and weak at
**architecture** (whole-board shape — will this design still be sound months
out?). The likely reason: architectural payoff is so far downstream that there's
little training signal for it, so the model can't have learned good taste here.
So architecture is the one place we should *inject human judgment* as an explicit,
persistent, prompt-level scaffold rather than hope the model supplies it.

Today the architectural principles are **scattered and under-delivered**:
- AGENTS.md "## Core Design Principles" (DRY, PURE, Simplicity, Root Cause) —
  human prose, never structured for machine consumption.
- `sdlc judge dry` / `sdlc judge pure` — two standalone judge prompts, each
  re-stating a principle, run as separate passes (not folded into the boundary
  review).
- The plan-quality judge checks executability but **not** architecture — yet
  planning is exactly where architecture is *decided*, so it's the highest-
  leverage place to inject these and the one currently missing them.

There's no single place to author an architectural principle once and have it
reach planning, plan-review, and code-review. (#71's "shim every external
service" is a future principle that needs exactly this to become enforceable
rather than aspirational.)
