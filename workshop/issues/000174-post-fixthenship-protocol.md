---
id: 000174
status: working
deps: []
github_issue:
created: 2026-07-14
updated: 2026-07-14
estimate_hours: 0.97
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module    design=0.1   impl=0.12
item: smaller-go-module    design=0.15  impl=0.15
item: smaller-go-module    design=0.05  impl=0.05
item: atlas-docs           design=0.05  impl=0.1
item: milestone-review     design=0.0   impl=0.15
design-buffer: 0.15
total: 0.97
```

Σdesign 0.35 × 1.15 + Σimpl 0.57 × 1.0 = 0.97. Items: leg A (FIX-THEN-SHIP
protocol block + formatter + 3 tests), leg C (publish-gate docs-only branch +
formatter + 4 subtests + mutation check), reclose refusal append + test,
helptext ×4 + atlas sweep, and the close-time boundary review. +15% buffer:
thorough plan doc. *Produced via
`brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

Durable design: `workshop/plans/000174-post-fixthenship-protocol-plan.md`.
Decision (operator, 2026-07-14): candidates A + C; B (`sdlc reverify`)
skipped — without teeth it's `--no-judge` with nicer bookkeeping. Plus one
leg: reclose refusal names the new-issue path. Single-pass, plain checkboxes.

- [x] brainstorm the mechanism (output-protocol vs `sdlc reverify` vs bookkeeping-ordering) and pick at design time
- [x] leg A: `formatFixThenShipProtocol` + verdict-conditional block in the closeFinalize arm (+ SHIP negative test)
- [x] leg C: `runPublishGate` docs-only pass via `hasCodePath` (#177, ARCH-DRY) + augmented refusal (pinned span preserved) + multi-issue subtest + mutation check
- [x] reclose refusal appends the new-issue recovery (append-only — gatesig + codex golden fixtures freeze the head span)
- [x] docs sweep: helptext close/milestone-close/merge/push + atlas publish-gate prose; bookkeeping

## Log

### 2026-07-14

Filed from #172 M4 (T3 findings 1–2).

Design discussion (operator + agent): the happy path already exists — close
doesn't commit, so bundling fixes + close mutations + bookkeeping into ONE
commit makes the anchor HEAD and every gate passes; nothing states it. The
anchor derivation (content read: newest commit leaving status codecomplete)
self-heals for issue-file-touching bookkeeping; the 6/6 publish-gate trips
are exactly the commits that don't touch it (lessons.md, plan ticks, atlas).
SHIP and FIX-THEN-SHIP are currently indistinguishable in close's output —
the protocol gap is total, hence the route-arounds. Plan reviewed fresh-eyes
(approved; multi-issue anchor soundness verified: an inter-anchor code commit
is inside the newest close's branch-point→HEAD review window, so
newestAnchor..HEAD docs-only is sound for the #160 invariant).

Implementation: all legs landed TDD (red→green per leg). formatFixThenShipProtocol
is verb-parameterized (closeVerb — the plan-quality judge's finding), so milestone
closes name `sdlc milestone-close` in the escape hatch, mirroring the REWORK arm
(ARCH-DRY). Leg C reuses hasCodePath (#177) + gitx.DiffNames; fail-closed on git
error; mutation-checked (inverted predicate reddens the docs-only subtests). All
new lines gatesig-collision-checked; pinned refusal heads preserved append-only
(gatesig + frozen codex golden fixtures). Docs: helptext close/milestone-close/
merge/push + atlas pre-merge-checks.md (canonical) + issue-lifecycle.md.
