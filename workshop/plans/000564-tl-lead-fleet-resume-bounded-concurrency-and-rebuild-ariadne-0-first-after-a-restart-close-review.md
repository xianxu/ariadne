# Boundary Review — ariadne#564 (whole-issue close)

| field | value |
|-------|-------|
| issue | 564 — tl-lead fleet resume: bounded concurrency, and rebuild ariadne:0 first after a restart |
| repo | ariadne |
| issue file | workshop/issues/000564-tl-lead-fleet-resume-bounded-concurrency-and-rebuild-ariadne-0-first-after-a-restart.md |
| boundary | whole-issue close |
| milestone | — |
| window | 260ea7dca61002c5b4624fe4b2b921d12be47b8f..efe4f364a43c37ce0a5b4b0da7844d7f78873153 |
| command | sdlc close --issue 564 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T18:43:26-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The bounded scheduler, wave barrier, receipt classification, cancellation joining, and documentation are well supported by passing tests. One blocking defect remains: unavailable agent evidence without an asleep marker becomes a normal skip, allowing an incorrect completion handoff. I reproduced this through a private build of the pinned head. The repository remains unchanged.

Done-when: current — it covers bounded concurrency, explicit asleep authority, ordered worker waves, rebuild-first guidance, and the steward-owned scheduling handoff.

```findings
findings:
  - id: new
    severity: Critical
    kind: defect
    family: recovery-evidence-integrity
    title: |
      Unavailable agent evidence can produce a successful worker-wave handoff
    detail: |
      cmd/tl-lead/fleet_recovery.go:110 returns on an absent asleep marker before validating agent evidence quality; recoveryWireAgent also discards error diagnostics. With an owned worker containing agent: {"state":"unavailable","error":"thread store unreadable"}, the pinned CLI prints "skipped", exits 0, and reports worker waves complete. This contradicts the plan's requirement that unavailable evidence remain unknown and block completion (ARCH-SECURE, ARCH-ORDER, ARCH-PURPOSE). Validate agent evidence before treating marker absence as an ordinary skip, retain relevant diagnostics, and add decoder and CLI regressions requiring a named unknown row and nonzero incomplete handoff.
```

1. **Strengths**

   - `fleet_batch.go:98` centralizes admission, result, and cancellation transitions, including duplicate-result rejection and the wave barrier.
   - `fleet_batch_run.go:76` explicitly joins admitted goroutines before returning.
   - Receipt handling preserves stable failure identity and refuses success after capture overflow or local pipe failure.
   - README, atlas, and rendered help consistently describe bounded worker waves and steward-only scheduling authority.

2. **Critical findings**

   - **Unavailable agent evidence becomes a successful skip**, at `cmd/tl-lead/fleet_recovery.go:110`. The private CLI reproduction returned exit **0** and “worker waves 1 and 2 complete” for an unreadable owned worker. Validate evidence quality before the absent-marker return. Sweep missing, malformed, unavailable, and conflicting agent evidence, while retaining valid older-producer marker omission as a skip.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Passed:

   - Full `cmd/tl-lead` and shared JSON validator tests.
   - Focused race tests for batch transitions, coordination, recovery decoding, and receipt classification.
   - All six private executable conformance cases, including SIGTERM and lost reply.
   - Isolated real-Couch receipt conformance for succeeded, refused, and failed outcomes.
   - Range whitespace checks.

   Existing fixtures omit unavailable agent evidence without a marker. Add a regression that fails on this pinned head. Native fleet wake was not exercised; private executable tests do not establish native agent qualification.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** shared strict JSON implementation and one recovery snapshot.
   - **ARCH-PURE — pass:** deterministic decoding, planning, and batch transitions remain separate from effects.
   - **ARCH-PURPOSE — flag:** the completion handoff bypasses unknown-evidence blocking in the finding above.
   - **ARCH-MOCK — pass:** production seams have stateful doubles and executable/receipt conformance.
   - **ARCH-CONSTRAINTS — pass:** bounded acquisition, capture, and in-flight calls; controlled tests cover capacity.
   - **ARCH-SECURE — flag:** unavailable external evidence loses its diagnostic and becomes an ordinary skip.
   - **ARCH-ORDER — flag:** that projection permits a completion state without sufficient evidence; scheduler transition ordering otherwise checks out.
   - **ARCH-FUNERAL — pass:** no new production durable queue or ledger; temporary fixture artifacts have cleanup.
   - **ARCH-EVOLVE — pass:** the additive report reader performs no rewrite, and validator extraction preserves existing behavior.

   Core-concept locations and classifications match the implementation. The decoder defect contradicts the documented evidence contract.

7. **Plan revision recommendations**

   Append a `## Revisions` entry recording the evidence-quality correction and its decoder-to-handoff regression coverage. Preserve the existing requirement that unavailable evidence blocks completion.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"32edff20-ecb2-413f-a8bf-33fb26cd501d","repository":"github.com/xianxu/ariadne","issue_id":"000564","recipe":"milestone-review","round":1,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/5ea2f2f61f5e2a46","reviewed_head":"efe4f364a43c37ce0a5b4b0da7844d7f78873153","slot":"ariadne:3","requested_at":"2026-10-25T01:40:39.238613Z"},"phase":"finished","pid":7538,"started_at":"2026-10-25T01:40:39.357156Z","finished_at":"2026-10-25T01:43:25.445864Z","session_id":"01a170b7-8e14-7f81-9a8a-584169f5ffd0","outcome":"returned"}] -->
