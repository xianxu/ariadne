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
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 9049b272b979f464c7f0afca39edfb576526aeaf
        tracker_head: 84921a6d968f98ca0d8571e544d08876b14af6f0
        branch_point: 4c52f8a4e712c01c3d764bf53340a5289940fc7b
        since: "2026-10-14T08:53:15-07:00"
        until: "2026-10-14T17:52:06-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T08:53:15-07:00"
              end: "2026-10-14T11:37:28-07:00"
              began: "2026-10-14T08:53:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
            - start: "2026-10-14T11:37:28-07:00"
              end: "2026-10-14T17:52:06-07:00"
              began: "2026-10-14T08:53:15-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a136db-6cd1-7a31-87f3-271411613545
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: ab049e7f9b9b2fd4cd39d559a2049dac956af547b6426efae172d5cb94eafd43
              started: "2026-10-13T20:02:12.897Z"
              spawned: false
              bytes: 19563577
              windows:
                - start: "2026-10-14T11:37:28-07:00"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 250a43b6f36d74011493b050d52ce7d14222ff4f6d03c5984a80d98d56cdf350
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 23360225
              windows:
                - start: "2026-10-14T08:53:15-07:00"
                  end: "2026-10-14T11:37:28-07:00"
                - start: "2026-10-14T11:37:28-07:00"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a13b21-7327-7c60-9006-edb3c90098ae
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c3162bf2b1730739eeab71cce50f8496d7da0898d39c3ef94f0cb469e5e0af89
              started: "2026-10-14T15:56:50.688Z"
              spawned: false
              bytes: 527136
              windows:
                - start: "2026-10-14T15:56:50.688Z"
                  end: "2026-10-14T11:37:28-07:00"
                - start: "2026-10-14T11:37:28-07:00"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a13bb6-8ece-7533-bf70-cb6022b40029
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: a7243862f79aa4e17102cda8f735c70c42cc802984be5c4e20fabce5e807085a
              started: "2026-10-14T18:39:42.497Z"
              spawned: false
              bytes: 576557
              windows:
                - start: "2026-10-14T18:39:42.497Z"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a13bb9-c9d4-74e0-8386-4967333fab42
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 39e12181e61983a2b02636e7dd5640454d664cf5e614e9d45901d17b0ef9cb16
              started: "2026-10-14T18:43:14.309Z"
              spawned: false
              bytes: 436808
              windows:
                - start: "2026-10-14T18:43:14.309Z"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a13bbc-0eea-7293-a1da-e6b517ba3bc5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b5b9167849378e1cfc8f4649223b2400a4e745d62160cde7230cbd5aa87c5ceb
              started: "2026-10-14T18:45:43.129Z"
              spawned: false
              bytes: 454631
              windows:
                - start: "2026-10-14T18:45:43.129Z"
                  end: "2026-10-14T17:52:06-07:00"
            - harness: codex
              id: 01a13bbc-c0df-7e53-9383-6398cd4255a9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 7f11c95c3b9871616ac14dd392bb0200ef5cfbe4f1ced84e182c4f1a79694b0b
              started: "2026-10-14T18:46:28.655Z"
              spawned: false
              bytes: 368889
              windows:
                - start: "2026-10-14T18:46:28.655Z"
                  end: "2026-10-14T17:52:06-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 45.166466666666665
        cuts:
            - dispatched: "2026-10-14T09:32:19-07:00"
              finalized: "2026-10-14T09:36:09-07:00"
              finalizing: false
            - dispatched: "2026-10-14T10:19:16-07:00"
              finalized: "2026-10-14T10:23:27-07:00"
              finalizing: false
      "n": 4
      timestamp: "2026-10-14T17:55:25-07:00"
      dispatched: "2026-10-14T17:52:33-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Automatic cleanup and backup state are removed. TestTrunkPublishRetryAfterPushInterruption verifies retained bytes and no second main commit; restoring the pre-retention cleanup makes this test fail.
          round: 4
        - id: BR-6
          disposition: addressed
          note: Publication performs no local restore, rename, or removal. TestTrunkPublishPreservesOpenDescriptorWriteAfterPush passes and fails with the pre-retention implementation restored in a disposable checkout.
          round: 4
        - id: BR-2
          disposition: addressed
          note: Plan Tasks 1–2 identify function-level strategies for eligibility, presence-aware decisions, publication, interruption, and local retention; corresponding tests exist.
          round: 4
        - id: BR-3
          disposition: addressed
          note: The plan has three implementation tasks and one subsequent issue-close boundary; the redundant M1 boundary is absent.
          round: 4
        - id: BR-4
          disposition: addressed
          note: Plan lines 41–45 explicitly declare the 1 MiB input envelope, two-minute remote budget, and three-attempt CAS budget. Enforcement of the input envelope has a separate finding below.
          round: 4
        - id: BR-5
          disposition: addressed
          note: Selected-path staging checks and cleanup guidance use literal Git pathspecs. Tracked and untracked wildcard-filename tests pass and verify preservation of neighboring files.
          round: 4
        - id: BR-7
          disposition: addressed
          note: README.md now documents trunk publish --path, its allowlist, conflict refusal, retry, local retention, and tracked/untracked cleanup guidance, consistent with the command.
          round: 4
      findings:
        - kind: defect
          id: BR-8
          severity: Important
          title: Enforce the prose limit during reading, not only before and after it
          detail: 'cmd/sdlc/trunkpublish.go:105–117 checks Lstat size, then calls unbounded os.ReadFile. A file growing between those operations is read fully before rejection, contrary to the plan''s promise to refuse oversized input before reading fully. Deterministic instrumentation in a disposable pinned-head checkout grew the file after Lstat: the reader consumed 4,194,304 bytes against a 1,048,576-byte limit. ARCH-CONSTRAINTS: this is the 2nd finding in family explicit-operating-envelope. Earlier rounds fixed declaration of the bounds; state and enforce the rule across both initial and post-publication reads: consume at most limit+1 bytes, then reject overflow. Add regression coverage for static oversized input and growth after validation.'
          family: explicit-operating-envelope
          round: 4
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

## Round 4 — 2026-10-14T17:55:25-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — Automatic cleanup and backup state are removed. TestTrunkPublishRetryAfterPushInterruption verifies retained bytes and no second main commit; restoring the pre-retention cleanup makes this test fail.
- BR-6 — addressed — Publication performs no local restore, rename, or removal. TestTrunkPublishPreservesOpenDescriptorWriteAfterPush passes and fails with the pre-retention implementation restored in a disposable checkout.
- BR-2 — addressed — Plan Tasks 1–2 identify function-level strategies for eligibility, presence-aware decisions, publication, interruption, and local retention; corresponding tests exist.
- BR-3 — addressed — The plan has three implementation tasks and one subsequent issue-close boundary; the redundant M1 boundary is absent.
- BR-4 — addressed — Plan lines 41–45 explicitly declare the 1 MiB input envelope, two-minute remote budget, and three-attempt CAS budget. Enforcement of the input envelope has a separate finding below.
- BR-5 — addressed — Selected-path staging checks and cleanup guidance use literal Git pathspecs. Tracked and untracked wildcard-filename tests pass and verify preservation of neighboring files.
- BR-7 — addressed — README.md now documents trunk publish --path, its allowlist, conflict refusal, retry, local retention, and tracked/untracked cleanup guidance, consistent with the command.

### Raised

- **BR-8** [Important] `explicit-operating-envelope` (kind: defect) Enforce the prose limit during reading, not only before and after it
  cmd/sdlc/trunkpublish.go:105–117 checks Lstat size, then calls unbounded os.ReadFile. A file growing between those operations is read fully before rejection, contrary to the plan's promise to refuse oversized input before reading fully. Deterministic instrumentation in a disposable pinned-head checkout grew the file after Lstat: the reader consumed 4,194,304 bytes against a 1,048,576-byte limit. ARCH-CONSTRAINTS: this is the 2nd finding in family explicit-operating-envelope. Earlier rounds fixed declaration of the bounds; state and enforce the rule across both initial and post-publication reads: consume at most limit+1 bytes, then reject overflow. Add regression coverage for static oversized input and growth after validation.

## Open findings

- **BR-8** [Important] `explicit-operating-envelope` (kind: defect) Enforce the prose limit during reading, not only before and after it
