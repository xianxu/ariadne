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
flow: {kind: quick, provenance: inferred, spec: "a78b6b3e", done: "e87c0e11"}
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

- [x] Draft the operator decision record at `workshop/plans/000390-test-economy-pattern-plan.md` from #376 attempt data and current close/merge contracts.
- [x] Have the operator review the policy choices; record approved decisions as a dated revision.
- [x] File the implementing issues and list them in the ops project; close this design-only issue after review.

## Log

### 2026-10-15
- 2026-10-15: closed — Operator approved the design; #392-#398 and pair#578 are published and listed in the ops project. BR-1 main-failure classification is corrected in the plan and side doc; docs pass git diff --check and make -k. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

Live `sdlc test-stats --since 7d` at 13:05 PDT: ariadne wall p50/p90
531.4/620.8 s (19/20 coverage), queue 602.0/1769.0 s (9/20); pair wall
555.6/786.1 s and queue 0.1/1841.5 s (24/24). Pair couchcore p90 is
481.3 s (n=21). The earlier 554 s queue median was a different snapshot;
the plan pins this newer one and reports sample coverage. Zero confirmed
same-head flakes in 9/24 complete attempts does not imply zero flakes overall.

## Revisions

### 2026-10-15 13:39 PDT — Priority for fleet-blocking repairs

The TL reported #391 waited fourth for a suite while it blocked `sdlc close`
fleet-wide. The plan now specifies a logged, time-bounded operator/TL grant
that lets a fleet-blocking candidate pass waiting normal candidates at the
next admission; it never preempts a running suite or waives landing evidence.

### 2026-10-15 13:49 PDT — Measurement precedes budgets and minimization

Operator review withdrew the plan's proposed couchcore 600 s ceiling and
240 s target pending several weekly trend snapshots. A new implementing issue
will save a weekly per-repo test-stats summary in brain and link it from the
ops project; full suites per landed issue is the headline evidence-tier metric.
Minimization now follows #328's traceability map feature by feature: overlap
nominates, mutation testing measures detection, and review decides removals.

### 2026-10-15 14:29 PDT — Operator's two inline design decisions

The operator set ariadne/pair as pilots but the adopted mechanisms as defaults
for every SDLC repo. Landing runtime data stays local and outside Git; a later
cloud backup must succeed before deletion. Both are recorded in the durable
plan's Revision. Three side docs were added for further operator review and
remain open.

### 2026-10-15 14:40 PDT — Side-doc decisions approved

The operator approved the per-repo landing queue and its storage decision,
the issue-backed known-flake rerun rule, and the feature-scoped minimization
algorithm. The durable plan now records all three; #390 remains open only to
file the implementing issues and link them from the ops project before close.

### 2026-10-15 14:43 PDT — Implementation cards filed

The approved follow-on work is filed and published, unclaimed: #392 weekly
measurement; #393 serial landing admission; #394 issue-backed flake landing;
#395 feature-scoped minimization; #396 shared guidance. Deferred follow-ups
are #397 measured budgets and #398 proven doc-only reuse; pair#578 covers
couchcore profiling. Existing #328, #327 and #380 remain in the rollout. The
TL has the full list for the ops project; its update is pending verification.

### 2026-10-15 14:44 PDT — Project linkage verified

The TL updated `ops:teams/devinfra/projects/test-economy/project.md` Tasks
with #392–#398, pair#578, and the existing #328/#327 work. The approved
design, independent cards and project view now agree; this issue made no code
or gate changes.

### 2026-10-15 14:48 PDT — Boundary review BR-1 addressed

The first close review identified an overbroad flake predicate in side doc 02:
a failure seen on main can be a persistent regression. The plan and side doc
now require same-head fail→pass evidence before known-flake eligibility, and
keep an unproven main failure blocking. #394 already names persistent main
failure as a separate acceptance case. Review rerun is pending.
