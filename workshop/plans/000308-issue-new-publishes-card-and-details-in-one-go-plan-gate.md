---
gate: plan-quality
issue: 308
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-12T07:29:17-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Minor
          title: --no-project redefinition removes the precise waiver for track-only milestones; die text at close.go:860 must change
          detail: After Task 9, ticking without a detail block requires --force (all gates). Update the die message and accept or mitigate the lost precise path.
          family: gate-flag-precision
          round: 1
        - id: PQ-2
          severity: Minor
          title: Peer origin publish precedes the close commit; state ordering and idempotent retry
          detail: publishPeerProject pushes from applyClose before close completes; note that re-derivation from fetched bytes makes retry a no-op commit.
          family: external-effect-ordering
          round: 1
        - id: PQ-3
          severity: Minor
          title: issueDirGuard must run before the repositoryTracked split to cover the legacy path (C5)
          detail: runIssueNew returns into runLegacyIssueNew at issue.go:238-242, before the plan's step (1).
          family: guard-placement-covers-all-paths
          round: 1
        - id: PQ-4
          severity: Minor
          title: runIssueNew keeps the detached-HEAD refusal and the stdout DestinationPath from the local-write era
          detail: issue.go:285 refuses with no branch and issue.go:364 prints a path that will no longer exist locally; drop the refusal or justify it, and check stdout consumers.
          family: stale-local-write-contract
          round: 1
        - id: PQ-5
          severity: Minor
          title: M2 transform should refuse when TLOwners fails to parse fetched peer bytes
          detail: Matches the local behaviour at close.go:813; the plan names only the owned case.
          family: untrusted-fetched-bytes-fail-closed
          round: 1
        - id: PQ-6
          severity: Minor
          title: issue new now makes three remote round trips; latency and retry-exhaustion behaviour unstated
          family: operating-envelope-unstated
          round: 1
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T07:30:48-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Task 9 rewrites the close.go:856-860 refusal and records the TL-accepted trade-off.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Ordering bullet states publish-before-close-commit and the no-commit retry from re-derivation.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: issueDirGuard runs before the repositoryTracked split at issue.go:239.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Detached-HEAD refusal dropped; stdout path kept with new meaning; consumer sweep added.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: TLOwners parse error refuses in the transform, matching close.go:812-813.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: ARCH-CONSTRAINTS note covers round trips, bounds, and exhaustion behaviour.
          round: 2
      blocked: false
content_hash: 0d491fe14e24adb558a16f256bc011513c946dfa3a8c78034d0cb4f61462a31c
---

# Gate ledger — ariadne#308 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T07:29:17-07:00 (claude) — passed

### Raised

- **PQ-1** [Minor] `gate-flag-precision` --no-project redefinition removes the precise waiver for track-only milestones; die text at close.go:860 must change
  After Task 9, ticking without a detail block requires --force (all gates). Update the die message and accept or mitigate the lost precise path.
- **PQ-2** [Minor] `external-effect-ordering` Peer origin publish precedes the close commit; state ordering and idempotent retry
  publishPeerProject pushes from applyClose before close completes; note that re-derivation from fetched bytes makes retry a no-op commit.
- **PQ-3** [Minor] `guard-placement-covers-all-paths` issueDirGuard must run before the repositoryTracked split to cover the legacy path (C5)
  runIssueNew returns into runLegacyIssueNew at issue.go:238-242, before the plan's step (1).
- **PQ-4** [Minor] `stale-local-write-contract` runIssueNew keeps the detached-HEAD refusal and the stdout DestinationPath from the local-write era
  issue.go:285 refuses with no branch and issue.go:364 prints a path that will no longer exist locally; drop the refusal or justify it, and check stdout consumers.
- **PQ-5** [Minor] `untrusted-fetched-bytes-fail-closed` M2 transform should refuse when TLOwners fails to parse fetched peer bytes
  Matches the local behaviour at close.go:813; the plan names only the owned case.
- **PQ-6** [Minor] `operating-envelope-unstated` issue new now makes three remote round trips; latency and retry-exhaustion behaviour unstated

## Round 2 — 2026-10-12T07:30:48-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Task 9 rewrites the close.go:856-860 refusal and records the TL-accepted trade-off.
- PQ-2 — addressed — Ordering bullet states publish-before-close-commit and the no-commit retry from re-derivation.
- PQ-3 — addressed — issueDirGuard runs before the repositoryTracked split at issue.go:239.
- PQ-4 — addressed — Detached-HEAD refusal dropped; stdout path kept with new meaning; consumer sweep added.
- PQ-5 — addressed — TLOwners parse error refuses in the transform, matching close.go:812-813.
- PQ-6 — addressed — ARCH-CONSTRAINTS note covers round trips, bounds, and exhaustion behaviour.

## Open findings

(none — every finding has been disposed)
