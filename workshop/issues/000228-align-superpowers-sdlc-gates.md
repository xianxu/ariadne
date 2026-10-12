---
id: 000228
status: working
deps: []
github_issue:
created: 2026-09-15
updated: 2026-10-12
estimate_hours:
card_mirror: '8c112364e4aa6ed05f99a281d03e16485e316210' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-12T16:34:13-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
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

Quick flow: prose edits to two adapted skills plus a doc test; no code.

Shadow sweep (ARCH-PURPOSE): the overlapping review loops live only in
`construct/adapted/superpowers-brainstorming/` (SKILL.md checklist step 8, graph,
"Spec Review Loop", "User Review Gate") and
`construct/adapted/superpowers-writing-plans/` (SKILL.md "Plan Review Loop",
"Execution Handoff") and their two reviewer-prompt templates (`Status: Approved`).
`construct/versions/*` are frozen snapshots and `construct/sources/*` the upstream;
neither is a live consumer. No Go code, helptext or atlas page restates either loop.
`superpowers-subagent-driven-development`'s spec reviewer checks code against spec
during execution, a different boundary, and stays.

- [ ] Brainstorming: the operator's approval of the written spec is the design gate;
  the spec-document review becomes optional and advisory, with criteria (broad,
  ambiguous or high-risk work); its findings are folded into `## Spec` or noted in
  `## Log`, never re-run to tick a checklist.
- [ ] Writing-plans: the plan-document review becomes an optional editing aid;
  `sdlc change-code`'s plan-quality gate is named as the one mandatory plan review
  (ARCH-DRY: one owner per boundary, one findings ledger). The handoff runs
  `change-code` before execution.
- [ ] Mark both reviewer-prompt templates advisory.
- [ ] Record Conversation 12 in `construct/intents/superpowers.md`.
- [ ] Name the review chain in `atlas/workflow/issue-lifecycle.md`.
- [ ] Doc test `cmd/sdlc/skillreviewauthority_test.go`: neither skill carries a
  mandatory automated approval loop; both name `sdlc change-code` as the plan-review
  owner; the chain is stated.

## Log

### 2026-09-15

Task filed from the review of local Superpowers versus SDLC. The desired policy is
that brainstorming establishes requirements with operator agreement; separate
automated spec review is optional. `sdlc change-code` remains the mandatory owner
of plan-quality review before implementation. No implementation changes made.
