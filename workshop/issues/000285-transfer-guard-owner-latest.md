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
  - [x] Task 1: pure verdict
  - [x] Task 2: collector + guard rewrite
  - [x] Task 3: `sdlc issue restore`
  - [x] Task 4: pair#365 end to end
  - [ ] Task 5: retire owner-as-branch wording; atlas

## Log

### 2026-10-08

Claimed in ariadne:2, start-plan. Design (plan D1–D7): protected set widens from handed-off to every *published* details file (main's history touched it); "changes" = prospective merge differs from main; owner = card claimant via `ownership()` (ARCH: branch names out entirely); based-on-latest = HEAD contains main's last commit to the path (Spec's literal condition, not a content compare); not-owner refusal offers a new narrow `sdlc issue restore --issue N`, owner-behind says merge main. Unreadable-card fail-closed narrows from repo-wide to the changed paths' cards. Survey: "owner = branch" lives only in `transferguard.go`; nothing else on close → pr → merge keys on the issue's branch name, so pair#365 is the guard alone.

### 2026-10-02

Task 2: guard rewritten (`changedDetails` + `detailsRefusal`; `ownerAuthored`/`checkTransferredPaths`/`validHandoffDestination` gone). Mutation checks: forcing owner=true fails the four non-owner rows; forcing based=true first survived because every behind fixture also conflicted, so I added a clean-merge behind case, which now catches it; forcing published=false fails four tests. The Conflict fact was unobservable, so I dropped it (plan Revisions). Lesson-worthy: a behind check whose only fixtures also trip another check is vacuous (fits the existing 'each rejection row asserts its own check' lesson; no new entry).

Task 3: `issue restore` shipped. The archived case found that merge-tree follows an archive's rename, so a stale edit to `workshop/issues/X` surfaces as a change to `workshop/history/issues/X`. Restore therefore makes every copy of a flagged issue's details equal main's (diff main..HEAD by basename), not only the path the guard named.

Task 4: `TestCloseOnARenamedBranchAfterAnOutsideMergeLands` (pair365_e2e_test.go): outside merge leaves the card working, close on `<branch>-close`, `pr --dry-run`, `merge` → card done. Restoring branch-name ownership in the guard fails it at `sdlc pr`, as pair#365 did. No other branch-name dependency surfaced.
