---
id: 000199
status: punt
deps: []
github_issue:
created: 2026-08-21
updated: 2026-09-02
estimate_hours:
---

# sdlc: expose query API to peer actors

## Problem

`sdlc`'s command line is the API between an agent and its own harness. In the
actor model being built in `pair#145` (couch), agents also need to answer
*queries from other actors* — "what is the state of `ariadne#111`" — without the
caller reaching into the callee's repo and without spending an LLM turn.

Two things are missing. There is no way for a command to declare that it is safe
to answer from outside, so today the choice is all-or-nothing. And there is no
way for a caller to discover what a given actor will answer, which matters
because binaries skew across repos and per-repo policy may differ.

The constraint that shapes the design: **the dialect is shared, the authority is
not.** Peer ariadne-based repos all understand the same vocabulary, but actor B
invoking `sdlc close` inside A's repo is permission laundering and breaks the
rule that only an actor interprets its own state. Test for the split: a call
that changes the *caller's* world is a query; one that changes the *callee's*
world is a command. Only queries get exposed.

Design context:
`brain/workshop/pensive/2026-08-20-01-pensive-couch-agent-switcher.md`.

## Spec

Three legs, each extending a pattern already in this tree:

1. **Per-command exposure annotation, default-deny.** A cobra `Annotations`
   entry per command declares whether it is externally answerable. A drift test
   asserts every command declares one explicitly, so a new verb can be neither
   silently exposed nor silently forgotten — the same shape as
   `processmanual.WorkflowVerbs` and its enumeration guard.
2. **`--json` as the wire format.** Already the convention (`sdlc state --json`).
   Orthogonal to the annotation and both are required: `actual` and
   `estimate-source` are machine-readable reads that should stay private
   (velocity data, brain calibration docs), so JSON-ness alone must not imply
   exposure.
3. **Vocabulary-derived response shapes.** Responses derive from
   `construct/vocabulary/*.cue` — already the single source the status enum
   derives from. This is the leg that makes it a *dialect* rather than a private
   API: a peer understands the answer without knowing internals.

**Capability discovery.** `sdlc --expose-manifest` prints the exposed verb set,
arg shapes, and version as JSON. Denied verbs are **omitted, not listed as
denied** — listing them leaks internals and invites asking for them. A caller
owning its own copy of `sdlc` is not a substitute for the manifest: the manifest
is the callee declaring *policy and version*, which is what survives binary skew
across repos.

**Not in scope here.** Transport is couch's (`pair#145`): couch hands
`{query, verb, args}` to the owning actor's shell, which runs the binary in its
own repo root under its own permissions. The caller never runs the binary. This
issue only makes `sdlc` declare and serve its exposed surface.

The generalisation matters — the annotation + `--json` + manifest convention
should be adoptable by any ariadne-family binary, not special-cased to `sdlc`.

## Done when

- Every `sdlc` command declares an exposure annotation, enforced by a drift test
  that fails on an undeclared command.
- `sdlc --expose-manifest` emits the exposed set with arg shapes and version;
  unexposed verbs are absent from the output entirely.
- Every exposed verb has a `--json` form whose shape derives from
  `construct/vocabulary/*.cue`, verified by a test rather than by inspection.
- No spine/mutating verb is exposed; a peer cannot reach `claim`, `close`,
  `merge`, or `push` through the manifest surface.
- The convention is documented well enough for a second binary to adopt it
  without copying `sdlc`'s implementation.

## Plan

- [ ] Annotation + default-deny drift test over the command tree.
- [ ] `--expose-manifest` with arg shapes and version; omit denied verbs.
- [ ] Vocabulary-derived response shapes for the exposed set, test-locked.
- [ ] Document the convention for adoption by other ariadne-family binaries.

## Log

### 2026-08-21

Filed as an enabler for `pair#145` (couch) M3. Design settled in the pensive
above: the deterministic query API is literally the CLI, filtered — no second
surface, so no drift between what the local agent sees and what a peer sees.

### 2026-09-02

Punted. The only consumers were `pair#147` (cluster transport) and `pair#148`
(brain advisor), both punted by the couch-lite rescope (`pair#170`). With couch
narrowed to a single-host switcher, no actor asks another actor anything, so the
exposure annotation has nothing to serve. The design stands if cross-actor
queries return.
