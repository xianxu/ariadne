# Boundary Review — ariadne#589 (whole-issue close)

| field | value |
|-------|-------|
| issue | 589 — Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them |
| repo | ariadne |
| issue file | workshop/issues/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 945a8168578afcbd94ed27965c283d0a6483aeb1..570681b8038d97147d5a9bd3c7529d528c0625f9 |
| command | sdlc close --issue 589 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T13:17:39-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation separates observation from classification, preserves warnings in ordinary saved-result replay, and adds useful lifecycle tests. Focused tests pass, but diagnostics can change timeout outcomes, creator attribution does not establish process incarnation, and mutable metadata can make existing segments appear new. The concept table also contradicts the pinned diff.

Done-when: current — warning IDs, an abandoned-segment fixture, harness guidance, and a filed downstream issue still cover the revised advisory scope.

```findings
findings:
  - id: new
    severity: Critical
    family: advisory-outcome-isolation
    title: |
      SysV observation time can turn a completed command into a timeout
    detail: |
      cmd/sdlc/localcheck_run.go:676-687 runs observations under an already-running command timeout, then line 722 checks that timeout after observation completes. A successful command followed by a slow after-snapshot can therefore become TimedOut/LaneOverrun; the before-snapshot also consumes command execution time. ARCH-CONSTRAINTS and ARCH-PURPOSE: isolate diagnostic budgets from execution budgets and capture the execution outcome before post-run diagnostics. Add regressions with delayed before/after observers proving diagnostics cannot change an otherwise successful outcome.
  - id: new
    severity: Critical
    family: creator-incarnation-evidence
    title: |
      A nonempty start time does not establish that the witnessed process created the segment
    detail: |
      cmd/sdlc/internal/sysvshm/report.go:29-33 matches only the numeric creator PID and checks that Witness.Start is nonempty. No code compares that incarnation with evidence from the segment. If a concurrent creator exits and its PID is reused by a run-group process, the unrelated segment receives that process's name and a witnessed-creator label. ARCH-SECURE and ARCH-ORDER: require evidence linking the creator incarnation to the allocation, otherwise retain unknown attribution. Test PID reuse into the owned group; the existing fake lookup returning false does not exercise this case.
  - id: new
    severity: Critical
    family: allocation-identity-stability
    title: |
      Mutable change time is treated as allocation identity
    detail: |
      cmd/sdlc/internal/sysvshm/report.go:11-26 uses the entire Segment, except Attachments, as identity. Changed comes from CTIME/ctime, which is the last shmctl change time rather than an immutable allocation generation. Updating permissions through IPC_SET can therefore make a pre-existing unattached segment appear newly allocated. ARCH-PURPOSE: separate stable identity from mutable metadata, treat ambiguous changes conservatively, and add a regression where one existing segment changes metadata without being replaced.
  - id: new
    severity: Critical
    family: concept-table-code-consistency
    title: |
      The Core concepts table claims a modification absent from the pinned range
    detail: |
      workshop/plans/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test-plan.md:40 marks savedAttemptRecord in cmd/sdlc/localcheck_republish.go as modified, but that file has no base-to-head diff. Replay already carries the supplied tail; the new behavior resides in statusDescription. Per the required Core concepts cross-check, correct the row to existing and append a Revisions entry explaining reuse of the existing validation path.
  - id: new
    severity: Important
    family: diagnostic-consumer-completeness
    title: |
      Resume publication drops warnings produced only by the child run
    detail: |
      cmd/sdlc/localcheck_resume_saved.go:117-138 builds the composite record using only the parent's tail, while cmd/sdlc/localcheck_resume.go:228 discards the child's hashAndTail tail. The newly instrumented resume batch can warn in its child log, yet fresh and saved composite publication omit the promised SysV warning prefix when the parent was clean. ARCH-PURPOSE: derive diagnostic status from both digest-verified logs and test a clean parent followed by a leaking child through publication and republish.
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README update appears missing for the new qualification usage
    detail: |
      atlas/workflow/sdlc-binary.md:3394 introduces TestNativeOwnedSysVQualification and WF_SYSV_QUALIFY=1, but README.md is unchanged in the pinned range. The requested README gate covers new usage steps and configuration keys. Add a concise README pointer explaining host warnings and linking the operator-cleared qualification procedure.
```

1. **Strengths**

   - Strict parsers reject partial tables rather than reporting fabricated empty observations.
   - Observation is explicitly injected into production constructors; unconfigured unit tests avoid live host probing.
   - Ordinary replay retains the digest-verified warning footer, with compatibility coverage for older logs.
   - Atlas and embedded help explain attribution uncertainty, VM exclusions, and PostgreSQL shutdown responsibilities.

2. **Critical findings**

   The four Critical findings above require fixes before admission. The concept-table correction is documentation-only; the three behavior corrections need meaningful regression tests.

3. **Important findings**

   Preserve child-run warnings in resume publication and update README usage documentation.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed on the pinned head:

   `go test ./cmd/sdlc/internal/sysvshm ./cmd/sdlc -run 'SysV|Sysv|TestVerbContextsReachTrackerReads' -count=1 -timeout 3m`

   Existing coverage validates ordinary warning persistence, cancellation, observation failure, admission refusal, and probe counts. It does not cover delayed diagnostics crossing the execution deadline, creator PID reuse into the owned group, mutable CTIME, or child-only resume warnings. Native allocation and private-executable qualification were not independently exercised in this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared classification, rendering, and status extraction.
   - **ARCH-PURE — pass:** deterministic parsing/classification separated from host observation.
   - **ARCH-PURPOSE — flag:** allocation classification and resume publication leave contract gaps.
   - **ARCH-MOCK — flag:** creator lifecycle tests do not reproduce the reuse scenario behind the attribution claim.
   - **ARCH-CONSTRAINTS — flag:** independently bounded diagnostics still affect execution deadlines.
   - **ARCH-SECURE — flag:** numeric PID matching substitutes for creator-incarnation evidence.
   - **ARCH-ORDER — flag:** creator exit/reuse between allocation and observation is not distinguished.
   - **ARCH-FUNERAL — pass:** observations remain scoped values; durable diagnostics reuse existing logs.
   - **ARCH-EVOLVE — pass:** no persisted schema change; older logs remain accepted.

7. **Plan revision recommendations**

   Append `## Revisions` entries documenting execution-budget isolation, conservative allocation/creator identity rules, resume diagnostic propagation, and `savedAttemptRecord` as an existing reused entity. Update completed test claims to name the added regressions.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"79949986-fe8e-4cc8-ac8f-ecea0b0d19a3","repository":"github.com/xianxu/ariadne","issue_id":"000589","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/f334c81f122f4d93","reviewed_head":"570681b8038d97147d5a9bd3c7529d528c0625f9","slot":"ariadne:3","requested_at":"2026-10-24T20:15:09.741732Z"},"phase":"finished","pid":10796,"started_at":"2026-10-24T20:15:09.901024Z","finished_at":"2026-10-24T20:17:38.753123Z","session_id":"01a16f8d-8eaf-7242-bed6-0be76fe8711f","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T13:33:54-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 589 — Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them |
| repo | ariadne |
| issue file | workshop/issues/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 945a8168578afcbd94ed27965c283d0a6483aeb1..e0b3df15d51488c2e22000b67ac0c6fc0dfea4b9 |
| command | sdlc close --issue 589 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T13:33:54-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation now isolates execution deadlines, uses conservative segment identity, preserves verified child warnings, and documents the new usage. Focused tests pass. Two items still block admission: cancellation during post-run diagnostics changes a completed result, and BR-2 lacks a regression that detects the former attribution behavior.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Execution timeout creation now follows the before-probe, and deadline state is captured before the after-probe. TestSysVDiagnosticsCannotConsumeExecutionBudget passes at HEAD and fails with those changes reversed through a temporary Go overlay.
  - id: BR-2
    disposition: not-addressed
    note: |
      Production witness attribution was removed, but the required regression evidence remains incomplete. TestSysVCurrentProcessNameIsNotCreatorProof supplies only a Segment, with no competing incarnation or witness. Restoring the former Witness-bearing Candidate and conditional attribution renderer through an overlay leaves this test green. Add a regression that detects restoration of the unsafe process-witness path.
  - id: BR-3
    disposition: addressed
    note: |
      Delta now compares numeric IDs independently of mutable metadata. TestDeltaMetadataChangesAreNotNewIDs passes at HEAD and fails when the former Segment-based identity comparison is restored through an overlay.
  - id: BR-4
    disposition: addressed
    note: |
      The plan's Core concepts row now says savedAttemptRecord is existing, reused; its Revisions explain reuse. The pinned range leaves localcheck_republish.go unchanged and adds diagnostic projection in localchecks.go and localcheck_resume_saved.go.
  - id: BR-5
    disposition: addressed
    note: |
      resolvedSavedRecord derives child warnings after proof validation. Saved and fresh publication regressions pass at HEAD and both fail when the child-warning union is removed through an overlay; child digest mutation is rejected.
  - id: BR-6
    disposition: addressed
    note: |
      README.md:1156-1160 explains advisory warnings and WF_SYSV_QUALIFY=1, linking the atlas section that describes operator-cleared qualification and exact-owned-segment cleanup.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: advisory-outcome-isolation
    title: |
      Cancellation during the after-snapshot discards a completed command outcome
    detail: |
      cmd/sdlc/localcheck_run.go:684-725 freezes execution deadline state before diagnostics but reads ownedCtx.Err and queue interruption afterward. A temporary regression using fakeSysVObserver.onAfter to cancel the caller after a successful true command returned Exit:-1 and queueInterruptedError instead of the completed success. ARCH-PURPOSE and ARCH-ORDER: this is the 2nd finding in family advisory-outcome-isolation. State and enforce the rule across deadline, caller/signal cancellation, and queue-cancellation inputs: events received after confirmed execution completion must not retroactively reclassify its outcome. Preserve authoritative cancellation during execution, and add controlled before/during/after completion tests covering both success and failure.
```

Done-when: current — the existing criteria still cover recorded leak warnings and harness guidance; the revised attribution and identity limits refine that behavior.

1. **Strengths**

   - Numeric-ID comparison avoids treating mutable CTIME as allocation evidence (`report.go:11-26`).
   - Parent and child diagnostics reuse existing digest-verified proof paths (`localcheck_resume_saved.go:117-134`).
   - Native allocation qualification is opt-in and limits cleanup to the independently verified child allocation.
   - README, help, and atlas describe advisory semantics and attribution limits consistently.

2. **Critical findings**

   - New cancellation defect detailed above.
   - BR-2 remains open for the required regression evidence. The production removal is sound; its current test does not establish protection against the former behavior.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Focused SysV and resume tests passed. Temporary overlays confirmed regression sensitivity for BR-1, BR-3, and BR-5, reproduced the cancellation defect, and showed the BR-2 test remains green with the former attribution renderer restored. Native allocation qualification was not rerun. Temporary files were removed; the checkout remains clean.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared rendering and saved-proof reuse.
   - **ARCH-PURE — pass:** parsing and classification remain separate from observation IO.
   - **ARCH-PURPOSE — flag:** post-run diagnostics can still change a completed outcome.
   - **ARCH-MOCK — pass:** private stateful tables exercise production observation seams; native conformance exists.
   - **ARCH-CONSTRAINTS — pass:** bounded snapshots and linear classification; deadline isolation verified.
   - **ARCH-SECURE — pass:** production no longer infers creator identity from current PID evidence.
   - **ARCH-ORDER — flag:** completion versus late cancellation lacks a frozen outcome boundary.
   - **ARCH-FUNERAL — pass:** diagnostics use existing log retention; no new durable sidecar.
   - **ARCH-EVOLVE — pass:** no persisted attempt-schema change.

7. **Plan revision recommendations**

   Append a `## Revisions` entry defining confirmed execution completion as the outcome boundary, enumerating cancellation events on either side, and naming the controlled sequence tests that enforce it.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"84621e07-6376-4472-995d-d4d6da88f542","repository":"github.com/xianxu/ariadne","issue_id":"000589","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/214e815bb791058e","reviewed_head":"e0b3df15d51488c2e22000b67ac0c6fc0dfea4b9","slot":"ariadne:3","requested_at":"2026-10-24T20:30:42.341002Z"},"phase":"finished","pid":25078,"started_at":"2026-10-24T20:30:42.44071Z","finished_at":"2026-10-24T20:33:48.022286Z","session_id":"01a16f9b-caa4-7000-bb87-5052ae3e0cb7","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T14:30:32-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 589 — Recorded suites report leaked SysV shared-memory segments; PostgreSQL test harnesses must not leak them |
| repo | ariadne |
| issue file | workshop/issues/000589-recorded-suites-report-leaked-sysv-shared-memory-segments-postgresql-test.md |
| boundary | whole-issue close |
| milestone | — |
| window | 04052c566441588889d84db71e22caa62b1126b8..6d8c6a42e05bfec33c0f9bd2e0813fe72adeb2e6 |
| command | sdlc close --issue 589 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T14:30:32-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised advisory-warning contract. Both open findings have regression tests that fail when the unsafe behavior is restored. Focused, race, deadline-composition and queue-retirement checks passed. No new blocking findings emerged; the checkout remains unchanged.

Done-when: current — the revised attribution and lifecycle boundaries still deliver recorded segment warnings and harness guidance covered by the existing completion criteria.

```findings
dispose:
  - id: BR-2
    disposition: addressed
    note: |
      Attribution machinery is removed; sysvwarning.go:30-32 always renders unknown command and unattributed relationship. Restoring the former Witness-bearing Candidate and unsafe renderer through a temporary Go overlay makes TestSysVCurrentProcessNameIsNotCreatorProof fail with false witnessed attribution.
  - id: BR-7
    disposition: addressed
    note: |
      localcheck_run.go:514-526 defers observation until outcome-mutating cleanup and retirement finish. Controlled caller/queue before/during/after tests cover success, failure and strict cleanup; signal tests cover default disposition. Restoring the early after-snapshot through an overlay makes the committed regression fail with Exit:-1 and queue interruption after successful execution.
  - id: BR-1
    disposition: addressed
    note: |
      The before-probe precedes execution timer creation, and deadline classification precedes final diagnostics. Controlled SysV budget tests and existing command-deadline composition tests pass.
  - id: BR-3
    disposition: addressed
    note: |
      report.go:11-26 compares numeric IDs conservatively. TestDeltaMetadataChangesAreNotNewIDs covers mutable metadata without creating candidates.
  - id: BR-4
    disposition: addressed
    note: |
      The current Core concepts table describes savedAttemptRecord as existing and reused; the pinned diff leaves localcheck_republish.go unchanged.
  - id: BR-5
    disposition: addressed
    note: |
      resolvedSavedRecord projects the independently verified child warning into statusDescription. Fresh child publication and saved replay regression tests pass.
  - id: BR-6
    disposition: addressed
    note: |
      README.md now documents advisory warnings and WF_SYSV_QUALIFY=1, linking the atlas passage that explains operator-cleared capacity and exact-owned-segment cleanup.
```

1. **Strengths**

   - Conservative ID comparison and explicitly unknown attribution avoid overstating snapshot evidence.
   - Final diagnostics use the original log descriptor and execute after lifecycle cleanup and queue retirement.
   - Fresh and resumed publication share warning projection from verified logs.
   - README, atlas and embedded help explain coverage limits and PostgreSQL shutdown guidance.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed focused SysV tests in both packages, race checks, caller-context checks, existing deadline-composition tests and serialized cancellation/retirement tests. Independently reproduced red controls for BR-2 and BR-7 using temporary overlays, subsequently removed. Native IPC allocation and opt-in private CLI qualification were not rerun; neither was the full repository suite.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared delta, rendering and warning extraction.
   - **ARCH-PURE — pass:** parsing and classification remain separate from injected observation IO; concept-table locations match.
   - **ARCH-PURPOSE — pass:** recorded warnings, replay consumers and harness guidance are delivered.
   - **ARCH-MOCK — pass:** stateful private tables exercise the production observation seam; native qualification exists separately.
   - **ARCH-CONSTRAINTS — pass:** bounded snapshots and linear classification; tests exercise input limits and probe counts.
   - **ARCH-SECURE — pass:** strict table validation, explicit unavailable observations and no production cleanup authority.
   - **ARCH-ORDER — pass:** existing serialized retirement governs cancellation; controlled sequence tests exercise the relevant ordering.
   - **ARCH-FUNERAL — pass:** no new durable sidecar or monitor; diagnostics reuse existing log retention.
   - **ARCH-EVOLVE — pass:** no persisted attempt-schema change; old logs remain readable.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"c7e0338e-82fe-4d45-95e1-277d858464e9","repository":"github.com/xianxu/ariadne","issue_id":"000589","recipe":"milestone-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/f8281ce597bbc622","reviewed_head":"6d8c6a42e05bfec33c0f9bd2e0813fe72adeb2e6","slot":"ariadne:3","requested_at":"2026-10-24T21:28:07.11862Z"},"phase":"finished","pid":81923,"started_at":"2026-10-24T21:28:07.287784Z","finished_at":"2026-10-24T21:30:31.537313Z","session_id":"01a16fd0-5ab7-7203-8524-e42e5eb0347d","outcome":"returned"}] -->
