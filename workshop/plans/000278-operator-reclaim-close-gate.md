---
gate: boundary-review
issue: 278
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-01T21:05:43-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: The wrong-owner test checks only the old workspace's start-plan; the plan's "new one passes" and change-code/close refusals are untested
          detail: reclaim_test.go:169-173. Add ownershipGate runs for change-code and close from the old slot, plus a passing gate from r.root, or revise plan step 2.
          family: plan-claimed-test-missing
          round: 1
        - id: BR-2
          severity: Minor
          title: The "already mine" rerun skips the local mirror refresh, so after a lost response that landed, details stay stale
          detail: reclaim.go:160-163 returns before refreshLocalMirror. Claim's own "already mine" path refreshes (claim.go:168-176), and the help promises the refresh.
          family: retry-path-skips-side-effects
          round: 1
        - id: BR-3
          severity: Minor
          title: reclaimHistory reads 50 card commits; the plan says 20 and the help says "every past reclaim"
          family: doc-code-bound-drift
          round: 1
        - id: BR-4
          severity: Minor
          title: --reason without --expect is silently ignored (inspect mode)
          family: silently-ignored-flag
          round: 1
        - id: BR-5
          severity: Minor
          title: TestReclaimIsOnlyOperatorInvoked scans only function bodies; package-level var initializers escape it
          family: guard-scope-gap
          round: 1
      recipe: milestone-review
      blocked: true
---

# Gate ledger — ariadne#278 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-01T21:05:43-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `plan-claimed-test-missing` The wrong-owner test checks only the old workspace's start-plan; the plan's "new one passes" and change-code/close refusals are untested
  reclaim_test.go:169-173. Add ownershipGate runs for change-code and close from the old slot, plus a passing gate from r.root, or revise plan step 2.
- **BR-2** [Minor] `retry-path-skips-side-effects` The "already mine" rerun skips the local mirror refresh, so after a lost response that landed, details stay stale
  reclaim.go:160-163 returns before refreshLocalMirror. Claim's own "already mine" path refreshes (claim.go:168-176), and the help promises the refresh.
- **BR-3** [Minor] `doc-code-bound-drift` reclaimHistory reads 50 card commits; the plan says 20 and the help says "every past reclaim"
- **BR-4** [Minor] `silently-ignored-flag` --reason without --expect is silently ignored (inspect mode)
- **BR-5** [Minor] `guard-scope-gap` TestReclaimIsOnlyOperatorInvoked scans only function bodies; package-level var initializers escape it

## Open findings

- **BR-1** [Important] `plan-claimed-test-missing` The wrong-owner test checks only the old workspace's start-plan; the plan's "new one passes" and change-code/close refusals are untested
- **BR-2** [Minor] `retry-path-skips-side-effects` The "already mine" rerun skips the local mirror refresh, so after a lost response that landed, details stay stale
- **BR-3** [Minor] `doc-code-bound-drift` reclaimHistory reads 50 card commits; the plan says 20 and the help says "every past reclaim"
- **BR-4** [Minor] `silently-ignored-flag` --reason without --expect is silently ignored (inspect mode)
- **BR-5** [Minor] `guard-scope-gap` TestReclaimIsOnlyOperatorInvoked scans only function bodies; package-level var initializers escape it
