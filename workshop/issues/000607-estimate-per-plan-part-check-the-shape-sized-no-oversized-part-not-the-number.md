---
id: 000607
status: open
deps: []
github_issue:
created: 2026-10-24
updated: 2026-10-24
estimate_hours:
card_mirror: 'fb5189d49c4f4d2b6fdcbf6e6f5295086a7645d7' # card fields mirrored from issue-cards; edit via sdlc
---

# Estimate per plan part: check the shape (sized, no oversized part), not the number

## Problem

## Spec

**Estimate per plan part, as a nudge to break work down (operator, 10-24).** The operator's framing: the estimation
system should be "a soft nudge for AI to break things down and reason about how complex the parts are" and feed future
estimates; it's auxiliary and shouldn't block landing. Today change-code asks for one `estimate_hours` number plus an
`## Estimate` block after plan-quality passes; the single number carries little of the reasoning.

Change:
1. **Estimate per part.** The `## Estimate` block becomes a short table: each plan step or milestone, its size and its
   main risk; the total is derived.
2. **Check the shape, not the number.** The gate checks that every part is sized and that no single part exceeds a
   few hours (a part too large to size is a signal to split the issue, raised as a plan-review concern); it never
   argues with the numbers themselves.
3. **Soft otherwise.** A missing or odd estimate produces a note in the plan review, not a refusal, except the
   oversized-part signal.

Side doc with the full proposal: ops `session.side/2026-10-24-11-estimation-system-proposal.md` (operator approved).

## Done when

- change-code accepts a per-part estimate table and derives the total; a test pins the parsing and derivation.
- An oversized or unsized part is flagged in plan review; other estimate problems are notes, not refusals.
- Help and the atlas describe the table.

## Plan

- [ ]

## Log

### 2026-10-24
