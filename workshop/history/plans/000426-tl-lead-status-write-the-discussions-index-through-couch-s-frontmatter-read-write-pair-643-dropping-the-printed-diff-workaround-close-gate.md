---
gate: boundary-review
issue: 426
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: df736ab6-ca82-4315-977e-34b329d0f70a
            repository: github.com/xianxu/ariadne
            issue_id: "000426"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/dced4f9afdca824c
            reviewed_head: 1278a02061304bf1509f3a016e0be5922e872539
            slot: ariadne:2
            requested_at: "2026-10-18T20:07:35.370427Z"
          phase: finished
          pid: 30792
          started_at: "2026-10-18T20:07:35.462545Z"
          finished_at: "2026-10-18T20:08:52.138448Z"
          session_id: cee08e22-3c6e-4862-9f10-119db1b8b584
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1278a02061304bf1509f3a016e0be5922e872539
        tracker_head: 13436d4f391c78c06eebfd2e60b3561a859dc33b
        branch_point: 810f125b368813c581bc80b8963fd29f69ed5550
        since: "2026-10-17T07:33:33-07:00"
        until: "2026-10-18T12:26:56-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T07:33:33-07:00"
              end: "2026-10-17T07:33:57-07:00"
              began: "2026-10-17T07:33:33-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            - start: "2026-10-18T12:22:08-07:00"
              end: "2026-10-18T12:26:56-07:00"
              began: "2026-10-18T12:22:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 560882b5833ada440578a6bb7086b74f757e46a3a9d58f80c5b770b7a10505a2
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 200544190
              windows:
                - start: "2026-10-17T07:33:33-07:00"
                  end: "2026-10-17T07:33:57-07:00"
                - start: "2026-10-18T12:22:08-07:00"
                  end: "2026-10-18T12:26:56-07:00"
      "n": 1
      timestamp: "2026-10-18T13:08:52-07:00"
      dispatched: "2026-10-18T13:07:35-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: unreadChanges and TestUnreadChanges are left with no production or conformance caller
          detail: unreadChanges (conformance_test.go:164) is called only by its own test, TestUnreadChanges (index_test.go:155-166); the conformance test no longer uses it, though the log says it moved to its one user. Delete both. No other instance of this in the window.
          family: dead-test-helper-after-caller-removal
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The fake's text() comment still refers to the removed WriteAll and its whole-file write
          detail: sessionsync_fake_test.go:162-163 still says WriteAll's parse returns each body byte. The other whole-file mentions in the fake (lines 29, 37) remain accurate (the view, MarkAllSeen).
          family: stale-comment-after-api-removal
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: syncIndex discards ok and rest from FrontBlock(next), and FrontBlock's trimmed delimiter differs from frontmatter()'s exact one
          detail: 'index.go:51 uses front, _, _ := FrontBlock(next). If the frontmatter contains an indented ''  ---'' line, a reconciled entry after it lands in rest and is never written, yet is reported as dropped or added. Guard with: refuse unless ok and rest == body.'
          family: frontmatter-delimiter-rule-divergence
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 1278a02061304bf1509f3a016e0be5922e872539
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f54c903e-efd6-4755-a2cd-043765527908
            repository: github.com/xianxu/ariadne
            issue_id: "000426"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/18036cfaff099227
            reviewed_head: 219cb76ce811ac966ab6585e1cc5ce7a7d3872be
            slot: ariadne:2
            requested_at: "2026-10-18T20:51:34.099061Z"
          phase: finished
          pid: 69924
          started_at: "2026-10-18T20:51:34.355896Z"
          finished_at: "2026-10-18T20:52:31.667749Z"
          session_id: 831c0077-6498-447f-8fde-7c079ae8dfe9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 219cb76ce811ac966ab6585e1cc5ce7a7d3872be
        tracker_head: 2c82f907fbf51988a1faeec77ac44bc639caeec7
        branch_point: 810f125b368813c581bc80b8963fd29f69ed5550
        since: "2026-10-17T07:33:33-07:00"
        until: "2026-10-18T13:09:40-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T07:33:33-07:00"
              end: "2026-10-17T07:33:57-07:00"
              began: "2026-10-17T07:33:33-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
            - start: "2026-10-18T12:22:08-07:00"
              end: "2026-10-18T13:09:40-07:00"
              began: "2026-10-18T12:22:08-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: b97e6370d9ac6d943763b1c193e532ca827834aab03962dbde075f9f81215c6c
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 201291958
              windows:
                - start: "2026-10-17T07:33:33-07:00"
                  end: "2026-10-17T07:33:57-07:00"
                - start: "2026-10-18T12:22:08-07:00"
                  end: "2026-10-18T13:09:40-07:00"
        queue_waits:
            - attempt_id: 312b8f706cfe74ffb90f94bd8882fd06
              slot: ariadne:2
              record_digest: f74d455ebc535ae643583d777f8b2c07fec78e1abc02a3b3b2d1ed9a76fe42f9
              start: "2026-10-18T19:30:14.524412Z"
              end: "2026-10-18T19:58:13.683368Z"
              queue_seconds: 1679.158956
              raw_seconds: 1679.158956
              exact: true
        cuts:
            - dispatched: "2026-10-18T13:07:35-07:00"
              finalized: "2026-10-18T13:08:52-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-18T13:52:33-07:00"
      dispatched: "2026-10-18T13:51:34-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: unreadChanges (conformance_test.go) and TestUnreadChanges (index_test.go) are both deleted; grep finds no remaining references; vet and tests pass.
          round: 2
        - id: BR-2
          disposition: addressed
          note: sessionsync_fake_test.go:162-163 now says a frontmatter write leaves section hashes alone; no WriteAll mentions remain in cmd/tl-lead.
          round: 2
        - id: BR-3
          disposition: addressed
          note: index.go:55-58 refuses unless ok and rest == body; TestIndexRefusesADelimiterCouchReadsDifferently goes red with the guard disabled (err nil, writes 1) and green with it restored.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 219cb76ce811ac966ab6585e1cc5ce7a7d3872be
      blocked: false
---

# Gate ledger — ariadne#426 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T13:08:52-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `dead-test-helper-after-caller-removal` (kind: defect) unreadChanges and TestUnreadChanges are left with no production or conformance caller
  unreadChanges (conformance_test.go:164) is called only by its own test, TestUnreadChanges (index_test.go:155-166); the conformance test no longer uses it, though the log says it moved to its one user. Delete both. No other instance of this in the window.
- **BR-2** [Minor] `stale-comment-after-api-removal` (kind: defect) The fake's text() comment still refers to the removed WriteAll and its whole-file write
  sessionsync_fake_test.go:162-163 still says WriteAll's parse returns each body byte. The other whole-file mentions in the fake (lines 29, 37) remain accurate (the view, MarkAllSeen).
- **BR-3** [Minor] `frontmatter-delimiter-rule-divergence` (kind: defect) syncIndex discards ok and rest from FrontBlock(next), and FrontBlock's trimmed delimiter differs from frontmatter()'s exact one
  index.go:51 uses front, _, _ := FrontBlock(next). If the frontmatter contains an indented '  ---' line, a reconciled entry after it lands in rest and is never written, yet is reported as dropped or added. Guard with: refuse unless ok and rest == body.

## Round 2 — 2026-10-18T13:52:33-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — unreadChanges (conformance_test.go) and TestUnreadChanges (index_test.go) are both deleted; grep finds no remaining references; vet and tests pass.
- BR-2 — addressed — sessionsync_fake_test.go:162-163 now says a frontmatter write leaves section hashes alone; no WriteAll mentions remain in cmd/tl-lead.
- BR-3 — addressed — index.go:55-58 refuses unless ok and rest == body; TestIndexRefusesADelimiterCouchReadsDifferently goes red with the guard disabled (err nil, writes 1) and green with it restored.

## Open findings

(none — every finding has been disposed)
