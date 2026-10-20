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
