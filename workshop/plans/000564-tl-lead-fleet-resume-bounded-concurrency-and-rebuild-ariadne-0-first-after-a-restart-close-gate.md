---
gate: boundary-review
issue: 564
id_prefix: BR
rounds:
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 32edff20-ecb2-413f-a8bf-33fb26cd501d
            repository: github.com/xianxu/ariadne
            issue_id: "000564"
            recipe: milestone-review
            round: 1
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/5ea2f2f61f5e2a46
            reviewed_head: efe4f364a43c37ce0a5b4b0da7844d7f78873153
            slot: ariadne:3
            requested_at: "2026-10-25T01:40:39.238613Z"
          phase: finished
          pid: 7538
          started_at: "2026-10-25T01:40:39.357156Z"
          finished_at: "2026-10-25T01:43:25.445864Z"
          session_id: 01a170b7-8e14-7f81-9a8a-584169f5ffd0
          outcome: returned
      "n": 1
      timestamp: "2026-10-24T18:43:27-07:00"
      dispatched: "2026-10-24T18:40:39-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      findings:
        - kind: defect
          id: BR-1
          severity: Critical
          title: Unavailable agent evidence can produce a successful worker-wave handoff
          detail: 'cmd/tl-lead/fleet_recovery.go:110 returns on an absent asleep marker before validating agent evidence quality; recoveryWireAgent also discards error diagnostics. With an owned worker containing agent: {"state":"unavailable","error":"thread store unreadable"}, the pinned CLI prints "skipped", exits 0, and reports worker waves complete. This contradicts the plan''s requirement that unavailable evidence remain unknown and block completion (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE). Validate agent evidence before treating marker absence as an ordinary skip, retain relevant diagnostics, and add decoder and CLI regressions requiring a named unknown row and nonzero incomplete handoff.'
          family: recovery-evidence-integrity
          round: 1
      recipe: milestone-review
      isolation: fenced
      blocked: true
    - reviewer_runs:
        - version: 1
          intent:
            disposable_root: true
            id: 7e28675a-9c6c-49a0-9da5-e8fc86a14168
            repository: github.com/xianxu/ariadne
            issue_id: "000564"
            recipe: milestone-review
            round: 2
            attempt: 1
            agent: codex
            root: /Volumes/TestScratch/review-checkouts/c5241420f2a0ddc3
            reviewed_head: 77425d4f8952fa41f8b20d4ae304eecf63608a7e
            slot: ariadne:3
            requested_at: "2026-10-25T01:51:19.843063Z"
          phase: finished
          pid: 69727
          started_at: "2026-10-25T01:51:19.954877Z"
          finished_at: "2026-10-25T01:53:54.021845Z"
          session_id: 01a170c1-5442-7122-94dc-98a60b800fa5
          outcome: returned
      "n": 2
      timestamp: "2026-10-24T18:53:54-07:00"
      dispatched: "2026-10-24T18:51:19-07:00"
      agent: codex
      reviewer: codex (openai; author openai)
      dispose:
        - id: BR-1
          disposition: addressed
          note: fleet_recovery.go:101 validates agent evidence before marker absence and retains diagnostics. Decoder and CLI regressions pass at HEAD; removing the guard in a disposable checkout makes both fail and reproduces the false successful handoff. The private executable unavailable-evidence case also passes. Disposable artifacts were removed.
          round: 2
      recipe: milestone-review
      isolation: fenced
      reviewed: 77425d4f8952fa41f8b20d4ae304eecf63608a7e
      blocked: false
---

# Gate ledger — ariadne#564 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-24T18:43:27-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `recovery-evidence-integrity` (kind: defect) Unavailable agent evidence can produce a successful worker-wave handoff
  cmd/tl-lead/fleet_recovery.go:110 returns on an absent asleep marker before validating agent evidence quality; recoveryWireAgent also discards error diagnostics. With an owned worker containing agent: {"state":"unavailable","error":"thread store unreadable"}, the pinned CLI prints "skipped", exits 0, and reports worker waves complete. This contradicts the plan's requirement that unavailable evidence remain unknown and block completion (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE). Validate agent evidence before treating marker absence as an ordinary skip, retain relevant diagnostics, and add decoder and CLI regressions requiring a named unknown row and nonzero incomplete handoff.

## Round 2 — 2026-10-24T18:53:54-07:00 (codex) — passed

### Disposed

- BR-1 — addressed — fleet_recovery.go:101 validates agent evidence before marker absence and retains diagnostics. Decoder and CLI regressions pass at HEAD; removing the guard in a disposable checkout makes both fail and reproduces the false successful handoff. The private executable unavailable-evidence case also passes. Disposable artifacts were removed.

## Open findings

(none — every finding has been disposed)
