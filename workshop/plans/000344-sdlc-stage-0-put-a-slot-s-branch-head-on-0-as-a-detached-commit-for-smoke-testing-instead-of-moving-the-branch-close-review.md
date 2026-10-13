# Boundary Review — ariadne#344 (whole-issue close)

| field | value |
|-------|-------|
| issue | 344 — sdlc stage :0: put a slot's branch HEAD on :0 as a detached commit for smoke testing, instead of moving the branch |
| repo | ariadne |
| issue file | workshop/issues/000344-sdlc-stage-0-put-a-slot-s-branch-head-on-0-as-a-detached-commit-for-smoke-testing-instead-of-moving-the-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19dafa47ec755f12dd5a22d3c83a769b76f65a81..177c355e034af85be21c60600fe663e3aa92d252 |
| command | sdlc close --issue 344 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T14:56:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've read the code diff; now running the test suite and a mutation check.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

`sdlc stage` matches the Spec. Staging keeps the source branch. A second stage replaces the first and prints what it replaced. `--clear` returns `:0` to its resting branch. A dirty target or a detached commit that no ref holds is refused without changing anything. A move into a staged `:0` replaces the stage and removes the record. The fleet reports a staged `:0` as the hold `staged`. Every new test passes: `TestStageKeepsTheBranchAndReplaces`, `TestStageRefusalsChangeNothing`, `TestMoveReplacesAStagedSlot`, `TestStageBuildReportsRevisions`, `TestCheckStageAndUnstage`, `TestCheckMove`, `TestRecoveryContractsAreProven`, and the fleet and recovery packages.

One gap should be fixed first, and it is cheap. A detached target counts as safe to leave in two ways: sdlc staged it (the record), or some branch or remote ref contains it (`git for-each-ref --contains`). No test covers the ref half. I forced it to false and every stage and move test still passed. That half is what makes the crash-recovery claim true: a crash after the switch but before the record is written leaves `:0` on a commit the slot's branch contains, and the rerun relies on the ref check to accept it.

Two tests in the package failed when my test filter happened to match them: `TestArchiveDoneIssues*`, with "no issue tracker is reachable". This diff does not touch that code, and the failure comes from the review checkout's environment.

1. **Strengths**
   - `cmd/sdlc/stageplan.go:18-42`: all of stage's refusal rules are in one pure function, `checkStage`. It reuses move's slot observation (`moveSide`) and its untracked-file check (`untrackedCollisions`), so stage and move cannot disagree about what counts as a clean target (ARCH-DRY).
   - `cmd/sdlc/internal/fleet/staged.go:27-40` and `inventory.go:370`: the record is trusted only when it equals the detached HEAD exactly. A stale or hand-edited record can never make some other commit look safe to leave (ARCH-SECURE).
   - `cmd/sdlc/stage.go:61-82`: the effect order is explicit. It observes, checks, observes again and compares, switches, verifies, and only then writes the record. The comment explains why a crash between the switch and the write is safe.
   - `cmd/sdlc/stage_test.go:31-35`: `slotState` covers HEAD, branch, status and the record for both slots, so the refusal tests really prove that a refused run changed nothing.
   - `cmd/sdlc/stagebuild.go:17`: `readBuildInfo` is a variable, so the revision check is tested with a fake `make build` and no real compile.

2. **Critical findings:** none.

3. **Important findings**
   - **The ref half of the safe-to-leave check has no test** (`cmd/sdlc/move.go:296-302`). Mutation-confirmed above.
     - What depends on it: the recovery catalog's lost-response promise (`internal/recovery/catalog.go`, the `LostResponse` field on the `stage` entry) and the plan's statement of the effect order.
     - Fix: add a real-git row where `:0` is detached at a commit a branch contains, with no record (for example `git -C :0 switch --detach <slot branch HEAD>`). Assert that `stage`, `--clear` and `move` all accept it. That same row also tests the crash-before-record case.
     - The plan's table test also promised to cross target states with stage, clear and move. In real git it covers only three states, and none with move.

4. **Minor findings**
   - `cmd/sdlc/internal/fleet/render.go:58-60`: `"\tstaged"` adds a bare word to output whose other columns are all `key=value`. `staged=true` would be consistent.
   - The "already staged" path in `stage.go:84-85` (staging the same HEAD again, which skips the switch and rewrites the record) has no test, though the catalog's `Repeat` field describes it.
   - `stagebuild.go:36-38`: any `readBuildInfo` error is treated as "not a Go binary", so an unreadable binary is silently skipped.
   - `vcs.modified` counts untracked files, so operator scratch files in `:0` will make the "not built from the staged commit exactly" warning fire every time. The help text could mention this.
   - In the issue, the item "the xx-tl change sent to the hub" is checked, but the Log says the hub "is told at landing". The checkbox claims something that has not happened yet.
   - The durable plan's task checkboxes are all still unchecked (`- [ ]`), while the issue's Plan list is fully checked.
   - The `make build` call (`stagebuild.go:24`) runs directly, with no injectable seam (see ARCH-MOCK below).

5. **Test coverage notes**
   - The pure rules have a decision table with no git: `TestCheckStageAndUnstage` (8 target states × stage/clear, plus 4 source cases) and `TestCheckMove` (2 new rows).
   - In real git, these are covered: dirty target, operator-detached commit with no ref, other branch, replace after amend, clear, clearing twice, move into a staged slot, and the fleet hold.
   - Not covered: ref-contained detached targets (Important above), real-git refusal of a mid-operation target, and repeating a stage at the same HEAD.

6. **Architectural notes**
   - **ARCH-DRY: pass.** Stage reuses `moveSide`, `observeMoveSide`, `untrackedCollisions` and `orDetached`. `outOfTheWay` produces the shared refusal message for both verbs.
   - **ARCH-PURE: pass.** `checkStage`, `checkUnstage`, `checkMove` and `buildRevision` are pure. `runStage` and `runUnstage` only observe, switch and record.
   - **ARCH-PURPOSE: pass, with a note.** The xx-tl skill lives in ops, which workers must not edit, so that part of Done-when is handed to the hub. That is a legitimate external constraint, but the checklist should not mark it as already sent.
   - **ARCH-MOCK: pass, with a note.** Git goes through the existing `execGitRunner` seam. `make` runs directly; tests use a real `make` with a fake Makefile.
   - **ARCH-CONSTRAINTS: pass.**
     - Each detached observation by stage or move adds one `for-each-ref --contains`, which costs the most when there are many remote refs.
     - The fleet inventory adds only one `rev-parse --git-path` call per detached row.
   - **ARCH-SECURE: pass.** The record is compared byte-for-byte with HEAD, and nothing secret is involved.
   - **ARCH-ORDER: pass.** The effect order is documented, facts are compared again before the switch, the command takes the repo lock (`markMutatingCommand`), and a lost response is handled by re-running.
   - **ARCH-FUNERAL: pass.**
     - The record is one line per worktree, overwritten in place.
     - Stage, `--clear` and move remove it, and git removes it along with a deleted worktree's git dir.
     - A stale record left by a manual switch stays harmless because of the exact-match rule.

7. **Plan revision recommendations**
   - Add a `## Revisions` entry: the planned real-git table (target states × stage/clear/move) shipped as three real-git refusal cases plus the pure decision table. Either add the ref-contained row, or state the reduced scope.
   - Tick the durable plan's task checkboxes, and reword the xx-tl item to "to be sent to the hub at landing".

```findings
findings:
  - id: new
    severity: Important
    family: load-bearing-branch-untested
    title: |
      The ref-containment half of moveSide.Leavable (for-each-ref --contains) has no test; mutating it to false keeps every stage and move test green
    detail: |
      cmd/sdlc/move.go:296-302. This path is what the stage recovery contract's LostResponse and the plan's effect-order argument depend on (a crash before the record is written). Add a real-git row with :0 detached at a branch-contained commit and no record, and assert that stage, --clear and move all accept it.
  - id: new
    severity: Minor
    family: render-format-consistency
    title: |
      The fleet render adds a bare tab-separated "staged" word to key=value output (render.go:58-60)
    detail: |
      Use staged=true, or a key=value form, to keep the line format consistent.
  - id: new
    severity: Minor
    family: load-bearing-branch-untested
    title: |
      The repeat stage at the same HEAD ("already staged", stage.go:84-85) has no test
  - id: new
    severity: Minor
    family: silent-error-swallowing
    title: |
      stageBuild treats every readBuildInfo error as "not a Go binary" and silently skips it (stagebuild.go:36-38)
  - id: new
    severity: Minor
    family: checklist-claims-match-evidence
    title: |
      The issue checks "xx-tl change sent to the hub", but the Log says the hub is told at landing; the durable plan's task boxes are all unchecked
```

---

## Re-review — 2026-10-13T15:15:56-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 344 — sdlc stage :0: put a slot's branch HEAD on :0 as a detached commit for smoke testing, instead of moving the branch |
| repo | ariadne |
| issue file | workshop/issues/000344-sdlc-stage-0-put-a-slot-s-branch-head-on-0-as-a-detached-commit-for-smoke-testing-instead-of-moving-the-branch.md |
| boundary | whole-issue close |
| milestone | — |
| window | 19dafa47ec755f12dd5a22d3c83a769b76f65a81..911b8afa4437f1ab911c32d2c819d3fb329c9274 |
| command | sdlc close --issue 344 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T15:15:56-07:00 |
| verdict | SHIP |

## Review

Checking that the targeted stage tests pass, then reverting the BR-1 fix to confirm the new test catches it.

```verdict
verdict: SHIP
confidence: high
```

`sdlc stage` does what the Spec asks. A slot detaches its HEAD on `:0` and keeps its branch. A dirty or mid-operation `:0` is refused, and so is one on another branch or detached at an unreferenced commit. A second stage replaces the first and says what it replaced. `--clear` and `move` both return `:0` to rest and drop the record, and the fleet view reports a staged `:0` as the hold `staged`. BR-1 is fixed and I checked it: I hard-wired the `for-each-ref --contains` result in `cmd/sdlc/move.go:301` to false, and all three subtests of `TestStageAcceptsADetachARefKeeps` (stage, clear, move) failed. With the code restored, every stage and move test passes. BR-3 and BR-5 are addressed. BR-2 and BR-4 have their code fixes, but no test fails without them, so under the claimed-fix rule they stay open as Minors. They don't block. The xx-tl part of "Done when" lives in ops, which workers may not edit, and the issue and plan now say plainly that the hub gets it in the landing report.

1. **Strengths**
   - `moveSide.Leavable` is one rule shared by stage, `--clear` and move (`cmd/sdlc/move.go:292-303`, `moveplan.go:47`). This follows ARCH-DRY, and the fleet trusts the record only on an exact HEAD match (`fleet/inventory.go:370`).
   - The effect order is stated and tested. The record is written only after the switch is verified (`stage.go`). The crash-before-record case is now a real-git test that I confirmed goes red when the ref check is removed.
   - `checkStage`/`checkUnstage` are pure and tested on literal facts (`stageplan_test.go`). The refusal tests assert that HEAD, branch, status and record are all unchanged (`stage_test.go` `slotState`).
   - `TreeRow.validate` rejects a staged row that is not detached (`types.go:324`), so that invalid state can't be serialised.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - **Recovery catalog proofs (new).** The `stage` entry's `LostResponse` (`recovery/catalog.go:206`) promises "rerun accepts a switch whose record was not written". Its `Proofs` list doesn't name `TestStageAcceptsADetachARefKeeps`, the one test that proves it.
   - BR-2 and BR-4 have no regression tests (see the dispositions below).

5. **Test coverage**
   - Stage, move, fleet and recovery tests pass.
   - The only failure in the package run was `TestArchiveDoneIssues_MovesAndClosesGH`. It failed on tracker reachability in this environment ("no issue tracker is reachable"), so it is unrelated to this diff.
   - `notGoBinary` matches on error-message text. That is fragile across Go versions, and nothing tests it.

6. **Architecture**
   - **ARCH-DRY:** pass. `moveSide`, `observeMoveSide`, `untrackedCollisions` and `outOfTheWay` are reused. `verifySlot` is close to move's own verify loop and could be unified later.
   - **ARCH-PURE:** pass. The rules are pure, and observe/switch is a thin layer.
   - **ARCH-PURPOSE:** pass. The xx-tl consumer can't be edited from here, and the hand-off is now recorded accurately.
   - **ARCH-MOCK:** pass. Tests use real-git temp fleets, and the build-info reader is injected.
   - **ARCH-CONSTRAINTS:** pass. `for-each-ref --count=1` runs only on detached targets.
   - **ARCH-SECURE:** pass. The record is trimmed and trusted only on an exact match, and a hand-edited record can't vouch for another commit.
   - **ARCH-ORDER:** pass. The order is observe, re-observe and compare, switch, verify, then record, and the crash window is tested.
   - **ARCH-FUNERAL:** pass. The record is one line per worktree, overwritten in place, and removed by `--clear` and `move`. A stale record is harmless.

7. **Plan revisions:** none needed. The plan's boxes and Core concepts table match the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestStageAcceptsADetachARefKeeps (stage_test.go:192) goes red in all three subtests when move.go:301's ref check is forced false; verified by mutation in the review checkout.
  - id: BR-2
    disposition: not-addressed
    note: |
      render.go:59 now emits staged=true, but no render_test row covers a staged row, so nothing fails without the fix; add one row to render_test.go.
  - id: BR-3
    disposition: addressed
    note: |
      stage_test.go:66-69 asserts "already staged on" and the unchanged stage on a repeat at the same HEAD.
  - id: BR-4
    disposition: not-addressed
    note: |
      stagebuild.go:37-41 now warns on non-"not Go" errors, but no test fakes readBuildInfo returning another error; the fix is unpinned (and notGoBinary matches on error text).
  - id: BR-5
    disposition: addressed
    note: |
      Plan task boxes are ticked, and the issue's Docs item now says the xx-tl change lives in ops and goes to the hub in the landing report, matching the Log.
findings:
  - id: new
    severity: Minor
    family: checklist-claims-match-evidence
    title: |
      The stage recovery contract's LostResponse claim does not cite TestStageAcceptsADetachARefKeeps as its proof
    detail: |
      This is the 2nd finding in family checklist-claims-match-evidence. The rule: every claim a durable contract makes (plan box, catalog field) names the test that proves it. Sweep the stage Contract in recovery/catalog.go:199-213 so that each of Effects, Repeat, LostResponse and Ends maps to a listed Proof.
```
