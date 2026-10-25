---
id: 000630
status: working
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours: 1.99
card_mirror: 'ad6efb64fb42ebaabd046947f82415d45f4e9531' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-25T09:51:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Shard timeouts with no failing test are 'incomplete', not failures: retry once, skip the main-baseline diagnosis

## Problem

## Spec

**An all-shard timeout under load is treated as an unattributable failure and triggers a full main-baseline run
(ariadne:5 on ariadne#616, 10-25).** #616's integrated merge suite (04:31–05:01) hit the 1800 s shard timeout on every
shard while the machine's load average was 30–50; 764 `cmd/sdlc` tests never ran (the same suite took 19–20 min
earlier). sdlc classified it as an unattributable failure and started a main-baseline diagnosis: another full ariadne
suite (both machine slots), queued behind everything, holding ariadne's landing line for an hour or more, to diagnose
a run that didn't fail any test.

A run where shards time out without a failing test isn't evidence against the code or main. Classify it as
"incomplete (timeout)", distinct from a failure: retry the same head's suite once (in its normal queue turn), and
diagnose against main only if a test actually fails. Report the load average at the timeout in the outcome.

## Done when

- A suite whose shards time out with no failing test records "incomplete (timeout)" with the load at the time,
  retries once, and runs no main-baseline diagnosis; a real test failure still diagnoses as now. Tests for both.
- **The retry runs only the packages that didn't finish** (operator, 10-25): Go reports results per package and sdlc records which passed in the attempt, so the retry selects the unfinished packages through the affected-run machinery instead of rerunning the whole suite; test-level granularity is a non-goal. Test: a timed-out attempt with some packages passed retries only the rest.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.05 impl=0.16
item: smaller-go-module      design=0.1  impl=0.16
item: smaller-go-module      design=0.1  impl=0.2
item: smaller-go-module      design=0.2  impl=0.2
item: smaller-go-module      design=0.1  impl=0.2
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.0  impl=0.32
design-buffer: 0.15
total: 1.99
```

The rows are:
- the shard runner's alarm marker and its parse;
- the classifier and the kind;
- the v5 attempt with the load, the typed error, and the guard and flake-view treatment (M1);
- the resume composite's incomplete branch: the v5 suite link, the pass rule and the version gates;
- the child command, its envelope and the automatic retry (M2);
- the atlas;
- two boundary reviews.

Familiarity is 1.0: the resume code was read in full for the plan (both plan-quality rounds checked its line references).

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: [000630-shard-timeouts-incomplete-plan.md](../plans/000630-shard-timeouts-incomplete-plan.md) (TL decisions
2026-10-25: a passing retry posts green; the retry covers the unfinished packages plus every package with a named
failure).

- [ ] M1 — classify, record, report, skip diagnosis: the shard runner's alarm marker, parsed process facts,
  `classifyIncomplete` and `flakes.IncompleteTimeout`, the v5 attempt plus the separate load record, the typed
  `suiteIncompleteError` (no main baseline, no train split), the guard and flake-view treatment.
- [ ] M2 — the one retry: an incomplete parent is resumable with its own package selection; `runLocalCheck` retries
  once through `resumeSaved`; a pass posts green, incomplete again stops, a real failure diagnoses as today.

## Log

### 2026-10-25
