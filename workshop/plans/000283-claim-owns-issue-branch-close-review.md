# Boundary Review — ariadne#283 (whole-issue close)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..bdea35b48f76e56d9279706b80df6bd90fe441a7 |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-02T13:07:09-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This change splits the lock from the lifecycle cleanly, and the code matches the narrowed Done-when. The model now has a separate `ownership` axis in `construct/vocabulary/issue.cue`. Its laws keep ownership events apart from lifecycle events and keep terminal statuses out of `holdable`. `open → working` is now the owner's `start`, and `pkg/vocab` exposes this with accessors that the claim, adopt, reclaim, start-plan and relocation code all read from, so nothing restates the status sets by hand. With an identity, claim leaves status alone; a legacy repo still does claim and start in one step. start-plan creates the branch first and writes the card second, so a lost card write can be fixed by re-running it. It carries edits to the issue's own details file and refuses any other dirty tracked file. change-code refuses a claimed but unstarted card, and reclaim accepts an owned `open` card. The atlas has the ownership-axis section and the terminology entry. The targeted tests (`TestStartDecision`, `TestClaimNeverMovesStatus`, `TestVerbContractTable`, `TestPlanningDirtyBlocking`, `TestShapeUnderClaimThenStart`, `TestChangeCodeRefusesUnstartedClaim`, `TestReclaimDecision`, `TestClaimDecisionOwnership`) and `pkg/vocab` pass when re-run here. Only Minor findings remain.

1. **Strengths**
   - `pkg/vocab/vocab.go:185-216`: one source of truth. `CanHoldOwner`, `OwnershipEvent("move").Statuses` and `TransitionForEvent(status,"start")` replace the old hardcoded lists in `claimdecision.go`, `reclaim.go` and `claim.go` (ARCH-DRY, ARCH-PURPOSE).
   - `cmd/sdlc/startdecision.go`: a pure decision. Its test covers every status × owner combination (none/me/other) from the model, and its expected values come from the model's `start` edge rather than copying the implementation (ARCH-PURE).
   - `cmd/sdlc/planningbranch.go:146`: `planningDirtyBlocking` is pure. Its tests cover rename and copy on either side, delete, add, another issue's details, and a details path with quotes and spaces.
   - The CUE laws `ownership-disjoint` and `holdable-nonterminal` make claim-moves-status unrepresentable in the model, not just untested.
   - `startplan.go:316`: the order is deliberate (branch, then card CAS), and an `ErrCardChanged` result gets a clear next action.

2. **Critical**: none.

3. **Important**: none.

4. **Minor**
   - `setstatus.go:274-279`: the #113 comment still says `sdlc claim` is "a cheap open→working lock broadcast". That is stale now.
   - The `start` edge declares the guard `owned`, but nothing enforces it by that name. `set-status working` on an unowned open card stamps the setter as claimant and starts the card in one step. That skips claim's check that the details are on main. This is existing #277 behaviour, but it is now a second, undocumented claim-and-start path.
   - `startplan.go:508`: the in-flight contention warning only counts `working`. A peer slot shaping a claimed `open` issue is not shown as in flight.
   - `claimant.go:162`: the relocation hint says "record it with `sdlc claim`". `claim.go:145` only finishes a relocation for `move`'s statuses, which are active only. So for an open, held card (start-plan's card write lost, then `sdlc move`) the hint leads to a refusal. This is an edge case.
   - Import grouping: `pkg/vocab` and `internal/tracker` are placed inside the stdlib import block in `changecode.go`, `reclaim.go` and `startplan.go`.

5. **Test coverage**
   - The claim/start-plan split is covered at three levels: the pure decision tables, the verb-contract table (new open/mine, open/foreign and codecomplete/unknown rows), and the end-to-end `TestShapeUnderClaimThenStart`. That e2e test checks the carried edit, the refused dirty code file (card and branch unchanged), the kept `started` stamp, and that a re-run changes nothing.
   - Not covered: start-plan's `ErrCardChanged` path (a lost or raced card write after the branch exists). Its recovery is that re-running start-plan is idempotent, and that is tested.

6. **Architecture**
   - ARCH-DRY: pass. `transitionFor` is now shared with `ProjectModel`.
   - ARCH-PURE: pass. The new decisions are pure, and the IO stays in `startPlanBranch`.
   - ARCH-PURPOSE: pass. All the narrowed Done-when items are delivered. The transfer guard's "owner" naming is explicitly handed to #285. The `owned` guard has no named implementation (see Minor).
   - ARCH-MOCK: pass. Tests use the existing tracker repository fixtures over real git.
   - ARCH-CONSTRAINTS: N/A. These are CLI verbs, and only one extra card CAS was added.
   - ARCH-SECURE: pass. The claimant parse error is surfaced, not ignored.
   - ARCH-ORDER: pass. The state is now two explicit axes in the model. The branch-then-card order is documented, and the lost-write case is handled by an idempotent re-run.
   - ARCH-FUNERAL: pass, with one note. No new durable artifact family is created. A start-plan that loses its card write leaves an issue branch behind, which belongs to the slot (as the error message says). Abandoning a claim without unclaiming is left to #284 and #286.
   - For #284: unclaim should read `OwnershipEvent("unclaim").Statuses`. Note that the model currently allows unclaim from `codecomplete`, but the round-1 Log says "not from codecomplete". The round-2 handoff decision supersedes it, so check this is intended.

7. **Plan revisions**: none needed.

```findings
findings:
  - id: new
    severity: Minor
    family: stale-claim-semantics-prose
    title: |
      setstatus.go:276 comment still describes claim as an open→working lock broadcast
    detail: |
      Since #283, claim records the owner and never moves status; the comment should say the estimate gate left claim/start-plan.
  - id: new
    severity: Minor
    family: model-guard-unenforced
    title: |
      The `owned` guard on the start edge has no named enforcement; set-status working claims and starts in one step
    detail: |
      statusDecision stamps the setter as claimant on an unowned open card, which skips claim's details-on-main readiness check. Either enforce `owned` in checkTransitionGuards or document set-status as the manual claim+start.
  - id: new
    severity: Minor
    family: in-flight-means-held
    title: |
      start-plan's contention warning counts only working cards, not open cards held by a shaping claim
  - id: new
    severity: Minor
    family: relocation-hint-status-scope
    title: |
      requireCardOwnership's moved-here hint points at `sdlc claim`, which finishes relocations only for active statuses
    detail: |
      For an open held card that was moved (start-plan's card write lost, then sdlc move), claim refuses with "claimed by" instead of finishing the relocation.
  - id: new
    severity: Minor
    family: import-grouping
    title: |
      Module imports placed inside the stdlib block in changecode.go, reclaim.go, startplan.go
```

---

## Re-review — 2026-10-02T13:18:08-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..e171d9569d3cb2c96a8d0888be7767194adcd6bd |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-02T13:18:08-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The issue's purpose is delivered, and so is every Done-when item. `issue.cue` now has a separate `ownership` axis. Two laws enforce it: ownership events must not share names with lifecycle events, and a terminal status cannot hold an owner. `claim` records the owner and leaves status alone. Only the owner can start work, and `start-plan` does it, branch first and card second, with the card write compare-and-swapped. Uncommitted edits to the issue's own details file carry onto the new branch; any other dirty tracked file blocks. `change-code` refuses an owned issue that hasn't started, and `reclaim` accepts an owned `open` card. The atlas defines the terminology. Of the five prior-round Minors, three are fixed with test evidence and two are justified deferrals to #284. One new Minor; nothing blocks the close.

**Strengths**
1. **The model is the single source for statuses.** `CanHoldOwner`, `OwnershipEvent("move").Statuses` and `TransitionForEvent(status,"start")` replace hand-written status lists in claim, adopt, reclaim, start-plan and the relocation check (`claim.go:145`, `reclaim.go:42`, `startdecision.go:24`). That passes ARCH-DRY and ARCH-PURPOSE.
2. **Tests are generated from the model.** `TestStartDecision` (`startdecision_test.go:15`) covers every status × owner (none, me, other) and computes the expected result from the model rather than a literal table.
3. **The dirty-tree check is pure and tested against tricky input.** `planningDirtyBlocking` (`planningbranch.go:145`) has cases for renames and copies on either side, paths with spaces and quotes, and another issue's details file (`planningdirty_test.go`).
4. **Ordering is safe in start-plan.** The branch is created before the card write, and the write is a CAS with an explicit `ErrCardChanged` recovery message (`startplan.go:316-339`). A lost write leaves a re-runnable state, never a `working` card with no branch.

**Prior findings**
- **BR-1:** addressed. The comment at `setstatus.go:282-286` is rewritten.
- **BR-2:** addressed. The guard is enforced in `statusDecision` (`setstatus.go:155-159`). `TestStatusDecisionRecordsOrRefusesTheClaimant` covers the unowned refusal, the `--force` waiver and the held start. The verb-contract table now runs unforced and expects "takes the lock first". Issue guards are hand-coded checks everywhere (the model has no runner registry for them), so enforcing it inline matches existing practice.
- **BR-3:** addressed. The plan's revision note 3 (`plan.md:302`) makes "in-flight = working" deliberate and hands the views of claimed-but-unstarted issues to #284.
- **BR-4:** addressed. The relocation-hint edge case is recorded in #284's Log, which is where the decision on `move` semantics for open claims belongs.
- **BR-5:** addressed. Imports are regrouped in all three files.

**Critical:** none.

**Important:** none.

**Minor**
- **`start-plan` now accepts more statuses than before** (`startdecision.go:24`). It admits any status that can hold an owner, so an owned `blocked` or `codecomplete` card now passes. Before this change only `working` passed. On an owned `codecomplete` card from a resting branch, start-plan would create a fresh issue branch from main under the details-derived name. That could be a name #148 has retired. The rule: a lifecycle verb's accepted statuses should come from the lifecycle axis (startable or already in progress), not the ownership axis. Either refuse `codecomplete` here or record the widening as intended.

**Test coverage notes:** the claim/start split is pinned at three levels:
- the pure decisions (`TestClaimNeverMovesStatus`, `TestStartDecision`);
- the verb-contract table;
- the end-to-end claim-then-start test with a carried shaping edit.

Two existing tests now pass `--force` deliberately, with a comment explaining why (`issue_test.go:111`, `planningreview_test.go:128`).

**Architecture**
- **ARCH-DRY:** pass. `transitionFor` is extracted and the project model reuses it.
- **ARCH-PURE:** pass. `startDecision` and `planningDirtyBlocking` are pure; the IO stays in `startPlanBranch`.
- **ARCH-PURPOSE:** pass. Every consumer of the status lists derives from the model.
- **ARCH-MOCK:** pass. Card updates go through the existing tracker fake.
- **ARCH-CONSTRAINTS:** N/A. The change only reads one card and one `git status`.
- **ARCH-SECURE:** pass. The claimant is parsed by the existing fail-closed `CardClaimant`, and errors propagate.
- **ARCH-ORDER:** pass. Transitions are explicit in the model; the CAS covers a second actor.
- **ARCH-FUNERAL:** pass. The plan records that an unstarted claim has no expiry, and that `unclaim` (#284) and `reclaim` release it.

**Plan revisions:** none needed. Optionally, a one-line Revision entry noting that start-plan accepts every status that can hold an owner, if that is intended.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      setstatus.go:282-286 rewritten; claim never moves status.
  - id: BR-2
    disposition: addressed
    note: |
      statusDecision enforces owned on the start edge (setstatus.go:155); covered by TestStatusDecisionRecordsOrRefusesTheClaimant and the unforced verb-contract row.
  - id: BR-3
    disposition: addressed
    note: |
      Intended per plan revision note 3 (plan.md:302); shaping-claim views handed to #284.
  - id: BR-4
    disposition: addressed
    note: |
      Recorded in #284 Log, which owns move semantics for open claims.
  - id: BR-5
    disposition: addressed
    note: |
      Imports regrouped in changecode.go, reclaim.go, startplan.go.
findings:
  - id: new
    severity: Minor
    family: verb-admission-axis
    title: |
      start-plan admits held blocked/codecomplete cards via CanHoldOwner, widening the old working-only gate
    detail: |
      startDecision (startdecision.go:24) gates on the ownership axis, so an owned codecomplete card now gets a fresh issue branch from main (possibly a retired name). A lifecycle verb's admission set should derive from the lifecycle axis; refuse codecomplete or record the widening as intended.
```

---

## Re-review — 2026-10-02T13:23:21-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 3b7315b7fde5083f036e377d41e8053193e52ef3..d8f753db416fd51ff18829024e0e776b7df057e4 |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-02T13:23:21-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

**Verdict: SHIP.** BR-6, the only open finding, is fixed. `start-plan` now decides which statuses it accepts from the `start` edge of the lifecycle model, not from the ownership model. That edge is defined in `construct/vocabulary/issue.cue:168` and leads from `open` to `working`. So `startDecision` only accepts a card that is `open` (to start it) or `working` (already started). A held `blocked` or `codecomplete` card is now refused. Before this fix, such a card could get a new issue branch from main.

I checked this rather than relying on the commit message:
- **The test would catch a revert.** Under the old `CanHoldOwner` gate, a `blocked` or `codecomplete` card owned by "me" passed. The code then found no `start` edge from that status and returned the card unchanged, with no error. The updated table test expects "nothing to plan" for those cells, so going back to the old code fails it.
- **The test passes:** `go test -run TestStartDecision` passes on HEAD.

I also looked for other commands with the same mistake (the `verb-admission-axis` family). Only `claimdecision.go` and `reclaim.go` still gate on `CanHoldOwner`. Those are `claim` and `reclaim`, which take or transfer ownership, and the model's ownership table lists exactly those statuses for them (`issue.cue:216-218`). Gating them on ownership is correct, so there is nothing left to fix.

**Strengths**
- `cmd/sdlc/startdecision.go:28`: the accepted statuses come from the model's `start` edge (`FirstTransitionForEvent("start")`), not a hard-coded list, so they will change if the model changes.
- `startdecision_test.go:32`: the test works out what to expect from the model and covers every status × owner combination.
- The help text was updated in the same commit (`start-plan.md`).

**Critical findings:** none.

**Important findings:** none.

**Minor findings**
- `startdecision_test.go:13`: the doc-comment line is longer than the lines around it. Cosmetic only, so I didn't raise it as a finding.

**Test coverage notes:** the status × owner table now covers the refused statuses (`blocked`, `codecomplete`, terminal), whoever owns the card.

**Architectural notes:** there is a general rule here. A command that moves a card through its lifecycle should decide which statuses it accepts from its lifecycle edges. A command that changes who holds a card should decide from the ownership table. `start-plan` now follows that rule.

**Plan revision recommendations:** none.

```findings
dispose:
  - id: BR-6
    disposition: addressed
    note: |
      startdecision.go:28 admits only the start edge's From/To; the TestStartDecision table now expects refusal for blocked/codecomplete owned cells, which the old CanHoldOwner gate admitted, so a revert goes red. Remaining CanHoldOwner callers are the ownership verbs (claim/reclaim), correctly on the ownership axis.
```

---

## Re-review — 2026-10-07T11:35:27-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1b11fc83f9f2ecf656c2e70eadb2aed6a28a7d88..a3628c46de2ce2ce41bcc2b4874864cced3af593 |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-07T11:35:27-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

This round adds the rebase fixup fd8aac13. It moves fleet inventory and the scheduling example onto the lock model. The code is correct. `mine()` and `ClaimAssociation.validate()` (`cmd/sdlc/internal/fleet/claims.go:169,213`) now take the claim's statuses from the model's ownership axis (`CanHoldOwner`) instead of `IsActive`, which is the right single source. The tests now treat an owned open card as a placed claim and a `done` card as holding no lock. The fleet, recovery and vocab packages pass, and so do the targeted `cmd/sdlc` tests (fleet, malformed card, lost-response reruns, recovery, start decision, the shape-under-claim e2e).

One thing should be fixed first. The recovery contract catalog (`cmd/sdlc/internal/recovery/catalog.go`) is the per-verb recovery contract that agents read, and it still describes the pre-#283 behavior of claim, start-plan and reclaim. Most seriously, it says start-plan has no remote effect, but start-plan now writes the card on the tracker. The window touched `recovery/page.go` but not `catalog.go`.

1. **Strengths**
   - `fleet/claims.go:169,213`: fleet now reads which statuses hold a claim from the model (ARCH-DRY / ARCH-PURPOSE shadow-sweep: one more consumer derives from the ownership axis).
   - `fleet/claims_test.go:40-41,136`: the tests flip the fixture classes in both directions. Owned open becomes placed, and the contract mutation now uses `done`, so the "holds no lock" refusal is still tested and not lost.
   - `recovery_proofs_test.go:66`: the lost-response rerun test now runs on the open→working start edge, the transition #283 created, instead of an edge that is no longer reachable from a fresh claim.
   - `startplan.go`: branch first, then card ("a lost card write leaves an idempotent re-run"). The ordering is explicit and the comment states the reason (ARCH-ORDER).

2. **Critical:** none.

3. **Important**
   - `cmd/sdlc/internal/recovery/catalog.go:14-18,32,49-53`: the recovery contracts are stale for the verbs #283 changed. **This is the 2nd finding in family `stale-claim-semantics-prose`.**
     - **The rule:** every rendered contract that states a verb's status effect, admission statuses or remote effects is part of that verb's surface. When a window changes the verb, the contract is swept in the same window.
     - **The enumeration** is every catalog/page entry for claim, start-plan, reclaim, set-status and change-code, plus the `page.go` Example. I found these stale entries:
       - **claim Effects** says "open → working, `started` and the claimant". Claim now records the owner and `started` only.
       - **claim Ends** says "the card leaves working". The lock now ends at release, a terminal status or reclaim.
       - **claim Repeat/Preconditions** says "A working card with no owner", but `--adopt` gates on `IsActive`.
       - **start-plan Effects** says "Nothing is pushed", and its LostResponse says "no remote effect". start-plan now does an `UpdateCard` compare-and-swap, open → working, on the tracker.
       - **start-plan Preconditions** says "the card is working … no tracked changes". It now admits an owned open card, and it carries edits confined to the issue's own details.
       - **start-plan Proofs** name no test that covers the start write. `TestStartDecision` and `TestShapeUnderClaimThenStart` should be listed.
       - **reclaim Preconditions** says "an owned working/blocked/codecomplete card". Owned open cards are now reclaimable.
       - **`page.go:69` Otherwise** still says "still open: the request may be lost". After #283 a successful claim is also `open`. The step's `assignment.relation` expectation is what tells success from failure, so the text should say "no recorded owner yet".

4. **Minor:** none beyond the above.

5. **Test coverage notes**
   - `TestRecoveryContractsAreProven` checks only that each named proof exists, not what the contract says. That is why this class can drift silently. A cheap guard would require start-plan's Proofs to include a test that exercises its card write.

6. **Architectural notes**
   - ARCH-DRY: pass. Fleet now derives from `CanHoldOwner`.
   - ARCH-PURE: pass. `startDecision`, `claimDecision` and `reclaimDecision` remain pure.
   - ARCH-PURPOSE: flag, as above. The catalog is a consumer of the lock model that the sweep missed.
   - ARCH-MOCK: pass. The tracker fakes are unchanged.
   - ARCH-CONSTRAINTS: pass, not material.
   - ARCH-SECURE: pass. `validate()` still rejects lockless statuses on parsed inventory input.
   - ARCH-ORDER: pass. start-plan writes the branch before the card, and a rerun converges.
   - ARCH-FUNERAL: pass. Nothing new is durable beyond the existing card fields.
   - For #284 (unclaim), give the catalog entry the full lifecycle from the start: release Effects, Ends, and LostResponse.

7. **Plan revision recommendations**
   - Add to `## Revisions`: "Recovery catalog contracts for claim/start-plan/reclaim (and the scheduling-example Otherwise text) updated to the lock model; start-plan's contract now records its tracker write."

```findings
findings:
  - id: new
    severity: Important
    family: stale-claim-semantics-prose
    title: |
      Recovery catalog contracts for claim/start-plan/reclaim still state pre-283 semantics; start-plan claims no remote effect
    detail: |
      2nd finding in this family. Rule: a verb's rendered recovery contract (catalog.go Effects/Preconditions/Repeat/LostResponse/Ends/Proofs plus page.go Example text) is part of the verb's surface and is swept in the same window that changes the verb. Stale now: claim Effects (open to working), claim Ends (leaves working), claim Repeat/Preconditions (a working card with no owner, though adopt gates on IsActive), start-plan Effects (Nothing is pushed) and LostResponse (no remote effect) despite its UpdateCard start write, start-plan Preconditions (card is working, no tracked changes), start-plan Proofs naming no test of the start write, reclaim Preconditions (excludes owned open), page.go:69 Otherwise (still open now also means success).
```

---

## Re-review — 2026-10-07T11:42:56-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1b11fc83f9f2ecf656c2e70eadb2aed6a28a7d88..2c59070896e09f5e4e401e1148658e8550f96f1a |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-07T11:42:56-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** The one open finding, BR-7, has been fixed. Commit `2c590708` updates the recovery catalog entries for claim, reclaim, start-plan and set-status, and the scheduling example's `Otherwise` text in `page.go`, so they now describe the lock model. They match the code: claim no longer moves status, start-plan writes the card with a compare-and-swap after creating the branch, reclaim admits an owned open card, and set-status guards `owned` on the start edge. All five newly cited tests exist. But the same commit copied start-plan's two new Proof rows into the **change-code** entry as well. Those rows say change-code starts an open card and carries details edits. It does neither: change-code now *refuses* an owned open card (`changecode.go:351-354`). The change-code entry's Preconditions also don't mention that new refusal, and its real test, `TestChangeCodeRefusesUnstartedClaim`, isn't cited. So the published contract states behaviour the verb doesn't have. It's cheap to fix, and it's the third finding in its family, so the rule needs fixing, not just this instance.

1. **Strengths**
   - `catalog.go:49-64` (start-plan): the "branch first, then card" order and the rerun behaviour (`the card is started only if it is still open`) match `startplan.go:316-341` exactly. That includes the `ErrCardChanged` refusal naming the branch left in the checkout.
   - `catalog.go:12-28` (claim): the legacy no-tracker carve-out is spelled out, and the `--adopt` precondition was widened to the model's started set. Both match what was fixed for BR-6.
   - `startdecision.go` is a clean pure core. Admission comes from the model's `start` edge (`FirstTransitionForEvent("start")`), not from hardcoded statuses.
   - `page.go:69` correctly says to judge a claim by assignment, not status, which is the one place an operator would most likely misread the new model.

2. **Critical**: none.

3. **Important**
   - `cmd/sdlc/internal/recovery/catalog.go:78-79` (change-code entry): two Proof rows were copied from start-plan and claim behaviour change-code doesn't have. The entry is also missing its own new precondition. **This is the 3rd finding in family `stale-claim-semantics-prose`.** Rule: every Proof row in a verb's catalog entry describes that verb's own behaviour, and any verb whose code changes in a window gets its whole entry checked, not just the entries a finding named. Fix:
     - Remove the two rows from change-code.
     - Add "the card is started (an owned open card refuses toward start-plan, #283)" to change-code's Preconditions.
     - Cite `TestChangeCodeRefusesUnstartedClaim` as its proof.
     - Write the rule in `workshop/lessons.md`. The lesson added this round covers sweeping, not copy-paste across entries.
     - Optionally enforce it: have `recovery_proofs_test` check that no two different verbs share an identical Proof claim string. That would have caught this one.

4. **Minor**: none new.

5. **Test coverage notes**
   - The change-code refusal is pinned by `changecode_tracker_test.go:52`. The gap is only that the catalog doesn't cite it.
   - The new start-plan proof is cited (`TestStartDecision`, `TestShapeUnderClaimThenStart`).
   - The set-status `owned` guard is waived by `--force`, as the catalog now says. That waiver is consistent with the code comment at `setstatus.go:118-120`.

6. **Architectural notes**
   - ARCH-DRY: flag. Pasting the same Proof rows into two entries is the copy-paste failure, and this time it published false behaviour.
   - ARCH-PURPOSE: flag. The BR-7 sweep was done but not checked against each verb's code; change-code's own change (a new refusal) was mis-stated.
   - ARCH-PURE: pass (`startDecision` is pure; `startplan.go` is a thin shell).
   - ARCH-MOCK: pass (tracker fakes, `loseResponses` seam).
   - ARCH-CONSTRAINTS: pass (one compare-and-swap per verb).
   - ARCH-SECURE: pass (card parse errors are surfaced, `CardClaimant` errors are wrapped).
   - ARCH-ORDER: pass (the start edge is read from the model; the branch-then-card order is explicit and its rerun converges).
   - ARCH-FUNERAL: pass (nothing new and durable beyond the review sidecars, which are archived with the issue).

7. **Plan revisions**: none. The plan matches the code; only the catalog entry is wrong.

```findings
dispose:
  - id: BR-7
    disposition: addressed
    note: |
      claim/start-plan/reclaim/set-status catalog fields and page.go Otherwise now match claim.go, startplan.go:296-341, reclaim and setstatus.go:155-158; cited tests exist.
findings:
  - id: new
    severity: Important
    family: stale-claim-semantics-prose
    title: |
      change-code catalog entry carries start-plan's Proof rows and omits its own new open-card refusal
    detail: |
      catalog.go:78-79 claims change-code "starts the owner's open card" and "carries only this issue's own details edits", but changecode.go:351-354 refuses an open card toward start-plan; Preconditions omit that refusal and TestChangeCodeRefusesUnstartedClaim is uncited. 3rd in family: rule = each catalog Proof describes its own verb's behaviour and every changed verb's whole entry is swept; consider a recovery_proofs_test check that no Proof claim string appears under two different verbs.
```

---

## Re-review — 2026-10-07T12:50:13-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 283 — Re-derive ownership from the claimant |
| repo | ariadne |
| issue file | workshop/issues/000283-claim-owns-issue-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1b11fc83f9f2ecf656c2e70eadb2aed6a28a7d88..d3874b2f3e94d0496e5b342decceac6b354c04c5 |
| command | sdlc close --issue 283 |
| reviewer | claude |
| timestamp | 2026-10-07T12:50:13-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

I recommend shipping. BR-8 is fixed. The change-code catalog entry (`cmd/sdlc/internal/recovery/catalog.go:66-80`) no longer carries start-plan's two Proof rows ("starts the owner's open card…" and "carries only this issue's own details edits"). Its Preconditions now state change-code's own refusal of a claimed card that is still open, and a new Proof row cites `TestChangeCodeRefusesUnstartedClaim` (`cmd/sdlc/changecode_tracker_test.go:52`). That test exercises the refusal at `changecode.go:351-353`.

The window also adds one scope fold-in: "a close survives a rebase" (f50b6ab3). The issue's Revisions record it, the Spec carries it as a Done-when line, and the catalog's Ends text and Proof rows were updated. Targeted tests pass, including `TestRecoveryContractsAreProven`, `TestChangeCodeRefusesUnstartedClaim`, `TestNewestCloseAfterARebase` and `TestCloseAncestorOf…`. Nothing blocks the boundary.

1. **Strengths**
   - `completeop.go:53-77`: `newestClose` keeps ancestry as the first test. It falls back to a check of which close is still on the branch only when a branch exists, and that check is asymmetric: an earlier close the rebase rewrote off the branch loses to a receipt the branch contains. A stale receipt from before the rebase, or a detached checkout, still refuses. `TestNewestCloseAfterARebase` covers all four of those cases over a small injected commit graph, with no I/O (ARCH-PURE).
   - `closeAncestorOf` (`trackerenv.go:148-157`) is scoped narrowly. Only the two close-generation call sites treat an unknown commit as preceding nothing; `ancestorOf` keeps its error for every other caller, and `TestCloseAncestorOfTreatsAnUnknownCommitAsNoAncestor` checks both behaviours.
   - Each catalog entry's Proofs now describe that entry's own verb. start-plan keeps the start-edge and dirty-carry proofs; change-code keeps the refusal proof.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - No test fails if either wiring site, `closetracker.go:418` or `issuerecovery.go:164`, goes back to `env.ancestorOf`. The two helpers are tested separately but not together; an end-to-end test of rebasing and then re-closing would pin the wiring.
   - The atlas does not mention the new supersession rule for close generations. Only the catalog's Ends text records it. One sentence in `atlas/workflow/issue-tracker.md`, next to the close/receipt prose, would cover it.

5. **Test coverage notes:** the pure decision has table coverage, and the git-backed helper has a real-repository test. The only gap is the integration wiring noted above.

6. **Architecture**
   - **ARCH-DRY:** pass. The "refuses a non-owner" Proof row appears under several verbs legitimately, because each of them runs the shared ownership gate.
   - **ARCH-PURE:** pass.
   - **ARCH-PURPOSE:** pass. The fold-in is documented as a scope revision, and BR-8 was fixed across change-code's whole catalog entry (Preconditions and Proofs both), not just the line the finding named.
   - **ARCH-MOCK:** pass. Git is reached through `gitTest`, and the test graph is injected through the existing `ancestor` seam.
   - **ARCH-CONSTRAINTS:** pass. The fallback costs at most two `merge-base` probes per close.
   - **ARCH-SECURE:** pass. Card bindings come from the tracker, and a reviewed commit this clone doesn't have now fails safe: it reads as "precedes nothing" instead of erroring.
   - **ARCH-ORDER:** pass. Supersession stays a pure decision over (binding, receipt, branch), and a stale receipt is still refused.
   - **ARCH-FUNERAL:** pass. Nothing durable is added.

7. **Plan revision recommendations:** none. The 2026-10-07 Revisions entry already covers the rebase fold-in.

```findings
dispose:
  - id: BR-8
    disposition: addressed
    note: |
      catalog.go change-code entry now lists only its own proofs plus "refuses an owned card that start-plan never started" citing TestChangeCodeRefusesUnstartedClaim (changecode_tracker_test.go:52, exercising changecode.go:351-353); Preconditions name the open-card refusal; recovery contract test passes.
findings:
  - id: new
    severity: Minor
    family: rebase-close-wiring-untested
    title: |
      No test fails if closetracker.go or issuerecovery.go reverts to env.ancestorOf instead of closeAncestorOf
    detail: |
      newestClose and closeAncestorOf are unit-tested separately; an e2e close-rebase-reclose test would pin the two call-site wirings.
  - id: new
    severity: Minor
    family: atlas-missing-surface
    title: |
      atlas does not mention the rebase-aware close-generation supersession rule
    detail: |
      Only the recovery catalog Ends text records it; add one sentence to atlas/workflow/issue-tracker.md near the close/receipt section.
```
