---
gate: boundary-review
issue: 504
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 80cb73ad-e614-4abb-9c54-b7c3f4b0dab1
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/25dc9c80b0c19429
            reviewed_head: ff57bc497da70f69e3e03043ea38168870e41899
            slot: ariadne:4
            requested_at: "2026-10-20T17:47:02.538692Z"
          phase: finished
          pid: 58982
          started_at: "2026-10-20T17:47:02.677981Z"
          finished_at: "2026-10-20T17:49:09.870174Z"
          session_id: 01a15a6c-821e-7eb3-bd5c-d1a71058d553
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: ff57bc497da70f69e3e03043ea38168870e41899
        tracker_head: 0072d9e4991b99e41ff38f88fe2aa6f4112be614
        branch_point: fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e
        since: "2026-10-20T09:56:17-07:00"
        until: "2026-10-20T10:46:40-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T09:56:17-07:00"
              end: "2026-10-20T10:46:40-07:00"
              began: "2026-10-20T09:56:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 64260ed5b95913c2221bd63750587aa4a12b5365282d022348dc9f43deea7c46
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 3892127
              windows:
                - start: "2026-10-20T09:56:17-07:00"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a3e-8844-7f72-b197-96fc83f81539
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 99f4fe4c8943f116c817c6f49191601750ba7359a2b97219b98d0b55624c89a9
              started: "2026-10-20T16:56:50.162Z"
              spawned: true
              bytes: 429479
              windows:
                - start: "2026-10-20T16:56:50.162Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a42-5379-7c20-88c4-31004da81ecc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4e4c9bc1741e2528e5051d11802a748f5f372ca5332b3434ded06fc5caf6adf7
              started: "2026-10-20T17:00:59.003Z"
              spawned: false
              bytes: 763820
              windows:
                - start: "2026-10-20T17:00:59.003Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a45-27e9-7720-875d-4c8a5b2772e1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1135a9640ff15cd6fbb4ea5f6dbd7032bd62470fa26f54b5b9887c31c685fc66
              started: "2026-10-20T17:04:04.565Z"
              spawned: false
              bytes: 426336
              windows:
                - start: "2026-10-20T17:04:04.565Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a46-07a6-72c0-82cb-44bec02002b5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5bfb3fe7a915a4d656ff9fc353b586004651c3a5db2eca26b8f02cd1dbe0c13f
              started: "2026-10-20T17:05:01.724Z"
              spawned: false
              bytes: 456616
              windows:
                - start: "2026-10-20T17:05:01.724Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a46-8683-7420-a06f-df0dff734fa1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d639dea7a97c7fda61a1498c3bd55ae064ba78bcc121eedf02ec8c40ab988e6e
              started: "2026-10-20T17:05:34.342Z"
              spawned: false
              bytes: 394281
              windows:
                - start: "2026-10-20T17:05:34.342Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a47-64cc-7310-b6d4-0d5c94a3f3cc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 820b4ce19aac4522ef3faa693fcbcb6f4509dc0881c4797ac9437133db6f0cbb
              started: "2026-10-20T17:06:30.831Z"
              spawned: true
              bytes: 611122
              windows:
                - start: "2026-10-20T17:06:30.831Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a47-8d09-72d0-bc5e-59ffcf634459
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef8f65fe4b671657e5c95d5c071a40a3eac4103d889709a3ab12ed99196d3573
              started: "2026-10-20T17:06:41.209Z"
              spawned: true
              bytes: 428852
              windows:
                - start: "2026-10-20T17:06:41.209Z"
                  end: "2026-10-20T10:46:40-07:00"
            - harness: codex
              id: 01a15a69-4c0c-7a52-ab62-17119854390f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ebd12c3a091cc94658e41e7188525eb145a39b575f44d918aef9073b74d621bc
              started: "2026-10-20T17:43:32.788Z"
              spawned: true
              bytes: 464087
              windows:
                - start: "2026-10-20T17:43:32.788Z"
                  end: "2026-10-20T10:46:40-07:00"
      "n": 1
      timestamp: "2026-10-20T10:49:10-07:00"
      dispatched: "2026-10-20T10:47:02-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Accepted YAML parent values can retain the deleted session path
          detail: 'cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: ''../tl-status-op.md'' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.'
          family: relocation-preserves-reference-targets
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Existing negated ignore rules defeat migration's local-file guarantee
          detail: 'cmd/tl-lead/status_move.go:102 treats an ignore pattern''s presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.'
          family: migration-enforces-effective-ignore-policy
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 58bf162b-d81c-4091-b03b-7702d6c559ee
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/421142a4a2e79973
            reviewed_head: cb7c96f5d4f7b4c3c8125cc2593fc6c9884dfbbc
            slot: ariadne:4
            requested_at: "2026-10-20T18:00:57.814703Z"
          phase: finished
          pid: 38908
          started_at: "2026-10-20T18:00:58.0935Z"
          finished_at: "2026-10-20T18:03:25.578912Z"
          session_id: 01a15a79-4176-7811-b5df-8ee0c3622713
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: cb7c96f5d4f7b4c3c8125cc2593fc6c9884dfbbc
        tracker_head: ec282920f2eb4cd54afc98e66c8dd94091ca3df8
        branch_point: fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e
        since: "2026-10-20T09:56:17-07:00"
        until: "2026-10-20T11:00:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T09:56:17-07:00"
              end: "2026-10-20T11:00:16-07:00"
              began: "2026-10-20T09:56:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 7fba531658aa2d030fccfd4ab32b2a8d1a3ee7ef7da5bfc80fe8c36c18c10940
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 4432866
              windows:
                - start: "2026-10-20T09:56:17-07:00"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a3e-8844-7f72-b197-96fc83f81539
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 99f4fe4c8943f116c817c6f49191601750ba7359a2b97219b98d0b55624c89a9
              started: "2026-10-20T16:56:50.162Z"
              spawned: true
              bytes: 429479
              windows:
                - start: "2026-10-20T16:56:50.162Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a42-5379-7c20-88c4-31004da81ecc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4e4c9bc1741e2528e5051d11802a748f5f372ca5332b3434ded06fc5caf6adf7
              started: "2026-10-20T17:00:59.003Z"
              spawned: false
              bytes: 763820
              windows:
                - start: "2026-10-20T17:00:59.003Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a45-27e9-7720-875d-4c8a5b2772e1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1135a9640ff15cd6fbb4ea5f6dbd7032bd62470fa26f54b5b9887c31c685fc66
              started: "2026-10-20T17:04:04.565Z"
              spawned: false
              bytes: 426336
              windows:
                - start: "2026-10-20T17:04:04.565Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a46-07a6-72c0-82cb-44bec02002b5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5bfb3fe7a915a4d656ff9fc353b586004651c3a5db2eca26b8f02cd1dbe0c13f
              started: "2026-10-20T17:05:01.724Z"
              spawned: false
              bytes: 456616
              windows:
                - start: "2026-10-20T17:05:01.724Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a46-8683-7420-a06f-df0dff734fa1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d639dea7a97c7fda61a1498c3bd55ae064ba78bcc121eedf02ec8c40ab988e6e
              started: "2026-10-20T17:05:34.342Z"
              spawned: false
              bytes: 394281
              windows:
                - start: "2026-10-20T17:05:34.342Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a47-64cc-7310-b6d4-0d5c94a3f3cc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 820b4ce19aac4522ef3faa693fcbcb6f4509dc0881c4797ac9437133db6f0cbb
              started: "2026-10-20T17:06:30.831Z"
              spawned: true
              bytes: 611122
              windows:
                - start: "2026-10-20T17:06:30.831Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a47-8d09-72d0-bc5e-59ffcf634459
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef8f65fe4b671657e5c95d5c071a40a3eac4103d889709a3ab12ed99196d3573
              started: "2026-10-20T17:06:41.209Z"
              spawned: true
              bytes: 428852
              windows:
                - start: "2026-10-20T17:06:41.209Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a69-4c0c-7a52-ab62-17119854390f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ebd12c3a091cc94658e41e7188525eb145a39b575f44d918aef9073b74d621bc
              started: "2026-10-20T17:43:32.788Z"
              spawned: true
              bytes: 464087
              windows:
                - start: "2026-10-20T17:43:32.788Z"
                  end: "2026-10-20T11:00:16-07:00"
            - harness: codex
              id: 01a15a75-3658-7670-b877-fa294a5481b8
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 6419148965fba2a62d577dc2fb57ef478282b8e925438264a943b4cbb3e97b80
              started: "2026-10-20T17:56:33.685Z"
              spawned: true
              bytes: 386945
              windows:
                - start: "2026-10-20T17:56:33.685Z"
                  end: "2026-10-20T11:00:16-07:00"
        cuts:
            - dispatched: "2026-10-20T10:47:02-07:00"
              finalized: "2026-10-20T10:49:10-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-20T11:03:26-07:00"
      dispatched: "2026-10-20T11:00:57-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      dispose:
        - id: BR-1
          disposition: addressed
          note: The revised spec explicitly permits plain parents only. TestSessionMoveParentRefusals passes at HEAD; removing the validation guard in a disposable checkout produces behavioral failures for quoted, commented, tagged, block and quoted-key forms.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Existing later negations now cause refusal before copies. TestSessionMoveEffectiveIgnores passes at HEAD and fails for session, side-tree and seen-file cases when the effective-ignore guard is removed. A distinct policy-change-during-copy case remains below.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Important
          title: Copied nested ignore files invalidate the pre-copy protection check
          detail: 'cmd/tl-lead/status_move_apply.go:145 verifies ignores before copying the side tree, which can itself contain .gitignore files. Reproduced with root rules "/session.side/", "!/session.side/", "/session.side/*" and a source side-tree .gitignore containing "!*": apply returns success and removes the sources, but git check-ignore --quiet -- session.side/image.bin exits 1 afterward. ARCH-PURPOSE / ARCH-SECURE: this is the 2nd finding in family migration-enforces-effective-ignore-policy. Earlier rounds fixed instances; state and enforce the rule that every published destination remains ignored under the resulting policy. Enumerate migrated ignore files at every depth and either refuse policy-changing inputs before publication or validate their prospective effects; recheck before switching and cleanup.'
          family: migration-enforces-effective-ignore-policy
          round: 2
        - kind: defect
          id: BR-4
          severity: Important
          title: The transition model records phases without authorizing migration effects
          detail: 'cmd/tl-lead/status_move.go:134 accepts three string events and returns only a phase; it receives no copy or registration observations, failures, or unconfirmed outcomes. cmd/tl-lead/status_move_apply.go:93 performs effects and decides recovery entirely in imperative control flow. ARCH-ORDER: this does not implement the plan''s promised observation-driven model that authorizes effects. Make progress, verification and uncertain outcomes explicit inputs to a pure transition function that returns permitted effects, route production execution through it, and test independent safety invariants across interruption sequences.'
          family: migration-state-model-authorizes-effects
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 021bbc20-f6b3-41fa-b9a9-6e80e029095a
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/b0e27c0bf424e400
            reviewed_head: 67b62bfca2fd2ef447480cf4dc639f39490fca16
            slot: ariadne:4
            requested_at: "2026-10-20T18:07:33.732092Z"
          phase: finished
          pid: 32800
          started_at: "2026-10-20T18:07:33.881768Z"
          finished_at: "2026-10-20T18:10:33.789142Z"
          session_id: 01a15a7f-4b6c-70e3-8849-9095af112878
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 67b62bfca2fd2ef447480cf4dc639f39490fca16
        tracker_head: ec282920f2eb4cd54afc98e66c8dd94091ca3df8
        branch_point: fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e
        since: "2026-10-20T09:56:17-07:00"
        until: "2026-10-20T11:06:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T09:56:17-07:00"
              end: "2026-10-20T11:06:51-07:00"
              began: "2026-10-20T09:56:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 6c7a63d43689a4345c693ddb7fe712e90ab00ab9846576fe731c4601440155ee
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 5224485
              windows:
                - start: "2026-10-20T09:56:17-07:00"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a3e-8844-7f72-b197-96fc83f81539
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 99f4fe4c8943f116c817c6f49191601750ba7359a2b97219b98d0b55624c89a9
              started: "2026-10-20T16:56:50.162Z"
              spawned: true
              bytes: 429479
              windows:
                - start: "2026-10-20T16:56:50.162Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a42-5379-7c20-88c4-31004da81ecc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4e4c9bc1741e2528e5051d11802a748f5f372ca5332b3434ded06fc5caf6adf7
              started: "2026-10-20T17:00:59.003Z"
              spawned: false
              bytes: 763820
              windows:
                - start: "2026-10-20T17:00:59.003Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a45-27e9-7720-875d-4c8a5b2772e1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1135a9640ff15cd6fbb4ea5f6dbd7032bd62470fa26f54b5b9887c31c685fc66
              started: "2026-10-20T17:04:04.565Z"
              spawned: false
              bytes: 426336
              windows:
                - start: "2026-10-20T17:04:04.565Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a46-07a6-72c0-82cb-44bec02002b5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5bfb3fe7a915a4d656ff9fc353b586004651c3a5db2eca26b8f02cd1dbe0c13f
              started: "2026-10-20T17:05:01.724Z"
              spawned: false
              bytes: 456616
              windows:
                - start: "2026-10-20T17:05:01.724Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a46-8683-7420-a06f-df0dff734fa1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d639dea7a97c7fda61a1498c3bd55ae064ba78bcc121eedf02ec8c40ab988e6e
              started: "2026-10-20T17:05:34.342Z"
              spawned: false
              bytes: 394281
              windows:
                - start: "2026-10-20T17:05:34.342Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a47-64cc-7310-b6d4-0d5c94a3f3cc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 820b4ce19aac4522ef3faa693fcbcb6f4509dc0881c4797ac9437133db6f0cbb
              started: "2026-10-20T17:06:30.831Z"
              spawned: true
              bytes: 611122
              windows:
                - start: "2026-10-20T17:06:30.831Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a47-8d09-72d0-bc5e-59ffcf634459
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef8f65fe4b671657e5c95d5c071a40a3eac4103d889709a3ab12ed99196d3573
              started: "2026-10-20T17:06:41.209Z"
              spawned: true
              bytes: 428852
              windows:
                - start: "2026-10-20T17:06:41.209Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a69-4c0c-7a52-ab62-17119854390f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ebd12c3a091cc94658e41e7188525eb145a39b575f44d918aef9073b74d621bc
              started: "2026-10-20T17:43:32.788Z"
              spawned: true
              bytes: 464087
              windows:
                - start: "2026-10-20T17:43:32.788Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a75-3658-7670-b877-fa294a5481b8
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 6419148965fba2a62d577dc2fb57ef478282b8e925438264a943b4cbb3e97b80
              started: "2026-10-20T17:56:33.685Z"
              spawned: true
              bytes: 386945
              windows:
                - start: "2026-10-20T17:56:33.685Z"
                  end: "2026-10-20T11:06:51-07:00"
            - harness: codex
              id: 01a15a7c-3b63-7521-a427-e5ef10e7f5ae
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: b5135e0c34b4f5b53779cdc33dd9243449e8392172086c2698e4f6975ab49130
              started: "2026-10-20T18:04:13.61Z"
              spawned: true
              bytes: 440990
              windows:
                - start: "2026-10-20T18:04:13.61Z"
                  end: "2026-10-20T11:06:51-07:00"
        cuts:
            - dispatched: "2026-10-20T10:47:02-07:00"
              finalized: "2026-10-20T10:49:10-07:00"
              finalizing: false
            - dispatched: "2026-10-20T11:00:57-07:00"
              finalized: "2026-10-20T11:03:26-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-20T11:10:34-07:00"
      dispatched: "2026-10-20T11:07:33-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      dispose:
        - id: BR-3
          disposition: addressed
          note: All-depth source/destination policy guards and fresh pre-switch/per-delete checks are reachable. Removing rejectMoveIgnoreFiles in a disposable checkout caused six pure cases and two production refusal cases to fail; restored checks pass.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Production dispatch now consumes transitionMove effect grants and reports confirmed, failed or unconfirmed outcomes. Disabling failure/uncertainty revocation makes TestSessionMoveTransition fail; restored sequence and production interruption tests pass.
          round: 3
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Migration deletes the target of accepted side-tree directory links
          detail: cmd/tl-lead/status_move.go:55 maps descendants through MoveUnder but omits the side-tree root itself. Adding [browse](tl-status-op.side/) to the production fixture makes apply succeed with [browse](teams/dev/tl-status-op.side/) in session.md, then cleanup deletes that target. This is the 2nd finding in family relocation-preserves-reference-targets. Earlier rounds fixed instances; state and enforce the rule that every relocated reference preserves its target. Enumerate session, side-root, descendant-directory, file and stationary targets across Markdown/reference/image links, parent and discussions consumers. Map the side-tree root explicitly and add regression coverage for the enumeration.
          family: relocation-preserves-reference-targets
          round: 3
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 0bb714f9-4e93-448f-bcb7-520a2af14b34
            repository: github.com/xianxu/ariadne
            issue_id: "000504"
            recipe: milestone-review
            round: 4
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/f56e119ac6dc8fec
            reviewed_head: 339ca08719fb439537a062638e467c4921952c50
            slot: ariadne:4
            requested_at: "2026-10-20T18:15:01.361925Z"
          phase: finished
          pid: 66714
          started_at: "2026-10-20T18:15:01.500746Z"
          finished_at: "2026-10-20T18:17:05.995207Z"
          session_id: 01a15a86-1fd8-7302-af13-a00dcfc9dd8e
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 339ca08719fb439537a062638e467c4921952c50
        tracker_head: f927c17794dcb484b4bcd0870349c03dda108a78
        branch_point: fdfc3ca3c297c1c6c2802e03f3ea54b73f4a2b3e
        since: "2026-10-20T09:56:17-07:00"
        until: "2026-10-20T11:14:31-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-20T09:56:17-07:00"
              end: "2026-10-20T11:14:31-07:00"
              began: "2026-10-20T09:56:17-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
        sources:
            - harness: codex
              id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 80b7591b2f7a031b861964bf10d38f25d3719d60dcd5296b1be4c7ce3df7c3a1
              started: "2026-10-20T16:55:54.591Z"
              spawned: false
              bytes: 5678192
              windows:
                - start: "2026-10-20T09:56:17-07:00"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a3e-8844-7f72-b197-96fc83f81539
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 99f4fe4c8943f116c817c6f49191601750ba7359a2b97219b98d0b55624c89a9
              started: "2026-10-20T16:56:50.162Z"
              spawned: true
              bytes: 429479
              windows:
                - start: "2026-10-20T16:56:50.162Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a42-5379-7c20-88c4-31004da81ecc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 4e4c9bc1741e2528e5051d11802a748f5f372ca5332b3434ded06fc5caf6adf7
              started: "2026-10-20T17:00:59.003Z"
              spawned: false
              bytes: 763820
              windows:
                - start: "2026-10-20T17:00:59.003Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a45-27e9-7720-875d-4c8a5b2772e1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 1135a9640ff15cd6fbb4ea5f6dbd7032bd62470fa26f54b5b9887c31c685fc66
              started: "2026-10-20T17:04:04.565Z"
              spawned: false
              bytes: 426336
              windows:
                - start: "2026-10-20T17:04:04.565Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a46-07a6-72c0-82cb-44bec02002b5
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: 5bfb3fe7a915a4d656ff9fc353b586004651c3a5db2eca26b8f02cd1dbe0c13f
              started: "2026-10-20T17:05:01.724Z"
              spawned: false
              bytes: 456616
              windows:
                - start: "2026-10-20T17:05:01.724Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a46-8683-7420-a06f-df0dff734fa1
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              digest: d639dea7a97c7fda61a1498c3bd55ae064ba78bcc121eedf02ec8c40ab988e6e
              started: "2026-10-20T17:05:34.342Z"
              spawned: false
              bytes: 394281
              windows:
                - start: "2026-10-20T17:05:34.342Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a47-64cc-7310-b6d4-0d5c94a3f3cc
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 820b4ce19aac4522ef3faa693fcbcb6f4509dc0881c4797ac9437133db6f0cbb
              started: "2026-10-20T17:06:30.831Z"
              spawned: true
              bytes: 611122
              windows:
                - start: "2026-10-20T17:06:30.831Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a47-8d09-72d0-bc5e-59ffcf634459
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ef8f65fe4b671657e5c95d5c071a40a3eac4103d889709a3ab12ed99196d3573
              started: "2026-10-20T17:06:41.209Z"
              spawned: true
              bytes: 428852
              windows:
                - start: "2026-10-20T17:06:41.209Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a69-4c0c-7a52-ab62-17119854390f
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: ebd12c3a091cc94658e41e7188525eb145a39b575f44d918aef9073b74d621bc
              started: "2026-10-20T17:43:32.788Z"
              spawned: true
              bytes: 464087
              windows:
                - start: "2026-10-20T17:43:32.788Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a75-3658-7670-b877-fa294a5481b8
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: 6419148965fba2a62d577dc2fb57ef478282b8e925438264a943b4cbb3e97b80
              started: "2026-10-20T17:56:33.685Z"
              spawned: true
              bytes: 386945
              windows:
                - start: "2026-10-20T17:56:33.685Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a7c-3b63-7521-a427-e5ef10e7f5ae
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: b5135e0c34b4f5b53779cdc33dd9243449e8392172086c2698e4f6975ab49130
              started: "2026-10-20T18:04:13.61Z"
              spawned: true
              bytes: 440990
              windows:
                - start: "2026-10-20T18:04:13.61Z"
                  end: "2026-10-20T11:14:31-07:00"
            - harness: codex
              id: 01a15a82-bd42-7803-8860-2975e90d17aa
              slot: ariadne:4
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
              parent_id: 01a15a3d-80b5-7451-ab0e-35331fe11c6f
              digest: d5ac9254be09b888a1001f0ad1760d90cb7f65d73775acd38dab54be6753a3ae
              started: "2026-10-20T18:11:20.074Z"
              spawned: true
              bytes: 432542
              windows:
                - start: "2026-10-20T18:11:20.074Z"
                  end: "2026-10-20T11:14:31-07:00"
        cuts:
            - dispatched: "2026-10-20T10:47:02-07:00"
              finalized: "2026-10-20T10:49:10-07:00"
              finalizing: false
            - dispatched: "2026-10-20T11:00:57-07:00"
              finalized: "2026-10-20T11:03:26-07:00"
              finalizing: false
            - dispatched: "2026-10-20T11:07:33-07:00"
              finalized: "2026-10-20T11:10:34-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-20T11:17:06-07:00"
      dispatched: "2026-10-20T11:15:01-07:00"
      agent: codex
      reviewer: 'codex (openai; author openai); degraded: same-family by choice'
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plain-parent validation remains reachable before effects; parent-refusal regressions pass.
          round: 4
        - id: BR-2
          disposition: addressed
          note: Effective Git ignore checks remain enforced; negation regressions pass.
          round: 4
        - id: BR-3
          disposition: addressed
          note: All-depth policy-file refusals and refreshed publication checks remain present; nested and late-policy regressions pass.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Production dispatch consumes transitionMove effect grants; transition, interruption, and uncertain-rename tests pass.
          round: 4
        - id: BR-5
          disposition: addressed
          note: status_move.go:57 explicitly maps the side root; directory planning preserves empty targets. The production reference matrix passes. Temporary overlays removing root mapping or directory planning independently make the corresponding regressions fail.
          round: 4
      recipe: milestone-review
      isolation: fenced
      reviewed: 339ca08719fb439537a062638e467c4921952c50
      blocked: false
---

# Gate ledger — ariadne#504 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-20T10:49:10-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `relocation-preserves-reference-targets` (kind: defect) Accepted YAML parent values can retain the deleted session path
  cmd/tl-lead/status_move.go:70 delegates parent rewriting to sidedoc.RebaseLinks, whose rebaseParent only handles plain values and uncommented double-quoted values. A valid parent: '../tl-status-op.md' or parent: ../tl-status-op.md # backlink passes snapshot metadata validation but remains unchanged during relocation; apply then deletes the original session and reports success. ARCH-PURPOSE / ARCH-SECURE: decode and rewrite supported parent scalars structurally, preserving surrounding metadata, or refuse unsupported forms before effects. Add migration regressions covering quotes, tags, and inline comments.
- **BR-2** [Important] `migration-enforces-effective-ignore-policy` (kind: defect) Existing negated ignore rules defeat migration's local-file guarantee
  cmd/tl-lead/status_move.go:102 treats an ignore pattern's presence anywhere as sufficient. With /session.md followed by !/session.md, it adds no later corrective rule, and runSessionMove copies and switches without checking effective ignored status. The same applies to the other migration-owned patterns. ARCH-PURPOSE: ensure the final rules take precedence and verify effective ignores before publishing copies; add a regression with later negations and assert git check-ignore results.

## Round 2 — 2026-10-20T11:03:26-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The revised spec explicitly permits plain parents only. TestSessionMoveParentRefusals passes at HEAD; removing the validation guard in a disposable checkout produces behavioral failures for quoted, commented, tagged, block and quoted-key forms.
- BR-2 — addressed — Existing later negations now cause refusal before copies. TestSessionMoveEffectiveIgnores passes at HEAD and fails for session, side-tree and seen-file cases when the effective-ignore guard is removed. A distinct policy-change-during-copy case remains below.

### Raised

- **BR-3** [Important] `migration-enforces-effective-ignore-policy` (kind: defect) Copied nested ignore files invalidate the pre-copy protection check
  cmd/tl-lead/status_move_apply.go:145 verifies ignores before copying the side tree, which can itself contain .gitignore files. Reproduced with root rules "/session.side/", "!/session.side/", "/session.side/*" and a source side-tree .gitignore containing "!*": apply returns success and removes the sources, but git check-ignore --quiet -- session.side/image.bin exits 1 afterward. ARCH-PURPOSE / ARCH-SECURE: this is the 2nd finding in family migration-enforces-effective-ignore-policy. Earlier rounds fixed instances; state and enforce the rule that every published destination remains ignored under the resulting policy. Enumerate migrated ignore files at every depth and either refuse policy-changing inputs before publication or validate their prospective effects; recheck before switching and cleanup.
- **BR-4** [Important] `migration-state-model-authorizes-effects` (kind: defect) The transition model records phases without authorizing migration effects
  cmd/tl-lead/status_move.go:134 accepts three string events and returns only a phase; it receives no copy or registration observations, failures, or unconfirmed outcomes. cmd/tl-lead/status_move_apply.go:93 performs effects and decides recovery entirely in imperative control flow. ARCH-ORDER: this does not implement the plan's promised observation-driven model that authorizes effects. Make progress, verification and uncertain outcomes explicit inputs to a pure transition function that returns permitted effects, route production execution through it, and test independent safety invariants across interruption sequences.

## Round 3 — 2026-10-20T11:10:34-07:00 (codex) — BLOCKED

### Disposed

- BR-3 — addressed — All-depth source/destination policy guards and fresh pre-switch/per-delete checks are reachable. Removing rejectMoveIgnoreFiles in a disposable checkout caused six pure cases and two production refusal cases to fail; restored checks pass.
- BR-4 — addressed — Production dispatch now consumes transitionMove effect grants and reports confirmed, failed or unconfirmed outcomes. Disabling failure/uncertainty revocation makes TestSessionMoveTransition fail; restored sequence and production interruption tests pass.

### Raised

- **BR-5** [Critical] `relocation-preserves-reference-targets` (kind: defect) Migration deletes the target of accepted side-tree directory links
  cmd/tl-lead/status_move.go:55 maps descendants through MoveUnder but omits the side-tree root itself. Adding [browse](tl-status-op.side/) to the production fixture makes apply succeed with [browse](teams/dev/tl-status-op.side/) in session.md, then cleanup deletes that target. This is the 2nd finding in family relocation-preserves-reference-targets. Earlier rounds fixed instances; state and enforce the rule that every relocated reference preserves its target. Enumerate session, side-root, descendant-directory, file and stationary targets across Markdown/reference/image links, parent and discussions consumers. Map the side-tree root explicitly and add regression coverage for the enumeration.

## Round 4 — 2026-10-20T11:17:06-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Plain-parent validation remains reachable before effects; parent-refusal regressions pass.
- BR-2 — addressed — Effective Git ignore checks remain enforced; negation regressions pass.
- BR-3 — addressed — All-depth policy-file refusals and refreshed publication checks remain present; nested and late-policy regressions pass.
- BR-4 — addressed — Production dispatch consumes transitionMove effect grants; transition, interruption, and uncertain-rename tests pass.
- BR-5 — addressed — status_move.go:57 explicitly maps the side root; directory planning preserves empty targets. The production reference matrix passes. Temporary overlays removing root mapping or directory planning independently make the corresponding regressions fail.

## Open findings

(none — every finding has been disposed)
