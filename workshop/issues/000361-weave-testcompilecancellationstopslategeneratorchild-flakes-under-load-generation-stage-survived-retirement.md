---
id: 000361
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '77bd89779dadc022f59c181b662e03c27c6fb40e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T19:18:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "991a59d0", done: "a3c27513"}
---

# weave: TestCompileCancellationStopsLateGeneratorChild flakes under load (generation stage survived retirement)

## Problem

`cmd/weave` `TestCompileCancellationStopsLateGeneratorChild`
(`generator_lifetime_test.go:75`) failed once in #353's full suite (2026-10-13, load ~38)
with `generation stage survived retirement: .generation-weave-…`. Alone it passed 60/60.

In this test the stage's owner is the live test process, so the retry's `Reclaim`
correctly never takes it; only the cancelled compile can remove its own stage. On
cancellation, `finishProducers` kills the generator's group and waits `staging.Settle`
(1s) for the inherited lease (`AwaitExclusive`); if the lease is still held after that,
it preserves the stage by design. The likely failure is that bound expiring under load
(a SIGKILLed child taking over 1s to release fd 3), or `staging.Remove`'s non-blocking
`Exclusive` meeting the same exit window #343 fixed for `Reclaim`. Not a #343 regression:
#343 only moved `finishProducers`' unchanged 1s loop into `AwaitExclusive`, and `Reclaim`'s
settle does not apply to a live owner.

## Spec

Confirm which path preserved the stage (instrument on failure under load, as #343 did),
then fix the cause: for example let `staging.Remove` settle like `Reclaim`, or size the
settle for load, without letting a genuinely live producer's stage be removed.

## Done when

- The failing path is identified with evidence; the test passes 50x under concurrent load,
  and the fix fails the test when reverted.

## Plan

- [ ] `pkg/procgroup.Kill(pgid, sig)`: ESRCH and EPERM → `os.ErrProcessDone` (documented, measured XNU semantics); unit test on a zombie-only group
- [ ] weave `finishProducers`: a done group is not an error; always settle on the lease (the proof); seam `killGroup` + deterministic test (an exiting holder, EPERM-equivalent kill) that fails when reverted
- [ ] Sweep the class: weave `cmd.Cancel` and sdlc `processgroup.Terminate` use the helper
- [ ] Prove under load: the flaky test 50x+ under CPU load with the fix; with the fix reverted it fails

## Log

### 2026-10-13
- 2026-10-13: unclaimed: filed from #353; independent flake, unclaimed for the hub to assign

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).
- 2026-10-13 (ariadne:1) root cause, with evidence. I traced `finishProducers` and `staging.Remove` (temporary instrumentation, now removed) and reproduced under 24–40 CPU burners: 4/60, 1/60, then 3/200 failures. The issue's hypothesis (the 1s settle expiring) is **wrong**: no failing run reached the preserve path, and each took about 0.3s. In every failure:
  1. The first `Exclusive` was busy.
  2. `kill(-pgid, SIGKILL)` returned **EPERM**.
  3. `finishProducers` returned that error at once, skipping `AwaitExclusive`.
  4. `Remove`'s non-blocking `Exclusive` then met the still-exiting holder (`lsof` a moment later showed none), and the stage survived.

  A measurement on this Mac: `kill(-pgid, 0|SIGKILL)` on a group whose only member is an unreaped zombie returns EPERM, and ESRCH only after the reap. XNU skips exiting/zombie processes in killpg and reports EPERM, but marks a process exiting before its descriptors close. So "group can't be signalled" ≠ "lease released". It's a real retirement bug in weave, not a test-timing assumption. The same EPERM-as-failure handling is in weave's `cmd.Cancel` and sdlc's `processgroup.Terminate` (the class).

