---
id: 000390
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'defa58e2dd6db303ad608c500786489db471a875' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-15
