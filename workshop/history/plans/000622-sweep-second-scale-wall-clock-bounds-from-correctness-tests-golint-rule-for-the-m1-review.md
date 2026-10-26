# Boundary Review — ariadne#622 (milestone M1)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..a101e5c2c5c2d24e605f999d699e13a276cf38d7 |
| command | sdlc milestone-close --issue 622 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T05:11:50-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm checking how other repos consume the lint catalog, then I'll run the tests.

Now probing the rule with adversarial inputs.

Confirmed Pair consumes the catalog. Running the #625 test once to confirm it passes at head.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 delivers what it set out to. The rule is widened as planned: any elapsed-time bound below `testgate.HangCeiling` is flagged, and so is any `time.Now().Add(d)` deadline that a loop later polls. Variables are keyed by their declaring object. All 13 group A sites in the plan were moved to `HangCeiling`, plus the three extra deadlines the object-keyed rule found. The #625 barrier now fails fast, and the re-read on an empty file is correct. The atlas was updated.

What I verified myself:
- `go test ./pkg/golint/` is green, including the repo-wide catalog test.
- `TestLandingTrainSplitDoesNotBlameUntouchedPackage` passes `-count=3` at head (155 s).
- I fed the rule a set of tricky inputs to see what it catches and misses.

Two Important gaps block a clean SHIP, both cheap:
1. **Fail-fast only at #625.** The plan's group A fix says to fail fast wherever a worker can exit before the event. Only #625 got that. Six other sites have a worker result channel but don't check it, so a test whose worker dies early now waits 2 minutes before failing, instead of 4–60 seconds.
2. **Pair is not considered.** Pair pins and runs `pkg/golint`'s catalog. The widened rule will flag Pair's tests on its next pin bump, and the ariadne-only pending list doesn't help Pair. The plan says nothing about this.

**1. Strengths**
- `pkg/golint/rules.go:587` `varKey`: keying by declaring object is right. `TestTightWallclockScopesNamesPerFunction` (`rules_test.go:430`) fails if variables are keyed by name instead: `bounded()`'s `elapsed` would be flagged, and `long()`'s assignment would hide `short()`'s.
- Only flagging deadlines that are actually polled is a precise way to separate them from timestamp fixtures (`rules_test.go:192`). The reassignment case at `rules_test.go:193` covers the `delete` branch.
- `hangBound` takes the ceiling from `testgate`, so there is one source for it (ARCH-DRY). `testfix.HangCeiling` is an alias of the same constant (`testfix/wait.go:13`).
- `landingtrain_attribution_test.go:148-175`: the #625 loop treats the file barrier as the event, checks both landing channels, and the comment explains why an empty read is not the event yet.
- `wallclockPending` names its end (removed in M2), as ARCH-FUNERAL asks.

**2. Critical findings**
None.

**3. Important findings**
- **Fail-fast missing at worker-backed group A sites** (ARCH-PURPOSE, ARCH-ORDER). None of these loops checks its worker's result channel:
  - `publishgate_cancel_test.go:37` (`done`)
  - `makebuild_publish_test.go:100` (`done`)
  - `queue_interruption_test.go:123` (`done`; the same file's loop at :48 already does this correctly)
  - `landingintegrate_test.go:158` and `merge_main_test.go:78` (the `worker` from `startBlockedMergeWorker`)
  - `landingtrain_test.go:611` (`one`, `two`)

  Fix sketch: add `select { case err := <-done: t.Fatalf(...); default: }` inside each loop, as already done at `queue_interruption_test.go:57` and `landingtrain_archive_test.go:172`.
- **Pair is not considered in the plan** (ARCH-PURPOSE; ariadne is the base layer). History shows Pair pins `ariadne/pkg/golint` and runs the catalog: #590 shipped a matcher change that Pair then reported as false positives, and #549 inspected Pair's pinned copy. The M1 rule enforces on every path not in an ariadne-specific pending list, so Pair's second-scale deadlines will go red when it bumps the pin. Fix sketch: add a plan Revision that states the downstream impact and a migration note, for example a sweep issue filed against Pair or a changelog line in the atlas entry.

**4. Minor findings**
- `hangBound` (`rules.go:633`) accepts any expression that mentions `HangCeiling`. So `HangCeiling/1200` and `HangCeiling-119*time.Second` pass, and the implementor's own mutation (`HangCeiling/1200`) would pass lint. No live instance exists today.
- The rule misses the inverted polling form `deadline.Before(time.Now())` / `.After(time.Now())`. No live instance exists today.
- There is no rule test for the `time.Until(v)` polling form, or for the `wallclockPending` skip itself.
- `wallclockPending[context.Path]` doesn't apply `path.Clean`, unlike the sibling rules (`rules.go:121,152`).
- `fnIdent` repeats the package-identifier check inside `isSince` (ARCH-DRY, small). `isSince` could use it.
- The new `testfix`/`testgate` imports sit inside the stdlib import group in about 9 files. gofmt is happy, but goimports grouping is inconsistent.
- `runner_setup_test.go:131` still calls `staging.AwaitExclusive(stage, staging.Settle)`, a 1 s product-contract wait the rule can't see. It's worth a look in M2's group C pass.

**5. Test coverage notes**
- The rule tests cover the shapes the plan seeds: #620's BR-16 shape red, a non-constant bound red, `After`/`Before` deadlines red, the fixture green, the `HangCeiling`/2 min bounds green, the reassignment case, and the per-function scoping case.
- The #625 fix is only covered by stress runs, which is acceptable for an ordering fix in a test. The logged mutation (100 ms deadline → red) is consistent with the code.
- I did not repeat the per-package `-count=10` runs. I ran the #625 test `-count=3`, which was green.

**6. Architectural notes**
- ARCH-DRY: pass, with the minor `fnIdent` duplication above.
- ARCH-PURE: pass. The rule is AST-only and tested without IO.
- ARCH-PURPOSE: flagged (the fail-fast gap and Pair).
- ARCH-MOCK: pass. No seams changed.
- ARCH-CONSTRAINTS: pass. `HangCeiling` is the only guard. The cost is slower failure, which ties back to the fail-fast finding.
- ARCH-SECURE: N/A. Only test code and the lint rule change.
- ARCH-ORDER: flagged (the fail-fast gap).
- ARCH-FUNERAL: pass. The pending list is deleted in M2.
- ARCH-EVOLVE: N/A. No persisted format changes.
- For #626 and M2: tighten `hangBound` before #626 reuses it.

**7. Plan revision recommendations**
- Add a Revision noting that the processgroup group C allow landed in M1, since processgroup is not on the pending list.
- Add a Revision stating the downstream impact on Pair and how it will be handled.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: polled-wait-fail-fast
    title: |
      Group A poll loops with a worker channel do not fail fast when the worker exits before the barrier
    detail: |
      The plan's group A fix and ARCH-ORDER say to fail fast where a worker can exit before the event, but only the #625 site does. These loops ignore their worker channels: publishgate_cancel_test.go:37, makebuild_publish_test.go:100, queue_interruption_test.go:123, landingintegrate_test.go:158, merge_main_test.go:78 and landingtrain_test.go:611. A premature worker exit now costs the full 2 min HangCeiling. Add a non-blocking select on the worker channel inside each loop, as queue_interruption_test.go:57 already does.
  - id: new
    severity: Important
    kind: defect
    family: base-layer-downstream-impact
    title: |
      Widened test-tight-wallclock rule ships with no plan for Pair, which runs the catalog
    detail: |
      Pair pins ariadne pkg/golint and runs its catalog (see #549 and #590 history). The M1 rule enforces on every path outside an ariadne-only pending list, so Pair's second-scale polling deadlines go red when it bumps the pin. Add a plan Revision stating the downstream impact and how it is handled, for example a Pair sweep issue or a migration note.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-loophole
    title: |
      hangBound accepts any expression that mentions HangCeiling, so HangCeiling/1200 passes
    detail: |
      rules.go:633 accepts HangCeiling divided by or minus anything. Accept only a bare HangCeiling identifier or selector, or HangCeiling plus a non-negative term. The inverted polling form deadline.Before(time.Now()) is also not matched.
  - id: new
    severity: Minor
    kind: defect
    family: rule-test-coverage
    title: |
      No rule test for the time.Until polling form or for the wallclockPending skip
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7e66c2fc-a4b6-456c-ab56-aed2f5a7f6b8","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/9cc7eb3194154c6a","reviewed_head":"a101e5c2c5c2d24e605f999d699e13a276cf38d7","slot":"ariadne:2","requested_at":"2026-10-25T12:06:32.83958Z"},"phase":"finished","pid":10596,"started_at":"2026-10-25T12:06:32.967354Z","finished_at":"2026-10-25T12:11:47.406777Z","session_id":"66b47e8d-c34f-4d53-96ac-be02c37e1add","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T06:56:54-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..a4a63ccffc59193d542ebfbbf85410311fcc88f3 |
| command | sdlc milestone-close --issue 622 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T06:56:54-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The golint tests pass and the changed packages vet clean. Next I'm checking the rule against adversarial inputs and running the changed tests.

Prior findings look addressed. Now running a sample of the changed cmd/sdlc tests once to confirm they compile and pass.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All four prior findings are fixed, and I checked each one against the code at the pinned head. The six inline loops named in BR-1 now stop as soon as their worker exits. The plan has a Revision covering pair (pair#885). `hangBound` now accepts only `HangCeiling` itself, `HangCeiling` plus a non-negative constant, or `HangCeiling` times a positive integer. The inverted poll form and `time.Until` are matched, and both have tests. I broke those fixes on purpose (allowing `/` and `-` in `hangBound`, disabling the inverted-poll match): `TestRules` failed on exactly the four new cases, and passed again once I restored the file. `go test ./pkg/golint/` passes, including the repo catalog test. Vet is clean on the changed packages, and a sample of seven changed `cmd/sdlc` tests passed in 58 s.

One thing still needs fixing before shipping. `waitQueued` and `waitVouched` now take a variadic list of workers to watch. Several callers leave out a worker that is certainly still running, and nothing forces them to pass it. This is the second finding in the `polled-wait-fail-fast` family.

**1. Strengths**
- `pkg/golint/rules.go:541-590`: the poll check makes a single pass in source order and judges each poll by the assignment it reads at that point. A later reassignment can't hide an earlier poll, and there are tests for both orders (`rules_test.go:193,195`).
- Variables are keyed by their declaring object (`varKey`). `TestTightWallclockScopesNamesPerFunction` checks that a name reused in another function isn't confused. The Log says this keying also caught three deadlines that keying by name had hidden.
- `hangBound` rejects local look-alikes: a local `HangCeiling` variable (`rules.go:648`) and a struct field (`rules.go:651`), each with a test (`rules_test.go:193-194`). The ceiling has one source: `testfix.HangCeiling = testgate.HangCeiling` (`testfix/wait.go:13`).
- The #625 barrier treats an empty read as "not yet" (`landingtrain_attribution_test.go:155-160`). That fixes a race the implementer found during their own proof runs.
- `failIfWorkerEnded` (`landingqueue_test.go:72`) is a small shared helper. Its comment explains why a worker that a deferred cleanup reads must be left out.

**2. Critical findings**
None.

**3. Important findings**
- **This is the 2nd finding in family `polled-wait-fail-fast`.** `waitVouched` is called without workers that are certainly still running:
  - `landingtrain_test.go:548`
  - `landingtrain_test.go:662`
  - `landingtrain_test.go:953` (both `one` and `two`)
  - `landingtrain_provenance_test.go:37`
  - `landingchecks_train_test.go:59,63`

  In each case the same worker was just passed to the preceding `waitQueued`, and it stays blocked by `s.blocker` until after the vouch wait. If it dies early, the test hangs for the full 2-minute `HangCeiling`. The root cause is the variadic, opt-in signature: leaving the worker out is the default and costs nothing at compile time. **Rule:** every polled wait helper takes the set of workers it awaits as a required argument. Make `workers []<-chan error` a required parameter, with an explicit `nil` and a comment where a cleanup owns the channel (as at `landingdrop_*`, `landingtrain_held_test.go:157`, `landingeconomy_test.go:167`). Then fix every caller in one pass rather than patching these six sites.

**4. Minor findings**
- **This is the 2nd finding in family `lint-rule-loophole`.** The deadline matcher recognises only one way of writing a deadline: a single `x := time.Now().Add(d)` assignment. My probes found five shapes it misses:
  - `var deadline = time.Now().Add(30*time.Second)`
  - a multi-value assignment such as `ok, deadline := …`
  - `time.Now().Compare(deadline)`
  - `time.Since(deadline) < 0`
  - a local `const HangCeiling = time.Second`, which counts as a hang bound

  None of these appear anywhere in the tree today. **Rule:** match any binding of `time.Now().Add(d)` (every element of an N:N `AssignStmt`, plus `ValueSpec`). Treat every comparison method on the bound value as a poll. Accept `HangCeiling` only as an imported package's constant, not a constant declared in the same file.
- **ARCH-DRY:** the poll loop "until a file exists or a worker ends, deadline `HangCeiling`" is copied by hand into about six files: merge_main, landingintegrate, makebuild_publish ×2, publishgate_cancel, queue_interruption, plus landingtrain:611. A `testfix` poll helper that takes a condition and workers would replace them all, and makes the required-workers rule above enforceable.
- The new `testfix` and `testgate` imports sit inside the standard-library import group. goimports would move them into their own group.

**5. Test coverage notes**
- The rule's tests now cover every adversarial input listed in the plan, plus `/`, `-`, `+` and `*` variants, `time.Until`, the inverted poll, and the pending-list skip. My mutations showed each new case is pinned.
- The new early-exit branches in the swept tests are diagnostics in test code, and no test exercises them. That's acceptable: whether they're correct shows up in review, and the plan doesn't claim a test for them.

**6. Architectural notes**
- ARCH-DRY: flag. The duplicated poll loop is listed under Minor findings.
- ARCH-PURE: pass. The rule is AST-only, and its tests do no IO.
- ARCH-PURPOSE: mostly pass. All 13 group-A sites plus #625 were swept. The partly-applied fail-fast on the wait helpers is the Important finding.
- ARCH-MOCK: pass. Only the oracles changed; no seams did.
- ARCH-CONSTRAINTS: pass. Waits are bounded by `HangCeiling` alone, and the Log records niced `-count=10` proof runs under load.
- ARCH-SECURE: N/A. Only test code and a lint rule change; nothing reads untrusted input or touches secrets.
- ARCH-ORDER: flag. The wait helpers can't see some live workers; see the Important finding.
- ARCH-FUNERAL: pass. `wallclockPending` and its test are named for deletion in M2.
- ARCH-EVOLVE: N/A. No persisted format changes.

For M2: deleting `wallclockPending` must also delete `TestTightWallclockSkipsOnlyPendingFiles`. Keep #626 (the `select`/`time.After` fail-timeouts) reusing `hangBound`.

**7. Plan revision recommendations**
- Add an "M1 boundary review round 2" entry to `## Revisions` stating that wait helpers take their workers as a required argument (an explicit `nil` only where a cleanup owns the channel).

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      All six named loops now select on their worker; verified in the diff (publishgate_cancel:42, makebuild_publish:108, queue_interruption:131, landingintegrate:166, merge_main:86, landingtrain:621).
  - id: BR-2
    disposition: addressed
    note: |
      Plan Revisions 'M1 boundary review round 1' names pair's catalog consumer, the 183-site impact, pair#885, and the one-line exclusion on pin bump.
  - id: BR-3
    disposition: addressed
    note: |
      hangBound only allows ADD with a non-negative constant or MUL by an int of 1 or more; the inverted poll is matched. Reverting either turns rules_test HangCeiling/1200, HangCeiling-time.Minute and the !deadline.Before case red.
  - id: BR-4
    disposition: addressed
    note: |
      rules_test covers time.Until; TestTightWallclockSkipsOnlyPendingFiles covers the pending skip; both pass.
findings:
  - id: new
    severity: Important
    kind: defect
    family: polled-wait-fail-fast
    title: |
      waitVouched callers omit workers known to be live; the variadic opt-in signature lets callers leave a worker out unnoticed
    detail: |
      landingtrain_test.go:548, 662 and 953 (one and two), landingtrain_provenance_test.go:37 and landingchecks_train_test.go:59,63 call waitVouched without the conductor they just passed to waitQueued. That conductor stays blocked by s.blocker, so an early exit costs the full HangCeiling. Second finding in this family. Rule: every polled wait helper takes its awaited workers as a required argument, with an explicit nil plus a comment where a cleanup owns the channel. Fix all callers in one pass.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-loophole
    title: |
      The deadline matcher only sees single x := time.Now().Add(d) assignments; var specs, multi-assign, Compare/Since polls and a local const HangCeiling slip through
    detail: |
      Probed: var deadline = time.Now().Add(30s), ok, deadline := true, time.Now().Add(30s), time.Now().Compare(deadline), time.Since(deadline) < 0 and a local const HangCeiling = time.Second are all unflagged. In-tree prevalence is 0. Rule: match every binding of time.Now().Add (N:N AssignStmt and ValueSpec), treat every comparison on the bound value as a poll, and accept HangCeiling only as an imported package's constant.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-poll-loop
    title: |
      ARCH-DRY: the poll-until-file-or-worker-exit loop is copied by hand into about six test files
    detail: |
      merge_main_test.go:79, landingintegrate_test.go:159, makebuild_publish_test.go:101,157, publishgate_cancel_test.go:37, queue_interruption_test.go:124 and landingtrain_test.go:613 repeat the same loop (deadline HangCeiling, stat, select on the worker, sleep). A testfix poll helper taking a condition and workers would replace them and make the required-workers rule enforceable.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"dfd05cb7-c31f-4c25-90f6-d0b1b3ef7cb2","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/724e88486db36789","reviewed_head":"a4a63ccffc59193d542ebfbbf85410311fcc88f3","slot":"ariadne:2","requested_at":"2026-10-25T13:53:32.228577Z"},"phase":"finished","pid":4311,"started_at":"2026-10-25T13:53:32.319389Z","finished_at":"2026-10-25T13:56:52.70314Z","session_id":"7ad61d54-3d06-4ff7-b5f9-f3ef8c7b56bf","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T08:24:44-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..dd5a634bd1240f19ecbfce6ae87a5b11e84a762d |
| command | sdlc milestone-close --issue 622 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T08:24:44-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the call sites that pass only one worker. The provenance one was named in BR-5.

Wait helper callers look right. Now reviewing the lint rule changes.

BR-6's named shapes are covered by tests. Next I'm probing nearby shapes and running the suites.

Looking at whether weave's `await` helper (same poll class, outside `cmd/sdlc`) was touched in this window.

Checking that the AwaitCondition test fails when the put-back line is removed.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All three open findings are fixed in this round. The rule-level fix answers the `polled-wait-fail-fast` family at its root. `testfix.AwaitCondition` takes the workers as a required, non-variadic `[]chan error`, so the compiler stops a caller from silently leaving one out. All 54 `waitQueued`/`waitVouched` call sites pass the workers they hold; the five sites that pass one worker each start only one conductor, and the other cars come from `queuedCar`. The hand-copied poll loops BR-7 named are now built on the helper. The lint matcher now covers every deadline binding shape BR-6 listed, and each shape has a rule test. Nothing blocks the boundary. One Important docs-gate gap remains and is cheap to fix: atlas does not document `AwaitCondition`. Two Minor findings belong to families already in play.

1. **Strengths**
   - `cmd/sdlc/internal/testfix/wait.go:58`: the required `workers []chan error` argument makes "forgot a worker" a compile-time error rather than a review finding. Putting the result back on the channel (`w <- err`) keeps deferred `finished(t, ch)` joins working. I removed the put-back line and `TestAwaitConditionFailsFastAndKeepsTheResult` failed ("the worker's result was consumed"); with the line restored, it passes.
   - `pkg/golint/rules.go`: the deadline matcher makes one pass in source order, and `varKey` keys each variable by its declaring object. That closes the reassignment and cross-function confusion cases, and each has a test.
   - `hangBound` is strict and tested both ways: `HangCeiling/1200` and `-time.Minute` are flagged; `+time.Second` and `2*` pass. A local `const`/`:=` `HangCeiling` or a struct field named `HangCeiling` is rejected.
   - Every worker channel the helper receives is buffered with room for one value (`make(chan error, 1)`), as its doc comment requires, and none of them is closed.
   - `go test ./pkg/golint/` passes. `go vet ./cmd/sdlc/ ./cmd/weave/... ./cmd/tl-lead/` is clean.

2. **Critical findings:** none.

3. **Important findings**
   - **The atlas update for `testfix.AwaitCondition` appears to be missing.** The test-helper section at `atlas/workflow/sdlc-binary.md:2622-2640` lists `Watch`, `AwaitExit`, `DeadlineOnCue` and `HangCeiling`, but not the new helper or its required-workers rule. That rule is the convention this round introduced. Fix: add a sentence next to `testfix.Watch` describing `AwaitCondition(t, what, cond, workers)`: the workers argument is required (nil only with a comment), each channel must be buffered, and the result is put back on the channel.

4. **Minor findings**
   - **`lint-rule-loophole`, 3rd finding.** BR-6's "every binding" rule was applied to deadlines only. The elapsed matcher (`rules.go:495`) still handles only a 1:1 `AssignStmt`. Probes that slip through: `var elapsed = time.Since(start)`, `ok, elapsed := true, time.Since(start)`, `deadline.Sub(time.Now()) > 0`, and `time.Now().UnixNano() < deadline.UnixNano()`. In-tree prevalence: 0. The rule that covers all of these: one shared binding collector over N:N `AssignStmt` and `ValueSpec`, used by both the elapsed and the deadline matchers, and any expression that combines a deadline variable with `time.Now()` counts as a poll.
   - **`duplicated-poll-loop`, 2nd finding.** `AwaitCondition` lives in `cmd/sdlc/internal/testfix`, so `cmd/weave` and `cmd/tl-lead` cannot import it. weave keeps its own poll helpers with no workers argument:
     - `release/main_test.go:36` `await`: its callers at `:127` and `:160` wait on a "started" file while a child process runs, so an early exit costs the full ceiling.
     - `weavefs/runner_setup_test.go:143`.
     
     About 18 hand-rolled `time.Now().Add(…HangCeiling)` loops remain repo-wide. The rule: every poll-until-condition loop in tests goes through one helper with required workers, placed in `pkg/testgate` next to the ceiling it uses. The atlas already notes that `testgate` sits in `pkg` "so cmd/weave shares it".
   - `cmd/tl-lead/doc_operating_test.go:7`: the `pkg/testgate` import sits inside the stdlib import group.
   - The plan's M1 checkboxes are still unticked.

5. **Test coverage notes:** the helper has a fail-fast test and a test that the result survives, and the put-back half is mutation-checked. The rule tests cover every BR-6 shape. One environment note: the full `testfix` package run fails in `realgit_test.go`, because the review sandbox refuses `git init` outside its directories. That is unrelated to this diff; the targeted `AwaitCondition`/`Watch` tests pass.

6. **Architectural notes**
   - ARCH-DRY: flag (Minor, above). The helper's reach stops at `cmd/sdlc`.
   - ARCH-PURE: pass. The rule is pure AST code; `AwaitCondition` is tested through a `fatalTB` stand-in, with no IO.
   - ARCH-PURPOSE: pass for M1's scope. The family fix is at rule level, and the residual lint shapes are noted above.
   - ARCH-MOCK: pass. No seams changed.
   - ARCH-CONSTRAINTS: pass. Waits are bounded only by `HangCeiling` and poll every 5 ms.
   - ARCH-SECURE: pass (N/A). Only test code and a lint rule change.
   - ARCH-ORDER: pass. Workers that end early fail the wait at once, and putting the result back keeps the join order intact.
   - ARCH-FUNERAL: pass. `wallclockPending` is marked for deletion in M2, and its skip has a test.
   - ARCH-EVOLVE: pass (N/A). No persisted formats change.

7. **Plan revision recommendations:** add a Revisions note that `AwaitCondition` is not reachable from `cmd/weave` or `cmd/tl-lead`, and either move it to `pkg/testgate` or scope it explicitly to `cmd/sdlc`.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      waitQueued/waitVouched take a required []chan error (landingqueue_test.go:55, landingtrain_test.go:520); all 54 callers pass their live workers, and AwaitCondition fail-fast plus put-back is mutation-checked red.
  - id: BR-6
    disposition: addressed
    note: |
      rules_test.go adds var, N:N, Compare, time.Since and local-const HangCeiling cases; the residual elapsed-binding shapes are raised as a separate Minor finding.
  - id: BR-7
    disposition: addressed
    note: |
      All named loops (merge_main, landingintegrate, makebuild_publish x2, publishgate_cancel, queue_interruption, landingtrain) now call testfix.AwaitCondition.
findings:
  - id: new
    severity: Important
    kind: defect
    family: atlas-surface-coverage
    title: |
      atlas update appears missing for testfix.AwaitCondition and its required-workers convention
    detail: |
      atlas/workflow/sdlc-binary.md:2622-2640 documents testfix.Watch, AwaitExit, DeadlineOnCue and HangCeiling but not AwaitCondition(t, what, cond, workers), its required workers argument (nil only with a comment), the one-slot buffer it requires, or the result put-back. Add one sentence beside testfix.Watch.
  - id: new
    severity: Minor
    kind: defect
    family: lint-rule-loophole
    title: |
      The elapsed matcher still sees only 1:1 x := time.Since(s); deadline.Sub(time.Now()) and UnixNano comparisons are not polls
    detail: |
      3rd finding in this family. Probed unflagged: var elapsed = time.Since(start); ok, elapsed := true, time.Since(start); deadline.Sub(time.Now()) > 0; time.Now().UnixNano() < deadline.UnixNano(). In-tree prevalence 0. Rule: one binding collector (N:N AssignStmt plus ValueSpec) shared by the elapsed and deadline matchers, and any expression combining a deadline variable with time.Now() counts as a poll.
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-poll-loop
    title: |
      AwaitCondition sits in cmd/sdlc/internal, so weave and tl-lead keep their own worker-less poll helpers
    detail: |
      2nd finding in this family. cmd/weave/internal/release/main_test.go:36 await (callers :127 and :160 poll while a child runs) and weavefs/runner_setup_test.go:143 re-implement the loop with no workers argument; about 18 hand-rolled HangCeiling deadline loops remain repo-wide. Rule: every test poll loop goes through one required-workers helper placed in pkg/testgate, next to HangCeiling, so every package can import it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9aed4ea8-bbae-4524-b373-4638c9e77e8c","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/cd46946621497a2c","reviewed_head":"dd5a634bd1240f19ecbfce6ae87a5b11e84a762d","slot":"ariadne:2","requested_at":"2026-10-25T15:21:51.373692Z"},"phase":"finished","pid":11841,"started_at":"2026-10-25T15:21:51.478516Z","finished_at":"2026-10-25T15:24:41.7956Z","session_id":"b6f6ccbe-1b69-4a7e-962a-ee2d7f339566","outcome":"returned"}] -->

---

## Re-review — 2026-10-25T08:45:10-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 622 — Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape |
| repo | ariadne |
| issue file | workshop/issues/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 2bc11af146651898f1506721f33fcc41f250f3b4..6fc8c03fcc7836a127bc1e15c23847b90f199e38 |
| command | sdlc milestone-close --issue 622 --milestone M1 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-25T08:45:10-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the open findings now: atlas and lint-rule tests look covered on read, so next I'm running the golint/testgate/weave tests and looking at how `AwaitCondition` callers set up their channels.

The broad `cmd/sdlc` test run went past the 10-minute foreground limit. I'm stopping it and running one narrow test instead.

The run hit Go's 10-minute package timeout; the only test frame left in the truncated output is `unclaim_test.go:87`, which this window doesn't touch. Running the #625 test on its own:

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

All three open findings are fixed at the pinned head, and nothing new blocks M1. The atlas now documents `AwaitCondition` (BR-8). One shared binding collector feeds both lint matchers, and every probe BR-9 named is now a red rule test (BR-10's fix is covered below). `pkg/golint` passes, including the repo-wide catalog test (so the `wallclockPending` list works). So do `pkg/testgate` and the weave `release` and `weavefs` packages, and the #625 test passes 3 of 3 runs on its own. I raise one new Minor: `AwaitCondition` documents but never checks that each worker channel has a one-value buffer and stays open. A caller that breaks that rule gets a hang or a panic instead of the fast failure the helper exists to give. This is the 3rd finding in family `polled-wait-fail-fast`, so the fix below states a rule rather than patching one site. It does not block the boundary.

One run did not finish: a broad `cmd/sdlc` subset (`-run 'Attribution|Queue|…' -count=2`) hit Go's 10-minute package timeout. The truncated dump shows only `unclaim_test.go:87`, which is outside this window, so I can't tie the timeout to a changed test. I did not repeat the plan's per-package `-count=10` proofs.

1. **Strengths**
   - `pkg/testgate/await.go:19` puts one helper with a required `workers` argument in a package every module can import. It takes a small `T` interface, so `testgate` needs no `testing` import. The put-back `w <- err` keeps a deferred join working, and `await_test.go:26` checks that it does.
   - `pkg/golint/rules.go` `bindings()` is one collector for `:=`, N:N `=` and `var`, and both matchers read through it. The source-order walk judges each poll by the assignment it reads at that point. A comparison that combines any variable with `time.Now()` counts as a poll, which closes the `UnixNano` / `Sub` shapes as a class, not one at a time.
   - `hangBound` refuses a `HangCeiling` this file declares itself (`v.Obj == nil`), and refuses `/` and `-` variants of the ceiling. Rule tests cover all of these.
   - The weave `release` fixture's `waitFor` routes every `cmd.Wait` through one channel, so the code never waits on a process twice. The `nil` worker sites carry their reason in a comment (`release/main_test.go:191`, `weavefs/runner_setup_test.go:145`).

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `pkg/testgate/await.go:24-26`: the one-slot-buffer precondition is never checked.
     - With an unbuffered worker channel, the `select` takes the worker's value and the put-back `w <- err` then blocks forever, so the test hangs until `go test` times out.
     - With a closed channel, the receive gets `nil` and the put-back panics with "send on closed channel".
     - Today's callers are fine: `mergeAsync`, `waitFor` and the train helpers all use `make(chan error, 1)`.
     - Fix: check every worker at entry (`cap(w) >= 1`, else `t.Fatalf`), and receive with `err, ok := <-w` so a closed channel fails cleanly instead of panicking.
   - `cmd/tl-lead/doc_operating_test.go:7`: the `testgate` import sits inside the standard-library group. gofmt accepts it, but goimports style would put it in its own group.

5. **Test coverage notes**
   - The BR-9 probes are all pinned in `rules_test.go`: `var elapsed`, `ok, elapsed :=`, `deadline.Sub(time.Now())`, `UnixNano`, and the struct field `x.HangCeiling`.
   - `TestAwaitConditionFailsFastAndKeepsTheResult` covers both the fast failure and the put-back.
   - Nothing tests a worker channel that is unbuffered or closed (see the Minor above).

6. **Architectural notes**
   - **ARCH-DRY: pass.** There is one helper; `testfix.AwaitCondition` and the weave sites alias or call it. The remaining hand-written loops (`release/main_test.go:64` and about 18 others) are listed as an M2 item.
   - **ARCH-PURE: pass.** The rule is pure AST code; the helper is a thin polling layer.
   - **ARCH-PURPOSE: pass.** The fix for BR-10 is a rule, not one site: there is one helper, and M2 converts every remaining loop to it.
   - **ARCH-MOCK: pass.** No seams change.
   - **ARCH-CONSTRAINTS: pass.** Waits are bounded only by `HangCeiling`.
   - **ARCH-SECURE: N/A.** Only test code and a lint rule change; no untrusted input or secrets.
   - **ARCH-ORDER: pass,** with the Minor above. Workers fail fast; `nil` sites give a reason.
   - **ARCH-FUNERAL: pass.** `wallclockPending` names its end: M2 deletes it, together with its test.
   - **ARCH-EVOLVE: N/A.** No persisted format changes.
   - For M2: keep requiring workers in the ~18 loop conversions. Struct-field deadlines are still a candidate for #626.

7. **Plan revisions:** none needed; the round-4 entry matches the code.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      atlas/workflow/sdlc-binary.md:2627 documents testgate.AwaitCondition(t, what, cond, workers), the required workers, nil only with a comment, the one-value buffer, the put-back, and that waitQueued and waitVouched are built on it.
  - id: BR-9
    disposition: addressed
    note: |
      rules.go bindings() covers AssignStmt N:N and ValueSpec for both matchers. A BinaryExpr ordering comparison mentioning time.Now() and the v.Sub(time.Now()) form count as polls. rules_test.go pins all four probes red, and pkg/golint passes.
  - id: BR-10
    disposition: addressed
    note: |
      AwaitCondition lives in pkg/testgate/await.go with testfix aliasing it. The weave release await takes workers via waitFor, and weavefs passes nil with its reason. The plan M2 checklist converts the remaining ~18 HangCeiling loops (TL-approved).
findings:
  - id: new
    severity: Minor
    kind: defect
    family: polled-wait-fail-fast
    title: |
      AwaitCondition never checks its one-slot-buffer precondition, so an unbuffered or closed worker channel makes it hang or panic
    detail: |
      3rd finding in family polled-wait-fail-fast. pkg/testgate/await.go:24-26: with an unbuffered worker the put-back w <- err blocks forever (the test hangs to the go test timeout), and with a closed channel the put-back panics. Rule: the helper checks every worker at entry (cap >= 1, else Fatalf) and receives with err, ok so a closed channel fails cleanly. Then no caller's channel can void the fast-failure guarantee. Current callers all use buffer 1, so prevalence is 0.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0415978a-7a64-4897-90a4-e2634805cb11","repository":"github.com/xianxu/ariadne","issue_id":"000622","recipe":"milestone-review","round":4,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/5c055fe98975cb64","reviewed_head":"6fc8c03fcc7836a127bc1e15c23847b90f199e38","slot":"ariadne:2","requested_at":"2026-10-25T15:31:35.755454Z"},"phase":"finished","pid":46083,"started_at":"2026-10-25T15:31:35.946068Z","finished_at":"2026-10-25T15:45:07.742064Z","session_id":"323b6578-042c-4695-af58-0b74f80364c8","outcome":"returned"}] -->
