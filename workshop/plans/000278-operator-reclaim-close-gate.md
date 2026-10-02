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
    - "n": 2
      timestamp: "2026-10-01T21:23:53-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: reclaim_test.go:164-171 runs start-plan, change-code and close from the old slot, each refused with judged=false; :190-198 shows the new owner's start-plan passes.
          round: 2
        - id: BR-2
          disposition: addressed
          note: reclaim.go:166-168 refreshes on the already-mine path; scratch mutation removing it turns TestReclaimStaleAndLostResponse red at reclaim_test.go:237.
          round: 2
        - id: BR-3
          disposition: addressed
          note: reclaimHistoryLimit=20 with --grep on the trailer (max-count applies after grep); help and plan both say last 20.
          round: 2
        - id: BR-4
          disposition: addressed
          note: reclaim.go:157-159 refuses --reason without --expect; tested at reclaim_test.go:187-189.
          round: 2
        - id: BR-5
          disposition: addressed
          note: Guard now walks GenDecl ValueSpec initializers (reclaim_test.go:334-341); the vacuity check confirms reclaimEffect's initializer is seen.
          round: 2
      findings:
        - id: BR-6
          severity: Important
          title: README's ownership paragraph lists claim, --adopt and move but not the new sdlc reclaim verb
          detail: 'README.md:26-30 describes #277 ownership verbs; add one sentence naming `sdlc reclaim` (inspect, then --expect/--reason) as the operator-directed transfer between owners.'
          family: docs-new-surface-missing
          round: 2
      recipe: milestone-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-01T21:24:50-07:00"
      agent: claude
      dispose:
        - id: BR-6
          disposition: addressed
          note: README.md:29-32 now names `sdlc reclaim --issue N`, inspect-first, then `--expect REV --reason`; it matches reclaim.go:119-120,61-62,223.
          round: 3
      findings:
        - id: BR-7
          severity: Minor
          title: README.md:32 reclaim sentence leaves an over-long unwrapped line in the ownership paragraph
          detail: Cosmetic only; reflow the paragraph. Not a missing-surface repeat, just formatting at the site of the BR-6 fix.
          family: docs-new-surface-missing
          round: 3
      recipe: milestone-review
      blocked: false
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

## Round 2 — 2026-10-01T21:23:53-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — reclaim_test.go:164-171 runs start-plan, change-code and close from the old slot, each refused with judged=false; :190-198 shows the new owner's start-plan passes.
- BR-2 — addressed — reclaim.go:166-168 refreshes on the already-mine path; scratch mutation removing it turns TestReclaimStaleAndLostResponse red at reclaim_test.go:237.
- BR-3 — addressed — reclaimHistoryLimit=20 with --grep on the trailer (max-count applies after grep); help and plan both say last 20.
- BR-4 — addressed — reclaim.go:157-159 refuses --reason without --expect; tested at reclaim_test.go:187-189.
- BR-5 — addressed — Guard now walks GenDecl ValueSpec initializers (reclaim_test.go:334-341); the vacuity check confirms reclaimEffect's initializer is seen.

### Raised

- **BR-6** [Important] `docs-new-surface-missing` README's ownership paragraph lists claim, --adopt and move but not the new sdlc reclaim verb
  README.md:26-30 describes #277 ownership verbs; add one sentence naming `sdlc reclaim` (inspect, then --expect/--reason) as the operator-directed transfer between owners.

## Round 3 — 2026-10-01T21:24:50-07:00 (claude) — passed

### Disposed

- BR-6 — addressed — README.md:29-32 now names `sdlc reclaim --issue N`, inspect-first, then `--expect REV --reason`; it matches reclaim.go:119-120,61-62,223.

### Raised

- **BR-7** [Minor] `docs-new-surface-missing` README.md:32 reclaim sentence leaves an over-long unwrapped line in the ownership paragraph
  Cosmetic only; reflow the paragraph. Not a missing-surface repeat, just formatting at the site of the BR-6 fix.

## Open findings

- **BR-7** [Minor] `docs-new-surface-missing` README.md:32 reclaim sentence leaves an over-long unwrapped line in the ownership paragraph
