---
id: 000103
status: working
deps: []
created: 2026-06-15
updated: 2026-06-17
estimate_hours: 1
---

# AGENTS.md instruction: user intention

we should instruct agent to establish their hypothesis of user intention and keep validating that mental model along a chat/coding session. Each turn will either positively and negatively move that user mental model, and agent should keep updating the model such that the model of user mental model is 1/ self consistent; 2/ fit observation from user interactions.

This issue owns the **live-maintenance** discipline (the canonical definition of
the user-model + its two criteria). Its **persistence** counterpart already
exists: `ariadne#105` added a `## Thread arc & user model` section to the
`continuation` datatype that checkpoints this model at park/handoff time, using
the same two criteria. When this issue lands its AGENTS.md instruction, keep the
criteria single-sourced — the continuation section defers here as canonical.

## Spec

Add a constitution instruction (in ariadne's `AGENTS.base.md`, the composed
source) establishing the **live user-model discipline**: across a chat/coding
session the agent holds a running hypothesis of the user's intent/mental-model,
and updates it every turn — each turn moves the model positively or negatively —
keeping it **(1) self-consistent** and **(2) fitting the observed user
interactions**. This is the canonical home of the user-model + its two criteria
(design settled in the Log resolutions above: single-source here; no dedicated
durable artifact).

Single-source: thin ariadne#105's `continuation` datatype section (`## Thread arc
& user model` in `construct/datatype/continuation.md`) so it **defers to AGENTS.md
by pointer** — remove the duplicated criteria text, keep the per-park checkpoint
purpose. (The persistence counterparts — the continuation checkpoint + the
per-session pensive flush — are unchanged in role; only the *criteria definition*
collapses to here.)

Voice/placement: match the existing constitution prose; site it where session-
conduct guidance lives (near the workflow/answering-questions principles), terse —
one short instruction, not an essay.

## Done when

- `AGENTS.base.md` carries the user-model live-maintenance instruction (the running
  hypothesis + the two criteria), in the constitution's voice and a sensible slot.
- `construct/datatype/continuation.md`'s `## Thread arc & user model` defers to the
  AGENTS.md instruction for the two criteria (duplicated criteria text removed),
  per the #105 single-source resolution; its checkpoint purpose stays.
- `weave compile` regenerates the composed `AGENTS.md`/`CLAUDE.md` cleanly (the
  instruction appears in the woven output); tree clean; `make harness-check` green.

## Plan

- [ ] Add the user-model live-maintenance instruction to `AGENTS.base.md`
      (canonical: running hypothesis + the two criteria).
- [ ] Thin `construct/datatype/continuation.md` `## Thread arc & user model` to a
      pointer to the AGENTS.md instruction (drop duplicated criteria text).
- [ ] `make weave` (regenerate) + verify clean tree + `make harness-check`.

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

