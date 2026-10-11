# Boundary Review — ariadne#274 (whole-issue close)

| field | value |
|-------|-------|
| issue | 274 — Handoff guard judges detail ownership by branch name |
| repo | ariadne |
| issue file | workshop/issues/000274-handoff-guard-ownership-by-branch-name.md |
| boundary | whole-issue close |
| milestone | — |
| window | c663f84a6ee901174cbe7403da74c97870a68557..960e5cc44467bc291bb0e5c8c544237d70f3ed96 |
| command | sdlc close --issue 274 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T10:24:01-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've reviewed the diff; now checking the abandon caller and stale doc statements, then running tests.

Running the targeted and full sdlc test suites.

Full suite is slow; running the targeted subset in the foreground instead.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The change works: one predicate, `mayNotEditDetails` (`cmd/sdlc/transferguard.go:152`), now drives both writers of published details. Those are `changedDetails`, used by pr, push, merge and restore, and `requireOwnedToPublish`, used by publish and abandon's archive. An unclaimed card can now be edited, and main's latest version is still checked for concurrency. A foreign claimant is still refused, and the refusal now names the owner and no longer suggests `claim`. I checked the `TestDetailsVerdict` truth table (`AHAFAAAF`) row by row against `detailsVerdict`. All three Done-when clauses have tests, and the targeted suite passes (`TestTransferGuard|TestDetailsVerdict|TestNoHelpCallsABranchTheOwner|TestIssueRestore|TestPublish|TestAbandon|TestHelp`, ok in 521s). One thing stands in the way of SHIP. The docs sweep missed `sdlc issue restore --help`, which still says the guard's offer is for "a checkout that does not own the issue" and tells the user to `sdlc claim --issue N` to keep the edit. That is exactly the advice this issue removes.

1. **Strengths**
   - Following ARCH-DRY, the guard and publish share one policy predicate (`transferguard.go:152`, `republish.go:214`) instead of two switches. This also fixes an existing mismatch: abandon's resume path already accepted an unclaimed card, but its archive `Allow` refused one.
   - `TestDetailsVerdict` covers all 8 fact combinations exhaustively, and the facts are pure booleans with no IO (ARCH-PURE).
   - `TestTransferGuardOverBranchShapes` now uses a three-way `holder` enum. It has rows for unclaimed rewrite after merging main, unclaimed direct push from main, and unclaimed stale re-add (behind). It also checks that no refusal mentions `sdlc claim`.
   - The pair#300 case (the owner's edit on another issue's branch) has its own row (`transferguard_test.go:120`).
   - Two older fixtures depended on "unclaimed is refused" by accident. They now set a foreign owner on purpose, with comments saying so.

2. **Critical findings**: none.

3. **Important findings**
   - **`cmd/sdlc/issuerestore.go:25-33`**: the restore help text was not updated. It says restore is offered "from a checkout that does not own the issue". It still says "To keep the edit instead, `sdlc claim --issue N` and land it as the owner", but claim can't take an issue another workspace holds, and that is the only case where restore is now offered. It still says "In the owner's own checkout restore refuses", but restore now also refuses an unclaimed issue that is behind main (the `verdictBehind` branch at :77). Rewrite it to match the new `detailsRefusal` wording: restore is for another workspace's issue, the alternative is to hand the edit to the owner (or the operator runs `sdlc reclaim`), and it refuses when this checkout may edit the issue (its own or unclaimed). Everything else in this family is already fixed in the window: `helptext/issue.md`, `atlas/workflow/issue-tracker.md`, `AGENTS.base.md`, the `detailsRefusal` string, and the `transferguard.go` header. The two remaining instances are this help text and the abandon comment below. `ownershipGateHelp` (`claimant.go:134`) belongs to the continuation gate, not detail editing, so it is correctly unchanged.

4. **Minor findings**
   - `cmd/sdlc/abandon.go:250`: the doc comment on `archiveAbandoned` still says "Only its owner publishes." Its `Allow` now also accepts an unclaimed card. Same family as above.
   - `issuerestore.go:77`: for an **unclaimed** issue whose branch carries a stale copy (for example, a filing branch after publication), restore is now refused in favour of "merge main and resolve". Before, restore was offered. This matches the plan ("Restore is only offered for another workspace's issue"), and merging main is a workable path. Still, a user who explicitly asks to drop an unclaimed stale copy is refused. Consider allowing restore for unclaimed issues, or at least test that refusal on purpose.

5. **Test coverage notes**
   - All three Done-when clauses are covered:
     - Unclaimed edit on main, then publish: `TestPublishUnclaimedDetails`, which also checks that publishing doesn't claim the issue, plus the guard row for an unclaimed direct push.
     - Owner editing from a branch not named for the issue: the `000013-meddler` row.
     - Stale-owner overwrites still refused: the `heldByOther` rows and the `TestTransferGuardFollowsTheSlot` row for another slot.
   - No test covers `issue restore` on an unclaimed issue (the behaviour change in the Minor finding above).
   - My full `go test ./cmd/sdlc/` run hit go test's default 10-minute timeout, and the package took over 600s. That timeout is not evidence of a failure. Only the targeted subset above was confirmed green. I did not repeat the implementor's mutation check, but the unclaimed rows are reachable: the `unclaimed` holder sets no claimant, so `ownership` returns `OwnershipUnknown`.

6. **Architectural notes**
   - ARCH-DRY: pass. One predicate, two consumers. I found no duplicated block.
   - ARCH-PURE: pass. `detailsVerdict` and `mayNotEditDetails` are pure, and the IO (`ownership`, `ancestorOf`) stays in `changedDetails`.
   - ARCH-PURPOSE: pass, except for the stale restore help text, which is still a hand-maintained statement of the old policy. That is the Important finding above. `OwnershipUnknown` also covers legacy cards claimed before #277, so those are now editable too. That matches the existing "no owner" wording in `ownershipGateHelp`.

7. **Plan revision recommendations**: none. The plan's checklist matches the code. Once the restore help is fixed, the Docs item should mention `issuerestore.go` Long text alongside the other docs.

```findings
findings:
  - id: new
    severity: Important
    family: doc-claims-track-policy
    title: |
      `sdlc issue restore --help` still recommends `sdlc claim` and describes the old non-owner policy
    detail: |
      cmd/sdlc/issuerestore.go:25-33 says restore is offered from a checkout that does not own the issue, says to claim to keep the edit (claim cannot take a foreign-held issue, which is now the only case restore is offered), and says it refuses only in the owner's own checkout, but it now also refuses unclaimed+behind (issuerestore.go:77). Family instances in window: this help text, plus the abandon.go:250 comment (Minor). Rewrite to match the detailsRefusal wording.
  - id: new
    severity: Minor
    family: doc-claims-track-policy
    title: |
      archiveAbandoned doc comment still says "Only its owner publishes."
    detail: |
      cmd/sdlc/abandon.go:250; its Allow (requireOwnedToPublish) now also accepts an unclaimed card.
  - id: new
    severity: Minor
    family: restore-scope
    title: |
      restore refuses to drop an unclaimed issue's stale copy (it used to be offered); there is no test for it
    detail: |
      issuerestore.go:77 treats unclaimed+behind like the owner's own work. This matches the plan, but a user who explicitly asks to drop an unclaimed stale copy is refused toward merging main; consider allowing it, or pin the refusal with a test.
```
