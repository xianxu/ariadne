---
id: 000287
status: open
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: 'dd197e6da8c6210c548643a357fe00e3ee5b78b9' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.

## Log

### 2026-10-02
