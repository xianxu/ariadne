---
gate: boundary-review
issue: 383
id_prefix: BR
rounds:
    - actual_inputs:
        version: 1
        method: active-time-v3-engagement-v1
        repository: github.com/xianxu/ariadne
        code_head: 18ce936d5a3e35dbec0885ef61c9fc80fd5be496
        tracker_head: 2eac65b33290466457f3d0c8d17192c7dab2b1fe
        branch_point: ef5e0bf028c33b130d93b7014be47d310812fc15
        since: "2026-10-14T22:12:14-07:00"
        until: "2026-10-14T22:28:13-07:00"
        brain_root: /Users/xianxu/workspace/brain
        gap_minutes: 15
        commit_weight: 1
        include_assistant: true
        engagements:
            - start: "2026-10-14T22:12:14-07:00"
              end: "2026-10-14T22:28:13-07:00"
              began: "2026-10-14T22:12:14-07:00"
              roots:
                - /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
        sources:
            - harness: claude
              id: edb1a6a7-e586-4b70-a8a5-085ea3105e3b
              cwd: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
              digest: 5c4a7b4dc5b376978c5c9bbbe4a1b50e2f20af58f8e2a0de2a3409c1de971d61
              started: "2026-10-14T00:19:33.465Z"
              spawned: false
              bytes: 7202781
              windows:
                - start: "2026-10-14T22:12:14-07:00"
                  end: "2026-10-14T22:28:13-07:00"
      "n": 1
      timestamp: "2026-10-14T23:21:10-07:00"
      dispatched: "2026-10-14T23:17:34-07:00"
      agent: claude
      findings:
        - kind: defect
          id: BR-1
          severity: Minor
          title: The "slot is free while merge polls and merges" clause is measured only at runLocalCheck's PostStatus
          detail: TestSuiteSlotCoversOnlyTheRun probes inside PostStatus, not inside sdlc merge's CI poll or PR merge. It holds by structure because acquireSuiteSlot has a single caller, localcheck_run.go:81, and the slot is freed before that function returns. This is the only instance in the window.
          family: done-when-measured-at-proxy
          round: 1
        - kind: defect
          id: BR-2
          severity: Minor
          title: Docs say the slot is freed "before the status post"; on Linux the watcher frees it up to about 0.5s later
          detail: 'test-lock.py wait_for_exit uses kqueue on macOS but checks kill(pid,0) every 0.5s elsewhere. Instances: atlas/workflow/sdlc-binary.md:1743 and the comment at cmd/sdlc/localcheck_run.go:112. Consider saying the release is started before the status post.'
          family: doc-claim-platform-timing
          round: 1
      recipe: small-diff-review
      isolation: fenced
      reviewed: 18ce936d5a3e35dbec0885ef61c9fc80fd5be496
      blocked: false
---

# Gate ledger — ariadne#383 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-14T23:21:10-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `done-when-measured-at-proxy` (kind: defect) The "slot is free while merge polls and merges" clause is measured only at runLocalCheck's PostStatus
  TestSuiteSlotCoversOnlyTheRun probes inside PostStatus, not inside sdlc merge's CI poll or PR merge. It holds by structure because acquireSuiteSlot has a single caller, localcheck_run.go:81, and the slot is freed before that function returns. This is the only instance in the window.
- **BR-2** [Minor] `doc-claim-platform-timing` (kind: defect) Docs say the slot is freed "before the status post"; on Linux the watcher frees it up to about 0.5s later
  test-lock.py wait_for_exit uses kqueue on macOS but checks kill(pid,0) every 0.5s elsewhere. Instances: atlas/workflow/sdlc-binary.md:1743 and the comment at cmd/sdlc/localcheck_run.go:112. Consider saying the release is started before the status post.

## Open findings

- **BR-1** [Minor] `done-when-measured-at-proxy` (kind: defect) The "slot is free while merge polls and merges" clause is measured only at runLocalCheck's PostStatus
- **BR-2** [Minor] `doc-claim-platform-timing` (kind: defect) Docs say the slot is freed "before the status post"; on Linux the watcher frees it up to about 0.5s later
