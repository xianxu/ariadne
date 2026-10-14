---
id: 000369
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '432936cc379234b38af8e7aea1797c7a0790f8d8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T11:06:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "bddf13a5", done: "70017678"}
---

# TestWaitCIGateBoundsRealMakeProbe/cancel is flaky on main: cancel sometimes returns after ~1.05s (the make child's sleep) instead of under 700ms; also TestSetupLeaseSurvivesRunnerParentDeath failed once under the full suite

## Problem

## Spec

Cancellation of the real make facts probe must retire its owned process group
and inherited pipe holders promptly. Preserve the existing 700ms assertion and
one-second WaitDelay; fix the observed process lifecycle, not a timeout.

## Done when

- Reproduce the cancellation failure and explain the surviving process/pipe.
- Existing cancel check passes 50 repeated runs after the fix; add a focused
  regression for the measured ordering and preserve dying-group semantics.
- Recorded suite and shared SDLC acceptance permit landing.

## Plan

- [x] Measure kernel signal results and descendant lifetime on cancellation.
- [x] Apply the smallest root-cause fix with a failing regression (ARCH-DRY).
- [x] Run 50 cancellations, relevant lifecycle tests and the recorded suite.

## Log

### 2026-10-14

#365 is checkpointed and paused on its own branch; #369 starts from main.
Baseline on this branch: one cancellation failure in50 runs, elapsed1.017309208s,
matching the1s WaitDelay. Investigation will distinguish signal failure from
late descendants rather than assuming the #361 hypothesis. ARCH-ORDER: process
exit evidence does not prove inherited pipes have been released.

### 2026-10-14 — measured root cause and regression

Baseline cancel failed1/50 (1.017309208s). A temporary real-make trace failed
3/100: syscall group termination returned nil, yet a live orphan sleep remained
in the SAME group (e.g. sleep19654,pgid19648). Therefore #361's EPERM mapping
is not this failure: a successful group signal is one traversal, not a fork
barrier; it can miss a child born during teardown that holds inherited pipes.

The deterministic regression signals only the ready parent on the first
successful pass, preserving a real child/pipe. It failed with1.001s drain and
a surviving child, then passed with bounded forceful re-signalling. Forceful
teardown repeats every5ms until the kernel reports done, capped at100ms;
graceful signalling and existing WaitDelay/700ms assertions stay unchanged.
Preserve nil after a successful first signal so exec.Cmd.Cancel cannot suppress
context cancellation when the parent exits zero; a separate regression caught
that return-value edge before correction.

Final repeat: budget0/50 failures, cancel0/50 failures. Full internal processgroup
and procgroup packages pass; judge13.517s, gitx90.813s pass. The reported lease
test passed10 repeats (0.896s). Repeating both probe modes additionally exposed
10/50 budget-only fixture failures: a100ms deadline legitimately expires before
the first shell marker is scheduled. Require start proof only for explicit
cancel (which waits for that marker); both modes still check timing/error and
worktree cleanup. No production timeout or elapsed assertion increased.

### 2026-10-14 — suite blocked by lease flake, diagnostic repair

Recorded suite on3c507cda18d9d357ada0c7328b4497185fe017ee passed all1342 root
tests but failed the known weavefs lease test (0.08s), total524.9s. Focused
lease100 repeats, whole-package50 repeats, and focused1000 repeats passed.
Peer :2 confirmed its earlier full-suite failure had no retained assertion
either. The summarizer wrongly assumes parent test2json output contains child
assertions and drops failed subtests. A pure recorded-event regression caught
that omission before removing the skip; it now runs from the root Go suite.
Next recorded suite will retain the actual lease assertion if it recurs; no
lease timing/production change is guessed without that evidence.

### 2026-10-14 — lease authority measured and repaired

Under other active suites the known lease case reproduced1/500: after
'released', AcquireSetup still returned ErrSetupInUse. Keeping the payload
alive on a second explicit barrier reproduced2/100; tracing500 repetitions
found four cases where the lease became available1.2–1.6ms later while the
child was still awaiting 'finish'. The message/process notification precedes
resource release; it is not evidence of persistent leakage.

The fixture now keeps the child alive through the resource probes, polls setup
release within existing staging.Settle, and uses existing AwaitExclusive for
the stage. Strict exclusion before release remains immediate. Fixed lease
case passed500 repeats. An omitted-close mutation failed in1.02s, proving a
real retained descriptor cannot pass merely because the child eventually exits.
Related package testing exposed the same post-Wait assumption in
TestInheritedLeaseOutlivesParentCopy; settle on that lease before Remove too
(ARCH-PURPOSE/ORDER). Both packages passed10 repeats16.213s/12.027s.

The queued follow-up suite was cancelled at admission (exit130) for these
measured repairs; no passing record is claimed. Next commit, record the full
suite on the repaired head, then shared close/PR/merge.

### 2026-10-14 — recorded full-suite acceptance

sdlc local-check run --name suite passed on
b74440ee82f01bbc41b982f8377bd1383c2c8553:1354 cmd/sdlc tests in12 shards,
54 other packages, zero failures,445.8s (7m26s). Recorded sdlc-local/suite
success, log sha256 prefixc47b0021ba8a. All code is in that tested head; this
checkpoint is docs-only. Next shared close review, PR and merge. #365 stays
paused at85b35d75, with its implementation/suite/landing status checkpointed.
