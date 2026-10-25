---
id: 000624
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: '675f3cf710a95d0bdb9062e8aa5b6a5112af1dbb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T03:07:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "e57bf477", done: "dc5eb85e"}
---

# fleet inventory --json reports merge intents (producer for pair#848)

## Problem

## Spec

**`sdlc fleet inventory --json` reports merge intents (producer for pair#848).** After a restart, Couch's recovery plan
must re-enter each landing a restart interrupted, once. sdlc already owns merge-intent records (validated v1 files,
lock-based liveness, exact re-entry command construction), but `fleet inventory` doesn't expose them, so Couch would
have to read sdlc's state files itself. pair#848's plan (pair `workshop/plans/000848-orphaned-merge-recovery-plan.md`,
TL-approved 10-25) makes inventory the single source.

Add per-row `merge_intents[]`: for each intent, its state (live, orphaned, terminal, stale), the slot and exact member
root it belongs to (dependencies included), its physical identity (for de-duplication across worktrees sharing one
Git common dir; group shared stores once), the exact re-entry command, and a carried `note` (e.g. queue paused). Each
row also states read quality (complete, partial, unreadable), and a corrupt record without a trustworthy root appears
once as a repository-level inspection diagnostic. Additive JSON fields only; an older Couch ignores them.

Transition: no intent format change; the new reader reads previous-build v1 files through the existing preserving
codec, pinned by legacy fixtures. A capability field lets Couch tell "no merge intents" from "this sdlc can't report
them" (it must then say upgrade and rerun, never claim none).

## Done when

- Inventory JSON carries merge intents with state, root, identity, re-entry command, note and read quality; tests
  over a stateful fake for live, orphaned, terminal, paused, unreadable, shared-store duplicates and corrupt-root
  records, plus legacy v1 fixtures.
- One inventory call per fleet; bounded output; the capability field is present and documented in `--help`.

## Plan

- [ ] Implement and verify the shared report, grouped read-only inventory adapter and additive wire contract in [the durable plan](../plans/000624-fleet-merge-intent-inventory-plan.md), after hub approval and change-code.
- [ ] Qualify private real-store/executable conformance, recorded pre-close check and close review; land and hand the final producer contract to pair:5.

## Log

### 2026-10-25

Claimed and entered start-plan from main after #617 landed (Git merge954c8b922af48d0fcc1a5bc148ccefb57be73c2d). No retained check or merge runner remains. Read the existing fleet collection, common-directory identity, slot/dependency rows, state/landing shared intent observer and preserving lock-based v1 store. Read pair#848's approved plan through its issue branch without changing that peer checkout; pair:5 is working elsewhere. Its consumer remains parked until this producer lands.

Pair:5 authenticated receipt bba6d841-4e31-400d-a24a-71d4b4648b3e confirms capabilities.merge_intents=1; row quality complete/partial/unreadable with error; exact canonical record identity and Root-only association; complete+[] after terminal removal, not new tombstones; malformed/legacy/mixed-store controls. The existing observer silently skips non-ENOENT open errors and drops earlier observations on a later branch error: the plan closes that uncertainty class through the same reader, rather than a second classifier (ARCH-DRY/ARCH-SECURE). Group by the inventory's existing physical common-dir identity, not a publication URL or branch name.

Durable plan drafted and checkpointed for ops:0 review before change-code. It includes the visible-state table, one shared classifier/command owner, one store read per group, bounded1MiB JSON matching the consumer, counted8/16 growth controls, preserving v1 fixtures and private real-lock conformance. Proposed rollout is private executable proof plus the hub's normal post-land primary rebuild; no live staging, pause, sweep or merge-recovery execution. No implementation or test edits and no new runtime state/schema.

## Revisions

### 2026-10-25 — Clarify terminal observation and freeze consumer vocabulary

The existing merge-intent lifecycle removes terminal records; inventory reports a completed empty read rather than inventing a terminal state or tombstone. Pair:5 agreed. Keep schema_version1 and unchanged v1 intents, with an explicit capability distinguishing an old unsupported producer from complete empty evidence. The new row-quality vocabulary describes enumeration completeness independently of each record's live/orphaned/stale/unreadable state.

### 2026-10-25 — Hub approved implementation

Authenticated ops:0 receipt565af516-d8bc-4ef8-b8c8-46042ce3b16e approves the durable plan and explicitly includes the existing reader's lost-partial-result bug. Its non-ENOENT failure and partial-observation controls are required tests. Private executable/store qualification plus the hub's post-land primary rebuild is approved; no live staging. Proceed to change-code before implementation.
