---
id: 000356
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '3a22d3c5f96619ba6e919580e051eebe043db91d' # card fields mirrored from issue-cards; edit via sdlc
---

# test: TestToolsDoubleColonFailurePropagates fails under make -k (inherited MAKEFLAGS=k reaches its nested make); make it hermetic

## Problem

The TL's rule since 2026-10-13 is to run full suites as `make -k test` (a cited known failure must never hide others). `make -k` exports `MAKEFLAGS=k` into the environment, and `cmd/weave/internal/startup` `TestToolsDoubleColonFailurePropagates` runs a nested `make` that expects to stop at the first failure; with the inherited `k` it keeps going and the test fails (ariadne:5 reproduced: FAIL with MAKEFLAGS=k, PASS 0.086 s with it unset).

## Spec

Make the test hermetic: the nested make runs with MAKEFLAGS (and MFLAGS, MAKELEVEL) cleared or explicitly set, so the caller's make options never reach it. Sweep the repo for other tests that shell out to make and apply the same rule (one helper).

## Done when

- The test passes under `make -k test` and plain `make test`; a guard covers every test that invokes make; tests prove it with MAKEFLAGS=k set.

## Plan

- [ ]

## Log

### 2026-10-13
