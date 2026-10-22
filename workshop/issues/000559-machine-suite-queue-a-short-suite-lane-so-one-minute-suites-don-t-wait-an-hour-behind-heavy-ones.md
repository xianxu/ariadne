---
id: 000559
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'de9e1f4444edbc94113977091b4658a4bc1f6c9d' # card fields mirrored from issue-cards; edit via sdlc
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

## Done when

- A repo whose recorded suite wall time is under the threshold runs immediately in the light slot while both normal
  slots are held by a weight-2 suite (test).
- A light run exceeding the bound is stopped and requeued normally, and eligibility is re-derived from measurements
  (test).
- `sdlc queue status` shows the lane and its holder.

## Plan

- [ ]

## Log

### 2026-10-22
