---
gate: boundary-review
issue: 411
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 120f3049-2005-48f1-b59e-d112ebd3d7a8
            repository: github.com/xianxu/ariadne
            issue_id: "000411"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/4a02288cdc53df8b
            reviewed_head: 09e1566a82c57068474e738f704a74c849e71b02
            slot: ariadne:5
            requested_at: "2026-10-16T23:20:10.563667Z"
          phase: finished
          pid: 99638
          started_at: "2026-10-16T23:20:10.665585Z"
          finished_at: "2026-10-16T23:21:38.118439Z"
          session_id: 01a14704-0fdf-7c00-97ba-2be1c2a6cb2f
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 09e1566a82c57068474e738f704a74c849e71b02
        tracker_head: 027722769dcba7a6b0af330c3329ca43b50130e3
        branch_point: 094cd4294c2d7a0a8102e8b291c638a8eb5d45d6
        since: "2026-10-16T15:06:50-07:00"
        until: "2026-10-16T16:19:51-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-16T15:06:50-07:00"
              end: "2026-10-16T16:19:51-07:00"
              began: "2026-10-16T15:06:50-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
        sources:
            - harness: codex
              id: 01a136db-6cd1-7a31-87f3-271411613545
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 211e95c362eb407097ec7582cea809d243e0e0c85a54384657f89134d4041c00
              started: "2026-10-13T20:02:12.897Z"
              spawned: false
              bytes: 41025315
              windows:
                - start: "2026-10-16T15:06:50-07:00"
                  end: "2026-10-16T16:19:51-07:00"
            - harness: codex
              id: 01a146c1-89b0-70a3-a6f7-d36195c1e712
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              parent_id: 01a136db-6cd1-7a31-87f3-271411613545
              digest: 1068623f9680867c2a64cb9bf0b5c185f3ceec2d701e8b0674e4a3d6ee614220
              started: "2026-10-16T22:07:31.442Z"
              spawned: true
              bytes: 823054
              windows:
                - start: "2026-10-16T22:07:31.442Z"
                  end: "2026-10-16T16:19:51-07:00"
            - harness: codex
              id: 01a146c6-7afd-7f20-953b-a22feb894466
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: e7bae74aa107bed4024e9aa869437fe1da2de733d62d9dfa635fe79367db2935
              started: "2026-10-16T22:12:55.715Z"
              spawned: false
              bytes: 603643
              windows:
                - start: "2026-10-16T22:12:55.715Z"
                  end: "2026-10-16T16:19:51-07:00"
            - harness: codex
              id: 01a146c7-9d2c-7ad2-8f7a-84a895b8235d
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 5a6b7a328de16be8f18d9f3a1971722cded091139f006f4f1f0a283bf4c893a3
              started: "2026-10-16T22:14:10.262Z"
              spawned: false
              bytes: 447253
              windows:
                - start: "2026-10-16T22:14:10.262Z"
                  end: "2026-10-16T16:19:51-07:00"
            - harness: codex
              id: 01a146c8-cac0-7591-8d7e-d5e067bc1582
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 308c1c0d0a2d2e5936a5304e666250be74e55e605808f5269dea172da2148fb6
              started: "2026-10-16T22:15:27.659Z"
              spawned: false
              bytes: 466887
              windows:
                - start: "2026-10-16T22:15:27.659Z"
                  end: "2026-10-16T16:19:51-07:00"
            - harness: codex
              id: 01a146c9-2fd7-76f1-80fa-e78d01c5de6a
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
              digest: 13dc112a246fe90f4326d725c9f4526f1c6008c501e8b6efd640cbc7a6f62c9e
              started: "2026-10-16T22:15:53.422Z"
              spawned: false
              bytes: 412608
              windows:
                - start: "2026-10-16T22:15:53.422Z"
                  end: "2026-10-16T16:19:51-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 10.169466666666667
      "n": 1
      timestamp: "2026-10-16T16:21:40-07:00"
      dispatched: "2026-10-16T16:20:10-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Gate deferral misses retirement of previously managed settings
          detail: 'cmd/weave/settings.go:20-39 creates DeferredOutput only for an existing MergeSettings action. After compiling settings and removing its manifest declaration, I reproduced weave settings exiting silently and refresh --substrate --defer-settings deleting .claude/settings.json (output: "retired .claude/settings.json"). Under write denial, this retirement can fail refresh again. ARCH-PURPOSE/ARCH-ORDER: protect the settings destination throughout reconciliation, including retirement and other action types; add a regression covering declaration removal, retained inventory/backups, denied mutations, and eventual operator cleanup.'
          family: protected-output-exclusion
          round: 1
        - kind: defect
          id: BR-2
          severity: Critical
          title: The Core concepts table labels filesystem ownership reconciliation PURE
          detail: workshop/plans/000411-protected-settings-plan.md:16 lists Managed output ownership under Pure entity, but ownership.go:200-208 performs publication reclamation and commits filesystem changes. Its regression at deferred_test.go:78-98 uses temporary files and deferredDenyFS. This contradicts the required Core concepts contract and ARCH-PURE. Move ownership reconciliation to INTEGRATION, identify its injected weavefs seam, and append a Revisions entry; retain only genuinely deterministic entities in the PURE table.
          family: core-concept-purity-classification
          round: 1
        - kind: unverifiable
          id: BR-3
          severity: Important
          title: Native protected-settings conformance could not execute at the pinned head
          detail: 'ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run ''^TestStartPlanProtectedSettingsNativeSandbox$'' -count=1 -v -timeout 8m failed in both subtests because sandbox-exec exited 71 with "sandbox_apply: Operation not permitted", before launching the child. This prevents verification of the real OS-denial invariant (ARCH-MOCK). Supply a successful run on a host permitting sandbox-exec at the reviewed head, including denied-write controls and stale/current notice assertions, or enable that capability in the review environment.'
          family: native-boundary-conformance
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
disputes:
    - id: BR-3
      rationale: The reviewer cannot nest sandbox-exec inside its fence; host conformance is now measured on the revised source head, including retirement. This is test-only native fencing, not an unsandboxed production write retry.
      evidence: 'At aa326d59db8cf05a962daa7dc9f5e40764305f27: ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run Test(StartPlan|RunInstalledWeave|GateSettings) -count=1 -v -timeout 10m passed 122.390s. Native subtest passed 56.41s: stale/current/retired all confirm denied write, real deferred refresh and first start-plan success, notice counts 1/0/1; retired settings, displaced backup and ownership inventory unchanged. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-revised-native-9iww01mg.log records source head and controls.'
      after: 1
      timestamp: "2026-10-16T23:34:46Z"
      by: Xian Xu on Xian’s MacBook Pro at ariadne:5 (/Users/xianxu/workspace/worktree/ariadne-slot5/ariadne)
---

# Gate ledger — ariadne#411 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-16T16:21:40-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `protected-output-exclusion` (kind: defect) Gate deferral misses retirement of previously managed settings
  cmd/weave/settings.go:20-39 creates DeferredOutput only for an existing MergeSettings action. After compiling settings and removing its manifest declaration, I reproduced weave settings exiting silently and refresh --substrate --defer-settings deleting .claude/settings.json (output: "retired .claude/settings.json"). Under write denial, this retirement can fail refresh again. ARCH-PURPOSE/ARCH-ORDER: protect the settings destination throughout reconciliation, including retirement and other action types; add a regression covering declaration removal, retained inventory/backups, denied mutations, and eventual operator cleanup.
- **BR-2** [Critical] `core-concept-purity-classification` (kind: defect) The Core concepts table labels filesystem ownership reconciliation PURE
  workshop/plans/000411-protected-settings-plan.md:16 lists Managed output ownership under Pure entity, but ownership.go:200-208 performs publication reclamation and commits filesystem changes. Its regression at deferred_test.go:78-98 uses temporary files and deferredDenyFS. This contradicts the required Core concepts contract and ARCH-PURE. Move ownership reconciliation to INTEGRATION, identify its injected weavefs seam, and append a Revisions entry; retain only genuinely deterministic entities in the PURE table.
- **BR-3** [Important] `native-boundary-conformance` (kind: unverifiable) Native protected-settings conformance could not execute at the pinned head
  ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run '^TestStartPlanProtectedSettingsNativeSandbox$' -count=1 -v -timeout 8m failed in both subtests because sandbox-exec exited 71 with "sandbox_apply: Operation not permitted", before launching the child. This prevents verification of the real OS-denial invariant (ARCH-MOCK). Supply a successful run on a host permitting sandbox-exec at the reviewed head, including denied-write controls and stale/current notice assertions, or enable that capability in the review environment.

## Disputes

Findings the implementer declined (`sdlc finding dispute`), each answered by the
next review round.

- **BR-3** — after round 1 (2026-10-16T23:34:46Z, Xian Xu on Xian’s MacBook Pro at ariadne:5 (/Users/xianxu/workspace/worktree/ariadne-slot5/ariadne))
  rationale: The reviewer cannot nest sandbox-exec inside its fence; host conformance is now measured on the revised source head, including retirement. This is test-only native fencing, not an unsandboxed production write retry.
  evidence: At aa326d59db8cf05a962daa7dc9f5e40764305f27: ARIADNE_NATIVE_SANDBOX=1 go test -trimpath ./cmd/sdlc -run Test(StartPlan|RunInstalledWeave|GateSettings) -count=1 -v -timeout 10m passed 122.390s. Native subtest passed 56.41s: stale/current/retired all confirm denied write, real deferred refresh and first start-plan success, notice counts 1/0/1; retired settings, displaced backup and ownership inventory unchanged. Log /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne411-revised-native-9iww01mg.log records source head and controls.

## Open findings

- **BR-1** [Critical] `protected-output-exclusion` (kind: defect) Gate deferral misses retirement of previously managed settings
- **BR-2** [Critical] `core-concept-purity-classification` (kind: defect) The Core concepts table labels filesystem ownership reconciliation PURE
