---
id: 000454
status: open
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'f05201aa2196d6f3d03161fc542f4a75d2643c75' # card fields mirrored from issue-cards; edit via sdlc
---

# ARCH principle: persisted formats evolve additively; readers ignore unknown fields; refusal only when ignoring is unsafe

## Problem

The ARCH registry has no principle for how persisted formats evolve. ARCH-SECURE covers reading input written by an *older* version, but nothing says what an older reader should do with data from a *newer* writer. Designs default to refusing unknown records (pair#684's first draft, precedent #198). The operator (2026-10-18): "typical robust strategy would be compatibility, that old binary will just ignore new attributes ... both simpler and also more robust ... if this is not your default, I wonder if we should add [it] to architecture".

## Spec

Add a principle to the registry (name to settle, e.g. ARCH-EVOLVE): persisted formats (records, ledgers, manifests, config) evolve additively. A reader ignores fields it doesn't know and keeps working; a writer never changes the meaning of an existing field. A version bump, with older readers refusing, is reserved for the cases where ignoring a new field would be unsafe (it changes how existing fields must be interpreted, or it guards a destructive action an old reader would otherwise take), and the design must say why. at-plan prompt: for each persisted format the design changes, say whether the change is additive, and if a reader must refuse, why ignoring is unsafe. Delivered through sdlc arch-principles, start-plan and the plan-quality judge like the others.

## Done when

- The principle is in the registry, printed by sdlc arch-principles and start-plan.
- The plan-quality judge cites it on a design that adds a field with a refusal but no stated reason (a test fixture).

## Plan

- [ ]

## Log

### 2026-10-18
