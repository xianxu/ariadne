---
id: 000360
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '4cd511df0921ae717b977f58e8035b5e81e856c6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T18:37:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ]

## Log

### 2026-10-13

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).
