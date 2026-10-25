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

---

## Re-review — 2026-10-24T18:53:54-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 564 — tl-lead fleet resume: bounded concurrency, and rebuild ariadne:0 first after a restart |
| repo | ariadne |
| issue file | workshop/issues/000564-tl-lead-fleet-resume-bounded-concurrency-and-rebuild-ariadne-0-first-after-a-restart.md |
| boundary | whole-issue close |
| milestone | — |
| window | 260ea7dca61002c5b4624fe4b2b921d12be47b8f..77425d4f8952fa41f8b20d4ae304eecf63608a7e |
| command | sdlc close --issue 564 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T18:53:54-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The pinned implementation satisfies the revised recovery contract. BR-1 is addressed with reachable validation and regressions that fail when the fix is removed. No new blocking findings emerged. The checkout remains unchanged.

Done-when: current — covers bounded concurrency, explicit asleep authorization, ordered worker waves, rebuild-first guidance, and the steward-owned scheduling handoff.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      fleet_recovery.go:101 validates agent evidence before marker absence and retains diagnostics. Decoder and CLI regressions pass at HEAD; removing the guard in a disposable checkout makes both fail and reproduces the false successful handoff. The private executable unavailable-evidence case also passes. Disposable artifacts were removed.
```

1. **Strengths**

   - Agent validation now distinguishes valid legacy marker omission from unavailable or conflicting evidence (`cmd/tl-lead/fleet_recovery.go:100`).
   - Explicit batch transitions enforce concurrency, wave barriers, and failure stopping; the effect shell joins admitted calls before returning (`cmd/tl-lead/fleet_batch_run.go:76`).
   - The shared JSON validator is byte-identical to the original implementation, with the SDLC import path preserved.
   - Help, README, and atlas consistently document rebuild-first recovery and steward-only wave three (`cmd/tl-lead/help.md:61`).

2. **Critical findings**

   None.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Independently passed:

   - Full `tl-lead` package and shared JSON tests.
   - Focused recovery, coordinator, batch, and receipt tests under `-race`.
   - All seven private executable conformance cases, including SIGTERM, lost replies, repeated failures, and unavailable evidence.
   - Real Couch CLI receipt conformance for succeeded, refused, and failed outcomes.

   BR-1’s decoder and CLI tests failed after guard removal in a disposable checkout. The private binary matched pinned HEAD with `vcs.modified=false`.

   Native fleet wake and the full repository suite were not run in this review; native qualification remains the explicitly documented limitation.

6. **Architectural notes for upcoming work**

   - **ARCH-DRY — pass:** one recovery projection and shared strict JSON authority.
   - **ARCH-PURE — pass:** projection, planning, receipt classification, and batch transitions separate from IO; core-concept classifications match their implementations.
   - **ARCH-PURPOSE — pass:** delivers both worker waves and the scoped completion handoff.
   - **ARCH-MOCK — pass:** stateful injected and executable fakes share production boundaries; real CLI conformance checks receipt behavior.
   - **ARCH-CONSTRAINTS — pass:** positive concurrency validation, bounded captures and subprocess budgets, and counted 8/16-worker tests.
   - **ARCH-SECURE — pass:** unavailable evidence fails visibly; command arguments are structural; conformance uses private storage.
   - **ARCH-ORDER — pass:** explicit transitions, controllable completion ordering, duplicate-result handling, cancellation, and joined calls.
   - **ARCH-FUNERAL — pass:** no new durable runtime artifacts; temporary fixtures have cleanup.
   - **ARCH-EVOLVE — pass:** additive report fields are tolerated without rewriting; duplicate-key validation remains shared and strict.

7. **Plan revision recommendations**

   None. The existing BR-1 revision accurately describes the correction and its regression coverage.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"7e28675a-9c6c-49a0-9da5-e8fc86a14168","repository":"github.com/xianxu/ariadne","issue_id":"000564","recipe":"milestone-review","round":2,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/c5241420f2a0ddc3","reviewed_head":"77425d4f8952fa41f8b20d4ae304eecf63608a7e","slot":"ariadne:3","requested_at":"2026-10-25T01:51:19.843063Z"},"phase":"finished","pid":69727,"started_at":"2026-10-25T01:51:19.954877Z","finished_at":"2026-10-25T01:53:54.021845Z","session_id":"01a170c1-5442-7122-94dc-98a60b800fa5","outcome":"returned"}] -->

---

## Re-review — 2026-10-24T19:15:14-07:00 (SHIP)

| field | value |
|-------|-------|
| issue | 564 — tl-lead fleet resume: bounded concurrency, and rebuild ariadne:0 first after a restart |
| repo | ariadne |
| issue file | workshop/issues/000564-tl-lead-fleet-resume-bounded-concurrency-and-rebuild-ariadne-0-first-after-a-restart.md |
| boundary | whole-issue close |
| milestone | — |
| window | e7a14009840ebed6f6319bd53dafadde368c9390..d6fdeba76cdfb6a3e24d3ed142e8067a723f1ffb |
| command | sdlc close --issue 564 |
| reviewer | codex (openai; author openai) |
| isolation | fenced |
| timestamp | 2026-10-24T19:15:14-07:00 |
| verdict | SHIP |

## Review

```verdict
verdict: SHIP
confidence: high
```

The interdiff correctly repairs command-discovery expectations after the shared `cmd/internal` extraction. No new runtime behavior, scope expansion, or blocking defect was found. The checkout remains unchanged.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      The prior correction remains intact: agent evidence is validated before marker absence. Decoder and CLI handoff regressions pass at pinned HEAD; this interdiff does not alter that behavior.
```

1. **Strengths**

   - All three corrected selectors exclude namespaces while retaining command entrypoints.
   - Owner-build expectations preserve the existing `weave` exclusion.
   - Sandbox expectations retain explicit error handling (`cmd/weave/internal/settingsx/base_settings_test.go:74`).
   - The plan revision records the CI correction without changing the recovery contract.

2. **Critical findings:** None.

3. **Important findings:** None.

4. **Minor findings:** None.

5. **Test coverage notes**

   Passed the portable Makefile fixture, targeted sandbox-policy test, both BR-1 regressions, and slot-dependency shell syntax check. Running the base versions of the portable fixture and sandbox test against HEAD reproduced their failures. Both Python selectors also passed independent command, namespace, library, and gateway controls.

   The full slot-dependency acceptance matrix and repository suite were not rerun.

6. **Architectural notes**

   - **ARCH-DRY — pass:** independent fixtures consistently apply the existing command-entrypoint convention.
   - **ARCH-PURE — pass:** changes remain within integration-test discovery.
   - **ARCH-PURPOSE — pass:** the correction covers all three affected expectations found in the sweep.
   - **ARCH-EVOLVE — pass:** no persisted-format semantics change.

   No new architectural or user-facing surface requires atlas or README updates.

7. **Plan revision recommendations:** None.

<!-- sdlc-review-usage [{"version":1,"intent":{"disposable_root":true,"id":"107a43f2-c56b-48f0-a07a-c286777ee92b","repository":"github.com/xianxu/ariadne","issue_id":"000564","recipe":"fix-review","round":3,"attempt":1,"agent":"codex","root":"/Volumes/TestScratch/review-checkouts/da52f2cc9ab072ad","reviewed_head":"d6fdeba76cdfb6a3e24d3ed142e8067a723f1ffb","slot":"ariadne:3","requested_at":"2026-10-25T02:14:12.241027Z"},"phase":"finished","pid":55957,"started_at":"2026-10-25T02:14:12.384068Z","finished_at":"2026-10-25T02:15:12.635271Z","session_id":"01a170d6-4432-7892-8547-1fc3e53cf60a","outcome":"returned"}] -->
