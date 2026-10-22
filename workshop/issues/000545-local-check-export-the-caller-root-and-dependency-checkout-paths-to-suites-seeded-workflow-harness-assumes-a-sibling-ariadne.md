---
id: 000545
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'fbf71c644057ee42da6e4b7a3a4c140b26e4ac24' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T00:00:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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
- Every managed command path (ordinary check, runner, affected selection and rerun batch) receives the same caller/dependency environment. Inherited reserved variables cannot supply stale dependency paths.
- Unique substrate/data kinds export absolute paths; identical resolved paths dedupe. Conflicting same-kind rows refuse before command execution, name both rows/paths and tell the caller to repair the declaration.
- A detached consumer-wrapper fixture proves explicit substrate selection without a sibling, direct-checkout fallback, and a clear `substrate not found` skip; dependency bytes remain unchanged. echo#10 owns the actual echo wrapper change.

## Plan

- [ ] Add red-before production runner fixtures for caller identity (including WorkflowAnchor), declared substrate/data paths, duplicate/conflicting rows, absent/malformed declarations and stale inherited variables; replay the wrapper selection contract from a detached checkout with no substrate sibling.
- [ ] Add one shared environment builder before managed command admission: read the tested checkout's bounded construct/deps with layergraph.ReadDeclaration/ParseRows, resolve substrate rows through the existing layergraph resolver against callerRoot (including its established primary fallback), resolve data mounts against callerRoot, and replace inherited SDLC_CALLER_ROOT/SDLC_DEP_* values. Reuse it for all runUnderResourceBatch callers; command cwd remains the exact tested tree. Export absent declared paths for the consumer to diagnose; never acquire or write dependencies.
- [ ] Document the contract and read-only dependency obligation in local-check help, README and atlas. Run targeted native runner/parser/wrapper controls, then the recorded full suite, close and land. Ask the hub whether the isolated detached production-path proof suffices for the live gate.

Design: this is per-run environment preparation, not new durable state. Reuse callerRoot and layergraph's parser/resolver (ARCH-DRY); keep kind/path collision and reserved-variable replacement deterministic, with filesystem/Git reads in the thin integration boundary (ARCH-PURE). Declarations come from the tested tree so uncommitted caller edits do not redefine its rows; relative paths are anchored at the initiating checkout, never the disposable run tree. A missing declaration exports only caller identity, and malformed/unreadable declarations refuse before launching or recording a test result. No dependency mutation or automatic acquisition. The existing runner owns process/cancellation/cleanup lifetime; these values die with its command environment (ARCH-ORDER, ARCH-FUNERAL). One bounded declaration read and one pass over its rows, with existing memoized primary lookup; no polling or new hot path. No persisted format changes (ARCH-EVOLVE).

Actions: run check, repair dependency declaration, repair missing substrate, retry check.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| No declaration, or valid unique/deduped rows | Command receives exact caller and declared absolute dependency paths | Run check | None |
| Invalid/unreadable declaration or conflicting same-kind paths | Preparation refusal naming declaration/rows; command never starts | Repair declaration, retry check | None |
| Declared substrate absent | Path exported; wrapper says `substrate not found` and skips its substrate-specific proof | Repair missing substrate, retry check | No substrate proof; no success claim for skipped coverage |

Covered: all shared managed command entry paths and the detached wrapper contract. Not covered: dependency version pinning/acquisition, schema changes, transitive environment exports, or echo's independently owned wrapper implementation. Stateful native Git/filesystem fixtures cover resolution and detached execution; no new external service seam or live fleet mutations (ARCH-MOCK, ARCH-SECURE). Shell wrappers quote paths; no shell-source construction from dependency content.

## Log

### 2026-10-21

### 2026-10-22

- Claimed and started from fresh main after #537 landed and cleaned up. Existing callerRoot already preserves WorkflowAnchor through disposable train trees; runUnderResourceBatch is shared by ordinary checks, resource runners and reruns. Ariadne's portable-ci.test.sh is self-rooted; the sibling assumption is in echo's wrapper, not an Ariadne-seeded test. Hub confirmed runner/docs/detached-fixture scope here, echo#10 owns its wrapper. Hub also approved one variable per unique kind, identical-path deduplication and loud conflicting-row refusal.

## Revisions

- 2026-10-22 00:03 PDT — Hub clarified the original harness assumption: Ariadne owns exports/docs and a detached replay of the wrapper contract; echo#10 owns env/fallback/skip in echo. For multiple same-kind dependency rows, identical resolved paths dedupe and conflicting paths refuse with both rows named. This supersedes the original Spec's missing-substrate failure wording for the wrapper: use the dispatched clear skip instead.
