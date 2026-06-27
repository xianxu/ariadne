# Close Actual N/A Implementation Plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy) to determine the appropriate execution approach: use superpowers-subagent-driven-development (if subagents are suitable per AGENTS.md) or superpowers-executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `sdlc close --no-actual` produce a schema-valid closed issue with an explicit `actual_hours: N/A` sentinel, while keeping measured numeric actuals as the calibration path.

**Architecture:** Add one shared sentinel contract for issue actuals and route the schema, close write path, validation tests, and ledger skip behavior through it (`ARCH-DRY`). Keep sentinel recognition pure and small; the IO seams remain `runClose`, `appendCalibrationRow`, and `vocabulary validate-instance` (`ARCH-PURE`). Cover every consumer named in the issue spec, including help/atlas docs and invalid-string validation, so the change fulfills the whole purpose rather than just making close write a string (`ARCH-PURPOSE`).

**Tech Stack:** Go, CUE, Cobra helptext, existing `cmd/vocabulary` validation, existing `cmd/sdlc` close and estimate tests.

---

## Chunk 1: N/A Actual Contract

### Core Concepts

Pure entities:

| Name | Lives in | Status |
|------|----------|--------|
| `ActualNotApplicableSentinel` | `cmd/sdlc/internal/issue/frontmatter.go` or a new focused sibling if existing file shape argues for it | new |
| `IsActualNotApplicable` | `cmd/sdlc/internal/issue/frontmatter.go` or same sibling | new |
| `IssueActualSchema` | `construct/vocabulary/issue.cue` | modified |
| `CalibrationLedgerRows` | `cmd/sdlc/internal/estimate/ledger.go` | modified |

- **ActualNotApplicableSentinel** — the single spelling for no measured actual: `N/A`.
  - **Relationships:** One sentinel is written by `sdlc close --no-actual`, accepted by `#Issue`, documented in help/atlas, and excluded from calibration.
  - **DRY rationale:** The issue calls for a small closed set of valid spellings. One exported constant avoids parallel literals in close, tests, and docs.
  - **Future extensions:** If the sentinel ever changes, the schema and docs should be updated in the same commit; do not add aliases unless there is an operator need.

- **IsActualNotApplicable** — pure predicate accepting only the exact sentinel after whitespace trimming.
  - **Relationships:** Used anywhere Go code needs to distinguish `N/A` from arbitrary strings.
  - **DRY rationale:** Prevents each caller from reimplementing string checks differently.
  - **Future extensions:** Natural place to widen if a future schema intentionally accepts another sentinel.

- **IssueActualSchema** — `actual_hours` permits positive numeric values, empty/null while open, and the exact `N/A` sentinel for closed issues.
  - **Relationships:** `#Issue` owns validation; `sdlc issue validate` and push/merge gates consume it.
  - **DRY rationale:** Keep the formal datatype model as the acceptance source rather than duplicating validation in the close command.
  - **Future extensions:** Non-done terminal statuses can be addressed later if their actual-hours semantics differ; this issue focuses on valid done closes from `sdlc close --no-actual`.

- **CalibrationLedgerRows** — parser continues to ignore malformed/non-numeric rows, with an explicit test that `actual=N/A` is excluded.
  - **Relationships:** `sdlc close` should not append N/A rows, and drift logic should never count them.
  - **DRY rationale:** The existing `shouldLogCalibration` gate is the main exclusion; parser-level skipping is a defense for existing/manual ledger rows.
  - **Future extensions:** If ledger schema gets typed, `actual` can become number-or-sentinel there too; for now, skipped rows preserve numeric-only calibration.

Integration points:

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| `runClose` issue write path | `cmd/sdlc/close.go` | modified | issue file write |
| `vocabulary validate-instance` | `cmd/vocabulary/validate_test.go` plus `construct/vocabulary/issue.cue` | modified | `cue vet` |
| `close helptext` | `cmd/sdlc/helptext/close.md` | modified | user-facing CLI docs |
| `workflow atlas docs` | `atlas/workflow/issue-lifecycle.md`, `atlas/workflow/vocabulary.md`, likely `atlas/workflow/sdlc-binary.md` | modified | repo map docs |

- **runClose issue write path** — when the actual gate is explicitly bypassed for a full issue close, write `actual_hours: N/A` and warn that calibration is skipped.
  - **Injected into:** Existing close tests should use temp issue files and existing close helpers rather than mocking file IO.
  - **Future extensions:** Milestone close can keep omitting partial actual if no actual is supplied, unless tests reveal it also mutates a full issue actual.

- **vocabulary validate-instance** — real CUE integration tests prove accepted/rejected frontmatter shapes.
  - **Injected into:** `sdlc issue validate` and merge/push validation gates already shell through this engine.
  - **Future extensions:** If a future validation API imports generated JSON instead of shelling to CUE, the same cases should move with it.

- **close helptext** — documents that `--no-actual` records `actual_hours: N/A`, is only for not-applicable/unavailable telemetry cases, and skips velocity calibration.
  - **Injected into:** Cobra help via embedded helptext.
  - **Future extensions:** Keep the short flag text and deep-dive section aligned.

- **workflow atlas docs** — update the map where it currently says done requires positive actuals.
  - **Injected into:** Human/operator reference, not code.
  - **Future extensions:** If project-file actual syntax later accepts N/A, document that separately.

### Task 1: Schema And Validation Tests

**Files:**
- Modify: `construct/vocabulary/issue.cue`
- Modify: `cmd/vocabulary/validate_test.go`
- Potentially modify generated files after weave: `construct/generated/vocabulary/issue.json`, `pkg/vocab/issue.json`

- [ ] **Step 1: Write failing validation tests**

Add real-CUE cases in `TestValidateInstance_ValidPasses` or a focused sibling:

```go
cases := []struct {
    name string
    fm   string
}{
    {"done numeric actual", "id: \"000001\"\nstatus: done\nactual_hours: 1.25\n"},
    {"done not applicable actual", "id: \"000001\"\nstatus: done\nactual_hours: N/A\n"},
}
```

Add a rejection case to `TestValidateInstance_RejectsMalformations`:

```go
{"done invalid actual string", "id: \"000001\"\nstatus: done\nactual_hours: unknown\n", "actual_hours", "not valid"},
```

- [ ] **Step 2: Run validation tests to verify RED**

Run: `go test ./cmd/vocabulary -run 'TestValidateInstance_(ValidPasses|RejectsMalformations)' -count=1`

Expected: `done not applicable actual` fails because `actual_hours: N/A` conflicts with the current numeric-only done guard.

- [ ] **Step 3: Update `#Issue` actual schema**

In `construct/vocabulary/issue.cue`, introduce an exact sentinel definition and use it in `actual_hours`:

```cue
#ActualNotApplicable: "N/A"
actual_hours?: (number & >0) | #ActualNotApplicable | null
if status == "done" {
    actual_hours!: (number & >0) | #ActualNotApplicable
}
```

Update nearby comments and lifecycle guard text from "measured actuals only" to "measured actuals or an explicit not-applicable sentinel".

- [ ] **Step 4: Run validation tests to verify GREEN**

Run: `go test ./cmd/vocabulary -run 'TestValidateInstance_(ValidPasses|RejectsMalformations|ActiveCorpusPasses)' -count=1`

Expected: PASS. Numeric done actuals and `N/A` pass; missing actual and arbitrary strings still fail.

- [ ] **Step 5: Regenerate vocabulary outputs**

Run: `make weave`

Expected: generated vocabulary files are fresh. If `issue.json` does not change because the exported JSON only carries concrete categories/lifecycle, note that in the issue log.

- [ ] **Step 6: Run vocabulary freshness checks**

Run: `vocabulary vet`

Expected: PASS.

Run: `vocabulary check --output construct/generated/vocabulary`

Expected: PASS.

### Task 2: Close Write Path

**Files:**
- Modify: `cmd/sdlc/close.go`
- Modify: `cmd/sdlc/close_test.go` or add a focused test near existing close write-path tests
- Modify: `cmd/sdlc/internal/issue/frontmatter.go` or add a focused sibling for the sentinel constant/predicate

- [ ] **Step 1: Write failing close test**

Add an integration-style test using existing `closeRepo` helpers if possible:

```go
func TestRunClose_NoActualWritesNotApplicableSentinel(t *testing.T) {
    issuesDir := closeRepo(t, 135)
    f := &closeFlags{
        Issue: 135, NoActual: true, Verified: "administrative close",
        NoAtlas: true, NoProject: true, IssuesDir: issuesDir,
        BrainDir: "../nonexistent-brain",
    }
    if err := runClose(io.Discard, f); err != nil {
        t.Fatalf("runClose: %v", err)
    }
    data, err := os.ReadFile(filepath.Join(issuesDir, "000135-close.md"))
    if err != nil { t.Fatal(err) }
    if !strings.Contains(string(data), "actual_hours: N/A") {
        t.Fatalf("missing actual sentinel:\n%s", data)
    }
}
```

Adjust the exact filename to match the helper output.

- [ ] **Step 2: Run the close test to verify RED**

Run: `go test ./cmd/sdlc -run TestRunClose_NoActualWritesNotApplicableSentinel -count=1`

Expected: FAIL because current close writes `status: done` without `actual_hours`.

- [ ] **Step 3: Add the sentinel contract and write it from close**

Add one Go source for:

```go
const ActualNotApplicableSentinel = "N/A"

func IsActualNotApplicable(s string) bool {
    return strings.TrimSpace(s) == ActualNotApplicableSentinel
}
```

Use it in `runClose`: when closing a full issue and `f.Actual == "" && f.skip("actual")`, set `actual_hours` to the sentinel and include it in the success message. Keep numeric `--actual` behavior unchanged.

Update the warning from "closing with NO actual_hours" to "closing with actual_hours: N/A; velocity calibration skipped".

- [ ] **Step 4: Run the close test to verify GREEN**

Run: `go test ./cmd/sdlc -run TestRunClose_NoActualWritesNotApplicableSentinel -count=1`

Expected: PASS.

- [ ] **Step 5: Add or update warmup/flag text tests if present**

Search for assertions on `--no-actual` wording. If found, update them to expect the sentinel and calibration skip message.

Run: `go test ./cmd/sdlc -run 'TestClose|TestHelptext|TestEmbed' -count=1`

Expected: PASS.

### Task 3: Calibration Ledger And Drift Exclusion

**Files:**
- Modify: `cmd/sdlc/close_ledger_test.go`
- Modify: `cmd/sdlc/internal/estimate/ledger_test.go`
- Modify: `cmd/sdlc/internal/estimate/ledger.go` only if tests show parser currently counts N/A rows as zero

- [ ] **Step 1: Write failing parser test for N/A ledger rows**

Add to `cmd/sdlc/internal/estimate/ledger_test.go`:

```go
func TestParseRows_SkipsNotApplicableActualRows(t *testing.T) {
    text := Header() + "\n" +
        "ariadne#135\t2.00\t0.20\t1.80\tN/A\t-\testimate-logic-v2\t-\tyes\t2026-06-26\n" +
        FormatRow(LedgerRow{Issue: "x", Estimate: 1, Actual: 1, Model: "estimate-logic-v2", Date: "2026-06-26"}) + "\n"
    if got := ParseRows(text); len(got) != 1 {
        t.Fatalf("N/A actual row should be skipped; got %d rows", len(got))
    }
}
```

- [ ] **Step 2: Run parser test to verify RED**

Run: `go test ./cmd/sdlc/internal/estimate -run TestParseRows_SkipsNotApplicableActualRows -count=1`

Expected: FAIL because current `atof("N/A")` returns zero and the row is counted.

- [ ] **Step 3: Update row parser**

In `ParseRows`, parse numeric fields with an ok-return helper. If `actual` is not numeric, skip the row. Keep existing behavior for ragged rows, comments, and header.

- [ ] **Step 4: Run parser test to verify GREEN**

Run: `go test ./cmd/sdlc/internal/estimate -run 'TestParseRows|TestRoundTrip|TestFormatRow' -count=1`

Expected: PASS.

- [ ] **Step 5: Prove close does not append N/A calibration rows**

Extend `TestShouldLogCalibration` or add a close wiring test for `NoActual: true, Actual: ""` to assert `shouldLogCalibration` remains false and no ledger file is written.

Run: `go test ./cmd/sdlc -run 'TestShouldLogCalibration|TestRunClose_LedgerGuardWiring' -count=1`

Expected: PASS.

### Task 4: Help And Atlas Docs

**Files:**
- Modify: `cmd/sdlc/helptext/close.md`
- Modify: `atlas/workflow/issue-lifecycle.md`
- Modify: `atlas/workflow/vocabulary.md`
- Modify: `atlas/workflow/sdlc-binary.md` if its close/actual sections still imply `--no-actual` omits the field

- [ ] **Step 1: Update help text**

Change the `--no-actual` flag description and deep-dive text to state:

- `--no-actual` writes `actual_hours: N/A`.
- Use it only when measurement is genuinely not applicable or telemetry is unavailable.
- N/A rows are excluded from velocity calibration and drift stats.
- Numeric `--actual` remains the normal measured path.

- [ ] **Step 2: Update atlas docs**

Adjust docs that say done issues require positive actuals so they say done issues require either a positive numeric actual or explicit `actual_hours: N/A`.

- [ ] **Step 3: Run help/doc tests**

Run: `go test ./cmd/sdlc -run 'TestHelptext|TestEstimate|TestEmbed' -count=1`

Expected: PASS.

### Task 5: End-To-End Validation

**Files:**
- Modify: `workshop/issues/000135-close-actual-na.md`
- Modify: `workshop/plans/000135-close-actual-na-plan.md`

- [ ] **Step 1: Run targeted tests**

Run:

```bash
go test ./cmd/vocabulary ./cmd/sdlc/internal/estimate ./cmd/sdlc -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Validate a synthetic N/A issue through the CLI**

Create a temp markdown issue with `status: done` and `actual_hours: N/A`, then run:

```bash
vocabulary validate-instance --type issue /tmp/issue-na.md
```

Expected: exit 0.

Create a sibling temp issue with `actual_hours: unknown`, then run the same command.

Expected: non-zero with an `actual_hours` diagnostic.

- [ ] **Step 4: Validate the active issue corpus**

Run:

```bash
sdlc issue validate --all
```

Expected: PASS.

- [ ] **Step 5: Review git diff for unrelated work**

Run:

```bash
git status --short --untracked-files=all
git diff -- workshop/issues/000135-close-actual-na.md workshop/plans/000135-close-actual-na-plan.md construct/vocabulary/issue.cue cmd/vocabulary/validate_test.go cmd/sdlc/close.go cmd/sdlc/close_test.go cmd/sdlc/close_ledger_test.go cmd/sdlc/internal/estimate/ledger.go cmd/sdlc/internal/estimate/ledger_test.go cmd/sdlc/helptext/close.md atlas/workflow/issue-lifecycle.md atlas/workflow/vocabulary.md atlas/workflow/sdlc-binary.md
```

Expected: only #135 files plus pre-existing unrelated dirty files remain.

