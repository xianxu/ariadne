# Boundary Review — ariadne#501 (whole-issue close)

| field | value |
|-------|-------|
| issue | 501 — sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain) |
| repo | ariadne |
| issue file | workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md |
| boundary | whole-issue close |
| milestone | — |
| window | cb8ffe354ad44da23b6adae5fbc0211cda4adb84..31cc33fcfefcee50db5dc2118afef03e98e7e7b4 |
| command | sdlc close --issue 501 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T11:19:48-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core recovery path appears sound: it serializes reconciliation, validates exact tombstone bytes and owner evidence, and syncs audit evidence before removal. Relevant tests pass. Before closing, update README and acceptance criteria, and add the durability-error tests claimed by the plan.

Done-when: stale — it still requires reconciling the original live tombstone to achieve complete=true, while the approved revision substitutes private executable conformance and read-only live observation.

```findings
findings:
  - id: new
    severity: Important
    family: user-surface-documentation
    title: |
      README omits the new queue reconcile command
    detail: |
      README.md:596 documents queue pause/resume/status, but this range introduces reconcile and its required flags without updating README. Add the operator/:0 invocation, descendant-attestation requirement, and audit/retry behavior; atlas and embedded help updates do not satisfy the separate README gate.
  - id: new
    severity: Important
    family: acceptance-criteria-consistency
    title: |
      Done-when still requires the superseded live-machine acceptance
    detail: |
      workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md:48 requires live complete=true after reconciling the original tombstone. Its approved revision at line 77 replaces that acceptance after reboot. Update the active Done-when to the approved private-conformance and read-only observation criteria, preserving the original requirement as history.
  - id: new
    severity: Important
    family: durability-failure-coverage
    title: |
      Checked-off append and sync failure controls are missing
    detail: |
      The operative plan's completed crash/retry task claims append/sync error coverage, but reconcile_test.go does not exercise errors from log.Write, log.Sync, or either directory sync in reconcile.go:235-258. The crash hook runs only after successful audit and root sync. Add controllable failure tests proving pre-unlink failures preserve the tombstone and post-unlink sync failure retries without duplicating audit evidence (ARCH-ORDER).
```

1. **Strengths**

   - Exact-byte SHA256 validation and operation/lease exclusion precede retirement.
   - A real subprocess crash tests recovery after durable audit publication, including unchanged audit bytes on retry.
   - Tests verify pause and interruption preservation, malformed evidence refusal, and real kernel owner lifecycle observations.
   - Immutable original bytes preserve opaque future fields; the shared numeric-token correction retains strict duplicate-key checking.

2. **Critical findings:** None identified.

3. **Important findings:** The three findings above.

4. **Minor findings:** None.

5. **Test coverage**

   Passed:

   - `go test ./cmd/sdlc/internal/queuegate ./cmd/sdlc/internal/jsonstrict`
   - `go test ./cmd/sdlc -run 'TestQueue' -count=1`

   These validate the main behavior. They do not establish the claimed append/sync failure coverage. Historical executable-conformance claims were not treated as independently verified evidence.

6. **Architecture**

   - **ARCH-DRY — pass:** reuses gate, locks, authority, process probe, and interruption preservation.
   - **ARCH-PURE — pass:** request/audit validation is separate from filesystem orchestration.
   - **ARCH-PURPOSE — pass:** implements same-boot operator-attested retirement.
   - **ARCH-MOCK — pass:** portable filesystem fixtures and injected owner observations share production boundaries; kernel conformance covers real process behavior.
   - **ARCH-CONSTRAINTS — pass:** bounded record reads and streaming audit scans stay outside admission.
   - **ARCH-SECURE — pass:** malformed, uncertain, live, and unsafe evidence refuses.
   - **ARCH-ORDER — flag:** successful crash/retry ordering is tested; durability-error ordering lacks the promised controls.
   - **ARCH-FUNERAL — pass against approved revision:** permanent audit retention and per-repair storage cost are explicitly documented.
   - **ARCH-EVOLVE — pass:** audit records remain immutable, original bytes survive, and strict known-field validation remains.

7. **Plan revisions**

   Add a `## Revisions` entry aligning active Done-when with approved post-reboot acceptance. Correct the completed durability-test claim until append/sync failure controls exist. The operative revision explicitly withdraws the historical model/bundle concept tables; those withdrawals match the implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"6b321ae9-ec4c-45dc-a499-529d5c459382","repository":"github.com/xianxu/ariadne","issue_id":"000501","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/044530cff625a2b0","reviewed_head":"31cc33fcfefcee50db5dc2118afef03e98e7e7b4","slot":"ariadne:3","requested_at":"2026-10-20T18:18:30.441583Z"},"phase":"finished","pid":49315,"started_at":"2026-10-20T18:18:30.594054Z","finished_at":"2026-10-20T18:19:48.152653Z","session_id":"01a15a89-50a4-7bf2-b941-6422bb5e2e79","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T11:27:51-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 501 — sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain) |
| repo | ariadne |
| issue file | workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md |
| boundary | whole-issue close |
| milestone | — |
| window | cb8ffe354ad44da23b6adae5fbc0211cda4adb84..fe03d734f03caabf9bd30cf238d24e9dcc8f8008 |
| command | sdlc close --issue 501 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T11:27:51-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation matches the operative, approved plan. All three prior findings are addressed, and the relevant tests pass. No new blocking findings were identified.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:605–612 now documents the command, required flags, operator/:0 authority, descendant attestation, permanent audit and retry behavior, consistent with queue_reconcile.go and reconcile.go.
  - id: BR-2
    disposition: addressed
    note: |
      The issue's active Done-when at lines 48–49 now specifies private executable recovery and read-only post-reboot observation. Revisions preserves the original live complete=true criterion as history, consistent with the approved acceptance change.
  - id: BR-3
    disposition: addressed
    note: |
      TestReconcileIOFailures covers nine append and log/root/post-unlink sync failures, checking tombstone preservation, torn-log refusal and audit-preserving retry. All pass. Reverting the production fault-seam wiring through a temporary Go overlay makes all nine fail; the overlay was removed without modifying repository files.
```

1. **Strengths**

   - Exact-byte digest validation, operation-record exclusion and tombstone locking precede removal (`reconcile.go:172–208`).
   - Audit and root-directory sync complete before unlink; post-unlink sync errors propagate (`reconcile.go:242–265`).
   - Real subprocess crash/retry and kernel lifecycle tests complement refusal and preservation controls.
   - Immutable audit bytes preserve future fields; the shared numeric-token fix keeps duplicate-key validation strict.

2. **Critical findings:** None.

3. **Important findings:** None remaining.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/queuegate ./cmd/sdlc/internal/jsonstrict -count=1`
   - `go test ./cmd/sdlc -run 'TestQueue' -count=1`

   The negative overlay check confirmed that all nine new IO controls reach the production wiring. The full repository suite and historical live-machine acceptance were not independently rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** reuses gate, flock, authority, kernel probe and interruption helpers.
   - **ARCH-PURE — pass:** deterministic request/audit validation is separated from filesystem orchestration.
   - **ARCH-PURPOSE — pass:** delivers exact, operator-attested same-boot retirement.
   - **ARCH-MOCK — pass:** injected observations and private filesystem fixtures share production seams; real kernel conformance checks owner lifecycles.
   - **ARCH-CONSTRAINTS — pass:** bounded records and streaming audit reads remain outside admission.
   - **ARCH-SECURE — pass:** malformed, changed, live and uncertain evidence refuses.
   - **ARCH-ORDER — pass:** durable ordering, crash recovery and before/after-effect IO failures are exercised.
   - **ARCH-FUNERAL — pass against the approved revision:** permanent audit lifetime and per-repair growth are explicitly documented.
   - **ARCH-EVOLVE — pass:** existing formats remain unchanged; immutable audit/original bytes preserve unknown members.

7. **Plan revision recommendations:** None. The operative revisions explicitly withdraw the historical model/bundle concept tables and match the delivered implementation.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ff52f87a-1f6b-45af-8c29-ead031b83408","repository":"github.com/xianxu/ariadne","issue_id":"000501","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/9ecd27334cbf5ccb","reviewed_head":"fe03d734f03caabf9bd30cf238d24e9dcc8f8008","slot":"ariadne:3","requested_at":"2026-10-20T18:26:09.264032Z"},"phase":"finished","pid":94922,"started_at":"2026-10-20T18:26:09.361557Z","finished_at":"2026-10-20T18:27:50.029833Z","session_id":"01a15a90-5063-7b11-be1a-62e724797d81","outcome":"returned"}] -->

---

## Re-review — 2026-10-20T14:36:22-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 501 — sdlc queue reconcile: an operator-attested verb to retire an unresolved unmanaged tombstone (blocks the live drain) |
| repo | ariadne |
| issue file | workshop/issues/000501-sdlc-queue-reconcile-an-operator-attested-verb-to-retire-an-unresolved-unmanaged-tombstone-blocks-the-live-drain.md |
| boundary | whole-issue close |
| milestone | — |
| window | 70ed52cedca4732bd1783e77383b92a5dd8fc44c..904e2a9ffb758a43ee764c12b8dad823a8c351b6 |
| command | sdlc close --issue 501 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-20T14:36:22-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff introduces no blocking defects. The conflict resolution preserves both branches’ intent: main’s shared scanner and depth controls remain, and #501’s opaque-number compatibility survives. The issue revision records integration work without changing scope. Previously disposed findings remain satisfied.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      README.md:605–612 retains the command, authority, attestation, audit and retry documentation.
  - id: BR-2
    disposition: addressed
    note: |
      Active Done-when retains approved private recovery and read-only live acceptance; the superseded criterion remains in Revisions. This window adds no acceptance change.
  - id: BR-3
    disposition: addressed
    note: |
      The previously verified IO failure controls and production wiring are unchanged; all nine controls pass within the queuegate package run.
```

1. **Strengths**

   - `jsonstrict.go:36–37` preserves `UseNumber` and delegates to the shared `Scan` implementation.
   - The resolved file matches landed main exactly, retaining depth limits, duplicate-key checks and stream handling.
   - Both scanner and reconciliation regressions remain present and pass.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed:

   - `go test ./cmd/sdlc/internal/jsonstrict ./cmd/sdlc/internal/queuegate -count=1`
   - `go test ./cmd/sdlc -run 'TestQueue' -count=1`
   - Pinned-range `git diff --check`

   The full suite was not rerun. Repository files remained unchanged.

6. **Architectural notes**

   - **ARCH-DRY — pass:** resolution retains the shared scanner.
   - **ARCH-PURE — pass:** JSON validation remains deterministic, without IO.
   - **ARCH-PURPOSE — pass:** both conflicting changes’ required behavior survives.
   - **ARCH-EVOLVE — pass:** opaque future numbers remain accepted; strict validation remains intact. No persisted format or rewrite behavior changes.

   This window adds no user-facing or architectural surface requiring further README or atlas updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"ce8b2607-1759-40c8-aa90-4b714c61d103","repository":"github.com/xianxu/ariadne","issue_id":"000501","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/f166f79965ff8072","reviewed_head":"904e2a9ffb758a43ee764c12b8dad823a8c351b6","slot":"ariadne:3","requested_at":"2026-10-20T21:35:25.496379Z"},"phase":"finished","pid":15015,"started_at":"2026-10-20T21:35:25.606211Z","finished_at":"2026-10-20T21:36:22.323527Z","session_id":"01a15b3d-98f5-7a51-b8d7-e3baab00f67d","outcome":"returned"}] -->
