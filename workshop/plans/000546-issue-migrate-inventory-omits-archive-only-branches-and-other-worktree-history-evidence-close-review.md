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
