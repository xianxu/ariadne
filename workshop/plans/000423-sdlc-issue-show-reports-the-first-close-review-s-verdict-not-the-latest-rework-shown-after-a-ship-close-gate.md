---
gate: boundary-review
issue: 423
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: e2b7f2f1-a74d-4196-937d-412659227373
            repository: github.com/xianxu/ariadne
            issue_id: "000423"
            recipe: small-diff-review
            round: 1
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/8975df82cc57282c
            reviewed_head: 4bea66dbd7c3275d35560124f274834b1505e16b
            slot: ariadne:5
            requested_at: "2026-10-17T11:23:17.606778Z"
          phase: finished
          pid: 88350
          started_at: "2026-10-17T11:23:17.763941Z"
          finished_at: "2026-10-17T11:24:25.184821Z"
          session_id: 01a1499a-1936-71b1-bd81-4532363ae12c
          outcome: returned
      "n": 1
      timestamp: "2026-10-17T04:24:26-07:00"
      dispatched: "2026-10-17T04:23:17-07:00"
      agent: codex
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Stale cards without completion bindings suppress the latest review
          detail: cmd/sdlc/internal/observe/checkpoints.go:186 uses valued(comp.Read) as evidence that a completion binding exists. assembleCompletion preserves State=stale when a stale card has no binding (assemble.go:198), leaving ReviewedHead empty. A working issue with a valid latest SHIP consequently becomes unknown with an empty verdict and error "no review entry matches the completion reviewed head ". The complete family in this window is this selection guard and its shared text/JSON consumers. Check actual binding presence separately from read quality; preserve latest-review selection when absent and completion-bound selection when present. Add regression coverage for both stale states.
          family: binding-presence-independent-of-read-quality
          round: 1
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 232ab55e-a8d6-4bd9-935d-82c57640c5e0
            repository: github.com/xianxu/ariadne
            issue_id: "000423"
            recipe: small-diff-review
            round: 2
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/31c8cc98ca1f4ce2
            reviewed_head: ea1a9ee9a2c0855563cc269aac6dd3cd8e216565
            slot: ariadne:5
            requested_at: "2026-10-17T12:28:08.51601Z"
          phase: finished
          pid: 62630
          started_at: "2026-10-17T12:28:08.613898Z"
          finished_at: "2026-10-17T12:29:35.460322Z"
          session_id: 01a149d5-7708-7312-96c4-6374b766dbd6
          outcome: returned
      "n": 2
      timestamp: "2026-10-17T05:29:36-07:00"
      dispatched: "2026-10-17T05:28:08-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: The guard now requires comp.Token alongside valued read quality. The four fresh/stale × binding-present/absent cases pass. Removing the token guard in a disposable pinned-head copy makes TestReviewRoundsSelectCompletionOrLatest/stale_absent fail with the original unknown-verdict error.
          round: 2
      findings:
        - kind: defect
          id: BR-2
          severity: Critical
          title: Timestamp drift corrupts review round numbers and history labels
          detail: 'cmd/sdlc/internal/observe/checkpoints.go:238-255 joins entries to ledger rounds by exact timestamp, silently substituting entry order on a miss. Sidecar persistence samples nowRFC3339 separately from ledger persistence (reviewsidecar.go:211 and close.go:1526); the pinned #423 sidecar records 04:24:25 while its ledger records 04:24:26. A disposable regression probe with ledger rounds 4 and 6 and a one-second mismatch on the completion-bound second entry reports round 2 and labels the earlier round 4 “later, not completion-bound.” The complete family is the collector’s timestamp map (observe.go:318-322), both historical/current number lookups, and the numeric relation comparison (checkpoints.go:197), affecting close and milestone reviews and their shared text/JSON consumers. Use stable review identity with compatibility handling for existing artifacts; derive chronological labels from entry order. Add coverage for timestamp drift where ledger numbers differ from entry ordinals.'
          family: review-provenance-requires-stable-identity
          round: 2
      recipe: small-diff-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: cb4c4b0b-c096-4c49-a464-23b4c1d13c76
            repository: github.com/xianxu/ariadne
            issue_id: "000423"
            recipe: milestone-review
            round: 3
            attempt: 1
            agent: codex
            root: /Users/xianxu/Library/Caches/sdlc/review-checkouts/b85b9c4da9ab46d5
            reviewed_head: 8838469f4f7e1a03761b118b16e19f7e3a5fa314
            slot: ariadne:5
            requested_at: "2026-10-17T13:10:50.396948Z"
          phase: finished
          pid: 95110
          started_at: "2026-10-17T13:10:50.572298Z"
          finished_at: "2026-10-17T13:13:34.897437Z"
          session_id: 01a149fc-8f1e-75e3-93ad-bb536ba6e7e9
          outcome: returned
      "n": 3
      timestamp: "2026-10-17T06:13:35-07:00"
      dispatched: "2026-10-17T06:10:50-07:00"
      agent: codex
      dispose:
        - id: BR-1
          disposition: addressed
          note: Confirmed the binding-presence guard and passing fresh/stale × present/absent regression matrix.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Reviewed-commit matching replaces timestamp identity across collection and assembly; history labels use entry order. The timestamp-drift regression passes at head and fails with assignment disabled. The snapshot regression passes at head and fails with mutable-ref reads. Both disposable checkouts were removed.
          round: 3
      recipe: milestone-review
      isolation: fenced
      reviewed: 8838469f4f7e1a03761b118b16e19f7e3a5fa314
      blocked: false
---

# Gate ledger — ariadne#423 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T04:24:26-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `binding-presence-independent-of-read-quality` (kind: defect) Stale cards without completion bindings suppress the latest review
  cmd/sdlc/internal/observe/checkpoints.go:186 uses valued(comp.Read) as evidence that a completion binding exists. assembleCompletion preserves State=stale when a stale card has no binding (assemble.go:198), leaving ReviewedHead empty. A working issue with a valid latest SHIP consequently becomes unknown with an empty verdict and error "no review entry matches the completion reviewed head ". The complete family in this window is this selection guard and its shared text/JSON consumers. Check actual binding presence separately from read quality; preserve latest-review selection when absent and completion-bound selection when present. Add regression coverage for both stale states.

## Round 2 — 2026-10-17T05:29:36-07:00 (codex) — BLOCKED

### Disposed

- BR-1 — addressed — The guard now requires comp.Token alongside valued read quality. The four fresh/stale × binding-present/absent cases pass. Removing the token guard in a disposable pinned-head copy makes TestReviewRoundsSelectCompletionOrLatest/stale_absent fail with the original unknown-verdict error.

### Raised

- **BR-2** [Critical] `review-provenance-requires-stable-identity` (kind: defect) Timestamp drift corrupts review round numbers and history labels
  cmd/sdlc/internal/observe/checkpoints.go:238-255 joins entries to ledger rounds by exact timestamp, silently substituting entry order on a miss. Sidecar persistence samples nowRFC3339 separately from ledger persistence (reviewsidecar.go:211 and close.go:1526); the pinned #423 sidecar records 04:24:25 while its ledger records 04:24:26. A disposable regression probe with ledger rounds 4 and 6 and a one-second mismatch on the completion-bound second entry reports round 2 and labels the earlier round 4 “later, not completion-bound.” The complete family is the collector’s timestamp map (observe.go:318-322), both historical/current number lookups, and the numeric relation comparison (checkpoints.go:197), affecting close and milestone reviews and their shared text/JSON consumers. Use stable review identity with compatibility handling for existing artifacts; derive chronological labels from entry order. Add coverage for timestamp drift where ledger numbers differ from entry ordinals.

## Round 3 — 2026-10-17T06:13:35-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — Confirmed the binding-presence guard and passing fresh/stale × present/absent regression matrix.
- BR-2 — addressed — Reviewed-commit matching replaces timestamp identity across collection and assembly; history labels use entry order. The timestamp-drift regression passes at head and fails with assignment disabled. The snapshot regression passes at head and fails with mutable-ref reads. Both disposable checkouts were removed.

## Open findings

(none — every finding has been disposed)
