---
id: 000069
status: working
deps: [000075]
github_issue:
created: 2026-06-02
updated: 2026-06-03
estimate_hours: 3
---

# merge the two per-milestone reviews into one boundary pass

## Problem

Each milestone boundary currently triggers **two independent fresh-context code
reviews** of the same diff:

1. The agent runs `superpowers-requesting-code-review` (a subagent) per AGENTS.md §3.
2. `sdlc milestone-close` **auto-dispatches** its own `sdlc judge milestone-review`
   (another agent) on the same window.

Observed in the 2026-06-02 `nous#41` session (4 milestones): the two passes were
**redundant** — the milestone-review judge mostly *confirmed* the superpowers review
rather than adding new findings — and **slow**: each `sdlc judge milestone-review` took
3–10 min (M3's worst), serializing the workflow with background waits the agent then had
to coordinate.

## Spec

> The two reviews should be merged. `superpowers-requesting-code-review` is borrowed from
> external (superpowers); `milestone-close` is home-grown. `sdlc milestone-close` intends
> to provide the **form**; the **essence** — in ariadne's design — should be the
> **adapted `superpowers-requesting-code-review`**. So fold milestone-close's judge into
> the adapted superpowers review: diff the two prompts and add milestone-close's tweaks
> (issue-ref awareness, the SHIP|FIX-THEN-SHIP|REWORK verdict line, the Review-Window
> trailer it emits) onto the adapted superpowers reviewer, then have milestone-close
> *invoke that one pass* instead of running a separate `judge milestone-review`.

There's already an adapted superpowers in the stack:
`construct/adapted/superpowers-writing-plans/` (sibling `superpowers-executing-plans/`) —
the requesting-code-review one should live alongside, and be what milestone-close calls.

## Refined design (2026-06-03)

Decided with operator (co-design session):

- **(A) Binary owns the review.** A review the agent can *skip* isn't a gate;
  binary-owned means it always runs, and the binary can also do the **cheap
  deterministic structural checks an agent forgets** (unticked-but-done boxes,
  missing `## Log` close entry, status not flipped) *before* spending tokens on
  the LLM pass. Division of labour: binary = cheap structural gate; one LLM
  review = judgment. So the agent runs `sdlc milestone-close`/`close`; it does the
  review; the agent does **not** separately dispatch `superpowers-requesting-code-
  review`.
- **One reviewer prompt** = reconcile the adapted superpowers `code-reviewer.md`
  (the general quality/architecture/testing/readiness checklist) with ariadne's
  milestone-review tweaks (issue-ref, the #70 `VERDICT:` contract line,
  `Review-Window` trailer, the **Atlas-update gate** and **Core-concepts
  cross-check**) **+ #75's `at-review` architectural lens**. Embed it as one
  source (à la #70/#75), used by the binary's review dispatch.
- **Close is a review boundary too.** `sdlc close` auto-dispatches the same
  review (whole-issue window): for a no-milestone issue that's the one review;
  for a multi-milestone issue it's an **end-of-issue review** (per-milestone
  reviews each see a slice — integration bugs + "do the milestones add up to the
  spec?" only show at the whole-issue diff). Plus the binary's structural checks.
- **Soft dep on #75** — the review consumes #75's `at-review` lens, so #75 lands
  first.

## Done when

- **One** fresh-context review per boundary (not two): the agent stops running a
  separate `superpowers-requesting-code-review` subagent; the binary's
  milestone-close/close review is the single pass.
- `sdlc close` runs the review (whole-issue window) + the cheap structural checks
  at the issue boundary; `sdlc milestone-close` does so at each milestone.
- The one reviewer prompt folds the superpowers checklist + ariadne tweaks + #75
  `at-review` lens; verdict feeds the `Review-Verdict:` trailer (#70 contract).
- AGENTS.md §3 + verb help reflect the single binary-owned pass (don't run both).

## Plan

Chosen shape: **Option B** — extract the reviewer prompt to an embedded markdown
file (`//go:embed`, à la #75's `architecture.md`); `code-review.md` is the single
source the binary renders. See `## Revisions` for why B over "enrich the Go
prompt", and why a *pointer* beats a *drift-tested duplicate*.

**Layering (settled in design):** two embedded files, two roles —
`architecture.md` = the **principles** registry (ARCH-* taste, at-plan/at-review
lenses, #75); `code-review.md` = the **review procedure** (checklist, severity,
output format, VERDICT). The procedure *refers* to markers; the registry *defines*
them; the binary co-locates them at dispatch (`render(code-review.md) +
ArchitectureBlock("at-review") + ContractPreamble + diff`). `code-review.md` must
NOT inline principle bodies (that would re-duplicate the registry — ARCH-DRY).

- [x] M1 — **the one embedded reviewer prompt + kill the double-run.**
  - New `cmd/sdlc/internal/judge/code-review.md` (`//go:embed`): reconcile the
    superpowers `code-reviewer.md` checklist (Code Quality / Architecture /
    Testing / Requirements / Production Readiness) with ariadne's milestone-review
    tweaks (issue-ref, Critical/Important/Minor severity, Core-concepts
    cross-check, Atlas-update gate, output format, `VERDICT: SHIP|FIX-THEN-SHIP|
    REWORK`). Placeholders `{{ISSUE_REF}}`/`{{BASE}}`/`{{HEAD}}`/`{{ARCH_STAR}}`.
    Its Architecture item is a *pointer* — "for each of `{{ARCH_STAR}}`, pass or
    flag; cite the marker" — never the principle text.
  - `ArchitectureMarkers() []string` — lift the `ARCH-([A-Z][A-Z-]*)` scan out of
    #75's drift test into one helper; used by **both** the drift test AND the
    `{{ARCH_STAR}}` substitution (ARCH-DRY — one extraction site, two consumers).
    Adding `ARCH-SHIM` (#71) flows into registry block + drift test + review
    checklist with zero edits to `code-review.md`.
  - Refactor `prompts.go` `MilestoneReview`: render `code-review.md` (substitute
    the four placeholders) + `ArchitectureBlock("at-review")` + `ContractPreamble`
    + diff. Render is pure string templating (ARCH-PURE); the `judge.Run` shim
    stays the thin IO seam. Strengthen `ArchitectureBlock`'s header → "work through
    **each** of the N entries below explicitly" (coverage, not just citation).
  - Reframe `construct/adapted/superpowers-requesting-code-review/`: SKILL.md → a
    pointer ("SDLC boundary reviews are binary-owned via `sdlc milestone-close`/
    `close`; don't run a separate pass. Canonical reviewer:
    `cmd/sdlc/internal/judge/code-review.md`. This skill remains for ad-hoc,
    non-SDLC reviews."). **NOT dropped** `code-reviewer.md` — at implementation
    it turned out a live sibling (`superpowers-subagent-driven-development`) still
    references it, so it's not an orphan duplicate; the root-cause fix is removing
    the *boundary mandate* (no double-run), not deleting the generic template.
  - AGENTS.md §3: the mandatory fresh-eyes review at a boundary IS the binary's
    milestone-close/close pass; the agent does NOT separately run
    `superpowers-requesting-code-review`. Keep the `BASE_SHA` = prev-boundary
    window semantics (the binary uses exactly that).
  - Tests: `code-review.md` embedded + non-empty; renders with `{{ARCH_STAR}}` →
    `ARCH-DRY, ARCH-PURE` and the four placeholders substituted; `MilestoneReview`
    composes body+registry+contract+diff; **guardrail** — `code-review.md` cites
    markers but does NOT contain principle bodies; `ArchitectureMarkers()` shared
    by the (refactored) #75 drift test.
- [x] M2 — **`sdlc close` as a review boundary** (+ shared dispatch helper).
  - Extract the review-dispatch from `milestoneclose.go` into one
    `dispatchBoundaryReview(window, issueRef)` helper (ARCH-DRY — close +
    milestone-close share it; emits the `Review-Verdict:`/`Review-Window:`
    trailer via the #70 contract).
  - `close.go`: on a standalone **issue** close, after the structural gates,
    dispatch the review on the whole-issue window (branch-point..HEAD). No-
    milestone issue → this is THE review (today it gets none — the biggest gap).
    Multi-milestone issue → the end-of-issue *integration* review (each milestone
    already reviewed its slice; integration + "do the milestones add up to the
    spec?" only show at the whole-issue diff).
  - **Guard (test it):** milestone-close calls `runClose` internally — that inner
    call must NOT dispatch (milestone-close dispatches its own milestone-window
    review). Fire close's dispatch only on the standalone issue-close path.
  - `--no-judge` on close (per-gate bypass, #67) to skip the review with explicit
    acknowledgment. Keep the existing milestone-verdict gate (both layers valued).
  - `side-quest:` fold the pending `warmupThreshold 2 → 1` fix (close.go) — warm
    up once, not twice (agents don't reread; #75's premise).
  - close.md helptext + AGENTS.md §3 + `atlas/workflow/sdlc-binary.md` updated to
    the single binary-owned pass at both boundaries.
  - Tests (stub `judge.Run`): close in issue mode dispatches on the whole-issue
    window + emits the trailer; milestone-close's inner `runClose` does NOT
    double-dispatch; `--no-judge` skips; existing milestone-close tests stay green.

## Revisions

### 2026-06-03 — Option B + design refinements (at design time)

- **Chose Option B** (extract to embedded `code-review.md`) over "enrich the Go
  prompt in place." Rationale: the operator's roadmap is *lift prompts into
  editable markdown* (cf. #75 `architecture.md`, #70 contract doc); B makes the
  reviewer prompt prose-edited data, not Go string-concat (ARCH-PURE), and is the
  first step of that migration. Cost (a second authoring style vs the six Go-built
  judge prompts) accepted as the direction, not an accident.
- **Pointer, not drift-tested duplicate.** The draft assumed two copies kept in
  sync by a test. B's whole point is ONE copy (`code-review.md`); the superpowers
  skill becomes a pointer, so there's nothing left to drift — strictly more DRY
  than a drift-tested duplicate (ARCH-DRY, Simplicity-First).
- **`{{ARCH_STAR}}` enumeration.** Glob `ARCH-*` resolves fine for *citation*
  (the registry is co-present in the composed prompt), but explicit enumeration
  improves *coverage* (per-principle pass/flag). Derive the list from the registry
  via the shared `ArchitectureMarkers()` (no hardcoding → no drift as the registry
  grows).

## Log

### 2026-06-03 — M2 review (FIX-THEN-SHIP) — findings addressed
- M2 boundary review (dogfooded through the new code-review.md prompt) returned
  FIX-THEN-SHIP with two Important + minors; all addressed before crossing:
  - **I1** `--no-judge` close skipped the log-line verdict annotation that
    milestone-close does → extracted `finishBoundaryReview` so BOTH the dispatched
    and the not-run/`--no-judge` paths emit the trailer AND annotate the log line.
    Test asserts the not-run annotation lands.
  - **I2 (ARCH-DRY)** `resolveReviewWindow` was a *parallel* reimplementation of
    close.go's atlas-gate commit scan (different match strings) — my comment had
    even claimed "same source". Extracted `firstCommitReferencing(refSubject)`,
    now the ONE scan both consume; atlas window ≡ review window by construction.
  - **minor** atlas command-table rows (`close`/`milestone-close`) refreshed to
    the unified review + `--no-judge`; gate count 7 → 8. Added readback assertions
    that the issue-close log line actually receives `; review verdict: …`.
  - left as noted (harmless): dry-run trailer asymmetry; `#69`-substring vs `#690`
    collision (pre-existing in the atlas gate, now commented in the shared helper).

### 2026-06-03 — M2
- 2026-06-03: closed M2 — dispatchBoundaryReview shared by close+milestone-close (ARCH-DRY); runCloseWithReview auto-dispatches the whole-issue review on standalone issue close; structural double-dispatch guard (only Milestone=="" dispatches; milestone-close calls runClose directly) verified by closereview_test.go; --no-judge skips; warmup 2->1; go test+vet+gofmt green; close --help shows the boundary review. atlas surface already documented in M1 note.; review verdict: FIX-THEN-SHIP
- Extracted the review dispatch into one `dispatchBoundaryReview(stdout, stderr,
  boundaryReviewParams)` shared by milestone-close and close (ARCH-DRY);
  generalized `resolveReviewWindow(refSubject)`, `emitTrailerBlock(_, _, kind)`,
  and `appendVerdictSuffix`/`annotateLogLineWithVerdict` to the no-milestone
  ("closed — ") case.
- `runCloseWithReview` wraps `runClose`: a standalone **issue** close
  auto-dispatches the boundary review on the whole-issue window (firstSHA^..HEAD,
  same source as the atlas gate) + emits the `close` trailer + mirrors the
  verdict into the log line. **Double-dispatch guard is structural:** only
  `f.Milestone == ""` reaches the dispatch, and milestone-close calls `runClose`
  directly (never `runCloseWithReview`), so a milestone is never reviewed twice.
  `--no-judge` (per-gate #67) skips with a not-run trailer.
- Folded the M1-review notes: the cwarn vocabulary now says "before crossing the
  boundary" (shared helper); `--no-judge` + `--agent` on close; close.md +
  header comments updated.
- `side-quest:` warmupThreshold 2 → 1 (warm up once; agents don't reread — #75's
  premise).
- Tests (`closereview_test.go`, stub `judge.Run`): issue close dispatches once on
  the whole-issue window + emits the trailer; **milestone close does NOT dispatch
  (the load-bearing guard)**; `--no-judge` skips; issue-close `appendVerdictSuffix`
  case. `go test ./cmd/sdlc/...` + vet + gofmt green; `close --help` shows the
  boundary review + flags.

### 2026-06-03 — M1 review (SHIP) — carry into M2
- M1 milestone-review verdict **SHIP (high confidence)**, two advisory notes,
  both M2 work: (1) `milestoneclose.go:387` cwarn still says "address before next
  milestone" — fold the vocabulary into M2's `dispatchBoundaryReview` extraction
  (and `close.go` boundary-review helptext) so all boundary prose says "boundary".
  (2) `archMarkerRE` (`ARCH-([A-Z][A-Z-]*)`) would greedily swallow a trailing
  hyphen if a marker were ever followed by `-lowercase` prose — harmless today
  (em-dash separators), glance when #71's ARCH-SHIM lands.

### 2026-06-03 — M1
- 2026-06-03: closed M1 — one embedded reviewer procedure (code-review.md) rendered by CodeReviewBody; {{ARCH_STAR}} from shared ArchitectureMarkers; MilestoneReview composes body+at-review block+contract+diff; no-inlined-bodies guardrail + render + markers tests; AGENTS.md §3 + skill reframed so boundaries are binary-owned (no double-run); go test+vet+gofmt green; composed prompt dogfood-rendered; review verdict: SHIP

- The exploration reframed scope: the binary **already** owns a fresh-context
  review (`milestone-close` → `judge.Dispatch` → `claude -p` with the
  MilestoneReview prompt, already carrying the at-review lens + atlas gate +
  core-concepts + VERDICT). The genuine redundancy was just AGENTS.md §3 telling
  the agent to *also* run a separate `superpowers-requesting-code-review` subagent.
- Shipped: `code-review.md` (embedded `codeReviewTemplate`, rendered by
  `CodeReviewBody`) as the one reviewer procedure; `ArchitectureMarkers()` (the
  shared ARCH-* extractor, reused by the #75 drift test → `{{ARCH_STAR}}`);
  `MilestoneReview` refactored to render body + `ArchitectureBlock("at-review")`
  + `ContractPreamble` + diff; `ArchitectureBlock` header strengthened to
  enumerate the N entries (ARCH-DRY coverage). Dogfood-rendered the composed
  prompt — reads clean.
- **Deviation from plan:** did NOT drop `construct/adapted/superpowers-requesting-
  code-review/code-reviewer.md`. A live sibling skill
  (`superpowers-subagent-driven-development/code-quality-reviewer-prompt.md`)
  references it, so it's not an orphan. Reframed SKILL.md + AGENTS.md §3 to drop
  the *boundary mandate* (kills the double-run) while keeping the generic template
  for ad-hoc/in-session use. Root cause over deletion.
- Tests: `TestArchitectureMarkers`, `TestCodeReviewBody_Renders`,
  `TestCodeReviewTemplate_DoesNotInlinePrincipleBodies` (the no-inlined-bodies
  guardrail); refactored the #75 narrative-drift test onto `ArchitectureMarkers()`.
  `go test ./cmd/sdlc/...` + vet + gofmt green.

### 2026-06-02

Filed from the sdlc tooling retro
(`workshop/pensive/2026-06-02-01-pensive-sdlc-tooling-retro.md`, finding F4). Operator:
"milestone-close = form; essence = adapted superpowers-requesting-code-review; tweak it
with the diff from milestone-close's judge."
