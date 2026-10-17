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

- [ ] Add a machine-readable stage build target to the sdlc sentinel, and document the marker in the generic build contract.
- [ ] Make stage --build run declared sentinel targets after make build, then require every Go binary in bin/ to have the staged, unmodified VCS revision. Require each sentinel main package to have a matching staged binary, including differently named outputs.
- [ ] Cover the stale sdlc binary, mismatched or unreadable revisions, and differently named sentinel outputs with tests; run the targeted package and make -k before closing.
- [ ] Reuse the build-and-verify path after `stage --clear`, including a retry from an already-resting target, and test that the binaries return to the resting HEAD.

## Log

### 2026-10-16

- Design: ARCH-DRY keeps build ownership in the existing Make target; the sentinel declares only the target. ARCH-PURE keeps revision and package matching separate from process execution. No new persistent state is created (ARCH-FUNERAL).

## Revisions

- 2026-10-16 19:45 PDT — TL observed that clear left staged binaries on main. Extend the same build-and-verify invariant to `stage --clear`, including retry after a failed clear; keep one shared check for both transitions.
