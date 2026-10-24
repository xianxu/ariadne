---
id: 000586
status: open
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
card_mirror: 'e2d7f6597ea7ef7fab562d72800c0d3ed816a07e' # card fields mirrored from issue-cards; edit via sdlc
---

# Resume an affected-check attempt after a registered flake, running the groups it never reached

## Problem

## Spec

**An affected-check attempt can't be resumed after a registered flake; only suite attempts can (pair:9 on pair#841,
10-23).** A pair `--pre-close` affected attempt failed on a single test, which pair:9 then proved was a flake and
registered (pair#854). The supported same-head resume refused: resume only supports suite attempts, and the original
affected run had stopped at the Go failure before running the script groups. So the only path was a docs-only commit
and a fresh affected run.

Fix: support resuming an affected attempt the same way as a suite attempt: rerun the registered failing tests, then
run whatever selected groups the original attempt never reached (the script groups here), and record the combined
result for the same head. Refuse, as today, when an unregistered failure remains.

## Done when

- An affected attempt that failed only on registered flakes, and stopped before some groups, resumes: reruns those
  tests, runs the unreached groups, and records a pass for the head (test).
- An unregistered failure still refuses the resume (test).

## Plan

- [ ]

## Log

### 2026-10-23
