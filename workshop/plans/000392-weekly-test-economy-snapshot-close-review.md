# Boundary Review — ariadne#392 (whole-issue close)

| field | value |
|-------|-------|
| issue | 392 — Weekly test-economy snapshot from local checks and landings |
| repo | ariadne |
| issue file | workshop/issues/000392-weekly-test-economy-snapshot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 826a55dd130c59a4ae255c147ffafb3e7a2f2c1b..fb020b30e17631086f207423ca8f7fe816f8393c |
| command | sdlc close --issue 392 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T23:39:16-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation has a sound separation between aggregation, evidence collection, and immutable publication. However, three reproduced defects break the measurement contract: expired history can produce a numeric ratio, incomplete retries can escape validation, and legacy attempts disappear from reconciliation. Documentation and operational acceptance evidence also remain incomplete.

**1. Strengths**

- Reuses existing quantile and suite aggregation logic while adding explicit unknown flake classifications.
- Verifies landing PR, branch, completion binding, and issue-side Git ancestry.
- Atomic no-replace publication correctly handles identical reruns, conflicting bytes, concurrent writers, and stale temporary files.
- Atlas documents the new snapshot surface.

**2. Critical findings / 3. Important findings**

```findings
findings:
  - id: new
    severity: Critical
    family: evidence-history-completeness
    title: |
      Expired source history can be certified as a complete numeric ratio.
    detail: |
      cmd/sdlc/weeklyattribution.go:326 publishes a ratio whenever no explicit warning exists, but the reader supplies no retention or collection-coverage boundary. A temporary regression test with a first enqueue 40 days before the week and no retained attempts produces ratio=0 with no unknown coverage, despite the plan at line 21 requiring null when earlier evidence might have expired. Existing pruning removes old attempt records at cmd/sdlc/teststats_store.go:473. Carry explicit history coverage into validation and require it to cover the entire cohort interval before certifying a ratio; test expired attempts and potentially expired earlier admissions (ARCH-SECURE, ARCH-PURPOSE).
  - id: new
    severity: Critical
    family: evidence-boundary-validation
    title: |
      Earlier enqueue validation skips intermediate retries after finding the earliest enqueue.
    detail: |
      cmd/sdlc/weeklyattribution.go:193 compares each prior enqueue against enqueued, which line 216 mutates during the scan. For chronological events containing a refused enqueue at -3h, an unrefused enqueue at -2h, and the successful enqueue at -1h, the first record moves the threshold to -3h and the incomplete intermediate ticket is skipped. A temporary regression test confirms that this contradictory lifecycle is accepted without unknown coverage. Enumerate all admissions against the original successful enqueue boundary, validate each independently, then calculate the earliest timestamp; test input permutations and multiple retries (ARCH-ORDER, ARCH-SECURE).
  - id: new
    severity: Critical
    family: metric-population-consistency
    title: |
      Legacy attempts counted in weekly totals vanish from reconciliation.
    detail: |
      cmd/sdlc/weeklyattribution.go:262 selects in-week rows using FinishedAt directly, while aggregateWeeklyRepo uses suiteAttemptCompletedAt at cmd/sdlc/weeklystats.go:101. Production legacy records populate StartedAt but leave FinishedAt zero, so they contribute to physical suite totals yet never enter unassignable_attempts or unknown_attempts. A temporary regression test produces physical runs=1 and unassignable=0. Reconcile the same selected population while preserving the legacy timestamp's uncertainty and excluding it from exact attribution (ARCH-PURPOSE).
  - id: new
    severity: Important
    family: readme-surface-documentation
    title: |
      README documentation is missing for the new snapshot command.
    detail: |
      README.md:498 documents only the existing test-stats view; README.md is unchanged in the pinned range despite adding snapshot, --end, and --brain-root. Add invocation examples, weekly-boundary behavior, the immutable output path, and coverage/conflict semantics. Embedded help and the atlas update do not satisfy the explicit README gate.
  - id: new
    severity: Important
    family: default-rollout-completeness
    title: |
      The durable pilot snapshot and rollout acceptance evidence are still outstanding.
    detail: |
      workshop/issues/000392-weekly-test-economy-snapshot.md:73 explicitly says the pilot wrote only to a temporary directory, and line 75 describes the saved snapshot as future work. No durable snapshot link, ops-project confirmation, or successful execution from both propagated pilot checkouts is recorded, although Done when lines 33–36 and plan line 68 require these. Complete the capture and record its exact path/commit, source coverage, stable-rerun evidence, project link confirmation, and pilot consumer verification before whole-issue close (ARCH-PURPOSE, ARCH-MOCK).
  - id: new
    severity: Important
    family: explicit-operating-envelope
    title: |
      Git collection expands beyond the declared weekly cohort.
    detail: |
      cmd/sdlc/weeklystats_inputs.go:191 inventories every reachable historical merge, and lines 246–282 request graphs for every completion with any retained landed event, irrespective of the requested week. This contradicts plan line 41's bound by weekly landed commits and candidate attempts; an irrelevant historical graph failure also nulls the requested week's ratio through global CoverageUnknown. Pass the window into collection, restrict graph work to relevant candidates, and test that unrelated historical landings cannot consume the budget or invalidate current coverage (ARCH-CONSTRAINTS).
```

**4. Minor findings**

None.

**5. Test coverage notes**

- Focused weekly tests passed, as did `TestVerbContextsReachTrackerReads`.
- `git diff --check` passed.
- Temporary Go test overlays reproduced all three Critical defects against production functions. The overlays were removed; the checkout remains unchanged.
- Existing tests omit expired history, three-admission sequences, and legacy reconciliation. Add permanent regressions for these cases.
- Full-suite verification was not rerun during this review.

**6. Architectural notes**

- **ARCH-DRY — pass:** existing statistics and evidence primitives are reused.
- **ARCH-PURE — pass:** aggregation and attribution remain free of IO; integration functions own collection and publication.
- **ARCH-PURPOSE — flag:** reconciliation defects and the outstanding durable pilot prevent delivery of the promised baseline.
- **ARCH-MOCK — flag:** portable Git fixtures exist, but the required live pilot/consumer conformance evidence remains outstanding.
- **ARCH-CONSTRAINTS — flag:** historical Git collection exceeds the declared cohort scope.
- **ARCH-SECURE — flag:** missing history and contradictory admissions can be treated as complete evidence.
- **ARCH-ORDER — flag:** lifecycle validation depends on mutable scan state and event ordering.
- **ARCH-FUNERAL — pass:** permanent snapshot ownership is documented, and temporary artifacts have automatic cleanup.

The Core concepts roles are present in the stated implementation areas; the PURE implementations and their unit tests do not require IO.

**7. Plan revision recommendations**

Append `## Revisions` entries naming:

- The explicit source-history coverage model used to certify cohort completeness.
- Order-independent validation of every prior admission.
- A shared population-selection rule for legacy reconciliation.
- Window-scoped evidence collection and its verification.
- Actual pilot capture, project linkage, and propagated-consumer verification evidence once completed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"49dd21a3-fb69-4088-afe2-636d048e9f91","repository":"github.com/xianxu/ariadne","issue_id":"000392","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/db824b5e234d4c2b","reviewed_head":"fb020b30e17631086f207423ca8f7fe816f8393c","slot":"ariadne:1","requested_at":"2026-10-18T06:35:26.311244Z"},"phase":"finished","pid":68460,"started_at":"2026-10-18T06:35:26.543529Z","finished_at":"2026-10-18T06:39:15.83681Z","session_id":"01a14db8-eb42-7903-8efb-4a97ccd1bc08","outcome":"returned"}] -->
