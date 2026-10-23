---
gate: boundary-review
issue: 566
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 013f0c7e-f49c-4cef-a668-38ce8cc73e61
            repository: github.com/xianxu/ariadne
            issue_id: "000566"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/52672d8f33959adc
            reviewed_head: 644766d85c579461760e8be9014524656b510df3
            slot: ariadne:5
            requested_at: "2026-10-23T10:48:47.02757Z"
          phase: finished
          pid: 87650
          started_at: "2026-10-23T10:48:47.178372Z"
          finished_at: "2026-10-23T10:54:25.447005Z"
          session_id: 7a0af5bc-09bd-4246-a505-a9982bc61815
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 644766d85c579461760e8be9014524656b510df3
        tracker_head: 880b1cbfc9b03c260175e582aaa125f0cc5f6640
        branch_point: 73492328ed3745236372ad859b538029f1488ca0
        since: "2026-10-23T03:16:37-07:00"
        until: "2026-10-23T03:48:18-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T03:16:37-07:00"
              end: "2026-10-23T03:48:18-07:00"
              began: "2026-10-23T03:16:37-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 7035850d7166a1cc79fe1be8b9f04285870aeb198fd07b4e029cfa87fd1fda54
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 2495845
              windows:
                - start: "2026-10-23T03:16:37-07:00"
                  end: "2026-10-23T03:48:18-07:00"
      "n": 1
      timestamp: "2026-10-23T03:54:26-07:00"
      dispatched: "2026-10-23T03:48:47-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: issue rename skips ownership/status checks whenever the card already sits at the new path, not only when the rename is finished
          detail: issuerename.go:565 runs renamable only if card.Path != newCard; in the half-done state (card moved, details old) or a pre-existing card/details mismatch on a working issue owned by someone else, a non-owner can move the details on main unchecked. Skip only when the first observation is RenameDone, otherwise apply requireOwnedToEdit before RenameDetails; add a refusal test for that case.
          family: guard-keyed-on-partial-state
          round: 1
        - kind: defect
          id: BR-2
          severity: Important
          title: Legacy issue new slug cap has no test (plan Task 1 Step 6 promised one in legacymode_test.go)
          detail: legacymode.go:90 now calls FileSlug but no legacy filer test asserts a 300-char title yields a basename within 6+1+80+3 bytes; reverting it to Slugify would pass the suite.
          family: consumer-without-regression-test
          round: 1
        - kind: defect
          id: BR-3
          severity: Minor
          title: commitpublication.go publishes source-commit trees without passing CheckPathLimits; not named in the plan's not-guarded list
          detail: Low risk since such trees come from a working tree, but the spec says refuse any oversized path; name the exclusion and its rationale in the plan.
          family: guard-coverage-enumeration
          round: 1
        - kind: defect
          id: BR-4
          severity: Minor
          title: Plan Core concepts omits DetailNames and Revisions omits the dropped plans-on-main precondition
          detail: Both are recorded only in the Log; add a Revisions entry and a PURE row for tracker.DetailNames.
          family: plan-code-drift
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 62ba3204-8a7d-4db1-b1e4-9032fec37274
            repository: github.com/xianxu/ariadne
            issue_id: "000566"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: claude
            root: /Volumes/TestScratch/review-checkouts/214ccabd26f40ca0
            reviewed_head: ec1d2831d02f2e7df3f3e7a25134663b4bcc080f
            slot: ariadne:5
            requested_at: "2026-10-23T10:58:55.171001Z"
          phase: finished
          pid: 47353
          started_at: "2026-10-23T10:58:55.295634Z"
          finished_at: "2026-10-23T11:02:53.919166Z"
          session_id: 4d933977-0093-4217-9444-18cb77687338
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: ec1d2831d02f2e7df3f3e7a25134663b4bcc080f
        tracker_head: 880b1cbfc9b03c260175e582aaa125f0cc5f6640
        branch_point: 73492328ed3745236372ad859b538029f1488ca0
        since: "2026-10-23T03:16:37-07:00"
        until: "2026-10-23T03:58:24-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-23T03:16:37-07:00"
              end: "2026-10-23T03:58:24-07:00"
              began: "2026-10-23T03:16:37-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: claude
              id: 03bd9c39-c2ac-4044-9095-3ae230ea12e1
              slot: ariadne:5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: e8c1072a4fab2a735b8466f1c8f09d7a5c38731a01648f74a5ae1f0030eeb27e
              started: "2026-10-23T08:59:44.621Z"
              spawned: false
              bytes: 2861513
              windows:
                - start: "2026-10-23T03:16:37-07:00"
                  end: "2026-10-23T03:58:24-07:00"
        cuts:
            - dispatched: "2026-10-23T03:48:47-07:00"
              finalized: "2026-10-23T03:54:26-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-23T04:02:55-07:00"
      dispatched: "2026-10-23T03:58:55-07:00"
      agent: claude
      reviewer: claude (anthropic; author anthropic)
      dispose:
        - id: BR-1
          disposition: addressed
          note: renamable now gates every non-Done step; reverting to card-only turns the half-done another-owner subtest RED.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestLegacyRepositoryIssueNewCapsLongTitleBasename added; RED with Slugify restored in legacymode.go.
          round: 2
        - id: BR-3
          disposition: addressed
          note: PublishCommitChecked checks the base-to-merged diff; removing it turns TestPublishCommitRefusesOverlongName RED.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Plan Core concepts lists DetailNames; Revisions records the dropped plans-on-main precondition.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Important
          title: issue rename checks renamable (status open, ownership) on a stale snapshot; MoveCard moves whatever bytes the fresh tip holds
          detail: 'issuerename.go:680-688 evaluates renamable on cur, then MovePath (rename.go:399) copies the fresh tip''s card without comparing it to cur. A claim landing in between gets its now-working card renamed, its branch stranded on the old stem, and further renames refused. The plan''s State design wrongly says no action is needed. Fix: pass the expected source bytes into MovePath and refuse on mismatch so a rerun re-checks; add a test that changes the card in the gap.'
          family: guard-observed-outside-cas
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#566 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-23T03:54:26-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `guard-keyed-on-partial-state` (kind: defect) issue rename skips ownership/status checks whenever the card already sits at the new path, not only when the rename is finished
  issuerename.go:565 runs renamable only if card.Path != newCard; in the half-done state (card moved, details old) or a pre-existing card/details mismatch on a working issue owned by someone else, a non-owner can move the details on main unchecked. Skip only when the first observation is RenameDone, otherwise apply requireOwnedToEdit before RenameDetails; add a refusal test for that case.
- **BR-2** [Important] `consumer-without-regression-test` (kind: defect) Legacy issue new slug cap has no test (plan Task 1 Step 6 promised one in legacymode_test.go)
  legacymode.go:90 now calls FileSlug but no legacy filer test asserts a 300-char title yields a basename within 6+1+80+3 bytes; reverting it to Slugify would pass the suite.
- **BR-3** [Minor] `guard-coverage-enumeration` (kind: defect) commitpublication.go publishes source-commit trees without passing CheckPathLimits; not named in the plan's not-guarded list
  Low risk since such trees come from a working tree, but the spec says refuse any oversized path; name the exclusion and its rationale in the plan.
- **BR-4** [Minor] `plan-code-drift` (kind: defect) Plan Core concepts omits DetailNames and Revisions omits the dropped plans-on-main precondition
  Both are recorded only in the Log; add a Revisions entry and a PURE row for tracker.DetailNames.

## Round 2 — 2026-10-23T04:02:55-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — renamable now gates every non-Done step; reverting to card-only turns the half-done another-owner subtest RED.
- BR-2 — addressed — TestLegacyRepositoryIssueNewCapsLongTitleBasename added; RED with Slugify restored in legacymode.go.
- BR-3 — addressed — PublishCommitChecked checks the base-to-merged diff; removing it turns TestPublishCommitRefusesOverlongName RED.
- BR-4 — addressed — Plan Core concepts lists DetailNames; Revisions records the dropped plans-on-main precondition.

### Raised

- **BR-5** [Important] `guard-observed-outside-cas` (kind: defect) issue rename checks renamable (status open, ownership) on a stale snapshot; MoveCard moves whatever bytes the fresh tip holds
  issuerename.go:680-688 evaluates renamable on cur, then MovePath (rename.go:399) copies the fresh tip's card without comparing it to cur. A claim landing in between gets its now-working card renamed, its branch stranded on the old stem, and further renames refused. The plan's State design wrongly says no action is needed. Fix: pass the expected source bytes into MovePath and refuse on mismatch so a rerun re-checks; add a test that changes the card in the gap.

## Open findings

- **BR-5** [Important] `guard-observed-outside-cas` (kind: defect) issue rename checks renamable (status open, ownership) on a stale snapshot; MoveCard moves whatever bytes the fresh tip holds
