---
id: 000492
status: codecomplete
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '8b8f4f61b09cc2c8f911104d3e3a6809586571db' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T08:41:58-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "27c81229", done: "fd8ecfc8"}
actual_hours: 0.38
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

- [x] Add one caller-root selector to localRunDeps: WorkflowAnchor when present, otherwise Root. Use it for SuiteAttempt.Slot, managed suite identity, cancellation lease ownership and WF_TEST_LOCK_OWNER; retain Root for train content/helper lookup and exact-head priority evidence.
- [x] RED/GREEN a real two-car train whose disposable tree differs from its conductor. Read persisted attempts and prove the conductor slot survives, including retry/baseline/rerun paths through the shared runner. Feed a produced attempt into the actual measurement regression: an overlapping foreign-slot issue measures without queue-provenance refusal; a missing-slot control still refuses.
- [x] Verify ordinary local checks retain their owner; cover real queue/lease/slot diagnostic boundaries and existing train instance/helper controls. Update atlas. Delivery gates remain the recorded suite and central close/landing; a passing record precedes merge.

ARCH-DRY: reuse WorkflowAnchor, already the verified checkout the train was made from, rather than add a second owner field or recover ownership from a temporary pathname/environment. ARCH-PURPOSE: sweep all lifetime/provenance writers using that root, not only the final JSON field. Root stays the train tree for content checks (#490) and helper resolution (#486). ARCH-EVOLVE: populate the existing slot field; no schema/version change. Train identity remains in existing landing-line train records; SuiteAttempt has no train-id field to reuse.

Alternatives: only changing the final slot lookup leaves queue/cancellation ownership wrong; adding another owner field duplicates the existing caller anchor. The root selector is O(1), adds no scans/processes, and dies with run dependencies; this batch completion path is not a new per-ticket hot loop. Existing suite/queue bounds remain unchanged.

Actions: run suite, restore evidence. Valid conductor identity produces the existing attributable record (run suite); a truly unresolved identity retains the existing unknown-slot measurement refusal (restore evidence). Historical missing-slot records are not guessed or rewritten; no new visible state or loss is introduced.

## Log

### 2026-10-20
- 2026-10-20: closed — Real two-car conductor writes its slot and caller-root queue/cancellation/holder identity; actual CLI measures the overlapping foreign-slot issue and still refuses missing-slot evidence. Removing the selector fails all four baseline/prefix attempts; restored targeted matrix passes in 295.135s and tracked Go lint passes. Recorded suite on the pushed implementation is queued; its passing record remains required before landing. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

### 2026-10-20 — claim and root-cause plan

Claimed immediately after #491 landed. The train deliberately changes localRunDeps.Root to its detached content tree and retains the original conductor in WorkflowAnchor; SuiteAttempt, queue enrollment, run cancellation lease and capacity-holder diagnostics nevertheless used Root. Reuse the existing anchor at those identity boundaries. Content/root and priority semantics stay unchanged. Quick-flow-sized one-selector fix, with real producer and measurement controls before code; no staging or live machine pause.

### 2026-10-20 — implementation and targeted evidence

The real train regression first returned an empty saved slot, temporary-tree queue/lease/holder ownership, and the exact actual CLI refusal "unknown slot provenance". After the one callerRoot selector was applied to those four identity writers, it passes; the missing-slot control still refuses. An initial measurement fixture used a local remote identity, so its records could not join GitHub suite evidence; the fixture now delegates identity reads to real Git and maps only GitHub transport to its existing private bare origin. All records pass the production decoder before reaching actuals. A queued second-car fixture keeps this ownership test independent of the second merge's cleanup timeout observed in #500.

A primary-checkout control also failed because relative Git common directories were anchored to the disposable content tree; the same selector fixes that path. The train-splitting control now asserts all prefix retries and its diagnostic baseline carry the conductor slot. Removing the selector made all four persisted attempts fail that assertion; restoring it passes. Ordinary local-check, real-lock baseline/reentry/rerun, cancellation, priority, queue lifetime, train helper/content, known-flake and verb-context controls all pass together (295.135s). The tracked Go lint catalog passes (0.535s). No new Go lint rule: caller versus content ownership is a semantic path boundary, enforced by the real writer/CLI tests rather than a lexical name rule.

ARCH-PURE: the selector is pure; existing IO resolves workspace identity. ARCH-MOCK: GitHub/capacity use existing private stateful seams, with real Git and the real actual CLI consumer. ARCH-CONSTRAINTS: constant-time selection replaces arguments without extra processes/scans or changed suite bounds. ARCH-SECURE: no environment/path-name inference, and unknown-slot evidence still refuses. ARCH-ORDER: immutable run dependencies add no state between events. ARCH-FUNERAL: no new artifact family, retention or lifetime; normal attempt and temporary-fixture cleanup remain in charge. ARCH-DRY/PURPOSE/EVOLVE are described in Plan.

## Revisions

### 2026-10-20 — complete the caller/content boundary

The root-use sweep found relative Git-common-directory paths also anchored to content Root even though the Git seam runs in the caller checkout. Resolve them through the same callerRoot selector, with a real primary-checkout regression. This adds no mechanism, field or format. The Plan's delivery suffix originally said "run the recorded suite and close/land with central sdlc"; those remain mandatory gates, outside the completed implementation checklist, so close review can overlap a queued recorded suite under tl-worker's rule.
