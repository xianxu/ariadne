---
id: 000432
status: working
deps: []
github_issue:
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
collaboration-mode: collaboration
card_mirror: '148ba2e0b06981f01e438d4f49dbe3cbe426e22d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T18:01:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "de8febc6", done: "2f1dc36a"}
---

# tl-worker: an early draft PR for CI overlap needs no hub approval

## Problem

Slots ask the hub for permission before opening an early draft PR so GitHub CI overlaps a queued local suite (pair:3 asked twice on 2026-10-17). The hub always says yes, because the suite, close review and merge gates still run before landing. Each ask costs a hub round trip while the suite queue is long.

## Spec

tl-worker help states it as a rule: an early draft PR for CI overlap needs no hub approval, provided the recorded suite, close review and `sdlc merge` gates still precede landing. Add it as a rules row like the other worker rules.

## Done when

- `tl-worker --help` carries the rule, and its rules test covers the phrase.

## Plan

- [x] Add a worker rules-table assertion for the early draft PR permission and confirm it fails against current help.
- [x] Add one help rule beside the existing permissions and run the focused tests.
- [x] Run the recorded suite on the pushed head.
- [x] Submit the tested branch to the close review and normal PR and merge gates.

## Log

### 2026-10-17
- 2026-10-17: closed — TDD rules-table failure observed before help edit; go test ./cmd/tl-worker -count=1 and make -k build passed; built tl-worker help shows draft-PR permission and preserved suite/review/merge gates; exact-head sdlc-local/suite passed on 729a94cb9085c92ca720ae502bfbec0dae25e015 (12 shards, zero failures).; review verdict: SHIP

The hub approved quick flow. The existing embedded help and `rules.tsv` remain the single source and coverage pattern (ARCH-DRY); no PR or landing behavior changes.

The new rule failed twice against the old help (both invocation forms), then `go test ./cmd/tl-worker -count=1` passed after the help change. `make -k build` passed, and the built `bin/tl-worker --help` prints the permission and landing-gate condition.

Opened PR #291 as a draft while the suite queued. The recorded `sdlc-local/suite` on 729a94cb9085c92ca720ae502bfbec0dae25e015 passed: 12 shards, 1577 sdlc tests, 67 other packages, zero failures (log sha256:f763c937b6ca). No new atlas surface is introduced by the help wording.
