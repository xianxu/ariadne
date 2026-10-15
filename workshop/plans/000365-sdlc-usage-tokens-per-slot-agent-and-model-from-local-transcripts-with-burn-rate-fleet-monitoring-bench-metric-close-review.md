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
