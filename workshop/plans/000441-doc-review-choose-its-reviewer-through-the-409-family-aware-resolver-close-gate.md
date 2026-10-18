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

## Open findings

- **BR-6** [Minor] `move-preserves-contract` (kind: defect) reviewpolicy Choice.Refusal() stays callable through the judge.ReviewChoice alias and drops judge's refusal contract
