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

---

## Re-review — 2026-10-17T04:52:29-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 420 — Actual: queued suite status polling bridges idle time into focused-hours samples |
| repo | ariadne |
| issue file | workshop/issues/000420-actual-queue-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..96da94cc8a46984bb2a8265cf991bb93cffb2b0f |
| command | sdlc close --issue 420 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T04:52:29-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Queue evidence retention, replay, phase splitting, and calibration trust are well covered. However, BR-2 remains reproducible: a poll reusing its yielded command’s call ID bypasses completion verification and fabricates work evidence. This blocks shipping.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      QueueBindingSourceIDs preserves binding transcripts even after all their activity is subtracted. TestActualQueueMeasurementRetainsBindingSource passes and verifies successful replay.
  - id: BR-2
    disposition: not-addressed
    note: |
      cmd/sdlc/internal/activetime/queue.go:133-142 registers polls without duplicate-ID validation and distinguishes polls from original results by comparing call IDs. A temporary pinned-head regression yielded stage command "stage" into session 7, then polled that session using call_id "stage" with output "unknown process". The parser emitted an 11-minute toolWork span without completion evidence. Existing unconfirmed-poll tests pass at head and fail against the pre-fix implementation, but omit this identity collision. This repeats family work-evidence-requires-confirmed-completion. Enforce the rule across every call registration: identities must be unique, poll status must be explicit, and yielded work may complete only on confirmed terminal evidence. Implement the requested transition model and cover duplicate IDs, concurrent poll aliases, and late results rather than adding another instance-specific guard (ARCH-ORDER, ARCH-SECURE, ARCH-PURPOSE).
  - id: BR-3
    disposition: addressed
    note: |
      actualcorrection.go includes QueueV1 in ownership-proven calibration trust. The parameterized legacy, engagement, and queue correction-trust regression passes; the production method-consumer sweep found no remaining omission.
```

1. **Strengths**

   - Queue subtraction uses source-scoped interval unions, preserving independent work without double-subtracting overlapping waits.
   - Retained snapshots preserve queue-binding sources with zero corrected activity, supporting replay.
   - Exact queue endpoints have strict validation while legacy records retain their documented interval convention.
   - #416’s referenced tracker-card blob confirms generation-bound exclusion: previous actual `1.12`, new actual `N/A`, with reason and completion binding preserved.

2. **Critical findings**

   - **BR-2 remains open — `cmd/sdlc/internal/activetime/queue.go:133–142`.** Poll registration bypasses the identity checks at lines 66–74. Reusing the original ID makes `ptr.id != p.CallID` false, allowing an unconfirmed result to finalize work. Separate operation state from call identity and route all registrations/results through an explicit transition model.

3. **Important findings**

   None newly raised.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed full `activetime`, `gatestate`, and `transcripts` package tests.
   - Passed targeted SDLC measurement, correction, selection, decoding, and recorder tests.
   - Confirmed the committed unconfirmed-poll regression fails against the pre-fix parser.
   - A temporary Go-overlay regression fails at pinned head with an unexpected 11-minute work span for the duplicate-ID sequence.
   - `git diff --check` passed; the reviewed checkout remains unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared decoder, interval helpers, and measurement path.
   - **ARCH-PURE — pass:** parsing and interval logic are directly testable without IO.
   - **ARCH-PURPOSE — flag:** unfinished work can still restore excluded queue time.
   - **ARCH-MOCK — pass:** inspected integration changes reuse existing dependency boundaries and portable fixtures.
   - **ARCH-CONSTRAINTS — pass:** bounded attempt reads and snapshots remain enforced.
   - **ARCH-SECURE — flag:** duplicate persisted identities can fabricate completion evidence.
   - **ARCH-ORDER — flag:** poll identity substitutes for explicit operation state.
   - **ARCH-FUNERAL — pass:** existing attempt retention and snapshot lifecycle cover the additions.

   Atlas documentation covers the new measurement flow. No new command, flag, or configuration key requires a README update.

7. **Plan revision recommendations**

   Append a `## Revisions` entry requiring one transition model for original calls and polls, with explicit identity uniqueness, terminal-state enforcement, and sequence tests for duplicate IDs, overlapping polls, and late results.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"710c6dda-e9a2-4dbd-9e86-c260fe6244b4","repository":"github.com/xianxu/ariadne","issue_id":"000420","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/dbf706a03bbfb60e","reviewed_head":"96da94cc8a46984bb2a8265cf991bb93cffb2b0f","slot":"ariadne:1","requested_at":"2026-10-17T11:49:22.097679Z"},"phase":"finished","pid":82020,"started_at":"2026-10-17T11:49:22.265368Z","finished_at":"2026-10-17T11:52:29.480943Z","session_id":"01a149b1-f84e-72b0-ae76-73827379e39e","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T06:07:23-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 420 — Actual: queued suite status polling bridges idle time into focused-hours samples |
| repo | ariadne |
| issue file | workshop/issues/000420-actual-queue-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 2bfbc4fa27ed30da2455ad38d4eaea0ddf82a774..3290d6f0f592307fb36466b54a2e112bc81f1517 |
| command | sdlc close --issue 420 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T06:07:23-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Exact queue endpoints, slot attribution, phase splitting, and retained replay are implemented and tested. BR-1 and BR-3 are verified fixes. BR-2 remains blocking: the Claude result path still converts unconfirmed tool outcomes into completed work, restoring queue time to calibrated actuals.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      TestActualQueueMeasurementRetainsBindingSource passes; temporarily removing binding-source retention makes it fail because the zero-contribution source disappears.
  - id: BR-2
    disposition: not-addressed
    note: |
      The poll-based completion path was removed, but the confirmed-completion rule remains violated at cmd/sdlc/internal/activetime/queue.go:135. Claude tool_result content is passed as raw JSON, and is_error is ignored. Temporary regressions with empty, unknown-process and transport-error content all emit completed work. An end-to-end fixture with is_error:true and content:"unknown process" restores 11 minutes during a recorded 60-minute queue instead of refusing measurement. This remains family work-evidence-requires-confirmed-completion (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER).
  - id: BR-3
    disposition: addressed
    note: |
      TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust passes; temporarily removing queue-v1 from the trust condition makes its queue case fail with false trust.
```

1. **Strengths**

   - Recorder endpoints and decoder validation establish a consistent, replayable queue interval.
   - Tests cover phase conservation, independent overlapping work, multiple sources, legacy evidence refusal, and replay without ambient attempt files.
   - Binding-source retention and calibration trust both have regressions that fail when their fixes are removed.

2. **Critical findings**

   **BR-2 — Unconfirmed Claude results become completed work.** At [queue.go:135](/Users/xianxu/Library/Caches/sdlc/review-checkouts/c22bf0eed05685f4/cmd/sdlc/internal/activetime/queue.go:135), `string(b.Content)` retains JSON quotation marks. Consequently, empty content becomes `""`, and failure text begins with a quotation mark, bypassing `isFailedToolOutput`. The parser also drops `is_error`. The shared result handler then emits `toolWork`, which restores the interval during queue subtraction.

   Fix the class: enumerate supported native result envelopes, decode them into shared typed completion/uncertainty outcomes, and preserve error metadata. Cover Claude strings, content blocks and error results alongside Codex scalar, object and block results. Unknown outcomes must produce uncertainty. **ARCH-DRY, ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER.**

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Passed: full `activetime`, `gatestate`, and `transcripts` package tests; targeted SDLC queue, ownership, correction and endpoint tests; recorder integration regression; `git diff --check`.
   - Temporary parser regressions failed for all three unconfirmed Claude outcomes.
   - Temporary production-engine regression failed with **11 active minutes and 11 retained-work minutes**, demonstrating the measurement impact.
   - BR-1/BR-3 counterfactual tests failed as expected. Temporary overlays were removed; tracked files remain unchanged.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — flag:** native result paths need shared outcome normalization; see BR-2.
   - **ARCH-PURE — pass:** interval arithmetic and transcript interpretation remain directly testable without IO.
   - **ARCH-PURPOSE — flag:** unconfirmed work can still contaminate focused effort; see BR-2.
   - **ARCH-MOCK — pass:** this change introduces no new external dependency; recorder verification uses existing seams.
   - **ARCH-CONSTRAINTS — pass:** existing bounded attempt readers and snapshot limits constrain the added evidence.
   - **ARCH-SECURE — flag:** result provenance and error metadata are lost before interpretation; see BR-2.
   - **ARCH-ORDER — flag:** failed/unknown outcomes still transition to completed work on one supported transcript path.
   - **ARCH-FUNERAL — pass:** attempts retain their existing pruning lifecycle; snapshots retain existing size bounds and archival ownership.

   Core-concept locations match the diff. Atlas documents the revised flow. No new command, flag or configuration syntax requires a README addition.

7. **Plan revision recommendations**

   Append a `## Revisions` entry enumerating every supported native result envelope and its completion/uncertainty contract. Require shared outcome decoding and cross-harness regression coverage for the complete BR-2 family before declaring it addressed.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"4311457e-973d-4b8e-bd74-ef51ce51e9db","repository":"github.com/xianxu/ariadne","issue_id":"000420","recipe":"milestone-review","round":4,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/c22bf0eed05685f4","reviewed_head":"3290d6f0f592307fb36466b54a2e112bc81f1517","slot":"ariadne:1","requested_at":"2026-10-17T13:02:39.137015Z"},"phase":"finished","pid":45750,"started_at":"2026-10-17T13:02:39.349373Z","finished_at":"2026-10-17T13:07:22.532649Z","session_id":"01a149f5-1100-76f3-aed9-894eae2881cc","outcome":"returned"}] -->

---

## Re-review — 2026-10-17T06:50:08-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 420 — Actual: queued suite status polling bridges idle time into focused-hours samples |
| repo | ariadne |
| issue file | workshop/issues/000420-actual-queue-wait-attribution.md |
| boundary | whole-issue close |
| milestone | — |
| window | 146796d884ccf9a9d589a8766d7627762b4571e6..6d1a922964acbd953d95d14c54a1733abf4f004f |
| command | sdlc close --issue 420 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-17T06:50:08-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The revised dispatch-only policy resolves BR-2’s result-completion problem, and the existing targeted tests pass. Two correctness gaps still allow unsupported calibration samples: native wrappers can hide interactive input behind a poll classification, and close can proceed despite incomplete queue evidence.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      QueueBindingSourceIDs preserves fully subtracted sources; TestActualQueueMeasurementRetainsBindingSource passes and verifies replay.
  - id: BR-2
    disposition: addressed
    note: |
      Queue attribution no longer derives work from results or poll completion. Yielded-work and Claude-error regression tests pass; disabling dispatch refusal in a temporary overlay makes both fail.
  - id: BR-3
    disposition: addressed
    note: |
      QueueV1 participates in correction calibration trust; TestActualCorrectionProvenOwnershipUpdatesCalibrationTrust passes for legacy, engagement and queue methods.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: work-evidence-requires-confirmed-completion
    title: |
      Native wrapper classification treats unknown interactive input as an empty poll
    detail: |
      cmd/sdlc/internal/activetime/queue.go:51-54 defaults to Poll=true when chars is present but is not a quoted literal. For example, const r=await tools.write_stdin({session_id:7,chars:String.fromCharCode(121)});text(r) sends nonempty input but bypasses overlapping-work refusal. A temporary regression asserting this is not a poll fails on head. This is the 2nd finding in family work-evidence-requires-confirmed-completion. State and enforce the class-wide rule: exempt only calls proven to contain exclusively empty-input polling; unknown expressions, spreads and other unrecognized argument forms must remain non-polls. Cover that enumeration rather than adding another isolated regex exception (ARCH-PURPOSE, ARCH-SECURE).
  - id: new
    severity: Critical
    kind: defect
    family: queue-evidence-completeness-before-calibration
    title: |
      Close admits calibration evidence after queue records become unreadable or unknown
    detail: |
      cmd/sdlc/actualinputs.go:328-355 rejects malformed-record and unknown-provenance warnings only for correction candidates; close rejects only reader limits and warnings containing legacy suite attempt. A temporary production-resolver regression with a truncated attempts/aabb.json fails: actualCloseCandidate succeeds with active-time-v3-engagement-v1 and an invalid-record warning. Likewise, actualqueue.go:32-36 skips missing queue_seconds or slot evidence with only a warning. The sample can therefore retain unexcluded queue activity and reach numeric adoption. Apply one completeness policy across close and correction: evidence that cannot establish relevant queue coverage must prevent calibration, unless it can be proven irrelevant. Add regressions for malformed, unreadable and missing-field records (ARCH-PURPOSE, ARCH-SECURE).
```

1. **Strengths**

   - Exact queue endpoints and duration share the same recorded instants, avoiding serialization disagreement.
   - Source-scoped interval subtraction preserves foreign-slot activity and unions overlapping waits.
   - Retained source prefixes support replay even when queue subtraction removes all activity from a source.
   - Removing result decoding substantially simplifies the queue-attribution contract.

2. **Critical findings**

   - `queue.go:51-54`: require positive evidence that a wrapper performs only empty-input polling.
   - `actualinputs.go:328-355`: enforce queue-evidence completeness consistently before numeric adoption.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed the full `activetime`, `gatestate` and `transcripts` package tests, targeted SDLC queue/ownership/correction/recorder tests, and range `git diff --check`. Both new adversarial overlay tests failed as described. Disabling dispatch refusal also made the existing yielded-work and Claude-error regressions fail, confirming meaningful BR-2 coverage. Temporary overlays were removed; tracked files remain unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** existing attempt validation and interval helpers are reused.
   - **ARCH-PURE — pass:** classification and interval arithmetic are independently testable without IO.
   - **ARCH-PURPOSE — flag:** the two findings leave incomplete coverage of the calibration contract.
   - **ARCH-MOCK — pass:** no new external dependency seam; recorder tests use existing injected infrastructure.
   - **ARCH-CONSTRAINTS — pass:** bounded attempt reading and snapshot limits remain enforced.
   - **ARCH-SECURE — flag:** unknown input and damaged persisted evidence can become trusted measurement inputs.
   - **ARCH-ORDER — pass:** queue attribution no longer carries yielded-call completion state.
   - **ARCH-FUNERAL — pass:** endpoint fields use the existing attempt retention lifecycle; snapshots retain compact selected evidence.

   Core-concept entities exist at their documented paths. Atlas and help describe the revised policy. No new command, flag or configuration key requires a README addition.

7. **Plan revision recommendations**

   Add `## Revisions` entries defining conservative poll classification across literal, computed and spread arguments, and one queue-evidence completeness policy shared by close and correction, with the adversarial regressions listed above.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"b59b7f5c-bff0-4899-8fdd-25d5a7bce1fc","repository":"github.com/xianxu/ariadne","issue_id":"000420","recipe":"milestone-review","round":5,"attempt":1,"agent":"codex","root":"/Users/xianxu/Library/Caches/sdlc/review-checkouts/0a9959ba0c418399","reviewed_head":"6d1a922964acbd953d95d14c54a1733abf4f004f","slot":"ariadne:1","requested_at":"2026-10-17T13:39:12.327163Z"},"phase":"finished","pid":72415,"started_at":"2026-10-17T13:39:12.421765Z","finished_at":"2026-10-17T13:50:08.216318Z","session_id":"01a14a16-8692-7d43-9b76-f81b6a4d1518","outcome":"returned"}] -->
