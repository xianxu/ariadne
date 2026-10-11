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

---

## Re-review — 2026-10-11T01:57:35-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 307 — Claimant identity is the slot, not the worktree path |
| repo | ariadne |
| issue file | workshop/issues/000307-claimant-identity-is-the-slot-not-the-worktree-path.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab4f690039d8f0f845a14baf4d8083d628fb1895..c8583dc12ed1c5f5e2ad8b666a0c3c5554f2fac6 |
| command | sdlc close --issue 307 |
| reviewer | claude |
| timestamp | 2026-10-11T01:57:35-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Clean merge with no conflict hunks. Now checking whether the new observe matcher duplicates the ownership gate's rule.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This fix window responds to all three prior Minors, and the merge from main was clean (`git show --cc` shows no conflict hunks). BR-2 and BR-3 are fully addressed. For BR-1, the fleet half is now a documented limitation. The observe half adds `isClaimantCheckout`, and that new matcher brings in a defect: it rewrites the ownership rule instead of calling `issue.MatchClaimant`, and it drops the machine check. So when the claim comes from a same-named slot on another machine, the local checkout of that slot is marked `IsClaimant`. This is display only, not a gate, and it is cheap to fix, so the verdict is FIX-THEN-SHIP.

1. **Strengths**
   - `cmd/sdlc/observe.go:200`: holdings now get their address from `id.HomeSlot()`, the same source `claimantIdentity` uses (`cmd/sdlc/claimant.go:42`). A slot's dependency clones now label the same way on both sides.
   - `cmd/sdlc/ownership_test.go:508-521`: `close` and `milestone-close` are now tested for acceptance from another checkout of the owning slot. The test checks that the ownership refusal ("may not continue") is absent and that the review actually ran (`judged`).
   - `atlas/workflow/issue-tracker.md:44-48`: the docs now state both the one-fleet-per-machine limit and how fleet treats a dangling claim.

2. **Critical:** none.

3. **Important**
   - `cmd/sdlc/internal/observe/checkpoints.go:66-74`: `isClaimantCheckout` rewrites the slot-then-worktree switch from `issue.MatchClaimant` (`internal/issue/claimant.go:173`) but leaves out the `Machine` and `Repository` checks. Failure case: an issue claimed by `pair:1` on laptop B, observed on laptop A while A's `pair:1` holds the branch. That checkout gets `IsClaimant=true`, which contradicts `Relation=other-workspace` and `ClaimantWorktree=other-machine` in the same output. Slot names repeat across machines by design, so this is much more likely than the old path collision. Fix: for each holding, build `issue.Claimant{Machine: in.Me.Machine, Repository: in.Me.Repository, Workspace: h.Address, Worktree: h.Path}` and test `issue.MatchClaimant(&recorded, that) == OwnershipMine`, returning false when `in.Me == nil`. Then delete `isClaimantCheckout`, or reduce it to that call. Add a test case for a claimant on another machine (ARCH-DRY, ARCH-PURPOSE).

4. **Minor:** `cmd/sdlc/internal/fleet/claims.go:102`: the comment change is documentation only. That is acceptable because fleet placement places a claim on a checkout row, which is a different question from who owns it.

5. **Test coverage notes:** `TestIsClaimantCheckoutFollowsTheSlot` covers slot versus path well but has no case where the machine differs. That gap is how the defect above got through.

6. **Architectural notes:** every "is this the claimant?" decision should go through `issue.MatchClaimant`. Today observe's `assembleAssignment` calls it, while `assembleWorkspaces` has its own copy. That copy is the second instance in the `ownership-shadow-consumer` family.

7. **Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      observe now keys IsClaimant on the slot (with a new defect, raised below); fleet placement is documented as recorded-checkout inventory in the atlas and the PlaceClaims comment.
  - id: BR-2
    disposition: addressed
    note: |
      atlas/workflow/issue-tracker.md:44-46 documents that slot addresses are unique per fleet and that a machine should run one fleet.
  - id: BR-3
    disposition: addressed
    note: |
      ownership_test.go:508-521 now tests that close and milestone-close from another checkout of the owning slot get past the ownership gate to the review.
findings:
  - id: new
    severity: Important
    family: ownership-shadow-consumer
    title: |
      observe isClaimantCheckout rewrites MatchClaimant without the machine check, so a same-named slot on another machine counts as claimant
    detail: |
      This is the 2nd finding in family ownership-shadow-consumer. Rule: every claimant/ownership judgment goes through issue.MatchClaimant; no consumer re-implements the slot/worktree comparison. checkpoints.go:66 should build a Claimant from in.Me (Machine, Repository) plus h.Address and h.Path and call MatchClaimant (false when in.Me is nil). Add a test case for a claimant on another machine. Today a pair:1 claim from laptop B marks laptop A's pair:1 checkout IsClaimant=true, which contradicts Relation=other-workspace in the same output.
```

---

## Re-review — 2026-10-11T01:59:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 307 — Claimant identity is the slot, not the worktree path |
| repo | ariadne |
| issue file | workshop/issues/000307-claimant-identity-is-the-slot-not-the-worktree-path.md |
| boundary | whole-issue close |
| milestone | — |
| window | ab4f690039d8f0f845a14baf4d8083d628fb1895..deb5d336487b4a9b1afbb4e2de05b5c03c6ead16 |
| command | sdlc close --issue 307 |
| reviewer | claude |
| timestamp | 2026-10-11T01:59:05-07:00 |
| verdict | SHIP |

## Review

Checking that `HomeSlot` is defined the same way for observe and for the claimant identity.

```verdict
verdict: SHIP
confidence: high
```

The only open finding was BR-4, and it is fixed in a way that covers the whole problem, not just the one line it named. `isClaimantCheckout` (`cmd/sdlc/internal/observe/checkpoints.go:63-76`) now hands the decision to `issue.MatchClaimant`. It starts from the queried identity (`in.Me`), which carries the machine and repository, and replaces only its slot and path with the holding checkout's. So a slot with the same name on another machine falls through to `OwnershipForeign` at `MatchClaimant`'s machine check. When `in.Me` is unknown, the function returns false. The new table test covers the BR-4 case ("same slot, other machine" → false), along with the no-slot fallbacks. `collectHolding` now uses `id.HomeSlot()`, the same function `claimantIdentity` uses (`cmd/sdlc/claimant.go:42`). Before, the two sides worked out "slot" differently. The interdiff makes no scope change and resolves no conflicts. The rest of the window is comments, atlas text, a lesson entry and the BR-3 acceptance test that was already disposed.

1. **Strengths**
   - `checkpoints.go:72-74` builds the identity to compare from `*in.Me` and overrides only the slot and path. The machine and repository come from the real identity, so the rule set by the earlier ownership findings now holds here: ownership is decided only in `MatchClaimant`.
   - `observe.go:201` and `claimant.go:42` both work out the slot with `HomeSlot()`. That leaves one definition of "the slot that owns claims from here" (ARCH-DRY).
   - `observe_test.go:381-409` checks both the slotted and no-slot branches and both outcomes of each. That includes the machine mismatch that BR-4 asked to have tested.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `isClaimantCheckout` copies the observe `Claimant` field by field into an `issue.Claimant`, and `assemble.go:150` does the same copy in the other direction. A small converter would keep the two from drifting apart. This is cosmetic.

5. **Test coverage:** The BR-4 regression test would fail if the fix were reverted: the old `Worktree == h.Path` comparison gives true for "same slot, other machine", because the paths match. The ownership acceptance path also has coverage now (`ownership_test.go:508-521`).

6. **Architecture**
   - **ARCH-DRY: pass.** The decision goes through `MatchClaimant`, and the slot comes from `HomeSlot`.
   - **ARCH-PURE: pass.** `isClaimantCheckout` is a pure function of its inputs and is tested without any IO. The IO stays in `collectHolding`.
   - **ARCH-PURPOSE: pass.** I counted three places that decide ownership: the gates, the observe relation and observe's per-checkout claimant flag. All three now go through `MatchClaimant`. The fleet's placement of claims by recorded checkout is a decision that is written down and disclosed in the atlas, not a copy of the ownership rule.

7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      isClaimantCheckout builds the comparison from in.Me (machine and repository) with the holding checkout's slot and path, calls issue.MatchClaimant, and returns false when in.Me is nil. The test case "same slot, other machine" pins this, and collectHolding now uses HomeSlot like claimantIdentity.
findings:
  - id: new
    severity: Minor
    family: ownership-shadow-consumer
    title: |
      Claimant is copied field by field between the observe and issue types in two places
    detail: |
      checkpoints.go:72 and assemble.go:150 hand-copy the same six fields in opposite directions. A single converter would stop a new Claimant field from being dropped in one of them. Cosmetic: ownership is still decided only in MatchClaimant.
```

---

## Re-review — 2026-10-11T02:08:52-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 307 — Claimant identity is the slot, not the worktree path |
| repo | ariadne |
| issue file | workshop/issues/000307-claimant-identity-is-the-slot-not-the-worktree-path.md |
| boundary | whole-issue close |
| milestone | — |
| window | ceb4d3c6e57e9d1c4fd4409228ee127429249559..124de8b635a55d249587925391912b821d042637 |
| command | sdlc close --issue 307 |
| reviewer | claude |
| timestamp | 2026-10-11T02:08:52-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This fix review covers six files and only one finding was still open: BR-5, the Minor about the six `Claimant` fields being copied by hand. It is addressed for the whole class, not just the two sites it named. All three hand-copies now use Go's direct struct conversion:
- `cmd/sdlc/internal/observe/checkpoints.go:71`: `issue.Claimant(*c)`
- `cmd/sdlc/internal/observe/assemble.go:150`: `Claimant(recorded)`
- `cmd/sdlc/internal/fleet/claims.go:182`: `Claimant(c)`

The fleet site was not named in the finding but belongs to the same class. That conversion only compiles when the field names, types and order match exactly (struct tags don't count). I confirmed the three structs match: `issue/claimant.go:25`, `observe/types.go:103`, `fleet/claims.go:60`. So adding a field to one struct without the others now fails at compile time instead of silently dropping the field. `go build ./...` and `go vet` pass, and `go test ./internal/observe ./internal/fleet` passes. The window has no conflict resolutions. The other changes are generated ledger and review files, and they change no scope (`## Spec`, `## Revisions` or `## Done when`).

1. **Strengths**
   - The struct conversion lets the compiler enforce that the fields stay in step, which no hand-copy can do. The comments added at `observe/types.go:100-102` and `fleet/claims.go:179-181` say why the conversion is used.
   - The sweep went past the two sites the finding named and also caught the fleet converter (ARCH-PURPOSE).
2. **Critical:** none.
3. **Important:** none.
4. **Minor:** none new. Two similar copies are outside this window and predate it, so they are not in scope. `issue/release.go:26,31` copies into `ReleaseBy`, a different type. `state.go:112` builds `IssueOwner`, which deliberately holds only some of the fields.
5. **Test coverage:** this is a refactor that keeps behaviour the same. The compiler enforces the mirrored fields, and the existing observe and fleet tests pass.
6. **Architecture:**
   - **ARCH-DRY: pass.** No hand-copies remain between the three `Claimant` types.
   - **ARCH-PURE: pass.** The conversions are pure, and no IO was added.
   - **ARCH-PURPOSE: pass.** The whole class was swept.
7. **Plan revisions:** none.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      All three hand-copies (observe checkpoints.go:71, assemble.go:150, fleet claims.go:182) now use direct struct conversion, so a field added to one Claimant fails to compile instead of being dropped. The fields match exactly across issue, observe and fleet; build, vet and the observe and fleet tests pass.
```
