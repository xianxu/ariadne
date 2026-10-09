---
id: 000287
status: working
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-08
estimate_hours:
card_mirror: 'a8135e651c51bb2fc2853a37928136b7643cbe82' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-08T17:48:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "9f0d7c07", done: "bb26c6af"}
---

# Reconcile merges done outside sdlc

## Problem

Part of project `claimant-ownership`. pair#365 started with a GitHub merge (web button) while the card was not `codecomplete`. sdlc never noticed, and the bookkeeping was finished by hand. There must always be an escape hatch for work landed outside sdlc.

## Spec

Detect after the fact; no dependency on GitHub CI (a required check is deferred). A routine read (`sdlc state`, and the next `merge`) notices "issue branch merged into main, but card not `done`":
- If the close evidence is on main, finish the bookkeeping: card → `done`, archive the details/plans, delete the remote branch.
- Otherwise print the next action (run `close` from the owner's slot, or claim the issue first if it has no owner).

Independent of #284–#286: it needs only the merge detection and the existing close evidence.

## Done when

- A fixture with a GitHub-merged issue branch and a non-`done` card is reported by `sdlc state`.
- With the close evidence on main, reconcile flips the card to done and archives; without it, it prints the next action and changes nothing.

## Plan

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [x] Durable plan `workshop/plans/000287-reconcile-external-merges-plan.md`; single close, no Mx.
  - [x] Task 1: pure verdict
  - [x] Task 2: extract `archiveIssueOnMain` from abandon
  - [x] Task 3: detection and `sdlc state` findings
  - [x] Task 4: finisher via reconcile and merge; docs

## Log

### 2026-10-08

Claimed in ariadne:2. Survey: the evidence-on-main → done half exists (`settleLandedCompletions`, run by reconcile and after every merge), but it doesn't archive, doesn't delete the remote branch, and nothing reports it; `sdlc state` only has a commit-subject heuristic. Design: detect by ancestry of the pushed branch tip (kept current by #286) and of the completion's evidence; `state` reports with the next action and never writes; reconcile and the next merge finish (done, archive via the narrow main commit extracted from abandon, leased remote delete). Squash/rebase merges outside sdlc are a non-goal: ancestry can't see them without the GitHub API.

### 2026-10-02

Built: `externalMergeVerdict` (pure); `archiveIssueOnMain` extracted from abandon; `externalMerges` detection. A branch counts as merged when its pushed tip is on main but off main's first-parent line, because a branch fresh from start-plan sits on that line; that rule is mutation-checked by `TestStateIgnoresAFreshBranch`. `sdlc state` reports with the next action and suppresses the commit-subject 'looks done' guess for the same issue. The finisher `finishLandedLeftovers` archives every done card whose details are still live on main, after a landing has archived its own issues. It runs from reconcile and at the end of merge (a landing settles before its own archive, so the IDs that settle returns can't drive it). The merge hook, the reconcile hook and the branch deletion are each mutation-checked. make test green except the sandbox-only processgroup test.
