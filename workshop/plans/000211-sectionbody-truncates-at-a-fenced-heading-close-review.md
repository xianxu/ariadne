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

---

## Re-review — 2026-09-02T23:55:37-07:00 (FIX-THEN-SHIP)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | whole-issue close |
| milestone | — |
| window | 318689b02e0d9c347897cacc9b41be73d175f3b6..657e374fe279cb76998ec3860f687df8a54cfb06 |
| command | sdlc close --issue 211 |
| reviewer | claude |
| timestamp | 2026-09-02T23:55:37-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The shipped code is correct and, for the first time in this issue's nine rounds, each claimed fix is pinned by a test that reds when *exactly that fix* is reverted — I verified all three of BR-21's corollaries independently in a git-backed scratch clone, plus BR-18's writer guard. `PlanSectionRE` and `fencedCodeRE` are gone, every plan-item reader routes through `PlanItemsBody`, `TickMilestone` is a pure function with both filters independently pinned, and `internal/project` consumes the moved scanner unchanged. What blocks a clean SHIP is not the code this window writes but the class it declined to finish: BR-11's enumeration is still 4-of-6 swept, and I reproduced its false PASS empirically (`retro-recorded` accepts a fenced retro heading; `LatestRetroDate` returns a date that exists only inside a quoted example). That plus the atlas sentence still asserting "Everything that finds a heading is fence-aware" — which is the documentation of a sweep that did not happen — is the residue. None of it is a regression, so it does not block the gate.

## 1. Strengths

- **`internal/issue/fence.go:66` `FenceSpans`** — the two-pass structure is the right answer to the policy problem: the unterminated decision genuinely cannot be made until EOF, and unwinding `openedAt..len` afterwards is simpler than threading a lookahead. `UnterminatedPolicy` as a parameter with a per-consumer table at `fence.go:32` is the single best thing in this diff, and `TestSectionBody_UnterminatedFenceKeepsLaterSections` pins both directions.
- **`internal/issue/plan.go:62` `TickMilestone`** — extracting the writer out of `computeClose`'s IO shell is the ARCH-PURE fix, not another test. I confirmed both filters now red independently: deleting `inside[i] ||` reds `TestMilestoneTick_OnlyTicksTheRealPlan`, and replacing `SectionByteBounds` with whole-body offsets reds it on two other assertions.
- **`commitpathspec_guard_test.go:238` `assertWiring`** — collapsing three AST walks into one shared helper as the third one arrived is exactly right (ARCH-DRY), and I verified `TestPlanItemWritersUseTickMilestone` reds on reverting `close.go:568` to `pat.ReplaceAllString(newBody, …)` while the rest of the package stays green.
- **`commitpathspec_guard_test.go:305` `planItemMatchers`** — deriving the reader set from the counting regexes instead of listing readers by hand, with a stale-exemption check that refuses a dead entry. This is the shape the rest of the sweep still needs.
- **`workshop/lessons.md`** — the union-fixture corollary is stated as a rule with the measurement that produced it, not as a narrative.

## 2. Critical findings

None.

## 3. Important findings

**BR-11 remains open (4th consecutive round).** `cmd/sdlc/internal/project/guards.go:48` and `cmd/sdlc/internal/project/retro.go:11` still match against raw `d.SectionBody("Log")`. I reproduced the defect rather than inferring it — a project `## Log` whose only retro heading sits inside a ` ```markdown ` fence makes `retro-recorded` return nil (false PASS on a presence-means-pass gate) and `LatestRetroDate` return `"2026-01-01"` from the quoted example. Two further members the earlier rounds surfaced are also unswept: `internal/issue/structural.go:180` (`bulletRE` over the raw `## Done when` — same presence-means-pass shape) and `guards.go:30` (`ParsePhaseA` over the raw `## Estimate`). The enumeration is greppable in one command — `grep -n 'SectionBody(' ` across `cmd/sdlc` and `internal/{issue,project}` returns 13 call sites, and the members are exactly those that then run a structural matcher over the returned text. The rule is already stated at `fence.go:157`; what is missing is a derived guard in the shape of `planItemMatchers` that fails when a new `SectionBody` consumer matches structure without `StripFenced`/`FindLineOutsideFences`.

**BR-7 / BR-22 residue: `atlas/workflow/issue-lifecycle.md:181`.** "**Everything that finds a heading is fence-aware**" is unqualified while `cmd/sdlc/issue.go:530`, `project/guards.go:48`, `project/retro.go:11` and `structural.go:180` are not; `atlas:170`'s "Two helpers close that level" says the same thing one paragraph earlier. This is the atlas documenting a completed sweep that is 4-of-6 done.

## 4. Minor findings

- `close.go:308`/`:315` still resolve the Log section twice, the second call dropping `ok` (BR-14). `body[logStart:0]` is what a false `ok` produces.
- `internal/issue/section.go:73` says the tick "takes `SectionByteBounds` + `FenceSpans` directly (close.go)" — it is `issue.TickMilestone` in `plan.go` since this window.
- `close_test.go:154` says `TestMilestoneTickRegex` "mirrors the regex in runClose's milestone path"; `runClose` no longer contains that regex, and the test is a verbatim copy of `TickMilestone`'s pattern testing nothing production reaches.
- `workshop/issues/000211-…:32` reads `- [x] M2 — NOT done` under a table at `:37` asserting 2 open items. It was `- [ ]` at the base commit; `ebad613` flipped it — the whole-body tick this issue fixed, executed on this issue's own file, still uncorrected.
- `fence_test.go:228` still `TrimSuffix`es both sides, so byte-identity is asserted only up to trailing newlines (BR-8).
- The M2 Log records no `sdlc issue validate --all` re-run after the `stripCodeFences` rebase, which the Done-when asks for (BR-9).

## 5. Test coverage notes

`go test ./cmd/... ./pkg/...` is green at HEAD except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which I confirmed pre-existing and unrelated — `workshop/plans/000200-…-plan.md` is absent at the base commit too, and it is tracked as #210. Two gaps worth naming: `TestSectionByteBounds_MatchesSectionBody` exercises `SectionByteBounds`/`SectionHeadingByteOffset` only through four hand-written bodies while the corpus walk covers `SectionBody` alone; and `TestMilestoneScan_SeesMilestonesAfterAFencedHeading` now requires an ambient git repository (`git log: exit status 128` in a non-git checkout of HEAD), which is the new finding below.

## 6. Architectural notes

- **ARCH-DRY — pass.** One scanner, one `SectionLineBounds`, one `assertWiring`. `planGateContent:738` open-codes a line walk but builds it on `FenceSpans`, which is reuse of the primitive rather than a fourth scanner.
- **ARCH-PURE — flag.** `TickMilestone`'s extraction is the principle applied correctly; the milestone *enumeration* is the same shape left undone — it is a pure decision over `PlanItemsBody` + `milestonePlanRE` that only exists inside `findMilestonesMissingVerdict`, which shells out to `git log` per milestone. Pinning it therefore dragged git into the issue's central regression test. Fix: `issue.PlanMilestones(body) []string` as a sibling of `TickMilestone`, called by both `findMilestonesMissingVerdict` and the test, with the routing covered by the existing wiring guard — the pattern this round already established.
- **ARCH-PURPOSE — flag.** Shadow-sweep of the single-source change: `SectionBody`, `PlanSectionBody`, `PlanItemsBody`, `logHasEntryToday`, `insertLogLine` (heading, end, and day-header search), `planGateContent`, the counters and the tick all derive. `project`'s three matchers, `structural.go:180` and `issue.go:530` do not, and the atlas states otherwise. That is the instance-not-class pattern the issue's own Spec argues against.
- **ARCH-MOCK — flag (minor).** `gitx.RunGit` has no fake behind the seam here; the milestone test resolves it by depending on the ambient repository. Package convention elsewhere is a real temp git repo per test; this one silently borrows the developer's.
- **ARCH-CONSTRAINTS — pass.** Scanning is O(lines) per call and the repeated `FenceSpans` passes (`PlanItemsBody` runs two) are trivial against ~600-line issue bodies and a ~400-file corpus; `internal/issue` tests run in 2.1s.
- **ARCH-SECURE — pass with one note.** Input is repo-local markdown; `SectionByteBounds:105-118` clamps every derived offset and `SectionHeadingByteOffset` degrades to `0, true` rather than panicking. The one unguarded slice is the BR-14 double-compute, unreachable today because both calls take identical arguments.

## 7. Plan revision recommendations

- The issue still carries no `## Revisions` section for the in-place `## Done when` rewrite (the "one LINE-oriented scanner plus a recorded exception" narrowing). Per AGENTS.md §1, a mid-stream revision appends rather than overwrites.
- The Problem section's demonstration at `:26-38` no longer demonstrates what its table claims. Restore `- [ ] M2 — NOT done`, and the incident is worth one line in the Log — the issue's own file was corrupted by the bug it fixes, which is stronger evidence for the Spec's "don't write markdown in issues is not a workable policy" argument than anything currently recorded.

```findings
dispose:
  - id: BR-7
    disposition: not-addressed
    note: |
      Every named symbol and the close.go doc block are fixed; atlas:181's unqualified "Everything that finds a heading is fence-aware" survives while issue.go:530, project/guards.go:48, project/retro.go:11 and structural.go:180 are not.
  - id: BR-8
    disposition: not-addressed
    note: |
      fence_test.go:228 still TrimSuffixes both sides, and SectionByteBounds/SectionHeadingByteOffset are still absent from the corpus walk.
  - id: BR-9
    disposition: not-addressed
    note: |
      Still no post-M2 record; grep for "validate --all" in the issue returns only the Done-when and M1's two entries.
  - id: BR-11
    disposition: not-addressed
    note: |
      4th round open. project/guards.go:48 and project/retro.go:11 untouched by 657e374; I reproduced the false PASS (retro-recorded accepts a fenced retro heading, LatestRetroDate returns the quoted date), plus structural.go:180 bulletRE and guards.go:30 ParsePhaseA as members 5 and 6.
  - id: BR-14
    disposition: not-addressed
    note: |
      close.go:308 and :315 still resolve the Log section twice, the second call discarding ok.
  - id: BR-15
    disposition: not-addressed
    note: |
      Atlas half fixed (PlanItemsBody is listed). Issue :37 and :143 still cite close.go:563 (guard is at :578), and :32 still carries the tick-corrupted "- [x] M2 - NOT done" - it was "- [ ]" at 318689b and ebad613 flipped it, so the example now contradicts its own table.
  - id: BR-18
    disposition: addressed
    note: |
      Revert-verified in a git-backed scratch clone: restoring close.go:568 to the whole-body FindAllStringIndex + ReplaceAllString reds TestPlanItemWritersUseTickMilestone with the rest of the package green; the tick test now drives issue.TickMilestone.
  - id: BR-21
    disposition: addressed
    note: |
      All three corollaries measured individually - deleting `inside[i] ||` reds TestMilestoneTick, replacing SectionByteBounds with whole-body offsets reds it separately, reverting close.go's routing reds the new writer guard, and reverting FindLineOutsideFences to the match range reds its new unanchored test.
  - id: BR-22
    disposition: not-addressed
    note: |
      structural.go's block is attached and true, but the enumeration it asked for was not run - atlas:170 and :181 still overclaim, section.go:73 places the tick in close.go when it is in plan.go, and close_test.go:154 still says TestMilestoneTickRegex mirrors runClose's milestone path.
findings:
  - id: new
    severity: Minor
    family: pure-decision-tested-through-io
    title: |
      Pinning the milestone scan on the production path routed a pure assertion through `git log`, so the issue's central regression fails outside a git worktree
    detail: |
      planfence_test.go:72 now calls findMilestonesMissingVerdict, which invokes
      gitx.RunGit once per milestone. Measured: in a checkout of 657e374 with no
      .git the test fails with "findMilestonesMissingVerdict: git log: exit
      status 128", even though the fact under test - that a fenced heading no
      longer hides M2 from the plan-order enumeration - is decided purely before
      the first git call and then discarded on error. This is the same ARCH-PURE
      shape TickMilestone's extraction fixed on the write side, left undone on
      the read side. Extract issue.PlanMilestones(body) []string beside
      TickMilestone, call it from findMilestonesMissingVerdict and from the test,
      and let the existing assertWiring guard cover the routing.
  - id: new
    severity: Minor
    family: guard-key-not-scoped
    title: |
      assertWiring and the reader guard key the call index on `basename:funcName` while parsing two packages, so a shared filename silently merges their call sets
    detail: |
      commitpathspec_guard_test.go:255 and :368 build `calls[filepath.Base(path)
      + ":" + fn.Name.Name]` across both "." and "internal/issue". No basename
      collides today (verified by comm over both directories), but internal/issue
      already has plan.go, section.go and structural.go, and cmd/sdlc has
      close.go, issue.go and structural-adjacent files - one future close.go or
      issue.go in internal/issue makes two different functions with the same name
      share a key, and the guard would report a wiring edge satisfied by the
      wrong package. This is the tree's safety net for every unpinnable call
      site, so it should not be ambiguous: key on the directory-qualified path.
```
