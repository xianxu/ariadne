---
id: 000449
status: open
deps: [ariadne#394]
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '887b764a9302cac08923d5e487b1efa4d1f9bccf' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ]

## Log

### 2026-10-18
