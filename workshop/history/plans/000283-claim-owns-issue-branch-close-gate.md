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
    - "n": 3
      timestamp: "2026-10-02T13:23:22-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: startdecision.go:28 admits only the start edge's From/To; the TestStartDecision table now expects refusal for blocked/codecomplete owned cells, which the old CanHoldOwner gate admitted, so a revert goes red. Remaining CanHoldOwner callers are the ownership verbs (claim/reclaim), correctly on the ownership axis.
          round: 3
      recipe: milestone-review
      blocked: false
    - "n": 4
      timestamp: "2026-10-07T11:35:28-07:00"
      agent: claude
      findings:
        - id: BR-7
          severity: Important
          title: Recovery catalog contracts for claim/start-plan/reclaim still state pre-283 semantics; start-plan claims no remote effect
          detail: '2nd finding in this family. Rule: a verb''s rendered recovery contract (catalog.go Effects/Preconditions/Repeat/LostResponse/Ends/Proofs plus page.go Example text) is part of the verb''s surface and is swept in the same window that changes the verb. Stale now: claim Effects (open to working), claim Ends (leaves working), claim Repeat/Preconditions (a working card with no owner, though adopt gates on IsActive), start-plan Effects (Nothing is pushed) and LostResponse (no remote effect) despite its UpdateCard start write, start-plan Preconditions (card is working, no tracked changes), start-plan Proofs naming no test of the start write, reclaim Preconditions (excludes owned open), page.go:69 Otherwise (still open now also means success).'
          family: stale-claim-semantics-prose
          round: 4
      recipe: milestone-review
      blocked: false
    - "n": 5
      timestamp: "2026-10-07T11:42:56-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: claim/start-plan/reclaim/set-status catalog fields and page.go Otherwise now match claim.go, startplan.go:296-341, reclaim and setstatus.go:155-158; cited tests exist.
          round: 5
      findings:
        - id: BR-8
          severity: Important
          title: change-code catalog entry carries start-plan's Proof rows and omits its own new open-card refusal
          detail: 'catalog.go:78-79 claims change-code "starts the owner''s open card" and "carries only this issue''s own details edits", but changecode.go:351-354 refuses an open card toward start-plan; Preconditions omit that refusal and TestChangeCodeRefusesUnstartedClaim is uncited. 3rd in family: rule = each catalog Proof describes its own verb''s behaviour and every changed verb''s whole entry is swept; consider a recovery_proofs_test check that no Proof claim string appears under two different verbs.'
          family: stale-claim-semantics-prose
          round: 5
      recipe: milestone-review
      blocked: false
    - "n": 6
      timestamp: "2026-10-07T12:50:13-07:00"
      agent: claude
      dispose:
        - id: BR-8
          disposition: addressed
          note: catalog.go change-code entry now lists only its own proofs plus "refuses an owned card that start-plan never started" citing TestChangeCodeRefusesUnstartedClaim (changecode_tracker_test.go:52, exercising changecode.go:351-353); Preconditions name the open-card refusal; recovery contract test passes.
          round: 6
      findings:
        - id: BR-9
          severity: Minor
          title: No test fails if closetracker.go or issuerecovery.go reverts to env.ancestorOf instead of closeAncestorOf
          detail: newestClose and closeAncestorOf are unit-tested separately; an e2e close-rebase-reclose test would pin the two call-site wirings.
          family: rebase-close-wiring-untested
          round: 6
        - id: BR-10
          severity: Minor
          title: atlas does not mention the rebase-aware close-generation supersession rule
          detail: Only the recovery catalog Ends text records it; add one sentence to atlas/workflow/issue-tracker.md near the close/receipt section.
          family: atlas-missing-surface
          round: 6
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

## Round 3 — 2026-10-02T13:23:22-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — startdecision.go:28 admits only the start edge's From/To; the TestStartDecision table now expects refusal for blocked/codecomplete owned cells, which the old CanHoldOwner gate admitted, so a revert goes red. Remaining CanHoldOwner callers are the ownership verbs (claim/reclaim), correctly on the ownership axis.

## Round 4 — 2026-10-07T11:35:28-07:00 (claude) — passed

### Raised

- **BR-7** [Important] `stale-claim-semantics-prose` Recovery catalog contracts for claim/start-plan/reclaim still state pre-283 semantics; start-plan claims no remote effect
  2nd finding in this family. Rule: a verb's rendered recovery contract (catalog.go Effects/Preconditions/Repeat/LostResponse/Ends/Proofs plus page.go Example text) is part of the verb's surface and is swept in the same window that changes the verb. Stale now: claim Effects (open to working), claim Ends (leaves working), claim Repeat/Preconditions (a working card with no owner, though adopt gates on IsActive), start-plan Effects (Nothing is pushed) and LostResponse (no remote effect) despite its UpdateCard start write, start-plan Preconditions (card is working, no tracked changes), start-plan Proofs naming no test of the start write, reclaim Preconditions (excludes owned open), page.go:69 Otherwise (still open now also means success).

## Round 5 — 2026-10-07T11:42:56-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — claim/start-plan/reclaim/set-status catalog fields and page.go Otherwise now match claim.go, startplan.go:296-341, reclaim and setstatus.go:155-158; cited tests exist.

### Raised

- **BR-8** [Important] `stale-claim-semantics-prose` change-code catalog entry carries start-plan's Proof rows and omits its own new open-card refusal
  catalog.go:78-79 claims change-code "starts the owner's open card" and "carries only this issue's own details edits", but changecode.go:351-354 refuses an open card toward start-plan; Preconditions omit that refusal and TestChangeCodeRefusesUnstartedClaim is uncited. 3rd in family: rule = each catalog Proof describes its own verb's behaviour and every changed verb's whole entry is swept; consider a recovery_proofs_test check that no Proof claim string appears under two different verbs.

## Round 6 — 2026-10-07T12:50:13-07:00 (claude) — passed

### Disposed

- BR-8 — addressed — catalog.go change-code entry now lists only its own proofs plus "refuses an owned card that start-plan never started" citing TestChangeCodeRefusesUnstartedClaim (changecode_tracker_test.go:52, exercising changecode.go:351-353); Preconditions name the open-card refusal; recovery contract test passes.

### Raised

- **BR-9** [Minor] `rebase-close-wiring-untested` No test fails if closetracker.go or issuerecovery.go reverts to env.ancestorOf instead of closeAncestorOf
  newestClose and closeAncestorOf are unit-tested separately; an e2e close-rebase-reclose test would pin the two call-site wirings.
- **BR-10** [Minor] `atlas-missing-surface` atlas does not mention the rebase-aware close-generation supersession rule
  Only the recovery catalog Ends text records it; add one sentence to atlas/workflow/issue-tracker.md near the close/receipt section.

## Open findings

- **BR-9** [Minor] `rebase-close-wiring-untested` No test fails if closetracker.go or issuerecovery.go reverts to env.ancestorOf instead of closeAncestorOf
- **BR-10** [Minor] `atlas-missing-surface` atlas does not mention the rebase-aware close-generation supersession rule
