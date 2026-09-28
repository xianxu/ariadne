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
