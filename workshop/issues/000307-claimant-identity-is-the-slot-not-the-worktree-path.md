---
id: 000307
status: codecomplete
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-11
estimate_hours:
card_mirror: '606dd931c4f6304981956ea0a042ab1ee3452e68' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-11T00:28:33-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "bc5d34a2", done: "a08684af"}
actual_hours: 0.97
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

- [x] `workspace.Identity.HomeSlot()` + table test (`TestHomeSlotOnlyWhereSlotsExist`, `TestHomeSlotOutsideEnvironments`), and assertions in `TestEnvironmentIndependentConformance` (slot, dependency, feature worktree → host address)
- [x] `resolveClaimantIdentity` records `HomeSlot()`; replace `slotLabel` and its test (moved to pkg/workspace)
- [x] `MatchClaimant` slot rule with the legacy fallback; table rows: same slot other worktree (mine), other slot same machine (foreign), same slot other machine (foreign), unlabeled sides (legacy)
- [x] In-process gate tests (`TestOwnershipGatesFollowTheSlot`, `TestTransferGuardFollowsTheSlot`): `requireCardOwnership` and the transfer guard accept the slot owner from another checkout path and refuse another slot
- [x] Messages name the slot (`claimantWhere`); sweep help, atlas, recovery catalog (no stale claim there), AGENTS (none)
- [x] close

## Log

### 2026-10-09
Filed from the robustness evidence review (ariadne-robustness-1). The operator decided the claimant is the slot (`pair:1`).

### 2026-10-11
- 2026-10-11: closed — Round-3 minor fixed: observe and fleet Claimant views convert from issue.Claimant by struct conversion (compile-time drift guard). fleet + observe packages pass; build ok. (fixed-then-ship); review verdict: SHIP
- 2026-10-11: closed — BR-4 fixed: observe isClaimantCheckout delegates to issue.MatchClaimant (machine + slot, legacy fallback; unknown identity false), TestIsClaimantCheckoutFollowsTheSlot covers other machine + unknown identity. Earlier minors: observe holdings use HomeSlot; close/milestone-close acceptance asserted; fleet/same-named-slot limits documented. Targeted tests pass after merging main. (fixed-then-ship); review verdict: SHIP
- 2026-10-11: closed — make test: 12 shards / 1049 cmd/sdlc tests pass; only processgroup failed in-sandbox (ps blocked), passes unsandboxed. New: TestHomeSlotOnlyWhereSlotsExist, TestHomeSlotOutsideEnvironments, HomeSlot asserts in TestEnvironmentIndependentConformance (real git: dependency clone + feature worktree -> host slot), TestMatchClaimant slot/legacy rows, TestOwnershipGatesFollowTheSlot (all 4 gates refuse other slot at same path; owning slot passes from another checkout), TestTransferGuardFollowsTheSlot. Live probe: issue show 307 = this-workspace from ariadne-slot1, other-workspace from pair-slot1/ariadne.; review verdict: SHIP
- Probed real identities with a scratch build: `pair-slot1/ariadne` resolves as kind `dependency`, no address, `environment_host` pair:1. So the home slot was already provable; only the claimant read ignored it (`slotLabel` needed `Address`).
- Messages: `claimantWhere` (slot + checkout) shared by `describeClaimant`, the continuation refusal and the claim success line.
- Prose sweep: claim help, `ownershipGateHelp`, atlas issue-tracker (ownership + transfer guard), atlas vocabulary, `issue.cue` comment.
- Live probe (scratch build): `sdlc issue show 307` reads `this-workspace` from ariadne-slot1 and `other-workspace` from `pair-slot1/ariadne` (pair:1), the card naming ariadne:1.
- A targeted run tripped the hermeticity guard on a leaked `.git/sdlc.lock`. That lock lives in the common dir shared by every ariadne slot; rerunning only the new tests was clean, so another slot's verb held it at the check.
- Close review round 1: SHIP, 3 Minors, all handled in-round:
  - observe's `IsClaimant` and holding label re-derived the old slot rule (`Address` + `UsesSlotLayout`). Fixed: holdings use `HomeSlot()`, and `isClaimantCheckout` matches by slot, else worktree (`TestIsClaimantCheckoutFollowsTheSlot`). Fleet placement stays on the recorded checkout: rows carry no slot, and threading one through is its own change. Documented in atlas and the `PlaceClaims` comment.
  - Same-named slots in two fleet roots on one machine share ownership: documented in atlas (one fleet per machine).
  - `TestOwnershipGatesFollowTheSlot` now asserts that close and milestone-close let the owning slot through to the review dispatch.
- Close review round 2 (fixed-then-ship), BR-4 Important: my `isClaimantCheckout` re-implemented the match without the machine check. Fixed the class: it now moves the queried identity to the holding checkout and asks `issue.MatchClaimant`, and an unknown identity is never the claimant. Test rows added for another machine and an unknown identity. A sweep for hand-rolled claimant comparisons leaves only fleet placement, a location join over this machine's claims (documented).
- Close review round 3: SHIP, 1 cosmetic Minor (the observe/fleet Claimant views hand-copied issue.Claimant's six fields). Fixed: they convert directly (`Claimant(recorded)`), so a field added to one fails to compile rather than being dropped.

