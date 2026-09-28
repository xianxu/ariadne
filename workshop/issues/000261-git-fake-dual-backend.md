---
id: 000261
status: open
deps: [ariadne#253]
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '9041d452009122935d31a4a71cf03726aa521d0e' # card fields mirrored from issue-cards; edit via sdlc
---

# Stateful git fake behind the consolidated seam; run e2e tests against both backends

## Problem

ARCH-MOCK asks for a stateful fake behind the same seam as every external
binary, with a live conformance check. For `git`, only the fleet collectors
have one (`cmd/sdlc/internal/fleet/fakegit_test.go`, read-only, with
`git_conformance_test.go`). The verbs' write paths (commit, push, fetch,
branch, worktree, merge) have stateless capture stubs only, so their tests
drive real git and dominate the suite's wall time (#253's timing report).

## Spec

- Grow a write-capable stateful git fake behind the seam #253 consolidates
  (one directory-explicit git runner), starting from fleet's `FakeGit`.
- Write each converted test body once against a backend parameter:
  `go test -short` runs it on the fake; the full run uses real git, and that
  run is the fake's live conformance check.
- Convert incrementally, verb-level tests first — tests of sdlc's own
  decisions (what it commits, which refs it moves). Tests of git's behavior
  (exact bytes, merge convergence, status parsing, ref ordering; see the #252
  and #259 lessons) stay real-git only.

## Done when

- A shared fake models the git surface the converted tests consume, and the
  same test bodies pass against both the fake and real git.
- The full run fails when the fake diverges from real git on a converted test.

## Plan

- [ ] Inventory the git surface the verb-level e2e tests consume (after #253's seam)
- [ ] Fake + two-backend test runner
- [ ] Convert the first verb family; measure the -short tier gain

## Log

### 2026-09-28

- Split out of #253 by the operator: #253 makes the real-git tests fast
  (tiers, shared binary, parallel-safe seams); the fake is this issue.
