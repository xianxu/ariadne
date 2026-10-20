---
id: 000477
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b40316cdad8f84c482b2b06592b7bf4b5b7080a7' # card fields mirrored from issue-cards; edit via sdlc
---

# Scratch fallback must override an inherited unwritable GOTMPDIR/TMPDIR (sandboxed go fails since pair#708 set them to the RAM disk)

## Problem

Since the 10-19 reboot, pair#708 gives every slot shell `GOTMPDIR=/Volumes/TestScratch/tmp/<repo>-slotN` (and the
RAM-disk temp folder), but the agent sandboxes deny writes to `/Volumes/TestScratch`. Every sandboxed `go build` /
`go test` fails with "go: creating work dir ... operation not permitted" (ops:0 reproduced: `touch` into its own
GOTMPDIR is denied). ariadne:4's sandboxed `sdlc merge` for #456 saw its integration suite and the main baseline both
fail in 0 s, so main looked red when it wasn't.

`prepareScratch` (cmd/sdlc/testscratch.go) already falls back to the normal temp folder when the scratch root isn't
writable, but on fallback it returns no env, so the child inherits the unwritable `GOTMPDIR` (and `TMPDIR` if
inherited) from the slot shell and fails anyway.

## Spec

- On fallback, `prepareScratch` returns env entries that point `TMPDIR` and `GOTMPDIR` at a writable normal temp
  folder, overriding inherited values, instead of returning none.
- More generally, before a suite or baseline runs, an inherited `GOTMPDIR`/`TMPDIR` that isn't writable is replaced
  (with a one-line note naming it), so an environment mismatch never shows up as a red suite in 0 s.
- A suite whose every package fails at "creating work dir" is classified as an environment failure, not a red main
  (merge says so and names the variable).

## Done when

- Test: inherited `GOTMPDIR` pointing at an unwritable dir + unwritable scratch root → the child runs with a writable
  `GOTMPDIR` and the suite runs.
- Test: an all-packages "creating work dir" failure is reported as an environment failure naming GOTMPDIR.

## Plan

- [ ]

## Log

- The real fix for the sandbox side is the operator adding `/Volumes/TestScratch` to the agents' sandbox write
  allowlist (a local-gates change, Needs you since #468). This issue makes sdlc robust either way.
