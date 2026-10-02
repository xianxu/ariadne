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
    - "n": 2
      timestamp: "2026-10-02T13:18:08-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: setstatus.go:282-286 rewritten; claim never moves status.
          round: 2
        - id: BR-2
          disposition: addressed
          note: statusDecision enforces owned on the start edge (setstatus.go:155); covered by TestStatusDecisionRecordsOrRefusesTheClaimant and the unforced verb-contract row.
          round: 2
        - id: BR-3
          disposition: addressed
          note: 'Intended per plan revision note 3 (plan.md:302); shaping-claim views handed to #284.'
          round: 2
        - id: BR-4
          disposition: addressed
          note: 'Recorded in #284 Log, which owns move semantics for open claims.'
          round: 2
        - id: BR-5
          disposition: addressed
          note: Imports regrouped in changecode.go, reclaim.go, startplan.go.
          round: 2
      findings:
        - id: BR-6
          severity: Minor
          title: start-plan admits held blocked/codecomplete cards via CanHoldOwner, widening the old working-only gate
          detail: startDecision (startdecision.go:24) gates on the ownership axis, so an owned codecomplete card now gets a fresh issue branch from main (possibly a retired name). A lifecycle verb's admission set should derive from the lifecycle axis; refuse codecomplete or record the widening as intended.
          family: verb-admission-axis
          round: 2
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

## Round 2 — 2026-10-02T13:18:08-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — setstatus.go:282-286 rewritten; claim never moves status.
- BR-2 — addressed — statusDecision enforces owned on the start edge (setstatus.go:155); covered by TestStatusDecisionRecordsOrRefusesTheClaimant and the unforced verb-contract row.
- BR-3 — addressed — Intended per plan revision note 3 (plan.md:302); shaping-claim views handed to #284.
- BR-4 — addressed — Recorded in #284 Log, which owns move semantics for open claims.
- BR-5 — addressed — Imports regrouped in changecode.go, reclaim.go, startplan.go.

### Raised

- **BR-6** [Minor] `verb-admission-axis` start-plan admits held blocked/codecomplete cards via CanHoldOwner, widening the old working-only gate
  startDecision (startdecision.go:24) gates on the ownership axis, so an owned codecomplete card now gets a fresh issue branch from main (possibly a retired name). A lifecycle verb's admission set should derive from the lifecycle axis; refuse codecomplete or record the widening as intended.

## Open findings

- **BR-6** [Minor] `verb-admission-axis` start-plan admits held blocked/codecomplete cards via CanHoldOwner, widening the old working-only gate
