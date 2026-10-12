# Boundary Review — ariadne#308 (milestone M1)

| field | value |
|-------|-------|
| issue | 308 — issue new publishes card and details in one go |
| repo | ariadne |
| issue file | workshop/issues/000308-issue-new-publishes-card-and-details-in-one-go.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..5b128e69aa622fdfa1ee3876e5691501d4c82b30 |
| command | sdlc milestone-close --issue 308 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T09:24:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Inspecting the CLI/tracker diffs is done; now checking receipt-driving call sites, mirror staleness after claim, and running targeted tests.

Targeted tests pass; skimming the docs and atlas diffs before writing the verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: medium
```

M1 does what the issue asks. `issue new` reserves the card, then publishes the details to main through a receipt-driven `creation.main` stage that shares `DetailPublication` with the transfer. It writes nothing to the checkout. A numbered slot claims what it files through claim's own decision and token, and `:0` leaves the issue unclaimed. Section presence now depends on status, `move-detail` is a hidden deprecated alias, and the docs and atlas were updated in the same range. `go vet` is clean. The tracker, issue and recovery packages pass. The targeted `cmd/sdlc` run (`IssueNew|Creat|Recovery|MoveDetail|Publish|LandingInstance|NoText|TrackerFullSlot|Fetch|Help`) passes too.

Nothing blocks the boundary. Two cheap gaps remain:
- **Stale classification:** `Receipt.NeedsSourceCheckout` still treats creation stage 1 as a local effect. The plan said it should become `false`, and leaving it causes spurious "must be finished from the checkout on X" refusals.
- **Untested legacy path:** the reconcile branch that releases old-style receipts has no test. The plan named `TestRecoveryLegacyCreationReceiptPointsAtPublish`, but only the predicate is tested.

### 1. Strengths
- **One shared helper:** `cmd/sdlc/internal/tracker/detailpublish.go:14`. `DetailPublication` is the single first-publication candidate, and transfer and creation both call it. `confirmedMain` was lifted out of `TransferOp` the same way (ARCH-DRY delivered as planned).
- **Main races handled:** `createop.go` `Refresh` tells a reservation race (allocate a new ID) apart from a main race (keep the intent and re-prepare). `TestCreationMainRaceRecomputesOnMovedMain` proves both commits survive.
- **Real git in tests:** the tests run against a bare origin. `assertUntouched` and `assertCheckoutUntouched` check HEAD, the index and the working tree including untracked files. That pins Done-when 1 and 2 directly, not through mocks.
- **Pure decisions:** `creatorClaims` (`issuenewclaim.go:22`) is pure and covered by a table test. The skeleton exemption is one edit in `CheckSectionsPresence` that every caller picks up, with a test built on the real scaffold render (`structural_test.go:228`).
- **Wording pinned by a test:** `TestNoTextSaysIssueNewWritesDetailsLocally` checks the help tree, atlas, AGENTS.base.md, README and helptext against pre-#308 phrasing.

### 2. Critical findings
None.

### 3. Important findings
- **`NeedsSourceCheckout` still describes the retired local stage** (`cmd/sdlc/internal/tracker/store.go:129-132`). It returns `Stage <= 1` for creation, and its comment still says "materializing created details". The plan's Core concepts say it becomes `false`, because no creation stage acts locally now.
  - **Effect:** a creation whose main push was uncertain can't be reconciled from another branch or worktree. `issuerecovery.go:131` refuses with `ErrForeignCheckout`, and `recovery list` says "finish from <branch>", even though `creation.main` needs only the shared object database.
  - **Fix:** return `false` for creation. Keep the legacy case working: either check `LegacyCreationDetail` before the foreign-checkout guard, or treat legacy receipts specially. Add a test that reconciles a stage-1 `creation.main` receipt from another branch.
  - (ARCH-PURPOSE shadow sweep, ARCH-ORDER.)
- **The legacy-receipt reconcile path has no test** (`issuerecovery.go:135-142`). Deleting a pre-#308 stage-1 receipt and pointing at `issue publish` is new recovery behavior on state an older binary wrote. Only the `LegacyCreationDetail` predicate is tested (`createop_test.go:172`).
  - **Fix:** add the plan's `TestRecoveryLegacyCreationReceiptPointsAtPublish`. Build a stage-1 receipt whose candidate is the detail blob, run `runRecoveryReconcile`, and assert three things: the receipt is gone, stderr names `issue publish --issue N`, and main is unchanged.
  - (ARCH-SECURE: input persisted by an older version.)

### 4. Minor findings
- No adapter-level test covers an uncertain `creation.main` push. `lostAck` only drops the PublishCard acknowledgement. The plan's `TestCreationMainUncertainIsRecoverable` was not written. The shared `probeEvent` path is covered by transfer tests.
- `issueDirGuard` runs after `repositoryTracked`, not first as the plan says. That's harmless, since it still runs before the tracked/legacy split.
- With `requireExisting`, a brand-new legacy repository with no `workshop/issues` yet is now refused with "not the repository you mean". That message misleads in that case. The scaffold seeds the directory (`construct/base.manifest:173`), so the impact is small.
- The `move-detail` alias's RunE repeats publish's argument and legacy checks but skips `guardSpineRepo`. Calling the publish command's RunE would remove the drift (ARCH-DRY).
- `issueDirGuard` mixes filesystem probing with the decision. A pure `(root, cwd, dir, exists func)` core would be easy to unit-test (ARCH-PURE nit).
- The details on main carry the pre-claim `card_mirror`. That is the same as today's claim-from-main behaviour, and `start-plan` refreshes it.

### 5. Test coverage notes
- **Covered:** filing from `:0` leaves the issue unclaimed with the tree untouched, and filing from `pair:1` claims it with a claim-kind token. On an issue branch nothing is committed. The C5 refusals cover a nested issues tree and a legacy repository with no issues dir. A main race recomputes on the new tip, a foreign destination refuses with `ErrDestinationExists`, an open skeleton lands, and a started issue still owes Plan and Done-when content.
- **F1 regression:** the plan asked for a skeleton on main plus an unrelated branch that lands. The existing #322 test covers main-side changes after the branch point. The new row covers a branch that carries an open skeleton. Together they're adequate.
- **Gaps:** the legacy reconcile path (Important) and an uncertain `creation.main` push (Minor).

### 6. Architectural notes for upcoming work
- **ARCH-DRY: pass.**
- **ARCH-PURE: pass**, with a nit on `issueDirGuard`.
- **ARCH-PURPOSE: flag.** `NeedsSourceCheckout` is a leftover restatement of the old model. Everything else in M1 is delivered.
- **ARCH-MOCK: pass.** Tests use real git with a bare origin. The identity seam (`claimantIdentity`) is a function variable.
- **ARCH-CONSTRAINTS: pass.** There are up to three bounded round trips, and a failed claim is a warning.
- **ARCH-SECURE: pass**, with the legacy-receipt test gap noted above.
- **ARCH-ORDER: flag.** The source-checkout classification is stale, and no test covers an uncertain main push.
- **ARCH-FUNERAL: pass.** Nothing new is durable, and legacy receipts are deleted.
- **For M2:** `publishPeerProject` should reuse the same CAS and refresh pattern and keep its transform pure (`projectCloseEdit`).

### 7. Plan revision recommendations
- `## Revisions`: "Task 2's `TestRecoveryLegacyCreationReceiptPointsAtPublish` is split into a predicate test (`TestLegacyCreationDetailReceipt`) plus a reconcile test (to add)." Also record that `NeedsSourceCheckout(creation)` returns `false`, or state why it stays `Stage <= 1`.
- `## Revisions`: "Task 3's `TestIssueNewRefusesWhenCwdIsNotTheRepoRoot` is replaced by `TestIssueNewRefusesNestedIssuesTreeUnderAnotherRepo` and `TestLegacyIssueNewRefusesRepoWithoutIssuesDir`, because the cwd==toplevel rule contradicted subdirectory filing." The issue's Log already says this; the plan doesn't.

```findings
findings:
  - id: new
    severity: Important
    family: retired-stage-restated-in-model
    title: |
      NeedsSourceCheckout still treats creation stage 1 as a local effect; plan says false
    detail: |
      store.go:129-132 returns Stage<=1 for creation, and the comment still says "materializing created details". creation.main needs no checkout, so reconcile from another branch or worktree refuses with ErrForeignCheckout (issuerecovery.go:131) and recovery list says "finish from X". Return false for creation (check legacy receipts before the guard) and add a cross-branch reconcile test.
  - id: new
    severity: Important
    family: untested-recovery-path
    title: |
      Legacy creation.detail receipt release in reconcile has no test
    detail: |
      issuerecovery.go:135-142 deletes a pre-#308 stage-1 receipt and points at issue publish. Only the LegacyCreationDetail predicate is tested. Add the plan's TestRecoveryLegacyCreationReceiptPointsAtPublish: receipt removed, stderr names publish, main unchanged.
  - id: new
    severity: Minor
    family: untested-recovery-path
    title: |
      No adapter-level test of an uncertain creation.main push (plan's TestCreationMainUncertainIsRecoverable)
  - id: new
    severity: Minor
    family: guard-message-precision
    title: |
      issueDirGuard refuses a fresh legacy repo without workshop/issues with a misleading "not the repository you mean"
  - id: new
    severity: Minor
    family: alias-duplicates-target-guards
    title: |
      move-detail alias re-implements publish's argument and legacy checks and skips guardSpineRepo
```

---

## Re-review — 2026-10-12T09:40:16-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 308 — issue new publishes card and details in one go |
| repo | ariadne |
| issue file | workshop/issues/000308-issue-new-publishes-card-and-details-in-one-go.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | ab362ac4c35f137125f25e36a45f7b994d7493ad..8d6c22b9a9d0ac0d9aec23283c1c22891a5c1520 |
| command | sdlc milestone-close --issue 308 --milestone M1 |
| reviewer | claude |
| isolation | checkout |
| timestamp | 2026-10-12T09:40:16-07:00 |
| verdict | SHIP |

## Review

I'm checking the BR-1 and BR-2 fixes now. Next I'll run their regression tests, then revert the BR-1 fix to confirm the test goes red.

Both claimed fixes go red when reverted. Now I'm checking leftover creation-specific code in recovery and running the broader package tests.

```verdict
verdict: SHIP
confidence: high
```

The M1 review is clean: all five round-1 findings are fixed, and I raised one new finding, a Minor.

- **BR-1 is fixed.** `NeedsSourceCheckout` no longer treats creation as needing its source checkout (`store.go:131`). I put the old `case "creation": return r.wire.Stage <= 1` back into a throwaway copy of the code, and `TestIssueNewStoppedAtMainFinishesFromAnotherCheckout` failed with `ErrForeignCheckout` ("must be finished from the checkout on refs/heads/main"). With the fix restored it passes.
- **BR-2 is fixed.** I turned off the legacy-receipt release at `issuerecovery.go:135`, and `TestRecoveryLegacyCreationReceiptPointsAtPublish` failed ("object … is a blob, not a commit"). It passes with the release in place.
- **Tests:** `go test ./cmd/sdlc/internal/...` passes, and so does a targeted `./cmd/sdlc` run covering IssueNew, Recovery, Creation, Sections, LandingInstance, MoveDetail, Publish and Tracker (≈475s). `go vet` is clean.

**1. Strengths**
- The recovery tests use a real bare origin with a `pre-receive` hook that refuses pushes to main (`issuenew_recovery_test.go:17-30`). That is a stateful stand-in for an origin outage, not a mock, and it drives the real CLI recovery path.
- `TestRecoveryLegacyCreationReceiptPointsAtPublish` builds the old receipt shape by rewriting a real receipt into it and asserting the predicate. So the fixture can't drift away from what `LegacyCreationDetail` actually detects.
- `NeedsSourceCheckout` is now the only place that decides where a receipt must be finished. Both `recovery list` (`issuerecovery.go:71`) and `reconcile` (`:131`) use it, so the BR-1 fix reaches both.
- The `move-detail` alias now runs the same `guardSpineRepo` → `runTrackedPublish` path as `publish`.

**2. Critical findings:** none.

**3. Important findings:** none.

**4. Minor findings**
- The new-checkout guard at `issuerecovery.go:131` still has the clause `!(r.Discardable() && r.Operation() == "creation")`, but `NeedsSourceCheckout` can no longer be true for creation, so the clause never applies. This is the 2nd finding in family `retired-stage-restated-in-model`. The rule that covers both: once a stage is retired, `Receipt.NeedsSourceCheckout` alone says where a receipt must be finished, and callers don't add per-operation exceptions. Fix by deleting the clause and grepping `issuerecovery.go` for any other `Operation() == "creation"` exception to the checkout rule.

**5. Test coverage notes**
- On BR-3: the hook refuses the push while main's tip is unchanged. In `publishEvent` that path goes PushRejected → probe (absent, same base) → re-push → probe again. That is the "rejected, cause unknown" path an uncertain push takes, and the receipt then resumes from another checkout. This covers the plan's `TestCreationMainUncertainIsRecoverable`, just at the CLI level rather than the adapter level.
- Not run: the full `make test`, the M2 paths (out of this window), and the live scratch-issue check. The issue's Log says the live check was replaced by `TestTrackerFullSlotCycle`.

**6. Architectural notes for upcoming work**
- **ARCH-DRY: pass.** `DetailPublication` is shared by creation and transfer. The alias's own legacy refusal is deliberate, because its wording differs from `publish`'s legacy `--commit` path.
- **ARCH-PURE: pass.** `creatorClaims` and `CheckSectionsPresence` are pure; the effects stay in the adapters.
- **ARCH-PURPOSE: pass for M1's scope.** The peer tick (C1) is M2, as the plan says.
- **ARCH-MOCK: pass.** Tests use a real git origin, with hooks to simulate failures.
- **ARCH-CONSTRAINTS: pass.** Retries stay bounded by the existing receipt engine.
- **ARCH-SECURE: pass.** Untrusted older receipts are classified by a structural predicate and released, not driven.
- **ARCH-ORDER: pass.** Uncertain outcomes are probed before being called failures, and legacy-stage receipts are handled explicitly.
- **ARCH-FUNERAL: pass.** Legacy receipts are deleted when released, and `issue new` creates no new durable artifacts.
- For M2: `publishPeerProject` should reuse the hook-based origin fixture from `issuenew_recovery_test.go` for its "origin refuses" and "TL appeared" cases.

**7. Plan revision recommendations:** none. The Revisions entry for round 1 already records the Task 2 and Task 3 test changes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      store.go:131 drops creation; TestIssueNewStoppedAtMainFinishesFromAnotherCheckout fails with ErrForeignCheckout when the old case is restored (verified).
  - id: BR-2
    disposition: addressed
    note: |
      TestRecoveryLegacyCreationReceiptPointsAtPublish goes red (blob-not-a-commit) when the issuerecovery.go:135 release is disabled (verified).
  - id: BR-3
    disposition: addressed
    note: |
      The pre-receive refusal with an unchanged base goes through publishEvent's probe/re-push path, and the recovery tests resume the creation from it.
  - id: BR-4
    disposition: addressed
    note: |
      issue.go:398 message now names the fresh-repo action; TestLegacyIssueNewRefusesRepoWithoutIssuesDir pins it.
  - id: BR-5
    disposition: addressed
    note: |
      Alias calls guardSpineRepo then runTrackedPublish; its distinct legacy refusal is deliberate (publish's legacy path is --commit).
findings:
  - id: new
    severity: Minor
    family: retired-stage-restated-in-model
    title: |
      issuerecovery.go:131 keeps a creation exemption in the foreign-checkout guard that can no longer fire
    detail: |
      2nd in family. Rule: NeedsSourceCheckout alone says where a receipt must be finished; no caller adds per-operation exceptions. Delete the clause and grep issuerecovery.go for other creation-specific checkout exceptions.
```
