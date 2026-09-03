---
id: 000211
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.71
started: 2026-09-02T18:20:40-07:00
---

# SectionBody truncates at a fenced heading

## Problem

`issue.SectionBody` (`cmd/sdlc/internal/issue/section.go:15`) ends a section with
`^## ` — which matches at any line start, including **inside a fenced code
block**. `PlanSectionRE` (`plan.go:15`) has the identical shape and the identical
bug. A section that quotes markdown containing a `##` heading is silently cut off
there.

**The severe consequence is a false PASS on close gates, not a false refusal.**
The word- and bullet-count checks are `≥ N` thresholds that truncation can only
push down, so they fail safe. Two gates count things whose *absence* means pass:

    ## Plan
    - [x] M1 — done
    - [x] Add the scaffold. Example of what it emits:
    ```markdown
    ## Some heading the issue is quoting          <- parser stops here
    ```
    - [x] M2 — NOT done
    - [ ] Wire the consumer

| guard | sees | truth |
| --- | --- | --- |
| plan-unchecked (`close.go:563`) | **0** open items | 2 |
| milestone-verdict (`findMilestonesMissingVerdict`) | `[M1]` | `[M1, M2]` |

So `sdlc close` would pass an issue with two unticked plan items and never demand
review evidence for M2. A quoted `##` anywhere in a Plan disarms both.

The false refusal is the visible half and how this was found: `sdlc change-code
--issue 208` refused with `` `## Spec` has 35 words; need ≥ 50 `` against a
600-word Spec, because #208 quotes the registry entries it adds.

**Why issues quote markdown at all** — this is structural, not a bad habit. In
this repo the deliverable often *is* a markdown document, so specifying one means
showing it verbatim: #208 quotes the two `## ARCH-*` registry entries, #030 the
target-datatype template, #035 the `## Postmortem` section the verb writes, #066
the `## Log` line `close` appends. The quoted headings are `##` because the
*target file* uses `##`. The tracker and the artifacts share a format, so the
tracker's delimiter appears inside its own content. Telling authors "don't do
that" is not a policy anyone can follow: this issue dodges its own bug only
because it happens to use 4-space indented blocks rather than fences.

**Measured exposure:** 6 of 209 issue files (active + history) currently quote a
`##` inside a fence; `markdown` is the 5th most common fence language in the
corpus. No live misverdict today — no `## Plan` in the corpus quotes a fenced
heading — so this is latent, and the rate rises with exactly the registry/
datatype/helptext work that has been accelerating.

## Spec

**Not a better regex.** Go's `regexp` is RE2: no backreferences (verified —
`invalid escape sequence: \1`), no lookahead. CommonMark's closing fence must be
the same character and *at least as long* as the opener, which is a matched
delimiter and therefore not regular. A regex that consumes fenced blocks as units
handles plain, tilde, unterminated and indented fences, and still gets a
a four-backtick fence containing a three-backtick line wrong. Measured 4 of 5.

**The real shape is consolidation, not a new feature.** The tree already has
three fence scanners and they disagree:

| implementation | backticks | tildes | width rule | unterminated |
| --- | --- | --- | --- | --- |
| `issue.fencedCodeRE` (`stripCodeFences`) | yes | no | no | prose |
| `issue.SplitFences` (#179 migrate) | yes | no | no | fenced |
| `project.scanMarkdownLines` | yes | yes | yes | fenced |

`SectionBody` and `PlanSectionRE` use none of them. Adding a fourth is the wrong
move (`ARCH-DRY`); `project.scanMarkdownLines` is the only CommonMark-correct one
and becomes the single source.

**Direction of the move.** `internal/project` imports `internal/issue`, so the
scanner moves DOWN into `issue`; `project` consumes it from the new home. Its
existing tests pin that this is behaviour-preserving on the project side.

**The unterminated-fence policy is the dangerous part, and it is a PARAMETER.**
`scanMarkdownLines` treats an unterminated fence as running to end-of-file, so
every heading after it disappears. Adopting that for `SectionBody` would be
strictly worse than the bug being fixed: instead of one section truncated, the
whole remainder of the issue vanishes — including `## Plan`, which is what the
close gates count. That is the same false pass, unbounded.

This is not hypothetical. **This issue's own first rewrite created one**: a line
of prose beginning with four backticks read as an opener, and under that policy
`## Done when`, `## Plan` and `## Log` all became invisible. Measured across the
209 issue files, an unterminated fence is the only case where the scanner's
policy removes a *real* section — the other five affected files lose only headings
that are genuinely quoted inside closed fences, which is correct.

So the scanner exposes the policy and each consumer sets it:

| consumer | unterminated fence ⇒ | why |
| --- | --- | --- |
| `SectionBody` / plan extraction | **prose** | a swallowed `## Plan` disarms the close gates; failing open on a malformed fence is worse than the truncation this issue fixes |
| `stripCodeFences` (word count) | **prose** | today's behaviour, deliberate — see its comment |
| `SplitFences` (#179 migrate) | **fenced** | today's behaviour, deliberate — a rewriter must not edit inside a possibly-code tail |
| `project` section scan | **fenced** | today's behaviour; unchanged by the move |

Implementation follows from that: fence spans are computed in one pass, and a
fence with no closer is simply not a span under the `prose` setting.

**Divergence axes (`SplitFences` is not just a policy change).** `SplitFences`
finds fences with an unanchored, indent-blind `strings.Index("```")`, so it
currently treats a **4-space-indented** triple backtick as a fence, while
`fenceMarker` requires line-start with ≤3 spaces of indent — verified. Re-basing
it flips those to prose and `migrate` would begin rewriting refs inside indented
code blocks. It also guarantees byte-exact reassembly (`Concatenating the Text of
every segment reproduces the input byte-for-byte`), which a line-visitor that
drops fence lines cannot provide without deliberate reconstruction. Every axis is
stated per consumer before any is changed:

| axis | `SplitFences` today | `fenceMarker` |
| --- | --- | --- |
| line-anchored | no (substring) | yes |
| indent | any | ≤3 spaces |
| tilde fences | no | yes |
| closer width rule | no | ≥ opener |
| unterminated | fenced | caller's choice |
| byte-exact reassembly | guaranteed | must be reconstructed |

`SplitFences` keeps byte-exactness and gains the width/tilde rules; whether it
adopts line-anchoring is a **behaviour change to `migrate` and is decided
explicitly**, with a test either way.

**`PlanSectionRE` is deleted, not fixed.** Enumerated by grep, not by memory —
an earlier draft named four and missed two:

| site | note |
| --- | --- |
| `close.go:563` | plan-unchecked guard |
| `close.go:1718` | `findMilestonesMissingVerdict` |
| `structural.go:160` | `checkPlan` |
| `sizing.go:63` | plan/milestone counts |
| `plan.go:30` | **`CountPlanItems`** — feeds `state.go:255`, same truncation bug |
| `close_test.go:440` | stops compiling on deletion |

Every production site takes `FindStringSubmatchIndex` only to slice
`body[m[2]:m[3]]` and work on the string, so each becomes
`SectionBody(body, "Plan")`. Two comments go stale with it and are part of the
change: `section.go:10-11` ("checkPlan keeps its own PlanSectionRE: it needs byte
offsets" — it does not) and `structural.go:23`.

**Preserve the deliberate disagreement.** `stripCodeFences` and `SplitFences`
differ on an unterminated fence on purpose — prose for the word-count gate,
fenced for the migrate rewriter, and both comments say so. Rebuilding them on one
scanner must make that a parameter, not erase it. This is the one place a
careless merge silently breaks #179.

**Out of scope:** CommonMark conformance beyond fenced blocks (setext headings,
HTML blocks, link reference definitions). The goal is that a fenced example stops
lying to the gates, not a compliant parser.

## Done when

- A section whose body contains a fenced block with `## ` inside is extracted in
  full, across all five fence forms: ``` and ~~~, a wider fence holding a
  narrower line, an unterminated fence, an indented fence — plus the adversarial
  "closed fence, then a real heading".
- `sdlc close` refuses an issue whose `## Plan` has unchecked items **after** a
  fenced `##`, and `findMilestonesMissingVerdict` sees milestones after one. A
  test drives both; this is the false pass the issue exists for.
- An unterminated fence NEVER hides a later `##` section from `SectionBody`. A
  test builds an issue whose `## Spec` opens a fence and never closes it, and
  asserts `## Plan`'s unchecked items are still counted by the close gate.
- One LINE-oriented fence scanner remains in `cmd/sdlc`: `PlanSectionRE` is
  gone, `stripCodeFences` and the plan-item counters are built on it, and the
  per-consumer unterminated-fence policies are asserted by a test that names why
  they differ. `SplitFences` keeps its own character-oriented scanner — a
  deliberate, recorded exception, not an oversight (see the M2 Log entry).
- `internal/project`'s behavior is unchanged: its existing suite passes with no
  edits beyond the import path.
- `sdlc issue validate --all` over `workshop/issues/` + `workshop/history/` is
  run before and after; any issue whose verdict CHANGES is listed in the Log.
  (Predicted: none. Measured before starting — `checkSpecWordCount` already
  strips fences, so 5 files' counts rise (e.g. #208 134 → 600) with zero
  spec-present flips across 209 files. A flip would mean the prediction was
  wrong and wants explaining, not silently accepting.)

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: smaller-go-module      design=0.04 impl=0.14
item: cross-cutting-refactor design=0.04 impl=0.20
item: smaller-go-module      design=0.05 impl=0.20
item: greenfield-go-module   design=0.05 impl=0.28
item: atlas-docs             design=0.02 impl=0.06
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
total: 1.71
```
In Plan order:

1. `smaller-go-module` **at 0.14, not the ceiling** — moving
   `scanMarkdownLines` + `fenceMarker` down is mirror-or-extend of working code;
   only the policy parameter is new. An earlier cut put every code row at exactly
   `0.5 × 0.40`, which erased the very risk difference the narrative claimed.
2. `cross-cutting-refactor` at the ceiling — `SectionBody` rebuilt,
   `PlanSectionRE` deleted, six sites and two stale comments rerouted.
3. `smaller-go-module` at the ceiling — **the riskiest row**: byte-exact
   reassembly must survive, the two consumers must keep disagreeing about
   unterminated fences on purpose, and `SplitFences`' line-anchoring change
   alters what `migrate` rewrites.
4. `greenfield-go-module` — the corpus-seeded property test over all 406
   `workshop/**/*.md` is **new machinery, not a table**, which is what that slug
   is for; scaled band 0.12–0.32, priced near the top because the invariant
   (no file loses a real section, visited+skipped reproduces the input) has to be
   designed, not just asserted.
5. `atlas-docs` — the section-parsing entry and the single-source note. Design
   discounted like every other row this time.
6. `milestone-review` ×3 — see below.

Design is `×0.2` spec-quality discounted throughout: the Spec names every
consumer by file:line from a grep, tables the six divergence axes, and fixes each
consumer's unterminated-fence policy. Buffer `+15%` (v3.1 step 4). Familiarity
`1.0`: Go, this package, line scanning.

**Three `milestone-review` rows, and the Plan is tagged `M1`/`M2` to match.** The
first cut priced two review rows against an untagged Plan, which is incoherent:
in this workflow an `Mx` tag *is* the review boundary, and the primitive is "one
milestone code review (one chunk)" — so rows must equal boundaries, not rounds.
The scope genuinely has two: M1 is the scanner and section extraction (verifiable
as "no section is lost"), M2 is the fence consumers (verifiable as "migrate still
rewrites the same refs"). Different risk, different evidence, worth closing
separately.

The third row is an explicit **rework allowance**, not a fourth boundary. Both
ledger rows to date under-ran on exactly this: #206 estimated 1.31 → actual 4.43
(0.30) having priced one review for six rounds; #208 estimated 1.13 → actual 1.72
(0.66) having priced two and taken two. Rework rounds are re-review of the same
chunk, which the primitive does not name, so they are budgeted here rather than
discovered at close.

**Not applying the realized ratios directly.** Extrapolating #208's 0.66 onto
this total gives ~2.3h and #206's gives ~5h, but those ratios largely encode
estimation errors this block has now corrected — undifferentiated ceilings, a
mirror-or-extend row priced as greenfield's neighbour, review rows that did not
match the boundary structure. Multiplying by the ratio on top of fixing its
causes double-counts. The rework row is the part of the miss that is
*systematic*, so that part is budgeted; the rest is left to be measured.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

Two review boundaries, because they carry different risk and different evidence:
M1 is verifiable as "no section is lost", M2 as "migrate still rewrites the same
refs". Tagged so the structure and the `## Estimate` agree.

- [x] M1 — Move `scanMarkdownLines` + `fenceMarker` from
      `internal/project/doc.go` into `internal/issue`; `project` consumes it from
      there. Give the scanner an explicit unterminated-fence policy.
- [x] M1 — Rebuild `SectionBody` on the scanner with the `prose` policy; delete
      its regex terminator.
- [x] M1 — Delete `PlanSectionRE`; route its six sites and two stale comments
      through `SectionBody(body, "Plan")`.
- [x] M1 — Corpus-seeded property test over all 406 `workshop/**/*.md`: no file
      loses a real section, and visited + skipped lines reproduce the input.
      Fold the before/after verdict diff into it so the invariant is mechanical.
- [x] M1 — `close` plan-unchecked + `findMilestonesMissingVerdict` regression:
      unchecked items and milestones after a fenced `##` must be counted. Plus an
      unterminated fence that must NOT hide later sections.
- [x] M2 — Sweep the rest of the heading-finding class (M1 review I4). Three
      production sites still locate `## <heading>` without fence awareness:
      `setstatus.go:302` `logHasEntryToday` (**live instance** — takes the FIRST
      `## Log`, which in `workshop/history/issues/000066-*.md` is at line 22
      inside a fence while the real one is at line 68, so the reopen guard reads
      the wrong region on a real file); `close.go:301` `insertLogLine` (whose
      last-match heuristic from #66 is a second, weaker workaround for exactly
      what `FenceSpans` now solves, and which fails for a fenced `## Log` after
      the real one — 1 of 406 corpus files already has that shape); and
      `changecode.go:721` `planGateContent`, whose own comment reasons about
      RE2's lack of lookahead. (`issue.go:530`'s structure peek is cosmetic —
      listed, not fixed.)
- [x] M2 — Filter the plan-item counters on `FenceSpans`: fenced `- [ ]` /
      `- [x] Mx` lines inside a quoted block now survive into `planBody` and are
      counted. Fails safe (spurious refusal), but it is free to be right.
- [x] M2 — Rebuild `stripCodeFences` on the scanner. `SplitFences` was
      attempted and DECLINED — its contract is character-oriented; see the Log.
- [x] M2 — `SplitFences` decision recorded at the function, in the atlas, and
      in the Log; `TestSplitFences` pins the character-oriented contract that
      makes it a different abstraction.
- [x] M2 — `fenceMarker` table over width/char/indent/info-string boundaries;
      one test each pinning why `stripCodeFences` and `SplitFences` differ.
- [x] M1 — Atlas: the section-parsing entry and the scanner's single-source note.
      (Moved from M2: the milestone-close atlas gate fired on M1, correctly — the
      architectural surface is the scanner and its policy, which landed here.)

## Log

### 2026-09-03 — close review round 8: converging, and one more of my own

Round 8: 0 repeat families, 9 disposed. Two new Minors, both real and both fixed:

- **A pure decision routed through IO.** Pinning the milestone scan on the
  production path made it call `findMilestonesMissingVerdict`, which shells to
  `git log` per milestone — so this issue's *central* regression failed outside a
  git worktree, on an error raised after the fact under test was already decided.
  Extracted `milestonesInPlanOrder` (pure); the test drives that. Same `ARCH-PURE`
  shape as `TickMilestone`, and I had just made the opposite mistake fixing the
  previous finding: moving a test onto the production path is right, moving it
  onto the production *IO* is not.
- **Guard keys were basename-scoped across two packages.** No collision today,
  but `internal/issue` already has `plan.go`/`section.go`/`structural.go`, and a
  future `close.go` there would have merged two functions' call sets and reported
  an edge satisfied by the wrong one. Keys are directory-qualified now.

**And the extraction quietly broke the guard I had just built.** Moving
`milestonePlanRE` into the pure helper removed it from `findMilestonesMissingVerdict`,
which dropped that caller from the derived reader set — the derivation finds
*regex users*, and the caller had stopped being one. I wrote an exemption saying
the helper was safe because "its caller is itself checked by this guard", then
probed it: reverting that caller fired nothing. **The rationale was false**, the
third such of mine this issue.

Fixed by covering what the derivation structurally cannot: `planItemBodySources`
names the callers that *obtain* a body and delegate the counting.
Revert-verified — routing `findMilestonesMissingVerdict` back to
`PlanSectionBody` → `close.go:findMilestonesMissingVerdict no longer calls
PlanItemsBody`.

Worth naming as a pattern: **extracting a helper can move a call site out of a
derived guard's population.** The derivation keys on a property (references a
counting regex) that refactoring relocates, so a refactor that looks purely
structural can silently shrink what the guard covers.

**BR-7 and BR-11 verified resolved in the tree**, not disposed by the ledger:
`logHasEntryToday` uses `StripFenced` and `insertLogLine` uses
`FindLineOutsideFences`, both revert-verified via
`TestWithinSectionMatchers`; the only remaining `logHeaderRE` /
`stripEstimateForHash` mentions are in prose *describing* the fix (this Log and
the auto-appended verification line), which necessarily name the old symbols.

### 2026-09-02 — close review: the same rule, measured three more ways

BR-21 is the fourth finding in `claimed-fix-unpinned-by-test`, and it caught a
**second** false revert claim of mine — this one written in the entry directly
below, hours after the first was corrected. The rule it states is now the
operative one:

> A fix is pinned only when reverting EXACTLY that fix — that filter, that call
> site, that return value — reds a named test.

Three corollaries, one per member, each measured before I accepted it:

1. **A fixture both filters happen to catch pins only their union.**
   `TickMilestone` installs two independent filters (section scoping, fence
   skipping) and my fixture put the quoted row in `## Problem` — which the
   scoping alone rejects. Verified: deleting `inside[i] ||` left
   `TestMilestoneTick` PASS. The fixture now carries a quoted row *inside*
   `## Plan` (only `FenceSpans` can reject it) and one in `## Log` (only
   `SectionByteBounds` can); each revert now reds independently.
2. **Where the wiring is not drivable, the guard needs a writer sibling.**
   Reverting `close.go`'s tick to the old whole-body `ReplaceAll` left the suite
   green, because the behavioural test drives `TickMilestone` directly.
   `TestPlanItemWritersUseTickMilestone` closes it, sharing one `assertWiring`
   with the reader and commit-pathspec guards rather than a third AST walk.
3. **A contract looser than every caller exercises needs the caller it exists
   for.** `FindLineOutsideFences` returns a line range, but its sole caller
   anchors `^...$`, so reverting to the match range changed nothing observable.
   The test now passes an unanchored pattern matching mid-line — the case the
   contract is about, since the caller splices at the returned offset.

Also fixed, same rule: this issue's own reason-to-exist regressions
re-implemented close's guard and the milestone scan over the *unfiltered*
section, so they never traversed the production path. The milestone one now
drives `findMilestonesMissingVerdict` (which is in-process drivable); the
plan-unchecked one reads `PlanItemsBody`, the extractor the routing guard pins.

### 2026-09-02 — M2 close (FIX-THEN-SHIP, three advisories fixed pre-commit)

The first advisory was the same shape a **third** time, and this round the fix
was structural rather than another test:

- **The tick test reproduced `close.go`'s loop line for line**, so it asserted a
  copy of the logic instead of the logic — reverting the real code would have
  left it green. The cause was placement: the loop lived inside the IO shell,
  where a test could only restate it (`ARCH-PURE`). Extracted as
  `issue.TickMilestone` (pure); the test now drives it. **The "reverting either
  filter reds it" claim first written here was false** — see the close-review
  entry below.
- **`planItemReaders` was a hand-maintained list** — the #208 restatement
  problem. Now DERIVED: any function referencing a counting regex
  (`PlanUncheckedRE`, `PlanItemRE`, `nonEmptyPlanItemRE`, `milestonePlanRE`,
  `milestoneLabelRE`) *is* a plan-item reader and must call `PlanItemsBody`. The
  exemption map is empty and the guard refuses a stale entry — proven by the one
  I tried to add for `TickMilestone`, which it rejected because that function
  builds its pattern inline and is never classed as a reader.
- **The tick warning guessed a cause** it no longer had: `n == 0` now also means
  "no `## Plan` section". `HasSection` distinguishes them.

Revert-verified: routing `checkPlan` back to `PlanSectionBody` →
`structural.go:checkPlan uses nonEmptyPlanItemRE but never calls PlanItemsBody`.

### 2026-09-02 — M2 review round 5: a false evidence claim of mine
- 2026-09-02: closed M2 — M2 complete; every finding from rounds 1-5 closed. `go test ./cmd/... ./pkg/...` -> green except the pre-existing unrelated fleet_plan test (#210). Evidence stated as command -> observed result, per BR-16 rule: (1) BR-11 — reverted setstatus.go StripFenced and close.go dayRE to fence-blind scans, `go test ./cmd/sdlc/ -run TestWithinSectionMatchers` -> --- FAIL on both assertions. (2) BR-10/BR-16 — the ROUTING, which my earlier claim did not verify: reverting all four call sites (close.go x2, sizing.go, structural.go) to PlanSectionBody, `go test ./cmd/sdlc/...` -> green except fleet_plan, i.e. unpinned. Fixed at the source since computeClose dies past the test seam: routing close.go back with TestPlanItemReadersUsePlanItemsBody -> 1 failure naming file, function and reason. (3) BR-12 — TestMilestoneTick_OnlyTicksTheRealPlan covers a quoted row, a real row, and a row outside ## Plan. (4) BR-3/BR-6 — reverting the bounded section and the fence-aware estimate scan -> --- FAIL on their named tests, each after an explicit build check because two earlier probes silently failed to compile and read as green. BR-17 doc drift swept across all four members (FindLineOutsideFences now returns the documented LINE range, which is the safe contract for a splice API; logHeaderRE and stripEstimateForHash were dead symbol names; section.go carried a rationale close.go had proven false). BR-5 dead parameter removed; BR-13 stale plan/Done-when claims corrected. SplitFences deliberately not rebased — TestSplitFences pins the character-oriented contract making it a different abstraction; recorded at the function, in the atlas, and in the Log. Two lessons.md rules added. Round 4 verdict unknown was an OAuth expiry in the reviewer subprocess, not code state.; review verdict: FIX-THEN-SHIP

Round 4 returned `verdict: unknown` — the reviewer subprocess failed OAuth
("session expired and could not be refreshed"), not a gate or code fault. The
operator re-authenticated; round 5 ran properly and disposed 11 findings.

**BR-16 caught me reporting evidence that was false.** The M2 Log said BR-10 and
BR-11 were "both revert-verified". BR-11 was. BR-10 was not, and the reason is
the exact trap `workshop/lessons.md` records from earlier today:

- What I reverted: the **helper** — I removed the `StripFenced` call inside
  `PlanItemsBody` and watched a test go red.
- What BR-10 was about: the **routing** — four call sites choosing
  `PlanItemsBody` over `PlanSectionBody`.

Measured myself before accepting the finding: reverting all four call sites
(`close.go` ×2, `sizing.go`, `structural.go`) and running `go test ./cmd/sdlc/...`
leaves every test green except the pre-existing `fleet_plan` one.
`TestPlanItemReadersAgree` calls `PlanItemsBody` directly and re-implements
close's guard rather than driving it, so it pins the helper and mocks the wiring
— which is what I wrote a lesson about six hours ago and then did.

**The rule, and the fix:** where the entry point is not in-process drivable, a
source-level guard asserts the wiring. `TestPlanItemReadersUsePlanItemsBody`
does that, in the same shape as #206's `TestVerbsWireTheirCommitHelpers`.
Revert-verified: routing `close.go`'s guard back to `PlanSectionBody` and running
`go test ./cmd/sdlc/ -run TestPlanItemReadersUsePlanItemsBody` → 1 failure naming
the file, the function, and why that reader must not read the raw section.

**BR-17 — doc drift, all four members swept**, not just the one named:
`FindLineOutsideFences` documented a line range and returned a match range (now
returns the line range, which is the safe contract for a splice API — the sole
caller anchors its pattern so the two coincide today, but an unanchored one
would splice mid-line); `close.go` named a deleted `logHeaderRE`; the atlas and
this issue named `stripEstimateForHash`, which does not exist (`planGateContent`);
and `section.go` carried the tick-writer rationale `close.go` had already proven
false.

The rule BR-16 states is worth keeping: **a Log line asserting evidence must name
the command run and the observed result**, and "revert-verified" is only true
when the named test went RED with the production change undone. A build check is
not a revert check. The BR-11 claim above is rewritten to that standard.

### 2026-09-02 — M2 review round 3, and a process error of mine

**I broke the FIX-THEN-SHIP protocol.** Round 2 returned FIX-THEN-SHIP, whose
rule is: fix, bundle into ONE commit, and do NOT re-run `milestone-close` —
"a second review of the same boundary is the #172 re-close loop". I fixed, made a
*new* commit, and re-ran. The boundary anchor advanced to that commit, so round
3's review window was `7752d83f..7752d83f` — **empty**. Three findings (BR-4,
BR-5, BR-7) were re-raised as not-addressed because the reviewer could not see
fixes that were already in the tree.

The findings it derived from reading the tree anyway are real:

- **BR-12** — the milestone TICK ran `ReplaceAll` over the whole body, so a
  `- [ ] Mx` inside a quoted example anywhere in the issue was ticked. Worse, my
  BR-4 note justified skipping the writer with "it rewrites a line it already
  matched" — which is false; it rewrites every matching row in the document.
  Fixed: scoped to the real Plan section and fence-filtered, with a test.
- **BR-13** — a ticked plan row still claimed `SplitFences` was rebuilt, and the
  Done-when pointed at a `## Revisions` section that does not exist on this
  issue (the content is in `## Log`). Both corrected.

**Both then closed rather than deferred:**

- **BR-10** — the routing onto `PlanItemsBody` was pinned by nothing. The
  invariant is AGREEMENT, not correctness of one site: BR-4 was "`sdlc state` and
  `sdlc close` report different things about the same Plan". So
  `TestPlanItemReadersAgree` asserts every reader — `CountPlanItems`, close's
  unchecked guard, the milestone scan, the structural gate — sees the same item
  set, and fails if any one is routed back to the unfiltered body.
- **BR-11** — the finding's own sentence is the rule: *bounding a search to a
  fence-aware section does not make the search fence-aware*, because a section
  can quote its own format. `FindLineOutsideFences` is the offset-returning
  answer for splicing callers (`insertLogLine`'s day-header lookup) and
  `StripFenced` the reading one (`logHasEntryToday`).

  Revert-verified, naming the command and the result:
  reverting `setstatus.go`'s `StripFenced` and `close.go`'s `dayRE` to a
  fence-blind scan, then `go test ./cmd/sdlc/ -run TestWithinSectionMatchers`
  → `--- FAIL` on both assertions. (An earlier draft of this line said "both
  revert-verified" covering BR-10 too; that was false. See below.)

### 2026-09-02 — M2 review (5 findings, all fixed pre-commit)

- **BR-3 (Critical)** — the `insertLogLine` fix was half a fix. Anchoring the
  `## Log` *heading* fence-aware still left the `### <date>` search running to
  EOF, so a quoted day header in a LATER section captured the insert: the same
  class one level down. Bounded to the section via `SectionByteBounds`.
  My first regression test for it passed against the broken code — the quoted
  header sat after the real one, so first-match found the right thing anyway.
  Rewritten so the real section has no matching day header, which is the only
  shape where the unbounded scan actually bites; now red on revert.
- **BR-4** — the fenced-item filter reached `CountPlanItems` and nothing else, so
  `sdlc state` and `sdlc close` disagreed about the same Plan. Filtering
  per-consumer was the error; there is now one `PlanItemsBody` extraction point
  that every counting reader uses. The milestone-tick WRITER keeps the unfiltered
  body, deliberately — it rewrites a line it already matched.
- **BR-5** — the indent parameter had zero callers after the `SplitFences`
  revert, plus a comment naming a caller that never invoked it. Removed.
- **BR-6/BR-7** — coverage and stale docs: `insertLogLine`'s doc block still
  specified the last-match anchor it no longer uses, and `SplitFences`' comment
  described the parameter above. Both rewritten to what the code does.

### 2026-09-02 — M2

Swept the heading-finding class the M1 review enumerated. `logHasEntryToday` and
`insertLogLine` now locate the real `## Log` through the scanner, and
`planGateContent`'s Estimate strip's line scan is fence-aware. Revert-verified on the live
instance: restoring the old first-match lookup reds both assertions —
today's real entry is not found, and a date existing only inside the quoted
example matches.

`insertLogLine`'s **last-match heuristic from #66 is retired.** It existed
because first-match filed a close line into #66's own quoted example — the same
defect, worked around 145 issues earlier — and it was only accidentally right:
it fails when a quoted `## Log` sits after the real one, which the new test
covers directly.

**`SplitFences` is NOT rebased, and that is the decision, not a shortfall.** The
attempt was made and its contract test refused it: `TestSplitFences` pins
`` a```one``` mid ```two```z `` as two fenced segments with prose between them —
inline pairs, mid-line boundaries. That is CHARACTER-oriented segmentation, and a
line classifier cannot express it without embedding a second scanner inside the
first. It also answers a different question ("may a rewriter edit these bytes"
vs "where does this section end"), so merging would change what `migrate`
rewrites across repos for no benefit to this class. Reverted, with the reasoning
recorded at the function and in the atlas, and the Done-when narrowed from "one
scanner" to "one line-oriented scanner plus a recorded exception".

An indent-policy parameter had been added to make that rebase
behavior-preserving (`SplitFences` is indent-blind where CommonMark allows ≤3
spaces). With the rebase declined it had zero callers and a doc naming one that
never invoked it, so it is **removed** — the M2 review was right that a parameter
kept "to make the axis explicit" is just dead code with a false comment. The axis
is documented in prose where the two implementations are compared.

The plan-item counters now filter through `StripFenced`, so a `- [ ]` in a quoted
example is not counted as open work.

### 2026-09-02 — M1 review (FIX-THEN-SHIP, no blocking findings)

Fixed before the boundary commit:

- **I5 (`ARCH-DRY`)** — the diff removed one duplicated section scanner and
  introduced another: `SectionBody`'s loop and `project.SectionLineBounds` were
  the same algorithm differing only in policy. Now one exported
  `issue.SectionLineBounds(lines, heading, policy)`; `SectionBody` joins its
  range and `project` wraps it with its own policy. Doing the exact thing this
  issue argues against, one package over, is worth naming.
- **I6** — `fence.go`'s header and the atlas policy table described M2's end
  state as landed: `stripCodeFences` and `SplitFences` are still on their own
  logic at this boundary. Both now say "M2 rebases it onto this", so a reader who
  greps finds what is actually there.
- Minors: `close_test.go`'s two inline copies of the deleted `PlanSectionRE`
  routed through `PlanSectionBody` (they were pinning a shadow of removed code
  and giving false coverage for the path this issue fixes); the stale "matching
  regex edit" comment; `workshopMarkdown` now skips on an absent corpus instead
  of swallowing the error and then fataling below 100 files.

**Carried into M2 rather than fixed here (I4):** the class is "code that locates
a `## <heading>` without fence awareness", and it has three more production
members — one with a *live* instance on a real issue file, where
`logHasEntryToday` reads a fenced `## Log` at line 22 instead of the real one at
line 68. Enumerated in the Plan this round, which is what the review asked for;
the sweep is M2-sized.

Also noted: `close.go`'s `insertLogLine` carries a last-match heuristic added by
#66 that is a second, weaker workaround for precisely the problem `FenceSpans`
solves. That is the shape of a class fix arriving late — the workaround predates
the scanner by 145 issues.

### 2026-09-02 — M1
- 2026-09-02: closed M1 — M1 complete. go test ./cmd/... ./pkg/... green except the pre-existing unrelated fleet_plan test (#210). Revert-verified against the old regex parser: all three close-gate regressions go red with real numbers (plan-unchecked 0 of 2 open items; milestone scan [M1] not [M1 M2]; CountPlanItems 2 of 4) plus 7 cases in the fence suite. Corpus property test verified 2875 real sections across 406 workshop markdown files. `sdlc issue validate --all` byte-identical before and after (16 conforming both ways) — the predicted zero-verdict-change held. project/ suite passes unchanged after consuming the moved scanner at UnterminatedIsFenced. Atlas: issue-lifecycle.md gains the section-parsing entry with the RE2 argument and the per-consumer policy table (row moved from M2 — the boundary gate correctly identified M1 as where the surface landed). Actual 0.65h is the measured window value; the engine flags it cross-attributed across #115/#208/#211 with #211 at 27.7m/71%, so it is a conservative upper bound rather than a clean per-issue figure.; review verdict: FIX-THEN-SHIP

Scanner moved down into `internal/issue` as `FenceSpans` / `ScanMarkdownLines`
with the unterminated policy as a parameter; `project` consumes it at
`UnterminatedIsFenced` and its suite passes with no other edit. `SectionBody`
rebuilt on it at `UnterminatedIsProse`; `PlanSectionRE` deleted and all six sites
plus both stale comments rerouted through `PlanSectionBody`.

Revert-verified, which is the only evidence that counts here: restoring the old
regex reds all three close-gate regressions with the real numbers — plan-unchecked
sees 0 of 2 open items, the milestone scan sees `[M1]` instead of `[M1 M2]`, and
`CountPlanItems` reports 2 of 4 — plus 7 cases in the fence suite.

Corpus property test verified **2875 real sections across 406 workshop markdown
files**, and `sdlc issue validate --all` is byte-identical before and after
(16 conforming both ways). The prediction that no verdict would move held.

The atlas row was planned under M2 and the M1 boundary gate refused without it —
rightly, since the surface M2 adds is consumers of a scanner M1 introduced. Moved
to M1 and written: `atlas/workflow/issue-lifecycle.md` gains the section-parsing
entry, the RE2 argument, and the per-consumer policy table.

One test expectation was wrong rather than the code: under `UnterminatedIsProse`
a `##` following a stray opener IS read as a real heading, so text after it lands
in that accidental section. That over-segmentation is the deliberate price —
visible and recoverable, where under-segmenting hides `## Plan` from the gates —
and it now has its own test saying so, so nobody "fixes" it later.

### 2026-09-02

Found while running `sdlc change-code --issue 208`, an issue whose Spec quotes
the two registry entries it adds — the parser stopped at the first quoted
`## ARCH-SECURE` and reported 35 words against a 600-word Spec. #208 was
unblocked by adding a summary paragraph *before* the fence, which the Spec wanted
anyway; the parser bug is untouched and lives here.

Plan-quality round 1 caught the thing the investigation missed, and it was the
same class as the bug: **the unterminated-fence policy.** My plan moved
`scanMarkdownLines` down wholesale, inheriting "unterminated ⇒ runs to EOF" —
which for section extraction deletes every heading after the fence, `## Plan`
included. That is the false pass this issue exists to fix, made unbounded.

The demonstration was this file. My own rewrite of the Problem section opened a
four-backtick fence in prose and never closed it, so under the proposed scanner
`## Done when`, `## Plan` and `## Log` all vanished from #211 itself. Fixed the
prose, and made the policy an explicit per-consumer parameter.

While adding the `## Estimate` block, a scripted insert at `index("## Plan\n")`
matched the *example* `## Plan` inside the Problem section — the indented one this
issue uses to demonstrate the bug — splitting that line and landing the whole
estimate inside Problem. `plan-present` then failed with "no non-empty checklist
items", because the real Plan's items had been left behind an unindented heading.

Third time today this issue's own text tripped the thing it describes: a
heading-shaped string in prose being read as structure. Worth noting as evidence
for the Spec's claim that "don't write markdown in issues" is not a workable
policy — the author here knew the bug intimately and still hit it twice while
writing the fix.

Investigated before planning rather than after. Three things the first draft of
this issue got wrong or missed: it led with the false refusal when the false PASS
on `close` is the real defect; it did not mention `PlanSectionRE`, which carries
the same bug and feeds both affected gates; and its plan said "track fence state
while scanning", which would have made a **fourth** scanner in a tree that
already has three that disagree.
