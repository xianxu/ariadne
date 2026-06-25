---
type: target
slug: issue-lifecycle
status: active
created: 2026-06-24
updated: 2026-06-24
sources: [workshop/pensive/2026-06-24-01-pensive-cue-schema-layer-nouns-verbs.md]
---

# Target: The issue lifecycle is a single, formally-modeled invariant

What we defend: an `issue` has exactly one lifecycle — a finite set of statuses
partitioned into *open* / *active* / *terminal*, connected by *named, guarded*
transitions — and that lifecycle lives in **one** authoritative source,
`construct/vocabulary/issue.cue`, from which every consumer (the `sdlc`
deterministic shell, parley, the prose in `AGENTS.md`, and the LLM that reasons
about issues) *derives* its understanding rather than restating it. The *what* —
which statuses exist, which transitions are legal, which guards each demands —
is the cue file's job and is not duplicated here. This target defends the *why*
and the *shape*.

Why we defend it: the lifecycle is the oldest and most-copied noun in the system.
Before this, the same six statuses and their legal moves were re-encoded
independently in sdlc's Go (scattered string literals), in parley's Lua, and in
base prose — three sources, none authoritative, drifting by construction. The
commitment is that there is now *one* source and the duplication is deleted, not
paralleled: a change to the lifecycle (a new status, a new transition) is a change
to the cue model, and consumers either derive it for free (category-based reads)
or fail closed (conformance checks) — never silently disagree.

Two properties are load-bearing and must not erode:
- **Transitions are gated, not free.** A status change is a *move along a declared
  edge* — and as of #122 M4 this is literally enforced: `sdlc issue set-status`
  refuses an edge the model doesn't declare (with a `--force` escape). An edge may
  also carry named guards (e.g. closing to `done` demands measured actuals,
  verification, an atlas update); the guard *names* live in the model, their
  effectful *implementations* in sdlc. Adding/removing an edge or a guard is a
  lifecycle change, made in the open in `issue.cue`, not an inline `if` somewhere.
- **Consumers branch on categories, not raw values.** `terminal`/`active`/`open`
  are the stable interface; individual status names are details under them. This
  is what lets the lifecycle grow (a new `active` status) without touching every
  consumer.

## Why now

Formalizing the lifecycle is the first instance of the broader bet — that a
system's nouns and verbs should be declared once, formally, and compiled to every
consumer (ariadne#122). `issue` is the guinea pig precisely because it is the most
scattered noun, so consolidating it both proves the pattern and pays down the
oldest duplication in the repo. CUE is the chosen source language, and this file's
existence also tests whether a formal model can serve as a *human/LLM design
interface*, not merely a build input.

## What this is NOT

- Not a redefinition of *what* the statuses are — that's `construct/vocabulary/issue.cue`; this target must never restate the transition table (a copy would drift, recreating the very problem it exists to kill).
- Not a general state-machine framework — it defends *this* lifecycle, not a DSL for arbitrary ones. Other nouns may get their own vocabulary; they don't inherit this one.
- Not a claim that every guard is compiled — effectful guards (active-time, atlas-updated) stay as named code; only the *contract* (the edge and its guard names) is in the model.

## Open questions

- Does the `parked` status (in-progress, blocked on an *external* dependency, distinct from `blocked`-on-another-issue) earn a place in the lifecycle, or does `blocked` cover it? (Used as the #122 acceptance scenario regardless.)
- How far do other nouns adopt this shape before the pattern itself (not just `issue`) becomes a defended target of its own?
