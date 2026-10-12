---
gate: plan-quality
issue: 309
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-10-12T11:18:02-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: Legacy merge path does not pin PR head; D1 claim holds only for the durable path
          detail: runMerge calls realGH.PRMerge (ghclient.go:153) with no --match-head-commit, so checks read on head H can land a later head. Either pin the head SHA in PRMerge and read checks against it, or declare the gap a non-goal; changing PRMerge's signature is a seam change the plan should state.
          family: unbacked-existing-behavior-claim
          round: 1
        - id: PQ-2
          severity: Important
          title: gh pr checks exits non-zero (1 failing, 8 pending) with valid JSON; plan treats non-zero as query failure
          detail: Parse stdout first and classify by bucket; a non-zero exit is a query failure only when stdout doesn't parse. Seed the parser test from real captured gh output for exit codes 0, 1 and 8 plus the no-checks message, so the fake matches the real binary.
          family: external-dependency-contract-unverified
          round: 1
        - id: PQ-3
          severity: Important
          title: D5 protected list is hand-picked and misses .codex/config.toml; no stated source for the enumeration
          detail: .codex/config.toml (Codex sandbox config, base.manifest:66-68) can loosen gates and isn't on the list. State the source the list is derived from (gate-relevant base.manifest rows plus local additions), sweep it, and note which entries exist only in the peer repos.
          family: protected-list-enumeration-source
          round: 1
        - id: PQ-4
          severity: Minor
          title: M1/M2 test bullets list test cases in prose; compress to one strategy line per risky function
          family: test-plan-enumerates-cases
          round: 1
        - id: PQ-5
          severity: Minor
          title: pull_request_target run block must take PR values via env and fetch refs/pull/N/head, never interpolate head.ref
          family: untrusted-input-in-workflow-script
          round: 1
        - id: PQ-6
          severity: Minor
          title: Confirm gh pr checks reports the latest guard run after a labeled re-run, not the earlier failure
          family: check-rerun-supersession
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T11:19:38-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: D1 changes PRMerge to take expectedHead with --match-head-commit (ghclient.go:149 today lacks it; ghlanding.go:163 durable path already pins).
          round: 2
        - id: PQ-2
          disposition: addressed
          note: Adapter parses stdout first; non-zero exit is a failure only when stdout is neither JSON nor the no-checks message; real-output fixtures.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: D5 derives the list from three stated sources, includes .codex/config.toml (base.manifest:68), names peer-only entries and exclusions.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Test bullets compressed to per-function strategies.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: 'Run block uses env: and fetches refs/pull/N/head; head.ref interpolation forbidden.'
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Live check confirms gh pr checks reports the latest guard run after labeling.
          round: 2
      blocked: false
    - "n": 3
      timestamp: "2026-10-12T11:20:23-07:00"
      agent: claude
      blocked: false
      protocol_error: no valid findings block
content_hash: 52cac80b666e066ea3fb851f20c5c180f48ce50f2d501ede14bec26c574bee9a
---

# Gate ledger — ariadne#309 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T11:18:02-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `unbacked-existing-behavior-claim` Legacy merge path does not pin PR head; D1 claim holds only for the durable path
  runMerge calls realGH.PRMerge (ghclient.go:153) with no --match-head-commit, so checks read on head H can land a later head. Either pin the head SHA in PRMerge and read checks against it, or declare the gap a non-goal; changing PRMerge's signature is a seam change the plan should state.
- **PQ-2** [Important] `external-dependency-contract-unverified` gh pr checks exits non-zero (1 failing, 8 pending) with valid JSON; plan treats non-zero as query failure
  Parse stdout first and classify by bucket; a non-zero exit is a query failure only when stdout doesn't parse. Seed the parser test from real captured gh output for exit codes 0, 1 and 8 plus the no-checks message, so the fake matches the real binary.
- **PQ-3** [Important] `protected-list-enumeration-source` D5 protected list is hand-picked and misses .codex/config.toml; no stated source for the enumeration
  .codex/config.toml (Codex sandbox config, base.manifest:66-68) can loosen gates and isn't on the list. State the source the list is derived from (gate-relevant base.manifest rows plus local additions), sweep it, and note which entries exist only in the peer repos.
- **PQ-4** [Minor] `test-plan-enumerates-cases` M1/M2 test bullets list test cases in prose; compress to one strategy line per risky function
- **PQ-5** [Minor] `untrusted-input-in-workflow-script` pull_request_target run block must take PR values via env and fetch refs/pull/N/head, never interpolate head.ref
- **PQ-6** [Minor] `check-rerun-supersession` Confirm gh pr checks reports the latest guard run after a labeled re-run, not the earlier failure

## Round 2 — 2026-10-12T11:19:38-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — D1 changes PRMerge to take expectedHead with --match-head-commit (ghclient.go:149 today lacks it; ghlanding.go:163 durable path already pins).
- PQ-2 — addressed — Adapter parses stdout first; non-zero exit is a failure only when stdout is neither JSON nor the no-checks message; real-output fixtures.
- PQ-3 — addressed — D5 derives the list from three stated sources, includes .codex/config.toml (base.manifest:68), names peer-only entries and exclusions.
- PQ-4 — addressed — Test bullets compressed to per-function strategies.
- PQ-5 — addressed — Run block uses env: and fetches refs/pull/N/head; head.ref interpolation forbidden.
- PQ-6 — addressed — Live check confirms gh pr checks reports the latest guard run after labeling.

## Round 3 — 2026-10-12T11:20:23-07:00 (claude) — passed

**Protocol error:** no valid findings block — this round contributed no findings.

## Open findings

(none — every finding has been disposed)
