---
id: 000449
status: working
deps: [ariadne#394]
github_issue:
created: 2026-10-18
updated: 2026-10-19
estimate_hours: 5.47
collaboration-mode: collaboration
card_mirror: '112b2cd1b9fdcb554e40334be5fbb0a7cfe5e229' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T00:24:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# landing train: batch ready PRs on an interval and land them with one full suite (extends #393)

## Problem

The landing line (#393) lands one PR at a time, each with its own full suite on the combined code. On a busy day that is one full suite per landing, and suites queued 20-50 minutes for the two machine slots (2026-10-18). The operator wants the suite frequency to be a knob.

## Spec

A landing train, extending #393: on a configurable interval (default hourly), take every PR that is admission-ready (review SHIP, branch suite recorded), merge them in admission order onto main in one integration, run one full suite, and land them all if it passes. On failure, find the culprit by splitting (bisect the batch) and land the clean prefix; a known, issue-backed flake (#394) gets its one rerun rather than splitting. A fleet-blocking grant lands alone and immediately, outside the train. Report per batch: PRs, suite result, splits.

## Done when

- With N ready PRs, one train runs one integration suite and lands all N when it passes (test with fakes).
- A failing batch is split and the culprit PR is reported and held while the rest land.
- The interval is configurable; a fleet-blocking grant bypasses the train.
- Depends on #394 (known-flake landing).

## Plan

Durable plan: `workshop/plans/000449-landing-train-plan.md` (state table, Q1–Q7 for the hub).

- [x] M1 — Pure core and readiness: schedule, readiness sidecar, boarding, train state machine, bisect, additive events, landing status, `.sdlc/landing.json`; interval 0 equals today.
- [ ] M2 — Conductor: train branch and PR, one suite, CI, merge; cars resume on merged; GitHub fake merges by reachability.
- [ ] M3 — Split on unexplained failure, known-flake rerun, grant bypass, conflict and CI-red handling, per-train report, live conformance, docs.

## Estimate

Derived against estimate-logic-v3.1 (calibration doc flagged stale). Familiarity 0.9: the landing line, the suite runner and #394's flake step are warm from the last item. Milestone reviews are priced higher than the default, because #394's converge took 1.66h over 10 rounds.

```estimate
model: estimate-logic-v3.1
familiarity: 0.9
item: greenfield-go-module   design=0.4 impl=0.5
item: greenfield-go-module   design=0.4 impl=0.8
item: smaller-go-module      design=0.2 impl=0.4
item: smaller-go-module      design=0.1 impl=0.3
item: smaller-go-module      design=0.2 impl=0.4
item: atlas-docs             design=0.0 impl=0.3
item: milestone-review       design=0.0 impl=0.5
item: milestone-review       design=0.0 impl=0.5
item: milestone-review       design=0.0 impl=0.5
design-buffer: 0.3
total: 5.47
```

Items in order:
1. The train core: schedule, boarding, state machine and bisect.
2. The conductor shell: train branch, PR, suite, CI and merge.
3. Car readiness and the resolved landed commit.
4. The fake that merges by reachability, plus the readiness sidecar.
5. Split, grant bypass and conflict holds.
6. Docs.
7. Three milestone reviews.

## Log

### 2026-10-18

### 2026-10-19

- Claimed and started from ariadne:2. Explored the #393 landing line (flock tickets, head.lock, events.jsonl; no daemon in sdlc, so the schedule rides on admission). Wrote the durable plan; sent its summary and Q1–Q7 to the hub before code.
- Plan-quality: 3 rounds. PQ-1: `.sdlc/*` IS local-gates guarded, corrected to the hub, so the pilot needs the operator label. PQ-2: the sidecar moved to `ready/`. PQ-3: every `MergeOID` consumer goes through one landed commit (M2). Estimate 5.47h.
- M1 built:
  - the pure core (4c755312): a mutation of split is caught by the random-history invariants;
  - the car side, config, admission and status (5918c6fd).
- Found while reading for M2: `parseLandingPRs` refuses a MERGED PR without `mergeCommit` (ghlanding.go). That is a fifth `MergeOID` consumer beyond PQ-3's list, and it moves into the landed-commit rule in M2.
- #460 audit (ariadne:1): the readiness sidecar keeps unknown members across a same-ticket rewrite, with strict known members and no duplicate keys, plus regressions.
- Verification: the landingline package; the new cmd tests under -race; `go test -run 'Landing|Merge|Integrat|CIGate|Local|Archive|Tracker' ./cmd/sdlc/` ok in 1922s, so interval 0 leaves landing unchanged.
