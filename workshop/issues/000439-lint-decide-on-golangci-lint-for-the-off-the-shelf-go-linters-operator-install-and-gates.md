---
id: 000439
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2c975afef5e7f9362b571a114cda065b7ab12cbf' # card fields mirrored from issue-cards; edit via sdlc
---

# lint: decide on golangci-lint for the off-the-shelf Go linters (operator: install and gates)

## Problem

#407 shipped the curated catalog as stdlib analyzers in pkg/golint, run by the suite. Off-the-shelf linters (usetesting, thelper, paralleltest/tparallel, forbidigo, noctx, bodyclose, errorlint, nolintlint) need golangci-lint, which is not installed: adopting it is a machine change (operator) and, at landing or in CI, a local-gates change (operator's label).

## Spec

## Done when

- The operator decides on golangci-lint (pinned version, where it runs); if adopted, `make lint` runs it with a config, existing findings are fixed or carry reasoned suppressions, and landing runs it.

## Plan

- [ ]

## Log

### 2026-10-18
