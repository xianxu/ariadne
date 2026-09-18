# Quick flow implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute with superpowers-executing-plans in the main session (the design rides on warm context); the SDLC gates own the fresh-context reviews. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Issues carry a `flow: {kind, provenance}` record that the gates infer and act on, so a small change runs claim → change-code → close with no plan-quality or estimate gates and one small-diff review, while anything outside a hard size shell is upgraded to the full flow.

**Architecture:** One pure package, `cmd/sdlc/internal/flow`, owns the record, the inference, the shell and the Done-when checks. `change-code` and `close`/`milestone-close` are thin IO shells around it. The review recipe reuses the existing boundary-review body, parametrized by a marker subset drawn from the single ARCH registry.

**Tech Stack:** Go (cobra, `go.yaml.in/yaml/v3`), CUE (`construct/vocabulary/issue.cue`), embedded markdown prompts, git.

Spec: `workshop/issues/000231-a-quick-path-for-small-diffs-scale-the-gate-set-and-the-review-recipe-to-size.md`.

---

## Core concepts

### Pure entities

| Name | Lives in | Status |
|------|----------|--------|
| `Flow`, `Kind`, `Provenance` | `cmd/sdlc/internal/flow/flow.go` | new |
| `flow.Parse` / `flow.Format` / `flow.FromFrontmatter` | `cmd/sdlc/internal/flow/flow.go` | new |
| `flow.Decide` | `cmd/sdlc/internal/flow/flow.go` | new |
| `flow.MaxCodeFiles`, `flow.MaxChangedLines` | `cmd/sdlc/internal/flow/shell.go` | new |
| `flow.Measure`, `flow.Crossings` | `cmd/sdlc/internal/flow/shell.go` | new |
| `flow.Surfaces`, `flow.ParseSurfaces` | `cmd/sdlc/internal/flow/surfaces.go` | new |
| `flow.DoneWhenPresent`, `flow.DoneWhenFresh` | `cmd/sdlc/internal/flow/donewhen.go` | new |
| `issue.MilestonesInPlanOrder` | `cmd/sdlc/internal/issue/plan.go` | modified (moved from `close.go`) |
| `issue.HasDoneWhenBullet` | `cmd/sdlc/internal/issue/structural.go` | new (factored out of `checkDoneWhen`) |
| `churn.IsDoc`, `churn.IsEmbedded`, `churn.IsCodeFile` | `cmd/sdlc/internal/churn/classify.go` | new |
| `churn.isTestPath` | `cmd/sdlc/internal/churn/classify.go` | modified (non-Go test layouts) |
| `judge.ArchitectureBlockFor`, `judge.ArchitectureSection` | `cmd/sdlc/internal/judge/architecture.go` | new (`ArchitectureBlock` delegates) |
| `judge.ReviewMarkers`, `judge.SmallDiffMarkers` | `cmd/sdlc/internal/judge/prompts.go` | new |
| `judge.SmallDiffReview` category | `cmd/sdlc/internal/judge/prompts.go` + `prompts/small-diff-review.md` | new |
| `judge.CodeReviewBody` | `cmd/sdlc/internal/judge/review.go` | modified (takes the marker set) |
| `estimate.LedgerRow` flow columns | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `estimate.driftSample` | `cmd/sdlc/internal/estimate/drift.go` | modified (excludes quick/upgraded) |

- **Flow** — `{Kind: full|quick, Provenance: inferred|operator}`, serialized on one line as `{kind: quick, provenance: inferred}` because the issue frontmatter helpers are line-based. `FromFrontmatter(fm)` returns `(Flow, recorded bool, err)`: absent reads as full and not recorded. A malformed value is an error, which callers treat as `full`, the stricter flow.
  - **Relationships:** 1:1 with an issue; written by change-code and by the close finalize.
  - **DRY rationale:** one parser/formatter for every reader (change-code, close, milestone-close, calibration). It reuses yaml/v3 as `project.DecodeMetadata` does.
  - **Future extensions:** #233 adds `Kind` `config`.
- **Decide** — `Decide(in DecideInput) (Flow, error)`, where `DecideInput{Recorded *Flow; Pin string; HasMilestones, HasPlan bool}`. Rules, in order:
  - A pin sets `{pin, operator}`. `quick` with milestones is an error, as is an unknown pin.
  - A recorded operator flow stands.
  - A recorded `full` stands, because gates never downgrade.
  - Otherwise the result is `{full if HasMilestones || HasPlan else quick, inferred}`.
- **Measure / Crossings** — `Measure(files []string, stats []churn.FileStat, s Surfaces, milestones []string) Size` counts code files (via `churn.IsCodeFile`), the insertions in those files, the shared surfaces touched and the Mx rows. `Crossings(Size) []string` returns one human reason per crossed limit, or nil. Both are pure and table-tested at each limit.
  - **DRY rationale:** the limits live once. The help text and the constitution read them through `{{QUICK_MAX_*}}` tokens and a pointer to `sdlc change-code --help`.
- **Surfaces** — parsed from `.sdlc/shared-surfaces`: one `path.Match` pattern per line, `#` comments, and a trailing `/` meaning "anything under". The declaration file always matches itself, so a branch cannot exempt its own diff.
  - A malformed line is returned as an error. Close turns that error into a crossing, so it fails toward `full`.
- **DoneWhenPresent / DoneWhenFresh** — `DoneWhenPresent(body) error` requires a Done-when bullet; `related:` does not count, because this is the review's oracle. `DoneWhenFresh(anchorBody, body string) error` refuses when `## Spec` differs from the anchor and `## Done when` does not.
- **churn classifiers** — `IsDoc` is #177's per-path docs rule and `IsEmbedded` is the `cmd/` rule from #174. `IsCodeFile(p) = ClassifyPath(p) == CodeProd && (!IsDoc(p) || IsEmbedded(p))`. `hasCodePath` and `publishGateHasCodeSurface` become thin `any()` loops over these, so there is one per-path rule and three readings.
  - `isTestPath` widens to the non-Go layouts the fleet uses: `*_spec.lua`, `*_test.lua`, `*_test.py`, `test_*.py`, `*.test.*`, `*.spec.*`, and path segments `test`, `tests`, `spec`, `__tests__`, `testdata`. The motivating repo (parley.nvim) keeps its tests under `tests/**/*_spec.lua`.
- **ArchitectureSection / ArchitectureBlockFor** — a production slicer (promoted from the `architectureEntry` test helper) that splits the registry losslessly into a preamble and per-marker sections. `ArchitectureBlockFor(lens, markers)` renders the header, the preamble and the selected sections. `ArchitectureBlock(lens)` equals `ArchitectureBlockFor(lens, ArchitectureMarkers())` byte for byte, which the goldens pin.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| change-code flow step | `cmd/sdlc/changecode.go` | modified | issue file read/write, plan-file lookup |
| close shell + Done-when step | `cmd/sdlc/close.go` (`computeClose`) | modified | `git diff --numstat`, `git log -G`, `git show`, `.sdlc/shared-surfaces` read |
| recipe selection | `cmd/sdlc/milestoneclose.go` (`boundaryReviewParams`, `boundaryReviewDispatchOptions`) | modified | judge subprocess (`judge.Run` seam) |
| calibration row | `cmd/sdlc/close.go` (`appendCalibrationRow`) | modified | brain TSV append |
| help tokens | `cmd/sdlc/main.go` (`renderLong`) | modified | embedded helptext |

- **change-code flow step** — runs after the issue and plan are read (`changecode.go:130`) and before the gate loop.
  - Inputs are the milestone rows (`issue.MilestonesInPlanOrder(PlanItemsBody)`) and `planContent != ""`, which is the same plan lookup plan-quality uses.
  - It writes the record with `Parse → SetField → Compose` unless `--dry-run`, and prints one info line.
  - Each gate closure skips on `quick`: structural, plan-quality, estimate, estimate-recon and estimate-quality. `changeCodeGateOrder()` still returns all five, so the guard tests keep their strength.
  - `planGateContent` strips `flow:` as well as `estimate_hours:`, so writing the record does not bust the plan-gate pass-through cache.
- **close shell + Done-when step** — runs inside `computeClose` after the atlas block (~`close.go:530`), where `windowBase`, `windowHead` and `diffFiles` already exist. That is the same window the atlas gate and the review use (#58).
  - File set: `diffFiles`, which includes binaries. Lines: `churn.ParseNumstat(git diff --numstat base..head)` through a new `windowFileStats(base, head)` shared with `churnForWindow`.
  - On a crossing, the upgrade goes into `newFM` plus a Log line, and `applyClose` writes it at finalize. A REWORK writes nothing (#139), so the next close re-derives the upgrade from the same window.
  - The Done-when checks run when the recorded kind is `quick`, including when the close is upgrading it.
  - The freshness anchor is the first commit that added a `flow:` line to the issue (`git log --reverse --format=%H -G'^flow: ' -- <issue>`), read with `git show`. With no anchor, close refuses and names `sdlc issue sync` or `--no-done-when-fresh`.
- **recipe selection** — a new `boundaryReviewParams.Category` field is set from `closeResult.flow` at `close.go:1024/:1037/:1078`. milestone-close always passes `MilestoneReview`. `boundaryReviewDispatchOptions` reads it for both `BuildPrompt` and `AllowedTools`.
- **Test seams (existing, reused):** `judge.Run` for the LLM, real git in `testfix.Repo` / `syncRepo` temp repos (git is the external binary and there is no network), `computeActualForCloseFn`, `die`/`expectDie`, and `executeSDLCTestCommand`. `TestMain` guards the real repo.

## Decisions and architecture

- **Changed lines means insertions in code files.** That is git's `+` count, the same number the churn report and ledger already use; deletions don't count. A 60-line modification is 60 changed lines, not 120. Files are counted from `--name-only`, so a binary change still counts as a file.
- **The flow upgrade is recorded at finalize, not before the review.** `computeClose` writes nothing, and #139's invariant is that a REWORK leaves the issue unwritten. The rule "no gate downgrades" is therefore about the *recorded* flow. The measurement is a pure function of the review window, so a REWORK followed by a re-close re-derives the same upgrade. It can only differ if the rework itself deleted code back inside the shell, and then the small-diff review correctly reviews what exists.
- **The shared-surface declaration is a plain line file**, `.sdlc/shared-surfaces`, not a vocabulary noun. It is repo-owned, like `.sdlc/fleet.json`, and not woven. `path.Match` plus a trailing-`/` prefix covers depth without a `**` dependency. ariadne gets a starter declaration in M2; parley.nvim's declaration is peer work, listed as a follow-up and not deferred purpose (the mechanism is the deliverable).
- **Refusals the quick flow keeps:** `change-code --flow quick` on a Plan with Mx rows, and at close an empty Done-when and a stale Done-when (skip with `--no-done-when-fresh`). Crossing the shell never refuses.
- **ARCH-DRY:** one flow record codec, one milestone parser (the colon-requiring `milestoneLabelRE` in `sizing.go` is retired onto it), one per-path classifier set in `churn`, one review body parametrized by markers, and one shared boundary-review tail (`{{BOUNDARY_TAIL}}`, expanded at template load so `milestone-review.md` renders byte-identically). Plan lookup reuses `readOptionalPlanFile`, the window reuses `boundaryWindowBase`, and the numstat call is shared with `churnForWindow`.
- **ARCH-PURE:** all decisions live in `internal/flow`, `internal/churn` and `internal/judge` and are tested on strings. The IO is three thin sites: the change-code flow step, the close measurement (git + file read) and the ledger append.
- **ARCH-PURPOSE:** every consumer derives from the record — change-code, close, milestone-close, recipe selection, calibration, help text, constitution and atlas. Test detection widens to non-Go layouts because the motivating case is a Lua repo.
- **ARCH-MOCK:** no new external dependency. git runs for real in temp repos (the established pattern), and the judge is faked at `judge.Run`.
- **ARCH-CONSTRAINTS:** a CLI batch step at close, with no UI path.
  - It adds one `git diff --numstat` over a window already diffed by `--name-only`. On quick issues only, it also adds one `git log -G` and one `git show`.
  - Cost scales with window size, which the shell itself keeps small on the quick path.
  - Memory: one window's numstat.
- **ARCH-SECURE:** the untrusted inputs are frontmatter (hand-editable), the surfaces file (repo-authored) and the `--flow` pin (agent-attested).
  - Each is parsed to a typed value at the boundary. A malformed flow or surfaces file resolves to the stricter flow instead of crashing or passing.
  - The pin is bounded by the shell, and editing the surfaces file counts as touching a surface. Push/merge instance validation catches a mistyped `flow:` via cue.
  - No credentials are involved.
- **ARCH-ORDER:** the flow is state held across events: change-code, close, milestone-close and REWORK.

  | State | change-code (no pin) | change-code `--flow X` | close/milestone-close: shell crossed | REWORK |
  |---|---|---|---|---|
  | absent | → inferred kind | → `{X, operator}` (quick refused if Mx) | treated as full: n/a | nothing written |
  | quick/inferred | re-infer; may become full | → `{X, operator}` | → full/inferred at finalize | nothing written |
  | quick/operator | stays | → `{X, operator}` | → full/inferred at finalize | nothing written |
  | full/inferred | stays (no downgrade) | → `{X, operator}` | stays | nothing written |
  | full/operator | stays | → `{X, operator}` | stays | nothing written |

  - Invariant: only an operator pin produces `quick` from `full`. The table is `Decide` and `Crossings` and is tested per cell.
  - Every mutation runs under the repo transaction lock. Close's two-lock split is untouched because the upgrade rides the existing finalize.
  - The event most likely to be mishandled is REWORK, covered in the decision above and by a test that closes with REWORK and re-closes.
- **ARCH-FUNERAL:**
  - The `flow:` line lives and archives with its issue.
  - The three ledger columns add width to rows that are already appended one per close, not new rows.
  - `.sdlc/shared-surfaces` is one repo-owned file, edited by its owner.
  - `small-diff-review.md` is static.
  - Nothing grows per launch or per session.

## Plan

### M1 — the flow record and change-code

Files:
- `cmd/sdlc/internal/flow/{flow,flow_test}.go`
- `cmd/sdlc/internal/issue/{plan,plan_test,structural,sizing,sizing_test,frontmatter_test}.go`
- `cmd/sdlc/{close,changecode,changecode_test,commitpathspec_guard_test}.go`
- `construct/vocabulary/issue.cue`
- `cmd/vocabulary/validate_test.go`
- `cmd/sdlc/helptext/{change-code,issue}.md`
- `atlas/workflow/{issue-lifecycle,sdlc-binary,vocabulary}.md`

- [ ] **Milestone parser, one home.**
  - Test first: `TestMilestonesInPlanOrder` in `internal/issue/plan_test.go` covers the em-dash form, the colon form, bold, ticked, fenced and duplicate rows. `TestComputeSizingCountsEmDashMilestones` in `sizing_test.go` must go red on today's colon-only regex.
  - Move `milestonePlanRE` + `milestonesInPlanOrder` into `internal/issue` as `MilestonesInPlanOrder`, and point `close.go` and `ComputeSizingFromContent` at it. Delete `milestoneLabelRE`.
  - Update `planItemMatchers` / `planItemBodySources` in `commitpathspec_guard_test.go`.
  - Run `go test ./cmd/sdlc/... -run 'Milestone|Sizing|PlanItem|Guard' -count=1`.
- [ ] **Flow codec + Decide.**
  - Test first: `TestFlowRoundTrip`; `TestFromFrontmatter` for absent, valid, malformed kind and non-map; `TestDecide` with one case per ARCH-ORDER cell plus both pin refusals.
  - Test first: `TestSetFieldRoundTripsFlowMap` in `internal/issue/frontmatter_test.go`, where `SetField` then `GetField` preserves the one-line map.
  - Implement `internal/flow/flow.go`, then run `go test ./cmd/sdlc/internal/flow/... ./cmd/sdlc/internal/issue/... -count=1`.
- [ ] **Model it in cue.**
  - Add `#Flow: {kind: "full" | "quick", provenance: "inferred" | "operator"}` and `flow?: #Flow` to `#Issue`.
  - Tests first in `cmd/vocabulary/validate_test.go`: a valid flow passes, and `flow: {kind: quikc, …}` fails with a `flow.kind` diagnostic.
  - Run `go test ./cmd/vocabulary/... -count=1` (needs `cue` on PATH) and `make vocab-embed`, which should show no `issue.json` diff because this is a definition-only change.
- [ ] **change-code infers, pins, records, skips.**
  - Pure tests first in `changecode_test.go`:
    - `TestChangeCodeFlowStep` covers inferred quick, full via Mx, full via plan file, operator pin, quick pin refused with Mx, and dry-run not writing.
    - `TestChangeCodeGatesSkipOnQuick`: every gate closure returns nil on quick with no estimate block.
    - `TestPlanGateContentIgnoresFlow` mirrors `TestPlanGateContentIgnoresEstimate`.
    - `TestFlowInfoLineNoGatesigCollision`.
  - Implement: add a `--flow` flag, a `changeCodeCtx.flow` field, the step after `changecode.go:130`, the closure guards, and the `planGateContent` strip.
  - Run `go test ./cmd/sdlc/... -run 'ChangeCode|PlanGate|GateOrder|ForceAck|Gatesig' -count=1`.
- [ ] **Docs for M1.**
  - `helptext/change-code.md`: the flow step, `--flow`, and which gates quick skips.
  - `helptext/issue.md`: the `flow` frontmatter field.
  - Atlas: `issue-lifecycle.md` (flow line + frontmatter template), `sdlc-binary.md` (change-code), `vocabulary.md` (`#Flow`).
  - Run `go test ./... -count=1`, then `git diff --check`.
- [ ] M1 — tick the rows above, then run `sdlc milestone-close --issue 231 --milestone M1 --verified '<test output>'`.

### M2 — the shell, the Done-when checks, and the small-diff review at close

Files:
- `cmd/sdlc/internal/flow/{shell,surfaces,donewhen}{,_test}.go`
- `cmd/sdlc/internal/churn/classify{,_test}.go`
- `cmd/sdlc/{close,close_atlasskip_test,publishgate,churnreport,milestoneclose,closereview_test,close_test,main}.go`
- `cmd/sdlc/internal/processmanual/gatesig.go`
- `cmd/sdlc/internal/judge/{architecture,prompts,review,judge_test,golden_test}.go`
- `cmd/sdlc/internal/judge/prompts/{milestone-review,small-diff-review}.md`
- `cmd/sdlc/internal/judge/boundary-tail.md`
- `.sdlc/shared-surfaces`
- `cmd/sdlc/helptext/{close,milestone-close,change-code}.md`
- `atlas/workflow/{gate-state,pre-merge-checks,architecture-principles,sdlc-binary}.md`
- `atlas/process-manual.md`, `atlas/workflow/process-manual.md` (regenerated)

- [ ] **One per-path classifier.**
  - Tests first in `churn/classify_test.go`:
    - `TestIsDoc` and `TestIsEmbedded`.
    - `TestIsCodeFile`: `cmd/**/*.md` is code, `README.md` isn't, `workshop/` and `atlas/` aren't, `tests/foo_spec.lua` isn't.
    - Extend `TestClassifyPath` with the non-Go test layouts.
  - Existing `TestHasCodePath` and the publish-gate tests must stay green unchanged.
  - Implement the classifiers and rewrite `hasCodePath` / `publishGateHasCodeSurface` as loops over them.
  - Run `go test ./cmd/sdlc/internal/churn/... ./cmd/sdlc/... -run 'Classify|CodePath|PublishGate|Churn' -count=1`.
- [ ] **Shell + surfaces.**
  - Tests first:
    - `TestCrossings`: 2 files / 100 lines stay quick; 3 files, 101 lines, one touched surface and one Mx row each cross; several crossings give several reasons.
    - `TestMeasure`: tests and docs are excluded, a binary counts as a file, and lines are summed only over code files.
    - `TestParseSurfaces`: comments, prefix entries, the bad-pattern error, and the self-match.
  - Implement `shell.go` and `surfaces.go`. Add ariadne's `.sdlc/shared-surfaces`:
    - `construct/vocabulary/*.cue`, `pkg/vocab/`, `construct/base.manifest`, `AGENTS.base.md`
    - `cmd/sdlc/internal/judge/architecture.md`, `cmd/sdlc/internal/judge/code-review.md`
    - `cmd/sdlc/internal/gatestate/`, `cmd/sdlc/internal/estimate/ledger.go`
- [ ] **Done-when checks.**
  - Tests first: `TestDoneWhenPresent`, where `related:` does not satisfy. `TestDoneWhenFresh` covers: Spec changed with Done-when unchanged refuses; both changed passes; neither changed passes; Log/Plan-only edits pass.
  - Implement `donewhen.go`, factoring `issue.HasDoneWhenBullet` out of `checkDoneWhen`.
- [ ] **close measures, upgrades, checks.**
  - Integration tests first in `closereview_test.go` on `closeRepo` fixtures that carry `flow:` and `## Done when`:
    - `TestCloseQuickWithinShellSelectsSmallDiff`: the stub judge sees the small-diff prompt, and `flow` stays quick.
    - `TestCloseQuickCrossingShellUpgrades`: `flow: {kind: full, provenance: inferred}` plus a Log reason, and the full prompt.
    - `TestCloseOperatorQuickStillUpgrades`.
    - `TestCloseQuickReworkThenReclose`: REWORK writes nothing, and the re-close upgrades again.
    - `TestCloseQuickEmptyDoneWhenRefuses`.
    - `TestCloseQuickStaleDoneWhenRefuses`, plus the skip via `--no-done-when-fresh`.
    - `TestCloseNoFlowUnchanged`: existing fixtures, full recipe, no new output.
    - `TestMilestoneCloseUpgradesQuick`.
  - Implement:
    - `windowFileStats`, shared with `churnForWindow`.
    - The step in `computeClose`, plus `closeResult.flow` / `.upgrade`.
    - The `--no-done-when-fresh` flag, its `skip()` key, and a `GateCatalog` row.
    - The upgrade in milestone mode.
    - Dry-run printing.
  - Run `go test ./cmd/sdlc/... -run 'Close|Milestone|GateCatalog|Skip' -count=1`.
- [ ] **The recipe.**
  - Tests first:
    - `TestArchitectureSectionsLossless`: preamble + sections re-join to the registry.
    - `TestArchitectureBlockForSubset`.
    - `TestSmallDiffRecipeCarriesExactlyThreeMarkers`: `markersIn(BuildPrompt(SmallDiffReview, in))` equals `SmallDiffMarkers`, and each marker exists in the registry.
    - `TestMilestoneReviewUnchanged`: the golden is byte-identical after `{{BOUNDARY_TAIL}}`.
    - `TestSmallDiffRecipeHasFocusAndContract`: family enumeration, the two-mode clause, doc-claim checks and `BoundaryReviewContract`.
  - Add `SmallDiffReview` to `AllInjectedCategories()` but not `AllCategories()`, since it is not a standalone `sdlc judge` check. Add its golden with `go test ./cmd/sdlc/internal/judge -run Golden -update` (confirm the flag name in `golden_test.go`), and extend `processmanual/collect_test.go`.
  - Implement `ArchitectureSection`, `ArchitectureBlockFor`, `ReviewMarkers`, `CodeReviewBody(in, markers)`, the `{{BOUNDARY_TAIL}}` include and `small-diff-review.md`.
  - Wire `boundaryReviewParams.Category`.
  - Run `go test ./cmd/sdlc/internal/judge/... ./cmd/sdlc/internal/processmanual/... -count=1`.
- [ ] **End to end.**
  - `TestQuickAndFullFlowsSameVerbSequence` uses `syncRepo`, `stubJudgeSeq` and a stubbed actual, and runs `executeSDLCTestCommand` for `claim`, `change-code --worktree=no` and `close --verified x` on two issues.
  - The small issue (1 code file, 20 lines, no plan) ends `flow: quick` and was reviewed by small-diff. The large one (3 code files) starts quick, is upgraded, and is reviewed by milestone-review.
  - Neither run passes a flow flag.
- [ ] **Help tokens + docs for M2.**
  - Add `{{QUICK_MAX_CODE_FILES}}` / `{{QUICK_MAX_CHANGED_LINES}}` to `renderLong`, used in `change-code.md` and `close.md`. `TestNoCommandLongHasSurvivingPlaceholder` pins the substitution.
  - `close.md` and `milestone-close.md` gain the shell, the upgrade, the Done-when checks and the recipe choice.
  - Atlas: `gate-state.md`, `pre-merge-checks.md`, `architecture-principles.md` (the subset recipe), `sdlc-binary.md` (close + `.sdlc/shared-surfaces`).
  - Regenerate the process manual with `sdlc process-manual`. Run `go test ./... -count=1` and `git diff --check`.
- [ ] M2 — tick, then run `sdlc milestone-close --issue 231 --milestone M2 --verified '<test output>'`.

### M3 — calibration, constitution, and the #263 fixture

Files:
- `cmd/sdlc/internal/estimate/{ledger,ledger_test,drift,drift_test}.go`
- `cmd/sdlc/{close,close_ledger_test}.go`
- `AGENTS.base.md` (then `make weave` for `AGENTS.md` / `CLAUDE.md`)
- `construct/local/sdlc/SKILL.md` (if it describes change-code's gates)
- `atlas/workflow/ledger-landscape.md`
- brain `data/life/42shots/velocity/SKILL.md` + TSV header note (peer)

- [ ] **Ledger columns.**
  - Tests first:
    - `TestRoundTripFlowColumns`.
    - `TestParseRowsKeepsChurnOnTwentyColumnRow`: widening the header must not zero the #187 block, so the width check is split per block at `ledger.go:133`.
    - `TestUpgradeHeaderAddsFlowColumns`.
    - In `close_ledger_test.go`, a quick close appends `quick`, `inferred`, `false`, and an upgraded close appends `full`, `inferred`, `true`.
  - Implement the `LedgerRow` fields `FlowKind`, `FlowProvenance` and `FlowUpgraded`, the header, `FormatRow` / `ParseRows`, and `appendCalibrationRow`.
- [ ] **Drift excludes quick and upgraded rows.**
  - Tests first: `TestDriftSampleExcludesQuickAndUpgraded`, and `TestDriftQuickRowLastDoesNotDisable`, where a trailing quick row with a blank model does not switch off the check for the next full close.
  - Implement the filters in `driftSample` and in the latest-row selection. `SpanThroughput` keeps quick hours, because they are real hours.
- [ ] **Constitution.**
  - `AGENTS.base.md` §2: "Non-trivial task (outside the quick-flow shell; `sdlc change-code --help` prints the limits)" replaces ">3 files or >100 lines".
  - One sentence on flow: the gates infer it, the operator pins it, and small work simply doesn't get milestones or a durable plan.
  - Avoid "estimate" within 80 chars of "start-plan" (`estimatetiming_test.go`). Run `make weave` and `go test ./... -count=1`.
- [ ] **#263 fixture run (manual, recorded in the Log).**
  - Render the small-diff prompt against parley.nvim#263's shipped range and run it once with the configured judge agent.
  - Record whether it surfaces the BR-1, BR-2 and BR-9 families, and how many ARCH lenses it spends.
  - This checks the recipe's aim; it is not a CI test, because the judge is an LLM.
- [ ] **Docs for M3.** `ledger-landscape.md` gets the three columns and the drift exclusion. Add a brain `velocity/SKILL.md` note on filtering by `flow_kind` for manual recalibration.
- [ ] M3 — tick, then run `sdlc close --issue 231 --verified '<evidence>'`.

## Follow-ups (not this issue)

- parley.nvim `.sdlc/shared-surfaces` (the keybinding registry, `config.lua` option schema): peer work in parley's tree.
- #191 (`exitWithCode` seam) is not needed here: the end-to-end test runs the gates' happy paths.
- #233 adds `kind: config`.
