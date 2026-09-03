# Boundary Review — ariadne#211 (whole-issue close)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | whole-issue close |
| milestone | — |
| window | 318689b02e0d9c347897cacc9b41be73d175f3b6..94a4fcc50023ae1eac5f1389196ee63e09f70cea |
| command | sdlc close --issue 211 |
| reviewer | claude |
| timestamp | 2026-09-02T23:36:46-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The core of #211 is delivered and correct: one fence-aware line scanner (`FenceSpans`/`ScanMarkdownLines`) lives in `internal/issue`, `project` consumes it at its own `UnterminatedIsFenced` policy, `PlanSectionRE` is gone with all six sites rerouted, and I verified the whole-issue window builds green (`go test ./cmd/... ./pkg/...` → only the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` failure from the missing `000200-*-plan.md`, #210). I independently confirmed the Done-when's verdict-stability claim across the *full* window: `sdlc issue validate --all` built at 318689b vs 94a4fcc is byte-identical output. What blocks a clean SHIP is not the shipped behavior but two things a fresh reader can measure: BR-11's enumeration was never finished — `project`'s `retro-recorded` guard still passes on a retro heading that exists only inside a fence (I reproduced it: guard PASS, `LatestRetroDate = "2026-01-01"` from a quoted example) — and three of this round's claimed fixes are pinned by nothing, including one the M2 Log asserts was revert-verified when it was not.

## 1. Strengths

- **`TestPlanItemReadersUsePlanItemsBody` is the real thing.** I reverted all four call sites (`close.go` ×2, `sizing.go:63`, `structural.go:160`) to `PlanSectionBody` and the guard reds with four named failures identifying file, function, and the counting regex that classes it a reader. Deriving readers from `planItemMatchers` rather than listing them (commitpathspec_guard_test.go:290-297) is the correct answer to BR-19, and the empty-exemption-map rationale is honest.
- **`UnterminatedPolicy` as a parameter, with the price documented and tested.** `fence.go:32-59` and `TestSectionBody_UnterminatedFenceIsProse` pin over-segmentation as the *deliberate* trade, so the next reader has to argue with a test rather than silently "fix" it. `TestUnterminatedPolicies_DisagreeOnPurpose` pins the fork between `stripCodeFences` and `SplitFences`.
- **`SectionByteBounds`' off-by-one reasoning is right and I checked it.** section.go:107-116 — dropping the byte `strings.Join` never emitted makes `body[start:end]` byte-identical to `SectionBody`, and it holds for the trailing-section, no-following-heading, and empty-section shapes I hand-walked.
- **Declining the `SplitFences` rebase and recording *why* at the function, the atlas and the Log** (structural.go:244-256) is the right call — character-oriented segmentation genuinely is a different abstraction, and it is pinned by `TestSplitFences`.
- **`TickMilestone` extracted as PURE** (plan.go:42-82) is the correct structural answer to BR-18's "the test could only restate the logic" — ARCH-PURE applied, not just cited.

## 2. Critical findings

None.

## 3. Important findings

**(a) BR-11 not-addressed — `project`'s within-section matchers, measured live.** `cmd/sdlc/internal/project/guards.go:48` and `retro.go:11` still run `retroHeadingRE`/`retroDateRE` over `d.SectionBody("Log")` raw. I probed it: `retro-recorded` — a presence-means-pass guard, the exact false-PASS shape this issue exists for — passes on a project whose only retro heading is inside a ```` ```markdown ```` block, and `LatestRetroDate` returns that quoted date into `RetroStale`. The enumeration was also short by one: `guards.go:30` `ParsePhaseA(d.SectionBody("Estimate"))` is a fifth member with the same shape. Prevalence: 1 of 349 corpus files already carries a fenced `### <date> — retro` (`workshop/history/plans/000180-*.md`) — latent exactly as `## Plan` truncation was.

**(b) Three fixes this round are unpinned, and one Log evidence line is false.** Details in §5; the family rule is below.

## 4. Minor findings

- BR-14 not-addressed: close.go:308 and :315 both resolve the Log section, and the second drops `ok` (`_, logEnd, _`). Unreachable today (identical args), but `body[logStart:0]` is what a false `ok` produces.
- BR-8 not-addressed: fence_test.go:228 still `TrimSuffix`es both sides, so it doesn't pin the byte-identity section.go:88 claims; `SectionHeadingByteOffset` has no direct test at all.
- BR-9 not-addressed: the Done-when's before/after `validate --all` is recorded only for M1. I measured it for the whole window — identical base→head — so the Log just needs the line.
- BR-15 half-open: the atlas now lists `PlanItemsBody`; the issue's `close.go:563` refs at lines 37 and 143 are now 578.
- BR-7 half-open: `atlas/workflow/issue-lifecycle.md:181` still asserts "**Everything that finds a heading is fence-aware**" unqualified, while `issue.go:530`'s structure peek deliberately is not (the Plan says "listed, not fixed") and `project`'s matchers are not swept. Same for the atlas's "Two helpers close that level".
- New: `structural.go:217-226` — the `stripCodeFences` doc block is detached by a blank line at :227 (so it isn't the function's doc at all) *and* is now false: "Naive — doesn't handle nested fences or indented code" and "NOT built on SplitFences, deliberately" describe the deleted `fencedCodeRE`, not `StripFenced`, which handles tildes, the width rule and ≤3-space indent.

## 5. Test coverage notes

Every claim below is a revert I ran in a scratch copy of 94a4fcc (baseline: 6 pre-existing failures from the absent `.git`; all reverts reproduced exactly those 6 unless stated).

| revert | result |
|---|---|
| `TickMilestone`'s `FenceSpans` filter removed | **PASS** — `go test ./cmd/sdlc -run TestMilestoneTick` green |
| `TickMilestone`'s `SectionByteBounds` scoping removed | FAIL (2 assertions) |
| `close.go:567` routing → old whole-body `ReplaceAll` | **PASS** — full suite green |
| `FindLineOutsideFences` → `off+loc[0], off+loc[1]` | **PASS** — full suite green |
| four `PlanItemsBody` call sites → `PlanSectionBody` | FAIL (4 named) ✅ |

The first row contradicts the M2 Log at issue line 318: *"reverting either filter reds it — `go test ./cmd/sdlc -run TestMilestoneTick` → two failures naming the quoted row and the row outside `## Plan`"*. Only one filter reds. The cause is the fixture: planfence_test.go:268-270 puts the quoted `- [ ] M1` in `## Problem`, so the *scoping* filter catches it and the fence filter is never exercised — there is no fenced plan row inside `## Plan` anywhere in the suite. Moving that fence into the Plan section fixes it in three lines.

Separately: `TestClosePlanGate_SeesItemsAfterAFencedHeading` and `TestMilestoneScan_SeesMilestonesAfterAFencedHeading` — the issue's reason-to-exist regressions — re-implement their gates over `PlanSectionBody` (the *unfiltered* body), so they don't traverse the production path at all. `findMilestonesMissingVerdict` is in-process drivable (close_test.go:512 does exactly that); calling it is nearly free.

## 6. Architectural notes

- **ARCH-DRY — pass.** Three scanners → one, with `project` delegating and the one survivor justified in prose. The M1 review's I5 (a second `SectionLineBounds` reintroduced one package over) was caught and collapsed.
- **ARCH-PURE — pass with one note.** `TickMilestone`'s extraction is the principle working as intended. `FenceSpans`/`SectionBody`/`StripFenced` are pure and unit-tested without IO; the corpus tests read the tree but skip on its absence.
- **ARCH-PURPOSE — flag.** The shadow-sweep does not close: `project`'s three within-section matchers derive from nothing, and `issue.go:530` is an undocumented-in-atlas exception. Per the principle, a family recurring across rounds is the ledger reporting that the enumeration was never written — `within-section matcher over an extracted section` is enumerable by one grep (`SectionBody(` + a regex in the same function).
- **ARCH-MOCK — N/A.** No external binary or service surface introduced; parsing is in-process.
- **ARCH-CONSTRAINTS — pass.** Two `FenceSpans` passes per section read on ~20 KB documents; `internal/issue` tests run in 1.5 s over 406 corpus files. No hot path.
- **ARCH-SECURE — pass.** Input is hand-editable markdown, which is exactly the untrusted-provenance case; `SectionByteBounds` clamps `start`/`end` and the parse degrades visibly (over-segmentation) rather than fabricating. The one residual is BR-14's discarded `ok`.

## 7. Plan revision recommendations

- The issue has **no `## Revisions` section** despite the Done-when having been rewritten in place mid-stream (from "one scanner" to "one line-oriented scanner plus a recorded exception"). AGENTS.md §1 asks for an appended `## Revisions` entry (timestamp + reason + delta) rather than an in-place edit narrated in the Log.
- The M2 Plan row at issue line 285-291 lists three production sites for the heading-finding sweep and carves out only `issue.go:530`. It should be revised to name `project/guards.go` (×2) and `project/retro.go` as either in-scope or explicitly deferred with a follow-up issue id — right now the Plan is ticked and the atlas claims closure over a class that has three live members.

```findings
dispose:
  - id: BR-7
    disposition: not-addressed
    note: |
      close.go doc block, :331, and logHeaderRE all fixed and stripEstimateForHash is gone from the tree and atlas; atlas:181 "Everything that finds a heading is fence-aware" is still unqualified while issue.go:530 and project's three matchers are not.
  - id: BR-8
    disposition: not-addressed
    note: |
      fence_test.go:228 still TrimSuffixes both sides; SectionByteBounds/SectionHeadingByteOffset are still absent from the corpus walk.
  - id: BR-9
    disposition: not-addressed
    note: |
      Still unrecorded. I measured it: sdlc built at 318689b vs 94a4fcc produces byte-identical `issue validate --all` output, so the Log line is all that is missing.
  - id: BR-11
    disposition: not-addressed
    note: |
      2 of 4 members swept (close.go dayRE, close.go tick). project/guards.go:48 and project/retro.go:11 remain, and I reproduced the false PASS: retro-recorded accepts a fenced retro heading and LatestRetroDate returns the quoted date. guards.go:30 ParsePhaseA is a 5th member the enumeration missed.
  - id: BR-12
    disposition: addressed
    note: |
      TickMilestone scopes to the real Plan and fence-filters; the scoping half reds on revert. The fence half is unpinned - folded into the new claimed-fix-unpinned-by-test finding.
  - id: BR-13
    disposition: addressed
    note: |
      Done-when now cites the M2 Log entry, the SplitFences plan row is corrected, logHeaderRE and stripEstimateForHash are gone. The issue still carries no `## Revisions` section for the in-place Done-when rewrite - noted as a plan-revision recommendation, not re-raised.
  - id: BR-14
    disposition: not-addressed
    note: |
      close.go:308 and :315 still resolve the Log section twice and the second discards ok.
  - id: BR-15
    disposition: not-addressed
    note: |
      Atlas half fixed (PlanItemsBody is listed). Issue lines 37 and 143 still cite close.go:563; the guard is at close.go:578.
  - id: BR-17
    disposition: addressed
    note: |
      All four doc-drift members swept and FindLineOutsideFences now returns the documented LINE range. The new contract is pinned by no test - reverting to the match range leaves the suite green; folded into the new finding rather than re-raised here.
  - id: BR-18
    disposition: not-addressed
    note: |
      The extraction landed and the test drives issue.TickMilestone, but the wiring guard the finding asked for was not written: reverting close.go:567 to the whole-body ReplaceAll leaves the full suite green, and the reader guard cannot see it because TickMilestone builds its pattern inline.
  - id: BR-19
    disposition: addressed
    note: |
      planItemReaders is now derived from planItemMatchers and I verified it reds on all four call sites with named failures.
  - id: BR-20
    disposition: addressed
    note: |
      close.go:571 now distinguishes "no `## Plan` section" from "no matching row" via issue.HasSection.
findings:
  - id: new
    severity: Important
    family: claimed-fix-unpinned-by-test
    title: |
      Three fixes this round are pinned by no failing test, and the M2 Log asserts a revert-verification for one of them that I measured to be false
    detail: |
      This is the 4th finding in family `claimed-fix-unpinned-by-test`. Earlier
      rounds fixed instances. Do NOT fix these three sites one at a time - the
      rule is what needs fixing.

      RULE: a fix is pinned only when reverting EXACTLY that fix - that filter,
      that call site, that return value - reds a named test. Three corollaries,
      one per member measured below: (1) when a fix installs two independent
      filters, the fixture must isolate each, because a fixture both filters
      happen to catch pins only their union; (2) where the wiring is not
      in-process drivable, a source-level guard must name it - the reader guard
      is the working example and needs a writer sibling; (3) where a contract is
      deliberately looser than every current caller exercises, the test must
      supply the caller the contract exists for.

      Measured at 94a4fcc in a scratch copy (baseline: 6 pre-existing failures
      from the absent .git; each revert reproduced exactly those 6):
      (a) Removing `inside[i] ||` from plan.go:72 leaves
      `go test ./cmd/sdlc -run TestMilestoneTick` PASS. Removing the
      SectionByteBounds scoping instead reds it (2 assertions). The M2 Log at
      issue line 318 claims "reverting either filter reds it - two failures"; only
      one does. Cause: planfence_test.go:268 puts the quoted `- [ ] M1` in
      `## Problem`, not inside `## Plan`, so no fenced plan row inside a Plan
      section exists anywhere in the suite.
      (b) Reverting close.go:567 to the old whole-body
      `pat.FindAllStringIndex` + `ReplaceAllString` leaves the full suite green.
      (c) Reverting fence.go:177 to `off+loc[0], off+loc[1]` leaves the full
      suite green - the line-vs-match contract BR-17 asked for is documented and
      unverified, because the sole caller anchors `^...$`.

      Also in scope of the same rule: planfence_test.go:45 and :63 -
      the issue's reason-to-exist regressions - re-implement close's guard and the
      milestone scan over the UNFILTERED `PlanSectionBody`, so they never traverse
      the production path. `findMilestonesMissingVerdict` is in-process drivable
      (close_test.go:512 drives it), so that one is nearly free to close.
  - id: new
    severity: Minor
    family: doc-drifts-from-code
    title: |
      stripCodeFences' doc block is detached by a blank line and describes the regex that was deleted
    detail: |
      This is the 6th finding in family `doc-drifts-from-code`. Earlier rounds
      fixed instances. Do NOT fix this instance alone - the rule (every symbol
      name and behavioral claim in a comment, the atlas or an issue must be
      greppable-true against the tree at the commit that writes it) is already
      stated by BR-17; what is missing is its enumeration being run over the
      files this window touched.

      structural.go:217-226 is separated from `func stripCodeFences` at :228 by a
      blank line at :227, so it is not the function's doc comment at all. Its
      content is also false at HEAD: "Naive - doesn't handle nested fences or
      indented code" and "NOT built on SplitFences, deliberately" describe the
      deleted `fencedCodeRE`; the function now delegates to `StripFenced`, which
      handles tildes, the closer-width rule and <=3-space indent. The
      unterminated-policy half of the paragraph is still true and is
      re-stated inside the function at :229-232, so the block is pure residue.

      Prevalence in this window, by grep: structural.go:217 (this one),
      atlas:181 "Everything that finds a heading is fence-aware" (BR-7 residue,
      contradicted by issue.go:530 and project/guards.go), atlas:176-179 "Two
      helpers close that level" (contradicted by the same), issue lines 37 and
      143 citing close.go:563 for a guard now at :578 (BR-15 residue).
```
