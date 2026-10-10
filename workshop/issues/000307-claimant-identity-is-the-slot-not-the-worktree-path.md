---
id: 000307
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '598a8dea371696d260b6400f5ca05f85414fa3c5' # card fields mirrored from issue-cards; edit via sdlc
---

# Claimant identity is the slot, not the worktree path

## Problem

A claim records `repository + machine + worktree` and ownership is an exact match on all three (`cmd/sdlc/internal/issue/claimant.go:167-176`). The operator's model, from the #283 discussion, is different: the claimant is a **slot on a machine** (e.g. `pair:1`). That's the durable home of one agent's thread of work, which may touch several repos.

The worktree-path match breaks cross-repo work. When pair:1 files or edits an ariadne issue, it works in `pair-slot1/ariadne`, the weave dependency copy. That path is not an ariadne slot worktree, and dependency checkouts aren't first-class workspaces: `sdlc move` refused "contextual workspace address from a dependency is ambiguous" (evidence `a2-0928:126`).

The match is also fragile within one repo:
- the same operator editing from :0 is refused;
- a `move` whose re-stamp failed is refused until `claim` runs again.

Evidence: `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`, Part 1 (#274 remaining cases 2–3) and "Operator decisions, round 3" D-1(a).

## Spec

- Ownership is decided by **slot identity**: machine fingerprint + slot address (`<family>:<N>`, e.g. `pair:1`, `ariadne:0`). Repository and worktree path stay on the card as descriptive fields.
- An issue in repo R claimed by `pair:1` is editable from any checkout belonging to slot `pair:1`, including its dependency copy of R.
- Migration: existing cards carrying worktree-path claimants resolve to their slot.
- `sdlc state` and the claim output name the slot.

## Done when

- A pair:1 session can claim, edit and publish an ariadne issue from `pair-slot1/ariadne`, and the transfer guard and the continuation gates accept it as the owner.
- The same issue edited from `ariadne-slot1` (a different slot) is refused as not-owner.
- Existing claimant cards keep working, with no flag day for older binaries (cf. G1 in the evidence).

## Plan

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). The operator decided the claimant is the slot (`pair:1`).
