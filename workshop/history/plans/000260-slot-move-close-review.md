# Boundary Review — ariadne#260 (whole-issue close)

| field | value |
|-------|-------|
| issue | 260 — sdlc move: move the current branch to another slot |
| repo | ariadne |
| issue file | workshop/issues/000260-slot-move.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f27d185111d247ad86e33aade2630dfc6c2088b..b2c3ee8899c5614c6dab98ad2a35957d163f0b3f |
| command | sdlc close --issue 260 |
| reviewer | claude |
| timestamp | 2026-09-28T12:32:28-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The branch delivers `sdlc move [:N]` as the Spec describes. The refusal rules live in the pure `checkMove` and `untrackedCollisions` functions, and they have unit tests with no IO. `observeMove` gathers both slots twice, and the move is refused if the two readings differ. The switches run in the right order: the source goes back to its resting branch first, then the target takes the issue branch. The failure path for the second switch has a real-Git test. The help text, the root command list, the README, AGENTS.base.md and the atlas all point to `sdlc move`, and the atlas now keeps only what the operator still decides. `go test ./cmd/sdlc -run 'Move|UntrackedCollisions|WorkspaceProcedure|Help'` passes (121s). One stated contract is not met, and the fix is cheap: after the move, the code checks that the source is on its resting *branch*, but not that the resting *HEAD* is unchanged. The Spec requires both.

**1. Strengths**
- `cmd/sdlc/moveplan.go:30`: every refusal rule sits in one pure `checkMove(moveFacts)`, with a table test for each rule (`moveplan_test.go`). This is a clean ARCH-PURE split.
- `cmd/sdlc/move.go:53`: comparing the two observations with `reflect.DeepEqual` checks everything `checkMove` looked at. A hand-picked field list could miss a change in HEAD, branch, untracked files or state.
- `move_test.go` `TestMoveRefusals`: each refusal case compares a snapshot of status, HEAD and refs across all three slots, so "nothing changed" is proven, not assumed.
- `TestMoveSecondSwitchFails`: an ignored file that the preflight cannot see triggers the real Git `--no-overwrite-ignore` guard. The test then checks that the branch ref, both resting branches and the ignored file are all intact.
- `gitOperationInProgress` now takes a git function and a root, and `issue move-detail` and `move` share it instead of each keeping a copy (ARCH-DRY).

**2. Critical findings**
None.

**3. Important findings**
- `cmd/sdlc/move.go:63`: the post-move check passes `head: ""` for the source, so it only confirms the source's *branch*. The Spec says "verify … the source is on its unchanged resting HEAD", and the old atlas procedure said the same. The fix: record the source's resting-branch SHA in `moveFacts` during `observeMove` (`git rev-parse <from.Resting>`), then pass it as `check.head` for the source row. The tests assert this only through `assertMoved`, which is test-side; the command itself never enforces it.

**4. Minor findings**
- `move.go:108`: if the target's resting branch has no configured upstream, the move stops with an error. This matches #248's "stop if no single configured upstream resolves", but it is missing from the refusal list in `helptext/move.md`.
- `move.go:66`: if the post-move check fails, the error does not say that both switches already ran. An operator reading it could think nothing changed.
- `untrackedCollisions` compares paths byte for byte. On macOS's case-insensitive filesystem, `Feature` and `feature` get past the preflight. Git's switch still refuses the overwrite, so the failure path handles it. Noting it only.
- No real-Git test covers the staged-change or dirty-submodule refusal. Both are covered only at the `checkMove` level, which is acceptable because `observeMoveSide` uses the same status parser for every change.

**5. Test coverage notes**
Every Done-when item has a test:
- moves to `:0` and to `:2`
- each listed refusal, with a no-change snapshot
- scratch files in the target survive the move
- a failed second switch leaves the refs intact
- `--dry-run` changes nothing

Replacing the two #248 tests, which ran the atlas shell block, is correct: that shell block no longer exists. The upstream-missing refusal and the head-mismatch branch of the post-move check have no tests.

**6. Architectural notes**
- **ARCH-DRY: pass.** The git-operation check is shared, and the procedure is no longer written out in both the atlas and code.
- **ARCH-PURE: pass.** Rules are in `checkMove`; `observeMove` and `runMove` are the thin IO layer.
- **ARCH-PURPOSE: pass.** All five steps and the carried-over checks are in the binary. The post-move build stays manual, which the operator chose (Revision 2).
- **ARCH-MOCK: pass.** Git is exercised through real-Git worktree fixtures, which is this repo's established convention for git-backed workspace behavior.
- **ARCH-CONSTRAINTS: pass.** This is a one-shot CLI with bounded git calls. `ls-tree -r` times untracked files is O(n·m), which is fine for the scratch-file counts involved.
- **ARCH-SECURE: pass.** Git output is parsed through `gitx.ParseStatusZ` and NUL-split `ls-tree`. Arguments are passed as argv arrays, not built into strings, and no secrets are involved.
- **ARCH-ORDER: pass with the Important note above.** The only state involved is the two worktrees across two switches, and the code handles each outcome explicitly:
  - The first switch fails: nothing has moved.
  - The second switch fails: the branch is intact, and the error names the retry.
  - The window between the preflight and the switch is narrowed by the second observation.

  The missing resting-HEAD check is the one gap in the "confirmed success" evidence.
- **ARCH-FUNERAL: pass.** The command creates nothing durable: no files or refs, only a checkout switch.

**7. Plan revision recommendations**
None needed if the resting-HEAD check is added. If it is deliberately left out instead, add a `## Revisions` entry saying the post-move check covers the source's branch only.

```findings
findings:
  - id: new
    severity: Important
    family: postcondition-verify-incomplete
    title: |
      Post-move verify does not check the source's resting HEAD is unchanged
    detail: |
      move.go:63 passes head "" for the source row, so only the branch name is checked; the Spec requires "source is on its unchanged resting HEAD". Record rev-parse of from.Resting in moveFacts during observeMove and verify it after the switch.
  - id: new
    severity: Minor
    family: help-refusal-list-incomplete
    title: |
      helptext/move.md omits the no-configured-upstream refusal
    detail: |
      observeMove errors when the target resting branch has no upstream; the help refusal list does not mention it.
  - id: new
    severity: Minor
    family: error-states-partial-progress
    title: |
      Post-move verify failure does not say both switches already ran
    detail: |
      move.go:66 error reads like a precondition failure; say the move ran and what state each slot is in.
```

---

## Re-review — 2026-09-28T12:38:33-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 260 — sdlc move: move the current branch to another slot |
| repo | ariadne |
| issue file | workshop/issues/000260-slot-move.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f27d185111d247ad86e33aade2630dfc6c2088b..aea641f83cc14e7c7922018ddab507a680c43669 |
| command | sdlc close --issue 260 |
| reviewer | claude |
| timestamp | 2026-09-28T12:38:33-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

I recommend fixing one thing before close: the BR-1 check has no test. `sdlc move` meets the Spec and the Done-when list. The pure refusal rules are in `checkMove`, so they can be tested without Git. `observeMove` only reads the two slots and `runMove` only switches them. The real-Git tests cover each refusal the Spec lists and show a refused move changes nothing. The move-related tests pass at HEAD aea641f8 (`go test ./cmd/sdlc/ -run 'Move|Procedure|Untracked'` → `ok … 142s`). Of the three findings from the last round, BR-2 is fully fixed. The code for BR-1 and BR-3 is correct, but neither fix has a test that would fail without it, which this gate's rules require for a behavior change. That test is cheap to write.

1. **Strengths**
   - `moveplan.go:33-64`: every refusal rule sits in one pure function. `moveplan_test.go` covers each case, including the file-versus-directory collisions in `untrackedCollisions`.
   - `move.go:52-58`: the second read of both slots is compared field by field with `reflect.DeepEqual` against the first. The comparison therefore always covers every field `checkMove` uses. Since aea641f8 that includes `RestHead` for both slots, which also tightens the "nothing changed since the preflight" check.
   - `TestMoveSecondSwitchFails` sets up a real failure of the second switch: an ignored file that the branch tracks. It then checks that the branch ref, both resting branches and the ignored file are all unchanged.
   - `gitOperationInProgress` was generalised to take a git function and a root, and `move-detail` now uses it too, so the logic is shared (ARCH-DRY).
   - The atlas section is reduced to the decisions a person still has to make. The long shell procedure it replaced is gone rather than left to drift.

2. **Critical findings:** none.

3. **Important findings:** only BR-1, which stays open (see the findings block). The fix at `move.go:68-75` is correct and runs on every move: it checks the source slot against `{from.Root, from.Resting, from.RestHead}`. But no test fails without it, because every existing fixture leaves the source's resting ref alone. A cheap test would add a `post-checkout` hook that runs only in slot 1 and does `git commit --allow-empty` on `main-slot1`. Hooks are shared across worktrees, so it fires after the first switch. The test would then assert the error contains `both switches ran`, which also covers BR-3.

4. **Minor findings**
   - BR-3 has the same gap as BR-1: the new "both switches ran" wording is not asserted by any test. The hook test above would cover it.

5. **Test coverage notes**
   - Every refusal the Done-when list names is covered, plus the missing-upstream case (`move_test.go:140`).
   - The success paths also check that scratch files in the target survive and that parked commits are reported.
   - There are no tests for a slot changing between the preflight and the switch, or after the switch. There is currently no way to inject that timing; a hook-based test is the practical option.

6. **Architectural notes**
   - **ARCH-DRY: pass.**
   - **ARCH-PURE: pass.** The rules are pure; the IO is confined to `observeMove` and the switch calls.
   - **ARCH-PURPOSE: pass.** The atlas, README, `AGENTS.base.md` and `sdlc --help` all point at `sdlc move`, and the old procedure tests were removed.
   - **ARCH-MOCK: pass, following repo convention.** Tests run real Git in temporary fixtures and nothing touches a remote.
   - **ARCH-CONSTRAINTS: N/A.** This is a single interactive command whose `git status`/`git ls-tree` calls cost about the same as an ordinary switch.
   - **ARCH-SECURE: pass.** Git arguments are passed as argv arrays, and slot paths come from workspace identity rather than being built from slot numbers.
   - **ARCH-ORDER: pass with a note.** The order is written down: preflight, second read, source switch, target switch, verify. Failure after the first switch says what state each slot is in. Only the ordering-injection test gap in BR-1 remains.
   - **ARCH-FUNERAL: pass.** The command creates nothing that persists.

7. **Plan revision recommendations:** none. The plan still matches the code.

```findings
dispose:
  - id: BR-1
    disposition: not-addressed
    note: |
      Code fix correct (RestHead observed, in DeepEqual, verified at move.go:68-75) but no test fails without it; add a slot1-only post-checkout hook that advances main-slot1 and assert the "both switches ran" error.
  - id: BR-2
    disposition: addressed
    note: |
      helptext/move.md now lists the no-configured-upstream refusal; TestMoveRefusals "destination without upstream" pins the refusal and that nothing changed.
  - id: BR-3
    disposition: not-addressed
    note: |
      Wording at move.go:71-74 is right, but no test asserts it; the BR-1 hook test covers both.
```

---

## Re-review — 2026-09-28T12:42:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 260 — sdlc move: move the current branch to another slot |
| repo | ariadne |
| issue file | workshop/issues/000260-slot-move.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8f27d185111d247ad86e33aade2630dfc6c2088b..e1104e8677e6015071ffc9245b520f4175aa9388 |
| command | sdlc close --issue 260 |
| reviewer | claude |
| timestamp | 2026-09-28T12:42:16-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

This round had two open findings, and both are fixed. For BR-1, `observeMoveSide` now records `RestHead` from `rev-parse --verify refs/heads/<resting>` (`cmd/sdlc/move.go:144-150`), and the check after the move compares the source's HEAD against it (`move.go:68-75`). I restored the old check (skip the HEAD comparison when it is empty) in a scratch copy, and `TestMoveVerifiesSourceRestingHead` then failed. With the real code it passes. So the regression test is real and reaches the fix. For BR-3, both failure messages after the switches now begin "both switches ran, but …", and the same test asserts that wording. The `-run 'Move|Procedure|MoveDetail'` subset of `./cmd/sdlc/` passes (108s). I found nothing new that blocks.

1. **Strengths**
   - `checkMove` (`cmd/sdlc/moveplan.go:34`) holds every refusal rule as one pure function over `moveFacts`. The IO shell only observes and switches.
   - The recheck just before switching compares the two observations with `reflect.DeepEqual`. That proves "nothing changed since the preflight" without restating each field (`move.go:53-58`).
   - Every refusal test takes a whole-state snapshot before and after, so it also proves nothing changed (`move_test.go:28-37`, `TestMoveRefusals`). This covers every refusal the Spec lists, plus the missing-upstream refusal.
   - `TestMoveSecondSwitchFails` uses a real ignored-file collision against `--no-overwrite-ignore` to exercise the partial-progress path, and checks the branch ref, both resting branches and the ignored file.
   - `gitOperationInProgress` now takes a git function and a root, so `move` and `issue move-detail` share it rather than keeping a copy (ARCH-DRY).

2. **Critical findings:** none.
3. **Important findings:** none.
4. **Minor findings:** none new.
   - The atlas says preserved untracked paths remain. The command does not check this after the move; `TestMoveToPrimary` covers it instead. The issue Spec's post-move check does not require it, so this is not a finding.

5. **Test coverage**
   - Every item in Done when has a real-Git test: both move paths, the refusals, scratch files surviving, the second switch failing, and dry-run.
   - The pure rules have unit tests in `moveplan_test.go`.
   - The two #248 tests that parsed the atlas shell block were removed and replaced by `move_test.go`.

6. **Architecture**
   - ARCH-DRY: pass.
   - ARCH-PURE: pass. `checkMove` and `untrackedCollisions` are pure and unit-tested.
   - ARCH-PURPOSE: pass. All five steps and every carried-over check are delivered. The atlas, `AGENTS.base.md`, README and `sdlc --help` all point at the verb.
   - ARCH-MOCK: pass. Git is exercised through real-Git fixtures, which is this repo's established seam for git.
   - ARCH-CONSTRAINTS: N/A. This is a one-shot operator command that runs a bounded number of git calls.
   - ARCH-SECURE: pass. Branch and ref values come from git and go into argv arrays, and Git forbids branch names that start with `-`.
   - ARCH-ORDER: pass. The ordering is explicit: observe, recheck, switch the source, switch the target, verify. Each failure point reports what state the slots are in: nothing moved, the second switch failed, or both switches ran.
   - ARCH-FUNERAL: pass. The command creates nothing durable; it only moves existing branch checkouts.

7. **Plan revisions:** none. The plan matches the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      RestHead is recorded in observeMoveSide (move.go:144-150) and verified after the move (move.go:68-75). TestMoveVerifiesSourceRestingHead fails once the old empty-head skip is restored in a scratch copy, confirmed.
  - id: BR-3
    disposition: addressed
    note: |
      Both failure messages after the switches now start "both switches ran, but ..." (move.go:71,74), and the new test asserts it.
```
