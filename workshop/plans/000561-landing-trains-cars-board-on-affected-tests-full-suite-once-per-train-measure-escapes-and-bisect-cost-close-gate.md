---
gate: boundary-review
issue: 561
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1fba3996-3fc4-4c88-a97e-1502c04ab174
            repository: github.com/xianxu/ariadne
            issue_id: "000561"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/cede9040195488b6
            reviewed_head: 2e6f906cbd6703002ee909b8691d165781af53b1
            slot: ariadne:2
            requested_at: "2026-10-23T08:29:48.512331Z"
          phase: finished
          pid: 8850
          started_at: "2026-10-23T08:29:48.599274Z"
          finished_at: "2026-10-23T08:34:48.559738Z"
          session_id: 059fe0c8-acb9-44eb-bb7b-1232faa0e1ed
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2e6f906cbd6703002ee909b8691d165781af53b1
        tracker_head: 18021d93093587e9677aadc631edfa5b78e5ccdc
        branch_point: 6318fc7bc700173acdcea67f3993d326f8947845
        since: "2026-10-23T00:12:42-07:00"
        until: "2026-10-23T01:29:04-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T00:12:42-07:00"
              end: "2026-10-23T01:29:04-07:00"
              began: "2026-10-23T00:12:42-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 35e15fe2-8f0d-4566-952d-248079e77abe
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 0851f315479769063bc9ac22dcb62b5561890d2be8b96e0c2ae9b939592c9403
              started: "2026-10-23T07:50:04.212Z"
              spawned: false
              bytes: 392978
              windows:
                - start: "2026-10-23T07:50:04.212Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ac60500aa08d9059ee23e51ca637753c08468c337b271385fdcadad6cde4a691
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 1876292
              windows:
                - start: "2026-10-23T07:49:37.819Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 209724edab26b675f92f1d17f3b4efd6d07e87ff44f2088fdbe20366b9b231f1
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 70313740
              windows:
                - start: "2026-10-23T00:12:42-07:00"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a1679b-daf3-7670-8b64-92acffdd9912
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 8600e756d7e557ec4b2e56aa6102e78a69d16bfebd0d438c358850dda70634f5
              started: "2026-10-23T07:13:49.893Z"
              spawned: true
              bytes: 704561
              windows:
                - start: "2026-10-23T07:13:49.893Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a1679c-0bce-73c3-a3b1-deb7349cfd29
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: e1cd9dde0d2f0ecbbd3440c6e23d690c09e3ea0f6a31707c3d0473f44b497004
              started: "2026-10-23T07:14:02.415Z"
              spawned: true
              bytes: 804726
              windows:
                - start: "2026-10-23T07:14:02.415Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a167a9-1cfd-7c42-b326-59702214ac99
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8dedd1bfe31983481bc0d4586b86f5df76dbf124a81decce1ce6f6037421c922
              started: "2026-10-23T07:28:19.079Z"
              spawned: false
              bytes: 919808
              windows:
                - start: "2026-10-23T07:28:19.079Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a167af-04b8-7e72-81d0-27b0a05cb910
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6a38189077060d4a9682ad7e4e739a16d36542f9442b52793311f9297f86f4cb
              started: "2026-10-23T07:34:46.026Z"
              spawned: false
              bytes: 501476
              windows:
                - start: "2026-10-23T07:34:46.026Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a167b1-2df2-7dd1-9ce4-befd102b8266
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ba8418ea76d3828851c4914c761562d877653459f8d0a3ef9591e956bdc02f30
              started: "2026-10-23T07:37:07.828Z"
              spawned: false
              bytes: 454992
              windows:
                - start: "2026-10-23T07:37:07.828Z"
                  end: "2026-10-23T01:29:04-07:00"
            - harness: codex
              id: 01a167b1-cacb-7f63-ad15-9a2f3efef053
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: d6391bbc2ddf0bc306e40693531bc098f3f3f129848d59e87215614ee803e1ac
              started: "2026-10-23T07:37:47.761Z"
              spawned: false
              bytes: 413051
              windows:
                - start: "2026-10-23T07:37:47.761Z"
                  end: "2026-10-23T01:29:04-07:00"
      "n": 1
      timestamp: "2026-10-23T01:34:49-07:00"
      dispatched: "2026-10-23T01:29:48-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic,openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: README still documents train_max_cars default 8 and omits --boarding
          detail: README.md:661 says "train_max_cars (default 8)" but landingline.DefaultMaxCars is now 2 (train.go:39); the affected-suite section near README.md:847 has no mention of --boarding, its full-selection refusal, or the seven-day manual ramp rule. Update to default 2 (#561) and add a short --boarding paragraph mirroring helptext/local-check.md.
          family: docs-drift-on-changed-default
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: cigate.go:675 hand-builds a local-check command bypassing localCheckCommand
          detail: For the affected stand-in this would recommend a run without --boarding; use localCheckCommand(name).
          family: recovery-command-single-source
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: Help rendering of TRAIN_RAMP_POLICY and wrapHelpText has no test
          detail: Only status asserts the policy text; nothing pins that sdlc landing --help renders it or that wrapping preserves the words.
          family: derived-surface-untested
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: runLocalCheckCmd passing mode into runAffectedCheck is not tested end to end
          detail: Tests call runAffectedCheck directly; passing affectedOrdinary at localcheck_cmd.go:242 would keep every test green.
          family: cli-wiring-untested
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 378110e8-62aa-4de0-88b1-46e5ec174f8a
            repository: github.com/xianxu/ariadne
            issue_id: "000561"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/b8f0a912985f0155
            reviewed_head: ac4b48a62f178e996cfbf400aa4e933fb1eb063d
            slot: ariadne:2
            requested_at: "2026-10-23T08:36:12.375477Z"
          phase: finished
          pid: 71608
          started_at: "2026-10-23T08:36:12.476121Z"
          finished_at: "2026-10-23T08:39:54.894024Z"
          session_id: c8f69ecb-422a-41f0-aa19-195912bbc13d
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: ac4b48a62f178e996cfbf400aa4e933fb1eb063d
        tracker_head: 410094be74c3d6a671b27c10edd9e8f052de6f96
        branch_point: 6318fc7bc700173acdcea67f3993d326f8947845
        since: "2026-10-23T00:12:42-07:00"
        until: "2026-10-23T01:35:50-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T00:12:42-07:00"
              end: "2026-10-23T01:35:50-07:00"
              began: "2026-10-23T00:12:42-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 35e15fe2-8f0d-4566-952d-248079e77abe
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 0851f315479769063bc9ac22dcb62b5561890d2be8b96e0c2ae9b939592c9403
              started: "2026-10-23T07:50:04.212Z"
              spawned: false
              bytes: 392978
              windows:
                - start: "2026-10-23T07:50:04.212Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: claude
              id: a26a620a-c2f7-4a53-b072-dec0ffd89fd2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c7e689fef70fa82b33e426827ab44eac1dbeb4543737a02810f8991dffebf49a
              started: "2026-10-23T07:49:37.819Z"
              spawned: false
              bytes: 2657187
              windows:
                - start: "2026-10-23T07:49:37.819Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 209724edab26b675f92f1d17f3b4efd6d07e87ff44f2088fdbe20366b9b231f1
              started: "2026-10-20T18:21:54.178Z"
              spawned: false
              bytes: 70313740
              windows:
                - start: "2026-10-23T00:12:42-07:00"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a1679b-daf3-7670-8b64-92acffdd9912
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: 8600e756d7e557ec4b2e56aa6102e78a69d16bfebd0d438c358850dda70634f5
              started: "2026-10-23T07:13:49.893Z"
              spawned: true
              bytes: 704561
              windows:
                - start: "2026-10-23T07:13:49.893Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a1679c-0bce-73c3-a3b1-deb7349cfd29
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              parent_id: 01a15a8b-4413-79d1-8ce1-b0de05e9b8ca
              digest: e1cd9dde0d2f0ecbbd3440c6e23d690c09e3ea0f6a31707c3d0473f44b497004
              started: "2026-10-23T07:14:02.415Z"
              spawned: true
              bytes: 804726
              windows:
                - start: "2026-10-23T07:14:02.415Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a167a9-1cfd-7c42-b326-59702214ac99
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 8dedd1bfe31983481bc0d4586b86f5df76dbf124a81decce1ce6f6037421c922
              started: "2026-10-23T07:28:19.079Z"
              spawned: false
              bytes: 919808
              windows:
                - start: "2026-10-23T07:28:19.079Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a167af-04b8-7e72-81d0-27b0a05cb910
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6a38189077060d4a9682ad7e4e739a16d36542f9442b52793311f9297f86f4cb
              started: "2026-10-23T07:34:46.026Z"
              spawned: false
              bytes: 501476
              windows:
                - start: "2026-10-23T07:34:46.026Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a167b1-2df2-7dd1-9ce4-befd102b8266
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ba8418ea76d3828851c4914c761562d877653459f8d0a3ef9591e956bdc02f30
              started: "2026-10-23T07:37:07.828Z"
              spawned: false
              bytes: 454992
              windows:
                - start: "2026-10-23T07:37:07.828Z"
                  end: "2026-10-23T01:35:50-07:00"
            - harness: codex
              id: 01a167b1-cacb-7f63-ad15-9a2f3efef053
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: d6391bbc2ddf0bc306e40693531bc098f3f3f129848d59e87215614ee803e1ac
              started: "2026-10-23T07:37:47.761Z"
              spawned: false
              bytes: 413051
              windows:
                - start: "2026-10-23T07:37:47.761Z"
                  end: "2026-10-23T01:35:50-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-23T01:29:48-07:00"
              finalized: "2026-10-23T01:34:49-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T01:39:56-07:00"
      dispatched: "2026-10-23T01:36:12-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic,openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: README.md:661 now says default 2 since 561 and keeps explicit caps, adds the manual seven-day ramp rule, and README.md:858 documents --boarding refusal and its alternative; matches DefaultMaxCars=2 and helptext/local-check.md.
          round: 2
        - id: BR-2
          disposition: not-addressed
          note: cigate.go:675 now uses localCheckCommand(name), but no test asserts this message; restoring the hand-built string keeps every test green.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestLandingHelpRendersTrainRampPolicy pins rendering and wrapping; removing the placeholder from landing.md turns it red (verified).
          round: 2
        - id: BR-4
          disposition: addressed
          note: TestLocalCheckCmdBoardingRefusesFullSelection drives runLocalCheckCmd; hard-coding affectedOrdinary at localcheck_cmd.go:242 turns it red (verified).
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Minor
          title: Recovery advice still hand-builds local-check run commands outside localCheckCommand
          detail: '2nd finding in this family. suiteenvironment.go:49 and localcheck_reclassify.go:70 format sdlc local-check run --name X by hand; isTestCheck includes affected, so a boarding car can be told to rerun affected without --boarding. Rule: every user-facing local-check run command derives from localCheckCommand (extended with optional --rev), enforced by a guard test that scans non-test cmd/sdlc sources for the literal outside that helper. Prevalence: 2 remaining sites after the BR-2 fix.'
          family: recovery-command-single-source
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: ac4b48a62f178e996cfbf400aa4e933fb1eb063d
      blocked: false
---

# Gate ledger — ariadne#561 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T01:34:49-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `docs-drift-on-changed-default` (kind: defect) README still documents train_max_cars default 8 and omits --boarding
  README.md:661 says "train_max_cars (default 8)" but landingline.DefaultMaxCars is now 2 (train.go:39); the affected-suite section near README.md:847 has no mention of --boarding, its full-selection refusal, or the seven-day manual ramp rule. Update to default 2 (#561) and add a short --boarding paragraph mirroring helptext/local-check.md.
- **BR-2** [Minor] `recovery-command-single-source` (kind: defect) cigate.go:675 hand-builds a local-check command bypassing localCheckCommand
  For the affected stand-in this would recommend a run without --boarding; use localCheckCommand(name).
- **BR-3** [Minor] `derived-surface-untested` (kind: defect) Help rendering of TRAIN_RAMP_POLICY and wrapHelpText has no test
  Only status asserts the policy text; nothing pins that sdlc landing --help renders it or that wrapping preserves the words.
- **BR-4** [Minor] `cli-wiring-untested` (kind: defect) runLocalCheckCmd passing mode into runAffectedCheck is not tested end to end
  Tests call runAffectedCheck directly; passing affectedOrdinary at localcheck_cmd.go:242 would keep every test green.

## Round 2 — 2026-10-23T01:39:56-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — README.md:661 now says default 2 since 561 and keeps explicit caps, adds the manual seven-day ramp rule, and README.md:858 documents --boarding refusal and its alternative; matches DefaultMaxCars=2 and helptext/local-check.md.
- BR-2 — not-addressed — cigate.go:675 now uses localCheckCommand(name), but no test asserts this message; restoring the hand-built string keeps every test green.
- BR-3 — addressed — TestLandingHelpRendersTrainRampPolicy pins rendering and wrapping; removing the placeholder from landing.md turns it red (verified).
- BR-4 — addressed — TestLocalCheckCmdBoardingRefusesFullSelection drives runLocalCheckCmd; hard-coding affectedOrdinary at localcheck_cmd.go:242 turns it red (verified).

### Raised

- **BR-5** [Minor] `recovery-command-single-source` (kind: defect) Recovery advice still hand-builds local-check run commands outside localCheckCommand
  2nd finding in this family. suiteenvironment.go:49 and localcheck_reclassify.go:70 format sdlc local-check run --name X by hand; isTestCheck includes affected, so a boarding car can be told to rerun affected without --boarding. Rule: every user-facing local-check run command derives from localCheckCommand (extended with optional --rev), enforced by a guard test that scans non-test cmd/sdlc sources for the literal outside that helper. Prevalence: 2 remaining sites after the BR-2 fix.

## Open findings

- **BR-2** [Minor] `recovery-command-single-source` (kind: defect) cigate.go:675 hand-builds a local-check command bypassing localCheckCommand
- **BR-5** [Minor] `recovery-command-single-source` (kind: defect) Recovery advice still hand-builds local-check run commands outside localCheckCommand
