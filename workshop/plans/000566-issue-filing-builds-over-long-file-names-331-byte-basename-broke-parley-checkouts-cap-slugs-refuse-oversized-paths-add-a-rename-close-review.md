# Boundary Review — ariadne#566 (whole-issue close)

| field | value |
|-------|-------|
| issue | 566 — Issue filing builds over-long file names (331-byte basename broke parley checkouts); cap slugs, refuse oversized paths, add a rename |
| repo | ariadne |
| issue file | workshop/issues/000566-issue-filing-builds-over-long-file-names-331-byte-basename-broke-parley-checkouts-cap-slugs-refuse-oversized-paths-add-a-rename.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73492328ed3745236372ad859b538029f1488ca0..644766d85c579461760e8be9014524656b510df3 |
| command | sdlc close --issue 566 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T03:54:25-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking whether the other commit-building paths (commit publication, rebased patch) get around the new path-length guard.

Guard regression confirmed (test goes red without it). Now checking whether README lists issue verbs.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** This range does what #566 asks. All filers now get their slug from one capped builder. Every candidate tree on `issue-tracker` and `main` is checked by one 255-byte guard in `TrunkFile.commitSet`, before anything can be pushed. The new `sdlc issue rename` decides each step from a small pure state machine that is tested over all 16 combinations, and a stopped rename finishes when you run it again. I ran all the targeted tests and they pass: `internal/issue`, `gitx`, `tracker`, `recovery`, the `cmd/sdlc` rename, filer and refusal tests, and `TestRecoveryContractsAreProven`. I also took the `CheckPathLimits` call out of `commitSet`, and `TestUpdateManyRefusesOverlongNameBeforePush` went red (`err = <nil>`), so the guard is reachable and its test pins it. Two cheap things should be fixed before shipping. First, the rename skips its ownership and "unstarted" checks whenever the card is already at the new name, so it can move a started issue's details on `main` for someone who doesn't own it. Second, the legacy-filer test the plan promised was never written.

Done-when: current — all three bullets have tests: capped basename for details and card (`TestIssueNewCapsLongTitleBasename`, `TestFlakeFilingCapsLongTitleBasename`), refusal before push (`TestUpdateManyRefusesOverlongNameBeforePush`), and rename on both refs (`TestIssueRenameMovesCardAndDetails`). The only revision moved the recorded suite into the release gates and doesn't change scope.

### 1. Strengths
- **One place builds every candidate tree.** The guard sits in `commitSet` (`cmd/sdlc/internal/gitx/updatemany.go:192-195`), which both `UpdateMany` and `PrepareCandidate` go through, so card reservation, publication, setters and rename are all covered by one check. My revert test confirms it.
- **The state machine is separate from the IO.** `NextRenameStep` (`internal/tracker/rename.go:286`) is pure, and its test (`rename_test.go:357`) writes out the expected table rather than deriving it from the code.
- **The resume bug was found and fixed.** The rerun used to take the old details name from the card, which the first step had already moved. Details are now looked up by issue ID (`DetailNames`), and the resume test covers it (`issuerename_test.go:728`).
- **Duplicate code removed.** The "explicit slug, else Slugify, refuse empty" logic that was copied in `issue.go` and `legacymode.go` is now the single function `FileSlug` (ARCH-DRY).
- **The filer test can actually fail.** It checks against a literal 80 rather than `MaxSlugBytes` (`issuenew_test.go:454`), because a bound derived from the constant would always pass.

### 2. Critical findings
None.

### 3. Important findings
- **Rename skips its permission and status checks once the card is at the new name** (`cmd/sdlc/issuerename.go:565-570`). The comment says only an already-finished rename needs no permission. But `card.Path != newCard` is also false in the half-done state (card moved, details still under the old name). It is also false when the card and details names already differ for an issue that is `working` and owned by someone else. In those cases a non-owner can run `rename --slug <card's slug>` and the details on `main` move with no ownership or status check.
  - Fix: skip `renamable` only when the first observation is `RenameDone`. Otherwise, at least run `requireOwnedToEdit` before the `RenameDetails` step.
  - Add a refusal case for "card already moved, other owner".
  - (ARCH-ORDER / ARCH-SECURE)
- **The legacy `issue new` path has no capping test** (`cmd/sdlc/legacymode.go:90`). Plan Task 1 Step 6 promised "Legacy: one case in `legacymode_test.go`", and the issue's plan item says "tracker + legacy". No such test exists: if the legacy path went back to plain `Slugify`, no test would fail. Fix: add a legacy filer test with a 300-character title that asserts the basename is at most 6+1+80+3 bytes. (ARCH-PURPOSE)

### 4. Minor findings
- Commits built by `commitpublication.go` (legacy `issue publish --commit SHA`) reuse the source commit's tree, so they never pass through `commitSet`. The risk is low, since a working tree can't normally hold a name over 255 bytes, but the spec says "refuse to publish *any* path". It would be worth a sentence in the plan's "not guarded" list.
- `issue new --slug` now refuses slugs that aren't already normalized (for example uppercase or `a--b`), where before it accepted anything. That's an intentional behavior change, but only the help text mentions it.
- The README doesn't mention `issue rename`. The README doesn't list the other card setters either, so this is consistent; the atlas and help were updated.
- `CheckPathLimits` no longer puts the component itself in the error message, unlike the plan sketch. That's sensible (it avoids printing a 331-byte string), but the plan text is slightly out of date.

### 5. Test coverage notes
- The path guard has a unit test and an end-to-end test showing origin doesn't move; the revert check confirms the end-to-end test catches removal.
- The rename tests cover: moving with bytes unchanged, deriving the slug from the title, resume plus idempotence, and five refusals that each leave both refs unchanged.
- Gaps:
  - No test for the reverse half-state (tracker at the old name, main at the new) through the command; only the pure function covers it.
  - No test for the permission-skip path described in Important #1.

### 6. Architectural notes (ARCH-*)
- **ARCH-DRY: pass.** One slug builder and one guard. The details path is assembled inline (`path.Join(dirs.Rel[0], newStem+".md")`), but that's minor.
- **ARCH-PURE: pass.** `FileSlug`, `CheckPathLimits`, `NextRenameStep` and `DetailNames` are pure and tested without IO.
- **ARCH-PURPOSE: flag (Important #2).** Every filer does go through `FileSlug`. The legacy consumer just has no test proving it.
- **ARCH-MOCK: pass.** Tests use the existing `newTrackerRepo` setup (real git against a local bare origin); no new external dependency.
- **ARCH-CONSTRAINTS: pass.** It's an interactive CLI, and the loop is capped at 3 iterations.
- **ARCH-SECURE: flag (Important #1).** Slugs are validated before any path is built, and bytes are copied without parsing. However, the permission check is skipped based on an observation that doesn't prove the rename is finished.
- **ARCH-ORDER: flag (Important #1).** The state machine itself is sound and every state change goes through it. The permission gate sits outside the model and is keyed on the wrong state.
- **ARCH-FUNERAL: pass.** The old paths are deleted in the same commit, and nothing durable is added.
- **ARCH-EVOLVE: pass.** No persisted format changes; card bytes, and therefore the card's blob ID, are unchanged.

### 7. Plan revision recommendations
- Add to `## Revisions`: "Dropped the 'no `workshop/plans/<stem>-*` on main' rename precondition; `status: open` plus no local or remote branch covers the unstarted case." This is already recorded in the Log but not in Revisions.
- Add to `## Revisions`: the details are found on main by ID through the new `DetailNames`, which the Core concepts table doesn't list; add it as a PURE entity in `internal/tracker/rename.go`.

```findings
findings:
  - id: new
    severity: Important
    kind: defect
    family: guard-keyed-on-partial-state
    title: |
      issue rename skips ownership/status checks whenever the card already sits at the new path, not only when the rename is finished
    detail: |
      issuerename.go:565 runs renamable only if card.Path != newCard; in the half-done state (card moved, details old) or a pre-existing card/details mismatch on a working issue owned by someone else, a non-owner can move the details on main unchecked. Skip only when the first observation is RenameDone, otherwise apply requireOwnedToEdit before RenameDetails; add a refusal test for that case.
  - id: new
    severity: Important
    kind: defect
    family: consumer-without-regression-test
    title: |
      Legacy issue new slug cap has no test (plan Task 1 Step 6 promised one in legacymode_test.go)
    detail: |
      legacymode.go:90 now calls FileSlug but no legacy filer test asserts a 300-char title yields a basename within 6+1+80+3 bytes; reverting it to Slugify would pass the suite.
  - id: new
    severity: Minor
    kind: defect
    family: guard-coverage-enumeration
    title: |
      commitpublication.go publishes source-commit trees without passing CheckPathLimits; not named in the plan's not-guarded list
    detail: |
      Low risk since such trees come from a working tree, but the spec says refuse any oversized path; name the exclusion and its rationale in the plan.
  - id: new
    severity: Minor
    kind: defect
    family: plan-code-drift
    title: |
      Plan Core concepts omits DetailNames and Revisions omits the dropped plans-on-main precondition
    detail: |
      Both are recorded only in the Log; add a Revisions entry and a PURE row for tracker.DetailNames.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"013f0c7e-f49c-4cef-a668-38ce8cc73e61","repository":"github.com/xianxu/ariadne","issue_id":"000566","recipe":"milestone-review","round":1,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/52672d8f33959adc","reviewed_head":"644766d85c579461760e8be9014524656b510df3","slot":"ariadne:5","requested_at":"2026-10-23T10:48:47.02757Z"},"phase":"finished","pid":87650,"started_at":"2026-10-23T10:48:47.178372Z","finished_at":"2026-10-23T10:54:25.447005Z","session_id":"7a0af5bc-09bd-4246-a505-a9982bc61815","outcome":"returned"}] -->

---

## Re-review — 2026-10-23T04:02:54-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 566 — Issue filing builds over-long file names (331-byte basename broke parley checkouts); cap slugs, refuse oversized paths, add a rename |
| repo | ariadne |
| issue file | workshop/issues/000566-issue-filing-builds-over-long-file-names-331-byte-basename-broke-parley-checkouts-cap-slugs-refuse-oversized-paths-add-a-rename.md |
| boundary | whole-issue close |
| milestone | — |
| window | 73492328ed3745236372ad859b538029f1488ca0..ec1d2831d02f2e7df3f3e7a25134663b4bcc080f |
| command | sdlc close --issue 566 |
| reviewer | claude (anthropic; author anthropic) |
| isolation | fenced |
| timestamp | 2026-10-23T04:02:54-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Checking the prior findings' fixes now: I'll revert each one in the disposable checkout and confirm its regression test fails.

Now reverting BR-1 (check only on card step) and BR-2 (Slugify in legacy) to confirm RED.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

All four findings from round 1 are fixed. I checked the three behaviour fixes by reverting each one in this checkout: its regression test fails with the fix removed and passes with it in place. The plan now contains the `DetailNames` row and the Revisions entry it was missing. The code delivers all three Done-when items: the slug cap is shared by all three filers, oversized paths are refused before anything is pushed, and the rename finishes or no-ops when rerun. Targeted tests pass at head (internal issue/gitx/tracker/recovery, plus the rename, filer and legacy tests in `cmd/sdlc`). One new Important finding should be fixed before landing. `sdlc issue rename` checks that the issue is renamable against an earlier snapshot, not inside the compare-and-swap that moves the card. A `claim` that lands between the two can leave a started issue whose card has the new name while its branch keeps the old one. The plan's State design says this case needs no action. The fix is cheap.

Done-when: current — the Spec/Revisions changes (dropping the plans-on-main precondition, guarding `PublishCommitChecked`, adding `DetailNames`) don't widen what the three Done-when items promise, and each item has a test: `TestIssueNewCapsLongTitleBasename`/`TestFlakeFilingCapsLongTitleBasename`/`TestLegacyRepositoryIssueNewCapsLongTitleBasename`, `TestUpdateManyRefusesOverlongNameBeforePush`/`TestPublishCommitRefusesOverlongName`, and `TestIssueRenameMovesCardAndDetails`.

**Prior findings**

| ID | Disposition | Evidence |
|---|---|---|
| BR-1 | addressed | `issuerename.go:680` now runs `renamable` before any step that moves a file, including the details step. With the check put back to card-step-only, the subtest `another owner finishing a half-done rename` fails with `err = <nil>, want "only its owner"`. |
| BR-2 | addressed | `legacymode_test.go:30`. With `Slugify` restored in `legacymode.go`, the test fails. |
| BR-3 | addressed | `commitpublication.go:149` checks the diff between base and the merged tree. With that check removed, `TestPublishCommitRefusesOverlongName` fails at `pathlimit_test.go:73`. The plan now names the unguarded writers. |
| BR-4 | addressed | The plan's Core concepts table now has a `DetailNames` row, and Revisions records why the plans-on-main precondition was dropped. |

### 1. Strengths
- **One slug builder.** `issue.FileSlug` (`internal/issue/scaffold.go:231`) replaces two copies of the same logic. It is shared by `runIssueNew`, the legacy filer and rename. `Slugify` is deliberately left uncapped, which keeps gate family slugs stable.
- **The path check sits in one place.** `CheckPathLimits` runs at the top of `commitSet` (`updatemany.go:192`), the one function that builds every tracker and main candidate tree. Merge-tree publication runs the same check separately.
- **The rename state machine is pure.** `tracker.NextRenameStep` is a pure function of what both refs hold. `TestNextRenameStep` covers all 16 combinations against an expected table written out independently of the code.
- **The resume test caught a real bug.** Before the fix, the details step looked up the old name from a card that had already moved. Details are now found on main by ID (`DetailNames`).
- **Real tests.** They run against real git with a local bare origin. The stateful `publicationGit` fake gained a `diff` case for the new call.

### 2. Critical findings
None.

### 3. Important findings
- **`cmd/sdlc/issuerename.go:680-688` and `internal/tracker/rename.go:399-426` (ARCH-ORDER): the guard is checked against a stale snapshot.**
  - **What happens:** `renamable` (status `open`, the owner or unclaimed) is evaluated on the snapshot `cur`. `MoveCard` then runs `MovePath` inside `UpdateMany` on a freshly fetched tip. It copies whatever bytes it finds there and never checks them against `cur`.
  - **Failure case:** a concurrent `sdlc claim` (or a claim by another workspace) lands between the snapshot and the push. The rename then moves a card that is now `working` or owned by someone else. The issue branch keeps the old stem, and after this the rename refuses (status `working`), so repairing it needs hand-written plumbing again.
  - **Plan gap:** the plan's State design row "concurrent claim between steps … none needed" only covers the gap after the card step, not the one before it.
  - **Fix sketch:** have `MoveCard`/`MovePath` take the expected source bytes (`cur`'s raw card) and refuse with "card changed since it was checked; rerun" when the pinned tip differs. Then re-checking on rerun re-runs `renamable`. Add a test that changes the card's status inside the gap, for example through a `MovePath` precondition hook, or by asserting that a changed source is refused.

### 4. Minor findings
- The legacy cap test (`legacymode_test.go:42`) only fails because the OS reports "file name too long" on a 331-byte name. A title that slugifies to 100–250 bytes would get past an uncapped filer. Using a title of about 150 characters would make the test's own byte-length check the oracle.
- In the half-done state, `renamable` derives the branch name from the moved card's path, so it checks for a branch under the new stem, not the old one. This is harmless today because a claim would also flip the status, but it is worth a comment.
- Other tree writers (`rebasedpatch.go`, `abandon.go`, `assetindex.go`, `recoveryref.go`) aren't in the plan's not-guarded list. They build trees from branches or working trees that already hold the names, so the risk is low. A one-line mention in the plan is enough.

### 5. Test coverage notes
- With each fix reverted: BR-1 subtest RED, BR-2 RED, BR-3 RED. All three pass at head.
- The resume and idempotence path is covered, including that a finished rename commits nothing.
- The ownership/status race before the card step has no test (see the Important finding).

### 6. Architectural notes
- **ARCH-DRY: pass.** FileSlug replaces the two copies in the filers. `MovePath` is shared by both rename steps.
- **ARCH-PURE: pass.** `FileSlug`, `CheckPathLimits`, `NextRenameStep` and `DetailNames` are pure and tested without IO. `runIssueRename` is thin glue.
- **ARCH-PURPOSE: pass.** All three filers use `FileSlug`, and every candidate writer is guarded or named as an exclusion.
- **ARCH-MOCK: pass.** Tests use real git against a local bare origin, and the stateful fake models `diff`.
- **ARCH-CONSTRAINTS: pass.** Interactive CLI with two pushes; the loop is bounded at 3.
- **ARCH-SECURE: pass.** The explicit slug is validated before any path is built (no `/`, `..` or uppercase), and bytes are copied, not parsed.
- **ARCH-ORDER: flag.** The guard is evaluated outside the compare-and-swap (Important finding above).
- **ARCH-FUNERAL: pass.** The old paths are deleted in the same commit, and nothing new is left behind.
- **ARCH-EVOLVE: pass.** No persisted format changes.

### 7. Plan revision recommendations
- **State design:** split "concurrent claim" into before the card step (now refused because the observed card bytes no longer match; rerun re-checks) and after it (compare-and-swap retry). Record this as a Revisions entry once the fix lands.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      renamable now gates every non-Done step; reverting to card-only turns the half-done another-owner subtest RED.
  - id: BR-2
    disposition: addressed
    note: |
      TestLegacyRepositoryIssueNewCapsLongTitleBasename added; RED with Slugify restored in legacymode.go.
  - id: BR-3
    disposition: addressed
    note: |
      PublishCommitChecked checks the base-to-merged diff; removing it turns TestPublishCommitRefusesOverlongName RED.
  - id: BR-4
    disposition: addressed
    note: |
      Plan Core concepts lists DetailNames; Revisions records the dropped plans-on-main precondition.
findings:
  - id: new
    severity: Important
    kind: defect
    family: guard-observed-outside-cas
    title: |
      issue rename checks renamable (status open, ownership) on a stale snapshot; MoveCard moves whatever bytes the fresh tip holds
    detail: |
      issuerename.go:680-688 evaluates renamable on cur, then MovePath (rename.go:399) copies the fresh tip's card without comparing it to cur. A claim landing in between gets its now-working card renamed, its branch stranded on the old stem, and further renames refused. The plan's State design wrongly says no action is needed. Fix: pass the expected source bytes into MovePath and refuse on mismatch so a rerun re-checks; add a test that changes the card in the gap.
```

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"62ba3204-8a7d-4db1-b1e4-9032fec37274","repository":"github.com/xianxu/ariadne","issue_id":"000566","recipe":"milestone-review","round":2,"attempt":1,"agent":"claude","root":"/Volumes/TestScratch/review-checkouts/214ccabd26f40ca0","reviewed_head":"ec1d2831d02f2e7df3f3e7a25134663b4bcc080f","slot":"ariadne:5","requested_at":"2026-10-23T10:58:55.171001Z"},"phase":"finished","pid":47353,"started_at":"2026-10-23T10:58:55.295634Z","finished_at":"2026-10-23T11:02:53.919166Z","session_id":"4d933977-0093-4217-9444-18cb77687338","outcome":"returned"}] -->
