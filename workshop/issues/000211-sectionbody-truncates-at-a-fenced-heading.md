---
id: 000211
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
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
```` ```` ````-fence containing a ``` line wrong. Measured 4 of 5.

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

**`PlanSectionRE` is deleted, not fixed.** All four consumers — `close.go:563`,
`sizing.go:63`, `findMilestonesMissingVerdict`, `checkPlan` — take
`FindStringSubmatchIndex` only to slice `body[m[2]:m[3]]` and work on the string.
None needs byte offsets, so each becomes `SectionBody(body, "Plan")`.

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

## Plan

- [ ] Move `scanMarkdownLines` + `fenceMarker` from `internal/project/doc.go`
      into `internal/issue`; `project` consumes it from there.
- [ ] Rebuild `SectionBody` on the scanner; delete its regex terminator.
- [ ] Delete `PlanSectionRE`; route its four consumers through
      `SectionBody(body, "Plan")`.
- [ ] Rebuild `stripCodeFences` + `SplitFences` on the scanner, keeping their
      unterminated-fence difference as an explicit parameter.
- [ ] Table-test the fence forms; regression-test the `close` false pass.
- [ ] Corpus verdict diff before/after; record the changed set (predicted empty).

## Log

### 2026-09-02

Found while running `sdlc change-code --issue 208`, an issue whose Spec quotes
the two registry entries it adds — the parser stopped at the first quoted
`## ARCH-SECURE` and reported 35 words against a 600-word Spec. #208 was
unblocked by adding a summary paragraph *before* the fence, which the Spec wanted
anyway; the parser bug is untouched and lives here.

Investigated before planning rather than after. Three things the first draft of
this issue got wrong or missed: it led with the false refusal when the false PASS
on `close` is the real defect; it did not mention `PlanSectionRE`, which carries
the same bug and feeds both affected gates; and its plan said "track fence state
while scanning", which would have made a **fourth** scanner in a tree that
already has three that disagree.
