---
gate: boundary-review
issue: 283
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-02T13:07:09-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: setstatus.go:276 comment still describes claim as an open→working lock broadcast
          detail: 'Since #283, claim records the owner and never moves status; the comment should say the estimate gate left claim/start-plan.'
          family: stale-claim-semantics-prose
          round: 1
        - id: BR-2
          severity: Minor
          title: The `owned` guard on the start edge has no named enforcement; set-status working claims and starts in one step
          detail: statusDecision stamps the setter as claimant on an unowned open card, which skips claim's details-on-main readiness check. Either enforce `owned` in checkTransitionGuards or document set-status as the manual claim+start.
          family: model-guard-unenforced
          round: 1
        - id: BR-3
          severity: Minor
          title: start-plan's contention warning counts only working cards, not open cards held by a shaping claim
          family: in-flight-means-held
          round: 1
        - id: BR-4
          severity: Minor
          title: requireCardOwnership's moved-here hint points at `sdlc claim`, which finishes relocations only for active statuses
          detail: For an open held card that was moved (start-plan's card write lost, then sdlc move), claim refuses with "claimed by" instead of finishing the relocation.
          family: relocation-hint-status-scope
          round: 1
        - id: BR-5
          severity: Minor
          title: Module imports placed inside the stdlib block in changecode.go, reclaim.go, startplan.go
          family: import-grouping
          round: 1
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#283 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-02T13:07:09-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `stale-claim-semantics-prose` setstatus.go:276 comment still describes claim as an open→working lock broadcast
  Since #283, claim records the owner and never moves status; the comment should say the estimate gate left claim/start-plan.
- **BR-2** [Minor] `model-guard-unenforced` The `owned` guard on the start edge has no named enforcement; set-status working claims and starts in one step
  statusDecision stamps the setter as claimant on an unowned open card, which skips claim's details-on-main readiness check. Either enforce `owned` in checkTransitionGuards or document set-status as the manual claim+start.
- **BR-3** [Minor] `in-flight-means-held` start-plan's contention warning counts only working cards, not open cards held by a shaping claim
- **BR-4** [Minor] `relocation-hint-status-scope` requireCardOwnership's moved-here hint points at `sdlc claim`, which finishes relocations only for active statuses
  For an open held card that was moved (start-plan's card write lost, then sdlc move), claim refuses with "claimed by" instead of finishing the relocation.
- **BR-5** [Minor] `import-grouping` Module imports placed inside the stdlib block in changecode.go, reclaim.go, startplan.go

## Open findings

- **BR-1** [Minor] `stale-claim-semantics-prose` setstatus.go:276 comment still describes claim as an open→working lock broadcast
- **BR-2** [Minor] `model-guard-unenforced` The `owned` guard on the start edge has no named enforcement; set-status working claims and starts in one step
- **BR-3** [Minor] `in-flight-means-held` start-plan's contention warning counts only working cards, not open cards held by a shaping claim
- **BR-4** [Minor] `relocation-hint-status-scope` requireCardOwnership's moved-here hint points at `sdlc claim`, which finishes relocations only for active statuses
- **BR-5** [Minor] `import-grouping` Module imports placed inside the stdlib block in changecode.go, reclaim.go, startplan.go
