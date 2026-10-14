---
id: 000361
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'caa96e389a15f7292d4870ac878034435a8b2035' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T19:18:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ]

## Log

### 2026-10-13

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).
