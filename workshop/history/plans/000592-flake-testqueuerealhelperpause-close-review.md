# Boundary Review — ariadne#592 (whole-issue close)

| field | value |
|-------|-------|
| issue | 592 — flake: TestQueueRealHelperPause |
| repo | ariadne |
| issue file | workshop/issues/000592-flake-testqueuerealhelperpause.md |
| boundary | whole-issue close |
| milestone | — |
| window | b3a1850018e1664c2dc38b21e7075a77fa748d8d..468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94 |
| command | sdlc close --issue 592 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T06:19:14-07:00 |
| verdict | SHIP |

## Review

Diff and helper read; now checking whether capacity-file reads at exit:231/271 share the same in-place-rewrite race.

The sweep looks complete: capacity reads are ordered behind the `ready` barrier. Next I'll try running the regression tests.

The tests pass. Now I'll revert the fixes in this disposable checkout to confirm the regression tests fail without them.

Both regressions fail 3 of 3 with the fix reverted, and the checkout is restored. Last, checking whether the test scripts already have a shared helper module (for the duplication question).

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** Nothing blocks the close. The fix addresses the root cause and doesn't paper over it with retries. `QueueLifetime.retire` (`scripts/test-lock.py:574`) and `QueueLifetime.rewrite` (`:540-544`) change queue state in transactions held under `control.lock`. The test fixtures used to read that state without the lock, so they could see a transaction half done. They now check their conditions only while holding `control.lock`, taken without blocking. I checked the regression evidence in this disposable checkout. With the fix, both new regressions and their unslowed siblings pass. When I put back the old unlocked waits, the pause regression failed 3 of 3 and the exit regression failed 3 of 3 with `JSONDecodeError: Expecting value`, which is the reported symptom. I then restored the checkout. No `## Done when` clause names two modes, and the single clause holds: the fix adds no retry loop and no longer timeout. The only findings are two Minor duplications under ARCH-DRY.

1. **Strengths**
   - **Release on close:** `try_lock` (`scripts/test-lock.py:1306`) uses `flock(LOCK_EX|LOCK_NB)`, so closing the descriptor in `committed()`'s `finally` releases the lock every time. A busy lock reads as "not yet", which can't deadlock the helper.
   - **The pause wait now checks the settled state:** at `test-lock-pause.test.py:282` it also requires the operation file to be gone. That is exactly what the later `assertFalse(path.exists())` asserts.
   - **The sweep is done by record family, and it is complete.** I re-checked every unlocked read in both scripts:
     - Every read of an `operations/` record now goes through `committed`/`committed_read`.
     - `unresolved/` and `interrupted/` are written once and atomically, so reading them after the locked wait is safe.
     - The capacity reads at exit:231/271 happen after the `ready` barrier. That barrier is reached only after the helper finishes its locked capacity write (`test-lock.py:1473-1476`) and returns.
     - Pause:221 reads capacity after the helper has exited.
     - `completion.jsonl` is read after the process has been waited on.
   - **The pid file is now written atomically** (temporary file, then `os.replace`) at pause:262-263. That closes the empty-read `int()` failure.
   - **One shared shim drives both regressions,** and it is reachable: the watcher the helper starts inherits `PYTHONPATH`, and the red runs above show the shim takes effect.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **ARCH-DRY:** `committed` and `committed_read` are copied word for word in `test-lock-exit.test.py:43-64` and `test-lock-pause.test.py:97-118`. The only difference is `HELPER.try_lock` versus `lock.try_lock`. The Python test-lock suites have no shared fixture module today, so pulling these out would mean creating one. That is reasonable to defer.
   - **ARCH-DRY:** the code that puts the shim on `PYTHONPATH` appears twice: `test-lock-exit.test.py:110-112` and `test-lock-pause.test.py:256-257`. It belongs with the helpers above.

5. **Test coverage notes**
   - Each regression fails without its fix and passes with it; I confirmed this here, 3 of 3 red when reverted.
   - The unslowed cases still run, so the normal path stays covered.
   - The other `committed_read` sites (supervisor, watcher-crash, the waiting and running phase checks, and the managed cases) have no slowed variant. They use the same wrapper the regressions already prove, so this is acceptable.

6. **Architectural notes**
   - **ARCH-DRY:** two Minor findings, listed above.
   - **ARCH-PURE:** passes. The change is limited to test fixtures and adds no business logic.
   - **ARCH-PURPOSE:** passes. The root cause is fixed, and the class was swept by record family rather than just the one line that failed.
   - **ARCH-EVOLVE:** passes. No persisted format changed.
   - **Docs:** no atlas or README update is needed, since the diff adds no user-facing surface.
   - **For later:** if more test suites need locked reads, a shared `scripts/test/fixtures` Python module holding `committed`, `committed_read` and a `slow_env()` helper would remove both duplications.

7. **Plan revision recommendations:** none. The plan still matches the code.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: duplicated-test-fixture-helper
    title: |
      ARCH-DRY: committed/committed_read and the shim PYTHONPATH setup are duplicated across the two test scripts
    detail: |
      Instances: test-lock-exit.test.py:43-64 and test-lock-pause.test.py:97-118 are word-for-word copies apart from HELPER.try_lock vs lock.try_lock. The shim PYTHONPATH construction appears at test-lock-exit.test.py:110-112 and test-lock-pause.test.py:256-257. Consider one shared fixture module under scripts/test/fixtures if more suites adopt locked reads.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"18c2d474-635e-4e18-82b6-b8b6ffc211eb","repository":"github.com/xianxu/ariadne","issue_id":"000592","recipe":"small-diff-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/73f1a73dc6e3f297","reviewed_head":"468ee2e4bcbd8df4c4c11b28f7bb781b9ac45a94","slot":"ariadne:5","requested_at":"2026-10-24T13:17:37.172065Z"},"phase":"finished","pid":74607,"started_at":"2026-10-24T13:17:37.306114Z","finished_at":"2026-10-24T13:19:14.296472Z","session_id":"8892a979-c8af-4f63-854d-0ff754996d41","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T08:35:13-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 592 — flake: TestQueueRealHelperPause |
| repo | ariadne |
| issue file | workshop/issues/000592-flake-testqueuerealhelperpause.md |
| boundary | whole-issue close |
| milestone | — |
| window | 494fd121a17bb638cad3f5383c15f0b622896031..6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba |
| command | sdlc close --issue 592 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-24T08:35:13-07:00 |
| verdict | SHIP |

## Review

Still checking the fix: the shared fixture reproduces the helper's lock semantics exactly. I'm sweeping the remaining `control.lock` sites to see whether the duplication was fixed everywhere.

```verdict
verdict: SHIP
confidence: high
```

This change fixes BR-1 completely. The duplicated helpers now live in one shared module, `scripts/test/fixtures/queue_reads.py`, with three functions: `committed`, `committed_read` and `slow_writes`. Both test scripts load it and call it through one-line wrappers. Behavior is unchanged. The shared fixture takes the lock with the same non-blocking `flock` call the helper's `try_lock` uses (`scripts/test-lock.py:1306-1311`), and a busy lock still reads as "not yet". Both scripts pass at the head: exit ran 11 tests OK and pause ran 19 tests OK. The issue file changes only status, `actual_hours`, the card hash and the Log, and the two `workshop/plans/` files are review records. There is no change to Spec, Revisions or Done-when, and no conflict resolution to judge.

1. **Strengths**
   - `queue_reads.py:1-6`: the module docstring keeps the reason for the locked reads: retire's write/write/unlink order and the in-place rewrite of the operation record.
   - The whole class was fixed, not just the named site: both test scripts and both shim-path setups (`test-lock-exit.test.py:96-97`, `test-lock-pause.test.py:241`) now go through `slow_writes`.
   - `committed_read` takes the test's own wait function (`await_`) as an argument, so the shared module doesn't depend on unittest.

2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. `queue_reads.py` is loaded with the same five lines of `importlib` code in each script. That is acceptable for script-style tests and matches how both scripts already load `test-lock.py`.
5. **Test coverage:** this is a test-only refactor. Both full scripts pass at the head. The regression tests that call these helpers (`test_fast_cancellation_waits_for_the_whole_retirement`, `test_cancellation_wait_reads_the_record_whole`) are in the passing runs. The fix doesn't change behavior, so no new regression test is needed.
6. **Architecture:**
   - **ARCH-DRY: pass.** I swept the other `control.lock` uses (`test-lock-pause.test.py:108,292`, `test-lock-exit.test.py:279`). They take the lock to block while writing, or check that it is free. That is a different job from the locked-read helper, not a copy of it.
   - **ARCH-PURE: pass.** The fixture's IO is limited to the lock and the file read, and the condition being tested is passed in as a function.
   - **ARCH-PURPOSE: pass.** All sites of the finding's class are fixed.
   - **ARCH-EVOLVE: N/A.** No persisted format changed.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      committed, committed_read and slow_writes moved to scripts/test/fixtures/queue_reads.py; both scripts delegate to it (exit:46-50,97; pause:100-104,241); same flock LOCK_NB semantics as try_lock; exit 11 OK and pause 19 OK at head.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"02ac161e-4e88-448b-848d-437be980c141","repository":"github.com/xianxu/ariadne","issue_id":"000592","recipe":"fix-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/2210dddbfe1a19d7","reviewed_head":"6eb2f26b668f5a19d8dbcf0b15c87ec2d36adfba","slot":"ariadne:5","requested_at":"2026-10-24T15:34:23.110993Z"},"phase":"finished","pid":9839,"started_at":"2026-10-24T15:34:23.231717Z","finished_at":"2026-10-24T15:35:13.012775Z","session_id":"4b50d2ed-2439-44f8-82c6-45bc05458031","outcome":"returned"}] -->
