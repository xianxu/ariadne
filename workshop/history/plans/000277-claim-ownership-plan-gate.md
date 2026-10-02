---
gate: plan-quality
issue: 277
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-01T13:42:52-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: milestone-close does not run prepareTrackerClose; plan's "same prepare path" is wrong
          detail: close.go:510 calls prepareTrackerClose only when mode == "issue". Name a separate pre-review ownership hook for milestone mode.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-2
          severity: Important
          title: Worktree-keyed MatchClaimant makes the owner Foreign after `sdlc move :N`, with no recovery path
          detail: 'adopt refuses owned cards and reassignment is deferred to #278. Either sdlc move re-stamps the claimant by CAS, or the plan states that move is blocked until #278.'
          family: ownership-key-vs-existing-workflow
          round: 1
        - id: PQ-3
          severity: Important
          title: Built-binary race/restart tests cannot use a package-var identity injection
          detail: claimremote_test.go runs buildFleetE2EBinary. Name a seam a subprocess can see (test build tag/ldflags), or derive distinct identities from real differences between the clones.
          family: test-seam-unreachable
          round: 1
        - id: PQ-4
          severity: Important
          title: One claimant card makes every older binary fail the whole tracker Snapshot, not just that card
          detail: reader.go:108 aborts Snapshot on the first ParseCard error. This is a fleet-wide flag day; it needs an explicit rollout and rebuild step or a tolerant reader shipped first.
          family: compat-blast-radius-unstated
          round: 1
        - id: PQ-5
          severity: Minor
          title: Test bullets enumerate cases in prose; compress to one strategy line per risky function
          detail: 'For example: ParseClaimant over hand-edited/old-version card YAML gets a fuzz/table test seeded with extra keys, non-strings, flow style and nulls.'
          family: test-prose-enumeration
          round: 1
        - id: PQ-6
          severity: Minor
          title: Plan cites PublicationTarget; the function is gitx.ResolvePublicationTarget
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-7
          severity: Minor
          title: 'No consolidated non-goals section (#278 reassignment, #279 observation, no claim ID or timestamp)'
          family: missing-non-goals
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-01T13:47:33-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: requireOwnership is a separate call in the computeClose tracker-era block for both modes, before the review (close.go:509 guard verified).
          round: 2
        - id: PQ-2
          disposition: addressed
          note: sdlc move re-stamps the claimant on the owner's own move; its ordering is raised as a new finding.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Built-binary tests use real differences (distinct worktree paths, host fingerprint); the injected identity is used only in-process.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Snapshot abort verified at reader.go:108-110; the documented flag-day rollout and rebuild step are in the plan and the close follow-up.
          round: 2
        - id: PQ-5
          disposition: not-addressed
          note: Only the ParseClaimant line was compressed; MatchClaimant, mirror, adopt and per-gate cases are still prose enumerations. Minor, carried to close.
          round: 2
        - id: PQ-6
          disposition: addressed
          round: 2
        - id: PQ-7
          disposition: addressed
          round: 2
      findings:
        - id: PQ-8
          severity: Important
          title: Move re-stamp CAS runs before the checkout switches; a failed switch leaves the owner Foreign in their own slot with no recovery
          detail: 'runMove (move.go:34-80) has a preflight, a re-observe, two switches and verification. If the claimant CAS lands first and a switch then fails, the gates and a rerun of move both refuse as Foreign in the source, and adopt refuses owned cards. State the effect order: re-stamp after both switches are verified, or let move re-stamp when the claimant matches the source or the destination on the same machine and repository. Also say how move maps the branch to an issue (non-issue branches skip), that dry-run makes no write, and whether the CAS needs the network (move is local today). Add a test that injects a switch failure after the CAS.'
          family: partial-progress-recovery-unstated
          round: 2
      blocked: true
    - "n": 3
      timestamp: "2026-10-01T13:48:30-07:00"
      agent: claude
      dispose:
        - id: PQ-5
          disposition: addressed
          note: Gate, adopt, relocation and parser tests are now one table-strategy line each; the residual MatchClaimant examples define the contract.
          round: 3
        - id: PQ-8
          disposition: addressed
          note: Relocation runs after the switches are verified; CAS failure is repaired by claim at the destination; mapping, dry-run, network and a failure-injection test are stated.
          round: 3
      blocked: false
    - "n": 4
      timestamp: "2026-10-01T13:49:30-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: 787934bb3595527f2b583f577cbfe37fee22124d6a632e7dd9feb6a9dbfd0de5
---

# Gate ledger — ariadne#277 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-01T13:42:52-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` milestone-close does not run prepareTrackerClose; plan's "same prepare path" is wrong
  close.go:510 calls prepareTrackerClose only when mode == "issue". Name a separate pre-review ownership hook for milestone mode.
- **PQ-2** [Important] `ownership-key-vs-existing-workflow` Worktree-keyed MatchClaimant makes the owner Foreign after `sdlc move :N`, with no recovery path
  adopt refuses owned cards and reassignment is deferred to #278. Either sdlc move re-stamps the claimant by CAS, or the plan states that move is blocked until #278.
- **PQ-3** [Important] `test-seam-unreachable` Built-binary race/restart tests cannot use a package-var identity injection
  claimremote_test.go runs buildFleetE2EBinary. Name a seam a subprocess can see (test build tag/ldflags), or derive distinct identities from real differences between the clones.
- **PQ-4** [Important] `compat-blast-radius-unstated` One claimant card makes every older binary fail the whole tracker Snapshot, not just that card
  reader.go:108 aborts Snapshot on the first ParseCard error. This is a fleet-wide flag day; it needs an explicit rollout and rebuild step or a tolerant reader shipped first.
- **PQ-5** [Minor] `test-prose-enumeration` Test bullets enumerate cases in prose; compress to one strategy line per risky function
  For example: ParseClaimant over hand-edited/old-version card YAML gets a fuzz/table test seeded with extra keys, non-strings, flow style and nulls.
- **PQ-6** [Minor] `unbacked-existing-behavior-claim` Plan cites PublicationTarget; the function is gitx.ResolvePublicationTarget
- **PQ-7** [Minor] `missing-non-goals` No consolidated non-goals section (#278 reassignment, #279 observation, no claim ID or timestamp)

## Round 2 — 2026-10-01T13:47:33-07:00 (claude) — BLOCKED

### Disposed

- PQ-1 — addressed — requireOwnership is a separate call in the computeClose tracker-era block for both modes, before the review (close.go:509 guard verified).
- PQ-2 — addressed — sdlc move re-stamps the claimant on the owner's own move; its ordering is raised as a new finding.
- PQ-3 — addressed — Built-binary tests use real differences (distinct worktree paths, host fingerprint); the injected identity is used only in-process.
- PQ-4 — addressed — Snapshot abort verified at reader.go:108-110; the documented flag-day rollout and rebuild step are in the plan and the close follow-up.
- PQ-5 — not-addressed — Only the ParseClaimant line was compressed; MatchClaimant, mirror, adopt and per-gate cases are still prose enumerations. Minor, carried to close.
- PQ-6 — addressed
- PQ-7 — addressed

### Raised

- **PQ-8** [Important] `partial-progress-recovery-unstated` Move re-stamp CAS runs before the checkout switches; a failed switch leaves the owner Foreign in their own slot with no recovery
  runMove (move.go:34-80) has a preflight, a re-observe, two switches and verification. If the claimant CAS lands first and a switch then fails, the gates and a rerun of move both refuse as Foreign in the source, and adopt refuses owned cards. State the effect order: re-stamp after both switches are verified, or let move re-stamp when the claimant matches the source or the destination on the same machine and repository. Also say how move maps the branch to an issue (non-issue branches skip), that dry-run makes no write, and whether the CAS needs the network (move is local today). Add a test that injects a switch failure after the CAS.

## Round 3 — 2026-10-01T13:48:30-07:00 (claude) — passed

### Disposed

- PQ-5 — addressed — Gate, adopt, relocation and parser tests are now one table-strategy line each; the residual MatchClaimant examples define the contract.
- PQ-8 — addressed — Relocation runs after the switches are verified; CAS failure is repaired by claim at the destination; mapping, dry-run, network and a failure-injection test are stated.

## Round 4 — 2026-10-01T13:49:30-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
