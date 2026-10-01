---
gate: boundary-review
issue: 277
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-01T14:01:43-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Claim race test rejects the foreign-owner refusal a late-reading loser now gets
          detail: A loser whose snapshot follows the winner's publish hits ownedBy, which returns "is working, claimed by ...". That message contains neither "not open" nor "changed while claiming", so the test is green only when both clones read before either publishes. Accept "claimed by" (or add a deterministic in-process loser test), and update the claim.md loser wording to match.
          family: test-oracle-single-interleaving
          round: 1
        - id: BR-2
          severity: Important
          title: Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
          detail: The Core concepts table and the M1 step describe a key parameter and a "different key gives a different value" test. The code uses a fixed domain prefix and tests separation between different raw IDs. The behaviour is fine; add a plan Revisions entry so the plan matches the code.
          family: plan-code-contract-drift
          round: 1
        - id: BR-3
          severity: Minor
          title: An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
          family: dry-run-writes
          round: 1
        - id: BR-4
          severity: Minor
          title: M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
          detail: Any claim made with this branch's binary breaks stale fleet binaries. Add the atlas rollout paragraph now, or avoid claiming with this binary until the work lands.
          family: rollout-doc-timing
          round: 1
        - id: BR-5
          severity: Minor
          title: The race-test skip on an unreadable host machine ID promised by the plan is not implemented
          family: plan-code-contract-drift
          round: 1
        - id: BR-6
          severity: Minor
          title: slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
          family: layout-knowledge-outside-owner
          round: 1
        - id: BR-7
          severity: Minor
          title: No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1
          family: missing-branch-unit-test
          round: 1
      boundary: M1
      recipe: milestone-review
      blocked: true
    - "n": 2
      timestamp: "2026-10-01T14:06:41-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Oracle now accepts "claimed by" (claimremote_test.go:182); help text updated; TestClaimDecisionOwnership pins the foreign-owner refusal deterministically.
          round: 2
        - id: BR-2
          disposition: addressed
          note: Plan Revisions entry 2026-10-01 records the fixed-key MachineFingerprint(raw); domain-separation test added.
          round: 2
        - id: BR-3
          disposition: addressed
          note: claim.go:138 dry-run guard; TestClaimDryRunOwnerRepeatWritesNothing goes red when the guard is reverted (verified in scratch worktree).
          round: 2
        - id: BR-4
          disposition: addressed
          note: Rollout paragraph now in atlas/workflow/issue-tracker.md within M1.
          round: 2
        - id: BR-5
          disposition: addressed
          note: TestClaimRaceHasExactlyOneWinner skips when machineID() errors.
          round: 2
        - id: BR-6
          disposition: addressed
          note: slotLabel delegates to pkg/workspace Identity.UsesSlotLayout, which reuses slotNumber.
          round: 2
        - id: BR-7
          disposition: addressed
          note: TestClaimDecisionOwnership covers stamp, Mine, Foreign, Unknown (--adopt), and non-working statuses.
          round: 2
      findings:
        - id: BR-8
          severity: Minor
          title: TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
          detail: Move the plain-primary assertion before writing worktree/ariadne-slotX so the no-slot-dirs state is still exercised.
          family: stale-test-precondition
          round: 2
      boundary: M1
      recipe: milestone-review
      blocked: false
---

# Gate ledger — ariadne#277 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-01T14:01:43-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `test-oracle-single-interleaving` Claim race test rejects the foreign-owner refusal a late-reading loser now gets
  A loser whose snapshot follows the winner's publish hits ownedBy, which returns "is working, claimed by ...". That message contains neither "not open" nor "changed while claiming", so the test is green only when both clones read before either publishes. Accept "claimed by" (or add a deterministic in-process loser test), and update the claim.md loser wording to match.
- **BR-2** [Important] `plan-code-contract-drift` Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
  The Core concepts table and the M1 step describe a key parameter and a "different key gives a different value" test. The code uses a fixed domain prefix and tests separation between different raw IDs. The behaviour is fine; add a plan Revisions entry so the plan matches the code.
- **BR-3** [Minor] `dry-run-writes` An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
- **BR-4** [Minor] `rollout-doc-timing` M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
  Any claim made with this branch's binary breaks stale fleet binaries. Add the atlas rollout paragraph now, or avoid claiming with this binary until the work lands.
- **BR-5** [Minor] `plan-code-contract-drift` The race-test skip on an unreadable host machine ID promised by the plan is not implemented
- **BR-6** [Minor] `layout-knowledge-outside-owner` slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
- **BR-7** [Minor] `missing-branch-unit-test` No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1

## Round 2 — 2026-10-01T14:06:41-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — Oracle now accepts "claimed by" (claimremote_test.go:182); help text updated; TestClaimDecisionOwnership pins the foreign-owner refusal deterministically.
- BR-2 — addressed — Plan Revisions entry 2026-10-01 records the fixed-key MachineFingerprint(raw); domain-separation test added.
- BR-3 — addressed — claim.go:138 dry-run guard; TestClaimDryRunOwnerRepeatWritesNothing goes red when the guard is reverted (verified in scratch worktree).
- BR-4 — addressed — Rollout paragraph now in atlas/workflow/issue-tracker.md within M1.
- BR-5 — addressed — TestClaimRaceHasExactlyOneWinner skips when machineID() errors.
- BR-6 — addressed — slotLabel delegates to pkg/workspace Identity.UsesSlotLayout, which reuses slotNumber.
- BR-7 — addressed — TestClaimDecisionOwnership covers stamp, Mine, Foreign, Unknown (--adopt), and non-working statuses.

### Raised

- **BR-8** [Minor] `stale-test-precondition` TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
  Move the plain-primary assertion before writing worktree/ariadne-slotX so the no-slot-dirs state is still exercised.

## Open findings

- **BR-8** [Minor] `stale-test-precondition` TestSlotLabelOnlyWhereSlotsExist's "no slots" check runs after the malformed slot dir is created, duplicating the prior assertion
