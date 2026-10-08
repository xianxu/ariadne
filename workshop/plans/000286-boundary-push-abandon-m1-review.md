# Boundary Review — ariadne#286 (milestone M1)

| field | value |
|-------|-------|
| issue | 286 — Boundary pushes and sdlc abandon |
| repo | ariadne |
| issue file | workshop/issues/000286-boundary-push-abandon.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 3f92e76e4940e04b23721d548148deab1874e130..b8be18062a5480d7ae4a59d079d770a4e131a4c9 |
| command | sdlc milestone-close --issue 286 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-08T16:15:52-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 delivers what it promised. There is one leased push helper (`leasedBranchPush`), shared by the tracker env and the landing runner. It is wired into start-plan, both milestone-close finalize paths, close, reconcile's close-mirror completion and `sdlc pr`. Merge deletes the remote branch, leased on the PR head. I checked the main correctness concerns and none of them are bugs:
- **Lease semantics:** `for-each-ref` returning empty correctly means "the branch must not exist on the remote".
- **Lost-response repeat:** git reports "up to date" before it checks the lease, so the repeat succeeds.
- **Merge resume:** a rerun from rest with the local branch already deleted goes `nextLandingAction` → `landingComplete` → `deleteRemoteBranch`, so the remote delete is retried.

The one thing to fix before shipping is test coverage. The judged milestone-close push (`close.go:1409`) is the path every real milestone-close takes, and no test exercises it. The end-to-end test only runs `milestone-close --no-judge`.

**Strengths**
- `boundarypush.go:44-61`: moving `pushIssueBranch` and generalizing it over `gitFn` gives the boundaries, unclaim and `sdlc pr` a single lease implementation (ARCH-DRY). `-u` is folded in, so the upstream is recorded too.
- `boundarypush_test.go:34-66`: `TestBoundaryPushLeases` runs every lease case against a real bare remote with a real second machine: first push, a rewrite, a lost response, and an unseen tip that is refused and left in place. These are tests of real behavior, not mocks.
- `deleteRemoteBranch` is leased on `pr.HeadOID`, so a push made after the PR merged is never thrown away. The "already gone" case is idempotent and tested (`boundarypush_test.go:93`).
- Each affected verb's recovery catalog `Effects` entry was updated and given a named Proof.
- The atlas section (`atlas/workflow/issue-tracker.md:268`) states the warn-versus-fail split (D2) accurately.

**Critical findings**
None.

**Important findings**
- **The judged milestone-close push has no regression test** (`cmd/sdlc/close.go:1409-1411`). `TestBoundaryVerbsPushTheIssueBranch` passes `--no-judge`, so it only reaches `milestoneclose.go:187`. Removing the `finalizeBoundaryReview` call breaks no test, because the close that follows pushes anyway. The Log says both paths were mutation-checked, but no surviving test pins this one.
  - Fix: add a `stubJudge(SHIP)` milestone-close step and call `atHead("milestone-close (judged)")` before the next close.

**Minor findings**
- `landing_test.go:770` says "a resumed landing whose remote branch is already gone still finishes", but the test only runs one merge. Either add a second `runMerge` from rest or drop the claim from the comment.
- The stale-lease hint for unclaim changed from "before handing off" to the generic "before the next push" (`boundarypush.go:56`). That is slightly less specific for unclaim's hard-fail path.
- When the remote delete in merge fails (`landing.go:491`), the warning gives no recovery step. It could name `sdlc merge --branch B --yes` (which the code path above shows retries the delete) or `git push --delete`.
- A boundary push on a network that hangs (rather than refusing) blocks the verb with no timeout. The plan's offline case (D2) only covers fast failures. This is ARCH-CONSTRAINTS, noted for the future.

**Test coverage notes**
- Covered: the lease matrix, skipping the resting branch, delete at, past and after its head, start-plan, milestone-close (no-judge), close, and a close after a rebase. `pr` with a rewritten branch and merge's remote delete are covered in the landing fixtures.
- Not covered: the judged milestone-close push (Important above), a merge resume after the local branch was deleted, and the reconcile `retryCloseMirror` push (the catalog's reconcile entry lists no proof for the #286 addition).

**Architecture**
- **ARCH-DRY: pass.** There is one lease push and one leased delete, shared across both runner types.
- **ARCH-PURE: pass.** These are thin IO helpers with no business logic buried in them.
- **ARCH-PURPOSE: pass for M1.** Every boundary named in the Spec pushes, and abandon is properly deferred to M2.
- **ARCH-MOCK: pass.** Tests use real bare git remotes, the established seam.
- **ARCH-CONSTRAINTS: pass, with the hang note above.**
- **ARCH-SECURE: pass.** No new untrusted parsing. Ref names come from local state.
- **ARCH-ORDER: pass.** The lease is the ordering guard against a second writer; a lost response and an unseen tip are both modeled and tested. Merge deletes local then remote, and the observation-driven resume reaches the remote delete again.
- **ARCH-FUNERAL: pass.** M1 removes residue: remote issue branches no longer pile up after merge.

**Plan revision recommendations**
- The Core concepts table lists `pushIssueBranch` as living in `handoff.go`, "(moved to `boundarypush.go`)". Add a Revisions note that `leasedBranchPush(gitFn, …)` is now the shared primitive and `pushIssueBranch` is a thin tracker-env wrapper, and add `leasedBranchPush` and `gitFn` as table rows.

```findings
findings:
  - id: new
    severity: Important
    family: boundary-push-path-coverage
    title: |
      The judged milestone-close push (close.go:1409) has no regression test; only the --no-judge path is exercised
    detail: |
      TestBoundaryVerbsPushTheIssueBranch runs milestone-close with --no-judge, and the close that follows pushes anyway, so removing the finalizeBoundaryReview milestonePush call fails no test. Add a stubJudge SHIP milestone-close step and assert that origin equals HEAD right after it.
  - id: new
    severity: Minor
    family: test-claim-exceeds-test
    title: |
      TestLandingDeletesTheRemoteBranch says it covers a resumed landing, but it runs only one merge
    detail: |
      Add a second runMerge from rest after the branch is gone, or drop the resume claim from the comment.
  - id: new
    severity: Minor
    family: warning-names-recovery
    title: |
      Merge's remote-delete warning names no recovery command
    detail: |
      Suggest naming `sdlc merge --branch B --yes`, which observation-resumes into deleteRemoteBranch.
  - id: new
    severity: Minor
    family: operating-envelope-unbounded-io
    title: |
      A boundary push has no timeout, so a hanging network blocks the verb despite D2's warn-only intent
```

---

## Re-review — 2026-10-08T16:23:27-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 286 — Boundary pushes and sdlc abandon |
| repo | ariadne |
| issue file | workshop/issues/000286-boundary-push-abandon.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 3f92e76e4940e04b23721d548148deab1874e130..f21fbf4a41dcfda51b24ccd7d7ea3c038c5ab93a |
| command | sdlc milestone-close --issue 286 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-10-08T16:23:27-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M1 delivers what it set out to do. There is one leased push helper (`leasedBranchPush`), and the tracker env, `sdlc pr` and the merge cleanup all share it. Start-plan, both milestone-close paths, close, and reconcile's close completion now push the branch. Merge deletes the remote branch, leased on the PR's head. Round 1's fixes hold up when checked against the code: the judged milestone-close push has its own test step, the resume claim is now tested, the merge warning names a recovery command, and the push has a time bound. One thing still blocks a clean SHIP. Reconcile's close-completion push (`closetracker.go:149`) is a boundary-push call site that no test checks, and it is the second finding in the `boundary-push-path-coverage` family. The plan's stale line numbers (BR-1) were also never fixed.

**1. Strengths**
- `boundarypush.go:58-77`: the lease reads `for-each-ref` on the remote-tracking ref, so an absent ref gives an empty lease that means "must not exist". This replaces the old `rev-parse` two-step in `handoff.go` with one simpler path. Unclaim, the boundaries and `pr` all use the same function (ARCH-DRY), wired through the `gitFn` seam.
- `TestBoundaryPushLeases` covers all four lease cases against a real bare origin: first push, a rebase force-pushed, a repeat push after a lost response, and a peer's unseen tip being refused and left in place.
- `TestBoundaryVerbsPushTheIssueBranch` now checks origin after a `--no-judge` M1 and separately after a judged M2. Removing `close.go:1409-1411` turns the M2 assertion red, because the M1-close and M2 commits sit between the two pushes. BR-2 is genuinely fixed.
- `trackerenv.go:110-115`: the process group, `Cancel` and `WaitDelay` copy gitx's existing pattern. `main.go:89`'s `NotifyContext` still cancels on SIGINT, so Ctrl-C still reaches git.
- The recovery catalog's Effects and Proofs entries were updated for every verb M1 touched.

**2. Critical findings:** none.

**3. Important findings**
- **This is the 2nd finding in family `boundary-push-path-coverage`.** BR-2 fixed one call site, but the class wasn't swept.
  - **The rule:** every `boundaryPush` / `milestonePush` / `leasedBranchPush` / `deleteRemoteBranch` call site needs an assertion on origin that fails when that one call is removed.
  - **Covered:** `startplan.go:344` (first run), `milestoneclose.go:187`, `close.go:1410`, `closetracker.go:427`, `landing.go:565` and `landing.go:490`.
  - **Not covered:** `closetracker.go:149`, the `retryCloseMirror` push during `issue recovery reconcile`. No test outside `boundarypush_test.go`/`handoff_test.go` reads the issue branch with `ls-remote`, and the reconcile contract (`catalog.go:128`) claims the push but has no proof row.
  - **Also missing:** the plan's Task 2 test, "start-plan's rerun pushes when origin lacks the branch".
  - **Fix:** add a reconcile case (an interrupted close mirror, then reconcile, then assert origin equals HEAD), plus a start-plan rerun with the remote branch deleted. Add both test names to the catalog proofs.

**4. Minor findings**
- BR-1 is still open: the plan still cites `landing.go:558`, `ghclient.go:149`, `landing.go:310` and `handoff.go:205`. `pushIssueBranch` has since moved to `boundarypush.go`, so that last anchor is now wrong too.
- `startPlanBranch` sends its push messages to stdout while every other site uses stderr. This is consistent with that function's existing `refreshLocalMirror` warning, so it's a cosmetic difference only.

**5. Test coverage notes**
- Bounded push: `TestBoundaryPushIsBounded` uses a sleeping pre-push hook, which really exercises the process-group kill.
- Merge resume: the resume-with-branch-gone case is covered. The remote-delete warning's own path (delete refused, then a retry) has no test, which is acceptable for a warn-only effect.

**6. Architectural notes**
- **ARCH-DRY:** pass (one leased push and one leased delete).
- **ARCH-PURE:** pass (thin IO glue with no logic buried in it).
- **ARCH-PURPOSE:** pass apart from the coverage sweep above.
- **ARCH-MOCK:** pass (real bare remotes, no stateless mocks).
- **ARCH-CONSTRAINTS:** pass (2-minute bound, warn-only).
- **ARCH-SECURE:** pass. The lease refuses tips this checkout never fetched, and no credentials are involved.
- **ARCH-ORDER:** pass. Lost response, rebase and peer push are all tested in sequence, and an unconfirmed push is never treated as success.
- **ARCH-FUNERAL:** pass, and better than before: remote issue branches now have a removal path at merge.
- **For M2:** `abandon`'s step 6 should reuse `deleteRemoteBranch` as it stands.

**7. Plan revision recommendations**
- Add a `## Revisions` entry that drops the line numbers in favour of function names. It should also note that `pushIssueBranch` now lives in `boundarypush.go` as a wrapper over `leasedBranchPush`, and that `deleteRemoteBranch` is called from `runDurableMerge` after `deleteLandingBranch`, not from inside it.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      Plan still cites landing.go:558, ghclient.go:149, landing.go:310, handoff.go:205 (pushIssueBranch has since moved to boundarypush.go).
  - id: BR-2
    disposition: addressed
    note: |
      TestBoundaryVerbsPushTheIssueBranch now runs a judged (stubJudge SHIP) M2 milestone-close and asserts origin == HEAD; removing close.go:1410 fails it.
  - id: BR-3
    disposition: addressed
    note: |
      TestLandingDeletesTheRemoteBranch now runs a second runMerge with --branch after the remote branch is gone.
  - id: BR-4
    disposition: addressed
    note: |
      landing.go:491 warning names `sdlc merge --branch B --yes`; the resume path reaches deleteRemoteBranch (exercised by the BR-3 test).
  - id: BR-5
    disposition: addressed
    note: |
      boundaryPushTimeout plus process-group cancel in gitRaw; TestBoundaryPushIsBounded uses a sleeping pre-push hook.
findings:
  - id: new
    severity: Important
    family: boundary-push-path-coverage
    title: |
      Reconcile's close-completion push (closetracker.go:149) and start-plan's rerun push have no regression test
    detail: |
      2nd in family. Rule: every boundary-push call site needs an origin assertion that fails when that call is removed. Covered: startplan first run, milestoneclose.go:187, close.go:1410, closetracker.go:427, landing pr and merge. Uncovered: closetracker.go:149 (no proof row in the reconcile contract) and the plan's start-plan-rerun test. Add both and list them in catalog Proofs.
```
