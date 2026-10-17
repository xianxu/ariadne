# Boundary Review — ariadne#420 (whole-issue close)

| field | value |
|-------|-------|
| issue | 420 — Actual: queued suite status polling bridges idle time into focused-hours samples |
| repo | ariadne |
| issue file | workshop/issues/000420-actual-queue-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0a1f6a454d8192e798fade61106cedede1774795..fe2155eda0572e95b2e5243c164698decddfcf3a |
| command | sdlc close --issue 420 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T02:51:32-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation correctly separates recorded queue waits from activity and preserves historical measurement generations. However, two reproduced correctness bugs block shipping: unfinished work can restore queue time, and a successful measurement can produce an unreplayable snapshot. Queue-aware corrections also omit an existing calibration-trust update.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: retained-evidence-covers-replay-dependencies
    title: |
      Snapshot filtering drops sources required to replay queue exclusions
    detail: |
      cmd/sdlc/actualmeasurement.go:65-79 retains only sources contributing activity after subtraction, while retaining every QueueWait. When a wait removes all activity from its source and another source contributes work, measurement succeeds but drops the wait's binding transcript. A temporary regression reproduced successful measurement followed by replay failure: "queue attempt aabb has no owning wait invocation". Retain all queue-binding source prefixes and provenance, including sources contributing zero corrected activity (ARCH-PURPOSE).
  - id: new
    severity: Critical
    kind: defect
    family: work-evidence-requires-confirmed-completion
    title: |
      A still-running poll finalizes yielded work as completed evidence
    detail: |
      cmd/sdlc/internal/activetime/queue.go:138 preserves running polls only for toolWait. A yielded stage command followed by a poll returning "still running" instead loses its yielded association and reaches result(), which emits a work span without completion evidence. A temporary regression produced an 11-minute work span from an unfinished invocation. Preserve running state for both kinds and emit work only after confirmed completion; cover repeated polls, missing completion and eventual completion through an explicit transition model (ARCH-ORDER, ARCH-SECURE).
  - id: new
    severity: Important
    kind: defect
    family: measurement-method-consumer-completeness
    title: |
      Queue-aware corrections fail to establish calibration trust
    detail: |
      cmd/sdlc/actualinputs.go:353 introduces MeasurementMethodQueueV1, but cmd/sdlc/actualcorrection.go:820 sets WindowTrusted only for MeasurementMethodEngagementV1. Correcting an untrusted sample with ownership-proven queue evidence therefore leaves it excluded from calibration. Substituting QueueV1 into the existing TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust reproduced "correction trust false, want true". Include the new method in this consumer and parameterize the regression across both ownership-proven methods (ARCH-PURPOSE).
```

1. **Strengths**

   - Exact queue endpoints and duration consistency are validated, while legacy reconstruction remains explicit.
   - Interval subtraction preserves overlapping work and isolates other sources; the native and phase-split regressions pass.
   - Historical replay and correction candidates have separate purposes.
   - Atlas and command help explain queue provenance and legacy timing limitations.

2. **Critical findings**

   - Retain queue-binding sources even when subtraction removes their entire activity.
   - Keep yielded work pending through running polls; require completion before restoring time.

3. **Important findings**

   - Extend calibration-trust handling to the queue-aware measurement method.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed the active-time and gatestate package tests, targeted CLI tests covering queue/ownership/endpoint/local-check behavior, and `git diff --check`.

   Three temporary overlay regressions reproduced the findings above. The overlays were removed; the checkout remains unchanged. The full suite was not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** no material duplication identified.
   - **ARCH-PURE — pass:** interval and parsing logic remain directly testable without IO; declared pure entities exist at their documented paths.
   - **ARCH-PURPOSE — flag:** replay evidence and calibration-method consumers are incomplete.
   - **ARCH-MOCK — pass:** new IO uses existing dependency seams; no new external dependency.
   - **ARCH-CONSTRAINTS — pass:** existing reader and snapshot bounds remain enforced.
   - **ARCH-SECURE — flag:** an unfinished operation becomes positive work evidence.
   - **ARCH-ORDER — flag:** yielded-work transitions mishandle running poll outcomes.
   - **ARCH-FUNERAL — pass:** attempt retention and bounded snapshot storage cover the added evidence.

7. **Plan revision recommendations**

   Add `## Revisions` entries requiring:

   - Retention of every source needed to replay queue binding, regardless of corrected contribution.
   - Explicit yielded-tool transitions and sequence tests for unfinished and eventually completed work.
   - Enumeration and regression coverage of all measurement-method consumers, including calibration trust.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"42c1741a-9784-4319-b090-3f12c52210e2","repository":"github.com/xianxu/ariadne","issue_id":"000420","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/eb5e6ff16e87a774","reviewed_head":"fe2155eda0572e95b2e5243c164698decddfcf3a","slot":"ariadne:1","requested_at":"2026-10-17T09:46:59.009541Z"},"phase":"finished","pid":45129,"started_at":"2026-10-17T09:46:59.129733Z","finished_at":"2026-10-17T09:51:31.199571Z","session_id":"01a14941-ec21-71e2-8cb9-f6e498416133","outcome":"returned"}] -->
