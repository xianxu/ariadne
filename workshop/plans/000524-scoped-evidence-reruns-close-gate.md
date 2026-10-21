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
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 38541188-1e97-43fd-bbe1-2a6102f6957c
            repository: github.com/xianxu/ariadne
            issue_id: "000524"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/d83c023049a68299
            reviewed_head: 744236f93156b5662e61fffdd8bfbad206e33b96
            slot: ariadne:3
            requested_at: "2026-10-21T12:15:58.292339Z"
          phase: finished
          pid: 76466
          started_at: "2026-10-21T12:15:58.40264Z"
          finished_at: "2026-10-21T12:18:18.619684Z"
          session_id: 01a15e63-c2c9-7fb3-80c2-4b3934137f9c
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 744236f93156b5662e61fffdd8bfbad206e33b96
        tracker_head: 0265ce812ca4e9785839d303a6b04bf4ef4c8bbe
        branch_point: 63ebb3385af267d64084921c6e8d1150a188c734
        since: "2026-10-21T04:42:56-07:00"
        until: "2026-10-21T05:15:33-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:42:56-07:00"
              end: "2026-10-21T05:15:33-07:00"
              began: "2026-10-21T04:42:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 5f47b3cbcc021c6ccd1378eaa0fadf9627ddc9a63f1f58ca22a4deaa884105f6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 111937757
              windows:
                - start: "2026-10-21T04:42:56-07:00"
                  end: "2026-10-21T05:15:33-07:00"
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
                  end: "2026-10-21T05:15:33-07:00"
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
                  end: "2026-10-21T05:15:33-07:00"
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
                  end: "2026-10-21T05:15:33-07:00"
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
                  end: "2026-10-21T05:15:33-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T05:07:53-07:00"
              finalized: "2026-10-21T05:10:29-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-21T05:18:19-07:00"
      dispatched: "2026-10-21T05:15:58-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestScopedSavedProofConsumers covers explicit/implicit selection, reuse and publication. Removing the parent-index invocation in a disposable pinned-head clone made the cross-check regression fail with unauthorized publication; the unchanged head passes.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Ran sh scripts/verify-pre524-reader.sh successfully against pinned pre-change source. Its retained test asserts the global warning, incomplete saved proof and guard/publication refusal; the equivalent upgraded-reader cases pass.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Important
          title: README update appears missing for the compatibility verification command
          detail: scripts/verify-pre524-reader.sh:1 introduces an operator-run verification command, documented at atlas/workflow/sdlc-binary.md:2646, but README.md has no change in the pinned range. Add a brief README usage entry or link covering the command and reader-first rollout barrier to satisfy the explicit README gate.
          family: user-surface-documentation
          round: 2
        - kind: defect
          id: BR-4
          severity: Minor
          title: Compatibility-control clones have no defined cleanup lifecycle
          detail: 'scripts/verify-pre524-reader.sh:7 creates a fresh clone on every invocation and retains it on success or failure without a cleanup path. ARCH-FUNERAL: define retention through inspection and its removal owner/command, or clean by default with an explicit retain option; update the plan''s claim that this work creates no new artifact family.'
          family: artifact-lifecycle
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 69a1d4f2-1d06-4fcf-ac74-4aa6a93712d0
            repository: github.com/xianxu/ariadne
            issue_id: "000524"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/8079621f91e40a08
            reviewed_head: f6dc3de33057609c69f6e85f4d71d6df9a8731ee
            slot: ariadne:3
            requested_at: "2026-10-21T14:25:14.110269Z"
          phase: finished
          pid: 83934
          started_at: "2026-10-21T14:25:14.233495Z"
          finished_at: "2026-10-21T14:27:35.027303Z"
          session_id: 01a15eda-1b45-79d1-9459-eb9a5806361a
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: f6dc3de33057609c69f6e85f4d71d6df9a8731ee
        tracker_head: e80b7aa63ef575209931b9232ccdbe9ccacbc3cf
        branch_point: 63ebb3385af267d64084921c6e8d1150a188c734
        since: "2026-10-21T04:42:56-07:00"
        until: "2026-10-21T05:22:11-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:42:56-07:00"
              end: "2026-10-21T05:22:11-07:00"
              began: "2026-10-21T04:42:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:42:56-07:00"
                  end: "2026-10-21T05:22:11-07:00"
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
                  end: "2026-10-21T05:22:11-07:00"
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
                  end: "2026-10-21T05:22:11-07:00"
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
                  end: "2026-10-21T05:22:11-07:00"
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
                  end: "2026-10-21T05:22:11-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T05:07:53-07:00"
              finalized: "2026-10-21T05:10:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T05:15:58-07:00"
              finalized: "2026-10-21T05:18:19-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-21T07:27:35-07:00"
      dispatched: "2026-10-21T07:25:14-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Production-consumer regressions pass. Removing parent-index resolution in a disposable pinned-head clone makes both cross-check publication and inverse-parent scope regressions fail.
          round: 3
        - id: BR-2
          disposition: addressed
          note: The committed compatibility fixture passes against pinned pre-change source, demonstrating global warnings, incomplete saved proof and publication refusal. Upgraded consumer tests pass.
          round: 3
        - id: BR-3
          disposition: addressed
          note: README.md:765 documents the verification command and reader-first rollout barrier, matching scripts/verify-pre524-reader.sh:5-15 and atlas/workflow/sdlc-binary.md:2637-2651.
          round: 3
        - id: BR-4
          disposition: addressed
          note: scripts/verify-pre524-reader.sh:13 names invoker cleanup; atlas/workflow/sdlc-binary.md:2653 defines retention through inspection and exact-directory removal. Plan lines 108 and 128 acknowledge the private clone lifecycle.
          round: 3
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: f6dc3de33057609c69f6e85f4d71d6df9a8731ee
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: d26bc2cc-ad3e-42d2-91cd-7632fe3243be
            repository: github.com/xianxu/ariadne
            issue_id: "000524"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/94256da7abb451c1
            reviewed_head: 1ff8b62a01066df5152816acc8c2d2b6880bd005
            slot: ariadne:3
            requested_at: "2026-10-21T14:43:52.827599Z"
          phase: finished
          pid: 19966
          started_at: "2026-10-21T14:43:52.933755Z"
          finished_at: "2026-10-21T14:47:37.188881Z"
          session_id: 01a15eeb-2cf5-7ab1-ab83-2cde8a2f21f9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1ff8b62a01066df5152816acc8c2d2b6880bd005
        tracker_head: e80b7aa63ef575209931b9232ccdbe9ccacbc3cf
        branch_point: 63ebb3385af267d64084921c6e8d1150a188c734
        since: "2026-10-21T04:42:56-07:00"
        until: "2026-10-21T07:43:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:42:56-07:00"
              end: "2026-10-21T07:43:31-07:00"
              began: "2026-10-21T04:42:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:42:56-07:00"
                  end: "2026-10-21T07:43:31-07:00"
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
                  end: "2026-10-21T07:43:31-07:00"
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
                  end: "2026-10-21T07:43:31-07:00"
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
                  end: "2026-10-21T07:43:31-07:00"
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
                  end: "2026-10-21T07:43:31-07:00"
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: eae658a57308dc6ca62f6101eee1f3664cac87c9329d68384eb609a254358616
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 2020480
              windows:
                - start: "2026-10-21T14:24:08.011Z"
                  end: "2026-10-21T07:43:31-07:00"
            - harness: codex
              id: 01a15edc-a908-7d03-b74b-f339fff0b5b0
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 362cfe17f9121a3605f424b067f7f672310fd5367618939b67b6e812635d892f
              started: "2026-10-21T14:28:02.073Z"
              spawned: true
              bytes: 564661
              windows:
                - start: "2026-10-21T14:28:02.073Z"
                  end: "2026-10-21T07:43:31-07:00"
            - harness: codex
              id: 01a15ee2-6d4d-7563-bfbd-69a76a92de09
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: c335cd2f156a43a89953e408a7ac9af51391586c682d46dbda43052bc75c5c3d
              started: "2026-10-21T14:34:20.012Z"
              spawned: true
              bytes: 888567
              windows:
                - start: "2026-10-21T14:34:20.012Z"
                  end: "2026-10-21T07:43:31-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T05:07:53-07:00"
              finalized: "2026-10-21T05:10:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T05:15:58-07:00"
              finalized: "2026-10-21T05:18:19-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T07:25:14-07:00"
              finalized: "2026-10-21T07:27:35-07:00"
              finalizing: true
      "n": 4
      timestamp: "2026-10-21T07:47:38-07:00"
      dispatched: "2026-10-21T07:43:52-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: PASS collection rejects valid Go subtest events and invents required ancestors
          detail: 'cmd/sdlc/internal/flakes/rerun.go:104-108 validates every terminal Test value using the narrower registry-name grammar: a passing t.Run("case(value)") invalidates the entire log, even when only its valid top-level test requires proof. Lines 124-132 also require a PASS event for every slash prefix, but t.Run("path/file") emits TestSlash/path/file and TestSlash, with no TestSlash/path event. Real go test -json runs exited zero in both cases; temporary overlays extending TestKnownFlakeExactPackageGate and TestKnownFlakeGateSubtest reproduced production outcomes of known-flake-rerun-inconclusive. ARCH-PURPOSE/ARCH-MOCK: distinguish registry eligibility from valid tool output, and require actual test outcomes without fabricating ancestors from name separators. Sweep legal subtest-name forms and add real-Go regressions for both cases while preserving package binding, conflict rejection, and sibling-failure blocking.'
          family: external-event-contract-fidelity
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 95170fca-1f42-4faa-904f-b382dfd98888
            repository: github.com/xianxu/ariadne
            issue_id: "000524"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/d12f61852ea4d6be
            reviewed_head: 2750b1ed88b29c30a74bb781578b2ba59150f7c1
            slot: ariadne:3
            requested_at: "2026-10-21T14:53:04.523798Z"
          phase: finished
          pid: 50917
          started_at: "2026-10-21T14:53:04.632039Z"
          finished_at: "2026-10-21T14:58:52.410133Z"
          session_id: 01a15ef3-9826-78f1-8df0-9996f5cc3104
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 2750b1ed88b29c30a74bb781578b2ba59150f7c1
        tracker_head: e80b7aa63ef575209931b9232ccdbe9ccacbc3cf
        branch_point: 63ebb3385af267d64084921c6e8d1150a188c734
        since: "2026-10-21T04:42:56-07:00"
        until: "2026-10-21T07:52:27-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-21T04:42:56-07:00"
              end: "2026-10-21T07:52:27-07:00"
              began: "2026-10-21T04:42:56-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 6a2e2bbef93f77238828232c2bc7ef14fea9726407b6153fa21ccf3fdfa1834b
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 112292489
              windows:
                - start: "2026-10-21T04:42:56-07:00"
                  end: "2026-10-21T07:52:27-07:00"
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
                  end: "2026-10-21T07:52:27-07:00"
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
                  end: "2026-10-21T07:52:27-07:00"
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
                  end: "2026-10-21T07:52:27-07:00"
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
                  end: "2026-10-21T07:52:27-07:00"
            - harness: codex
              id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 85df962a862ad85939319427ee22a9074b409eb0200848bac4afd843035c0288
              started: "2026-10-21T14:24:08.011Z"
              spawned: false
              bytes: 2518293
              windows:
                - start: "2026-10-21T14:24:08.011Z"
                  end: "2026-10-21T07:52:27-07:00"
            - harness: codex
              id: 01a15edc-a908-7d03-b74b-f339fff0b5b0
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: 7aadeab570fff25f1eb3a09602fbedd505d30523fb4848566a4e13999bca2b96
              started: "2026-10-21T14:28:02.073Z"
              spawned: true
              bytes: 660206
              windows:
                - start: "2026-10-21T14:28:02.073Z"
                  end: "2026-10-21T07:52:27-07:00"
            - harness: codex
              id: 01a15ee2-6d4d-7563-bfbd-69a76a92de09
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a15ed8-b1d1-7783-855a-322febf4f2fe
              digest: c335cd2f156a43a89953e408a7ac9af51391586c682d46dbda43052bc75c5c3d
              started: "2026-10-21T14:34:20.012Z"
              spawned: true
              bytes: 888567
              windows:
                - start: "2026-10-21T14:34:20.012Z"
                  end: "2026-10-21T07:52:27-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-21T05:07:53-07:00"
              finalized: "2026-10-21T05:10:29-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T05:15:58-07:00"
              finalized: "2026-10-21T05:18:19-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-21T07:25:14-07:00"
              finalized: "2026-10-21T07:27:35-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-21T07:43:52-07:00"
              finalized: "2026-10-21T07:47:38-07:00"
              finalizing: false
      "n": 5
      timestamp: "2026-10-21T07:59:25-07:00"
      dispatched: "2026-10-21T07:53:04-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-5
          disposition: addressed
          note: Real-Go TestKnownFlakeExactPackageGate and TestKnownFlakeGateFlatSlashSubtest pass on the pinned head. Restoring the pre-fix collector in a disposable checkout makes both fail with known-flake-rerun-inconclusive; the pure regression also fails. Package binding, required-event conflicts, and sibling-failure controls remain green. The disposable checkout was removed.
          round: 5
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: 2750b1ed88b29c30a74bb781578b2ba59150f7c1
      blocked: false
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

## Round 2 — 2026-10-21T05:18:19-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — TestScopedSavedProofConsumers covers explicit/implicit selection, reuse and publication. Removing the parent-index invocation in a disposable pinned-head clone made the cross-check regression fail with unauthorized publication; the unchanged head passes.
- BR-2 — addressed — Ran sh scripts/verify-pre524-reader.sh successfully against pinned pre-change source. Its retained test asserts the global warning, incomplete saved proof and guard/publication refusal; the equivalent upgraded-reader cases pass.

### Raised

- **BR-3** [Important] `user-surface-documentation` (kind: defect) README update appears missing for the compatibility verification command
  scripts/verify-pre524-reader.sh:1 introduces an operator-run verification command, documented at atlas/workflow/sdlc-binary.md:2646, but README.md has no change in the pinned range. Add a brief README usage entry or link covering the command and reader-first rollout barrier to satisfy the explicit README gate.
- **BR-4** [Minor] `artifact-lifecycle` (kind: defect) Compatibility-control clones have no defined cleanup lifecycle
  scripts/verify-pre524-reader.sh:7 creates a fresh clone on every invocation and retains it on success or failure without a cleanup path. ARCH-FUNERAL: define retention through inspection and its removal owner/command, or clean by default with an explicit retain option; update the plan's claim that this work creates no new artifact family.

## Round 3 — 2026-10-21T07:27:35-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Production-consumer regressions pass. Removing parent-index resolution in a disposable pinned-head clone makes both cross-check publication and inverse-parent scope regressions fail.
- BR-2 — addressed — The committed compatibility fixture passes against pinned pre-change source, demonstrating global warnings, incomplete saved proof and publication refusal. Upgraded consumer tests pass.
- BR-3 — addressed — README.md:765 documents the verification command and reader-first rollout barrier, matching scripts/verify-pre524-reader.sh:5-15 and atlas/workflow/sdlc-binary.md:2637-2651.
- BR-4 — addressed — scripts/verify-pre524-reader.sh:13 names invoker cleanup; atlas/workflow/sdlc-binary.md:2653 defines retention through inspection and exact-directory removal. Plan lines 108 and 128 acknowledge the private clone lifecycle.

## Round 4 — 2026-10-21T07:47:38-07:00 (codex) — BLOCKED

### Raised

- **BR-5** [Critical] `external-event-contract-fidelity` (kind: defect) PASS collection rejects valid Go subtest events and invents required ancestors
  cmd/sdlc/internal/flakes/rerun.go:104-108 validates every terminal Test value using the narrower registry-name grammar: a passing t.Run("case(value)") invalidates the entire log, even when only its valid top-level test requires proof. Lines 124-132 also require a PASS event for every slash prefix, but t.Run("path/file") emits TestSlash/path/file and TestSlash, with no TestSlash/path event. Real go test -json runs exited zero in both cases; temporary overlays extending TestKnownFlakeExactPackageGate and TestKnownFlakeGateSubtest reproduced production outcomes of known-flake-rerun-inconclusive. ARCH-PURPOSE/ARCH-MOCK: distinguish registry eligibility from valid tool output, and require actual test outcomes without fabricating ancestors from name separators. Sweep legal subtest-name forms and add real-Go regressions for both cases while preserving package binding, conflict rejection, and sibling-failure blocking.

## Round 5 — 2026-10-21T07:59:25-07:00 (codex) — passed

### Disposed

- BR-5 — addressed — Real-Go TestKnownFlakeExactPackageGate and TestKnownFlakeGateFlatSlashSubtest pass on the pinned head. Restoring the pre-fix collector in a disposable checkout makes both fail with known-flake-rerun-inconclusive; the pure regression also fails. Package binding, required-event conflicts, and sibling-failure controls remain green. The disposable checkout was removed.

## Open findings

(none — every finding has been disposed)
