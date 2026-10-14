---
gate: boundary-review
issue: 351
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-13T17:17:44-07:00"
      dispatched: "2026-10-13T17:16:37-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Sibling empty-array expansion under set -u at construct/local/introspect/scripts/introspect-extract.sh:102
          detail: SEGMENTS=("${TRIMMED[@]}") dies on bash 3.2 when --limit 0 leaves TRIMMED empty; same class as the folded-in line 108 fix. A shelllint rule would cover the class.
          family: bash32-empty-array-set-u
          round: 1
        - id: BR-2
          severity: Minor
          title: Folded-in bash 3.2 pids fix in parallel-checks.sh has no automated regression test
          detail: Verified only by a manual stub run per the Log; also MAX_PARALLEL_CHECKS=0 still spins and hits empty "${pids[@]}" at line 105 (pre-existing).
          family: shell-script-regression-coverage
          round: 1
      recipe: milestone-review
      isolation: checkout
      reviewed: a77603f1f6dfa5b2d8e2d1b5fe640416b34e7865
      blocked: false
---

# Gate ledger — ariadne#351 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-13T17:17:44-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `bash32-empty-array-set-u` Sibling empty-array expansion under set -u at construct/local/introspect/scripts/introspect-extract.sh:102
  SEGMENTS=("${TRIMMED[@]}") dies on bash 3.2 when --limit 0 leaves TRIMMED empty; same class as the folded-in line 108 fix. A shelllint rule would cover the class.
- **BR-2** [Minor] `shell-script-regression-coverage` Folded-in bash 3.2 pids fix in parallel-checks.sh has no automated regression test
  Verified only by a manual stub run per the Log; also MAX_PARALLEL_CHECKS=0 still spins and hits empty "${pids[@]}" at line 105 (pre-existing).

## Open findings

- **BR-1** [Minor] `bash32-empty-array-set-u` Sibling empty-array expansion under set -u at construct/local/introspect/scripts/introspect-extract.sh:102
- **BR-2** [Minor] `shell-script-regression-coverage` Folded-in bash 3.2 pids fix in parallel-checks.sh has no automated regression test
