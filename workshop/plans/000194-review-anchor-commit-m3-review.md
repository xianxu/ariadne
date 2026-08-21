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

---

## Re-review — 2026-08-20T22:37:26-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 5e8a3e5e12f60c1b8c1d9e7d54fb0c1638bf46b4..93a3bc421148fc420e4ad4d7349841e9d997519f |
| command | sdlc milestone-close --issue 194 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-08-20T22:37:26-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The Critical from round 5 is genuinely fixed and — unusually — I can prove it from the prompt I was handed: this review's own prior-findings block carried nine families across the M2 and M3 rounds and escalated me on `family-plumbing-incomplete`, which is exactly the cross-boundary behavior `BR-20` said was missing. I mutation-verified the fixes rather than trusting the commit message: reverting `RenderPriorFindingsScoped` → `RenderPriorFindings` fails `TestBoundaryPriorFindings_FamiliesSpanMilestones` on all three assertions, and removing `f.Family = normalizeText(f.Family)` from `canonical()` fails `FuzzRenderParseRoundTrip/seed#7` instantly. `go build`, `go vet`, `gofmt -l`, `go test ./...` are all clean on a clean tree. What keeps this off SHIP is not a code defect — it is that **two of the six fixes landed with no test at all**: reverting `Family: f.Family` in `seedFromPlanGate` (BR-21) and reverting `r.N >= round` to the original `r.N == round` (BR-22) both leave the entire suite green, and each finding had explicitly named the test to add. That is the second instance of `test-pins-the-invariant`, so per the escalation I am not asking for those two tests as an instance fix — I am asking for the rule.

## 1. Strengths

- **The feature caught its own author, live.** The prompt for this round rendered the family vocabulary and the escalation block from the full issue-wide ledger. `BR-20` claimed the M3 prompt read *"This is the FIRST round"* with three milestones of families in the ledger; that is now demonstrably false. Self-hosting proof beats any unit test here.
- **`RenderPriorFindingsScoped(scoped, full)` is the right seam** (`cmd/sdlc/internal/gatestate/prompt.go:36`). It widens only the family input, leaves `Decide`/`OpenFindings` scoping untouched, keeps `RenderPriorFindings(l) = Scoped(l, l)` so plan-quality needs no edit, and the doc comment states *why* the two views differ. The fix also caught the second layer the finding named — the empty-scope early return short-circuiting before the family block (`prompt.go:38-48`).
- **BR-23 was fixed at the write boundary, not just the read boundary,** and pinned by fuzz rather than by an example. The recorded crasher corpus entry was *migrated* to three args (`testdata/fuzz/FuzzRenderParseRoundTrip/417cf3fd96f47e3d`) rather than dropped, and seed#7 `family="\n0"` is the exact shape that produced the unreadable ledger. Mutation-verified: it fails within 0.3s without the fix.
- **BR-24's consolidation kept the wrapper.** `NormalizeFamily` now calls `issue.Slugify` but survives as a named function so the paragraph about what normalization does *not* catch (`family.go:29-37`) has somewhere to live. Deleting the wrapper would have deleted the honest statement of the residual risk with it.
- **`TestBoundaryPriorFindings_FamiliesSpanMilestones` tests the production path** (`boundaryPriorFindings`), not the pure function — which is precisely the gap that let BR-20 ship. It also covers the `Milestone: ""` whole-issue case, the one that dropped everything.

## 2. Critical findings

None. BR-20 is fixed and pinned.

## 3. Important findings

**I1 — a gate fix without a failing-first test is not a fix.** *(2nd instance in family `test-pins-the-invariant`.)*

Measured, not asserted — I reverted each fix in place and re-ran `go test ./cmd/sdlc/...`:

| fix | revert to | suite result |
|---|---|---|
| BR-20 `RenderPriorFindingsScoped` | `RenderPriorFindings(FilterBoundary(...))` | **FAIL** (3 assertions) |
| BR-23 `canonical()` family normalize | drop the line | **FAIL** (fuzz seed#7) |
| BR-21 `Family: f.Family` in the seed | drop the field | **PASS — nothing caught it** |
| BR-22 `if r.N >= round` | `if r.N == round` (the original bug) | **PASS — nothing caught it** |

`TestConvergenceLine` does not discriminate BR-22 because it only calls round 2 of a 2-round ledger; the tools#1 fixture only calls rounds 2 and 4, where both families already exist. `ConvergenceLine(full, 3)` is the discriminating call (round 3 debuts `oracle-blind-direction`) and nothing makes it. `TestBoundaryReview_SeedsDeferredPlanGateFindings` (`boundaryledger_test.go:149`) asserts severity, title and the `PQ-1` provenance note — the finding named this test, and Family was not added to it.

Two more instances in the same family, so the rule is not a two-test patch:
- BR-29's requested items are both still absent — no round-trip test for `family`, and `TestFindingRenderBlockInstruction` (`pkg/vocab/finding_test.go:85`) still does not assert `family:`.
- The one *new* negative assertion is inert. `boundaryledger_test.go:543-548` guards on `strings.Contains(m2, "MUST dispose")`, but in the state under test `m2` is the first-round branch, which never emits that header — so the `t.Error` is unreachable. It also slices `m2[idx:idx+400]` unbounded, which would panic if the block were ever shorter than 400 chars past `OPEN FINDINGS`.

**The rule, stated:** *a fix for a gate finding is complete only when a test fails without it — verified by reverting the fix, not by inspection — and the `## Log` records that verification.* This repo already does this correctly once: M2's Log says C1 is *"pinned by TestCloseCommand\_LiveReviewSeesPriorFindings, verified to FAIL when the fix is reverted."* The rule exists in practice and was applied to 2 of 6 fixes this round. Measured prevalence on this issue: 4 instances (BR-21, BR-22, BR-29's two items), plus one inert assertion. The gate's failure mode is *silence* — an unpinned convergence bug reports "Converging" forever and nobody learns otherwise — which is why this family deserves the rule rather than four more one-liners.

**I2 — the plan's Core concepts table now contradicts the code in five rows.** *(2nd instance in family `plan-artifact-lags-code`.)*

`workshop/plans/000194-review-anchor-commit-plan.md:198-205`:

| table says | code says |
|---|---|
| `gatestate.FamilyCounts` → `ledger.go` | `family.go:45` |
| `gatestate.normalizeFamily` → `ledger.go` | `NormalizeFamily` (exported), `family.go:38` |
| `gatestate.ConvergenceLine` → `prompt.go` | `family.go:118` |
| `#Finding.family` → `finding.cue:78` | `finding.cue:91` |
| `gatestate.Finding.Family` → `ledger.go:34` | `ledger.go:43` |

Absent entirely: `RenderPriorFindingsScoped` — a **new exported API** introduced by this round's fix, which is the one entity downstream gates will consume. `## Revisions` has no M3 entry at all, so it records neither the scoped/full split nor that Task 3.4's stated *"reusing `DispositionCounts`"* was silently substituted with `len(r.Dispositions)`.

I am calling this Important rather than the checklist's blanket Critical, and saying why: every entity exists, is genuinely PURE, and is genuinely tested — only the stated paths are wrong. That is navigational drift, not behavior drift.

**The rule, stated:** BR-25 ticked the boxes; the table drifted in the same commit — so "update the plan" is not a rule an agent can follow, because it has no failing state. The enforceable version: **the `plan-unchecked` close gate reads only the *issue's* `## Plan`** (`cmd/sdlc/close.go:569`, `milestoneclose.go:113`) — the durable plan file in `workshop/plans/` has no gate at all, which is exactly why it lags. Extend that gate to the durable plan's checkboxes for the milestone being closed (it already has a `--no-plan-check` bypass and the flag convention from AGENTS.md §5). That is a follow-up issue, not M3 scope; file it, correct the five rows here, and the family closes.

## 4. Minor findings

- `construct/vocabulary/finding.cue:71-92` — the plan's Task 3.1 and the issue's `## Log` both justify the model-first ordering with *"closed schema — an unmodeled key fails instance validation."* Nothing validates a finding instance against `#Finding`: `grep -rn '#Finding'` across `*.sh`/`*.go`/`Makefile` returns **zero** hits outside the file itself, `cue export` yields only `[categories, discovery, dispositions, hardBlocking, when, whenDisposed]`, and `pkg/vocab/finding.json` contains no `family` key. Round 5 said in prose *"worth not restating as if it were live"*; the Log restated it. Either add a `cue vet … -d '#Finding'` instance case to `construct/vocabulary/vet_test.sh` (it already does exactly this for `#Project`), or drop the enforcement claim.
- `cmd/sdlc/boundaryledger.go:118-120` — `Family: f.Family,` and `Detail:` are separated by a blank line inside one struct literal, apparently to dodge gofmt's alignment group. `gofmt` accepts it; it reads as an accidental paste.
- `gatestate` now imports `cmd/sdlc/internal/issue`, whose package does filesystem work (`NextID` → `os.ReadDir`). `Slugify` itself is pure and the build has no cycle, so this is fine today — but the package doc says *"no filesystem"*, and `issue` can never import `gatestate` now. A shared `slug` leaf package would be the cleaner end state if a third caller appears.

## 5. Test coverage notes

- Mutation coverage is the honest measure and it is **2 of 4** on this round's code fixes (table in I1). BR-20 and BR-23 are pinned; BR-21 and BR-22 are not.
- The escalation/vocabulary rendering is well covered on the pure side (`family_test.go`) and now on the production side (`boundaryledger_test.go:496`). `TestFamilyCounts_TrueSynonymsAreNotMerged_AcceptedResidualRisk` remains the best test in the diff — its *name* is the documentation of the limit.
- `ConvergenceLine` is still never exercised through a full `close`/`milestone-close` run against M2's stateful reviewer fake; `TestBoundaryReview_EmitsConvergenceLine` calls `persistBoundaryRound` directly. Adequate as a unit test, but the stderr emission path is untested end-to-end.
- `ParseFindingsBlock`'s family normalization (`parse.go:46`) has no direct unit test; it is belt-and-braces behind `canonical()`, which the fuzz target does pin.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass.** BR-24 consolidated `NormalizeFamily` onto `issue.Slugify`; the seed still mints ids through `AssignIDsAt`; `readGateLedger`/`writeGateLedger` stay shared. The one residual over-share is `pluralFindings` doing double duty in the vocabulary list and the convergence line (open as BR-26).
- **ARCH-PURE — pass.** `family.go` is deterministic and IO-free end to end. `ConvergenceLine` is computed in the pure layer and merely *emitted* from `boundaryledger.go:193`. `RenderPriorFindingsScoped` widened the pure signature rather than pushing a boundary parameter into `Decide` — the same caller-side-transform discipline `FilterBoundary` established.
- **ARCH-PURPOSE — pass on the milestone, one shadow left.** Shadow-sweep on the `family` single source: consumers are the prompt vocabulary (derives ✓), the escalation (derives ✓), the convergence line (derives ✓), the plan-gate seed (derives ✓ as of BR-21), and the human prose projection (does **not** — open as BR-28). The CUE `#Finding` "source" derives nothing at all (Minor above): the Go struct and the `family: <slug>` literal in `RenderBlockInstruction` are hand-maintained restatements. That is the pre-existing pattern for every field, not M3's regression — but it means "model first" bought documentation here, not enforcement, and the artifacts should stop saying otherwise.
- **ARCH-MOCK — pass.** No new external dependency. The tools#1 four-round history is a copied `testdata/` fixture rather than a sibling-checkout read, so the acceptance test is hermetic. Git stays behind the real-repo fixture; the reviewer stays behind `judge.Run`.
- **For the whole-issue close.** `FilterBoundary(l, "")` will hide *every* M1/M2/M3 round — including the still-open `BR-10`, `BR-17`, `BR-18`, `BR-19` (M2) and whatever remains of `BR-26`…`BR-30` (M3). Families now cross the boundary; open findings still do not. So the close reviewer will be shown a nine-family vocabulary, told *"Earlier rounds fixed instances"*, and simultaneously told *"This is the FIRST round … no prior findings to dispose of."* That is D1 working as designed plus BR-26 unfixed, but the combination is newly visible because of this round's fix, and it deserves an explicit decision before `sdlc close` runs.

## 7. Plan revision recommendations

`workshop/plans/000194-review-anchor-commit-plan.md` has **no `## Revisions` entry for M3 at all**. It needs four:

- **"Core concepts table corrected (M3 boundary review, I2)"** — `FamilyCounts`, `normalizeFamily`→`NormalizeFamily`, and `ConvergenceLine` all landed in a new `cmd/sdlc/internal/gatestate/family.go`, not `ledger.go`/`prompt.go`; `#Finding.family` is `finding.cue:91`, `Finding.Family` is `ledger.go:43`. Add a row for `gatestate.RenderPriorFindingsScoped` (`prompt.go:36`, new, PURE) and for `renderFamilyVocabulary`/`renderFamilyEscalation`.
- **"D1's seam is the render signature, not `FamilyCounts` alone"** — D1 (`plan.md:81`) says *"`FamilyCounts(l)` takes the unfiltered ledger"*, and BR-20 proved that wording is unenforceable at the call site. Record `RenderPriorFindingsScoped(scoped, full)` as the actual seam so the next gate to adopt the ledger inherits the rule instead of re-deriving it.
- **"Task 3.4's `DispositionCounts` reuse dropped"** — `ConvergenceLine` counts `len(r.Dispositions)` for the target round instead. `DispositionCounts` is ledger-wide and keyed by disposition state, so it cannot answer "how many did *this round* dispose". State the substitution and why.
- **"M3 fix round: BR-21/BR-22 shipped unpinned"** — record the measured prevalence of `test-pins-the-invariant` (4 instances this issue) under `Limits`, per the escalation contract, alongside the follow-up issue extending the `plan-unchecked` gate to the durable plan file.

```findings
dispose:
  - id: BR-20
    disposition: addressed
    note: |
      RenderPriorFindingsScoped split; mutation-verified — reverting fails 3 assertions in TestBoundaryPriorFindings_FamiliesSpanMilestones, and this round's own prompt carried 9 families.
  - id: BR-21
    disposition: addressed
    note: |
      Family is carried by the seed; the test the finding named was not added — folded into the test-pins-the-invariant rule finding.
  - id: BR-22
    disposition: addressed
    note: |
      Now `r.N >= round`; no discriminating fixture — reverting to the original `== round` leaves the suite green.
  - id: BR-23
    disposition: addressed
    note: |
      normalizeText on both canonical() and ParseFindingsBlock, fuzz target at three args, crasher corpus entry migrated; mutation-verified via seed#7.
  - id: BR-24
    disposition: addressed
    note: |
      NormalizeFamily now calls issue.Slugify, wrapper retained so the not-caught note survives.
  - id: BR-25
    disposition: addressed
    note: |
      Tasks 2.3, 3.1-3.7 and all four Verification boxes ticked; the table drift is raised separately as the second instance of the family.
  - id: BR-26
    disposition: not-addressed
    note: |
      Unchanged, and now live: this round's prompt escalated on families with a count of 1, naming only the top of nine.
  - id: BR-27
    disposition: not-addressed
    note: |
      boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
  - id: BR-28
    disposition: not-addressed
    note: |
      render.go's prose projection still prints id/severity/title/detail only.
  - id: BR-29
    disposition: not-addressed
    note: |
      Neither item landed; subsumed by the test-pins-the-invariant rule finding below.
  - id: BR-30
    disposition: not-addressed
    note: |
      cinfo still sits between the demotion comment and its loop; finding.go:78 and the window test placement unchanged.
findings:
  - id: new
    severity: Important
    family: test-pins-the-invariant
    title: |
      Two of the six BR-20..BR-25 fixes shipped with no test — mutation-verified, and this is the 2nd instance of the family
    detail: |
      Reverting `Family: f.Family` (boundaryledger.go:118, BR-21) and reverting `r.N >= round`
      to the original `r.N == round` (family.go:124, BR-22) each leave `go test ./cmd/sdlc/...`
      fully green; BR-20 and BR-23 both fail loudly when reverted. Each finding had named the
      test to add. BR-29's two requested items are also still absent, and the one new negative
      assertion (boundaryledger_test.go:543-548) is unreachable in the state it tests and slices
      m2 400 bytes past an index without a bound. Do NOT fix these four instances. The RULE:
      a fix for a gate finding is complete only when a test fails without it, verified by
      reverting the fix rather than by inspection, and the Log records that verification. M2
      already did this once for C1 ("verified to FAIL when the fix is reverted"). Apply it as
      the standing bar for every disposition of `addressed`; measured prevalence this issue is
      4 instances plus one inert assertion.
  - id: new
    severity: Important
    family: plan-artifact-lags-code
    title: |
      The durable plan's Core concepts table contradicts the code in five rows and omits the new exported API
    detail: |
      plan.md:198-205 puts FamilyCounts, normalizeFamily and ConvergenceLine in ledger.go/prompt.go
      when all three are in family.go; normalizeFamily is exported NormalizeFamily; #Finding.family
      is finding.cue:91 not :78; Finding.Family is ledger.go:43 not :34. RenderPriorFindingsScoped
      — a new exported API downstream gates will consume — is absent, and there is no M3 Revisions
      entry at all, so the scoped/full split and Task 3.4's dropped DispositionCounts reuse are
      unrecorded. BR-25 ticked the boxes and the table drifted in the same commit, so this is the
      2nd instance. Do NOT stop at correcting the rows. The RULE, and it is enforceable: the
      plan-unchecked close gate (close.go:569, milestoneclose.go:113) reads only the ISSUE's
      "## Plan" — the durable plan in workshop/plans/ has no gate, which is why it lags. File a
      follow-up extending that gate (it already has --no-plan-check) to the durable plan's
      checkboxes for the milestone being closed.
  - id: new
    severity: Minor
    family: doc-claim-exceeds-enforcement
    title: |
      The "closed schema, an unmodeled key fails instance validation" rationale is enforced nowhere
    detail: |
      grep for "#Finding" across *.sh, *.go and Makefile returns zero hits outside finding.cue;
      cue export drops it, so pkg/vocab/finding.json has no family key and the Go struct plus the
      "family: <slug>" literal in RenderBlockInstruction are hand-maintained restatements. Round 5
      said in prose "worth not restating as if it were live"; Task 3.1 and the issue Log restate it.
      Either add a cue vet -d '#Finding' instance case to construct/vocabulary/vet_test.sh — it
      already does exactly that for #Project — or drop the enforcement claim from both artifacts.
```

---

## Re-review — 2026-08-20T22:49:07-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 5e8a3e5e12f60c1b8c1d9e7d54fb0c1638bf46b4..76de5dc5aeacd506906be21e0c416958b025eca6 |
| command | sdlc milestone-close --issue 194 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-08-20T22:49:07-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3's substance is delivered and, for the first time on this issue, I could verify the *verification* rather than trust it. I reverted each of the four BR-20..BR-23 code fixes in place and re-ran the suite: `Family: f.Family` in `seedFromPlanGate` now fails `TestSeedFromPlanGate_CarriesFamilyAcrossGates`, and `r.N >= round` → `r.N == round` now fails `TestConvergenceLine_LaterRoundsAreNotPriorFamilies`. Both were green-either-way at round 6; both are genuinely pinned now, and the commit's claim to have mutation-verified them is accurate. `go build`, `go vet`, `gofmt -l`, `go test ./...`, `construct/vocabulary/vet_test.sh`, and `sdlc process-manual` are all clean on a clean tree. BR-32's five table rows are corrected and independently checked against the code (`finding.cue:91`, `ledger.go:43`, all three symbols in `family.go`, `RenderPriorFindingsScoped` added), the M3 `## Revisions` entry exists, and #198 is filed against the right root cause. What keeps this off SHIP is two things, both of which I am raising as rules rather than instance fixes because both belong to families already in play: the assertion the commit says it *repaired* is still unreachable — I probed it and its body never executes — and the escalation prompt directs the reviewer to record prevalence in a `Limits` section that exists in no artifact model anywhere in the repo.

## 1. Strengths

- **The mutation-verification claim holds under independent re-testing.** I did not take the commit message's word for it: both reverts go red, and both go green on restore. That is the first time on this issue a fix-round's testing claim survived a check rather than being the thing the check caught.
- **The rule was applied, not just recorded.** `workshop/lessons.md:941-959` states the revert-to-verify rule with its origin *and* its measured prevalence, and the two instances were fixed in the same commit. BR-31 asked for the rule and got the rule plus the instances — the stronger of the two available responses.
- **#198 names the right root cause.** It does not stop at "keep the plan updated"; it identifies that the durable plan is the only major SDLC artifact with no automated check, cites the tell (the table drifted in the *same commit* that ticked its boxes), and observes the table is deterministically checkable without an LLM because every row names a path and a symbol. That is a fileable rule, not a resolution to be more careful.
- **`RenderPriorFindingsScoped` (`prompt.go:37`) remains the right seam** and its first-round branch (`prompt.go:39-49`) correctly carries families through the empty-scope case — the second layer BR-20 named.
- **`TestConvergenceLine_LaterRoundsAreNotPriorFamilies` (`family_test.go:176`) is a well-built discriminating fixture**: three rounds, a family debuting in the *middle*, and assertions at both round 2 (must not repeat) and round 3 (must repeat). It fails on the exact mutation it was written to catch.

## 2. Critical findings

None.

## 3. Important findings

**I1 — the assertion the commit says it repaired is still unreachable.** *(3rd instance in family `test-pins-the-invariant`.)*

`cmd/sdlc/boundaryledger_test.go:546-554`. I probed it by inserting `if !strings.Contains(m2, "OPEN FINDINGS") { t.Fatal(...) }` ahead of the guard and running the test:

```
boundaryledger_test.go:546: PROBE: m2 has NO OPEN FINDINGS section - assertion body never runs
```

The fixture's rounds are all `Boundary: "M1"`, so `FilterBoundary(l, "M2")` yields zero rounds, `RenderPriorFindingsScoped` takes the first-round branch (`prompt.go:39`), and the `"OPEN FINDINGS"` header (`prompt.go:55`) is never emitted. The outer `if` is false, the body never runs, and `t.Errorf` at :552 is dead — exactly as it was before. The unbounded `m2[idx:idx+400]` slice *was* genuinely fixed. But the comment at :542-545 now asserts a repair that did not happen: it says *"assert on a boundary-scoped OPEN finding instead"*, and the code still asserts on `BR-1`, which this fixture disposes at round 2. A comment claiming a test does something it does not is worse than the original, because the next reader stops checking.

Mitigating, and why this is Important rather than Critical: the invariant itself is covered elsewhere — `gatestate/boundary_test.go:14,54,79,97` pin `FilterBoundary`'s scoping directly, including that a seeded `BoundaryAll` finding is visible at every boundary while a milestone-scoped one is not. This is redundant dead coverage, not a hole.

**The rule, stated — and it is a genuine sharpening, not a restatement.** The rule adopted this round (revert the fix, watch the test go red) operates at *test* granularity, and that is precisely why it missed here: `TestBoundaryPriorFindings_FamiliesSpanMilestones` **does** go red when BR-20 is reverted — on its other three assertions — so the revert check reports the test as pinning, while one assertion inside it never executes. The sharper rule:

> **An assertion nested inside a runtime guard is unfalsifiable, because a dead branch and a passing branch are indistinguishable. Assert the guard, don't test under it:** write `if !cond { t.Fatal("precondition: …") }` followed by the unconditional assertion, or build the fixture so no guard is needed. `if cond { …assert… }` is only acceptable when `cond` is itself the assertion (`if strings.Contains(x, forbidden) { t.Error }`), which is the common and correct shape.

This is mechanically greppable and its prevalence here is exactly one: `boundaryledger_test.go:546` wrapping `:551` is the only nested-guard assertion in `cmd/sdlc/**_test.go` — every other `if strings.Contains(` in the suite is either a direct negative assertion or a loop-body filter. So the rule is cheap to state, cheap to enforce, and closes the gap the round-6 rule left open. Measured prevalence of `test-pins-the-invariant` on this issue is now 5 (BR-29's two items, BR-21, BR-22, and this).

## 4. Minor findings

**M1 — the escalation's fallback branch writes to a section that does not exist.** *(2nd instance in family `escalation-copy-precision`; ARCH-PURPOSE.)* `family.go:109-110` instructs the reviewer: *"If the rule cannot be stated, say why, and record the family in `Limits` with its measured prevalence."* There is no `Limits` section in `construct/datatype/`, `construct/vocabulary/`, any issue, any plan, or any skill — a repo-wide grep for `## Limits` and `` `Limits` `` returns four hits, and all four are this prompt string or a quotation of it (`family.go:110`, `000194-*.md:93`, `000194-*-plan.md:469`, and this review's own round-6 recommendation echoing it). So the escalation's escape hatch — the branch that preserves the measurement precisely when the rule is *not* writable, which is the harder and more valuable case — names no consumer. A reviewer following it either invents a heading or drops the prevalence data.

This originates in the issue's own Spec (`000194-review-anchor-commit.md:93`), so the implementation faithfully rendered what it was asked for; the gap is that nobody checked the referent exists. Since BR-26 (the first instance of this family) is still open and unfixed, the honest framing is not "state the rule so this stops recurring" — no instance has been fixed yet — but **fold this into BR-26's disposition**: the escalation block is hand-written prose that asserts things the ledger does not support (`counts[fam] >= 1` always true, "Earlier rounds fixed instances" asserted for open and withdrawn findings, one family's ordinal printed for all of them, a family total rendered as *"N new findings"*) and now also points at an artifact section that does not exist. The rule covering all five: **the escalation block must be generated from the data it describes and may reference only sections the artifact model defines** — either add `## Limits` to the issue/plan datatype, or point the instruction at `## Log`, which does exist.

- `atlas/index.md:14`'s one-line blurb for `gate-state.md` still enumerates only #187's surface — no families, no convergence line — though the linked file itself was correctly updated. The link requirement of AGENTS.md §8 is met; the description is stale.
- `prompt.go:46` returns the first-round branch without the `strings.TrimRight(…, "\n")` that :108 applies, so the two branches differ by a trailing newline.
- `parse.go:46` reads `normalizeText(rr.New[i].Family)` while the two lines above it read `normalizeText(f.Title)` / `normalizeText(f.Detail)` from the loop copy. Equivalent here, but the inconsistency invites a future edit that isn't.

## 5. Test coverage notes

- Mutation coverage on this round's two additions is **2 of 2** — I verified both independently rather than reading the claim. Combined with round 6's verified BR-20 and BR-23, all four code fixes across the fix rounds are now genuinely pinned.
- BR-29's first item is, contrary to round 6's disposition, effectively satisfied: `FuzzRenderParseRoundTrip` (`render_test.go:117`) now round-trips `family` through `Render → ParseSidecar → Render → ParseSidecar` with a dedicated seed (`family="\n0"`) and a migrated crasher-corpus entry. Its second item is not — `TestFindingRenderBlockInstruction` (`pkg/vocab/finding_test.go:85`) still asserts severities, dispositions and `id: new`, but not `family:`. The goldens cover it indirectly; the model↔prompt drift guard is where the invariant belongs, since a judge never told to emit `family` defeats the milestone silently.
- Still untested end-to-end: `ConvergenceLine`'s stderr emission through a full `close` / `milestone-close` run against M2's stateful reviewer fake. `TestBoundaryReview_EmitsConvergenceLine` calls `persistBoundaryRound` directly.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass.** `NormalizeFamily` is a one-line wrapper over `issue.Slugify` (`family.go:38`) with the residual-risk paragraph retained above it; id minting stays in `AssignIDsAt`; `readGateLedger`/`writeGateLedger` stay shared. The one live over-share is `pluralFindings` serving both the vocabulary list and the convergence line (BR-26, open).
- **ARCH-PURE — pass.** `family.go` is deterministic and IO-free end to end; the package's tests need exactly one `os.ReadFile` (for the `tools#1` fixture) and no mocks. `ConvergenceLine` is computed in the pure layer and emitted from `boundaryledger.go:193`.
- **ARCH-PURPOSE — flag (M1).** Shadow-sweep on the `family` single source: prompt vocabulary derives ✓, escalation derives ✓, convergence line derives ✓, plan-gate seed derives ✓ (verified by revert), human prose projection does not (BR-28, open), the CUE `#Finding` "source" enforces nothing so the Go struct and the `family: <slug>` literal are hand restatements (BR-33, open). New this round: the escalation's `Limits` consumer does not exist at all — a named sink with no artifact behind it, which is a stronger form of "doesn't derive" than a hand-maintained restatement.
- **ARCH-MOCK — pass.** No new external dependency in the window. The `tools#1` four-round history remains a copied `testdata/` fixture rather than a sibling-checkout read, so the acceptance test is hermetic; git stays behind the real-repo fixture and the reviewer behind `judge.Run`.
- **For the whole-issue close.** With families now crossing boundaries but open findings still boundary-scoped, the close reviewer will be handed a large family vocabulary and an escalation reading *"Earlier rounds fixed instances"* alongside *"This is the FIRST round … no prior findings to dispose of"* — because `FilterBoundary(l, "")` drops every M1/M2/M3 round, including the still-open `BR-10`, `BR-17`, `BR-18`, `BR-19`, `BR-26`–`BR-30` and `BR-33`. That combination is D1 working as designed plus BR-26 unfixed, but it is worth an explicit decision before `sdlc close` runs rather than discovering it in the prompt.

## 7. Plan revision recommendations

The plan is now accurate — the five drifted rows are corrected, `RenderPriorFindingsScoped` is listed, and the M3 `## Revisions` entry records both the scoped/full split and the dropped `DispositionCounts` reuse. Two small additions:

- **Extend the existing M3 `## Revisions` entry** with the two mutation verifications performed this round, per the rule the same entry adopts: reverting `Family: f.Family` fails `TestSeedFromPlanGate_CarriesFamilyAcrossGates`, and reverting `r.N >= round` fails `TestConvergenceLine_LaterRoundsAreNotPriorFamilies`. The rule says the Log records the verification; recording it for the two fixes that motivated the rule is the first application.
- **Add a Core-concepts row for `renderFamilyVocabulary` / `renderFamilyEscalation`** (`family.go:77`, `family.go:92`, new, PURE). They are unexported but they are the two entities the milestone's escalation behavior actually lives in, and #198's proposed path/symbol check would want them named.

```findings
dispose:
  - id: BR-26
    disposition: not-addressed
    note: |
      All four sub-items unchanged at family.go:95, :106-110, :85, :106; a fifth is added below (the `Limits` referent does not exist).
  - id: BR-27
    disposition: not-addressed
    note: |
      boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
  - id: BR-28
    disposition: not-addressed
    note: |
      render.go:104-116 still prints id/severity/title/detail with no family.
  - id: BR-29
    disposition: not-addressed
    note: |
      Item one IS now covered — FuzzRenderParseRoundTrip round-trips family with a dedicated seed. Item two is not: finding_test.go:85 still does not assert "family:".
  - id: BR-30
    disposition: not-addressed
    note: |
      cinfo still sits between the demotion comment and its loop; finding.go:78 and the window-test placement unchanged.
  - id: BR-31
    disposition: addressed
    note: |
      Rule written to lessons.md with origin and prevalence, AND both instances fixed — I independently reverted each and confirmed the named test goes red. The still-inert assertion is re-raised below as the family's 3rd instance with a sharpened rule.
  - id: BR-32
    disposition: addressed
    note: |
      Five rows corrected and verified against the code, RenderPriorFindingsScoped added, M3 Revisions entry written, and #198 filed against the root cause rather than the symptom.
  - id: BR-33
    disposition: not-addressed
    note: |
      vet_test.sh still has no -d '#Finding' instance case, and Task 3.1 plus the issue Log both still restate the enforcement claim.
findings:
  - id: new
    severity: Important
    family: test-pins-the-invariant
    title: |
      The assertion the commit says it repaired is still unreachable — 3rd instance, and the round-6 rule cannot catch it
    detail: |
      boundaryledger_test.go:546-554. Probed, not inferred: inserting a t.Fatal ahead of the
      guard prints "m2 has NO OPEN FINDINGS section". All fixture rounds are Boundary M1, so
      FilterBoundary(l, "M2") is empty, RenderPriorFindingsScoped takes the first-round branch,
      and the "OPEN FINDINGS" header at prompt.go:55 is never emitted. The slice-past-index bug
      WAS fixed; the unreachability was not, and the comment at :542-545 now claims a repair
      that did not happen. Do NOT fix this instance. The round-6 rule (revert the fix, watch it
      go red) structurally cannot catch this: the test DOES go red on its other three
      assertions. The sharpened RULE — an assertion nested inside a runtime guard is
      unfalsifiable, so assert the guard rather than testing under it; write
      "if !cond { t.Fatal(precondition) }" then assert unconditionally, or build the fixture so
      no guard is needed. "if cond { assert }" is fine only when cond IS the assertion.
      Greppable, and prevalence in cmd/sdlc is exactly one. The invariant itself is already
      covered by gatestate/boundary_test.go:14,54,79,97, so this is redundant dead coverage
      rather than a hole. test-pins-the-invariant now measures 5 instances on this issue.
  - id: new
    severity: Minor
    family: escalation-copy-precision
    title: |
      The escalation tells the reviewer to record prevalence in a `Limits` section that exists in no artifact model
    detail: |
      family.go:109-110 instructs "record the family in `Limits` with its measured prevalence".
      A repo-wide grep for "## Limits" and "`Limits`" returns four hits and all four are this
      prompt string or a quotation of it (family.go:110, the issue Spec at :93, plan.md:469,
      and round 6's own recommendation). There is no Limits section in construct/datatype/,
      construct/vocabulary/, any issue, any plan, or any skill. So the escalation's FALLBACK
      branch — the one that preserves the measurement when the rule cannot be stated, i.e. the
      harder case — names a sink that does not exist (ARCH-PURPOSE). It originates in the
      issue's Spec, so the implementation rendered faithfully; nobody checked the referent.
      BR-26 is still open, so no instance of this family has been fixed yet: fold this in
      rather than patching it separately. The rule covering all five sub-items — the escalation
      block must be GENERATED from the data it describes and may reference only sections the
      artifact model defines. Either add `## Limits` to the issue/plan datatype, or point the
      instruction at `## Log`, which exists.
```

---

## Re-review — 2026-08-20T22:58:22-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 194 — boundary reviews: anchor to the reviewed commit, and remember across rounds |
| repo | ariadne |
| issue file | workshop/issues/000194-review-anchor-commit.md |
| boundary | milestone M3 |
| milestone | M3 |
| window | 5e8a3e5e12f60c1b8c1d9e7d54fb0c1638bf46b4..7398a35a647c83ab3d263ddba0d4b540a72c2fbb |
| command | sdlc milestone-close --issue 194 --milestone M3 |
| reviewer | claude |
| timestamp | 2026-08-20T22:58:22-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

Verified the tree independently: `go build`, `go vet`, `gofmt -l`, `go test ./...`, `construct/vocabulary/vet_test.sh` (→ `ok`), and `sdlc process-manual` are all clean at `7398a35`.

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

M3 is substantively done and the last Important finding is genuinely closed — not repaired but **deleted**, which was the right call and which I verified rather than accepted: `gatestate/boundary_test.go:14-50` asserts unconditionally that `FilterBoundary(l,"M2")` yields only `BR-2`/`BR-3` and never M1's `BR-1`, so the guarded copy in `boundaryledger_test.go` really was redundant dead coverage and its removal loses no invariant. The `lessons.md` entry states the sharpened rule correctly, with a worked `if !cond { t.Fatal }` form and an honest account of why mutation-verification could not have caught it. I re-checked BR-32's five table rows against the filesystem myself — `finding.cue:91`, `ledger.go:43`, `FamilyCounts`/`NormalizeFamily`/`ConvergenceLine` all in `family.go:44/38/118`, `RenderPriorFindingsScoped` at `prompt.go:37` — and the plan now matches the code in every row, so `plan-artifact-lags-code` stays closed. **Nothing blocks this gate**: every remaining open finding is Minor and no Critical or Important is open. What keeps it off SHIP is that the escalation block — the *substance* of this milestone — is misfiring on this issue's own ledger, and I watched it do so in the prompt I was handed: `escalation-copy-precision` has two findings, neither ever fixed, yet the prompt told me it is "a rule that has already been patched at least once" and that "Earlier rounds fixed instances." That is BR-26 sub-item 4, still open after three rounds, and it is a live contradiction of Spec C's "≥1 **disposed** finding" in the one mechanism this milestone exists to deliver.

## 1. Strengths

- **The fix was deletion, and the reasoning survives scrutiny.** `boundaryledger_test.go:542-548` now carries a comment explaining *why* there is no assertion there and where the invariant actually lives. I checked the pointer: `TestFilterBoundary_ScopesRoundsToOneBoundary` covers it directly, with the filter as the subject rather than as a precondition. Replacing a test with a comment that names its real coverage is the honest move, and rarer than the repair.
- **`lessons.md:960-989` states a rule that generalizes past its origin.** "`if cond { assert }` is legitimate only when `cond` IS the thing being asserted" is mechanically checkable and correctly bounds itself — it doesn't outlaw the common negative-assertion shape.
- **Task 3.6's window test is not redundant with its neighbours**, which I checked rather than assumed: `TestBoundaryWindowBase_WholeIssueIgnoresPriorBoundary` (milestonewindow_test.go:159) has a trailer but runs on `main`; `TestBoundaryWindowBase_WholeIssueBasesOnBranchPoint` (:180) runs on a branch but plants no trailer. `WholeIssueStaysAtMergeBase` is the intersection — feature branch *and* a finalized `Review-Verdict:` boundary — and it is the only test that keeps M4 rejected.
- **The `family` shadow-sweep is now complete on the write path.** Only one non-test `gatestate.Finding{...}` construction exists in the tree (`boundaryledger.go:114`, the plan-gate seed), and it carries `Family`. `canonical()` and `ParseFindingsBlock` both normalize it; `FuzzRenderParseRoundTrip` fuzzes it with a migrated crasher seed.
- **#198 is filed against the root cause.** It cites the sharp tell — the table drifted in the same commit that ticked its boxes — and observes the check is deterministic (every row names a path and a symbol), so it does not need an LLM.

## 2. Critical findings

None.

## 3. Important findings

None open. BR-34 is disposed `addressed`.

## 4. Minor findings

**The convergence line's helptext shows output the formatter cannot produce, and the line ships markdown into a plain terminal.** *(3rd instance in family `escalation-copy-precision`.)*

`cmd/sdlc/helptext/close.md:72-73` documents:

```
round 4 — 2 new findings, 0 repeat families, 6 disposed. Converging.
round 2 — 3 new findings, 2 repeat families. Not converging: fix rules.
```

`ConvergenceLine` (`family.go:149-150`) formats `"round %d — %s, %s, %d disposed. %s"` with `verdict` being the literal `**Converging.**` / `**Not converging: fix rules, not instances.**`. So the second example is unreachable — the `, N disposed` segment is unconditional — and both examples drop the asterisks. `cinfo` (`term.go:34`) writes the string raw with only an ANSI prefix, so what an operator actually sees is `==> boundary gate: round 2 — 3 new findings, 2 repeat families, 1 disposed. **Not converging: fix rules, not instances.**` — markdown emphasis rendered nowhere. `TestBoundaryReview_EmitsConvergenceLine` asserts `Contains(out, "Not converging")`, which passes either way, so nothing pins the shape.

Both examples were copied from the issue's Spec D rather than from the formatter. Per the escalation contract I am **not** asking for these two lines to be patched — the rule is the one BR-35 already states, generalized one notch: *prose describing the gate's own output must be generated from that output or pinned to it by a test; a hand-written restatement of a formatter drifts.* Applying it here means either deriving the helptext example from `ConvergenceLine` (it is pure and takes a ledger — a doc-generation test can render one) or dropping the `**` so the string is honest about its only sink. Note honestly, against the escalation template's wording: **no instance of this family has been fixed**, so "Earlier rounds fixed instances" is false here — which is itself BR-26.

- `atlas/index.md:14`'s blurb for `gate-state.md` still enumerates only #187's surface — no families, no convergence line — though the linked file was correctly updated. AGENTS.md §8's link requirement is met; the one-line description is stale. (Raised in round 7's prose without an id; carrying it forward here rather than minting one.)
- `prompt.go:46` returns the first-round branch without the `strings.TrimRight(…, "\n")` that `:108` applies, so the two branches of one function differ by a trailing newline.
- `parse.go:46` reads `normalizeText(rr.New[i].Family)` while the two lines above read from the loop copy `f`. Equivalent today; the asymmetry invites an edit that isn't.

## 5. Test coverage notes

- The suite is green and the deletion cost nothing: `TestBoundaryPriorFindings_FamiliesSpanMilestones` still fails on its three remaining assertions if `RenderPriorFindingsScoped` is reverted (round 6 mutation-verified this; the deleted assertion was never one of the three).
- Mutation coverage on the four BR-20..BR-23 code fixes is now **4 of 4**, per round 7's independent reverts. I did not re-run those reverts this round — round 7's account is specific enough to check against the code, and both named tests (`TestSeedFromPlanGate_CarriesFamilyAcrossGates`, `TestConvergenceLine_LaterRoundsAreNotPriorFamilies`) exist and are structured to discriminate.
- Still absent, and still worth having: `TestFindingRenderBlockInstruction` (`pkg/vocab/finding_test.go:85`) does not assert `family:`. It is the model↔prompt drift guard; a judge never told to emit `family` defeats the milestone silently, and the goldens only cover it indirectly. (BR-29 item two, open.)
- Still untested end-to-end: the convergence line's stderr emission through a full `close` / `milestone-close` run against M2's stateful reviewer fake. `TestBoundaryReview_EmitsConvergenceLine` calls `persistBoundaryRound` directly.
- `ConvergenceLine(l, len(l.Rounds))` is safe against an N/index mismatch — I checked `persistBoundaryRound`'s `n := len(l.Rounds) + 1` and `Apply`'s append-only contract, so the last round's `N` always equals `len(l.Rounds)`. A mismatch would have silently reported "0 new findings … Converging"; it cannot arise on this path.

## 6. Architectural notes for upcoming work

- **ARCH-DRY — pass.** `NormalizeFamily` is a one-line wrapper over `issue.Slugify` with the residual-risk paragraph retained above it (`family.go:28-38`); id minting stays in `AssignIDsAt`; `readGateLedger`/`writeGateLedger` stay shared. The one live over-share is `pluralFindings` serving both the vocabulary list and the convergence line (BR-26, open).
- **ARCH-PURE — pass.** `family.go` and `prompt.go` are deterministic and IO-free end to end; the package's tests need exactly one `os.ReadFile` for the `tools#1` fixture and no mocks. `ConvergenceLine` is computed in the pure layer and merely emitted from `boundaryledger.go:193`.
- **ARCH-PURPOSE — flag, all instances already carried.** Shadow-sweep on the `family` single source: prompt vocabulary ✓, escalation ✓, convergence line ✓, plan-gate seed ✓, write-path normalizers ✓. Not deriving: the human prose projection (`render.go:110-116`, BR-28), the CUE `#Finding` "source" which enforces nothing so the Go struct and the `family: <slug>` literal are hand restatements (BR-33), and the escalation's `Limits` sink which has no artifact behind it at all (BR-35). I re-verified BR-35 independently — a repo-wide grep for `Limits` outside `workshop/history/` returns `family.go:110`, the issue Spec, `plan.md:469`, this review family's own text, and an unrelated `hardLimitsHeader` in `processmanual/session.go`. No `## Limits` exists in `construct/datatype/`, `construct/vocabulary/`, any issue, or any skill.
- **ARCH-MOCK — pass.** No external dependency added in the window. The `tools#1` four-round history is a copied `testdata/` fixture, so the acceptance test is hermetic; git stays behind the real-repo fixture and the reviewer behind `judge.Run` plus M2's stateful fake.
- **For the whole-issue close, one thing to decide before running it.** `FilterBoundary(l, "")` drops every M1/M2/M3 round, so the close reviewer will be shown the full family vocabulary (families cross boundaries now) *and* "This is the FIRST round of this gate at this boundary — there are no prior findings to dispose of." Simultaneously, the still-open `BR-10`, `BR-17`, `BR-18`, `BR-19`, `BR-26`–`BR-30`, `BR-33` and `BR-35` become invisible to every future gate — nine Minor findings that no later reviewer will ever be asked to dispose of. That is D1 working as designed, but the disposal of those nine is a decision to make deliberately, not one to discover by their absence.
- **Docs gate.** `atlas/workflow/gate-state.md` covers the new surface; `close.md` carries the family + convergence contract and `milestone-close.md:107` points at it as "the same gate". M3 introduces no new subcommand, flag, keybinding or config key, so README needs no change — confirmed by grep, not assumed.

## 7. Plan revision recommendations

The plan now matches the code; the M3 `## Revisions` entry records the scoped/full split, the dropped `DispositionCounts` reuse, and both adopted rules. Two small additions, both carried from round 7 and neither landed:

- **Extend the M3 `## Revisions` entry with the mutation verifications**, per the rule that same entry adopts: reverting `Family: f.Family` fails `TestSeedFromPlanGate_CarriesFamilyAcrossGates`; reverting `r.N >= round` fails `TestConvergenceLine_LaterRoundsAreNotPriorFamilies`. The rule says the record carries the verification; recording it for the two fixes that motivated the rule is its first application.
- **Add a Core-concepts row for `renderFamilyVocabulary` / `renderFamilyEscalation`** (`family.go:77`, `family.go:92`, new, PURE). They are unexported, but they are where the milestone's escalation behavior actually lives, and #198's proposed path/symbol check would want them named.

```findings
dispose:
  - id: BR-34
    disposition: addressed
    note: |
      The guarded assertion is deleted, not repaired, and the replacement comment names where the invariant really lives — I verified gatestate/boundary_test.go:14-50 covers it unconditionally. Rule landed in lessons.md with the worked precondition form.
  - id: BR-26
    disposition: not-addressed
    note: |
      All four sub-items unchanged at family.go:95, :106-110, :85, :106 — and sub-item 4 misfired on THIS round's prompt, which told me escalation-copy-precision had "already been patched at least once" when neither of its two findings has ever been fixed.
  - id: BR-27
    disposition: not-addressed
    note: |
      boundaryledger.go:193 still passes len(l.Rounds) rather than CountedRounds.
  - id: BR-28
    disposition: not-addressed
    note: |
      render.go:110-116 still prints id/severity/title/detail with no family, and the Open-findings projection likewise.
  - id: BR-29
    disposition: not-addressed
    note: |
      Item one remains covered by FuzzRenderParseRoundTrip. Item two is not - pkg/vocab/finding_test.go:85 still does not assert "family:".
  - id: BR-30
    disposition: not-addressed
    note: |
      All three sub-items unchanged - cinfo still between the demotion comment and its loop, pkg/vocab/finding.go:79 still says "for the plan-quality prompt", and the window test still sits in boundaryledger_test.go:486 rather than beside its five siblings in milestonewindow_test.go.
  - id: BR-33
    disposition: not-addressed
    note: |
      Ran vet_test.sh myself - it vets finding.cue at line 58 but has no -d '#Finding' instance case, unlike the -d '#Project' cases at :45 and :48.
  - id: BR-35
    disposition: not-addressed
    note: |
      family.go:110 still names `Limits`; re-verified by grep that the only non-quotation hit repo-wide is an unrelated hardLimitsHeader in processmanual/session.go.
findings:
  - id: new
    severity: Minor
    family: escalation-copy-precision
    title: |
      The convergence line's helptext shows output the formatter cannot produce, and the line emits markdown into a plain terminal
    detail: |
      close.md:72-73 documents "round 2 — 3 new findings, 2 repeat families. Not converging:
      fix rules." but family.go:149-150 formats "round %d — %s, %s, %d disposed. %s" with the
      disposed segment unconditional and the verdict literally "**Not converging: fix rules,
      not instances.**". Both examples also drop the asterisks, and cinfo (term.go:34) writes
      the string raw, so an operator sees "**Not converging: ...**" with the markers. Nothing
      pins the shape - TestBoundaryReview_EmitsConvergenceLine asserts only Contains("Not
      converging"). Copied from the issue's Spec D rather than from the formatter. Third
      instance, so do NOT patch these two lines. The rule, generalizing BR-35 one notch:
      prose describing the gate's own output must be GENERATED from that output or pinned to
      it by a test. Note that unlike the escalation template's wording, no instance of this
      family has been fixed yet - which is BR-26 sub-item 4 - so fold this in with BR-26 and
      BR-35 rather than treating it as a separate patch.
```
