---
gate: plan-quality
issue: 231
id_prefix: PQ
rounds:
    - "n": 1
      timestamp: "2026-09-17T20:18:41-07:00"
      agent: claude
      findings:
        - id: PQ-1
          severity: Important
          title: start-plan, AGENTS.base.md and root help still tell every issue to write a durable plan, so inference routes it to full
          detail: startplan.go:196-204 planPointer always directs superpowers-writing-plans into workshop/plans, and its framing (startplan.go:67-69) and helptext/start-plan.md say the same. AGENTS.base.md:25 restates the flow as claim, start-plan, design into workshop/plans, change-code. helptext/root.md:16-19 says change-code always runs plan-quality and the estimate gate. A durable plan existing is an inference input, so an agent following start-plan on a small issue lands on full. List every surface that tells agents to write a durable plan and make each conditional on the shell (a pointer to change-code --help); M3 currently rewrites only the section-2 threshold.
          family: single-source-consumer-sweep
          round: 1
        - id: PQ-2
          severity: Important
          title: DoneWhenFresh only watches the Spec section, but the constitution records mid-stream reframes in an appended Revisions section
          detail: 'AGENTS.md section 1 says to revise plan artifacts by appending a Revisions section, not by overwriting. A reframe recorded that way leaves Spec unchanged, the check passes, and the single close review judges against stale Done-when criteria, which is the case #263 hit. Compare Spec plus Revisions at the anchor against now, and add a Revisions-only-edit case that refuses.'
          family: acceptance-oracle-freshness
          round: 1
        - id: PQ-3
          severity: Minor
          title: Stripping the quick field breaks three tests that expect the registry verbatim; moving the milestone parser breaks more
          detail: judge_test.go:328-333, archprinciples_test.go:28 and startplan_test.go:41 check that the rendered output contains ArchitectureRegistry verbatim, which stops holding once quick lines are stripped. Point them at the stripped rendering rather than deleting them. Moving milestonesInPlanOrder and milestonePlanRE also breaks references at planfence_test.go:79,411 and close_test.go:444.
          family: plan-omits-affected-tests
          round: 1
        - id: PQ-4
          severity: Minor
          title: Decide keeps quick/operator as quick when change-code runs again after Mx rows were added
          detail: Rule 2 (a recorded operator flow stands) comes before any Mx check, so change-code skips plan-quality and estimate on an issue whose Mx row crosses the shell. Spec section 1 says a gate that finds the shell crossed upgrades whatever its provenance. Add that cell to the ARCH-ORDER table.
          family: shell-limit-enforced-at-every-gate
          round: 1
        - id: PQ-5
          severity: Minor
          title: The shared-surfaces declaration is read from the working tree, so an uncommitted edit can exempt the diff
          detail: The self-match only catches committed edits to the declaration. Read it at the window base (git show base:.sdlc/shared-surfaces), or union base and head, so a branch cannot loosen its own shell.
          family: trust-boundary-read-location
          round: 1
        - id: PQ-6
          severity: Minor
          title: The freshness anchor, the first commit that added a flow line, cannot move when change-code runs again
          detail: A later change-code run fixes the contract again but leaves the flow line unchanged, so git log -G never sees it. That causes false refusals the operator must bypass with --no-done-when-fresh. The log also lacks --follow, so a renamed issue file loses its anchor.
          family: acceptance-oracle-freshness
          round: 1
        - id: PQ-7
          severity: Minor
          title: Widening isTestPath changes what churn_prod and churn_test mean partway through the ledger history
          detail: 'For Lua and Python repos, rows written before the change count spec files as churn_prod and rows written after count them as churn_test, so any #187 cost comparison across the change is skewed. Say so in ledger-landscape.md or the Log.'
          family: ledger-column-semantics-stable
          round: 1
        - id: PQ-8
          severity: Minor
          title: Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
          detail: 'Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.'
          family: test-prose-enumeration
          round: 1
      blocked: true
    - "n": 2
      timestamp: "2026-09-17T20:25:39-07:00"
      agent: claude
      dispose:
        - id: PQ-1
          disposition: addressed
          note: Derived grep sweep lists every directive surface, now in M1; my re-run finds no missed surface.
          round: 2
        - id: PQ-2
          disposition: addressed
          note: spec hash covers Spec + Revisions; Revisions-only edit is a named test case.
          round: 2
        - id: PQ-3
          disposition: addressed
          note: Field kept at render so ArchitectureBlock stays verbatim; moved-parser test refs listed.
          round: 2
        - id: PQ-4
          disposition: addressed
          note: Decide rule 3 puts Mx rows ahead of a recorded operator flow; table cell added.
          round: 2
        - id: PQ-5
          disposition: addressed
          note: Surfaces read as committed at base union head.
          round: 2
        - id: PQ-6
          disposition: addressed
          note: Contract hashes recorded on every change-code run replace the git-log anchor.
          round: 2
        - id: PQ-7
          disposition: addressed
          note: Cut-over recorded in ledger-landscape.md and the Log.
          round: 2
        - id: PQ-8
          disposition: not-addressed
          note: Tests named and both parsers fuzzed, but the case lists remain for FlowStep/DoneWhenFresh/Crossings/CloseQuick.
          round: 2
      findings:
        - id: PQ-9
          severity: Minor
          title: Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
          detail: 'Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan''s own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.'
          family: record-roundtrips-every-reader
          round: 2
        - id: PQ-10
          severity: Minor
          title: Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
          detail: 'Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.'
          family: acceptance-oracle-freshness
          round: 2
      blocked: false
content_hash: b60f493d2202af5394b95554108e52a79d7d14a493f7ee2ee8aba139da1f03a8
---

# Gate ledger — ariadne#231 (plan-quality)

Findings this gate raised, the stable ids the binary assigned them, and how
later rounds disposed of them. Generated — edit the gate, not this file.

## Round 1 — 2026-09-17T20:18:41-07:00 (claude) — BLOCKED

### Raised

- **PQ-1** [Important] `single-source-consumer-sweep` start-plan, AGENTS.base.md and root help still tell every issue to write a durable plan, so inference routes it to full
  startplan.go:196-204 planPointer always directs superpowers-writing-plans into workshop/plans, and its framing (startplan.go:67-69) and helptext/start-plan.md say the same. AGENTS.base.md:25 restates the flow as claim, start-plan, design into workshop/plans, change-code. helptext/root.md:16-19 says change-code always runs plan-quality and the estimate gate. A durable plan existing is an inference input, so an agent following start-plan on a small issue lands on full. List every surface that tells agents to write a durable plan and make each conditional on the shell (a pointer to change-code --help); M3 currently rewrites only the section-2 threshold.
- **PQ-2** [Important] `acceptance-oracle-freshness` DoneWhenFresh only watches the Spec section, but the constitution records mid-stream reframes in an appended Revisions section
  AGENTS.md section 1 says to revise plan artifacts by appending a Revisions section, not by overwriting. A reframe recorded that way leaves Spec unchanged, the check passes, and the single close review judges against stale Done-when criteria, which is the case #263 hit. Compare Spec plus Revisions at the anchor against now, and add a Revisions-only-edit case that refuses.
- **PQ-3** [Minor] `plan-omits-affected-tests` Stripping the quick field breaks three tests that expect the registry verbatim; moving the milestone parser breaks more
  judge_test.go:328-333, archprinciples_test.go:28 and startplan_test.go:41 check that the rendered output contains ArchitectureRegistry verbatim, which stops holding once quick lines are stripped. Point them at the stripped rendering rather than deleting them. Moving milestonesInPlanOrder and milestonePlanRE also breaks references at planfence_test.go:79,411 and close_test.go:444.
- **PQ-4** [Minor] `shell-limit-enforced-at-every-gate` Decide keeps quick/operator as quick when change-code runs again after Mx rows were added
  Rule 2 (a recorded operator flow stands) comes before any Mx check, so change-code skips plan-quality and estimate on an issue whose Mx row crosses the shell. Spec section 1 says a gate that finds the shell crossed upgrades whatever its provenance. Add that cell to the ARCH-ORDER table.
- **PQ-5** [Minor] `trust-boundary-read-location` The shared-surfaces declaration is read from the working tree, so an uncommitted edit can exempt the diff
  The self-match only catches committed edits to the declaration. Read it at the window base (git show base:.sdlc/shared-surfaces), or union base and head, so a branch cannot loosen its own shell.
- **PQ-6** [Minor] `acceptance-oracle-freshness` The freshness anchor, the first commit that added a flow line, cannot move when change-code runs again
  A later change-code run fixes the contract again but leaves the flow line unchanged, so git log -G never sees it. That causes false refusals the operator must bypass with --no-done-when-fresh. The log also lacks --follow, so a renamed issue file loses its anchor.
- **PQ-7** [Minor] `ledger-column-semantics-stable` Widening isTestPath changes what churn_prod and churn_test mean partway through the ledger history
  For Lua and Python repos, rows written before the change count spec files as churn_prod and rows written after count them as churn_test, so any #187 cost comparison across the change is skewed. Say so in ledger-landscape.md or the Log.
- **PQ-8** [Minor] `test-prose-enumeration` Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
  Name each test and replace the case lists with one line per risky function. For example: fuzz flow.FromFrontmatter seeded with malformed, nested and duplicate-key maps (never panics, malformed resolves to full); fuzz ParseSurfaces (a bad pattern becomes a crossing); DoneWhenFresh section extraction over fenced headings.

## Round 2 — 2026-09-17T20:25:39-07:00 (claude) — passed

### Disposed

- PQ-1 — addressed — Derived grep sweep lists every directive surface, now in M1; my re-run finds no missed surface.
- PQ-2 — addressed — spec hash covers Spec + Revisions; Revisions-only edit is a named test case.
- PQ-3 — addressed — Field kept at render so ArchitectureBlock stays verbatim; moved-parser test refs listed.
- PQ-4 — addressed — Decide rule 3 puts Mx rows ahead of a recorded operator flow; table cell added.
- PQ-5 — addressed — Surfaces read as committed at base union head.
- PQ-6 — addressed — Contract hashes recorded on every change-code run replace the git-log anchor.
- PQ-7 — addressed — Cut-over recorded in ledger-landscape.md and the Log.
- PQ-8 — not-addressed — Tests named and both parsers fuzzed, but the case lists remain for FlowStep/DoneWhenFresh/Crossings/CloseQuick.

### Raised

- **PQ-9** [Minor] `record-roundtrips-every-reader` Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
  Verified with local cue: spec 12345678 gives "mismatched types int and string" against the plan's own #Flow. About 2.3 percent per hash, roughly 4.6 percent of quick records. validategate.go runs this on every changed issue at push and merge, and yaml.v3 hides it on the Go side. Rule: a persisted value must parse to the same type under every reader. Fix: Format quotes the hashes, and the round-trip test, FuzzFromFrontmatter and the validate_test flow case are seeded with an all-digit hash.
- **PQ-10** [Minor] `acceptance-oracle-freshness` Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
  Third finding in this family. The rule for all of them: a revision that renames or reshapes an artifact Done-when names restates that bullet in the same edit. Apply it now by checking every name in Done-when against the plan (both drifts here) before implementing. DoneWhenFresh will enforce this mechanically once it ships.

## Open findings

- **PQ-8** [Minor] `test-prose-enumeration` Tests are listed case by case in prose; the risky parsers have no one-line adversarial strategy
- **PQ-9** [Minor] `record-roundtrips-every-reader` Unquoted 8-hex hashes in the flow record become YAML ints, and cue checks at push/merge refuse them
- **PQ-10** [Minor] `acceptance-oracle-freshness` Spec section 4 and Done-when still say quick yes/no after the plan renamed it quick-flow, and Done-when's Flow omits the hash fields
