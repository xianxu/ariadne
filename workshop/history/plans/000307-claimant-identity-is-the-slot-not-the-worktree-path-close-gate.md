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
    - "n": 2
      timestamp: "2026-10-11T01:57:36-07:00"
      dispatched: "2026-10-11T01:56:40-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: observe now keys IsClaimant on the slot (with a new defect, raised below); fleet placement is documented as recorded-checkout inventory in the atlas and the PlaceClaims comment.
          round: 2
        - id: BR-2
          disposition: addressed
          note: atlas/workflow/issue-tracker.md:44-46 documents that slot addresses are unique per fleet and that a machine should run one fleet.
          round: 2
        - id: BR-3
          disposition: addressed
          note: ownership_test.go:508-521 now tests that close and milestone-close from another checkout of the owning slot get past the ownership gate to the review.
          round: 2
      findings:
        - id: BR-4
          severity: Important
          title: observe isClaimantCheckout rewrites MatchClaimant without the machine check, so a same-named slot on another machine counts as claimant
          detail: 'This is the 2nd finding in family ownership-shadow-consumer. Rule: every claimant/ownership judgment goes through issue.MatchClaimant; no consumer re-implements the slot/worktree comparison. checkpoints.go:66 should build a Claimant from in.Me (Machine, Repository) plus h.Address and h.Path and call MatchClaimant (false when in.Me is nil). Add a test case for a claimant on another machine. Today a pair:1 claim from laptop B marks laptop A''s pair:1 checkout IsClaimant=true, which contradicts Relation=other-workspace in the same output.'
          family: ownership-shadow-consumer
          round: 2
      recipe: fix-review
      blocked: true
    - "n": 3
      timestamp: "2026-10-11T01:59:05-07:00"
      dispatched: "2026-10-11T01:58:26-07:00"
      agent: claude
      dispose:
        - id: BR-4
          disposition: addressed
          note: isClaimantCheckout builds the comparison from in.Me (machine and repository) with the holding checkout's slot and path, calls issue.MatchClaimant, and returns false when in.Me is nil. The test case "same slot, other machine" pins this, and collectHolding now uses HomeSlot like claimantIdentity.
          round: 3
      findings:
        - id: BR-5
          severity: Minor
          title: Claimant is copied field by field between the observe and issue types in two places
          detail: 'checkpoints.go:72 and assemble.go:150 hand-copy the same six fields in opposite directions. A single converter would stop a new Claimant field from being dropped in one of them. Cosmetic: ownership is still decided only in MatchClaimant.'
          family: ownership-shadow-consumer
          round: 3
      recipe: fix-review
      reviewed: deb5d336487b4a9b1afbb4e2de05b5c03c6ead16
      blocked: false
    - "n": 4
      timestamp: "2026-10-11T02:08:53-07:00"
      dispatched: "2026-10-11T02:02:57-07:00"
      agent: claude
      dispose:
        - id: BR-5
          disposition: addressed
          note: All three hand-copies (observe checkpoints.go:71, assemble.go:150, fleet claims.go:182) now use direct struct conversion, so a field added to one Claimant fails to compile instead of being dropped. The fields match exactly across issue, observe and fleet; build, vet and the observe and fleet tests pass.
          round: 4
      recipe: fix-review
      reviewed: 124de8b635a55d249587925391912b821d042637
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

## Round 2 — 2026-10-11T01:57:36-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — observe now keys IsClaimant on the slot (with a new defect, raised below); fleet placement is documented as recorded-checkout inventory in the atlas and the PlaceClaims comment.
- BR-2 — addressed — atlas/workflow/issue-tracker.md:44-46 documents that slot addresses are unique per fleet and that a machine should run one fleet.
- BR-3 — addressed — ownership_test.go:508-521 now tests that close and milestone-close from another checkout of the owning slot get past the ownership gate to the review.

### Raised

- **BR-4** [Important] `ownership-shadow-consumer` observe isClaimantCheckout rewrites MatchClaimant without the machine check, so a same-named slot on another machine counts as claimant
  This is the 2nd finding in family ownership-shadow-consumer. Rule: every claimant/ownership judgment goes through issue.MatchClaimant; no consumer re-implements the slot/worktree comparison. checkpoints.go:66 should build a Claimant from in.Me (Machine, Repository) plus h.Address and h.Path and call MatchClaimant (false when in.Me is nil). Add a test case for a claimant on another machine. Today a pair:1 claim from laptop B marks laptop A's pair:1 checkout IsClaimant=true, which contradicts Relation=other-workspace in the same output.

## Round 3 — 2026-10-11T01:59:05-07:00 (claude) — passed

### Disposed

- BR-4 — addressed — isClaimantCheckout builds the comparison from in.Me (machine and repository) with the holding checkout's slot and path, calls issue.MatchClaimant, and returns false when in.Me is nil. The test case "same slot, other machine" pins this, and collectHolding now uses HomeSlot like claimantIdentity.

### Raised

- **BR-5** [Minor] `ownership-shadow-consumer` Claimant is copied field by field between the observe and issue types in two places
  checkpoints.go:72 and assemble.go:150 hand-copy the same six fields in opposite directions. A single converter would stop a new Claimant field from being dropped in one of them. Cosmetic: ownership is still decided only in MatchClaimant.

## Round 4 — 2026-10-11T02:08:53-07:00 (claude) — passed

### Disposed

- BR-5 — addressed — All three hand-copies (observe checkpoints.go:71, assemble.go:150, fleet claims.go:182) now use direct struct conversion, so a field added to one Claimant fails to compile instead of being dropped. The fields match exactly across issue, observe and fleet; build, vet and the observe and fleet tests pass.

## Open findings

(none — every finding has been disposed)
