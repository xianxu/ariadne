---
gate: boundary-review
issue: 338
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T15:40:08-07:00"
      dispatched: "2026-10-12T15:23:21-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Minor
          title: Fake reviewer exits silently when PLANNING_SPINS is unset instead of failing loudly
          detail: 'planningreview_test.go:73 turns a missing PLANNING_SPINS into an immediate exit 1 after writing ready, which is the intermittent failure this issue chased. Add a guard such as : "${PLANNING_SPINS:?}" (and the same for PLANNING_BARRIER) so a partial env fails every time.'
          family: fixture-env-fails-silently
          round: 1
        - id: BR-2
          severity: Minor
          title: Log does not record the 10-run and make test results that Done-when requires
          detail: The Plan's Verify item is checked, but the Log has no line with the run count or the make test outcome.
          family: done-when-evidence-unrecorded
          round: 1
      recipe: milestone-review
      isolation: fenced
      reviewed: 436a1f42e317dc7fd1fe0fb48ea9bf5c4aacec9b
      blocked: false
    - "n": 2
      timestamp: "2026-10-12T15:52:59-07:00"
      dispatched: "2026-10-12T15:49:37-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: 'planningreview_test.go:70 adds : "${PLANNING_BARRIER:?}" "${PLANNING_SPINS:?}" before ready is written; I confirmed it in sh, and the competing test still passes (-count=2).'
          round: 2
        - id: BR-2
          disposition: addressed
          note: 'The issue Log''s Evidence line records 3/11 failing before the fix, 10/10 passing after it (-count=10), and a make test run whose only failure is pre-existing #340.'
          round: 2
      recipe: fix-review
      isolation: fenced
      reviewed: de37f528cd778ef486f79f01faa4804515d28ca2
      blocked: false
---

# Gate ledger — ariadne#338 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T15:40:08-07:00 (claude) — passed

### Raised

- **BR-1** [Minor] `fixture-env-fails-silently` Fake reviewer exits silently when PLANNING_SPINS is unset instead of failing loudly
  planningreview_test.go:73 turns a missing PLANNING_SPINS into an immediate exit 1 after writing ready, which is the intermittent failure this issue chased. Add a guard such as : "${PLANNING_SPINS:?}" (and the same for PLANNING_BARRIER) so a partial env fails every time.
- **BR-2** [Minor] `done-when-evidence-unrecorded` Log does not record the 10-run and make test results that Done-when requires
  The Plan's Verify item is checked, but the Log has no line with the run count or the make test outcome.

## Round 2 — 2026-10-12T15:52:59-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — planningreview_test.go:70 adds : "${PLANNING_BARRIER:?}" "${PLANNING_SPINS:?}" before ready is written; I confirmed it in sh, and the competing test still passes (-count=2).
- BR-2 — addressed — The issue Log's Evidence line records 3/11 failing before the fix, 10/10 passing after it (-count=10), and a make test run whose only failure is pre-existing #340.

## Open findings

(none — every finding has been disposed)
