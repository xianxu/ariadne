---
id: 000600
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '790d36791b000adc5a70827be5711750ed7bbfbd' # card fields mirrored from issue-cards; edit via sdlc
---

# tl-lead: combined operator-testing moment (batch green items into one staged restart with one checklist)

## Problem

## Spec

**Make the combined operator-testing moment a TL procedure, with a small tl-lead verb (operator request, 10-24).** On
10-24 the TL improvised one combined live check: four green items (pair#851, #856, #838, #842) plus parley.nvim#419
staged together for a single Couch restart, one checklist side doc, per-issue PASS or first problem, results relayed
to each owner. The operator: "the combined testing works really well … organize product testing of many landings and
create a combined testing moment that makes sense." Today tl-lead's rules cover one staged item at a time.

The procedure that worked (ops session side doc 2026-10-24-04 has the full write-up):
1. **Collect:** items that are green (recorded suite passed, close SHIP) and need an operator live look, per machine
   target (pair:0, parley.nvim:0).
2. **Decide the moment:** batch them into one restart when the batch is worth the operator's time (a morning, or a
   priority item ready); exclude anything not green rather than wait for it.
3. **Build once:** one coordinator slot builds a staging-only combined merge on the target with provenance (exact
   heads, every binary at one revision) via `sdlc stage`; that merge never lands.
4. **One checklist:** one side doc, a heading per issue, plain-language steps (from each item's operator steps),
   PASS or the first problem under each, plus any restart caveats.
5. **Relay:** each result goes to its owner; passing items close and land individually; failing ones are fixed and
   ride the next moment; the coordinator clears the stage after the last check.

## Done when

- `tl-lead --help` describes the combined-moment procedure (when to batch, who coordinates, exclusion rule, relay,
  clear).
- A tl-lead verb lists items awaiting an operator look (green, with operator steps) per target, and drafts the
  checklist side doc from their steps; staging itself stays `sdlc stage`.
- Tested against a fake fleet state with mixed green/not-green items and two targets.

## Plan

- [ ]

## Log

### 2026-10-24
