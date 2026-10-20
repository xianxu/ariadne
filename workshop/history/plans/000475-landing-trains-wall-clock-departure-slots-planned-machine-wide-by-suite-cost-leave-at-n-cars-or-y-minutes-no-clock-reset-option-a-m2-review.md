# Boundary Review — ariadne#475 (milestone M2)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 2fa3a152e5eeb43923578c69bc4b3d5397337970..d67b8db5ac58de96a3c1cab4ee5cc461d14deee2 |
| command | sdlc milestone-close --issue 475 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T05:27:53-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I've read the M2 diff and am now checking the README and running the tests.

Found a likely status/admission divergence; now running the targeted tests to confirm the suite is green.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M2's pure core is solid. `DecodePlan` is strict and keeps unknown members, `PackOffsets` is deterministic and reproduces the operator's example layout, and the `since_stamp` change closes the stale-carry family as a rule rather than one instance. The landingline tests, the targeted `cmd/sdlc` tests (`LandingPlan|ScheduleSource|PlannedRepos|LandingStatus|TestStats|ParseLandingConfig`) and `go vet ./cmd/sdlc/...` all pass at the head. Four things need fixing first:
- **Status ignores the plan.** `sdlc landing status` builds its schedule without the machine plan, but its own comment and the atlas both say it shows "the same decision admission makes". Once a plan exists, it reports a different next departure than admission uses.
- **No test for the fleet inputs.** T9's "golden test over a fixture fleet" is ticked, but `fleetPlanInputs` has no test at all.
- **README.** It doesn't mention `sdlc landing plan`.
- **Silent overflow.** `PackOffsets` can place suites over the machine's slot count without any warning.

None of these change how landings run, and each is cheap to fix.

1. **Strengths**
   - `cmd/sdlc/internal/landingline/plan.go:50-118`: the decoder is strict. It refuses trailing data, non-objects, wrong types, a missing `offset_minutes`, and out-of-range offsets. It keeps unknown members at the top level and in each repo entry, and `MarshalJSON` never lets an unknown member override a known one. `TestMachinePlanCodec` covers a real round trip.
   - `PackOffsets` (`plan.go:183-227`) is pure, sorts heaviest first with ties broken by name, and has two kinds of test. The golden test matches the operator's example (ariadne :00; pair and parley :15), and a 500-case seeded property test checks that load never exceeds the slots. I re-derived the golden spans by hand: 15, 20 and 10 minutes.
   - `runLandingPlan --write` refuses to overwrite an unreadable plan, keeps other repos' entries and unknown members, and writes atomically via tmp+rename. The test covers all three behaviours.
   - `SinceStamp.Equal(r.At)` (`train.go:276`) states the rule "trust a derived field only on a record written by an sdlc that maintains it". It covers every reset an older writer can't apply, and `TestStaleSinceFromAnOlderSdlc`'s "ready again, same head" case would fail under the earlier `since_head` check.
   - Splitting `buildTestStatsReport` and `fleetRepoPaths` out of `runTestStats` is a clean reuse with no change in behaviour (ARCH-DRY pass).

2. **Critical findings:** none.

3. **Important findings**
   - **Status ignores the machine plan** (ARCH-DRY, ARCH-PURPOSE). `cmd/sdlc/landingcmd.go:182` calls `readLandingSchedule` directly, so `sched.Offset` is always nil. Admission gets its schedule from `landingScheduleSource` (`landingqueue.go:248-258`), which applies the plan. Example with offset :15, interval 60:
     - a full train leaves at 08:40;
     - admission's next departure is the 09:15 slot;
     - status prints 09:40 ("interval").
     
     Status also never prints the state table's "no machine plan…" or "[!] plan … (reason)" lines. **This is the 2nd finding in family `shared-helper-not-repeated-predicate`.** The rule: every consumer that needs a repository's effective schedule gets it from one constructor that combines the config and the plan. For example, `effectiveSchedule(git, rev, repo) (Schedule, planNote, error)`, called by both `landingScheduleSource` and `runLandingStatus`. The sweep covers every caller of `readLandingSchedule`:
     - `landingqueue.go:237` is the source and should move to the constructor;
     - `landingcmd.go:182` is status and should move to the constructor;
     - `landingplan.go:181` legitimately needs only the config.
     
     Add a status test that uses a plan file and an early departure.
   - **T9's fixture-fleet test is missing, and other plan claims don't match delivery** (ARCH-MOCK, ARCH-PURPOSE). `fleetPlanInputs` (`landingplan.go:161-202`) has no test. Its untested logic is the trains-on filter, the weight read from local-checks, the 15-minute default when there are no samples, and the join between test-stats rows and fleet names. It hard-codes `execGitRunner{}`, so there is no seam for a fake. The same kind of mismatch appears in three other places:
     - T7 claims path resolution via `WF_TEST_LOCK_FILE`, but every test sets it to "".
     - T9 says the repo list comes from `sdlc fleet inventory`, but the code uses `fleetRepoPaths`.
     - The issue's M2 checkbox claims "live dry run", but T11 is unchecked.
     
     **This is the 4th finding in family `plan-claim-matches-delivery`.** The rule: a task is ticked only when each artifact it names exists. Each named test must be in the tree and exercise the named entity, and each named data source must be the one the code calls. Otherwise the plan text is revised in the same commit. Sweep T6–T11 and the issue's M2 line against that rule, not just T9.
   - **README doesn't cover the new subcommand.** README.md:521-557 describes the departure rule but not `sdlc landing plan [--write]`, the `landing-plan.json` location, or `WF_TEST_SUITES`. **This is the 2nd finding in family `user-docs-track-new-surface`.** The rule: every new subcommand, flag, config key, env var or state-file path added in a range gets a README line in the same range. Sweep `newLandingPlanCmd`'s flags and `machinePlanPath`/`machineSuiteSlots`' env inputs.
   - **Packing can exceed the slots silently** (ARCH-CONSTRAINTS). When no offset fits, `PackOffsets` takes the offset with the lowest peak (`plan.go:208-216`) even if that peak is over `slots`. Neither `planFromCosts` nor `landing plan` warns, so the plan's "≤ slots" promise can break unseen. The property test skips exactly that case (`total > 60`). Fix: return each placement's peak and have `planFromCosts` emit `[!] <repo> placed at :MM over capacity (peak N > slots)`. Then test an infeasible input.

4. **Minor findings**
   - Existing trains-on e2e tests call `landingScheduleSource` without isolating `XDG_STATE_HOME`/`WF_TEST_LOCK_FILE`, so they read the real machine plan. That's not hermetic (ARCH-SECURE). Set both to a temp dir in the shared landing fixture.
   - `filepath.Base(t.Root)` (`landing.go:388`) and `fleetRepoPaths`' `filepath.Base(p)` are two separate derivations of the plan's repo key. A shared `planRepoName` helper would make the join explicit.
   - `strictObject` decodes into a map, so a duplicate key silently keeps the last value. ARCH-EVOLVE asks for duplicate keys to stay strict.
   - With no train-enabled repos, the header prints "every 0m". When intervals differ, the plan packs on the longest and a shorter-interval repo's offset is taken modulo its own interval, so windows can collide; this is only warned.
   - `Offset()` takes an offset larger than the interval modulo it, while the plan says an offset outside `0..interval-1` is an error. That needs a plan revision.
   - `--write` always uses the same `path+".tmp"` name, so two concurrent writes would race, and a failed rename leaves the tmp file behind. This is low risk because only the steward runs it.

5. **Test coverage notes**
   - Pure entities are tested without IO.
   - `TestScheduleSourceReadsThePlan` covers placement, a repo with no entry, and the warn-once fallback.
   - `TestPlannedReposNeverDepartInEachOthersWindow` runs `DepartureDue` on a schedule built by hand, not through `landingScheduleSource`→`trainAdmission`. Together with the source test, that's acceptable.
   - Gaps: `fleetPlanInputs`; resolving the plan path via `WF_TEST_LOCK_FILE`; `PackOffsets` with infeasible costs; status with a plan file present.

6. **Architectural notes for upcoming work**
   - ARCH-DRY: flag (status assembles the schedule separately).
   - ARCH-PURE: pass (`DecodePlan`, `PackOffsets` and `planFromCosts` are pure; the glue is thin).
   - ARCH-PURPOSE: flag (status doesn't use the plan).
   - ARCH-MOCK: flag (`fleetPlanInputs` calls `execGitRunner{}` directly, with no seam).
   - ARCH-CONSTRAINTS: flag (silent over-capacity packing). The O(waiters) poll cost and the 30 s cached plan read are fine.
   - ARCH-SECURE: pass, with the Minor hermeticity note. The plan is parsed into typed values and an invalid plan falls back visibly.
   - ARCH-ORDER: pass. The decision holds no state between events because it is recomputed from the events log, the waiters and the clock under `head.lock`.
   - ARCH-FUNERAL: pass. It's one file per machine, overwritten in place, with entries bounded by fleet size.
   - ARCH-EVOLVE: pass. Unknown members survive both top-level and nested rewrites. Refusing another version only falls back to the interval, never refuses service; the duplicate-key nit is under Minor findings.

7. **Plan revision recommendations**
   - `## Revisions` entry: "M2: the plan offset is taken modulo the interval (not refused) when the interval changes under the plan. T9 reads repos via `fleetRepoPaths` (fleet discovery plus primary), not `sdlc fleet inventory`. T11's live dry run is still open, so the issue's M2 'live dry run' claim waits on T11."
   - Once fixed: "status reads the same effective schedule as admission (config + plan) through one constructor."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: shared-helper-not-repeated-predicate
    title: |
      landing status builds its schedule without the machine plan, so it reports a different next departure than admission
    detail: |
      landingcmd.go:182 calls readLandingSchedule directly (Offset always nil); admission uses landingScheduleSource, which applies the plan (landingqueue.go:248). Example: offset 15, a full train leaves 08:40; admission's next departure is the 09:15 slot, status says 09:40. The state table's plan-missing and plan-unreadable lines are never printed. 2nd finding in this family. Rule: every consumer of a repo's effective schedule gets it from one constructor that combines config and plan. Sweep every readLandingSchedule caller: landingqueue.go:237 and landingcmd.go:182 move to it; landingplan.go:181 needs config only.
  - id: new
    severity: Important
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      T9's fixture-fleet golden test is ticked but fleetPlanInputs has no test; T7 and T9 sources and the M2 live dry-run claim also don't match delivery
    detail: |
      fleetPlanInputs (landingplan.go:161) has no test: trains-on filter, local-checks weight, the no-sample default, the test-stats name join. It hard-codes execGitRunner{}, so it can't be faked. Also: T7's WF_TEST_LOCK_FILE path is untested (tests set it to ""); T9 names sdlc fleet inventory but the code uses fleetRepoPaths; the issue's M2 checkbox claims a live dry run while T11 is unchecked. 4th finding in this family. Rule: tick a task only when each named artifact (test, data source, live step) exists as named; otherwise revise the plan text in the same commit. Sweep T6-T11 and the issue's M2 line.
  - id: new
    severity: Important
    kind: defect
    family: user-docs-track-new-surface
    title: |
      README update appears missing for sdlc landing plan [--write], the landing-plan.json path and WF_TEST_SUITES
    detail: |
      README.md:521-557 describes the departure rule but not the new subcommand or its state file. 2nd finding in this family. Rule: every new subcommand, flag, config key, env var or state-file path in a range gets a README line in the same range. Sweep newLandingPlanCmd's flags, plus the env inputs of machinePlanPath and machineSuiteSlots.
  - id: new
    severity: Important
    kind: defect
    family: silent-out-of-envelope
    title: |
      PackOffsets can place suites over the machine's slot count and landing plan never warns (ARCH-CONSTRAINTS)
    detail: |
      When no offset fits, plan.go:208-216 takes the lowest-peak offset even if the peak is over slots; planFromCosts emits no warning, and the property test skips exactly that case (total > 60). Return each placement's peak, warn when it exceeds slots, and test an infeasible input.
  - id: new
    severity: Minor
    kind: defect
    family: tests-isolate-user-state
    title: |
      Trains-on e2e tests read the real user's landing-plan.json via landingScheduleSource
    detail: |
      Only the two new plan tests set XDG_STATE_HOME and WF_TEST_LOCK_FILE. Set both to a temp dir in the shared landing fixture.
  - id: new
    severity: Minor
    kind: defect
    family: strict-known-fields
    title: |
      strictObject silently accepts duplicate JSON keys in the plan file (last one wins)
    detail: |
      ARCH-EVOLVE asks for duplicate keys to stay strict. Detect them with a token-level decode.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"40efdd64-4cd8-46fb-84f7-9868fcb0db5a","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/2e9a64e93116eecf","reviewed_head":"d67b8db5ac58de96a3c1cab4ee5cc461d14deee2","slot":"ariadne:4","requested_at":"2026-10-20T12:24:38.408162Z"},"phase":"finished","pid":19138,"started_at":"2026-10-20T12:24:38.568946Z","finished_at":"2026-10-20T12:27:53.021787Z","session_id":"390f745e-756b-46cd-ac8f-b0502e20737e","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T06:01:08-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 2fa3a152e5eeb43923578c69bc4b3d5397337970..30d34a5a4f7cf329e6bc570f760f6669454efb3b |
| command | sdlc milestone-close --issue 475 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T06:01:08-07:00 |
| verdict | REWORK |

## Review

Checking the round-1 fixes against the code; running tests next.

All landing/plan tests pass; now reverting individual fixes to confirm their regression tests go red.

```verdict
verdict: REWORK
confidence: high
```

Most round-1 fixes hold up. I reverted four of them one at a time, and each regression test failed without its fix: the status plan line (BR-8's test), duplicate keys (BR-13) and the overload report (BR-11). All tests matching `Landing|Plan|Machine|TestStats|Fleet` in `./cmd/sdlc` pass, and so does `./cmd/sdlc/internal/landingline`. The e2e set is slow: it took two runs of about 7 and 9 minutes, with nothing failing. One thing blocks SHIP. BR-8 is only fixed when `--repo` is an absolute path. With the default `--repo .`, status still disagrees with admission. `runLandingStatus` passes `filepath.Base(repo)` as the plan key, and `filepath.Base(".")` is `"."` (I checked this). So a plain `sdlc landing status` run inside the repo prints "the machine plan has no entry for ." and the interval-based next departure. Admission uses `filepath.Base(id.WorktreeRoot)` and gets the plan slot. The test passes because it gives an absolute `--repo`. That makes it the same class BR-8 named: two callers each work out an input to the effective schedule in their own way.

**1. Strengths**
- `effectiveSchedule` (`cmd/sdlc/landingconfig.go:96`) is a single constructor that both admission (`landingqueue.go:236`) and status (`landingcmd.go:183`) call. `fleetPlanInputs` keeps reading the setting alone, with a comment saying why.
- `fleetPlanInputs` now takes its sources as arguments, and `TestFleetPlanInputs` covers the trains-on filter, the declared weight, the default when there are no samples, and the name join, using a fake git.
- `PackOffsets` now returns the overloaded placements with their peaks, and `TestPackOffsetsNamesAnOverload` fails if that is removed.
- The duplicate-key check reads tokens at every level. Three new decode cases cover the top level, `repos` and an entry.
- `isolateMachinePlan` is called in the shared `landingFixture`, so `integrationFixture` and the train e2e tests get it too. The unit tests that call `landingScheduleSource` all have trains off, so they never read the plan.

**2. Critical findings:** none.

**3. Important findings**
- **BR-8 not-addressed** (`cmd/sdlc/landingcmd.go:183`, ARCH-DRY / ARCH-PURPOSE). The plan key is still worked out separately at each call site: `filepath.Base(repo)` in status, where it can be `.` or a subdirectory, and `filepath.Base(t.Root)` at `landing.go:388`. Fix: have `effectiveSchedule` take the checkout root and derive the key with one shared helper. Status should first resolve `--repo` to the worktree top level (`git rev-parse --show-toplevel`). Add a status test that runs with `--repo .` from inside the fixture repo, or from a subdirectory.

**4. Minor findings**
- `planFromCosts` builds the "does not fit" warning (`landingplan.go:53-61`), but no test checks it. Deleting the loop leaves every test green. Add an overloaded `stubPlanInputs` case to `TestLandingPlanCommand` that checks for `"does not fit"`.

**5. Test coverage notes**
- `TestLandingStatusShowsTrainsAndReadiness` only uses an absolute `--repo`, so it misses the default invocation.
- T9's "golden test over a fixture fleet" is covered by two tests together: `TestLandingPlanCommand` (stub inputs, checks output) and `TestFleetPlanInputs` (fake fleet). I accept that.

**6. Architectural notes**
- ARCH-DRY: flagged; the plan key has two derivations (the BR-8 note above).
- ARCH-PURE: pass. `PackOffsets` and the plan codec are pure; the plan inputs are now injected.
- ARCH-PURPOSE: flagged, for the same partial BR-8 sweep.
- ARCH-MOCK: pass. The fake git and the injected path are the same seams production uses.
- ARCH-CONSTRAINTS: pass. Overloads are reported, apart from the untested warning noted above.
- ARCH-SECURE: pass. The plan is parsed strictly, and an unreadable plan is reported and never overwritten.
- ARCH-ORDER: pass. The schedule source still uses the mutex and TTL cache, and the warning flags are set once.
- ARCH-FUNERAL: pass. The plan is one file per machine that `--write` replaces.
- ARCH-EVOLVE: pass. Unknown members are kept, and duplicate keys are now rejected.

**7. Plan revision recommendations**
- In the M2 round-1 Revisions entry for BR-8, say that the constructor also owns the plan key (derived from the checkout's top level), so no caller passes its own name.

```findings
dispose:
  - id: BR-8
    disposition: not-addressed
    note: |
      effectiveSchedule is shared, but status keys the plan on filepath.Base(repo) and --repo defaults to ".", so the key is "."; admission uses Base(WorktreeRoot). The default `sdlc landing status` still reports the interval departure when admission uses the slot. Derive the key inside the constructor from the resolved top level, and test --repo . or a subdirectory.
  - id: BR-9
    disposition: addressed
    note: |
      fleetPlanInputs is injectable and TestFleetPlanInputs covers the filter, weight, no-sample default and name join. TestMachinePlanPath covers WF_TEST_LOCK_FILE and XDG. T9 now names fleetRepoPaths, and the live dry run moved to an unchecked line in the issue.
  - id: BR-10
    disposition: addressed
    note: |
      README.md:540-554 documents sdlc landing plan [--write], the plan path, WF_TEST_LOCK_FILE, WF_TEST_SUITES and the fallback; train_max_wait_minutes is at :523.
  - id: BR-11
    disposition: addressed
    note: |
      PackOffsets returns overloads with peaks; reverting the bestPeak > slots check fails TestPackOffsetsNamesAnOverload. The planFromCosts warning text has no test (new Minor below).
  - id: BR-12
    disposition: addressed
    note: |
      isolateMachinePlan is in landingFixture, which integrationFixture also uses; the remaining unit callers of landingScheduleSource have trains off.
  - id: BR-13
    disposition: addressed
    note: |
      The token-level duplicate check works at every level; reverting it makes TestMachinePlanCodec accept all three new duplicate cases.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: silent-out-of-envelope
    title: |
      landing plan's "does not fit" warning has no test; deleting it keeps the suite green
    detail: |
      This is the 2nd finding in family silent-out-of-envelope. The rule: every out-of-envelope signal is pinned by a test at the surface the operator reads, not only in the pure helper. PackOffsets' over map is tested; the planFromCosts loop that turns it into the warning (landingplan.go:53-61) is not. Add an overloaded stubPlanInputs case to TestLandingPlanCommand that checks for "does not fit".
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5a3f636c-4d8b-4c38-8982-9f6fc759b654","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/e39ea9a134db5f9c","reviewed_head":"30d34a5a4f7cf329e6bc570f760f6669454efb3b","slot":"ariadne:4","requested_at":"2026-10-20T12:31:44.411385Z"},"phase":"finished","pid":9152,"started_at":"2026-10-20T12:31:44.572169Z","finished_at":"2026-10-20T13:00:53.351977Z","session_id":"6a22593e-ed3c-41ae-bec6-c3e918be71de","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T06:09:41-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 475 — landing trains: wall-clock departure slots planned machine-wide by suite cost; leave at N cars or Y minutes; no clock reset (option A) |
| repo | ariadne |
| issue file | workshop/issues/000475-landing-trains-wall-clock-departure-slots-planned-machine-wide-by-suite-cost-leave-at-n-cars-or-y-minutes-no-clock-reset-option-a.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 2fa3a152e5eeb43923578c69bc4b3d5397337970..61af3a94f9a6f05669bbadd80bd5ab4d9d44aebf |
| command | sdlc milestone-close --issue 475 --milestone M2 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-20T06:09:41-07:00 |
| verdict | SHIP |

## Review

Now revert-checks for BR-8 (status uses the effective schedule) and BR-14 (the overload warning).

I've finished checking the code and tests; writing up the verdict now.

```verdict
verdict: SHIP
confidence: high
```

Both open findings are fixed, and I confirmed each by reverting the fix in this checkout. When `landing status` went back to the bare `readLandingSchedule`, `TestLandingStatusShowsTrainsAndReadiness` failed with "status ignores the machine plan". When the "does not fit" loop in `planFromCosts` was removed, `TestLandingPlanCommand` failed. The landingline package passes, and so do the targeted `cmd/sdlc` landing, plan and test-stats tests (154 s). The M2 work matches the Spec. The plan file is strictly typed, keeps unknown members, refuses repeated keys and degrades visibly to the old clock. Packing is deterministic and names any overload. Admission and status now share one effective-schedule constructor. The only new finding is Minor: two sentences in the plan describe behaviour the code doesn't have. It does not block SHIP.

1. **Strengths**
   - `effectiveSchedule` (`cmd/sdlc/landingconfig.go:93`) is the single source for the schedule. Admission (`landingqueue.go:234`) and status (`landingcmd.go:182`) both use it. `fleetPlanInputs` reads only the setting, and its comment explains why.
   - `DecodePlan` and `strictObject` (`internal/landingline/plan.go:49-157`) are strict about known fields and duplicate keys at every level, and keep unknown members on rewrite. `TestMachinePlanCodec` covers all of this.
   - `PackOffsets` returns the overloads instead of failing silently. A property test checks that whenever the costs can fit, the summed weight stays within the slots at every minute.
   - `fleetPlanInputs` takes all three sources injected (ARCH-PURE/ARCH-MOCK). `TestFleetPlanInputs` checks the trains-on filter, weight, no-sample default and the name join without any IO.
   - The plan key is the base name of `rev-parse --show-toplevel`. That gives the repo name for `<fleet>/worktree/<repo>-slotN/<repo>` checkouts and lines up with `fleetRepoPaths`. A test covers status run with `--repo .` from a subdirectory.
2. **Critical:** none.
3. **Important:** none.
4. **Minor**
   - Two plan sentences don't match the code:
     - The plan says an offset outside 0..interval-1 is an error. `MachinePlan.Offset` (`plan.go:38`) accepts 0..1439 and reduces it modulo the interval.
     - The plan says a failed read warns "once per TTL". `landingScheduleSource` warns once for the life of the source.
5. **Test coverage**
   - The plan-wired admission path (`TestScheduleSourceReadsThePlan`) and the planned-window fixture (`TestPlannedReposNeverDepartInEachOthersWindow`) are both real-logic tests. The fixture calls `DepartureDue` directly rather than going through `trainAdmission`. That's acceptable because the schedule source has its own test.
   - The `SinceStamp` change is pinned for both stale cases (a new head, and ready again at the same head), plus the case where a maintained `Since` is trusted.
6. **Architecture**
   - **ARCH-DRY: pass.** `fleetRepoPaths` and `buildTestStatsReport` were split out and are reused, not copied.
   - **ARCH-PURE: pass.**
   - **ARCH-PURPOSE: pass.** The shadow-sweep finds every schedule consumer deriving from `effectiveSchedule`.
   - **ARCH-MOCK: pass.** git goes through `landingGit`, and the plan path is injectable.
   - **ARCH-CONSTRAINTS: pass.** The plan read and one `rev-parse` sit under the existing 30 s cache, and overloads are named.
   - **ARCH-SECURE: pass.** Parsing is typed, and an unreadable plan degrades visibly and is never overwritten.
   - **ARCH-ORDER: pass.** No new state is held between events. Concurrent `--write` runs share one `.tmp` name; that's acceptable for a verb only the steward runs.
   - **ARCH-FUNERAL: pass.** There is one plan file per machine and it is overwritten in place. A failed rename can leave `.tmp` behind, which is negligible.
   - **ARCH-EVOLVE: pass.** Unknown members round-trip at the top level and in each entry. The `since_head` → `since_stamp` rename only touches a field from M1 on this branch that never shipped (the base is the rebased branch state).
   - For upcoming work: when repos have different intervals, packing uses the longest one and a shorter-interval repo's offset is reduced modulo its own interval, so the no-collision spacing no longer holds. Today that only produces a warning; it should become a real rule if mixed intervals show up.
7. **Plan revisions**
   - Add a "## Revisions" entry saying offsets 0..1439 are accepted and reduced modulo the interval.
   - Change "one warning per TTL" to "one warning per schedule source".

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      effectiveSchedule is used by landingScheduleSource and runLandingStatus; reverting status to readLandingSchedule fails TestLandingStatusShowsTrainsAndReadiness ("status ignores the machine plan").
  - id: BR-14
    disposition: addressed
    note: |
      TestLandingPlanCommand has an overloaded stubPlanInputs case; deleting the planFromCosts "does not fit" loop turns it red (verified by revert).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: plan-claim-matches-delivery
    title: |
      Plan prose says out-of-range offsets are errors and plan-read failures warn once per TTL; the code reduces offsets modulo the interval and warns once per source
    detail: |
      This is the 5th finding in family plan-claim-matches-delivery. Rule: every behavioural sentence in the plan's Core concepts names a test or code site that enforces it, or the sentence goes in Revisions. Instances: plan.go:38 (MachinePlan.Offset accepts 0..1439 and takes it modulo the interval, never refuses) and landingqueue.go:242 (planWarned is set once for the source's lifetime, not per TTL). Fix by adding a Revisions entry for both, not by changing the code.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8afdfe85-1b12-4840-aed6-ba1fbfab7c97","repository":"github.com/xianxu/ariadne","issue_id":"000475","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/1dd7f04750a43b8e","reviewed_head":"61af3a94f9a6f05669bbadd80bd5ab4d9d44aebf","slot":"ariadne:4","requested_at":"2026-10-20T13:04:56.076592Z"},"phase":"finished","pid":32275,"started_at":"2026-10-20T13:04:56.244633Z","finished_at":"2026-10-20T13:09:40.377767Z","session_id":"75f34f1f-15d8-4dc4-9061-ee3bf1c3e240","outcome":"returned"}] -->
