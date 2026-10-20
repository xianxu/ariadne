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
---

# Gate ledger — ariadne#477 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-19T18:30:16-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Important] `artifact-creation-cleanup-parity` (kind: defect) Repaired checkout roots are absent from stale-worktree cleanup
  cmd/sdlc/internal/scratchdisk/scratchdisk.go:133-137 now creates detached checkouts under NormalTempDir's repaired root, but cmd/sdlc/localcheck_run.go:562-575 still sweeps only scratch and os.TempDir(). With unusable inherited TMPDIR, a killed run leaves an sdlc-local-* checkout under Darwin's per-user temp or /tmp that subsequent runs never collect; prepareScratch's sweep excludes that prefix, and git worktree prune does not delete existing checkout directories. ARCH-FUNERAL: include repaired normal roots in checkout cleanup enumeration and add a regression covering stale and fresh checkouts there.

## Open findings

- **BR-1** [Important] `artifact-creation-cleanup-parity` (kind: defect) Repaired checkout roots are absent from stale-worktree cleanup
