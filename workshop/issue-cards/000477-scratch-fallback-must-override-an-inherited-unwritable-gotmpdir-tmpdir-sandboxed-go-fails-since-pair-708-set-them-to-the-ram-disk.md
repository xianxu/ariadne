---
id: 000477
status: open
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T17:59:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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
