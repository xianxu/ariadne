---
id: 000548
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.93
severity: critical
collaboration-mode: collaboration
card_mirror: '167331c577222f4c9e0a65d5fa45e9c54454c42b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T01:34:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Actual queue evidence must belong to the measured issue head

## Problem

Echo#9 actual preview refuses because another issue's recorded suite attempt lies in its measurement period while echo:0 holds more than one claim. The suite attempt belongs to echo#3, not echo#9. This blocks echo#9's fix and therefore echo#10, #3 and #5 suites/landings.

## Spec

A suite attempt may affect actual measurement only for the issue whose candidate head it ran on. Do not adopt or demand engagement proof for another issue's attempt merely because repository, slot and clock overlap. Preserve strict refusal for a selected issue's missing or ambiguous pinned ownership; do not drop uncertain evidence that really belongs to it.

## Done when

- A real fixture holds two issues in one slot, records a suite for issue A, and measures issue B without adopting or refusing on A's attempt.
- A's actual still uses its own exact recorded queue endpoints, and missing/ambiguous ownership for a relevant attempt still refuses.
- Head identity and issue lineage, rather than current branch-name guesses or timestamp overlap alone, decide relevance.
- Echo#9 actual preview passes the owner-run live check before adopting hours.

## Plan

Durable plan: [suite head attribution](../plans/000548-suite-head-attribution-plan.md).

- [ ] Reproduce and enforce immutable issue-head membership before queue attribution.
- [ ] Verify the owned/foreign/unknown and train controls, owner-run Echo preview, suite, close and landing.

## Log

- 2026-10-22: #544 landed PR388 with integrated2198-test/76-package suite PASS and close round3 SHIP. Imported the plan-only branch from its independent clone into the canonical slot; start-plan refreshed the resting checkout to Git-verified origin/main9b6896e4123f0a8d2790b6ec3b6f3a54c0d6a792, then selected this existing issue branch. Clean main integration is Git-verified393e3bc60df22840f26bfe9a887e4aa266b4b96e; it contains only landed #544 code, not a stack of unlanded work. Added exact pure/IO function contracts to the approved plan. The bounded graph reader and strict parser exceed the100-line production shell, so implementation uses the full gates. No implementation code changed yet; plan-quality comes before estimating.

- 2026-10-22: TL approved the durable plan (delivery c4635b43-eb9c-41eb-9592-c4174b66fe93): immutable tested-head membership precedes slot/time ownership; reuse local-lead grammar, follow first-parent neutral integration/metadata only, honor exact train composition, skip proven foreign runs, keep unknown/unreadable strict, and add no persisted fields. Implementation starts after #544 lands. ops-echo:0 owns read-only actual9 and actual3 live previews.
- Started planning from Git-verified main909bf77db0b84bfe73a4b494bec775c40d47b78e in a separate independent ordinary clone, as TL authorized while #544 queues. #544's runner remains intact in the canonical slot. Workspace gate refused a linked dependency worktree; the untouched worktree was removed and the required independent clone passed the gate. No identity override used.
- The exact offending attempt SHA fc679c0745bbdd09896163978d3f24d7516b4564 leads with #3; selectOwnedQueueAttempts does not inspect SHA at all. A slot/time-only selector is insufficient when two issue claims coexist. Design uses shared local-lead grammar plus bounded immutable lineage proof before existing ownership selection. No code changed; awaiting TL plan review.

2026-10-22: Discovered in #544 live conformance; ops-echo:0 delivery f8174005-af47-42ac-af83-c57928432164. Read-only actual9 at its pushed branch head on clean #544 binary b1ea682 refused: suite attempt fcf15650129203dfd1408728d625962b has 0 matching pinned slot engagements. Echo owner identifies it as echo#3's failed recorded run. echo:0 had kept #3 blocked at close and taken #9 per ops:0, so both claims coexist legitimately. The #544 repair and mixed legacy/tracker measurement for #3 passed independently. TL delivery59ea794e-0b24-4d8f-85cf-c1b571d0117a keeps this outside #544; delivery7ebee200-0aa0-49ac-9606-ee9ac3af6f02 assigns it to ariadne:1 immediately after #544 lands, plan first, with parallel planning from main allowed while #544 queues.

- 2026-10-22: Plan-quality round1 blocked on replacement-sensitive immutable reads and an unspecified aggregate resource envelope. Revised the narrow reader contract to disable Git replacements and specify command-global node/byte plus per-lineage depth ceilings, with boundary tests through the production walker; compressed the repeated test inventory. No implementation changed.

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only; calibration is flagged stale by estimate-source. Two single-concern greenfield modules cover pure membership and the memoized walker; bounded Git IO and existing measurement wiring are smaller extensions. Git's cat-file protocol and existing bounded runner remove novel-stack discovery; no new library is needed. Thorough approved plan applies ×0.2 design discount and15% design buffer; implementation values are already40% of v2 ranges, familiarity1.0.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.20 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.03 impl=0.20
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.02 impl=0.08
item: milestone-review design=0.02 impl=0.20
design-buffer: 0.15
total: 1.93
```
