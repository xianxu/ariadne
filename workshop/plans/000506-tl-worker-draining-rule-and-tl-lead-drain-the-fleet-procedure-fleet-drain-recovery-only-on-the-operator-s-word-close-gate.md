---
gate: boundary-review
issue: 506
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: f8b854d8-807c-45ba-b436-bd0b1d306113
            repository: github.com/xianxu/ariadne
            issue_id: "000506"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/dd0622cbdb164413
            reviewed_head: 27d3c860c3f6fdf2f8f6d35e7ff6b8f21dcc965c
            slot: ariadne:5
            requested_at: "2026-10-22T10:03:50.544475Z"
          phase: finished
          pid: 26491
          started_at: "2026-10-22T10:03:50.655352Z"
          finished_at: "2026-10-22T10:05:19.893023Z"
          session_id: 01a16311-273d-7d70-9bfc-10feba40fdc9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: 27d3c860c3f6fdf2f8f6d35e7ff6b8f21dcc965c
        tracker_head: f4d5cdd3cf9227ac54311a35cc3e882c8401c9e5
        branch_point: fce345a89d417394aa6d1b7c09c758b02707066b
        since: "2026-10-22T02:11:50-07:00"
        until: "2026-10-22T03:03:26-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-22T02:11:50-07:00"
              end: "2026-10-22T03:03:26-07:00"
              began: "2026-10-22T02:11:50-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: dba0f306a916f3a48ab4711b60c9b72f5ed0f96f1f821f47f90cafdd915e3269
              started: "2026-10-20T15:58:37.432Z"
              spawned: false
              bytes: 38839755
              windows:
                - start: "2026-10-22T02:11:50-07:00"
                  end: "2026-10-22T03:03:26-07:00"
            - harness: codex
              id: 01a162e6-6626-70f3-a0d9-e8ea3bb4e614
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a15a09-0f59-7b93-ad53-4f28144d42d9
              digest: a08cbb77ce98c7aba47a0843e70716ad0db315ff05dfd88e2047f7160778a806
              started: "2026-10-22T09:17:09.197Z"
              spawned: true
              bytes: 552627
              windows:
                - start: "2026-10-22T09:17:09.197Z"
                  end: "2026-10-22T03:03:26-07:00"
        queue_waits:
            - attempt_id: fe3705e457697018c73456fee2b3b0dd
              slot: ariadne:5
              record_digest: 57e439abe5db38e38644f37622bb7b1da27729bf8e020a308d6a688145a38d82
              start: "2026-10-22T09:30:02.573494Z"
              end: "2026-10-22T09:49:02.447499Z"
              queue_seconds: 1139.874005
              raw_seconds: 1139.874005
              exact: true
      "n": 1
      timestamp: "2026-10-22T03:05:20-07:00"
      dispatched: "2026-10-22T03:03:50-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Remove the obsolete acknowledgment protocol from remaining drain guidance
          detail: cmd/sdlc/helptext/queue.md:70-71 still instructs workers to "acknowledge the authenticated drain ID"; go run ./cmd/sdlc queue --help exits 0 and renders that instruction. The other remaining instance is atlas/workflow/sdlc-binary.md:2150, which describes a supervisor "notice/ack formatter". These conflict with cmd/tl-worker/help.md:122 and leave the ARCH-DRY/ARCH-PURPOSE contract sweep incomplete. Replace both restatements with pointers to the authoritative worker/Couch procedures, and extend rendered-help regression coverage to reject the removed acknowledgment requirement in queue help. Preserve the deliberately historical fixtures.
          family: drain-contract-consistency
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#506 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-22T03:05:20-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `drain-contract-consistency` (kind: defect) Remove the obsolete acknowledgment protocol from remaining drain guidance
  cmd/sdlc/helptext/queue.md:70-71 still instructs workers to "acknowledge the authenticated drain ID"; go run ./cmd/sdlc queue --help exits 0 and renders that instruction. The other remaining instance is atlas/workflow/sdlc-binary.md:2150, which describes a supervisor "notice/ack formatter". These conflict with cmd/tl-worker/help.md:122 and leave the ARCH-DRY/ARCH-PURPOSE contract sweep incomplete. Replace both restatements with pointers to the authoritative worker/Couch procedures, and extend rendered-help regression coverage to reject the removed acknowledgment requirement in queue help. Preserve the deliberately historical fixtures.

## Open findings

- **BR-1** [Important] `drain-contract-consistency` (kind: defect) Remove the obsolete acknowledgment protocol from remaining drain guidance
