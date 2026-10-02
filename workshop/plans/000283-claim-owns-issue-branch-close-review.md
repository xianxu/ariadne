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
