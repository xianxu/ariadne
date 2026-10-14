# Boundary Review — ariadne#361 (whole-issue close)

| field | value |
|-------|-------|
| issue | 361 — weave: TestCompileCancellationStopsLateGeneratorChild flakes under load (generation stage survived retirement) |
| repo | ariadne |
| issue file | workshop/issues/000361-weave-testcompilecancellationstopslategeneratorchild-flakes-under-load-generation-stage-survived-retirement.md |
| boundary | whole-issue close |
| milestone | — |
| window | 41c30e804a9c96aab1628f443a97c66476e5afb3..1f174297427fced9cc9138068030fd1a315bc80a |
| command | sdlc close --issue 361 |
| reviewer | claude |
| isolation | fenced |
| timestamp | 2026-10-13T22:02:29-07:00 |
| verdict | SHIP |

## Review

Now revert the fix in the procgroup helper (drop EPERM) and confirm the regressions go red.

I'm checking whether a separate failure in the original weave test comes from this change or from the review environment.

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** The fix targets the right cause and covers every place the code signals a process group. On macOS, `killpg` returns EPERM when all of a group's members are zombies or already exiting. Before the fix, `finishProducers` treated that as a hard error and returned without waiting on the lease. The stage could then outlive the run while a member was still closing fd 3. The new `pkg/procgroup.Kill` treats both ESRCH and EPERM as `os.ErrProcessDone`. `finishProducers` now always waits on the lease, which is the only real proof that the producer let go. I put each change back the old way in the disposable checkout and confirmed that all three new regression tests fail without it ("operation not permitted") and pass with it. The original test passed 20/20 with `TMPDIR=/tmp`. Inside this checkout it failed every time at the "generator did not signal" step, before any cancellation happens. That is consistent with the review sandbox setting `TMPDIR` inside the repo, so I don't attribute it to this diff. Nothing blocks shipping. The three findings below are all minor wording and plan-accuracy items.

1. **Strengths**
   - The investigation corrected the issue's own guess: the log shows the 1s settle was never the cause, EPERM was. The kernel behaviour is measured and the evidence is recorded in the issue.
   - The fix covers every place that sends a signal to a process group: weave's `cmd.Cancel` (`cmd/weave/internal/weavefs/runner.go:68`), `finishProducers` (`runner.go:121`) and sdlc's `processgroup.Terminate` (`cmd/sdlc/internal/processgroup/process_unix.go:32`). A grep finds no raw `syscall.Kill(-pgid, …)` left outside tests.
   - `procgrouptest.ZombieGroup` builds a real zombie-only group and, on macOS, waits until the kernel actually answers EPERM. So the tests check the real field condition instead of a mock of it.
   - The doc comment at `pkg/procgroup/procgroup.go:22-24` says plainly that "done" does not mean "released". That is the key rule here, and the atlas states it too.
   - The `finishProducers` test (`finish_producers_test.go`) has a separate process hold the lease and release it late. That is exactly the situation the shipped bug got wrong.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `pkg/procgroup/procgroup.go:4`: the package comment says it is "the one place weave and sdlc read the kernel's answer". That is too broad. `cmd/weave/internal/staging/stage.go:85` and `cmd/sdlc/repolock.go:176` also read `kill`'s answer, for single-process liveness checks where EPERM deliberately means "alive". It should say "for process-group signals".
   - The plan's item 2 describes a `killGroup` seam and an "EPERM-equivalent kill" fake. Neither exists. The test uses a real zombie group through `procgroup.Kill`, which is better, but the plan describes a different design.
   - `pkg/procgroup/procgrouptest/zombie.go:16-17`: one doc-comment line is far longer than the others.

5. **Test coverage**
   - The tests cover the zombie-only group (EPERM on macOS), the empty group (ESRCH) and the live group (signal delivered), plus `finishProducers` waiting after a "done" answer and `Terminate` in both force modes.
   - On Linux, a zombie still accepts signals, so the `finishProducers` test passes there without reaching the EPERM branch. That's acceptable because the bug only exists on macOS.
   - The "50x under load" clause in Done-when rests on the implementor's logged runs: 0 failures in 200 with the fix, 6 in 200 with it reverted. I independently checked the revert-turns-red behaviour and a 20x run without load. I did not reproduce the load run.

6. **Architecture**
   - **ARCH-DRY: pass.** Three copies of the "map ESRCH to `ErrProcessDone`" code became one helper.
   - **ARCH-PURE: pass.** `procgroup.Kill` is a thin IO wrapper and the waiting logic stays in its caller.
   - **ARCH-PURPOSE: pass.** All production process-group signal sites were swept, and the issue's purpose (stop stage-survival flakes) is met at the root cause.
   - If a future caller ever runs groups that contain processes owned by other users, reading EPERM as "done" is wrong for them. The comment at `procgroup.go:18-19` documents that assumption.

7. **Plan revision recommendations**
   - Add a "## Revisions" entry: "Plan item 2: no `killGroup` seam was introduced; the deterministic test drives `finishProducers` against a real zombie-only group (`procgrouptest.ZombieGroup`) whose `killpg` answer is the kernel's EPERM on macOS."

```findings
findings:
  - id: new
    severity: Minor
    family: doc-claim-accuracy
    title: |
      procgroup package doc claims to be "the one place" weave and sdlc read the kernel's kill answer
    detail: |
      pkg/procgroup/procgroup.go:4 overstates it. weave staging/stage.go:85 and sdlc repolock.go:176 also read kill(pid,0) answers, for single-process liveness checks where EPERM means alive. Narrow the claim to process-group signals. No other count or "only" claim in the window is wrong; the atlas text at atlas/workflow/weave.md:353-358 checks out.
  - id: new
    severity: Minor
    family: plan-matches-code
    title: |
      Plan item 2 names a killGroup seam and EPERM-equivalent fake kill that were never built
    detail: |
      The finishProducers test uses a real zombie group through procgroup.Kill, not a seam. Add a Revisions entry so the plan describes the delivered design.
  - id: new
    severity: Minor
    family: comment-formatting
    title: |
      ZombieGroup doc comment has one line much longer than the rest
    detail: |
      pkg/procgroup/procgrouptest/zombie.go:16-17. It is the only such line in the window.
```
