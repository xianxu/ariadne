---
gate: boundary-review
issue: 441
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 1a9c75c4-0956-4650-90ef-74613071a845
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/dc36896d98479771
            reviewed_head: 0fcc618f8f6848441a40a9dd5bfec00db29a1cb8
            slot: ariadne:2
            requested_at: "2026-10-18T16:23:57.619564Z"
          phase: finished
          pid: 55386
          started_at: "2026-10-18T16:23:57.809457Z"
          finished_at: "2026-10-18T16:44:17.539951Z"
          session_id: f5a73cbb-61b9-4e92-927d-9ba28c3e4a94
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 0fcc618f8f6848441a40a9dd5bfec00db29a1cb8
        tracker_head: e655ff87803ccaad9f0054eadd97f53eb740d0f9
        branch_point: 4c2fa13e834dfdf4caec56b2c47b864b5465cefe
        since: "2026-10-18T09:10:46-07:00"
        until: "2026-10-18T09:23:32-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T09:10:46-07:00"
              end: "2026-10-18T09:23:32-07:00"
              began: "2026-10-18T09:10:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 4945cf1c-b2de-4b88-8487-960125262396
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bfd2a2a02db84fc7901f95f82e724d16044f613e3e25965fc5af0641a010abe9
              started: "2026-10-18T16:13:07.152Z"
              spawned: false
              bytes: 540455
              windows:
                - start: "2026-10-18T16:13:07.152Z"
                  end: "2026-10-18T09:23:32-07:00"
            - harness: claude
              id: 96feb223-0a1d-41a2-8c6b-18fb25b1e8b2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ade489c7ff729c62ba31b994f6802871c1ef39e98009ca682757e66f4736e55d
              started: "2026-10-18T16:16:05.648Z"
              spawned: false
              bytes: 383034
              windows:
                - start: "2026-10-18T16:16:05.648Z"
                  end: "2026-10-18T09:23:32-07:00"
            - harness: claude
              id: 99ead2dc-ffdc-4e2f-8618-76195520d5d3
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c0666deb100115d75698688ba06499b1733b8839bcf0db3bfa61cc5854749768
              started: "2026-10-18T16:14:46.747Z"
              spawned: false
              bytes: 499870
              windows:
                - start: "2026-10-18T16:14:46.747Z"
                  end: "2026-10-18T09:23:32-07:00"
            - harness: claude
              id: d0a88980-85ac-4e44-a6cb-cfc8b055a1be
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cd9d07196f3ec139376dc4c68d4eeae203fbf2cd7ab5520f6c3e4ba6dcc74d58
              started: "2026-10-18T16:15:41.447Z"
              spawned: false
              bytes: 476863
              windows:
                - start: "2026-10-18T16:15:41.447Z"
                  end: "2026-10-18T09:23:32-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c1aed187b2794ca40572e2a8aaa1fc6724ba67177f067f12cb21eee8448245e3
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 186806380
              windows:
                - start: "2026-10-18T09:10:46-07:00"
                  end: "2026-10-18T09:23:32-07:00"
      "n": 1
      timestamp: "2026-10-18T09:44:18-07:00"
      dispatched: "2026-10-18T09:23:57-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: judge's Choice.Refusal() no longer unwraps to ErrReviewDidNotRun; the test pinning it was rewritten
          detail: 'pkg/reviewpolicy/choice.go:76 NoReviewerError.Unwrap returns ErrNoReviewer only, so the Refusal() errors returned by changecode.go:732,988, milestoneclose.go:903 and judge.go:166 stop matching judge.ErrReviewDidNotRun (only DispatchChoice''s noReviewer wrapper still does). The moved TestResolve (policy_test.go:164) swapped its assertion to ErrNoReviewer, against the plan''s "tests pass unchanged"; the hint text also changed from "name one with --agent" to "name one explicitly". Fix: in judge, return noReviewer{} from the refusal path and restore a judge test asserting both errors.Is(ErrReviewDidNotRun) and errors.As(*NoReviewerError).'
          family: move-preserves-contract
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Plan Core-concepts rows name files and signatures the code doesn't have
          detail: Record is in choice.go (not record.go) with signature (used, choice, skips); AuthorFamilies takes []agentenv.Commit, not []string; judge/policyalias.go doesn't exist (the aliases are in family.go, agent_resolver.go, reviewchoice.go, dispatch.go). Add a Revisions entry correcting the table.
          family: plan-table-matches-code
          round: 1
        - kind: defect
          id: BR-3
          severity: Important
          title: judge family_test.go and reviewchoice_test.go fail gofmt -l
          detail: Runs of blank lines are left where the moved tests were cut out; run gofmt -w on both files.
          family: gofmt-clean
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: policy_test.go helper comments still name ParseReviewerRecord / ReviewerRecord
          detail: Rename the comments to ParseRecord / Record.
          family: stale-identifier-in-comment
          round: 1
        - kind: defect
          id: BR-5
          severity: Minor
          title: pkg/agentenv has no tests, and reviewpolicy Load/AuthorFamilies are only tested from cmd/sdlc
          detail: doc-review (M2) will import these directly; move or add package-level tests alongside them.
          family: shared-package-owns-its-tests
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 94f721a6-48b0-4fb6-bbca-3c26bd7df1cc
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/fad372f250fe19fa
            reviewed_head: 88fddabe648729d8270c338b9a9606c03983cece
            slot: ariadne:2
            requested_at: "2026-10-18T16:46:45.95791Z"
          phase: finished
          pid: 55727
          started_at: "2026-10-18T16:46:46.180434Z"
          finished_at: "2026-10-18T17:09:28.48915Z"
          session_id: 50a915aa-a5f0-40a3-a13f-11997634b2aa
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 88fddabe648729d8270c338b9a9606c03983cece
        tracker_head: b529ce35026932a9477ef23c38f1372e147336a7
        branch_point: 4c2fa13e834dfdf4caec56b2c47b864b5465cefe
        since: "2026-10-18T09:10:46-07:00"
        until: "2026-10-18T09:46:17-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T09:10:46-07:00"
              end: "2026-10-18T09:46:17-07:00"
              began: "2026-10-18T09:10:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 4945cf1c-b2de-4b88-8487-960125262396
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bfd2a2a02db84fc7901f95f82e724d16044f613e3e25965fc5af0641a010abe9
              started: "2026-10-18T16:13:07.152Z"
              spawned: false
              bytes: 540455
              windows:
                - start: "2026-10-18T16:13:07.152Z"
                  end: "2026-10-18T09:46:17-07:00"
            - harness: claude
              id: 96feb223-0a1d-41a2-8c6b-18fb25b1e8b2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ade489c7ff729c62ba31b994f6802871c1ef39e98009ca682757e66f4736e55d
              started: "2026-10-18T16:16:05.648Z"
              spawned: false
              bytes: 383034
              windows:
                - start: "2026-10-18T16:16:05.648Z"
                  end: "2026-10-18T09:46:17-07:00"
            - harness: claude
              id: 99ead2dc-ffdc-4e2f-8618-76195520d5d3
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c0666deb100115d75698688ba06499b1733b8839bcf0db3bfa61cc5854749768
              started: "2026-10-18T16:14:46.747Z"
              spawned: false
              bytes: 499870
              windows:
                - start: "2026-10-18T16:14:46.747Z"
                  end: "2026-10-18T09:46:17-07:00"
            - harness: claude
              id: d0a88980-85ac-4e44-a6cb-cfc8b055a1be
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cd9d07196f3ec139376dc4c68d4eeae203fbf2cd7ab5520f6c3e4ba6dcc74d58
              started: "2026-10-18T16:15:41.447Z"
              spawned: false
              bytes: 476863
              windows:
                - start: "2026-10-18T16:15:41.447Z"
                  end: "2026-10-18T09:46:17-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 91f4b976e4b3e60548879b86867330a735c8c3f3002aaa271166b9b33547a1de
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 189188450
              windows:
                - start: "2026-10-18T09:10:46-07:00"
                  end: "2026-10-18T09:46:17-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T09:23:57-07:00"
              finalized: "2026-10-18T09:44:18-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-18T10:09:30-07:00"
      dispatched: "2026-10-18T09:46:45-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: judge.Refusal (reviewchoice.go:67) matches both NoReviewerError and ErrReviewDidNotRun and keeps the --agent hint; all 4 sites use it; TestRefusalKeepsJudgesContract pins it.
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'Plan Revisions (M1 round 1) table checked against the head: Record in choice.go, AuthorFamilies takes []agentenv.Commit, aliases listed by their real files.'
          round: 2
        - id: BR-3
          disposition: addressed
          note: gofmt -l on cmd and pkg lists no file in the diff; the extra blank lines are gone in family_test.go and reviewchoice_test.go.
          round: 2
        - id: BR-4
          disposition: addressed
          note: policy_test.go:209-210 comments now say ParseRecord / Record; no ReviewerRecord left in pkg/.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Added pkg/agentenv/agentenv_test.go (Env, ParseLog, LogFormat) and pkg/reviewpolicy/load_test.go (Load, AuthorFamilies); both pass.
          round: 2
      findings:
        - kind: defect
          id: BR-6
          severity: Minor
          title: reviewpolicy Choice.Refusal() stays callable through the judge.ReviewChoice alias and drops judge's refusal contract
          detail: '2nd finding in this family. Rule: a method a moved type brings along through an alias must keep its old callers'' contract, or sdlc must be unable to call it. Fix the rule, not the call site: remove Choice.Refusal (choice.go:55), or add a judge test that fails when non-test code outside pkg/ calls .Refusal(). Today no site test would fail if a site went back to choice.Refusal(): TestBoundaryReviewRefusesWithNoOtherFamily only checks the text "no reviewer available".'
          family: move-preserves-contract
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: fenced
      reviewed: 88fddabe648729d8270c338b9a9606c03983cece
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 9ed6e410-73b7-4476-9541-8520fc9569ac
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/5b0857c0d46fe800
            reviewed_head: a57f01df3f2eb012cc6a88a2e394da92b9383909
            slot: ariadne:2
            requested_at: "2026-10-18T17:14:35.019626Z"
          phase: finished
          pid: 55612
          started_at: "2026-10-18T17:14:35.193984Z"
          finished_at: "2026-10-18T17:17:06.542053Z"
          session_id: 9effd4c5-06be-4db4-9c56-c178602c9f7b
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: a57f01df3f2eb012cc6a88a2e394da92b9383909
        tracker_head: 45358b7114d2340036d1d059390e3203f97f259d
        branch_point: 4c2fa13e834dfdf4caec56b2c47b864b5465cefe
        since: "2026-10-18T09:10:46-07:00"
        until: "2026-10-18T10:14:08-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T09:10:46-07:00"
              end: "2026-10-18T10:14:08-07:00"
              began: "2026-10-18T09:10:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 4945cf1c-b2de-4b88-8487-960125262396
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bfd2a2a02db84fc7901f95f82e724d16044f613e3e25965fc5af0641a010abe9
              started: "2026-10-18T16:13:07.152Z"
              spawned: false
              bytes: 540455
              windows:
                - start: "2026-10-18T16:13:07.152Z"
                  end: "2026-10-18T10:14:08-07:00"
            - harness: claude
              id: 96feb223-0a1d-41a2-8c6b-18fb25b1e8b2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ade489c7ff729c62ba31b994f6802871c1ef39e98009ca682757e66f4736e55d
              started: "2026-10-18T16:16:05.648Z"
              spawned: false
              bytes: 383034
              windows:
                - start: "2026-10-18T16:16:05.648Z"
                  end: "2026-10-18T10:14:08-07:00"
            - harness: claude
              id: 99ead2dc-ffdc-4e2f-8618-76195520d5d3
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c0666deb100115d75698688ba06499b1733b8839bcf0db3bfa61cc5854749768
              started: "2026-10-18T16:14:46.747Z"
              spawned: false
              bytes: 499870
              windows:
                - start: "2026-10-18T16:14:46.747Z"
                  end: "2026-10-18T10:14:08-07:00"
            - harness: claude
              id: d0a88980-85ac-4e44-a6cb-cfc8b055a1be
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cd9d07196f3ec139376dc4c68d4eeae203fbf2cd7ab5520f6c3e4ba6dcc74d58
              started: "2026-10-18T16:15:41.447Z"
              spawned: false
              bytes: 476863
              windows:
                - start: "2026-10-18T16:15:41.447Z"
                  end: "2026-10-18T10:14:08-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 4aacbec96d4c214560d2c6e98b27ffb94c2d5f0372f5ffd9fe99c4b18780c2e9
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 193115978
              windows:
                - start: "2026-10-18T09:10:46-07:00"
                  end: "2026-10-18T10:14:08-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T09:23:57-07:00"
              finalized: "2026-10-18T09:44:18-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T09:46:45-07:00"
              finalized: "2026-10-18T10:09:30-07:00"
              finalizing: true
      "n": 3
      timestamp: "2026-10-18T10:17:07-07:00"
      dispatched: "2026-10-18T10:14:35-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-7
          severity: Critical
          title: doc-review self-reviews under explicit mode when the session agent has no pair, and the record does not mark it degraded
          detail: 'chooseReviewers (cmd/doc-review/choose.go:70-76) passes explicit policies straight to Resolve. Plan then asks Policy.Reviewer(main), which returns main when the session has no pair (pkg/reviewpolicy/policy.go:120-124). Probe: Pairs{codex:claude}, a claude session and a claude-authored doc give candidates=[claude] and record "claude (anthropic; author anthropic)" with no degraded marker. A pair whose reviewer shares the authors'' family also slips through. This contradicts the help ("never reviews a document with its author''s own family") and the M2 commit. Fix: in chooseReviewers, for a non-forced choice, move candidates whose family is among the authors into Skips (refusing if none remain). Add the case to TestChooseReviewers and remove the Explicit exemption at choose_test.go:140.'
          family: never-self-review-holds-in-every-mode
          round: 3
        - kind: defect
          id: BR-8
          severity: Important
          title: The plan's as-built table still lists Choice.Refusal, which this window deleted from pkg/reviewpolicy/choice.go
          detail: 'This is the 2nd finding in family plan-table-matches-code. Rule: any commit that adds, renames or removes an identifier listed in the as-built table updates the table in the same commit, checked by grepping every table identifier against the head at boundary close. Here: remove Refusal from plan.md:90''s Choice row and add a Revisions note that the shared package''s Choice.Refusal API was dropped (refusals are now &NoReviewerError{Skips}).'
          family: plan-table-matches-code
          round: 3
        - kind: defect
          id: BR-9
          severity: Minor
          title: cmd/doc-review/main.go:2-3 and :15 still say the reviewer defaults to codex
          detail: 'This is the 2nd finding in family stale-identifier-in-comment. Rule: when a behaviour or identifier changes, grep the package''s comments for the old claim in the same commit (rg ''codex by default|defaults to codex|DefaultAgent'' cmd/doc-review).'
          family: stale-identifier-in-comment
          round: 3
        - kind: defect
          id: BR-10
          severity: Minor
          title: When an explicitly named agent fails to launch, the refusal tells the user to name one explicitly
          detail: A forced candidate that fails to launch ends in NoReviewerError, whose text says "name one explicitly", though the user already did. Return the launch error directly when choice.Forced.
          family: refusal-names-applicable-remedy
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cbe342c3-4ee6-4d58-90b3-3d4db48692a4
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/2d82673fb6471c63
            reviewed_head: 208ed53a35799b810557fc46986d7ebbc659495f
            slot: ariadne:2
            requested_at: "2026-10-18T17:18:39.156824Z"
          phase: finished
          pid: 98105
          started_at: "2026-10-18T17:18:39.332472Z"
          finished_at: "2026-10-18T17:20:20.029283Z"
          session_id: fdb320cd-8668-48bb-aee5-6dc42f92fdb8
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 208ed53a35799b810557fc46986d7ebbc659495f
        tracker_head: 5381acd2b80ee58e81dcb189bf9e63fa564f8880
        branch_point: 4c2fa13e834dfdf4caec56b2c47b864b5465cefe
        since: "2026-10-18T09:10:46-07:00"
        until: "2026-10-18T10:18:10-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T09:10:46-07:00"
              end: "2026-10-18T10:18:10-07:00"
              began: "2026-10-18T09:10:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 4945cf1c-b2de-4b88-8487-960125262396
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bfd2a2a02db84fc7901f95f82e724d16044f613e3e25965fc5af0641a010abe9
              started: "2026-10-18T16:13:07.152Z"
              spawned: false
              bytes: 540455
              windows:
                - start: "2026-10-18T16:13:07.152Z"
                  end: "2026-10-18T10:18:10-07:00"
            - harness: claude
              id: 96feb223-0a1d-41a2-8c6b-18fb25b1e8b2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ade489c7ff729c62ba31b994f6802871c1ef39e98009ca682757e66f4736e55d
              started: "2026-10-18T16:16:05.648Z"
              spawned: false
              bytes: 383034
              windows:
                - start: "2026-10-18T16:16:05.648Z"
                  end: "2026-10-18T10:18:10-07:00"
            - harness: claude
              id: 99ead2dc-ffdc-4e2f-8618-76195520d5d3
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c0666deb100115d75698688ba06499b1733b8839bcf0db3bfa61cc5854749768
              started: "2026-10-18T16:14:46.747Z"
              spawned: false
              bytes: 499870
              windows:
                - start: "2026-10-18T16:14:46.747Z"
                  end: "2026-10-18T10:18:10-07:00"
            - harness: claude
              id: d0a88980-85ac-4e44-a6cb-cfc8b055a1be
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cd9d07196f3ec139376dc4c68d4eeae203fbf2cd7ab5520f6c3e4ba6dcc74d58
              started: "2026-10-18T16:15:41.447Z"
              spawned: false
              bytes: 476863
              windows:
                - start: "2026-10-18T16:15:41.447Z"
                  end: "2026-10-18T10:18:10-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 6cd9aca0446c731069c5a962f30648dc2bd85df5dda6b60ff302aa1cd9e4dd15
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 194517799
              windows:
                - start: "2026-10-18T09:10:46-07:00"
                  end: "2026-10-18T10:18:10-07:00"
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T09:23:57-07:00"
              finalized: "2026-10-18T09:44:18-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T09:46:45-07:00"
              finalized: "2026-10-18T10:09:30-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-18T10:14:35-07:00"
              finalized: "2026-10-18T10:17:07-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-18T10:20:20-07:00"
      dispatched: "2026-10-18T10:18:39-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-7
          disposition: addressed
          note: choose.go:60-74 skips non-forced author-family candidates; reverting it turns TestChooseReviewers (2 explicit cases) and TestChooseReviewersRecordsTheSelfReviewSkip red.
          round: 4
        - id: BR-8
          disposition: addressed
          note: plan.md:90 Choice row now lists only First; all as-built identifiers grepped present at head; Revisions entry records the Refusal removal and the same-commit rule.
          round: 4
        - id: BR-9
          disposition: addressed
          note: main.go:2-4,16 rewritten; rg 'codex by default|defaults to codex|DefaultAgent' over cmd/doc-review, fix SKILL.md and docflow atlas returns nothing.
          round: 4
        - id: BR-10
          disposition: addressed
          note: review.go:99-101 returns launchErr when Forced; removing it turns TestRunReviewForcedLaunchFailure red with the old name-one-explicitly text.
          round: 4
      findings:
        - kind: defect
          id: BR-11
          severity: Minor
          title: A refusal caused by the never-self-review skip still advises retry later, though only a config change or a named agent can help
          detail: 'This is the 2nd finding in family refusal-names-applicable-remedy. Rule: the remedy a NoReviewerError names comes from its skip reasons: transient skips (not on PATH, failed launch, usage limit) suggest retry later; policy skips (shares the author family) suggest fixing reviewer_pairs or naming an agent; forced failures return their own error. help.md:26-27 should also say that an explicit pair into the author''s family, or an unpaired session, refuses.'
          family: refusal-names-applicable-remedy
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: fenced
      reviewed: 208ed53a35799b810557fc46986d7ebbc659495f
      blocked: false
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 24e3e023-f0ff-4989-a8e5-cd79a32cdb38
            repository: github.com/xianxu/ariadne
            issue_id: "000441"
            recipe: milestone-review
            round: 5
            attempt: 1
            agent: claude
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/beea8d778a44af75
            reviewed_head: b7c033ab62cfa957c71465574ff4aaa93a74468b
            slot: ariadne:2
            requested_at: "2026-10-18T18:17:29.807684Z"
          phase: finished
          pid: 10310
          started_at: "2026-10-18T18:17:29.92047Z"
          finished_at: "2026-10-18T18:18:53.05861Z"
          session_id: 03d0af34-0961-492b-b5f0-da09d5190f05
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: b7c033ab62cfa957c71465574ff4aaa93a74468b
        tracker_head: c5bc654a3c3906d0d5a8589cd8414418906c27e1
        branch_point: 4c2fa13e834dfdf4caec56b2c47b864b5465cefe
        since: "2026-10-18T09:10:46-07:00"
        until: "2026-10-18T10:48:02-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-18T09:10:46-07:00"
              end: "2026-10-18T10:48:02-07:00"
              began: "2026-10-18T09:10:46-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: 4945cf1c-b2de-4b88-8487-960125262396
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: bfd2a2a02db84fc7901f95f82e724d16044f613e3e25965fc5af0641a010abe9
              started: "2026-10-18T16:13:07.152Z"
              spawned: false
              bytes: 540455
              windows:
                - start: "2026-10-18T16:13:07.152Z"
                  end: "2026-10-18T10:48:02-07:00"
            - harness: claude
              id: 96feb223-0a1d-41a2-8c6b-18fb25b1e8b2
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: ade489c7ff729c62ba31b994f6802871c1ef39e98009ca682757e66f4736e55d
              started: "2026-10-18T16:16:05.648Z"
              spawned: false
              bytes: 383034
              windows:
                - start: "2026-10-18T16:16:05.648Z"
                  end: "2026-10-18T10:48:02-07:00"
            - harness: claude
              id: 99ead2dc-ffdc-4e2f-8618-76195520d5d3
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: c0666deb100115d75698688ba06499b1733b8839bcf0db3bfa61cc5854749768
              started: "2026-10-18T16:14:46.747Z"
              spawned: false
              bytes: 499870
              windows:
                - start: "2026-10-18T16:14:46.747Z"
                  end: "2026-10-18T10:48:02-07:00"
            - harness: claude
              id: d0a88980-85ac-4e44-a6cb-cfc8b055a1be
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: cd9d07196f3ec139376dc4c68d4eeae203fbf2cd7ab5520f6c3e4ba6dcc74d58
              started: "2026-10-18T16:15:41.447Z"
              spawned: false
              bytes: 476863
              windows:
                - start: "2026-10-18T16:15:41.447Z"
                  end: "2026-10-18T10:48:02-07:00"
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              slot: ariadne:2
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: f5ebb78219742f86fb6725f3b3c3e9e4900924e44cef2866b641011a8f31aa60
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 195184182
              windows:
                - start: "2026-10-18T09:10:46-07:00"
                  end: "2026-10-18T10:48:02-07:00"
        queue_waits:
            - attempt_id: dd86d2996958c8e19e84dd8b9c49ee61
              slot: ariadne:2
              record_digest: 88fa31cfbfa77f1a36dc434f4bfe5ef427dff08c61f6e632b59666f877c498f6
              start: "2026-10-18T17:18:16.439065Z"
              end: "2026-10-18T17:28:43.782497Z"
              queue_seconds: 627.343432
              raw_seconds: 627.343432
              exact: true
        cuts:
            - boundary: M1
              dispatched: "2026-10-18T09:23:57-07:00"
              finalized: "2026-10-18T09:44:18-07:00"
              finalizing: false
            - boundary: M1
              dispatched: "2026-10-18T09:46:45-07:00"
              finalized: "2026-10-18T10:09:30-07:00"
              finalizing: true
            - boundary: M2
              dispatched: "2026-10-18T10:14:35-07:00"
              finalized: "2026-10-18T10:17:07-07:00"
              finalizing: false
            - boundary: M2
              dispatched: "2026-10-18T10:18:39-07:00"
              finalized: "2026-10-18T10:20:20-07:00"
              finalizing: true
      "n": 5
      timestamp: "2026-10-18T11:18:54-07:00"
      dispatched: "2026-10-18T11:17:29-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-6
          disposition: addressed
          note: Choice.Refusal no longer exists in pkg/reviewpolicy/choice.go; grep finds only judge.Refusal (reviewchoice.go:67), so the alias cannot expose it.
          round: 5
        - id: BR-11
          disposition: addressed
          note: Skip.Policy is set by the never-self-review filter (choose.go:69), the remedy drops retry-later for policy-only skips, TestRefusalRemedyFollowsTheSkips pins it, help.md:27-29 updated; the empty-skips sibling is raised as a new finding.
          round: 5
      findings:
        - kind: defect
          id: BR-12
          severity: Minor
          title: An empty refusal under other mode recommends reviewer_pairs, which Validate rejects outside explicit mode
          detail: 'This is the 3rd finding in family refusal-names-applicable-remedy. Earlier rounds fixed instances: a fixed sentence, then a single Policy boolean. Rule: every cause of refusal carries its own remedy, chosen by the cause together with the effective mode. Cause 1, a CLI that is unavailable: retry later. Cause 2, an explicit pair into the author family, or an unpaired session: fix reviewer_pairs. Cause 3, an other-mode plan that leaves no family (all families authored, or reviewer_order covers only author families): no config remedy except widening reviewer_order; name one explicitly. Cause 4, a skip for availability whose candidate would also have been policy-skipped: report the policy cause, not the transient one. Fix the rule: replace the Policy bool with a remedy or cause enum on Skip, plus an explicit cause for an empty plan. choice.go:68 currently sets policy=true when there are no skips. A probe in this checkout (chooseReviewers under Other with authors anthropic, openai and google) printed "pair this agent with another family in reviewer_pairs", and Policy{Mode: Other, Pairs: ...}.Validate() fails with "reviewer_pairs applies only to boundary_reviewer explicit". The same text reaches sdlc through judge.noReviewer. Add a table test that covers all four causes.'
          family: refusal-names-applicable-remedy
          round: 5
      recipe: milestone-review
      isolation: fenced
      reviewed: b7c033ab62cfa957c71465574ff4aaa93a74468b
      blocked: false
---

# Gate ledger — ariadne#441 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-18T09:44:18-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `move-preserves-contract` (kind: defect) judge's Choice.Refusal() no longer unwraps to ErrReviewDidNotRun; the test pinning it was rewritten
  pkg/reviewpolicy/choice.go:76 NoReviewerError.Unwrap returns ErrNoReviewer only, so the Refusal() errors returned by changecode.go:732,988, milestoneclose.go:903 and judge.go:166 stop matching judge.ErrReviewDidNotRun (only DispatchChoice's noReviewer wrapper still does). The moved TestResolve (policy_test.go:164) swapped its assertion to ErrNoReviewer, against the plan's "tests pass unchanged"; the hint text also changed from "name one with --agent" to "name one explicitly". Fix: in judge, return noReviewer{} from the refusal path and restore a judge test asserting both errors.Is(ErrReviewDidNotRun) and errors.As(*NoReviewerError).
- **BR-2** [Important] `plan-table-matches-code` (kind: defect) Plan Core-concepts rows name files and signatures the code doesn't have
  Record is in choice.go (not record.go) with signature (used, choice, skips); AuthorFamilies takes []agentenv.Commit, not []string; judge/policyalias.go doesn't exist (the aliases are in family.go, agent_resolver.go, reviewchoice.go, dispatch.go). Add a Revisions entry correcting the table.
- **BR-3** [Important] `gofmt-clean` (kind: defect) judge family_test.go and reviewchoice_test.go fail gofmt -l
  Runs of blank lines are left where the moved tests were cut out; run gofmt -w on both files.
- **BR-4** [Minor] `stale-identifier-in-comment` (kind: defect) policy_test.go helper comments still name ParseReviewerRecord / ReviewerRecord
  Rename the comments to ParseRecord / Record.
- **BR-5** [Minor] `shared-package-owns-its-tests` (kind: defect) pkg/agentenv has no tests, and reviewpolicy Load/AuthorFamilies are only tested from cmd/sdlc
  doc-review (M2) will import these directly; move or add package-level tests alongside them.

## Round 2 — 2026-10-18T10:09:30-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — judge.Refusal (reviewchoice.go:67) matches both NoReviewerError and ErrReviewDidNotRun and keeps the --agent hint; all 4 sites use it; TestRefusalKeepsJudgesContract pins it.
- BR-2 — addressed — Plan Revisions (M1 round 1) table checked against the head: Record in choice.go, AuthorFamilies takes []agentenv.Commit, aliases listed by their real files.
- BR-3 — addressed — gofmt -l on cmd and pkg lists no file in the diff; the extra blank lines are gone in family_test.go and reviewchoice_test.go.
- BR-4 — addressed — policy_test.go:209-210 comments now say ParseRecord / Record; no ReviewerRecord left in pkg/.
- BR-5 — addressed — Added pkg/agentenv/agentenv_test.go (Env, ParseLog, LogFormat) and pkg/reviewpolicy/load_test.go (Load, AuthorFamilies); both pass.

### Raised

- **BR-6** [Minor] `move-preserves-contract` (kind: defect) reviewpolicy Choice.Refusal() stays callable through the judge.ReviewChoice alias and drops judge's refusal contract
  2nd finding in this family. Rule: a method a moved type brings along through an alias must keep its old callers' contract, or sdlc must be unable to call it. Fix the rule, not the call site: remove Choice.Refusal (choice.go:55), or add a judge test that fails when non-test code outside pkg/ calls .Refusal(). Today no site test would fail if a site went back to choice.Refusal(): TestBoundaryReviewRefusesWithNoOtherFamily only checks the text "no reviewer available".

## Round 3 — 2026-10-18T10:17:07-07:00 (claude) — BLOCKED

### Raised

- **BR-7** [Critical] `never-self-review-holds-in-every-mode` (kind: defect) doc-review self-reviews under explicit mode when the session agent has no pair, and the record does not mark it degraded
  chooseReviewers (cmd/doc-review/choose.go:70-76) passes explicit policies straight to Resolve. Plan then asks Policy.Reviewer(main), which returns main when the session has no pair (pkg/reviewpolicy/policy.go:120-124). Probe: Pairs{codex:claude}, a claude session and a claude-authored doc give candidates=[claude] and record "claude (anthropic; author anthropic)" with no degraded marker. A pair whose reviewer shares the authors' family also slips through. This contradicts the help ("never reviews a document with its author's own family") and the M2 commit. Fix: in chooseReviewers, for a non-forced choice, move candidates whose family is among the authors into Skips (refusing if none remain). Add the case to TestChooseReviewers and remove the Explicit exemption at choose_test.go:140.
- **BR-8** [Important] `plan-table-matches-code` (kind: defect) The plan's as-built table still lists Choice.Refusal, which this window deleted from pkg/reviewpolicy/choice.go
  This is the 2nd finding in family plan-table-matches-code. Rule: any commit that adds, renames or removes an identifier listed in the as-built table updates the table in the same commit, checked by grepping every table identifier against the head at boundary close. Here: remove Refusal from plan.md:90's Choice row and add a Revisions note that the shared package's Choice.Refusal API was dropped (refusals are now &NoReviewerError{Skips}).
- **BR-9** [Minor] `stale-identifier-in-comment` (kind: defect) cmd/doc-review/main.go:2-3 and :15 still say the reviewer defaults to codex
  This is the 2nd finding in family stale-identifier-in-comment. Rule: when a behaviour or identifier changes, grep the package's comments for the old claim in the same commit (rg 'codex by default|defaults to codex|DefaultAgent' cmd/doc-review).
- **BR-10** [Minor] `refusal-names-applicable-remedy` (kind: defect) When an explicitly named agent fails to launch, the refusal tells the user to name one explicitly
  A forced candidate that fails to launch ends in NoReviewerError, whose text says "name one explicitly", though the user already did. Return the launch error directly when choice.Forced.

## Round 4 — 2026-10-18T10:20:20-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — choose.go:60-74 skips non-forced author-family candidates; reverting it turns TestChooseReviewers (2 explicit cases) and TestChooseReviewersRecordsTheSelfReviewSkip red.
- BR-8 — addressed — plan.md:90 Choice row now lists only First; all as-built identifiers grepped present at head; Revisions entry records the Refusal removal and the same-commit rule.
- BR-9 — addressed — main.go:2-4,16 rewritten; rg 'codex by default|defaults to codex|DefaultAgent' over cmd/doc-review, fix SKILL.md and docflow atlas returns nothing.
- BR-10 — addressed — review.go:99-101 returns launchErr when Forced; removing it turns TestRunReviewForcedLaunchFailure red with the old name-one-explicitly text.

### Raised

- **BR-11** [Minor] `refusal-names-applicable-remedy` (kind: defect) A refusal caused by the never-self-review skip still advises retry later, though only a config change or a named agent can help
  This is the 2nd finding in family refusal-names-applicable-remedy. Rule: the remedy a NoReviewerError names comes from its skip reasons: transient skips (not on PATH, failed launch, usage limit) suggest retry later; policy skips (shares the author family) suggest fixing reviewer_pairs or naming an agent; forced failures return their own error. help.md:26-27 should also say that an explicit pair into the author's family, or an unpaired session, refuses.

## Round 5 — 2026-10-18T11:18:54-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — Choice.Refusal no longer exists in pkg/reviewpolicy/choice.go; grep finds only judge.Refusal (reviewchoice.go:67), so the alias cannot expose it.
- BR-11 — addressed — Skip.Policy is set by the never-self-review filter (choose.go:69), the remedy drops retry-later for policy-only skips, TestRefusalRemedyFollowsTheSkips pins it, help.md:27-29 updated; the empty-skips sibling is raised as a new finding.

### Raised

- **BR-12** [Minor] `refusal-names-applicable-remedy` (kind: defect) An empty refusal under other mode recommends reviewer_pairs, which Validate rejects outside explicit mode
  This is the 3rd finding in family refusal-names-applicable-remedy. Earlier rounds fixed instances: a fixed sentence, then a single Policy boolean. Rule: every cause of refusal carries its own remedy, chosen by the cause together with the effective mode. Cause 1, a CLI that is unavailable: retry later. Cause 2, an explicit pair into the author family, or an unpaired session: fix reviewer_pairs. Cause 3, an other-mode plan that leaves no family (all families authored, or reviewer_order covers only author families): no config remedy except widening reviewer_order; name one explicitly. Cause 4, a skip for availability whose candidate would also have been policy-skipped: report the policy cause, not the transient one. Fix the rule: replace the Policy bool with a remedy or cause enum on Skip, plus an explicit cause for an empty plan. choice.go:68 currently sets policy=true when there are no skips. A probe in this checkout (chooseReviewers under Other with authors anthropic, openai and google) printed "pair this agent with another family in reviewer_pairs", and Policy{Mode: Other, Pairs: ...}.Validate() fails with "reviewer_pairs applies only to boundary_reviewer explicit". The same text reaches sdlc through judge.noReviewer. Add a table test that covers all four causes.

## Open findings

- **BR-12** [Minor] `refusal-names-applicable-remedy` (kind: defect) An empty refusal under other mode recommends reviewer_pairs, which Validate rejects outside explicit mode
