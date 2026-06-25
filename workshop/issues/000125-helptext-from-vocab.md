---
id: 000125
status: working
deps: [ariadne#122]
github_issue:
created: 2026-06-25
updated: 2026-06-25
estimate_hours:
started: 2026-06-25T14:33:20-07:00
---

# sdlc embeds a help-text fragment GENERATED from the issue vocabulary (stop hand-maintaining lifecycle prose)

## Problem

`sdlc`'s embedded help text (`cmd/sdlc/helptext/*.md`) hand-restates lifecycle facts and
**drifts**. #122 M4 made `set-status.md`'s "All other transitions are allowed without
guards" *false* (the lifecycle gate now refuses non-modeled flips); nothing caught it
automatically — the #122 whole-issue fresh-eyes review did (FIX-THEN-SHIP), and it was
patched by hand. So the operator-facing prose is a **hand-maintained shadow** of the
model, not derived — the exact drift the vocabulary layer (#122) exists to kill, for the
one consumer we never wired. We already generate-from-source where the target is
*templated* — code → `issue.json` via `go:embed`; the vocabulary skill `SKILL.md` via the
`.dynamic-skill` renderer — the gap is the *free-form help prose*.

## Spec

Generate the model-derived portion of the relevant help text from the vocabulary, so it
can't drift; keep the hand-written framing prose. **Generate the facts, reference the
source** (don't re-enumerate edges by hand). Likely shape:

- Extend the vocabulary renderer (the one that already emits the skill breadcrumb) to
  emit a **help fragment** for the issue lifecycle — the legal transitions / the gate
  description — as markdown.
- `sdlc` includes that fragment in `set-status` help (and any other help that states
  lifecycle facts — `claim`/`close`) via `go:embed` of the generated fragment, regenerated
  by the same generic `go generate` / `make vocab-embed`, with the generic git-diff
  freshness check (like the embedded JSON, #122 M3/M4).
- Reuses the per-language binding insight: the help fragment is just another generated
  face of the model.

## Done when

- Changing `construct/vocabulary/issue.cue`'s lifecycle (add/remove an edge) + regenerating
  updates `sdlc issue set-status --help` with **no hand edit**.
- A stale help fragment fails the freshness check (CI-catchable).
- The hand-maintained lifecycle claims (e.g. the "all other transitions allowed" class)
  are gone — the facts are model-derived.

## Plan

- [ ] Design at `start-plan`: which help surfaces restate lifecycle facts; fragment shape; embed vs include
- [ ] Renderer emits the lifecycle help fragment from the model
- [ ] Wire it into the helptext build (`go:embed`/include) + the generic freshness/diff gate
- [ ] Delete the hand-maintained lifecycle prose; tests

## Log

### 2026-06-25

- Filed as a #122 follow-up (the prose-consumer half of "compiled to consumers"). Motivated
  by the M4 help-text drift the close review caught — `issue.cue` was, for the help surface,
  still just-documentation that didn't derive from the model. The general principle (lessons):
  *every prose surface that restates the model is a shadow; generate it or reference it.*
