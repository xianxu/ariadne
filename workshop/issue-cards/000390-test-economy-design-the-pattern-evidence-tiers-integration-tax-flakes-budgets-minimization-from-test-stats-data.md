---
id: 000390
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
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
