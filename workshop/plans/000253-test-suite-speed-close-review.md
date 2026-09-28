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
