---
id: 000439
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# lint: decide on golangci-lint for the off-the-shelf Go linters (operator: install and gates)

## Problem

#407 shipped the curated catalog as stdlib analyzers in pkg/golint, run by the suite. Off-the-shelf linters (usetesting, thelper, paralleltest/tparallel, forbidigo, noctx, bodyclose, errorlint, nolintlint) need golangci-lint, which is not installed: adopting it is a machine change (operator) and, at landing or in CI, a local-gates change (operator's label).
