---
id: 000103
status: open
deps: []
created: 2026-06-15
updated: 2026-06-15
---

# AGENTS.md instruction: user intention

we should instruct agent to establish their hypothesis of user intention and keep validating that mental model along a chat/coding session. Each turn will either positively and negatively move that user mental model, and agent should keep updating the model such that the model of user mental model is 1/ self consistent; 2/ fit observation from user interactions.

This issue owns the **live-maintenance** discipline (the canonical definition of
the user-model + its two criteria). Its **persistence** counterpart already
exists: `ariadne#105` added a `## Thread arc & user model` section to the
`continuation` datatype that checkpoints this model at park/handoff time, using
the same two criteria. When this issue lands its AGENTS.md instruction, keep the
criteria single-sourced — the continuation section defers here as canonical.

## Done when

-

## Spec


## Plan

- [ ]

## Log

### 2026-06-15
- Added an #105 back-pointer: this issue is canonical for the live user-model
  discipline; ariadne#105's `## Thread arc & user model` continuation section is
  its persistence counterpart (same two criteria, single-sourced here).
- **Open-question resolutions** (operator sign-off during the pair#61 continuation
  dogfood — these settle the design before this issue is worked):
  - **Single source for the two criteria → collapse to #103.** This issue holds
    the canonical definition of the user-model + its two criteria (self-consistent;
    fits observation). The continuation datatype section must defer here by
    *pointer*, not restate the criteria. When landing the AGENTS.md instruction,
    also thin ariadne#105's continuation section to a reference (remove the
    duplicated criteria text).
  - **User-model stays cross-cutting — no dedicated durable artifact.** The model
    persists as: the `## Thread arc & user model` continuation section (per-park
    checkpoint) + this AGENTS.md instruction (live maintenance) + the per-session
    `pensive` flush (durable cross-session view). Rejected: a standalone
    user-model `target`/profile datatype that the continuation merely updates —
    too much new surface for the value; the pensive already carries the durable
    cross-session throughline.

