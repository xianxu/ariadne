---
id: 000209
status: open
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
---

# queue datatype for advisory work ordering

## Problem

**Sequence has nowhere to land.** The existing nouns each hold something else:
an issue holds *what*, a project holds *scope*, `deps:` holds *hard blocking*,
a roadmap holds the month-level aggregate, a project's `## Breakdown` holds
ordering *within* that project. Nothing holds "of the loose issues that could
be done in any order, which one is next, and why."

So it lives in conversation and dies with it. On 2026-09-02 a single advisory
session produced five issues across two repos — `pair#170`, `#171`, `#172`,
`ariadne#206`, `#207` — with a real order between them (`#207` after `#206`,
same dispatch; `#172` after `#170`, needs its switch semantics; `#171`'s
measurement before `#171`'s design). None of that ordering is recorded anywhere.
Meanwhile `sdlc state` prints ariadne's 16 open issues in ID order with no
ranking at all.

## Spec

Add **`construct/datatype/queue.md`**, a datatype describing one file per repo
at `workshop/queue.md`. Scope is the datatype prototype and nothing else — no
verb, no gate, no tooling (see *Deferred*).

**Semantics: advisory, never binding.** The queue indicates a section of work
that will likely happen — it is not a commitment. Execution still goes through
the hard blockers: `deps:` is authoritative and the queue never overrides it. If
a line is really a block, it belongs in `deps`, or there are two truths about
blocking that will disagree.

**Line format: one ref per line plus a few words of why-now**, with an optional
project tag for grouping.

```
- pair#171 — floor under attention; the menu case may prove the trigger wrong [couch]
- ariadne#207 — after #206 merges, same dispatch [sdlc]
```

**Two kinds of entry, and the datatype must distinguish them:**

- An **issue** line is a next action — something that can be started.
- A **project** line is a declared intent to work in that area. Coarser, and it
  should be replaced by an issue line once it becomes the actual next thing.
  Name the failure mode in the prose: a project line that never becomes
  actionable is permanently present, carries no ordering information, and
  trains the reader to stop reading the file.

**Why storing the rationale is safe.** The couch pensive's distinction applies:
labels tolerate staleness, state does not. A why-now note is a *label* — stale,
it is still better than an order you cannot evaluate. This is also why the queue
must not carry status: that would be state, and state goes confidently wrong.

**Relationship to siblings, to be stated in the datatype:**

- `deps:` — hard blocking. The queue is soft preference among the *unblocked*.
- project `status` (`paused`/`committed`/`executing`) and `roadmap` — these are
  where *project-level* prioritization already lives. The queue does not replace
  them.
- a project's `## Breakdown` — authoritative for ordering within that project.
  **The queue is for loose ends**, or the two orderings will disagree.

**Usage stays freeform.** No new verb. The operator says "pick the next thing
off the queue" and the agent reads the file. Let real use reveal the shape
before anything is mechanized.

## Deferred, with reasons

**Close-gate removal.** The rot risk is real and this repo has the scars — both
metis project files froze on 2026-07-22, before the phase that mattered. The
fix, when it comes: `sdlc close` parses queue lines with the **same resolver
`sdlc resolve` uses** (not grep — `#11` substring-matches `#111`, and a bare
`#11` is ambiguous across repos), removes any line resolving to the closing
issue, and announces the removal loudly. That mirrors the close gate's existing
fleet-wide project sweep, so it is one more consumer of an existing mechanism
rather than a parallel one. Deferred deliberately: the format is not known until
the file has been used, and building enforcement for an unused format is how
couch overran. Revisit after a few weeks of real use.

**Cross-repo ordering.** Per-repo only. The operator's real attention order
spans repos and `workshop/queue.md` cannot express "ariadne#207 before
pair#171". Single-threaded execution means picking a repo first, so this is a
stated limit, not a blocker. A brain-hosted fleet queue is the candidate.
Unsettled: the brain charter excludes SDLC process artifacts, and the operator
does not consider a queue to be one — so whether the charter bars it is an open
disagreement, not a settled no.

## Done when

- `construct/datatype/queue.md` exists and follows the sibling prototypes'
  shape (`type: type`, `name: queue`, a discovery `description`, then narrative
  prose).
- Its description triggers on the natural phrasings — "queue", "what's next",
  "pick the next thing", "plan the sequence".
- The prose states: advisory-not-binding and its relationship to `deps`; the
  division of labour with project `status`/`roadmap` and with a project's
  `## Breakdown`; the line format; and the issue-line vs project-line
  distinction with its failure mode.
- The prose names the rot risk and points at the deferred close-gate removal,
  so a later reader finds a recorded decision rather than an apparent gap.
- `construct/generated/datatype/SKILL.md` picks `queue` up **without a
  hand-edit** — the `datatype` binary DAG-merges `construct/datatype/*.md` and
  generates that description list, so regeneration is the only step. If it needs
  a manual edit anywhere, that is a bug in the single-sourcing, not a task here.

## Plan

- [ ] Write `construct/datatype/queue.md`.
- [ ] Regenerate and confirm `queue` appears in the generated datatype SKILL
      description with no hand-edit.
- [ ] Seed `ariadne/workshop/queue.md` with the real current ordering as the
      first exercise of the format — it is the cheapest possible test of
      whether the shape is right.

## Log

### 2026-09-02

Operator scoping: "it's just a datatype; how it's being used for now can be
freeform." Hence no verb and no gate in this issue — the enforcement design is
recorded under *Deferred* so it is not lost, and not built until the format has
survived contact with use.

Design point settled during the discussion: the queue is a **hybrid** — it
indicates a section of work likely to happen, not a commitment. That sits
between a roadmap's month-level intent and an issue's concrete unit, and it is
what justifies allowing project lines at all despite a project not being a next
action.
