---
id: 000545
status: open
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: '7315f950a65322dec3b163d829d39fd6418140bf' # card fields mirrored from issue-cards; edit via sdlc
---

# local-check: export the caller root and dependency checkout paths to suites (seeded workflow harness assumes a sibling ../ariadne)

## Problem

## Spec

From ops-echo:0 (10-21): echo's recorded suite fails before product code because echo#1's
`tests/ci-workflow.test.sh` (adopted from ariadne's upstream workflow harness) reads `../ariadne` beside the
checkout. `sdlc local-check` runs the suite in a fresh worktree under `/Volumes/TestScratch/local-checks/sdlc-local-N`,
which has no sibling, and nothing tells a suite where its dependency checkouts are. The only related variable today is
the incidental `WF_TEST_LOCK_OWNER` (the caller's root), which isn't a contract. ariadne's own copy of the harness
likely has the same sibling assumption for every consumer.

Fix: local-check exports a documented contract to the suite: `SDLC_CALLER_ROOT` (the initiating checkout) and, per
`construct/deps` row, `SDLC_DEP_<NAME>` resolved against the caller root (e.g. `SDLC_DEP_SUBSTRATE`, the ariadne
checkout). The upstream harness uses `SDLC_DEP_SUBSTRATE`, falling back to `../ariadne` for direct `make test`, and
fails with a clear message when neither exists. Never write to a dependency checkout from a suite.

## Done when

- A local-check run of a fixture suite sees `SDLC_CALLER_ROOT` and `SDLC_DEP_SUBSTRATE` pointing at the right paths (test).
- ariadne's seeded workflow harness finds the substrate in a local-check worktree with no sibling (test), and still
  works from a normal checkout.
- Documented in the local-check help.

## Plan

- [ ]

## Log

### 2026-10-21
