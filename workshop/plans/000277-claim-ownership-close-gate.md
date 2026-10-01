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

## Open findings

- **BR-1** [Important] `test-oracle-single-interleaving` Claim race test rejects the foreign-owner refusal a late-reading loser now gets
- **BR-2** [Important] `plan-code-contract-drift` Plan promises MachineFingerprint(raw, key) plus a key test; code has MachineFingerprint(raw)
- **BR-3** [Minor] `dry-run-writes` An owner's repeat claim with --dry-run still refreshes and writes the local details mirror
- **BR-4** [Minor] `rollout-doc-timing` M1 already writes claimant cards to the shared tracker, but the flag-day rollout note is deferred to M2
- **BR-5** [Minor] `plan-code-contract-drift` The race-test skip on an unreadable host machine ID promised by the plan is not implemented
- **BR-6** [Minor] `layout-knowledge-outside-owner` slotLabel re-derives the slot directory layout in cmd/sdlc instead of asking pkg/workspace
- **BR-7** [Minor] `missing-branch-unit-test` No claimDecision table test with a non-nil claimant; the Unknown (legacy working card) refusal is untested in M1
