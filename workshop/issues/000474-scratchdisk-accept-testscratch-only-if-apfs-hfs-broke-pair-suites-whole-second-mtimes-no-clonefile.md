---
id: 000474
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c3ffdedd62c5b8927f38a2687c554879af0657f8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T15:12:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] Add red fixtures for APFS acceptance, HFS/unknown refusal with a reason,
  and fallback placement while preserving rejected roots for cleanup.
- [ ] Read Darwin statfs filesystem names and enforce APFS through Choose,
  including explicit overrides; keep non-Darwin fixture behavior unchanged.
- [ ] Update AGENTS.base.md and the atlas; run focused tests and the recorded
  suite, close on SHIP, open the PR and land through the landing line.

ARCH-PURE/ARCH-DRY: Read supplies one filesystem fact; Choose remains the one
admission rule shared by suite and review roots. The existing statfs call
provides it in constant time without capability-test writes or subprocesses.
No new persisted state or background work. Absent/off stays silent; unreadable,
non-APFS on Darwin, and low-space roots fall back with one existing note.
Actions: run normally (fallback), or have the operator fix the mount and retry;
no test data is moved or lost. Cleanup still scans previously used roots.
A lock cannot improve a filesystem-format check: this process does not own the
operator's mount lifetime. Live validation is read-only statfs of TestScratch,
plus a disposable local harness; no shared staging or volume changes.

## Log

### 2026-10-19

### 2026-10-19 — approved scope

Hub confirmed Darwin-only APFS enforcement, overrides included. Linux CI uses
explicit temporary fixture roots and never mounts TestScratch; retain that
behavior. Prefer statfs to capability writes. The issue is a filesystem
precondition correction, not a new persisted mechanism or a general lintable
API misuse.

## Revisions

### 2026-10-19 15:13 PDT — platform boundary

Use the Spec's APFS statfs option on Darwin. Non-Darwin overrides retain their
existing behavior, as approved by the hub; no capability probes or mount edits.
