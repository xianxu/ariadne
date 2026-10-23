---
id: 000563
status: done
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'cff633d1b982da7209281c417b9565cdbcc02e05' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T21:59:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "13dad640", done: "14e17f19"}
actual_hours: 0.13
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
- 2026-10-22: closed — Captured real argparse stderr RED then GREEN; TestSuiteQueue passed 50.636s, focused race 1.985s, Go/shell lint pass. Private production CLI confirms graceful old-supplier diagnostics from Pair/Parley/older Ariadne and available order on current Ariadne; private old-helper acquire/release succeeds. Central recorded final-code suite is queued and must pass before landing (tl-worker permitted overlap).; review verdict: SHIP

Reproduced from pair-slot1 using ariadne:0 sdlc. Selected helper is pair-slot1/ariadne/scripts/test-lock.py, reached via Makefile.workflow symlink. Its stderr: usage: test-lock.py [-h] --watch WATCH [--slots SLOTS] {acquire}; error: argument cmd: invalid choice: inspect (choose from acquire). ariadne-slot2 gives the same result. Explicit central helper inspect succeeds from the identical Pair cwd, excluding cwd/weight/declaration as the cause. Source observation invokes inspect --supervised; acquireSuiteSlot separately invokes acquire --watch PID --slots WEIGHT. Both old helpers acquired and released slots in private WF_TEST_LOCK_FILE/WF_SDLC_QUEUE_ROOT paths with exit0/acquired; native release confirmed. No live admission or merge was driven. Cause reported to ops:0 before implementation.

### 2026-10-22 — quick implementation and focused verification

Used systematic debugging and TDD. The first replay fixture incorrectly used shell quoting as Python source; corrected the fixture before treating its result as a RED control. The corrected real-capture regression then failed in0.740s with actual exit2 and lost stderr/path. The production boundary now formats only explicit argparse command/option incompatibility as unsupported inspection, with an actionable supplier path; other exit2/runtime failures retain their own bounded diagnostics. It adds no helper probe or scheduler substitution and changes no admission path. Existing status process-group ownership, 3-second enclosing budget, output cap and frozen JSON remain intact.

All TestSuiteQueue controls passed50.636s, including the real priority helper, provider descendants, cancellation, status overlap/join, malformed/overflow refusal, captured errors, 1024-byte excerpt controls and an acquire-only supplier exercised through acquireSuiteSlot with unchanged --watch/--slots argv. The earlier real older-helper private probes from Pair and an older Ariadne branch independently confirmed acquisition and native release. No live queue, waiter or merge was changed. README/atlas explain recovery; the committed live stderr fixture defends the semantic compatibility boundary, which a lexical lint rule cannot infer from an exit code. Final race/catalog/private production-CLI checks and recorded suite remain pending.

Focused compatibility/JSON/admission race checks passed1.985s; Go lint1.077s and shell lint0.391s passed. Code change is22 added production lines. Checkpoint before building the private production CLI and queueing recorded verification.

### 2026-10-22 — production CLI conformance complete; prepare recorded suite

Built a private sdlc from the clean pushed code checkpoint using Go1.27.3; build metadata confirmed revision4d1550c8f2afeb3b06ed9dbe6604d90e4359bd4e and modified=false. Read-only queue status from Pair-slot1, Parley-slot1 and older Ariadne-slot2 all exited0, retained the clean main queue snapshot and printed the new explicit unsupported-inspection message with the correct actual supplier path and real argparse stderr. The same binary from current Ariadne-slot3 exited0 and displayed available ordered tickets. The private-binary fleet-path warning was expected; all lifecycle commands continue through central sdlc. No staging, rebuild of :0, peer edit or live queue mutation occurred.

Private CLI: /Volumes/TestScratch/tmp/ariadne-slot3/ariadne563-cli.Od5td0/sdlc. This worker removes that one diagnostic directory after acceptance; it is not a production artifact family. Captured producer stderr is committed and replayed in normal tests. The diagnostic depends on no agent identity or screen detector and adds no Couch harness/protocol surface.

Next: queue the required recorded suite on the pushed branch; this repository has no .sdlc/affected.json, so the full named suite is required. Close review may overlap its queue under tl-worker, but a final-code recorded pass must precede landing. Keep the queue/controller through waits and restarts; no priority or staging is requested for this status-only correction.

### 2026-10-22 — first-round SHIP; recorded suite retained

Close review returned SHIP with no findings or revisions. Reviewer independently ran all TestSuiteQueue controls (50.734s), focused race checks (1.749s), mutation-checked the old error path to reproduce the regression, and exercised real older helpers acquiring/releasing two private slots. The binary recorded codecomplete with its evidence binding; no source fix followed review.

The original recorded suite controller remains queued under lease-1824173798 on pushed head6a0b35faea2b351b5b1ec19a310f6ea1d19f1948, raw attempt2cb913014a018f68422fdf9ef52180cf. Last observed normal FIFO position12 behind the running Ariadne-slot2 suite; it has not run or passed. Retain this controller and queue position. No continuation or restart while it is queued/running. Open the PR for CI overlap; once the final-code recorded suite passes, enter central sdlc merge --yes and retain its normal landing turn. No staging or priority requested.

### 2026-10-22 — recorded full suite completed; automatic known-flake rerun retained

The original controller acquired both slots after4567s without cancellation or requeue. Its full Go phase completed in903.1s: all2367 cmd/sdlc tests across eight shards passed, with76 other packages exercised. Exactly one test failed: cmd/tl-lead TestOperatingCancellationStopsGitHookDescendants reported descendant state ?E at doc_operating_test.go:419. An unchanged targeted run from the exact detached suite checkout (Git HEAD verified6a0b35faea2b351b5b1ec19a310f6ea1d19f1948) passed in0.986s. Preserved failure events and targeted output under common test-stats/evidence/000563. Filed unclaimed ariadne#567 and registered that exact test under the worker flake rule; no source change or blind full retry.

The same retained controller automatically recognized #567 and queued one exact-package rerun on that head, lease-839917404. It currently waits for one suite slot (initial position10); it has not passed. Keep unified terminal68401 and the queued rerun intact. All PR406 hosted checks remain green; close remains first-round SHIP. Once the linked recorded result passes, enter central sdlc merge --yes promptly and retain its landing turn. No priority/staging requested.

Authenticated hub dispatch9c270b83-96f0-4235-b536-7f809314451c queues the next item after this landing: diagnose why pair-slot2 sdlc actual --issue162 requires old pair#151 adopted-window evidence, file/claim an Ariadne issue, distinguish real restorable evidence from labelled exclusion, and report. Do not alter Pair evidence speculatively. Pair#162 waits but is not urgent before08:00. #533 remains parked.

### 2026-10-23 — linked recorded result passed; enter landing

The retained automatic rerun acquired one slot after3671s and ran exactly cmd/tl-lead TestOperatingCancellationStopsGitHookDescendants on the original tested head. It passed (test0.65s, package0.806s), rerun attempt91ac7226a1f4076f391a475aec5fa101. SDLC exited0 and posted sdlc-local/suite success on6a0b35fa with known flake #567 linked, preserving the full failed attempt2cb913014a018f68422fdf9ef52180cf and its log at common test-stats/failed/2cb913014a018f68422fdf9ef52180cf.log (sha256 prefixf9e5f9de3ec9). The full-run command exit2 is retained inside that linked passing result; no failure was erased. All2367 SDLC tests and every tested package now have positive exact-head proof. No source change followed SHIP. Next: central sdlc merge --yes, retaining its normal landing turn and any authoritative integration suite. After landing, report PR406 and Git-derived merge evidence to the hub, remove only the private CLI diagnostic directory, then start the queued pair#162 measurement diagnosis.
