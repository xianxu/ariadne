---
id: 000176
status: open
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours:
---

# spine guards for off-workflow invocations: change-code on done issues + non-SDLC repos

## Problem

#172's friction audit found the spine gets invoked where the workflow state
says it shouldn't be, with no guard until close-time:

1. **Working on a `done` issue is un-gated.** `change-code` (and every other
   verb) runs silently against a done issue; only re-close is guarded.
   Firing-order measured 11 change-code-after-close inversions (8 brain,
   2 pair, 1 ariadne-dogfood).
2. **brain runs the spine against its own charter.** brain is a Drive-like
   capture repo, not an SDLC repo, yet it concentrates 19 bypasses and 8/11
   firing-order anomalies — every gate becomes noise to route around there,
   and it pollutes cross-repo friction measurement.

## Spec

Candidates:
- `change-code` (and `start-plan`) warn or refuse when the issue is
  `status: done`: "reopen via the REWORK path or open a new issue".
- a repo-level guard: spine verbs refuse (or loudly warn) in repos without the
  SDLC layout (no `workshop/issues/`), or in repos marked non-SDLC (brain's
  `.brain/config.md`). Constitution already says don't run sdlc in brain — the
  binary should own the gate (#69 pattern: the binary, not agent memory).

## Done when

- Running `sdlc change-code --issue N` on a done issue produces a
  warn/refusal naming the reopen path.
- Spine verbs in a non-SDLC repo (brain) refuse/warn instead of half-working.
- Re-measure: brain's bypass + anomaly concentration drops out of the
  friction report.

## Plan

- [ ] design the two guards (done-issue warn on change-code/start-plan; non-SDLC-repo refusal) as one gate family

## Log

### 2026-07-14

Filed from #172 M4 (T3 findings 3–4).
