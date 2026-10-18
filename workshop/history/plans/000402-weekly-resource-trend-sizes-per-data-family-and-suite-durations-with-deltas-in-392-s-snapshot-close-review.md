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

---

## Re-review — 2026-10-18T04:03:28-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 402 — Weekly resource trend: sizes per data family and suite durations, with deltas, in #392's snapshot |
| repo | ariadne |
| issue file | workshop/issues/000402-weekly-resource-trend-sizes-per-data-family-and-suite-durations-with-deltas-in-392-s-snapshot.md |
| boundary | whole-issue close |
| milestone | — |
| window | c1e63fe96e857efb9ef65c7135361fc2d7db21db..9ffb0ecdcc9faa408f39ee0edb7819de0b8594f6 |
| command | sdlc close --issue 402 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T04:03:28-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All five prior findings are addressed, with regression evidence for the three behavior fixes. The pinned head’s focused tests pass, and the live pilot matches its recorded hash. One new correctness defect blocks shipping: device trends cannot recover after a reboot or missing sample because comparison failure overwrites observation validity.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      weeklyresource_collect.go marks later blocked-file claimants incomplete. The committed regression passes at head and fails with the pre-fix collector restored; existing fixtures also cover earlier assignments and multiple claimants.
  - id: BR-2
    disposition: addressed
    note: |
      weeklystats_cmd.go returns validated existing publications before collection and recovers a concurrent winner. Clock-advance and concurrent-publication tests pass at head and both fail with the pre-fix command restored.
  - id: BR-3
    disposition: addressed
    note: |
      Both persisted readers call validatePersistedWeeklyResources. Malformed-measurement fixtures pass at head and fail with the pre-fix previous-week reader restored, including negative values, duplicate identities and invalid quantiles.
  - id: BR-4
    disposition: addressed
    note: |
      The pinned plan diff replaces the five nonexistent names with actual functions. Every Core concepts row now resolves to its declared file; the PURE functions have direct tests without IO. The first-boundary-review Revisions entry records the reconciliation.
  - id: BR-5
    disposition: addressed
    note: |
      Plan line 40 explicitly states permanent retention, no cleanup job, estimated incremental and total annual growth, and the measurement directory. weeklystats_write.go retains dated publications; the inspected pilot is 41,887 bytes, consistent with the documented total.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: observation-comparison-state-separation
    title: |
      Device trends remain unknown permanently after a recoverable comparison failure
    detail: |
      cmd/sdlc/weeklyresource_device.go:70–78 changes a valid current observation to unknown after an epoch change; the next week rejects that stored observation because its state is unknown, and repeats indefinitely. weeklyresource_machine.go:113–124 has the same failure for wear after a missing sample, while lines 100–107 propagate it after invalid prior evidence or an observation-time gap. Two disposable regression probes fail at head: boot2 values 10 then 30 never yield delta 20, and readable wear values 8 then 9 after a missing sample never yield delta 1. Preserve measurement validity separately from comparison status across both gauge types, and test recovery sequences through persisted weekly sections (ARCH-ORDER, ARCH-PURPOSE).
```

1. **Strengths**

   - Reuses #392’s producer, suite statistics and immutable writer.
   - Persisted evidence receives shared structural and semantic validation before comparison.
   - Metadata collection avoids file-content reads, skips symlinks and bounds traversal.
   - README and atlas describe schema v2, observation timing and manual publication.

2. **Critical findings**

   The device-state defect above requires separating “current measurement valid” from “delta unavailable.” Preserve the valid sample so the following comparable week can recover.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused resource, snapshot and device tests pass. BR-1–BR-3 tests demonstrably fail without their fixes. Both new recovery probes fail against the pinned head. The pilot’s SHA-256 matches `df3f21f55d3392a95c2b6b7c377efb78e2797a0141c9596758466fdc6ead21bc`; it contains 17 families, ten ledgers and 16 suite rows. Full suite not rerun during review. Disposable changes were removed; the checkout is clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared inventory, validator and publication authority.
   - **ARCH-PURE — pass:** comparison logic is directly testable without IO.
   - **ARCH-PURPOSE — flag:** device trends stop recovering after ordinary disruptions.
   - **ARCH-MOCK — pass:** command and capacity seams support controlled outputs and failures.
   - **ARCH-CONSTRAINTS — pass:** traversal, persisted reads and subprocess execution have explicit bounds.
   - **ARCH-SECURE — pass:** malformed persisted measurements become named unknown coverage.
   - **ARCH-ORDER — flag:** device state conflates observation validity with comparison outcome.
   - **ARCH-FUNERAL — pass:** permanent retention and its growth cost are explicit.

7. **Plan revision recommendations**

   Add a `## Revisions` entry defining device observation validity separately from comparison status, including recovery after reboot, missing samples, invalid prior evidence and timing gaps. Require multiweek sequence tests for both write and wear gauges.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"169a0263-61e7-4f0b-8fa6-adef344430a5","repository":"github.com/xianxu/ariadne","issue_id":"000402","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/4cbcc6c7a059796a","reviewed_head":"9ffb0ecdcc9faa408f39ee0edb7819de0b8594f6","slot":"ariadne:1","requested_at":"2026-10-18T11:00:31.757132Z"},"phase":"finished","pid":32612,"started_at":"2026-10-18T11:00:31.845585Z","finished_at":"2026-10-18T11:03:28.370063Z","session_id":"01a14eab-9ccb-7710-98e8-4767140526de","outcome":"returned"}] -->

---

## Re-review — 2026-10-18T04:33:55-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 402 — Weekly resource trend: sizes per data family and suite durations, with deltas, in #392's snapshot |
| repo | ariadne |
| issue file | workshop/issues/000402-weekly-resource-trend-sizes-per-data-family-and-suite-durations-with-deltas-in-392-s-snapshot.md |
| boundary | whole-issue close |
| milestone | — |
| window | c1e63fe96e857efb9ef65c7135361fc2d7db21db..a8f715692083efa7fcedc984606430e858a65677 |
| command | sdlc close --issue 402 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T04:33:55-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation fulfills the resource-trend scope, and BR-6 is addressed with meaningful regression evidence. Focused tests pass; restoring the old comparison guards makes all four recovery scenarios fail. No new blocking findings were identified.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Blocked files invalidate later claimants; the three-claimant fixture passes in both root orders.
  - id: BR-2
    disposition: addressed
    note: |
      Exact-week validation precedes collection; advancing-clock retry and concurrent-publication fixtures pass.
  - id: BR-3
    disposition: addressed
    note: |
      Previous-week and existing-week readers share persisted-resource validation; missing-field and semantic-corruption fixtures pass.
  - id: BR-4
    disposition: addressed
    note: |
      The plan's Core concepts table at lines 27–33 names existing entities at their declared paths; PURE functions contain no IO.
  - id: BR-5
    disposition: addressed
    note: |
      Plan line 41 explicitly specifies permanent brain retention and annual growth costs; the inspected 41,939-byte pilot supports the stated total-size estimate.
  - id: BR-6
    disposition: addressed
    note: |
      Device comparisons consume prior.Observed independently of comparison state. JSON-round-trip recovery tests pass at head and fail when the former prior-state guards are restored, reproducing lost write and wear deltas.
```

1. **Strengths**

   - Device validity survives comparison failure in `weeklyresource_device.go:62` and `weeklyresource_machine.go:116`.
   - Shared validation protects both persisted comparison evidence and immutable retries.
   - Collection fixtures cover overlap, symlink avoidance, ledger ranking, and traversal limits.
   - README and atlas document schema v2, observation timing, and manual publication.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Focused resource, device, previous-week, and snapshot-command tests passed. The BR-6 mutation check failed as expected; restoring head returned tests to green. `git diff --check` passed.

   Independently inspected the recorded temp-brain pilot: its SHA-256 matches the issue, with 17 complete families, ten ledgers, 16 suite rows, and observed device baselines. The full suite was not rerun during this review. The disposable checkout was removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** Reuses the inventory, suite metrics, writer, and validator.
   - **ARCH-PURE — pass:** Comparison logic is deterministic and tested without IO.
   - **ARCH-PURPOSE — pass:** Delivers the declared resource and suite trends, including recovery.
   - **ARCH-MOCK — pass:** Machine probes use an injected runner; fixtures and the live pilot exercise the boundary.
   - **ARCH-CONSTRAINTS — pass:** Traversals, subprocess execution, output, and prior-file reads have explicit limits.
   - **ARCH-SECURE — pass:** Persisted evidence is validated; subprocess arguments remain structured.
   - **ARCH-ORDER — pass:** Observation validity and comparison outcome are separate; recovery sequences are tested.
   - **ARCH-FUNERAL — pass:** Permanent snapshot retention and its measured growth cost are explicit.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"696e6e5e-722d-4487-acc5-715172e0a7a8","repository":"github.com/xianxu/ariadne","issue_id":"000402","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/bf9d2c394e51393e","reviewed_head":"a8f715692083efa7fcedc984606430e858a65677","slot":"ariadne:1","requested_at":"2026-10-18T11:32:08.364288Z"},"phase":"finished","pid":85468,"started_at":"2026-10-18T11:32:08.456669Z","finished_at":"2026-10-18T11:33:55.055272Z","session_id":"01a14ec8-8d76-7e43-adb5-294ac7e4cb04","outcome":"returned"}] -->
