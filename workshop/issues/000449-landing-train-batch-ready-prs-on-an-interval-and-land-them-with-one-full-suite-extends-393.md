---
id: 000449
status: working
deps: [ariadne#394]
github_issue:
created: 2026-10-18
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '866c71b7d333057f1facb70751b2c460d527ea6b' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T00:24:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] M1 — Pure core and readiness: schedule, readiness sidecar, boarding, train state machine, bisect, additive events, landing status, `.sdlc/landing.json`; interval 0 equals today.
- [ ] M2 — Conductor: train branch and PR, one suite, CI, merge; cars resume on merged; GitHub fake merges by reachability.
- [ ] M3 — Split on unexplained failure, known-flake rerun, grant bypass, conflict and CI-red handling, per-train report, live conformance, docs.

## Log

### 2026-10-18

### 2026-10-19

- Claimed and started from ariadne:2. Explored the #393 landing line (flock tickets, head.lock, events.jsonl; no daemon in sdlc, so the schedule rides on admission). Wrote the durable plan; sent its summary and Q1–Q7 to the hub before code.
