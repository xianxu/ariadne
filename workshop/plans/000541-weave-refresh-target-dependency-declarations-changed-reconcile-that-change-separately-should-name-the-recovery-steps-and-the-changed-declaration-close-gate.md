---
gate: boundary-review
issue: 541
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cafd23b9-bfa7-4b60-b085-71ebb1ae3888
            repository: github.com/xianxu/ariadne
            issue_id: "000541"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/fdac5adef85ff8b7
            reviewed_head: 048fc292ccbef0f535748ed1b50fc6e3df5f358c
            slot: ariadne:5
            requested_at: "2026-10-22T08:06:50.634705Z"
          phase: finished
          pid: 91467
          started_at: "2026-10-22T08:06:50.843227Z"
          finished_at: "2026-10-22T08:09:43.775316Z"
          session_id: 01a162a6-0aae-7583-a510-f89b7ea16ff9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 048fc292ccbef0f535748ed1b50fc6e3df5f358c
        tracker_head: 4b2631325c51af6ed04495d1ec9faf482bb1b6c5
        branch_point: f5aae36f628a4f9ea04439543f46995a45d61b37
        since: "2026-10-22T00:31:21-07:00"
        until: "2026-10-22T01:06:23-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T00:31:21-07:00"
              end: "2026-10-22T01:06:23-07:00"
              began: "2026-10-22T00:31:21-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 8bed52ca5d231e87b85c9ccf2fa76ee41616573877c33a7f89423d44ebc816f9
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 36440551
              windows:
                - start: "2026-10-22T00:31:21-07:00"
                  end: "2026-10-22T01:06:23-07:00"
        queue_waits:
            - attempt_id: f893b414f1e45b4f1e41b31036fa9636
              slot: ariadne:5
              record_digest: 6e3adc67a92a29c56a4e97e6f0b04e994e7e7f7a7d356eddf7a84cd420225c7f
              start: "2026-10-22T07:38:34.655566Z"
              end: "2026-10-22T07:53:12.846585Z"
              queue_seconds: 878.191019
              raw_seconds: 878.191019
              exact: true
      "n": 1
      timestamp: "2026-10-22T01:09:44-07:00"
      dispatched: "2026-10-22T01:06:50-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README update appears missing for declaration-drift recovery
          detail: 'README.md:372–376 still instructs readers to “reconcile that declaration change separately,” while cmd/weave/internal/refresh/refresh.go:157–161 introduces concrete recovery commands. The complete family in this window is this declaration-drift recovery flow: its atlas documentation was updated at atlas/workflow/setup-and-replication.md:66–85, but its README entry was not. Update the README entry with the clean-behind prerequisite, fast-forward/setup/retry steps, and ahead/divergent qualification, linking to the atlas recovery section.'
          family: user-facing-recovery-docs
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#541 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T01:09:44-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `user-facing-recovery-docs` (kind: defect) README update appears missing for declaration-drift recovery
  README.md:372–376 still instructs readers to “reconcile that declaration change separately,” while cmd/weave/internal/refresh/refresh.go:157–161 introduces concrete recovery commands. The complete family in this window is this declaration-drift recovery flow: its atlas documentation was updated at atlas/workflow/setup-and-replication.md:66–85, but its README entry was not. Update the README entry with the clean-behind prerequisite, fast-forward/setup/retry steps, and ahead/divergent qualification, linking to the atlas recovery section.

## Open findings

- **BR-1** [Important] `user-facing-recovery-docs` (kind: defect) README update appears missing for declaration-drift recovery
