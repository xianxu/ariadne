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

