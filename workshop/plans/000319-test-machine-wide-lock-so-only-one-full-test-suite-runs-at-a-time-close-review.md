# Boundary Review — ariadne#319 (whole-issue close)

| field | value |
|-------|-------|
| issue | 319 — test: machine-wide lock so only one full test suite runs at a time |
| repo | ariadne |
| issue file | workshop/issues/000319-test-machine-wide-lock-so-only-one-full-test-suite-runs-at-a-time.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5d364c69e8ffda6f5416e3f3dbd236625f299964..ae84fbb6c298c2caaa5d28e4b77f307e41aeef9d |
| command | sdlc close --issue 319 |
| reviewer | claude |
| timestamp | 2026-10-10T11:36:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing: I'm checking how pair's and parley.nvim's test targets reach the lock, and how close.go runs tests.

My scratch-directory setup failed, so I'm re-running the malformed-lock-file probe inside a directory in the repo.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The design works. Taking the lock while make parses its makefiles, through `Makefile.workflow`, means it is held before any prerequisite runs. That reaches pair's prerequisite-heavy `test`, and parley.nvim's `test`, whose sub-makes run under the outer lock. All three repos symlink `Makefile.workflow` to ariadne's, so a single wiring covers them. The lock is released by the kernel, so killing the holder frees it, and a nested `make test` passes straight through. I ran `scripts/test-lock.test.sh` myself and all 12 checks passed. One cheap fix is worth making before close. When the lock is contended, a waiter crashes if the record in the lock file is valid JSON without the expected keys. I reproduced this: a `{"pid":1}` record gives `KeyError: 'repo'`. `make test` then fails with "helper failed" instead of waiting with the holder shown. That breaks the Spec's "waiting is visible" contract whenever an older or newer version, or a hand edit, has written the record.

**1. Strengths**
- `Makefile.workflow:13-23`: taking the lock at parse time avoids GNU make 3.81 running the target's own prerequisites first. The comment explains why, and every woven repo picks it up without editing its own makefile.
- `scripts/test-lock.py:105-111`: the forked watcher inherits the lock file descriptor, closes its standard input and output so make's `$(shell)` returns, and watches make's pid for exit. Any exit or crash releases the lock with no stale-lock cleanup. On macOS, which has no `flock(1)`, this is the right primitive.
- Waits are reported as waits: `WF_TEST_LOCK_TIMEOUT` makes make fail with "full-suite lock not taken (timeout)" and "a lock wait, not a test failure". This meets the Spec's gate requirement.
- `scripts/test-lock.test.sh` gives each case its own lock file, so it never touches the real machine-wide lock. It unsets an inherited `WF_TEST_LOCK_HELD_BY`, so it still behaves correctly when run inside a locked `make test`. It also checks the make wiring end to end with `make -n`.

**2. Critical:** none.

**3. Important**
- `scripts/test-lock.py:42-52` (ARCH-SECURE): `holder()` handles invalid JSON (`ValueError`) but not the wrong shape. `describe()` indexes `rec['repo']` and the rest directly, and `rec.get` fails on anything that isn't a dict. Fix: in `holder()`, return the record only if it is a dict, otherwise `None`. In `describe()`, use `rec.get(...)` with a `?` default. Add a test case that writes a valid-JSON record missing the keys while the lock is held.

**4. Minor**
- `Makefile.workflow:14`: the `$(wildcard scripts/test-lock.py)` guard skips the lock without saying so in repos that haven't run `weave refresh`. `$(WF_WORKFLOW_SOURCE_DIR)scripts/test-lock.py` always resolves to ariadne's copy, so it would drop the dependency on refresh and the need for the manifest row. Otherwise, print a warning when the helper is missing.
- `scripts/test-lock.py:105`: the watcher stays in make's process group and keeps the default signal handling. A test that signals its own process group (`kill 0`) would kill the watcher and release the lock mid-suite. The watcher could ignore SIGINT and SIGHUP and rely on watching make's exit.
- `scripts/test-lock.py:63-68`: on Linux, where kqueue is unavailable, the polling fallback catches only `ProcessLookupError`. A reused pid owned by another user raises `PermissionError`, which crashes the watcher and releases the lock early.
- `scripts/test-lock.py:91`: there is a window of tens of milliseconds between taking the lock and writing the record. A waiter in that window reads the previous run's record. This is harmless as written, but worth a comment.

**5. Test coverage**
- Covered: two runs serialize with the holder shown, handover on holder exit, release on `kill -9`, the same make and a nested make are re-entrant, timeout, the off switch, and the make wiring for `test` versus `help`.
- Not covered: a malformed record (see Important).
- The Done-when criterion "at least … pair's full-test targets use it after `weave refresh`" was simulated, not run against a real refreshed pair slot. That is reasonable before ariadne:0 is updated, but the close evidence should say so.
- The tests depend on timing (≤2s handover, 1.5s sleeps). Acceptable for a standalone script.

**6. Architecture**
- **ARCH-DRY: pass.** One helper and one wiring point.
- **ARCH-PURE: pass.** It is a small IO script and there is no business logic to separate out.
- **ARCH-PURPOSE: pass.** ariadne, pair and parley.nvim all reach the lock through the shared `Makefile.workflow`.
- **ARCH-MOCK: pass.** There is no external service; the `WF_TEST_LOCK_FILE` override is the test seam.
- **ARCH-CONSTRAINTS: pass.** Polling is every 1s and reports every 30s. The unbounded default wait is a documented choice, with the timeout as the bound.
- **ARCH-SECURE: flag.** The record in the shared state file is trusted to have the right shape (see Important).
- **ARCH-ORDER: pass.** The kernel owns the lock's free, held and waiting states. The re-entrancy check reads a record that can briefly be stale, and that is harmless.
- **ARCH-FUNERAL: pass.** There is one lock file per user, rewritten on every run, and the watcher process ends when make exits.

**7. Plan revisions:** none. The plan matches the code.

```findings
findings:
  - id: new
    severity: Important
    family: untrusted-persisted-record-shape
    title: |
      A waiter crashes (KeyError or AttributeError) on a valid-JSON lock record with an unexpected shape
    detail: |
      holder() handles only ValueError; describe() indexes rec['repo'] and the rest directly. Reproduced: a held lock with record {"pid":1} makes the waiter traceback, and make test fails with "helper failed" instead of waiting. Validate in holder() (a dict, else None), use rec.get with defaults in describe(), and add a test case.
  - id: new
    severity: Minor
    family: silent-guard-bypass
    title: |
      The wildcard guard on scripts/test-lock.py skips the lock without a word in repos not yet refreshed
    detail: |
      $(WF_WORKFLOW_SOURCE_DIR)scripts/test-lock.py always resolves to ariadne's helper and removes the dependency on weave refresh; otherwise warn when the helper is missing.
  - id: new
    severity: Minor
    family: watcher-lifetime-coupling
    title: |
      The watcher keeps default signal handling in make's process group, so a group signal releases the lock mid-suite
    detail: |
      Ignore SIGINT and SIGHUP in the watcher and rely on watching make's exit; separately, the Linux polling fallback should treat PermissionError as the pid still being alive.
```

---

## Re-review — 2026-10-10T11:38:42-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 319 — test: machine-wide lock so only one full test suite runs at a time |
| repo | ariadne |
| issue file | workshop/issues/000319-test-machine-wide-lock-so-only-one-full-test-suite-runs-at-a-time.md |
| boundary | whole-issue close |
| milestone | — |
| window | 5d364c69e8ffda6f5416e3f3dbd236625f299964..67d51fbebd8b7fc37eb7592647326fc68f6c8391 |
| command | sdlc close --issue 319 |
| reviewer | claude |
| timestamp | 2026-10-10T11:38:42-07:00 |
| verdict | SHIP |

## Review

Environment restricts temp dirs; use a repo-local scratch dir.

```verdict
verdict: SHIP
confidence: high
```

All three prior findings are fixed, and the fixes are confirmed. `holder()` now validates the record's shape: it must be a dict with all four keys, and anything else reads as unidentified. I reproduced BR-1's exact `{"pid":1}` case against HEAD. The waiter now prints "unidentified holder" and returns `timeout` instead of crashing with a traceback. The new `[1, 2]` regression case would fail without the fix, because the old code called `rec.get` on a list and raised AttributeError. The helper path now comes from `$(WF_WORKFLOW_SOURCE_DIR)`, and a missing helper produces a `$(warning …)` instead of a silent skip. The watcher ignores SIGINT and SIGHUP, which a test covers, and the poll fallback treats PermissionError as alive. `scripts/test-lock.test.sh` passes 15/15 locally. Nothing blocks SHIP.

1. **Strengths**
   - Taking the lock at parse time (`Makefile.workflow:15-28`) runs it before any prerequisite, with one wiring point for every woven repo (ARCH-DRY).
   - The lock is released by the kernel through a forked watcher that holds the inherited descriptor, so there is never a stale lock. The kill -9 test proves this.
   - Lock waits stay separate from test failures: the `timeout` outcome surfaces as "full-suite lock not taken (timeout)" with a "lock wait, not a test failure" line on stderr.
   - The tests use a private `WF_TEST_LOCK_FILE` and never touch the real machine-wide lock (ARCH-SECURE). They drive the real Makefile wiring through `make -n`.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `describe(None)` says "record not yet written" even when the file holds a foreign or malformed record. "unidentified holder" alone would be accurate. This is cosmetic only.
   - A malformed `WF_TEST_LOCK_TIMEOUT` (e.g. `5m`) makes the helper fail with a traceback, and make then reports "helper failed". The failure is visible, but the message is crude.

5. **Test coverage notes:** The tests cover:
   - two runs serializing, with the waiter naming the holder;
   - prompt handover when the holder exits;
   - the watcher surviving SIGINT and SIGHUP;
   - a kill -9 of the watcher releasing the lock;
   - re-entry for the same make and for a nested make;
   - timeout, the `off` escape hatch, and a foreign record;
   - make wiring both ways (`test` locks, `help` doesn't).

   The missing-keys dict case (`{"pid":1}`) has no test of its own, but the same validation path covers it, and I verified it by hand.

6. **Architecture**
   - **ARCH-DRY:** pass. There is one helper and one wiring point.
   - **ARCH-PURE:** pass. This is a thin IO script with no business logic to separate out.
   - **ARCH-PURPOSE:** pass. Every Done-when item is delivered. Pair picks the change up through its `Makefile.workflow` symlink, and the issue Log records that this was simulated.
   - **ARCH-MOCK:** pass. The only external dependency is the kernel's flock, and the tests exercise the real thing on private files.
   - **ARCH-CONSTRAINTS:** pass. The wait is unbounded by default but optionally bounded, and it reports every 30s. The atlas tells agents to run the suite in the background.
   - **ARCH-SECURE:** pass after BR-1. The persisted record is validated before use, and the tests can't touch real state.
   - **ARCH-ORDER:** pass. The only state is the kernel lock plus a single holder record, and the reentrant, timeout and handover orderings are tested.
   - **ARCH-FUNERAL:** pass. There is one per-user file, rewritten on each run, so it never grows.

7. **Plan revisions:** The Plan's first checkbox still mentions a "manifest symlink row" and says the helper is "symlinked". Round 1 replaced that with resolving the helper beside `Makefile.workflow`'s real path. The Log records the change, but a one-line `## Revisions` entry would keep the Plan in step with the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      holder() requires a dict with all four keys; the {"pid":1} repro now reads as unidentified and returns timeout. The [1, 2] regression case would raise AttributeError without the fix.
  - id: BR-2
    disposition: addressed
    note: |
      The helper resolves via $(WF_WORKFLOW_SOURCE_DIR), and a missing helper emits $(warning) that the run is unlocked (Makefile.workflow:15-18).
  - id: BR-3
    disposition: addressed
    note: |
      The watcher ignores SIGINT and SIGHUP (tested by "watcher survives SIGINT and SIGHUP"), and the poll fallback treats PermissionError as alive.
findings:
  - id: new
    severity: Minor
    family: untrusted-persisted-record-shape
    title: |
      describe(None) says "record not yet written" for a foreign or malformed record too
    detail: |
      This is message wording only; the rule (validate the shape before use) is already applied in holder(). "unidentified holder" alone would be accurate.
```
