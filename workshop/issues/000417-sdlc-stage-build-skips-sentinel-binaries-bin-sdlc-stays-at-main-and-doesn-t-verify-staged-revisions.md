---
id: 000417
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '02db92c51cda4f44cef860b0b3e415e97d0b0e92' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T19:43:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# sdlc stage --build skips sentinel binaries (bin/sdlc stays at main) and doesn't verify staged revisions

## Problem

## Spec

Found staging #416 on 2026-10-16. `sdlc stage --build` in ariadne:0 built every binary except `sdlc`. The generic `make build` scanner skips `cmd/sdlc` because of its `.skip-make-build` sentinel (#31: sdlc builds through `make sdlc-build`). So `bin/sdlc` stayed at main (0e32b0eb) while the stage reported 00558776. A live check of an sdlc change would have run the old binary. The TL rebuilt it by hand.

In ariadne, `stage --build` should build each binary through its own build target: for a sentinel-skipped binary, the target its sentinel names. It should then verify that every staged `bin/*` carries the staged `vcs.revision`, and refuse when one doesn't.

Clearing a stage must restore the same invariant for the resting branch: rebuild its binaries from that branch and verify their revisions. A failed rebuild or verification must remain retryable through `stage --clear`.

## Done when

- `sdlc stage --build` in an ariadne-style repo leaves `bin/sdlc` at the staged revision.
- The stage checks each built binary's `vcs.revision` and refuses, naming the binary, when one doesn't match; a test covers that refusal.
- `sdlc stage --clear` rebuilds from the restored main HEAD, checks the same revision invariant, and can repair a prior failed clear on retry; tests cover both.

## Plan

- [x] Add a machine-readable stage build target to the sdlc sentinel, and document the marker in the generic build contract.
- [x] Make stage --build run declared sentinel targets after make build, then require every Go binary in bin/ to have the staged, unmodified VCS revision. Require each sentinel main package to have a matching staged binary, including differently named outputs.
- [x] Cover the stale sdlc binary, mismatched or unreadable revisions, and differently named sentinel outputs with tests; run the targeted package and make -k before closing.
- [x] Reuse the build-and-verify path after `stage --clear`, including a retry from an already-resting target, and test that the binaries return to the resting HEAD.

## Log

### 2026-10-16
- 2026-10-16: closed — Targeted TestStage and TestSentinel tests passed, including red-first corrupt XCOFF, Plan 9 and Mach-O cases; missing clean stamps, Make argument injection, stateful target retry and real Make conformance are also covered. The exact-head local-check suite is queued on pushed commit a7f66f30; earlier make -k passed before the review fixes. Live ariadne:0 stage and clear awaits a fresh TL grant before landing.; review verdict: SHIP
- 2026-10-16: flow upgraded quick → full — 167 added lines in code files (limit 100); an earlier round of this close already ran the full review

- Design: ARCH-DRY keeps build ownership in the existing Make target; the sentinel declares only the target. ARCH-PURE keeps revision and package matching separate from process execution. No new persistent state is created (ARCH-FUNERAL).
- Targeted stage/clear tests passed after watching the stale-revision, skipped-sentinel and unrepaired-clear cases fail first. A real `make build` skipped sdlc; `make sdlc-build` produced `bin/sdlc` at the branch HEAD with `vcs.modified=false`.
- Full `make -k test` passed: 12 shards, 1,527 `cmd/sdlc` tests, 66 other packages, zero failures, 505.7s suite wall time. The live :0 stage/clear check awaits the TL's fresh grant because #418 currently holds that slot. TL requested the close review first.
- Close boundary review round 1 returned REWORK: corrupt native binaries or missing `vcs.modified` could pass, marker values could become Make options, Make sequencing lacked a stateful fake, and README omitted the changed user flow. Added red-first regressions for the validation gaps, an injected Make boundary with failure/retry state, a real-Make fixture, and README/atlas guidance. Re-running targeted and full checks before close round 2.
- Close round 2 disposed the Make argument, process-seam and README findings. It caught XCOFF, Plan 9 and a Mach-O variant omitted from the native-format guard, plus a mistaken PURE label for a file-reading parser. Added a red-first cross-format regression, completed the format set and corrected the plan classification. The exact-head suite queued on the earlier commit cannot cover this fix; a new run is required after publication.

## Revisions

- 2026-10-16 19:45 PDT — TL observed that clear left staged binaries on main. Extend the same build-and-verify invariant to `stage --clear`, including retry after a failed clear; keep one shared check for both transitions.
