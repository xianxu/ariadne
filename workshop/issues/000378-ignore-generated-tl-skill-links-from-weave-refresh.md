---
id: 000378
status: codecomplete
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
severity: low
card_mirror: '5d4371153a3ed104313f2fd614108b263795138f' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T15:05:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ad02ad32", done: "7c4caa8d"}
actual_hours: 0.91
---

# Ignore generated TL skill links from weave refresh

## Problem

The #350 TL skills make `weave refresh` generate four local skill links that are not ignored. A refreshed slot then has a dirty checkout, which blocks `sdlc start-plan` for its next issue.

## Spec

Add the four generated TL skill paths to the existing `.gitignore` skill-link groups, in their sorted positions. The paths are `/.agents/skills/xx-tl-lead`, `/.agents/skills/xx-tl-worker`, `/.claude/skills/xx-tl-lead`, and `/.claude/skills/xx-tl-worker`. Keep the change to `.gitignore`; the weave output already defines the paths (ARCH-DRY).

## Done when

- All four generated TL skill paths are ignored by the exact new `.gitignore` patterns.
- The non-issue branch patch changes only `.gitignore` by those four lines, and the required quick-flow gates pass.

## Plan

- [x] Carry the exact four-line generated diff onto this issue branch and verify its scope and ignore matches.
- [x] Run the required build and suite, and verify the exact patch before close.

Delivery follows the close review: land the PR, repeat `weave refresh` from the resting branch, confirm it stays clean, and report the merge commit.

## Log

### 2026-10-14
- 2026-10-14: closed — make -k tools passed; four exact generated TL skill paths matched the added ignore patterns; git diff origin/main...HEAD -- .gitignore showed only those four additions; git diff --check passed; recorded sdlc-local/suite passed on baab3397 (12 shards, 1353 sdlc tests, 59 packages). No atlas surface: ignore entries only.; review verdict: SHIP

- `weave refresh` on main-slot1 at `7a7223628f1ec78a213ad4dd6ea4f0f9761c9f31` changed only `.gitignore` by the four specified additions. The generated diff was stashed while `start-plan` created this issue branch from main.
- On the issue branch, `git check-ignore -v` matched each of the four paths to its exact added line. `weave refresh` refuses a branch ahead of main, so the clean-refresh check moves to the resting branch after landing.
- `make -k tools` passed. The recorded `sdlc local-check run --name suite` passed on `baab3397b68763918c8d530b83df78aebd28b62f`: 12 shards, 1,353 `cmd/sdlc` tests, and 59 other packages. A prior suite run hit an intermittent `cmd/weave` process-death recovery failure; that focused pair of tests passed three repetitions, and the complete rerun passed. The transient GitHub API interruption cleared and the passing result posted.

## Revisions

- 2026-10-14 — The first verification wording required running `weave refresh` on the ahead issue branch, which weave refuses. The pre-close contract now checks all four generated paths and the exact diff; the operational refresh check follows the merge on the supported resting branch. The four-line scope is unchanged.
