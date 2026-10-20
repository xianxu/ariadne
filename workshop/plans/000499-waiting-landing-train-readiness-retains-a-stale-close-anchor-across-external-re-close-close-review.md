# Boundary Review — ariadne#499 (whole-issue close)

| field | value |
|-------|-------|
| issue | 499 — Waiting landing-train readiness retains a stale close anchor across external re-close |
| repo | ariadne |
| issue file | workshop/issues/000499-waiting-landing-train-readiness-retains-a-stale-close-anchor-across-external-re-close.md |
| boundary | whole-issue close |
| milestone | — |
| window | e725110a66b5585debff6181e2bf0b0fc7f27d26..8d74ea7a3132393ba070737dc0a1ac87b0db6439 |
| command | sdlc close --issue 499 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T09:10:00-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The change fixes stale records across readiness passes, and both new regressions fail without it. However, production readiness still reads multiple tracker generations within one pass, violating the Spec’s coherence requirement. The new counting test replaces the production gates and misses this path.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: pass-scoped-authority
    title: |
      Production readiness does not share one tracker generation across all checks
    detail: |
      cmd/sdlc/landingcar.go:43 scopes loadIssueRecords, but runLandingPublishGate first calls guardTransferredDetails (landinggate.go:84), whose changedDetails reads env.repo.Snapshot directly (transferguard.go:82). Issue selection subsequently reads through the scoped cache (landingarchive.go:770). An external write between these checks produces different generations within one readiness pass; a disposable-head probe retaining both production checks confirmed distinct tracker SHAs. Route both consumers through one pass-local snapshot and add a production-path regression. The complete affected family comprises this direct-snapshot/scoped-record split, the coherence claim at landingcar.go:40–43, its atlas restatement at atlas/workflow/sdlc-binary.md:455, and the fake-gate coverage at landingcar_records_test.go:88–142, which cannot detect that split. ARCH-PURPOSE: the observation boundary must cover every tracker consumer in the pass.
```

1. **Strengths**

   - Reuses `withIssueRecordsScope`, preserving parent context and isolating the parent cache.
   - The re-close regression exercises real close and publish logic against a local bare remote.
   - The status regression verifies external visibility and parent-scope isolation at both 8 and 16 reads.

2. **Critical findings**

   - The finding above blocks shipping. Share the tracker snapshot between transferred-details checks and completion selection; retain a regression that changes the remote between those consumers.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Focused readiness tests passed; scope, context, and waiting-loop checks also passed.
   - Removing the fix in a disposable checkout made both new regressions fail.
   - A separate disposable production-path probe exposed mixed generations after an intervening external title update.
   - Disposable checkouts were removed. The reviewed checkout remains unchanged.
   - The full suite was not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing scope helper reused.
   - **ARCH-PURE — pass:** cache lifetime remains in the integration boundary.
   - **ARCH-PURPOSE — flag:** production consumers do not all derive from the pass-local generation.
   - **ARCH-EVOLVE — pass:** no persisted format changes.

   Atlas was updated, though its coherence claim needs the correction above. No new command, flag, or configuration surface requires a README update.

7. **Plan revision recommendations**

   Add a `## Revisions` entry identifying the direct snapshot consumer and extending the implementation and regression plan to cover coherent generations through the production publish gate. The documented separation of reopen policy from cache visibility can remain.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"d61398c3-b61a-4fb1-bcb4-ff7a94a2538e","repository":"github.com/xianxu/ariadne","issue_id":"000499","recipe":"small-diff-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5981a55a8c1b48a1","reviewed_head":"8d74ea7a3132393ba070737dc0a1ac87b0db6439","slot":"ariadne:5","requested_at":"2026-10-20T16:06:10.711306Z"},"phase":"finished","pid":63085,"started_at":"2026-10-20T16:06:10.803729Z","finished_at":"2026-10-20T16:09:59.41483Z","session_id":"01a15a10-2969-7022-8fd8-5bd98abcb468","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T09:37:49-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 499 — Waiting landing-train readiness retains a stale close anchor across external re-close |
| repo | ariadne |
| issue file | workshop/issues/000499-waiting-landing-train-readiness-retains-a-stale-close-anchor-across-external-re-close.md |
| boundary | whole-issue close |
| milestone | — |
| window | e725110a66b5585debff6181e2bf0b0fc7f27d26..2af32ce2021a9388d45e21bde7f3605ed048d590 |
| command | sdlc close --issue 499 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T09:37:49-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned changes satisfy the revised Spec and Plan. Readiness refreshes tracker authority between passes while sharing one generation within each pass. BR-1 is addressed, with regression evidence that fails without the correction. No new findings.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      transferguard.go:35,82 now uses scoped records; issuerecord.go:68 shares repository authority across detail directories. TestCarReadinessPublishChecksShareGeneration passes at HEAD and fails when the former direct-snapshot reader is restored in a disposable checkout.
```

1. **Strengths**

   - `landingcar.go:43` reuses the existing scope helper and isolates each readiness pass.
   - `landingcar_records_test.go:154` retains production publish checks and changes external authority between them.
   - `issuerecord.go:128` invalidates older directory views when stale authority upgrades to fresh.
   - Snapshot recomposition preserves quarantine paths, freshness metadata, and previous views.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   - Tracker record tests passed.
   - Readiness, records-scope, restore, and context checks passed.
   - Transferred-details ownership and publish compatibility tests passed.
   - Removing the per-pass scope makes the re-close and both status-count regressions fail.
   - Restoring the old direct-snapshot reader makes the BR-1 regression fail.
   - Disposable checkouts were removed; the reviewed checkout remains unchanged.
   - Full suite was not rerun; its recorded pass remains required before landing.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing scope and composition helpers reused.
   - **ARCH-PURE — pass:** IO remains in integration readers; composition remains pure.
   - **ARCH-PURPOSE — pass:** ownership checks, completion selection, and flake-registry readers use scoped authority.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.
   - Atlas documents the new flow. No new user-facing command, flag, or configuration requires a README update.

7. **Plan revision recommendations:** None. Existing revisions cover the production-authority correction and the separate reopen-policy issue.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7908d70b-e7b9-4487-a8a0-4ae3ae82ef2a","repository":"github.com/xianxu/ariadne","issue_id":"000499","recipe":"small-diff-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/2d5da2fbaecd8d0c","reviewed_head":"2af32ce2021a9388d45e21bde7f3605ed048d590","slot":"ariadne:5","requested_at":"2026-10-20T16:34:31.787767Z"},"phase":"finished","pid":12714,"started_at":"2026-10-20T16:34:31.948835Z","finished_at":"2026-10-20T16:37:48.712515Z","session_id":"01a15a2a-1f2d-77c0-b058-08d326a43b8b","outcome":"returned"}] -->
