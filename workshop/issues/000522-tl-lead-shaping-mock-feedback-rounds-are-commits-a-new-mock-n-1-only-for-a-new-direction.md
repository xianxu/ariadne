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
flow: {kind: quick, provenance: inferred, spec: "dfb66709", done: "f7ef2dfd"}
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

- [x] Extend existing SHAPING rendered-help phrase guards for round commits/revert, direction-only folder versions, retained URLs and root fixtures; observe RED at both help entry points.
- [x] Update the embedded mock conventions concisely, separating undoing a round from revisiting an earlier direction; observe GREEN and run focused TL tests.
- [x] Push and record the central suite pass for delivery preparation; close review, PR, central landing and hub report follow through their gates.

## Log

### 2026-10-21
- 2026-10-21: closed — Existing SHAPING guards failed on the missing round/version rule at both help entry points, then passed with the embedded rule (0.156s); full cmd/tl-lead tests passed (14.324s), actual go run help inspected. Central recorded suite on 8af32958a7a3597b8b896af63e23b35d3e1cc2ec passed 1921 SDLC tests +75 packages with zero failures/reruns in721.9s, log hash e50a8b7e2b31. One commit per feedback round, agent-owned revert/reload, direction-only folder versions, retained URLs and root fixtures are all delivered; no runtime or local-gates change. (done-when: contract moved since change-code; judged by the close review); review verdict: SHIP

Authenticated dispatch a88deb45-1953-4489-a2df-acffd86d5d7f authorizes this fixed operator rule while #520 waits. Preserved #520's committed branch; claimed and start-plan created this independent branch from main. Quick plan posted to the hub, proceeding unless objected: one embedded-help edit plus existing phrase guards, no runtime mechanism. ARCH-DRY retains help as the rule's one source; ARCH-PURPOSE covers round commits, agent-owned revert/reload, separate mock folders and shared fixtures together. Existing Git history owns the round lifecycle; no new parser/state/store, external dependency, secret, hot path or artifact family is introduced (ARCH-PURE/MOCK/SECURE/ORDER/CONSTRAINTS/FUNERAL/EVOLVE). No extra design approval is needed for the already operator-approved wording in the Spec.

### 2026-10-21 — Phrase guards RED/GREEN and delivered help

Extended the existing TestLeadHelpShapingMockAndShippingBoundary guard, which renders both --help and no-argument help. RED reported the missing round/commit/revert, direction-only folder/URL/root-fixture and switch-back clauses at both entry points. After the compact embedded-help edit, targeted shaping tests passed (0.156s) and the full cmd/tl-lead package passed (14.324s). Actual go run ./cmd/tl-lead --help output includes every new convention. Atlas adds only a test-surface pointer. No staging or browser action is necessary for this fixed text rule; no #520 code or plan was changed. Proceed to the central recorded suite and mandatory close/merge gates.


### 2026-10-21 — Recorded suite passed

Central sdlc local-check run --name suite passed on 8af32958a7a3597b8b896af63e23b35d3e1cc2ec after1086s queued: 1921 cmd/sdlc tests across eight shards and75 other packages, all exit0, zero failures/reruns, wall721.9s. Status sdlc-local/suite recorded; log /Users/xianxu/workspace/ariadne/.git/sdlc/local-checks/8af32958a7a3597b8b896af63e23b35d3e1cc2ec/suite.log, sha256 prefix e50a8b7e2b31. PR365 is open and all three CI checks passed on the code head. The source remained frozen throughout the recorded run; no replacement suite or queue change. Delivery preparation is complete; mandatory close and central merge still follow.

## Revisions

### 2026-10-21 — Delivery checkbox records pre-close evidence

The final checklist row records pushed-head suite evidence, with close/PR/landing/report explicitly following through their gates. This avoids requiring a future landing before entering close; scope and required gates are unchanged.
