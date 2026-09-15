---
id: 000228
status: open
deps: []
github_issue:
created: 2026-09-15
updated: 2026-09-15
estimate_hours:
---

# Align local superpowers with SDLC review gates

## Problem

The local adaptation of the Superpowers brainstorming and writing-plans skills
requires review loops that overlap with Ariadne's SDLC gates. In particular,
brainstorming requires an automated spec-document review, and writing-plans
requires a plan-document review. `sdlc change-code` then performs the mandatory
fresh-context plan-quality review before implementation, with persistent findings
and disposition tracking.

This creates duplicate review authority and makes the intended workflow unclear.
The local skills should preserve useful design discipline without requiring agents
to pass multiple substantially overlapping automated approvals.

## Spec

Adapt the local Superpowers guidance so the workflow boundary is explicit:

> Brainstorming establishes requirements and gets the operator's agreement;
> separate automated spec review is optional when useful.

The responsibilities should be divided as follows:

- Brainstorming explores context, clarifies intent, presents alternatives, and
  records the agreed requirements in the issue's Spec. Human agreement is the
  design approval; the skill may recommend an advisory spec review for unusually
  broad, ambiguous, or high-risk work, but it must not create a second mandatory
  gate by default.
- Writing-plans turns the approved requirements into the durable implementation
  plan. Its reviewer may be retained as an advisory editing aid, but it must not
  be represented as an approval boundary independent of SDLC.
- `sdlc change-code` owns the mandatory implementation-entry review: deterministic
  structural checks plus the stateful plan-quality judge. Its findings, acceptance
  cache, estimate ordering, and refusal behavior remain authoritative.
- Boundary reviews at milestone/close remain the separate implementation-result
  review. This issue does not change those gates.

The adaptation must explain how advisory findings are carried into the durable
issue/plan and then surfaced to `change-code`, without asking the agent to repeat
the same review merely to satisfy two skill checklists. Keep the distinction
between human approval of intent and automated verification that an implementation
plan is executable. Cite ARCH-DRY: one mandatory owner per review boundary and one
source of truth for its findings.

## Done when

- The adapted brainstorming skill says that human agreement establishes the
  requirements and that automated spec review is optional, with criteria for when
  it is useful.
- The adapted writing-plans skill no longer presents its plan reviewer as a second
  mandatory approval gate when `sdlc change-code` will run the plan-quality review.
- The skills and SDLC documentation name the handoff and authority clearly:
  brainstorming/spec → durable plan → `change-code` plan-quality gate → code →
  milestone/close review.
- A regression or documentation test demonstrates that the generated workflow
  instructions do not require duplicate automated approval for one plan.
- Existing required SDLC gates remain intact and their persistent findings are not
  bypassed or silently duplicated.

## Plan

- [ ] Map the adapted skill text, source intent, generated consumers, and current
  `change-code` plan-quality contract (ARCH-PURPOSE shadow sweep).
- [ ] Revise brainstorming guidance to make operator agreement the design gate and
  automated spec review an optional advisory path.
- [ ] Revise writing-plans guidance to defer mandatory plan-quality authority to
  `sdlc change-code`, while preserving useful plan editing/review advice.
- [ ] Update tests/golden prompts/derived skill outputs that encode the old review
  sequence; assert the handoff and absence of duplicate mandatory approval.
- [ ] Run the relevant architecture, skill, and SDLC CLI tests and update atlas or
  workflow vocabulary if the ownership boundary is a new documented surface.

## Log

### 2026-09-15

Task filed from the review of local Superpowers versus SDLC. The desired policy is
that brainstorming establishes requirements with operator agreement; separate
automated spec review is optional. `sdlc change-code` remains the mandatory owner
of plan-quality review before implementation. No implementation changes made.
