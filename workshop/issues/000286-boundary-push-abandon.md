---
id: 000286
status: working
deps: [ariadne#283]
github_issue:
created: 2026-10-02
updated: 2026-10-08
estimate_hours: 2.52
card_mirror: 'e31a4406c39887c424a8dfc5840ddba68c1e827c' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-08T14:22:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
---

# Boundary pushes and sdlc abandon

## Problem

Part of project `claimant-ownership`. Started work is only recoverable after a handoff or a forced `reclaim` if it is on origin. Ending an issue as `wontfix`/`punt` today flips only the card and leaves the local branch, the remote branch and the owner behind.

## Spec

- **Boundary pushes:** push the issue branch at every sdlc boundary (`start-plan`, `milestone-close`, `close`, `unclaim`). The owner is the only writer (claim lock + #272 no-stacking), so rebasing then `push --force-with-lease` is safe. `issue sync` stays local.
- **Cleanup:** `merge` deletes the remote issue branch.
- **`sdlc abandon --issue N --as wontfix|punt --reason …`:**
  - Requires that you own the issue (or claim it first); refuses on a dirty tree.
  - Writes the reason to `## Log` and commits it.
  - Pushes the branch tip to `refs/ariadne/abandoned/NNNNNN`, records that ref on the card, and deletes the issue branch locally and on origin.
  - Narrowly publishes the final details to main and archives them; the card goes terminal, and the owner stays as attribution.
  - The work is kept for both `wontfix` and `punt`. Reopening restores the branch from the archive ref.
- `set-status` to `wontfix`/`punt` from an active status redirects to `abandon`.

## Done when

- After each boundary verb, origin's issue branch equals local HEAD; a rebase followed by the next boundary force-pushes with lease.
- `merge` leaves no remote issue branch.
- `abandon` leaves no issue branch locally or on origin, an archive ref holding the tip, a terminal card, and the details archived on main. Reopening a punted issue restores the branch at that tip.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module      design=0.1 impl=0.2
item: smaller-go-module      design=0.05 impl=0.12
item: greenfield-go-module   design=0.3 impl=0.32
item: greenfield-go-module   design=0.3 impl=0.28
item: smaller-go-module      design=0.05 impl=0.16
item: atlas-docs             design=0.05 impl=0.04
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
item: milestone-review       design=0.0 impl=0.14
design-buffer: 0.15
total: 2.52
```

Items, in order: the boundary push helper and its four call sites; merge's remote delete and pr's lease push; `sdlc abandon` (six convergent steps); the reopen restore (seven steps plus un-archive); the `Abandoned` record, the terminal-status archive policy and the set-status redirect; atlas; the M1, M2 and close reviews. Impl is 40% of the v2 table midpoints (v3.1); the design buffer is 15% because the plan is thorough.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

- [x] Claim, run start-plan, and design against the Spec and the project PRD (`workshop/projects/claimant-ownership.md`); size the flow at change-code.
- [x] M1 — Boundary pushes at start-plan, milestone-close, close (and reconcile's close completion); lease push in `pr`; merge deletes the remote branch. Plan Tasks 1–3.
- [x] M2 — `sdlc abandon`, the set-status redirect, and a reopen that restores the branch and un-archives the details. Plan Tasks 4–7.

Durable plan: `workshop/plans/000286-boundary-push-abandon-plan.md`.

## Log

### 2026-10-08
- 2026-10-08: closed — make test green except sandbox-only processgroup. Done-when: TestBoundaryVerbsPushTheIssueBranch + TestStartPlanRerunPushes + TestReconciledClosePushes (origin==HEAD after each boundary; rebase then next boundary force-pushes with lease); TestLandingDeletesTheRemoteBranch (merge leaves no remote branch); TestAbandonStartedWork (no local/remote branch, archive ref at tip, terminal card, details archived on main); TestReopenRestoresAnAbandonedIssue (punted issue restored at that tip, details un-archived, lands). Every push site and abandon/reopen step mutation-checked.; review verdict: FIX-THEN-SHIP
- 2026-10-08: closed M2 — make test green (except sandbox-only processgroup); abandon: StartedWork/OpenIssue/Refusals/Decision/RerunResumes/KeepsTheBranchsPlans/OverwritesAnOrphanArchiveRef; reopen: RestoresAnAbandonedIssue/StopsOnAConflictThenResumes/RerunAfterAFailedPush/RerunVariants; set-status: RedirectsStartedWorkToAbandon, AbandonedWorkLeavesOnlyByReopen; every fix mutation-checked; review verdict: SHIP
- 2026-10-08: closed M1 — make test green (except sandbox-only processgroup); each boundary push site pinned by its own mutation-checked test: TestBoundaryVerbsPushTheIssueBranch (start-plan, milestone-close --no-judge and judged, close, rebase+re-close), TestStartPlanRerunPushes, TestReconciledClosePushes, TestLandingPRPublishesARewrittenBranch, TestLandingDeletesTheRemoteBranch (+resume); TestBoundaryPushLeases, TestBoundaryPushIsBounded; review verdict: FIX-THEN-SHIP

Claimed in ariadne:2. Survey: `pushIssueBranch` (handoff.go) is the lease push to reuse; nothing pushes at start-plan/close today; milestone-close makes no commit (it pushes HEAD as-is, D3); the durable merge (`gh pr merge` without `--delete-branch`) leaves the remote branch; set-status has no wontfix/punt guard; no reopen un-archives details. Operator decision: reopen of an abandoned issue restores the branch AND un-archives the details (merge main, move history/X back), so it lands cleanly; done-reopen's same gap goes to a follow-up. Boundary push failures warn, never fail the verb (D2); the set-status redirect can't be forced (D8).

### 2026-10-02

M1 built: `boundarypush.go` (`boundaryPush`, `leasedBranchPush`, `deleteRemoteBranch`, all over a `gitFn` so the tracker env and the landing runner share one leased push/delete). Pushes at start-plan, milestone-close (both paths), close, and reconcile's close completion; `sdlc pr` (durable) uses the leased push; merge deletes the remote branch leased on the PR head. Every push site and both landing changes are mutation-checked. Pinned that a repeat push after a lost response succeeds (git reports up to date before checking the lease). Two fixtures assumed no remote branch and were updated to the new world: the #301 prune variant now force-pushes the rebase before gc, and the observe lifecycle's hand-landing deletes the remote branch as merge now does. make test green except the sandbox-only processgroup test.

M1 review round 1 (FIX-THEN-SHIP, BR-2 Important): the judged milestone-close push had no test of its own (my mutation disabled both paths at once). The test now runs a --no-judge M1 and a judged M2 and checks origin after each. Minors fixed in the same round: the landing test also runs a resumed merge with the remote branch gone; merge's remote-delete warning names `sdlc merge --branch B --yes`; boundary pushes are bounded (`boundaryPushTimeout`, 2m). The bound first failed: git died on time, but a hook's child held the output pipe. So `trackerEnv.gitRaw` now runs git in its own process group, killed whole on cancel, with `WaitDelay`, the same pattern as gitx's trunk writes (`TestBoundaryPushIsBounded`; passes sandboxed and unsandboxed).

M2 Tasks 4/4a/5: `Abandoned` record (ref, branch, head; all or none), `archivedDetails` accepts terminal statuses, `sdlc abandon` shipped with its contract. The first run found the remote branch is normally an earlier boundary push (behind the kept tip), so step 6 deletes it when the kept tip contains it. Rerun detection (the note commit; a resume recognised from the card before the branch checks) is mutation-checked across four interruption points, including a rerun from the resting branch.

M2 Task 6: the set-status redirect (not --force-able) and the reopen restore (`restoreAbandoned` before the card write; after it, `finishReopen` commits the mirror refresh, pushes and drops the archive ref). The first reopen run found the merge conflicts on the details: the archive commit's mirror refresh drops similarity below git's rename threshold, so main's archive reads as delete-plus-add and the branch's note as modify/delete. Restore resolves a conflict that touches only the issue's own details and plans by taking main's archived copy, and leaves any other conflict to the agent, with the card unchanged. Follow-up #305 filed (reopening a done issue doesn't un-archive) and published. The TempDir cleanup flake #284 logged (a detached auto-gc writing into a fixture repo at cleanup) got worse with this branch's extra pushes: 1 failure in 4 running alone. The test binary now disables auto-gc and auto-maintenance (`testfix.QuietBackgroundGit`), and it passes 6/6 and in the full suite.

M2 review round 4 (FIX-THEN-SHIP; BR-7 and BR-8 Important): all eight fixed. BR-7: started work's plans now archive from the kept tip, not main's older copy, and a reopen brings that version back. The reopen also had to create `workshop/plans/` again, because the merge had removed its last file. BR-8: README entry. BR-9: a resume checks ownership. BR-10: plans and history roots come from the environment in both verbs. BR-11: plan table. BR-12: `TestReopenRerunVariants` (after the branch is recreated; after the card write), and a `set-status working` rerun finishes an unfinished reopen. BR-13: a kept branch leaves terminal only by the restoring reopen. BR-14: `remoteRefTip`. Each code fix is mutation-checked. One scripted edit silently missed its anchor (the helper wasn't inserted); vet caught it, and that is the lessons.md rule.

Close review (FIX-THEN-SHIP, BR-9): the resume's ownership check had no test of its own. `TestAbandonResumeRefusesAnotherWorkspace` interrupts after main's archive, via a pre-push hook that rejects only the branch deletion. That is the one point where nothing later stops another workspace. Removing the guard makes the foreign resume succeed, so the test fails.
