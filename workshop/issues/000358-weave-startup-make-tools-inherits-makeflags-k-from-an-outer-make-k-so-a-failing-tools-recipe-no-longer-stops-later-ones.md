---
id: 000358
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: 'eb91e918117bc032be2ee2fefe2ca52f84062e36' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T18:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# weave: startup make tools inherits MAKEFLAGS=k from an outer make -k, so a failing tools recipe no longer stops later ones

## Problem

`startup.Tools` (`cmd/weave/internal/startup/tools.go:35`) runs the owner's
`make --no-print-directory -f Makefile -f - tools` with the inherited environment, so a
`MAKEFLAGS` from an outer make is honored. Under an outer `make -k` (MAKEFLAGS=k), a
failing `tools::` recipe no longer stops the later ones: weave's "a tools failure stops
the build" contract silently changes with how the caller was invoked.

Found during #353's full suite, which ran as `make -k test` (the TL's 2026-10-13 rule
for full suites): `TestToolsDoubleColonFailurePropagates` fails with "later recipe ran".
Deterministic: `MAKEFLAGS=k go test ./cmd/weave/internal/startup -run
TestToolsDoubleColonFailurePropagates` fails every run; without it, it passes (`MAKEFLAGS=j4`
passes too). So every slot that runs `make -k test` sees this failure.

## Spec

Pass `-S` (`--no-keep-going`, cancels an inherited `-k`) in the tools argv, so the
contract holds whatever the caller's MAKEFLAGS; decide whether any other inherited make
flag (e.g. `-i`, `-n`) also needs neutralizing. Add a regression test that sets
`MAKEFLAGS=k` itself and asserts the failure still stops later recipes.

## Done when

- `MAKEFLAGS=k go test ./cmd/weave/internal/startup` passes, and the new test fails when
  `-S` is removed.
- `make -k test` is green.

## Plan

- [ ]

## Log

### 2026-10-13

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).
