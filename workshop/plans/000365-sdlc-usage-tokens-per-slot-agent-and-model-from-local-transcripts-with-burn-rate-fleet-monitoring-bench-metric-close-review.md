# Boundary Review — ariadne#365 (whole-issue close)

| field | value |
|-------|-------|
| issue | 365 — sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric) |
| repo | ariadne |
| issue file | workshop/issues/000365-sdlc-usage-tokens-per-slot-agent-and-model-from-local-transcripts-with-burn-rate-fleet-monitoring-bench-metric.md |
| boundary | whole-issue close |
| milestone | — |
| window | 18915f6c17d9b4aeb5e618d3f736705ec355d2b3..2723e1f5892053ae932f5c86bf14818454ae62ad |
| command | sdlc close --issue 365 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T21:11:20-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The implementation has solid normalization, bounded storage, and documentation, but two reproduced defects allow incomplete evidence to appear complete. The plan also contains contradictory Core concepts entries and claims an integration regression that is not present. The checkout remains unchanged.

### 1. Strengths

- Deduplication precedes window filtering and attribution; tests cover revisions, conflicting finals, compaction copies, and overflow.
- Reviewer receipts use explicit Requested → Started → Finished transitions, CAS persistence, and detached cancellation cleanup.
- Meter writes are bounded, private, and atomic.
- README and atlas updates cover the new command, counter semantics, calibration, and coverage limitations.

### 2. Critical findings

```findings
findings:
  - id: new
    severity: Critical
    family: cumulative-coverage-reconciliation
    title: |
      Modern cumulative evidence hides later uncovered Codex spend
    detail: |
      cmd/sdlc/internal/usage/codex.go:218 uses the legacy cumulative baseline only when no modern thread_token_usage exists. A production-decoder probe containing a modern response and thread total of 100 input/20 output, followed by a legacy cumulative snapshot of 500/100, returned only 120 tokens with spend_complete=true and no uncovered_segments diagnostic. This can authorize calibration using an incomplete denominator. Reconcile all applicable cumulative evidence without adding it to response totals, and add this mixed-format regression. ARCH-PURPOSE, ARCH-SECURE.
  - id: new
    severity: Critical
    family: invalid-accounting-evidence-must-remain-visible
    title: |
      Invalid reviewer bindings disappear without incomplete coverage
    detail: |
      cmd/sdlc/usage_reviewers.go:95 ignores ledger parse failures, while :136 silently discards records failing reviewusage.Validate. A production-loader probe with a version-99 reviewer record returned empty coverage, attribution_complete=true, and worker attribution for the affected owned session. Validate retained records at ingestion and report failures; usable identity evidence must not silently become ordinary worker evidence. Sweep ledger and comment readers, preferably sharing the duplicated decoding in reviewer_accounting.go:136. ARCH-SECURE, ARCH-DRY.
  - id: new
    severity: Critical
    family: core-concept-table-must-match-code
    title: |
      Core concepts tables contradict the delivered paths and statuses
    detail: |
      workshop/plans/000365-usage-plan.md:30 lists provenance.go as modified, but it is unchanged in the pinned range. Line 31 names nonexistent internal/gatestate/reviewer.go, and line 44 names nonexistent usage_render.go. Binding validation actually lives in internal/reviewusage/model.go and rendering in usage.go. Under the requested Core concepts gate, these contradictions are blocking. Append a Revisions entry and reconcile the tables with the delivered implementation and PURE/INTEGRATION ownership. ARCH-PURE.
  - id: new
    severity: Important
    family: production-path-regression-evidence
    title: |
      The promised gate-dispatch-to-usage regression is missing
    detail: |
      workshop/plans/000365-usage-plan.md:139 marks production boundary/plan/estimate dispatch-through-usage coverage complete. The judge tests supply UsageIntent and UsageRecorder directly; root accounting tests exercise cleanup and serialization, and usage_attribution_test.go:159 fabricates bindings and worker windows. None verifies the production gate attachment and persistence through usage loading, including interruption without semantic gate advancement. Add the promised integration regressions and make omission of production accounting wiring fail them. ARCH-PURPOSE.
```

### 3. Important findings

The missing production integration regression is detailed above. No README or atlas omission was found.

### 4. Minor findings

None.

### 5. Test coverage notes

Passed:

- Complete `usage`, `reviewusage`, `transcripts`, and `judge` package tests.
- Focused root usage, attribution, authority, and accounting tests.
- Pinned-range whitespace checks.

Both temporary regression probes failed on the defects described above. They used external Go overlays, which were removed afterward. The full repository suite and opt-in live CLI checks were not rerun.

### 6. Architectural notes

- **ARCH-DRY — flag:** Reviewer artifact decoding is duplicated; consolidate it with consistent validation and diagnostics.
- **ARCH-PURE — pass implementation; flag plan:** Pure counting/calibration and IO adapters are separated, but the concept map needs correction.
- **ARCH-PURPOSE — flag:** False completeness undermines trustworthy burn projections and issue benchmarking; production integration evidence is missing.
- **ARCH-MOCK — pass:** Stateful fake reviewer tests share the production process seam; opt-in live conformance exists.
- **ARCH-CONSTRAINTS — pass:** Explicit candidate, line, response, deadline, and storage bounds have focused coverage.
- **ARCH-SECURE — flag:** Invalid persisted accounting and uncovered counter evidence can disappear without visible degradation.
- **ARCH-ORDER — pass:** Attempt transitions preserve uncertain progress and reject rewritten confirmations.
- **ARCH-FUNERAL — pass:** Meter replacement, receipt capacity, and exact committed-copy cleanup define artifact lifecycles.

### 7. Plan revision recommendations

Append `## Revisions` entries that:

- Reconcile every Core concepts path, status, and classification with the delivered code.
- Require reconciliation of all applicable cumulative evidence before declaring spend complete.
- Require visible diagnostics for invalid reviewer evidence.
- Correct Task2’s integration-test completion claim until the production regressions exist.

---

## Re-review — 2026-10-14T22:55:38-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 365 — sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric) |
| repo | ariadne |
| issue file | workshop/issues/000365-sdlc-usage-tokens-per-slot-agent-and-model-from-local-transcripts-with-burn-rate-fleet-monitoring-bench-metric.md |
| boundary | whole-issue close |
| milestone | — |
| window | 8699bc8dbdec809db55ddbd242eb3ecffb990a6d..ead74a59d03dda4a26a91c2515ca093df96f388d |
| command | sdlc close --issue 365 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-14T22:55:38-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

All four prior findings are addressed, with regression evidence for the behavior changes. Two additional accounting defects block shipping: discovery and decoding disagree on Codex session identity, and issue filtering can turn another issue’s known reviewer into worker spend. The focused tests pass; both new defects reproduce against the pinned head.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Mixed legacy/thread and per-turn reconciliation tests pass. Overlays restoring each pre-fix decoder make the corresponding regressions fail.
  - id: BR-2
    disposition: addressed
    note: |
      Ledger, comment and private-store invalid-evidence tests pass. Removing invalid-identity quarantine makes all three worker-exclusion regressions fail.
  - id: BR-3
    disposition: addressed
    note: |
      Inspected the before/after plan diff: Core concepts now names reviewusage/model.go and usage.go, and labels provenance.go and tracker/engagement.go reused unchanged. Filesystem and pinned diff confirm those paths and statuses; the Revisions entry records the correction.
  - id: BR-4
    disposition: addressed
    note: |
      Production boundary, plan and estimate dispatch-through-loading tests pass, including cancellation without semantic advancement. Independently omitting their accounting attachments makes each intended regression fail.
findings:
  - id: new
    severity: Critical
    family: session-identity-contract-must-be-shared
    title: |
      Codex decoding ignores an executing identity accepted by discovery
    detail: |
      cmd/sdlc/internal/usage/codex.go:35-42 reads only session_meta.payload.id, while internal/transcripts/usage_discovery.go:332-335 accepts thread_id as its fallback. A production discovery→Decode probe verified a child identified by thread_id, then returned both its 12 tokens and 120 copied parent tokens with spend_complete=true. Share the immutable executing-identity contract across discovery and decoding, and cover both identity spellings plus later copied metadata. ARCH-DRY, ARCH-SECURE.
  - id: new
    severity: Critical
    family: role-exclusion-requires-unfiltered-evidence
    title: |
      Issue filtering removes reviewer evidence before exclusive role selection
    detail: |
      cmd/sdlc/usage_reviewers.go:66-67 skips other issues’ accounting artifacts, and usage.go:249 supplies the selected issue as that filter. Once a private receipt is collected into its committed artifact, a known in-place reviewer for another issue can therefore become the selected issue’s worker during overlapping ownership intervals. A production-loader probe with a valid issue-366 plan-review ledger returned zero records, complete coverage and worker Include=true for issue 365. Load relevant reviewer identity evidence before issue filtering; filter included reviewer totals only after exclusive role selection. Cover ledger/sidecar persistence and receipt-cleanup parity. ARCH-PURPOSE, ARCH-SECURE.
```

1. **Strengths**

   - Cumulative snapshots remain diagnostic evidence rather than additional spend.
   - Invalid accounting retains usable identity for quarantine through a shared artifact decoder.
   - Production attachment regressions detect missing boundary, plan and estimate wiring.
   - README and atlas updates cover the command, calibration and receipt lifecycle.

2. **Critical findings**

   The two findings above require fixes and regression tests before crossing the boundary.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Usage, reviewusage, transcripts and judge packages passed, as did targeted root usage/accounting tests. Mutation checks confirmed the prior fixes are exercised. Both new probes failed against unchanged production code. Temporary overlays were removed; the checkout remains clean. I did not rerun the full suite or live provider conformance.

6. **Architectural notes**

   - **ARCH-DRY — flag:** session identity parsing differs between consumers.
   - **ARCH-PURE — pass:** normalization, aggregation and transitions have direct pure tests.
   - **ARCH-PURPOSE — flag:** premature issue filtering breaks exclusive reviewer attribution.
   - **ARCH-MOCK — pass:** portable repositories and stateful fake reviewer processes exercise shared boundaries.
   - **ARCH-CONSTRAINTS — pass:** explicit scan, line, response and deadline bounds are implemented.
   - **ARCH-SECURE — flag:** the new defects admit incorrect attribution despite available identity evidence.
   - **ARCH-ORDER — pass:** receipt transitions and cancellation preserve confirmed progress separately from gate acceptance.
   - **ARCH-FUNERAL — pass:** private receipts have bounded retention and committed-evidence cleanup; permanent metadata has a documented per-attempt cost.

7. **Plan revision recommendations**

   Append `## Revisions` entries establishing one immutable Codex identity contract across consumers and requiring reviewer exclusion evidence to survive issue filtering and receipt cleanup. Add both regressions to Task2’s verification scope.

---

## Re-review — 2026-10-15T00:05:25-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 365 — sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric) |
| repo | ariadne |
| issue file | workshop/issues/000365-sdlc-usage-tokens-per-slot-agent-and-model-from-local-transcripts-with-burn-rate-fleet-monitoring-bench-metric.md |
| boundary | whole-issue close |
| milestone | — |
| window | 1643ba28edb46a03318f90c89c225d1a33ac13a0..e7c54fe3ccb287e9a5a96464fd4055d426e0e34b |
| command | sdlc close --issue 365 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T00:05:25-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-6 is fixed, and its cleanup-parity regression detects the original failure. BR-5’s alias handling is fixed, but executing-identity enforcement remains incomplete: responses preceding the first metadata header bypass filtering, and headerless files can report complete spend. A production discovery→Decode probe reproduced 132 tokens instead of the child’s 12. This keeps BR-5 blocking.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Cumulative reconciliation remains implemented; thread/turn high-water and reset regressions pass in the usage package.
  - id: BR-2
    disposition: addressed
    note: |
      Invalid accounting retains quarantine identities and explicit coverage; malformed-artifact, invalid-binding and cleanup-protection regressions pass.
  - id: BR-3
    disposition: addressed
    note: |
      The plan’s Core concepts tables identify delivered paths, statuses and pure versus integration ownership, including the shared Codex parser and mixed adapter files; inspected definitions match.
  - id: BR-4
    disposition: addressed
    note: |
      Production boundary, planning and cancellation tests enter real dispatch wiring through fake CLI processes, persist accounting and verify reviewer attribution; these tests pass.
  - id: BR-5
    disposition: not-addressed
    note: |
      cmd/sdlc/internal/usage/codex.go:145 filters only when s.thread is already known. A foreground overlay regression using production DiscoverUsageEvidence→Decode placed a 120-token parent response before the first child session_meta, followed by a 12-token child response: discovery verified child, but Decode returned 132 tokens with spend_complete=true and no diagnostics. Omitting the header also returned 132 with spend_complete=true. The alias regression correctly fails under the old-decoder mutation, but does not cover these identity-ordering cases. Complete the existing session-identity-contract-must-be-shared family rule across every response position and absent identity; filter pending records after identity resolution or visibly reject uncertain ordering. ARCH-SECURE, ARCH-ORDER.
  - id: BR-6
    disposition: addressed
    note: |
      loadUsageReviewers retains cross-issue/repository identities before issue inclusion. Ledger/comment tests pass through actual committed receipt cleanup; restoring the early issue filter makes both fail by charging 20 reviewer tokens as the selected issue’s worker tokens.
```

1. **Strengths**

   - Reviewer exclusion survives private-receipt cleanup into both ledger and comment artifacts.
   - Accounting uses validated Requested→Started→Finished transitions and bounded CAS persistence.
   - README, atlas and command help document the new surface and its coverage limits.
   - Regression tests exercise production dispatch and persistence paths with isolated fake CLIs.

2. **Critical findings**

   **BR-5 remains open — identity filtering depends on record order.** At `cmd/sdlc/internal/usage/codex.go:145`, records arriving before `session_meta` enter the modern response collection unchecked. Resolving the identity later never revisits them. Discovery accepts that same late header, so copied parent spend can acquire the child’s attribution.

   Enforce one rule across ordinary and compacted responses, before/after metadata, both aliases, and missing/invalid metadata. Buffer and reconcile against the immutable identity, or report uncertain ordering as incomplete. Missing executing identity must prevent confident spend coverage; missing cwd alone need not.

3. **Important findings**

   None additional.

4. **Minor findings**

   None.

5. **Test coverage notes**

   - Usage, transcripts, reviewusage and judge packages passed.
   - Targeted root usage/accounting tests passed; production dispatch and binding tests also passed.
   - Old-decoder and early-issue-filter mutations both caused the intended regressions to fail.
   - The new absent/late-header probe failed against pinned HEAD.
   - Full-suite, live-provider and real-store performance checks were not rerun during this review.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared identity parser and ownership-window helper.
   - **ARCH-PURE — pass:** normalization, aggregation and transitions remain directly testable.
   - **ARCH-PURPOSE — flag:** BR-5 still permits incorrect issue totals.
   - **ARCH-MOCK — pass:** production dispatch uses fake CLI seams; opt-in live conformance exists.
   - **ARCH-CONSTRAINTS — pass:** deadlines, scan bounds and incomplete-coverage paths are explicit.
   - **ARCH-SECURE — flag:** pre-header records gain attribution without executing-identity validation.
   - **ARCH-ORDER — flag:** identity resolution does not reconcile previously admitted responses.
   - **ARCH-FUNERAL — pass:** meter retention is bounded; private receipts have committed-evidence cleanup and admission limits.

7. **Plan revision recommendations**

   Add a `## Revisions` entry enumerating the complete executing-identity rule: absent, invalid, early and late headers; both aliases; copied later headers; ordinary and compacted responses. Require production discovery→Decode regressions proving that record order cannot admit another thread’s spend as complete.

---

## Re-review — 2026-10-15T01:43:33-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 365 — sdlc usage: tokens per slot, agent and model from local transcripts, with burn rate (fleet monitoring + bench metric) |
| repo | ariadne |
| issue file | workshop/issues/000365-sdlc-usage-tokens-per-slot-agent-and-model-from-local-transcripts-with-burn-rate-fleet-monitoring-bench-metric.md |
| boundary | whole-issue close |
| milestone | — |
| window | 0820d860ceb54656f8a7c6cab6f209e2b6a65729..ab74cbf2e2e9d3ca1cfa8754071b40d55abc78a3 |
| command | sdlc close --issue 365 |
| reviewer | codex |
| isolation | fenced |
| timestamp | 2026-10-15T01:43:33-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

BR-5 is addressed, with regression tests that fail when its guards are removed. The focused suites pass, and README/atlas updates cover the new surface. One separate blocking defect remains: a known launched reviewer with no matching local transcript can disappear from the report while both completeness flags remain true.

```findings
dispose:
  - id: BR-5
    disposition: addressed
    note: |
      Discovery and decoding share CodexExecutingID. Production matrix tests cover aliases, copied metadata, preidentity responses/snapshots and absent identity. Four independent guard-removal mutations fail the regression tests.
  - id: BR-1
    disposition: addressed
    note: |
      Thread, turn and legacy cumulative reconciliation tests pass; disabling reconciliation makes the missing-spend regressions fail.
  - id: BR-2
    disposition: addressed
    note: |
      Invalid artifact/private evidence remains visible and quarantined. Removing artifact validation diagnostics fails the ledger/comment regressions. The new finding concerns schema-valid receipts that cannot resolve to transcript evidence.
  - id: BR-3
    disposition: addressed
    note: |
      The plan's Core concepts tables now match delivered paths and distinguish pure decoding from IO adapters; provenance.go and engagement derivation are correctly recorded as reused unchanged.
  - id: BR-4
    disposition: addressed
    note: |
      Production boundary and planning dispatch tests pass. Independently disabling boundary, plan or estimate accounting attachment fails the corresponding production regression.
  - id: BR-6
    disposition: addressed
    note: |
      Cross-issue exclusion and cleanup lifecycle tests pass. Restoring issue-prefiltered records fails both ledger and comment exclusion regressions.
findings:
  - id: new
    severity: Critical
    kind: defect
    family: invalid-accounting-evidence-must-remain-visible
    title: |
      Launched reviewers without usable transcript bindings silently report complete coverage
    detail: |
      cmd/sdlc/usage_reviewers.go:184-199 drops a retained attempt when no session matches, and returns only bindings without unresolved-attempt diagnostics. Valid Finished records pass reviewer_accounting.go:154-157 even when identity capture failed. Two temporary regression probes reproduced empty coverage and complete=true for a launched receipt with missing-session-identity and for an exact session ID whose transcript is unavailable. A full collectUsage probe confirmed zero tokens with spend_complete=true and attribution_complete=true despite the known launched reviewer. This violates whole-system issue accounting and uncertainty preservation (ARCH-PURPOSE, ARCH-SECURE, ARCH-ORDER). This is the 2nd finding in family invalid-accounting-evidence-must-remain-visible. State and enforce the rule for every relevant launched attempt: reconcile it to usable transcript evidence or retain an explicit coverage gap. Sweep private, ledger and comment evidence, missing/exact identities, and unavailable/invalid sources; preserve confirmed launch-failed as a distinct non-spending outcome.
```

1. **Strengths**

   - The shared immutable identity parser removes discovery/decoder drift.
   - Deduplication precedes window and role selection; cumulative counters remain coverage evidence rather than additional charges.
   - Reviewer receipts use explicit transitions, bounded storage and exact committed-evidence cleanup.
   - README and atlas document commands, counter semantics and coverage limitations.

2. **Critical findings**

   - **Unresolved launched attempts disappear**, at `cmd/sdlc/usage_reviewers.go:184-199`. Return binding-resolution diagnostics and merge them into report coverage. Process completion must not imply accounting completeness.

3. **Important findings**

   None.

4. **Minor findings**

   None.

5. **Test coverage notes**

   Usage, transcripts, reviewusage and judge packages pass. Focused command, attribution, accounting and production-dispatch tests pass (`16.405s`). Mutations confirm regression sensitivity for BR-1/2/4/5/6.

   The new probes fail on pinned HEAD, including through `collectUsage`. They ran through temporary overlays, which were removed. The checkout remains unchanged. I did not rerun the full suite or live-provider conformance.

6. **Architectural notes**

   - **ARCH-DRY — pass:** shared identity and ownership contracts.
   - **ARCH-PURE — pass:** normalization, aggregation and transitions have separate IO adapters.
   - **ARCH-PURPOSE — flag:** missing reviewer evidence can produce a falsely complete cost report.
   - **ARCH-MOCK — pass:** portable repositories, fake CLI processes and opt-in conformance exist.
   - **ARCH-CONSTRAINTS — pass:** cancellation, deadlines and input bounds are enforced.
   - **ARCH-SECURE — flag:** absent transcript evidence becomes affirmative completeness.
   - **ARCH-ORDER — flag:** Finished process state is conflated with resolved accounting.
   - **ARCH-FUNERAL — pass:** meter retention and reviewer receipt bounds/cleanup are explicit.

7. **Plan revision recommendation**

   Append a `## Revisions` entry requiring each relevant launched attempt to resolve to usable transcript evidence or an explicit gap. Add production collector regressions across all three persistence forms, covering missing identity, unavailable transcript, invalid metadata and confirmed launch failure.
