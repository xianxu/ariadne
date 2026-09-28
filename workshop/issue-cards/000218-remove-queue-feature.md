---
id: '000218'
status: done
started: 2026-09-09T10:48:12-07:00
created: 2026-09-09
updated: 2026-09-09
estimate_hours: 2.75
actual_hours: 1.80
---

# Remove the sdlc queue verb and workshop/queue.md

## Problem

`sdlc queue` and its trunk-backed `workshop/queue.md` (ariadne#209, merged
2026-09-07) are being withdrawn one day after landing. The operator's existing
practice — comments in pair's draft nvim pane as sticky notes — already worked,
is more versatile, and the queue never displaced it.

**The design error, precisely.** The chain was: *file in a repo* -> *repos have
branches* -> *branches make it stale* -> *read from the trunk* -> *needs a verb*
-> *needs worktree-free CAS writes with intent replay* -> eight review rounds.
Every step follows from the one before, but the second created the problem the
rest solved. A list of what to do next has no reason to differ per branch; it
only did because it was stored somewhere that versions.

A plain tracked `workshop/queue.md` would have been fine. The trunk-backing
solved staleness that, for an advisory list nobody is blocked on, was not worth
solving.

**Three deeper reasons it was the wrong shape**, worth recording because they
generalize:

1. **Wrong subject.** Every noun in the vocabulary describes the *object* of work
   — an issue is a unit of it, a project a container with a `done_when`, a target
   an invariant, a roadmap a month of it. A queue describes the *subject*: the
   operator's own evolving relationship to the work. Persistent, not
   deliverable-shaped, possibly multi-repo, constantly revised, made of labels
   rather than state. Nothing in `workshop/` fits that, which is why the
   container fought the content the whole way.

2. **Wrong scope.** Per-repo could never express "ariadne#207 before pair#171",
   and the original issue recorded that as a "stated limit" — usually the phrase
   for *the container is wrong*. The cross-repo ordering turns out not to need an
   artifact at all: it is a judgment remade cheaply each session when the operator
   chooses which context to work in, and a derived view (an LLM reading the local
   notes) beats a stored one.

3. **Wrong cost curve.** With a single operator there is no coordination cost, so
   freezing scope buys nothing and re-prioritization *is* the advantage. The queue
   put its most expensive machinery — validation, CAS, a commit on main — directly
   on the highest-frequency operation. Reordering should be free.

**Not a failure of execution.** The implementation was reviewed across eight
rounds and is correct; ariadne#209 closed at est 3.06h / actual 4.49h with the
findings resolved. This removes a thing that works because it should not exist,
which is a different judgement from "it is broken".
