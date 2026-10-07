---
id: 000285
status: open
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'bd21c617969322579484cf734fbf6984b1165508' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.

## Log

### 2026-10-02
