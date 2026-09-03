---
id: 000211
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.55
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
    - [ ] M2 — NOT done
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
- Exactly ONE fence scanner remains in `cmd/sdlc`. `PlanSectionRE` is gone;
  `stripCodeFences` and `SplitFences` are built on the shared scanner and still
  disagree about an unterminated fence, asserted by a test that names why.
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
item: smaller-go-module      design=0.05 impl=0.20
item: cross-cutting-refactor design=0.05 impl=0.20
item: smaller-go-module      design=0.05 impl=0.20
item: smaller-go-module      design=0.05 impl=0.20
item: atlas-docs             design=0.05 impl=0.06
item: milestone-review       design=0.00 impl=0.20
item: milestone-review       design=0.00 impl=0.20
total: 1.55
```

In Plan order:

1. `smaller-go-module` — move `scanMarkdownLines` + `fenceMarker` down into
   `internal/issue` and give the scanner its explicit unterminated-fence policy.
   Mirror-or-extend of working code, but the policy parameter is new.
2. `cross-cutting-refactor` — rebuild `SectionBody` on it, delete
   `PlanSectionRE`, route its six sites and two stale comments.
3. `smaller-go-module` — rebuild `stripCodeFences` and `SplitFences`. **The
   riskiest row**: byte-exact reassembly has to survive, the two consumers must
   keep disagreeing about unterminated fences on purpose, and `SplitFences`'
   line-anchoring change alters what `migrate` will rewrite.
4. `smaller-go-module` — the tests: a corpus-seeded property test over all 406
   `workshop/**/*.md`, per-function tables, and the `close` false-pass
   regression. Priced as its own row rather than folded into the others because
   the property test is new machinery, not a table.
5. `atlas-docs` — the section-parsing entry and the scanner's single-source note.
6. `milestone-review` ×2 — see below.

Design is `×0.2` spec-quality discounted: the Spec names every consumer by
file:line from a grep, tables the six divergence axes, and fixes each consumer's
policy — the remaining design is reading. Buffer `+15%` (v3.1 step 4).
Familiarity `1.0`: Go, this package, line scanning; nothing novel in the stack.

**Two `milestone-review` rows, both at the scaled ceiling**, following #208's
correction: the lever for "I expect N rounds" is N rows, not one row inflated.
Two is the read — plan-quality already took two rounds here, and the class has
more members than #208's (six call sites, five fence forms, three scanners being
collapsed, one behaviour change to a different verb). Today's ledger: #206 at
0.3× priced one round for six; #208 at 0.7× priced two and took two.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [ ] Move `scanMarkdownLines` + `fenceMarker` from `internal/project/doc.go`
      into `internal/issue`; `project` consumes it from there.
- [ ] Rebuild `SectionBody` on the scanner; delete its regex terminator.
- [ ] Delete `PlanSectionRE`; route its four consumers through
      `SectionBody(body, "Plan")`.
- [ ] Give the scanner an explicit unterminated-fence policy; `SectionBody` and
      the plan extraction take `prose`, `project` and `SplitFences` keep `fenced`.
- [ ] Rebuild `stripCodeFences` + `SplitFences` on the scanner, preserving
      byte-exact reassembly and deciding `SplitFences`' line-anchoring change
      explicitly (it is a `migrate` behaviour change either way).
- [ ] Tests, named per function with the adversarial strategy for each:
      `fenceMarker` — table over width/char/indent/info-string boundaries;
      `scanMarkdownLines` — **corpus-seeded property test** over all 406
      `workshop/**/*.md`, asserting no file loses a real section and that
      concatenating visited + skipped lines reproduces the input;
      `SectionBody` — the five fence forms plus "closed fence, then a real
      heading" plus an unterminated fence that must NOT hide later sections;
      `stripCodeFences` / `SplitFences` — one test each pinning the
      unterminated-fence disagreement and naming why they differ;
      `close` plan-unchecked + `findMilestonesMissingVerdict` — the false pass.
- [ ] Fold the corpus diff into that property test so the invariant is mechanical
      rather than a one-off run; record any verdict change (predicted none).

## Log

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
