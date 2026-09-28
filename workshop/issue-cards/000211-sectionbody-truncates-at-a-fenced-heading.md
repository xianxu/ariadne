---
id: '000211'
status: done
started: 2026-09-02T18:20:40-07:00
created: 2026-09-02
updated: 2026-09-03
estimate_hours: 1.71
actual_hours: 4.33
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
