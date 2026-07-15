---
id: 000174
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours:
started: 2026-07-14T19:35:30-07:00
---

# close: specify the post-FIX-THEN-SHIP protocol (stop the re-close loop and the bookkeeping publish-gate trip)

## Problem

#172's friction audit (M4 Findings) measured two recurring post-close loops,
both rooted in the same gap: **nothing specifies what happens between a
FIX-THEN-SHIP verdict and merge.**

1. **The re-close loop.** FIX-THEN-SHIP finalizes the close; the agent then
   commits fixes and — with no sanctioned way to "record" them — re-runs
   `sdlc close`, hits the reclose guard, and bypasses it. Evidence: the reclose
   guard is the ONLY gate whose refusals resolve 3/3 *via bypass*; 25 codex
   re-close bypasses (claude's dominant no-judge-bypass rationale is the same
   loop: "review ALREADY RAN, SHIP, twice"); live trace ariadne#145
   (close → FIX-THEN-SHIP → fixes → re-close → re-close).
2. **The bookkeeping publish-gate trip.** Post-close bookkeeping commits (issue
   Log, lessons, plan ticks) move HEAD past the reviewed anchor, so the
   merge/push publish gate (#160 reviewed-HEAD-unchanged) refuses and agents
   pass `--no-judge`. Evidence: 6/6 merge/push publish-gate refusals in the
   corpus are this shape.

## Spec

Candidates (pick at design time):
- close's FIX-THEN-SHIP output states the protocol explicitly: "commit the
  fixes (with the Review-Verdict trailer); do NOT re-close".
- a lightweight `sdlc reverify --issue N` that updates `--verified` evidence /
  records fix commits without re-running the close pipeline.
- close orders its own bookkeeping writes BEFORE finalizing, so the anchor is
  the last commit; and/or the publish gate tolerates doc-only deltas
  (workshop/, atlas/, lessons.md) after the anchor.

## Done when

- After a FIX-THEN-SHIP verdict, the agent-facing next action is unambiguous
  (stated in close's output) and does not require bypassing any gate.
- Post-close bookkeeping no longer trips the publish gate (or the gate's
  refusal explains the sanctioned path).
- Re-measure with `sdlc process-manual --friction-report`: reclose-guard
  via-bypass resolutions and merge/push no-judge bypasses drop.

## Plan

- [ ] brainstorm the mechanism (output-protocol vs `sdlc reverify` vs bookkeeping-ordering) and pick at design time

## Log

### 2026-07-14

Filed from #172 M4 (T3 findings 1–2).
