---
gate: boundary-review
issue: 329
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-10-12T17:25:29-07:00"
      dispatched: "2026-10-12T17:11:27-07:00"
      agent: claude
      findings:
        - id: BR-1
          severity: Important
          title: Cycle error can name a car downstream of the cycle, not on it (trainforecast.go:150)
          detail: 'After Kahn''s algorithm stalls, the first car in order with indeg>0 is reported, and that car may only be reachable from the cycle. Repro: trains Y: r#5 and X: r#3 → r#4, edges r#4 ⇒ r#5 and r#4 ⇒ r#3 give "cycle through r#5". D6 promises a car on the cycle. Fix: walk stalled predecessors until a node repeats, and add a test where the downstream car comes first.'
          family: diagnostic-names-wrong-entity
          round: 1
        - id: BR-2
          severity: Important
          title: D1's promised test (validate passes with an archived status done + deadline record) was not written
          detail: TestValidateInstance_ProjectDeadlineBan only checks the live corpus. Nothing pins that project validate --all and the changed-instance gate skip workshop/history/projects, so widening them would start rejecting the eight frozen done records with nothing to catch it.
          family: plan-promised-test-missing
          round: 1
        - id: BR-3
          severity: Minor
          title: Third copy of the WF_CALIB_LEDGER-else-VelocityPath rule, plus Abs+resolveWorkspace repeated per ref in projectIssueTerminalFn
          family: duplicated-path-resolution
          round: 1
        - id: BR-4
          severity: Minor
          title: CalibrateForecast divides by r.Estimate, but the eligibility check excludes only Estimate == 0, so a hand-edited negative estimate gives a negative multiplier
          family: untrusted-ledger-values
          round: 1
        - id: BR-5
          severity: Minor
          title: runProjectForecast writes warnings to os.Stderr instead of an injected writer
          family: injected-io-writer
          round: 1
        - id: BR-6
          severity: Minor
          title: D6 specifies --path P, but the forecast verb takes a positional file argument
          family: plan-code-drift
          round: 1
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 2
      timestamp: "2026-10-12T17:46:24-07:00"
      dispatched: "2026-10-12T17:29:58-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: carOnCycle walks stalled preds; reverting it turns TestForecastTrainsCycleAndEmpty red ("cycle through r#5"), restoring it turns it green.
          round: 2
        - id: BR-2
          disposition: addressed
          note: TestProjectValidationScopeExcludesArchive pins that validate --all and gateForPath both skip workshop/history/projects.
          round: 2
        - id: BR-3
          disposition: addressed
          note: close.go:1030 and convergereport now call calibrationLedgerPath; projectCheckout consolidates Abs+resolveWorkspace and resolves once per closure.
          round: 2
        - id: BR-4
          disposition: addressed
          note: Eligibility now requires Estimate > 0 and ConvergeHours >= 0; forecastcal_test rows 7 and 8 pin it via the N == 3 assertion.
          round: 2
        - id: BR-5
          disposition: addressed
          note: runProjectForecast takes an injected stderr (cmd.ErrOrStderr()); no os.Stderr remains in the added non-test code.
          round: 2
        - id: BR-6
          disposition: addressed
          note: Plan Revisions records D6's --path P becoming a positional [file]; it matches Use "forecast [file]".
          round: 2
      findings:
        - id: BR-7
          severity: Minor
          title: parseLedgerFloat accepts Inf/NaN, so a hand-edited Inf estimate zeroes the forecast multiplier
          detail: '2nd finding in this family. BR-4 filtered negatives at one reader. The rule covering the class (ARCH-SECURE): a ledger number becomes a typed value at the parse boundary, so parseLedgerFloat (ledger.go:271) should reject non-finite values (the row reads as absent) and downstream filters stop re-checking. Today an Inf estimate passes Estimate > 0, gives the row m = 0 and an infinite prior, and the forecast prints 0h or Inf with no warning.'
          family: untrusted-ledger-values
          round: 2
        - id: BR-8
          severity: Minor
          title: deadline ban accepts the sentinel string itself as a valid value
          detail: 'deadline?: "retired by ariadne#329 ..." rejects every other value, but a record carrying exactly that string validates. The trade-off for a readable cue diagnostic is documented, so this is a note only.'
          family: schema-ban-bypassable
          round: 2
      boundary: M1
      recipe: milestone-review
      isolation: checkout
      reviewed: 6d8a469e0a9daac749bfe29a126ef37ef05e25c6
      blocked: false
---

# Gate ledger — ariadne#329 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-10-12T17:25:29-07:00 (claude) — BLOCKED

### Raised

- **BR-1** [Important] `diagnostic-names-wrong-entity` Cycle error can name a car downstream of the cycle, not on it (trainforecast.go:150)
  After Kahn's algorithm stalls, the first car in order with indeg>0 is reported, and that car may only be reachable from the cycle. Repro: trains Y: r#5 and X: r#3 → r#4, edges r#4 ⇒ r#5 and r#4 ⇒ r#3 give "cycle through r#5". D6 promises a car on the cycle. Fix: walk stalled predecessors until a node repeats, and add a test where the downstream car comes first.
- **BR-2** [Important] `plan-promised-test-missing` D1's promised test (validate passes with an archived status done + deadline record) was not written
  TestValidateInstance_ProjectDeadlineBan only checks the live corpus. Nothing pins that project validate --all and the changed-instance gate skip workshop/history/projects, so widening them would start rejecting the eight frozen done records with nothing to catch it.
- **BR-3** [Minor] `duplicated-path-resolution` Third copy of the WF_CALIB_LEDGER-else-VelocityPath rule, plus Abs+resolveWorkspace repeated per ref in projectIssueTerminalFn
- **BR-4** [Minor] `untrusted-ledger-values` CalibrateForecast divides by r.Estimate, but the eligibility check excludes only Estimate == 0, so a hand-edited negative estimate gives a negative multiplier
- **BR-5** [Minor] `injected-io-writer` runProjectForecast writes warnings to os.Stderr instead of an injected writer
- **BR-6** [Minor] `plan-code-drift` D6 specifies --path P, but the forecast verb takes a positional file argument

## Round 2 — 2026-10-12T17:46:24-07:00 (claude) — passed

### Disposed

- BR-1 — addressed — carOnCycle walks stalled preds; reverting it turns TestForecastTrainsCycleAndEmpty red ("cycle through r#5"), restoring it turns it green.
- BR-2 — addressed — TestProjectValidationScopeExcludesArchive pins that validate --all and gateForPath both skip workshop/history/projects.
- BR-3 — addressed — close.go:1030 and convergereport now call calibrationLedgerPath; projectCheckout consolidates Abs+resolveWorkspace and resolves once per closure.
- BR-4 — addressed — Eligibility now requires Estimate > 0 and ConvergeHours >= 0; forecastcal_test rows 7 and 8 pin it via the N == 3 assertion.
- BR-5 — addressed — runProjectForecast takes an injected stderr (cmd.ErrOrStderr()); no os.Stderr remains in the added non-test code.
- BR-6 — addressed — Plan Revisions records D6's --path P becoming a positional [file]; it matches Use "forecast [file]".

### Raised

- **BR-7** [Minor] `untrusted-ledger-values` parseLedgerFloat accepts Inf/NaN, so a hand-edited Inf estimate zeroes the forecast multiplier
  2nd finding in this family. BR-4 filtered negatives at one reader. The rule covering the class (ARCH-SECURE): a ledger number becomes a typed value at the parse boundary, so parseLedgerFloat (ledger.go:271) should reject non-finite values (the row reads as absent) and downstream filters stop re-checking. Today an Inf estimate passes Estimate > 0, gives the row m = 0 and an infinite prior, and the forecast prints 0h or Inf with no warning.
- **BR-8** [Minor] `schema-ban-bypassable` deadline ban accepts the sentinel string itself as a valid value
  deadline?: "retired by ariadne#329 ..." rejects every other value, but a record carrying exactly that string validates. The trade-off for a readable cue diagnostic is documented, so this is a note only.

## Open findings

- **BR-7** [Minor] `untrusted-ledger-values` parseLedgerFloat accepts Inf/NaN, so a hand-edited Inf estimate zeroes the forecast multiplier
- **BR-8** [Minor] `schema-ban-bypassable` deadline ban accepts the sentinel string itself as a valid value
