---
gate: boundary-review
issue: 482
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 01315642-de00-4f87-b471-23668c374b18
            repository: github.com/xianxu/ariadne
            issue_id: "000482"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b3516f20cd8e94c2
            reviewed_head: 885657462e486119f46f41619a0b894f5c8af136
            slot: ariadne:4
            requested_at: "2026-10-20T06:28:14.624786Z"
          phase: finished
          pid: 25939
          started_at: "2026-10-20T06:28:14.760979Z"
          finished_at: "2026-10-20T06:29:05.804689Z"
          session_id: b6973a74-5ca9-49ce-ba8d-467e21a1a142
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 885657462e486119f46f41619a0b894f5c8af136
        tracker_head: e726af3e0747e9fb4f7f3f7bcb3ebf3194d1741d
        branch_point: 718afe6dbad3fb6db1b71981e6402a6a0bce8c92
        since: "2026-10-19T22:48:28-07:00"
        until: "2026-10-19T22:49:30-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T22:48:28-07:00"
              end: "2026-10-19T22:49:30-07:00"
              began: "2026-10-19T22:48:28-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: claude
              id: 299d33d0-1f10-4950-b82c-b729d6c107a0
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: a1cd97974f5b010db049e2907b2c0a0b9f6b39107ddb36f28a65b4f2695266a8
              started: "2026-10-16T06:28:11.678Z"
              spawned: false
              bytes: 93315715
              windows:
                - start: "2026-10-19T22:48:28-07:00"
                  end: "2026-10-19T22:49:30-07:00"
      "n": 1
      timestamp: "2026-10-19T23:29:06-07:00"
      dispatched: "2026-10-19T23:28:14-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Soak hold is to be written in the binary-rendered Slots section, which overwrites it and lists the held slot as free
          detail: 'help.md:424 says to show "hold: soak until X" on the slot''s line in the Slots section, but tl-lead status slots re-renders that whole section every pass (slots.go:61-117, status.go:320, help.md:266-274). An idle slot on its resting branch goes on the free: line (slots.go:71-72), so the hand-written note is lost and the Spec clause "the hold shows in the status file''s Slots line" is not delivered. Second instance of the same family: help.md:397''s free-for-new-work definition was not updated to exclude held slots, so a held slot can still be dispatched. Fix: record the hold in a hand-written section and amend help.md:397, or teach renderSlots a hold marker and pin it with a slots test.'
          family: hand-state-in-rendered-section
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: No atlas pointer for the new keep-conversation and soak-hold tl-lead rules
          detail: atlas/workflow/tl-skills.md points to other help.md rules (for example the 433 lifecycle notices) but was not updated for these two new rules.
          family: atlas-pointer-for-help-rules
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#482 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T23:29:06-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `hand-state-in-rendered-section` (kind: defect) Soak hold is to be written in the binary-rendered Slots section, which overwrites it and lists the held slot as free
  help.md:424 says to show "hold: soak until X" on the slot's line in the Slots section, but tl-lead status slots re-renders that whole section every pass (slots.go:61-117, status.go:320, help.md:266-274). An idle slot on its resting branch goes on the free: line (slots.go:71-72), so the hand-written note is lost and the Spec clause "the hold shows in the status file's Slots line" is not delivered. Second instance of the same family: help.md:397's free-for-new-work definition was not updated to exclude held slots, so a held slot can still be dispatched. Fix: record the hold in a hand-written section and amend help.md:397, or teach renderSlots a hold marker and pin it with a slots test.
- **BR-2** [Minor] `atlas-pointer-for-help-rules` (kind: defect) No atlas pointer for the new keep-conversation and soak-hold tl-lead rules
  atlas/workflow/tl-skills.md points to other help.md rules (for example the 433 lifecycle notices) but was not updated for these two new rules.

## Open findings

- **BR-1** [Important] `hand-state-in-rendered-section` (kind: defect) Soak hold is to be written in the binary-rendered Slots section, which overwrites it and lists the held slot as free
- **BR-2** [Minor] `atlas-pointer-for-help-rules` (kind: defect) No atlas pointer for the new keep-conversation and soak-hold tl-lead rules
