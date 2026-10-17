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
---

# Gate ledger — ariadne#423 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-17T04:24:26-07:00 (codex) — BLOCKED

### Raised

- **BR-1** [Critical] `binding-presence-independent-of-read-quality` (kind: defect) Stale cards without completion bindings suppress the latest review
  cmd/sdlc/internal/observe/checkpoints.go:186 uses valued(comp.Read) as evidence that a completion binding exists. assembleCompletion preserves State=stale when a stale card has no binding (assemble.go:198), leaving ReviewedHead empty. A working issue with a valid latest SHIP consequently becomes unknown with an empty verdict and error "no review entry matches the completion reviewed head ". The complete family in this window is this selection guard and its shared text/JSON consumers. Check actual binding presence separately from read quality; preserve latest-review selection when absent and completion-bound selection when present. Add regression coverage for both stale states.

## Open findings

- **BR-1** [Critical] `binding-presence-independent-of-read-quality` (kind: defect) Stale cards without completion bindings suppress the latest review
