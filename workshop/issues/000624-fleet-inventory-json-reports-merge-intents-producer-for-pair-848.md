---
id: 000624
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'e428c6050171c72cc29a914a6fe9c11b940b63c7' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-25
