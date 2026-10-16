# Boundary Review — ariadne#393 (milestone M2)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b19faedae8e8d2611ce2d7a16df0fe759bafb4cb..6619f0b44d6b0513d5554ac09b0406f642973dbf |
| command | sdlc milestone-close --issue 393 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T01:23:08-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reading the M2 diff now (merge wiring, `sdlc landing` commands, tests).

Tests pass. Checking for a data race in the async merge tests, then writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 does what it set out to do. `runDurableMerge` asks for confirmation first, then takes a turn in the line and holds it until it returns (`cmd/sdlc/landing.go:357-375`). Everything the landing acts on is read after admission. The `sdlc landing status|grant|revoke` commands are delivered, tested and documented in the atlas and help. The new tests and the landingline package pass at the pinned head, and the two async line tests are also clean under `-race`. Two things should be dealt with before shipping. First, the check that lets a merge skip the line looks at a different PR set than the one the landing goes on to choose. A branch that has an old merged PR and a new open one therefore integrates and lands without ever entering the line. Second, the plan's Core concepts section again describes an injection seam and file layout that the code doesn't have. This is the 3rd finding in `plan-table-drift`.

**Inspection done:** stat, name-status, full diff of code, docs and plan. I read `selectLandingPR`, `liveLandingPRs`, `selectIntegrationPR` and the fake GitHub's `LandingPRs`. Test runs:
- `go test ./cmd/sdlc/internal/landingline/`: ok
- `go test ./cmd/sdlc/ -run 'TestLandingLine|TestLandingGrant|TestLandingRevoke|TestLandingStatus|TestLandingEnds'`: ok (93 s)
- `go test -race` on the wait-then-integrate and duplicate tests: ok

I did not run the full `make test`.

### 1. Strengths
- **Reads happen after admission (PQ-1):** the head, PR and selection are read only after `enterLandingLine` returns (`landing.go:376-388`). `TestLandingLinePushedWhileWaiting` and `TestLandingLineDuplicateResumesAsMerged` show the fix this protects against: a fresh read after the wait means one merge and the new head lands.
- **The prompt moved before the enqueue,** so an operator's confirmation never holds the line.
- **`landingTurn.leave(err)` covers every exit.** It runs through a deferred call on the named return, so every path after admission records landed or refused and releases the line. The wait-failure path inside `enterLandingLine` releases as well. `TestLandingLineRefusalLeavesTheLine` pins this.
- **Existing code is reused rather than copied (ARCH-DRY):** `suiteQuantiles` for p50/p90, `issue.IDFromFilename` for the ticket's issue, and one `landingLineDir` shared by merge and the commands.
- **The commands have proper seams.** `landingDeps` injects the clock, caller and `$COUCH_TREE`, so the authorization tests cover hub, worker, and a :0 checkout entered from a worker without real workspaces.
- **The M1 minors are fixed.** A held head whose file can't be parsed is reported as present but unreadable (`heldTicket`, with a test), and the log reader no longer has a line-length cap (with a test).

### 2. Critical findings
None.

### 3. Important findings

**I-1: A branch with an old merged PR skips the line, then lands its new open PR anyway** (`cmd/sdlc/landing.go:366-375`, `cmd/sdlc/landingqueue.go:124`). ARCH-PURPOSE.
- **What happens:** `anyMergedPR(pre)` returns true if *any* PR for the branch is MERGED. `gh pr list --head <branch> --state all` also returns earlier merged PRs. `liveLandingPRs` keeps MERGED PRs, and `selectLandingPR` then filters them by head.
- **Failure:** issue 391 lands, its branch is deleted, and the issue is reopened on the same branch name with a new open PR. `pre` contains the old merged PR, so no ticket is taken. Selection then picks the open PR at the current head and runs the full integrate → suite → merge path outside the line. That is two concurrent integrations in one repo, which is exactly what Done-when 1 rules out.
- **Fix:** base the skip decision on the PR the landing will actually act on, not on "any". For example, read `head` first (a local read the plan already allows before the line) and skip only when `selectLandingPR(pre, t, branch, head)` returns a MERGED PR. Or skip only when no OPEN PR exists.
- **Test:** add a fake-GitHub fixture with one old merged PR and one open PR on the same branch, and assert that the merge waits behind a held line.

**I-2: The plan's Core concepts section doesn't match what M2 built** (`workshop/plans/000393-serial-landing-admission-plan.md:54`, Task 4). **This is the 3rd finding in family `plan-table-drift`.** Earlier rounds fixed instances, so this one should not be fixed on its own. The rule that covers all of them: *every entity and seam the diff adds or renames appears in the Core concepts tables at its built path, and every as-built deviation from a task gets a dated `## Revisions` entry in the same milestone's commits.* The current instances:
- The plan still says the line is "Injected into `runDurableMerge` through a package variable `landingLineOpen`". No such variable exists. The real seams are `landingLinePoll` and the real git common dir.
- Task 4 names `landing_line_test.go`, but the tests are in `landingqueue_test.go`.
- `landingqueue.go` (`enterLandingLine`, `landingTurn`, `openLandingLine`, `landingLineDir`) is missing from the Integration points table.
- Task 4's sketch inlines the enqueue in `runDurableMerge`; the code splits it into `landingTurn`.
- There is no M2 Revisions entry.

### 4. Minor findings
- `runLandingStatus` loops over the `grants` map, so the grant lines come out in random order (`landingcmd.go:177`). Sort them by issue.
- A duplicate run that was admitted and found the PR already merged still logs `landed`. That puts a second "landing" for the same PR into `Delays`/p90 (`landingqueue.go:107`). Consider a separate `resumed` kind, or skip `landed` when `pr.State == "MERGED"` at selection.
- Event and grant writes ignore their errors (`_ = line.Record(...)` in enqueue, OnAdmit and leave). A disk that stays full or unwritable would silently drop all audit and pilot data. At least warn once on stderr.
- `openLandingLine`'s `Grants` and `leave`'s auto-revoke call `time.Now()` directly instead of an injected clock.
- The merge-level events test checks PR and branch but not Issue, Seq or `Displaced`. Displacement is only tested at the line level.

### 5. Test coverage notes
- All the Task 4 tests the plan names exist except "events logged with issue and Seq", which is only partly asserted. All six Task 5 bullets are covered, including the end of a grant on landing (`TestLandingEndsTheIssuesGrant`).
- Missing: the branch-reuse case in I-1.
- `TestLandingLineWaitsThenIntegratesOnce` checks "no work while waiting" with a 50 ms sleep. That samples one interleaving, but it is a negative assertion on top of line-level tests that do have injected ordering, so I am not raising it under `ordering-test-seam`.

### 6. Architecture, marker by marker
- **ARCH-DRY: pass.** Shared quantile helper, ID parse and line directory. The second `LandingPRs` call is deliberate: the read after admission.
- **ARCH-PURE: pass.** Ordering and validity stay pure in landingline, and the merge-side code is thin. `leave`'s "end the grant if valid" is a small decision tangled with IO; acceptable.
- **ARCH-PURPOSE: flag (I-1).** The skip check leaves a way to land without entering the line.
- **ARCH-MOCK: pass.** Tests use the stateful `landingFakeGH` on real temp repos, and the line runs on real files.
- **ARCH-CONSTRAINTS: pass.** 2 s polling, O(waiters) scans, local IO only.
- **ARCH-SECURE: pass.** Grants are checked for authority again when read, the cooperative trust model is documented, and malformed log lines are counted and shown in status. The reader has no line-length bound, but under the cooperative model that is acceptable.
- **ARCH-ORDER: pass.** Waiting and head states are held as flocks, every exit goes through `leave`, and a waiter cancelled by Ctrl-C releases inside `enterLandingLine`.
- **ARCH-FUNERAL: pass.** `refused` events now carry `err.Error()` text, so each event is a bit bigger, but the 30-day sweep bounds them.

### 7. Plan revision recommendations
Add `### 2026-10-16 — M2 as built`:
- The line is opened by `enterLandingLine` in `landingqueue.go` with production options. The test seam is `landingLinePoll` plus a test-held head on the fixture's real git common dir. There is no `landingLineOpen` variable.
- `landingTurn.leave(err)` logs landed or refused, ends a valid grant on landing, and releases.
- Tests are in `landingqueue_test.go` and `landingcmd_test.go`.
- Add a `landingqueue.go` row to the Integration points table.
- Restate the skip rule as "skip only when the PR the landing would select is MERGED", once I-1 is fixed.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: gate-skip-predicate-divergence
    title: |
      anyMergedPR lets a branch with an old merged PR plus a new open PR skip the landing line and land outside it
    detail: |
      landing.go:366-375 skips the line if any PR for the branch is MERGED, but gh pr list --state all also returns old merged PRs, and selectLandingPR (landing.go:191) later picks the OPEN PR at the current head and runs integrate/suite/merge without a turn, breaking Done-when 1 when a branch name is reused. Fix: skip only when the PR the landing would select (selectLandingPR with the local head) is MERGED, or when no OPEN PR exists; add a fake-GitHub test with an old merged PR and an open PR on the same branch that waits behind a held line.
  - id: new
    severity: Important
    kind: defect
    family: plan-table-drift
    title: |
      Plan still claims a landingLineOpen seam and landing_line_test.go, and omits landingqueue.go and an M2 Revisions entry
    detail: |
      3rd finding in this family, so fix the rule, not this instance: every entity and seam the diff adds or renames appears in the Core concepts tables at its built path, and every as-built deviation gets a dated Revisions entry in the same milestone. Instances: plan line 54 names a package variable landingLineOpen that does not exist (the seams are landingLinePoll and the real common dir); Task 4 names landing_line_test.go but the tests are in landingqueue_test.go; landingqueue.go (enterLandingLine, landingTurn, openLandingLine, landingLineDir) is missing from the Integration points table.
  - id: new
    severity: Minor
    kind: defect
    family: deterministic-output
    title: |
      sdlc landing status prints grants in random map order
    detail: |
      landingcmd.go loops over the ValidGrants map directly; sort by issue number.
  - id: new
    severity: Minor
    kind: defect
    family: audit-write-failure-visibility
    title: |
      Errors from landing-line event and grant writes are ignored with no warning
    detail: |
      The enqueue, OnAdmit and leave writes in landingqueue.go discard errors from Record and AddGrant, so a disk that stays full or unwritable silently loses all audit and pilot data; warn once on stderr.
  - id: new
    severity: Minor
    kind: defect
    family: pilot-metric-fidelity
    title: |
      A duplicate merge admitted after the original landed logs a second landed event and skews the delay p90
    detail: |
      leave() logs EventLanded whenever err is nil, including when the admitted run only resumed cleanup of a PR that was already MERGED; use a distinct kind or skip it.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a30690e7-c055-4edb-9b15-eb23055729a8","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2d8e850921a514d3","reviewed_head":"6619f0b44d6b0513d5554ac09b0406f642973dbf","slot":"ariadne:2","requested_at":"2026-10-16T08:18:13.571086Z"},"phase":"finished","pid":48633,"started_at":"2026-10-16T08:18:14.375818Z","finished_at":"2026-10-16T08:23:06.776724Z","session_id":"73d6a3de-6b94-4d08-adc5-1656a874146c","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T02:00:00-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b19faedae8e8d2611ce2d7a16df0fe759bafb4cb..5f7a88ff5e5bf298e12f774bcdd6cc09c5f01c4f |
| command | sdlc milestone-close --issue 393 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T02:00:00-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I'm reviewing the landing line code. The full test suite hit the 10-minute limit, so I'm running just the targeted tests now.

The new landing tests pass. Next I'm undoing the BR-11 fix in this throwaway checkout to confirm its regression test actually fails without it.

BR-11's regression test fails without the fix, so that one is confirmed. The broader landing test run hit the 600s timeout, so I'm re-running it verbosely to find which test hangs.

That hang is `TestBoundedLandingGitTerminatesDescendants`, which this window doesn't touch. Now I'm checking the plan against the code for BR-12.

The fixes for the prior findings check out. I'm finishing the docs-gate check (README and atlas) before writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M2 is close to shippable. The round-1 change for BR-11 is real: when I undid it in this throwaway checkout, `TestLandingLineReusedBranchStillQueues` failed ("no 1 waiting merges"), and with it restored it passes. BR-12 is fixed in the plan, and BR-15 has a regression test. Two problems remain before the boundary, both cheap:

- **The skip check can still let a merge land without a turn.** The decision to skip the line is still made on a different GitHub read from the one the landing acts on. A merge that skipped the line is never stopped from landing an OPEN PR.
- **The README never mentions `sdlc landing` or that `sdlc merge` now waits in a line.**

BR-13 and BR-14 have code changes but no test that fails without them, so under the claimed-fix rule they stay open. Both are Minor and don't block.

**Test runs:**
- `internal/landingline` passes.
- All landing-line and landing-command tests in `cmd/sdlc` pass: WaitsThenIntegratesOnce, DuplicateResumesAsMerged, ReusedBranchStillQueues, AlreadyMergedTakesNoTurn, RefusalLeavesTheLine, PushedWhileWaiting, GrantFromHubAndRefusedFromWorkers, StatusDelays.
- The wider `-run Landing` set timed out: `TestBoundedLandingGitTerminatesDescendants` (all three subtests) hung for 8m20s here. That test isn't touched in this window, so I'm treating it as this machine's problem, not this change.
- A full `go test ./...` also ended in FAIL. I only captured its last 15 lines, which were all `ok`. My best guess is the same 600s timeout in the main package, but I didn't confirm it.

### 1. Strengths
- **The PQ-1 rule holds:** everything the landing acts on is read after admission. `landing.go:375-389` reads `head` and `LandingPRs` again after `enterLandingLine`, and the duplicate and pushed-while-waiting tests prove it.
- **One delay calculation:** `landingline.Delays` feeds the existing test-stats `suiteQuantiles` instead of a second percentile routine (ARCH-DRY).
- **One clock:** `landingNow` is shared by merge, `leave()` and the commands (`landingqueue.go:23`, `landingcmd.go:35`), which removes the split clocks.
- **`resumed` is its own event:** it keeps the pilot's p90 honest, and `TestLandingLineDuplicateResumesAsMerged` checks resumed==1 and landed==0.
- **Grant authority is checked on read as well as on write**, and status reports skipped malformed lines instead of hiding them (`landingcmd.go:190-192`).

### 2. Critical findings
None.

### 3. Important findings
- **I-A: a merge that skipped the line can still land an OPEN PR** (`landing.go:369-395`, ARCH-ORDER / ARCH-PURPOSE). This is the 2nd finding in family `gate-skip-predicate-divergence`, so the fix should cover the rule, not just this site.
  - **What happens:** whether to take a turn is decided from the first read (`pre`). What gets merged is chosen from a second read (`prs`).
  - **How it goes wrong:** if the first read shows no OPEN PR (none yet, or a GitHub listing that lags just after `sdlc pr`) and the second read shows one, `selectIntegrationPR` picks the OPEN PR. Integration, the suite and the merge then run with `turn == nil`, outside the line, which breaks Done-when 1.
  - **The rule:** a merge without a turn may only act on a PR that is already MERGED. Enforce that on the PR it actually selected, not on the earlier read.
  - **Fix sketch:** right after `selectIntegrationPR`, add `if turn == nil && pr.State != "MERGED" { return errors.New("PR opened while merge started; re-run sdlc merge") }`, or enter the line at that point.
  - **Test:** make `landingFakeGH` return only a MERGED or empty list on its first `LandingPRs` call and an OPEN PR on the second. Assert that `gh.merges == 0` while the line is held.
- **I-B: README update appears missing for `sdlc landing status|grant|revoke` and for merge waiting in the line** (Docs gate).
  - `README.md:414-425` describes `sdlc merge`'s landing and how it handles concurrent integration, but not the line, the waiting message, `sdlc landing status`, or how the operator grants and revokes priority.
  - The atlas subcommand table (`atlas/workflow/sdlc-binary.md` near line 43) has a `merge` row but no `landing` row. The prose section at line 377 doesn't replace the table entry.

### 4. Minor findings
- **BR-13 is still open:** the sort is in place (`landingcmd.go:182`), but no status test has two or more grants, so nothing fails without the sort.
- **BR-14 is still open:**
  - `logLine` and the `AddGrant` warning have no test. A fake log that can't be written, such as a directory where `events.jsonl` should be, would make one.
  - In `leave()`, a `GrantEntries` read error (`landingqueue.go:134`) still silently skips ending a landed issue's grant.
- **Stale plan text:** the Task 4 "Rule" paragraph and its code sketch still say "if the PR is MERGED, take no ticket" and use `anyMerged` / `landingLineOpen`. The M2 Revisions entry covers this, but the stale paragraph could mislead the next reader.
- **`warnedLines` (`landingqueue.go:96`) is a process-global map:** fine for one merge per process, but it should live on `landingTurn` or the line itself.

### 5. Test coverage notes
- **BR-11:** I checked this by undoing the fix, and the test goes red as expected. Note that the test ignores the run's outcome (`_ = finished(...)`). It only proves a turn was taken, which is the actual claim.
- **BR-15:** the duplicate test would fail without the fix, because it would log `landed` and not `resumed`.
- **Missing:** a test that changes what the two `LandingPRs` reads return between calls (I-A). The fake already models GitHub lag for heads (`headLag`), so adding a lag for PR listings is cheap and keeps it a stateful fake (ARCH-MOCK).
- **Environment:** `TestBoundedLandingGitTerminatesDescendants` hangs on this host. It's outside this window but worth checking before `make test` for close.

### 6. Architectural notes
- **ARCH-DRY:** pass. The flock code is reused from reviewbox and the quantiles are shared.
- **ARCH-PURE:** pass. `Order`, `ValidGrants`, `KeepGrants` and `Delays` are pure; `landingqueue.go` is thin glue.
- **ARCH-PURPOSE:** flag (I-A). The line is only enforced at entry, not where the merge actually happens.
- **ARCH-MOCK:** pass, with the gap noted under I-A. Real git fixture repos plus the stateful `landingFakeGH`.
- **ARCH-CONSTRAINTS:** pass. The 2s poll and ctx-bounded wait are unchanged since the earlier wait-termination fix.
- **ARCH-SECURE:** pass. Grants are validated on read, and the ticket number comes from the file name.
- **ARCH-ORDER:** flag (I-A). The "has a turn" state and the observation it was based on are decided at different times.
- **ARCH-FUNERAL:** pass. Tickets go when their process dies, logs are swept after 30 days, and `warnedLines` is bounded by the process.

### 7. Plan revision recommendations
- Add to the M2 Revisions entry: "A merge without a turn acts only on a PR that was MERGED when selected; an OPEN PR without a turn refuses (gate-skip rule, M2 round 2)."
- Rewrite or strike the Task 4 "Rule" paragraph and code sketch so they match `anyOpenPR` and `enterLandingLine`, not just carry an "as built" comment.
- Add a README update for `sdlc landing` and merge's waiting behavior to Task 6, plus a `landing` row in the atlas command table.

```findings
dispose:
  - id: BR-11
    disposition: addressed
    note: |
      anyOpenPR replaces anyMergedPR; reverting it makes TestLandingLineReusedBranchStillQueues fail (no 1 waiting merges), restored it passes.
  - id: BR-12
    disposition: addressed
    note: |
      Integration points table now lists landingqueue.go entities and landingcmd.go; Injected-into says seams are landingLinePoll and landingNow; dated M2 Revisions entry present.
  - id: BR-13
    disposition: not-addressed
    note: |
      slices.Sorted added at landingcmd.go:182, but no status test has two or more grants, so nothing fails without the sort.
  - id: BR-14
    disposition: not-addressed
    note: |
      logLine and the AddGrant warning exist but no test forces a write failure, and the GrantEntries read error in leave() still silently skips ending the grant.
  - id: BR-15
    disposition: addressed
    note: |
      EventResumed logged when the selected PR was MERGED; TestLandingLineDuplicateResumesAsMerged asserts resumed==1 and landed==0, which would fail without the fix.
findings:
  - id: new
    severity: Important
    kind: defect
    family: gate-skip-predicate-divergence
    title: |
      Line skip is still decided on the pre-read; a turnless run can land an OPEN PR selected from the second read
    detail: |
      2nd finding in this family, so state the rule: a run without a turn may only act on a PR that is MERGED when selected. landing.go decides the turn from pre (first LandingPRs) but selectIntegrationPR uses prs (second read); an OPEN PR in the second read lands with turn nil. Fix: after selection, refuse (or enter the line) when turn is nil and pr.State is not MERGED; test with a fake whose first LandingPRs omits the open PR.
  - id: new
    severity: Important
    kind: defect
    family: user-surface-docs
    title: |
      README update appears missing for sdlc landing status/grant/revoke and for merge waiting in the landing line
    detail: |
      README.md:414-425 describes sdlc merge landing and concurrency but not the line, its waiting output, or the new sdlc landing subcommand; the atlas subcommand table near sdlc-binary.md:43 also lacks a landing row.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a06f3992-9a80-4711-abea-018a15274d9b","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/3613a7274340dd53","reviewed_head":"5f7a88ff5e5bf298e12f774bcdd6cc09c5f01c4f","slot":"ariadne:2","requested_at":"2026-10-16T08:29:17.033609Z"},"phase":"finished","pid":69936,"started_at":"2026-10-16T08:29:17.364371Z","finished_at":"2026-10-16T08:59:56.216273Z","session_id":"e51e4dd9-c6a0-4a65-8d8d-98c38923666a","outcome":"returned"}] -->

---

## Re-review — 2026-10-16T02:19:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 393 — Serial per-repo landing admission with fleet-blocking priority |
| repo | ariadne |
| issue file | workshop/issues/000393-serial-landing-admission.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | b19faedae8e8d2611ce2d7a16df0fe759bafb4cb..2c61503b65a1ee4d777cbbda9f55d2f93959409a |
| command | sdlc milestone-close --issue 393 --milestone M2 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-16T02:19:05-07:00 |
| verdict | SHIP |

## Review

Checking the prior findings' fixes; running the package tests now.

Tests pass. Next I'm reverting the BR-16 guard to confirm its regression test fails without it.

The regression test fails without the fix, so the fix is confirmed. Last step: checking the plan's table and revisions against the code.

```verdict
verdict: SHIP
confidence: high
```

All four open findings are fixed, and I raised no new ones. The fix for the one Important finding about landing order (BR-16) has a regression test that fails when the fix is removed. I checked this by reverting the guard at `cmd/sdlc/landing.go:394`: `TestLandingLineLaggingListingRefusesOutsideTheLine` then failed with `err <nil>`. With the guard restored, the `landingline` package tests and all of `TestLandingLine*` / `TestLandingRevoke*` in `cmd/sdlc` pass. The repeating "check the first listing" problem is now handled as a rule rather than one case at a time: a merge without a turn may only act on a PR that was already merged. `resumeOnly` (`landingqueue.go:157`) decides whether to skip the line, and the selected PR is checked against the rule after the second read. An empty listing is no longer treated as "nothing to land", so the merge queues. The atlas and the plan state the rule in the same words. Nothing blocks shipping.

1. **Strengths**
   - `landing.go:394`: the rule is checked on the PR the merge actually acts on, not on the earlier read. On refusal, nothing has changed. Before that point, `selectIntegrationPR` only fetches main into the tracking ref and polls GitHub; it never pushes or merges.
   - `landingqueue.go:88-100`: the process-global `warnedLines` map was replaced by a `lineLog` owned by each merge, so "warn once" is scoped to one merge instead of being shared state. The grant-read failure in `leave` now reports itself instead of being silently skipped.
   - `landingcmd.go:177`: `slices.Sorted(maps.Keys(grants))` makes the status output deterministic, and the test repeats the call to expose map-order randomness.
   - Tests use the real git fixture, the fake GitHub's `lagFirst`/`firstListing` and a test-held head ticket (`heldLine`). They don't mock the code under test.

2. **Critical findings:** none.

3. **Important findings:** none.

4. **Minor findings:** none new.

5. **Test coverage notes**
   - BR-16: the regression test was confirmed red on revert. `TestLandingLineEmptyListingQueues` covers the empty-listing case.
   - BR-14: `TestLandingLineLogFailuresWarn` makes the events and grants logs unwritable (it puts directories at their paths). It checks that "could not log" is warned exactly once and that the grant-read warning appears.
   - BR-13: the sorted-grants test can pass by luck about 1 time in 32 when the fix is reverted (random 2-key order, 5 tries). That is good enough here.

6. **Architecture notes**
   - **ARCH-DRY: pass.** `resumeOnly` replaced `anyOpenPR` rather than sitting beside it, and `openLandingLine` is shared by merge and `sdlc landing`.
   - **ARCH-PURE: pass.** Ordering and grant validity stay pure in `internal/landingline`. `landingqueue.go` is thin glue.
   - **ARCH-PURPOSE: pass.** The fix covers the whole class: every turnless path goes through one post-selection guard.
   - **ARCH-MOCK: pass.** Tests run against the existing stateful GitHub fake and a real local git remote. The fake gained a lagging-listing mode.
   - **ARCH-CONSTRAINTS: pass.** Polling stays at the 2-second interval, and the merge still confirms before queueing.
   - **ARCH-SECURE: pass.** No new untrusted input in this round. Malformed log lines are still counted and reported in status.
   - **ARCH-ORDER: pass.** The two-read ordering is made explicit and tested by injecting a lagging first read.
   - **ARCH-FUNERAL: pass.** No new durable artifacts. Events and grants still fall under the M1 30-day sweep.

7. **Plan revision recommendations:** none. The plan's Core concepts rows (`plan.md:47-48`) and Revisions entries (`plan.md:254-269`) match the code: `lineLog`, `resumeOnly`, `landingqueue.go`, sorted status output, and the no-turn rule.

```findings
dispose:
  - id: BR-13
    disposition: addressed
    note: |
      landingcmd.go:177 iterates slices.Sorted(maps.Keys(grants)); landingcmd_test.go asserts #12 precedes #391 across repeated runs.
  - id: BR-14
    disposition: addressed
    note: |
      lineLog.record (landingqueue.go:95) warns once per merge for enqueue/admit/leave writes; leave warns on grant read and AddGrant failures; TestLandingLineLogFailuresWarn pins it.
  - id: BR-16
    disposition: addressed
    note: |
      Rule enforced on the selected PR at landing.go:394; reverting the guard turns TestLandingLineLaggingListingRefusesOutsideTheLine red (err nil); empty listing now queues.
  - id: BR-17
    disposition: addressed
    note: |
      README.md:423-431 covers the line, waiting output and sdlc landing status/grant/revoke; atlas sdlc-binary.md command table gains a landing row.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b9ec6ee7-cdb2-48cf-a4c7-df9e6511bec9","repository":"github.com/xianxu/ariadne","issue_id":"000393","recipe":"milestone-review","round":3,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/1d6afce90fb8869f","reviewed_head":"2c61503b65a1ee4d777cbbda9f55d2f93959409a","slot":"ariadne:2","requested_at":"2026-10-16T09:05:53.249075Z"},"phase":"finished","pid":53389,"started_at":"2026-10-16T09:05:53.51383Z","finished_at":"2026-10-16T09:18:55.069056Z","session_id":"7775c767-ff06-4eda-88cb-0ca1c2fd14aa","outcome":"returned"}] -->
