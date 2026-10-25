# Boundary Review — ariadne#603 (milestone M1)

| field | value |
|-------|-------|
| issue | 603 — Measured actuals never block close: record exact, partial or unmeasured instead of refusing |
| repo | ariadne |
| issue file | workshop/issues/000603-measured-actuals-never-block-close-record-exact-partial-or-unmeasured-instead-of.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 41238256b57c82978e4ba83a4b17a8f7969ce0e8..ecaf2d6712d7cfc32e4fa119263dd4c66450b043 |
| command | sdlc milestone-close --issue 603 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T20:03:23-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm partway through the diff: I've read the classifier, the close wiring and the measurement seams, and now I'm checking the queue-degradation indexing and the gatestate and estimate changes.

Tracing is done; now I'm running the test suites for the affected packages.

The targeted batch didn't fail on an assertion: it hit Go's default 10-minute test timeout because the selection was too broad. I'm rerunning narrower subsets with a longer timeout.

While the broader batch runs, I'm writing a quick check of whether the summary writer can produce more coalesced gaps than its own validator accepts.

Both broad runs timed out, so I'm running one with `-v` to see which tests are slow and whether any of them were changed in this window.

Still no failures in the close batch, but it's also slow. I'm narrowing to just the test functions in the test files this window changed.

Every test in the changed test files passes (177s). Next, a quick mutation: removing the rule that makes an unknown queue attempt an upper bound, to check a test catches it.

The mutation turned `TestClassifyActual` red, and the checkout is restored. Last step: the README, atlas and help diffs.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

**Summary.** M1 does what the spec asks. Every refusal shape in the issue now records a confidence instead of refusing: unknown or detached branches, a source with no single owner, the source cap at the branch proof, an oversized snapshot, and incomplete queue evidence. `close` adopts every outcome, and `--no-actual` is no longer offered as the fix for a measurement failure. The code builds. All tests in the touched internal packages pass, and so does every test in the changed `cmd/sdlc` test files (177s). A mutation that drops the queue-unknown upper-bound rule turns `TestClassifyActual` red.

Two cheap issues should be fixed before the boundary, and they share one rule: a writer emits records bigger than its own validator accepts. When that happens, `ValidateShape` or the summary check fails, and a close that should be `exact` or `partial` silently becomes `unmeasured`. I confirmed the first one by running it: coalescing 5000 credits leaves 2952 widened gaps against a limit of 2048.

### 1. Strengths
- `classifyActual` (`cmd/sdlc/actualconfidence.go`) is the one pure mapping from a measurement to its confidence. It is shared by `sdlc actual`, close, milestone-close and correction. `Token()` and `String()` are kept apart, so frontmatter never stores reason text, and the cue regex accepts exactly the set `Token()` can produce.
- `Validate` is split into `ValidateShape` and `ValidateBounds`, and the bounds failure is a typed `OversizeError`. This fixes the pair#838 and #851 shape (the cap was checked before the sources were trimmed). `actualsourcecap_test.go` covers #838, #851 and #866 through the real resolver.
- Ambiguous owners are handled carefully (`actualbranch.go:176-187`): the claim stays filed under its real owner key, and only the shared root is marked ambiguous.
- The ledger change is additive. Corrected rows keep their older width, and rows without the new columns read as `exact`. `TestCalibrationWeightProperty` checks that a weight-0 row changes nothing and that an all-exact ledger gives today's numbers.
- Corrections still demand exact evidence (`actualcorrection.go:739`, plus the `!degrade` branches in `actualinputs.go`). So a correction row, which carries no confidence columns, cannot launder a partial measurement into `exact`.

### 2. Critical
None.

### 3. Important
- **The summary writer doesn't bound `Widened`.** `cmd/sdlc/internal/gatestate/measurement_summary.go` `CoalesceCredits` adds one widened gap per coalescing step, so there are `merged − limit` of them. `validateSummary` refuses more than `MaxSummaryCredits` widened gaps. Any summary with more than 4096 merged credits therefore fails in `summariseMeasurement`, and the close drops to `unmeasured`. That breaks D2 ("exact plus a summary") and the plan's claim that the bound sits on the writer (ARCH-FUNERAL: a cap on the reader with no matching bound on the writer is a cliff). Reproduced: 5000 credits give `out=2048 widened=2952`.
  - **Fix:** bound `widened` on the writer, for example by coalescing the widened list itself. Over-reporting lost minutes is the safe direction.
  - **Test:** add a regression test with more than `2*MaxSummaryCredits` credits that asserts `out.Validate()` passes.
- **Unattributed reasons aren't bounded where they are written.** `actualbranch.go` (`skip`) builds reasons from the worktree `root` and `capture.Unavailable` with no length limit. `MeasurementUnattributed.Validate` refuses anything over 400 bytes, so a long path or capture error turns the close `unmeasured`. `queueUnattributed` does truncate, but by slicing bytes (`reason[:397]`), which can cut a multi-byte character in half. There are now three truncation rules: 240 bytes in `confidenceReason`, 397 in `queueUnattributed`, and none in `skip`. That is an ARCH-DRY issue too.
  - **Fix:** one UTF-8-safe bounding helper in gatestate, called by every writer of `MeasurementUnattributed`, with a test that feeds it a reason over the limit.

### 4. Minor
- `summariseMeasurement` records `Summary.Bytes = 0` when `json.Marshal` of the full snapshot fails, for example past the structural cap. It should take the byte count from the `OversizeError` instead of making one up (ARCH-SECURE).
- Plan Task 10 says the pair#856, #812, #863 and #830 fixtures each run "through the close seam" and `classifyActual`. In fact they are `SelectBranchActivity` tests in activetime that assert `Unknown` is empty. That is adequate, but the plan overstates it.
- In the help text, "the line says how much the hours may overstate" only holds when the overlap is greater than zero.

### 5. Test coverage notes
- **Passed:** all tests in the touched internal packages (gatestate, estimate, activetime, project, processmanual), and every test in the changed `cmd/sdlc` test files.
- **Not run in full:** two broad `cmd/sdlc` batches hit the 10-minute foreground timeout. Each logged 215 passes and no failures before stopping. The timeouts landed in `TestLeftoverBranchIsLockedUntilCaughtUpThenWorks` and `TestCloseSpreadButTinyAtTheLimit`, neither of which this window changes.
- **Gap:** `computeOwnedActualFor` can produce `unmeasured` if summarising itself errors, and no test covers that path. The first Important finding is exactly that path.

### 6. Architectural notes
- **ARCH-DRY:** flag (reason truncation, above). Otherwise passes: one classifier and renderer, and the shared `validateCredits` and `overlappingCredits` helpers.
- **ARCH-PURE:** pass. The classifier, span math, coalescer and weighted calibration are pure and tested without IO; the IO changes only turn errors into records.
- **ARCH-PURPOSE:** pass. Every refusal shape is handled, and the calibration readers (forecast, converge, drift, train forecast) all derive from `CalibrationWeight`.
- **ARCH-MOCK:** pass. The existing seams are reused and no new external surface is added.
- **ARCH-CONSTRAINTS:** pass. Close is a batch, and coalescing is O(groups×spans) per step, which is fine at a few thousand credits.
- **ARCH-SECURE:** pass with a Minor (the fabricated `Bytes`). Reasons in the trailer are shown relative to the home directory and bounded.
- **ARCH-ORDER:** pass. Close is a single pass, and the evidence commit is still the one durable write.
- **ARCH-FUNERAL:** flag (the widened gaps).
- **ARCH-EVOLVE:** pass. `unattributed` and the summary keep unknown fields. Version 5 is refused by older readers, with a stated reason that it is unsafe to ignore. The ledger columns are appended.

### 7. Plan revision recommendations
- Add a `## Revisions` entry: the pair#856, #812, #863 and #830 fixtures are selection-level `SelectBranchActivity` tests, not close-seam runs.
- Once the first finding is fixed, add a `## Revisions` entry saying `Widened` is bounded on the writer (and how).

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: writer-bound-matches-validator
    title: |
      CoalesceCredits leaves Widened unbounded, so a large summary fails validation and the close drops to unmeasured
    detail: |
      measurement_summary.go CoalesceCredits adds one widened gap per coalescing step (merged minus limit), but validateSummary refuses more than MaxSummaryCredits; reproduced 5000 credits giving out=2048 widened=2952. summariseMeasurement then errors and computeOwnedActualFor records unmeasured, breaking D2 (exact plus summary). Bound widened on the writer (e.g. coalesce the widened list too) and add a test with more than 2x MaxSummaryCredits credits asserting out.Validate() passes.
  - id: new
    severity: Important
    kind: defect
    family: writer-bound-matches-validator
    title: |
      Unattributed reasons are not bounded at the writer, while the validator refuses anything over 400 bytes
    detail: |
      actualbranch.go skip() builds reasons from the worktree root and capture.Unavailable with no length limit, so a long one fails ValidateShape and turns the close unmeasured. queueUnattributed truncates by byte slicing (reason[:397]), which can cut a UTF-8 character. Three truncation rules exist (240, 397, none); use one UTF-8-safe helper in gatestate for every MeasurementUnattributed writer and test it with a reason over the limit.
  - id: new
    severity: Minor
    kind: defect
    family: no-fabricated-evidence
    title: |
      summariseMeasurement records Summary.Bytes as 0 when marshalling the full snapshot fails
    detail: |
      Past the structural encoder cap json.Marshal errors and Bytes stays 0, a made-up value; take the size from the OversizeError counts instead.
  - id: new
    severity: Minor
    kind: defect
    family: plan-claims-match-code
    title: |
      Plan Task 10 says the pair 856, 812, 863 and 830 fixtures run through the close seam; they are selection-level tests
    detail: |
      branch_resting_test.go and branch_staging_test.go call SelectBranchActivity and assert Unknown is empty; add a Revisions entry or route them through the close seam.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d9a119a5-2f9f-4aca-9291-9c95409dab52","repository":"github.com/xianxu/ariadne","issue_id":"000603","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/9dc63c63d475bfa4","reviewed_head":"ecaf2d6712d7cfc32e4fa119263dd4c66450b043","slot":"ariadne:2","requested_at":"2026-10-25T02:14:01.817818Z"},"phase":"finished","pid":50044,"started_at":"2026-10-25T02:14:01.953301Z","finished_at":"2026-10-25T03:03:22.42211Z","session_id":"4a3e7520-5ee7-44bf-8ac6-a69cc5790722","outcome":"returned"}] -->
