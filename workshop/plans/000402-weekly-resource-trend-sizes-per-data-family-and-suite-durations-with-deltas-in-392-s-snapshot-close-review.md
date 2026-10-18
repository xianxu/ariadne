# Boundary Review — ariadne#402 (whole-issue close)

| field | value |
|-------|-------|
| issue | 402 — Weekly resource trend: sizes per data family and suite durations, with deltas, in #392's snapshot |
| repo | ariadne |
| issue file | workshop/issues/000402-weekly-resource-trend-sizes-per-data-family-and-suite-durations-with-deltas-in-392-s-snapshot.md |
| boundary | whole-issue close |
| milestone | — |
| window | c1e63fe96e857efb9ef65c7135361fc2d7db21db..d8e59a282ee3fe622c792b7e341c48d2fc2a0f4f |
| command | sdlc close --issue 402 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T03:11:04-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation delivers the resource section and preserves a useful separation between metadata collection and pure comparisons. Existing focused tests pass, and the recorded live snapshot matches its documented hash. Three reproduced behavior defects, an inaccurate Core concepts table, and an unspecified artifact lifecycle block shipment.

1. **Strengths**

   - Point-resource comparisons enforce the 6–8-day observation interval while suite comparisons retain their exact weekly windows.
   - Device-counter comparisons reject reboot epochs, device changes and counter resets.
   - The collector uses declared inventory selectors, skips symlinks and ranks ledgers during the metadata traversal.
   - README and atlas updates describe schema v2 and the manual publication schedule.

2. **Critical findings**

```findings
findings:
  - id: new
    severity: Critical
    family: overlap-invalidates-all-claimants
    title: |
      Later claimants of an already-blocked file remain falsely complete
    detail: |
      cmd/sdlc/weeklyresource_collect.go:161–162 returns immediately for a blocked file without marking its newly matched family incomplete. A temporary regression fixture with two overlapping families under the first root and a third family under a later root reports the third family as Complete:true, Bytes:0, Files:0 despite its matching file. Mark every matching claimant unknown regardless of traversal order, and test permutations and additional claimants. This violates the declared overlap contract and ARCH-PURPOSE.
  - id: new
    severity: Critical
    family: immutable-publication-retry
    title: |
      Advancing the observation clock breaks same-week publication retries
    detail: |
      cmd/sdlc/weeklystats_cmd.go:140–150 recollects resources and inserts a fresh observed_at before calling the byte-comparing immutable writer. A regression probe with unchanged source fixtures and a one-second clock advance fails with “already exists with different bytes.” Preserve and return an existing validated publication on retry, including concurrent-publication recovery, without rewriting its observation time. The committed rerun test freezes the clock and misses this ARCH-ORDER contract violation.
  - id: new
    severity: Critical
    family: persisted-measurement-validation
    title: |
      Invalid persisted measurements are accepted as complete comparison evidence
    detail: |
      cmd/sdlc/weeklyresource_previous.go:60–63 checks section presence but does not validate measurement semantics. A regression probe supplying a complete family with bytes:-100 is accepted as complete; compareResourceBytes then permits it to produce a numeric delta. Validate persisted rows before comparison, including nonnegative measurements, required values and unique identities, and return named unknown coverage for invalid evidence. Add malformed-but-valid-JSON fixtures across the measurement types (ARCH-SECURE).
  - id: new
    severity: Critical
    family: core-concepts-match-implementation
    title: |
      Five Core concepts entries do not exist at their declared locations
    detail: |
      workshop/plans/000402-weekly-resource-trend-plan.md:28–32 names ResourceDelta, RenderResourceSection, ResourceCollector, PreviousResourceReader and WeeklySnapshotAttachment, but repository searches find none of these code entities. The implementation instead uses helpers including compareResourceBytes, renderWeeklyResourceSection, collectResourceDetails, readPreviousResourceSection and runWeeklySnapshot. Append a Revisions entry mapping the actual entities, paths and PURE/INTEGRATION classifications, and reconcile the table and descriptions. The required Core concepts cross-check classifies this contradiction as Critical (ARCH-PURPOSE).
  - id: new
    severity: Important
    family: durable-artifact-lifecycle
    title: |
      Larger weekly artifacts have no explicit retention or permanent-growth contract
    detail: |
      workshop/plans/000402-weekly-resource-trend-plan.md:40 delegates retention to the brain charter without identifying a removal routine, bound, or explicit permanent-retention cost. The new resources payload adds approximately 13.4 KB of formatted JSON to the inspected weekly artifact, while cmd/sdlc/weeklystats_write.go retains dated publications indefinitely. Name the applicable cleanup policy or explicitly document permanent retention, expected growth and its measurement location (ARCH-FUNERAL).
```

3. **Important findings**

   The artifact lifecycle finding above requires a concrete retention policy or documented permanent-retention cost.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused resource and weekly-snapshot tests passed, as did the pinned-range whitespace check. Three temporary Go-overlay regression tests failed on the reviewed head, reproducing the overlap, retry and invalid-predecessor defects. The overlays were removed; tracked files were unchanged.

   Independently inspected the recorded 41,887-byte pilot: SHA-256 `b4d1c143405bb2f8d549b9114173dd23c0f1135b6b91b4eb19ffa53f930b1681`, schema v2, 17 complete families, ten ledgers and 16 suite rows. The full suite was not rerun during this review.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — Pass:** inventory definitions and the existing snapshot writer are reused.
   - **ARCH-PURE — Pass:** comparison logic is directly tested without IO; adapters own collection.
   - **ARCH-PURPOSE — Flag:** overlap handling remains order-dependent, and the concept table does not match code.
   - **ARCH-MOCK — Pass:** machine probes share an injected boundary with fixture tests; live pilot evidence is present.
   - **ARCH-CONSTRAINTS — Pass:** entry/output limits and command timeouts bound collection; traversal scaling has fixtures.
   - **ARCH-SECURE — Flag:** persisted numeric evidence lacks semantic validation.
   - **ARCH-ORDER — Flag:** publication retries fail when the clock advances.
   - **ARCH-FUNERAL — Flag:** the expanded durable payload lacks an explicit lifecycle contract.

7. **Plan revision recommendations**

   Append dated `## Revisions` entries identifying the actual code entities, defining retry behavior for an already-published week, and naming the weekly artifact retention policy or measured permanent-growth cost. Expand the regression matrix to cover overlap ordering and semantically invalid predecessor rows.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"8557d2d9-08f3-4b05-9078-85445ca09e1b","repository":"github.com/xianxu/ariadne","issue_id":"000402","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/a2497d12bfb46a88","reviewed_head":"d8e59a282ee3fe622c792b7e341c48d2fc2a0f4f","slot":"ariadne:1","requested_at":"2026-10-18T10:07:26.57511Z"},"phase":"finished","pid":81507,"started_at":"2026-10-18T10:07:26.682458Z","finished_at":"2026-10-18T10:11:04.331967Z","session_id":"01a14e7b-02cf-7940-a0a6-9a698db10237","outcome":"returned"}] -->
