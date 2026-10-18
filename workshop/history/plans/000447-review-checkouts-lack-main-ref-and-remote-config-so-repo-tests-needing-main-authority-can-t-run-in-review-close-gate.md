---
gate: boundary-review
issue: 447
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7e377110-ae60-4ef1-8559-65139d46009a
            repository: github.com/xianxu/ariadne
            issue_id: "000447"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/9207fed21adc0d72
            reviewed_head: 126eb88b52d3a8b4b671895272ba6b2214d418e1
            slot: ariadne:2
            requested_at: "2026-10-18T21:28:27.97644Z"
          phase: finished
          pid: 78431
          started_at: "2026-10-18T21:28:28.183367Z"
          finished_at: "2026-10-18T21:30:02.907021Z"
          session_id: 16988d97-20d6-4b89-b7b9-dc28c42796a2
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 126eb88b52d3a8b4b671895272ba6b2214d418e1
        tracker_head: 95575a4fe546616decb95c52b252e689aa61a14d
        branch_point: 60ed08fd6e643c0bd38cdbe164c39497224c7766
        since: "2026-10-18T13:57:34-07:00"
        until: "2026-10-18T14:01:34-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T13:57:34-07:00"
              end: "2026-10-18T14:01:34-07:00"
              began: "2026-10-18T13:57:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: df8ca35833fee1411e9eff24936465d81fe6b9411d9dbd6e1dbeddb943b87b7e
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 205372721
              windows:
                - start: "2026-10-18T13:57:34-07:00"
                  end: "2026-10-18T14:01:34-07:00"
      "n": 1
      timestamp: "2026-10-18T14:30:04-07:00"
      dispatched: "2026-10-18T14:28:27-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: A remote branch named feature/main creates a stray remote origin/feature in the review checkout
          detail: authority.go:49 accepts any refs/remotes/<X>/main, so refs/remotes/origin/feature/main makes addRemote("origin/feature"). copyMainAuthority then sets remote.origin/feature.url, which git accepts (verified). It is harmless but adds a remote the slot does not have; accept only remote names without a slash.
          family: allowlist-shape-too-loose
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: The main-authority config read covers global and system scope, not just the slot's repository
          detail: checkout.go:140 runs git config --get-regexp without --local, so a user-level branch.main.* or remote.*.fetch is copied as if it were the slot's. A valueless key anywhere in scope makes checkout creation fail and blocks the review; add --local.
          family: authority-source-scope
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 126eb88b52d3a8b4b671895272ba6b2214d418e1
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0acfb2cd-1b7e-4feb-8def-4df842a90477
            repository: github.com/xianxu/ariadne
            issue_id: "000447"
            recipe: fix-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/7b14e37e3794c2bd
            reviewed_head: 12b3eb3f91cc603aaf142a10d8730f0a20bb6f9d
            slot: ariadne:2
            requested_at: "2026-10-18T21:57:57.712906Z"
          phase: finished
          pid: 94327
          started_at: "2026-10-18T21:57:57.803854Z"
          finished_at: "2026-10-18T21:58:37.964554Z"
          session_id: dbec5979-e9e6-48b7-b51b-c7ca58995cf1
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 12b3eb3f91cc603aaf142a10d8730f0a20bb6f9d
        tracker_head: 0e6efae87ae8f6ae2b87a160848ea6bd71776bcb
        branch_point: 60ed08fd6e643c0bd38cdbe164c39497224c7766
        since: "2026-10-18T13:57:34-07:00"
        until: "2026-10-18T14:31:46-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T13:57:34-07:00"
              end: "2026-10-18T14:31:46-07:00"
              began: "2026-10-18T13:57:34-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 2f9e80856821f1a2d9704b73b325edca749ebc4a968ce5280d6fa35fe3dfe091
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 206712866
              windows:
                - start: "2026-10-18T13:57:34-07:00"
                  end: "2026-10-18T14:31:46-07:00"
        queue_waits:
            - attempt_id: 2c28a8b33d5a808b8e75731906e150cb
              slot: ariadne:2
              record_digest: 70b3dbc0b79d0b0dc25051653ded55e292e34a10444008ef3744d05cce8ca631
              start: "2026-10-18T21:01:41.142009Z"
              end: "2026-10-18T21:19:20.638179Z"
              queue_seconds: 1059.49617
              raw_seconds: 1059.49617
              exact: true
        cuts:
            - dispatched: "2026-10-18T14:28:27-07:00"
              finalized: "2026-10-18T14:30:04-07:00"
              finalizing: true
      "n": 2
      timestamp: "2026-10-18T14:58:38-07:00"
      dispatched: "2026-10-18T14:57:57-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: remoteMain (authority.go:85) admits only a one-segment R; new TestMainAuthority row goes red when the slash check is reverted.
          round: 2
        - id: BR-2
          disposition: addressed
          note: checkout.go:140 now passes --local; TestCheckoutCopiesOnlyTheSlotsConfig goes red (copies remote.elsewhere and branch.main) when --local is reverted.
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: 12b3eb3f91cc603aaf142a10d8730f0a20bb6f9d
      blocked: false
---

# Gate ledger — ariadne#447 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T14:30:04-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `allowlist-shape-too-loose` (kind: defect) A remote branch named feature/main creates a stray remote origin/feature in the review checkout
  authority.go:49 accepts any refs/remotes/<X>/main, so refs/remotes/origin/feature/main makes addRemote("origin/feature"). copyMainAuthority then sets remote.origin/feature.url, which git accepts (verified). It is harmless but adds a remote the slot does not have; accept only remote names without a slash.
- **BR-2** [Minor] `authority-source-scope` (kind: defect) The main-authority config read covers global and system scope, not just the slot's repository
  checkout.go:140 runs git config --get-regexp without --local, so a user-level branch.main.* or remote.*.fetch is copied as if it were the slot's. A valueless key anywhere in scope makes checkout creation fail and blocks the review; add --local.

## Round 2 — 2026-10-18T14:58:38-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — remoteMain (authority.go:85) admits only a one-segment R; new TestMainAuthority row goes red when the slash check is reverted.
- BR-2 — addressed — checkout.go:140 now passes --local; TestCheckoutCopiesOnlyTheSlotsConfig goes red (copies remote.elsewhere and branch.main) when --local is reverted.

## Open findings

(none — every finding has been disposed)
