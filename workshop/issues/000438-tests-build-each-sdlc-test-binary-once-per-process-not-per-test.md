---
id: 000438
status: working
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e61ebec54b82ad52e20558e0c20b49640cb31a01' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T02:40:05-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "6e340b9c", done: "4ee7ab6b"}
---

# tests: build each sdlc test binary once per process, not per test

## Problem

Six cmd/sdlc tests run `go build` for their own sdlc binary, a compile per test run (the pair#602 class, caught by the `test-go-build` rule in pkg/golint, #407): settings_native_test.go (weave), leftover_e2e_test.go, trackedlegacy_test.go, fleet_e2e_test.go, issueids_test.go, issuemigrate_test.go. Each carries a `//lint:allow test-go-build` citing this issue.

## Spec

**Measured (2026-10-18, ariadne:2):** the binary-building tests were run as a compiled test binary with logging
`go`/`make` wrappers first on PATH. Running them under `go test` doesn't work: it puts `$GOROOT/bin` first on PATH,
so the wrappers are bypassed. Result: **18 `go build -o` runs of sdlc, 17.2 MB each, 310 MB of output per suite**,
plus the linker's own temporary copy, and 1 `go run ../vocabulary`. A warm build takes 0.6–0.9 s on an idle machine
(Go 1.27 appears to cache the link), so the cost is mostly the bytes written plus process start, and it grows with
every new end-to-end test. The 4 `make build` tests use a fake go and cost nothing.

**Design:** `testfix.BuiltBinary(t, dir, pkg, buildArgs...) string` builds each (directory, package, flags) binary
at most once per test process and hands every caller the same path.
- **Memoized and safe under parallel tests:** a mutex-guarded map of per-key entries, each built once. A build
  failure is memoized too, so every caller sees the same error.
- **Lifetime (ARCH-FUNERAL):** binaries go in one per-process directory, `os.MkdirTemp("", "sdlc-testbin-*")`,
  created lazily. `testfix.RemoveBuiltBinaries()` deletes it after `m.Run` in cmd/sdlc's TestMain. A process killed
  mid-suite leaves its directory behind, so the first build in any later process sweeps `sdlc-testbin-*` directories
  older than 24 h. Bound: one 17 MB binary per package per live test process.
- **Sweep the class (ARCH-PURPOSE):** all six `go build` sites (`buildFleetE2EBinary` with its 15 callers,
  leftover_e2e, trackedlegacy, issueids, issuemigrate, and settings_native's weave build with `-trimpath`) and both
  `go run ../vocabulary` calls in project_crud_test move to the helper. Their six `//lint:allow test-go-build`
  suppressions go, and the golint rule then guards the class. The helper lives in a non-test file, which the rule
  (TestOnly) doesn't scan.
- **Expected effect:** builds drop from one per test (18 now, rising) to at most one per package per test process.
  That means at most one per shard that runs such a test in the 12-shard suite, and 1 in a serial run. I'll count
  this after the change with the same wrapper method.
- **Not in scope:** sharing one binary across shards would need the sharded runner (Makefile.workflow, base layer)
  to prebuild it and pass the path in. That's a possible follow-up if the per-shard count is still high.

## Done when

- The package builds each binary at most once per test process (a shared helper or TestMain), the six call sites use it, and their suppressions are gone.

## Plan

- [x] `testfix.BuiltBinary` + `RemoveBuiltBinaries` + stale sweep, with tests (one build for N callers, parallel callers, a memoized failure, cleanup)
- [x] Move the 6 build sites and 2 `go run` calls to it; drop the 6 suppressions
- [x] Re-measure builds and bytes with the wrapper method; record before/after in Log

## Log

### 2026-10-18

- Implemented `testfix.BuiltBinary` and `RemoveBuiltBinaries` (called from cmd/sdlc's TestMain after `m.Run`), with a
  sweep of stale `sdlc-testbin-*` directories. Tests: `TestBuiltBinaryBuildsOncePerProcess` (16 parallel callers, 1
  build; another package or flag set is another build; RemoveBuiltBinaries deletes it),
  `TestBuiltBinaryFailureIsSharedNotRetried`, `TestSweepBuiltDirsRemovesOnlyStaleOnes`. Moved `buildFleetE2EBinary`,
  leftover_e2e, trackedlegacy, issueids, issuemigrate, the settings_native weave build and both `go run ../vocabulary`
  calls to it; the six `test-go-build` suppressions are gone and pkg/golint stays clean.
- **Before/after**, same tests, same wrapper method, one serial process: before, 18 `go build` of sdlc plus 1
  `go run` (310 MB of binaries); after, 1 sdlc build plus 1 vocabulary build (20.9 MB). No `sdlc-testbin-*` directory
  is left after the process exits. In the 12-shard suite the bound is one build per package per shard that runs
  such a test.
- Close review round 1 (FIX-THEN-SHIP): BR-1 adds an atlas paragraph next to the other testfix conventions. Minors
  fixed in the round: the `test-go-build` message names `testfix.BuiltBinary`; builds have a 5-minute bound again
  (the weave site had one); the doc comment says a build runs under the first caller's environment; a test covers
  two spellings of one directory sharing a build. Tidied the issue file's duplicate headings and empty plan item.
