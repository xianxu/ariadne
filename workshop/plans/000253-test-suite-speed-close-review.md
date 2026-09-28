# Boundary Review — ariadne#253 (whole-issue close)

| field | value |
|-------|-------|
| issue | 253 — Keep the sdlc test suite fast: tiers, shared binary, parallel-safe e2e |
| repo | ariadne |
| issue file | workshop/issues/000253-test-suite-speed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f27d185111d247ad86e33aade2630dfc6c2088b..491ded173bfd6183e25edef59c590dd26a4cff54 |
| command | sdlc close --issue 253 |
| reviewer | claude |
| timestamp | 2026-09-28T13:19:23-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: medium
```

The branch does what the twice-narrowed Plan asks, and every checked Plan item is present in the code. `testfix.PreferRealGit()` is called from a `TestMain` in each package that runs git: cmd/sdlc, gitx, tracker, fleet and activetime. churn only showed up in my search because a test mentions a testfix file path, so it doesn't need one. `scripts/test-shard.py` compiles the test binary once, splits the tests into balanced shards (longest first), and emits `test2json` output. It exits non-zero if any test fails or any process dies. The three fixes the sharded run surfaced are in, and `lockMayBeOurs` is a pure function with a table test. The atlas and AGENTS.local.md explain how to run the suite and why it is sharded. The Log records the timings: 181 s sharded against 1,531 s serial, which meets the "≤ 5 min via `make test`" Done-when. Nothing blocks shipping; the findings below are minor. My confidence is medium because I reviewed the diff statically and did not re-run the suite.

1. **Strengths**
   - `cmd/sdlc/testmain_test.go:90` `lockMayBeOurs` keeps the decision pure and passes in the clock-like dependency (`alive`). `TestLockMayBeOurs` covers five cases: held by this process, held by another slot, holder dead, metadata still being written, and another host.
   - `scripts/test-shard.py`: the timings file has a clear end of life. Each run overwrites it, it is pruned to the tests that still exist, it is written atomically, and it lives in the git common dir, so it goes away with the clone (ARCH-FUNERAL passes).
   - `report()` prints the tail of a shard's output when the shard dies without any failing test. That addresses the "silent exit-1 shard" item in the Plan.
   - `realgit.go` splits the pure `PrependPath`, which has a table test, from the side-effecting `PreferRealGit`, which has a behavioural test that restores PATH.
   - `hangGuard` explains why the bound is loose: it only turns a hang into a failure.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `cmd/sdlc/issuemovedetail.go:38`: the help text says "escape patch" where it means "escape hatch", and the sentence is ungrammatical. This change is an unrelated side-quest, though it is correctly labelled as one in its commit.
   - `scripts/test-shard.py` is committed as 100644 (not executable), but its usage line says `scripts/test-shard.py`. `scripts/test-timing.py` is 100755.
   - `TEST_NAME` only matches `Test` and `Fuzz`, so any future `Example*` in cmd/sdlc would silently never run under `make test`. There are none today.
   - `processAlive` treats EPERM from `kill(pid, 0)` as a dead process. A lock held by another user's live process would therefore count as a leak. This is an edge case.
   - The rule "a new git-driving package needs the same TestMain" exists only in the atlas. A source guard could enforce it; this can wait for #261 or #262.

5. **Test coverage notes:** the pure pieces (`PrependPath`, `lockMayBeOurs`, `snapshotDiff`) have unit tests. The two Python scripts have no tests, which is acceptable for developer tooling that each run exercises end to end. The regression fixes (`SetArgs([]string{})`, `hangGuard`) are covered by the tests they fix.

6. **Architecture notes**
   - **ARCH-DRY: pass.** Go needs one `TestMain` per package, so the five copies are the minimum, and they share `testfix.PreferRealGit`.
   - **ARCH-PURE: pass.** `lockMayBeOurs` and `PrependPath` are pure, and the IO around them is kept thin.
   - **ARCH-PURPOSE: pass.** The scope was narrowed explicitly in `## Revisions`, and the deferred parts are tracked as #261 and #262.
   - **ARCH-MOCK: pass (explicitly deferred).** Tests still run real git; the stateful fake is #261.
   - **ARCH-CONSTRAINTS: pass.** Wall times are measured and logged. The shard count is bounded by the core count.
   - **ARCH-SECURE: pass.** The timings file is parsed defensively (`OSError`/`ValueError` leads to an empty timings map), and there are no credentials involved.
   - **ARCH-ORDER: pass.** The runner keeps no state between events. Shard processes are bounded by the pool's lifetime, and the temp dir is cleaned up on exit.
   - **ARCH-FUNERAL: pass.** The timings file is bounded as described above, and the temp dir is scoped to the run.

7. **Plan revision recommendations:** none. The two existing Revisions entries match the code.

```findings
findings:
  - id: new
    severity: Minor
    family: help-text-accuracy
    title: |
      move-detail help text says "escape patch" (means escape hatch) and the sentence is ungrammatical
    detail: |
      cmd/sdlc/issuemovedetail.go:38, side-quest commit f3e7c907.
  - id: new
    severity: Minor
    family: script-invocation-contract
    title: |
      scripts/test-shard.py committed 100644 though its usage line invokes it directly
  - id: new
    severity: Minor
    family: shard-test-selection-completeness
    title: |
      test-shard TEST_NAME excludes Example functions, so future cmd/sdlc examples would silently not run under make test
  - id: new
    severity: Minor
    family: guidance-without-enforcement
    title: |
      "new git-driving package needs PreferRealGit TestMain" is enforced only by atlas prose, not a source guard
```

---

## Re-review — 2026-09-28T13:21:25-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 253 — Keep the sdlc test suite fast: tiers, shared binary, parallel-safe e2e |
| repo | ariadne |
| issue file | workshop/issues/000253-test-suite-speed.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f27d185111d247ad86e33aade2630dfc6c2088b..3f11b26fe786c88616d165056e42e9813f15172c |
| command | sdlc close --issue 253 |
| reviewer | claude |
| timestamp | 2026-09-28T13:21:25-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The issue's narrowed Done-when is met. It has a timing script with its baseline in the Log, and `make test` runs the full suite in 181 s against 1,531 s before. The target was 5 minutes. The guidance line and the atlas section are both in place. Real git on PATH is wired through one helper, `testfix.PreferRealGit`, and a source guard now enforces it. The three fixes that came out of the sharded run are pinned by tests: `lockMayBeOurs` is pure and table-tested, the `SetArgs` fix is a one-liner, and there is a shared hang guard. The new guard and lock tests pass on HEAD (`go test -run 'TestGitDrivingPackagesPreferRealGit|TestLockMayBeOurs|TestSnapshotDiff'` → ok). Two prior Minors are still open: the help-text typo, and a runner name-filter change with no test covering it. Nothing blocks SHIP.

**1. Strengths**
- `lockMayBeOurs` (`cmd/sdlc/testmain_test.go:90`) is a pure decision. Process liveness, PID and host are injected, and the table covers all five cases: self, another live slot, dead holder, metadata still being written, and another host. This is ARCH-PURE applied well.
- `PrependPath` is pure and table-tested (`internal/testfix/realgit_test.go:11`). `PreferRealGit` is a no-op when git or its exec-path binary is missing, and the second test checks the result end to end with `exec.LookPath`.
- `realgit_guard_test.go` turns the atlas rule into a check (BR-4). A grep of each package shows that removing the `TestMain` from fleet, tracker, gitx or activetime would make it fail: each has spawn ≥ 3 and prefer = 1.
- The runner keeps its timings file bounded. It holds only tests that still exist, is written atomically, and merges results even from failing runs (ARCH-FUNERAL is documented in the script header).
- The runner reports a process that dies without any failing test and prints the tail of its output. This closes the "silent exit-1 shard" gap.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- BR-1 is not fixed. `cmd/sdlc/issuemovedetail.go:38` still reads "Consult operator before move, this is an escape patch".
- BR-3 is not fixed per the regression-evidence rule. The regex now accepts `Example`, but nothing tests it. The tree also has no `Example` functions, so no fixture reaches the change. `scripts/test/` already hosts shell tests where a small name-filter check could live.
- The guard's regex (`realgit_guard_test.go:15`) detects git use only through `exec.Command(... "git"` or `testfix.(Git|Capture|Repo)`. A package whose tests reach git only through `gitx` helpers would not be caught. No such package exists today (`internal/project` only mentions `gitx.IsBrainRepo` in comments), so I'm noting this rather than raising it as a finding.

**5. Test coverage notes**
- The new pure logic (`PrependPath`, `lockMayBeOurs`, `snapshotDiff`) is covered. The Python runner has no automated tests: `lpt`, the name filter and `report` are all unpinned. It was validated by the logged runs instead.

**6. Architectural notes**
- **ARCH-DRY: pass.** One helper is called from five thin `TestMain`s. `test-timing.py`'s `GIT_RE` and the guard's `spawnsGit` duplicate a prefix. This is tolerable because they live in different languages and serve different purposes.
- **ARCH-PURE: pass.** The decisions are pure and the IO is at the edges (`readSnapshot`, `PreferRealGit`).
- **ARCH-PURPOSE: pass.** Scope was narrowed explicitly in `## Revisions`. The deferred items went to #261 and #262 as separable work; they are not the point of this issue.
- **ARCH-MOCK: pass.** The stateful git fake is #261, deferred at the operator's direction and recorded in the atlas.
- **ARCH-CONSTRAINTS: pass.** The measured envelope is logged. The hang guard was widened with a stated reason, and the per-process timeout is explicit. The oversubscription (N shards plus a `go test` of the other packages) is accepted for a dev-machine runner.
- **ARCH-SECURE: pass.** The timings JSON is parsed defensively (`OSError`/`ValueError` fall back to `{}`), and lock metadata errors map to "not ours".
- **ARCH-ORDER: pass.** The guard reads a before and after snapshot, and the race with another slot's live lock is handled explicitly. There is one leftover false positive: another slot's verb crashing during the run would look like a leak. That is rare and would show up as a visible failure.
- **ARCH-FUNERAL: pass.** The timings file is overwritten each run, bounded to current tests, and deleted with the clone. Temporary shard output lives in a `TemporaryDirectory`.

**7. Plan revision recommendations:** none. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      cmd/sdlc/issuemovedetail.go:38 still reads "Consult operator before move, this is an escape patch" at HEAD 3f11b26f.
  - id: BR-2
    disposition: addressed
    note: |
      git ls-tree 3f11b26f shows scripts/test-shard.py (and test-timing.py) as 100755.
  - id: BR-3
    disposition: not-addressed
    note: |
      TEST_NAME now admits Example, but no test pins it and no Example exists in cmd/sdlc, so no fixture reaches the change; scripts/test/ could host the check.
  - id: BR-4
    disposition: addressed
    note: |
      cmd/sdlc/realgit_guard_test.go enforces it; per-dir grep shows fleet/tracker/gitx/activetime would fail without their TestMain call; passes at HEAD.
```

---

## Re-review — 2026-09-28T15:29:02-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 253 — Keep the sdlc test suite fast: tiers, shared binary, parallel-safe e2e |
| repo | ariadne |
| issue file | workshop/issues/000253-test-suite-speed.md |
| boundary | whole-issue close |
| milestone | — |
| window | e8415cb0b218ae713b4f9ae06e0018ef727ddd24..dd78d856fd93bacab04c607e44cc78ab94a05eea |
| command | sdlc close --issue 253 |
| reviewer | claude |
| timestamp | 2026-09-28T15:29:02-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The two speed-ups in the narrowed scope are delivered and they work. `testfix.PreferRealGit()` is called from a `TestMain` in every package that drives git (cmd/sdlc, gitx, tracker, fleet, activetime), and a source guard now enforces it (`TestGitDrivingPackagesPreferRealGit`). The sharded runner is in `scripts/test-shard.py` behind `make test`. The two fixes the sharded run needed are in (`SetArgs([]string{})`, and `lockMayBeOurs` for the shared `sdlc.lock`). The atlas and `AGENTS.local.md` both document how to run the suite. At HEAD dd78d856 the targeted tests pass: `TestGitDrivingPackagesPreferRealGit`, `TestLockMayBeOurs`, `TestSnapshotDiff` and the `testfix` package. Nothing blocks. Two prior Minors are still open. BR-1 was never touched: the `move-detail` help text still says "escape patch". BR-3 widened the regex, but the rule behind that finding still has no check.

**1. Strengths**
- `cmd/sdlc/testmain_test.go:90` `lockMayBeOurs` is a pure decision with the process-liveness check injected. The table in `testmain_guard_test.go:81` covers live, dead, mid-initialisation and other-host holders (ARCH-PURE).
- `internal/testfix/realgit.go` is a no-op when git or its exec-path binary is missing. `PrependPath` is idempotent and has its own table test.
- `realgit_guard_test.go` turns the "new git-driving package needs a `TestMain`" rule from atlas prose into a failing test. That closes BR-4 properly.
- `test-shard.py` bounds the timings file: it is overwritten each run, pruned to the tests that still exist, and lives in the git common dir, so it goes away with the clone. Its comment says all of this (ARCH-FUNERAL).
- `closereview_test.go:184` `hangGuard` fixes the root cause of the load-sensitive 5 s waits: the bound exists only to catch a hang, so it is now generous rather than tight.

**2. Critical:** none.

**3. Important:** none.

**4. Minor**
- BR-1 (still open): `cmd/sdlc/issuemovedetail.go:38` still reads "Consult operator before move, this is an escape patch, not for regular use." It should say "hatch", and the run-on sentence needs splitting, e.g. "Consult the operator before moving: this is an escape hatch, not for regular use."
- BR-3 (still open, 2nd in family `shard-test-selection-completeness`): the regex now includes `Example`, which fixes that one case. The rule it belongs to is still unchecked: every name `-test.list` prints (except `Benchmark*`) must run in exactly one shard. The runner could compute `set(names) - {tests seen with a pass/fail/skip event}` and exit non-zero when that set is non-empty. That catches this whole family (a regex gap, a sharding bug, a name dropped from a `-test.run` pattern), not just the `Example` case. Right now nothing fails if a selected test silently never runs.
- `scripts/test-shard.py:151`: `write_atomic` uses a fixed `sdlc-test-timings.tmp` in the shared common dir. Two slots running `make test` at once can race on that temp file. A `tempfile.NamedTemporaryFile(dir=...)` would avoid it. This is low impact, since the file only feeds shard balancing.

**5. Test coverage notes**
- The guard's regex only sees direct git spawns in test source: `exec.Command…"git"` and `testfix.Git/Capture/Repo`. A package whose tests reach git only through production code (for example by calling into gitx) would not be flagged. I found no such package today, so this is not a finding; it's a known blind spot.
- `test-shard.py` and `test-timing.py` have no automated tests. The runner's selection and failure reporting were checked only by live runs, which is why BR-3 is still open.

**6. Architecture pass**
- **ARCH-DRY: pass.** The five per-package `TestMain` files are near-identical, but that is how Go requires it.
- **ARCH-PURE: pass.** `lockMayBeOurs` and `PrependPath` are pure.
- **ARCH-PURPOSE: pass** against the twice-revised Done-when. The `-short` tier, the budget guard and the shared binary build were deferred explicitly to #261, with the reasons logged.
- **ARCH-MOCK: pass.** The stateful git fake is scoped to #261, and this change adds no new external call.
- **ARCH-CONSTRAINTS: pass.** Shard count defaults to the core count, and the per-process timeout is explicit (30m).
- **ARCH-SECURE: pass.** A timings file that fails to parse degrades to round-robin dealing. It is trusted only for balancing.
- **ARCH-ORDER: pass.** The only ordering concern is lock ownership across slots, and it is handled with an explicit pure decision.
- **ARCH-FUNERAL: pass.** The timings file is bounded and overwritten, and the temp directory is cleaned up by its context manager.

**7. Plan revision recommendations:** none. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      issuemovedetail.go:38 at dd78d856 still reads "this is an escape patch" in a comma-spliced sentence; unchanged since f3e7c907.
  - id: BR-3
    disposition: not-addressed
    note: |
      Regex widened to Example (3f11b26f), but no test and no class check; add a runner coverage check that every -test.list name except Benchmark got a terminal event, else exit non-zero.
findings:
  - id: new
    severity: Minor
    family: shared-state-write-race
    title: |
      test-shard.py write_atomic uses a fixed .tmp path in the shared git common dir, so concurrent make test runs from two slots can race
    detail: |
      scripts/test-shard.py:151 always writes sdlc-test-timings.tmp. Use tempfile.NamedTemporaryFile(dir=path.parent, delete=False) and then os.replace.
```
