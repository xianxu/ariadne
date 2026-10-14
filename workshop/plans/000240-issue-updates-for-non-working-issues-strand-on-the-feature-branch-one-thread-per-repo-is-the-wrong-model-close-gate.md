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
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 48255461dfd57311a8fd24a74d818826f97320d3
        tracker_head: 7b1e911628ebfc46694388bf87ee155ce914c023
        branch_point: 407e48a39057503fab7cdecaf6c69ce338ce7631
        since: "2026-10-14T08:53:15-07:00"
        until: "2026-10-14T10:18:40-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T08:53:15-07:00"
              end: "2026-10-14T10:18:40-07:00"
              began: "2026-10-14T08:53:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 4c103e4721ebcf00f573a128ff39404a99d50e011491319463187e09f09ad283
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 12621213
              windows:
                - start: "2026-10-14T08:53:15-07:00"
                  end: "2026-10-14T10:18:40-07:00"
            - harness: codex
              id: 01a13b21-7327-7c60-9006-edb3c90098ae
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c3162bf2b1730739eeab71cce50f8496d7da0898d39c3ef94f0cb469e5e0af89
              started: "2026-10-14T15:56:50.688Z"
              spawned: false
              bytes: 527136
              windows:
                - start: "2026-10-14T15:56:50.688Z"
                  end: "2026-10-14T10:18:40-07:00"
        cuts:
            - dispatched: "2026-10-14T09:32:19-07:00"
              finalized: "2026-10-14T09:36:09-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-14T10:23:27-07:00"
      dispatched: "2026-10-14T10:19:16-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: not-addressed
          note: 'ARCH-ORDER / ARCH-FUNERAL: cmd/sdlc/trunkpublish.go:276 creates an intermediate renamed-backup state that retry cannot recover. A deterministic interruption at trunkProseAfterMove reproduced both failures: tracked retry refuses at line 183 as a local deletion; untracked retry reports already clean at line 185 while retaining the backup. Define authoritative cleanup states and safe recovery for every intermediate effect, including process death, with production-transition sequence tests.'
          round: 3
        - id: BR-2
          disposition: addressed
          note: Plan Tasks 1–2 now name adversarial inputs and mechanical guards for the path predicate, publication decision, command, and cleanup. Lexical eligibility is separated from filesystem inspection; the pure tests invoke the predicates directly without IO. Cleanup strategy completeness remains covered by BR-1 and BR-6.
          round: 3
        - id: BR-3
          disposition: addressed
          note: The issue Plan uses plain checkboxes; the implementation plan contains no M1 boundary or milestone-close. This matches AGENTS.base.md:51's single-pass close rule.
          round: 3
        - id: BR-4
          disposition: addressed
          note: The plan specifies a 1 MiB prose envelope, two-minute shared remote deadline, cancellation, and three CAS attempts. cmd/sdlc/trunkpublish.go enforces the file-size limit and invokes BoundRemote; the existing Git seam supplies cancellation and the CAS budget.
          round: 3
        - id: BR-5
          disposition: addressed
          note: 'Selected-path staging checks use literal pathspecs and cleanup reads the exact pinned HEAD blob instead of Git restore patterns. TestTrunkPublishTreatsSelectedWildcardFilenameLiterally passes at pinned head and fails when the pre-fix cleanup is restored in a disposable checkout: unrelated other.md loses its dirty edit.'
          round: 3
        - id: BR-6
          disposition: not-addressed
          note: 'ARCH-ORDER: cmd/sdlc/trunkpublish.go:313–317 and 332–336 still separate the final backup-byte check from unlink. A disposable overlay added only an ordering hook immediately before unlink; writing through a descriptor opened before rename made both tracked and untracked cleanup return nil with no backup, while the later edit disappeared. Exclusive creation also does not exclude subsequent writers during replacement.Write at line 327. Preserve recoverable data throughout cleanup and test all destructive-effect boundaries, rather than only the after-move event.'
          round: 3
        - id: BR-7
          disposition: addressed
          note: README.md:28–37 adds command usage, eligibility, moved-main refusal, uncertain-push retry guidance, and an atlas link. These passages match the allowlist and publication checks; atlas/workflow/sdlc-binary.md also documents the new surface.
          round: 3
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

## Round 3 — 2026-10-14T10:23:27-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — not-addressed — ARCH-ORDER / ARCH-FUNERAL: cmd/sdlc/trunkpublish.go:276 creates an intermediate renamed-backup state that retry cannot recover. A deterministic interruption at trunkProseAfterMove reproduced both failures: tracked retry refuses at line 183 as a local deletion; untracked retry reports already clean at line 185 while retaining the backup. Define authoritative cleanup states and safe recovery for every intermediate effect, including process death, with production-transition sequence tests.
- BR-2 — addressed — Plan Tasks 1–2 now name adversarial inputs and mechanical guards for the path predicate, publication decision, command, and cleanup. Lexical eligibility is separated from filesystem inspection; the pure tests invoke the predicates directly without IO. Cleanup strategy completeness remains covered by BR-1 and BR-6.
- BR-3 — addressed — The issue Plan uses plain checkboxes; the implementation plan contains no M1 boundary or milestone-close. This matches AGENTS.base.md:51's single-pass close rule.
- BR-4 — addressed — The plan specifies a 1 MiB prose envelope, two-minute shared remote deadline, cancellation, and three CAS attempts. cmd/sdlc/trunkpublish.go enforces the file-size limit and invokes BoundRemote; the existing Git seam supplies cancellation and the CAS budget.
- BR-5 — addressed — Selected-path staging checks use literal pathspecs and cleanup reads the exact pinned HEAD blob instead of Git restore patterns. TestTrunkPublishTreatsSelectedWildcardFilenameLiterally passes at pinned head and fails when the pre-fix cleanup is restored in a disposable checkout: unrelated other.md loses its dirty edit.
- BR-6 — not-addressed — ARCH-ORDER: cmd/sdlc/trunkpublish.go:313–317 and 332–336 still separate the final backup-byte check from unlink. A disposable overlay added only an ordering hook immediately before unlink; writing through a descriptor opened before rename made both tracked and untracked cleanup return nil with no backup, while the later edit disappeared. Exclusive creation also does not exclude subsequent writers during replacement.Write at line 327. Preserve recoverable data throughout cleanup and test all destructive-effect boundaries, rather than only the after-move event.
- BR-7 — addressed — README.md:28–37 adds command usage, eligibility, moved-main refusal, uncertain-push retry guidance, and an atlas link. These passages match the allowlist and publication checks; atlas/workflow/sdlc-binary.md also documents the new surface.

## Open findings

- **BR-1** [Important] `partial-progress-recovery` Define recovery after publication succeeds but only part of cleanup completes
- **BR-6** [Critical] `cleanup-preserves-concurrent-work` (kind: defect) Cleanup can overwrite edits made after its validation
