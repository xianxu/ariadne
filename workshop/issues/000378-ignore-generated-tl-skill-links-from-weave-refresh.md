---
id: 000378
status: working
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
severity: low
card_mirror: '2acdefde628c6a1180fd34a3268c049b6a32a856' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-14T15:05:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "ad02ad32", done: "7c4caa8d"}
---

# Ignore generated TL skill links from weave refresh

## Problem

The #350 TL skills make `weave refresh` generate four local skill links that are not ignored. A refreshed slot then has a dirty checkout, which blocks `sdlc start-plan` for its next issue.

## Spec

Add the four generated TL skill paths to the existing `.gitignore` skill-link groups, in their sorted positions. The paths are `/.agents/skills/xx-tl-lead`, `/.agents/skills/xx-tl-worker`, `/.claude/skills/xx-tl-lead`, and `/.claude/skills/xx-tl-worker`. Keep the change to `.gitignore`; the weave output already defines the paths (ARCH-DRY).

## Done when

- `weave refresh` leaves no tracked changes in a clean checkout containing the new ignore entries.
- The branch patch changes only `.gitignore` by those four lines, and the required quick-flow gates pass.

## Plan

- [ ] Carry the exact four-line generated diff onto this issue branch, verify its scope and repeat `weave refresh`.
- [ ] Run the required build and suite, and verify the exact patch before close.

Delivery follows the close review: land the PR and report its merge commit.

## Log

### 2026-10-14

- `weave refresh` on main-slot1 at `7a7223628f1ec78a213ad4dd6ea4f0f9761c9f31` changed only `.gitignore` by the four specified additions. The generated diff was stashed while `start-plan` created this issue branch from main.
