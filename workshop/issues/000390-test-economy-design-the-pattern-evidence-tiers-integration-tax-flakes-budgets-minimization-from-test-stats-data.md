---
id: 000390
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '4b230c432e31a158915d4ee628337c4a61c6a19b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-15T13:05:00-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# test-economy: design the pattern (evidence tiers, integration tax, flakes, budgets, minimization) from test-stats data

## Problem

The ops project **test-economy** (`teams/devinfra/projects/test-economy/project.md`; pensive
`workshop/pensive/2026-10-14-01-pensive-test-economy.md`) now has measurement (#376 `sdlc test-stats`) and wants its
next task: **design the pattern** that keeps pair's and ariadne's tests healthy and fast as agents add tests. First
numbers (`sdlc test-stats --since 1d`, 10-15): ariadne 20 runs, wall p50/p90 531/621 s, **queue p50/p90 554/1769 s**,
slowest shards ~600 s each; pair 24 runs, wall 556/786 s, queue p90 1842 s, **couchcore alone 481 s** of the suite.
10-14/15 also showed the **integration tax**: every landing forces other ready branches into another full merged-head
suite (#376 needed three integrations and hit merge's 3-retry cap; #528 several).

## Spec

Write the design (a plan doc in this issue, for the operator's review) answering, with the data:
1. **Evidence tiers:** what close requires (targeted tests via a traceability map, ariadne#328?) vs. what landing
   requires (the full suite), and whether doc-only drift needs a rerun.
2. **The integration tax:** a merge queue that integrates several ready branches in one suite run, or rules that skip
   reruns for non-code drift; with expected savings from the queue data.
3. **Flake policy:** root cause within a day, never retry or raise a timeout; who owns the standing list; how
   test-stats' flake detection feeds it.
4. **Budgets:** per-package time budgets in `.sdlc/` (couchcore 481 s as the first case) and what happens on overrun.
5. **Minimization** (ariadne#327): how redundant tests are found (mutation or coverage overlap).

Each answer names its implementing issue(s) (existing ones: ariadne#328, #327, #380; new ones as needed) and the expected
effect on wall and queue time.

## Done when

- The design is in this issue's plan, reviewed by the operator; implementing issues are filed and listed in the ops
  project file.

## Plan

- [ ] Draft the operator decision record at `workshop/plans/000390-test-economy-pattern-plan.md` from #376 attempt data and current close/merge contracts.
- [ ] Have the operator review the policy choices; record approved decisions as a dated revision.
- [ ] File the implementing issues and list them in the ops project; close this design-only issue after review.

## Log

### 2026-10-15

Live `sdlc test-stats --since 7d` at 13:05 PDT: ariadne wall p50/p90
531.4/620.8 s (19/20 coverage), queue 602.0/1769.0 s (9/20); pair wall
555.6/786.1 s and queue 0.1/1841.5 s (24/24). Pair couchcore p90 is
481.3 s (n=21). The earlier 554 s queue median was a different snapshot;
the plan pins this newer one and reports sample coverage. Zero confirmed
same-head flakes in 9/24 complete attempts does not imply zero flakes overall.
