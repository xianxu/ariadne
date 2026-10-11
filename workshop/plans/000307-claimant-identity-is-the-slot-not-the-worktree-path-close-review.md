# Boundary Review — ariadne#307 (whole-issue close)

| field | value |
|-------|-------|
| issue | 307 — Claimant identity is the slot, not the worktree path |
| repo | ariadne |
| issue file | workshop/issues/000307-claimant-identity-is-the-slot-not-the-worktree-path.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73b717f853f34208840eada67b7614653894ba32..89a716b83cee0d8d59986d9376bdea94bf4d13a7 |
| command | sdlc close --issue 307 |
| reviewer | claude |
| timestamp | 2026-10-11T01:50:10-07:00 |
| verdict | SHIP |

## Review

Checking the remaining code that matches by worktree path (observe, fleet claims), then I'll write up the verdict.

Nearly done. I'm running the pure-package tests the diff touches, then I'll give the verdict.

```verdict
verdict: SHIP
confidence: medium
```

**Verdict: SHIP.** The change does what the issue asks. Ownership is now decided by slot on a machine, through a single function, `issue.MatchClaimant`. Every ownership gate goes through it via `ownership()`: the continuation gate, the transfer guard, publish/republish, move, set-status, claim, unclaim, reclaim, abandon, external-merge and observe. So the one change in `MatchClaimant` covers all of them. The home slot is worked out in one place, `workspace.Identity.HomeSlot()`. It reuses the environment-host facts that `pkg/workspace` already proves, so a dependency clone like `pair-slot1/ariadne` correctly resolves to `pair:1`.

Migration works without a flag day. The card schema is unchanged and the slot rides in the existing `workspace` key. When either side records no slot, the old rule applies (same repository, machine and worktree), so older cards and binaries keep working. I ran the `pkg/workspace` and `internal/issue` tests at the head commit and both pass. I did not run the cmd/sdlc gate tests. What's left is minor: two display/inventory paths still match by worktree, and the "accepts the owner" path is only tested on two of the four continuation gates.

1. **Strengths**
   - `cmd/sdlc/internal/issue/claimant.go:173`: the rule is ordered clearly (machine first, then slot, then the old fallback). The table covers same slot from another checkout, another slot at the same path, another machine, and cards with no slot recorded on either side.
   - `pkg/workspace/address.go:71`: `HomeSlot` replaces `slotLabel` and moves it into `pkg/`, so it isn't duplicated (ARCH-DRY pass). It's a pure method on `Identity`, tested both from a table and against real resolved identities (`TestEnvironmentIndependentConformance` asserts the host address for slot, dependency and feature checkouts).
   - The schema-neutral migration avoids breaking older binaries, which reject unknown keys when parsing a card.
   - `claimantWhere` is shared by `describeClaimant`, the refusal message and the claim success line, so the slot wording comes from one place.
   - The in-process gate tests (`TestOwnershipGatesFollowTheSlot`, `TestTransferGuardFollowsTheSlot`) cover both sides of the Done-when: a different slot is refused even at the card's own path, and the owning slot is accepted from another checkout.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Two places still match by worktree, not slot.**
     - `cmd/sdlc/internal/observe/checkpoints.go:72`: `IsClaimant` is set only when `Claimant.Worktree == h.Path`.
     - `cmd/sdlc/internal/fleet/claims.go:129` and `:149`: claims are attached to inventory rows by worktree, and reported as `dangling_claim` when that worktree isn't present. A slot-owned claim made from a checkout that has since been removed (say a feature worktree in pair-slot1) will show as dangling, although the slot still owns it.

     These only affect display and inventory, not any gate. Either key them on `Machine + Workspace` when both sides have a slot (ideally sharing `MatchClaimant`), or document them as checkout-level views.
   - **Multiple fleet roots on one machine.** Two fleet roots (e.g. `~/workspace` and `~/other`), each with an `ariadne-slot1`, now count as the same owner. The worktree used to tell them apart. This follows from the Spec as written; it just isn't stated anywhere.
   - **Acceptance tested on only two gates.** `TestOwnershipGatesFollowTheSlot` tests acceptance from another checkout only for start-plan and change-code; close and milestone-close are tested only for refusal. They share `requireCardOwnership`, so the risk is low.
   - **Wrapping.** `atlas/workflow/vocabulary.md:164` and `cmd/sdlc/helptext/claim.md:66` have one over-long line each; the rest is wrapped.
   - **Leftover duplicate.** `cmd/sdlc/observe.go:200` still has the old `Address != nil && UsesSlotLayout()` check. It describes the holding checkout's own address, not its home slot, so it's not a duplicate in meaning. It's worth a comment so nobody "fixes" it to `HomeSlot`.

5. **Test coverage notes:** The pure rule and `HomeSlot` are tested without IO apart from temp directories (ARCH-PURE pass). The gate tests run the real gate code paths, with the claimant identity swapped via `withClaimant`. Both Done-when outcomes (owning slot accepted, other slot refused) are covered, as is the case where only the current checkout has no slot recorded.

6. **Architecture**
   - **ARCH-DRY:** pass. `slotLabel` was removed and replaced by `HomeSlot`; `claimantWhere` is shared.
   - **ARCH-PURE:** pass. The rule lives in `issue.MatchClaimant`, and the only change in the IO layer is one line in `resolveClaimantIdentity`.
   - **ARCH-PURPOSE:** pass. Every gate that decides ownership goes through `MatchClaimant`. The only consumers still keyed on worktree are the observe and fleet display paths in the first Minor item, which aren't the point of the issue.

7. **Plan revisions:** none needed. If the observe and fleet views stay worktree-keyed on purpose, a one-line Log note saying so would help.

```findings
findings:
  - id: new
    severity: Minor
    family: ownership-shadow-consumer
    title: |
      observe IsClaimant and fleet claim association/dangling still key on worktree, not slot
    detail: |
      cmd/sdlc/internal/observe/checkpoints.go:72 and cmd/sdlc/internal/fleet/claims.go:129,149 match Claimant.Worktree to a checkout path; a slot-owned claim whose recorded checkout is gone reports as dangling though the slot still owns it. Display/inventory only, not a gate.
  - id: new
    severity: Minor
    family: slot-identity-uniqueness
    title: |
      same-machine fleets with identically named slots now share ownership
    detail: |
      Two fleet roots on one machine each with ariadne-slot1 both label ariadne:1; the worktree used to separate them. This follows from the Spec but is undocumented.
  - id: new
    severity: Minor
    family: done-when-both-states
    title: |
      close and milestone-close are tested only for refusal, not acceptance from another checkout of the owning slot
    detail: |
      TestOwnershipGatesFollowTheSlot tests acceptance only for start-plan and change-code; all four share requireCardOwnership.
```
