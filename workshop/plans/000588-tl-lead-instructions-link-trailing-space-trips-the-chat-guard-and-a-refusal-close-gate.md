---
gate: boundary-review
issue: 588
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 69fdfbeb-9054-4498-bd4f-4853c8333095
            repository: github.com/xianxu/ariadne
            issue_id: "000588"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/3fd8cc2f6171ed64
            reviewed_head: 54d9672559571d87593307d1a9386a28948031ad
            slot: ariadne:5
            requested_at: "2026-10-24T07:14:19.358258Z"
          phase: finished
          pid: 81981
          started_at: "2026-10-24T07:14:19.482178Z"
          finished_at: "2026-10-24T07:16:28.722395Z"
          session_id: 49bf248d-8148-4d55-89e4-221663659803
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 54d9672559571d87593307d1a9386a28948031ad
        tracker_head: d8eaa549b520e75faae63188495aecc2d5657092
        branch_point: 500f48f1ea8add6e2b05fc39c9a7fd72edc9157a
        since: "2026-10-23T21:18:07-07:00"
        until: "2026-10-24T00:13:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T21:18:07-07:00"
              end: "2026-10-24T00:13:51-07:00"
              began: "2026-10-23T21:18:07-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: b626bb5601908cc60c097128c3a2fe867f35579c4631396003249885d3dd4072
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 14943844
              windows:
                - start: "2026-10-23T21:18:07-07:00"
                  end: "2026-10-24T00:13:51-07:00"
      "n": 1
      timestamp: "2026-10-24T00:16:30-07:00"
      dispatched: "2026-10-24T00:14:19-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: linkInstruction's comment says a reconcile undoes the index entry and "leaves the file as it was"
          detail: 'cmd/tl-lead/instructions.go:154-158. The main recovery path in unlist restores the frontmatter byte for byte; the reconcile is only the fallback. On the fallback path, or when the undo itself is refused, the file is not exactly as it was. This is the only instance in the window: unlist''s own comment and atlas/workflow/tl-skills.md describe it correctly.'
          family: doc-claim-from-memory
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: No test drives a chat-guard refusal through link and the undo
          detail: TestIndexLinkUnderChatGuard reaches the guard by calling WriteSection directly. Through link, the undo is exercised only by a stale write (TestIndexLinkRefusedLeavesFile) and by the lock, where the undo is refused too. A guard refusal followed by a successful undo, which is the issue's reported case, runs the same branch as the lock case but is not exercised.
          family: done-when-path-not-end-to-end
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 54d9672559571d87593307d1a9386a28948031ad
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 37ea8ae2-aa0f-4488-8690-dfe2d588ee48
            repository: github.com/xianxu/ariadne
            issue_id: "000588"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/cd1ef6f6986a94a7
            reviewed_head: 584fcb609daa49b807fbc6f8e88a4ac62a89039f
            slot: ariadne:5
            requested_at: "2026-10-24T07:21:24.24614Z"
          phase: finished
          pid: 59955
          started_at: "2026-10-24T07:21:24.541322Z"
          finished_at: "2026-10-24T07:23:46.181866Z"
          session_id: 5f3f0c40-cc86-462f-9366-f0b00b70db24
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 584fcb609daa49b807fbc6f8e88a4ac62a89039f
        tracker_head: 7c951840770a98dfb0e3e358e0f4d3b254be8e7c
        branch_point: 500f48f1ea8add6e2b05fc39c9a7fd72edc9157a
        since: "2026-10-23T21:18:07-07:00"
        until: "2026-10-24T00:20:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T21:18:07-07:00"
              end: "2026-10-24T00:20:31-07:00"
              began: "2026-10-23T21:18:07-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 73d431007100f82f39cb463adbc9dc4485c34106dd9af1492ced62037300c53b
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 15245043
              windows:
                - start: "2026-10-23T21:18:07-07:00"
                  end: "2026-10-24T00:20:31-07:00"
        cuts:
            - dispatched: "2026-10-24T00:14:19-07:00"
              finalized: "2026-10-24T00:16:30-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-24T00:23:59-07:00"
      dispatched: "2026-10-24T00:21:24-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: instructions.go:156-157 now says unlist restores the frontmatter's old bytes, or reconciles when the frontmatter moved; matches unlist at instructions.go:214-235.
          round: 2
        - id: BR-2
          disposition: addressed
          note: index_test.go:102-109 drives a refuseAt guard refusal through link; with the unlist call removed it fails at index_test.go:108, and it passes on head.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 584fcb609daa49b807fbc6f8e88a4ac62a89039f
      blocked: false
---

# Gate ledger — ariadne#588 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T00:16:30-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `doc-claim-from-memory` (kind: defect) linkInstruction's comment says a reconcile undoes the index entry and "leaves the file as it was"
  cmd/tl-lead/instructions.go:154-158. The main recovery path in unlist restores the frontmatter byte for byte; the reconcile is only the fallback. On the fallback path, or when the undo itself is refused, the file is not exactly as it was. This is the only instance in the window: unlist's own comment and atlas/workflow/tl-skills.md describe it correctly.
- **BR-2** [Minor] `done-when-path-not-end-to-end` (kind: defect) No test drives a chat-guard refusal through link and the undo
  TestIndexLinkUnderChatGuard reaches the guard by calling WriteSection directly. Through link, the undo is exercised only by a stale write (TestIndexLinkRefusedLeavesFile) and by the lock, where the undo is refused too. A guard refusal followed by a successful undo, which is the issue's reported case, runs the same branch as the lock case but is not exercised.

## Round 2 — 2026-10-24T00:23:59-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — instructions.go:156-157 now says unlist restores the frontmatter's old bytes, or reconciles when the frontmatter moved; matches unlist at instructions.go:214-235.
- BR-2 — addressed — index_test.go:102-109 drives a refuseAt guard refusal through link; with the unlist call removed it fails at index_test.go:108, and it passes on head.

## Open findings

(none — every finding has been disposed)
