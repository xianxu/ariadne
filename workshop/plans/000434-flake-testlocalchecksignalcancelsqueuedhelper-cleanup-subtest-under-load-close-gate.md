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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: af1dde4b-7107-456f-aa90-ac2696d0a007
            repository: github.com/xianxu/ariadne
            issue_id: "000434"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/b6d34d571a64bcd4
            reviewed_head: 80083a7ce362c383bd88a562d6eaf69f80e6ffc9
            slot: ariadne:2
            requested_at: "2026-10-18T05:18:20.609933Z"
          phase: finished
          pid: 76992
          started_at: "2026-10-18T05:18:20.701165Z"
          finished_at: "2026-10-18T05:19:38.822621Z"
          session_id: 4059df4f-5862-49c0-aea3-e69708538989
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 80083a7ce362c383bd88a562d6eaf69f80e6ffc9
        tracker_head: ca9356c2b17c27c9a164575cb1a220d8d0d987f0
        branch_point: d6be01d5c70701215c16224ead1834f66e587576
        since: "2026-10-17T21:15:30-07:00"
        until: "2026-10-17T22:18:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:15:30-07:00"
              end: "2026-10-17T22:18:04-07:00"
              began: "2026-10-17T21:15:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 736408101b55157af9a27fe0d3de09d02ed9ad064583f914e5c02f4b7ff62441
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 102014868
              windows:
                - start: "2026-10-17T21:15:30-07:00"
                  end: "2026-10-17T22:18:04-07:00"
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
        cuts:
            - dispatched: "2026-10-17T22:14:47-07:00"
              finalized: "2026-10-17T22:16:23-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-17T22:19:40-07:00"
      dispatched: "2026-10-17T22:18:20-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: localchecks_make_process_test.go:61-65 asserts errors.Is(err, context.DeadlineExceeded) for hanging-shell; exited-parent rightly keeps err != nil (it ends on the drain bound), as the issue Log records. Ran -count=3 green.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: testfix.AwaitPIDFile now covers awaitLandingGitChild, the make-probe wait and the signal-test cleanup wait, but localcheck_run_process_test.go:50-61 is still a hand-written copy of the same pid-file loop; switch it to AwaitPIDFile(pidfile, testfix.HangCeiling, nil). Minor, non-blocking.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 80083a7ce362c383bd88a562d6eaf69f80e6ffc9
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8c5282b8-0c40-4255-aff1-74d39f0cb688
            repository: github.com/xianxu/ariadne
            issue_id: "000434"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/1723f06f80018cef
            reviewed_head: 41f0e729593a39a5a1aa2635aa622646dbc5a04e
            slot: ariadne:2
            requested_at: "2026-10-18T05:32:01.563036Z"
          phase: finished
          pid: 43652
          started_at: "2026-10-18T05:32:01.727825Z"
          finished_at: "2026-10-18T05:34:25.019232Z"
          session_id: 2eaa68b6-ca88-4bf4-b538-d375af735d7d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 41f0e729593a39a5a1aa2635aa622646dbc5a04e
        tracker_head: 5729268e2b7ba2b8e581a0584fa6c75f6bca0658
        branch_point: d6be01d5c70701215c16224ead1834f66e587576
        since: "2026-10-17T21:15:30-07:00"
        until: "2026-10-17T22:31:35-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T21:15:30-07:00"
              end: "2026-10-17T22:31:35-07:00"
              began: "2026-10-17T21:15:30-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 9a44cf3a3fe0078b9cd778b88c6589eb79b5ed71c25047bc3cbedcb34d291eaa
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 103451986
              windows:
                - start: "2026-10-17T21:15:30-07:00"
                  end: "2026-10-17T22:31:35-07:00"
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
        cuts:
            - dispatched: "2026-10-17T22:14:47-07:00"
              finalized: "2026-10-17T22:16:23-07:00"
              finalizing: true
            - dispatched: "2026-10-17T22:18:20-07:00"
              finalized: "2026-10-17T22:19:40-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-17T22:34:25-07:00"
      dispatched: "2026-10-17T22:32:01-07:00"
      agent: claude
      dispose:
        - id: BR-2
          disposition: addressed
          note: localcheck_run_process_test.go:48 and cli_signal_unix_test.go:122 now use testfix.AwaitPIDFile, closing the last cmd/sdlc copy of the loop that BR-2 named; vet clean, touched tests green. Similar loops in internal/{processgroup,gitx,judge} tests are outside this window.
          round: 3
      recipe: fix-review
      isolation: fenced
      reviewed: 41f0e729593a39a5a1aa2635aa622646dbc5a04e
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

## Round 2 — 2026-10-17T22:19:40-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — localchecks_make_process_test.go:61-65 asserts errors.Is(err, context.DeadlineExceeded) for hanging-shell; exited-parent rightly keeps err != nil (it ends on the drain bound), as the issue Log records. Ran -count=3 green.
- BR-2 — not-addressed — testfix.AwaitPIDFile now covers awaitLandingGitChild, the make-probe wait and the signal-test cleanup wait, but localcheck_run_process_test.go:50-61 is still a hand-written copy of the same pid-file loop; switch it to AwaitPIDFile(pidfile, testfix.HangCeiling, nil). Minor, non-blocking.

## Round 3 — 2026-10-17T22:34:25-07:00 (claude) — passed

### Disposed

- BR-2 — addressed — localcheck_run_process_test.go:48 and cli_signal_unix_test.go:122 now use testfix.AwaitPIDFile, closing the last cmd/sdlc copy of the loop that BR-2 named; vet clean, touched tests green. Similar loops in internal/{processgroup,gitx,judge} tests are outside this window.

## Open findings

(none — every finding has been disposed)
