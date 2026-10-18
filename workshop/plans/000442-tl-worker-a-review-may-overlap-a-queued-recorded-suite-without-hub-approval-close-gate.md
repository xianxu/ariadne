---
gate: boundary-review
issue: 442
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 51e80b50-9f09-4592-ae58-151c936f1fdb
            repository: github.com/xianxu/ariadne
            issue_id: "000442"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7f32db816d9c4151
            reviewed_head: 04363208a954f28e818a9b640b3c5f790aadcf42
            slot: ariadne:2
            requested_at: "2026-10-18T15:31:40.086073Z"
          phase: finished
          pid: 69400
          started_at: "2026-10-18T15:31:40.173447Z"
          finished_at: "2026-10-18T15:32:11.57313Z"
          session_id: c3331ca0-35d4-4994-a8a2-2df246f59de3
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 04363208a954f28e818a9b640b3c5f790aadcf42
        tracker_head: 6afcab021e6bfc8f75dfe1e586fd3872db037dae
        branch_point: 48b3b97b7cc45ab813214baddf6150549037225b
        since: "2026-10-18T08:30:28-07:00"
        until: "2026-10-18T08:31:13-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T08:30:28-07:00"
              end: "2026-10-18T08:31:13-07:00"
              began: "2026-10-18T08:30:28-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f2b184d36caf682fb11edabf55fe9fd8a11e168ad3e0d37ac9cc8c2d58c4e6d8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 175872517
              windows:
                - start: "2026-10-18T08:30:28-07:00"
                  end: "2026-10-18T08:31:13-07:00"
      "n": 1
      timestamp: "2026-10-18T08:32:12-07:00"
      dispatched: "2026-10-18T08:31:40-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Spec condition (1), the suite pass on the final head before landing, has no rules.tsv phrase guard
          detail: 'help.md:35 "a recorded suite pass on the final head still precedes landing" matches no rules.tsv row (early-draft-gates matches line 33 only), so deleting the safety clause keeps TestWorkerHelpCarriesEveryRule green. Enumerated family in window: overlap permission (guarded), condition 2 requeue (guarded), condition 1 (unguarded, the only instance). Fix: add an own-scope row, e.g. review-suite-pass-first with phrase "recorded suite pass on the final head still precedes landing", and update the Plan line from two rows to three.'
          family: rule-clause-unguarded
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: review-suite-overlap phrase stops at "without hub" and omits "approval"
          detail: The line wrap at help.md:34-35 means the guard does not pin "approval"; the no-approval wording could drift without the test noticing.
          family: rule-clause-unguarded
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7240fa34-2277-4039-9e09-288e1a36218a
            repository: github.com/xianxu/ariadne
            issue_id: "000442"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/a8a1a9fa7d5e6ca6
            reviewed_head: f4dee0dcde01fb7d590b79c65ce9c2c641176f2c
            slot: ariadne:2
            requested_at: "2026-10-18T15:33:06.615053Z"
          phase: finished
          pid: 75690
          started_at: "2026-10-18T15:33:06.706667Z"
          finished_at: "2026-10-18T15:33:27.239772Z"
          session_id: d9ef123e-c3c6-4399-a70e-99fd907882ca
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f4dee0dcde01fb7d590b79c65ce9c2c641176f2c
        tracker_head: 6afcab021e6bfc8f75dfe1e586fd3872db037dae
        branch_point: 48b3b97b7cc45ab813214baddf6150549037225b
        since: "2026-10-18T08:30:28-07:00"
        until: "2026-10-18T08:32:49-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T08:30:28-07:00"
              end: "2026-10-18T08:32:49-07:00"
              began: "2026-10-18T08:30:28-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ed4bdcd197577fdc871329135c29584fbdd11e66c906f8267124aebf8cf3ee4a
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 176804630
              windows:
                - start: "2026-10-18T08:30:28-07:00"
                  end: "2026-10-18T08:32:49-07:00"
        cuts:
            - dispatched: "2026-10-18T08:31:40-07:00"
              finalized: "2026-10-18T08:32:12-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T08:33:27-07:00"
      dispatched: "2026-10-18T08:33:06-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: rules.tsv adds review-suite-pass-first; deleting help.md condition (1) line turns TestWorkerHelpCarriesEveryRule red (verified by revert).
          round: 2
        - id: BR-2
          disposition: addressed
          note: Reflowed so "without hub approval, under two conditions" is one line, guarded by review-suite-approval; rewording "approval" turns the test red (verified).
          round: 2
      recipe: small-diff-review
      isolation: fenced
      reviewed: f4dee0dcde01fb7d590b79c65ce9c2c641176f2c
      blocked: false
---

# Gate ledger — ariadne#442 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T08:32:12-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `rule-clause-unguarded` (kind: defect) Spec condition (1), the suite pass on the final head before landing, has no rules.tsv phrase guard
  help.md:35 "a recorded suite pass on the final head still precedes landing" matches no rules.tsv row (early-draft-gates matches line 33 only), so deleting the safety clause keeps TestWorkerHelpCarriesEveryRule green. Enumerated family in window: overlap permission (guarded), condition 2 requeue (guarded), condition 1 (unguarded, the only instance). Fix: add an own-scope row, e.g. review-suite-pass-first with phrase "recorded suite pass on the final head still precedes landing", and update the Plan line from two rows to three.
- **BR-2** [Minor] `rule-clause-unguarded` (kind: defect) review-suite-overlap phrase stops at "without hub" and omits "approval"
  The line wrap at help.md:34-35 means the guard does not pin "approval"; the no-approval wording could drift without the test noticing.

## Round 2 — 2026-10-18T08:33:27-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — rules.tsv adds review-suite-pass-first; deleting help.md condition (1) line turns TestWorkerHelpCarriesEveryRule red (verified by revert).
- BR-2 — addressed — Reflowed so "without hub approval, under two conditions" is one line, guarded by review-suite-approval; rewording "approval" turns the test red (verified).

## Open findings

(none — every finding has been disposed)
