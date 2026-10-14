---
id: 000358
status: working
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '43fa88ad972a083a2969cdb1b5bfab87cd01ed07' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T18:04:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "22db63ab", done: "129ba1c5"}
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

- [x] Pin clean and inherited keep-going environments in the real make failure regression; verify RED.
- [x] Add -S to the shared tools argv and align the process double; verify targeted GREEN and the revised acceptance run below.
- [x] Record acceptance evidence, then enter close; land after SHIP.

## Log

### 2026-10-13
- 2026-10-13: unclaimed: filed from #353; independent weave fix, unclaimed for the hub to assign

Filed by ariadne:3 from #353's full suite (unrelated to #353's diff).

## Revisions

- 2026-10-14 — Hub assigned the production fix: cancel inherited keep-going with -S at the existing tools IO seam (ARCH-PURPOSE, ARCH-PURE). Preserve other make options, including job settings and explicitly selected ignore-error/dry-run modes; removing the whole environment is broader than this bug. Reuse the current real double-colon fixture and injected process double (ARCH-DRY, ARCH-MOCK). No new state, concurrency, artifacts or input surface is introduced (ARCH-ORDER, ARCH-FUNERAL, ARCH-CONSTRAINTS, ARCH-SECURE). #356 is superseded by this production correction.

- 2026-10-14 — RED: TestToolsDoubleColonFailurePropagates/MAKEFLAGS=k fails with later recipe ran (0.352s), while the explicit clean environment passes. GREEN after -S: MAKEFLAGS=k go test ./cmd/weave/internal/startup -count=1 passes (1.948s), including real Make and the existing stateful process double. The only production change is the tools argv; no architecture surface changed. Full make -k test next.

- 2026-10-14 — Acceptance revision authorized by hub receipt bada6a90-b130-4909-b7cf-868d8b8cc798: replace the full-suite criterion with all weave packages, startup under both MAKEFLAGS=k and empty, and green PR CI. Full suite starved at the gate (ariadne#357); change confined to weave's tools argv. Cancelled the queued make and its own waiter after over 40 minutes, before any suite tests ran. No bypass of the machine lock.
- 2026-10-14 — Revised acceptance passed at 865d3556d486dd7139125122d7fd597983b0a215: go test ./cmd/weave/... -count=1 (13 tested packages pass, one with no tests); startup MAKEFLAGS=k 0.830s and MAKEFLAGS empty 0.462s. Full weave log /tmp/ariadne358-weave-20261013.log; cancelled queue log /tmp/ariadne358-full-20261013.log. No architecture surface changed; this restores the existing fail-stop contract. Close and PR CI remain subsequent gates.
