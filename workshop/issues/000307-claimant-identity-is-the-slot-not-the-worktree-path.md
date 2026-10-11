---
id: 000307
status: working
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-11
estimate_hours:
card_mirror: '432175185a420438eb48741fc62bfde9bf59c326' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T00:28:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "bc5d34a2", done: "a08684af"}
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

### Design (2026-10-11)

- **Home slot** = the slot whose environment holds the checkout. `pkg/workspace` already proves it: a dependency clone (`pair-slot1/ariadne`, kind `dependency`) and a feature worktree inside an environment carry `EnvironmentHost` = `pair:1`; a slot's host is itself; a `:0` primary has no host, so its own address counts where the slot layout is in use (today's rule). New `workspace.Identity.HomeSlot()` owns this read (one place, `pkg/`, per the #289 lesson). A plain clone or a fleet-level feature worktree has none.
- **The card schema does not change.** The existing optional `workspace` key holds the home slot. A new key would fail older binaries' closed schema parse (the G1 flag day); a value change in a descriptive field does not. Older binaries keep matching on the worktree, which is still recorded, so a claim made by a new binary stays theirs in the same checkout.
- **Decision** (`issue.MatchClaimant`): same machine, and
  - both sides name a slot → same slot (repository and worktree are descriptive);
  - either side has none (a plain clone, a card claimed before the label covered dependencies) → today's rule, same repository and worktree. This is the migration: such a card keeps working from the checkout that claimed it. No lazy re-stamp; a repeat claim stays a no-op.
- Every gate already funnels through `MatchClaimant` (`ownership()`: continuation gate, transfer guard, move, set-status, unclaim/reclaim, claim, observe), so one change covers them.
- Not changed: `:0` stays its own slot per repository (an operator editing from `:0` an issue held by `:1` is still refused; that is reclaim's job). `sdlc move` from a dependency (the ambiguous-address refusal) is out of scope.
- Output: the continuation refusal and the claim success line name the slot; `describeClaimant` already does. Sweep help/atlas/recovery prose that says "repository, machine and worktree".
- ARCH: creates nothing durable (no new record or field).

## Plan

- [ ] `workspace.Identity.HomeSlot()` + table test, and assertions in `TestEnvironmentIndependentConformance` (slot, dependency, feature worktree → host address)
- [ ] `resolveClaimantIdentity` records `HomeSlot()`; replace `slotLabel` and its test
- [ ] `MatchClaimant` slot rule with the legacy fallback; table rows: same slot other worktree (mine), other slot same machine (foreign), same slot other machine (foreign), unlabeled sides (legacy)
- [ ] In-process gate tests: `requireCardOwnership` and the transfer guard accept the slot owner from another checkout path and refuse another slot
- [ ] Messages name the slot; sweep help, atlas, recovery catalog, AGENTS restatements
- [ ] close

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). The operator decided the claimant is the slot (`pair:1`).
