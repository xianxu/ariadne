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

## Open findings

- **BR-1** [Important] `diagnostic-names-wrong-entity` Cycle error can name a car downstream of the cycle, not on it (trainforecast.go:150)
- **BR-2** [Important] `plan-promised-test-missing` D1's promised test (validate passes with an archived status done + deadline record) was not written
- **BR-3** [Minor] `duplicated-path-resolution` Third copy of the WF_CALIB_LEDGER-else-VelocityPath rule, plus Abs+resolveWorkspace repeated per ref in projectIssueTerminalFn
- **BR-4** [Minor] `untrusted-ledger-values` CalibrateForecast divides by r.Estimate, but the eligibility check excludes only Estimate == 0, so a hand-edited negative estimate gives a negative multiplier
- **BR-5** [Minor] `injected-io-writer` runProjectForecast writes warnings to os.Stderr instead of an injected writer
- **BR-6** [Minor] `plan-code-drift` D6 specifies --path P, but the forecast verb takes a positional file argument
