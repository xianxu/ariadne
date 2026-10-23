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

### 2026-10-22 — quick implementation and focused verification

Used systematic debugging and TDD. The first replay fixture incorrectly used shell quoting as Python source; corrected the fixture before treating its result as a RED control. The corrected real-capture regression then failed in0.740s with actual exit2 and lost stderr/path. The production boundary now formats only explicit argparse command/option incompatibility as unsupported inspection, with an actionable supplier path; other exit2/runtime failures retain their own bounded diagnostics. It adds no helper probe or scheduler substitution and changes no admission path. Existing status process-group ownership, 3-second enclosing budget, output cap and frozen JSON remain intact.

All TestSuiteQueue controls passed50.636s, including the real priority helper, provider descendants, cancellation, status overlap/join, malformed/overflow refusal, captured errors, 1024-byte excerpt controls and an acquire-only supplier exercised through acquireSuiteSlot with unchanged --watch/--slots argv. The earlier real older-helper private probes from Pair and an older Ariadne branch independently confirmed acquisition and native release. No live queue, waiter or merge was changed. README/atlas explain recovery; the committed live stderr fixture defends the semantic compatibility boundary, which a lexical lint rule cannot infer from an exit code. Final race/catalog/private production-CLI checks and recorded suite remain pending.

Focused compatibility/JSON/admission race checks passed1.985s; Go lint1.077s and shell lint0.391s passed. Code change is22 added production lines. Checkpoint before building the private production CLI and queueing recorded verification.

### 2026-10-22 — production CLI conformance complete; prepare recorded suite

Built a private sdlc from the clean pushed code checkpoint using Go1.27.3; build metadata confirmed revision4d1550c8f2afeb3b06ed9dbe6604d90e4359bd4e and modified=false. Read-only queue status from Pair-slot1, Parley-slot1 and older Ariadne-slot2 all exited0, retained the clean main queue snapshot and printed the new explicit unsupported-inspection message with the correct actual supplier path and real argparse stderr. The same binary from current Ariadne-slot3 exited0 and displayed available ordered tickets. The private-binary fleet-path warning was expected; all lifecycle commands continue through central sdlc. No staging, rebuild of :0, peer edit or live queue mutation occurred.

Private CLI: /Volumes/TestScratch/tmp/ariadne-slot3/ariadne563-cli.Od5td0/sdlc. This worker removes that one diagnostic directory after acceptance; it is not a production artifact family. Captured producer stderr is committed and replayed in normal tests. The diagnostic depends on no agent identity or screen detector and adds no Couch harness/protocol surface.

Next: queue the required recorded suite on the pushed branch; this repository has no .sdlc/affected.json, so the full named suite is required. Close review may overlap its queue under tl-worker, but a final-code recorded pass must precede landing. Keep the queue/controller through waits and restarts; no priority or staging is requested for this status-only correction.
