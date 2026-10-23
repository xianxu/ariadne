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
