---
id: 000592
status: done
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: '2acfe447403337a2ff54d81c965ac37415d3ac05' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-24T04:45:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause]
flow: {kind: quick, provenance: inferred, spec: "096ee224", done: "a0bbe98b"}
actual_hours: 0.52
---

# flake: TestQueueRealHelperPause

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestQueueRealHelperPause: failed in attempt c5fff316216ec03a781b4796f770a3e5, passed on the same head in attempt f3122a02c4fdbbaba2e5dd1f03b2b89f

## Spec

It isn't a timeout. Both recorded failures (attempt 81a77a06 on b9611a8a, and 7e509591 on the
#588 merge head 7108f40d) fail at `test-lock-pause.test.py:255`, `assertFalse(path.exists())`:
the operation file is still there.

`QueueLifetime.retire` (`scripts/test-lock.py:574`) runs one transaction under `control.lock`:
write `unresolved/<id>.json`, write `interrupted/<id>.json`, unlink `operations/<id>.json`. The
test's wait (`:246`) polls for the two evidence files without that lock, so it can observe the
transaction half done; load widens the window between the second write and the unlink.
Production readers don't have the race: queuegate `Observe` reads under `WithGate`, the same
lock.

Reproduced deterministically: a `sitecustomize` shim on `PYTHONPATH` that sleeps 0.5s before
the helper unlinks an `operations/*.json` fails line 255 in 6 of 6 runs.

The same class, found by a sweep of the queue-helper tests' unlocked reads of state the helper
writes in one locked transaction:
- `test-lock-exit.test.py:129` (`unknown()`) waits for `unresolved/` only, then `:273` reads
  `interrupted/`, which the same transaction writes second.
- `test-lock-pause.test.py:229-232`: the child owner writes its pid with a non-atomic
  `write_text`, and the test reads it as soon as the file exists. A slowed run read it empty
  (`ValueError: invalid literal for int()`), 1 in 10.

Other waits read files written atomically or ordered by a barrier (`make.pid` before `ready`;
`priority:436` guards with `isdigit`).

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [x] Fixture waits that observe queue state take the helper's `control.lock` without blocking
      (`lock.try_lock`): a predicate is evaluated only while the lock is held, so it sees
      committed transactions only. Used at `pause:246`, where it waits for the settled state (both
      evidence files present, operation file gone), and in `exit` `unknown()`, where the
      unresolved and interrupted files then appear together.
- [x] The child-owner fixture writes its pid file atomically (temporary file + replace).
- [x] Regression: the pause test case also runs with the slowed-unlink shim on the helper's
      `PYTHONPATH`; it fails before the fix and passes after.
- [x] Verify: the shim case repeated 20 times, both scripts plain, and `TestQueueRealHelperPause`
      and `TestQueueRealHelperCompletion`.

## Log

### 2026-10-24
- 2026-10-24: closed — Round-1 SHIP Minor (ARCH-DRY duplicated committed/committed_read and shim PYTHONPATH setup) fixed in 6eb2f26b: one shared scripts/test/fixtures/queue_reads.py. The pause, exit and decision scripts pass, both slowed-write regressions pass 10/10, TestQueueRealHelperPause, TestQueueRealHelperCompletion and ./scripts/test pass. Pre-close suite on 6eb2f26b: the only failure, TestHelperRefreshRunnerResourceExcluded (unrelated Go code), was registered as a known flake (#596) and passed on its same-head rerun. --no-atlas: test-fixture-only change. (fixed-then-ship); review verdict: SHIP
- 2026-10-24: closed — Root cause: unlocked test reads of queue state the helper changes in transactions under control.lock (retire writes unresolved/, interrupted/, then unlinks operations/; rewrite truncates and pwrites operations/ in place). Fixture waits and every operations/ read now hold control.lock (committed, committed_read); the child pid file is written atomically. Regressions under the shared slow-queue-writes shim: pause test_fast_cancellation_waits_for_the_whole_retirement and exit test_cancellation_wait_reads_the_record_whole each fail 3/3 with the wait unlocked (line-255 assertion; the #589 JSONDecodeError) and pass 10/10 fixed. The pause, exit and decision scripts pass; TestQueueRealHelperPause and TestQueueRealHelperCompletion pass; full pre-close suite exit 0 on 468ee2e4 (sdlc-local/suite). --no-atlas: test-fixture-only change, no new surface.; review verdict: SHIP
- Root cause: an unlocked fixture read of the helper's locked retire transaction, not an overrun; reproduced with a slowed-unlink shim (see Spec).
- Implemented as approved by the TL. A `committed(predicate)` fixture wrapper in each script evaluates the predicate only while it holds `control.lock` (`try_lock`, non-blocking; busy reads as not yet). The pause case now waits for the operation file to be gone too.
- Regression `test_fast_cancellation_waits_for_the_whole_retirement` runs the case with the slowed-unlink `sitecustomize` shim. With the old unlocked wait put back it fails 5 of 5 at the original assertion; fixed, both cases pass 20 of 20. Both scripts pass in full (19 and 10 tests), and `TestQueueRealHelperPause` and `TestQueueRealHelperCompletion` pass. These need to run unsandboxed (boot identity).
- TL flagged ariadne:3's #589 suite: a `JSONDecodeError` in `test-lock-exit.test.py:269` (`test_cancellation_wins_before_recipe_completion`). It is the same class, and my first sweep missed it because it listed exists-then-read patterns, not reads inside wait predicates. `QueueLifetime.rewrite` truncates and rewrites the operation record in place, because its descriptor is the lifetime lease and can't be renamed. Any unlocked read while a helper or watcher is alive can find it empty. Reproduced: a shim that sleeps 0.3s before `os.pwrite` fails that wait 3 of 3 with the same error.
- Sweep, by record family: `operations/` is rewritten in place, so every test read of it now goes through `committed_read` or `committed`. That covers exit `start_make`, the supervisor and watcher-crash cases and the cancellation wait, plus pause waiting/running and both managed cases. `unresolved/` and `interrupted/` are written once, atomically, so a read after the locked wait is safe. The priority test already reads its tickets under the tickets lock, and the decision test reads its in-process gate with no concurrent writer.
- One shared shim, `scripts/test/fixtures/slow-queue-writes/sitecustomize.py`, slows `os.pwrite` and the unlink of operation records. Regressions: `test_fast_cancellation_waits_for_the_whole_retirement` (pause) and `test_cancellation_wait_reads_the_record_whole` (exit). Each fails 3 of 3 with its wait reverted to unlocked and passes 10 of 10 fixed. The pause, exit and decision scripts pass in full, and so do `TestQueueRealHelperPause` and `TestQueueRealHelperCompletion`.
- Close review round 1: SHIP, one Minor (ARCH-DRY): `committed`/`committed_read` and the shim's `PYTHONPATH` setup were duplicated across the two scripts. Fixed in the same round by moving them into `scripts/test/fixtures/queue_reads.py` (`committed`, `committed_read`, `slow_writes`), which both scripts load. The scripts pass, both regressions pass 10 of 10, the Go wrappers pass, and so does `./scripts/test`.
