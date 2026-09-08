---
type: type
name: queue
description: Use when deciding or recording WHAT TO DO NEXT among work that could be done in any order — the soft ordering that currently lives in conversation and dies with it. Triggers on "what's next", "pick the next thing off the queue", "queue this up", "plan the sequence", "what should I work on", "/xx-datatype queue". One file per repo at workshop/queue.md, stored on the trunk and read through `sdlc queue`. Advisory, never binding — distinct from deps (hard blocking), from project status/roadmap (project-level prioritization), and from a project's Breakdown (ordering within one project). The queue is for loose ends.
---

# queue

A queue is the **soft ordering layer**: of the loose issues that could be started
in any order, which one is next, and why.

Every other noun in this system holds something else. An issue holds *what*. A
project holds *scope*. `deps:` holds *hard blocking*. A roadmap holds the
month-level aggregate. A project's `## Breakdown` holds ordering *within* that
project. Nothing held "these five are all unblocked; here is the order I'd take
them in, and here is why" — so it lived in conversation and died with it.

## Advisory, never binding

The queue indicates a section of work that will *likely* happen. It is not a
commitment, and it never overrides `deps:`.

That boundary is load-bearing. `deps:` is the authoritative record of hard
blocking; the queue is soft preference among the **unblocked**. If a line is
really a block, it belongs in `deps` — otherwise there are two truths about
blocking, and they will disagree.

For the same reason the queue carries **no status**. The distinction is between
labels and state: labels tolerate staleness, state does not. A why-now note is a
label — stale, it is still more useful than an order you cannot evaluate. A
status would be state, and stale state goes confidently wrong.

## Two kinds of entry

An **issue line** is a next action — something that can be started now.

A **project line** is a declared intent to work in that area. It is coarser than
a next action, and it should be *replaced* by an issue line once it becomes the
actual next thing. This is the hybrid that justifies the datatype: the queue sits
between a roadmap's month-level intent and an issue's concrete unit.

The failure mode to watch for: a project line that never becomes actionable is
permanently present, carries no ordering information, and trains the reader to
stop reading the file. A project line that has survived several passes is a
prompt to write the issue, not a fixture.

## The line format

One entry per line:

```
- ariadne#207 — after #206 merges, same dispatch [sdlc]
- project:sdlc-fleet — the whole policy area is next [sdlc]
```

A ref, an em-dash surrounded by spaces, a few words of why-now, and an optional
trailing `[tag]` for grouping. A `project:` prefix on the ref marks a project
line; without it the entry is an issue line. The why-now carries no newline,
control character, ` — ` or brackets — those are rejected rather than escaped,
because the format's whole value is that a human reads it as a list, and an
escape costs exactly that.

A line that does not parse is **preserved verbatim, never discarded** — the file
is co-authored by hand and by the verb, and silently losing an operator's line
would be the worst failure this datatype could have. Such a line is not listed as
an entry, and `sdlc queue` reports how many it found; a hand-edit that used a
hyphen instead of an em-dash is the usual cause.

## It lives on the trunk

One file per repo at `workshop/queue.md`, and **`origin/main` is the permanent
base**. Every read and every edit goes through `sdlc queue`, which talks to the
trunk directly — so the queue reads the same from a feature branch, a detached
HEAD, a second worktree, or another machine.

**Read it through the verb, not by opening the file.** The file is tracked, so
your branch has a copy, and that copy is whatever your branch last saw. That
staleness is the exact thing the trunk-backed design exists to remove, and
`cat`-ing the file quietly puts it back.

The *operational* contract — the flags, the interleaving policy when two
checkouts edit at once, and what happens offline — lives in `sdlc queue --help`
and is not restated here. The split is by kind, not by convenience: the format
above is part of what a queue entry **is**, so it belongs to the noun; the
operations are what the verb **does**.

## Distinct from sibling datatypes

- `deps:` — hard blocking, authoritative. The queue is preference among the
  unblocked, and defers to it always.
- `project` `status` (`paused`/`committed`/`executing`) and `roadmap` — these are
  where *project-level* prioritization already lives. The queue does not replace
  them; a project line in the queue is a pointer into that world, not a second
  copy of it.
- a project's `## Breakdown` — authoritative for ordering *within* that project.
  **The queue is for loose ends.** Queuing the individual steps of a project that
  already has a breakdown gives you two orderings that will disagree, and the
  breakdown is the one that wins.
- `pensive` / `parley` — where an ordering is *worked out*. The queue is where
  the conclusion lands so it outlives the session.

## The rot risk, and what is deliberately not built yet

A queue nobody prunes becomes a queue nobody reads. This repo has the scars:
both metis project files froze on 2026-07-22, before the phase that mattered.

The fix, when it comes, is that `sdlc close` removes any line resolving to the
closing issue and announces the removal — using the **same resolver `sdlc
resolve` uses**, not grep, because `#11` substring-matches `#111` and a bare
`#11` is ambiguous across repos. That mirrors the close gate's existing
fleet-wide project sweep, so it would be one more consumer of an existing
mechanism rather than a parallel one.

It is deliberately **not built yet**. The format is not known until the file has
been used, and building enforcement for an unused format is how couch overran.
Revisit after a few weeks of real use. This paragraph exists so a later reader
finds a recorded decision rather than an apparent gap.

## Known limit: per-repo only

`workshop/queue.md` cannot express "ariadne#207 before pair#171". Real attention
order spans repos; this does not. Since execution is single-threaded and picking
a repo comes first, it is a stated limit rather than a blocker. A brain-hosted
fleet queue is the candidate — unsettled, because the brain charter excludes SDLC
process artifacts and whether a queue counts as one is an open disagreement, not
a settled no.
