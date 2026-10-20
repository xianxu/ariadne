---
gate: boundary-review
issue: 477
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 188f3c30-106d-44ad-9220-653f84d2defd
            repository: github.com/xianxu/ariadne
            issue_id: "000477"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/5adbcedf7fb8b059
            reviewed_head: fe4ac563928f22a4a5daa0fde6c3ebf567fc5bfa
            slot: ariadne:3
            requested_at: "2026-10-20T01:24:55.00195Z"
          phase: finished
          pid: 43549
          started_at: "2026-10-20T01:24:55.091548Z"
          finished_at: "2026-10-20T01:30:15.906133Z"
          session_id: 01a156e9-57aa-71c0-a8ee-08f197de2d92
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: fe4ac563928f22a4a5daa0fde6c3ebf567fc5bfa
        tracker_head: 8305da00350c3b53ac2ed6731e4e0cf5099f5df4
        branch_point: 4afeb37dae25f204eb6df98b337dcccfdea92baf
        since: "2026-10-19T17:59:41-07:00"
        until: "2026-10-19T18:24:40-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T17:59:41-07:00"
              end: "2026-10-19T18:24:40-07:00"
              began: "2026-10-19T17:59:41-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 52e4d0ba1707bd252ce367925cc088ccf62da79f64716e34ebf8487df306e4e6
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 55065892
              windows:
                - start: "2026-10-19T17:59:41-07:00"
                  end: "2026-10-19T18:24:40-07:00"
            - harness: codex
              id: 01a156d2-b360-7ef3-9529-95d4464eb428
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 14e8912110884204b5966e4299a9727a7506a73c7dd1075dbaa9f224029045cd
              started: "2026-10-20T01:00:11.574Z"
              spawned: true
              bytes: 604678
              windows:
                - start: "2026-10-20T01:00:11.574Z"
                  end: "2026-10-19T18:24:40-07:00"
            - harness: codex
              id: 01a156d8-63c9-7390-8df0-21f754df95c2
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 50a9e6b2320ea9b8322f8f994fd77ebba8bf34a8f10dd43da9ffd2e9f6f2a7b9
              started: "2026-10-20T01:06:24.509Z"
              spawned: true
              bytes: 556520
              windows:
                - start: "2026-10-20T01:06:24.509Z"
                  end: "2026-10-19T18:24:40-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 5.523383333333333
      "n": 1
      timestamp: "2026-10-19T18:30:16-07:00"
      dispatched: "2026-10-19T18:24:55-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Important
          title: Repaired checkout roots are absent from stale-worktree cleanup
          detail: 'cmd/sdlc/internal/scratchdisk/scratchdisk.go:133-137 now creates detached checkouts under NormalTempDir''s repaired root, but cmd/sdlc/localcheck_run.go:562-575 still sweeps only scratch and os.TempDir(). With unusable inherited TMPDIR, a killed run leaves an sdlc-local-* checkout under Darwin''s per-user temp or /tmp that subsequent runs never collect; prepareScratch''s sweep excludes that prefix, and git worktree prune does not delete existing checkout directories. ARCH-FUNERAL: include repaired normal roots in checkout cleanup enumeration and add a regression covering stale and fresh checkouts there.'
          family: artifact-creation-cleanup-parity
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: b85fd7f9-b491-4cf2-a5c2-bf6beb49c7e3
            repository: github.com/xianxu/ariadne
            issue_id: "000477"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/fde0deef7f352565
            reviewed_head: 28f3780778da5facae19adc6a981912d357b4b45
            slot: ariadne:3
            requested_at: "2026-10-20T01:34:09.392619Z"
          phase: finished
          pid: 40489
          started_at: "2026-10-20T01:34:09.483718Z"
          finished_at: "2026-10-20T01:37:10.863456Z"
          session_id: 01a156f1-cd33-7fa3-84a5-ee47f54c3c65
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 28f3780778da5facae19adc6a981912d357b4b45
        tracker_head: 8305da00350c3b53ac2ed6731e4e0cf5099f5df4
        branch_point: 4afeb37dae25f204eb6df98b337dcccfdea92baf
        since: "2026-10-19T17:59:41-07:00"
        until: "2026-10-19T18:33:39-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T17:59:41-07:00"
              end: "2026-10-19T18:33:39-07:00"
              began: "2026-10-19T17:59:41-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: b107a71d0c9d8474496a069eb0ba5d38690e74682ebd6b3dc50070c74beb9b43
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 55482744
              windows:
                - start: "2026-10-19T17:59:41-07:00"
                  end: "2026-10-19T18:33:39-07:00"
            - harness: codex
              id: 01a156d2-b360-7ef3-9529-95d4464eb428
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 14e8912110884204b5966e4299a9727a7506a73c7dd1075dbaa9f224029045cd
              started: "2026-10-20T01:00:11.574Z"
              spawned: true
              bytes: 604678
              windows:
                - start: "2026-10-20T01:00:11.574Z"
                  end: "2026-10-19T18:33:39-07:00"
            - harness: codex
              id: 01a156d8-63c9-7390-8df0-21f754df95c2
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 50a9e6b2320ea9b8322f8f994fd77ebba8bf34a8f10dd43da9ffd2e9f6f2a7b9
              started: "2026-10-20T01:06:24.509Z"
              spawned: true
              bytes: 556520
              windows:
                - start: "2026-10-20T01:06:24.509Z"
                  end: "2026-10-19T18:33:39-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 5.523383333333333
        cuts:
            - dispatched: "2026-10-19T18:24:55-07:00"
              finalized: "2026-10-19T18:30:16-07:00"
              finalizing: false
      "n": 2
      timestamp: "2026-10-19T18:37:11-07:00"
      dispatched: "2026-10-19T18:34:09-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: scratchdisk.Roots now enumerates repaired normal roots. TestSweepStaleLocalWorktreesRepairedRoot passes at HEAD and fails when only that enumeration fix is removed through a temporary Go source overlay.
          round: 2
      findings:
        - kind: defect
          id: BR-2
          severity: Important
          title: Repaired normal roots still strand abandoned per-run directories
          detail: 'cmd/sdlc/testscratch.go:35,60,70 sweeps only the inherited root, usable scratch root, and currently selected normal root. A killed fallback run leaves sdlc-run-* under Darwin''s per-user temp or /tmp; after inherited TMPDIR recovers, subsequent runs never sweep that former fallback. A private Darwin fixture reproduced this at HEAD: TestReviewRepairedRunRootAfterInheritedRecovery failed because the stale directory survived. ARCH-FUNERAL / ARCH-PURPOSE: this is the 2nd finding in family artifact-creation-cleanup-parity. Earlier rounds fixed instances. Do NOT fix only this instance: state and enforce the rule that every disposable artifact family sweeps every root where creation can place it, independently of current root selection. Enumerate checkout and per-run families, share normal-root enumeration, and test stale removal plus fresh/live preservation across fallback and recovery.'
          family: artifact-creation-cleanup-parity
          round: 2
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 4e7af666-9ba0-4857-9bd5-3f534df34edb
            repository: github.com/xianxu/ariadne
            issue_id: "000477"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/e6b538ad33c89e5f
            reviewed_head: d45e5ce7bd1a79fcd6267586d2a00869f42f75f0
            slot: ariadne:3
            requested_at: "2026-10-20T01:41:14.335959Z"
          phase: finished
          pid: 36374
          started_at: "2026-10-20T01:41:14.551819Z"
          finished_at: "2026-10-20T01:43:19.020166Z"
          session_id: 01a156f8-4ad5-7b10-9160-08780a2596f9
          outcome: returned
      actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: d45e5ce7bd1a79fcd6267586d2a00869f42f75f0
        tracker_head: 8305da00350c3b53ac2ed6731e4e0cf5099f5df4
        branch_point: 4afeb37dae25f204eb6df98b337dcccfdea92baf
        since: "2026-10-19T17:59:41-07:00"
        until: "2026-10-19T18:40:25-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-19T17:59:41-07:00"
              end: "2026-10-19T18:40:25-07:00"
              began: "2026-10-19T17:59:41-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
        sources:
            - harness: codex
              id: 01a14367-9218-7872-89c3-f9a28c524834
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              digest: 48eeaf0847c712e8c4f980891fed5c61f2eb359064c66d9613b8e34bf42daa5f
              started: "2026-10-16T06:30:23.8Z"
              spawned: false
              bytes: 55834465
              windows:
                - start: "2026-10-19T17:59:41-07:00"
                  end: "2026-10-19T18:40:25-07:00"
            - harness: codex
              id: 01a156d2-b360-7ef3-9529-95d4464eb428
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 14e8912110884204b5966e4299a9727a7506a73c7dd1075dbaa9f224029045cd
              started: "2026-10-20T01:00:11.574Z"
              spawned: true
              bytes: 604678
              windows:
                - start: "2026-10-20T01:00:11.574Z"
                  end: "2026-10-19T18:40:25-07:00"
            - harness: codex
              id: 01a156d8-63c9-7390-8df0-21f754df95c2
              slot: ariadne:3
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
              parent_id: 01a14367-9218-7872-89c3-f9a28c524834
              digest: 50a9e6b2320ea9b8322f8f994fd77ebba8bf34a8f10dd43da9ffd2e9f6f2a7b9
              started: "2026-10-20T01:06:24.509Z"
              spawned: true
              bytes: 556520
              windows:
                - start: "2026-10-20T01:06:24.509Z"
                  end: "2026-10-19T18:40:25-07:00"
        excluded_brain_sessions: 1
        excluded_brain_minutes: 5.523383333333333
        cuts:
            - dispatched: "2026-10-19T18:24:55-07:00"
              finalized: "2026-10-19T18:30:16-07:00"
              finalizing: false
            - dispatched: "2026-10-19T18:34:09-07:00"
              finalized: "2026-10-19T18:37:11-07:00"
              finalizing: false
      "n": 3
      timestamp: "2026-10-19T18:43:20-07:00"
      dispatched: "2026-10-19T18:41:14-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: Shared root enumeration remains in checkout cleanup; TestSweepStaleLocalWorktreesRepairedRoot passes for broken and recovered TMPDIR.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Both disposable families consume scratchdisk.Roots independently of current placement. TestPrepareScratchRepairedRunRootAfterRecovery passes at HEAD and fails after restoring pre-fix testscratch.go in a disposable copy, which was removed afterward.
          round: 3
      findings:
        - kind: defect
          id: BR-3
          severity: Minor
          title: Scratch preparation failure leaves the opened run log unclosed
          detail: 'cmd/sdlc/localcheck_run.go:307 returns on prepareScratch failure after os.Create opens logf, bypassing the explicit close later in the function. ARCH-FUNERAL: defer logf.Close immediately after successful creation so this new setup-error path releases the descriptor without waiting for GC or process exit.'
          family: resource-handles-close-on-every-exit
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: d45e5ce7bd1a79fcd6267586d2a00869f42f75f0
      blocked: false
---

# Gate ledger — ariadne#477 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T18:30:16-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `artifact-creation-cleanup-parity` (kind: defect) Repaired checkout roots are absent from stale-worktree cleanup
  cmd/sdlc/internal/scratchdisk/scratchdisk.go:133-137 now creates detached checkouts under NormalTempDir's repaired root, but cmd/sdlc/localcheck_run.go:562-575 still sweeps only scratch and os.TempDir(). With unusable inherited TMPDIR, a killed run leaves an sdlc-local-* checkout under Darwin's per-user temp or /tmp that subsequent runs never collect; prepareScratch's sweep excludes that prefix, and git worktree prune does not delete existing checkout directories. ARCH-FUNERAL: include repaired normal roots in checkout cleanup enumeration and add a regression covering stale and fresh checkouts there.

## Round 2 — 2026-10-19T18:37:11-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — scratchdisk.Roots now enumerates repaired normal roots. TestSweepStaleLocalWorktreesRepairedRoot passes at HEAD and fails when only that enumeration fix is removed through a temporary Go source overlay.

### Raised

- **BR-2** [Important] `artifact-creation-cleanup-parity` (kind: defect) Repaired normal roots still strand abandoned per-run directories
  cmd/sdlc/testscratch.go:35,60,70 sweeps only the inherited root, usable scratch root, and currently selected normal root. A killed fallback run leaves sdlc-run-* under Darwin's per-user temp or /tmp; after inherited TMPDIR recovers, subsequent runs never sweep that former fallback. A private Darwin fixture reproduced this at HEAD: TestReviewRepairedRunRootAfterInheritedRecovery failed because the stale directory survived. ARCH-FUNERAL / ARCH-PURPOSE: this is the 2nd finding in family artifact-creation-cleanup-parity. Earlier rounds fixed instances. Do NOT fix only this instance: state and enforce the rule that every disposable artifact family sweeps every root where creation can place it, independently of current root selection. Enumerate checkout and per-run families, share normal-root enumeration, and test stale removal plus fresh/live preservation across fallback and recovery.

## Round 3 — 2026-10-19T18:43:20-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Shared root enumeration remains in checkout cleanup; TestSweepStaleLocalWorktreesRepairedRoot passes for broken and recovered TMPDIR.
- BR-2 — addressed — Both disposable families consume scratchdisk.Roots independently of current placement. TestPrepareScratchRepairedRunRootAfterRecovery passes at HEAD and fails after restoring pre-fix testscratch.go in a disposable copy, which was removed afterward.

### Raised

- **BR-3** [Minor] `resource-handles-close-on-every-exit` (kind: defect) Scratch preparation failure leaves the opened run log unclosed
  cmd/sdlc/localcheck_run.go:307 returns on prepareScratch failure after os.Create opens logf, bypassing the explicit close later in the function. ARCH-FUNERAL: defer logf.Close immediately after successful creation so this new setup-error path releases the descriptor without waiting for GC or process exit.

## Open findings

- **BR-3** [Minor] `resource-handles-close-on-every-exit` (kind: defect) Scratch preparation failure leaves the opened run log unclosed
