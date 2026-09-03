# Boundary Review — ariadne#211 (milestone M2)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | e56f1a7b4519db6cb6e295817b0e4f8f549ea530..ebad6130219bb942d05edd7e0562a2761fce9e4f |
| command | sdlc milestone-close --issue 211 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-02T20:35:10-07:00 |
| verdict | FIX-THEN-SHIP |

## Review

```verdict
verdict: FIX-THEN-SHIP
confidence: high
```

The M2 sweep gets the hard part right: `logHasEntryToday`, `insertLogLine` and `planGateContent` now locate `## <heading>` through the shared fence-aware scanner, the `SplitFences` non-rebase is a well-argued and well-recorded decision rather than a shortfall, and I independently confirmed the new `SectionByteBounds` is byte-identical to `SectionBody` across 2886 real sections with no panics. What blocks a clean SHIP is that three of the milestone's own claims are partial in ways their tests don't catch: `insertLogLine` fixes the *anchor* but leaves the *search window* running to EOF, so I reproduced the close line being spliced inside a fenced example — the exact #66 defect the commit says is "solved properly", passing its new test only because the fixture's dates don't collide; the fence filter reached 1 of 5 plan-item consumers, so `sdlc state` and `sdlc close` now disagree about the same Plan; and the indent-policy parameter kept "to make the axis explicit" is wired to zero call sites with a comment claiming it serves `SplitFences`, which never calls it. All fixes are localized (~25 lines + 3 tests). `go test ./cmd/sdlc/... ` is green apart from the pre-existing `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` failure (#210, missing archived plan).

## 1. Strengths

- **The `SplitFences` non-rebase is the right call and is recorded three times over** (`structural.go:245-257`, `fence.go:16-19`, `atlas/workflow/issue-lifecycle.md:158`). "Character-oriented vs line-oriented, and it answers a different question" is a real distinction, and the Log explains the attempt was made and refused by `TestSplitFences`' inline-pair case. Narrowing the Done-when from "one scanner" to "one line-oriented scanner plus a recorded exception" instead of forcing the merge is exactly right (ARCH-DRY correctly *not* applied).
- **`SectionByteBounds` is correct where it counts.** I cross-checked `body[start:end] == SectionBody(...)` over every non-fenced `##` in `workshop/**/*.md`: 2886 sections, 0 mismatches, no index panics. The off-by-one reasoning in the comment at `section.go:85-88` is right.
- **The `stripCodeFences` rebase is measurably verdict-neutral.** I compared old `(?s)```.*?``` ` against the new `StripFenced` over every `## Spec` in the corpus: 1 file moves (`000147-*.md`, 246→247 words), 0 crossings of the ≥50 gate threshold. The rebase gained tildes + the width rule without moving a single gate verdict.
- **`TestLogHasEntryToday_IgnoresAQuotedLogSection`** (`planfence_test.go:118`) is a well-built regression: it asserts both directions (real date found, quoted-only date not matched), so it reds on the old first-match lookup for a real reason rather than incidentally.
- **`fence.go`'s policy table and the "over-segmentation is the deliberate price" test** (`fence_test.go:81-105`) pin a decision that a future reader would otherwise "fix". That's the kind of comment that earns its keep.

## 2. Critical findings

**C1 — `insertLogLine` locates the heading fence-aware but scans for the day header to EOF, so it still files the close line inside a quoted fence** (`cmd/sdlc/close.go:308-330`)

`section := body[logStart:]` is "the real Log section **+ anything after it**", and `dayRE.FindStringIndex(section)` searches all of it — including fenced examples in later sections. Reproduced:

```
## Log

### 2026-01-01
- an older entry

## Appendix
```markdown
## Log

### 2026-09-02
- a QUOTED example entry
```
```
`insertLogLine(body, "- 2026-09-02: closed")` writes the close line under the `### 2026-09-02` **inside the fence**. The real `## Log` gains nothing, and `logHasEntryToday` (now fence-aware) will not see it.

`TestInsertLogLine_TargetsTheRealLogSection` (`planfence_test.go:132`) does not catch this: its `trailing` fixture quotes `### 2020-01-01` while the inserted line is dated `2026-09-02`, so it passes on a date mismatch, not on the window being bounded. This is the #194 pattern — a test written from the same mental model as the fix, asserting whatever the fix happens to do. (Not a regression from `main`: last-match anchored inside the fence too. It is an incompletely delivered fix that the milestone claims is complete.)

*Fix sketch:* `SectionByteBounds` — introduced in this same diff — already returns the end. Use it and bound both the `dayRE` and `insertRE` searches to the section:
```go
bodyStart, bodyEnd, ok := issue.SectionByteBounds(body, "Log", issue.UnterminatedIsProse)
// … derive logStart from SectionHeadingByteOffset, but slice section = body[logStart:bodyEnd]
```
Pin it with the fixture above (quoted `### <today>` after the real Log, real Log lacking today's header) — it reds without the bound. ARCH-SECURE: this is the write path degrading invisibly on hand-authored input rather than failing loudly.

## 3. Important findings

**I1 — the plan-item fence filter reached 1 of 5 consumers, and the one write path isn't section-scoped at all**

> **This is the 2nd finding in family `consumer-enumeration-incomplete`.** Earlier rounds fixed instances. Do NOT fix this instance — state the rule that covers all of them, and fix that.

**The rule:** *a fence filter on plan-item text belongs at the extraction boundary (`PlanSectionBody`), not at individual call sites; and any code that reads or writes `- [s] …` rows must take its window from `SectionLineBounds`/`SectionByteBounds`, never from the whole body.* Measured prevalence — 5 read sites, 1 compliant; 1 write site, 0 compliant:

| site | filtered? |
| --- | --- |
| `internal/issue/plan.go:30` `CountPlanItems` | ✅ `StripFenced` |
| `close.go:568` plan-unchecked close gate | ❌ |
| `close.go:1727` `findMilestonesMissingVerdict` | ❌ |
| `internal/issue/structural.go:160` `checkPlan` | ❌ |
| `internal/issue/sizing.go:63-65` `PlanItems`/`Milestones` | ❌ |
| `close.go:558` milestone tick (**writes**, whole-body regex) | ❌ — not even scoped to `## Plan` |

Measured on one document whose `## Plan` holds a fenced example plus one real ticked row: `CountPlanItems` → (1 total, 1 ticked) so `sdlc state` reports 100%, while `close.go`'s guard sees **1 unchecked** and refuses with the quoted `- [ ] M9` in the error text. `sizing` sees 3 items. `checkPlan` accepts a Plan whose *only* items are fenced examples (false pass, the issue's own defect class). `findMilestonesMissingVerdict` would demand a `Review-Verdict:` commit for a milestone that exists only inside a code fence — an unsatisfiable refusal escapable only with `--no-verdict`. Before M2 all five agreed (all wrong); M2 made them disagree.

The write site is the sharpest instance and it fired in this very window: `pat := ^(- )\[[ .]\]( M2\b)` over `newBody` ticks a `- [ ] M2` in a fenced example anywhere in the file (verified: 2 matches, both rewritten). And `workshop/issues/000211-*.md:32` — the Problem section's worked example — was changed this commit from `- [ ] M2 — NOT done` to `- [x] M2 — NOT done` by an unanchored replace, breaking the demonstration (the table two lines below still says plan-unchecked truth = **2** open items; the example now shows 1, and a row labelled "NOT done" reads as ticked). The commit that closes the class committed an instance of it.

*Fix sketch:* strip inside `PlanSectionBody` (all five readers want the stripped body; none needs offsets into the original), and scope the milestone tick to `SectionByteBounds(body, "Plan", …)` with fenced lines excluded. Then delete `StripFenced` from `plan.go:30` — it becomes redundant.

**I2 — the indent-policy parameter is wired to nothing, and its comment claims a caller that doesn't exist** (`cmd/sdlc/internal/issue/fence.go:68-82, 128-133`)

`maxIndentAny` has **zero** references outside its own declaration. `fenceSpans(…, maxIndent)` and `fenceMarkerIndent(…, maxIndent)` are each called only with `commonMarkMaxIndent`. The doc says "maxIndentAny disables the check for SplitFences" — `SplitFences` (`structural.go:263`) uses `strings.Index` and never calls either function. So the comment describes a wiring that does not exist, and the parameter is documentation dressed as code — the "reads as protection while doing nothing" shape.

Compounding it: the Plan row `- [x] M2 — Decide SplitFences' line-anchoring change explicitly … **with a test pinning the choice**` is ticked, but no test pins the indent axis. `structural_test.go` is untouched in this window; the only policy test added (`TestUnterminatedPolicies_DisagreeOnPurpose`) pins the *unterminated* axis, and character-orientation is pinned by a pre-existing #179 case. The axis the Spec singled out as the actual `migrate` behaviour change — a 4-space-indented ``` is a fence for `SplitFences` and prose for `FenceSpans` — is pinned by nothing.

*Fix sketch:* delete `maxIndentAny` and the `maxIndent` threading, and add the missing pin to `TestSplitFences`: `{"indented fence is still a fence", "text\n    ```\nx\n    ```\n", […]}` plus the counterpart assertion that `FenceSpans` classifies the same lines as prose. That makes the divergence explicit *by test*, which is what the plan row promised — one real test beats a parameter no caller passes. (ARCH-DRY / Simplicity First.)

**I3 — `planGateContent`'s fence-awareness is unpinned by any test** (`cmd/sdlc/changecode.go:740-751`)

One of the three sites in the M2 sweep. I confirmed the fix is reachable and correct (an issue quoting `## Estimate` inside a fence keeps its real Problem prose in the hash), but nothing tests it — `changecode_test.go:602` only exercises estimate-edit invariance and passes either way. Per #194 a fix is complete only when a test fails without it; revert `!inside[i] &&` and no test reds.

*Fix sketch:* an issue body whose `## Problem` quotes ` ```markdown / ## Estimate ` followed by real prose; assert the real prose survives `planGateContent`. Reds without the fence check.

**I4 — the doc block above `insertLogLine` still documents the design that was just removed, and three docs name a symbol that doesn't exist** (`close.go:292-299`, `close.go:330-336`; `atlas/workflow/issue-lifecycle.md:165`)

`close.go:292` still reads "Anchor: the **last** `## Log` header, not the first. … All offsets below are taken relative to that last header" — three lines above the new comment saying the opposite. `close.go:331` repeats "so it anchors to the last header". `close.go:335` references `logHeaderRE`, deleted from `setstatus.go` in this same commit — a grep for it now returns only that comment. Separately, `stripEstimateForHash` does not exist anywhere in the tree (the function is `planGateContent`), yet it is named in the atlas, the Plan row and the M2 Log. This is the same class as the M1 review's I6 ("so a reader who greps finds what is actually there"), recurring in the commit that closed it.

Also overstated: the atlas asserts "**Everything that finds a heading is fence-aware**" — `issue.go:530` (`sdlc issue show`'s structure peek) is not, and the Plan deliberately lists it as unfixed. Fine as a decision; the atlas sentence should carry the exception.

*Fix sketch:* delete the superseded paragraph rather than layering a correction under it; rename `stripEstimateForHash` → `planGateContent` in atlas + Plan + Log; qualify the atlas sentence with the `issue show` carve-out.

## 4. Minor findings

- `fence_test.go:220` — `TestSectionByteBounds_MatchesSectionBody` trims trailing `\n` from **both** sides before comparing, so it does not pin the byte-identity `section.go:87` explicitly claims. (Identity does hold — 2886/2886 corpus sections — so this is coverage, not a bug.)
- The corpus property test covers `SectionBody` only; `SectionByteBounds`/`SectionHeadingByteOffset` carry all the index arithmetic and are exercised by four hand-written bodies. Folding them into the existing corpus walk is ~6 lines.
- `close.go:335`'s "shouldn't happen in practice" fallback is now unreachable by a different argument than the comment gives (the heading came from `SectionLineBounds`, not a regex) — worth restating or dropping.
- The M2 Log doesn't record re-running `sdlc issue validate --all` after the `stripCodeFences` rebase, which is a Done-when item and the one change in this window that could move a gate verdict. It's safe (measured above: 1 file ±1 word, 0 flips) — just unrecorded.

## 5. Test coverage notes

Two of the three swept sites are properly pinned and revert-verifiable (`logHasEntryToday`, `insertLogLine`-anchor). The gaps are where the bug this diff could ship actually lives: no test for `planGateContent` fence-awareness (I3), no test for the `insertLogLine` search *window* (C1 — the existing one passes incidentally), no test for the plan-counter divergence between `sdlc state` and `sdlc close` (I1), no test for the milestone-tick write hitting a fenced row (I1), no test for the indent divergence the plan row promised (I2). All five are cheap table/fixture cases in files that already exist. The corpus property test remains the strongest asset here — it is a genuine invariant over supplied inputs, not a golden.

## 6. Architectural notes

- **ARCH-DRY — flag (I1, I2).** The scanner consolidation itself is exemplary: one `FenceSpans`, one `SectionLineBounds` shared with `project` via a policy parameter, and a *reasoned* exception for `SplitFences`. But `StripFenced` is applied ad hoc at one call site where the shared boundary (`PlanSectionBody`) would cover all five, and `fenceSpans`/`fenceMarkerIndent` are duplicated-by-parameter for a caller that never arrives.
- **ARCH-PURE — pass.** `FenceSpans`, `StripFenced`, `SectionByteBounds`, `planGateContent`, `insertLogLine`, `logHasEntryToday` are all pure string functions tested directly. The corpus tests read the filesystem, but they are explicitly corpus-*seeded* with invariant assertions and skip when the corpus is absent — the right shape.
- **ARCH-PURPOSE — flag (C1, I1).** M2's stated purpose is sweeping the **class**, and the Log/atlas assert it is swept. Three claims are the instance rather than the class: the counters (1 of 5), the log-line window (anchor but not bounds), and the `- [ ] Mx` write path (never enumerated at all). The enumeration for the read sites was already written in the Spec's grep table — it just wasn't re-run for the counter filter.
- **ARCH-MOCK — pass / N/A.** No external binary or service surface changes here. `milestoneHasVerdictCommit` still shells to git behind the existing `gitx` seam, untouched.
- **ARCH-CONSTRAINTS — pass.** Line scanning over kilobyte documents; the corpus test runs 406 files × 2886 sections in 0.85s. `off()` is O(n) called twice per `SectionByteBounds` — not a hot path. No unbounded fan-out or blocking work introduced.
- **ARCH-SECURE — flag (C1).** Issue files are hand-edited, cross-version, operator-authored input, and the read paths degrade *visibly* here (over-segmentation under `UnterminatedIsProse`, pinned by its own test — good). The write paths do not: C1 splices bytes inside a fenced block and I1's milestone tick rewrites a quoted example, both silently. Where this diff parses untrusted-shaped input it is careful; where it writes, it inherits an unbounded window.

## 7. Plan revision recommendations

Append a `## Revisions` entry to `workshop/issues/000211-*.md` covering:

1. **Done-when correction.** "`stripCodeFences` and **the plan-item counters** are built on it" is not delivered — 1 of 5 counters filters, and the milestone-tick writer is unscoped. Either sweep the enumeration in I1 or narrow the Done-when to name which counters are in scope and why the rest are deferred.
2. **The `- [x] M2 — Decide SplitFences' line-anchoring change … with a test pinning the choice` row** is ticked without its test half. Either add the indent-axis test (I2) or restate the row as "decision recorded in prose; no test pins the indent axis" so the checkbox stops claiming coverage.
3. **Rename `stripEstimateForHash` → `planGateContent`** in the M2 Plan row, the M2 Log paragraph, and `atlas/workflow/issue-lifecycle.md` — the symbol does not exist.
4. **Restore `- [ ] M2 — NOT done`** at `000211-*.md:32`. The Problem-section example was ticked by an unanchored replace this commit; as it stands the worked example contradicts the guard table directly beneath it (`truth: 2` open items vs 1 shown), and the row labelled "NOT done" reads as done.
5. **Add the milestone-tick writer to the class enumeration** in the M2 row — `close.go:558` is a member the sweep never listed, and it is the only one that writes.

```findings
dispose:
  - id: BR-1
    disposition: addressed
    note: |
      Plan row 3 now reads "six sites"; verified the Spec table enumerates exactly six and the Estimate row 2 and Log both agree — 3 of 3 references consistent with the single grep-produced source.
  - id: BR-2
    disposition: addressed
    note: |
      TestSectionBody_CorpusLosesNoRealSection states corpus-supplies-inputs / assertion-is-invariant explicitly and carries no golden; workshopMarkdown names its repo-root resolution and skips on an absent corpus.
findings:
  - id: new
    severity: Critical
    family: unbounded-section-scan-window
    title: |
      insertLogLine locates the `## Log` heading fence-aware but scans for the day header to EOF, still filing the close line inside a quoted fence
    detail: |
      close.go:312 sets section = body[logStart:] ("the real Log section + anything after it"), so dayRE matches a `### <today>` inside a fenced example in a later section. Reproduced: with the real Log lacking today's day header and an `## Appendix` quoting `### 2026-09-02` in a ```markdown block, the close line is written inside the fence and logHasEntryToday cannot see it. TestInsertLogLine_TargetsTheRealLogSection passes only because its fixture quotes 2020-01-01 while the inserted line is dated 2026-09-02 — a test written from the same mental model as the fix (#194). SectionByteBounds, added in this same diff, already returns the end; bound both the dayRE and insertRE searches to it. ARCH-SECURE: the write path degrades invisibly on hand-authored input.
  - id: new
    severity: Important
    family: consumer-enumeration-incomplete
    title: |
      The plan-item fence filter reached 1 of 5 read sites and 0 of 1 write sites, so `sdlc state` and `sdlc close` now disagree about the same Plan
    detail: |
      2nd finding in family consumer-enumeration-incomplete — do not fix the instance. RULE - the fence filter belongs at the extraction boundary (PlanSectionBody), not at individual call sites, and any code reading or writing `- [s] ...` rows must take its window from SectionLineBounds/SectionByteBounds rather than the whole body. Measured - plan.go:30 CountPlanItems filtered; close.go:568 plan-unchecked, close.go:1727 findMilestonesMissingVerdict, structural.go:160 checkPlan, sizing.go:63-65 all unfiltered; close.go:558 milestone tick writes over the whole body, not scoped to Plan at all. On one fixture CountPlanItems reports 1/1 (state says 100%) while close refuses citing a quoted `- [ ] M9`; checkPlan accepts a Plan whose only items are fenced examples; findMilestonesMissingVerdict would demand an unsatisfiable verdict commit for a fenced milestone. Verified the tick regex rewrites 2 of 2 rows including the fenced one — and workshop/issues/000211-*.md:32 was ticked by an unanchored replace this very commit, breaking the Problem-section demonstration (the table below still says truth = 2 open items). Fix once inside PlanSectionBody and scope the tick to the Plan section, then drop StripFenced from plan.go:30. ARCH-PURPOSE, ARCH-DRY.
  - id: new
    severity: Important
    family: unwired-policy-parameter
    title: |
      The indent-policy parameter has zero call sites and its comment names a caller (SplitFences) that never invokes it; the plan row promising a test pinning the choice is unpinned
    detail: |
      maxIndentAny (fence.go:74) is referenced nowhere outside its declaration; fenceSpans and fenceMarkerIndent are only ever called with commonMarkMaxIndent. SplitFences uses strings.Index and calls neither, so the doc at fence.go:68-71 describes a wiring that does not exist. Meanwhile the ticked Plan row "Decide SplitFences' line-anchoring change explicitly ... with a test pinning the choice" has no test for the indent axis — structural_test.go is untouched in this window and the only added policy test covers the unterminated axis. Delete the dead parameter and add the real pin - an indented ``` is a fence for SplitFences and prose for FenceSpans, asserted side by side. RULE - a policy parameter with no non-default call site is documentation, not code; the axis must be pinned by test, not by an unused constant. ARCH-DRY.
  - id: new
    severity: Important
    family: claimed-fix-unpinned-by-test
    title: |
      planGateContent's fence-awareness — one of the three swept sites — is covered by no test
    detail: |
      changecode.go:740-751 now skips `## ` lines inside fences when computing the plan-gate hash. Verified reachable and correct (an issue quoting `## Estimate` in a fence keeps its real Problem prose), but changecode_test.go:602 only exercises estimate-edit invariance and passes with the fence check reverted. Per #194 the fix is not complete until a test reds without it - assert that real prose following a fenced `## Estimate` survives planGateContent.
  - id: new
    severity: Important
    family: doc-drifts-from-code
    title: |
      insertLogLine's doc block still specifies the last-match anchor that was just deleted, and three docs name a symbol that does not exist
    detail: |
      close.go:292-299 still reads "Anchor: the last `## Log` header, not the first ... All offsets below are taken relative to that last header", three lines above the new comment saying the opposite; close.go:331 repeats it; close.go:335 references logHeaderRE, deleted from setstatus.go in this same commit. Separately stripEstimateForHash exists nowhere in the tree (the function is planGateContent) yet is named in atlas/workflow/issue-lifecycle.md:165, the M2 Plan row and the M2 Log. The atlas also asserts "Everything that finds a heading is fence-aware" while issue.go:530's structure peek deliberately is not. Same class as the M1 review's I6, recurring in the commit that closed it - delete the superseded prose rather than layering a correction beneath it.
  - id: new
    severity: Minor
    family: test-asserts-weaker-than-contract
    title: |
      TestSectionByteBounds_MatchesSectionBody trims trailing newlines from both sides, so it does not pin the byte-identity section.go:87 claims
    detail: |
      Byte-identity does hold — I verified body[start:end] == SectionBody(...) across all 2886 non-fenced sections in workshop/**/*.md with no panics — so this is a coverage gap, not a defect. Drop the TrimSuffix and fold SectionByteBounds/SectionHeadingByteOffset into the existing corpus walk; they carry all the index arithmetic and are currently exercised by four hand-written bodies.
  - id: new
    severity: Minor
    family: unrecorded-gate-measurement
    title: |
      The M2 Log does not record re-running `sdlc issue validate --all` after the stripCodeFences rebase, a Done-when item
    detail: |
      The rebase from `(?s)```.*?``` ` to line-based StripFenced is the one change in this window that could move a gate verdict. It is safe — comparing old against new over every `## Spec` in the corpus, 1 file moves (000147-*.md, 246 to 247 words) and 0 cross the >=50 threshold — but the measurement is absent from the Log the Done-when asks for.
```

---

## Re-review — 2026-09-02T21:03:11-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 7752d83f9b0a1a91b8efa932bcf19cde9d2b97b5..7752d83f9b0a1a91b8efa932bcf19cde9d2b97b5 |
| command | sdlc milestone-close --issue 211 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-02T21:03:11-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

The M2 read path is correct and two of the five prior findings are properly closed — BR-3 and BR-6 both go red on revert, which I verified in a scratch copy. What blocks the boundary is that BR-4's fix reaches the readers but not the writer, is pinned by no test at all, and leaves a corrupted artifact in the tree: reverting all four `PlanItemsBody` call sites (`close.go:572`, `close.go:1726`, `plan.go:26`, `structural.go:160`, `sizing.go:63`) leaves the entire suite exactly as green as before — the same #194 rule the finding was raised under, violated by the commit closing it. Meanwhile `sdlc milestone-close`'s tick still `ReplaceAllString`s every `- [ ] Mx` row anywhere in the body (verified: it ticks a fenced quoted row and reports "ticked M3 in `## Plan`" when the *only* match is fenced), and `workshop/issues/000211-…:32` still reads `- [x] M2 — NOT done`, contradicting the table three lines below it and the test fixture at `planfence_test.go:29`. BR-5 and BR-7 are each half done. None of this is a correctness bug in what ships, and all of it is cheap — but two consecutive non-blocking verdicts is how these evaporate, so this one blocks.

### 1. Strengths

- **`planfence_test.go:156` (BR-3) is a real pin.** Reverting `section := body[logStart:logEnd]` → `body[logStart:]` reds `TestInsertLogLine_IgnoresAQuotedDayHeaderLater`. The fixture is built the only way that bites (the real Log has *no* matching day header), and the comment saying why it asserts by order rather than exact whitespace is the right call.
- **`planfence_test.go:200` (BR-6) pins the invariant, not the symptom.** Reverting the `!inside[i]` guard reds it, and the final assertion — estimate-only edit ⇒ identical hash input — tests what the pass-through actually promises.
- **`section.go:75` `PlanItemsBody` is the right shape.** One extraction point beats five per-consumer filters, and the doc names the reader/writer split rather than leaving it implicit (`ARCH-DRY` pass on the read side).
- **`fence.go` is a clean pure core.** Two-pass, policy-parameterised, no IO; the whole package tests without mocks (`ARCH-PURE` pass). `maxIndentAny` / `fenceMarkerIndent` are fully gone — grep returns zero references.
- **`SectionByteBounds` byte-identity holds under stress.** I probed 9 hand-built edge shapes (CRLF, heading with no trailing newline, heading-immediately-followed-by-heading, unterminated tail, empty body): `body[start:end] == SectionBody(...)` in every case, no panics.

### 2. Critical findings

None.

### 3. Important findings

**I1 — `close_test.go:240,257,440` + `planfence_test.go:44,63`: the tests that mirror close's plan gates call the wrong extractor, so BR-4's fix is pinned by nothing.**
Measured: reverting all four production routings back to `PlanSectionBody` leaves `./cmd/sdlc/...` and `./cmd/sdlc/internal/issue/...` exactly as green as HEAD (same 6 environmental failures in my no-`.git` scratch, identical set). The comment at `close_test.go:238` — *"Routed through the production extractor (#211)"* — is now false; production's extractor for that path is `PlanItemsBody`. The only coverage is `fence_test.go:262`, a unit test on `StripFenced` itself, which passes whether or not any consumer calls it.
*Fix sketch:* route those five mirrors to `PlanItemsBody` **and** give `planWithFencedHeading` a fenced `- [ ] M9 …` row. I wrote that pin in scratch and confirmed it: green at HEAD, `CountPlanItems = (5,3), want (3,3)` on revert.
**This is the 2nd finding in family `claimed-fix-unpinned-by-test`.** The rule, which covers both instances: *a test standing in for a production path must call the same extractor that path calls, and its fixture must contain the shape that extractor exists to filter.* A mirror pinned to a sibling function gives coverage for a path nothing runs. `ARCH-MOCK` at-review lens: production flow and test flow must share the boundary.

**I2 — a structural search inside an extracted section still isn't fence-filtered; BR-3 bounded the window but not its contents.**
`close.go:328` `dayRE` now searches only the real Log section, but a fenced `### <date>` *inside* that section still captures the insert. Same shape at `project/guards.go:17` (`retro-recorded` passes on a quoted retro heading — a guard whose *presence* means pass, the dangerous direction) and `project/retro.go:8` (`LatestRetroDate`). Measured prevalence across all 409 `workshop/**/*.md`: 0 fenced day-headers in a Log section, 0 fenced retro headings — latent today, exactly as the original `## Plan` bug was latent when this issue was filed.
**This is the 3rd finding in family `consumer-enumeration-incomplete`.** Do not fix these three sites. The rule: *bounding a search to a fence-aware section does not make the search fence-aware — any structural match **within** an extracted section must run over `StripFenced(section)` or a `FenceSpans`-filtered line walk.* The enumeration that rule implies is the four `(?m)^###`/`^- \[` searches above plus `close.go:547`; write it once as a helper and sweep it.

**I3 — `close.go:547-552`: the milestone tick rewrites every matching row in the document, and its stated rationale is factually wrong.**
The fix commit says the writer "keeps the unfiltered body deliberately; it rewrites a line it already matched." It does not — `pat.ReplaceAllString(newBody, …)` rewrites *all* matches, anywhere in the body, with no Plan scoping and no fence filter. Demonstrated: a body with `- [ ] M2` quoted in a fenced Problem-section example and one real row ticks **both**; a body where the only `- [ ] M3` is fenced gives `n > 0`, so close reports `ticked M3 in … ## Plan` while the real Plan is untouched. Readers (`PlanItemsBody`) see only the real row, so writer and readers now disagree about the same Plan — the disagreement BR-4 was raised to end.
*Fix sketch:* scope the replace to `SectionByteBounds(body, "Plan", …)` and skip lines `FenceSpans` marks inside, then splice back. Same helper as I2.

**I4 — the plan artifact claims work that was deliberately not done, and points at a `## Revisions` section that does not exist.**
The issue file has no `## Revisions` section (headings are Problem, Spec, Done when, Estimate, Plan, Log), yet Done-when line 190 says *"a deliberate, recorded exception, not an oversight (see Revisions)"*. AGENTS.md §1 requires appending one when a plan artifact is revised mid-stream, and Done-when *was* rewritten in place (from "one fence scanner" to "one LINE-oriented fence scanner plus a recorded exception"). Separately, the ticked row at line 296 still reads *"Rebuild `stripCodeFences` + `SplitFences` on the scanner"* — `SplitFences` was deliberately not rebuilt, which the Log explains at length but the row still claims.
**This is the 2nd finding in family `doc-drifts-from-code`.** The rule covering it and BR-7: *when a decision reverses what an artifact says, delete or rewrite the superseded text in the same commit — never leave it standing with a correction layered beneath, and never leave a cross-reference pointing at a section you didn't write.* Applying that rule to the current tree yields the enumeration in the BR-7 note below.

### 4. Minor findings

- `close.go:317` calls `SectionByteBounds` a second time (`SectionHeadingByteOffset` already computed it) and discards its `ok`; were it ever false, `body[logStart:0]` panics. One call, then derive the heading offset. `ARCH-DRY`, `ARCH-SECURE` (a discarded ok on a splice path).
- `workshop/issues/000211-…:37` cites `close.go:563` for the plan-unchecked guard; it is now `close.go:572`.
- `atlas/workflow/issue-lifecycle.md:166-168` lists the fence-aware heading finders but omits `PlanItemsBody`, the single extraction point M2 actually introduced.

### 5. Test coverage notes

- **Pinned and verified:** BR-3 (`planfence_test.go:156`), BR-6 (`planfence_test.go:200`) — both red on revert.
- **Unpinned:** the whole BR-4 consumer routing (I1). This is the only fix in the window that a revert leaves green.
- **Unpinned:** the indent axis BR-5 asked for — `TestFenceMarker:26` pins that `    ```` ` is not a fence for the scanner, but nothing asserts that `SplitFences` still treats it *as* one, which is the `migrate` behaviour change the ticked plan row promises a test for.
- `TestSectionByteBounds_MatchesSectionBody:229` still `TrimSuffix`es both sides, so it does not assert the byte-identity `section.go:87` claims. Identity does hold (I re-verified on 9 edge shapes); this is coverage, not a defect.
- Suite state: `go test ./cmd/... ./pkg/...` green except `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory`, which fails on a missing `workshop/plans/000200-…` — pre-existing and tracked as #210, unrelated to this window.

### 6. Architectural notes

- `ARCH-DRY` — **flag** (I2, I3): the read side consolidated correctly, but the write side and the within-section searches each carry their own copy of "find a row/heading in markdown". One `StripFenced`-scoped helper closes all five.
- `ARCH-PURE` — **pass**: `fence.go`, `section.go`, `plan.go`, `planGateContent`, `insertLogLine` are all pure and tested without IO; the one filesystem-touching test (`TestSectionBody_CorpusLosesNoRealSection`) skips cleanly on an absent corpus.
- `ARCH-PURPOSE` — **flag** (I1, I3): the shadow-sweep finds five consumers deriving from `PlanItemsBody` and one writer that does not, plus a documentation surface (`atlas:166`) still asserting *"Everything that finds a heading is fence-aware"* while `issue.go:530` deliberately is not. A single-source change is not finished while a consumer restates the model by hand.
- `ARCH-MOCK` — **pass on dependencies, flag on seam** (I1): no new external binary or service. But the test flow and production flow no longer share the extraction boundary, which is the same property this principle protects.
- `ARCH-CONSTRAINTS` — **pass**: `FenceSpans` is O(lines) with one `[]bool`; the corpus test walks 409 files in 0.13s. `insertLogLine` scans the body three times per close (see the Minor), which is irrelevant at issue-file scale.
- `ARCH-SECURE` — **pass with one note**: issue bodies are hand-edited input from outside the process, and the `UnterminatedIsProse` choice is the right failure direction (fail toward seeing *more* of the document). The residual is the discarded `ok` in the Minor above; every other parse path degrades to a visible fallback rather than a fabricated value.

### 7. Plan revision recommendations

Add a `## Revisions` section to `workshop/issues/000211-…` (it is referenced but absent) with:

1. **Done-when narrowed, M2 round 1** — "one fence scanner" → "one LINE-oriented fence scanner plus a recorded exception." Reason: `TestSplitFences` pins character-oriented segmentation (inline pairs, mid-line boundaries) that a line classifier cannot express. Delta: `SplitFences` keeps its own scanner.
2. **Plan row rewritten** — line 296 "Rebuild `stripCodeFences` + `SplitFences` on the scanner" → "Rebuild `stripCodeFences` on the scanner; decide `SplitFences` explicitly (declined — see Log)". The row as ticked claims work the Log says was deliberately reverted.
3. **Restore the Problem-section demonstration** — line 32 back to `- [ ] M2 — NOT done`, and note in Revisions that an unanchored replace during M2 ticked it, which is why the table at lines 37-38 stopped matching the block above it.

```findings
dispose:
  - id: BR-3
    disposition: addressed
    note: |
      Revert-verified: restoring body[logStart:] reds TestInsertLogLine_IgnoresAQuotedDayHeaderLater.
  - id: BR-4
    disposition: not-addressed
    note: |
      Readers routed to PlanItemsBody, but the tick writer is unscoped, no test pins the routing, and issue line 32 is still ticked.
  - id: BR-5
    disposition: not-addressed
    note: |
      Dead parameter and false comment removed; the indent-axis pin the finding asked for is still absent while the plan row claiming it stays ticked.
  - id: BR-6
    disposition: addressed
    note: |
      Revert-verified: dropping the !inside[i] guard reds TestPlanGateContent_IgnoresAQuotedEstimateHeading.
  - id: BR-7
    disposition: not-addressed
    note: |
      close.go:292 and :331 fixed; close.go:340 still names logHeaderRE (dead), and stripEstimateForHash survives at atlas:168 plus issue lines 289 and 333.
  - id: BR-8
    disposition: not-addressed
    note: |
      Test still TrimSuffixes both sides; byte-identity independently re-verified on 9 edge shapes including CRLF.
  - id: BR-9
    disposition: not-addressed
    note: |
      M2 Log still records no validate run; I measured it — PlanItemsBody routing leaves `sdlc issue validate --all` byte-identical.
findings:
  - id: new
    severity: Important
    family: claimed-fix-unpinned-by-test
    title: |
      BR-4's consumer routing is pinned by no test — reverting all four call sites leaves the suite green
    detail: |
      2nd in family. RULE - a test standing in for a production path must call the same
      extractor that path calls, and its fixture must contain the shape that extractor
      exists to filter. Measured - close_test.go:240,257,440 and planfence_test.go:44,63
      all call PlanSectionBody while close.go:572, close.go:1726, plan.go:26,
      structural.go:160 and sizing.go:63 now call PlanItemsBody; close_test.go:238's
      comment still claims it routes "through the production extractor". Reverting every
      production routing to PlanSectionBody changes no test result. A correctly-written
      pin (fixture with a fenced `- [ ] M9` row, read through PlanItemsBody) reds on
      revert with CountPlanItems = (5,3) want (3,3) — I verified both directions in
      scratch. ARCH-MOCK - production and test flow no longer share the boundary.
  - id: new
    severity: Important
    family: consumer-enumeration-incomplete
    title: |
      Bounding a search to a fence-aware section does not make the search fence-aware — four within-section matchers remain unfiltered
    detail: |
      3rd in family — do NOT fix these sites individually. RULE - any structural match
      run INSIDE an extracted section must go over StripFenced(section) or a
      FenceSpans-filtered line walk, not the raw section text. Measured members -
      close.go:328 dayRE (a fenced `### <date>` inside the real Log section still
      captures the close-line insert, which is BR-3's defect one level further down),
      project/guards.go:17 retroHeadingRE (retro-recorded is a presence-means-pass guard,
      so a quoted retro heading satisfies it), project/retro.go:8 retroDateRE, and
      close.go:547 (see the sibling finding). Prevalence across all 409 workshop/**/*.md -
      0 fenced day-headers in a Log section, 0 fenced retro headings; latent today,
      exactly as the `## Plan` truncation was latent when this issue was filed. Write the
      helper once and sweep the enumeration.
  - id: new
    severity: Important
    family: consumer-enumeration-incomplete
    title: |
      The milestone tick rewrites every matching row in the whole body, and its "rewrites a line it already matched" rationale is false
    detail: |
      close.go:547-552 does pat.ReplaceAllString(newBody, ...) with no Plan scoping and no
      fence filter, so it rewrites ALL matches anywhere in the document — not one already
      matched line. Verified - a body quoting `- [ ] M2` in a fenced Problem-section
      example plus one real row ticks both (matches=2); a body whose ONLY `- [ ] M3` is
      fenced still yields n>0, so close prints `ticked M3 in ... ## Plan` while the real
      Plan is untouched. PlanItemsBody sees only the real row, so writer and readers now
      disagree about the same Plan — the disagreement BR-4 was raised to end. Scope the
      replace to SectionByteBounds(body, "Plan", ...) and skip FenceSpans-inside lines.
  - id: new
    severity: Important
    family: doc-drifts-from-code
    title: |
      Done-when cites a `## Revisions` section that does not exist, and a ticked plan row still claims SplitFences was rebuilt
    detail: |
      2nd in family. RULE - when a decision reverses what an artifact says, delete or
      rewrite the superseded text in the same commit; never layer a correction beneath it,
      and never leave a cross-reference pointing at a section you did not write. Measured -
      the issue file's headings are Problem, Spec, Done when, Estimate, Plan, Log, with no
      `## Revisions`, yet Done-when line 190 says "a deliberate, recorded exception ... (see
      Revisions)" and AGENTS.md section 1 requires appending one on mid-stream revision (the
      Done-when was rewritten in place). Plan line 296 is ticked claiming "Rebuild
      stripCodeFences + SplitFences on the scanner" while the Log explains SplitFences was
      deliberately not rebuilt. Same rule covers BR-7's residue - close.go:340 names the
      deleted logHeaderRE, and stripEstimateForHash (a symbol that exists nowhere) survives
      at atlas/workflow/issue-lifecycle.md:168 and issue lines 289 and 333.
  - id: new
    severity: Minor
    family: redundant-recompute-drops-error
    title: |
      insertLogLine computes the Log section twice and discards the second ok, leaving a latent slice panic
    detail: |
      close.go:310 calls SectionHeadingByteOffset (which internally calls
      SectionByteBounds), then close.go:317 calls SectionByteBounds again with `_` for ok.
      The two agree today because the arguments are identical, so the panic is
      unreachable — but body[logStart:0] is what a false ok would produce. Take
      start/end/ok from one SectionByteBounds call and derive the heading offset from it.
      ARCH-DRY, ARCH-SECURE.
  - id: new
    severity: Minor
    family: doc-drifts-from-code
    title: |
      Stale line reference and a missing entry in the atlas's fence-aware inventory
    detail: |
      workshop/issues/000211-...:37 cites close.go:563 for the plan-unchecked guard, now
      close.go:572. atlas/workflow/issue-lifecycle.md:166-168 lists the fence-aware
      heading finders but omits PlanItemsBody, the single extraction point M2 introduced.
```

---

## Re-review — 2026-09-02T22:25:55-07:00 (unknown)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 7752d83f9b0a1a91b8efa932bcf19cde9d2b97b5..468ab8cf469b557f1d3fc503a91a54857be64d22 |
| command | sdlc milestone-close --issue 211 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-02T22:25:55-07:00 |
| verdict | unknown |

## Review

Failed to authenticate: OAuth session expired and could not be refreshed

---

## Re-review — 2026-09-02T22:46:09-07:00 (REWORK)

| field | value |
|-------|-------|
| issue | 211 — SectionBody truncates at a fenced heading |
| repo | ariadne |
| issue file | workshop/issues/000211-sectionbody-truncates-at-a-fenced-heading.md |
| boundary | milestone M2 |
| milestone | M2 |
| window | 7752d83f9b0a1a91b8efa932bcf19cde9d2b97b5..468ab8cf469b557f1d3fc503a91a54857be64d22 |
| command | sdlc milestone-close --issue 211 --milestone M2 |
| reviewer | claude |
| timestamp | 2026-09-02T22:46:09-07:00 |
| verdict | REWORK |

## Review

```verdict
verdict: REWORK
confidence: high
```

Two of this window's three behavior changes are correct and one of them (`insertLogLine`'s day-header lookup + `logHasEntryToday`) is genuinely pinned — I reverted both and `TestWithinSectionMatchers_SkipFencesInsideTheSection` went red with the close line filed inside the quoted fence. But the round does not cross the boundary for three measured reasons. First, BR-11 named a four-member enumeration and the sweep covered two: `project/guards.go:17` `retroHeadingRE` and `project/retro.go:8` `retroDateRE` still scan `d.SectionBody("Log")` raw — the 4th consecutive round of `consumer-enumeration-incomplete`, and the atlas paragraph this diff *wrote* asserts the class-wide invariant those two break. Second, both new tests that stand in for a production path re-implement it instead of calling it: I reverted `close.go`'s milestone tick to the old whole-body `ReplaceAll` and `TestMilestoneTick_OnlyTicksTheRealPlan` still passed; I reverted `close.go:596`, `close.go:1750`, `sizing.go:63` and `structural.go:160` back to `PlanSectionBody` and the whole suite still passed — only `CountPlanItems` (1 of 5 readers) is actually pinned, against a test comment claiming all four are. Third, the M2 Log records "Both revert-verified with an explicit build check first" for BR-10 and BR-11; that is true for BR-11 and false for BR-10, and a gate that accepts a fabricated verification claim has no evidence property left.

### 1. Strengths

- **`FindLineOutsideFences` is the right abstraction at the right level** (`cmd/sdlc/internal/issue/fence.go:164`). Pure, no IO, offset-preserving, and it names the distinction that the whole family turns on — readers take `StripFenced`, splicers take offsets. The doc block states the rule rather than the symptom.
- **The BR-11 fix is real and pinned.** Reverting `close.go:328-333` to `(?m)` + `FindStringIndex(section)` *and* `setstatus.go:311` to `strings.Contains(section, today)` reds `planfence_test.go:310` on both assertions, with the failure output showing the close line spliced inside the ```` ```markdown ```` block. This is the shape every other fix in this round should have.
- **The tick's behavior change is correct and non-regressive.** I compared old-vs-new tick over all 209 issue files × M1–M5: 6 milestone cases with matches, 0 differing outputs, 0 files where the old path ticked but no `## Plan` section exists. The fix is latent-only on the corpus, as claimed.
- **`PlanItemsBody` is now the single extraction point in fact, not just in prose** — `plan.go:26`, `structural.go:160`, `sizing.go:63`, `close.go:596`, `close.go:1750` all route through it, and the double-strip in `CountPlanItems` is gone. The *routing* half of BR-4 is complete.
- **Honest self-correction in the commit body and in `close.go:564-570`** — naming your own earlier rationale as false, in the code, is the cheapest possible way to stop the next reader inheriting it.

### 2. Critical findings

None. No correctness bug ships in this diff; the blockers are enumeration, pinning and evidence.

### 3. Important findings

All are dispositions of open prior findings — see the `dispose:` block. Summarised:

- `project/guards.go:17` + `project/retro.go:8` — 2 of BR-11's 4 enumerated members unswept.
- `planfence_test.go:256` and `planfence_test.go:350` — tests assert against their own copy of the algorithm; measured green on production revert.
- Log's "Both revert-verified" claim for BR-10 — measured false.

### 4. Minor findings

- `fence.go:156-157` — the doc says "byte range … of the first **line** matching re"; the code returns the **match** range (`off+loc[0]`, `off+loc[1]`). They coincide only because the sole caller anchors `^…$`. An unanchored regex hands a splicing caller a mid-line offset.
- `close.go:308` / `close.go:315` — BR-14, unchanged.
- `fence_test.go:228` — BR-8, `TrimSuffix` on both sides still weakens the byte-identity assertion.
- `close.go:342` names `logHeaderRE`; `atlas:184` and issue `:289`/`:369` name `stripEstimateForHash`. Neither symbol exists.
- `workshop/issues/000211-…:32` still reads `- [x] M2 — NOT done` under a table asserting truth = 2 open items (it is 1). `:37` and `:143` still cite `close.go:563`; the guard is at `close.go:596`.

### 5. Test coverage notes

- `go test ./cmd/sdlc/...` at HEAD: one failure, `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` — `workshop/plans/000200-sdlc-fleet-thread-inventory-plan.md` was archived to `workshop/history/plans/`. I confirmed the path is absent at the base commit too, so it is pre-existing and outside this window, but the suite is red at this boundary and a bare "tests pass" claim in `--verified` would be false.
- The gap that matters: **no test drives `computeClose` through the milestone tick.** `grep computeClose` over `*_test.go` returns only `closereview_test.go:419`'s comment. The tick is business logic living inside an IO-shell function, which is why the test copied it instead of calling it.
- `TestPlanItemReadersAgree`'s fixture is right (a fenced `- [ ] M9` / `- [x] M8` pair) — only the routing is wrong. Pointing its three later assertions at `findMilestonesMissingVerdict` and the real unchecked guard, instead of at a `planBody` the test extracted, would make it the pin its own comment promises.

### 6. Architectural notes

- **ARCH-PURE — flag.** `close.go:571-587` is pure string logic (split → `FenceSpans` → per-line replace → rejoin) embedded in `computeClose`, which needs a git repo, a repo lock and flag state to run. Extracting it as `issue.TickPlanItem(body, milestone) (string, int)` makes it unit-testable, removes the test's duplicate, and closes BR-10/BR-12's pinning gap in one move. This is the single highest-leverage change for the re-run.
- **ARCH-DRY — flag.** Three copies of the walk now exist: `close.go:574-586`, `planfence_test.go:275-288`, and (in spirit) `FindLineOutsideFences`/`StripFenced`. The consolidation is the extraction above.
- **ARCH-PURPOSE — flag.** BR-11 named the enumeration explicitly and the diff swept half of it. The `family:` slug has now fired four times; the ledger is reporting that the enumeration exists on paper but was never executed.
- **ARCH-MOCK — flag.** Production and test flow no longer share the boundary for the tick or for three of the five plan readers. Not an external-dependency issue, but the same failure mode the principle exists to prevent.
- **ARCH-CONSTRAINTS — pass.** All new work is O(lines) over KB-scale issue files; `logHasEntryToday`'s extra `StripFenced` allocation is negligible. No unbounded fan-out, no blocking work on an interactive path.
- **ARCH-SECURE — pass, with one note.** Issue bytes are untrusted input (hand-edited, older versions). Both new failure paths degrade visibly: `SectionByteBounds` → `ok=false` → tick warns and writes nothing; `FindLineOutsideFences` → `found=false` → falls back to top-of-Log. No fabricated values. The one latent hazard is BR-14's discarded `ok` (`body[logStart:0]` on a false negative), unreachable today because both calls share arguments.
- **Docs gate:** atlas updated for the new surface (the two helpers and the write-side rule are both documented) — pass on presence, fail on accuracy (`stripEstimateForHash`, and the unqualified "everything" claim). No new user-facing flag, subcommand or config key, so README is correctly untouched.

### 7. Plan revision recommendations

The issue file is the plan of record (no `Core concepts` table, so that cross-check is N/A). It needs a `## Revisions` entry, or an equivalent in-place correction, covering:

- The `## Problem` demonstration was corrupted at `:32` (`- [ ] M2` → `- [x] M2`) and the table at `:35-38` no longer matches it — restore the row or restate the truth column.
- `close.go:563` → `close.go:596` at `:37` and `:143`.
- `stripEstimateForHash` → `planGateContent` at `:289` and `:369`, and at `atlas/workflow/issue-lifecycle.md:184`.
- The M2 Log's "Both revert-verified" sentence — BR-11 was, BR-10 was not; say what was actually measured.
- The Done-when item at `:185` (`sdlc issue validate --all` before/after) is still unrecorded for the M2 window.

```findings
dispose:
  - id: BR-4
    disposition: addressed
    note: |
      All five readers route through PlanItemsBody and the tick is Plan-scoped; residual corrupted demo moved under BR-15.
  - id: BR-5
    disposition: not-addressed
    note: |
      maxIndentAny is gone (grep: zero references), but the promised indent-axis pin still does not exist — TestSplitFences (structural_test.go:188) has no indented case and no side-by-side FenceSpans counterpart.
  - id: BR-7
    disposition: not-addressed
    note: |
      close.go:292/:331 fixed; close.go:342 still names logHeaderRE (absent from the tree), atlas:184 still names stripEstimateForHash (function is planGateContent, changecode.go), and atlas:181's "Everything that finds a heading is fence-aware" is still unqualified while issue.go:530 is deliberately not.
  - id: BR-8
    disposition: not-addressed
    note: |
      fence_test.go:228 still TrimSuffix-es both sides; SectionByteBounds/SectionHeadingByteOffset are still exercised only by the four hand-written bodies, not the corpus walk.
  - id: BR-9
    disposition: not-addressed
    note: |
      No M2 Log entry records re-running `sdlc issue validate --all`; the only record (issue :434, :448) is M1's.
  - id: BR-10
    disposition: not-addressed
    note: |
      Measured by revert - reverting close.go:596, close.go:1750, sizing.go:63 and structural.go:160 to PlanSectionBody leaves the entire suite green. Only CountPlanItems (plan.go:26) reds, at (4,2) want (2,1). TestPlanItemReadersAgree extracts planBody itself and applies the regexes inline, so it pins 1 of 5 readers against a comment claiming 4.
  - id: BR-11
    disposition: not-addressed
    note: |
      2 of the 4 enumerated members swept (close.go:328 dayRE, close.go:547 tick) plus logHasEntryToday; project/guards.go:17 retroHeadingRE and project/retro.go:8 retroDateRE still match against raw d.SectionBody("Log").
  - id: BR-12
    disposition: not-addressed
    note: |
      The close.go:564-588 fix is correct and corpus-verified non-regressive, but unpinned - reverting it to the whole-body ReplaceAll leaves TestMilestoneTick_OnlyTicksTheRealPlan PASSING, because the test re-implements the loop rather than calling computeClose.
  - id: BR-13
    disposition: not-addressed
    note: |
      Done-when's Revisions cross-reference and the SplitFences plan row are both corrected; close.go:342 (logHeaderRE), atlas:184 + issue :289/:369 (stripEstimateForHash) survive, and section.go:72-74 still says the tick writer "keeps PlanSectionBody ... rewrites a specific line it already matched" — the rationale close.go:568 declares false in the same commit, about code that no longer calls PlanSectionBody at all.
  - id: BR-14
    disposition: not-addressed
    note: |
      close.go:308 and close.go:315 still compute the Log bounds twice, second call discarding ok.
  - id: BR-15
    disposition: not-addressed
    note: |
      Atlas now lists PlanItemsBody (addressed), but issue :37 and :143 still cite close.go:563 (guard is at close.go:596) and :32 still carries the tick-corrupted `- [x] M2 — NOT done` under a table asserting 2 open items.
findings:
  - id: new
    severity: Important
    family: unrecorded-gate-measurement
    title: |
      The M2 Log records a revert-verification for BR-10 that I measured to be false
    detail: |
      This is the 2nd finding in family `unrecorded-gate-measurement`. Do not fix
      the sentence - fix the rule. RULE: a Log line asserting evidence must name
      the command run and the observed result, and a "revert-verified" claim is
      only true when the named test went RED with the production change undone.
      Measured: issue :332-341 says BR-10 and BR-11 were "Both revert-verified
      with an explicit build check first". BR-11 is true (I reverted close.go's
      dayRE and setstatus.go's StripFenced; planfence_test.go:310 reds on both
      assertions). BR-10 is false (reverting close.go:596, close.go:1750,
      sizing.go:63, structural.go:160 leaves every test green). Prevalence in
      this issue: 2 of 3 revert claims in the M2 Log are unverifiable as written
      - the third, "Revert-verified on the live instance" at :368, names no test.
      A build check is not a revert check.
  - id: new
    severity: Important
    family: doc-drifts-from-code
    title: |
      FindLineOutsideFences documents a line range and returns a match range, on a splice-offset API
    detail: |
      This is the 4th finding in family `doc-drifts-from-code`. Do not fix this
      instance alone. RULE: every symbol name and every behavioral claim written
      into a comment, the atlas or an issue must be greppable-true against the
      tree at the commit that writes it; the enumeration is one grep per named
      symbol plus one check per "everything/all X" claim. Prevalence at HEAD, all
      four members verified by grep: fence.go:156 says "byte range of the first
      LINE matching re" while :171 returns off+loc[0], off+loc[1] (the match);
      close.go:342 names logHeaderRE, absent from the tree; atlas:184 and issue
      :289/:369 name stripEstimateForHash, absent (it is planGateContent);
      section.go:72-74 states a tick-writer rationale that close.go:568 calls
      false in the same commit. The FindLineOutsideFences case is the only one
      with a byte-corruption path: the sole caller anchors ^...$ so match == line
      today, but close.go:333 splices at the returned offset, so an unanchored
      regex from a future caller inserts mid-line into an issue file. Either
      return the line bounds or say "match range" and rename the returns.
```
