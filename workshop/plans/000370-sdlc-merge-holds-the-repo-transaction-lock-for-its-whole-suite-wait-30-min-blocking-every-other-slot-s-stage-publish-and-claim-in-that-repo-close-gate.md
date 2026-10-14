---
gate: boundary-review
issue: 370
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 7f42a98077dc560b215f99c01c49b22f35bea3fc
        tracker_head: 1b00608b8780f93ecaac2233c930115f90093fe2
        branch_point: 675dccf184b69de76bf1f8228c2b818459794a75
        since: "2026-10-14T11:38:01-07:00"
        until: "2026-10-14T12:19:13-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T11:38:01-07:00"
              end: "2026-10-14T12:19:13-07:00"
              began: "2026-10-14T11:38:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: fc63f154955a0a505650e9c0d8a933d24c5bf7ea34016923a2e3d61934777db3
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 19521789
              windows:
                - start: "2026-10-14T11:38:01-07:00"
                  end: "2026-10-14T12:19:13-07:00"
            - harness: codex
              id: 01a13bb6-8ece-7533-bf70-cb6022b40029
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: a7243862f79aa4e17102cda8f735c70c42cc802984be5c4e20fabce5e807085a
              started: "2026-10-14T18:39:42.497Z"
              spawned: false
              bytes: 576557
              windows:
                - start: "2026-10-14T18:39:42.497Z"
                  end: "2026-10-14T12:19:13-07:00"
            - harness: codex
              id: 01a13bb9-c9d4-74e0-8386-4967333fab42
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 39e12181e61983a2b02636e7dd5640454d664cf5e614e9d45901d17b0ef9cb16
              started: "2026-10-14T18:43:14.309Z"
              spawned: false
              bytes: 436808
              windows:
                - start: "2026-10-14T18:43:14.309Z"
                  end: "2026-10-14T12:19:13-07:00"
            - harness: codex
              id: 01a13bbc-0eea-7293-a1da-e6b517ba3bc5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b5b9167849378e1cfc8f4649223b2400a4e745d62160cde7230cbd5aa87c5ceb
              started: "2026-10-14T18:45:43.129Z"
              spawned: false
              bytes: 454631
              windows:
                - start: "2026-10-14T18:45:43.129Z"
                  end: "2026-10-14T12:19:13-07:00"
            - harness: codex
              id: 01a13bbc-c0df-7e53-9383-6398cd4255a9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 7f11c95c3b9871616ac14dd392bb0200ef5cfbe4f1ced84e182c4f1a79694b0b
              started: "2026-10-14T18:46:28.655Z"
              spawned: false
              bytes: 368889
              windows:
                - start: "2026-10-14T18:46:28.655Z"
                  end: "2026-10-14T12:19:13-07:00"
      "n": 1
      timestamp: "2026-10-14T12:24:05-07:00"
      dispatched: "2026-10-14T12:19:27-07:00"
      agent: codex
      findings:
        - id: BR-1
          severity: Critical
          title: Legacy merge can adopt a newer HEAD after publish validation
          detail: 'cmd/sdlc/merge.go:582 re-reads origin/<branch> after the unlocked publish gate and confirmation windows, then validates and merges against that newer head at lines 608–622. If another actor commits and pushes during that gap, the new commit can pass CI and reach PRMerge without repeating instance/publish validation; the original branchHead is checked only during cleanup at line 663, after the irreversible merge. ARCH-ORDER/ARCH-SECURE: pin one head before validation, preserve it across every unlocked phase, and refuse drift before effects; add a regression that commits and pushes after publish validation and asserts zero merge calls.'
          family: pinned-evidence-across-lock-gaps
          round: 1
        - id: BR-2
          severity: Important
          title: Durable merge omits the planned checkpoint after registration polling
          detail: 'cmd/sdlc/landing.go:444–478 proceeds from integrateLandingMain, including unlocked syncIntegrationPR polling, directly into gates using state.main. The plan explicitly requires a locked exact-head/worktree check and fresh-main observation after polling and before expensive gates (plan lines 69 and 81); main movement during registration therefore starts an already-stale suite and is detected only afterward. ARCH-CONSTRAINTS/ARCH-PURPOSE: add that checkpoint, reintegrate before gates when main moved, and test main advancement during delayed registration.'
          family: pre-gate-freshness
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fb2c0fa0093df3b50bfa0a6631040315c0e91d6c
        tracker_head: 1b00608b8780f93ecaac2233c930115f90093fe2
        branch_point: 675dccf184b69de76bf1f8228c2b818459794a75
        since: "2026-10-14T11:38:01-07:00"
        until: "2026-10-14T12:39:41-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T11:38:01-07:00"
              end: "2026-10-14T12:39:41-07:00"
              began: "2026-10-14T11:38:01-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
        sources:
            - harness: codex
              id: 01a138fb-9249-7fe0-86f2-862b8adc68ef
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: c745b3761cbc2d8872af6cba936043e4faa7fd43f9c2caed173421a27ae1ae46
              started: "2026-10-14T05:57:01.791Z"
              spawned: false
              bytes: 20036184
              windows:
                - start: "2026-10-14T11:38:01-07:00"
                  end: "2026-10-14T12:39:41-07:00"
            - harness: codex
              id: 01a13bb6-8ece-7533-bf70-cb6022b40029
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: a7243862f79aa4e17102cda8f735c70c42cc802984be5c4e20fabce5e807085a
              started: "2026-10-14T18:39:42.497Z"
              spawned: false
              bytes: 576557
              windows:
                - start: "2026-10-14T18:39:42.497Z"
                  end: "2026-10-14T12:39:41-07:00"
            - harness: codex
              id: 01a13bb9-c9d4-74e0-8386-4967333fab42
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 39e12181e61983a2b02636e7dd5640454d664cf5e614e9d45901d17b0ef9cb16
              started: "2026-10-14T18:43:14.309Z"
              spawned: false
              bytes: 436808
              windows:
                - start: "2026-10-14T18:43:14.309Z"
                  end: "2026-10-14T12:39:41-07:00"
            - harness: codex
              id: 01a13bbc-0eea-7293-a1da-e6b517ba3bc5
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: b5b9167849378e1cfc8f4649223b2400a4e745d62160cde7230cbd5aa87c5ceb
              started: "2026-10-14T18:45:43.129Z"
              spawned: false
              bytes: 454631
              windows:
                - start: "2026-10-14T18:45:43.129Z"
                  end: "2026-10-14T12:39:41-07:00"
            - harness: codex
              id: 01a13bbc-c0df-7e53-9383-6398cd4255a9
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
              digest: 7f11c95c3b9871616ac14dd392bb0200ef5cfbe4f1ced84e182c4f1a79694b0b
              started: "2026-10-14T18:46:28.655Z"
              spawned: false
              bytes: 368889
              windows:
                - start: "2026-10-14T18:46:28.655Z"
                  end: "2026-10-14T12:39:41-07:00"
        cuts:
            - dispatched: "2026-10-14T12:19:27-07:00"
              finalized: "2026-10-14T12:24:05-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-14T12:45:30-07:00"
      dispatched: "2026-10-14T12:39:55-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: TestRunMerge_LegacyRefusesHeadMovedAfterPublishGate passes on pinned HEAD. Removing the correction in a disposable checkout produces one PRMerge call before refusal; pinned HEAD produces zero.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestLandingMainMovesDuringRegistrationBeforeFirstSuite passes on pinned HEAD. Removing the landing.go:450 checkpoint produces two suites, including a stale first suite; pinned HEAD runs one suite containing the advanced main.
          round: 2
      findings:
        - kind: defect
          id: BR-3
          severity: Critical
          title: Legacy recovery requires a remote branch that successful merging deletes
          detail: 'cmd/sdlc/merge.go:418–421 unconditionally calls revalidateLegacyMerge with requireRemote=true before discovering an already-merged PR; its fetch at lines 775–776 fails when the remote feature branch was deleted by production PRMerge''s --delete-branch (ghclient.go:199). Extending the existing merged-PR recovery fixture with server-side branch deletion reproduces “fatal: couldn''t find remote ref feature”; making this early checkpoint local-only allows cleanup to finish. ARCH-ORDER/ARCH-MOCK: distinguish pre-merge remote-head requirements from confirmed-landing recovery, retain strict checks before a new merge, and add deletion-aware recovery coverage.'
          family: recovery-preconditions-respect-completed-effects
          round: 2
        - kind: defect
          id: BR-4
          severity: Important
          title: Blocked-suite tests release signals without joining merge workers on failure
          detail: 'cmd/sdlc/landingintegrate_test.go:137,153 and cmd/sdlc/merge_main_test.go:20,32 start merge goroutines, but cleanup only writes a release file. Fatal startup, lock-acquisition, or completion failures bypass the finished-channel receive, allowing fixture cleanup to restore cwd and package-global adapters while runMerge is still executing Git effects. ARCH-ORDER/ARCH-SECURE: give both workers cancellation contexts and cleanup that releases, cancels, and joins them before restoring fixture state; consolidate this lifecycle helper.'
          family: concurrent-fixtures-own-worker-lifetime
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
---

# Gate ledger — ariadne#370 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T12:24:05-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `pinned-evidence-across-lock-gaps` Legacy merge can adopt a newer HEAD after publish validation
  cmd/sdlc/merge.go:582 re-reads origin/<branch> after the unlocked publish gate and confirmation windows, then validates and merges against that newer head at lines 608–622. If another actor commits and pushes during that gap, the new commit can pass CI and reach PRMerge without repeating instance/publish validation; the original branchHead is checked only during cleanup at line 663, after the irreversible merge. ARCH-ORDER/ARCH-SECURE: pin one head before validation, preserve it across every unlocked phase, and refuse drift before effects; add a regression that commits and pushes after publish validation and asserts zero merge calls.
- **BR-2** [Important] `pre-gate-freshness` Durable merge omits the planned checkpoint after registration polling
  cmd/sdlc/landing.go:444–478 proceeds from integrateLandingMain, including unlocked syncIntegrationPR polling, directly into gates using state.main. The plan explicitly requires a locked exact-head/worktree check and fresh-main observation after polling and before expensive gates (plan lines 69 and 81); main movement during registration therefore starts an already-stale suite and is detected only afterward. ARCH-CONSTRAINTS/ARCH-PURPOSE: add that checkpoint, reintegrate before gates when main moved, and test main advancement during delayed registration.

## Round 2 — 2026-10-14T12:45:30-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — TestRunMerge_LegacyRefusesHeadMovedAfterPublishGate passes on pinned HEAD. Removing the correction in a disposable checkout produces one PRMerge call before refusal; pinned HEAD produces zero.
- BR-2 — addressed — TestLandingMainMovesDuringRegistrationBeforeFirstSuite passes on pinned HEAD. Removing the landing.go:450 checkpoint produces two suites, including a stale first suite; pinned HEAD runs one suite containing the advanced main.

### Raised

- **BR-3** [Critical] `recovery-preconditions-respect-completed-effects` (kind: defect) Legacy recovery requires a remote branch that successful merging deletes
  cmd/sdlc/merge.go:418–421 unconditionally calls revalidateLegacyMerge with requireRemote=true before discovering an already-merged PR; its fetch at lines 775–776 fails when the remote feature branch was deleted by production PRMerge's --delete-branch (ghclient.go:199). Extending the existing merged-PR recovery fixture with server-side branch deletion reproduces “fatal: couldn't find remote ref feature”; making this early checkpoint local-only allows cleanup to finish. ARCH-ORDER/ARCH-MOCK: distinguish pre-merge remote-head requirements from confirmed-landing recovery, retain strict checks before a new merge, and add deletion-aware recovery coverage.
- **BR-4** [Important] `concurrent-fixtures-own-worker-lifetime` (kind: defect) Blocked-suite tests release signals without joining merge workers on failure
  cmd/sdlc/landingintegrate_test.go:137,153 and cmd/sdlc/merge_main_test.go:20,32 start merge goroutines, but cleanup only writes a release file. Fatal startup, lock-acquisition, or completion failures bypass the finished-channel receive, allowing fixture cleanup to restore cwd and package-global adapters while runMerge is still executing Git effects. ARCH-ORDER/ARCH-SECURE: give both workers cancellation contexts and cleanup that releases, cancels, and joins them before restoring fixture state; consolidate this lifecycle helper.

## Open findings

- **BR-3** [Critical] `recovery-preconditions-respect-completed-effects` (kind: defect) Legacy recovery requires a remote branch that successful merging deletes
- **BR-4** [Important] `concurrent-fixtures-own-worker-lifetime` (kind: defect) Blocked-suite tests release signals without joining merge workers on failure
