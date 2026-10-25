---
id: 000622
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours: 3.84
card_mirror: 'd8efc2fbd21a5516e947e2bc234275d4c92b9bdf' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T04:05:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Sweep second-scale wall-clock bounds from correctness tests; golint rule for the shape

## Problem

## Spec

**Second-scale wall-clock bounds in correctness tests (found by ariadne:2 fixing ariadne#620, 10-25).** ariadne#620's
BR-16 subtest asserted that a run finished within 8 s of wall time; under fleet load it failed on main and blocked
every ariadne landing. ariadne:2 fixed it by asserting the outcome (the rerun's saved attempt timed out) instead of
the elapsed time. The same shape remains elsewhere: `cmd/sdlc/ghlanding_test.go:153`,
`helpercapability_test.go:167`, `landingattribution_impact_test.go:126`, and tests in tl-lead and weave.

Test-economy's rule (project file, 10-15): no wall-clock assertions in correctness tests; wait on signals, inject
clocks, assert outcomes. Sweep these to outcome or injected-clock assertions, and add a `pkg/golint` rule that flags
a duration comparison against `time.Since`/`time.Now` deltas in `_test.go` files (with a reviewed escape hatch for
true performance tests), so new ones are caught at lint time.

## Done when

- Each named site asserts an outcome or uses an injected clock; each change proven with -count=10 under load.
- The golint rule flags the old shape (a fixture showing it red on the pre-fix BR-16) and passes the swept tree.

## Plan

Durable plan: `workshop/plans/000622-sweep-second-scale-wall-clock-bounds-from-correctness-tests-golint-rule-for-the-plan.md`.

- [x] M1 — the widened `test-tight-wallclock` rule (a transitional per-file pending list), group A polled deadlines, and #625
- [x] M2 — groups B, C, D; delete the pending list so the rule enforces; close #625 with this issue

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module        design=0.2 impl=0.2
item: cross-cutting-refactor   design=0.5 impl=0.3
item: cross-cutting-refactor   design=0.8 impl=0.5
item: cross-cutting-refactor   design=0.4 impl=0.25
item: milestone-review         design=0.0 impl=0.2
item: milestone-review         design=0.0 impl=0.2
design-buffer: 0.15
total: 3.84
```

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.* Items: the rule extends an existing one (smaller module); three refactor sweeps, one per site family: group A polled deadlines (13, mechanical), group B kill/cancel outcomes (~14, each needs its own outcome oracle, so the most design), groups C+D allows or events (~8); two milestone reviews. `impl=` at 40% of the v2 table midpoints; design from the v2 ranges, with a +15% buffer since the plan is thorough. Proof runs (`-count=10` per package, one at a time) and the mutation checks are inside impl; the estimate-quality review (INFO) judged the flat 0.2 impl too low for group B and the serial proofs, so A, B and C+D impl were raised to 0.3, 0.5 and 0.25 before implementation began.

## Log

### 2026-10-25
- 2026-10-25: closed M1 — Round 5. BR-8: atlas sdlc-binary.md documents AwaitCondition (required workers, nil with a reason, one-value buffer, put-back). duplicated-poll-loop: AwaitCondition moved to pkg/testgate (testgate.T, no testing import), testfix aliases it; weave release await takes workers via waitFor, weavefs passes nil with its reason; the remaining ~18 repo-wide loops are a TL-approved M2 task. Lint: one bindings collector feeds elapsed and deadline matchers; comparisons combining a deadline with time.Now() and v.Sub(time.Now()) are polls; rule tests for each. Proof: testgate AwaitCondition -count=10, weave release and weavefs -count=10 niced, sdlc callers via the alias green; pkg/golint green incl. repo catalog. Earlier rounds: 93 changed-file tests once and 8 representatives -count=10 green.; review verdict: FIX-THEN-SHIP

### 2026-10-25

- TL dispatch: sweep plus rule, fold in #625. The widened rule flags 35 sites (6 named). TL scope call: all four groups here, full flow, M1 = rule + group A + #625, M2 = B/C/D + enforce; proofs per package `-count=10`, one at a time.
- Rule shape: deadlines flagged only when polled (`time.Now().After/Before(v)`, `time.Until(v)`), which cut 73 raw hits (fixture timestamps) to 35.
- M1 built: widened rule with object-keyed variables (it also caught 3 deadlines the name-keyed version hid: queue_interruption:47, landingtrain:520, makebuild_publish:99), transitional `wallclockPending` for the M2 files, 16 group A deadlines to `HangCeiling`, #625 fails fast when a landing ends before its barrier, processgroup's 5s shutdown contract under a reasoned allow.
- Proof found my own bug: #625's barrier file exists before the script writes the SHA; my early `break` took an empty read as the event (5/20 runs). Fixed to re-read on empty.
- M1 proofs (niced, one run at a time): cmd/sdlc 12 group A tests -count=10 green at load ~11 apart from the split pair; split pair -count=10 green after the fix at load 35–59; tl-lead TestOperatingCancellationStopsGitHookDescendants, weavefs TestSetupLeaseSurvivesRunnerParentDeath, weave/release package -count=10 green. Mutation: #625 deadline at HangCeiling/1200 (100ms) → red with "never reached the registration barrier"; restored → green. pkg/golint green including the repo catalog test.
- M2 built: group B on `testgate.Returns` (children outlive `HangCeiling`), group C reasoned allows, the remaining ~18 test poll loops on `AwaitCondition` with explicit workers or reasoned nils, `wallclockPending` deleted, so the rule enforces repo-wide.
- Proof found my own race: `golist` cancelled on a 100ms timer could fire before its descendant existed (1 of 10 under load). It now cancels on the descendant's pid cue; -count=10 green.
- M2 proofs (niced, one package at a time, -count=10): testgate, testfix, affected, fleet, gitx, judge, tl-lead, weave, weave/acquire, weave/release, and the 21 changed cmd/sdlc tests, all green at load ~25. Mutation: `runBounded` with the timer's kill removed fails with "the bounded build returned: not reached within 2m0s" (the hang guard, not a timing check); restored green. pkg/golint green with the rule enforcing on every file.
- Close review round 5 (FIX-THEN-SHIP): BR-12 atlas now documents `testgate.Returns`, its outlive-the-ceiling convention, the up-front worker-channel refusal and the orphaned-child note. Minors: duplicated-poll-loop scope stated (AwaitCondition owns HangCeiling-bounded polls; test-deadline-budgeted `awaitDropCondition`/`mergeStageAwait` and value-returning `AwaitPIDFile`/`AwaitExit` stay, recorded in plan Revisions and atlas); testgate imports moved to the module group in 15 files; plan M1/M2 ticked.
