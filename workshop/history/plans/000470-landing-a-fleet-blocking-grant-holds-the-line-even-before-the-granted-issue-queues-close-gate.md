---
gate: boundary-review
issue: 470
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 94b96ee8-2870-4045-a2dd-1959433e500a
            repository: github.com/xianxu/ariadne
            issue_id: "000470"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/ee510c4184d641cf
            reviewed_head: ea1e7197580fcd4c9dea6df83f0550beeef0b57a
            slot: ariadne:1
            requested_at: "2026-10-20T02:52:56.591363Z"
          phase: finished
          pid: 17160
          started_at: "2026-10-20T02:52:56.677342Z"
          finished_at: "2026-10-20T03:04:04.628176Z"
          session_id: 01a15739-eecd-7833-94fb-9f6a71822bb1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: ea1e7197580fcd4c9dea6df83f0550beeef0b57a
        tracker_head: c087f8d56fcf1d31733bd62367d975230766579e
        branch_point: 259c9b93331fcb4d8bdea1716eadcc109f106e92
        since: "2026-10-19T18:22:36-07:00"
        until: "2026-10-19T19:52:30-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T18:22:36-07:00"
              end: "2026-10-19T19:52:30-07:00"
              began: "2026-10-19T18:22:36-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a156e6-ad1c-73b0-a82f-2043dce7ed9e
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 857079894588b8f11189547e1885f362ebfab9e4d456c96225531ece3fa29e58
              started: "2026-10-20T01:22:09.69Z"
              spawned: false
              bytes: 2574842
              windows:
                - start: "2026-10-19T18:22:36-07:00"
                  end: "2026-10-19T19:52:30-07:00"
        queue_waits:
            - attempt_id: 1cf9e912385eddd3dde5df905cb5901b
              slot: ariadne:1
              record_digest: d4f00a09fa2a75af792d6155f758812b859651590fb72ebbedc5ac22e2c860a2
              start: "2026-10-20T01:26:21.129228Z"
              end: "2026-10-20T01:49:25.697177Z"
              queue_seconds: 1384.567949
              raw_seconds: 1384.567949
              exact: true
            - attempt_id: a783152183fa8e85c5134f539b12016d
              slot: ariadne:1
              record_digest: da700c842a2271d7c6b72a960e5d316c38cc61db13ea860ac20caa328f0c1a24
              start: "2026-10-20T02:01:17.426851Z"
              end: "2026-10-20T02:40:04.175683Z"
              queue_seconds: 2326.748832
              raw_seconds: 2326.748832
              exact: true
        excluded_brain_sessions: 1
        excluded_brain_minutes: 19.821183333333334
      "n": 1
      timestamp: "2026-10-19T20:04:06-07:00"
      dispatched: "2026-10-19T19:52:56-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Done-when still promises unchanged grant-ordering tests
          detail: 'workshop/issues/000470-landing-a-fleet-blocking-grant-holds-the-line-even-before-the-granted-issue-queues.md:43 says existing grant-ordering tests remain unchanged. Both instances in this window intentionally change: order_test.go:40 reverses the absent-grant expectation, and line_test.go:109 clears the completed grant before normal admission. Record a Revisions entry replacing the literal unchanged-test promise with preservation of ordering and non-preemption guarantees under the revised hold policy.'
          family: acceptance-text-matches-delivery
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: ea1e7197580fcd4c9dea6df83f0550beeef0b57a
      blocked: false
---

# Gate ledger — ariadne#470 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T20:04:06-07:00 (codex) — passed

### Raised

- **BR-1** [Minor] `acceptance-text-matches-delivery` (kind: defect) Done-when still promises unchanged grant-ordering tests
  workshop/issues/000470-landing-a-fleet-blocking-grant-holds-the-line-even-before-the-granted-issue-queues.md:43 says existing grant-ordering tests remain unchanged. Both instances in this window intentionally change: order_test.go:40 reverses the absent-grant expectation, and line_test.go:109 clears the completed grant before normal admission. Record a Revisions entry replacing the literal unchanged-test promise with preservation of ordering and non-preemption guarantees under the revised hold policy.

## Open findings

- **BR-1** [Minor] `acceptance-text-matches-delivery` (kind: defect) Done-when still promises unchanged grant-ordering tests
