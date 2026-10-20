# Boundary Review — ariadne#481 (milestone M1)

| field | value |
|-------|-------|
| issue | 481 — sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain |
| repo | ariadne |
| issue file | workshop/issues/000481-sdlc-queue-pause-resume-cancel-running-and-the-tl-worker-drain-rule-for-couch-drain.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | 8a44224ff0f867837283368f62c0344bcd322b25..7bedb19c29d5c52404bd2922a9d8df1e9e454102 |
| command | sdlc milestone-close --issue 481 --milestone M1 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T01:58:18-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

M1 provides a useful backend boundary: owner composition, serialized admission, strict codecs, and conservative legacy observations are implemented, and targeted race tests pass. Cancellation recovery and cleanup remain incomplete, the declared command deadline does not cover transaction work, and the Core concepts table contradicts the implemented API. M2’s admission wiring and peer conformance are explicitly deferred and are not treated as missing M1 deliverables.

```findings
findings:
  - id: new
    severity: Critical
    family: durable-cancellation-reconciliation
    title: |
      Durable cancellation intent is not reconciled after partial publication or owner death
    detail: |
      cmd/sdlc/internal/queuegate/store.go:244 writes cancel/<id>.json before rewriting the lease, but lease.go:156-168 checks only the lease phase. If the pause process dies between those writes, CancellationRequested returns false and Observe never consumes the persisted intent. Separately, Observe at store.go:333-343 converts a dead interrupting lease into unresolved evidence without retaining the promised interruption record or removing its cancellation intent; prior-boot reconciliation later removes only the unresolved record. This loses interruption diagnostics and leaves permanent cancellation files that can eventually exceed operationFiles' reader limit. ARCH-ORDER/ARCH-FUNERAL: reconcile exact-ID intents across partial publication and stale-lease recovery, preserve interruption evidence alongside descendant uncertainty, and clean intents after their last reader. Add deterministic crash-window and repeated recovery/boot tests.
  - id: new
    severity: Critical
    family: core-concept-traceability
    title: |
      The Core concepts table names entities and transition behavior absent from the implementation
    detail: |
      workshop/plans/000481-queue-drain-plan.md:31-35 lists PauseSet, OperationPhase and Interruption in model.go and says Transition returns requested interruption IDs with owner state protected from direct mutation. The implementation instead exposes Control, string phases and Operation-based interruption records; Transition returns only (Control, error). LegacyObserver at plan line 46 is also implemented as observeLegacyQueue rather than the named entity. Under the required Core concepts cross-check, these are plan/code contradictions. ARCH-PURPOSE/ARCH-ORDER: append a Revisions entry reconciling the table and structural-enforcement claims with the actual API, or implement the promised model.
  - id: new
    severity: Important
    family: transaction-deadline-enforcement
    title: |
      Queue command deadlines stop applying once the control mutex is acquired
    detail: |
      cmd/sdlc/internal/queuegate/store.go:67-72 checks context only while acquiring the mutex, then invokes a callback without deadline checks. requestInterruptions at lines 209-258 can perform durable intent writes and lease rewrites for 4096 operations while holding that mutex; Observe likewise scans and reconciles records without checking context inside its transaction. Consequently queuecmd.go:140's three-second timeout does not enforce the plan's under-four-second target or the peer's five-second ceiling. ARCH-CONSTRAINTS: propagate cancellation through long transaction loops, retain recoverable partial progress, and measure representative upper-bound workloads.
  - id: new
    severity: Minor
    family: shared-json-boundary-validation
    title: |
      Strict JSON validation is duplicated across the new boundary readers
    detail: |
      cmd/sdlc/queue_legacy.go:285-350 repeats the recursive duplicate-key, depth, UTF-8 and trailing-data validation in internal/queuegate/codec.go:17-85. ARCH-DRY: extract a shared strict JSON inspection helper while retaining each reader's schema-specific decoding.
```

1. **Strengths**

   - Owner transitions preserve unrelated holds, original reasons, and cancellation upgrades; tests exercise replay and matching-owner resume.
   - Real-flock tests force admission/pause ordering and verify that waiting operations remain excluded from activity.
   - Codecs reject malformed known fields and nested duplicate keys while preserving opaque additive values.
   - README, atlas, and embedded help were updated in this range. Legacy observations cannot authorize signaling.

2. **Critical findings**

   - Cancellation publication and recovery need one coherent lifecycle spanning durable intent, lease phase, interruption evidence, unresolved descendants, and cleanup.
   - The Core concepts table needs reconciliation before it can serve as an accurate contract for M2 consumers.

   Locations and fix sketches are in the findings block above.

3. **Important findings**

   - Enforce deadlines during registry transactions, with recoverable partial effects and representative workload measurements.

4. **Minor findings**

   - Consolidate the duplicated JSON validation routine.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/queuegate -race -count=1`
   - `go test ./cmd/sdlc -run '^TestQueue' -race -count=1`
   - `go test ./cmd/sdlc/helptext -count=1`
   - `go test ./pkg/golint -count=1`
   - Pinned-range `git diff --check`

   Existing tests cover successful cancellation publication and ordinary dead-lease uncertainty. They do not cover death between intent publication and phase rewrite, interrupted-owner recovery with evidence retention, cancellation-residue cleanup, or transaction deadline expiry after lock acquisition. No prior findings required disposition. No repository files were changed.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — flag:** duplicated JSON validator.
   - **ARCH-PURE — pass:** model and codec tests operate without IO; storage is separate.
   - **ARCH-PURPOSE — flag:** reconcile the declared core model; M2 deferrals otherwise match the milestone split.
   - **ARCH-MOCK — pass for M1:** portable stores, real locks, and injected observation seams support isolated verification. Actual peer executable conformance remains M2.
   - **ARCH-CONSTRAINTS — flag:** transaction deadlines and representative maximum-workload evidence are missing.
   - **ARCH-SECURE — pass:** strict persisted-input checks and conservative unknown states prevent malformed evidence from proving quiescence.
   - **ARCH-ORDER — flag:** partial cancellation publication lacks recovery through the production state model.
   - **ARCH-FUNERAL — flag:** dead-owner cancellation intents lack a complete removal path.
   - **ARCH-EVOLVE — pass:** inspected control and operation rewrites preserve unknown members; codec tests exercise additive round trips.

7. **Plan revision recommendations**

   Append `## Revisions` entries that:

   - Map the Core concepts table to actual symbols and accurately describe transition outputs and enforcement.
   - Specify recovery for each cancellation publication crash window, including interruption evidence and intent cleanup.
   - State how transaction deadlines bound work after mutex acquisition and how that envelope is measured.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"5b887303-5339-4b9b-83e7-6ca77bfa298f","repository":"github.com/xianxu/ariadne","issue_id":"000481","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c96aaf0e0e803e05","reviewed_head":"7bedb19c29d5c52404bd2922a9d8df1e9e454102","slot":"ariadne:3","requested_at":"2026-10-20T08:55:51.943938Z"},"phase":"finished","pid":22460,"started_at":"2026-10-20T08:55:52.085778Z","finished_at":"2026-10-20T08:58:18.363307Z","session_id":"01a15886-335e-7d91-a2b3-e17641373b01","outcome":"returned"}] -->
