# Boundary Review — ariadne#376 (whole-issue close)

| field | value |
|-------|-------|
| issue | 376 — Test economy: measure suite wall time, per-package time, queue wait and flake rate from recorded local checks (sdlc test-stats) |
| repo | ariadne |
| issue file | workshop/issues/000376-test-economy-measure-suite-wall-time-per-package-time-queue-wait-and-flake-rate-from-recorded-local-checks-sdlc-test-stats.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..44c39478add6be7e5f8c56ace079b462926dea24 |
| command | sdlc close --issue 376 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T22:22:10-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation has a well-separated calculation layer, useful persistence tests, and both required documentation updates. Two reproduced correctness bugs block shipping: JSON suite logs lose their measurements, and non-suite records suppress legacy suite evidence. Crash cleanup and the Core concepts inventory also need correction.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: contracted-log-format-coverage
    title: |
      The parser silently ignores the Spec's go test -json format
    detail: |
      cmd/sdlc/teststats_parse.go:34 recognizes only plain-text summaries; JSON package/test events produce empty failures and units without warnings. TestBoundary376JSON reproduced this on the pinned head using valid test and package fail events. ARCH-PURPOSE: support the contracted JSON format, including package timings and package-qualified failure identities, with regression fixtures for pass, fail, and malformed events.
  - id: new
    severity: Critical
    kind: defect
    family: evidence-deduplication-scope
    title: |
      A non-suite attempt hides the legacy suite on the same SHA
    detail: |
      cmd/sdlc/teststats_cmd.go:138 marks every attempt SHA as known before filtering by check name. Since runLocalCheck records non-suite checks too, a smoke record suppresses an existing suite.log; aggregation then discards the smoke record. TestBoundary376Legacy reproduced an empty repository report with no warning despite an available suite log. ARCH-PURPOSE: deduplicate by the covered check and SHA, and add a mixed-check regression test.
  - id: new
    severity: Critical
    kind: defect
    family: core-concept-inventory-accuracy
    title: |
      The Core concepts table names the wrong location for SuiteAttempt
    detail: |
      workshop/plans/000376-test-stats-plan.md:15 places SuiteAttempt in teststats_parse.go, but its definition is cmd/sdlc/teststats.go:11. This contradicts the required Core concepts cross-check. Correct the location and append a Revisions entry; make PURE and INTEGRATION classifications explicit in the inventory. This is a documentation correction and needs no wording-presence test.
  - id: new
    severity: Important
    kind: defect
    family: crash-artifact-lifecycle
    title: |
      Crash-left attempt publication files have no removal path
    detail: |
      cmd/sdlc/teststats_store.go:225 creates attempts/.attempt-* files whose cleanup depends on a defer. Process death bypasses that defer, while pruneSuiteAttempts skips every non-.json file at line 352 and only otherwise sweeps inflight directories. ARCH-FUNERAL: associate publication temporaries with attempt ownership and reclaim aged dead-owner residue, with a crash fixture that also proves active writers remain protected.
```

1. **Strengths**

   - Parser and aggregation tests exercise pure logic directly, including quantiles, missing coverage, and completion ordering.
   - Persistence tests verify immutable failure logs survive retries and failed status posting.
   - The decoder rejects missing outcomes, duplicate keys, malformed records, and unsafe paths.
   - README and atlas updates cover the new command; both documentation gates are satisfied.

2. **Critical findings**

   Three listed above: contracted JSON parsing, check-scoped deduplication, and the required Core concepts correction.

3. **Important findings**

   Crash-left publication temporaries need an ownership-aware cleanup path.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused parser, aggregation, command, decoder, publication, pruning, runner, and hashing tests passed. Both additional regression checks failed against the pinned head using temporary Go overlays; those overlays were removed. `git diff --check` passed, and the checkout remains unchanged. I did not rerun the full suite or treat the issue log’s claimed suite result as independently verified evidence.

6. **Architectural notes for upcoming work**

   | Marker | Result |
   | --- | --- |
   | ARCH-DRY | Pass: shared parser and existing failure extractor are reused. |
   | ARCH-PURE | Pass: calculations remain independent of IO. |
   | ARCH-PURPOSE | Flag: parsing and deduplication lose required evidence. |
   | ARCH-MOCK | Pass: runner tests use portable repositories and existing status/lock seams. |
   | ARCH-CONSTRAINTS | Pass for tested record intake limits; log-envelope claims need broader verification. |
   | ARCH-SECURE | Pass: persisted records are validated before aggregation. |
   | ARCH-ORDER | Pass: publication precedes status posting; tests cover retry preservation and intermediate publication/pruning order. |
   | ARCH-FUNERAL | Flag: publication temporaries escape crash cleanup. |

7. **Plan revision recommendations**

   Append `## Revisions` entries correcting the entity inventory and specifying publication-temporary ownership and reclamation. Add explicit JSON-event and mixed-check deduplication regression cases to the existing tasks.
