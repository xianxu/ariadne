# Boundary Review — ariadne#421 (whole-issue close)

| field | value |
|-------|-------|
| issue | 421 — sdlc stage: one stage at a time in :0, a lease naming the holder; staging over another issue's live stage refuses |
| repo | ariadne |
| issue file | workshop/issues/000421-sdlc-stage-one-stage-at-a-time-in-0-a-lease-naming-the-holder-staging-over-another-issue-s-live-stage-refuses.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3dd288111bc10a361618e7f167abb864bd687b63..8ffb676c1b1458bc3cfbbb07f88bca5f294718a5 |
| command | sdlc close --issue 421 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T01:44:41-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I've read the diff and am now running the lease tests and checking the switch-failure path.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change does what the issue asks. The staged record is now also the lease, so there is still only one file per slot. The rules are pure functions in `stageplan.go` (`checkLease`, `checkClearLease`, `stageHolder`), and the three entry points that change the target (`stage`, `stage --clear` and `move`) all read and decide on the lease under a non-blocking flock. Each Done-when line has an integration test on `moveFixture`. I ran `TestStageLease*`, `TestCheckLease`, every `TestStage*` and `TestMove*`, `TestEveryCommitSiteStamps`, the recovery catalog tests and the fleet `StageLease` tests in this checkout, and they all pass. Two Important findings stand in the way of SHIP, and both are cheap to fix:
- **Failed switch leaves a lease behind:** when `git switch` fails, `runStage` keeps the lease it wrote just before the switch, but tells the user "nothing was staged".
- **README not updated:** README's "Concurrent issue work" section describes `stage` and `--clear` but not the new `--abandoned` flag or the one-stage-at-a-time refusal.

**1. Strengths**
- **One file, not two:** the staged record is extended into the lease instead of adding a second file to keep in step (ARCH-DRY). Old one-line records still parse as a commit with no holder (`internal/fleet/staged.go:127-156`).
- **Damaged leases are never read as free:** `ParseStageLease` requires all five keys in order and checks the time. `readStageLease` turns a parse error into a refusal that names the file (`stage.go:297-303`). Both the unit test and the integration test cover this.
- **Pure rules:** `checkLease` and `checkClearLease` are pure and table-tested. `--abandoned` only frees the lease when it names the current holder, and it accepts `418`, `#418` or `000418` (`stageplan.go:85-110`).
- **Lock covers re-check and write:** `runStage` takes the flock before its second `observeStage` and its lease write. A concurrent stage refuses, and a test covers that (`stage_test.go:582-589`).
- **Copyable recovery command:** the refusal prints the target's full address (for example `ariadne:0`), so the command it suggests can be copied from any slot.

**2. Critical findings**
None.

**3. Important findings**
- **Failed switch leaves the lease held** (`cmd/sdlc/stage.go:94-100`, ARCH-ORDER)
  - The holder's lease is written at line 94, before `git switch --detach` runs. If the switch then fails at line 98 (for example `--no-overwrite-ignore` hitting an ignored file), the function returns "…nothing was staged", but the target now carries this issue's lease on an unchanged checkout.
  - Every other issue's stage, `--clear` and `move` into that slot then refuses until someone runs `--clear` or `--abandoned`.
  - The plan accepts a held lease on an unchanged target after a *crash*. This is the ordinary error path, though, and the message is wrong about what happened.
  - Fix: if the switch fails, write back the previous lease `held` (or remove the record when the slot had none) before returning. Add a test that forces the switch to fail and checks that the record is back to what it was.
- **README update missing for `--abandoned` and the refusal** (`README.md:17-24`)
  - The "Concurrent issue work" section documents `sdlc stage --build` and `--clear` but says nothing about one stage at a time, the refusal, or `sdlc stage --clear :0 --abandoned <holder>`.
  - The atlas and `--help` text were updated; README was not.

**4. Minor findings**
- **Dry-run message:** `runUnstage` with `--dry-run` on a resting target that still has a lease prints "nothing staged", but a real run would release the lease (`stage.go:175-187`).
- **Stale observation in `runUnstage`:** `to` and `checkUnstage` are read before the lock and never re-read under it, unlike `runStage` and `runMove`. The lease read under the lock blocks the dangerous cross-holder case, so the remaining risk is small.
- **Leftover files:** `sdlc-staged.lock`, and a stray `.tmp` if a rename fails, stay in each worktree's git dir forever. That is bounded at one each per slot, so it is acceptable, but it isn't written down.
- **Fleet view:** `sdlc fleet` doesn't show the holder of a lease left on a resting slot, so the only way to find a held slot is to hit the refusal.
- **Asymmetric helpers:** `removeStaged` lives in `package main` while `WriteStageLease` lives in `fleet`. A `fleet.RemoveStageLease` would keep the record's whole lifecycle in one place.

**5. Test coverage notes**
- Every Done-when line is covered by `TestStageLeaseOneStageAtATime`: B is refused for stage, clear and move, the refusal names #1, its slot and time and the `--abandoned` command, and `:0` is unchanged afterwards. The same test covers `--abandoned` with the wrong holder and the right one, and the original holder restaging.
- `TestStageLeaseEdges` covers a lease left on a resting slot, a damaged lease, and a busy lock.
- Gap: no test fails the switch after the lease is written (see the first Important finding).
- Gap: no test for a legacy one-line record during a real stage. The unit tests cover parsing it, but not staging over it end to end.

**6. Architectural notes**
- **ARCH-DRY:** pass. One record serves as both the staged record and the lease. `takeStageLease` is shared by `stage` and `move`.
- **ARCH-PURE:** pass. The rules are pure and tested without IO. Locking and file writes happen in the thin `stage.go`/`move.go` layer.
- **ARCH-PURPOSE:** pass. All three verbs that change the target take the lease, so there is no path that bypasses it.
- **ARCH-MOCK:** pass. It follows the existing real-git fixture pattern and adds no new external dependency.
- **ARCH-CONSTRAINTS:** pass. The lock is non-blocking and is released before `--build`.
- **ARCH-SECURE:** pass. A damaged lease refuses, and the lease file is parsed into a typed value at the boundary.
- **ARCH-ORDER:** flag. The write order holds for crashes, but the switch-failure path leaves the lease held with no rollback (Important finding above).
- **ARCH-FUNERAL:** pass with a note. The record is removed by `--clear` or a move. The lock file lives as long as the slot.

**7. Plan revision recommendations**
- Add a Revisions entry under "Write order": "A *failed* switch (not a crash) restores the prior lease, so a refusal never leaves the target held."

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: failure-path-rollback-of-pre-written-state
    title: |
      runStage leaves the new holder's lease on the target when git switch fails, yet reports nothing was staged
    detail: |
      stage.go:94 writes the holder's lease before the switch at :98. A switch error returns "nothing was staged" without restoring the prior lease, so an unchanged target stays held and every other issue's stage, clear and move refuses. Fix: on switch failure, restore `held` (or remove the record when the slot had none) and add a test that forces the switch to fail.
  - id: new
    severity: Important
    kind: defect
    family: readme-missing-new-surface
    title: |
      README update missing for the stage lease refusal and the new --abandoned flag
    detail: |
      README.md:17-24 (Concurrent issue work) documents sdlc stage --build and --clear but not one stage at a time, the refusal naming the holder, or `sdlc stage --clear :0 --abandoned <holder>`. The atlas and --help text were updated; README was not.
  - id: new
    severity: Minor
    kind: defect
    family: dry-run-report-diverges-from-real-run
    title: |
      stage --clear --dry-run on a resting target with a held lease prints nothing staged
    detail: |
      stage.go:175-187: a real run releases the lease, but the dry run says nothing is staged and doesn't mention the lease it would release.
  - id: new
    severity: Minor
    kind: defect
    family: observe-outside-lock
    title: |
      runUnstage observes the target before taking the stage lock and never re-reads it under the lock
    detail: |
      runStage and runMove re-observe under the lock; runUnstage relies on `to` from before the lock. The lease read under the lock blocks cross-holder races, so the remaining risk is small.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"a1d6bdb4-d325-4746-9420-a9353b50e024","repository":"github.com/xianxu/ariadne","issue_id":"000421","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c486143584f05acd","reviewed_head":"8ffb676c1b1458bc3cfbbb07f88bca5f294718a5","slot":"ariadne:2","requested_at":"2026-10-17T08:38:46.976885Z"},"phase":"finished","pid":4814,"started_at":"2026-10-17T08:38:47.190071Z","finished_at":"2026-10-17T08:44:40.384619Z","session_id":"27fb9441-7c8c-4354-af52-9b05f31956e8","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T01:59:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 421 — sdlc stage: one stage at a time in :0, a lease naming the holder; staging over another issue's live stage refuses |
| repo | ariadne |
| issue file | workshop/issues/000421-sdlc-stage-one-stage-at-a-time-in-0-a-lease-naming-the-holder-staging-over-another-issue-s-live-stage-refuses.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3dd288111bc10a361618e7f167abb864bd687b63..767320e0dc23de298cf54b1951404f7efd6f2863 |
| command | sdlc close --issue 421 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-17T01:59:05-07:00 |
| verdict | SHIP |

## Review

The tests pass. Next I'll revert the BR-1 fix and the BR-3 fix one at a time to confirm the new tests fail without them.

```verdict
verdict: SHIP
confidence: high
```

All four prior findings are dealt with in the code. For BR-1 and BR-3 I removed each fix in the disposable checkout, and its new test failed. With the fixes back in place, the targeted stage, lease and move tests pass, along with the `fleet` and `recovery` packages. The README now covers the lease, the refusal and `--abandoned` (BR-2). BR-4's reorder is visible in the code: `runUnstage` now takes the lock, re-resolves the target, then observes it (`cmd/sdlc/stage.go:145-158`). But no test fails without the reorder, so under the claimed-fixes rule it stays `not-addressed`. It is Minor and doesn't block. I found nothing new that blocks the boundary. Every Done-when line and Plan checkbox has code and a test behind it.

**1. Strengths**
- **One file for both jobs.** The existing staged record is also the lease (`fleet/staged.go:119-207`), so there is no second file to keep in step (ARCH-DRY). A damaged record is an error and is never read as free (`ParseStageLease`).
- **The rule is pure and tested on its own.** `checkLease`/`checkClearLease` in `stageplan.go:723-741` have table tests in `TestCheckLease`. The IO code only locks, reads and writes.
- **`--abandoned` is a real acknowledgement.** It must name the current holder exactly, so it can't free a lease that has changed hands (`stageplan.go:737`).
- **BR-1 is fixed for the whole class, not just the named site.** Two things can fail after the lease is written. If the switch fails, the target never moved, so the prior lease is restored (`stage.go:99-104`). If the verify fails after the switch, the target did move, so it stays held, which is correct. `move.go` writes nothing before its switches. Nothing else in the class remains.
- **The lock's lifetime is bounded.** `runStage` releases it before `--build` (`stage.go:126-127`), and the kernel drops the flock if the process dies.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- **BR-4 (still open, test missing):** the reorder needs a regression test that fails under the old order. Example: hold `LockStage(:0)`, make `:0` dirty, run `runUnstage`, and assert the lock refusal ("another sdlc stage or move"). The old order returned the dirty-target error instead.
- **Comment and message overstate a failed switch:** the comment at `stage.go:100` says a failed switch "leaves the target as it was", and the error ends "nothing was staged". The new test's own comment says git had already updated the index and work tree before HEAD failed. Pre-existing #344 wording; worth rewording sometime.

**5. Test coverage notes**
- Revert checks I ran:
  - With `restoreStageLease` removed, `TestStageLeaseSwitchFailureRestores` fails: "the refused stage left a lease: {Holder:#1 …}".
  - With the dry-run branch disabled, `TestStageLeaseEdges` fails: the output reads "nothing staged" instead of "Would release".
- `TestStageLeaseOneStageAtATime` covers each Done-when line on three slots: another issue's stage, clear and move are refused and leave `:0` unchanged; `--abandoned` refuses the wrong holder and frees the right one; re-staging by the holder works; and B can stage once A has cleared.
- A busy lock and a damaged lease are both covered in `TestStageLeaseEdges`.

**6. Architecture**

| Principle | Result | Why |
|---|---|---|
| ARCH-DRY | pass | One record serves as both lease and staged record; `takeStageLease` is shared by stage and move |
| ARCH-PURE | pass | The rules and the holder naming are pure; the IO code is thin |
| ARCH-PURPOSE | pass | stage, `--clear` and move all go through the same lease check, so no consumer bypasses it |
| ARCH-MOCK | pass | No new external dependency; uses the existing real-git fixtures |
| ARCH-CONSTRAINTS | pass | The flock is non-blocking and refuses instead of waiting |
| ARCH-SECURE | pass | A damaged or partial record refuses and names the file; a bad time field is an error |
| ARCH-ORDER | pass | Write order is explicit: holder first, then switch, then commit; failure at each step is handled; lease state is read under the lock |
| ARCH-FUNERAL | pass | One lease per slot, removed by clear or move; the `.lock` file is one per slot and bounded |

Gap: tests can't inject ordering for the lock-versus-observe race (see BR-4).

For upcoming work: holder identity compares `stageHolder` strings built from two different sources, `observeMoveSide` and `workspace.Resolve`. Today they agree, and the test pins the slot format. If either address format changes, a detached holder would no longer match its own lease.

**7. Plan revision recommendations:** none. The Plan and Log match the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      restoreStageLease at stage.go:99-104; reverting it makes TestStageLeaseSwitchFailureRestores fail (lease Holder:#1 left behind); sibling post-switch failure paths correctly keep the lease.
  - id: BR-2
    disposition: addressed
    note: |
      README.md Concurrent issue work now names one stage at a time, the refusal naming the holder, and sdlc stage --clear :0 --abandoned <holder>; matches checkLease/checkClearLease behavior.
  - id: BR-3
    disposition: addressed
    note: |
      stage.go:185-188 prints Would release; disabling it makes TestStageLeaseEdges fail with nothing staged output.
  - id: BR-4
    disposition: not-addressed
    note: |
      Code reorder is correct (lock, re-resolve, observe at stage.go:145-158) but no test fails without it; add one holding LockStage on a dirty :0 and asserting the lock refusal, not the dirty-target error.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"2b986523-e6d8-482f-abae-cc81d1f6a0e9","repository":"github.com/xianxu/ariadne","issue_id":"000421","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c2a9a885c9fb8914","reviewed_head":"767320e0dc23de298cf54b1951404f7efd6f2863","slot":"ariadne:2","requested_at":"2026-10-17T08:50:03.668699Z"},"phase":"finished","pid":3284,"started_at":"2026-10-17T08:50:03.787787Z","finished_at":"2026-10-17T08:59:04.942463Z","session_id":"254a28b6-ee42-445c-8f38-29ef6048e1e0","outcome":"returned"}] -->
