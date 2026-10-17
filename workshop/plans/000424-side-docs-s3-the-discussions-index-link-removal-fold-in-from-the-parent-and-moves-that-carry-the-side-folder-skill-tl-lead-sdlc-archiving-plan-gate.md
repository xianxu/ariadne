---
reviewer_runs:
    - version: 1
      intent:
        id: 6dbda443-fe99-4a31-b436-4e65e7ba12a0
        repository: github.com/xianxu/ariadne
        issue_id: "000424"
        recipe: estimate-quality
        round: 1
        attempt: 1
        agent: claude
        root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        reviewed_head: 07f53a1160fbc289e86fd32883c6199db3a1ba15
        slot: ariadne:2
        requested_at: "2026-10-17T10:12:36.060045Z"
      phase: finished
      pid: 54406
      started_at: "2026-10-17T10:12:36.23607Z"
      finished_at: "2026-10-17T10:12:55.604017Z"
      session_id: 31cb4c5b-54e7-4751-84a4-4156bff95a5d
      outcome: returned
gate: plan-quality
issue: 424
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: a48e0915-2583-4c57-9b76-b6783ec1a163
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: e37159dcf6b5083cd10ac8a60e2ffe350c4f6c54
            slot: ariadne:2
            requested_at: "2026-10-17T10:08:48.549012Z"
          phase: finished
          pid: 32026
          started_at: "2026-10-17T10:08:48.642489Z"
          finished_at: "2026-10-17T10:09:41.352909Z"
          session_id: b39752e8-b0b1-4ce8-851a-e4eda5026d9d
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T03:09:41-07:00"
      agent: claude
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: M3 relies on a rollback that doesn't exist, misses interrupted-archive recovery and the trunk retry branches; it should reuse the asset archive stages
          detail: merge.go:880-912 renames one file at a time and returns partial moves on error (no rollback). push.go preparedArchiveMoves (:452-548) puts any .side path in other, so resuming an interrupted archive refuses. archivemain.go:63-110 (retry when live details are gone) and :80-102 (successor episode) need a rule for an already-archived .side tree. archiveassets.go (prepare/finish/recover/verify, plus assetFilesFromView) already solves this; reuse it (ARCH-DRY), list every archive path, and test a resumed interrupted archive (ARCH-ORDER).
          family: archive-move-reuse-asset-pipeline
          round: 1
        - kind: defect
          id: PQ-2
          severity: Minor
          title: Tasks 1, 3, 4 and 6 list test cases one by one instead of one strategy line per risky function
          detail: Compress to the function name plus the adversarial input class and guard, e.g. ReconcileIndex over hand-edited frontmatter, fuzz-seeded, refuse and leave the input unchanged.
          family: plan-test-prose-enumeration
          round: 1
        - kind: defect
          id: PQ-3
          severity: Minor
          title: The export bump to version 2 depends on parley.nvim#373, but frontmatter deps is empty
          detail: 'If parley syncs before #373 lands, its sync script refuses version 2. Declare the dependency or state which repository lands first.'
          family: undeclared-cross-repo-dep
          round: 1
        - kind: unverifiable
          id: PQ-4
          severity: Minor
          title: The claim that Couch already supports write --all can't be checked from ariadne
          detail: tl-lead only calls read --all today (sessionsync.go:152). M2's Step 3 conformance test should confirm write --all with --if-hash before the fake models it.
          family: unbacked-existing-behavior-claim
          round: 1
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            id: 25d4a28b-306c-44c9-aae0-9d36a6652239
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: plan-quality
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 07f53a1160fbc289e86fd32883c6199db3a1ba15
            slot: ariadne:2
            requested_at: "2026-10-17T10:10:58.108343Z"
          phase: finished
          pid: 44709
          started_at: "2026-10-17T10:10:58.206667Z"
          finished_at: "2026-10-17T10:11:26.002772Z"
          session_id: 0774b98f-c3cc-48a3-9410-19aaf0bd754b
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T03:11:26-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: M3 now reuses the asset archive stages (prepare, shared finisher, recover, verify), lists paths (a)-(f), and tests a resumed interrupted archive; I checked the function names against archiveassets.go and push.go.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: The test steps are now one strategy line per risk; Task 3's lines describe fixtures, not test cases.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: 'The landing order (ariadne first; parley refuses version 2 and keeps its snapshot until #373) is stated in Scope.'
          round: 2
        - id: PQ-4
          disposition: addressed
          note: M2's conformance step proves write --all with --if-hash against a live Couch before the fake models it.
          round: 2
      findings:
        - kind: defect
          id: PQ-5
          severity: Minor
          title: Task 6 never builds the tl-lead status index subcommand that the M2 row and help.md name
          detail: The issue's M2 row and Task 6 Step 4 name a status index command, but Step 2 wires only instructions link and status submit. Also, the Done-when says three archive paths while the plan lists six, (a)-(f). Align both.
          family: plan-step-surface-mismatch
          round: 2
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            id: 7bfc06fa-7e65-492a-bdad-a0827e28b377
            repository: github.com/xianxu/ariadne
            issue_id: "000424"
            recipe: plan-quality
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 07f53a1160fbc289e86fd32883c6199db3a1ba15
            slot: ariadne:2
            requested_at: "2026-10-17T10:12:16.237066Z"
          phase: finished
          pid: 53106
          started_at: "2026-10-17T10:12:16.400356Z"
          finished_at: "2026-10-17T10:12:34.916804Z"
          session_id: 049ae318-4ca4-4916-8e01-864f55b8fd05
          outcome: returned
      "n": 3
      timestamp: "2026-10-17T03:12:35-07:00"
      agent: claude
      dispose:
        - id: PQ-5
          disposition: addressed
          note: Task 6 Step 2 now builds `status index`; "three archive paths" (merge, push, trunk) and their recovery and retry branches (a) to (f) agree.
          round: 3
      blocked: false
content_hash: 4cefaf506b7bb07f5067a192562b3bb5e3e3c03a89e0f52323c7b55b089ac282
---

# Gate ledger — ariadne#424 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T03:09:41-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `archive-move-reuse-asset-pipeline` (kind: defect) M3 relies on a rollback that doesn't exist, misses interrupted-archive recovery and the trunk retry branches; it should reuse the asset archive stages
  merge.go:880-912 renames one file at a time and returns partial moves on error (no rollback). push.go preparedArchiveMoves (:452-548) puts any .side path in other, so resuming an interrupted archive refuses. archivemain.go:63-110 (retry when live details are gone) and :80-102 (successor episode) need a rule for an already-archived .side tree. archiveassets.go (prepare/finish/recover/verify, plus assetFilesFromView) already solves this; reuse it (ARCH-DRY), list every archive path, and test a resumed interrupted archive (ARCH-ORDER).
- **PQ-2** [Minor] `plan-test-prose-enumeration` (kind: defect) Tasks 1, 3, 4 and 6 list test cases one by one instead of one strategy line per risky function
  Compress to the function name plus the adversarial input class and guard, e.g. ReconcileIndex over hand-edited frontmatter, fuzz-seeded, refuse and leave the input unchanged.
- **PQ-3** [Minor] `undeclared-cross-repo-dep` (kind: defect) The export bump to version 2 depends on parley.nvim#373, but frontmatter deps is empty
  If parley syncs before #373 lands, its sync script refuses version 2. Declare the dependency or state which repository lands first.
- **PQ-4** [Minor] `unbacked-existing-behavior-claim` (kind: unverifiable) The claim that Couch already supports write --all can't be checked from ariadne
  tl-lead only calls read --all today (sessionsync.go:152). M2's Step 3 conformance test should confirm write --all with --if-hash before the fake models it.

## Round 2 — 2026-10-17T03:11:26-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — M3 now reuses the asset archive stages (prepare, shared finisher, recover, verify), lists paths (a)-(f), and tests a resumed interrupted archive; I checked the function names against archiveassets.go and push.go.
- PQ-2 — addressed — The test steps are now one strategy line per risk; Task 3's lines describe fixtures, not test cases.
- PQ-3 — addressed — The landing order (ariadne first; parley refuses version 2 and keeps its snapshot until #373) is stated in Scope.
- PQ-4 — addressed — M2's conformance step proves write --all with --if-hash against a live Couch before the fake models it.

### Raised

- **PQ-5** [Minor] `plan-step-surface-mismatch` (kind: defect) Task 6 never builds the tl-lead status index subcommand that the M2 row and help.md name
  The issue's M2 row and Task 6 Step 4 name a status index command, but Step 2 wires only instructions link and status submit. Also, the Done-when says three archive paths while the plan lists six, (a)-(f). Align both.

## Round 3 — 2026-10-17T03:12:35-07:00 (claude) — passed

### Disposed

- PQ-5 — addressed — Task 6 Step 2 now builds `status index`; "three archive paths" (merge, push, trunk) and their recovery and retry branches (a) to (f) agree.

## Open findings

(none — every finding has been disposed)
