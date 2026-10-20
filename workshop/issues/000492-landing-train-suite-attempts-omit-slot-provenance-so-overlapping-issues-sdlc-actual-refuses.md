---
id: 000492
status: working
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3f9b63993f092a63d49c3f1e19a1d29f8101d8cf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T08:41:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# Landing-train suite attempts omit slot provenance, so overlapping issues' sdlc actual refuses

## Problem

Landing-train suite attempts are saved without a `slot` field. pair's train 174 suite (attempt
f6a4337408d36adfd9d00ddb9d298d99, head c2c6b927, conductor pair:7) was recorded as version 2 with no `slot`, so
`sdlc actual` for an unrelated issue whose window overlapped it (pair#725) refused: "suite attempt … has unknown
slot provenance … ask the TL/hub to restore the named measurement evidence". The TL restored it by hand on 10-20
(added `"slot": "pair:7"`; original kept). Every train will do this to every overlapping issue's actuals.

## Spec

The train conductor records its suite attempt with `slot` (the conductor's workspace) like any local-check suite,
plus whatever identifies it as a train run (train id) if the schema allows, so measurement attributes it correctly.

## Done when

- A train-suite attempt written by a fixture train carries `slot` (test), and `sdlc actual` for an overlapping issue
  measures without refusal (test).

## Plan

- [ ] Add one caller-root selector to localRunDeps: WorkflowAnchor when present, otherwise Root. Use it for SuiteAttempt.Slot, managed suite identity, cancellation lease ownership and WF_TEST_LOCK_OWNER; retain Root for train content/helper lookup and exact-head priority evidence.
- [ ] RED/GREEN a real two-car train whose disposable tree differs from its conductor. Read persisted attempts and prove the conductor slot survives, including retry/baseline/rerun paths through the shared runner. Feed a produced attempt into the actual measurement regression: an overlapping foreign-slot issue measures without queue-provenance refusal; a missing-slot control still refuses.
- [ ] Verify ordinary local checks retain their owner; cover real queue/lease/slot diagnostic boundaries and existing train instance/helper controls. Update atlas, run the recorded suite and close/land with central sdlc.

ARCH-DRY: reuse WorkflowAnchor, already the verified checkout the train was made from, rather than add a second owner field or recover ownership from a temporary pathname/environment. ARCH-PURPOSE: sweep all lifetime/provenance writers using that root, not only the final JSON field. Root stays the train tree for content checks (#490) and helper resolution (#486). ARCH-EVOLVE: populate the existing slot field; no schema/version change. Train identity remains in existing landing-line train records; SuiteAttempt has no train-id field to reuse.

Alternatives: only changing the final slot lookup leaves queue/cancellation ownership wrong; adding another owner field duplicates the existing caller anchor. The root selector is O(1), adds no scans/processes, and dies with run dependencies; this batch completion path is not a new per-ticket hot loop. Existing suite/queue bounds remain unchanged.

Actions: run suite, restore evidence. Valid conductor identity produces the existing attributable record (run suite); a truly unresolved identity retains the existing unknown-slot measurement refusal (restore evidence). Historical missing-slot records are not guessed or rewritten; no new visible state or loss is introduced.

## Log

### 2026-10-20

### 2026-10-20 — claim and root-cause plan

Claimed immediately after #491 landed. The train deliberately changes localRunDeps.Root to its detached content tree and retains the original conductor in WorkflowAnchor; SuiteAttempt, queue enrollment, run cancellation lease and capacity-holder diagnostics nevertheless used Root. Reuse the existing anchor at those identity boundaries. Content/root and priority semantics stay unchanged. Quick-flow-sized one-selector fix, with real producer and measurement controls before code; no staging or live machine pause.
