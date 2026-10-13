---
gate: boundary-review
issue: 342
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T16:04:27-07:00"
      dispatched: "2026-10-13T16:03:39-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
          detail: '`TMPDIR=/x mktemp -d`, `{ mktemp -d; }`, `sudo mktemp -d` and `time mktemp -d` report no finding; commandPosition (pkg/shelllint/mktemp.go:89) should accept these prefixes.'
          family: lint-command-position-coverage
          round: 1
        - id: BR-2
          severity: Minor
          title: cd-into-substitution rule does not cover pushd "$(mktemp ...)"
          detail: pushd has the same empty-path failure mode as cd; extend the cdIntoMktemp regex at mktemp.go:26 to (cd|pushd).
          family: lint-command-position-coverage
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: e3ff505c5fb11818cf82b9d98fe20c055668b957
      blocked: false
---

# Gate ledger — ariadne#342 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T16:04:27-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `lint-command-position-coverage` MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
  `TMPDIR=/x mktemp -d`, `{ mktemp -d; }`, `sudo mktemp -d` and `time mktemp -d` report no finding; commandPosition (pkg/shelllint/mktemp.go:89) should accept these prefixes.
- **BR-2** [Minor] `lint-command-position-coverage` cd-into-substitution rule does not cover pushd "$(mktemp ...)"
  pushd has the same empty-path failure mode as cd; extend the cdIntoMktemp regex at mktemp.go:26 to (cd|pushd).

## Open findings

- **BR-1** [Minor] `lint-command-position-coverage` MktempFindings misses mktemp after prefix words/assignments ({, sudo, time, VAR=x)
- **BR-2** [Minor] `lint-command-position-coverage` cd-into-substitution rule does not cover pushd "$(mktemp ...)"
