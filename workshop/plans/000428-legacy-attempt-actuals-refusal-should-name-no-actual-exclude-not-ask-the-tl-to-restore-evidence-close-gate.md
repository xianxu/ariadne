---
gate: boundary-review
issue: 428
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: ec464232-8e97-449f-8bae-6102aef6414e
            repository: github.com/xianxu/ariadne
            issue_id: "000428"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/e14a2e514f1b61c9
            reviewed_head: b0119e262de3e2a86679ddc24182bc970f1ace75
            slot: ariadne:1
            requested_at: "2026-10-17T19:33:07.400287Z"
          phase: finished
          pid: 6070
          started_at: "2026-10-17T19:33:07.493936Z"
          finished_at: "2026-10-17T19:34:43.668604Z"
          session_id: 01a14b5a-8c08-7643-800a-c4e458535dc6
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: b0119e262de3e2a86679ddc24182bc970f1ace75
        tracker_head: 4786741d70c49f70a4527aac726511827efc40b5
        branch_point: f0c955edb8014d073ebe7f007a4eee3d6f319dca
        since: "2026-10-17T11:19:16-07:00"
        until: "2026-10-17T11:53:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-17T11:19:16-07:00"
              end: "2026-10-17T11:53:10-07:00"
              began: "2026-10-17T11:19:16-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: d79477b6733a4952f7fdebc4cc067f0db3be28d4288532e4e8a6ff26e5851a4c
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 66341113
              windows:
                - start: "2026-10-17T11:19:16-07:00"
                  end: "2026-10-17T11:53:10-07:00"
        queue_waits:
            - attempt_id: ae857f6c213e0d5db96163a66c76c1f9
              slot: ariadne:1
              record_digest: c5645ecddef9d464f468743115959c36d2a327b69df29c09a8019080445b04f0
              start: "2026-10-17T18:24:44.272753Z"
              end: "2026-10-17T18:42:02.12819Z"
              queue_seconds: 1037.855437
              raw_seconds: 1037.855437
              exact: true
      "n": 1
      timestamp: "2026-10-17T12:34:44-07:00"
      dispatched: "2026-10-17T12:33:07-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Legacy exclusion advice is selected without checking every blocker
          detail: 'cmd/sdlc/actual.go:381 classifies the window from one diagnostic, while cmd/sdlc/actualinputs.go:340 forwards only notes[0]. A temporary regression test using the production selector reproduced a legacy attempt followed by an exact attempt with unknown slot provenance: the output recommends exclusion and suppresses recovery guidance; reversing the order retains recovery guidance. This violates the Spec''s all-legacy condition. The family includes legacy-first mixtures with either other selector-warning category (unknown queue provenance or unknown slot provenance), across actual previews with and without --milestone, issue close, and milestone close. Preserve and classify the complete blocker set, select exclusion-only guidance only when all blockers are legacy, and test both orders.'
          family: complete-blocker-classification
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#428 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T12:34:44-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `complete-blocker-classification` (kind: defect) Legacy exclusion advice is selected without checking every blocker
  cmd/sdlc/actual.go:381 classifies the window from one diagnostic, while cmd/sdlc/actualinputs.go:340 forwards only notes[0]. A temporary regression test using the production selector reproduced a legacy attempt followed by an exact attempt with unknown slot provenance: the output recommends exclusion and suppresses recovery guidance; reversing the order retains recovery guidance. This violates the Spec's all-legacy condition. The family includes legacy-first mixtures with either other selector-warning category (unknown queue provenance or unknown slot provenance), across actual previews with and without --milestone, issue close, and milestone close. Preserve and classify the complete blocker set, select exclusion-only guidance only when all blockers are legacy, and test both orders.

## Open findings

- **BR-1** [Critical] `complete-blocker-classification` (kind: defect) Legacy exclusion advice is selected without checking every blocker
