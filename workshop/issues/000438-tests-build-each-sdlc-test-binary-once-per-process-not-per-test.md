---
id: 000438
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: '6788d3f6e4c68633e93476e0120f01b0ca5f4e6c' # card fields mirrored from issue-cards; edit via sdlc
---

# tests: build each sdlc test binary once per process, not per test

## Problem

Six cmd/sdlc tests run `go build` for their own sdlc binary, a compile per test run (the pair#602 class, caught by the `test-go-build` rule in pkg/golint, #407): settings_native_test.go (weave), leftover_e2e_test.go, trackedlegacy_test.go, fleet_e2e_test.go, issueids_test.go, issuemigrate_test.go. Each carries a `//lint:allow test-go-build` citing this issue.

## Spec

## Done when

- The package builds each binary at most once per test process (a shared helper or TestMain), the six call sites use it, and their suppressions are gone.

## Plan

- [ ]

## Log

### 2026-10-18
