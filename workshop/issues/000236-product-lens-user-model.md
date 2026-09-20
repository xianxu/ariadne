---
id: 000236
status: open
deps: []
github_issue:
created: 2026-09-18
updated: 2026-09-18
estimate_hours:
---

# Product lens: PROD-* principles and a per-product user model

## Problem

Agents design from the builder's seat. The brainstorming skill's design checklist
("Cover: architecture, components, data flow, error handling, testing" —
`construct/adapted/superpowers-brainstorming/SKILL.md:111`) never asks what the
user has to hold in their head, so the implementation's state machine is what
users end up seeing. ARCH-* has no user-side counterpart. Specimen: parley#266,
where the operator had to make or correct five product calls: undo grouping,
holding the write turn, failed tool → error result, Stop flushing tool pairs,
and removing the crashed-tool lock.

## Spec

**Not ready. The design is being worked out in a pensive:**
`brain/workshop/pensive/2026-09-18-01-pensive-product-lens-user-model.md`.
Evidence: `brain/probes/prod-lens-replay/` (r1). Don't `claim` or `start-plan`
until the pensive converges.

Current hypothesis (after r1–r4 and spec-stage s1), two parts, no new gates:

1. **A `prd` skill, run before claim:** a simplified PRD (the user journey × how
   user and system errors interrupt it). Each cell is either settled or turned
   into a policy question to the operator in user terms. It writes `## PRD` on
   the issue, and `start-plan` maps it to engineering.
2. **An always-on product section** in AGENTS.md / AGENTS.local.md: the
   principles plus the operator's written policies. Policies grant the agent
   authority on trade-offs, such as tool safety handled in-band with the user
   not acting as a safety gate, or undo in units of one answer.

## Done when

- The pensive's open questions are settled enough to write a real Spec: replay
  arm C (does a recorded fact carry over?) has been run, plus held-out cases
  from outside #266.
- This Spec and Done-when are rewritten from the converged design.

## Plan

- [ ] `prd` binary + `start-plan` gate → ariadne#237.
- [ ] Converge the design in the pensive (arm C replay, held-out cases).
- [ ] Rewrite Spec / Done when; then claim.

## Log

### 2026-09-18

Filed from a brain advisor session as a pointer to the pensive. r1 results:
c1/c3/c4 pass with no lens (the failure at the time was session context: asking
habits, defending what it had just built); c2/c5 fail in both arms (the model
pictured the wrong user; missing fact: "operator will not pay attention to tool
call").

Later the same day:

- **r2–r4:** always-on context works at least as well as a prompt placed at the
  decision. c2/c5 are safety-versus-simplicity trade-offs the model won't make
  on its own. A written policy that grants authority passes 15/15 (leaked).
- **Product spec before claim:** the #266 transcript shows the operator's
  pre-claim framing went straight to an engineering invariant. Replaying that
  framing through a journey × interruptions method surfaces all five decisions
  before claim: 0/30 missed, against 8/15 for #266 as filed.
- **Naming:** the pre-claim skill is **`prd`** and writes a `## PRD` section.
  Its first sentence: *"This is not your typical PRD, but a simplified version
  focused on the user journey and how user and system errors interrupt it."*
