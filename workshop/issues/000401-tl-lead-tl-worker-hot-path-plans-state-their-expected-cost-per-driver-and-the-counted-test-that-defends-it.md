---
id: 000401
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
card_mirror: '1298dcb3904e5921fca87932852421a982253a6e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T20:07:31-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "dd9f3c53", done: "d118cb10"}
---

# tl-lead/tl-worker: hot-path plans state their expected cost per driver and the counted test that defends it

## Problem

Resource regressions (pair#579's 61× writes, pair#584's 15 s decode) passed plan review because nobody asked what an
operation should cost. Ops project resource-observability decided (10-15) that hot-path changes state their expected
cost and count it.

## Spec

- tl-lead's plan-review checklist gains one question for changes on a hot path (per-byte, per-event, per-launch or
  per-scan code): "What does this cost per unit of its driver (TTY bytes, events, launches, slots), and which counted
  test defends it?" Counts, not timings.
- tl-worker's planning rules gain the matching line, so plans arrive with it answered.
- Both phrase-guarded in the help tests.

## Done when

- `tl-lead --help` and `tl-worker --help` carry the question; help tests guard the phrases.

## Plan

- [x] Add phrase guards to both help tests for the growth drivers, expected work per unit, and a counted test; verify red against current help.
- [x] Add a short HOT-PATH COST rule and a log-reader example to shared `pkg/tlrules/planning.md`, composed into both binaries. Require expected cost for every driver and a deterministic operation-count assertion at representative sizes; timing alone is not the defense.
- [ ] Update the atlas pointer, run focused and built-help checks, then the recorded exact-head suite, close review, and normal PR/merge gates.

## Log

### 2026-10-15

2026-10-17: Use the existing shared PLANNING source rather than separate worker and lead paragraphs (ARCH-DRY). The worked example contrasts one pass over N log entries, O(N), with rescanning all prior entries for each one, O(N²); a counted read/scan assertion at N and 2N defends linear growth. The rule asks for expected operations per byte, event, launch, scan, or other driver; timing-only benchmarks are supplemental.

The hub approved the N/2N counted-work design before code.

Six new phrase guards in each help test failed against old output for both invocation forms, then `go test ./cmd/tl-worker ./cmd/tl-lead ./pkg/tlrules -count=1` passed after adding the shared rule. `make -k build` passed, and both built help outputs print the same five-line HOT-PATH COST rule. The atlas now points to the shared source (ARCH-DRY).
