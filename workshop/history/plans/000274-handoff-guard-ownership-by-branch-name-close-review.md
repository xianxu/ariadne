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

---

## Re-review — 2026-10-11T10:39:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 274 — Handoff guard judges detail ownership by branch name |
| repo | ariadne |
| issue file | workshop/issues/000274-handoff-guard-ownership-by-branch-name.md |
| boundary | whole-issue close |
| milestone | — |
| window | c663f84a6ee901174cbe7403da74c97870a68557..2858f8d226353c932a5090659a9fed76a37215b1 |
| command | sdlc close --issue 274 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T10:39:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Diff read; now running the targeted tests and grepping for leftover old-policy wording.

Found leftover "owner's edits" wording in the publish summaries. Waiting on the test run.

Tests pass. Running one quick mutation check to confirm the unclaimed-restore fix has a test that would catch its removal.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Verdict: fix-then-ship.** The policy change is right and the targeted tests pass (`go test -run 'TestTransferGuard|TestDetailsVerdict|TestNoHelpCallsABranchTheOwner|TestIssueRestore|TestPublish'` → ok, 295s). One rule, `mayNotEditDetails` (`transferguard.go:153`), now decides both who may land an edit (the guard, through `changedDetails`) and who may publish one (`issue publish`, through `requireOwnedToPublish`). All three prior findings are fixed. I reverted the unclaimed-restore fix and ran the test: `TestIssueRestoreMakesAStaleFilingBranchLand/unclaimed` fails without it, so the test really guards the fix. What keeps this from a clean ship is one more case in the `doc-claims-track-policy` family. Several short summaries of `issue publish` still say only the owner publishes later edits.

### 1. Strengths
- **One rule, two writers:** a single predicate is read by the guard and by publish (ARCH-DRY passes). `requireOwnedToPublish` dropped its separate switch and now just calls it (`republish.go:211`).
- **Logic kept apart from IO:** `detailsVerdict` stays a pure function of `detailsFacts`. `TestDetailsVerdict` covers all 8 combinations of the inputs ("AHAFAAAF") with no IO (ARCH-PURE passes).
- **Both cases of each Done-when clause are tested:**
  - Unclaimed edits: accepted when the branch is up to date with main, refused when it is behind, and accepted on a direct push from local main.
  - Owner editing from another issue's branch (`000013-meddler`): accepted.
  - Another workspace's overwrite: still refused, and the refusal names the owner.
- **Refusal text is checked mechanically** (`transferguard_test.go:175-180`): no refusal may recommend `sdlc claim`, and `restore` is offered exactly when the issue isn't this workspace's own.
- **Old tests fixed for the right reason:** two tests had depended on an unclaimed card being refused. They now set a card held by another slot, which is the case they were written to protect.

### 2. Critical
None.

### 3. Important
- **Short summaries of `issue publish` still state the old rule.** This is the **3rd finding in family `doc-claims-track-policy`.**
  - **Rule:** any text saying who may publish or land an edit to published details must state the `mayNotEditDetails` policy: refused only for another workspace's issue.
  - **Fix the rule, not one sentence:** put the policy in one shared help constant, the way `ownershipGateHelp` / `{{OWNERSHIP_GATE}}` already works (`claimant.go:134`). Have the help and atlas lines use or quote it, and add a test that rejects "owner's later edits" / "owned issues'" wording in those surfaces.
  - **Instances still in the tree (5):**
    - `cmd/sdlc/helptext/issue.md:11`: "or the owner's later edits"
    - `atlas/workflow/issue-sync.md:9`: "or the owner's later edits"
    - `atlas/workflow/sdlc-binary.md:49`: "or the owner's edits in one narrow commit"
    - `atlas/workflow/sdlc-binary.md:72`: "the owned issues' details"
    - `cmd/sdlc/republish.go:56`: the name `republishOwned` and "publishes owned issues'"; only the second half of that comment was updated.
  - `helptext/issue.md:15` ("offer to a non-owner") is still literally true and can stay.

### 4. Minor
None new.

### 5. Test coverage
- Every Done-when clause is covered in both of its states.
- The unclaimed restore change fails its test when reverted (I ran this).
- Per the Log, setting `mayRestore` to always true fails `TestIssueRestoreRefusesTheOwnersOwnEdit` and the "owner is behind main" row of the guard test.
- `issue.OwnershipUnknown` covers both never-claimed cards and legacy cards claimed before #277 (no claimant recorded), so both are now editable. This matches how `ownershipGateHelp` already treats them, so it's consistent rather than a defect.

### 6. Architecture
- **ARCH-DRY:** passes. One predicate serves both writers.
- **ARCH-PURE:** passes. The verdict is a pure function, and the IO stays in `changedDetails`.
- **ARCH-PURPOSE:** passes for the code, which delivers the purpose (both refusals are lifted and the protection is kept). The written policy still doesn't come from one source, which is the Important finding above.

### 7. Plan revisions
None. The Plan and Log match the code.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      issuerestore.go:26-34 help now names foreign/unclaimed-behind cases and "hand it to its owner"; no claim advice; catalog.go:295 updated too.
  - id: BR-2
    disposition: addressed
    note: |
      abandon.go:250 now reads "Another workspace's card is refused (requireOwnedToPublish)", matching mayNotEditDetails.
  - id: BR-3
    disposition: addressed
    note: |
      mayRestore (transferguard.go:180) allows unclaimed; TestIssueRestoreMakesAStaleFilingBranchLand/unclaimed fails when reverted (verified by mutation).
findings:
  - id: new
    severity: Important
    family: doc-claims-track-policy
    title: |
      publish summaries still say only the owner publishes later edits
    detail: |
      3rd in family. Rule: every statement of who may publish/land published-details edits must state the mayNotEditDetails policy (refused only for another workspace's issue); derive it from one shared help constant (like ownershipGateHelp/{{OWNERSHIP_GATE}}) plus a test rejecting owner-only wording. Remaining instances: helptext/issue.md:11, atlas/workflow/issue-sync.md:9, atlas/workflow/sdlc-binary.md:49 and :72, republish.go:56 (republishOwned "publishes owned issues'").
```

---

## Re-review — 2026-10-11T10:52:29-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 274 — Handoff guard judges detail ownership by branch name |
| repo | ariadne |
| issue file | workshop/issues/000274-handoff-guard-ownership-by-branch-name.md |
| boundary | whole-issue close |
| milestone | — |
| window | c663f84a6ee901174cbe7403da74c97870a68557..ab95ac3bcad9c297c2b4d01560ecabe4b951f812 |
| command | sdlc close --issue 274 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T10:52:29-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I've read the review window and I'm now checking the BR-4 sites, sweeping for leftover owner-only wording, and running the tests.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The code change is correct and well tested. One policy, `mayNotEditDetails` (`cmd/sdlc/transferguard.go:157`), now decides both who may land published details and who may publish them: a foreign claimant blocks, while this workspace's own issue and an unclaimed one may be edited, always only against main's latest version. All three `## Done when` clauses have tests. Unclaimed edits succeed in the guard rows and in `TestPublishUnclaimedDetails`. The owner's edit lands from another issue's branch (`000013-meddler`). A stale overwrite is refused both for a foreign owner and, for an unclaimed issue, as behind main. The targeted test run passes: `ok github.com/xianxu/ariadne/cmd/sdlc 309.644s`.

BR-1, BR-2 and BR-3 hold up. BR-4 is only half done. The sites it listed are all fixed, and the right mechanism is in place: the `{{DETAILS_EDITORS}}` placeholder plus `TestNoTextSaysOnlyTheOwnerEditsDetails`. But that test's pattern only matches "**the** owner's (later )?edits". It misses "**an** owner's edits", and it doesn't scan README.md. So four owner-only statements are still in the tree and the test passes anyway. Fixing these is cheap and doesn't block shipping, which is why the verdict is FIX-THEN-SHIP.

1. **Strengths**
   - `transferverdict.go:26` is pure, and `TestDetailsVerdict` covers all eight combinations of its inputs. The foreign/behind split is easy to follow.
   - The guard and `issue publish` share one predicate, and restore's eligibility (`mayRestore`) is a separate predicate. Ownership and restore offers are not conflated.
   - The guard test rows now also check that a refusal never suggests `sdlc claim`, and that it offers restore exactly when the issue isn't this workspace's own (`transferguard_test.go:176-181`). This tests the purpose of the change, not just its happy path.
   - Two old tests depended on an unclaimed card being refused. They now set a foreign owner, which is the case they actually mean to protect.
   - Who may edit published details is now stated once, in help, as a placeholder, and a test enforces it.

2. **Critical:** none.

3. **Important**
   - BR-4, re-raised as not-addressed. **This is the 4th finding in family `doc-claims-track-policy`.**
     - **The rule:** any text that says who publishes published-details edits must either use `{{DETAILS_EDITORS}}` or state the unclaimed exception. `TestNoTextSaysOnlyTheOwnerEditsDetails` must match "an owner's" as well as "the owner's", and must scan README.md.
     - **Remaining instances:**
       - `cmd/sdlc/issuepublish.go:21`: the `Short` text says "first publication or an owner's edits". It reaches users through command listings.
       - `cmd/sdlc/issuemovedetail.go:41`: "also publishes an owner's later edits".
       - `README.md:23`: "or the owner's later edits (#284)". The test doesn't scan README.md.
       - `cmd/sdlc/republish.go:2`: the file header says "the owner's edits reach main".
     - **Fix sketch:** widen the pattern to `(?i)(the|an) owner's (later )?edits`, add README.md to the files the test checks, and fix those four lines.

4. **Minor**
   - `republish.go:214`: "only its owner republishes its details" is still accurate, because this message is only reached for a foreign claimant. No change needed.
   - `transferguard_test.go:59`: the comment on `refuseForeign` says the foreign refusal "offers restore". It does, but the constant only checks that the owner is named.

5. **Test coverage notes**
   - The guard rows cover all three holders: another workspace's, this workspace's own, and unclaimed. Restore is tested for both foreign and unclaimed holders. The publish tests cover an unclaimed card succeeding and a foreign card being refused.
   - The wording test is the only enforcement for the doc family, and it has a blind spot, described in item 3.
   - I did not revert any fix to confirm a test goes red. The implementor's Log records mutation checks, and the test design is consistent with them.

6. **Architectural notes**
   - **ARCH-DRY: pass.** One predicate serves both writers, and the help text has a single source.
   - **ARCH-PURE: pass.** The verdict and predicates are pure. Only `ownership()` and `ancestorOf` do IO, and they stay in `changedDetails`.
   - **ARCH-PURPOSE: flag (BR-4).** The help-text source is enforced for `issue.md`, but the `Short`/`Long` strings in Go and the README restate the policy by hand, and the test doesn't catch them.

7. **Plan revision recommendations:** none. The plan's checked items match the code.

```findings
dispose:
  - id: BR-4
    disposition: not-addressed
    note: |
      Listed sites fixed and a placeholder plus wording test added, but the test regex misses "an owner's" and skips README.md; remaining: issuepublish.go:21 Short, issuemovedetail.go:41, README.md:23, republish.go:2.
```

---

## Re-review — 2026-10-11T10:57:39-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 274 — Handoff guard judges detail ownership by branch name |
| repo | ariadne |
| issue file | workshop/issues/000274-handoff-guard-ownership-by-branch-name.md |
| boundary | whole-issue close |
| milestone | — |
| window | c663f84a6ee901174cbe7403da74c97870a68557..e9a4f1b0696d3185b58530fb07c74227e6951825 |
| command | sdlc close --issue 274 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-11T10:57:39-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the remaining policy statements across the tree for BR-4.

Checking the background test run, then writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The behaviour change is right and well tested. `mayNotEditDetails` (`cmd/sdlc/transferguard.go`) is now the one policy that both the guard (`changedDetails`) and publish (`requireOwnedToPublish`) read. The verdict was cleanly renamed from `NotOwner` to `Foreign`. The foreign refusal names the owner and no longer recommends `claim`. `mayRestore` never offers to drop this workspace's own record. The Done-when clauses are each covered: an unclaimed issue's edit lands and publishes; the owner's edit from another issue's branch lands; a foreign overwrite is refused.

What still blocks a clean SHIP is BR-4, which is not addressed. The fix put `{{DETAILS_EDITORS}}` into the `issue` help, which is good. But its enforcement is a blacklist of past phrasings (`TestNoTextSaysOnlyTheOwnerEditsDetails`), which only catches sentences already reported. A grep for the policy's words across the scanned surfaces still finds four owner-only statements that the regex misses. One of them is in the atlas section that the README links to.

1. **Strengths**
   - `transferverdict.go`: `detailsFacts.Foreign` plus `detailsVerdict` stay a small pure table, and `TestDetailsVerdict` tests it without IO (ARCH-PURE passes).
   - One predicate, `mayNotEditDetails`, is shared by both writers of published details (ARCH-DRY passes for code).
   - `TestTransferGuardOverBranchShapes` asserts on every refusal row that the refusal never says `sdlc claim`, and that restore is offered exactly when the holder is not this workspace. The family is tested, not just one instance.
   - `TestPublishUnclaimedDetails` checks that main received the edit and that publishing did not claim the issue.
   - Fail-closed is kept: a malformed claimant makes `ownership` return an error, which propagates, so a bad card is never treated as unclaimed.

2. **Critical**: none.

3. **Important**
   - BR-4 (`doc-claims-track-policy`) is not addressed. These sites still state the owner-only rule:
     - `atlas/workflow/issue-tracker.md:351`: "**Republish** (details on main): the owner only."
     - `cmd/sdlc/internal/recovery/catalog.go:253` (move-detail `Ends`): "the owner publishes later edits with `issue publish`"
     - `cmd/sdlc/internal/recovery/catalog.go:328` (issue publish `Preconditions`): "Republish: this workspace owns each card". This is now false for an unclaimed card.
     - `cmd/sdlc/helptext/set-status.md:62-63`: "lands only from the owner's checkout (#285) … or claim the issue first". An unclaimed issue now lands from anywhere, and a foreign one cannot be claimed.
     - Lesser: `republish.go:46` "one owned issue's publication".
     
     Fix sketch: patch every site above. Then change the test so it checks the rule rather than past phrasings. Either:
     - assert that each surface naming republish or later edits also contains the policy marker (`unclaimed` / `#274`), or
     - generate the catalog strings from the same help text (exported from a small shared package).

4. **Minor**
   - `ownership()` returns `OwnershipUnknown` both for "never claimed" and for "claimed before #277 with no claimant recorded". Both are now editable by anyone. That matches `ownershipGateHelp`, which already treats the two the same, but the atlas could say so.

5. **Test coverage**
   - A targeted run of `TestNoTextSaysOnlyTheOwnerEditsDetails|TestDetailsVerdict|TestPublishUnclaimed` passed.
   - The wider targeted run (`TestTransferGuard|…|TestIssueRestore|TestPublish`) took longer than the 120 s tool timeout. I moved it to the background, and it had not reported a result when I wrote this, so I did not observe it passing.
   - The guard rows cover all three holders (another workspace, this workspace, unclaimed) for both the based and behind cases.

6. **Architecture**
   - **ARCH-DRY: flag (documentation only).** The policy is still restated by hand in the atlas, the README, `AGENTS.base.md` and the catalog.
   - **ARCH-PURE: pass.**
   - **ARCH-PURPOSE: flag.** This is the BR-4 shadow-sweep: the catalog and atlas consumers are not derived from the policy text and are not enforced by the test.

7. **Plan revisions**: none needed. The Plan matches the code. The Log's round-3 claim of having "Fixed all four sites" should note the additional sites listed above once they are fixed.

```findings
dispose:
  - id: BR-4
    disposition: not-addressed
    note: |
      The regex blacklist misses remaining owner-only statements. atlas/workflow/issue-tracker.md:351 "Republish (details on main): the owner only."; recovery/catalog.go:253 "the owner publishes later edits"; recovery/catalog.go:328 "Republish: this workspace owns each card"; helptext/set-status.md:62-63 "lands only from the owner's checkout ... or claim the issue first"; minor republish.go:46 "one owned issue's publication". Fix all of them, and make the test assert that the policy is present (e.g. "unclaimed"/#274 on every republish/later-edit statement) rather than blacklisting past phrasings.
```

The wider targeted test run that was still going in the background has now passed: `ok github.com/xianxu/ariadne/cmd/sdlc 189.854s`. It covered `TestTransferGuard`, `TestDetailsVerdict`, `TestNoHelpCallsABranchTheOwner`, `TestNoTextSaysOnlyTheOwnerEditsDetails`, `TestIssueRestore` and `TestPublish`.

So the code passes its tests, but the verdict stays **FIX-THEN-SHIP**, because BR-4 is not about failing tests. The wording test passes while four places still state the old rule that only the owner may publish later edits:
- `atlas/workflow/issue-tracker.md:351`
- `cmd/sdlc/internal/recovery/catalog.go:253`
- `cmd/sdlc/internal/recovery/catalog.go:328`
- `cmd/sdlc/helptext/set-status.md:62-63`
