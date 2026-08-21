# Boundary Review — ariadne#194 (milestone M3)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 5e8a3e5e12f60c1b8c1d9e7d54fb0c1638bf46b4..261621d5e197f9d5ebc20dd669d5a0e179cdc1c8 |
| command | sdlc milestone-close --issue 194 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-08-20T22:18:14-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 delivers the substance it claims: `family` is modeled in CUE first, threaded into `gatestate.Finding`, the block instruction the judge is actually told to emit, and both golden prompts; `NormalizeFamily`/`FamilyCounts`/`ConvergenceLine` are genuinely pure and unit-tested with zero IO; the `tools#1` fixture is copied in rather than reaching for a sibling checkout; and Task 3.6's window regression test is real (it discriminates — the finalized boundary commit is not the merge-base). `go build`, `go vet`, `gofmt`, `go test ./...`, `vet_test.sh` and `sdlc process-manual` are all clean. What blocks SHIP is one wiring gap: the family vocabulary and escalation are computed from the **boundary-filtered** ledger, not the whole one — the exact opposite of what `family.go`'s own doc comment, `FilterBoundary`'s doc comment, plan D1, and `milestone-close --help` all assert, and it silently voids the sole justification for D1's one-ledger-per-issue decision. The acceptance test passes only because it calls `RenderPriorFindings` on an unfiltered ledger directly, which is precisely the "passes by construction while the live behavior fails" trap D3 was written to prevent.

## 1. Strengths

- **`family.go` is genuinely pure** and the whole convergence policy tests on in-memory ledgers with no mocks (`family_test.go` needs one `os.ReadFile` for the fixture and nothing else). ARCH-PURE holds.
- **The residual risk is pinned by a test's *name*** — `TestFamilyCounts_TrueSynonymsAreNotMerged_AcceptedResidualRisk` (`family_test.go:78`) documents the limit instead of pretending normalization solves synonyms. That is the honest version of D3.
- **`TestBoundaryReview_EmitsConvergenceLine` (`boundaryledger_test.go:456`) tests through the IO shell**, not just the pure function, and uses a *different spelling* for the repeat family — so it pins normalization end-to-end rather than by construction.
- **`TestBoundaryWindowBase_WholeIssueStaysAtMergeBase` (`boundaryledger_test.go:487`) actually discriminates**: it plants a finalized `Review-Verdict:` boundary commit that M4 would have adopted as the base, and asserts merge-base wins anyway. The failure message states *why* M4 was rejected — a rejected-alternative test that teaches.
- **`NormalizeFamily` is correct on the edges** — `"---"` → `""`, `"a-"` → `"a"`, leading hyphens suppressed by the `lastHyphen: true` initializer rather than a second trim pass.

## 2. Critical findings

**C1 — `cmd/sdlc/boundaryledger.go:81`: family counts are boundary-scoped, so a family never recurs across milestones.**

```go
return gatestate.RenderPriorFindings(gatestate.FilterBoundary(l, p.Milestone))
```

`RenderPriorFindings` calls `FamilyCounts(l)` on whatever ledger it is handed (`prompt.go:80`). In production it is handed the *filtered* view. Consequences, all live:

- At the whole-issue close `p.Milestone == ""` (`close.go:1053`), so `FilterBoundary(l, "")` drops every `M1`/`M2`/`M3` round. Verified against this issue's own ledger: all four rounds carry `boundary: M2`, and this very M3 review's prompt opens with *"This is the FIRST round of this gate for this issue"* — the M2 rounds are invisible. The close review will likewise see zero families.
- The issue's own evidence is the case that breaks: *"tools#1's rule spanned M1 rounds **and the close review**"* — the recurrence D1 cites as the reason for one file per issue is exactly the one this wiring cannot see.
- Three artifacts now assert behavior the code does not have: `family.go:44` ("across the WHOLE ledger — not a boundary-filtered view … a per-boundary count could not have seen it"), `ledger.go` `FilterBoundary` ("FamilyCounts deliberately takes the UNFILTERED ledger"), and `cmd/sdlc/helptext/milestone-close.md:100` ("finding families stay visible across milestones, which is the point of one file rather than several").
- The prompt text itself becomes false: `renderFamilyVocabulary` says *"Families already in play **on this issue**"* while listing one boundary's.

`TestFamilyEscalation_AgainstRealFourRoundHistory` cannot catch this — it calls `RenderPriorFindings(afterRound1)` directly, and every fixture round is `boundary: M1`, so filtering is a no-op there.

Fix sketch — keep the open-findings/cap scoping, widen only the family input:

```go
// gatestate/prompt.go
func RenderPriorFindings(l Ledger) string { return RenderPriorFindingsScoped(l, l) }
// families come from `full`; open/disposed lists from `scoped`
func RenderPriorFindingsScoped(scoped, full Ledger) string { … FamilyCounts(full) … }

// boundaryledger.go:81
return gatestate.RenderPriorFindingsScoped(gatestate.FilterBoundary(l, p.Milestone), l)
```

Add a test through `boundaryPriorFindings` (or extend `TestCloseCommand_LiveReviewSeesPriorFindings`) asserting an `M1`-round family appears in the `M2` prompt and at the whole-issue close. That test is what makes the fix stick.

## 3. Important findings

**I1 — `cmd/sdlc/boundaryledger.go:112-117`: `seedFromPlanGate` drops `Family`.** The seed copies `Severity`, `Title`, `Detail` but not `Family`, while D2 specifies *"re-issued as `BR-n` with `severity` **and `family`** preserved"*. This diff makes plan-quality findings carry families (`plan-quality.prompt` golden), so a rule first named at the plan gate arrives at the boundary anonymous — the earliest cross-gate recurrence, the one worth catching most, cannot escalate. One-field fix; assert it in `TestBoundaryReview_SeedsDeferredPlanGateFindings`.

**I2 — `cmd/sdlc/internal/gatestate/family.go:131`: `ConvergenceLine`'s "prior" set includes *later* rounds.** `if r.N == round { continue }` excludes only the target round, so every other round — including rounds after it — seeds `priorFamilies`. A family first raised at round 3 is reported as a repeat at round 3. Latent in production (the caller always passes the last round) but the package's own tests call it for round 2 and round 4 of a 4-round ledger, so the API is already used historically. Fix: `if r.N >= round { continue }`, plus a fixture where a family debuts in the middle round.

**I3 — `Family` bypasses the write-boundary normalizer, re-opening the unreadable-ledger hazard.** `canonical()` (`render.go:22-55`) normalizes `Title`, `Detail`, `Note`, `ID`, `Agent`, `Timestamp`, `Forced`, `ProtocolError` — its comment says *"no code path may produce a ledger document that cannot be read back, because that would permanently destroy the gate's memory for the issue"*. `ParseFindingsBlock` normalizes `Title` and `Detail`. `Family` is normalized in neither, and `FuzzRenderParseRoundTrip` (`render_test.go:104`) fuzzes only `title`/`detail`. A `family` with a leading newline (a block/folded scalar from the judge) hits the yaml/v3 emitter bug that fuzz target was built for; the ledger writes, then fails to read, and `blockOnLedgerFailure` wedges the boundary until the operator deletes the gate's memory. Fix: `f.Family = normalizeText(f.Family)` in `canonical` and in `ParseFindingsBlock`, and add a third fuzz argument for family.

**I4 — ARCH-DRY: `NormalizeFamily` (`family.go:32`) duplicates `issue.Slugify` (`cmd/sdlc/internal/issue/scaffold.go:58`).** Same algorithm — lowercase, non-alphanumeric → hyphen, collapse runs, trim edges — reimplemented with a builder instead of the regex. Both live under `cmd/sdlc/internal/`, so reuse is importable. Either call `issue.Slugify` (keeping `NormalizeFamily` as the named, documented wrapper so the comment about what it does *not* catch survives), or lift the slug helper into one shared spot. If gatestate is deliberately kept dependency-free, say so in the comment — right now there is no recorded reason for the second copy.

**I5 — the durable plan still says M3 is unstarted.** `workshop/plans/000194-review-anchor-commit-plan.md` has Tasks 3.1–3.6 (and 2.3, and the whole `## Verification` list) as `- [ ]`, while the issue's `## Plan` ticks M3 as done. AGENTS.md §8 puts the plan on the same per-milestone discipline as the atlas. Tick 2.3 and 3.1–3.6, and tick the four Verification boxes — I confirmed all four pass (`go test ./...` clean, `sdlc process-manual` renders, `construct/vocabulary/vet_test.sh` → `ok`, and the self-hosting ledger exists at `workshop/plans/000194-review-anchor-commit-close-gate.md`).

## 4. Minor findings

- `family.go:104`: `if counts[fam] >= 1` can never be false — `FamilyCounts` only stores counts ≥ 1. It reads as a tunable threshold but is dead; drop it or make it real.
- `family.go:110-118`: the escalation blockquote names only `repeats[0]` and computes the ordinal for that one family. With ≥2 families in play a reviewer following the template verbatim attributes the wrong family and count.
- `family.go:94`: `renderFamilyVocabulary` reuses `pluralFindings`, so a family's running total renders as *"3 new findings"* in the in-play list — that wording belongs to the convergence line only. A shared formatter whose text only fits one of its two call sites is DRY misapplied.
- `family.go:114`: "Earlier rounds fixed instances" is asserted for any prior finding, including ones still open or `withdrawn`. Spec C conditions on *"≥1 **disposed** finding"`.
- `boundaryledger.go:189`: the convergence line's round number is the raw index, so the D2 seed round (`NoCap`, no reviewer) makes the first real review print "round 2". `CountedRounds` exists for exactly this distinction.
- `boundaryledger.go:183-190`: the `cinfo` was inserted between the demotion comment paragraph and the `for … range d.Demoted` loop it documents, so the comment now reads as documenting the convergence line.
- `render.go:104-111`: the ledger's human prose projection prints `id / severity / title / detail` with no family, so a human reading `-close-gate.md` cannot see the families the gate is tracking.
- The convergence line is stderr-only — not in the trailer, the `-close-review.md` sidecar, or the ledger prose. Derivable later, but invisible on re-read.
- `boundaryledger_test.go:487`: `TestBoundaryWindowBase_WholeIssueStaysAtMergeBase` is a window test living in the ledger test file, borrowing `milestonewindow_test.go`'s helpers; its siblings (`TestBoundaryWindowBase_MilestoneBasesOnPriorBoundary`) are next door.
- `pkg/vocab/finding.go:78`: doc still says the instruction is "for the plan-quality prompt"; `milestone-review.md` has consumed it since M2.

## 5. Test coverage notes

- **The gap that hides C1:** nothing tests `boundaryPriorFindings` — the only production caller — for families. Every family test either calls the pure function or calls `persistBoundaryRound` (which uses the unfiltered ledger and so looks fine). The one test that would have caught it is a live-path test asserting a prior boundary's family reaches the next boundary's prompt.
- **No round-trip test for `family`**, though Task 3.1 says "Test the round-trip". It is structurally safe today (`yaml.Marshal` on the whole struct, `canonical` copies by value), but nothing pins it, and I3 shows the write path is where the field is under-handled.
- `TestFindingRenderBlockInstruction` (`pkg/vocab/finding_test.go:85`) is the model↔prompt drift guard and does not assert `family:`. The goldens cover it indirectly; the invariant belongs in the model test too — a judge never told to emit `family` defeats the entire milestone.
- `TestBoundaryReview_EmitsConvergenceLine` exercises `persistBoundaryRound` directly rather than M2's stateful reviewer fake. Reasonable unit choice, but the convergence line is never seen through a full `close`/`milestone-close` run.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — flag.** I4 (`NormalizeFamily` vs `issue.Slugify`) and the `pluralFindings` over-share. Positive: the seed correctly routes id minting through `AssignIDsAt` rather than hand-rolling `BR-<i+1>`.
- **ARCH-PURE — pass.** `family.go` is deterministic and IO-free end to end; `ConvergenceLine` is computed in the pure layer and merely *emitted* from the shell. No test needs a mock to run a "pure" entity.
- **ARCH-PURPOSE — flag (C1, I1).** The shadow-sweep: the source is the `family` slug; the consumers are the prompt vocabulary, the escalation, the convergence line, the seed, and the human prose projection. Two of the five do not derive from the whole ledger (C1) or from the field at all (I1, and the prose projection). The escalation *looks* implemented and, across milestones, is inert — which is the failure mode D3 named in advance.
- **ARCH-MOCK — pass.** No new external dependency. Git stays behind the hermetic real-repo fixture; the reviewer stays behind `judge.Run`; the `tools#1` history is a copied testdata file rather than a sibling-checkout read. `#Finding` in CUE still has no instance-validation consumer (`vet_test.sh` only asserts the *concrete* export blocks, because `#`-definitions do not export) — so the commit message's "an unmodeled key fails instance validation" is not enforced anywhere in this repo. Pre-existing (raised on #187's close), not M3's to fix, but worth not restating as if it were live.
- **For the whole-issue close:** `FilterBoundary(l, "")` also hides M2's still-open `BR-10` (`disposition: not-addressed`, round 4, never re-disposed) from every future boundary. That is D1 working as designed, but the consequence — a `not-addressed` finding that no later gate will ever surface — deserves an explicit decision before the close review runs.

## 7. Plan revision recommendations

Add to `workshop/plans/000194-review-anchor-commit-plan.md`:

- **`## Revisions` — "Core concepts table corrected (M3 boundary review)":** `gatestate.FamilyCounts` and `gatestate.normalizeFamily` were listed at `ledger.go`, and `gatestate.ConvergenceLine` at `prompt.go`. All three landed in a new `cmd/sdlc/internal/gatestate/family.go`, and `normalizeFamily` is exported as `NormalizeFamily` (it is called from `family_test.go` and is part of the package's documented ingest contract). The new file is the better home — record it so the table stops pointing at the wrong path.
- **`## Revisions` — "Task 3.4's `DispositionCounts` reuse dropped":** `ConvergenceLine` counts `len(r.Dispositions)` for the target round instead. `DispositionCounts` is ledger-wide and keyed by disposition state, so it does not answer "how many did *this round* dispose". State the substitution and why, or reuse it.
- **`## Revisions` — the D1/C1 seam, once C1 is fixed:** D1 says "`FamilyCounts(l)` takes the **unfiltered** ledger", but the seam that enforces it is the *`RenderPriorFindings` signature*, not `FamilyCounts` alone. Record the split (`RenderPriorFindingsScoped(scoped, full)`) so the next gate to adopt the ledger inherits the rule rather than re-deriving it.
- **Tick** Tasks 2.3, 3.1–3.6 and the four `## Verification` boxes.

```findings
findings:
  - id: new
    severity: Critical
    family: family-plumbing-incomplete
    title: |
      Family counts come from the boundary-FILTERED ledger, so a family never recurs across milestones
    detail: |
      boundaryledger.go:81 hands RenderPriorFindings the FilterBoundary view, and
      RenderPriorFindings calls FamilyCounts on whatever it is given (prompt.go:80). At the
      whole-issue close Milestone is "", so every M1/M2/M3 round is dropped and the family
      vocabulary is empty. This contradicts family.go:44, FilterBoundary's own doc, plan D1,
      and helptext/milestone-close.md:100, and voids the sole justification for one ledger
      per issue. Verified on this issue's ledger: all four rounds are boundary M2 and the M3
      prompt reads "This is the FIRST round". Fix by passing the unfiltered ledger for
      families only (RenderPriorFindingsScoped(scoped, full)) and testing through
      boundaryPriorFindings.
  - id: new
    severity: Important
    family: family-plumbing-incomplete
    title: |
      seedFromPlanGate drops Family, so a plan-gate rule arrives at the boundary anonymous
    detail: |
      boundaryledger.go:112-117 copies Severity, Title and Detail but not Family, while D2
      specifies severity AND family preserved. This diff makes plan-quality findings carry
      families, so the earliest cross-gate recurrence cannot escalate. One-field fix; assert
      it in TestBoundaryReview_SeedsDeferredPlanGateFindings.
  - id: new
    severity: Important
    family: prior-means-strictly-earlier
    title: |
      ConvergenceLine counts LATER rounds as prior families
    detail: |
      family.go:131 skips only r.N == round, so rounds after the target seed priorFamilies. A
      family debuting at round 3 is reported as a repeat at round 3. Latent in production
      (the caller passes the last round) but the package's own tests already call it for
      round 2 and 4 of a 4-round ledger. Fix: if r.N >= round { continue }, plus a fixture
      where a family debuts mid-history.
  - id: new
    severity: Important
    family: agent-text-normalization
    title: |
      Family bypasses canonical() and ParseFindingsBlock, re-opening the unreadable-ledger hazard
    detail: |
      render.go:22-55 normalizes every other agent-authored string precisely so no code path
      can emit a ledger that cannot be read back; ParseFindingsBlock normalizes Title and
      Detail. Family is normalized in neither and FuzzRenderParseRoundTrip fuzzes only
      title/detail, so the yaml/v3 leading-newline emitter bug is reachable again through
      family. Its consequence is blockOnLedgerFailure and a boundary wedged until the
      operator deletes the gate's memory. Add normalizeText on both paths and a third fuzz
      argument.
  - id: new
    severity: Important
    family: existing-helper-not-reused
    title: |
      NormalizeFamily duplicates issue.Slugify rather than reusing it
    detail: |
      cmd/sdlc/internal/issue/scaffold.go:58 already implements the identical algorithm
      (lowercase, non-alphanumeric to hyphen, collapse runs, trim edges). Both packages live
      under cmd/sdlc/internal so reuse is importable. Consolidate, or record in the comment
      why gatestate keeps a second copy.
  - id: new
    severity: Important
    family: plan-artifact-lags-code
    title: |
      The durable plan still shows M3 (and Task 2.3, and every Verification box) as unstarted
    detail: |
      workshop/plans/000194-review-anchor-commit-plan.md has Tasks 2.3 and 3.1-3.6 and the
      whole Verification list as unticked while the issue ticks M3 done. AGENTS.md section 8
      puts the plan on the same per-milestone discipline as the atlas. All four Verification
      items were confirmed passing during this review.
  - id: new
    severity: Minor
    family: escalation-copy-precision
    title: |
      The escalation block names only the top family, reuses convergence-line wording, and has a dead threshold
    detail: |
      family.go:104 (counts[fam] >= 1 can never be false), family.go:110-118 (the blockquote
      hardcodes repeats[0] and its ordinal, so with two families in play a reviewer copying
      the template attributes the wrong count), family.go:94 (pluralFindings renders a
      family total as "3 new findings"), and family.go:114 ("Earlier rounds fixed instances"
      asserted for findings that are still open or withdrawn, where Spec C conditions on a
      DISPOSED prior finding).
  - id: new
    severity: Minor
    family: counted-rounds-consistency
    title: |
      The convergence line's round number counts the no-cap seed round
    detail: |
      boundaryledger.go:189 passes len(l.Rounds), so after a D2 seed the first real review
      prints "round 2". CountedRounds exists for exactly this distinction and is what the
      cap uses.
  - id: new
    severity: Minor
    family: family-plumbing-incomplete
    title: |
      The ledger's human prose projection omits family
    detail: |
      render.go:104-111 prints id, severity, title and detail per finding. A human reading
      NNNNNN-slug-close-gate.md cannot see the families the gate is tracking, and the
      convergence line is stderr-only, so nothing durable shows them either.
  - id: new
    severity: Minor
    family: test-pins-the-invariant
    title: |
      No round-trip test for family, and the model-drift guard does not pin the family key
    detail: |
      Task 3.1 asks for a round-trip test; none exists. TestFindingRenderBlockInstruction
      (pkg/vocab/finding_test.go:85) is the prompt-model drift guard and does not assert
      "family:" — the goldens cover it indirectly, but a judge never told to emit family
      defeats the milestone, so the invariant belongs in the model test.
  - id: new
    severity: Minor
    family: comment-anchor-drift
    title: |
      The convergence cinfo was inserted between the demotion comment and the loop it documents
    detail: |
      boundaryledger.go:183-190. Also pkg/vocab/finding.go:78 still says the block
      instruction is "for the plan-quality prompt" though milestone-review has consumed it
      since M2, and boundaryledger_test.go:487's window regression test lives in the ledger
      test file rather than beside its siblings in milestonewindow_test.go.
```
