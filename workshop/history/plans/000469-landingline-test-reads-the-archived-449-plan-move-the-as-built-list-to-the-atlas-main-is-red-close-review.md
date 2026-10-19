# Boundary Review — ariadne#469 (whole-issue close)

| field | value |
|-------|-------|
| issue | 469 — landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red) |
| repo | ariadne |
| issue file | workshop/issues/000469-landingline-test-reads-the-archived-449-plan-move-the-as-built-list-to-the-atlas-main-is-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | cabcee34f5008627158452f85ee7b811dc3a0bf2..7879351a725d5ac52d9ea00e2ced1c362ba0dde7 |
| command | sdlc close --issue 469 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T11:44:15-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: checking how the new lint rule gets run across the repo.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The fix does what the issue asks. `TestPlanNamesEveryTrainEventAndEffect` now reads `atlas/workflow/sdlc-binary.md`. The as-built events, effects and phases are now listed on that page. A new `test-reads-workshop` rule is in the `pkg/golint` catalog, with fixtures that prove both outcomes: a match and a non-match. `TestTrackedGoFilesPassTheCatalog` already runs the catalog over every tracked Go file, so the guard is enforced. `go test ./cmd/sdlc/internal/landingline/ ./pkg/golint/` passes at head, including the whole-repo catalog scan. I did not run the full suite. One problem should be fixed before shipping. The atlas section "The Go lint catalog" names every rule in the catalog with its incident, and the new rule was not added there. The rest are minor doc and coverage nits.

**1. Strengths**
- The test points at durable documentation instead of a process artifact, which removes the root cause (`trainstate_test.go:341`).
- The rule's regex `^(\.\./)+workshop/(plans|issues)(/|$)` only matches paths that climb into the repo's own workshop. Temp-repo fixture paths (`workshop/issues/…`, `../dst/workshop/…`) and `workshop/history/…` are not matched, and the fixture rows prove each case (`rules_test.go:96-103`).
- The rule is registered with `Incident: "ariadne#469"` and `TestOnly`, and its fixtures include the non-test-file case.
- The atlas list is kept in step with the code by the test, so it is enforced and not just a copy.

**2. Critical findings**
None.

**3. Important findings**
- **Missing atlas entry for the new rule** (`atlas/workflow/sdlc-binary.md:2027-2030`). The sentence "each rule names the incident that taught it (…)" lists all 8 earlier rules but not `test-reads-workshop` (#469). Fix: add "`test-reads-workshop` #469" to that list.

**4. Minor findings**
- **Stale plan link** (`atlas/workflow/sdlc-binary.md:462`). The line still says "Plan: `workshop/plans/000449-landing-train-plan.md`", but that plan was moved to `workshop/history/plans/`. It sits right next to the edited lines and is the same mistake class this issue fixes. Other atlas pages have the same pattern, all older than this change: `harness-integration.md:104`, `introspect.md:325` and `issue-tracker-migration.md:98`.
- **Phases are not checked** (`atlas/workflow/sdlc-binary.md:457`). The atlas says `TestPlanNamesEveryTrainEventAndEffect` "keeps this list in step". The test only checks events and effects (the backtick-prefixed type names). The Phases line is not checked at all.
- **Easy bypass of the rule.** A path split into segments, such as `filepath.Join("..", "..", "workshop", "plans", …)`, is not matched. It is acceptable for the stated "Done when", but worth knowing.
- **Out-of-date test name.** `TestPlanNamesEveryTrainEventAndEffect` still says "Plan" although it now reads the atlas. Renaming it means updating the atlas line that names it as well.

**5. Test coverage notes**
- The rule's fixtures cover a match and a non-match, the test-file and non-test-file cases, and the history/atlas/fixture-path exclusions.
- I confirmed the archived plan now lives only at `workshop/history/plans/000449-landing-train-plan.md`, so the old test path would fail. The new path exists, and the test passes against it.

**6. Architectural notes**
- **ARCH-DRY: pass.** The rule reuses the existing `calls` and `stringLit` helpers.
- **ARCH-PURE: pass.** The rule is a pure AST matcher, tested on source text without IO.
- **ARCH-PURPOSE: flagged once.** The test and the guard are delivered. The atlas rule list is a hand-maintained restatement of the catalog that was not updated; that is the Important finding above.
- **ARCH-EVOLVE: not applicable.** No persisted format changes.

**7. Plan revision recommendations**
None. The plan is an empty quick-flow checklist, and the Log matches the code.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: atlas-restatement-drift
    title: |
      The atlas Go lint catalog list leaves out the new test-reads-workshop rule (#469)
    detail: |
      atlas/workflow/sdlc-binary.md:2027-2030 names every catalog rule with its incident. It lists all 8 earlier rules but not test-reads-workshop, which pkg/golint/rules.go:64 adds. Add "`test-reads-workshop` #469" to that list. No other Catalog-restating list exists in this window.
  - id: new
    severity: Minor
    kind: defect
    family: doc-links-archived-workshop-path
    title: |
      The atlas still links the archived #449 plan at workshop/plans
    detail: |
      atlas/workflow/sdlc-binary.md:462 says "Plan: `workshop/plans/000449-landing-train-plan.md`", but the file now lives at workshop/history/plans/. Other atlas lines with the same pattern, all older than this change: harness-integration.md:104, introspect.md:325, issue-tracker-migration.md:98.
  - id: new
    severity: Minor
    kind: defect
    family: doc-claim-exceeds-enforcement
    title: |
      The atlas says the test keeps the Phases list in step, but it only checks events and effects
    detail: |
      atlas/workflow/sdlc-binary.md:457 (Phases) falls under the claim "TestPlanNamesEveryTrainEventAndEffect keeps this list in step". The test matches only backtick-prefixed event and effect type names. Either narrow the claim or extend the test to phases.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-literal-only
    title: |
      test-reads-workshop misses paths split into Join segments
    detail: |
      filepath.Join("..", "..", "workshop", "plans", "x.md") passes the rule, because only a single literal that climbs into workshop/plans or workshop/issues is matched. This is acceptable for the Done-when; consider matching adjacent "workshop","plans"/"issues" arguments that follow ".." in the same Join call.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d6bab811-356f-4993-bbdb-fef48e6915e9","repository":"github.com/xianxu/ariadne","issue_id":"000469","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/69b25147586cac0d","reviewed_head":"7879351a725d5ac52d9ea00e2ced1c362ba0dde7","slot":"ariadne:2","requested_at":"2026-10-19T18:43:08.990555Z"},"phase":"finished","pid":60831,"started_at":"2026-10-19T18:43:09.147762Z","finished_at":"2026-10-19T18:44:14.90636Z","session_id":"f2a77207-4be8-42be-b365-bddac8c8c3b9","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T11:45:30-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 469 — landingline: test reads the archived #449 plan; move the as-built list to the atlas (main is red) |
| repo | ariadne |
| issue file | workshop/issues/000469-landingline-test-reads-the-archived-449-plan-move-the-as-built-list-to-the-atlas-main-is-red.md |
| boundary | whole-issue close |
| milestone | — |
| window | cabcee34f5008627158452f85ee7b811dc3a0bf2..7b2bff5878d86def1d2aa555929b5136b90e1ceb |
| command | sdlc close --issue 469 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T11:45:30-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This change does what the issue set out to do. `TestPlanNamesEveryTrainEventAndEffect` now reads `atlas/workflow/sdlc-binary.md` instead of the archived #449 plan. The new lint rule `test-reads-workshop` refuses the exact path that broke main, and a fixture row proves it. `go test ./pkg/golint/ ./cmd/sdlc/internal/landingline/` passes at the pinned head. The fix for BR-1 is real: I removed the backticks around `test-reads-workshop` in the atlas, and `TestAtlasNamesEveryRule` went red (`rules_test.go:208: ... does not name test-reads-workshop`); I then restored the file. Three Minor findings from the earlier round are still open, and BR-2 takes one line to fix. None of them blocks the gate.

1. **Strengths**
   - `pkg/golint/rules.go:351-368`: the regex is anchored to paths that climb with `../` into `workshop/plans` or `workshop/issues`. Its tests check the boundaries in both directions: fixture paths inside a temp repo (`../dst/workshop/…`, a bare `workshop/issues/…`) and `workshop/history/plans` all pass the rule.
   - `pkg/golint/rules_test.go:98`: the first fixture row is the exact path from the incident, so the Done-when clause "the guard refuses a test reading `workshop/plans/…`" is proved literally. The rule is also checked in both test and non-test files (the last row shows non-test files are not flagged).
   - `TestAtlasNamesEveryRule` (`rules_test.go:199`) turns the hand-kept catalog list in the atlas into a list a test enforces. That closes the BR-1 class, not just the one missing rule (ARCH-PURPOSE).
   - The as-built list is durable documentation, and it now lives in the atlas next to the landing-train section (`sdlc-binary.md:457-460`).

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** BR-2, BR-3 and BR-4 are still open; their dispositions are below. I found nothing new.

5. **Test coverage notes:**
   - BR-1's regression evidence is confirmed by reverting the fix, as described above.
   - `TestAtlasNamesEveryRule` searches the whole atlas file, not only the catalog paragraph, so a rule name in backticks anywhere in the file would satisfy it. That is acceptable for now; I am not raising it as a finding.

6. **Architectural notes:**
   - **ARCH-DRY: pass.** `workshopReads` reuses the existing `calls` and `stringLit` helpers.
   - **ARCH-PURE: pass.** The lint rule works only on the syntax tree, with no IO.
   - **ARCH-PURPOSE: pass.** The test that broke main is fixed, and the guard against the whole class is in the catalog, which runs over every tracked file.
   - **ARCH-EVOLVE: pass, not applicable.** No persisted format changed.

7. **Plan revision recommendations:**
   - The `## Plan` section is still an empty `- [ ]`. The quick flow allows that.
   - The Spec suggested a dedicated atlas page linked from `atlas/index.md`. The change uses the existing landing-train section of `sdlc-binary.md` instead, which meets the "create it if absent" wording. The Log could record that choice.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      sdlc-binary.md:2030 now names test-reads-workshop with incident #469. TestAtlasNamesEveryRule pins it: removing the backticks turned the test red, verified at the pinned head and then restored.
  - id: BR-2
    disposition: not-addressed
    note: |
      sdlc-binary.md:462 still says Plan: workshop/plans/000449-landing-train-plan.md, but the file is gone from workshop/plans. Change it to workshop/history/plans/.
  - id: BR-3
    disposition: not-addressed
    note: |
      sdlc-binary.md:457 still claims the test keeps the whole list in step, including Phases, but the test checks only backticked Event and Effect names.
  - id: BR-4
    disposition: not-addressed
    note: |
      Still literal-only: filepath.Join("..","..","workshop","plans",...) passes the rule. This stays acceptable under the Done-when.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2ed2f97d-e321-4c11-8d18-425e27c612fb","repository":"github.com/xianxu/ariadne","issue_id":"000469","recipe":"small-diff-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/e9b052e10f1f941a","reviewed_head":"7b2bff5878d86def1d2aa555929b5136b90e1ceb","slot":"ariadne:2","requested_at":"2026-10-19T18:45:01.198229Z"},"phase":"finished","pid":83073,"started_at":"2026-10-19T18:45:01.31092Z","finished_at":"2026-10-19T18:45:30.061751Z","session_id":"f7ca9c3a-9cb9-4173-9f7d-b91f549db18e","outcome":"returned"}] -->
