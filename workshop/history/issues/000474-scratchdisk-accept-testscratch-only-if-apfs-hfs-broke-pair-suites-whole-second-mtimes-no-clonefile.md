---
id: 000474
status: done
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: '2aa55b8a5eacd47dee18068917eebd3dade439dc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-19T15:12:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "77f31c55", done: "a94496f9"}
actual_hours: 0.26
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

- [x] Add red fixtures for APFS acceptance, HFS/unknown refusal with a reason,
  and fallback placement while preserving rejected roots for cleanup.
- [x] Read Darwin statfs filesystem names and enforce APFS through Choose,
  including explicit overrides; keep non-Darwin fixture behavior unchanged.
- [x] Update AGENTS.base.md and the atlas; run focused tests and the recorded
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
- 2026-10-19: closed — Red: HFS/unknown Choose fixtures admitted silently; green: APFS accepted and Darwin HFS/unknown rejected with reason. Focused scratchdisk/reviewbox/suite/review caller tests passed on APFS-backed normal temp and HFS-backed owned temp; HFS suite preparation inherits temp, prints one note and creates no scratch folder. Read-only live statfs observed TestScratch hfs and rejection. Linux/amd64 scratchdisk test compilation passed. Central sdlc recorded suite on 28837330d0a076e2d831c39d82ed8e854d384fee passed 12 shards / 1721 SDLC tests and 72 other packages, zero failures/retries in 10m03s, digest dd051507e069. AGENTS.base section 10 and atlas updated; no mount changes, deadline bumps, shared staging or private gate binary. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

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

### 2026-10-19 — implementation and focused evidence

Red: after adding the probe's filesystem fact, TestChoose rejected neither HFS
nor unknown types (use=true, empty note). Green: APFS accepted; HFS/unknown
return the normal-folder note through the same Choose rule. Native statfs
reported the live TestScratch as hfs and visibly rejected it, without writes.
The caller-level HFS fixture verifies no child env override, one warning and
no scratch subdirectory creation. Cleanup retains the rejected root.
Focused scratchdisk, reviewbox and suite/review caller tests passed both with
ordinary temp and with TMPDIR on an owned HFS fixture. Existing fixture
expectations now follow the actual filesystem's admission, preserving both
acceptance and fallback coverage. Linux/amd64 scratchdisk test compilation
passed. No timeout changes or mount operations. A private slot binary was
built but not used for gates; per hub direction all recorded checks, close and
merge use the normal central sdlc.

### 2026-10-19 — recorded suite

The central sdlc recorded suite passed on implementation head
28837330d0a076e2d831c39d82ed8e854d384fee: 12 shards / 1,721 SDLC tests and
72 other packages, zero failures, no rerun, 602.5 seconds (10m03s). Attempt
a2de36e136044751012eec0d172b26fa; saved log digest dd051507e069. The suite
ran on the existing HFS TestScratch through the normal central binary; the new
guard's HFS fallback fixtures passed there too. PR #331 CI is green.
