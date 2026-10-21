---
gate: boundary-review
issue: 524
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: bb628093-9896-4c9a-bff2-484e823157c2
            repository: github.com/xianxu/ariadne
            issue_id: "000524"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/9ea6798693905b79
            reviewed_head: be03c359a52d9f73806509cb326155edd7e70967
            slot: ariadne:3
            requested_at: "2026-10-21T12:07:53.965993Z"
          phase: finished
          pid: 75537
          started_at: "2026-10-21T12:07:54.11021Z"
          finished_at: "2026-10-21T12:10:28.998035Z"
          session_id: 01a15e5c-5f47-7f20-8abb-b2547dabed9f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: be03c359a52d9f73806509cb326155edd7e70967
        tracker_head: 0265ce812ca4e9785839d303a6b04bf4ef4c8bbe
        branch_point: 63ebb3385af267d64084921c6e8d1150a188c734
        since: "2026-10-21T04:42:56-07:00"
        until: "2026-10-21T05:07:20-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:42:56-07:00"
              end: "2026-10-21T05:07:20-07:00"
              began: "2026-10-21T04:42:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 22d0a67c20d8e049e7c009e5d2a2e7d40eb355c2631382f03e62ef4876e4de86
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 111403668
              windows:
                - start: "2026-10-21T04:42:56-07:00"
                  end: "2026-10-21T05:07:20-07:00"
            - harness: codex
              id: 01a15e49-a810-7030-882b-a3730b969269
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 7640be368b0dcba34d16bbaf45eeffe4b139d26c0d3df6f5bbac547f0ce5d8e8
              started: "2026-10-21T11:47:28.214Z"
              spawned: false
              bytes: 610529
              windows:
                - start: "2026-10-21T11:47:28.214Z"
                  end: "2026-10-21T05:07:20-07:00"
            - harness: codex
              id: 01a15e4c-2f6f-7b32-8342-ddb6680a0213
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6424b68893ef2158a339ccdd018088d74402e5f6b1a9a92e33b32320bf93bfe2
              started: "2026-10-21T11:50:13.931Z"
              spawned: false
              bytes: 446975
              windows:
                - start: "2026-10-21T11:50:13.931Z"
                  end: "2026-10-21T05:07:20-07:00"
            - harness: codex
              id: 01a15e4d-ca88-7cd2-99df-2aa22f58516a
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 78a26db060628253c048662e72898683455ac4ea3989dff2aebb9a9aa21cb077
              started: "2026-10-21T11:51:59.23Z"
              spawned: false
              bytes: 421816
              windows:
                - start: "2026-10-21T11:51:59.23Z"
                  end: "2026-10-21T05:07:20-07:00"
            - harness: codex
              id: 01a15e4e-4de2-7cb2-84aa-011086f3ced2
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 9f3bd5c328d7cde086fb1e697a8542fec1e74685f1082c28d9c0d2034d045a82
              started: "2026-10-21T11:52:32.902Z"
              spawned: false
              bytes: 383643
              windows:
                - start: "2026-10-21T11:52:32.902Z"
                  end: "2026-10-21T05:07:20-07:00"
      "n": 1
      timestamp: "2026-10-21T05:10:29-07:00"
      dispatched: "2026-10-21T05:07:53-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Implicit saved-proof selection bypasses unsupported evidence linked to its parent
          detail: 'cmd/sdlc/localcheck_republish.go:73 checks unsupported authority before selecting a supported attempt, using an empty ID for implicit selection. A decoder-accepted future record with the same repository/head, check "vm", and rerun_of naming a supported suite parent is therefore missed: proofError only consults byParent when an ID is supplied (teststats_evidence.go:72-78), and the suite subject lookup excludes "vm". loadSavedAttempt returns the parent as newest, and guardSuitePass repeats the ID-less lookup, allowing success publication despite linked uncertainty. Check the selected candidate''s ID before returning it and make the success guard enforce linked-parent authority, or reject inconsistent parent/check identities as unscopable. Add a production-consumer regression covering explicit and implicit selection of this fixture. ARCH-SECURE / ARCH-PURPOSE.'
          family: linked-proof-authority
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: The completed M1 old-reader control is not retained
          detail: workshop/plans/000524-scoped-evidence-reruns-plan.md:90 marks a pre-change decoder/guard control complete, but teststats_evidence_test.go:65-71 only checks decodeSuiteAttempt rejection before invoking the upgraded consumers. Decoder rejection does not prove the old reader's global warning, saved-proof incompleteness or unrelated success refusal. Retain a reproducible control against the pinned pre-change reader/guard and compare it with the upgraded reader on equivalent mixed-store fixtures; record the commands and results before marking this requirement complete. ARCH-PURPOSE.
          family: compatibility-control-evidence
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#524 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-21T05:10:29-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `linked-proof-authority` (kind: defect) Implicit saved-proof selection bypasses unsupported evidence linked to its parent
  cmd/sdlc/localcheck_republish.go:73 checks unsupported authority before selecting a supported attempt, using an empty ID for implicit selection. A decoder-accepted future record with the same repository/head, check "vm", and rerun_of naming a supported suite parent is therefore missed: proofError only consults byParent when an ID is supplied (teststats_evidence.go:72-78), and the suite subject lookup excludes "vm". loadSavedAttempt returns the parent as newest, and guardSuitePass repeats the ID-less lookup, allowing success publication despite linked uncertainty. Check the selected candidate's ID before returning it and make the success guard enforce linked-parent authority, or reject inconsistent parent/check identities as unscopable. Add a production-consumer regression covering explicit and implicit selection of this fixture. ARCH-SECURE / ARCH-PURPOSE.
- **BR-2** [Important] `compatibility-control-evidence` (kind: defect) The completed M1 old-reader control is not retained
  workshop/plans/000524-scoped-evidence-reruns-plan.md:90 marks a pre-change decoder/guard control complete, but teststats_evidence_test.go:65-71 only checks decodeSuiteAttempt rejection before invoking the upgraded consumers. Decoder rejection does not prove the old reader's global warning, saved-proof incompleteness or unrelated success refusal. Retain a reproducible control against the pinned pre-change reader/guard and compare it with the upgraded reader on equivalent mixed-store fixtures; record the commands and results before marking this requirement complete. ARCH-PURPOSE.

## Open findings

- **BR-1** [Critical] `linked-proof-authority` (kind: defect) Implicit saved-proof selection bypasses unsupported evidence linked to its parent
- **BR-2** [Important] `compatibility-control-evidence` (kind: defect) The completed M1 old-reader control is not retained
