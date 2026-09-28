# Boundary Review — ariadne#256 (whole-issue close)

| field | value |
|-------|-------|
| issue | 256 — issue migrate: refuse a branch that archives an issue main still has active |
| repo | ariadne |
| issue file | workshop/issues/000256-migrate-refuse-branch-archive.md |
| boundary | whole-issue close |
| milestone | — |
| window | f1f03ebd6cfda3d8baa70509b2df895c9a69cadc..cf9d95f359e946dc2c61600f3cb99fc3c1fa7147 |
| command | sdlc close --issue 256 |
| reviewer | claude |
| timestamp | 2026-09-27T23:10:24-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

The fix is correct and both new tests fail when it's removed. I'm returning FIX-THEN-SHIP because the same gap is still open in the post-cutover branch reconcile.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

**Summary.** The dry run now refuses a branch that deletes an issue file while main still has that issue active. It stays quiet in two cases: the branch still has an active file for the same ID (a slug rename), or main has archived the issue too. The branch edits come from the diff against the merge base (`cmd/sdlc/issuemigrate.go:262-266`), not the diff against main's tip. So a branch forked before main created an issue does not look like it deletes that issue, and there is no false positive. A branch that landed by squash or rebase shows as a deletion, but main has archived the issue too, so it stays quiet.

I checked that the tests pin both halves by editing a scratch copy and restoring it afterwards. Without the rename check (the `kept` map), the unit test fails on `feat-g` and the end-to-end test fails on the `000002-two` rename. Without the new case at all, `feat-e` is no longer refused and the unit test fails.

Both `## Done when` clauses are met. The refusal names the branch, and a landed archive is exercised in both its forms: a branch that renamed the file, and a branch whose issue main archived too.

The one thing between this and SHIP: the post-cutover branch reconcile (`cmd/sdlc/issuemigrate.go:569-571`) still skips any deleted details file as "its own archive move". So a branch the dry run never saw can still archive an issue whose card stays open, which is the exact failure this issue is about.

**Strengths**
- `migration.go:197-201` builds the `kept` map per branch name, so a local branch and its remote copy each decide on their own, and the existing grouping still merges them into one refusal naming both branches.
- The new `b.Raw == nil` case comes first in the `switch`. Before, a deletion was dropped with `continue`, so no deletion can now reach the parse-and-reconcile case below it.
- The unit test replaced the fixture that encoded the gap (`feat-e`) rather than adding next to it, and added fixtures for both silent paths: `feat-g` (rename) and `feat-h` (main archived it too).
- The end-to-end test uses real git moves (`git mv` into history, and a slug rename), so the merge-base diff is exercised.
- The atlas step 4 now includes the new refusal.

**Critical findings.** None.

**Important findings**
- **`cmd/sdlc/issuemigrate.go:569-571` still has the bug this issue fixes.** The post-cutover branch reconcile treats any removed details file as "its own archive move". That covers a branch in another clone, or one never pushed to the publication remote, which the dry run could not see. If that branch archived an issue whose imported card is still active, reconcile passes it silently: the card stays open, and landing the branch later conflicts with the converted details. That is the nous #48 scenario after the cutover instead of before it.
  - Fix: when `!present`, add the same rename check as the dry run. Refuse unless the imported card is done, or `changed` still holds another active file for that ID. Add a test for both outcomes.
  - This is the only other instance of the family in the tree: `grep "own archive move"` finds only this line.

**Minor findings**
- The `kept` map is keyed by `[2]string` (branch, id). A small named struct would read better, but it is local to the function, so this is fine as is.
- The rename check only looks for another active file for the ID on the same branch; it does not check that file's content. That is fine, because the new file goes through the reconcile case on its own.

**Test coverage notes.** The unit test covers refuse, rename-silent and main-archived-silent. The end-to-end test covers refuse and rename-silent with real git. The squash- or rebase-landed deletion (main archived it too) is covered only at unit level via `feat-h`; that is acceptable because the input shape is the same. Both packages pass: `go test ./internal/tracker/ -run TestPlanTrackerMigration` and `go test . -run TestIssueMigrate`.

**Architecture**
- **ARCH-DRY: pass.** The new refusal reuses the existing `branchProblem` grouping, with no duplicated code.
- **ARCH-PURE: pass.** The decision lives in the pure `PlanTrackerMigration`, and the input gathering is unchanged.
- **ARCH-PURPOSE: flag** (the Important finding above). The purpose is that an unlanded archive of an active issue must not leave its card open forever. The dry run is one of two places that decide what a branch's deletion means; the reconcile path is the other, and it is still unguarded.

**Plan revision recommendations.** Add a `## Revisions` entry that either extends the scope to the reconcile path, or explicitly defers it with a reason. Plain deferral is weak here, because the reconcile path is the same purpose.

```findings
findings:
  - id: new
    severity: Important
    family: branch-deletion-assumed-own-archive
    title: |
      Post-cutover branch reconcile still treats any deleted details file as the branch's own archive move
    detail: |
      cmd/sdlc/issuemigrate.go:569-571 skips a removed details file ("its own archive move") without checking the imported card. A branch the dry run never saw (another clone, never pushed) that archived a still-active issue reconciles silently, and the card stays open. This is the #256 scenario after cutover. Family sweep: this is the only remaining instance (grep "own archive move"). Fix: refuse unless the card is done or the branch keeps other active details for the ID. Test both outcomes.
```

---

## Re-review — 2026-09-27T23:13:29-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 256 — issue migrate: refuse a branch that archives an issue main still has active |
| repo | ariadne |
| issue file | workshop/issues/000256-migrate-refuse-branch-archive.md |
| boundary | whole-issue close |
| milestone | — |
| window | f1f03ebd6cfda3d8baa70509b2df895c9a69cadc..6152b1faa8af3c8ff08cdde80cd39359b38b5d5d |
| command | sdlc close --issue 256 |
| reviewer | claude |
| timestamp | 2026-09-27T23:13:29-07:00 |
| verdict | SHIP |

## Review

Reviewed the diff and reconcile path; writing the verdict now.

```verdict
verdict: SHIP
confidence: high
```

This closes BR-1 at the rule level, not just at the one site. A single predicate, `tracker.RemovalArchivesActive(keptOther, closedOnMain)`, now decides both paths that handle a deleted details file. The dry run (`migration.go:222-226`) passes it "main has no active file for the ID". Reconcile (`issuemigrate.go:579-585`) passes "the imported card is terminal". Both pass "the branch still has other details for the ID", so renames stay silent. `grep "own archive move"` now only hits the updated test comment. Every Done-when clause has a test: the unlanded archive is refused and names the branch (dry-run e2e plus unit `feat-e`), and a landed archive stays silent (unit `feat-h`). Reconcile has a regression test for the refusal (`TestIssueMigrateReconcileRefusesAnUnlandedArchive`), and that test fails without the fix. One small test gap remains, listed under Minor.

**Strengths**
- One shared rule and one reason string (`ArchivesActiveReason`) serve the dry run and reconcile, so the family is fixed in one place (ARCH-DRY passes).
- `RemovalArchivesActive` is a pure bool function. The IO (the ls-tree listing and parsing the card) stays in the thin `runMigrateReconcile` shell (ARCH-PURE passes).
- The unit test `migration_test.go:159-163` replaces the old `feat-e` fixture, which had encoded the gap. It also pins both silent outcomes: a rename (`feat-g`) and main archiving the issue too (`feat-h`).
- The reconcile test checks that a refused reconcile leaves HEAD unchanged.
- The atlas step 4 now covers the new refusal.

**Critical:** none.

**Important:** none.

**Minor**
- In reconcile, the silent path for a card that is closed on the tracker has no test. The test's silent case is the rename (`kept`). If `cardClosed` always returned false, no test would fail. Fix: close #1's card after apply (or seed a done card), then assert that reconcile does not refuse on `ArchivesActiveReason`.
- When the imported card doesn't parse, `cardClosed` (`issuemigrate.go:680-686`) returns false and quietly turns the parse error into "open". That fails safe (it refuses), but the refusal message would then wrongly say "still open on the tracker".

**Test coverage:** The dry run is covered by the unit test and the e2e test (both outcomes). Reconcile covers refusal and the rename. The closed-card silent case is missing, as above.

**Architecture:** ARCH-DRY passes, ARCH-PURE passes, and ARCH-PURPOSE passes: both consumers of the removal rule derive from the shared predicate, and the shadow sweep found no remaining restatement.

**Plan revisions:** none. The `## Plan` log already records the BR-1 disposition.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      issuemigrate.go:579-585 now refuses via shared tracker.RemovalArchivesActive (kept + terminal card); TestIssueMigrateReconcileRefusesAnUnlandedArchive pins the refusal; grep "own archive move" finds no other code site.
findings:
  - id: new
    severity: Minor
    family: test-both-outcomes
    title: |
      Reconcile's closed-card silent path (cardClosed true) has no test
    detail: |
      The reconcile test's only silent case is a rename (kept). Stubbing cardClosed to false would stay green. Add a case where #1's card is done and assert that reconcile does not refuse with ArchivesActiveReason.
  - id: new
    severity: Minor
    family: parse-error-coerced-to-state
    title: |
      cardClosed maps an unparseable card to open, so the refusal wrongly says the issue is still open
    detail: |
      issuemigrate.go:680-686 ignores the ParseCard error. This fails safe, but the message is inaccurate. Consider surfacing the parse error as its own refusal.
```
