---
gate: boundary-review
issue: 231
id_prefix: BR
rounds:
    - "n": 1
      timestamp: "2026-09-17T20:53:40-07:00"
      agent: sdlc
      findings:
        - id: BR-1
          severity: Minor
          title: Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
          detail: |-
            Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.
            (carried from plan-quality PQ-8, deferred to the boundary review)
          family: test-prose-enumeration
          round: 1
        - id: BR-2
          severity: Minor
          title: Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
          detail: |-
            Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan's own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.
            (carried from plan-quality PQ-9, deferred to the boundary review)
          family: record-roundtrips-every-reader
          round: 1
        - id: BR-3
          severity: Minor
          title: Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
          detail: |-
            Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.
            (carried from plan-quality PQ-10, deferred to the boundary review)
          family: acceptance-oracle-freshness
          round: 1
      boundary: '*'
      no_cap: true
      blocked: false
    - "n": 2
      timestamp: "2026-09-17T20:53:40-07:00"
      agent: claude
      findings:
        - id: BR-4
          severity: Important
          title: decideChangeCodeFlow's PlanItemsBody routing is unpinned — not in planItemBodySources, no fenced-Mx test
          detail: changecode_flow.go:48 hands a plan body to the exempt helper issue.MilestonesInPlanOrder, whose exemption says callers are covered by planItemBodySources (commitpathspec_guard_test.go:341), which lists only close.go. Reverting to PlanSectionBody in a scratch copy left every Flow/Guard/PlanItem/ChangeCode test green. Add a fenced-Mx case to TestDecideChangeCodeFlow (must stay quick) and sweep all MilestonesInPlanOrder callers (ARCH-PURPOSE).
          family: hand-listed-guard-incomplete
          round: 2
        - id: BR-5
          severity: Important
          title: change-code writes the flow record before its gates, so a refused run leaves a sticky full/inferred
          detail: applyChangeCodeFlow runs at changecode.go:137 before the gate loop; a structural/plan-quality/estimate refusal exits with flow full/inferred written (uncommitted), and Decide keeps a recorded full, so removing the plan or Mx rows and re-running stays full. The plan's ARCH-ORDER table says REWORK writes nothing. Decide before the gates, write after they pass (before syncIssue); test that a refused run leaves the file byte-identical (ARCH-ORDER).
          family: state-persisted-on-refused-transition
          round: 2
        - id: BR-6
          severity: Minor
          title: M1 docs describe close-time flow behaviour (small-diff review, upgrade, Done-when refusal) that M2 hasn't built
          detail: helptext/change-code.md THE FLOW, helptext/issue.md flow field, AGENTS.base.md section 2, atlas issue-lifecycle item 6 and flowInfoLine state it in the present tense; at HEAD close ignores the flow entirely. Acceptable only if M2 lands before any merge; note it in the Log.
          family: doc-claim-contradicts-code
          round: 2
        - id: BR-7
          severity: Minor
          title: flowReason re-derives Decide's rule order instead of Decide returning the rule that fired
          detail: changecode_flow.go:66-81 mirrors flow.go Decide rules; a reorder in Decide silently mislabels the reason (ARCH-DRY).
          family: decision-logic-restated
          round: 2
        - id: BR-8
          severity: Minor
          title: Rewriting a block-form hand-edited flow record orphans its child lines into invalid frontmatter
          detail: SetField replaces only the flow line; a probe produced flow full/inferred followed by indented kind/provenance lines, while the warning says the record is being rewritten (ARCH-SECURE).
          family: untrusted-input-rewrite
          round: 2
        - id: BR-9
          severity: Minor
          title: FuzzFromFrontmatter does not assert that the error path resolves to full
          detail: The plan names "every error path resolves to full" as the fuzz oracle; the body returns on error without checking Kind == Full.
          family: test-oracle-weaker-than-plan
          round: 2
        - id: BR-10
          severity: Minor
          title: issue.cue carries unrelated cue-fmt whitespace realignment of categories, Status and Transition
          family: unrelated-churn
          round: 2
      boundary: M1
      blocked: true
    - "n": 3
      timestamp: "2026-09-17T21:07:39-07:00"
      agent: claude
      dispose:
        - id: BR-1
          disposition: addressed
          note: Plan names every test. FuzzFromFrontmatter plus the fenced-heading case in TestContractHashes cover M1's parsers; FuzzParseSurfaces is M2 scope; keeping case lists is a recorded decision.
          round: 3
        - id: BR-2
          disposition: addressed
          note: Format quotes hashes with %q (flow.go:66-72); scratch revert to %s turns TestFlowRoundTrip red on 12345678; validate_test covers quoted-valid and unquoted-invalid through real cue.
          round: 3
        - id: BR-3
          disposition: addressed
          note: Spec section 4 and Done-when say quick-flow and the Flow bullet carries spec/done; old names survive only in dated Revisions.
          round: 3
        - id: BR-4
          disposition: addressed
          note: Scratch revert of changecode_flow.go:50 or close.go:1719 to PlanSectionBody turns TestPlanItemReadersUsePlanItemsBody red; the fenced-Mx case in TestDecideChangeCodeFlow also goes red.
          round: 3
        - id: BR-5
          disposition: addressed
          note: Moving recordChangeCodeFlow before the gate loop turns TestRunChangeCodeRecordsFlowAfterGates red. The fix opened a stale-snapshot window, raised separately.
          round: 3
        - id: BR-6
          disposition: addressed
          note: The 2026-09-17 "M1 built" Log entry records that the present-tense close docs hold only because M2 lands before merge.
          round: 3
        - id: BR-7
          disposition: addressed
          note: Decide returns the Rule that fired, flowReason is gone, and TestDecide asserts the rule for each case.
          round: 3
        - id: BR-8
          disposition: addressed
          note: SetField drops the indented block continuation (frontmatter.go:73-99), pinned by TestSetFieldReplacesBlockValue.
          round: 3
        - id: BR-9
          disposition: addressed
          note: Scratch revert of FromFrontmatter's error value to Flow{} turns FuzzFromFrontmatter's seed corpus red.
          round: 3
        - id: BR-10
          disposition: addressed
          note: The issue.cue diff is now 17 pure additions, with no realignment.
          round: 3
      findings:
        - id: BR-11
          severity: Important
          title: change-code writes the flow record from issue text read before the gates, clobbering edits made while plan-quality ran
          detail: 'recordChangeCodeFlow (changecode_flow.go:100-108) writes d.content, built at changecode.go:138 from the step-2 read, after the multi-minute gate loop. A scratch probe (write, report, edit the file on disk, record) lost the edit, and syncIssue then commits it and, on main, pushes it. It fires on every issue''s first successful run, because the record is always new. The repo lock does not cover editor or agent edits; close handles the same hazard by validating its snapshot after the review (close.go:1218-1231). Fix: re-read issuePath at record time and apply SetField to the fresh frontmatter, recomputing quick hashes over the fresh body, or refuse as close does. Add the probe as a regression test (ARCH-ORDER, second actor on the same state).'
          family: write-from-stale-snapshot
          round: 3
        - id: BR-12
          severity: Minor
          title: flow.Parse accepts spec "" and done "" that cue Flow rejects, contradicting Parse's own doc comment
          detail: 'This is the 2nd finding in record-roundtrips-every-reader. Rule: the Go codec and the cue model are two readers of one record and must accept exactly the same set of values. Fix the rule, not the instance: a shared testdata corpus of records with expected accept or reject, asserted by both flow_test (Parse) and cmd/vocabulary validate_test (cue). That also catches #233 adding config to one reader only. Also reject empty hashes in Parse. Prevalence: 1 divergence among 3 probed edge shapes.'
          family: record-roundtrips-every-reader
          round: 3
        - id: BR-13
          severity: Minor
          title: helptext/estimate.md:5 still says change-code refuses unless the Estimate block reconciles, with no flow qualification
          detail: 'This is the 2nd finding in doc-claim-contradicts-code. Rule: when a change makes a gate conditional, every sentence asserting that gate''s action is flow-qualified, and the enumeration comes from grepping the gate''s action verbs, not from a hand-copied surface list. PQ-1''s own regex matched estimate.md through estimate_hours, but the list dropped it. Measured: git grep over helptext, atlas, AGENTS.base.md and construct for change-code with refus, requir, demand, parses or asks-for, plus "estimate gate" and "plan-quality gate", gives about 15 hits; 1 is unqualified and false.'
          family: doc-claim-contradicts-code
          round: 3
        - id: BR-14
          severity: Minor
          title: The no-frontmatter branch of decideChangeCodeFlow decides full itself, bypassing Decide and its pin validation
          detail: 'This is the 2nd finding in decision-logic-restated. changecode_flow.go:35-39 returns full with an ad hoc Rule string and empty provenance. An invalid --flow fast, or a --flow quick pin, is silently ignored, and the info line prints "flow: full (: the issue has no frontmatter)". Rule: Decide is the only producer of a (Flow, Rule) pair; callers pass degenerate input (Recorded nil) through it and handle only "nowhere to write". Prevalence: 1 site; line 46 builds Decide''s input, which is fine.'
          family: decision-logic-restated
          round: 3
      boundary: M1
      blocked: true
    - "n": 4
      timestamp: "2026-09-17T21:21:22-07:00"
      agent: claude
      dispose:
        - id: BR-11
          disposition: addressed
          note: Scratch revert (write ran.content, skip the fresh read) turns TestRecordChangeCodeFlowKeepsConcurrentEdit and TestRecordChangeCodeFlowRefusesFlowChangingEdit red.
          round: 4
        - id: BR-12
          disposition: addressed
          note: Reverting Parse's seen[key] hash check to h != "" turns TestFlowRecordCorpus red on spec "" and done ""; the cue corpus test ran (cue on PATH) and passes.
          round: 4
        - id: BR-13
          disposition: addressed
          note: estimate.md:5, claim.md:17 and issue.md:93 are flow-qualified; re-running the recorded sweep finds no unqualified gate-action claim. A different verb class is raised separately.
          round: 4
        - id: BR-14
          disposition: addressed
          note: Restoring the ad hoc no-frontmatter branch turns TestDecideChangeCodeFlow and TestOnlyFlowPackageBuildsFlowValues red. The guard's holes are raised separately.
          round: 4
      findings:
        - id: BR-15
          severity: Minor
          title: change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
          detail: 'This is the 3rd finding in doc-claim-contradicts-code. helptext/change-code.md:4 (step 0 infers and records), atlas/workflow/issue-lifecycle.md:54 (first records, then runs structural checks) and atlas/workflow/sdlc-binary.md:36 (First it infers and records) were written in 24ea2b9; 1bd3b03 moved the write after the gates and past the dry-run return without updating them. The BR-13 sweep keyed on gate-action verbs and could not see an effect-timing change. Rule: every behaviour change in a fix round names its effect verb and subject (records ... flow) and that pair joins the sweep regex in the same round, so the regex derives from the diff''s changed effects, not a fixed verb list. Prevalence: 5 sentences describe writing the flow record; the 3 that state an order are all false. The new flowDrift refusal is also undocumented in change-code.md.'
          family: doc-claim-contradicts-code
          round: 4
        - id: BR-16
          severity: Minor
          title: TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
          detail: 'This is the 2nd finding in test-oracle-weaker-than-plan. Probe: rewriting BR-14''s original branch as var fl flow.Flow; fl.Kind = flow.Full, with rule set to a string literal, leaves the guard green; it scans only composite literals and conversions. Rule: a pin claimed for "only package X builds Y" must fail on every syntactic shape of the violation, so prefer compiler enforcement. Fix: make Flow''s fields unexported with accessors and have Decide be the only producer of Rule values (ARCH-ORDER structural enforcement), before M2''s close-time upgrade becomes the first new consumer.'
          family: test-oracle-weaker-than-plan
          round: 4
        - id: BR-17
          severity: Minor
          title: The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
          detail: 'This is the 3rd finding in record-roundtrips-every-reader. Deleting the seen[k] duplicate check in a scratch copy leaves every flow test green; Go would then read {kind: full, kind: quick, provenance: operator} as quick/operator while cue rejects it (both reject at HEAD, probed). Rule: the corpus enumerates Parse''s reject branches; give each error return a reason code, tag reject rows with it, and assert every code has at least one row, so no branch can exist unpinned. Prevalence: 8 reject branches, 6 in the corpus; the unparseable-YAML branch is backstopped by the non-map check, so duplicate-key is the one real gap.'
          family: record-roundtrips-every-reader
          round: 4
      boundary: M1
      blocked: false
---

# Gate ledger — ariadne#231 (boundary-review)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-17T20:53:40-07:00 (sdlc) — passed

### Raised

- **BR-1** [Minor] `test-prose-enumeration` Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
  Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.
  (carried from plan-quality PQ-8, deferred to the boundary review)
- **BR-2** [Minor] `record-roundtrips-every-reader` Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
  Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan's own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.
  (carried from plan-quality PQ-9, deferred to the boundary review)
- **BR-3** [Minor] `acceptance-oracle-freshness` Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
  Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.
  (carried from plan-quality PQ-10, deferred to the boundary review)

## Round 2 — 2026-09-17T20:53:40-07:00 (claude) — BLOCKED

### Raised

- **BR-4** [Important] `hand-listed-guard-incomplete` decideChangeCodeFlow's PlanItemsBody routing is unpinned — not in planItemBodySources, no fenced-Mx test
  changecode_flow.go:48 hands a plan body to the exempt helper issue.MilestonesInPlanOrder, whose exemption says callers are covered by planItemBodySources (commitpathspec_guard_test.go:341), which lists only close.go. Reverting to PlanSectionBody in a scratch copy left every Flow/Guard/PlanItem/ChangeCode test green. Add a fenced-Mx case to TestDecideChangeCodeFlow (must stay quick) and sweep all MilestonesInPlanOrder callers (ARCH-PURPOSE).
- **BR-5** [Important] `state-persisted-on-refused-transition` change-code writes the flow record before its gates, so a refused run leaves a sticky full/inferred
  applyChangeCodeFlow runs at changecode.go:137 before the gate loop; a structural/plan-quality/estimate refusal exits with flow full/inferred written (uncommitted), and Decide keeps a recorded full, so removing the plan or Mx rows and re-running stays full. The plan's ARCH-ORDER table says REWORK writes nothing. Decide before the gates, write after they pass (before syncIssue); test that a refused run leaves the file byte-identical (ARCH-ORDER).
- **BR-6** [Minor] `doc-claim-contradicts-code` M1 docs describe close-time flow behaviour (small-diff review, upgrade, Done-when refusal) that M2 hasn't built
  helptext/change-code.md THE FLOW, helptext/issue.md flow field, AGENTS.base.md section 2, atlas issue-lifecycle item 6 and flowInfoLine state it in the present tense; at HEAD close ignores the flow entirely. Acceptable only if M2 lands before any merge; note it in the Log.
- **BR-7** [Minor] `decision-logic-restated` flowReason re-derives Decide's rule order instead of Decide returning the rule that fired
  changecode_flow.go:66-81 mirrors flow.go Decide rules; a reorder in Decide silently mislabels the reason (ARCH-DRY).
- **BR-8** [Minor] `untrusted-input-rewrite` Rewriting a block-form hand-edited flow record orphans its child lines into invalid frontmatter
  SetField replaces only the flow line; a probe produced flow full/inferred followed by indented kind/provenance lines, while the warning says the record is being rewritten (ARCH-SECURE).
- **BR-9** [Minor] `test-oracle-weaker-than-plan` FuzzFromFrontmatter does not assert that the error path resolves to full
  The plan names "every error path resolves to full" as the fuzz oracle; the body returns on error without checking Kind == Full.
- **BR-10** [Minor] `unrelated-churn` issue.cue carries unrelated cue-fmt whitespace realignment of categories, Status and Transition

## Round 3 — 2026-09-17T21:07:39-07:00 (claude) — BLOCKED

### Disposed

- BR-1 — addressed — Plan names every test. FuzzFromFrontmatter plus the fenced-heading case in TestContractHashes cover M1's parsers; FuzzParseSurfaces is M2 scope; keeping case lists is a recorded decision.
- BR-2 — addressed — Format quotes hashes with %q (flow.go:66-72); scratch revert to %s turns TestFlowRoundTrip red on 12345678; validate_test covers quoted-valid and unquoted-invalid through real cue.
- BR-3 — addressed — Spec section 4 and Done-when say quick-flow and the Flow bullet carries spec/done; old names survive only in dated Revisions.
- BR-4 — addressed — Scratch revert of changecode_flow.go:50 or close.go:1719 to PlanSectionBody turns TestPlanItemReadersUsePlanItemsBody red; the fenced-Mx case in TestDecideChangeCodeFlow also goes red.
- BR-5 — addressed — Moving recordChangeCodeFlow before the gate loop turns TestRunChangeCodeRecordsFlowAfterGates red. The fix opened a stale-snapshot window, raised separately.
- BR-6 — addressed — The 2026-09-17 "M1 built" Log entry records that the present-tense close docs hold only because M2 lands before merge.
- BR-7 — addressed — Decide returns the Rule that fired, flowReason is gone, and TestDecide asserts the rule for each case.
- BR-8 — addressed — SetField drops the indented block continuation (frontmatter.go:73-99), pinned by TestSetFieldReplacesBlockValue.
- BR-9 — addressed — Scratch revert of FromFrontmatter's error value to Flow{} turns FuzzFromFrontmatter's seed corpus red.
- BR-10 — addressed — The issue.cue diff is now 17 pure additions, with no realignment.

### Raised

- **BR-11** [Important] `write-from-stale-snapshot` change-code writes the flow record from issue text read before the gates, clobbering edits made while plan-quality ran
  recordChangeCodeFlow (changecode_flow.go:100-108) writes d.content, built at changecode.go:138 from the step-2 read, after the multi-minute gate loop. A scratch probe (write, report, edit the file on disk, record) lost the edit, and syncIssue then commits it and, on main, pushes it. It fires on every issue's first successful run, because the record is always new. The repo lock does not cover editor or agent edits; close handles the same hazard by validating its snapshot after the review (close.go:1218-1231). Fix: re-read issuePath at record time and apply SetField to the fresh frontmatter, recomputing quick hashes over the fresh body, or refuse as close does. Add the probe as a regression test (ARCH-ORDER, second actor on the same state).
- **BR-12** [Minor] `record-roundtrips-every-reader` flow.Parse accepts spec "" and done "" that cue Flow rejects, contradicting Parse's own doc comment
  This is the 2nd finding in record-roundtrips-every-reader. Rule: the Go codec and the cue model are two readers of one record and must accept exactly the same set of values. Fix the rule, not the instance: a shared testdata corpus of records with expected accept or reject, asserted by both flow_test (Parse) and cmd/vocabulary validate_test (cue). That also catches #233 adding config to one reader only. Also reject empty hashes in Parse. Prevalence: 1 divergence among 3 probed edge shapes.
- **BR-13** [Minor] `doc-claim-contradicts-code` helptext/estimate.md:5 still says change-code refuses unless the Estimate block reconciles, with no flow qualification
  This is the 2nd finding in doc-claim-contradicts-code. Rule: when a change makes a gate conditional, every sentence asserting that gate's action is flow-qualified, and the enumeration comes from grepping the gate's action verbs, not from a hand-copied surface list. PQ-1's own regex matched estimate.md through estimate_hours, but the list dropped it. Measured: git grep over helptext, atlas, AGENTS.base.md and construct for change-code with refus, requir, demand, parses or asks-for, plus "estimate gate" and "plan-quality gate", gives about 15 hits; 1 is unqualified and false.
- **BR-14** [Minor] `decision-logic-restated` The no-frontmatter branch of decideChangeCodeFlow decides full itself, bypassing Decide and its pin validation
  This is the 2nd finding in decision-logic-restated. changecode_flow.go:35-39 returns full with an ad hoc Rule string and empty provenance. An invalid --flow fast, or a --flow quick pin, is silently ignored, and the info line prints "flow: full (: the issue has no frontmatter)". Rule: Decide is the only producer of a (Flow, Rule) pair; callers pass degenerate input (Recorded nil) through it and handle only "nowhere to write". Prevalence: 1 site; line 46 builds Decide's input, which is fine.

## Round 4 — 2026-09-17T21:21:22-07:00 (claude) — passed

### Disposed

- BR-11 — addressed — Scratch revert (write ran.content, skip the fresh read) turns TestRecordChangeCodeFlowKeepsConcurrentEdit and TestRecordChangeCodeFlowRefusesFlowChangingEdit red.
- BR-12 — addressed — Reverting Parse's seen[key] hash check to h != "" turns TestFlowRecordCorpus red on spec "" and done ""; the cue corpus test ran (cue on PATH) and passes.
- BR-13 — addressed — estimate.md:5, claim.md:17 and issue.md:93 are flow-qualified; re-running the recorded sweep finds no unqualified gate-action claim. A different verb class is raised separately.
- BR-14 — addressed — Restoring the ad hoc no-frontmatter branch turns TestDecideChangeCodeFlow and TestOnlyFlowPackageBuildsFlowValues red. The guard's holes are raised separately.

### Raised

- **BR-15** [Minor] `doc-claim-contradicts-code` change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
  This is the 3rd finding in doc-claim-contradicts-code. helptext/change-code.md:4 (step 0 infers and records), atlas/workflow/issue-lifecycle.md:54 (first records, then runs structural checks) and atlas/workflow/sdlc-binary.md:36 (First it infers and records) were written in 24ea2b9; 1bd3b03 moved the write after the gates and past the dry-run return without updating them. The BR-13 sweep keyed on gate-action verbs and could not see an effect-timing change. Rule: every behaviour change in a fix round names its effect verb and subject (records ... flow) and that pair joins the sweep regex in the same round, so the regex derives from the diff's changed effects, not a fixed verb list. Prevalence: 5 sentences describe writing the flow record; the 3 that state an order are all false. The new flowDrift refusal is also undocumented in change-code.md.
- **BR-16** [Minor] `test-oracle-weaker-than-plan` TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
  This is the 2nd finding in test-oracle-weaker-than-plan. Probe: rewriting BR-14's original branch as var fl flow.Flow; fl.Kind = flow.Full, with rule set to a string literal, leaves the guard green; it scans only composite literals and conversions. Rule: a pin claimed for "only package X builds Y" must fail on every syntactic shape of the violation, so prefer compiler enforcement. Fix: make Flow's fields unexported with accessors and have Decide be the only producer of Rule values (ARCH-ORDER structural enforcement), before M2's close-time upgrade becomes the first new consumer.
- **BR-17** [Minor] `record-roundtrips-every-reader` The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
  This is the 3rd finding in record-roundtrips-every-reader. Deleting the seen[k] duplicate check in a scratch copy leaves every flow test green; Go would then read {kind: full, kind: quick, provenance: operator} as quick/operator while cue rejects it (both reject at HEAD, probed). Rule: the corpus enumerates Parse's reject branches; give each error return a reason code, tag reject rows with it, and assert every code has at least one row, so no branch can exist unpinned. Prevalence: 8 reject branches, 6 in the corpus; the unparseable-YAML branch is backstopped by the non-map check, so duplicate-key is the one real gap.

## Open findings

- **BR-15** [Minor] `doc-claim-contradicts-code` change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
- **BR-16** [Minor] `test-oracle-weaker-than-plan` TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
- **BR-17** [Minor] `record-roundtrips-every-reader` The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
