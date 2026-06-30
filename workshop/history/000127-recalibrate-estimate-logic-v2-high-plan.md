# Recalibrate Estimate Logic v2 High Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** Recalibrate the estimate model numerically for `sdlc actual` ship-wall-clock, then make `sdlc` use that new model without stale-model drift noise.

**Architecture:** The numeric hour-range change belongs in the brain velocity calibration docs, not in `vocab.go`; ariadne owns the shared grammar, recognized model set, default model pointer, and drift guard. Keep the current model as a pure shared fact in `cmd/sdlc/internal/estimate` and route command defaults through it (ARCH-DRY). Keep drift selection pure and tested in `internal/estimate`; the close command remains a thin IO appender (ARCH-PURE). This plan includes the brain-side model files because recalibration is the issue's purpose, not a follow-up (ARCH-PURPOSE).

**Tech Stack:** Go, Cobra command tests, markdown helptext, brain velocity calibration docs as external source.

---

## Core Concepts

| Name | Lives in | Status |
|------|----------|--------|
| `EstimateLogicV3_1` | `/Users/xianxu/workspace/brain/data/life/42shots/velocity/estimate-logic-v3.1.md` | new |
| `BaselineV3_1` | `/Users/xianxu/workspace/brain/data/life/42shots/velocity/baseline-v3.1.md` | new |
| `CurrentModel` | `cmd/sdlc/internal/estimate/vocab.go` | new |
| `DriftSample` | `cmd/sdlc/internal/estimate/drift.go` | modified |

**EstimateLogicV3_1** — numerical calibration patch over v3 for the `sdlc actual` unit.
- **Relationships:** Pairs 1:1 with `BaselineV3_1`; consumed by agents when writing future `## Estimate` blocks.
- **DRY rationale:** Keeps hour ranges in the existing brain calibration source instead of duplicating ranges in ariadne code.
- **Future extensions:** v4 can add actual-quality metadata or new primitives if 40% impl scaling proves structurally wrong.

**BaselineV3_1** — calibration evidence and backtest for the v3.1 scale factor.
- **Relationships:** Extends `baseline-v3.md`; owns the ledger analysis and outlier classification.
- **DRY rationale:** Puts the why next to the model patch so future estimators do not infer from issue prose.
- **Future extensions:** Add rows as v3.1 closes accumulate; promote actual-quality fixes to #92 or a new issue if needed.

**CurrentModel** — the default estimate model provenance used by `sdlc` pushes and pulls.
- **Relationships:** 1:many from model vocabulary to `estimate-source`, `start-plan`, and help examples.
- **DRY rationale:** Replaces hard-coded `"estimate-logic-v2"` defaults spread across commands.
- **Future extensions:** A later v3.1/v4 bump changes one constant plus docs/tests.

**DriftSample** — the pure selection logic behind `DriftVerdict`.
- **Relationships:** 1:1 with the calibration ledger stream; selects the latest eligible rows for one model.
- **DRY rationale:** Keeps all row eligibility rules in the estimate package instead of scattering filters in `close.go`.
- **Future extensions:** Can add explicit actual-quality fields later if active-time writes commit-count/window metadata.

## Integration Points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| Brain velocity current-version pointer | `/Users/xianxu/workspace/brain/data/life/42shots/velocity/SKILL.md` | modified | estimator skill metadata |
| `estimate-source` default | `cmd/sdlc/estimatesource.go` | modified | CLI flag default + brain file stat |
| `start-plan` estimate source push | `cmd/sdlc/startplan.go` | modified | planning checkpoint output |
| `sdlc close` drift warning | `cmd/sdlc/close.go` | unchanged caller | ledger append + warning display |

**Brain velocity current-version pointer** — tells estimators which paired logic/baseline files are current.
- **Injected into:** Human/agent estimation workflow outside ariadne.
- **Future extensions:** Same pointer updates on each model bump.

**estimate-source default** — resolves the calibration source for the current model unless `--model` overrides it.
- **Injected into:** `estimateSourceStatus`, which remains the IO seam.
- **Future extensions:** A future default model can move without touching Cobra command call sites.

**start-plan estimate source push** — tells agents which calibration source to read during planning.
- **Injected into:** `runStartPlan`, using the pure current-model value.
- **Future extensions:** Could render both current and override provenance if the workflow grows per-repo model pins.

**sdlc close drift warning** — appends the ledger row and calls `estimate.DriftVerdict`.
- **Injected into:** No new dependency; the pure selection change is internal to `DriftVerdict`.
- **Future extensions:** If #92 adds actual-quality metadata, this call site should still only pass rows.

## Chunk 1: Numerical Recalibration

### Task 1: Diagnose and write v3.1 calibration docs

**Files:**
- Create: `/Users/xianxu/workspace/brain/data/life/42shots/velocity/estimate-logic-v3.1.md`
- Create: `/Users/xianxu/workspace/brain/data/life/42shots/velocity/baseline-v3.1.md`
- Modify: `/Users/xianxu/workspace/brain/data/life/42shots/velocity/SKILL.md`
- Modify: `/Users/xianxu/workspace/brain/data/life/42shots/velocity/calibration-findings.md` if it summarizes current open calibration state

- [x] **Step 1: Preserve existing brain WIP**

Run: `git -C /Users/xianxu/workspace/brain status --short --branch`
Expected: existing unrelated parley/tracker dirt is visible; stage/edit only velocity files.

- [x] **Step 2: Record the diagnosis**

Use the latest unique trusted v2 rows. Treat rows with very low actuals or known single-commit/degenerate windows as actual-side artifacts; use the clean rows around pair#67, ariadne#122/#126/#124/#125, pair#71, parley.nvim#144/#147 to choose the scale.

- [x] **Step 3: Create `baseline-v3.1.md`**

Document the backtest: old ratios around 2.0-2.5 on clean rows become roughly 0.6-1.35 when estimates use `design*1.15 + impl*0.40`; high residual rows remain marked as actual-quality outliers, not additional model scale evidence.

- [x] **Step 4: Create `estimate-logic-v3.1.md`**

State the algorithm patch:
- v3.1 extends v3, which extends v2.1.
- Keep v2.1's library-availability check.
- Keep design hours unchanged.
- Use +15% design buffer for thorough plans.
- For ariadne-style AI-paired work measured by `sdlc actual`, write item `impl=` values at 40% of the v2/v2.1 primitive-table implementation hours.
- Keep provenance as `estimate-logic-v3.1`.

- [x] **Step 5: Update the velocity skill current-version table**

Point current algorithm to `estimate-logic-v3.1.md` and baseline to `baseline-v3.1.md`.

## Chunk 2: Current Model Provenance

### Task 2: Add a single current-model source

**Files:**
- Modify: `cmd/sdlc/internal/estimate/vocab.go`
- Modify: `cmd/sdlc/internal/estimate/vocab_test.go`

- [x] **Step 1: Write the failing test**

Add `TestCurrentModel` asserting `CurrentModel() == "estimate-logic-v3.1"` and `KnownModel(CurrentModel())`.

- [x] **Step 2: Run test to verify it fails**

Run: `go test ./cmd/sdlc/internal/estimate -run TestCurrentModel -count=1`
Expected: FAIL because `CurrentModel` does not exist.

- [x] **Step 3: Implement minimal code**

Add a private constant and exported function:

```go
const currentModel = "estimate-logic-v3.1"

func CurrentModel() string { return currentModel }
```

Also add `"estimate-logic-v3.1"` to the recognized model set.

- [x] **Step 4: Run test to verify it passes**

Run: `go test ./cmd/sdlc/internal/estimate -run TestCurrentModel -count=1`
Expected: PASS.

### Task 3: Route command defaults through `CurrentModel`

**Files:**
- Modify: `cmd/sdlc/estimatesource.go`
- Modify: `cmd/sdlc/estimatesource_test.go`
- Modify: `cmd/sdlc/startplan.go`
- Modify: `cmd/sdlc/startplan_test.go`

- [x] **Step 1: Write failing tests**

Add coverage that `NewEstimateSourceCmd` has `--model` defaulting to `estimate.CurrentModel()`, and that `runStartPlan` output includes `estimate.CurrentModel()`.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./cmd/sdlc -run 'TestEstimateSourceCmd_Registered|TestRunStartPlan_RendersAtPlanLens' -count=1`
Expected: FAIL while defaults still hard-code v2.

- [x] **Step 3: Implement minimal code**

Use `estimate.CurrentModel()` in `NewEstimateSourceCmd` and `runStartPlan`.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./cmd/sdlc -run 'TestEstimateSourceCmd_Registered|TestRunStartPlan_RendersAtPlanLens' -count=1`
Expected: PASS.

## Chunk 3: Drift Window Semantics

### Task 4: Make drift compare one model revision at a time

**Files:**
- Modify: `cmd/sdlc/internal/estimate/drift.go`
- Modify: `cmd/sdlc/internal/estimate/drift_test.go`

- [x] **Step 1: Write failing tests**

Add tests for:
- a latest v3 row does not inherit four prior v2 over-estimates;
- repeated rows for the same issue/model count once, using the latest row;
- blank or unknown model rows do not become the active drift sample.

- [x] **Step 2: Run tests to verify failure**

Run: `go test ./cmd/sdlc/internal/estimate -run 'TestDrift_' -count=1`
Expected: FAIL because current drift uses the last `n` trusted rows regardless of model or duplicate issue rows.

- [x] **Step 3: Implement minimal code**

Inside `DriftVerdict`, select the latest eligible known model, walk rows backward, keep only trusted rows with that model and `Actual > 0`, dedupe by issue, then evaluate the latest `n` unique rows.

- [x] **Step 4: Run tests to verify pass**

Run: `go test ./cmd/sdlc/internal/estimate -run 'TestDrift_' -count=1`
Expected: PASS.

## Chunk 4: Docs, Backtest, And Atlas

### Task 5: Update shared-method docs

**Files:**
- Modify: `cmd/sdlc/helptext/estimate.md`
- Modify: `cmd/sdlc/helptext/estimate-source.md`
- Modify: `cmd/sdlc/internal/estimate/block.go`
- Modify: `cmd/sdlc/internal/judge/prompts.go` if it still hard-codes v2-only guidance.

- [x] **Step 1: Update docs after code is green**

Replace default examples with `estimate-logic-v3.1`, explain that the primitive vocabulary descends from v2 while v3.1 is the current provenance, and preserve the statement that hour ranges live in brain calibration docs.

- [x] **Step 2: Run focused tests**

Run: `go test ./cmd/sdlc/internal/estimate ./cmd/sdlc -count=1`
Expected: PASS.

### Task 6: Backtest recent ledger behavior

**Files:**
- Modify: `workshop/issues/000127-recalibrate-estimate-logic-v2-high.md`

- [x] **Step 1: Compute recent samples**

Run an `awk` or `go test` fixture-backed check against `brain/.../calibration-ledger.tsv` showing:
- clean latest unique v2 rows re-center under the v3.1 formula;
- low-actual/high-residual rows are explicitly classified as actual-quality artifacts;
- a new v3.1 row will not inherit old v2 drift because drift samples one model revision at a time.

- [x] **Step 2: Log the diagnosis**

Append the backtest result to `## Log`, explicitly separating model bias from actual-side artifacts and naming the brain files changed.

### Task 7: Update atlas

**Files:**
- Modify: `atlas/workflow/sdlc-binary.md`
- Modify: `atlas/workflow/ledger-landscape.md` if the drift semantics are documented there.
- Modify: `atlas/index.md` only if a new atlas page is added.

- [x] **Step 1: Update relevant workflow map**

Document current-model provenance and per-model drift windows; do not add a new page unless existing pages cannot hold the map.

- [x] **Step 2: Final verification**

Run: `go test ./cmd/sdlc/internal/estimate ./cmd/sdlc -count=1`
Expected: PASS.

Run: `git diff --check`
Expected: no whitespace errors.

## Revisions

### 2026-06-29 — plan-quality rescope to real recalibration

Reason: the first `sdlc change-code` plan-quality check returned `VERDICT: FAILURE`; adopting v3 as current would not change estimate numbers because v3 is a pure actuals-method shim over v2.1.

Delta: added Chunk 1 for brain-side `estimate-logic-v3.1.md` / `baseline-v3.1.md`, changed `CurrentModel` target from v3 to v3.1, and changed the backtest acceptance from "new model window does not inherit v2 drift" to "clean rows re-center under the v3.1 formula while artifact rows stay called out."
