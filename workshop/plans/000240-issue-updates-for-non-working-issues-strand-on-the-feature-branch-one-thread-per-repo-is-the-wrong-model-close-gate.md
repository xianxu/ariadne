---
gate: boundary-review
issue: 240
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-14T09:36:09-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Important
          title: Define recovery after publication succeeds but only part of cleanup completes
          detail: |-
            ARCH-ORDER: Restored tracked files no longer equal published bytes, and removed untracked files fail admission, so repeating the original selection cannot finish under the stated rules. Name the authoritative state/event model and structural transition boundary, define how already-cleaned paths are recognized safely, and specify controllable interruption testing that proves retry neither republishes nor overwrites later edits.
            (carried from plan-quality PQ-1, deferred to the boundary review)
          family: partial-progress-recovery
          round: 1
        - id: BR-2
          severity: Important
          title: Replace prose case inventories with named function-level test strategies
          detail: |-
            Tasks 1 and 2 enumerate test cases rather than stating an adversarial input class and mechanical guard per risky function. Give strategies for TrunkProsePath, prosePublishDecision, runTrunkPublish and the cleanup/transition component; ARCH-PURE also requires separating lexical eligibility from filesystem type and identity checks.
            (carried from plan-quality PQ-2, deferred to the boundary review)
          family: function-level-test-strategy
          round: 1
        - id: BR-3
          severity: Important
          title: Remove the single M1 milestone and its redundant milestone-close
          detail: |-
            The issue has one atomic implementation boundary; verification and shipping do not constitute a second milestone. Replace M1 with plain checkboxes and use sdlc close for the mandatory review, consistent with AGENTS.md §3.
            (carried from plan-quality PQ-3, deferred to the boundary review)
          family: atomic-review-boundary
          round: 1
        - id: BR-4
          severity: Minor
          title: State remote time and selected-input bounds
          detail: |-
            ARCH-CONSTRAINTS: Explicit paths and three CAS attempts do not bound pinned bytes or a stalled remote operation while holding the shared repository lock. State the relevant budgets or supported ranges, their basis, and exceeded-budget behavior; identify how command cancellation and the existing BoundRemote facility apply.
            (carried from plan-quality PQ-4, deferred to the boundary review)
          family: explicit-operating-envelope
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: c66f2889426d7d9024853f455ab9eeac0e206efc
        tracker_head: 3c84a24f3c320393999bfe8b91d700ea06ffee26
        branch_point: 407e48a39057503fab7cdecaf6c69ce338ce7631
        since: "2026-10-14T08:53:15-07:00"
        until: "2026-10-14T09:32:01-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T08:53:15-07:00"
              end: "2026-10-14T09:32:01-07:00"
              began: "2026-10-14T08:53:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 3c0c91a2917a38c4f5fd5e6fa9008a7a72f177aa164e87f8da424f1e017eec5e
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 11627560
              windows:
                - start: "2026-10-14T08:53:15-07:00"
                  end: "2026-10-14T09:32:01-07:00"
            - harness: codex
              id: 01a13b21-7327-7c60-9006-edb3c90098ae
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c3162bf2b1730739eeab71cce50f8496d7da0898d39c3ef94f0cb469e5e0af89
              started: "2026-10-14T15:56:50.688Z"
              spawned: false
              bytes: 527136
              windows:
                - start: "2026-10-14T15:56:50.688Z"
                  end: "2026-10-14T09:32:01-07:00"
      "n": 2
      timestamp: "2026-10-14T09:36:09-07:00"
      dispatched: "2026-10-14T09:32:19-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-5
          severity: Critical
          title: Cleanup interprets the selected filename as a Git pattern and discards unrelated edits
          detail: cmd/sdlc/trunkpublish.go:250 passes p to git restore without literal-pathspec handling, although trunkProsePath accepts wildcard characters. A temporary regression overlay invoking runTrunkPublish with a tracked workshop/parley/*.md filename published successfully but restored an unrelated dirty other.md to its base bytes. Use literal pathspecs consistently for selected-path Git operations and add regression coverage for wildcard filenames.
          family: literal-selected-path
          round: 2
        - kind: defect
          id: BR-6
          severity: Critical
          title: Cleanup can overwrite edits made after its validation
          detail: 'cmd/sdlc/trunkpublish.go:232-253 reads and compares the file, then separately restores or removes it; these destructive effects are not conditional on the validated file still being current. A deterministic temporary overlay injected an edit when the restore subprocess started: finishTrunkPublish returned nil and replaced that edit with base bytes. ARCH-ORDER: redesign cleanup to preserve intervening writes, with explicit cleanup transitions and controllable ordering tests for tracked restore, untracked removal, and HEAD/index changes.'
          family: cleanup-preserves-concurrent-work
          round: 2
        - kind: defect
          id: BR-7
          severity: Important
          title: README update appears missing for trunk publish --path
          detail: cmd/sdlc/main.go:151 introduces a user-facing command and cmd/sdlc/trunkpublish.go:34 introduces its --path flag, but README.md has no change in the pinned range. Add usage, eligibility, conflict, and retry guidance alongside README.md:15's concurrent-work documentation, linking to the updated atlas.
          family: readme-user-surface
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#240 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T09:36:09-07:00 (sdlc) — passed

### Raised

- **BR-1** [Important] `partial-progress-recovery` Define recovery after publication succeeds but only part of cleanup completes
  ARCH-ORDER: Restored tracked files no longer equal published bytes, and removed untracked files fail admission, so repeating the original selection cannot finish under the stated rules. Name the authoritative state/event model and structural transition boundary, define how already-cleaned paths are recognized safely, and specify controllable interruption testing that proves retry neither republishes nor overwrites later edits.
  (carried from plan-quality PQ-1, deferred to the boundary review)
- **BR-2** [Important] `function-level-test-strategy` Replace prose case inventories with named function-level test strategies
  Tasks 1 and 2 enumerate test cases rather than stating an adversarial input class and mechanical guard per risky function. Give strategies for TrunkProsePath, prosePublishDecision, runTrunkPublish and the cleanup/transition component; ARCH-PURE also requires separating lexical eligibility from filesystem type and identity checks.
  (carried from plan-quality PQ-2, deferred to the boundary review)
- **BR-3** [Important] `atomic-review-boundary` Remove the single M1 milestone and its redundant milestone-close
  The issue has one atomic implementation boundary; verification and shipping do not constitute a second milestone. Replace M1 with plain checkboxes and use sdlc close for the mandatory review, consistent with AGENTS.md §3.
  (carried from plan-quality PQ-3, deferred to the boundary review)
- **BR-4** [Minor] `explicit-operating-envelope` State remote time and selected-input bounds
  ARCH-CONSTRAINTS: Explicit paths and three CAS attempts do not bound pinned bytes or a stalled remote operation while holding the shared repository lock. State the relevant budgets or supported ranges, their basis, and exceeded-budget behavior; identify how command cancellation and the existing BoundRemote facility apply.
  (carried from plan-quality PQ-4, deferred to the boundary review)

## Round 2 — 2026-10-14T09:36:09-07:00 (codex) — BLOCKED

### Raised

- **BR-5** [Critical] `literal-selected-path` (kind: defect) Cleanup interprets the selected filename as a Git pattern and discards unrelated edits
  cmd/sdlc/trunkpublish.go:250 passes p to git restore without literal-pathspec handling, although trunkProsePath accepts wildcard characters. A temporary regression overlay invoking runTrunkPublish with a tracked workshop/parley/*.md filename published successfully but restored an unrelated dirty other.md to its base bytes. Use literal pathspecs consistently for selected-path Git operations and add regression coverage for wildcard filenames.
- **BR-6** [Critical] `cleanup-preserves-concurrent-work` (kind: defect) Cleanup can overwrite edits made after its validation
  cmd/sdlc/trunkpublish.go:232-253 reads and compares the file, then separately restores or removes it; these destructive effects are not conditional on the validated file still being current. A deterministic temporary overlay injected an edit when the restore subprocess started: finishTrunkPublish returned nil and replaced that edit with base bytes. ARCH-ORDER: redesign cleanup to preserve intervening writes, with explicit cleanup transitions and controllable ordering tests for tracked restore, untracked removal, and HEAD/index changes.
- **BR-7** [Important] `readme-user-surface` (kind: defect) README update appears missing for trunk publish --path
  cmd/sdlc/main.go:151 introduces a user-facing command and cmd/sdlc/trunkpublish.go:34 introduces its --path flag, but README.md has no change in the pinned range. Add usage, eligibility, conflict, and retry guidance alongside README.md:15's concurrent-work documentation, linking to the updated atlas.

## Open findings

- **BR-1** [Important] `partial-progress-recovery` Define recovery after publication succeeds but only part of cleanup completes
- **BR-2** [Important] `function-level-test-strategy` Replace prose case inventories with named function-level test strategies
- **BR-3** [Important] `atomic-review-boundary` Remove the single M1 milestone and its redundant milestone-close
- **BR-4** [Minor] `explicit-operating-envelope` State remote time and selected-input bounds
- **BR-5** [Critical] `literal-selected-path` (kind: defect) Cleanup interprets the selected filename as a Git pattern and discards unrelated edits
- **BR-6** [Critical] `cleanup-preserves-concurrent-work` (kind: defect) Cleanup can overwrite edits made after its validation
- **BR-7** [Important] `readme-user-surface` (kind: defect) README update appears missing for trunk publish --path
