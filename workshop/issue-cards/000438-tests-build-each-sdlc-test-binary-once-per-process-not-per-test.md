---
id: 000438
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
---

# tests: build each sdlc test binary once per process, not per test

## Problem

Six cmd/sdlc tests run `go build` for their own sdlc binary, a compile per test run (the pair#602 class, caught by the `test-go-build` rule in pkg/golint, #407): settings_native_test.go (weave), leftover_e2e_test.go, trackedlegacy_test.go, fleet_e2e_test.go, issueids_test.go, issuemigrate_test.go. Each carries a `//lint:allow test-go-build` citing this issue.
