# Boundary Review — ariadne#231 (milestone M1)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | adaa53a60caea9815780269c76a852df03496f77..24ea2b951e9369a90c3fcfd5fb54a5836492ec72 |
| command | sdlc milestone-close --issue 231 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-17T20:53:40-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what the plan claims. There is a pure `internal/flow` package: a typed codec for the record, a `Decide` transition function, the shell limits and the contract hashes. There is now one milestone parser, and the colon-only regex is gone. `#Flow` is modelled in cue as a closed definition. `change-code` infers the flow, accepts an operator pin and skips its gates on quick. The live surfaces that routed every issue into a durable plan (start-plan, the constitution, help text, the brainstorming skill, atlas) now depend on the flow. I re-ran the `git grep` sweep; only archived `construct/versions/` and `construct/sources/` copies still send every issue to writing-plans, which is correct.

The flow, issue and vocabulary packages and the targeted `cmd/sdlc` tests pass at HEAD. Two cheap Important items stand before this boundary:
1. The new plan-body reader in change-code is not pinned by any test or guard. I confirmed this in a scratch copy: reverting it leaves the suite green.
2. The flow record is written before the gates run, so a refused `change-code` leaves behind a `full` record that then can't be undone without an operator pin.

**Strengths**
- `flow.Parse` (`cmd/sdlc/internal/flow/flow.go:83-131`) turns hand-editable frontmatter into a typed value at the boundary: closed sets, no duplicate keys, and hashes must be `!!str`. Every error resolves to full. The fuzz target and `TestParseRejects` cover the malformed inputs a person could write, including the int/float retyping of hashes (PQ-9).
- `Decide` (`flow.go:167-192`) is the single pure transition function, and `TestDecide` has one case per ARCH-ORDER table cell plus the pin refusals.
- One milestone parser (`internal/issue/plan.go:84-122`). `TestComputeSizingCountsEmDashMilestones` pins a real bug fix: the old parser reported zero milestones for the `M1 —` form.
- `activeChangeCodeGates` (`changecode_flow.go:118-124`) filters the gate list instead of guarding each closure, so `changeCodeGates` stays the complete declaration its ordering guards test. `flowInfoLine` builds the skipped-gate names from `changeCodeGateOrder()` rather than restating them.
- `ShellSummary()` is the single source for the limits: help reads it via `{{QUICK_SHELL}}` (`main.go:48`), start-plan and the info line read it, and tests pin both. The constitution and the skill point at `change-code --help` instead of restating the numbers.

**Critical findings:** none.

**Important findings**
1. **Nothing pins the plan-body routing in `decideChangeCodeFlow` (ARCH-PURPOSE).**
   - `changecode_flow.go:48` gets a body with `issue.PlanItemsBody` and passes it to `issue.MilestonesInPlanOrder`. That helper is exempt from the derived guard. Its exemption text says its callers are covered by `planItemBodySources` instead (`commitpathspec_guard_test.go:341`), but that list still names only `close.go:findMilestonesMissingVerdict`.
   - `TestDecideChangeCodeFlow` has no case with a milestone row inside a fence.
   - Scratch probe: I replaced `PlanItemsBody` with `PlanSectionBody` and ran the `Flow|Guard|PlanItem|Wiring|PlanFence|Milestone|ChangeCode` tests. All green apart from two failures caused only by the scratch copy not being a git repo.
   - Consequence: a regression would push any issue that quotes `- [ ] M1 — …` in a fenced example onto the full flow.
   - Fix: add a fenced-Mx case that must still infer quick. The function is pure, so a behavioural test is better than another wiring edge. Then sweep every caller of `MilestonesInPlanOrder`: `sizing.go:ComputeSizingFromContent` is still covered because it also references `PlanItemRE`; `decideChangeCodeFlow` is the only one missing.
2. **The flow record is written before the gates, and it sticks (ARCH-ORDER).**
   - `applyChangeCodeFlow` runs at `changecode.go:137`, before the gate loop. On the full flow, a structural, plan-quality or estimate refusal exits and leaves an uncommitted `flow: {kind: full, provenance: inferred}` in the issue.
   - `Decide` then treats that recorded `full` as permanent. An operator who deletes a stray durable plan or the Mx rows and re-runs gets "already full — gates never downgrade". Only `--flow quick` gets them back to quick.
   - The plan's ARCH-ORDER table says a REWORK writes nothing.
   - It is not a real defence against gaming either: the line is uncommitted, so `git checkout` removes it.
   - Fix: decide before the gates (they need `ctx.flow`), but write only after the loop passes, just before `syncIssue`. Add a test that a refused run leaves the issue file byte-identical.

**Minor findings**
- **Docs describe close behaviour that M2 hasn't built yet.** These passages describe it in the present tense:
  - `helptext/change-code.md` THE FLOW (close measures the diff, upgrades, refuses a moved contract)
  - the `flow` field in `helptext/issue.md` ("written by … `sdlc close`")
  - `AGENTS.base.md` §2 ("one small-diff review at close")
  - `atlas/workflow/issue-lifecycle.md` item 6
  - `flowInfoLine` ("close runs the small-diff review")

  At HEAD, close knows nothing about the flow: it runs the full milestone-review, with no Done-when checks and no upgrade. That is fine only if M2 lands before any merge; record that in the Log.
- **`flowReason` restates `Decide` (ARCH-DRY).** It recomputes `Decide`'s rule order (`changecode_flow.go:66-81`). `Decide` should return the rule that fired.
- **A block-form `flow:` produces invalid YAML when rewritten (ARCH-SECURE).** For a hand-edited multi-line map (`flow:\n  kind: quick\n  provenance: inferred`), `SetField` replaces only the `flow:` line and leaves the indented children behind. Probe output: `flow: {kind: full, provenance: inferred}\n  kind: quick\n  provenance: inferred`. The frontmatter is now invalid, while the warning says the record is being rewritten.
- **The fuzz oracle is weaker than the plan says.** `FuzzFromFrontmatter` returns on error without asserting `Kind == Full`, yet the plan names "every error path resolves to full" as its oracle.
- **Unrelated formatting churn.** `construct/vocabulary/issue.cue` has `cue fmt` whitespace changes in `categories`, `#Status` and `#Transition`.

**Test coverage notes**
- `go test` passes for `./cmd/sdlc/internal/flow/...`, `./cmd/sdlc/internal/issue/...` and `./cmd/vocabulary/...`. The targeted `cmd/sdlc` run (ChangeCode, PlanGate, GateOrder, Gatesig, Flow, StartPlan, EstimateTiming, Milestone, Sizing, PlanItem, Guard, PlanFence, Help, Placeholder, Wiring) also passes.
- The full `go test ./...` has exactly one failure: `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`. It was already failing before this window: it opens `workshop/plans/000200-…-plan.md`, which was archived to `workshop/history/plans/` before base `adaa53a6`, and nothing in this window touches it. Don't cite `go test ./...` as all-green in `--verified`.
- Probes: a fenced Mx example currently infers quick (correct). CRLF frontmatter fails to parse and falls to full with no write (fails safe).

**Architecture, marker by marker**
- ARCH-DRY: flagged, Minor (`flowReason`). Otherwise passes: one parser, one codec, the limits single-sourced, `HasDoneWhenBullet` factored out.
- ARCH-PURE: passes. The decisions live in `internal/flow` and the pure `decideChangeCodeFlow`, and `applyChangeCodeFlow` is a thin IO shell.
- ARCH-PURPOSE: flagged, Important (#1). The M1 purpose and the sweep of directive surfaces are otherwise delivered.
- ARCH-MOCK: passes. No new external dependency, and the tests use real temp files.
- ARCH-CONSTRAINTS: passes. A trivial CLI step.
- ARCH-SECURE: passes with a Minor (the block-form rewrite). Typed parse at the boundary, closed cue model, fuzzed.
- ARCH-ORDER: flagged, Important (#2). Otherwise every production write goes through `Decide`. The repo lock is applied by the command auto-wrap.
- ARCH-FUNERAL: passes. The `flow:` line lives and archives with its issue, and the `-plan-gate.md` ledger is covered by the existing archive sweep.
- Docs gate: atlas is updated for the M1 surface. The README documents no change-code flags, so there was nothing to update there.

**Architectural notes for upcoming work**
- M2: #2's fix and M2's close-time upgrade should share one rule: persist the flow only on a transition that passes.
- `ContractHashes` hashes the first `## Revisions` through the next `##` heading. Agents append Revisions after `## Log`, so a Log entry written at the end of the file would fall under Revisions and move the `spec` hash. M2's freshness check will turn that into a false refusal. Consider a test for it.
- M3: `driftSample` currently turns drift off when the latest row has an unknown model. A quick row (model `""`) does exactly that, so M3's "trailing quick row" test is the right target.

**Plan revision recommendations**
- `## Revisions`: the change-code gates are skipped by filtering the list (`activeChangeCodeGates`), not by per-closure guards. The test names changed to `TestDecideChangeCodeFlow`, `TestActiveChangeCodeGates` and `TestApplyChangeCodeFlowWritesUnlessDryRun`, replacing `TestChangeCodeFlowStep` and `TestChangeCodeGatesSkipOnQuick`.
- Core concepts: add `flow.WithContract` (pure, `flow.go`… actually `donewhen.go`) and the `decideChangeCodeFlow` / `applyChangeCodeFlow` split to the change-code integration row.
- If #2 is fixed, change the integration-point text "It writes `flow:` … unless `--dry-run`" to say the record is written after the gates pass.

```findings
findings:
  - id: new
    severity: Important
    family: hand-listed-guard-incomplete
    title: |
      decideChangeCodeFlow's PlanItemsBody routing is unpinned — not in planItemBodySources, no fenced-Mx test
    detail: |
      changecode_flow.go:48 hands a plan body to the exempt helper issue.MilestonesInPlanOrder, whose exemption says callers are covered by planItemBodySources (commitpathspec_guard_test.go:341), which lists only close.go. Reverting to PlanSectionBody in a scratch copy left every Flow/Guard/PlanItem/ChangeCode test green. Add a fenced-Mx case to TestDecideChangeCodeFlow (must stay quick) and sweep all MilestonesInPlanOrder callers (ARCH-PURPOSE).
  - id: new
    severity: Important
    family: state-persisted-on-refused-transition
    title: |
      change-code writes the flow record before its gates, so a refused run leaves a sticky full/inferred
    detail: |
      applyChangeCodeFlow runs at changecode.go:137 before the gate loop; a structural/plan-quality/estimate refusal exits with flow full/inferred written (uncommitted), and Decide keeps a recorded full, so removing the plan or Mx rows and re-running stays full. The plan's ARCH-ORDER table says REWORK writes nothing. Decide before the gates, write after they pass (before syncIssue); test that a refused run leaves the file byte-identical (ARCH-ORDER).
  - id: new
    severity: Minor
    family: doc-claim-contradicts-code
    title: |
      M1 docs describe close-time flow behaviour (small-diff review, upgrade, Done-when refusal) that M2 hasn't built
    detail: |
      helptext/change-code.md THE FLOW, helptext/issue.md flow field, AGENTS.base.md section 2, atlas issue-lifecycle item 6 and flowInfoLine state it in the present tense; at HEAD close ignores the flow entirely. Acceptable only if M2 lands before any merge; note it in the Log.
  - id: new
    severity: Minor
    family: decision-logic-restated
    title: |
      flowReason re-derives Decide's rule order instead of Decide returning the rule that fired
    detail: |
      changecode_flow.go:66-81 mirrors flow.go Decide rules; a reorder in Decide silently mislabels the reason (ARCH-DRY).
  - id: new
    severity: Minor
    family: untrusted-input-rewrite
    title: |
      Rewriting a block-form hand-edited flow record orphans its child lines into invalid frontmatter
    detail: |
      SetField replaces only the flow line; a probe produced flow full/inferred followed by indented kind/provenance lines, while the warning says the record is being rewritten (ARCH-SECURE).
  - id: new
    severity: Minor
    family: test-oracle-weaker-than-plan
    title: |
      FuzzFromFrontmatter does not assert that the error path resolves to full
    detail: |
      The plan names "every error path resolves to full" as the fuzz oracle; the body returns on error without checking Kind == Full.
  - id: new
    severity: Minor
    family: unrelated-churn
    title: |
      issue.cue carries unrelated cue-fmt whitespace realignment of categories, Status and Transition
```

---

## Re-review — 2026-09-17T21:07:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | adaa53a60caea9815780269c76a852df03496f77..5f73ddc325ba4410249d8ec68ffd2123b25ceee2 |
| command | sdlc milestone-close --issue 231 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-17T21:07:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what the plan claims. The flow record is parsed strictly at the frontmatter boundary. `Decide` is a pure function with a test case for every row of the ARCH-ORDER table. change-code infers the flow, accepts an operator pin, and skips its gates on the quick flow. The shell limits come from one source, and every surface that sent work into a durable plan is now conditional on the flow. I disposed all ten prior findings as `addressed`. For each behaviour fix I undid it in a scratch copy and confirmed a test goes red: BR-4 at both production callers, plus BR-5, BR-2 and BR-9.

One Important finding stops this from being SHIP. The BR-5 fix moved the flow-record write to after the gates, but it still writes the issue text read *before* them. Any edit made to the issue while plan-quality runs, which takes minutes, is silently overwritten and then committed. A scratch probe reproduced the loss. The other three new findings are Minor, and each repeats a family from earlier rounds, so each states the rule that covers all instances.

**1. Strengths**
- **BR-4 was fixed for the whole class, not one site.** Listing `MilestonesInPlanOrder` in `planItemMatchers` (`commitpathspec_guard_test.go:320`) derives all three callers: `changecode_flow.go:50`, `close.go:1719` and `sizing.go:62`. Undoing the fix at either production site turns `TestPlanItemReadersUsePlanItemsBody` red.
- **`flow.Decide` is a real pure state machine.** It returns the `Rule` that fired instead of letting callers re-derive it (`flow.go:180-205`). `TestDecide` covers every ARCH-ORDER cell and checks the rule too.
- **`flow.Parse` is strict at the trust boundary.** It checks the `!!str` type, rejects duplicate and unknown keys, and enforces the closed sets. `FuzzFromFrontmatter` checks that every error resolves to `full`.
- **The limits have one source.** `ShellSummary()` feeds `planPointer`, `{{QUICK_SHELL}}` and `flowInfoLine`. `AGENTS.base.md` and the brainstorming skill point at `--help`. A tree-wide grep found no restated numbers.
- **One filter decides which gates run.** `activeChangeCodeGates` is the only place that skips gates, so `changeCodeGates` stays the complete declaration its ordering guards test.

**2. Critical findings**
None.

**3. Important findings**
- **The flow record is written from a stale copy of the issue** (`changecode_flow.go:100-108`, `changecode.go:138/197`).
  - `d.content` is built from the step-2 read. It is written after the gate loop, which includes the multi-minute plan-quality judge.
  - My probe (write, report, edit the file on disk, record) lost the edit, and `syncIssue` then commits it (and pushes it on main).
  - It triggers on every issue's first successful run, because the record is always new then.
  - The repo lock doesn't cover editor or agent edits. `close` already handles this case by re-checking its snapshot after its review (`close.go:1218-1231`).
  - Fix: re-read `issuePath` inside `recordChangeCodeFlow`, apply `SetField` to the fresh frontmatter, and recompute the quick hashes over the fresh body. Alternatively, refuse the way `close` does. Turn the probe into a regression test (ARCH-ORDER: a second actor on the same state).

**4. Minor findings**
- `flow.Parse` accepts `spec: ""` and `done: ""`, but cue `#Flow` rejects them. This contradicts `Parse`'s own doc comment. It is the second finding in `record-roundtrips-every-reader`.
- `helptext/estimate.md:5` still says change-code "refuses to proceed unless it reconciles", with no flow qualification. It is the second finding in `doc-claim-contradicts-code`.
- The no-frontmatter branch in `decideChangeCodeFlow` bypasses `Decide`, so a bad `--flow` pin is silently ignored and the info line prints `flow: full (: …)`. It is the second finding in `decision-logic-restated`.
- Nits (not ledgered):
  - `git diff --check` flags a blank line at EOF in the plan file (`…-plan.md:385`).
  - In `changecode_flow_test.go` the `go/ast`, `go/parser` and `go/token` imports sit in the module import group.

**5. Test coverage notes**
- `go test ./cmd/sdlc/` has one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`. It already fails at base: the #200 plan was moved to `workshop/history/plans/` before `adaa53a6`. The M1 `--verified` evidence should say so rather than claim a green suite.
- The flow, issue and vocabulary packages pass. cue is on PATH, so the `#Flow` cases actually ran rather than skipping.
- `DoneWhenPresent` and `DoneWhenFresh` have no production caller until M2, which matches the plan.

**6. Architectural notes for upcoming work**
- M2's upgrade at finalize should reuse `close`'s existing snapshot check rather than add a second stale-write pattern.
- Before #233 adds `config`, consider deriving Go's closed sets for `kind` and `provenance` from the cue model (as `pkg/vocab` does for statuses). Today they are stated twice, in `flow.go` and `issue.cue`.

**7. Plan revision recommendations**
- Add a Revisions entry for the stale-write fix: `recordChangeCodeFlow` re-reads the issue. Name its regression test.
- ARCH-ORDER's "Mutations run under the repo lock" should add that the lock doesn't cover direct file edits, and that change-code's post-gate write re-reads the issue.
- The PQ-1 surface list in Decisions should add `helptext/estimate.md`. The plan's own sweep regex matched it through `estimate_hours`.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan names every test. FuzzFromFrontmatter plus the fenced-heading case in TestContractHashes cover M1's parsers; FuzzParseSurfaces is M2 scope; keeping case lists is a recorded decision.
  - id: BR-2
    disposition: addressed
    note: |
      Format quotes hashes with %q (flow.go:66-72); scratch revert to %s turns TestFlowRoundTrip red on 12345678; validate_test covers quoted-valid and unquoted-invalid through real cue.
  - id: BR-3
    disposition: addressed
    note: |
      Spec section 4 and Done-when say quick-flow and the Flow bullet carries spec/done; old names survive only in dated Revisions.
  - id: BR-4
    disposition: addressed
    note: |
      Scratch revert of changecode_flow.go:50 or close.go:1719 to PlanSectionBody turns TestPlanItemReadersUsePlanItemsBody red; the fenced-Mx case in TestDecideChangeCodeFlow also goes red.
  - id: BR-5
    disposition: addressed
    note: |
      Moving recordChangeCodeFlow before the gate loop turns TestRunChangeCodeRecordsFlowAfterGates red. The fix opened a stale-snapshot window, raised separately.
  - id: BR-6
    disposition: addressed
    note: |
      The 2026-09-17 "M1 built" Log entry records that the present-tense close docs hold only because M2 lands before merge.
  - id: BR-7
    disposition: addressed
    note: |
      Decide returns the Rule that fired, flowReason is gone, and TestDecide asserts the rule for each case.
  - id: BR-8
    disposition: addressed
    note: |
      SetField drops the indented block continuation (frontmatter.go:73-99), pinned by TestSetFieldReplacesBlockValue.
  - id: BR-9
    disposition: addressed
    note: |
      Scratch revert of FromFrontmatter's error value to Flow{} turns FuzzFromFrontmatter's seed corpus red.
  - id: BR-10
    disposition: addressed
    note: |
      The issue.cue diff is now 17 pure additions, with no realignment.
findings:
  - id: new
    severity: Important
    family: write-from-stale-snapshot
    title: |
      change-code writes the flow record from issue text read before the gates, clobbering edits made while plan-quality ran
    detail: |
      recordChangeCodeFlow (changecode_flow.go:100-108) writes d.content, built at changecode.go:138 from the step-2 read, after the multi-minute gate loop. A scratch probe (write, report, edit the file on disk, record) lost the edit, and syncIssue then commits it and, on main, pushes it. It fires on every issue's first successful run, because the record is always new. The repo lock does not cover editor or agent edits; close handles the same hazard by validating its snapshot after the review (close.go:1218-1231). Fix: re-read issuePath at record time and apply SetField to the fresh frontmatter, recomputing quick hashes over the fresh body, or refuse as close does. Add the probe as a regression test (ARCH-ORDER, second actor on the same state).
  - id: new
    severity: Minor
    family: record-roundtrips-every-reader
    title: |
      flow.Parse accepts spec "" and done "" that cue Flow rejects, contradicting Parse's own doc comment
    detail: |
      This is the 2nd finding in record-roundtrips-every-reader. Rule: the Go codec and the cue model are two readers of one record and must accept exactly the same set of values. Fix the rule, not the instance: a shared testdata corpus of records with expected accept or reject, asserted by both flow_test (Parse) and cmd/vocabulary validate_test (cue). That also catches #233 adding config to one reader only. Also reject empty hashes in Parse. Prevalence: 1 divergence among 3 probed edge shapes.
  - id: new
    severity: Minor
    family: doc-claim-contradicts-code
    title: |
      helptext/estimate.md:5 still says change-code refuses unless the Estimate block reconciles, with no flow qualification
    detail: |
      This is the 2nd finding in doc-claim-contradicts-code. Rule: when a change makes a gate conditional, every sentence asserting that gate's action is flow-qualified, and the enumeration comes from grepping the gate's action verbs, not from a hand-copied surface list. PQ-1's own regex matched estimate.md through estimate_hours, but the list dropped it. Measured: git grep over helptext, atlas, AGENTS.base.md and construct for change-code with refus, requir, demand, parses or asks-for, plus "estimate gate" and "plan-quality gate", gives about 15 hits; 1 is unqualified and false.
  - id: new
    severity: Minor
    family: decision-logic-restated
    title: |
      The no-frontmatter branch of decideChangeCodeFlow decides full itself, bypassing Decide and its pin validation
    detail: |
      This is the 2nd finding in decision-logic-restated. changecode_flow.go:35-39 returns full with an ad hoc Rule string and empty provenance. An invalid --flow fast, or a --flow quick pin, is silently ignored, and the info line prints "flow: full (: the issue has no frontmatter)". Rule: Decide is the only producer of a (Flow, Rule) pair; callers pass degenerate input (Recorded nil) through it and handle only "nowhere to write". Prevalence: 1 site; line 46 builds Decide's input, which is fine.
```

---

## Re-review — 2026-09-17T21:21:22-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 231 — A quick path for small diffs: scale the gate set and the review recipe to size |
| repo | ariadne |
| issue file | workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | adaa53a60caea9815780269c76a852df03496f77..5e9b7f2be4b323060f5110d023bd724cc22c7382 |
| command | sdlc milestone-close --issue 231 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-09-17T21:21:22-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

All four open findings from round 2 are fixed. For each one I undid the fix in a scratch copy of HEAD and confirmed a test goes red:
- **BR-11:** writing the stale text again turns `TestRecordChangeCodeFlowKeepsConcurrentEdit` and `…RefusesFlowChangingEdit` red.
- **BR-12:** going back to the old `h != ""` hash check turns `TestFlowRecordCorpus` red on `spec: ""` and `done: ""`. cue is on PATH, so the cue half of the corpus actually ran.
- **BR-14:** restoring the ad hoc no-frontmatter branch turns `TestDecideChangeCodeFlow` and `TestOnlyFlowPackageBuildsFlowValues` red.
- **BR-13:** I re-ran the recorded verb sweep and found no unqualified gate-action claim left.

The three new findings are all Minor, and each repeats a family from earlier rounds. So each one states the rule to fix rather than the instance:
- Three docs still say change-code records the flow *before* its gates. BR-5 moved that write to after them.
- The source guard added for BR-14 misses field assignment and `flow.Rule` strings.
- The shared corpus dropped `Parse`'s duplicate-key rejection when it replaced `TestParseRejects`.

Nothing blocks the M1 boundary.

**1. Strengths**
- **BR-11 fixed as a rule.** `recordChangeCodeFlow` (`changecode_flow.go:100-121`) re-derives the record from a fresh read, so the contract hashes describe the edited text. `flowDrift` refuses when an edit changed the flow the gates ran under. Both are tested with a second writer. The lesson in `workshop/lessons.md` states the general rule.
- **BR-12 now checks both readers the same way.** One corpus (`construct/vocabulary/testdata/flow_records.txt`) is asserted by the Go codec and by real cue. It already paid off: it corrected the plan's claim that unquoted `12e45678` is a float. I probed truncated, duplicate-key and empty values, and both readers agree on all of them today.
- **`flow.Recorded`** is now the one place a malformed record resolves to full/inferred. The inline switch in change-code is gone.
- **The BR-13 sweep command is recorded in the plan**, so a reviewer can re-run it. That is how I found the gap in finding 1 below.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**

- **Three docs say change-code records the flow before its gates (ARCH-PURPOSE).**
  - This is the 3rd finding in family `doc-claim-contradicts-code`.
  - `helptext/change-code.md:4` says step 0 "infers the issue's flow and records it".
  - `atlas/workflow/issue-lifecycle.md:54` says change-code "first records the issue's flow … then runs structural checks".
  - `atlas/workflow/sdlc-binary.md:36` says "First it infers and records".
  - All three were written in 24ea2b9. BR-5 (1bd3b03) then moved the write after the gates and past the dry-run return, and left them unchanged.
  - The BR-13 sweep keyed only on gate-action verbs (refus, requir…), so it could not see a changed effect (records).
  - **Rule:** every behaviour change in a fix round names its effect verb and subject ("records … flow"), and that pair joins the sweep regex in the same round. The regex comes from the diff's changed effects, not a fixed list of verbs.
  - **Prevalence:** 5 sentences describe writing the flow record. The 3 that state an order are all false; the other 2 say nothing about order.
  - The new flow-drift refusal is also undocumented in `change-code.md`.
- **The BR-14 guard misses field writes and `flow.Rule` strings (ARCH-ORDER).**
  - This is the 2nd finding in family `test-oracle-weaker-than-plan`.
  - `TestOnlyFlowPackageBuildsFlowValues` only scans for composite literals and conversions.
  - Probe: I rewrote BR-14's original defect as `var fl flow.Flow; fl.Kind = flow.Full` with `rule: "the issue has no frontmatter"`. The guard stays green.
  - **Rule:** when "only package X builds Y" is claimed as pinned, the pin must fail on every syntactic shape of the violation. The plan's Revisions claim it is pinned; it is not.
  - **Fix:** use the compiler, as ARCH-ORDER's structural enforcement asks. Give `Flow` unexported fields with accessors, and make `Decide` the only thing that produces a `Rule` value.
  - Do this before M2: the close-time upgrade is the first new consumer, and it is exactly where `rec.Kind = flow.Full` would get written.
- **The corpus doesn't cover `Parse`'s duplicate-key rejection (ARCH-SECURE).**
  - This is the 3rd finding in family `record-roundtrips-every-reader`.
  - Moving `TestParseRejects` into the corpus dropped its `{kind: quick, kind: full, provenance: inferred}` case.
  - Deleting the `seen[k]` check in a scratch copy leaves every flow test green. Go would then read `{kind: full, kind: quick, provenance: operator}` as quick/operator, which cue rejects.
  - **Rule:** the corpus enumerates `Parse`'s reject branches. Give each error return a reason code, tag each reject row with its reason, and assert that every code has at least one row. Then no branch can exist unpinned.
  - **Prevalence:** 8 reject branches, 6 in the corpus. The unparseable-YAML branch is backstopped by the non-map check, so the duplicate-key branch is the only real gap.
- **Nits (not ledgered):**
  - `git diff --check` still flags a blank line at EOF in the plan (`…-plan.md:405`).
  - In `changecode_flow_test.go`, the `go/ast`, `go/parser`, `go/token` and `io/fs` imports sit in the module import group.
  - The `<verdict>\t<value>` corpus line parser is written twice, in `flow_test.go` and in `validate_test.go`.
  - `atlas/workflow/vocabulary.md` doesn't mention the shared corpus, and #233 must extend it.

**5. Test coverage notes**
- `go test` passes for `./cmd/sdlc/internal/flow/...`, `./cmd/sdlc/internal/issue/...` and `./cmd/vocabulary/...`. The targeted `cmd/sdlc` run (ChangeCode, Flow, PlanGate, GateOrder, Gatesig, StartPlan, EstimateTiming, Milestone, Sizing, PlanItem, Guard, PlanFence, Help, Placeholder, Wiring) also passes.
- The full `go test ./...` in the scratch copy fails in two places, neither caused by this window:
  - `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` already fails at base: the #200 plan was archived. Don't cite `go test ./...` as all-green in `--verified`.
  - `TestArchitecture_NarrativeRoutesToArchPrinciples` fails only because the scratch copy lacks the woven, untracked `AGENTS.md`.

**6. Architectural notes for upcoming work**
- ARCH-DRY: pass (the duplicated corpus parser is a nit).
- ARCH-PURE: pass. `flowDrift` and `decideChangeCodeFlow` are pure, and `recordChangeCodeFlow` is a thin IO shell.
- ARCH-PURPOSE: flagged, Minor (finding 1).
- ARCH-MOCK: pass. There is no new external dependency, and cue runs for real.
- ARCH-CONSTRAINTS: pass.
- ARCH-SECURE: flagged, Minor (finding 3).
- ARCH-ORDER: BR-11's second-actor gap is closed. Flagged, Minor, for the guard's holes (finding 2).
- ARCH-FUNERAL: pass. The corpus is static testdata.
- Docs gate: atlas covers the M1 surface, apart from finding 1's ordering claims. README documents no change-code flags, so it needed no change.
- M2's close-time upgrade should build its `{full, inferred}` through package `flow`, for example a `flow.Upgrade`, and not by assigning fields. Finding 2's encapsulation fix makes that the only option.

**7. Plan revision recommendations**
- Add a Revisions note:
  - `Decide` returns `(Flow, Rule, error)`; Core concepts line 47 still says `(Flow, error)`.
  - Add `flow.Recorded`, `flow.WithContract` and `Rule` to the pure-entities table.
- If finding 2 is fixed by encapsulation, record that `Flow`'s fields are unexported and that `TestOnlyFlowPackageBuildsFlowValues` is retired or narrowed.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      Scratch revert (write ran.content, skip the fresh read) turns TestRecordChangeCodeFlowKeepsConcurrentEdit and TestRecordChangeCodeFlowRefusesFlowChangingEdit red.
  - id: BR-12
    disposition: addressed
    note: |
      Reverting Parse's seen[key] hash check to h != "" turns TestFlowRecordCorpus red on spec "" and done ""; the cue corpus test ran (cue on PATH) and passes.
  - id: BR-13
    disposition: addressed
    note: |
      estimate.md:5, claim.md:17 and issue.md:93 are flow-qualified; re-running the recorded sweep finds no unqualified gate-action claim. A different verb class is raised separately.
  - id: BR-14
    disposition: addressed
    note: |
      Restoring the ad hoc no-frontmatter branch turns TestDecideChangeCodeFlow and TestOnlyFlowPackageBuildsFlowValues red. The guard's holes are raised separately.
findings:
  - id: new
    severity: Minor
    family: doc-claim-contradicts-code
    title: |
      change-code help and two atlas pages say the flow is recorded first, but since BR-5 it is written after the gates
    detail: |
      This is the 3rd finding in doc-claim-contradicts-code. helptext/change-code.md:4 (step 0 infers and records), atlas/workflow/issue-lifecycle.md:54 (first records, then runs structural checks) and atlas/workflow/sdlc-binary.md:36 (First it infers and records) were written in 24ea2b9; 1bd3b03 moved the write after the gates and past the dry-run return without updating them. The BR-13 sweep keyed on gate-action verbs and could not see an effect-timing change. Rule: every behaviour change in a fix round names its effect verb and subject (records ... flow) and that pair joins the sweep regex in the same round, so the regex derives from the diff's changed effects, not a fixed verb list. Prevalence: 5 sentences describe writing the flow record; the 3 that state an order are all false. The new flowDrift refusal is also undocumented in change-code.md.
  - id: new
    severity: Minor
    family: test-oracle-weaker-than-plan
    title: |
      TestOnlyFlowPackageBuildsFlowValues misses field assignment and untyped Rule strings, so BR-14's own defect passes it
    detail: |
      This is the 2nd finding in test-oracle-weaker-than-plan. Probe: rewriting BR-14's original branch as var fl flow.Flow; fl.Kind = flow.Full, with rule set to a string literal, leaves the guard green; it scans only composite literals and conversions. Rule: a pin claimed for "only package X builds Y" must fail on every syntactic shape of the violation, so prefer compiler enforcement. Fix: make Flow's fields unexported with accessors and have Decide be the only producer of Rule values (ARCH-ORDER structural enforcement), before M2's close-time upgrade becomes the first new consumer.
  - id: new
    severity: Minor
    family: record-roundtrips-every-reader
    title: |
      The shared flow corpus dropped Parse's duplicate-key reject branch when it replaced TestParseRejects
    detail: |
      This is the 3rd finding in record-roundtrips-every-reader. Deleting the seen[k] duplicate check in a scratch copy leaves every flow test green; Go would then read {kind: full, kind: quick, provenance: operator} as quick/operator while cue rejects it (both reject at HEAD, probed). Rule: the corpus enumerates Parse's reject branches; give each error return a reason code, tag reject rows with it, and assert every code has at least one row, so no branch can exist unpinned. Prevalence: 8 reject branches, 6 in the corpus; the unparseable-YAML branch is backstopped by the non-map check, so duplicate-key is the one real gap.
```
