---
id: 000438
status: codecomplete
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T02:40:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.45
tracker:
    version: 1
    completion:
        token: close-bf4b3be77136
        repository: github.com/xianxu/ariadne
        reviewed_head: f2f3d2ebb4f7cfab41bfbcaf229af02b3f786c11
        evidence_commit: 66731214feaba7fdbe12c7e56428d0b4da2959d9
---

# tests: build each sdlc test binary once per process, not per test

## Problem

Six cmd/sdlc tests run `go build` for their own sdlc binary, a compile per test run (the pair#602 class, caught by the `test-go-build` rule in pkg/golint, #407): settings_native_test.go (weave), leftover_e2e_test.go, trackedlegacy_test.go, fleet_e2e_test.go, issueids_test.go, issuemigrate_test.go. Each carries a `//lint:allow test-go-build` citing this issue.
