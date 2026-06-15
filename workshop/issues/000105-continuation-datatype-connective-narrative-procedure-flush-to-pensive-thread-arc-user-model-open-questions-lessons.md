---
id: 000105
status: working
deps: []
github_issue:
created: 2026-06-15
updated: 2026-06-15
estimate_hours: 2
---

# continuation datatype: connective-narrative procedure (flush-to-pensive, thread-arc/user-model, open-questions, lessons)

## Problem

The `continuation` datatype (`construct/datatype/continuation.md`) distills a
session's human-meaningful state into a durable handoff doc. Dogfooding (pair#61)
surfaced five gaps against what a high-quality handoff needs:

1. **No flush-first step.** The procedure assumes the session's durable artifacts
   already exist; nothing first captures un-recorded key exchanges, so insights
   die in the transcript.
2. **No lessons.** It records *this-work* decisions/dead-ends but not the
   transferable **lessons** learned in the session.
3. **Lists, doesn't connect.** "Pointers" enumerates files/issues but never
   explains their **history, reasoning, and connections** — the continuation
   isn't built *around* the durable artifacts.
4. **No thread arc / user model.** It captures no **arc of the thread** (where we
   started, the pivots, their underlying connection), no **model of the user's
   mental model** / latent intention, and no channel for the **open questions**
   that model leaves unresolved (to ask on resume).
5. **NEXT ACTION stands alone** — concrete, but not connected to the arc/lessons.

Relationships: mirror of **pair#61** (the dogfood origin). **ariadne#103** wants
the user-model *maintained live every turn*; this issue **persists/checkpoints**
that model into the continuation (the two are complementary halves). **ariadne#90**
(docflow suspend/resume + auto-summary) is the sibling-domain analog.

## Spec

Reframe the continuation as the **connective narrative over a session's durable
artifacts**: detail lives in the flushed artifacts; the continuation explains how
they connect and where the user's head is. This reframe is what lets the new
reflective sections stay concise instead of bloating the doc — it preserves the
standing "Resumable, not exhaustive" rule rather than fighting it.

**Procedure — add Step 0 (flush).** Before distilling, scan the session for
results of key exchanges not yet captured durably and **flush them to `pensive`**
(the low-friction sink for not-yet-structured insight). Automated flush is
`pensive`-only: do **not** auto-create `target`s (promoting an insight to a target
is human-review-gated) and `meeting-notes` is not the right fit here. Issues
already in flight may be updated as normal work. If it's all already captured, say
so and proceed. Rationale: the continuation *builds on* durable artifacts, so they
must exist first — and pushing detail into them is what keeps the continuation
terse.

**Body skeleton** (`*` always present, `†` omit if truly empty):

- `## NEXT ACTION` `*` — kept; now explicitly **tied to the arc/lessons** (gap 5).
- `## State of play` `*` — kept (per-issue status; point at `sdlc state`/issues).
- `## Thread arc & user model` — **NEW** (gap 4). Where the session started, how it
  pivoted, the underlying connection among the pivots, and the inferred latent
  intention → a model of the user's mental model. Two constraints (shared with
  ariadne#103): the model must be (a) internally self-consistent and (b) fit the
  observed interactions. Concise — a few tight paragraphs, not a transcript.
- `## Open questions` `†` — **NEW** (gap 4). Internal inconsistencies in that model
  / unresolved ambiguities. **Lead the section with the embedded resume directive:**
  *"On resume, resolve these open questions with the user before continuing with
  the NEXT ACTION."* This is how "ask on resume" is delivered — embedded in the
  generated file, not via a pair seed-prompt change (pair#61 decision 5).
- `## Artifact map` — **REFRAME** of "Pointers" (gap 3). The flushed
  pensive/issues/targets/key files **with their history, reasoning & connections**
  — a narrative ("pensive X, written because …, constrains issue Y, toward target
  Z"), not a bare list — plus read-first ordering, branch/worktree, cross-repo
  paths.
- `## Live deliberations` `†` — kept.
- `## Decisions & dead ends` `†` — kept.
- `## Lessons learned` `†` — **NEW** (gap 2). Transferable meta-lessons (codebase,
  process, tooling, working with this user), distinct from this-work decisions.

**Rules / authoring updates.**
- Reaffirm "Resumable, not exhaustive": the new reflective sections are concise and
  high-signal and **reference** artifacts rather than restate them; the flush-first
  step is what earns the terseness.
- Writer stays minimal: `pair-continuation` still enforces only `## NEXT ACTION`
  (pair#61 decision 4) — the richer skeleton is authoring discipline, not a binary
  guard.
- Add **Search recipes** for the new sections (e.g. open questions across
  continuations).

## Done when

- `continuation.md` procedure has Step 0 (flush → `pensive`, with the
  `target`/`meeting-notes` carve-outs).
- Body skeleton carries Thread arc & user model, Open questions (with the embedded
  resume directive), Lessons learned, and the Artifact-map reframe; the NEXT ACTION
  rule ties it to the arc/lessons.
- Rules reaffirm concision / narrative-over-artifacts and record "writer enforces
  NEXT ACTION only."
- Search recipes cover the new sections.
- Cross-links to pair#61, ariadne#103, ariadne#90 are present.

## Plan

- [x] Rewrite `construct/datatype/continuation.md` — Authoring step 1 (flush →
      pensive, with target/meeting-notes carve-outs).
- [x] Body skeleton: Thread arc & user model, Open questions (verbatim resume
      directive), Artifact-map reframe, Lessons learned; NEXT ACTION tied to arc.
- [x] Rules (concision / narrative-over-artifacts + "writer enforces NEXT ACTION
      only") and Search recipes for the new sections.
- [x] Cross-links: pair#61, ariadne#103 (single-sourced user-model), ariadne#90.
- [x] Confirm nothing pins the old skeleton (no Go test does; no datatype lint exists).
- [x] Atlas synced (`atlas/workflow/data-artifacts.md` — connective-narrative framing).

End-to-end dogfood validation is **pair#61**'s deliverable (cross-repo, not a
blocker for this datatype change); it is intentionally not a checkbox here.

## Log

### 2026-06-15
- Created from the pair#61 dogfood. Design + decisions agreed with operator
  (see pair#61 decisions 1–6): sections/order, flush = pensive-only, writer
  enforces NEXT ACTION only, resume directive embedded in the generated file.
- Rewrote the datatype. **Base-layer propagation:** `construct/datatype` is a
  `symlink` entry (`construct/base.manifest:152`), so this flows to every
  downstream ariadne-styled repo by symlink — consistent, no per-repo merge, and
  no recompose needed for content. Risk is low: the richer skeleton is authoring
  guidance, not a parsed schema, so continuations already written downstream under
  the old skeleton don't break.
- DRY (plan-quality finding): user-model definition single-sourced — ariadne#103
  is canonical for the discipline; the continuation section defers to it and adds
  a back-pointer in #103.
