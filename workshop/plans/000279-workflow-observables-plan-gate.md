---
gate: plan-quality
issue: 279
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-01T23:58:31-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Review verdicts for a landed issue have no named source once a squash/rebase merge and branch -D remove the evidence commits
          detail: merge.go:365 says the server merge may squash or rebase, leaving evidence commits off main, and merge.go:610/624 deletes the branch. "The landed range" then usually holds no Review-Verdict trailers. Name the source (completion.evidence_commit if reachable, or archived ledgers), say it reports unknown rather than absent when unreachable, and test a squash-style landing.
          family: evidence-source-after-landing
          round: 1
        - id: PQ-2
          severity: Important
          title: '"gatestate.Decide over its ledger" misdescribes the ledgers: two kinds, issue-wide, scoped via FilterBoundary'
          detail: There is a plan-quality ledger (planreview.go:28) and an issue-wide boundary ledger (boundaryledger.go:38) scoped per milestone by FilterBoundary (boundaryledger.go:60), with a round cap read from an env var (:187); readGateLedger reads the filesystem, not a ref. Per-boundary open_blocking needs DecideScoped over the filtered ledger and a fixed or recorded cap.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-3
          severity: Important
          title: landing.state mixes outcome (landed/not-landed) with the read-quality enum (present/absent/stale/unknown)
          detail: With a stale tracker, a landed issue cannot express both. schema_version 1 is the hard-to-reverse part of this issue, so keep state as read quality only and move the value to its own field (e.g. landing.outcome) across all sections.
          family: contract-field-conflation
          round: 1
        - id: PQ-4
          severity: Minor
          title: Test bullets list cases; replace them with one strategy line per risky function (fuzz Assemble over malformed card/details/ledger bytes)
          family: test-strategy-not-enumeration
          round: 1
        - id: PQ-5
          severity: Minor
          title: No fetch timeout or latency budget, and no bound on per-worktree git status fan-out (ARCH-CONSTRAINTS)
          family: missing-operating-envelope
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-01T23:59:51-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Reviews read from archived artifacts (branch plans/, then main history/plans/ via archivePlanArtifacts push.go:315); evidenced-but-unfound gives unknown; squash-landing test added.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Both ledgers named; FilterBoundary (gatestate/ledger.go:141) + DecideScoped (decide.go:59) over ref bytes, default cap, OpenBlocking only.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: state is read quality only in every section; landing.outcome, relation and verdict are separate fields; rule stated for future sections.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Assemble now has fuzz plus one table; the remaining real-git bullets are lifecycle fixtures, not an enumeration of edge cases.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: 'Operating envelope added: bounded fetch, per-issue git fan-out, budget asserted by counting runner calls.'
          round: 2
      blocked: false
content_hash: 471f7de4b20572dc03b34c724d63c06dd9ea7d9464f655b027950e37d38d6184
---

# Gate ledger — ariadne#279 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-01T23:58:31-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `evidence-source-after-landing` Review verdicts for a landed issue have no named source once a squash/rebase merge and branch -D remove the evidence commits
  merge.go:365 says the server merge may squash or rebase, leaving evidence commits off main, and merge.go:610/624 deletes the branch. "The landed range" then usually holds no Review-Verdict trailers. Name the source (completion.evidence_commit if reachable, or archived ledgers), say it reports unknown rather than absent when unreachable, and test a squash-style landing.
- **PQ-2** [Important] `unbacked-existing-behavior-claim` "gatestate.Decide over its ledger" misdescribes the ledgers: two kinds, issue-wide, scoped via FilterBoundary
  There is a plan-quality ledger (planreview.go:28) and an issue-wide boundary ledger (boundaryledger.go:38) scoped per milestone by FilterBoundary (boundaryledger.go:60), with a round cap read from an env var (:187); readGateLedger reads the filesystem, not a ref. Per-boundary open_blocking needs DecideScoped over the filtered ledger and a fixed or recorded cap.
- **PQ-3** [Important] `contract-field-conflation` landing.state mixes outcome (landed/not-landed) with the read-quality enum (present/absent/stale/unknown)
  With a stale tracker, a landed issue cannot express both. schema_version 1 is the hard-to-reverse part of this issue, so keep state as read quality only and move the value to its own field (e.g. landing.outcome) across all sections.
- **PQ-4** [Minor] `test-strategy-not-enumeration` Test bullets list cases; replace them with one strategy line per risky function (fuzz Assemble over malformed card/details/ledger bytes)
- **PQ-5** [Minor] `missing-operating-envelope` No fetch timeout or latency budget, and no bound on per-worktree git status fan-out (ARCH-CONSTRAINTS)

## Round 2 — 2026-10-01T23:59:51-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Reviews read from archived artifacts (branch plans/, then main history/plans/ via archivePlanArtifacts push.go:315); evidenced-but-unfound gives unknown; squash-landing test added.
- PQ-2 — addressed — Both ledgers named; FilterBoundary (gatestate/ledger.go:141) + DecideScoped (decide.go:59) over ref bytes, default cap, OpenBlocking only.
- PQ-3 — addressed — state is read quality only in every section; landing.outcome, relation and verdict are separate fields; rule stated for future sections.
- PQ-4 — addressed — Assemble now has fuzz plus one table; the remaining real-git bullets are lifecycle fixtures, not an enumeration of edge cases.
- PQ-5 — addressed — Operating envelope added: bounded fetch, per-issue git fan-out, budget asserted by counting runner calls.

## Open findings

(none — every finding has been disposed)
