---
id: 000544
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9cdf509175715c8768fbf00884a640e65c56be42' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T00:02:32-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# issue migrate drops a legacy close's completion record; later issues' actual measurement then refuses (echo#1 → echo#3)

## Problem

Tracker cutover preserves adopted numbers on archived issues but drops the evidence that bounds them. The conservative exclusion reader consequently blocks later issues, including echo#3.

## Spec

From ops-echo:0 (10-21 evening). echo#1 closed under the legacy workflow (merge 08124b3, archive 112fd43,
actual_hours 0.80) just before echo migrated to the tracker (`sdlc issue migrate --apply`, 6f0fe49). After the
migration, `sdlc issue show 1` reads `status done, completion: no close recorded`. echo#3's close now stops before
review: `sdlc actual --issue 3` refuses "cannot auto-compute (issue #1 adopted actual has no trustworthy saved
window: completion is missing)", from the adopted-actual exclusion (`cmd/sdlc/actualexclusion.go`, `adoptionRefuse`
with no saved window).

So the migration carries a legacy-closed issue's status and `actual_hours`, but not its close binding or saved
measurement window, and every later issue whose window overlaps it is blocked from measuring. Every legacy repo that
migrates with recently closed issues hits this.

Fix: `sdlc issue migrate` carries each legacy close's completion record (the close commit and saved measurement
window) onto the tracker. For already-migrated repos, `--reconcile` (or a narrow repair verb) backfills it from the
legacy close commit and archive, with the same evidence the legacy close used. Never invent a window: if none is
recoverable, mark the adopted actual as windowless-legacy, and let the exclusion treat a windowless issue whose
close predates the candidate's engagement as disjoint (skip), not refuse.

## Done when

- A born-archived legacy import with no close remains visibly unbound and warns; proven archive/done provenance strictly before the earliest new engagement permits later actual measurement, while overlap remains strict (test).

- Migrating a legacy fixture with a closed issue keeps its completion record; `sdlc issue show` shows the close (test).
- In an already-migrated fixture, the repair restores it, and a later issue's `sdlc actual` measures (test).
- A legacy-closed issue with no recoverable window, closed before the candidate's claim, doesn't block measurement
  (test); an overlapping one still refuses.

## Plan

Durable plan: [legacy completion preservation and repair](../plans/000544-legacy-completion-plan.md).

- [ ] Recover and import verified legacy close bindings, retaining exact saved measurement inputs or explicit windowless provenance.
- [ ] Add a digest-reviewed single-card repair for repositories already migrated, preserving current state and supporting uncertain-push retry.
- [ ] Use proven windowless legacy close ordering for disjointness, with overlapping/unknown evidence still refusing.
- [ ] Verify production-path migration/repair/actual controls, full suite, approved Echo conformance and documentation; close and land.

## Log

### 2026-10-21

### 2026-10-22

- Claimed and started planning from fresh main after #540 landed. Read current worker/SDLC contracts and issue vocabulary. Parallel read-only investigations confirmed migrationInventory and PlanTrackerMigration bind only active codecomplete records; archived done uses ArchivedCard without completion. Existing --reconcile converts branch details and does not repair this loss.
- Inspected the issue-referenced Echo evidence read-only after AGENTS.local.md (no MEMORY.md present). Git verified latest close 476b6b12e39e87ed1febcda2a4b7ef94c8f4ec73, archive 112fd43, cutover 6f0fe498956349b2c3318000f524bf93920b3b98 and tracker root 601c7b41a7d40e2adaaea4fa1d70b8798aedfa70. The imported #1 card is done with actual0.80 and no completion. The close has SHIP/Review-Window trailers and an issue1 ledger whose final reviewed head agrees; it contains no retained ActualInputs. This is a real windowless legacy case, not permission to synthesize a measurement window. Echo checkout contains unrelated feature/untracked work; none was changed.
- ARCH-DRY/PURE: use one immutable legacy-evidence resolver for cutover and repair, then feed proven earlier windowless closes into the existing windowDisjoint decision. Candidate cutoff is Snapshot.Since, the earliest admitted engagement, never the latest claim or mutable started card field. ARCH-EVOLVE: add explicit legacy provenance; malformed modern records and damaged saved inputs remain strict. Plan is awaiting TL approval; no code or Echo state changed.

- TL approved the plan and exact repair command; ops-echo:0 owns the real Echo dry-run/apply/repeat/actual3 conformance in a temporary worktree. Entering change-code, no implementation yet. Approval deliveries d46eb30c-7e6c-4da2-9368-129e6b4cdb7c and 904d435d-de8b-446e-8fea-489c1c90ddbe.

- Plan-quality round1 PQ-1 requests named function-level test strategies and compressed prose inventories. Refined that contract without changing approved mechanisms. Existing born-archived fixture records have no actual historical close; sent the hub the narrow compatibility distinction before code.

## Revisions

### 2026-10-22 — unbound terminal imports

TL delivery 82679d7d-74d5-4320-9051-4232bdd6f97f extends the same strictly-earlier rule to unbound legacy imports using a proven archive/done commit. Preserve absence of a close, never fabricate review/window evidence, and keep overlapping cases strict. The plan now uses one LegacyAdoption envelope record across close-backed and terminal-only imports.
