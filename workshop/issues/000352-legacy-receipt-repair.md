---
id: 000352
status: open
deps: [ariadne#337]
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '38d018da85e5f880cf00f7a8bcf3fab96534369a' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T17:10:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Repair invalid legacy transfer receipts safely

## Problem

Issue assets (#337) reject a persisted transfer candidate whose body index or
asset bytes are invalid. A receipt prepared before that validation existed is
immutable: editing the source body does not repair its pinned candidate. Recovery
currently has list/reconcile, but no safe transfer restart procedure.

Priority: low. A read-only fleet inventory on 2026-10-14T00:09:36Z found **zero
recovery receipts**, including zero pre-feature transfer receipts, across 68
checkouts and 44 unique Git repositories on this machine, with zero scan errors.
The defect is reproduced by production recovery tests in #337; there is no
measured local backlog requiring repair.

## Spec

Provide a supported repair path for the exact legacy receipt named by #337's
refusal. Derive safety from remote publication evidence: distinguish prepared,
uncertain-not-landed and already-landed candidates. Preserve source evidence and
never delete a receipt that may guard published or outstanding work. The operator
must not edit raw receipt JSON or guess which Git ref to discard.

## Done when

- A concrete supported command/procedure repairs or safely supersedes invalid
  legacy transfers while preserving source and remote work.
- Tests cover prepared, uncertain-not-landed, already-landed, repeated repair,
  foreign ownership and concurrent remote changes.
- Refusals in #337 can point directly to that repair path; documentation names
  its inspection, proof and recovery steps.

## Plan

- [ ] Design and implement the safe legacy-transfer repair after #337 lands.

## Log

### 2026-10-13

- Filed unassigned at hub direction (delivery d2bc4634-77fa-4a4e-a9d0-d9ce372e0afa).
  Measurement used `sdlc fleet inventory --json`, deduplicated its Git repository
  identities, and read each repository's `refs/sdlc/recovery/` namespace. No
  checkout or receipt was changed. Local evidence:
  /tmp/ariadne337-receipt-inventory-report.json.
