# Boundary Review — ariadne#473 (whole-issue close)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0956e5e1c71bd6734f195929fc151e0c752dbc14..a3b1a8cb593559d6254cf0df09b0833e4fafcb46 |
| command | sdlc close --issue 473 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T03:12:02-07:00 |
| verdict | SHIP |

## Review

Start with BR-12 (the last commit).

Meanwhile, read the core code.

The targeted tests are still running, so I'm going through the CI-gate and flakes changes while I wait.

I'm still running the new `cmd/sdlc` tests. That package's test run gives no output, so I'm checking whether it's hanging.

I'm checking how the docs rule handles `.md` files that tests read, then the guard/republish paths.

```verdict
verdict: SHIP
confidence: medium
```

The open finding, BR-12, is fixed. `testStatsRows` (`cmd/sdlc/teststats_cmd.go:163-171`) now applies the `Runs > 0` filter to both the suite row and the affected row. `TestTestStatsRowsNeedRuns` pins this. I put back the old unconditional suite row in this checkout and the test failed (`affected only: [{Repo:pair Check: Runs:0 …} …]`); with the fix restored it passes. The rest of the change matches the Spec, Plan and Revisions: the pure planner and the `go list` seam, the `affected` runner with its full, empty, subset and no-change paths, the car stand-in that only the car sets, the success guard covering both checks, subset passes that never confirm a flake, a separate stats row per check, and the docs and atlas updates. All 20 new or changed tests I targeted pass, as do the `affected`, `flakes` and `jsonstrict` packages. The full `cmd/sdlc` package run hit its 600 s timeout: the train e2e test `TestLandingCarMergedWithoutAWordFinishes` got stuck on a git call. Separately, `internal/testfix` failed because the reviewer sandbox refused a `git init` outside its directories (#339). I can't tell from here whether the train e2e stall is also the sandbox or a real hang. Neither test touches #473 code. I found three Minor issues and nothing that blocks.

1. **Strengths**
   - `affected.Plan` is pure and fails closed. A path no package or group claims, a bad glob, or a stale group `go` root each forces Full (`plan.go:46-87`). An unloadable package graph also forces Full (`localcheck_affected.go:39-42`).
   - Package ownership is exact and a union: the package's own directory, its `testdata/`, and every package that embeds the file (`graph.go:34-50`). A real `go list` run over a fixture module checks this (`golist_test.go`).
   - The stand-in lives only in the car's facts (`landingcar.go:117-129`). Merge never sets it, so merge still requires `suite`. A broken `affected.json` makes the car fall back to requiring `suite` instead of stopping the train. The CI runner now refuses a check name it can't resolve instead of running an empty check (`cigate.go:649-654`).
   - One resolver: `affectedCheckDecl` is the only place the name `affected` becomes a check (`localcheck_affected.go:73-76`), so run and republish share it.
   - `flakes.Attempt.Subset` keeps a partial pass from confirming a fail→pass, but its failures still need explaining (`evidence.go:24-44`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `plan.go:30`: the docs rule runs before ownership. A `.md` file outside `cmd/` that a package embeds or keeps in `testdata/` counts as docs. If a branch changes only such files, `runAffectedCheck` doesn't load the graph and records a docs-only `affected` pass. This is the 4th finding in family `unclassified-path-fails-closed`. Rule: a path may be ignored only when no package owns it and no group matches it. The docs rule should apply only to unowned paths, and the graph should load whenever any path changed.
   - `plan:169`: the plan's ARCH-ORDER section says the record names the merge-base it planned against. The status only carries the scope, e.g. "3 packages, groups lua". This is the 3rd finding in family `plan-tracks-code`.
   - `jsonstrict` compares keys case-sensitively, but `encoding/json` matches field names case-insensitively. So `{"command":…,"COMMAND":…}` passes both checks and the last value wins. This is the 2nd finding in family `strict-persisted-format`.
   - Not filed: `savedAttemptRecord` doesn't keep `Scope`, so a republished `affected` status loses its "N packages" prefix. This is cosmetic only.

5. **Test coverage notes**
   - Each outcome has a test: the contract fixture, the full fallback, docs-only, the no-change refusal, car boarding with `affected` while merge still needs `suite`, a broken config, the guard spanning both checks, subset passes, and the stats rows.
   - Missing tests: an embedded or `testdata` `.md` file outside `cmd/`; a case-folded duplicate key in `affected.json`.
   - The full-package result for `cmd/sdlc` is unverified here, for the reason given above.

6. **Architectural notes**
   - Pass: ARCH-DRY (`isTestCheck`, `affectedCheckDecl`, the shared docs rule), ARCH-PURE, ARCH-MOCK (the real `go list` conformance fixture), ARCH-CONSTRAINTS (about 1.5 s of `go list` per run in a batch path), ARCH-ORDER (exact-sha records, no new stateful component), ARCH-FUNERAL (the worktree is cleaned up by `defer`; logs and attempts go in existing stores), ARCH-SECURE (config read from main and parsed strictly; the selection reaches the command through `env` argv), ARCH-EVOLVE (`RepoStats.Check` is additive with `omitempty`; refusing an unknown `affected.json` version is justified, because ignoring a new selector would skip tests).
   - ARCH-PURPOSE passes for code. The one-week pair measurement is still owned by the TL rollout (D6).
   - Root-level `go.mod` is owned by the root package when one exists. A repo whose `full` list leaves out `go.mod`/`go.sum` would select only the root package and its dependents. Consider forcing Full for those files in the planner itself.

7. **Plan revision recommendations**
   - Amend the ARCH-ORDER bullet: the record names the scope only, or add the merge-base to the status.
   - If the docs-rule finding is fixed, revise Core concepts rule 1: docs are ignored only when no package owns them.

```findings
dispose:
  - id: BR-12
    disposition: addressed
    note: |
      testStatsRows filters both rows on Runs>0; reverting the suite filter turns TestTestStatsRowsNeedRuns red (verified).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: unclassified-path-fails-closed
    title: |
      Docs rule runs before ownership, so an embedded or testdata .md outside cmd/ records a docs-only affected pass
    detail: |
      plan.go:30 skips isDoc paths before owners() and runAffectedCheck skips the graph when publishGateHasCodeSurface is false; churn.IsEmbedded only exempts cmd/. 4th in family: rule is "a path may be ignored only when no package owns it and no group matches it" - apply isDoc only to unowned paths and load the graph for any nonempty change set.
  - id: new
    severity: Minor
    kind: defect
    family: plan-tracks-code
    title: |
      Plan ARCH-ORDER claims the affected record names its merge-base; the status carries only the scope
    detail: |
      plan line 169 vs affectedScope/statusDescription (localcheck_affected.go:79, localchecks.go:315). 3rd in family: rule is "each plan claim about a code property is delivered or revised in Revisions" - sweep Core concepts and ARCH sections against code at close.
  - id: new
    severity: Minor
    kind: defect
    family: strict-persisted-format
    title: |
      affected.json strict parse misses case-folded duplicate keys (command vs COMMAND), last one wins
    detail: |
      jsonstrict.RejectDuplicateKeys compares keys case-sensitively but encoding/json matches field names case-insensitively, so DisallowUnknownFields accepts both. Rule: a strict decode rejects any key that is not an exact tag match.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"05bc1e1c-46c6-4788-b8aa-2d9f40b49b0e","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"milestone-review","round":6,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e99f14f8d5df9597","reviewed_head":"a3b1a8cb593559d6254cf0df09b0833e4fafcb46","slot":"ariadne:2","requested_at":"2026-10-20T09:46:35.868741Z"},"phase":"finished","pid":81339,"started_at":"2026-10-20T09:46:35.988473Z","finished_at":"2026-10-20T10:12:01.910331Z","session_id":"2ee48323-5fad-4b4e-a672-b1a99ccf4fce","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T03:27:17-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 473 — test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B') |
| repo | ariadne |
| issue file | workshop/issues/000473-test-economy-affected-suite-as-close-evidence-the-full-suite-runs-once-in-the-train-or-at-merge-option-b.md |
| boundary | whole-issue close |
| milestone | — |
| window | 87f7ad03ff6a853b8b4106bccbb894b07308160b..ed4b83b402a5a80103c576efcde93f78811e6dcf |
| command | sdlc close --issue 473 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | checkout |
| timestamp | 2026-10-20T03:27:17-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Progress: the diff is read. Next I'm checking the CI-gate docs rule and running the tests.

Tests pass. Now reverting each fix to confirm its regression test goes red.

The reverts all fail as expected. Last, checking whether the case-fold fix covers non-ASCII folding.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The changes since the close review are tight and in scope. BR-13 is fixed for the whole class, not just the one file the finding cited. `affected.Plan` now asks who owns a path before it applies the docs rule. `runAffectedCheck` loads the package graph for every non-empty change set. The README, atlas, helptext and lessons all describe the new rule. BR-14 is fixed by a `## Revisions` entry, and that entry only records the fix; it does not widen the scope.

BR-15 is only partly fixed. Duplicate keys that differ in ASCII case are now refused, but the code folds case with `strings.ToLower`, and `encoding/json` folds differently. I confirmed a gap with a probe: `{"groups":["a"],"groupſ":["b"]}` passes `DisallowUnknownFields`, the last value wins (`[b]`), and `strings.ToLower` does not treat the two keys as equal. It's Minor and doesn't block, but the fix is one line, so I'm asking for it before shipping.

**Verification.** I ran the affected and jsonstrict package tests, and the `cmd/sdlc` affected, local-check, CI-gate and car tests; all pass. I then reverted each fix and re-ran its tests:
- reverting `plan.go` turns 3 new `TestPlan` rows red;
- reverting `jsonstrict.go` turns `TestRejectDuplicateKeys` red;
- reverting `localcheck_affected.go` turns `TestAffectedOwnedDocRunsItsPackage` red.

So every fix has a test that fails without it.

**1. Strengths**
- `cmd/sdlc/internal/affected/plan.go:57` applies the docs rule only to a path nothing claims, which is exactly the rule BR-13 stated. The empty result now comes from "no package or group claimed anything" (`:61`), not from a separate `code` flag.
- `cmd/sdlc/internal/affected/plan_test.go:60-63` adds one row per way a package can own a doc: an embedded file outside `cmd/`, a testdata file, and a file in the package directory.
- `cmd/sdlc/localcheck_affected_test.go:155` tests the change end to end: a doc in `leaf/` sets the env contract to `leaf`, `mid` and `tonly`.
- If the graph fails to load, even for a docs-only diff, the run falls back to the full suite. It never records a pass that nothing ran for.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- BR-15 is not fully fixed. `cmd/sdlc/internal/jsonstrict/jsonstrict.go:55` folds keys with `strings.ToLower`, which misses non-ASCII folds that `encoding/json` matches, such as `ſ`→`s` and the Kelvin sign. A key spelled that way can still fill a field twice. The fix: compute the key with the same folding `encoding/json` uses (`strings.ToUpper(strings.ToLower(k))`, or compare with `strings.EqualFold` against keys already seen), and add a test row for `groupſ`.

**5. Test coverage:** each fix has a test that fails without it (shown above). The one gap is a non-ASCII fold case for jsonstrict.

**6. Architecture**
- **ARCH-DRY: pass.** The docs rule still comes from the single `publishGateHasCodeSurface` predicate.
- **ARCH-PURE: pass.** The ownership logic stays in the pure `Plan`; the graph is still loaded through the injected seam.
- **ARCH-PURPOSE: pass for BR-13, flag for BR-15.** BR-13's fix covers every kind of owned doc. BR-15's fix covers ASCII only, so a sibling case remains (see the BR-15 finding below).
- **ARCH-EVOLVE: pass.** The duplicate-key check got stricter; no known-field strictness was weakened.

`cigate.go:556` and `landingtrain.go:509` still use the old publish-gate docs rule. That code predates this window, so I did not count it as a finding.

**7. Plan revisions:** none beyond the entry already added. If BR-15 is fixed as above, its line in that entry should say keys are folded the way `encoding/json` folds them.

```findings
dispose:
  - id: BR-13
    disposition: addressed
    note: |
      plan.go:57 applies isDoc only to unclaimed paths; runAffectedCheck always loads the graph; plan_test rows plus TestAffectedOwnedDocRunsItsPackage go red when the fix is reverted.
  - id: BR-14
    disposition: addressed
    note: |
      Plan Revisions entry (2026-10-20) changes the ARCH-ORDER claim to the status scope (selection), which matches statusDescription.
  - id: BR-15
    disposition: not-addressed
    note: |
      ASCII case folding now refuses, but strings.ToLower differs from encoding/json folding; a probe showed {"groups":..,"groupſ":..} decoding with the last value winning. Fold keys as encoding/json does and add a test row.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"70a41027-e73a-4367-97e3-94daa74f1455","repository":"github.com/xianxu/ariadne","issue_id":"000473","recipe":"fix-review","round":7,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/264af91e9c4815db","reviewed_head":"ed4b83b402a5a80103c576efcde93f78811e6dcf","slot":"ariadne:2","requested_at":"2026-10-20T10:15:43.573064Z"},"phase":"finished","pid":19683,"started_at":"2026-10-20T10:15:43.660056Z","finished_at":"2026-10-20T10:27:16.934227Z","session_id":"41698dae-c926-4527-aa24-84a591d7b462","outcome":"returned"}] -->
