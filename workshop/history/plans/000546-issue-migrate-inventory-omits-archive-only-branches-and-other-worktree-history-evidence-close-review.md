# Boundary Review — ariadne#546 (whole-issue close)

| field | value |
|-------|-------|
| issue | 546 — issue migrate inventory omits archive-only branches and other-worktree history evidence |
| repo | ariadne |
| issue file | workshop/issues/000546-issue-migrate-inventory-omits-archive-only-branches-and-other-worktree-history-evidence.md |
| boundary | whole-issue close |
| milestone | — |
| window | 23ffbc7a2bdd7cf913e331cdede798267b275344..345fdc36883e03cca3b98d68586010ce77926e03 |
| command | sdlc close --issue 546 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T03:41:28-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The implementation delivers the documented inventory scope while preserving main-only imports and existing migration recovery. Targeted native Git, pure policy, lint, and recovery tests passed. No Critical or Important findings; one Minor cleanup remains.

```findings
findings:
  - id: new
    severity: Minor
    kind: defect
    family: obsolete-production-paths
    title: |
      Remove unused inventory compatibility helpers
    detail: |
      cmd/sdlc/issuemigrate_inventory.go:175 introduces dirtyIssuePaths with no callers, while cmd/sdlc/issuemigrate.go:207 retains migrationBranchRefs with only test callers. ARCH-DRY: remove these obsolete production paths and move the relevant tests onto publicationBranchTips or observeIssueInventory so they exercise the current observation contract.
```

1. **Strengths**

   - Shared classification derives roots from `issue.IDDirs`, preserving filename bytes and excluding nested bundles.
   - Archive evidence remains separate from active reconciliation; pure matching distinguishes related and unrelated histories.
   - Native regressions cover both original reproductions, detached/inherited evidence, ignored files, index flags, normalization, refusal preservation, and changed evidence before apply.
   - README and atlas describe the new scope, conservative refusals, and remaining writer-freeze requirement.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** Remove the unused compatibility helpers identified above.

5. **Test coverage notes**

   Passed:

   - Native migration/init/inventory suite: **256.929s**.
   - Targeted gitx, issue, tracker, and existing `PlanTrackerMigration` tests.
   - Go lint and recovery tests.
   - Pinned-range `git diff --check`.

   The native suite includes growth guards and existing close-proof, rename/archive, retry, and reconciliation regressions. The full repository suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — flag, Minor:** obsolete production helpers retain an unused alternative path.
   - **ARCH-PURE — pass:** listed pure entities perform no IO; semantic policy remains in the planner.
   - **ARCH-PURPOSE — pass:** observation covers the documented roots and carriers.
   - **ARCH-MOCK — pass:** portable native repositories/bare origins follow this repo’s established stateful Git fixture approach; publication fake coverage remains intact.
   - **ARCH-CONSTRAINTS — pass:** bounded reads, serial processing, OID deduplication, and growth tests support the declared envelope.
   - **ARCH-SECURE — pass:** malformed observations and unsupported kinds fail closed; paths retain NUL framing.
   - **ARCH-ORDER — pass:** apply recollects evidence before digest validation; publication recovery remains unchanged.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifact family.
   - **ARCH-EVOLVE — pass:** no persisted-format change.

7. **Plan revision recommendations:** None required for the delivered behavior.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"37118dd5-60a9-497d-ad6c-fb51182120b9","repository":"github.com/xianxu/ariadne","issue_id":"000546","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/18c9939a0aa4ce00","reviewed_head":"345fdc36883e03cca3b98d68586010ce77926e03","slot":"ariadne:3","requested_at":"2026-10-22T10:36:09.986285Z"},"phase":"finished","pid":66917,"started_at":"2026-10-22T10:36:10.134036Z","finished_at":"2026-10-22T10:41:27.54641Z","session_id":"01a1632e-bfa9-7710-bfe5-46fec3f6247c","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T04:47:10-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 546 — issue migrate inventory omits archive-only branches and other-worktree history evidence |
| repo | ariadne |
| issue file | workshop/issues/000546-issue-migrate-inventory-omits-archive-only-branches-and-other-worktree-history-evidence.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9c846f17395634d47f4dd2f9084f1a2d4f990a87..8797b4a4a7e57b60ff83e29f56a245c40926d4d3 |
| command | sdlc close --issue 546 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T04:47:10-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff addresses BR-1 and introduces no identified defects. The fixture correction preserves the original asset-refusal assertions. Spec and Done-when remain unchanged; the plan revision records the bounded correction.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Removed migrationBranchRefs and the unused dirtyIssuePaths wrapper. Tests now exercise publicationBranchTips and observeIssueInventory; the retained dirtyIssuePaths has a live migrationInventory caller consuming pinned worktrees. Repository search confirms no obsolete helper references remain.
```

1. **Strengths**
   - Removed both obsolete production paths without changing runtime policy.
   - Ref-category tests now check the shared inventory’s exact names and OIDs.
   - Asset fixture repair retains conversion/content-name refusal and checkout-preservation assertions.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage:** Affected asset, migration-inventory and ref-category tests passed (`126.465s`). Pinned-range `git diff --check` passed. Cleanup changes no executable behavior requiring a new regression test. The full recorded suite was not rerun in this review.

6. **Architecture:** ARCH-DRY passes: obsolete parallel paths removed. ARCH-PURE passes: IO boundaries unchanged. ARCH-PURPOSE passes: both BR-1 instances addressed and tests moved to live consumers. ARCH-EVOLVE passes: no persisted-format changes. No new user-facing or architectural surface requires README or atlas updates.

7. **Plan revisions:** None needed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"0b69eff7-ac13-46da-8b80-313b4900d2ef","repository":"github.com/xianxu/ariadne","issue_id":"000546","recipe":"fix-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/07a423a39a92b53a","reviewed_head":"8797b4a4a7e57b60ff83e29f56a245c40926d4d3","slot":"ariadne:3","requested_at":"2026-10-22T11:44:20.006488Z"},"phase":"finished","pid":93383,"started_at":"2026-10-22T11:44:20.163984Z","finished_at":"2026-10-22T11:47:09.705346Z","session_id":"01a1636d-2882-7d41-9443-1f523e71cf6b","outcome":"returned"}] -->

---

## Re-review — 2026-10-22T08:25:05-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 546 — issue migrate inventory omits archive-only branches and other-worktree history evidence |
| repo | ariadne |
| issue file | workshop/issues/000546-issue-migrate-inventory-omits-archive-only-branches-and-other-worktree-history-evidence.md |
| boundary | whole-issue close |
| milestone | — |
| window | a437d609d13e8b2efbb69bb113f1e60555bda5b5..eaad89ade92b1faa464ae407615a6e803cdcda06 |
| command | sdlc close --issue 546 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-22T08:25:05-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff preserves both sides of the conflict: complete inventory observation and main’s legacy-completion verification. BR-1 remains addressed. No new defects or scope changes were identified; affected native tests passed.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Repository search confirms the obsolete migrationBranchRefs helper and dirtyIssuePaths wrapper remain absent. Tests exercise publicationBranchTips and observeIssueInventory; the retained dirtyIssuePaths consumes pinned worktrees through migrationInventory.
```

1. **Strengths**

   - `cmd/sdlc/issuemigrate_inventory.go:110–140` preserves main’s completion-proof collection byte-for-byte and connects it to the complete inventory.
   - The collector propagates proof errors and uses the planner’s selected source generation.
   - `atlas/workflow/issue-tracker-migration.md:34,79` retains both inventory guidance and legacy-completion recovery. Both legacy sections are preserved verbatim.
   - Spec and Done-when remain unchanged; the plan changes only an acceptance checkbox.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Nine targeted tests passed, including their subtests, in **39.944s**. Coverage includes archived completion preservation, provable/landed closes, committed inventory scopes, reserved branch names, both archive layouts, immutable reads, fetch isolation, and changed-evidence rejection. Pinned-range `git diff --check` passed. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** one inventory implementation remains; proof collection reuses the existing planner and verifier.
   - **ARCH-PURE — pass:** semantic selection remains pure; Git evidence collection stays in the integration shell.
   - **ARCH-PURPOSE — pass:** conflict resolution preserves complete observation and existing completion authority.
   - **ARCH-EVOLVE — pass:** no persisted-format changes.
   
   Atlas updates cover the resolution. This interdiff introduces no new user-facing command or configuration requiring a README update.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"99ffc415-0233-4de9-b742-10bae4a6f47d","repository":"github.com/xianxu/ariadne","issue_id":"000546","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/fffe4d866e38acdd","reviewed_head":"eaad89ade92b1faa464ae407615a6e803cdcda06","slot":"ariadne:3","requested_at":"2026-10-22T15:23:14.66332Z"},"phase":"finished","pid":22039,"started_at":"2026-10-22T15:23:14.819637Z","finished_at":"2026-10-22T15:25:04.658755Z","session_id":"01a16435-93ef-79d3-9f3b-99db69d63674","outcome":"returned"}] -->
