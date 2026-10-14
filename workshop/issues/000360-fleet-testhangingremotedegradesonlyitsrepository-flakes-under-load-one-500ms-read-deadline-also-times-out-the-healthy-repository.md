---
id: 000360
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '6120f55a74c2252ad2cc2e0060ac7969c5760369' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T18:37:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "c2e45336", done: "422b95dc"}
---

# fleet: TestHangingRemoteDegradesOnlyItsRepository flakes under load; one 500ms read deadline also times out the healthy repository

## Problem

`cmd/sdlc/internal/fleet` `TestHangingRemoteDegradesOnlyItsRepository`
(`recordsdeadline_test.go:102`) failed in #353's full suite (2026-10-13, after a 1376s
queue, under load) with:

    healthy repository: unknown "…/healthy has a fetched issue tracker but its publication
    target is unusable: invalid resting branch: context deadline exceeded: "

It passes alone (3/3, 4s). `shortDeadline(t, 500*time.Millisecond)` sets the one
`recordsReadDeadline` that governs both repositories' reads: the hanging one needs it to
expire, but under load the healthy one's several git reads also exceed 500ms.

## Spec

Keep the property the test proves (one dead remote degrades only itself, within a
bound) without making the healthy repository race a 500ms budget: the stuck remote hangs
indefinitely, so a deadline with real headroom (a few seconds) still expires for it,
with the elapsed bound set from that deadline. If production should give each
repository its own deadline rather than one shared budget, decide that here too.

## Done when

- The test passes 50x under concurrent load (for example alongside a full suite), and
  still fails if the stuck repository's deadline handling is removed.

## Plan

- [x] Root cause: `recordsReadDeadline` wrapped the whole repository read
  (`context.WithTimeout` in `loadRepoRecords`), and that context is bound into the
  `tracker.Repository`/`gitx.TrunkFile`, so it governed every git call, local and
  remote. Its purpose is "a remote that does not answer", yet under load the healthy
  repository's local work (`workspace.Resolve`, `ResolvePublicationTarget`) spent it
  ("invalid resting branch: context deadline exceeded").
- [x] Design fix: `gitx.TrunkFile.BoundRemote(d)` (via `tracker.Repository.BoundRemote`)
  is a remote budget, applied only to the five remote-contacting commands (fetch, two
  ls-remote, two push) through `runRemote`, shared across them from the first one (a
  hang cannot multiply it per call). Local work runs on the caller's context. Fleet sets
  `recordsReadDeadline` as that budget instead of wrapping the read.
- [x] Consequence: a hanging remote now degrades like any unreachable remote under
  `PreferFresh` (stale, answered from its last fetch, naming the deadline) instead of
  unknown, because the read survives to its local fallback. The fleet test expects that.
- [x] Deterministic tests in `gitx` (latency injected through `runGitInContext`): slow
  local work does not spend the budget, while the old whole-operation deadline fails the
  same read; a hanging remote is bounded and named; the budget is shared.
- [x] Prove the fleet test under real load (in a loop alongside a full suite).

## Log

### 2026-10-13
- 2026-10-13: unclaimed: filed from #353; independent flake, unclaimed for the hub to assign

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).

### 2026-10-13 — ariadne:3

Under real load: test binaries built from origin/main and from this branch ran the fleet
test alternately while a full suite ran (load 30 rising to 90): **origin/main failed 26 of
100, this branch 0 of 100.** Two stresses that failed to reproduce, for the record: a
CPU-only busy loop (load ~10, 0/20), and a slow `git` wrapper on PATH (the test harness
resolves git's absolute path once, so the wrapper saw a single call). The deterministic
proof is in `gitx` through the `runGitInContext` seam instead.

That suite run caught my own `TestRemoteBudgetIsSharedAcrossRemoteCommands` failing under
load (its first remote call plus a 150ms injected delay exceeded a 200ms budget: the same
tight-margin fragility this issue fixes). Reworked both timing tests to rely on wide
margins rather than sums: local work 500ms per command against a 1.5s budget, and for the
shared budget, the first real call gets the whole 2s and the test then waits past the
budget's end, so the second call fails on the shared deadline however fast it is. The
shared test still fails with the budget disabled.
