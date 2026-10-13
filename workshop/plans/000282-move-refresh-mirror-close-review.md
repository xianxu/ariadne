# Boundary Review — ariadne#282 (whole-issue close)

| field | value |
|-------|-------|
| issue | 282 — sdlc move refreshes the destination's details mirror after relocating the owner |
| repo | ariadne |
| issue file | workshop/issues/000282-move-refresh-mirror.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..74787863b48a8881667e5c7b0c778b43330979b9 |
| command | sdlc close --issue 282 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:04:36-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Reviewing diff; now checking how `moveRelocation` and other callers derive the details path.

Checking whether the move's mirror edit leaves the destination dirty and blocks a move back.

Writing a throwaway test to probe move-back after the refresh.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix then ship.** The core change is correct and the tests prove it. Mirror placement now has one rule, applied in one function: the details mirror is refreshed only on the issue's own branch (`cmd/sdlc/claim.go:368`). `sdlc move` refreshes the destination's details after a successful relocation, and a refusal only warns.

I reverted each fix in the disposable checkout and the matching tests failed:
- Restoring the old `onRest()` guard fails `TestMirrorRefreshesOnlyOnTheIssueBranch`.
- Removing the `refreshMovedMirror` call fails both `TestMoveRelocatesItsOwner` and `TestMoveWarnsWhenTheDestinationMirrorRefuses`.

The targeted test runs and `go vet` are clean. One real problem remains, and I confirmed it with a probe test I then deleted. The move's refresh is an uncommitted edit, so it leaves the destination slot dirty. The documented "Moving back" step (`sdlc move :N` from the destination) then refuses with "has uncommitted changes". Neither the move help nor the atlas mentions this. That should be fixed, or at least documented, before close.

### 1. Strengths
- **One place decides where the mirror is written** (`claim.go:367-372`). All five worktree writers go through it: claim, reclaim, start-plan, the card setters and move. The atlas row at `atlas/workflow/issue-tracker.md:421` lists exactly those five.
- **`TestMirrorRefreshesOnlyOnTheIssueBranch`** covers every case in both directions: the resting branch, another issue's branch, a detached HEAD (all must stay clean), and the issue's own branch (must refresh).
- **`TestMoveRelocatesItsOwner`** checks the actual `card_mirror` value against the current card blob, not just whether some text appears.
- **The refusal test** checks all three things: the move succeeds, the owner is relocated, and the hand-edited details are kept.
- **The `move` refresh only runs after a successful relocation** (`move.go:113-117`), so a failed or non-applicable relocation never touches the mirror.

### 2. Critical findings
None.

### 3. Important findings
- **`move` now leaves the destination dirty, which blocks moving back** (`cmd/sdlc/move.go:115`, `cmd/sdlc/helptext/move.md:4-11`, `atlas/workflow/workspace-branching.md:95-121`).
  - **Reproduced:** after `sdlc move :0`, `git status` in :0 shows ` M workshop/issues/000001-procedure.md`. Then `sdlc move :1` from :0 fails with `sample:0 has uncommitted changes … commit or remove them first`.
  - **Why it matters:** the Spec left a choice between committing the refresh and leaving it as a worktree edit. The worktree edit was picked without recording it in the Log. That choice clashes with move's own requirement of a clean tree in both slots, and it breaks the documented round trip of moving to :0 for testing and back.
  - **Fix, preferred:** commit the refresh as a narrow mirror commit on the issue branch, as close's mirror commit does (#275).
  - **Fix, minimum:** say in `move.md`, in the atlas `sdlc move` paragraph and in its "Moving back" bullet that the destination's details are rewritten and must be committed before moving back. Then add a Log/Revisions entry recording the choice.
  - **Other docs to update:** `move.md`'s OWNERSHIP section still describes only the claimant re-stamp and says nothing about the mirror refresh.
- **One Done-when clause is only tested on one side.** The clause "A non-applicable relocation leaves details untouched" has no test.
  - **Unattributed card:** `TestMoveLeavesAnUnattributedIssueUnknown` commits no details, so "untouched" holds trivially.
  - **Card owned by another slot:** no test at all.
  - The code is right by construction (the refresh only runs when `err == nil`), but nothing would catch a refactor that moves it.
  - **Fix:** call `commitMovedDetails` in the unattributed test and check the details file is byte-identical and the tree is clean after the move.

### 4. Minor findings
- **ARCH-DRY: `refreshMovedMirror` repeats work `moveRelocation` already did** (`move.go:130-137` vs `move.go:148-153`).
  - It repeats the cutover-marker stat, with a mismatched check: `err != nil` vs `errors.Is(err, os.ErrNotExist)`.
  - It reopens the tracker env and takes a fresh snapshot.
  - It builds the details path from `branch+".md"`, while every other caller uses `path.Join(dir, path.Base(card.Path))` (`cardsetters.go:53`, `startplan.go:316`, `reclaim.go:163`, `claim.go:136`).
  - **Fix:** do the refresh inside the `moveRelocation` flow, with the env and card it already holds (`refreshMirrorFrom` exists for that).
- **Spurious warning when the card is missing.** When `moveRelocation` returns `nil` because there is no card (`move.go:159-160`), the refresh still runs. If details exist, it then warns "mirror not refreshed: …no card". The move should skip the refresh when no card exists.

### 5. Test coverage notes
- `TestRevertedMirrorConvergesOnTheNextRefresh` still passes with the old guard. It describes behaviour that already existed (B2 c), so that is fine, since the plan claims no code change for B2 c. It is evidence for the help-text claim, not a regression test.
- The plan says the move refreshes after a relocation "or one already done" (`OwnershipMine` returns `nil`). That path is reachable but has no dedicated test.

### 6. Architectural notes
- **ARCH-DRY: flagged.** See the Minor finding about `refreshMovedMirror` repeating `moveRelocation`'s work. The placement rule itself correctly lives in one place.
- **ARCH-PURE: pass.** The guard is a pure branch comparison. The IO stays in `refreshLocalMirrorAt`.
- **ARCH-PURPOSE: pass.** I enumerated every caller of `refreshLocalMirror` and `refreshLocalMirrorAt`; all the worktree writers derive from the one guard. The committed writers (`closetracker.go:279`, the archive at `claim.go:402`) are placed separately, as the plan says.

### 7. Plan revision recommendations
- Add a Log/Revisions entry recording the choice between committing the move's refresh and leaving it as a worktree edit, and what it costs (the destination stays dirty and moving back is refused until it is committed). If the fix commits it instead, record that.

```findings
findings:
  - id: new
    severity: Important
    family: verb-leaves-tree-dirty-against-clean-precondition
    title: |
      move's worktree mirror refresh dirties the destination; the documented move-back then refuses
    detail: |
      Reproduced: after sdlc move :0 the destination shows " M workshop/issues/<id>.md", and sdlc move :1 from there fails with "has uncommitted changes". Instances: cmd/sdlc/move.go:115 (the refresh), helptext/move.md OWNERSHIP section (silent on the mirror edit), atlas/workflow/workspace-branching.md move paragraph and its "Moving back" bullet. Fix: commit the refresh narrowly on the issue branch (as close's mirror commit does, #275), or at minimum document that it must be committed before moving back, and log the decision.
  - id: new
    severity: Important
    family: done-when-clause-tested-one-side
    title: |
      "A non-applicable relocation leaves details untouched" has no test
    detail: |
      TestMoveLeavesAnUnattributedIssueUnknown commits no details, so "untouched" holds trivially; the case of a card owned by another slot has no test at all. Add commitMovedDetails to the unattributed test and assert the details are unchanged and the tree is clean.
  - id: new
    severity: Minor
    family: arch-dry-restated-precondition
    title: |
      refreshMovedMirror repeats moveRelocation's cutover check, tracker open and details-path derivation
    detail: |
      move.go:130-137 vs move.go:148-153: the cutover stat is repeated with a mismatched check (err != nil vs errors.Is ErrNotExist), the tracker is opened and snapshotted again, and the details path comes from branch+".md" while every other caller uses path.Base(card.Path) (cardsetters.go:53, startplan.go:316, reclaim.go:163, claim.go:136). Refresh inside moveRelocation's flow using refreshMirrorFrom on the card it already read.
  - id: new
    severity: Minor
    family: refresh-after-noop-relocation
    title: |
      A move with no card (moveRelocation's ErrNoCard path returns nil) still tries the refresh and warns
    detail: |
      move.go:159-160 returns nil when the card is missing, so refreshMovedMirror runs; if details exist, snap.Require fails and the move prints a spurious "mirror not refreshed" warning. Skip the refresh when no card exists.
```

---

## Re-review — 2026-10-13T13:36:03-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 282 — sdlc move refreshes the destination's details mirror after relocating the owner |
| repo | ariadne |
| issue file | workshop/issues/000282-move-refresh-mirror.md |
| boundary | whole-issue close |
| milestone | — |
| window | 08afcec554b1d5cb4ff03b84f57352274df5db1d..395975ae0dce0ab03477ed411fff4f78b2cc7c74 |
| command | sdlc close --issue 282 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-13T13:36:03-07:00 |
| verdict | SHIP |

## Review

Still reviewing; the broad test run timed out, so I'm narrowing it to the move/mirror tests.

Running the close-mirror and claim-mirror tests to confirm the shared `commitMirror` refactor didn't regress them.

```verdict
verdict: SHIP
confidence: high
```

All four prior findings are fixed, and I found no new Critical or Important issues. Move now commits the destination's mirror refresh as one narrow commit, through `commitMirror`, which was generalized from close's mirror commit. The destination stays clean, so moving back works. I checked BR-1 by reverting the fix: I swapped `commitMirror` for a plain worktree `refreshLocalMirror` write, and `TestMoveRelocatesItsOwner` failed with "the refresh left the destination dirty: M workshop/issues/000001-procedure.md". With the fix restored, it passes. The other targeted runs pass: the move and mirror-placement tests, close-mirror, claim-mirror, reclaim and start-plan. A full `cmd/sdlc` run with the `Close` pattern hit Go's default 10-minute test timeout while inside `TestCloseQuotedDocAndTestNamesDoNotCount`. That is the suite's runtime, not a failing assertion. The Done-when clauses are covered in both modes: the relocated case and the refusal case, and both non-applicable cases (unattributed, and owned by another slot).

1. **Strengths**
   - Placement is enforced in one place, `refreshLocalMirror` (`cmd/sdlc/claim.go:367`), keyed on `issue.BranchName(detailPath)`. All six worktree callers inherit the rule (ARCH-DRY pass).
   - `commitMirror` (`cmd/sdlc/closetracker.go:232`) is one shared narrow-commit path for close and move, and only the commit subject differs. This is reuse, not a copy (ARCH-DRY pass).
   - `movedCard` (`cmd/sdlc/move.go:153`) unifies the cutover, tracker and card read. The details path is now built from `path.Base(card.Path)`, the same way other callers build it.
   - The tests use real git and pin behaviour: a clean tree, the HEAD subject, the current `card_mirror` OID, and an actual move back.
   - The Spec's open decision (commit vs worktree edit) is settled and recorded in the Log, and both the help text and the atlas reflect it.

2. **Critical:** none.

3. **Important:** none.

4. **Minor**
   - `ownership_test.go` `assertMovedDetailsUntouched`: the subject check compares both `"#1: details\n"` and the trimmed form. Comparing `strings.TrimSpace(subject)` alone is enough.
   - `claim.go:368`: the `env.branch == ""` guard is redundant, because `BranchName` is never empty. It is harmless and does document the detached-HEAD case.
   - `move.go:168`: the no-card path in `moveRelocation` (BR-4) has no explicit test. The code is clearly correct.

5. **Test coverage:** BR-1 has a regression test that I confirmed fails without the fix. BR-2 now commits details in both non-applicable fixtures and checks for a clean tree and no mirror commit. One case is untested: the already-mine relocation (OwnershipMine → nil → refresh). It is low risk.

6. **Architecture:**
   - ARCH-DRY: pass. One worktree writer and one committed writer.
   - ARCH-PURE: pass. The new logic is thin IO glue around existing pure projection (`refreshMirror`).
   - ARCH-PURPOSE: pass. The issue's broadened purpose is delivered at the choke point, not just move's slice. All worktree callers derive the rule from that one function.
   - For later work: move's mirror commit is never pushed. That is consistent with move being local, but a later verb's push will carry it.

7. **Plan revisions:** none needed. The Log already records the decision to commit the refresh and the round-1 dispositions.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      refreshMovedMirror commits via commitMirror; reverting it to a worktree refresh makes TestMoveRelocatesItsOwner fail with the destination dirty (move back is asserted); help and atlas updated.
  - id: BR-2
    disposition: addressed
    note: |
      Unattributed and another-slot tests commit details and assert a clean tree plus no mirror commit (assertMovedDetailsUntouched).
  - id: BR-3
    disposition: addressed
    note: |
      movedCard shares the cutover, tracker and card read; the details path now uses path.Base(card.Path). The re-snapshot is needed to read the post-relocation card.
  - id: BR-4
    disposition: addressed
    note: |
      movedCard returns ok=false on ErrNoCard, so refreshMovedMirror returns "" with no warning.
findings:
  - id: new
    severity: Minor
    family: redundant-test-assertion
    title: |
      assertMovedDetailsUntouched compares the subject twice (raw with newline and trimmed)
    detail: |
      ownership_test.go assertMovedDetailsUntouched: strings.TrimSpace(subject) != "#1: details" alone is enough. Only instance in the window.
```
