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
