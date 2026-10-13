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
    - "n": 3
      timestamp: "2026-10-12T18:09:51-07:00"
      dispatched: "2026-10-12T18:03:50-07:00"
      agent: claude
      findings:
        - id: BR-9
          severity: Important
          title: M2 has no Revisions entry for unplanned scope, and the D10 patch artifact is not recorded anywhere
          detail: '2nd finding in this family. Rule: at every milestone close, match each checklist item and D-decision to a delivered artifact and record any difference in Revisions, in both directions. Promised but missing: D10 says to write a patch file for the TL, but no patch is in the range and the Log gives no path, so Done-when 4 cannot be checked. Delivered but unplanned: train-only matching (D5 only said fields are filled for matched files), the close-tick RowRef skip, the project validate section check, and the claimant-ownership.md migration, none recorded in Revisions.'
          family: plan-code-drift
          round: 3
        - id: BR-10
          severity: Important
          title: Prose descriptions of the project match rule still say marker-only after train-only matching was added
          detail: The DiscoverByIssueRef doc comment (discover.go:186-192) and helptext/resolve.md:54-60 both say a match means the body contains the [repo#id] marker. They omit train-car matches and the new train/car JSON fields. Update both, or point both at DiscoverInRoots as the single source of the rule.
          family: stale-match-rule-restatement
          round: 3
        - id: BR-11
          severity: Minor
          title: TestResolveRun_KindProjectJSONCarriesTrain does not check that t.md is in the results
          detail: The loop only checks matching entries, so the test passes if the train-only file is missing from res.Files.
          family: vacuous-test-assertion
          round: 3
        - id: BR-12
          severity: Minor
          title: project validate treats every scaffold section as required
          detail: missingProjectSections iterates over the scaffold Sections(), so a future optional scaffold section would make every existing project fail. Consider a required flag in the model.
          family: scaffold-as-requirement
          round: 3
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      blocked: true
    - "n": 4
      timestamp: "2026-10-12T18:16:09-07:00"
      dispatched: "2026-10-12T18:12:41-07:00"
      agent: claude
      dispose:
        - id: BR-9
          disposition: addressed
          note: tl-plan.md Revisions 2026-10-12 M2 entry matches D5/D10 to deliveries and lists the four unplanned items; patch committed at workshop/plans/000329-ops-conformance.patch.
          round: 4
        - id: BR-10
          disposition: addressed
          note: discover.go:187-206 points at DiscoverInRoots as the single rule; helptext/resolve.md:54-58 names row marker OR train car plus train/car JSON; sweep found no remaining marker-only restatement.
          round: 4
        - id: BR-11
          disposition: addressed
          note: projectfind_test.go now asserts sawTrain, so a missing t.md fails the test.
          round: 4
        - id: BR-12
          disposition: addressed
          note: 'Section.Optional + #ScaffoldSection.optional; TestMissingSectionsSkipsOptional fails without the !sec.Optional guard in missingSections.'
          round: 4
      findings:
        - id: BR-13
          severity: Minor
          title: D10 Revisions omit that the ops patch leaves couch-broadcast-web at status defined and drops its Estimate text
          detail: 'This is the 3rd finding in family plan-code-drift. The covering rule is the one BR-9 stated: at milestone close, every D-decision clause is matched to the artifact in both directions. D10 says the patch "sets status: executing", but only ariadne-robustness-1 is changed; couch-broadcast-web stays defined (its trains are empty) and its "Not yet. Sized after..." line is deleted rather than kept as Sizing. Record both in the D10 Revisions bullet.'
          family: plan-code-drift
          round: 4
        - id: BR-14
          severity: Minor
          title: The ops conformance patch already fails git apply --check against current ops; the appended Decisions hunks no longer match the grown Log tails
          detail: Run against ~/workspace/ops, git apply --check fails at ariadne-robustness-1:687 and couch-broadcast-web:754, and succeeds with -C1. Tell the TL to apply with --3way or reduced context, or regenerate the patch at handoff.
          family: stale-handoff-artifact
          round: 4
      boundary: M2
      recipe: milestone-review
      isolation: checkout
      reviewed: 2e5ef5a3c43b6528d82b14a8fbadd8a78dcaa9a2
      blocked: false
    - "n": 5
      timestamp: "2026-10-12T18:40:34-07:00"
      dispatched: "2026-10-12T18:26:43-07:00"
      agent: claude
      dispose:
        - id: BR-7
          disposition: addressed
          note: ledger.go:276 rejects Inf/NaN at the parse boundary; forecastcal_test.go:53-67 fails without the guard.
          round: 5
        - id: BR-8
          disposition: withdrawn
          note: Accepted, documented trade-off for a readable cue diagnostic; Minor note only, nothing to fix.
          round: 5
        - id: BR-13
          disposition: addressed
          note: Plan Revisions (after M2) record couch stays defined with an empty trains table and keeps Sizing; script output confirms both.
          round: 5
        - id: BR-14
          disposition: addressed
          note: Handoff is now ops-conform.py; run against current ops, both outputs pass sdlc project validate. The saved patch is stale again but is labeled preview-only.
          round: 5
      recipe: milestone-review
      isolation: checkout
      reviewed: 6c57938b885a945dd12f9269bc7b70dcf40bab52
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

## Round 3 — 2026-10-12T18:09:51-07:00 (claude) — BLOCKED

### Raised

- **BR-9** [Important] `plan-code-drift` M2 has no Revisions entry for unplanned scope, and the D10 patch artifact is not recorded anywhere
  2nd finding in this family. Rule: at every milestone close, match each checklist item and D-decision to a delivered artifact and record any difference in Revisions, in both directions. Promised but missing: D10 says to write a patch file for the TL, but no patch is in the range and the Log gives no path, so Done-when 4 cannot be checked. Delivered but unplanned: train-only matching (D5 only said fields are filled for matched files), the close-tick RowRef skip, the project validate section check, and the claimant-ownership.md migration, none recorded in Revisions.
- **BR-10** [Important] `stale-match-rule-restatement` Prose descriptions of the project match rule still say marker-only after train-only matching was added
  The DiscoverByIssueRef doc comment (discover.go:186-192) and helptext/resolve.md:54-60 both say a match means the body contains the [repo#id] marker. They omit train-car matches and the new train/car JSON fields. Update both, or point both at DiscoverInRoots as the single source of the rule.
- **BR-11** [Minor] `vacuous-test-assertion` TestResolveRun_KindProjectJSONCarriesTrain does not check that t.md is in the results
  The loop only checks matching entries, so the test passes if the train-only file is missing from res.Files.
- **BR-12** [Minor] `scaffold-as-requirement` project validate treats every scaffold section as required
  missingProjectSections iterates over the scaffold Sections(), so a future optional scaffold section would make every existing project fail. Consider a required flag in the model.

## Round 4 — 2026-10-12T18:16:09-07:00 (claude) — passed

### Disposed

- BR-9 — addressed — tl-plan.md Revisions 2026-10-12 M2 entry matches D5/D10 to deliveries and lists the four unplanned items; patch committed at workshop/plans/000329-ops-conformance.patch.
- BR-10 — addressed — discover.go:187-206 points at DiscoverInRoots as the single rule; helptext/resolve.md:54-58 names row marker OR train car plus train/car JSON; sweep found no remaining marker-only restatement.
- BR-11 — addressed — projectfind_test.go now asserts sawTrain, so a missing t.md fails the test.
- BR-12 — addressed — Section.Optional + #ScaffoldSection.optional; TestMissingSectionsSkipsOptional fails without the !sec.Optional guard in missingSections.

### Raised

- **BR-13** [Minor] `plan-code-drift` D10 Revisions omit that the ops patch leaves couch-broadcast-web at status defined and drops its Estimate text
  This is the 3rd finding in family plan-code-drift. The covering rule is the one BR-9 stated: at milestone close, every D-decision clause is matched to the artifact in both directions. D10 says the patch "sets status: executing", but only ariadne-robustness-1 is changed; couch-broadcast-web stays defined (its trains are empty) and its "Not yet. Sized after..." line is deleted rather than kept as Sizing. Record both in the D10 Revisions bullet.
- **BR-14** [Minor] `stale-handoff-artifact` The ops conformance patch already fails git apply --check against current ops; the appended Decisions hunks no longer match the grown Log tails
  Run against ~/workspace/ops, git apply --check fails at ariadne-robustness-1:687 and couch-broadcast-web:754, and succeeds with -C1. Tell the TL to apply with --3way or reduced context, or regenerate the patch at handoff.

## Round 5 — 2026-10-12T18:40:34-07:00 (claude) — passed

### Disposed

- BR-7 — addressed — ledger.go:276 rejects Inf/NaN at the parse boundary; forecastcal_test.go:53-67 fails without the guard.
- BR-8 — withdrawn — Accepted, documented trade-off for a readable cue diagnostic; Minor note only, nothing to fix.
- BR-13 — addressed — Plan Revisions (after M2) record couch stays defined with an empty trains table and keeps Sizing; script output confirms both.
- BR-14 — addressed — Handoff is now ops-conform.py; run against current ops, both outputs pass sdlc project validate. The saved patch is stale again but is labeled preview-only.

## Open findings

(none — every finding has been disposed)
