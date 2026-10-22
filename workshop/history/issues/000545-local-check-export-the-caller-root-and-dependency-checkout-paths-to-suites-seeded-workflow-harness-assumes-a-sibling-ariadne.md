---
id: 000545
status: done
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9272d97ce4a7bfa7451df95b3f26e5aa8760badb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T00:00:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "b19164e0", done: "0abe5136"}
actual_hours: 0.50
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

- [x] Add red-before production runner fixtures for caller identity (including WorkflowAnchor), declared substrate/data paths, duplicate/conflicting rows, absent/malformed declarations and stale inherited variables; replay the wrapper selection contract from a detached checkout with no substrate sibling.
- [x] Add one shared environment builder before managed command admission: read the tested checkout's bounded construct/deps with layergraph.ReadDeclaration/ParseRows, resolve substrate rows through the existing layergraph resolver against callerRoot (including its established primary fallback), resolve data mounts against callerRoot, and replace inherited SDLC_CALLER_ROOT/SDLC_DEP_* values. Reuse it for all runUnderResourceBatch callers; command cwd remains the exact tested tree. Export absent declared paths for the consumer to diagnose; never acquire or write dependencies.
- [x] Document the contract and read-only dependency obligation in local-check help, README and atlas. Run targeted native runner/parser/wrapper controls, then the recorded full suite, close and land. Ask the hub whether the isolated detached production-path proof suffices for the live gate.

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
- 2026-10-22: closed — Native regressions fail before and pass after caller/dependency export for suite and runner resources, stale inherited values, duplicate/conflicting/malformed rows, detached wrapper selection and missing-path skip, direct fallback, primary checkout fallback and shared command batches. Broader native runner, exact-package flake and recovery checks passed; layergraph, golint/shelllint, actual portable CI harness and compiled help passed. Recorded suite lease-177733865 on pushed fb664c301c6d1e78033bece283658f71cb750881 passed 2096 SDLC tests plus 75 packages, zero failures, 867.9s, log 9f4c249d0935; server checks green. Hub approved isolated detached production-path proof, no :0 stage; echo#10 owns its wrapper change.; review verdict: SHIP

- Claimed and started from fresh main after #537 landed and cleaned up. Existing callerRoot already preserves WorkflowAnchor through disposable train trees; runUnderResourceBatch is shared by ordinary checks, resource runners and reruns. Ariadne's portable-ci.test.sh is self-rooted; the sibling assumption is in echo's wrapper, not an Ariadne-seeded test. Hub confirmed runner/docs/detached-fixture scope here, echo#10 owns its wrapper. Hub also approved one variable per unique kind, identical-path deduplication and loud conflicting-row refusal.
- Hub approved the full checkpointed plan and isolated production-path detached execution as sufficient live proof; no :0 stage. After landing, report to ops-echo:0 so echo#10 exercises its wrapper under a real recorded check.
- RED: native suite and runner fixtures inherited stale caller/substrate/data values and undeclared reserved keys; conflicting substrate/data and malformed rows launched and recorded commands; detached wrapper skipped because its sibling substrate was absent. GREEN after one shared builder before queue/command admission: tested declaration, initiating caller/WorkflowAnchor paths, shared resolver primary fallback, same-kind deduplication/refusal and reserved-variable replacement. No dependency acquisition or writes. The wrapper fixture proves explicit selection without a sibling, missing declared path skip, direct fallback, and unchanged dependency sentinel bytes; echo remains untouched.
- Targeted proof: first corrected runner group passed (6.665s), expanded environment group passed (2.479s), then broader native local-check, exact-package known-flake and recovery controls passed (42.430s). One added fallback assertion initially required a nonexistent trailing newline on the final environment entry; corrected the assertion, not production behavior. Layergraph, golint and shelllint packages passed (3.273s / 1.861s / 0.523s). Actual portable-ci harness passed all workflow-block controls, and compiled local-check help printed the new environment contract. Diff check clean. Preparing the pushed head for the required recorded full suite.
- The invariant is path provenance across caller/tested trees, not a lexical ban on sibling fallback (which is explicitly retained for direct wrapper runs). Native detached/primary-fallback/batch regressions and the existing bounded declaration parser defend this semantic contract; adding a helper-name lint would not prove it (ARCH-PURPOSE).
- Recorded full suite PASS on pushed fb664c301c6d1e78033bece283658f71cb750881: lease-177733865 retained through 643s FIFO wait, 2,096 SDLC tests and 75 other packages, zero failures, 867.9s (14m28s), log sha256 prefix 9f4c249d0935. Saved under .git/sdlc/local-checks/fb664c301c6d1e78033bece283658f71cb750881/suite.log. No flake retry/resume needed. GitHub merge-check and local-gates checks passed on the same head. Preparing mandatory close review without code changes or manual main integration.
- First close preflight refused the remaining unchecked documentation/verification row before review dispatch. Its documentation, targeted/full verification and hub live-proof decision are complete; marked that implementation row complete for the close boundary. Close review and landing remain the next lifecycle actions, not yet claimed complete. No gate bypass or code change.

## Revisions

- 2026-10-22 00:03 PDT — Hub clarified the original harness assumption: Ariadne owns exports/docs and a detached replay of the wrapper contract; echo#10 owns env/fallback/skip in echo. For multiple same-kind dependency rows, identical resolved paths dedupe and conflicting paths refuse with both rows named. This supersedes the original Spec's missing-substrate failure wording for the wrapper: use the dispatched clear skip instead.
