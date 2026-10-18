---
gate: plan-quality
issue: 394
id_prefix: PQ
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            id: 1a0fb2f0-0aaf-44c3-b8eb-8b829426c28a
            repository: github.com/xianxu/ariadne
            issue_id: "000394"
            recipe: plan-quality
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            reviewed_head: 4ac6cedcd31350bcd01b7596bd4533f68987ed6b
            slot: ariadne:2
            requested_at: "2026-10-18T22:52:56.731598Z"
          phase: finished
          pid: 19602
          started_at: "2026-10-18T22:52:56.888729Z"
          finished_at: "2026-10-18T22:54:04.99539Z"
          session_id: 23a2a6e8-f430-49c9-938f-6a09729f7067
          outcome: returned
      "n": 1
      timestamp: "2026-10-18T15:54:05-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: PQ-1
          severity: Important
          title: 'The targeted rerun "under the held slot" conflicts with #383''s early slot, lease and signal release in runLocalCheck'
          detail: 'runLocalCheck calls release(), releaseLease() and stopSignals() right after the suite command exits (localcheck_run.go, the #383 comment). State whether the rerun keeps the slot (undoing #383), takes it again, and stays reachable by local-check cancel. Also state the rerun''s Weight and the WF_TEST_LOCK_HELD_BY env, since it bypasses c.Command.'
          family: unstated-lock-contract-change
          round: 1
        - kind: defect
          id: PQ-2
          severity: Important
          title: PersistentMain needs merge's baseline, but diagnoseSuiteFailure runs only after runLocalCheck returns, where the plan classifies and posts
          detail: suitebaseline.go:14-27 runs the main baseline after runLocalCheck has already failed. Specify where Classify runs for merge, whether a passing known-flake rerun skips the baseline, and which component posts the final status. Reuse classifyFailures (suitefailures.go:67).
          family: gate-effect-ordering-unstated
          round: 1
        - kind: defect
          id: PQ-3
          severity: Important
          title: 'Schema v2: decodeSuiteAttempt accepts only version 1 and a fixed key allowlist, and the success guard must fail closed on records it cannot read'
          detail: teststats_store.go:37 and the Version check. A v2 reader must still accept v1 records. A failed attempt the guard cannot decode must block the success post rather than being skipped, or an unreadable failure turns green.
          family: absent-evidence-reads-as-success
          round: 1
        - kind: defect
          id: PQ-4
          severity: Important
          title: flake register writes through issue publish, which only the owner may run; the flake issue is usually claimed by another slot
          detail: issue new from a numbered slot claims the issue, and publish refuses non-owners. Name who runs register (the worker, or the TL as D4 implies) and how a worker registers on a flake issue it does not own.
          family: unstated-ownership-contract
          round: 1
        - kind: defect
          id: PQ-5
          severity: Minor
          title: M2 and M3 list e2e test cases in prose; compress to one strategy line per risky function
          family: test-prose-enumeration
          round: 1
        - kind: defect
          id: PQ-6
          severity: Minor
          title: The new flake_tests frontmatter field is not added to the vocabulary model
          family: vocabulary-single-source
          round: 1
        - kind: defect
          id: PQ-7
          severity: Minor
          title: 'Done-when 3 still requires CI coverage that D3 moved to #451; add a Revision to Done-when'
          family: done-when-drift
          round: 1
      blocked: true
---

# Gate ledger — ariadne#394 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T15:54:05-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unstated-lock-contract-change` (kind: defect) The targeted rerun "under the held slot" conflicts with #383's early slot, lease and signal release in runLocalCheck
  runLocalCheck calls release(), releaseLease() and stopSignals() right after the suite command exits (localcheck_run.go, the #383 comment). State whether the rerun keeps the slot (undoing #383), takes it again, and stays reachable by local-check cancel. Also state the rerun's Weight and the WF_TEST_LOCK_HELD_BY env, since it bypasses c.Command.
- **PQ-2** [Important] `gate-effect-ordering-unstated` (kind: defect) PersistentMain needs merge's baseline, but diagnoseSuiteFailure runs only after runLocalCheck returns, where the plan classifies and posts
  suitebaseline.go:14-27 runs the main baseline after runLocalCheck has already failed. Specify where Classify runs for merge, whether a passing known-flake rerun skips the baseline, and which component posts the final status. Reuse classifyFailures (suitefailures.go:67).
- **PQ-3** [Important] `absent-evidence-reads-as-success` (kind: defect) Schema v2: decodeSuiteAttempt accepts only version 1 and a fixed key allowlist, and the success guard must fail closed on records it cannot read
  teststats_store.go:37 and the Version check. A v2 reader must still accept v1 records. A failed attempt the guard cannot decode must block the success post rather than being skipped, or an unreadable failure turns green.
- **PQ-4** [Important] `unstated-ownership-contract` (kind: defect) flake register writes through issue publish, which only the owner may run; the flake issue is usually claimed by another slot
  issue new from a numbered slot claims the issue, and publish refuses non-owners. Name who runs register (the worker, or the TL as D4 implies) and how a worker registers on a flake issue it does not own.
- **PQ-5** [Minor] `test-prose-enumeration` (kind: defect) M2 and M3 list e2e test cases in prose; compress to one strategy line per risky function
- **PQ-6** [Minor] `vocabulary-single-source` (kind: defect) The new flake_tests frontmatter field is not added to the vocabulary model
- **PQ-7** [Minor] `done-when-drift` (kind: defect) Done-when 3 still requires CI coverage that D3 moved to #451; add a Revision to Done-when

## Open findings

- **PQ-1** [Important] `unstated-lock-contract-change` (kind: defect) The targeted rerun "under the held slot" conflicts with #383's early slot, lease and signal release in runLocalCheck
- **PQ-2** [Important] `gate-effect-ordering-unstated` (kind: defect) PersistentMain needs merge's baseline, but diagnoseSuiteFailure runs only after runLocalCheck returns, where the plan classifies and posts
- **PQ-3** [Important] `absent-evidence-reads-as-success` (kind: defect) Schema v2: decodeSuiteAttempt accepts only version 1 and a fixed key allowlist, and the success guard must fail closed on records it cannot read
- **PQ-4** [Important] `unstated-ownership-contract` (kind: defect) flake register writes through issue publish, which only the owner may run; the flake issue is usually claimed by another slot
- **PQ-5** [Minor] `test-prose-enumeration` (kind: defect) M2 and M3 list e2e test cases in prose; compress to one strategy line per risky function
- **PQ-6** [Minor] `vocabulary-single-source` (kind: defect) The new flake_tests frontmatter field is not added to the vocabulary model
- **PQ-7** [Minor] `done-when-drift` (kind: defect) Done-when 3 still requires CI coverage that D3 moved to #451; add a Revision to Done-when
