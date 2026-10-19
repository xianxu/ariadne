---
gate: boundary-review
issue: 471
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 74c0d0d1-139e-4362-a673-7b0f40943a21
            repository: github.com/xianxu/ariadne
            issue_id: "000471"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/762119f53ffa786e
            reviewed_head: c2d924406b8704c12eb36f01b723865f5406acab
            slot: ariadne:1
            requested_at: "2026-10-19T23:30:21.016538Z"
          phase: finished
          pid: 9072
          started_at: "2026-10-19T23:30:21.254625Z"
          finished_at: "2026-10-19T23:36:20.378293Z"
          session_id: 01a15680-758c-7cd0-b1bf-137650e44e1a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: c2d924406b8704c12eb36f01b723865f5406acab
        tracker_head: 6c13e4dcad96e4bda2a08c04ec0232edb4ccc14e
        branch_point: c1fbcfe858e5d72512ee2c63555c7c5d15b7fb9b
        since: "2026-10-19T13:17:35-07:00"
        until: "2026-10-19T16:29:29-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T13:17:35-07:00"
              end: "2026-10-19T16:29:29-07:00"
              began: "2026-10-19T13:17:35-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 8f9583be372e0a0b311daed0cd8d974873042d0c443ccc5ce6ca3a6db48557d4
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 157248652
              windows:
                - start: "2026-10-19T13:17:35-07:00"
                  end: "2026-10-19T16:29:29-07:00"
        queue_waits:
            - attempt_id: d6c170ccbcf30af2ac74524d1dcafbc6
              slot: ariadne:1
              record_digest: 26493f659dcb344acdfb201afb28e036ba907ee19e759d9248d24aeab4c10941
              start: "2026-10-19T21:08:10.165505Z"
              end: "2026-10-19T21:34:37.752655Z"
              queue_seconds: 1587.58715
              raw_seconds: 1587.58715
              exact: true
            - attempt_id: 1e13daa36ff5109648641478eb687626
              slot: ariadne:1
              record_digest: d87adecb8b327517c006a5d635bfcfaeb0421518220a641368d1661bdcc084df
              start: "2026-10-19T23:18:36.955772Z"
              end: "2026-10-19T23:18:37.101793Z"
              queue_seconds: 0.146021
              raw_seconds: 0.146021
              exact: true
      "n": 1
      timestamp: "2026-10-19T16:36:21-07:00"
      dispatched: "2026-10-19T16:30:21-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: Local branch deletion can remove a tip that was never archived
          detail: cmd/sdlc/abandon.go:422 checks the local tip before remote operations and checkout, but line 475 later deletes it using unconditional git branch -D. A temporary regression moved that ref to an unarchived commit in the post-checkout hook; abandon deleted it and returned nil. Use an atomic deletion conditioned on rec.Head, and commit a deterministic regression covering movement between observation and deletion. This violates the Spec and ARCH-ORDER.
          family: destructive-ref-cas
          round: 1
        - id: BR-2
          severity: Important
          title: README update appears missing for changed branch requirements
          detail: README.md:84 and :110 describe attachment and abandonment without the new checkout requirements, and README has no change in the pinned range. Document canonical-branch admission for started work, resting-branch attachment for open issues, and terminal abandon retry requirements. Atlas and abandon help updates do not satisfy the separate README gate.
          family: user-surface-docs
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#471 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T16:36:21-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `destructive-ref-cas` Local branch deletion can remove a tip that was never archived
  cmd/sdlc/abandon.go:422 checks the local tip before remote operations and checkout, but line 475 later deletes it using unconditional git branch -D. A temporary regression moved that ref to an unarchived commit in the post-checkout hook; abandon deleted it and returned nil. Use an atomic deletion conditioned on rec.Head, and commit a deterministic regression covering movement between observation and deletion. This violates the Spec and ARCH-ORDER.
- **BR-2** [Important] `user-surface-docs` README update appears missing for changed branch requirements
  README.md:84 and :110 describe attachment and abandonment without the new checkout requirements, and README has no change in the pinned range. Document canonical-branch admission for started work, resting-branch attachment for open issues, and terminal abandon retry requirements. Atlas and abandon help updates do not satisfy the separate README gate.

## Open findings

- **BR-1** [Critical] `destructive-ref-cas` Local branch deletion can remove a tip that was never archived
- **BR-2** [Important] `user-surface-docs` README update appears missing for changed branch requirements
