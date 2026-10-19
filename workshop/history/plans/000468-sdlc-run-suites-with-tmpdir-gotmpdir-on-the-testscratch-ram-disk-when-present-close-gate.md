---
gate: boundary-review
issue: 468
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 43236dff-a7c3-44e8-a00f-4975cfc789c5
            repository: github.com/xianxu/ariadne
            issue_id: "000468"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/08489f19f6d85445
            reviewed_head: f0927a34254769c8179bd8aa501cc88587ff3398
            slot: ariadne:4
            requested_at: "2026-10-19T17:57:22.340162Z"
          phase: finished
          pid: 25536
          started_at: "2026-10-19T17:57:22.806844Z"
          finished_at: "2026-10-19T18:10:15.680014Z"
          session_id: d5248414-85a1-4cb3-95dd-0f538e3d2a08
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: f0927a34254769c8179bd8aa501cc88587ff3398
        tracker_head: 7ad6b317721a2a0a1e5a00e1a1a87298e2f2fb1b
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T09:32:22-07:00"
        until: "2026-10-19T10:56:28-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T09:32:22-07:00"
              end: "2026-10-19T10:56:28-07:00"
              began: "2026-10-19T09:32:22-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 6e6e6007ac958578b9267de69e0708c4fed61f2dc0688d9a2eedb5efb6942486
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 64184666
              windows:
                - start: "2026-10-19T09:32:22-07:00"
                  end: "2026-10-19T10:56:28-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
      "n": 1
      timestamp: "2026-10-19T11:10:17-07:00"
      dispatched: "2026-10-19T10:57:22-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: The run-folder sweep never matches folders made by hyphenated check names, so a crashed run's folder stays on the RAM disk for good
          detail: 'testscratch.go:53 runDirRE requires the check name to match [A-Za-z0-9_.]+, but localchecks.go:77 allows hyphens ([a-z0-9-]). prepareScratch names a folder <check>-<pid>-<rand>, so a crashed run of a check such as go-lint leaves a folder that neither the dead-pid nor the 24 h age rule matches (ARCH-FUNERAL). Fix: ^.+-(\d+)-\d+$, plus a hyphenated case in TestSweepStaleTemp.'
          family: sweep-pattern-narrower-than-writer
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: The leak warning omits the sizes and full entry names that Revision 2a promised
          detail: Revision 2a says the warning reports the count and size, and each entry's name; plan step 2a says "plus sizes". reportScratchLeaks (testscratch.go:201) prints counts and one example per package, and never measures size. Either deliver size and the full list, or add a Revisions entry narrowing the contract.
          family: revision-contract-drift
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: The review-checkout fallback retries on any Create error and blames the RAM disk
          detail: reviewisolation.go:42 retries on the cache root for any error, such as a bad SHA or a git failure, and the warning says the RAM disk failed. Retry only on write or permission errors, or reword the warning neutrally.
          family: fallback-on-unrelated-error
          round: 1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d2b0c5ce-34ec-4758-a9dd-c767d2e00569
            repository: github.com/xianxu/ariadne
            issue_id: "000468"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/627dce757f9e23c8
            reviewed_head: a1145892568eb97974a4f73a4fd52e3bb3b65158
            slot: ariadne:4
            requested_at: "2026-10-19T18:12:59.679767Z"
          phase: finished
          pid: 77871
          started_at: "2026-10-19T18:12:59.994785Z"
          finished_at: "2026-10-19T18:15:16.049884Z"
          session_id: 05596eb7-026a-4065-afd1-f27e82c9e77b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: a1145892568eb97974a4f73a4fd52e3bb3b65158
        tracker_head: 7ad6b317721a2a0a1e5a00e1a1a87298e2f2fb1b
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T09:32:22-07:00"
        until: "2026-10-19T11:11:34-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T09:32:22-07:00"
              end: "2026-10-19T11:11:34-07:00"
              began: "2026-10-19T09:32:22-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5dc078e93bd3ee08df7a463679c59a7257a49673922d90fdc0e1d17e06110d2c
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 65638888
              windows:
                - start: "2026-10-19T09:32:22-07:00"
                  end: "2026-10-19T11:11:34-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
        cuts:
            - dispatched: "2026-10-19T10:57:22-07:00"
              finalized: "2026-10-19T11:10:17-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T11:15:17-07:00"
      dispatched: "2026-10-19T11:12:59-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: runDirRE now ^.+-(\d+)-\d+$; reverting it makes TestSweepStaleTemp fail on go-lint-100-4 (verified in this checkout).
          round: 2
        - id: BR-2
          disposition: addressed
          note: reportScratchLeaks gives per-package and total size and every name up to leakListCap; TestScratchLeaksAttributesOnlyThisRun pins them; a Revisions entry narrows sizing to attributed entries.
          round: 2
        - id: BR-3
          disposition: not-addressed
          note: The narrowing and the neutral wording are in, but no test shows a non-permission Create error skips the fallback; TestReviewIsolationFallsBackFromScratch passes with or without the errors.Is guard.
          round: 2
      findings:
        - kind: defect
          id: BR-4
          severity: Minor
          title: The run-folder sweep pattern is now wider than prepareScratch's names, so it can delete foreign folders on the shared RAM disk
          detail: 'This is the 2nd finding in family sweep-pattern-narrower-than-writer. The rule covering both: a sweep''s match must be built from the same constant as its writer, neither narrower nor wider. The local-check sweep already does this with localCheckTmpPrefix. testscratch.go:56 ^.+-(\d+)-\d+$ matches any x-N-N folder under /Volumes/TestScratch/tmp, which AGENTS.base.md now advertises for general scratch work, and removes it when the pid looks dead or it is a day old. Fix: one shared prefix constant (e.g. sdlc-run-) used by both the MkdirTemp at testscratch.go:44 and the sweep regex, plus a test case with a foreign pr-468-2 folder that must survive.'
          family: sweep-pattern-narrower-than-writer
          round: 2
      recipe: milestone-review
      isolation: checkout
      reviewed: a1145892568eb97974a4f73a4fd52e3bb3b65158
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 2f8fbf2c-0a4e-43c1-a579-b710a1c9a521
            repository: github.com/xianxu/ariadne
            issue_id: "000468"
            recipe: fix-review
            round: 3
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/c8a9f1459c47543c
            reviewed_head: 9e4c4f89b69e712e4ad6bd608f3416530827c64a
            slot: ariadne:4
            requested_at: "2026-10-19T18:17:20.578651Z"
          phase: finished
          pid: 79644
          started_at: "2026-10-19T18:17:20.6951Z"
          finished_at: "2026-10-19T18:18:51.646812Z"
          session_id: a726d180-d42d-412d-adbb-988a7efef389
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 9e4c4f89b69e712e4ad6bd608f3416530827c64a
        tracker_head: 7ad6b317721a2a0a1e5a00e1a1a87298e2f2fb1b
        branch_point: 931a4ad6898a9883e1324701e92860bca401f2bc
        since: "2026-10-19T09:32:22-07:00"
        until: "2026-10-19T11:16:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T09:32:22-07:00"
              end: "2026-10-19T11:16:51-07:00"
              began: "2026-10-19T09:32:22-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d04bf96f327f1a50142bc4ddfb5216c590a60562de75590bc952fc44376c936b
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 67022313
              windows:
                - start: "2026-10-19T09:32:22-07:00"
                  end: "2026-10-19T11:16:51-07:00"
        queue_waits:
            - attempt_id: bab05729e5aa66707d18abc487e82c82
              slot: ariadne:4
              record_digest: 4269c85bb4cee1e5d6d512fb1ca487093fc523f8f4e0c964b593b99172217c84
              start: "2026-10-19T16:56:57.010726Z"
              end: "2026-10-19T17:00:27.968039Z"
              queue_seconds: 210.957313
              raw_seconds: 210.957313
              exact: true
            - attempt_id: f40b02c7b758c556fd24519f024975c3
              slot: ariadne:4
              record_digest: 2e5a51b6f6979f0e1b47fdafd876e549427c0fdad41895c8eadcc5909daf5884
              start: "2026-10-19T17:33:42.737584Z"
              end: "2026-10-19T17:44:43.407011Z"
              queue_seconds: 660.669427
              raw_seconds: 660.669427
              exact: true
        cuts:
            - dispatched: "2026-10-19T10:57:22-07:00"
              finalized: "2026-10-19T11:10:17-07:00"
              finalizing: false
            - dispatched: "2026-10-19T11:12:59-07:00"
              finalized: "2026-10-19T11:15:17-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-19T11:18:53-07:00"
      dispatched: "2026-10-19T11:17:20-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-3
          disposition: addressed
          note: reviewisolation.go:46 falls back only on fs.ErrPermission/EROFS; TestReviewIsolationKeepsOtherErrors goes red when reverted to the old any-error condition.
          round: 3
        - id: BR-4
          disposition: addressed
          note: scratchRunPrefix (testscratch.go:57) is shared by MkdirTemp (:44) and runDirRE (:62); the foreign other-tool-100-6 case goes red when the prefix is removed.
          round: 3
      recipe: fix-review
      isolation: checkout
      reviewed: 9e4c4f89b69e712e4ad6bd608f3416530827c64a
      blocked: false
---

# Gate ledger — ariadne#468 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T11:10:17-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `sweep-pattern-narrower-than-writer` (kind: defect) The run-folder sweep never matches folders made by hyphenated check names, so a crashed run's folder stays on the RAM disk for good
  testscratch.go:53 runDirRE requires the check name to match [A-Za-z0-9_.]+, but localchecks.go:77 allows hyphens ([a-z0-9-]). prepareScratch names a folder <check>-<pid>-<rand>, so a crashed run of a check such as go-lint leaves a folder that neither the dead-pid nor the 24 h age rule matches (ARCH-FUNERAL). Fix: ^.+-(\d+)-\d+$, plus a hyphenated case in TestSweepStaleTemp.
- **BR-2** [Important] `revision-contract-drift` (kind: defect) The leak warning omits the sizes and full entry names that Revision 2a promised
  Revision 2a says the warning reports the count and size, and each entry's name; plan step 2a says "plus sizes". reportScratchLeaks (testscratch.go:201) prints counts and one example per package, and never measures size. Either deliver size and the full list, or add a Revisions entry narrowing the contract.
- **BR-3** [Minor] `fallback-on-unrelated-error` (kind: defect) The review-checkout fallback retries on any Create error and blames the RAM disk
  reviewisolation.go:42 retries on the cache root for any error, such as a bad SHA or a git failure, and the warning says the RAM disk failed. Retry only on write or permission errors, or reword the warning neutrally.

## Round 2 — 2026-10-19T11:15:17-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — runDirRE now ^.+-(\d+)-\d+$; reverting it makes TestSweepStaleTemp fail on go-lint-100-4 (verified in this checkout).
- BR-2 — addressed — reportScratchLeaks gives per-package and total size and every name up to leakListCap; TestScratchLeaksAttributesOnlyThisRun pins them; a Revisions entry narrows sizing to attributed entries.
- BR-3 — not-addressed — The narrowing and the neutral wording are in, but no test shows a non-permission Create error skips the fallback; TestReviewIsolationFallsBackFromScratch passes with or without the errors.Is guard.

### Raised

- **BR-4** [Minor] `sweep-pattern-narrower-than-writer` (kind: defect) The run-folder sweep pattern is now wider than prepareScratch's names, so it can delete foreign folders on the shared RAM disk
  This is the 2nd finding in family sweep-pattern-narrower-than-writer. The rule covering both: a sweep's match must be built from the same constant as its writer, neither narrower nor wider. The local-check sweep already does this with localCheckTmpPrefix. testscratch.go:56 ^.+-(\d+)-\d+$ matches any x-N-N folder under /Volumes/TestScratch/tmp, which AGENTS.base.md now advertises for general scratch work, and removes it when the pid looks dead or it is a day old. Fix: one shared prefix constant (e.g. sdlc-run-) used by both the MkdirTemp at testscratch.go:44 and the sweep regex, plus a test case with a foreign pr-468-2 folder that must survive.

## Round 3 — 2026-10-19T11:18:53-07:00 (claude) — passed

### Disposed

- BR-3 — addressed — reviewisolation.go:46 falls back only on fs.ErrPermission/EROFS; TestReviewIsolationKeepsOtherErrors goes red when reverted to the old any-error condition.
- BR-4 — addressed — scratchRunPrefix (testscratch.go:57) is shared by MkdirTemp (:44) and runDirRE (:62); the foreign other-tool-100-6 case goes red when the prefix is removed.

## Open findings

(none — every finding has been disposed)
