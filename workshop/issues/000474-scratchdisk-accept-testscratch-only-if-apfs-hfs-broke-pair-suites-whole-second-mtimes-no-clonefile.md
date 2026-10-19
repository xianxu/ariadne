---
id: 000474
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b06208ad139fa95dbb362c8ede1bc71c14c5f571' # card fields mirrored from issue-cards; edit via sdlc
---

# scratchdisk: accept TestScratch only if APFS (HFS+ broke pair suites: whole-second mtimes, no clonefile)

## Problem

2026-10-19: with ariadne#468, recorded suites run on `/Volumes/TestScratch`, which the ops login item had formatted HFS+.
HFS+ keeps whole-second timestamps and has no clonefile, so pair's suites failed on every landing: same-size rewrites with a
restored mtime went undetected (pair:6 A/B on one head: passes with TMPDIR=/private/tmp, fails on TestScratch), pid-mtime
bounds failed, and clone-based write budgets blew up (couchcore 1.9 GB vs a 128 MB cap). pair:7 measured 20/20 undetected
rewrites and parent churn on TestScratch vs 0/20 on /private/tmp. The ops script now formats APFS (probed: ns timestamps,
clonefile work); but sdlc accepted the HFS+ volume silently.

## Spec

The scratch-disk admission (`cmd/sdlc/internal/scratchdisk`) accepts `/Volumes/TestScratch` only when its file system has the
semantics the suites rely on: APFS (check `statfs` f_fstypename), or more generally nanosecond timestamps and clonefile
support probed once per run (write a file, check mtime has a non-zero fractional part across two writes; clonefile a file).
Otherwise fall back to today's temp folder with one visible note naming the reason ("TestScratch is hfs: no ns timestamps").

## Done when

- Fixture: an HFS+ (or probe-failing) scratch root is refused with the note and the suite falls back; APFS is accepted.
- The AGENTS.base.md section 10 line says the RAM disk must be APFS.

## Plan

- [ ]

## Log

### 2026-10-19
