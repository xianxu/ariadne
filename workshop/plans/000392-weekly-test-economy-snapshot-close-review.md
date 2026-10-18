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

---

## Re-review — 2026-10-18T00:30:05-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 392 — Weekly test-economy snapshot from local checks and landings |
| repo | ariadne |
| issue file | workshop/issues/000392-weekly-test-economy-snapshot.md |
| boundary | whole-issue close |
| milestone | — |
| window | 826a55dd130c59a4ae255c147ffafb3e7a2f2c1b..336692bd0b707fbdc6e14d64605fd7ca19b4b84a |
| command | sdlc close --issue 392 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-18T00:30:05-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Five prior findings are addressed with supporting evidence. BR-5 remains open: the durable snapshot and ops-project link are still deferred. A new adversarial test also demonstrates that duplicate landing identities can produce a certified numeric ratio, contrary to the Plan’s validation contract.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The retained-history floor is consumed by validation and populated by the command. TestValidateLandingEvidenceRequiresRetainedWholeIssueHistory passes at head and fails when the history guard is removed.
  - id: BR-2
    disposition: addressed
    note: |
      Prior admissions are compared against the original successful enqueue. TestValidateLandingEvidenceChecksEveryPriorAdmission covers chronological and reordered retries; reverting the threshold fix makes it fail.
  - id: BR-3
    disposition: addressed
    note: |
      Reconciliation now uses suiteAttemptCompletedAt. TestLegacyWeeklyAttemptUsesPhysicalCompletionForReconciliation passes at head and fails when FinishedAt-only selection is restored.
  - id: BR-4
    disposition: addressed
    note: |
      README.md:505–516 now documents snapshot invocation, both flags, weekly boundaries, immutable publication and nullable coverage, consistent with weeklystats_cmd.go and weeklystats_write.go.
  - id: BR-5
    disposition: not-addressed
    note: |
      The hashed temporary pilot and reported pair rerun improve the evidence, but issue Log line 91 still explicitly defers durable brain publication and the ops link. The immutability rationale does not fulfill the unchanged Done-when requirements; a committed snapshot path and confirmed project link remain outstanding.
  - id: BR-6
    disposition: addressed
    note: |
      Collection now receives the window and restricts merge inventory and graph candidates. TestReadWeeklyEvidenceInventoriesOnlyRequestedWeek passes at head and fails when the collection restrictions are reverted.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: evidence-boundary-validation
    title: |
      Different issues can certify the same landing identity twice.
    detail: |
      cmd/sdlc/weeklyattribution.go:194–211 detects repeated issue IDs but does not enforce uniqueness of PR, branch or bound merge across issues. A disposable regression fixture with two done cards and separate valid tickets claiming the same PR/branch/merge produces landed_issues=2, ratio=0 and no coverage warnings, violating plan line 23. This is the 2nd finding in family evidence-boundary-validation: state and enforce the cohort-wide identity rule across sequence, issue, branch, PR and merge bindings, reject conflicting aliases as unknown, and test reordered cross-record collisions rather than fixing only this fixture (ARCH-SECURE, ARCH-PURPOSE).
```

1. **Strengths**

   - Physical weekly counts and landing-cohort attribution remain explicitly separated.
   - History certification conservatively rejects expired or insufficient recorder history.
   - Publication uses atomic no-replace creation, with tests for conflicting writers, identical reruns and stale temporary-file cleanup.
   - README, atlas and embedded help now describe the new surface.

2. **Critical findings**

   - The duplicate-identity defect above blocks certification. Validate identities across the entire cohort before emitting accepted landings; contradictions must null coverage rather than silently enlarge the denominator.

3. **Important findings**

   - **BR-5 remains open**, at `workshop/issues/000392-weekly-test-economy-snapshot.md:91`. Complete durable capture and project-link confirmation before whole-issue acceptance.

4. **Minor findings**

   - None.

5. **Test coverage notes**

   The focused weekly tests passed. In a disposable checkout, reverting each behavioral correction for BR-1, BR-2, BR-3 and BR-6 made its permanent regression fail. The new duplicate-identity probe failed against pinned head with `2 landings, ratio=0, unknown=[]`. The disposable checkout was removed; the reviewed checkout remains unchanged. No full suite was rerun.

6. **Architectural notes**

   | Principle | Result |
   |---|---|
   | ARCH-DRY | Pass: existing attempt reader, aggregation and quantiles are reused. |
   | ARCH-PURE | Pass: calculations and attribution are directly tested without IO; command dependencies isolate collection/publication. |
   | ARCH-PURPOSE | Flag: duplicate evidence can certify a false denominator; durable acceptance remains incomplete. |
   | ARCH-MOCK | Flag: portable Git fixtures exercise the production seam, but BR-5’s rollout acceptance remains pending. |
   | ARCH-CONSTRAINTS | Pass: weekly graph selection and explicit input, output and timeout limits are implemented. |
   | ARCH-SECURE | Flag: cross-record landing identity conflicts remain trusted. |
   | ARCH-ORDER | Pass: prior-admission validation now survives input reordering; publication arbitrates concurrent writers atomically. |
   | ARCH-FUNERAL | Pass: temporary files have cleanup; permanent snapshots have declared retention and backup ownership. |

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining the cohort-wide identity invariant and collision-test enumeration. Reconcile the post-SHIP publication sequence with the issue’s durable-capture acceptance requirement so the boundary has one explicit completion rule.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6b5af64f-5c34-4e7b-a48d-c8bf57fceb8c","repository":"github.com/xianxu/ariadne","issue_id":"000392","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/2201d26168ab239c","reviewed_head":"336692bd0b707fbdc6e14d64605fd7ca19b4b84a","slot":"ariadne:1","requested_at":"2026-10-18T07:27:35.753578Z"},"phase":"finished","pid":17415,"started_at":"2026-10-18T07:27:35.853541Z","finished_at":"2026-10-18T07:30:05.31033Z","session_id":"01a14de8-aa9a-7d71-a325-e2fb39d06ea3","outcome":"returned"}] -->
