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
