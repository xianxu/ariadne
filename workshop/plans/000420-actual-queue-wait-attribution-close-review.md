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

---

## Re-review — 2026-10-17T04:16:39-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 420 — Actual: queued suite status polling bridges idle time into focused-hours samples |
| repo | ariadne |
| issue file | workshop/issues/000420-actual-queue-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 409e789d83910de366d11b8a536c2c6d8dc9900e..1b66ce929c739364bea85940917c7d571fbf5bde |
| command | sdlc close --issue 420 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T04:16:39-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-1 and BR-3 are verified fixed, with regression tests that fail when their fixes are removed. BR-2’s running-poll case is repaired, but failed or empty polls still finalize unfinished work as positive evidence. That remaining correctness gap blocks SHIP.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      actualmeasurement.go retains QueueBindingSourceIDs alongside contributing sources. TestActualQueueMeasurementRetainsBindingSource passes at head and fails when that retention is removed; it also verifies successful replay.
  - id: BR-2
    disposition: not-addressed
    note: |
      queue.go:75-91 and 138-150 still treat any poll result lacking a recognized running marker as completion, deleting the yielded association before emitting work. A temporary pinned-head regression yielded stage work at minute 1, then polled at minute 10: transport-error, empty, and unknown-process responses at minute 11 each produced an 11-minute toolWork span without completion evidence. The running-poll regression passes and fails without its fix, but the required confirmed-completion rule remains unenforced. Preserve unconfirmed outcomes through an explicit transition model and test failed/empty polls followed by missing and eventual terminal results (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).
  - id: BR-3
    disposition: addressed
    note: |
      actualcorrection.go:820 includes QueueV1 in ownership-proven calibration trust. The parameterized legacy/engagement/queue regression passes at head; removing QueueV1 makes its queue case fail with correction trust false, want true. The production measurement-method consumer sweep found no remaining omission.
```

1. **Strengths**

   - Snapshot retention now preserves zero-activity sources required for replay, with an end-to-end regression.
   - Queue endpoint validation preserves legacy compatibility while rejecting inconsistent new records.
   - Source-scoped interval tests verify overlap conservation and isolation from other active sources.
   - Atlas and help document queue attribution, retained evidence, and historical replay semantics.

2. **Critical findings**

   **BR-2 remains open — `cmd/sdlc/internal/activetime/queue.go:91` and `:141`.** Absence of a running marker is insufficient evidence of completion. Failed or empty poll responses currently restore unfinished work inside queue exclusions and discard the association needed to follow later completion.

   Model running, confirmed terminal, and unconfirmed outcomes explicitly. Poll failures must preserve uncertainty and must not emit work.

3. **Important findings**

   None newly raised.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Activetime, gatestate, and transcripts package tests passed, as did targeted measurement, correction-trust, selection, decoder, and recorder tests. All three prior fixes were mutation-checked. The additional failed-poll regression failed on pinned-head code as described above. `git diff --check` passed. The disposable mutation checkout was removed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing attempt decoder and interval helpers are reused.
   - **ARCH-PURE — pass:** selection, parsing, and subtraction logic are directly testable without IO.
   - **ARCH-PURPOSE — flag:** BR-2 still permits unsupported positive work attribution.
   - **ARCH-MOCK — pass:** changed integration paths reuse existing seams and portable fixtures.
   - **ARCH-CONSTRAINTS — pass:** attempt-reader and snapshot bounds remain enforced.
   - **ARCH-SECURE — flag:** unknown poll output becomes fabricated completion evidence.
   - **ARCH-ORDER — flag:** poll uncertainty lacks an explicit transition model.
   - **ARCH-FUNERAL — pass:** attempt retention and bounded snapshot lifecycles cover the added evidence.

   Core-concept entities exist at their declared paths. Atlas coverage is present; no new command, flag, or config surface requires a README addition.

7. **Plan revision recommendations**

   Add a `## Revisions` entry specifying that failed, empty, and unrecognized poll outcomes remain unconfirmed, preserve their yielded association, and cannot restore work. Name regression sequences covering both absent and eventual confirmed completion.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"9d6d4e65-fa0b-4d35-a569-6dcab083436f","repository":"github.com/xianxu/ariadne","issue_id":"000420","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c1e66e49e965bbea","reviewed_head":"1b66ce929c739364bea85940917c7d571fbf5bde","slot":"ariadne:1","requested_at":"2026-10-17T11:12:57.16443Z"},"phase":"finished","pid":41574,"started_at":"2026-10-17T11:12:57.377331Z","finished_at":"2026-10-17T11:16:38.5981Z","session_id":"01a14990-a2f0-7c21-a6c0-ed96d348eab8","outcome":"returned"}] -->
