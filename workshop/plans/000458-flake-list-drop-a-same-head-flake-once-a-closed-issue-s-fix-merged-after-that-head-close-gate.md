---
gate: boundary-review
issue: 458
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 8d765e44-9a65-4e75-91b9-905d30a95b5e
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/28a4f2053ee07160
            reviewed_head: 1b56f757d2b7852f06c162fc2b39a97c0078eb14
            slot: ariadne:1
            requested_at: "2026-10-19T12:25:12.330439Z"
          phase: finished
          pid: 47557
          started_at: "2026-10-19T12:25:12.427098Z"
          finished_at: "2026-10-19T12:27:07.102181Z"
          session_id: 01a1541f-7f0e-7d72-a761-660c9890a1fd
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 1b56f757d2b7852f06c162fc2b39a97c0078eb14
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T04:55:16-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T04:55:16-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: ee47cbf5e7b1d5e81968b86e16a1827fbcd2c412de47c79251443eb97bb5cf6a
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 141839773
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T04:55:16-07:00"
      "n": 1
      timestamp: "2026-10-19T05:27:07-07:00"
      dispatched: "2026-10-19T05:25:12-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - id: BR-1
          severity: Critical
          title: 'ARCH-PURE: the declared pure decision performs injected Git IO'
          detail: cmd/sdlc/flake_fixed.go:219–223 invokes contains, wired to real Git in production and in TestFixedFlakeAgainstRealGitMergeHistory. The plan’s Core concepts table at lines 19–22 instead declares a pure FixedFlakeDecision taking ancestry answers; neither that symbol nor FixedFlakeIssue exists under the declared names. Move ancestry execution into the integration shell and pass explicit results into a pure decision, or revise the table to name the actual entities and classify the orchestration as INTEGRATION.
          family: pure-core-contract
          round: 1
        - id: BR-2
          severity: Important
          title: README update appears missing for flake list -v
          detail: cmd/sdlc/flake_cmd.go:261 introduces --verbose/-v and changes default list behavior, but README.md:525 still describes only the previous daily view. Update that passage with default suppression, verbose attribution, and recurrence behavior; atlas coverage does not replace README usage.
          family: user-surface-documentation
          round: 1
        - id: BR-3
          severity: Important
          title: Command tests do not pin default suppression or recurrence rendering
          detail: cmd/sdlc/flake_fixed_test.go:83 tests the classifier against real Git, while lines 124–147 exercise only verbose fixed-row rendering with synthetic ancestry answers. No test asserts that the default command hides a proven fixed row or that a recurrence stays visible with its annotation. Add command-level tests covering both modes and conservative failure cases, using the real-history seam required by Done when.
          family: observable-behavior-regression-coverage
          round: 1
        - id: BR-4
          severity: Important
          title: Per-row evidence failures discard the promised coverage reason
          detail: cmd/sdlc/flake_fixed.go:203–225 collapses unreadable details, ambiguous ownership, and ancestry errors into flakeUnknown without a reason. cmd/sdlc/flake_cmd.go:487 prints those rows unchanged; its coverage message handles only catalogue-wide FixedErr. Preserve a reason in the decision result and render it so these cases satisfy the plan’s line-34 promise that unknown rows remain visible with a coverage reason.
          family: uncertainty-remains-observable
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 09f5f5d6-fdf5-4217-9b31-b3337ecf0308
            repository: github.com/xianxu/ariadne
            issue_id: "000458"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/feb76cf032713ca5
            reviewed_head: e4e8590d273723c84d7cc000c381abc16d622cdd
            slot: ariadne:1
            requested_at: "2026-10-19T12:59:54.84211Z"
          phase: finished
          pid: 84333
          started_at: "2026-10-19T12:59:54.936514Z"
          finished_at: "2026-10-19T13:02:41.515882Z"
          session_id: 01a1543f-45bc-7b93-b21d-fabbef7cb8e7
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-queue-v1
        repository: github.com/xianxu/ariadne
        code_head: e4e8590d273723c84d7cc000c381abc16d622cdd
        tracker_head: d45848d988ab04490cf0d9bfce88a1eaed8da451
        branch_point: a11504c417e9c8789327362895bc88c9fad042df
        since: "2026-10-19T04:32:38-07:00"
        until: "2026-10-19T05:32:44-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T04:32:38-07:00"
              end: "2026-10-19T05:32:44-07:00"
              began: "2026-10-19T04:32:38-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              slot: ariadne:1
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 850973e687bb57065e56875c285c9ea04f9d928b2bf2d8347071c2aed9033dd6
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 142709310
              windows:
                - start: "2026-10-19T04:32:38-07:00"
                  end: "2026-10-19T05:32:44-07:00"
        queue_waits:
            - attempt_id: 47932ec61a1603c497832d736497e4f0
              slot: ariadne:1
              record_digest: 26e688dc36fb19ea4a4f2f0ba6b5a0c1a7640426e201e1a37b7e660686969336
              start: "2026-10-19T11:55:43.408394Z"
              end: "2026-10-19T12:12:29.80791Z"
              queue_seconds: 1006.399516
              raw_seconds: 1006.399516
              exact: true
        cuts:
            - dispatched: "2026-10-19T05:25:12-07:00"
              finalized: "2026-10-19T05:27:07-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T06:02:42-07:00"
      dispatched: "2026-10-19T05:59:54-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan lines 19–23 now name the implemented entities and classify classifyFixedFlake as INTEGRATION, matching its injected Git calls. fixedIssueCovers remains independently testable without IO.
          round: 2
        - id: BR-2
          disposition: addressed
          note: README.md:527–533 documents default suppression, flake list -v, recurrence visibility, and uncertain evidence; these descriptions match flake_cmd.go:474–486.
          round: 2
        - id: BR-3
          disposition: addressed
          note: TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default and verbose output, recurrence, and ambiguous coverage using real merge history. Temporary source overlays disabling suppression or recurrence annotation each made it fail.
          round: 2
        - id: BR-4
          disposition: addressed
          note: classifyFixedFlake now returns reasons for unreadable details, ambiguous ownership, missing merges, and ancestry failures. Decision tests assert these reasons; removing coverage-reason rendering made the real-history command test fail.
          round: 2
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: 'ARCH-PURPOSE: package-basename matching can hide an unrelated test'
          detail: cmd/sdlc/flake_fixed.go:171 reduces package identity to path.Base(pkg). For a failure named github.com/xianxu/ariadne/cmd/internal/launcher.TestX, details reading “TestX fails in pkg/launcher/run_test.go” incorrectly establish ownership. A temporary regression test confirmed that classifyFixedFlake returns flakeFixed when the issue’s merge is on main but absent from the failing head, causing default rendering to suppress the unrelated failure. bareCollision does not protect this case when only one package occurs in retained attempts. Require unambiguous package identity for textual ownership, retain uncertain rows, and cover same-basename packages across substantive sections and dedicated Log lines.
          family: exact-test-ownership
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#458 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T05:27:07-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `pure-core-contract` ARCH-PURE: the declared pure decision performs injected Git IO
  cmd/sdlc/flake_fixed.go:219–223 invokes contains, wired to real Git in production and in TestFixedFlakeAgainstRealGitMergeHistory. The plan’s Core concepts table at lines 19–22 instead declares a pure FixedFlakeDecision taking ancestry answers; neither that symbol nor FixedFlakeIssue exists under the declared names. Move ancestry execution into the integration shell and pass explicit results into a pure decision, or revise the table to name the actual entities and classify the orchestration as INTEGRATION.
- **BR-2** [Important] `user-surface-documentation` README update appears missing for flake list -v
  cmd/sdlc/flake_cmd.go:261 introduces --verbose/-v and changes default list behavior, but README.md:525 still describes only the previous daily view. Update that passage with default suppression, verbose attribution, and recurrence behavior; atlas coverage does not replace README usage.
- **BR-3** [Important] `observable-behavior-regression-coverage` Command tests do not pin default suppression or recurrence rendering
  cmd/sdlc/flake_fixed_test.go:83 tests the classifier against real Git, while lines 124–147 exercise only verbose fixed-row rendering with synthetic ancestry answers. No test asserts that the default command hides a proven fixed row or that a recurrence stays visible with its annotation. Add command-level tests covering both modes and conservative failure cases, using the real-history seam required by Done when.
- **BR-4** [Important] `uncertainty-remains-observable` Per-row evidence failures discard the promised coverage reason
  cmd/sdlc/flake_fixed.go:203–225 collapses unreadable details, ambiguous ownership, and ancestry errors into flakeUnknown without a reason. cmd/sdlc/flake_cmd.go:487 prints those rows unchanged; its coverage message handles only catalogue-wide FixedErr. Preserve a reason in the decision result and render it so these cases satisfy the plan’s line-34 promise that unknown rows remain visible with a coverage reason.

## Round 2 — 2026-10-19T06:02:42-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Plan lines 19–23 now name the implemented entities and classify classifyFixedFlake as INTEGRATION, matching its injected Git calls. fixedIssueCovers remains independently testable without IO.
- BR-2 — addressed — README.md:527–533 documents default suppression, flake list -v, recurrence visibility, and uncertain evidence; these descriptions match flake_cmd.go:474–486.
- BR-3 — addressed — TestFlakeListRendersFixedRecurrenceAndCoverageFromRealGit exercises default and verbose output, recurrence, and ambiguous coverage using real merge history. Temporary source overlays disabling suppression or recurrence annotation each made it fail.
- BR-4 — addressed — classifyFixedFlake now returns reasons for unreadable details, ambiguous ownership, missing merges, and ancestry failures. Decision tests assert these reasons; removing coverage-reason rendering made the real-history command test fail.

### Raised

- **BR-5** [Critical] `exact-test-ownership` (kind: defect) ARCH-PURPOSE: package-basename matching can hide an unrelated test
  cmd/sdlc/flake_fixed.go:171 reduces package identity to path.Base(pkg). For a failure named github.com/xianxu/ariadne/cmd/internal/launcher.TestX, details reading “TestX fails in pkg/launcher/run_test.go” incorrectly establish ownership. A temporary regression test confirmed that classifyFixedFlake returns flakeFixed when the issue’s merge is on main but absent from the failing head, causing default rendering to suppress the unrelated failure. bareCollision does not protect this case when only one package occurs in retained attempts. Require unambiguous package identity for textual ownership, retain uncertain rows, and cover same-basename packages across substantive sections and dedicated Log lines.

## Open findings

- **BR-5** [Critical] `exact-test-ownership` (kind: defect) ARCH-PURPOSE: package-basename matching can hide an unrelated test
