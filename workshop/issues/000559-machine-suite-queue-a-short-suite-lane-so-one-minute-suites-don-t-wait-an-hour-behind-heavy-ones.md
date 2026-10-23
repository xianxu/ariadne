---
id: 000559
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'd0d3e0926b6c457b8f61b47a7d48ae20cfbf50d4' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T06:42:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# Machine suite queue: a short-suite lane so one-minute suites don't wait an hour behind heavy ones

## Problem

## Spec

From ops-echo:0 (10-22): echo's recorded suite runs in 30–60 s but waits 1–1.5 h in the machine-wide suite queue
(echo#4 at position 14, echo#11 at 12), mostly behind ariadne's sharded suite (weight 2, ~27 min) and pair's
~12-minute runs. Three echo issues today sat finished for over an hour waiting for a one-minute suite. The same queue
delays every small repo's landing behind the heavy suites.

Proposal: a **short-suite lane** in the machine test budget (#319/#341/#346). A suite whose recorded wall time (from
its last N recorded attempts on that repo, not self-declared) is under a threshold (about 2 minutes) may run in a
dedicated light slot that heavy suites never hold, so it isn't queued behind them. Bounded: one light slot, a light
run that exceeds the threshold by a margin is stopped and requeued as normal (and its repo loses lane eligibility
until it measures short again), and the light slot's load is small enough not to disturb a running heavy suite's
timing.

**Operator and ops-echo additions (10-22):** "nice to have, but be careful of starvation; echo will surely grow to be
as complex as pair." Bound starvation both ways: heavy suites keep their FIFO turn however many short ones arrive
(the light slot never borrows a normal slot or delays a heavy suite's admission), and a repo whose recorded wall time
grows past the limit leaves the lane automatically by measurement, never by a flag. Also: the machine's CPU sits at
100% constantly, so CPU, not slot count, limits landing throughput. The light slot must be cheap enough (one short
suite, niced) not to slow a running heavy suite measurably; measure that, don't assume it.

## Done when

- A repo whose recorded suite wall time is under the threshold runs immediately in the light slot while both normal
  slots are held by a weight-2 suite (test).
- A light run exceeding the bound is stopped and requeued normally, and eligibility is re-derived from measurements
  (test).
- `sdlc queue status` shows the lane and its holder.
- Heavy suites keep their FIFO admission under a continuous stream of short ones (test); a running heavy suite's wall time with the light slot busy is measured and within a stated bound.

## Plan

Durable plan: `workshop/plans/000559-machine-suite-queue-a-short-suite-lane-so-one-minute-suites-don-t-wait-an-hour-behind-heavy-ones-plan.md` (scope choices a–e approved by the TL, 10-23).

- [ ] Eligibility (pure) and the attempt `lane` field
- [ ] Helper light line (`--lane light`, `light_order`, `lane-v1`, inspect `lane`)
- [ ] Capability probe and lane-aware acquisition
- [ ] Run order and overrun requeue
- [ ] Status and atlas
- [ ] Verification: targeted tests, recorded suite, heavy-impact measurement

## Log

### 2026-10-23
- Claimed and planned (ariadne:5). Measured: echo's recorded suites run 61–88 s and queued 2000–4480 s today; ariadne holds 445 attempt records.
- TL approved all five scope choices (a)–(e); add an atlas note that the lane is a third concurrent run beyond WF_TEST_SUITES=2, bounded to light suites. (scope choices a–e approved)

### 2026-10-22
