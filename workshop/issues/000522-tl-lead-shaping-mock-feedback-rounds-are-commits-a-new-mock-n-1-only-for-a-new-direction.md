---
id: 000522
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'e08f7478b3271e1179b205b71daa1a7d22c127c7' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T03:02:01-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead shaping: mock feedback rounds are commits; a new mock-N+1 only for a new direction

## Problem

## Spec

Add the operator-approved (10-20) mock feedback-round convention to tl-lead's SHAPING section (landed in #518):

- **Each round of operator feedback refines the current mock version in place and is one commit** to that version
  (e.g. `mock-3: round 4: …`). Undo/redo of a round is git: "undo the last round" means the agent reverts that commit
  and the mock server reloads. The operator never touches git.
- **A new `mock-N+1` only for a different direction worth comparing side by side** (the operator says "try it another
  way", or a rework big enough that the old one should stay beside it). Older versions stay servable by URL.
- So versions stay few and meaningful (alternatives), rounds stay cheap (commits).
- Versions are separate `mock-N/` folders committed to git (not branches), sharing one root `fixtures.json`
  (pair#755's `couch mock serve` layout).

Source: ops `session.side/2026-10-20-34-mock-server-plan.md`; product-shaping project Log 10-20 23:05.

## Done when

- tl-lead's SHAPING section states the round/version rule above, concisely, next to the mock conventions.
- `tl-lead --help` (or its shaping subsection) shows it; existing tests for the help text pass.

## Plan

- [ ] Extend existing SHAPING rendered-help phrase guards for round commits/revert, direction-only folder versions, retained URLs and root fixtures; observe RED at both help entry points.
- [ ] Update the embedded mock conventions concisely, separating undoing a round from revisiting an earlier direction; observe GREEN and run focused TL tests.
- [ ] Push, run the central recorded suite, close through review, open/land the PR and report to the hub.

## Log

### 2026-10-21

Authenticated dispatch a88deb45-1953-4489-a2df-acffd86d5d7f authorizes this fixed operator rule while #520 waits. Preserved #520's committed branch; claimed and start-plan created this independent branch from main. Quick plan posted to the hub, proceeding unless objected: one embedded-help edit plus existing phrase guards, no runtime mechanism. ARCH-DRY retains help as the rule's one source; ARCH-PURPOSE covers round commits, agent-owned revert/reload, separate mock folders and shared fixtures together. Existing Git history owns the round lifecycle; no new parser/state/store, external dependency, secret, hot path or artifact family is introduced (ARCH-PURE/MOCK/SECURE/ORDER/CONSTRAINTS/FUNERAL/EVOLVE). No extra design approval is needed for the already operator-approved wording in the Spec.
