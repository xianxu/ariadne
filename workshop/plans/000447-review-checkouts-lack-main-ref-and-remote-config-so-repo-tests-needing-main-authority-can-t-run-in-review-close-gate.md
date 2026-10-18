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

## Open findings

- **BR-1** [Minor] `allowlist-shape-too-loose` (kind: defect) A remote branch named feature/main creates a stray remote origin/feature in the review checkout
- **BR-2** [Minor] `authority-source-scope` (kind: defect) The main-authority config read covers global and system scope, not just the slot's repository
