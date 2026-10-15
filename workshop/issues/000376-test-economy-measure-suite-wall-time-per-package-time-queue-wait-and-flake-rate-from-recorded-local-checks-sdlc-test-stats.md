---
id: 000376
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours: 3.48
card_mirror: '59a5dbff06fab52bc6a7a0daf717f88f901fb6b6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T20:40:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
---

# Test economy: measure suite wall time, per-package time, queue wait and flake rate from recorded local checks (sdlc test-stats)

## Problem

The operator (2026-10-14) made test health and performance a train ("test economy" in the ops TL status file): find a sustainable pattern to keep pair's and ariadne's tests fast and trustworthy. On 10-14 the suite queue set the fleet's pace: ariadne:3's recorded suite waited ~50 min for its turn; pair#528 spent most of the day waiting for or rerunning ~10-minute suites; flakes (#369's two, pair#423, pair#554, the #514 class) cost whole extra runs. Nobody can say where the time goes, because nothing aggregates it.

## Spec

`sdlc test-stats [--since 7d] [--repo …] [--json]`, read-only, from what already exists: recorded local checks (#345/#332 `.git/sdlc/local-checks/<sha>/`), suite logs (`go test -json` / shard summaries), and the suite lock's ticket log (#357) for queue waits. Report per repo: runs, wall time (p50/p90), time per package or shard (top N), queue wait (p50/p90), and **flakes**: tests that failed and then passed on the same head, or failed on main in a baseline run, with counts. It's the first car of the test-economy train: it says where to aim targeted runs (#328) and minimization sweeps (#327).

## Done when

- `sdlc test-stats --since 7d` on this machine prints the per-repo table for ariadne and pair, with the top slow packages and the flake list (tests on recorded fixtures for the parsers).

## Plan

- [ ] Parse recorded suite logs and aggregate per-repository wall, queue, slow-unit and flake metrics; test pure calculations with ariadne and pair fixtures.
- [ ] Preserve one compact record per future suite attempt so retries, queue waits and baseline runs are measurable; test retention and lost status posts.
- [ ] Expose `sdlc test-stats --since 7d [--repo …] [--json]`, verify the real ariadne/pair output, update CLI documentation, run the suite and close review.

Detailed design: [test-stats plan](../plans/000376-test-stats-plan.md). The TL approved recording future attempts because current logs cannot reconstruct all requested metrics.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.4 impl=0.32
item: greenfield-go-module design=0.4 impl=0.32
item: greenfield-go-module design=0.4 impl=0.32
item: smaller-go-module design=0.1 impl=0.2
item: atlas-docs design=0.1 impl=0.08
item: milestone-review design=0.1 impl=0.2
item: issue-spec design=0.2 impl=0.08
design-buffer: 0.15
total: 3.48
```

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. Three greenfield Go concerns are log parsing, aggregation, and durable attempt storage (design 2.0 hours each discounted ×0.2 for the detailed plan, implementation 0.8 hours each scaled ×0.4); the smaller module is CLI wiring on existing Cobra/fleet helpers. The calibration source reports stale relative to its ledger, so this estimate is provisional.

## Log

### 2026-10-14

- Claimed in ariadne:1. Source inspection found ariadne's retained suite logs contain shard summaries and pair's contain Go package timings, but the single log per SHA is overwritten by a retry; queue tickets are ephemeral; pair logs lack total wall time. A per-attempt compact record is needed to measure the full Spec going forward. Historical fields that cannot be recovered will be marked unknown (ARCH-PURPOSE), not estimated from file times.
- Plan-quality round 1 raised PQ-1–PQ-3 (concurrent publication, persisted-input validation, named function tests) and two Minor findings. The durable plan now specifies a private in-flight directory and atomic record marker, safe cleanup, bounded validation, controlled-order tests, reuse of `failingTests`, and an operating envelope. Re-running the gate on that revision.

## Revisions

### 2026-10-14 — TL-approved recorder scope

Record verified slot and nullable agent/model/effort compatible with #367. Save exact head, exit and log hash before status posting so #380 can retry a failed upload from the saved result. Preserve each failed attempt's log across same-head retries for 30 days, together with its metadata. Suite commands and protected gates remain unchanged; historical unknowns stay unknown.
