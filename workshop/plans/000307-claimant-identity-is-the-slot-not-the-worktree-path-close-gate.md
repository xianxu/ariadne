---
gate: boundary-review
issue: 307
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-11T01:50:11-07:00"
      dispatched: "2026-10-11T01:47:31-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: observe IsClaimant and fleet claim association/dangling still key on worktree, not slot
          detail: cmd/sdlc/internal/observe/checkpoints.go:72 and cmd/sdlc/internal/fleet/claims.go:129,149 match Claimant.Worktree to a checkout path; a slot-owned claim whose recorded checkout is gone reports as dangling though the slot still owns it. Display/inventory only, not a gate.
          family: ownership-shadow-consumer
          round: 1
        - id: BR-2
          severity: Minor
          title: same-machine fleets with identically named slots now share ownership
          detail: Two fleet roots on one machine each with ariadne-slot1 both label ariadne:1; the worktree used to separate them. This follows from the Spec but is undocumented.
          family: slot-identity-uniqueness
          round: 1
        - id: BR-3
          severity: Minor
          title: close and milestone-close are tested only for refusal, not acceptance from another checkout of the owning slot
          detail: TestOwnershipGatesFollowTheSlot tests acceptance only for start-plan and change-code; all four share requireCardOwnership.
          family: done-when-both-states
          round: 1
      recipe: small-diff-review
      reviewed: 89a716b83cee0d8d59986d9376bdea94bf4d13a7
      blocked: false
---

# Gate ledger — ariadne#307 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-11T01:50:11-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `ownership-shadow-consumer` observe IsClaimant and fleet claim association/dangling still key on worktree, not slot
  cmd/sdlc/internal/observe/checkpoints.go:72 and cmd/sdlc/internal/fleet/claims.go:129,149 match Claimant.Worktree to a checkout path; a slot-owned claim whose recorded checkout is gone reports as dangling though the slot still owns it. Display/inventory only, not a gate.
- **BR-2** [Minor] `slot-identity-uniqueness` same-machine fleets with identically named slots now share ownership
  Two fleet roots on one machine each with ariadne-slot1 both label ariadne:1; the worktree used to separate them. This follows from the Spec but is undocumented.
- **BR-3** [Minor] `done-when-both-states` close and milestone-close are tested only for refusal, not acceptance from another checkout of the owning slot
  TestOwnershipGatesFollowTheSlot tests acceptance only for start-plan and change-code; all four share requireCardOwnership.

## Open findings

- **BR-1** [Minor] `ownership-shadow-consumer` observe IsClaimant and fleet claim association/dangling still key on worktree, not slot
- **BR-2** [Minor] `slot-identity-uniqueness` same-machine fleets with identically named slots now share ownership
- **BR-3** [Minor] `done-when-both-states` close and milestone-close are tested only for refusal, not acceptance from another checkout of the owning slot
