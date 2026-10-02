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
