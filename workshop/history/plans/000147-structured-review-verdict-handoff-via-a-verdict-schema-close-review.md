# Boundary Review — ariadne#147 (whole-issue close)

| field | value |
|-------|-------|
| issue | 147 — Structured review verdict handoff via a verdict schema |
| repo | ariadne |
| issue file | workshop/issues/000147-structured-review-verdict-handoff-via-a-verdict-schema.md |
| boundary | whole-issue close |
| milestone | — |
| window | 9f3de32c8190806bbceff03580cda822a8f9aca9..HEAD |
| command | sdlc close --issue 147 |
| reviewer | claude |
| timestamp | 2026-06-30T16:03:05-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

I have completed a thorough review. The `#147` surface is verified green (`internal/judge` and `pkg/vocab` pass with real, non-cached times; `cue vet` + `vet_test.sh` green; `validate-instance --type verdict` verified by hand). Here is my verdict.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

VERDICT: FIX-THEN-SHIP (confidence: high)

The core mechanism lands correctly and is the right design: `verdict.cue` is a real single-source, `ParseVerdict` is now block-authoritative (the session's `"the verdict stands: FIX-THEN-SHIP"` + a block resolves to FIX-THEN-SHIP, not `unknown`), the drift guard pins the SHIP-family consumers, and the generic validator handles `--type verdict` with no binary change. All five Done-when items are met and there are **no correctness bugs**. What keeps it from a clean SHIP is a shipped-prompt contradiction (the legacy `ContractPreamble` still hard-orders "VERDICT: line MUST lead" while the new instruction says "emit the block first"), an atlas description that now mis-states the handoff, and a plan table that advertises a change that was deferred — all cheap, none blocking.

### 1. Strengths
- **The regression target is genuinely fixed and proven.** `ParseVerdict` is block-first → prose-fallback → unknown (`classify.go:204`), and `TestParseVerdict_BlockBeatsProse` (`verdictblock_test.go:54`) pins the exact session failure mode. Verified passing.
- **The prompt's own example block can't self-parse.** The template placeholder `verdict: <SHIP | …>` captures `<SHIP`, which fails `IsEmitted`, so `ParseVerdictBlock` returns `ok=false` for it (`classify.go:177`). A nice, deliberate safety property given the judges review this very parser.
- **The single-source is real, not cosmetic.** `verdictFor` collapsed its hardcoded switch to `vocab.Verdict().IsEmitted` (`classify.go:148`); the prompt renders from `RenderBlockInstruction()` (`review.go:43`). These derive, they don't restate.
- **Generic validation confirmed end-to-end:** `validate-instance --type verdict` passes a valid instance (exit 0) and rejects `MAYBE` with `want: FIX-THEN-SHIP|REWORK|SHIP` — the Log's claim holds, no binary change needed.
- **`TestVerdictConformance` is fail-closed and model-derived** (partition + predicate agreement), mirroring the issue conformance pattern (`conformance_test.go`).

### 2. Critical findings
None.

### 3. Important findings

**I1 — The milestone-review prompt ships two mutually-exclusive "must lead" instructions** (`prompts.go:359` appends `ContractPreamble`, `contract.go:34`). `ContractPreamble` still says *"Your response's FIRST line MUST be exactly: VERDICT: <TOKEN>"* and *"Do not put a title, heading, or any preamble above the VERDICT line; it must lead"* — which directly contradicts the new `code-review.md` text *"emit [the block] as the first thing in your response."* The plan's Task 4 named `contract.go (ContractPreamble)` as a file to modify, but the diff leaves it untouched. Not a parse break (a leading `VERDICT:` line still resolves via the fallback), but it under-serves the issue's purpose: an agent obeying `ContractPreamble` literally would emit only the prose line and skip the block, re-exposing the buried-prose risk #147 exists to kill. *Fix sketch:* `ContractPreamble` is shared across all judges, so don't blanket-edit it — render a review-specific preamble (or a one-line carve-out) that yields first-line precedence to the verdict block for the boundary-review category, keeping the `VERDICT:` line as the documented fallback.

**I2 — Atlas mis-states the handoff post-#147** (`atlas/workflow/sdlc-binary.md:333`). The "Judge → classifier contract" section still describes the verdict handoff as prose-only (`ParseVerdictToken` scans for the `VERDICT:` line) with no mention that the fenced verdict block is now the *authoritative* path via `ParseVerdictBlock`. The plan's Task 6 named `sdlc-binary.md`, but only `vocabulary.md` was updated. Done-when #5 is *partially* satisfied by the vocabulary bullet, yet the doc that actually describes the parse mechanism is now wrong. *Fix sketch:* one sentence at `sdlc-binary.md:333` noting block-first resolution (block authoritative, prose `VERDICT:` line the logged fallback).

**I3 — Plan Core-concepts table claims a modification that didn't ship.** The Integration-points table (plan line 36) lists `sidecar verdict frontmatter | cmd/sdlc/reviewsidecar.go | modified`, and Task 5 is an unchecked actionable, but `reviewsidecar.go` is untouched (Task 5 deferred — recorded in the issue Log, not in the plan body). The review's core-concepts cross-check treats a table↔code contradiction as needing a plan `## Revisions` entry. The deferral itself is legitimate and non-Done-when; the only defect is the plan still advertising the change. *Fix sketch:* see §7.

### 4. Minor findings
- **Drift-guard is narrower than the plan states.** `TestVerdictDriftGuard` pins only `verdictTokenRE` (`verdictblock_test.go:99`); the plan's disposition row named all three prose regexes — `verdictTokenLineRE` (`classify.go:56`) and `verdictConfidenceRE` (`classify.go:141`) are unguarded. All three match every emitted token *today*, so no live bug, but a future token would be caught by only one, and a dev could miss the siblings. Extend the loop to all three.
- **`milestoneclose.go:517`** hardcodes `'SHIP | FIX-THEN-SHIP | REWORK'` in the "no verdict found" warning — an unpinned 6th restatement that could read `strings.Join(vocab.Verdict().Emitted(), " | ")`.
- **`last-block-wins` vs. "lead with the block."** `ParseVerdictBlock` inspects only the *last* block and, if it's invalid, returns `ok=false` (it doesn't fall back to an earlier *valid* block). In the self-review scenario (reviewing this parser — #147's recurring case), a reviewer who quotes an illustrative verdict block after their real leading one would have the trailing block win or force a drop to prose. `TestParseVerdictBlock`'s "last block wins" case pins this deliberately, so it's a design reconsideration (first-block, or last-*valid*-block), not a bug.
- **Indented closing fence** isn't tolerated — `verdictBlockRE` requires backticks immediately after the newline, so `\n    ` + fence won't close. Unlikely from agents; note only.
- **`TestValidateInstance_VerdictGeneralizes`** (promised in the plan's Test surface) wasn't added; the capability is verified manually but unpinned. Low priority while the sidecar consumer is deferred.

### 5. Test coverage notes
- **#147 surface is green.** `go test ./cmd/sdlc/internal/judge/ ./pkg/vocab/` → `ok` with real (non-cached) times; every new test passes (`TestParseVerdictBlock`, `TestParseVerdict_BlockBeatsProse`, `TestVerdictDriftGuard`, `TestDispatch_ResolvesVerdictBlock`, `TestVerdictConformance`, `TestCodeReviewBody_Renders`, `TestBuildPrompt_MilestoneReview_HasContract`). `vet_test.sh` green.
- **Environmental, NOT a regression:** `go test ./cmd/sdlc/` (top-level package) times out in `TestSetStatusAlias_BothPathsMutate` at `acquireRepoLockForCommand`. Root cause confirmed via `.git/sdlc.lock/meta.json`: the real repo lock is held by `sdlc close --issue 147` (PID 8848) — **the very close operation running this review**. That test acquires the *real* `.git` lock (a pre-existing test-isolation trait), so it blocks on the in-flight close. **Do not clear the lock** — it belongs to the running close. Re-run `./cmd/sdlc/` after the close completes to confirm green.
- The process fixture covers `Dispatch → ParseVerdict`; the cmd/sdlc-level `dispatchBoundaryReview → trailer/sidecar` integration is covered by existing `closereview`/`milestoneclose` tests (unmodified), which I could not re-run here due to the lock.

### 6. Architectural notes
- **ARCH-DRY — pass with note.** The derived consumers (prompt, parser, `verdictFor`) genuinely single-source from `verdict.cue`. The residual hand-maintained sites (Verdict enum, `ContractTokens`, `blockingTokens`, 3 regexes, the `milestoneclose` warning) are an *acknowledged* trade-off, pinned by equality/subset tests because they also carry the deferred tri-state tokens. The DRY gaps are Minor 1 (2 of 3 regexes unpinned) and Minor 2 (the warning string).
- **ARCH-PURE — pass.** `ParseVerdictBlock`/`blockField`/`verdictFor`/`RenderBlockInstruction` are pure (string in, value/string out, no IO); the IO seam (`Dispatch` exec, sidecar write) stays in cmd/sdlc. The process fixture injects via the `Run` package var rather than mocking the pure core — correct.
- **ARCH-PURPOSE — pass with note.** Shadow-sweep: the consumers the issue's purpose *requires* to derive (prompt-accepted set, parser-accepted set, `verdictFor`) do derive; the block is authoritative; the buried-prose regression is fixed. The two deferrals (sidecar frontmatter → #136/#139; close finalize-policy → #139) are genuinely separable, not the point of this issue. The one place the purpose is under-served is I1 — leaving a conflicting "lead" instruction in the shipped prompt slightly weakens "reliable structured emission."

### 7. Plan revision recommendations
Add a `## Revisions` entry to `workshop/plans/000147-verdict-schema-handoff-plan.md`:
- **2026-06-30 — boundary-review (FIX-THEN-SHIP).** Record that **Task 5** (sidecar verdict frontmatter, `reviewsidecar.go`) and **`TestValidateInstance_VerdictGeneralizes`** are **deferred** (separable, non-Done-when; sidecar convergence tracked with #136/#139). Change the Integration-points table row `sidecar verdict frontmatter` status from *modified* → *deferred*.
- Note that **Task 4's `ContractPreamble` modification was not made** (only `code-review.md` changed), and that the drift guard pins only `verdictTokenRE` of the three prose regexes the disposition table named — so amend Task 4 / the disposition table to match what shipped (or implement the remainder per I1 + Minor 1).
