---
id: 000548
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.93
severity: critical
collaboration-mode: collaboration
card_mirror: '7effeaab18994f0514495708c737b1efdba40b2d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T01:34:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.54
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

- [x] Reproduce and enforce immutable issue-head membership before queue attribution.
- [ ] Verify the owned/foreign/unknown and train controls, owner-run Echo preview, suite, close and landing.

## Log


- 2026-10-22: closed — BR-1 corrected both filter-location plan tables and recorded split revision; no code delta. Real foreign/own/unknown tracker and train controls pass; owned/hybrid targets81.814s, gitx and full golint pass. Echo owner actual9 passes0.76h and actual3 unchanged0.50h. Recorded suite lease-191091060 alive first in line; permitted review overlap, no-plan-check acknowledges only suite/landing release gates pending. Final recorded pass precedes landing; hub owns project. TL approved branch binary to adopt corrected inputs for this fix itself.; review verdict: SHIP
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

- 2026-10-22: Full change-code passed plan-quality and estimate-quality (INFO). The module budgets include their adversarial tests: parser/train controls, reader replacement/framing controls, walker resource/growth tests, and real tracker measurement fixture. The final review budget includes live preview coordination and landing verification; machine queue and suite idle are not counted twice as implementation work.
- 2026-10-22: Real Git/tracker two-claim fixture reproduced the exact old failure before wiring: foreign failed suite fcf15650129203dfd1408728d625962b reported0 matching pinned slot engagements. Its own crossing wait still refused and its own exact60s wait remained admitted. Added immutable head classification before the shared owned selector, preserving #544 suffix clipping and saved replay. Integration receives the canonical issue ID explicitly; no branch-name inference or persisted fields.
- 2026-10-22: Bounded reader real replacement-ref, missing/noncommit/cancel and framing checks passed. Stateful graph tests count each node once at8/16 heads with shared tails, test aggregate limits and a real Git train/reintegration excluding main/base cars. An in-session resource audit found depth was checked after discovery and neutral memberships copied per descendant; behavioral red tests measured1000 reads past the256-edge ceiling and unwanted membership copies, then fixes bounded discovery and shared immutable neutral membership. No producer change or landingtrain.go edit.
- 2026-10-22: Combined targeted behavioral tests passed, but the real-repository hermeticity guard correctly refused the run because this session changed the dirty-path set during execution. No test defect or passing suite evidence is claimed from that run; freeze/stage all edits and repeat on the final source before the recorded suite.

- 2026-10-22: Frozen-source final targeted run passed: cmd/sdlc81.814s (membership, resource limits, real tracker foreign/own/unknown, existing owned actual/correction and hybrid-cutover controls), gitx0.661s; separate full pkg/golint test pass0.963s. Diff checks clean. Next: clean branch build to Echo owner for actual9/actual3 previews and recorded full suite on the pushed head.

- 2026-10-22: Owner-run Echo live conformance PASS (ops-echo:0 delivery c7f2abfd-3287-4a0e-8186-10349d89234e), read-only at pushed issue heads on clean ebf48d5e4feab1b94abce80c2241189908189a54. actual9 at dc91d5b exits0 and measures0.76h, legacy-prefix/owned-queue provenance,98.95 raw/excluded minutes across2 own waits. actual3 at a40d090 exits0 and remains0.50h,1.13 raw/1.12 excluded minutes across1 own wait. Existing source/mention warnings remain visible. No Echo close or adoption performed by this worker.
- 2026-10-22: PR393 opened. Recorded suite lease-191091060 on ebf48d5e is alive, queued for both slots (initial position6); no running suite interrupted and no duplicate runner. Close review overlaps the queued suite under the worker contract; the unchecked verification row now contains only suite completion and normal landing. The precise no-plan-check acknowledgment is for this release gate, not unfinished implementation. A recorded pass on final code remains mandatory before landing; code fixes require fresh checks.

- 2026-10-22: Close round1 REWORK raised one traceability finding BR-1: both plan tables named the head-parser file for the pure attempt filter, which was split into actualqueue_filter.go. Corrected both tables, added a file-split Revisions entry, and audited named locations. Review reported no behavior/architecture defects; no production code changed, so the queued suite remains valid for final code.
- 2026-10-22: The installed fleet binary still applies pre-fix queue selection to this fix's own actual. Same pinned implementation head preview measured1.02h installed versus1.40h branch because #544's foreign wait was wrongly subtracted; interrupted the installed close before publication (card remained working, no completion) and reran with clean branch bin/sdlc, which retained corrected inputs and measured1.44h at the later log edge. TL explicitly approved this use (delivery4cd53738-e8b5-4231-9536-53fd64448623). All normal gates remain enabled except the documented suite-overlap plan acknowledgment and TL-owned project. TL granted #548 one-hour priority; the same live suite is now first, with no running suite interrupted.
