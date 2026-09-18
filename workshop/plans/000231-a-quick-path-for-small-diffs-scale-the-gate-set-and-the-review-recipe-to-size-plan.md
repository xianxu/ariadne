# Quick flow implementation plan

> **For agentic workers:** Consult AGENTS.md Section 3 (Subagent Strategy). Execute with superpowers-executing-plans in the main session (the design rides on warm context); the SDLC gates own the fresh-context reviews. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Issues carry a `flow:` record that the gates infer and act on. A small change then runs claim → change-code → close with no durable plan, no plan-quality or estimate gates, and one small-diff review. Anything outside a hard size shell is upgraded to the full flow.

**Architecture:** One pure package, `cmd/sdlc/internal/flow`, owns the record, the inference, the shell limits and the Done-when checks. `change-code`, `start-plan` and `close`/`milestone-close` are thin IO shells around it. The review recipe reuses the existing boundary-review body, parametrized by the markers the ARCH registry marks `quick-flow: yes`.

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
| `flow.MaxCodeFiles`, `flow.MaxChangedLines`, `flow.ShellSummary` | `cmd/sdlc/internal/flow/limits.go` | new |
| `flow.ContractHashes` | `cmd/sdlc/internal/flow/donewhen.go` | new |
| `flow.DoneWhenPresent`, `flow.DoneWhenFresh` | `cmd/sdlc/internal/flow/donewhen.go` | new |
| `flow.Measure`, `flow.Crossings` | `cmd/sdlc/internal/flow/shell.go` | new |
| `flow.Surfaces`, `flow.ParseSurfaces` | `cmd/sdlc/internal/flow/surfaces.go` | new |
| `issue.MilestonesInPlanOrder` | `cmd/sdlc/internal/issue/plan.go` | modified (moved from `close.go`) |
| `issue.HasDoneWhenBullet` | `cmd/sdlc/internal/issue/structural.go` | new (factored out of `checkDoneWhen`) |
| `churn.IsDoc`, `churn.IsEmbedded`, `churn.IsCodeFile` | `cmd/sdlc/internal/churn/classify.go` | new |
| `churn.isTestPath` | `cmd/sdlc/internal/churn/classify.go` | modified (non-Go test layouts) |
| `judge.ArchitectureSection`, `judge.ArchitectureBlockFor` | `cmd/sdlc/internal/judge/architecture.go` | new (`ArchitectureBlock` delegates) |
| `judge.QuickMarkers`, `judge.ReviewMarkers` | `cmd/sdlc/internal/judge/architecture.go`, `prompts.go` | new |
| ARCH registry `quick-flow:` field | `cmd/sdlc/internal/judge/architecture.md` | modified (every entry declares `quick-flow: yes|no`) |
| `judge.SmallDiffReview` category | `cmd/sdlc/internal/judge/prompts.go` + `prompts/small-diff-review.md` | new |
| `judge.CodeReviewBody` | `cmd/sdlc/internal/judge/review.go` | modified (takes the marker set) |
| `startplan.planPointer`, `startplan.estimateNudge` | `cmd/sdlc/startplan.go` | modified (flow-conditional) |
| `estimate.LedgerRow` flow columns | `cmd/sdlc/internal/estimate/ledger.go` | modified |
| `estimate.driftSample` | `cmd/sdlc/internal/estimate/drift.go` | modified (excludes quick/upgraded) |

- **Flow** — `{Kind: full|quick, Provenance: inferred|operator, Spec, Done string}`, serialized on one line, e.g. `{kind: quick, provenance: inferred, spec: "1a2b3c4d", done: "5e6f7a8b"}`. `Format` always quotes the hashes: an unquoted all-digit hash is a YAML int to cue and yaml.v3 (PQ-9). Other unquoted hex shapes were measured to read as strings under both readers; the shared corpus pins the exact set (BR-12). It stays on one line because the issue frontmatter helpers are line-based. `Spec`/`Done` are optional, and change-code writes them for `quick` (see ContractHashes).
  - `FromFrontmatter(fm)` returns `(Flow, recorded bool, err)`. An absent field reads as full and not recorded. A malformed value is an error, which every caller resolves to `full`, the stricter flow.
  - **Relationships:** 1:1 with an issue. It is written by change-code and by the close finalize.
  - **DRY rationale:** one codec for every reader (change-code, start-plan, close, milestone-close, calibration). It reuses yaml/v3 as `project.DecodeMetadata` does.
  - **Future extensions:** #233 adds `Kind` `config`.
- **Decide** — `Decide(in DecideInput) (Flow, error)`, with `DecideInput{Recorded *Flow; Pin string; HasMilestones, HasPlan bool}`. Rules, in order:
  1. An unknown pin, or a `quick` pin with milestones, is an error.
  2. A pin sets `{pin, operator}`.
  3. With milestones the result is `{full, inferred}` whatever was recorded — unless the record is already full, which stays as it is (matching the ARCH-ORDER table) — because Mx rows cross the shell and a gate that finds the shell crossed upgrades regardless of provenance (PQ-4).
  4. A recorded operator flow stands.
  5. A recorded `full` stands, because gates never downgrade.
  6. Otherwise the result is `{full if HasPlan else quick, inferred}`.
- **Limits** — `MaxCodeFiles = 2`, `MaxChangedLines = 100`. `ShellSummary()` renders the one sentence that start-plan, the help tokens and error text all use.
  - **DRY rationale:** the numbers live once. The constitution points at `sdlc change-code --help` instead of restating them.
- **ContractHashes / DoneWhenPresent / DoneWhenFresh** — the Done-when checks are pure over the issue, with no git archaeology.
  - `ContractHashes(body) (spec, done string)` hashes the fence-aware section bodies. `spec` covers `## Spec` + `## Revisions`, because the constitution records a mid-stream reframe by appending Revisions (PQ-2). `done` covers `## Done when`. Each hash is the 8-hex prefix of sha256 over whitespace-trimmed text.
  - change-code writes both into the quick record on every run, so the anchor moves whenever the contract is re-fixed (PQ-6).
  - `DoneWhenPresent(body) error` requires a Done-when bullet; `related:` does not count, because this is the review's oracle.
  - `DoneWhenFresh(rec Flow, body) error` refuses when the recorded `spec` differs from the current one and `done` does not. A record without hashes means "no anchor" and refuses with the fix named: re-run `sdlc change-code`, or pass `--no-done-when-fresh`.
- **Measure / Crossings** — `Measure(files []string, stats []churn.FileStat, s Surfaces, milestones []string) Size` counts code files (via `churn.IsCodeFile`), the insertions in those files, the shared surfaces touched and the Mx rows. `Crossings(Size) []string` returns one reason per crossed limit, or nil.
- **Surfaces** — parsed from `.sdlc/shared-surfaces`: one `path.Match` pattern per line, `#` comments, and a trailing `/` meaning "anything under". The declaration file always matches itself.
  - Close parses the committed file at the window base and at HEAD and takes the union. A branch can add a surface but cannot remove one from its own shell, and an uncommitted edit is never read (PQ-5).
  - A malformed line is an error, which close turns into a crossing, so it fails toward `full`.
- **churn classifiers** — `IsDoc` is #177's per-path docs rule and `IsEmbedded` is #174's `cmd/` rule. `IsCodeFile(p) = ClassifyPath(p) == CodeProd && (!IsDoc(p) || IsEmbedded(p))`. `hasCodePath` and `publishGateHasCodeSurface` become `any()` loops over these, giving one per-path rule with three readings.
  - `isTestPath` widens to the non-Go layouts the fleet uses: `*_spec.lua`, `*_test.lua`, `*_test.py`, `test_*.py`, `*.test.*`, `*.spec.*`, and path segments `test`, `tests`, `spec`, `__tests__`, `testdata`.
  - That shifts what `churn_prod`/`churn_test` mean for non-Go repos from this issue onward. `ledger-landscape.md` and the Log record the cut-over (PQ-7).
- **ArchitectureSection / ArchitectureBlockFor / QuickMarkers** — the registry becomes addressable per entry.
  - `ArchitectureSection` is a production slicer, promoted from the `architectureEntry` test helper. It splits the registry losslessly into a preamble and per-marker sections, and reports an entry whose `- **quick-flow:**` field is missing or is not `yes`/`no` as an error. So a new principle cannot be added without deciding.
  - `QuickMarkers()` returns, in registry order, the markers marked `yes`. **This is where the quick flow's principles are chosen.**
  - `ArchitectureBlockFor(lens, markers)` renders the header, the preamble and the selected sections verbatim, including the field, since it tells a reader which flow checks the principle. `ArchitectureBlock(lens)` delegates with all markers, so it remains "the registry verbatim".
  - `ReviewMarkers(c)` returns `QuickMarkers()` for `SmallDiffReview` and all markers otherwise.

### Integration points

| Name | Lives in | Status | Wraps |
|------|----------|--------|-------|
| change-code flow step | `cmd/sdlc/changecode.go` | modified | issue file read/write, plan-file lookup |
| start-plan guidance | `cmd/sdlc/startplan.go` | modified | stdout |
| close shell + Done-when step | `cmd/sdlc/close.go` (`computeClose`) | modified | `git diff --numstat`, `git show <rev>:.sdlc/shared-surfaces` |
| recipe selection | `cmd/sdlc/milestoneclose.go` (`boundaryReviewParams`, `boundaryReviewDispatchOptions`) | modified | judge subprocess (`judge.Run` seam) |
| calibration row | `cmd/sdlc/close.go` (`appendCalibrationRow`) | modified | brain TSV append |
| help tokens | `cmd/sdlc/main.go` (`renderLong`) | modified | embedded helptext |

- **change-code flow step** — runs after the issue and plan are read (`changecode.go:130`) and before the gate loop.
  - Inputs: `issue.MilestonesInPlanOrder(PlanItemsBody)`, and `planContent != ""`, which is the plan lookup plan-quality uses.
  - `reportChangeCodeFlow` decides and prints the flow before the gates. `recordChangeCodeFlow` writes `flow:` (plus the hashes for quick) with `Parse → SetField → Compose` only after the gates pass and past the dry-run return, right before the sync commit. A refused run leaves the issue untouched (BR-5).
  - On `quick` the gate loop iterates `activeChangeCodeGates`, which is empty, rather than a skip in each closure. `changeCodeGates` stays the complete declaration, and `changeCodeGateOrder()` still returns all five.
  - `planGateContent` strips `flow:` alongside `estimate_hours:`.
- **start-plan guidance** — `planPointer` and `estimateNudge` become conditional on the shell. The pointer says to write a durable plan only for work outside the shell, and that change-code infers quick otherwise. The nudge says there is no estimate on the quick flow.
- **close shell + Done-when step** — runs inside `computeClose` after the atlas block (~`close.go:530`), on the window already computed there (`windowBase`, `windowHead`, `diffFiles`).
  - Lines come from `churn.ParseNumstat(git diff --numstat base..head)`, via `windowFileStats(base, head)`, which `churnForWindow` then reuses.
  - On a crossing, the upgrade goes into `newFM` plus a Log reason, and `applyClose` writes it at finalize.
  - The Done-when checks run when the recorded kind is `quick`, including when this close upgrades it.
- **recipe selection** — `boundaryReviewParams.Category` is set from `closeResult.flow` at `close.go:1024/:1037/:1078`. milestone-close always passes `MilestoneReview`.
- **Test seams (existing, reused):** `judge.Run`, real git in `testfix.Repo`/`syncRepo` temp repos, `computeActualForCloseFn`, `die`/`expectDie`, `executeSDLCTestCommand`, and `TestMain`'s real-repo guard.

## Decisions and architecture

- **Changed lines** are insertions in code files — git's `+` count, the same number the churn report and ledger use. Files are counted from `--name-only`, so a binary change still counts as a file.
- **The upgrade is recorded at finalize.** `computeClose` writes nothing, and a REWORK leaves the issue unwritten (#139). So "no gate downgrades" is about the *recorded* flow. The measurement is a pure function of the window, so a re-close after REWORK re-derives it.
- **Every surface that routes work into a durable plan becomes flow-conditional (PQ-1).**
  - The directive surfaces were derived with `git grep -n -E 'superpowers-writing-plans|durable plan|writing-plans skill|plan-quality|estimate_hours'`, over agent-facing files outside `workshop/`:
    - `AGENTS.base.md:11,23,24,25`
    - `cmd/sdlc/startplan.go` (framing `:67-71`, `planPointer :191-204`, `estimateNudge :220-233`)
    - `cmd/sdlc/helptext/start-plan.md:3,13,28,34-35,46`
    - `cmd/sdlc/helptext/root.md:16-19`
    - `cmd/sdlc/helptext/change-code.md`
    - `cmd/sdlc/helptext/issue.md:37,46`
    - `cmd/sdlc/helptext/set-status.md:15`
    - `construct/adapted/superpowers-brainstorming/SKILL.md:43,60,76,149` (with a Conversation entry in `construct/intents/superpowers.md`, per lessons)
  - The descriptive atlas mentions (`artifact-hierarchy.md:12,21`, `issue-lifecycle.md:6,35`, `sdlc-binary.md:35,1092`) go with each milestone's docs step.
  - This moves into M1: the quick flow must be reachable on the documented path before the trial.
- **The shared-surface declaration** is a plain line file, `.sdlc/shared-surfaces`. It is repo-owned like `.sdlc/fleet.json` and is not woven. ariadne gets a starter declaration; parley.nvim's is peer follow-up work.
- **Refusals the quick flow keeps:** `change-code --flow quick` on a Plan with Mx rows, and at close an empty Done-when and a stale Done-when (skip with `--no-done-when-fresh`). Crossing the shell never refuses.
- **ARCH-DRY:**
  - Which principles the quick flow checks is a registry field beside each principle's lenses, not a Go list.
  - One flow codec; one milestone parser (the colon-only `milestoneLabelRE` retires); one per-path classifier set; one review body parametrized by markers; one shared boundary-review tail (`{{BOUNDARY_TAIL}}`, expanded at template load).
  - One set of limits, read by start-plan, help, error text and the constitution's pointer.
  - Plan lookup reuses `readOptionalPlanFile`, the window reuses `boundaryWindowBase`, and numstat is shared with `churnForWindow`.
- **ARCH-PURE:** every decision lives in `internal/flow`, `internal/churn` and `internal/judge`, and is tested on strings. The IO is four thin sites: the change-code step, start-plan output, the close measurement (git + `git show` of the declaration) and the ledger append.
- **ARCH-PURPOSE:** every consumer derives from the record and the limits: change-code, start-plan, the constitution and skills, close, milestone-close, recipe selection, calibration, help and atlas. Test detection widens to non-Go layouts because the motivating case is Lua.
- **ARCH-MOCK:** no new external dependency. git runs for real in temp repos, and the judge is faked at `judge.Run`.
- **ARCH-CONSTRAINTS:** a CLI batch step at close. It adds one `git diff --numstat` over a window already diffed by `--name-only`, plus two `git show`s of one small file. Cost scales with the window, which the shell keeps small on the quick path. There is no UI path.
- **ARCH-SECURE:**
  - Frontmatter is hand-editable, so `flow:` is parsed to a typed value. Malformed resolves to full, and cue validation at push/merge catches typos.
  - The surfaces file is read only as committed, at base ∪ head, so a branch cannot loosen its own shell.
  - The `--flow` pin is agent-attested and bounded by the shell.
  - The contract hashes are integrity hints, not security. Hand-editing them only weakens the agent's own freshness check, and the full review still runs if the shell is crossed.
  - No credentials are involved.
- **ARCH-ORDER:** the flow is state held across events: change-code, close, milestone-close and REWORK.

  | State | change-code (no pin) | change-code `--flow X` | Mx rows present at change-code | close/milestone-close: shell crossed | REWORK |
  |---|---|---|---|---|---|
  | absent | → inferred kind | → `{X, operator}` | → full/inferred (quick pin refused) | treated as full | nothing written |
  | quick/inferred | re-infer; hashes refreshed | → `{X, operator}` | → full/inferred | → full/inferred at finalize | nothing written |
  | quick/operator | stays; hashes refreshed | → `{X, operator}` | → full/inferred (quick pin refused) | → full/inferred at finalize | nothing written |
  | full/inferred | stays | → `{X, operator}` | stays | stays | nothing written |
  | full/operator | stays | → `{X, operator}` | stays | stays | nothing written |

  - Invariant: only an operator pin produces `quick` from `full`, and nothing produces `quick` while Mx rows exist.
  - Each cell is a `TestDecide` / `TestCrossings` case. Mutations run under the repo lock, and close's two-lock split is untouched.
  - The event most likely to be mishandled is REWORK, which gets its own integration test.
- **ARCH-FUNERAL:**
  - The `flow:` line lives and archives with its issue.
  - The ledger columns widen existing per-close rows.
  - `.sdlc/shared-surfaces` is one repo-owned file.
  - The prompt file is static.
  - Nothing grows per launch or per session.

## Plan

### M1 — the flow record, change-code, and a reachable quick path

Files:
- `cmd/sdlc/internal/flow/{flow,limits,donewhen}{,_test}.go`
- `cmd/sdlc/internal/issue/{plan,plan_test,structural,sizing,sizing_test,frontmatter_test}.go`
- `cmd/sdlc/{close,changecode,changecode_test,planfence_test,close_test,commitpathspec_guard_test,startplan,startplan_test}.go`
- `construct/vocabulary/issue.cue`
- `cmd/vocabulary/validate_test.go`
- `cmd/sdlc/helptext/{change-code,issue,start-plan,root,set-status}.md`
- `AGENTS.base.md` (then `make weave`)
- `construct/adapted/superpowers-brainstorming/SKILL.md`, `construct/intents/superpowers.md`
- `atlas/workflow/{issue-lifecycle,sdlc-binary,vocabulary,artifact-hierarchy}.md`

- [x] **One milestone parser.**
  - Tests first, in `internal/issue`:
    - `TestMilestonesInPlanOrder` covers the em-dash, colon, bold, ticked, fenced and duplicate forms.
    - `TestComputeSizingCountsEmDashMilestones` goes red on today's colon-only regex.
  - Move `milestonePlanRE` and `milestonesInPlanOrder` into `internal/issue`, and point `close.go` and `ComputeSizingFromContent` at them. Delete `milestoneLabelRE`.
  - Update the references in `planfence_test.go:79,411`, `close_test.go:444` and `commitpathspec_guard_test.go` (`planItemMatchers`, `planItemBodySources`).
  - Run `go test ./cmd/sdlc/... -run 'Milestone|Sizing|PlanItem|Guard|PlanFence' -count=1`.
- [x] **Flow codec, limits, Decide, contract hashes.**
  - Tests first:
    - `TestFlowRoundTrip`, seeded with an all-digit hash and an `NNeNNNNN` hash (PQ-9).
    - `TestFromFrontmatter`: absent, valid, and a malformed kind.
    - `FuzzFromFrontmatter`: seeded with nested, duplicate-key, quoted and truncated maps. It never panics, and every error path resolves to full.
    - `TestDecide`: one case per ARCH-ORDER cell, plus the pin refusals.
    - `TestContractHashes`: a Revisions-only edit changes `spec`, a Log/Plan-only edit changes neither, and a `## Done when` inside a code fence is not the section.
    - `TestDoneWhenPresent`, where `related:` does not satisfy it.
    - `TestDoneWhenFresh`: Spec changed with Done unchanged refuses; a Revisions-only change refuses; both changed passes; no hashes refuses with the fix named.
    - `TestSetFieldRoundTripsFlowMap`.
  - Implement `flow.go`, `limits.go`, `donewhen.go`, and factor out `issue.HasDoneWhenBullet`.
  - Run `go test ./cmd/sdlc/internal/flow/... ./cmd/sdlc/internal/issue/... -count=1`, and `go test ./cmd/sdlc/internal/flow -fuzz FuzzFromFrontmatter -fuzztime 20s`.
- [x] **Model it in cue.**
  - Add `#Flow: {kind: "full" | "quick", provenance: "inferred" | "operator", spec?: =~"^[0-9a-f]{8}$", done?: =~"^[0-9a-f]{8}$"}` and `flow?: #Flow`.
  - Tests first in `cmd/vocabulary/validate_test.go`: a valid flow with a quoted all-digit hash passes; a bad kind fails on `flow.kind`; a bad hash fails.
  - Run `go test ./cmd/vocabulary/... -count=1` and `make vocab-embed`, expecting no `issue.json` diff.
- [x] **change-code infers, pins, records, skips.**
  - Tests first:
    - `TestChangeCodeFlowStep`: inferred quick writes hashes; full via Mx; full via plan; operator pin; quick pin refused with Mx; quick/operator plus a new Mx row becomes full; dry-run doesn't write.
    - `TestChangeCodeGatesSkipOnQuick`.
    - `TestPlanGateContentIgnoresFlow`.
    - `TestFlowInfoLineNoGatesigCollision`.
  - Implement the `--flow` flag, `changeCodeCtx.flow`, the step after `changecode.go:130`, the closure guards and the `planGateContent` strip.
  - Run `go test ./cmd/sdlc/... -run 'ChangeCode|PlanGate|GateOrder|ForceAck|Gatesig' -count=1`.
- [x] **Make the quick flow reachable (PQ-1).**
  - Tests first in `startplan_test.go`:
    - `TestPlanPointerConditionalOnShell`: names `flow.ShellSummary()`, says to write a durable plan only outside it, and still names `superpowers-writing-plans` for that case.
    - `TestEstimateNudgeMentionsQuick`.
  - Update `startplan.go` and every directive surface in the Decisions list, including `AGENTS.base.md` §2 ("Non-trivial task (outside the quick-flow shell; `sdlc change-code --help` prints the limits)"), plus one flow sentence.
  - The brainstorming skill's "invoke writing-plans" step becomes conditional, with a Conversation entry and Verify clauses in `construct/intents/superpowers.md`.
  - Run `make weave`, then `go test ./cmd/sdlc/... -run 'StartPlan|EstimateTiming' -count=1`. `estimatetiming_test.go` forbids "estimate" within 80 chars of "start-plan".
- [x] **Docs for M1.**
  - `helptext/change-code.md` (flow step, `--flow`, skipped gates) and `helptext/issue.md` (the `flow` field).
  - Atlas: `issue-lifecycle.md`, `sdlc-binary.md`, `vocabulary.md`, `artifact-hierarchy.md`.
  - Run `go test ./... -count=1`, then `git diff --check`.
- [ ] M1 — tick, then run `sdlc milestone-close --issue 231 --milestone M1 --verified '<test output>'`.

### M2 — the shell, the Done-when checks, and the small-diff review at close

Files:
- `cmd/sdlc/internal/flow/{shell,surfaces}{,_test}.go`
- `cmd/sdlc/internal/churn/classify{,_test}.go`
- `cmd/sdlc/{close,publishgate,churnreport,milestoneclose,closereview_test,close_test,main}.go`
- `cmd/sdlc/internal/processmanual/{gatesig.go,collect_test.go}`
- `cmd/sdlc/internal/judge/{architecture,prompts,review,judge_test,golden_test}.go`
- `cmd/sdlc/internal/judge/architecture.md`, `cmd/sdlc/internal/judge/boundary-tail.md`
- `cmd/sdlc/internal/judge/prompts/{milestone-review,small-diff-review}.md`
- `cmd/sdlc/internal/judge/testdata/golden/*`
- `.sdlc/shared-surfaces`
- `cmd/sdlc/helptext/{close,milestone-close,change-code}.md`
- `atlas/workflow/{gate-state,pre-merge-checks,architecture-principles,sdlc-binary}.md`
- `atlas/process-manual.md`, `atlas/workflow/process-manual.md` (regenerated)

- [ ] **One per-path classifier.**
  - Tests first:
    - `TestIsDoc`.
    - `TestIsEmbedded`.
    - `TestIsCodeFile`: `cmd/**/*.md` is code; `README.md`, `workshop/`, `atlas/` and `tests/x_spec.lua` aren't.
    - Extend `TestClassifyPath` with the non-Go layouts.
  - `TestHasCodePath` and the publish-gate tests must stay green unchanged.
  - Implement the classifiers and reduce the two old classifiers to loops over them.
  - Run `go test ./cmd/sdlc/internal/churn/... ./cmd/sdlc/... -run 'Classify|CodePath|PublishGate|Churn' -count=1`.
- [ ] **Shell + surfaces.**
  - Tests first:
    - `TestCrossings`: 2 files / 100 lines stay quick; 3 files, 101 lines, one surface and one Mx row each cross; several crossings give several reasons.
    - `TestMeasure`: tests and docs excluded, a binary counts as a file, lines summed over code files only.
    - `TestParseSurfaces`.
    - `FuzzParseSurfaces`: never panics, and a bad pattern yields an error.
    - `TestSurfacesUnion`: removing a pattern on the branch still matches.
  - Implement `shell.go` and `surfaces.go`, and add ariadne's `.sdlc/shared-surfaces`:
    - `construct/vocabulary/*.cue`, `pkg/vocab/`, `construct/base.manifest`, `AGENTS.base.md`
    - `cmd/sdlc/internal/judge/architecture.md`, `cmd/sdlc/internal/judge/code-review.md`
    - `cmd/sdlc/internal/gatestate/`, `cmd/sdlc/internal/estimate/ledger.go`
- [ ] **close measures, upgrades, checks.**
  - Integration tests first in `closereview_test.go`, on `closeRepo` fixtures that carry `flow:` and `## Done when`:
    - `TestCloseQuickWithinShellSelectsSmallDiff`.
    - `TestCloseQuickCrossingShellUpgrades`.
    - `TestCloseOperatorQuickStillUpgrades`.
    - `TestCloseQuickReworkThenReclose`.
    - `TestCloseQuickEmptyDoneWhenRefuses`.
    - `TestCloseQuickStaleDoneWhenRefuses`, plus its `--no-done-when-fresh` skip.
    - `TestCloseSurfacesUncommittedEditIgnored`.
    - `TestCloseNoFlowUnchanged`.
    - `TestMilestoneCloseUpgradesQuick`.
  - Implement:
    - `windowFileStats`, shared with `churnForWindow`.
    - The step in `computeClose`, plus `closeResult.flow` / `.upgrade`.
    - `--no-done-when-fresh`, its `skip()` key and a `GateCatalog` row.
    - The milestone-mode upgrade.
    - Dry-run printing.
  - Run `go test ./cmd/sdlc/... -run 'Close|Milestone|GateCatalog|Skip' -count=1`.
- [ ] **The recipe.**
  - Add `- **quick-flow:** yes` to ARCH-DRY, ARCH-PURE and ARCH-PURPOSE, and `- **quick-flow:** no` to the other five.
  - Tests first:
    - `TestArchitectureSectionsLossless`.
    - `TestEveryPrincipleDeclaresQuickFlow`, where a fixture that is missing the field, or says `maybe`, errors.
    - `TestQuickMarkers`: DRY, PURE, PURPOSE, in order.
    - `TestArchitectureBlockForSubset`.
    - `TestSmallDiffRecipeCarriesExactlyQuickMarkers`: `markersIn(BuildPrompt(SmallDiffReview, in))` equals `QuickMarkers()`.
    - `TestMilestoneReviewRendersViaBoundaryTail`: the same text as before, apart from the new registry lines.
    - `TestSmallDiffRecipeHasFocusAndContract`: family enumeration, the two-mode clause, doc-claim checks and `BoundaryReviewContract`.
  - Add `SmallDiffReview` to `AllInjectedCategories()` only. Regenerate goldens with `go test ./cmd/sdlc/internal/judge -run Golden -update-golden` and review the diff: only the new `quick-flow:` lines and the new category should change. Extend `processmanual/collect_test.go`.
  - Implement the registry functions, `CodeReviewBody(in, markers)`, the `{{BOUNDARY_TAIL}}` include, `small-diff-review.md` and `boundaryReviewParams.Category`.
  - Run `go test ./cmd/sdlc/internal/judge/... ./cmd/sdlc/internal/processmanual/... ./cmd/sdlc/... -run 'Arch|Prompt|Golden|StartPlan|Collect' -count=1`.
- [ ] **End to end.**
  - `TestQuickAndFullFlowsSameVerbSequence` uses `syncRepo`, `stubJudgeSeq` and a stubbed actual, and runs `executeSDLCTestCommand` for `claim`, `change-code --worktree=no` and `close --verified x` on two issues.
  - The small issue (1 code file, 20 lines, no plan) ends quick with the small-diff review. The large one (3 code files) ends upgraded with milestone-review.
  - Neither passes a flow flag.
- [ ] **Help tokens + docs for M2.**
  - `{{QUICK_SHELL}}` in `renderLong`, reading `flow.ShellSummary()`, used in `change-code.md` and `close.md`. `TestNoCommandLongHasSurvivingPlaceholder` pins it.
  - `close.md` and `milestone-close.md` gain the shell, upgrade, Done-when checks and recipe choice.
  - Atlas: `gate-state.md`, `pre-merge-checks.md`, `architecture-principles.md` (the `quick-flow:` field), `sdlc-binary.md` (close + `.sdlc/shared-surfaces`).
  - Record the churn test-bucket cut-over in the issue Log.
  - Regenerate the process manual. Run `go test ./... -count=1` and `git diff --check`.
- [ ] M2 — tick, then run `sdlc milestone-close --issue 231 --milestone M2 --verified '<test output>'`.

### M3 — calibration and the #263 fixture

Files:
- `cmd/sdlc/internal/estimate/{ledger,ledger_test,drift,drift_test}.go`
- `cmd/sdlc/{close,close_ledger_test}.go`
- `atlas/workflow/ledger-landscape.md`
- brain `data/life/42shots/velocity/SKILL.md` (peer)

- [ ] **Ledger columns.**
  - Tests first:
    - `TestRoundTripFlowColumns`.
    - `TestParseRowsKeepsChurnOnTwentyColumnRow`: split the width check per block at `ledger.go:129`.
    - `TestUpgradeHeaderAddsFlowColumns`.
    - `close_ledger_test.go` quick and upgraded rows.
  - Implement the `LedgerRow` fields `FlowKind`, `FlowProvenance` and `FlowUpgraded`, the header, `FormatRow`/`ParseRows`, and `appendCalibrationRow`.
- [ ] **Drift excludes quick and upgraded rows.**
  - Tests first: `TestDriftSampleExcludesQuickAndUpgraded` and `TestDriftQuickRowLastDoesNotDisable`.
  - Filter in `driftSample` and in the latest-row selection. `SpanThroughput` keeps quick hours.
- [ ] **#263 fixture run (manual, recorded in the Log).** Render the small-diff prompt against parley.nvim#263's shipped range and run it once. Record whether it surfaces the BR-1, BR-2 and BR-9 families, and how many ARCH lenses it spends.
- [ ] **Docs for M3.** `ledger-landscape.md` gets the three columns, the drift exclusion and the churn cut-over. The brain `velocity/SKILL.md` gets a filtering note.
- [ ] M3 — tick, then run `sdlc close --issue 231 --verified '<evidence>'`.

## Follow-ups (not this issue)

- parley.nvim `.sdlc/shared-surfaces`: peer work in parley's tree.
- `sdlc propagate-base` to carry the `AGENTS.base.md` change to downstream repos: the operator's call.
- #233 adds `kind: config`.

## Revisions

### 2026-09-17 — the quick principles live in the registry

Reason: the operator asked where to change which principles the quick flow
checks, and chose the registry over a Go list.

Delta: `judge.SmallDiffMarkers` (a Go var) is replaced by a `quick: yes|no` field
on every entry of `judge/architecture.md`, read by `QuickMarkers()`. A missing or
invalid field is an error, and the renderer strips the field so existing goldens
stay byte-identical. M2's recipe tests follow the registry instead of a fixed
list of three.

### 2026-09-17 — plan-quality round 1 (PQ-1..PQ-8)

Reason: the plan-quality judge raised two Important and six Minor findings.

Delta:

- **PQ-1.** Every directive surface that routes work into a durable plan
  becomes flow-conditional. They are derived by a `git grep` sweep and listed in
  Decisions. The work moves into M1, because the quick flow must be reachable
  before the trial. `flow.ShellSummary()` is their single source.
- **PQ-2 and PQ-6.** Freshness no longer digs through git. change-code records
  `spec`/`done` hashes in the quick record on every run, and `spec` covers Spec +
  Revisions.
- **PQ-3.** The registry field is renamed `quick-flow:` and is not stripped at
  render. That keeps the "registry verbatim" tests true, and goldens change by the
  new lines only.
- **PQ-4.** `Decide` gives Mx rows precedence over a recorded operator quick.
- **PQ-5.** Surfaces are read as committed at base ∪ head.
- **PQ-7.** The churn cut-over is documented.
- **PQ-8.** Every test is named, and fuzz targets cover the two hand-edited
  parsers.

### 2026-09-17 — advisory findings PQ-8, PQ-9, PQ-10

- **PQ-9.** `Format` quotes the hashes, and the round-trip, fuzz and cue tests
  are seeded with all-digit and exponent-shaped hashes. The rule: a persisted
  value must parse to the same type under every reader.
- **PQ-10.** The issue's §4 and Done-when were restated in the same edit.
- **PQ-8** is kept deliberately. The per-test case lists are the red-first
  oracle the implementer writes against, and the boundary review checks them
  against the diff. Collapsing them to one-line strategies would move that oracle
  into the implementer's head. The two hand-edited parsers get fuzz targets, which
  was the substantive half of the finding.

### 2026-09-17 — M1 boundary review round 1 (BR-4..BR-10)

- **BR-4.** The guard now derives callers of `issue.MilestonesInPlanOrder` by
  listing it in `planItemMatchers`, replacing the hand-kept
  `planItemBodySources` that missed change-code. There is also a fenced-Mx case
  that must stay quick.
- **BR-5.** The record is decided before the gates and written after them
  (`reportChangeCodeFlow` / `recordChangeCodeFlow`), pinned by
  `TestRunChangeCodeRecordsFlowAfterGates`.
- **BR-6.** The M1 docs describe close behaviour that M2 builds. That holds
  because M2 lands before any merge; logged.
- **BR-7.** `Decide` returns the `Rule` that fired, and `flowReason` is gone.
- **BR-8.** `SetField` replaces a block-form value whole. It is also literal now,
  so there is no `$` expansion.
- **BR-9.** The fuzz oracle asserts that error paths resolve to full.
- **BR-10.** The `cue fmt` churn in `issue.cue` is reverted.
- **Test names as built:** `TestDecideChangeCodeFlow`,
  `TestRecordChangeCodeFlowWritesUnlessDryRun`, `TestActiveChangeCodeGates` and
  `TestRunChangeCodeRecordsFlowAfterGates`, replacing `TestChangeCodeFlowStep`
  and `TestChangeCodeGatesSkipOnQuick`. `flow.WithContract` lives in
  `donewhen.go`.

### 2026-09-17 — M1 boundary review round 2 (BR-11..BR-14), fixed as rules

- **BR-11** (write from a stale snapshot). `recordChangeCodeFlow` re-derives
  the record from a fresh read at write time. It refuses if the edit changed the
  flow itself (`flowDrift`). Both paths are tested with a second writer.
- **BR-12** (a record round-trips through every reader). One corpus,
  `construct/vocabulary/testdata/flow_records.txt`, is asserted by the Go codec
  (`TestFlowRecordCorpus`) and by cue (`TestValidateInstance_FlowRecordCorpus`).
  `Parse` rejects a present-but-empty hash. The corpus also corrected a claim:
  unquoted `12e45678` reads as a string under both readers, not a float.
- **BR-13** (a doc claim contradicts the code). A verb sweep over the gates'
  action words, `git grep -n -i -E
  'change-code.{0,60}(refus|requir|demand|parses|asks? for|runs)|(refus|requir|demand).{0,60}change-code|estimate gate|plan-quality gate'`
  over helptext, atlas, `AGENTS.base.md` and construct, found 26 hits. Three
  were unqualified (`estimate.md:5`, `claim.md:17`, `issue.md:93`), and all three
  are now qualified.
- **BR-14** (decision logic restated). Only `package flow` builds flow values,
  pinned by `TestOnlyFlowPackageBuildsFlowValues`. A missing frontmatter
  refuses, and a malformed record is resolved by `flow.Recorded`.

