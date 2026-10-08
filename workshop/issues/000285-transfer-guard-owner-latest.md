---
id: 000285
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-08
estimate_hours:
card_mirror: '00f8142a312c05cb695f1b5f6a7af363915f06e8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-08T13:01:20-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "af3ac220", done: "674e29fb"}
---

# Transfer guard: owner plus based-on-latest

## Problem

Part of project `claimant-ownership`. `cmd/sdlc/transferguard.go` calls the branch named after the details file the "owner". That stranded pair#365: after #148 retired the branch name, nothing could land the issue's own close.

## Spec

The guard protects only published details files: it is optimistic concurrency on them. For each published details file a landing would change, the change is accepted only if:
1. the landing runs from that issue's **owner's checkout** (compared with the card, since commits don't carry the claimant), and
2. HEAD is **based on main's latest version** of the file (it contains the commit that last published it).

Otherwise the prospective merge must leave the file exactly as main has it. Branch names play no part, so a reopened issue needs no fresh branch name.

Resolution: when a non-owner branch changed details, sdlc restores main's version mechanically (and says: claim the issue to keep the edit). When the owner is behind main after a handoff, the owner slot's LLM merges main and resolves the prose. Retire all "owner = branch" wording.

## Done when

- A fixture reproduces pair#365 (close after the PR merged, on a renamed branch); claim, then `sdlc pr`, then `sdlc merge` flips the card to done.
- A stale filing branch that still carries a published details file is refused, and the offered restore makes it land.
- An owner branch behind main's latest version is refused with a merge-main next action; after the merge it lands.
- No "owner" in sdlc output or help refers to a branch.

## Plan

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [ ] Durable plan: `workshop/plans/000285-transfer-guard-owner-latest-plan.md` (full flow: guard rewrite + `issue restore` exceed the quick-flow code limit). Single close, no Mx.
  - [ ] Task 1: pure verdict
  - [ ] Task 2: collector + guard rewrite
  - [ ] Task 3: `sdlc issue restore`
  - [ ] Task 4: pair#365 end to end
  - [ ] Task 5: retire owner-as-branch wording; atlas

## Log

### 2026-10-08

Claimed in ariadne:2, start-plan. Design (plan D1–D7): protected set widens from handed-off to every *published* details file (main's history touched it); "changes" = prospective merge differs from main; owner = card claimant via `ownership()` (ARCH: branch names out entirely); based-on-latest = HEAD contains main's last commit to the path (Spec's literal condition, not a content compare); not-owner refusal offers a new narrow `sdlc issue restore --issue N`, owner-behind says merge main. Unreadable-card fail-closed narrows from repo-wide to the changed paths' cards. Survey: "owner = branch" lives only in `transferguard.go`; nothing else on close → pr → merge keys on the issue's branch name, so pair#365 is the guard alone.

### 2026-10-02
