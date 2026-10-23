# Boundary Review — ariadne#561 (milestone M1)

| field | value |
|-------|-------|
| issue | 561 — Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost |
| repo | ariadne |
| issue file | workshop/issues/000561-landing-trains-cars-board-on-affected-tests-full-suite-once-per-train-measure-escapes-and-bisect-cost.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6318fc7bc700173acdcea67f3993d326f8947845..2e6f906cbd6703002ee909b8691d165781af53b1 |
| command | sdlc milestone-close --issue 561 --milestone M1 |
| reviewer | claude (anthropic; author anthropic,openai) |
| isolation | fenced |
| timestamp | 2026-10-23T01:34:49-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I've read the diff and the plan. Now checking the car readiness and help paths, then I'll run the tests.

Targeted tests pass. Next I'm rendering the help output and checking whether the new tests fail when the fix is reverted.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The M1 code does what the plan asks. `--boarding` is checked before any IO by a pure helper (`affectedModeFor`). In boarding mode, every full selection (a `full` glob, an unclassified path, an unloadable package graph, or a failed `graph_prepare`) refuses before anything runs or posts. The refusal names the reason, the `affected.json` repair, `--name suite --rev SHA` followed by `sdlc merge --yes`, and the serial-landing prerequisite. Ordinary affected runs keep their suite fallback. `DefaultMaxCars` drops from 8 to 2, and explicit caps keep their own value. `FormatTrainRampPolicy` is a single source that both `sdlc landing --help` and status render.

I ran the targeted tests on the pinned head and they pass: `go test ./cmd/sdlc/internal/landingline` and the named `cmd/sdlc` tests. I also put back the old behaviour twice to check the new tests can fail:
- Disabling the boarding refusal turns `TestAffectedBoardingNeverEscalatesToSuite` red.
- Setting `DefaultMaxCars` back to 8 turns `TestBoard/an_omitted_cap…` and `TestFormatTrainRampPolicy` red.

Both edits were reverted and the checkout is clean. The one thing to fix before shipping is that `README.md` still says the default cap is 8, which is now wrong, and doesn't mention `--boarding`.

### 1. Strengths
- **Refusal sits in the right place.** It is one `switch` arm in `runAffectedCheck` (`cmd/sdlc/localcheck_affected.go:83`). `affected.Plan` still owns selection and readiness isn't touched, as the plan required.
- **Bad flag combinations fail before any work.** `affectedModeFor` runs before any git or network call in `runLocalCheckCmd` (`cmd/sdlc/localcheck_cmd.go:143`), and its test (`TestLocalCheckBoardingFlag`) covers suite, mixed names and `--all`.
- **Recovery messages share one source.** `localCheckCommand` (`cmd/sdlc/cigate.go:237`) feeds both the `ciVerdict.refusal` text and the car-readiness "merge stopped" command. `landingConfigPath` now comes from `landingline.ConfigPath`, and `SerialLandingPrerequisite` is reused by both the policy text and the boarding refusal.
- **The boarding test checks real side effects.** It uses the native fixture and asserts the marker file, the consumer output and the posted statuses, not mocks. The new `affected boarding` train subtest shows three cars holding only affected records landing on the train's single suite.
- **Fixtures were updated honestly.** Multi-car fixtures now set `train_max_cars: 8` explicitly instead of silently depending on the old default.

### 2. Critical findings
None.

### 3. Important findings
- **README update appears missing for the default cap and `--boarding`.**
  - `README.md:661` still says "`train_max_cars` (default 8)". The code now uses 2 (`cmd/sdlc/internal/landingline/train.go:39`).
  - The affected-suite section around `README.md:847` doesn't mention `--boarding`, its refusal, or the seven-day manual ramp rule.
  - Fix: change it to "default 2 (#561)" with a one-line pointer to the ramp rule, and add a short `--boarding` paragraph matching `helptext/local-check.md:249-256`.

### 4. Minor findings
- `cmd/sdlc/cigate.go:675` still builds `sdlc local-check run --name %s` by hand. For the affected stand-in it would recommend a run without `--boarding`. Using `localCheckCommand(name)` fixes it (ARCH-DRY).
- No test checks that `sdlc landing --help` renders `{{TRAIN_RAMP_POLICY}}`, or covers `wrapHelpText` (`cmd/sdlc/main.go:74`). Only the status rendering is tested.
- `runLocalCheckCmd` passing `mode` through to `runAffectedCheck` isn't tested end to end. The tests call `runAffectedCheck` directly, so passing `affectedOrdinary` there would still pass them.
- From M1 onward, status and help describe measurements (escapes, denominator, unresolved coverage) that M2 hasn't built yet. That's acceptable if M2 ships soon; otherwise the rule points at evidence that doesn't exist yet.

### 5. Test coverage notes
- The boarding-refusal and default-cap tests fail when the change is reverted (confirmed by the two revert runs above).
- `graph_prepare` failure in boarding mode is covered only structurally: it goes through the same `graphErr` branch as "package graph unloadable". That's acceptable.
- The train test posts affected statuses directly instead of going through the CLI. Combined with the per-unit tests, that's acceptable for M1.

### 6. Architectural notes for upcoming work
- **ARCH-DRY:** pass, apart from the Minor at `cigate.go:675`.
- **ARCH-PURE:** pass. `affectedModeFor`, `boardingRefusal` and `FormatTrainRampPolicy` are pure, and IO stays in `productionAffectedDeps`.
- **ARCH-PURPOSE:** pass for M1. Every affected recovery path recommends `--boarding` except the Minor above.
- **ARCH-MOCK:** pass. Tests use native git fixtures and the stateful `trainFakeGH` / `ciFakeGH`.
- **ARCH-CONSTRAINTS:** pass. The refusal comes after graph loading, which costs the same as before, and no suite is launched.
- **ARCH-SECURE:** N/A. Nothing new reads untrusted input or handles secrets, and the `landing.json` parse is unchanged.
- **ARCH-ORDER:** pass. Nothing is kept between events; the mode is a per-call value.
- **ARCH-FUNERAL:** pass. A refusal leaves nothing behind: no record, status or attempt.
- **ARCH-EVOLVE:** pass, with a note. No persisted format changes. The meaning of an omitted `train_max_cars` changes from 8 to 2 for repos that already have trains on; the hub approved this and the atlas and help document it, and the README fix above completes that.
- For M2: the `BoardingProof` codec and `AdvanceMeasurement` are where ARCH-EVOLVE and ARCH-ORDER will carry the real weight.

### 7. Plan revision recommendations
None. The code matches the M1 rows of the Core concepts table: `affectedExecutionMode`, `DefaultMaxCars`, `TrainRampReviewWindow` and `FormatTrainRampPolicy` all exist at the stated paths.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: docs-drift-on-changed-default
    title: |
      README still documents train_max_cars default 8 and omits --boarding
    detail: |
      README.md:661 says "train_max_cars (default 8)" but landingline.DefaultMaxCars is now 2 (train.go:39); the affected-suite section near README.md:847 has no mention of --boarding, its full-selection refusal, or the seven-day manual ramp rule. Update to default 2 (#561) and add a short --boarding paragraph mirroring helptext/local-check.md.
  - id: new
    severity: Minor
    kind: defect
    family: recovery-command-single-source
    title: |
      cigate.go:675 hand-builds a local-check command bypassing localCheckCommand
    detail: |
      For the affected stand-in this would recommend a run without --boarding; use localCheckCommand(name).
  - id: new
    severity: Minor
    kind: defect
    family: derived-surface-untested
    title: |
      Help rendering of TRAIN_RAMP_POLICY and wrapHelpText has no test
    detail: |
      Only status asserts the policy text; nothing pins that sdlc landing --help renders it or that wrapping preserves the words.
  - id: new
    severity: Minor
    kind: defect
    family: cli-wiring-untested
    title: |
      runLocalCheckCmd passing mode into runAffectedCheck is not tested end to end
    detail: |
      Tests call runAffectedCheck directly; passing affectedOrdinary at localcheck_cmd.go:242 would keep every test green.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"1fba3996-3fc4-4c88-a97e-1502c04ab174","repository":"github.com/xianxu/ariadne","issue_id":"000561","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/cede9040195488b6","reviewed_head":"2e6f906cbd6703002ee909b8691d165781af53b1","slot":"ariadne:2","requested_at":"2026-10-23T08:29:48.512331Z"},"phase":"finished","pid":8850,"started_at":"2026-10-23T08:29:48.599274Z","finished_at":"2026-10-23T08:34:48.559738Z","session_id":"059fe0c8-acb9-44eb-bb7b-1232faa0e1ed","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T01:39:55-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 561 — Landing trains: cars board on affected tests; full suite once per train; measure escapes and bisect cost |
| repo | ariadne |
| issue file | workshop/issues/000561-landing-trains-cars-board-on-affected-tests-full-suite-once-per-train-measure-escapes-and-bisect-cost.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 6318fc7bc700173acdcea67f3993d326f8947845..ac4b48a62f178e996cfbf400aa4e933fb1eb063d |
| command | sdlc milestone-close --issue 561 --milestone M1 |
| reviewer | claude (anthropic; author anthropic,openai) |
| isolation | fenced |
| timestamp | 2026-10-23T01:39:55-07:00 |
| verdict | SHIP |

## Review

Disposed findings check out so far. Next I'm grepping the remaining doc consumers and running the targeted tests.

Prior fixes mostly hold; now confirming BR-4's test goes red when the mode wiring is reverted.

```verdict
verdict: SHIP
confidence: high
```

M1 does what the plan says. `--boarding` is wired from the CLI into `runAffectedCheck`. In boarding mode, a full selection (a `full` glob, an unclassified path, or an unavailable graph) is refused before anything runs or posts. The default cap is now 2, and explicit caps are left alone. `FormatTrainRampPolicy` is the only source for the ramp rule, and both `sdlc landing --help` and status print it. I ran the targeted tests at head and they pass: `go test ./cmd/sdlc/internal/landingline` ok, and the targeted `./cmd/sdlc` affected/boarding/config/status/car/train set ok (74s). Three of the four prior findings are fixed, and for BR-3 and BR-4 I reverted the fix and saw the tests fail. BR-2's code change is correct, but no test catches a regression, and two sibling recovery strings still build the command by hand. Both of those are Minor, so nothing blocks SHIP.

### 1. Strengths
- `cmd/sdlc/localcheck_affected.go:83`: the boarding refusal is a single `switch` arm placed ahead of the suite fallback, and `affected.Plan` still decides the selection. Graph failures reach the same arm, because `graphErr` is turned into `Full` first (line 76).
- `affectedModeFor` (`localcheck_affected.go:33`) is a pure check that runs before any git or IO in `runLocalCheckCmd`. It refuses `--boarding` with suite, mixed names or `--all` rather than ignoring the flag.
- `TestLocalCheckCmdBoardingRefusesFullSelection` (`localcheck_graphprepare_test.go:150`) tests the real CLI path against a native repo and the stateful fake. The same head is refused with the flag and falls back to the suite without it, which pins both modes.
- Train fixtures that need several cars now set `train_max_cars: 8` explicitly. That keeps them separate from the onboarding default, and `TestParseLandingConfig` pins the literal 2 plus explicit 1 and 8.
- `ConfigPath` and `SerialLandingPrerequisite` now live in one place in `trainpolicy.go`, and `landingConfigPath` points at that constant.

### 2. Critical findings
None.

### 3. Important findings
None.

### 4. Minor findings
- **BR-2 is still open: no regression test.** The change at `cigate.go:675` is right, now `localCheckCommand(name)`, and the branch can be reached when the gate runs local checks and affected is the stand-in. No test asserts this message, so putting back the hand-built string would leave everything green.
- **New, 2nd finding in family `recovery-command-single-source`.** `suiteenvironment.go:49` and `localcheck_reclassify.go:70` still build `sdlc local-check run --name %s` by hand. Since `isTestCheck` includes affected (`teststats.go:112`), both can tell a boarding car to rerun affected without `--boarding`.
  - Instead of patching these two sites, adopt a rule: every user-facing local-check run command comes from `localCheckCommand`, extended with an optional `--rev`.
  - Back the rule with a guard test that scans non-test `cmd/sdlc/*.go` for the literal `local-check run --name` outside that helper.
  - Prevalence: 2 sites left after the BR-2 fix.
- `internal/landingline/train.go:38`: the comment refers to `TrainRampPolicy`, but the function is `FormatTrainRampPolicy`.

### 5. Test coverage notes
- **BR-3:** I swapped `{{TRAIN_RAMP_POLICY}}` out of `landing.md` and `TestLandingHelpRendersTrainRampPolicy` failed.
- **BR-4:** I hard-coded `affectedOrdinary` at `localcheck_cmd.go:242` and `TestLocalCheckCmdBoardingRefusesFullSelection` failed. Both reverts were restored and the checkout is clean.
- Boarding with `graph_prepare` failing or timing out is not tested directly. It goes through the same `graphErr → Full` arm that the "package graph unloadable" case covers, so the risk is low.
- The `affected boarding` train subtest posts affected statuses directly instead of running boarding. That is acceptable: it is a train-gate test, and the CLI test covers the boarding side.

### 6. Architecture (each principle)
- **ARCH-DRY: flag (Minor).** The single source for local-check commands is not used everywhere yet; see the new finding.
- **ARCH-PURE: pass.** `affectedModeFor`, `FormatTrainRampPolicy` and `boardingRefusal` are pure, and their tests do no IO.
- **ARCH-PURPOSE: pass.** M1's purpose is delivered: no per-car suite, train gate unchanged, default of 2. The rest is M2.
- **ARCH-MOCK: pass.** Tests use native bare remotes, `ciFakeGH` and `trainFakeGH`, and production and tests share the same seam.
- **ARCH-CONSTRAINTS: pass.** Nothing new on any runtime path, and the cap of 2 bounds bisect cost.
- **ARCH-SECURE: pass.** The only new input is a boolean flag. No secrets or persisted input are involved.
- **ARCH-ORDER: pass.** The mode exists only for one call and holds no state between events.
- **ARCH-FUNERAL: pass.** A refusal leaves nothing durable behind: no status, run or log.
- **ARCH-EVOLVE: pass, with a note.** The `landing.json` format is unchanged. Leaving out `train_max_cars` now means 2 instead of 8, which changes behaviour for existing adopters. The hub approved this in the Spec, and README, atlas and help all document it.

### 7. Plan revision recommendations
None. The plan still matches the code for M1. The Core concepts entries that M1 owns (`affectedExecutionMode`, `DefaultMaxCars`, `TrainRampReviewWindow`, `FormatTrainRampPolicy`) are all at the stated paths. The other rows belong to M2.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:661 now says default 2 since 561 and keeps explicit caps, adds the manual seven-day ramp rule, and README.md:858 documents --boarding refusal and its alternative; matches DefaultMaxCars=2 and helptext/local-check.md.
  - id: BR-2
    disposition: not-addressed
    note: |
      cigate.go:675 now uses localCheckCommand(name), but no test asserts this message; restoring the hand-built string keeps every test green.
  - id: BR-3
    disposition: addressed
    note: |
      TestLandingHelpRendersTrainRampPolicy pins rendering and wrapping; removing the placeholder from landing.md turns it red (verified).
  - id: BR-4
    disposition: addressed
    note: |
      TestLocalCheckCmdBoardingRefusesFullSelection drives runLocalCheckCmd; hard-coding affectedOrdinary at localcheck_cmd.go:242 turns it red (verified).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: recovery-command-single-source
    title: |
      Recovery advice still hand-builds local-check run commands outside localCheckCommand
    detail: |
      2nd finding in this family. suiteenvironment.go:49 and localcheck_reclassify.go:70 format sdlc local-check run --name X by hand; isTestCheck includes affected, so a boarding car can be told to rerun affected without --boarding. Rule: every user-facing local-check run command derives from localCheckCommand (extended with optional --rev), enforced by a guard test that scans non-test cmd/sdlc sources for the literal outside that helper. Prevalence: 2 remaining sites after the BR-2 fix.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"378110e8-62aa-4de0-88b1-46e5ec174f8a","repository":"github.com/xianxu/ariadne","issue_id":"000561","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/b8f0a912985f0155","reviewed_head":"ac4b48a62f178e996cfbf400aa4e933fb1eb063d","slot":"ariadne:2","requested_at":"2026-10-23T08:36:12.375477Z"},"phase":"finished","pid":71608,"started_at":"2026-10-23T08:36:12.476121Z","finished_at":"2026-10-23T08:39:54.894024Z","session_id":"c8f69ecb-422a-41f0-aa19-195912bbc13d","outcome":"returned"}] -->
