# Boundary Review — ariadne#187 (milestone M1)

| field | value |
|-------|-------|
| issue | 187 — tune the change-code gate: stateful plan review, estimate after plan, churn metric |
| repo | ariadne |
| issue file | workshop/issues/000187-tune-change-code-gate.md |
| boundary | milestone M1 |
| milestone | M1 |
| window | f59f49cb2bc7f9d165f1f7b8ea962954424c5135^..HEAD |
| command | sdlc milestone-close --issue 187 --milestone M1 |
| reviewer | claude |
| timestamp | 2026-07-29T12:26:04-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The stateful plan gate is well-built: `gatestate` is genuinely pure (no fs/clock/subprocess, every convergence test runs on in-memory strings with zero mocks), the vocabulary is single-sourced in CUE and actually derived by the prompt/parser/decision, the two fuzz targets are the repo's first and one already caught a real yaml/v3 emitter bug, and the demotion's safety argument is wired — `code-review.md`'s carry-forward reached me verbatim in this prompt. What blocks SHIP is one silent regression and two undelivered Task 9 surfaces: collapsing the five per-gate `--force` messages into one templated string (`changecode.go:137`) renamed two bypass ACKs that `processmanual/gatesig.go` matches by literal, so forced `structural` / `estimate-recon` bypasses are now invisible to the friction classifier — degrading the accepted-vs-forced instrumentation #187 D3 is built on. Separately, the `ContentHash` pass-through does not achieve the purpose asserted for it in three places, because the estimate-gate retry by construction edits the issue text the hash covers. And `helptext/change-code.md` still prints the pre-B1 gate order while `helptext/start-plan.md` still says "this is where you set the estimate" — both named in Task 9's file list, under a commit that claims "across every surface that tells it." **Caveat: Bash is unavailable in this session (harness EPERM on `~/.claude/session-env`), so I could not run `go test` / `go vet`; every finding below is from reading the files.**

## 1. Strengths

- **`ARCH-PURE` is exemplary here.** `cmd/sdlc/internal/gatestate/` touches no filesystem, clock, or subprocess; `planreview.go` is the single seam and says so. The result is that `Decide`, `OpenFindings`, `ParseFindingsBlock`, `Render`/`ParseSidecar` and `RenderPriorFindings` are all tested without a mock in sight — the claim in the package doc is actually true of the code.
- **The blocking decision genuinely moved out of the LLM.** `decide.go:42-79` reads the accumulated ledger, not the verdict token, and `finding.cue`'s `dispositions.closing`/`open` partition means `closedSet` (`ledger.go:173-182`) can't acquire an unhandled case. `TestFindingConformance` derives every assertion from the model rather than maintaining a list.
- **The two fuzz targets earn their place.** `FuzzRenderParseRoundTrip` (`render_test.go:105`) pins "Render never emits a document ParseSidecar can't read" *and* canonical-form-is-a-fixed-point, with the crasher corpus committed (`testdata/fuzz/FuzzRenderParseRoundTrip/417cf3fd96f47e3d`). `TestParseFindingsBlockBlockScalarSurvivesHash` (`parse_test.go:157`) pins both halves of the ` #`-truncation hazard, so the block-scalar template can't be "simplified" back.
- **`ARCH-DRY` extractions are the right ones:** `sidecarPathFor` (`reviewsidecar.go:36`), `classifyFallback` (`changecode.go:536`), `contains` promoted into `lifecycle.go:21`, and `RenderBlockInstruction` placed as a `FindingModel` method mirroring `VerdictModel` rather than split across packages.
- **Gates-as-data is a real guard, not a restatement.** `runChangeCode` iterates the same `changeCodeGates` literal that `changeCodeGateOrder` reads (`changecode.go:130`, `:220`), so reordering the literal reorders execution — `TestGateOrderPlanBeforeEstimate` fails on the regression it exists to catch.

## 2. Critical findings

**C1 — `cmd/sdlc/changecode.go:137`: the consolidated `--force` message silently kills two friction-classifier ACK patterns.**

The five hand-written bypass messages became one template: `fmt.Sprintf("%s gate bypassed (--force: %s)", g.name, f.Force)`. With `g.name` from `changeCodeGates` (`changecode.go:209-215`), the emitted strings are now `structural gate bypassed` and `estimate-recon gate bypassed`. `processmanual/gatesig.go` still matches the old wording:

- `gatesig.go:115` — `AckPat: "structural gates bypassed \\(--force:"` (plural **gates**) → dead.
- `gatesig.go:121` — `AckPat: "estimate-reconciliation gate bypassed \\(--force:"` → dead.

(`plan-quality`/`estimate-quality` at `:112` and `estimate` at `:118` still match.)

Failure scenario: operator runs `sdlc change-code --issue N --force "hotfix"` with a structural failure. The binary prints `structural gate bypassed (--force: hotfix)`; the classifier's `ackRE` doesn't match. Because both rows carry `SilentAlone: true` (`gatesig.go:114`, `:120`), the ACK line is the **only** observable evidence that gate was bypassed — so a forced structural or estimate-recon bypass becomes invisible to `sdlc process-manual` and every retro built on it. Nothing catches this: `TestGateCatalogMatchesRegisteredFlags` (`gates_test.go:20`) only reconciles (command, flag) pairs against cobra flags, and `friction_test.go:123`/`:209` only exercise the plan-quality pattern.

Fix sketch: update `gatesig.go:115` → `` `structural gate bypassed \(--force:` `` and `:121` → `` `estimate-recon gate bypassed \(--force:` `` (this also brings both in line with the G2 grammar already documented at `gatesig.go:19`, `"<gate> gate bypassed (--force: <reason>)"`). Then close the drift permanently: add a test in `cmd/sdlc` that, for every `change-code` row in `GateCatalog`, asserts its `ackRE` matches the exact string `fmt.Sprintf("%s gate bypassed (--force: x)", name)` for the corresponding name from `changeCodeGateOrder()` — deriving the emitted string from the same declaration the runner iterates, in the spirit of B1's own guard. (`ARCH-DRY`: the consolidation was correct; the consumer sweep was skipped.)

## 3. Important findings

**I1 — `cmd/sdlc/changecode.go:424` / `internal/gatestate/ledger.go:230`: the pass-through does not prevent the re-dispatch it was built to prevent.**

`ContentHash(issueContent, planContent)` hashes the *entire* issue file — frontmatter and body. The retry it is supposed to make cheap is, by construction, the one that edits both: after plan-quality passes and the estimate gate refuses, B2's instruction is to add `estimate_hours:` to frontmatter **and** an itemized `## Estimate` block to the body. So on the next invocation the hash differs, `PassesUnchanged` returns false, and the judge re-dispatches.

Failure scenario: fresh issue, no estimate (per B2). Invocation 1 → plan-quality dispatches ~4 min, passes, stamps `content_hash` → estimate gate fails, exit 1. Operator adds `estimate_hours: 8.45` + the `## Estimate` block. Invocation 2 → `issueContent` changed → **full re-dispatch (~4 min)**. That is precisely the cost round 4 of this issue's own plan gate flagged as "B1 as specced would have made the gate more expensive," and the mechanism added to fix it doesn't cover the case. It's asserted otherwise in three places: `changecode.go:418-423` ("that retry is guaranteed on every issue"), `ledger.go:84-88`, and `atlas/workflow/gate-state.md:86-90`. Note the live exercise on #187 could not have caught this — #187 already carried `estimate_hours` from the pre-M1 rounds, so the ledger (2 rounds, one `content_hash`) never traversed it.

Fix sketch: hash what the plan gate is actually asked to review. B1 removed the estimate from plan-quality's remit entirely (`judge_test.go:331-337` now *forbids* `estimate_hours` in the prompt), so an estimate-only edit must not invalidate its acceptance. Strip the `## Estimate` section (`issue.EstimateSection`, already used at `changecode.go:315`) and the `estimate_hours:` frontmatter field before hashing — e.g. a pure `gatestate`-adjacent `planGateContent(issueContent) string` in `changecode.go`, unit-tested. Add the missing test: *"adding `estimate_hours` + an `## Estimate` block between invocations must NOT re-dispatch"* (sibling to `TestPlanQualityPassThroughOnUnchangedContent`, `changecode_test.go:402`). (`ARCH-PURPOSE`.)

**I2 — `cmd/sdlc/helptext/change-code.md:4-21`: `--help` still documents the pre-B1 gate order and never mentions the stateful gate.**

The numbered list still reads `1. Structural sanity → 2. Estimate gate → 3. Quality judges → 4. Branching`. The actual sequence is structural → plan-quality → estimate → estimate-recon → estimate-quality → branching. Task 9's file list mandated exactly this edit ("renumber the gate list to the new order… and add to the plan-quality entry: The gate is **stateful** (#187)…"). Three consequences:

- The agent-facing contract (`sdlc change-code --help`) contradicts the atlas row that *was* retimed and *is* guarded (`estimatetiming_test.go:82-96`).
- Line 17's "missing test surface" describes the ask C1 deleted from `plan-quality.md`.
- `WF_PLAN_ROUND_CAP` — a new operator-settable env var (`changecode.go:575`) — is documented only in `atlas/workflow/gate-state.md:73`. Every other `WF_*` var this binary reads is surfaced in helptext (`claim.md:62`, `merge.md:86`, `close.md:181`, `judge.md:45`), so this is a gap against the repo's own convention and the Docs update gate's "config keys" clause.

Fix sketch: renumber to the executed order, add the stateful-gate paragraph (ledger path `workshop/plans/NNNNNN-slug-plan-gate.md`, dispose-first, only *undisposed* Critical/Important block, Minor carried to the close review, `WF_PLAN_ROUND_CAP` default 3), drop the stale "missing test surface" phrase, and list `--no-estimate-recon` under FLAGS (it's referenced at `:12` but absent from `:52-60`).

**I3 — `cmd/sdlc/helptext/start-plan.md:30-32`: still tells the agent start-plan is where the estimate is set.**

> "Then a non-blocking `estimate_hours` nudge (#113): this is where you set the estimate — post-design, when scope is knowable — because `change-code` requires it (claim no longer does)."

`startplan.go:204-208` was retimed to say the opposite, so the verb's runtime output and its `--help` now contradict each other on the one policy B2 exists to unify. The Spec names this surface explicitly ("`sdlc start-plan`'s output"), and Done-when says "the helptext, `start-plan`, `helptext/estimate.md` and base-layer `AGENTS.md` all say so consistently."

The semantic guard misses it structurally: `estimateTimingRE` (`estimatetiming_test.go:21`) needs the literal `start-plan` within 80 chars of `estimate` **on the same line** — in `start-plan.md` the identifier is the *filename*, not the prose, so no line matches. And `helptext/start-plan.md` is absent from `TestEstimateTimingStatedPositively`'s list (`estimatetiming_test.go:72-78`). The guard built to catch exactly this class is blind to its most obvious surface.

Fix sketch: retime the prose ("a non-blocking `estimate_hours` nudge (#113, retimed by #187): do **not** derive it here — `change-code` runs plan-quality first and asks only after the plan clears"), and add `"cmd/sdlc/helptext/start-plan.md"` to the positive-assertion list so a revert fails.

**I4 — `cmd/sdlc/internal/gatestate/ledger.go:214-221`: `DispositionCounts` branches on disposition string literals.**

```go
switch state[f.ID] {
case "addressed":  addressed++
case "withdrawn":  withdrawn++
default:           open++
}
```

This is the exact posture `finding.cue:38-42` argues against at length ("a flat list plus a prose gloss would put the closes-vs-leaves-open decision in a Go switch… adding `deferred` would pass validation, match no case"), that `Closes()` exists to prevent, and that `closedSet` (`ledger.go:178`) correctly honors. Failure scenario: a future closing disposition (say `obsolete`) is added to `dispositions.closing`; `TestFindingConformance` passes, `OpenFindings` correctly settles it — and `DispositionCounts` silently reports it as **open**, under-counting settled findings in the one metric meant to answer "did the gate's findings get acted on, or worked around?". It's unwired today (M2 consumer), so it ships as a latent defect rather than a live one — cheap to fix now.

Fix sketch: drive the tally off the model — increment per closing-partition member (`for _, d := range m.Dispositions["closing"]`) or key a `map[string]*int`, so the counters are derived rather than enumerated. Extend `TestDispositionCounts` (`ledger_test.go:134`) to assert the buckets cover `AllDispositions()`. (`ARCH-DRY`, and the same shadow-sweep gap as C1.)

## 4. Minor findings

- `cmd/sdlc/changecode.go:3-12` — the package doc still says "Composes four gates" and lists the pre-B1 order (estimate #2, plan-quality #3). First thing a reader of the changed file sees.
- `workshop/plans/000187-tune-change-code-gate-plan.md:2440` — "the calibration-ledger schema row gains the **seven** columns"; Step 3 (`:2340`) and Step 4 (`:2388`) both say ten. This is carry-forward finding **PQ-3**, still partly open (see §7).
- `cmd/sdlc/estimatetiming_test.go:31` — the allowlist reason for `startplan.go` says "asserted positively below", but the positive assertion is in `startplan_test.go:127-136`, not below in this file.
- `cmd/sdlc/changecode_test.go:525` — `TestRoundCapFromEnv` reads the ambient `WF_PLAN_ROUND_CAP` before setting it; prepend `t.Setenv("WF_PLAN_ROUND_CAP", "")` for hermeticity.
- `cmd/sdlc/changecode.go:240` — `structural()` prints "fix the failures above, OR re-run with `--force <reason>`" even when `--force` was already supplied (the loop then warns it was bypassed). Pre-#187 the two paths printed different text.
- `cmd/sdlc/planreview.go:45` — with `--name` and no `--issue`, `f.Issue` is 0, so the ledger frontmatter records `issue: 0` and `Render` emits `# Gate ledger — ariadne#0 (plan-quality)`.
- `construct/vocabulary/finding.cue:70-76` — the `discovery` block exports to `finding.json:33-36` but `FindingModel` has no field for it, so `planreview.go:23`'s `planGateSuffix = "plan-gate"` is a hand-maintained restatement whose comment asserts a match nothing enforces. Consistent with the `verdict.cue` precedent (also declared, also unconsumed), so a note rather than a defect — but if `discovery` is never going to be derived for either noun, consider dropping it from both or adding the accessor `Issue()`/`Project()` already have.
- With `--force`, a corrupt ledger (`changecode.go:411-416`) causes plan-quality to return *before dispatching*, so the review never runs at all rather than running-and-being-overridden. Defensible under "bypass gate refusals", worth a line in the helptext paragraph from I2.

## 5. Test coverage notes

- **`runChangeCode`'s gate loop is untested end-to-end.** `exitWithCode` (`term.go:56`) is a plain func calling `os.Exit` — unlike `die`, which is a swappable var precisely so refusal paths are testable (`term.go:46-50`). So the loop's halt-on-first-failure and, more importantly, the `--force` continuation branch at `changecode.go:134-138` — the branch that emits the strings C1 is about — have no in-process coverage. Everything is tested one level down. Either promote `exitWithCode` to a var (matching `die`'s established seam) or add a subprocess test; the C1 guard is the cheap version that covers the message contract without either.
- **The estimate-edit pass-through case has no test** (I1). The three pass-through tests cover identical content, edited-plan content, and post-blocking content — none covers the one edit the new gate order guarantees.
- **The G2 ACK strings have no test** (C1). `friction_test.go` covers only `plan-quality`.
- Fuzz corpora: `FuzzRenderParseRoundTrip` ships its crasher seed; `FuzzParseFindingsBlock` has no `testdata/` entry, consistent with a clean run.
- `TestEstimateTimingConsistency` skips (rather than fails) when `git` is unavailable (`estimatetiming_test.go:105`, `:118`) and sweeps only tracked files — a newly-authored-but-unstaged doc escapes it. Acceptable; noting the boundary.
- I could not execute `go test ./...` or `go vet ./...`; no compile or runtime verification was performed.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — flag** (C1, I4). Both are the same shape: a value that became single-sourced in one place while a literal restatement of it survived elsewhere. When M2 wires `DispositionCounts` and the ten ledger columns, run the consumer sweep on *strings the binary emits* as well as on types — `processmanual`'s catalog is a machine-read consumer of `sdlc`'s stdout/stderr and should be treated as an API surface, not prose.
- **ARCH-PURE — pass.** The strongest part of the diff; no changes recommended. M2's `churnForWindow` is the one new IO seam and the plan already routes it through `gitx.RunGit` for error fidelity.
- **ARCH-PURPOSE — flag** (I1, I2, I3). The shadow-sweep on "one source, every consumer derives" comes out: prompt ✓, parser ✓, `Decide` ✓, `code-review.md` severities ✓ (drift-tested), `DispositionCounts` ✗, `planGateSuffix` ✗ (minor). And B2's "one story across every surface" has two live surfaces still telling the old one. The pattern to carry into M2: when a guard is written *because* a surface list proved incomplete, check that the guard's matching rule can actually reach every surface in that list — I3 is a guard that passes while the thing it guards is broken.
- **ARCH-MOCK — pass.** No new external dependency; `stubJudgeSeq` (`changecode_test.go:320`) extends the existing `judge.Run` var rather than standing up a second fake, and production and test flow share that boundary. Live conformance for the ` ```findings ` fence across `codex`/`gemini` is genuinely unproven, but it is recorded as Risk 5 with a stated fail-closed trigger (<5% → drop the fallback, >20% → simplify the schema), and the degradation path — prose fallback plus a persisted `protocol_error` round — is implemented and tested (`changecode_test.go:460`). That is the right posture for M1; M2's `gate_rounds` is the measurement that closes it.
- For **#183**: `Ledger.Gate`/`IDPrefix` as data does look genuinely reusable. The one thing to settle before the second consumer lands is whether `ContentHash`'s inputs are gate-specific — I1 shows the hash's *scope* is a per-gate decision (what that gate reviews), not a universal `issue+plan`. Consider making it `Ledger.ContentHash` + a per-gate content extractor supplied by the shell.

## 7. Plan revision recommendations

- **PQ-3 (carry-forward, Minor, still valid).** `workshop/plans/000187-tune-change-code-gate-plan.md:2440` still says "the calibration-ledger schema row gains the **seven** columns". Fix in place to "ten" — the count is now correct at `:2340` and `:2388`, so this is the last survivor. (PQ-1, PQ-2 and PQ-4 are confirmed addressed: the Task 14 harness now drives `runPlanQualityJudge` (`:2469-2481`), `churnForWindow` mandates `gitx.RunGit` including in the Integration-points table (`:163-166`), and `ClassifyPath` names `code-prod` as the explicit default with the lockfile/CI/manifest cases pinned (`:2196-2210`).)
- **Add a `## Revisions` entry** recording the M1-as-shipped deltas, so the plan stops claiming what the code doesn't deliver:
  - *Task 9 shipped incomplete.* `cmd/sdlc/helptext/change-code.md` (gate renumbering + the stateful-gate paragraph, Task 9 Step 3) and `cmd/sdlc/helptext/start-plan.md` (OUTPUT section retiming, Task 9 Step 3) were listed and not edited. The consistency guard cannot see `start-plan.md` because its regex requires the literal `start-plan` in the line body; the positive-assertion list must carry that file instead.
  - *Task 8 Step 4a's pass-through is narrower than specified.* The plan claims the short-circuit removes the guaranteed post-estimate-failure re-dispatch; as implemented, hashing the whole issue means the estimate edit invalidates it. Record the scope correction (hash the plan-relevant content, eliding the `## Estimate` section and `estimate_hours`) and the missing test.
  - *Task 8's `--force` consolidation has a downstream consumer the plan didn't name.* `processmanual/gatesig.go`'s G2 `AckPat` rows are literal-coupled to the five old messages; add them to the task's file list with the derive-from-`changeCodeGateOrder` guard.
