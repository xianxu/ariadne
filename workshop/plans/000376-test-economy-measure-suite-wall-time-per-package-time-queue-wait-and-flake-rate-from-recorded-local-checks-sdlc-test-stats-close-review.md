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

---

## Re-review — 2026-10-14T23:12:22-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 376 — Test economy: measure suite wall time, per-package time, queue wait and flake rate from recorded local checks (sdlc test-stats) |
| repo | ariadne |
| issue file | workshop/issues/000376-test-economy-measure-suite-wall-time-per-package-time-queue-wait-and-flake-rate-from-recorded-local-checks-sdlc-test-stats.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..d7770d349b0e78ed4edc690c375179b330202473 |
| command | sdlc close --issue 376 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T23:12:22-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior findings are addressed, with regression tests that fail when the behavioral fixes are removed. The pinned head still accepts conflicting JSON fields as successful-run evidence, omits failed-package timings from text logs, and does not enforce the planned parser line limit. These findings block closing the issue.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      JSON pass/fail/malformed fixtures pass on head. Removing JSON event handling makes TestSuiteLogFactsGoTestJSON and TestSuiteLogFactsMalformedGoTestJSONWarnsWithoutInventingData fail.
  - id: BR-2
    disposition: addressed
    note: |
      TestSuiteStatsCommandSmokeRecordDoesNotHideLegacySuite passes on head and fails with the prior unconditional SHA deduplication restored.
  - id: BR-3
    disposition: addressed
    note: |
      The before/after plan diff corrects SuiteAttempt to cmd/sdlc/teststats.go, matching its definition at line 11. The inventory explicitly identifies PURE entities and INTEGRATION persistence, and the close-review Revisions entry records the correction.
  - id: BR-4
    disposition: addressed
    note: |
      Publication temps now carry owner IDs. TestPruneSuiteAttemptsReclaimsOnlyDeadOwnerPublicationTemps protects active writers and reclaims dead-owner residue; restoring the prior store implementation makes it fail.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: persisted-evidence-canonical-decoding
    title: |
      Conflicting JSON field casing can turn a failed attempt into success
    detail: |
      cmd/sdlc/teststats_store.go:75 checks duplicate keys case-sensitively, but the struct decode at line 102 accepts case-insensitive field matches. A valid failed record containing "exit":2 followed by "Exit":0 is accepted with Exit=0, despite DisallowUnknownFields. TestBoundary376CaseAlias reproduced this on head. ARCH-SECURE: this admitted record can become false successful-run and flake-confirmation evidence. Reject noncanonical field names before decoding, and test aliases and conflicting keys across outcome/provenance fields.
  - id: new
    severity: Important
    kind: defect
    family: contracted-log-format-coverage
    title: |
      Plain-text failed-package summaries lose their measured duration
    detail: |
      cmd/sdlc/teststats_parse.go:86 extracts package timing only from ok summaries; line 97 recognizes FAIL without extracting its timing. TestBoundary376FailedPackageTiming reproduced "FAIL example.com/slow 90.0s" yielding no units and no warning, excluding slow failing runs from package p90. ARCH-PURPOSE: this is the 2nd finding in family contracted-log-format-coverage. Earlier work fixed an instance. State and enforce the rule that every supported format preserves available package durations for both passing and failing outcomes; enumerate text, JSON, and shard outcome fixtures and sweep that matrix in this round.
  - id: new
    severity: Important
    kind: defect
    family: declared-parser-resource-bounds
    title: |
      The planned 1 MiB log-line limit is not enforced
    detail: |
      cmd/sdlc/teststats_parse.go:47-52 passes the entire log to failingTests and splits it without checking line lengths. The plan at workshop/plans/000376-test-stats-plan.md:37 requires a 1 MiB line limit and unavailable facts plus a warning for over-limit logs. TestBoundary376LineLimit supplied an oversized line and received a package measurement without warnings. ARCH-CONSTRAINTS: enforce the bound before either parser consumes the input, using a shared path for recorder and legacy reads, and add boundary-size regressions.
```

1. **Strengths**

   - Immutable failed logs survive same-head retries and failed status publication.
   - Pure aggregation tests pin quantiles, coverage, chronological flake classification, and input-order independence.
   - README and atlas updates cover the new command and its partial-history semantics.

2. **Critical findings**

   - Conflicting JSON aliases can fabricate success. Reject noncanonical keys at the persisted-input boundary.

3. **Important findings**

   - Preserve timings from failed text-package outcomes.
   - Enforce the documented parser line bound before failure extraction and timing parsing.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused parser, aggregation, command, storage, local-check, identity, and hashing tests passed. The three new reproduction tests failed on head. Behavioral fixes for BR-1, BR-2, and BR-4 were removed in disposable copies and their regressions failed as expected. `git diff --check` passed. I did not rerun the full suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** recorder and legacy reader share parsing.
   - **ARCH-PURE — pass:** calculation and parser tests run without IO.
   - **ARCH-PURPOSE — flag:** failed-package measurements remain incomplete.
   - **ARCH-MOCK — pass:** integration tests use disposable repositories and existing dependency seams.
   - **ARCH-CONSTRAINTS — flag:** total intake caps exist; the promised line cap is missing.
   - **ARCH-SECURE — flag:** case-insensitive decoding bypasses duplicate-field protection.
   - **ARCH-ORDER — pass:** publication precedes status posting; staged ordering tests protect active evidence.
   - **ARCH-FUNERAL — pass:** prior publication-temp cleanup now has ownership and regression evidence.

7. **Plan revision recommendations**

   Add `## Revisions` entries defining canonical persisted-field decoding and the complete format-by-outcome parsing matrix. Record the implementation and regression coverage for the existing line-limit requirement.
