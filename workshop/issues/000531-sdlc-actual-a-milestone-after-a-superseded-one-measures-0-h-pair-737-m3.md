---
id: 000531
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 1.38
collaboration-mode: collaboration
card_mirror: '7aad8d75d3016078e9df3aead092cf3fda7fc1e0' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T06:11:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# sdlc actual: a milestone after a superseded one measures 0 h (pair#737 M3)

## Problem

## Spec

pair#737 (pair:1, 10-21): after the operator-approved re-scope (old M2 superseded, new work tagged M3 per the TL's
instruction), `sdlc` measures **0.00 h** active time for M3 despite 21 admitted sessions and substantive M3 commits
since the re-scope. The worker correctly refuses to hand-type hours. Likely cause: the milestone's measurement window
is anchored on a boundary (the superseded M2's start/close, or the M3 tag's first appearance) that excludes the
actual M3 work, or a superseded milestone isn't modelled at all.

Fix: measure a milestone that follows a superseded one from the supersession point (the last closed boundary or the
revision that superseded it) to the milestone's close, and make the measurement explain its window (start commit,
end, session count) so a 0 h result is visibly wrong rather than silently adopted.

## Done when

- A fixture with M1 closed, M2 superseded (never closed), M3 worked on measures M3 from after M1's close (or the
  supersession revision) and attributes the M3 sessions; a regression for the 0 h case.
- `sdlc actual` prints its window and session count.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; source is flagged stale. Pinned Plan resolver uses existing Git and Markdown libraries (library availability discount); selector, diagnostics and consumer wiring extend existing modules. Detailed approved design applies the 0.2 design discount, familiarity 1.0 and 15% design buffer. Implementation entries use 40% of v2 ranges; review includes isolated production proof.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.05 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: smaller-go-module design=0.04 impl=0.12
item: smaller-go-module design=0.06 impl=0.20
item: milestone-review design=0.02 impl=0.20
item: atlas-docs design=0.03 impl=0.04
design-buffer: 0.15
total: 1.38
```

## Plan

Full-flow proposal: [implementation plan](../plans/000531-retired-boundary-accounting-plan.md).

- [ ] Replay pair#737 and define retired-boundary accounting from the pinned canonical Plan.
- [ ] Share cut selection across fresh tracker and legacy measurements; preserve retained replay cuts.
- [ ] Explain effective milestone phases and source counts; verify adversarial and full-suite coverage.
- [ ] Close through the mandatory review and land.

## Log

### 2026-10-21

### 2026-10-22 — claim and root-cause investigation

Claimed in ariadne:2 and started planning. BuildTimeline sends an unfinalized boundary's convergence to the complete measurement right edge, so its monotonic emitter leaves no M3 phases. Neither BoundaryRound nor gatestate.Round / MeasurementCut has supersession metadata.

Read-only pair#737 evidence in pair-slot1: M1 finalized 10-20 21:23:47 -07; M2 had four REWORK rounds, first dispatched 10-21 00:01:03 and last returned 10:31:26, never finalized. The issue's canonical Plan replaces M2's executable checkbox with a historical plain bullet; M3 is its own executable row. The durable plan explicitly preserves findings/cap/history without a SHIP claim. Git records that approved retirement at 10-21 10:59:39 -07. Its Log records M3 measuring 0.00h through 11:48:39 despite 21 admitted sessions and work during 10:53–11:31. The first M3 dispatch occurred later, 14:05:10, proving that waiting for another review is insufficient.

Proposal for hub review: use the existing last-finalized-boundary fallback allowed by Done when, rather than inventing a prose-derived supersession timestamp. Before producing fresh cuts, omit unfinalized named boundaries absent from the exact code pin's canonical issue Plan. Keep finalized boundaries, current executable boundaries and whole-issue rounds. Report the retired boundaries and effective selected phase range. Missing/ambiguous required Plan evidence refuses with an exact restore/commit/retry action; absence is never an empty authority set. Saved snapshots continue replaying their exact original Cuts, without applying today's Plan. ARCH-DRY/PURE: one pure selector before both fresh consumers, no duplicated timeline implementation. No code or test implementation has started; awaiting hub scope approval.

Hub approved the full-flow membership/fallback policy and isolated production-path proof (authenticated submitted delivery a5d4ada9-9b0c-47c0-836b-6fafdb35a465). No supersession event or schema change; proceed through change-code.

Plan-quality round 1 raised PQ-1 (function-level test contract). Named selector, pinned resolver, phase diagnostic and production compute/replay functions; compressed the executable-case inventories into adversarial strategies and mechanical guards. Policy unchanged; re-running admission.

Plan-quality round 2 CLEAN: PQ-1 addressed; no remaining findings. Estimate derived only after acceptance: 1.38h, provisional stale calibration. Baseline pure timeline/boundary/conservation tests PASS.
