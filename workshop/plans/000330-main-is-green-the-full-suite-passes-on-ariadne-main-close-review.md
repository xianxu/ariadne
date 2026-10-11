# Boundary Review — ariadne#330 (whole-issue close)

| field | value |
|-------|-------|
| issue | 330 — Main is green: the full suite passes on ariadne main |
| repo | ariadne |
| issue file | workshop/issues/000330-main-is-green-the-full-suite-passes-on-ariadne-main.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e80818d6213ac3e9f7c55306a9ccce111cf7cd4..2fe0e43e59fe05c8796b9ae768db230a354994ab |
| command | sdlc close --issue 330 |
| reviewer | claude |
| timestamp | 2026-10-11T00:47:49-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Still reviewing — checking testfix's non-test importers and other liveness probes.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The work does what the Spec asks. The two root-cause fixes are correct and each has a test that fails without it. The first is `processAlive` treating EPERM as alive. The second is the testgate nested-helper deadlock. The `ps` probes are now one exec-free helper with a wait ceiling. `make test` now covers the whole module. The CI decision is recorded and filed as #331, which exists. Both quarantines name #325, which exists. One cheap gap remains before shipping: #325's own file says nothing about the quarantine, so the step that ends it ("#325's fix deletes both calls") is recorded only in #330's Log and a TL message. Everything else is Minor.

Verified locally at HEAD 2fe0e43e:
- `testfix`, `testgate`, `processgroup` and `judge` pass.
- `TestProcessAlive*` passes.
- `testfix` vets for `GOOS=linux` and builds for `GOOS=windows`.
- No non-test file imports `testfix`, so the `golang.org/x/sys` dependency is test-only, as the Log claims.

1. **Strengths**
   - `cmd/sdlc/repolock.go:172-177`: this is a real lock-safety fix for shared `.git/sdlc.lock`, with the reason documented. The atlas lock section is updated to match.
   - `pkg/testgate/testgate.go:121-123` and `testgate_test.go:82`: the fix is small and the test reproduces the actual hang, a nested re-exec under a single slot. `runHelper` correctly clears the new variable so the helper doesn't inherit it.
   - `testfix/process.go`: a condition wait with a ceiling instead of a one-shot probe. It also catches the async group-SIGKILL race that the slower `ps` call used to hide. Two judge checks that passed vacuously now really check.
   - `scripts/test-shard.py:86,119`: widening the run to `./...` closes the coverage gap, and the Log measures what it costs.
   - The lessons entry states the general rule ("an error ignored from an environment probe is a check that silently passes"), not just this one case.

2. **Critical:** none.

3. **Important**
   - The end of each quarantine is recorded only in #330. `workshop/issues/000325-*.md` doesn't mention `testfix.Quarantine` or the two test names. Whoever fixes #325 has no durable prompt to delete the two calls (`planningreview_test.go:122`, `review_concurrency_unix_test.go:81`), so the skips can outlive the fix (ARCH-FUNERAL). Fix: add a Done-when line to #325 ("delete both `testfix.Quarantine(t, "#325", …)` calls; `grep -rn 'testfix.Quarantine(' cmd` is empty for #325"). The owner or TL can publish it with `sdlc issue publish`. If that can't happen in this round, say so in `--verified`.

4. **Minor**
   - `review_concurrency_unix_test.go:81`: the quarantine skips all 14 subtests of `TestReviewConcurrencySchedules`, but only `milestone-close/branch` failed, once in two runs. That drops coverage of cancel, relock and the other cases. Quarantining only the subtests that failed would keep that coverage.
   - `repolock.go:171`: the comment says "Only ESRCH is gone", but the code treats every error other than EPERM as dead. `weave/internal/staging/stage.go:85` uses the opposite reading (`!ESRCH` means alive). Either align the comment, or share one liveness predicate (ARCH-DRY).
   - `WF_TEST_GATE_HELD_BY` is inherited by every descendant, not only re-exec'd helpers. A gated test that runs `go test` on a fixture module would run that child ungated, with no parallel cap. This is intended per the comment ("covers its process tree"), but the value is never checked against a live ancestor.
   - `process_linux.go:20`: if the stat line has no `)`, `LastIndexByte` returns -1 and the whole line gets parsed. Harmless, but not explicit.
   - `atlas/workflow/sdlc-binary.md:1430`: "About 4 minutes" now sits next to the measured 7.5–8 minutes. The stale figure should just be replaced. The same "about 4 minutes" appears in the CLAUDE.md overlay.

5. **Test coverage**
   - `TestAwaitExit` covers a live process, a zombie and a reaped process.
   - `TestNestedHelperStandsAside` fails before the fix and passes after.
   - `TestProcessAliveCountsUnsignalableProcesses` uses pid 1. As root (for example in the CI container planned in #331) pid 1 answers nil, so the test still passes but no longer exercises EPERM. Worth a note in #331.
   - The Linux `processRunning` path compiles but has never run. #331 is where it will first run.

6. **Architecture**
   - ARCH-DRY: flag (Minor). The kill-0 liveness rule is now in two places with different readings: repolock and weave staging.
   - ARCH-PURE: pass. The helpers are a thin IO seam, and `Quarantine` is trivial.
   - ARCH-PURPOSE: pass. The whole module is now in the suite, and the shadow sweep of `ps` probes found all three sites.
   - ARCH-MOCK: pass. Not applicable; the change is test-infra kernel queries.
   - ARCH-CONSTRAINTS: pass. AwaitExit polls every 5ms against a 2s ceiling, and the widened run's wall cost is measured.
   - ARCH-SECURE: pass. The lock now fails safe: unknown means alive.
   - ARCH-ORDER: pass. The group-SIGKILL ordering is handled by waiting, not by sampling once.
   - ARCH-FUNERAL: flag (Important, above). The quarantine's end is not recorded on the owning issue.

7. **Plan revisions:** none. The Plan matches the code.

```findings
findings:
  - id: new
    severity: Important
    family: removal-path-unrecorded-at-owner
    title: |
      Quarantine end-condition lives only in #330's Log; owning issue #325 does not mention deleting the two Quarantine calls
    detail: |
      workshop/issues/000325 has no quarantine/test-name reference, so #325's fix can land with the skips still in place. Add a Done-when line to #325 (delete both testfix.Quarantine(t, "#325", ...) calls) via sdlc issue publish, or note why not in --verified.
  - id: new
    severity: Minor
    family: quarantine-scope-wider-than-failure
    title: |
      TestReviewConcurrencySchedules quarantined whole (14 subtests) though only milestone-close/branch failed, once in two runs
    detail: |
      Quarantining inside the failing subtest keeps cancel/relock/etc. coverage while #325 is open.
  - id: new
    severity: Minor
    family: single-liveness-predicate
    title: |
      kill-0 liveness rule implemented twice with different readings (repolock: non-EPERM error = dead; weave staging: non-ESRCH = alive)
    detail: |
      repolock.go comment says "only ESRCH is gone" but code returns dead for any non-EPERM error; align the code or comment, or share one helper.
  - id: new
    severity: Minor
    family: stale-doc-figure
    title: |
      atlas sdlc-binary.md keeps "About 4 minutes" beside the measured 7.5-8 minutes; CLAUDE.md overlay repeats the stale figure
```

---

## Re-review — 2026-10-11T00:51:07-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 330 — Main is green: the full suite passes on ariadne main |
| repo | ariadne |
| issue file | workshop/issues/000330-main-is-green-the-full-suite-passes-on-ariadne-main.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3e80818d6213ac3e9f7c55306a9ccce111cf7cd4..86401ddf688ac0b18a2437972d81e411ec6658be |
| command | sdlc close --issue 330 |
| reviewer | claude |
| timestamp | 2026-10-11T00:51:07-07:00 |
| verdict | SHIP |

## Review

Still reviewing: all four round-1 fixes check out against the code, so I'm writing up the verdict.

```verdict
verdict: SHIP
confidence: high
```

All four round-1 findings are fixed, and I checked each one against the pinned diff (`86401ddf`), not the commit message. #325's Done when now names both quarantine calls and gives a grep that must come back empty. The `TestReviewConcurrencySchedules` quarantine now sits only on the 20s start-deadline failure path, so the other 13 subtests keep running and failing normally. `processAlive` now uses the same rule as weave staging: only ESRCH means the process is gone. The timing figures are fixed in `AGENTS.local.md`, which is the source; `CLAUDE.md` is generated from it and gitignored (`.gitignore:56`), so its stale working-tree copy will update on the next weave refresh. I found nothing new that blocks the close.

1. **Strengths**
   - `cmd/sdlc/review_concurrency_unix_test.go:139-141`: the skip runs only after the subtest has failed to start, and only after `cancel()`/`release()` and the 7s stop check. Cleanup and the stop assertion still run. With `WF_TEST_QUARANTINED=run`, the test falls through to `t.Fatal`, so the escape hatch stays honest.
   - `cmd/sdlc/repolock.go:176` and `cmd/weave/internal/staging/stage.go:85` now use the same `!ESRCH` reading. `TestProcessAliveCountsUnsignalableProcesses` (pid 1 → EPERM) pins the lock-safety bug.
   - The #325 Done when (`workshop/issues/000325…:26`) gives a machine-checkable end condition (`grep -rn 'Quarantine(t, "#325"' cmd` is empty). The quarantine's removal is now tracked in #325 itself, the issue that owns the fix.
   - The atlas text at `sdlc-binary.md:1446-1448` adds the "place it at the failing step" rule. That covers the `quarantine-scope-wider-than-failure` family as a rule, not just this one instance.

2. **Critical:** none.

3. **Important:** none.

4. **Minor:** none new.

5. **Test coverage:** the EPERM path is covered by the pid-1 test. The narrowed quarantine is reachable only through the start-timeout branch. That's correct, and it isn't worth a test of its own.

6. **Architecture**
   - ARCH-DRY: pass. There are still two `kill(pid, 0)` call sites (repolock and weave staging), but they now read the result the same way. Sharing a helper across separate cmd modules would be optional.
   - ARCH-PURE: pass.
   - ARCH-PURPOSE: pass. Every quarantined test has an owner with an end condition.
   - ARCH-MOCK: pass. Nothing external was added in this round.
   - ARCH-CONSTRAINTS: pass. The timings are measured.
   - ARCH-SECURE: pass.
   - ARCH-ORDER: pass. No state is held between events.
   - ARCH-FUNERAL: pass. Each quarantine call is removed by #325's Done when.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      000325 Done when gains a line requiring both testfix.Quarantine calls deleted, with a grep check; committed in 86401ddf.
  - id: BR-2
    disposition: addressed
    note: |
      Quarantine moved inside the 20s start-timeout branch (review_concurrency_unix_test.go:140) after cancel/release; other subtests unskipped.
  - id: BR-3
    disposition: addressed
    note: |
      repolock.go:176 now returns !errors.Is(kill0, ESRCH), matching comment and weave staging; pid-1 EPERM test pins it.
  - id: BR-4
    disposition: addressed
    note: |
      atlas sdlc-binary.md and AGENTS.local.md say about 8 minutes; CLAUDE.md is a gitignored generated overlay of AGENTS.local.md.
```
