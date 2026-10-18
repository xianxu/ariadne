---
gate: boundary-review
issue: 434
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e7bced6c-dc9e-4ec5-b574-7d403d111d65
            repository: github.com/xianxu/ariadne
            issue_id: "000434"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/bd1141e8e5e31227
            reviewed_head: 0e15fbe282dc936b12296577e873e4ab353603b0
            slot: ariadne:2
            requested_at: "2026-10-18T05:14:47.23119Z"
          phase: finished
          pid: 11732
          started_at: "2026-10-18T05:14:47.532613Z"
          finished_at: "2026-10-18T05:16:22.414344Z"
          session_id: ee20cce2-33c3-4017-a1d0-53c058a0077c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 0e15fbe282dc936b12296577e873e4ab353603b0
        tracker_head: b7866452351e213c35441efe06dec9a22b2589c4
        branch_point: d6be01d5c70701215c16224ead1834f66e587576
        since: "2026-10-17T21:15:30-07:00"
        until: "2026-10-17T22:14:19-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:15:30-07:00"
              end: "2026-10-17T22:14:19-07:00"
              began: "2026-10-17T21:15:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 7b251eb38bd7ed9a36a12fb8a37556e1574ee1b89f7a1c800cde471c953845f8
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 100921735
              windows:
                - start: "2026-10-17T21:15:30-07:00"
                  end: "2026-10-17T22:14:19-07:00"
        queue_waits:
            - attempt_id: fd3257c5944743de8e7ffe2885a526aa
              slot: ariadne:2
              record_digest: 710ce5a943bc6674c841209741279cd89af0d3c0509ea76b7201351085e2127e
              start: "2026-10-18T04:32:34.684523Z"
              end: "2026-10-18T04:42:05.667398Z"
              queue_seconds: 570.982875
              raw_seconds: 570.982875
              exact: true
            - attempt_id: e2567b2317dc1070ec0a982a85f9cb0d
              slot: ariadne:2
              record_digest: 7be03dced4bcf1749acde420529d969ddb60f4187c0684de23c89b27c0de2e40
              start: "2026-10-18T04:55:36.890988Z"
              end: "2026-10-18T05:04:07.210632Z"
              queue_seconds: 510.319644
              raw_seconds: 510.319644
              exact: true
      "n": 1
      timestamp: "2026-10-17T22:16:23-07:00"
      dispatched: "2026-10-17T22:14:47-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: Make-probe deadline test asserts only err != nil; Spec claims DeadlineExceeded
          detail: localchecks_make_process_test.go:64 accepts any error, but the Spec says both changed tests assert DeadlineExceeded. Add errors.Is(err, context.DeadlineExceeded). The only other family member, landinggit_process_test.go:67, already checks it.
          family: deadline-test-asserts-error-kind
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Hand-written pid-file polling loop repeated (ARCH-DRY)
          detail: localcheck_signal_unix_test.go:177-191 has the same shape as the helper/owner poll at :136-163 and as awaitLandingGitChild in landinggit_process_test.go. A testfix "await pid file" helper could replace all three.
          family: test-poll-loop-duplication
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 0e15fbe282dc936b12296577e873e4ab353603b0
      blocked: false
---

# Gate ledger — ariadne#434 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T22:16:23-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `deadline-test-asserts-error-kind` (kind: defect) Make-probe deadline test asserts only err != nil; Spec claims DeadlineExceeded
  localchecks_make_process_test.go:64 accepts any error, but the Spec says both changed tests assert DeadlineExceeded. Add errors.Is(err, context.DeadlineExceeded). The only other family member, landinggit_process_test.go:67, already checks it.
- **BR-2** [Minor] `test-poll-loop-duplication` (kind: defect) Hand-written pid-file polling loop repeated (ARCH-DRY)
  localcheck_signal_unix_test.go:177-191 has the same shape as the helper/owner poll at :136-163 and as awaitLandingGitChild in landinggit_process_test.go. A testfix "await pid file" helper could replace all three.

## Open findings

- **BR-1** [Minor] `deadline-test-asserts-error-kind` (kind: defect) Make-probe deadline test asserts only err != nil; Spec claims DeadlineExceeded
- **BR-2** [Minor] `test-poll-loop-duplication` (kind: defect) Hand-written pid-file polling loop repeated (ARCH-DRY)
