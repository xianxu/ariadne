---
id: 000563
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f3e17e85086eede893e130d0dee360cb975c7168' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T21:59:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "13dad640", done: "14e17f19"}
---

# Queue status must explain older supplier helper incompatibility

## Problem

After #560, central sdlc queue status invokes inspect --supervised on the helper selected by the caller Makefile.workflow supplier. Older Pair/Parley pinned substrates and pre-#560 Ariadne issue branches have an acquire-only helper, which argparse rejects with exit2. The status display hides stderr and suggests an ineffective retry.

## Spec

Keep selecting the workflow supplier shared with admission; do not silently substitute a newer scheduler that could describe policy the old scheduler does not enforce. Gracefully identify the captured acquire-only inspection incompatibility, name the actual helper path and the supplying-checkout update recovery, and preserve bounded stderr/path for other helper failures. This is optional text-order diagnostics only. Frozen queue status JSON and acquire admission stay unchanged.

## Done when

- Replayed real pre-#560 argparse stderr produces an explicit unsupported-inspection diagnostic naming the selected helper and an update recovery, rather than an unexplained exit2/retry.
- Generic inspect failures retain bounded stderr and helper identity, and never invent available order or classify unrelated exit2 as version skew.
- Supported inspection, output caps and owned cancellation/descendant cleanup remain passing; frozen status JSON does not query supplemental inspection.
- Private real older helpers from Pair and an older Ariadne branch acquire and release isolated suite slots successfully; no live queue or waiter is changed.

## Plan

- Preserve the live capture as a fixture and write RED diagnostic tests through the actual observation seam.
- Add a small pure diagnostic formatter for the recognized old command/option refusal, reusing current bounded subprocess capture and supplier resolution. Keep admission untouched.
- Run focused normal/race and private current/old helper conformance, document the recovery, then recorded affected check, close/PR/landing through central sdlc.

## Log

### 2026-10-22 — diagnosis before code

Reproduced from pair-slot1 using ariadne:0 sdlc. Selected helper is pair-slot1/ariadne/scripts/test-lock.py, reached via Makefile.workflow symlink. Its stderr: usage: test-lock.py [-h] --watch WATCH [--slots SLOTS] {acquire}; error: argument cmd: invalid choice: inspect (choose from acquire). ariadne-slot2 gives the same result. Explicit central helper inspect succeeds from the identical Pair cwd, excluding cwd/weight/declaration as the cause. Source observation invokes inspect --supervised; acquireSuiteSlot separately invokes acquire --watch PID --slots WEIGHT. Both old helpers acquired and released slots in private WF_TEST_LOCK_FILE/WF_SDLC_QUEUE_ROOT paths with exit0/acquired; native release confirmed. No live admission or merge was driven. Cause reported to ops:0 before implementation.
