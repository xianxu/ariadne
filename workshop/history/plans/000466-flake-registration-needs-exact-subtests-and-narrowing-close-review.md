# Boundary Review — ariadne#466 (whole-issue close)

| field | value |
|-------|-------|
| issue | 466 — flake registration needs exact subtests and narrowing |
| repo | ariadne |
| issue file | workshop/issues/000466-flake-registration-needs-exact-subtests-and-narrowing.md |
| boundary | whole-issue close |
| milestone | — |
| window | a0e8766188c855529ac82a38d66dcf30085f47b4..3e1c744a6513c1278812620c0ca5def6db84e2e6 |
| command | sdlc close --issue 466 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T13:54:19-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing the gate rerun path and log scanner next.

Checking whether the success-post guard (`Unexplained`) can see subtest registrations.

Running the targeted tests and a quick probe that the success guard ignores subtest registrations.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core of #466 is correct and well tested. `ParseName`, `Covers`, `Leaves` and `Registry.Cover` behave as the plan describes, and the classifier judges a failing test by its failing leaf subtests. The rerun runs the whole top-level test, and a sibling that fails in the rerun is reported as unknown. The Done-when test uses a real nested Go package. I ran the targeted tests (`internal/flakes`, `internal/issue`, `internal/recovery`, and `cmd/sdlc -run 'Flake|FailingSubtests|KnownFlake|Recovery|Catalog'`) and they all pass. Two things keep this from SHIP:

- **Two readers still use exact top-level lookup.** The success-post guard (`flakes.Unexplained`) and the "not registered" section of `flake list` both call `Registry.Lookup` on the top-level failures stored in records. A registered subtest therefore never explains an earlier unknown failure. On that head, the documented recovery (register the flake, then republish) keeps refusing. A small probe confirmed this: after registering `pkg.TestA/sub`, an Unknown attempt failing `pkg.TestA` is still unexplained.
- **The plan's "records unchanged" claim is wrong.** Attempt records' `known_flakes[].test` can now hold subtest names, and older binaries refuse such records. That hazard isn't in the README or help.

1. **Strengths**
   - `cmd/sdlc/internal/flakes/registry.go:72`: `ParseName` splits at the leftmost dot that gives a valid import path and a `Test` function, and rejects bad segments outright rather than trying another split. The `dev-aliases.test.sh` and `TestA/sub.TestB` cases are pinned in `TestParseName`.
   - `classify.go:19`: `explain` is one pure function used for both the attempt and the rerun. A sibling failing in the rerun becomes `Unknown` with its name, not "failed twice".
   - `localcheck_flakes_test.go`, `TestKnownFlakeGateSubtest`: a real `go test -v` package with nested subtests drives the real rerun argv (`-run ^(TestScripts)$`) and checks the posted status. This is the Done-when, tested properly.
   - `BuildRegistry` refuses overlapping names (a test and its own subtest) across two issues, and `flake register` refuses them up front. That keeps "which issue owns this failure" to one answer.
   - `requireOwnedToEdit` (`republish.go:565`) reuses the details-publish ownership rule instead of copying it.

2. **Critical findings**
   - None.

3. **Important findings**
   - **The success guard and `flake list` don't recognise subtest registrations** (`internal/flakes/evidence.go:74`, `flake_cmd.go:610`; ARCH-PURPOSE shadow-sweep).
     - `Unexplained` checks `r.Lookup(f)` against the top-level failures stored in records. So "registering later explains a failure the rule never judged" never applies to a subtest name.
     - Effect: `guardSuitePass` keeps refusing a head whose Unknown failure was a now-registered subtest, and the hint at `localcheck_republish.go:189` ("`sdlc flake register`, then republish") leads nowhere.
     - Narrowing #465's broad registration has the same effect on heads the broad name used to explain.
     - `flake list` also reports a confirmed fail→pass covered by a registered subtest as "not registered".
     - Fix sketch: pass a per-attempt subtest reader into `Unexplained` and the list (the `flakeDeps.subtests` seam already exists). Judge each failure by `Leaves` + `Registry.Cover`, failing closed when the log is unreadable. Add tests for: registered subtest explains it; sibling does not; log missing stays unexplained.
   - **The plan's persisted-formats section is wrong, and the record hazard isn't stated in user-facing docs** (ARCH-EVOLVE).
     - `known_flakes[].test` in v2 attempt records may now hold a subtest name. A pre-#466 reader rejects such a record (`teststats_store.go:172`, through `ValidTest`).
     - `guardSuitePass` refuses on any store warning, so a stale binary sharing the store blocks every suite-success post, not only the known-flake rule.
     - The issue Log mentions this, but the README, the help text and the plan's "records unchanged / no older-reader refusal" do not.
     - Fix: correct the plan, and add the record hazard to the README rollout bullet and the `register` help.
   - **No test proves concurrent additions survive `unregister`/`narrow`, or that a repeat writes nothing** (test coverage, ARCH-ORDER).
     - The `flakeWorld.editTests` fake applies edits directly, with no race.
     - `TestFlakeEditOnTrackerCards` only checks that a repeated narrow returns no error. It doesn't check for `ErrNoChange` or an unchanged commit, and it has no racing registration.
     - The catalog proof for "a repeat writes nothing" therefore isn't pinned. Add a racing-addition case (the `raceAdd` pattern) and assert no new tracker commit on a repeat.

4. **Minor findings**
   - `RerunMissing` is called with `flakes.Tests(o.Known)` only; the plan says the top-level test's PASS line is also required. Exit 0 already implies it passed, so this is plan drift, not a safety gap.
   - `MaxRerun` counts known entries, not the top-level tests actually rerun, so two subtests of one test count twice.
   - In `PersistentMain`, `topOf(u)` matches main's baseline at top-level granularity, so a sibling failure can be labelled "persistent main" when main failed a different subtest.
   - `unregister` on a card that never had `flake_tests` writes an empty `flake_tests: ` line rather than no-op.
   - When several failed attempts exist on the same head, `subtestFailedIn` checks only the one `SameHeadFailPass` picked. That fails closed, but it can refuse valid evidence.
   - The overlap check in `runFlakeRegister` restates `BuildRegistry`'s check in both directions (ARCH-DRY nit).

5. **Test coverage notes**
   - Covered well: pure name, coverage and leaf tests; classification of subtests (registered, sibling, both, parent alone, rerun outcomes); subtest parsing in all three log formats, plus a check that subtests don't reach the persisted JSON; the real-package gate e2e; register evidence through the saved-log seam.
   - Gaps:
     - `Unexplained`/`guardSuitePass` with a subtest registration.
     - The default `subtests` dep reading a real `failed/<id>.log`.
     - Edit-verb races and no-op repeats.

6. **Architectural notes**
   - Pass: ARCH-DRY (shared scanner `failingNames`, shared ownership rule), ARCH-PURE (pure decisions and classification; thin IO through injected deps), ARCH-MOCK (real tracker repo and real Go package), ARCH-CONSTRAINTS (rerun stays one bounded `go test` invocation within the existing timeout), ARCH-SECURE (names validated at every boundary; subtest segments never reach `-run`, only the top-level test does, and the segment charset is argv-safe), ARCH-FUNERAL (creates nothing durable).
   - ARCH-ORDER: passes on mechanism (the CAS decision runs on the current card, and a concurrent change refuses with "rerun"), but the race isn't tested.
   - Flagged: ARCH-PURPOSE and ARCH-EVOLVE, as above.
   - Since records keep only top-level failures, every reader that judges coverage needs the log's subtests. Put that "attempt → leaves" lookup behind one injected seam so `Classify`, `Unexplained` and the list all derive from the same source.

7. **Plan revision recommendations**
   - `## Revisions`: "Persisted formats: suite attempt records' `known_flakes[].test` may carry subtest names under v2. A pre-#466 reader refuses such records, and its success guard then refuses every post on that store. Rollout: rebuild :0 and the slots before any subtest is registered. README and help state this."
   - `## Revisions`: "The success guard (`Unexplained`) and `flake list` judge failures by subtest coverage from the saved log, not by exact top-level lookup" (or record why this is deferred, and change the guard's hint so it doesn't promise that registering a subtest helps).

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: coverage-consumers-must-derive
    title: |
      Unexplained (success guard) and flake list use exact top-level Lookup, so a registered subtest never explains a failure
    detail: |
      evidence.go:74 and flake_cmd.go:610 call Registry.Lookup on records' top-level failures. A probe confirmed that an Unknown attempt failing pkg.TestA stays unexplained once pkg.TestA/sub is registered, so guardSuitePass keeps refusing and its "register then republish" hint fails; flake list shows the case as not registered. Fix: judge each failure by Leaves plus Registry.Cover, using the attempt log's subtests through the existing flakeDeps.subtests seam and failing closed when the log is unreadable; add tests.
  - id: new
    severity: Important
    kind: defect
    family: persisted-format-change-declared
    title: |
      Attempt records' known_flakes may now hold subtest names, but the plan says records are unchanged; README and help omit the older-reader refusal
    detail: |
      teststats_store.go:172 validates known_flakes[].test with ValidTest, so a pre-466 reader refuses such a v2 record, and guardSuitePass refuses on any store warning, which blocks every suite-success post on that store. Correct the plan's Persisted formats section and add this hazard to the README rollout bullet and the register help.
  - id: new
    severity: Important
    kind: defect
    family: done-when-evidence-pinned
    title: |
      No test proves concurrent additions survive unregister/narrow or that a repeat writes nothing
    detail: |
      The flakeWorld.editTests fake applies edits with no race, and TestFlakeEditOnTrackerCards only checks that a repeated narrow returns no error. Add a racing-addition case on real tracker cards and assert ErrNoChange (no new commit) on a repeat, as the catalog proof claims.
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      RerunMissing does not also require the top-level test's PASS line, as the plan says
    detail: |
      Exit 0 already implies the parent passed, so this is plan drift only; fix the plan text or pass RerunTests to it too.
  - id: new
    severity: Minor
    kind: defect
    family: subtest-granularity-labels
    title: |
      MaxRerun counts names, not top-level tests rerun; PersistentMain matches at top-level granularity
  - id: new
    severity: Minor
    kind: defect
    family: idempotent-setter-no-write
    title: |
      unregister on a card without flake_tests writes an empty field instead of no-op
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a4eb394f-4bbe-4df5-a635-1e1749cdb8ed","repository":"github.com/xianxu/ariadne","issue_id":"000466","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/7bf7d43ca245c258","reviewed_head":"3e1c744a6513c1278812620c0ca5def6db84e2e6","slot":"ariadne:2","requested_at":"2026-10-19T20:49:43.607381Z"},"phase":"finished","pid":54877,"started_at":"2026-10-19T20:49:43.707778Z","finished_at":"2026-10-19T20:54:18.935532Z","session_id":"f3c633d4-10fb-4739-aa9b-d9edbaa122b3","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T14:19:57-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 466 — flake registration needs exact subtests and narrowing |
| repo | ariadne |
| issue file | workshop/issues/000466-flake-registration-needs-exact-subtests-and-narrowing.md |
| boundary | whole-issue close |
| milestone | — |
| window | a0e8766188c855529ac82a38d66dcf30085f47b4..e7ed662d6420d15d58b13c317224c8e42404d136 |
| command | sdlc close --issue 466 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T14:19:57-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Focus first on the latest fix commit for the prior findings.

Tests pass at head. Next I'll revert the wiring for the guard and the flake list to see whether any test catches it.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The code fix for BR-1 is correct, but nothing tests it where the bug actually was. `flakes.Explained` is now the single coverage rule. The gate (`explain`), the success guard (`Unexplained`) and `flake list` all decide through it. Attempts carry the subtests read from their saved logs (`withLogSubtests` / `storeSubtests`), and an unreadable log fails closed. The fix sat in two IO-side call sites, and nothing tests either one. I reverted both in this disposable checkout: `localcheck_republish.go:165` back to `flakes.Unexplained(flakeAttempts(ev), sha, reg)` and `flake_cmd.go:597` back to `flakes.Sort(flakeAttempts(mine))`. `go test . -run 'Flake|Guard|SuitePass|Republish|LocalCheck|KnownFlake'` stayed green (392s). The new `TestUnexplainedUsesLeafCoverage` only tests the pure function, which already worked when given subtests. The bug was that the callers never passed any. So by the claimed-fix rule, BR-1 is `not-addressed`. BR-2 (documentation only) and BR-3 (a real-tracker test) are addressed. BR-4..6 are Minor and unchanged.

1. **Strengths**
   - `cmd/sdlc/internal/flakes/evidence.go:87-94`: `Explained` passes through to `explain`. Coverage is now decided by one function rather than three copies (ARCH-DRY, ARCH-PURPOSE).
   - `cmd/sdlc/teststats.go:236-268`: `withLogSubtests` fails closed on an unreadable log, so a registered subtest never covers its parent without evidence from the log (ARCH-SECURE).
   - `cmd/sdlc/flake_cmd_test.go:421-475`: on real tracker cards, a registration that lands mid-edit makes the compare-and-swap refuse, a rerun keeps both names, and a repeated narrow or unregister leaves the `issue-tracker` tip unchanged. This settles BR-3.
   - The registry refuses overlapping names on two issues (`registry.go:162-168`), so which issue covers a failure has one answer.

2. **Critical:** none.

3. **Important**
   - **BR-1, still open** (no regression test for the consumers). There are two ways to close it:
     - **Guard test:** a `guardSuitePass` test with a stored failed attempt whose log shows `T/sub` failing and a registry naming `T/sub`. It should assert no refusal, plus a sibling case that is refused.
     - **List test:** a `runFlakeList` test with `d.subtests` set, asserting that a confirmed attempt is not listed as unregistered.
     - Either test must go red when the wiring is reverted as above.

4. **Minor**
   - **Wrong name format in the guard's hint** (`localcheck_republish.go:190`): the refusal says `--test <pkg.TestName>`, with no subtest form. This is the 2nd finding in family `subtest-granularity-labels`. The rule that covers both: every user-facing text that names a registerable test uses the one grammar `PKG.TestName[/sub…]`. A grep of `cmd/sdlc` for `pkg.TestName>` finds this as the only remaining site.
   - **BR-4..6** are unchanged. Details are in the dispositions below.

5. **Test coverage notes**
   - The pure core is well covered: `Leaves`, `Cover`, `Classify`, and `Unexplained` with subtests.
   - The IO glue that feeds subtests into the guard and the list has no test. That is exactly the bug class BR-1 named.
   - The flakes package passes; the targeted `cmd/sdlc` run passes (399s).

6. **Architecture notes** (one line each)
   - **ARCH-DRY:** pass. One coverage rule, and `storeSubtests` was pulled out of `defaultFlakeDeps`.
   - **ARCH-PURE:** pass. The decision is pure; the IO is the `subtests` function that gets injected.
   - **ARCH-PURPOSE:** pass on code; the consumer wiring still needs its test (BR-1).
   - **ARCH-MOCK:** pass. Card verbs are tested on a real tracker repo.
   - **ARCH-CONSTRAINTS:** pass. Reading the logs happens only for failed attempts on one SHA, within the store's record limit.
   - **ARCH-SECURE:** pass. Names are parsed at the boundary, and an unreadable log fails closed.
   - **ARCH-ORDER:** pass. Compare-and-swap editing is tested with an injected interleaving.
   - **ARCH-FUNERAL:** pass. Nothing new is stored; closing an issue unregisters its tests.
   - **ARCH-EVOLVE:** pass after BR-2. The older reader refusing `known_flakes` subtest names is now a documented rollout hazard.

7. **Plan revisions**
   - Fix the `RerunMissing` sentence: it requires each known entry's exact PASS line, not also the top-level test's (BR-4).
   - Note that `MaxRerun` counts registered names, not top-level tests rerun (BR-5).

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      Code fix correct, but reverting withLogSubtests at localcheck_republish.go:165 and flake_cmd.go:597 keeps all targeted cmd/sdlc tests green; add guardSuitePass and runFlakeList regression tests that fail without the wiring.
  - id: BR-2
    disposition: addressed
    note: |
      README rollout bullet, flake help (flake_cmd.go:298-299), atlas sdlc-binary.md Rollout line and plan Revisions now state the older-reader known_flakes refusal.
  - id: BR-3
    disposition: addressed
    note: |
      TestFlakeEditKeepsAConcurrentRegistration on real tracker cards asserts "changed concurrently" on the raced edit, both names after rerun, and an unchanged issue-tracker tip on repeats; passes.
  - id: BR-4
    disposition: not-addressed
    note: |
      Plan still says RerunMissing requires the top-level PASS line too; code checks entries only.
  - id: BR-5
    disposition: not-addressed
    note: |
      Unchanged; MaxRerun still counts known entries (names).
  - id: BR-6
    disposition: not-addressed
    note: |
      Unchanged; editFlakeTests still calls SetCardField(..., "") when the field is absent.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: subtest-granularity-labels
    title: |
      guardSuitePass refusal hint says --test <pkg.TestName>, omitting the subtest form
    detail: |
      This is the 2nd finding in family subtest-granularity-labels. The rule is that every user-facing text naming a registerable test uses the grammar PKG.TestName[/sub…]. A sweep of cmd/sdlc for pkg.TestName> finds only localcheck_republish.go:190.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"106b5287-e10e-47ac-9b79-5281985be73a","repository":"github.com/xianxu/ariadne","issue_id":"000466","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/15ac6d4d01c68eb4","reviewed_head":"e7ed662d6420d15d58b13c317224c8e42404d136","slot":"ariadne:2","requested_at":"2026-10-19T21:05:08.020127Z"},"phase":"finished","pid":25814,"started_at":"2026-10-19T21:05:08.200323Z","finished_at":"2026-10-19T21:19:56.812052Z","session_id":"c30f1b78-e776-44ba-90e8-2a82e05ae25b","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T14:25:40-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 466 — flake registration needs exact subtests and narrowing |
| repo | ariadne |
| issue file | workshop/issues/000466-flake-registration-needs-exact-subtests-and-narrowing.md |
| boundary | whole-issue close |
| milestone | — |
| window | a0e8766188c855529ac82a38d66dcf30085f47b4..ee4d22050707b73f6c81b9a365082237d3ac687e |
| command | sdlc close --issue 466 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T14:25:40-07:00 |
| verdict | SHIP |

## Review

Checking BR-1's fix wiring; next I'll run tests and revert the wiring to confirm the regression tests go red.

Disposing BR-4..6 (code unchanged at those sites); now sweeping user-facing name grammar for BR-7.

```verdict
verdict: SHIP
confidence: high
```

The blocking finding from earlier rounds, BR-1, is fixed and I verified it. `flakes.Explained` (`cmd/sdlc/internal/flakes/evidence.go:83`) is now the single coverage rule, and three places use it: the gate (`explain`), the success guard (`Unexplained`) and `flake list`. Each failed attempt's failing subtests are read from its saved log through `withLogSubtests` / `storeSubtests` (`cmd/sdlc/teststats.go:236`). If a log can't be read, no subtests are added, so only a registered whole test can explain that failure: it fails closed. In my checkout I put back the old wiring at both call sites (`localcheck_republish.go:165`, `flake_cmd.go:597`). Both `TestSuitePassGuardReadsSubtestsFromTheLog` and `TestFlakeListJudgesSubtestsFromTheLog` then failed, and with the fix restored they pass. The `internal/flakes` tests and the targeted `cmd/sdlc` flake, subtest and guard tests pass (56s). The only remaining exact `Registry.Lookup` call is in `Revalidate` (`classify.go:219`), and it is correct there because it checks registered entries, not failures. BR-7 is fixed. BR-4, BR-5 and BR-6 are unchanged; all three are Minor, so they don't block shipping.

**What I could not verify here:** the recorded full suite, which the Done-when requires. That evidence belongs at the hub's landing.

1. **Strengths**
   - One coverage rule (`explain`, `classify.go:72`) is shared by the gate, the guard and the list. That is the right way to close the class of bug BR-1 belonged to, not just the one site it named (ARCH-DRY, ARCH-PURPOSE).
   - The regression tests drive real `runLocalCheck` plus `republishSavedPass` with a nested-subtest log. Both the "registered subtest lands" and "only the sibling is registered, so it blocks" cases are covered (`localcheck_flakes_test.go:440`).
   - `ParseName`'s leftmost-valid-boundary split handles dotted subtests and subtests that look like test names, and each segment is checked against a character allow-list at the boundary (ARCH-SECURE).
   - `BuildRegistry` refuses overlapping names across issues, so which issue owns a failure has exactly one answer.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **BR-4 (still open):** `localcheck_flakes.go:104` checks the rerun for PASS lines of the known entries only. The plan says it also checks the top-level test. Either update the plan or pass the `RerunTests` names too.
   - **BR-5 (still open):** `classify.go:166` compares the number of covering names to `MaxRerun`, not the number of top-level tests rerun.
   - **BR-6 (still open):** `editFlakeTests` with an empty result calls `SetField`, which adds `flake_tests:` to a card that never had the field. So `unregister` on such a card writes instead of doing nothing.
   - **Log reads in the guard and list (ARCH-CONSTRAINTS):** every guard or list call re-reads and parses every failed attempt's full saved log. The guard could filter `ev` to `sha` before reading logs.

5. **Test coverage notes:** the guard and list wiring is now covered by tests that fail without the fix. There is a pure test for fail-closed on an unreadable log through `Attempt.Subtests` being empty. There is no wiring-level test where the log file is missing.

6. **Architectural notes**
   - ARCH-DRY: pass. There is one coverage rule.
   - ARCH-PURE: pass. Coverage and card decisions are pure, and log IO is behind `subtestsOf`.
   - ARCH-PURPOSE: pass. All three consumers derive from `explain`.
   - ARCH-MOCK: pass. Tests use real tracker cards and a real `go test` fixture.
   - ARCH-CONSTRAINTS: minor note above.
   - ARCH-SECURE: pass. Names are validated at the boundary, and an unreadable log fails closed.
   - ARCH-ORDER: pass. Card edits go through compare-and-swap `runCardUpdate`, and the concurrency test runs on real cards.
   - ARCH-FUNERAL: pass. Nothing new is persisted; subtests are read from existing logs.
   - ARCH-EVOLVE: pass. `known_flakes` may now hold a subtest name, an older sdlc refuses such records, and the README, help and atlas document this.

7. **Plan revisions:** add a Revisions line that either corrects the `RerunMissing` "and its top-level test" claim (BR-4) or records that it is deferred.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Guard and list both use withLogSubtests and flakes.Explained; reverting either wiring site makes TestSuitePassGuardReadsSubtestsFromTheLog / TestFlakeListJudgesSubtestsFromTheLog fail (verified locally).
  - id: BR-4
    disposition: not-addressed
    note: |
      localcheck_flakes.go:104 still passes flakes.Tests(o.Known) only; plan text unchanged.
  - id: BR-5
    disposition: not-addressed
    note: |
      classify.go:166 still compares len(known) covering names against MaxRerun.
  - id: BR-6
    disposition: not-addressed
    note: |
      editFlakeTests sets flake_tests to "" via SetField, which appends the key to a card that lacked it.
  - id: BR-7
    disposition: addressed
    note: |
      localcheck_republish.go:190 now reads PKG.TestName[/sub…]; flake help, register Use and flags agree; README and atlas follow the synopsis with explicit subtest text.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ccaf6a69-91b3-4f4e-a649-972298649eab","repository":"github.com/xianxu/ariadne","issue_id":"000466","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/d7abc595aeb1da9b","reviewed_head":"ee4d22050707b73f6c81b9a365082237d3ac687e","slot":"ariadne:2","requested_at":"2026-10-19T21:23:12.296918Z"},"phase":"finished","pid":93534,"started_at":"2026-10-19T21:23:12.584286Z","finished_at":"2026-10-19T21:25:39.345602Z","session_id":"87c09e13-6272-4816-9059-f31006c9a08b","outcome":"returned"}] -->

---

## Re-review — 2026-10-19T17:53:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 466 — flake registration needs exact subtests and narrowing |
| repo | ariadne |
| issue file | workshop/issues/000466-flake-registration-needs-exact-subtests-and-narrowing.md |
| boundary | whole-issue close |
| milestone | — |
| window | 4e640fe30cf6b3746378a800fc5eabdc904d28a2..c0bc9b48b215b10f936e147fb142d2421aaeab3b |
| command | sdlc close --issue 466 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-19T17:53:25-07:00 |
| verdict | SHIP |

## Review

Reviewing the conflict resolution in `flake_cmd.go`; I'll build and run the flake tests next.

```verdict
verdict: SHIP
confidence: high
```

This window contains nothing but the merge-conflict resolution. At base `4e640fe`, `atlas/workflow/sdlc-binary.md` and `cmd/sdlc/flake_cmd.go` still had conflict markers committed in them, and head `c0bc9b4` resolves both. Each resolution keeps the intent of both sides. In `flake_cmd.go`, #458's `flake list -v` flag, fixed-issue filtering and recurrence labelling sit together with #466's `unregister`/`narrow` verbs, the subtest help text and the shared `flakes.Explained` coverage rule. The atlas row keeps #458's fixed-filter sentence and folds #466's subtest bullets into one row-safe paragraph without dropping any content. No conflict markers are left in `atlas/`, `cmd/` or `README.md`. `go build ./...` and `go vet ./cmd/sdlc/` are clean, and the `Flake|flake` tests in `cmd/sdlc` and all of `cmd/sdlc/internal/flakes` pass. Nothing in this window touches the three open Minor findings (BR-4/5/6), so they stay open. None of them blocks.

1. **Strengths**
   - `cmd/sdlc/flake_cmd.go:626`: the resolved `flake list` sorts `withLogSubtests(...)` attempts, so #458's fixed-issue loop now runs over subtest-aware attempts and the `-v` path sees the same order.
   - `cmd/sdlc/flake_cmd.go:658`: the unregistered check uses `flakes.Explained` (#466) and then hands uncovered rows to #458's `classifyFixedFlake`. Both behaviours survive, and nothing is hidden when evidence is ambiguous.
   - The help text (`flake_cmd.go:298-309`) lists all four verbs, including `list [-v]`. The `-v` flag is registered next to the new verbs (`:391-392`).
   - The atlas row is one table cell again. A multi-line bullet list would have broken the markdown table.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/flake_fixed.go:104` (reached from `flake_cmd.go:661`): #458's fixed-issue matcher `fixedIssueCovers` compares a terminal card's `flake_tests` entries to the top-level failure name by exact equality. Since the merge, those entries can be subtest names, so a landed issue that fixed `PKG.TestX/sub` never hides, or labels as "recurred after", a later `PKG.TestX` row. This fails safe: the row stays visible as unregistered.

5. **Test coverage:** existing `cmd/sdlc` flake tests and all `internal/flakes` tests pass at head. No test combines the #458 fixed-issue filter with a subtest-registered terminal card; that is the gap behind the Minor above.

6. **Architectural notes**
   - **ARCH-DRY:** pass. The resolution adds no duplicate logic.
   - **ARCH-PURE:** pass. `classifyFixedFlake` still takes ancestry through injection.
   - **ARCH-PURPOSE:** flagged Minor. The fixed-issue matcher is one coverage consumer that does not use the shared `Covers`/`Explained` rule (see the Minor above).
   - **ARCH-EVOLVE:** pass. No persisted format changes in this window.

7. **Plan revision recommendations:** none from this window. BR-4's plan-text drift is still open from earlier rounds.

```findings
dispose:
  - id: BR-4
    disposition: not-addressed
    note: |
      Interdiff is only the conflict resolution; plan text and RerunMissing call are unchanged.
  - id: BR-5
    disposition: not-addressed
    note: |
      Unchanged in this window; MaxRerun still counts covering names.
  - id: BR-6
    disposition: not-addressed
    note: |
      Unchanged in this window; editFlakeTests still writes an empty flake_tests field.
findings:
  - id: new
    severity: Minor
    kind: defect
    family: coverage-consumers-must-derive
    title: |
      Merged fixed-issue filter (fixedIssueCovers) matches top-level names exactly, not via the shared subtest coverage rule
    detail: |
      This is the 2nd finding in family coverage-consumers-must-derive. Rule: every consumer that asks "does a card's flake_tests cover this failure" must call flakes.Covers/Explained (leaf or ancestor) rather than string equality; sweep all such matchers, of which flake_fixed.go:104 is the one left after the #458/#466 merge. It fails safe (a subtest-registered fixed issue never hides or labels a row as recurred), so this is Minor.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c835c65b-ba31-4a83-83f3-09912ab40028","repository":"github.com/xianxu/ariadne","issue_id":"000466","recipe":"fix-review","round":4,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/41feaf1b2beb96bc","reviewed_head":"c0bc9b48b215b10f936e147fb142d2421aaeab3b","slot":"ariadne:2","requested_at":"2026-10-20T00:52:14.681863Z"},"phase":"finished","pid":81535,"started_at":"2026-10-20T00:52:14.783161Z","finished_at":"2026-10-20T00:53:25.25073Z","session_id":"af0abda6-fe5a-43a4-857d-8dc434d645c2","outcome":"returned"}] -->
