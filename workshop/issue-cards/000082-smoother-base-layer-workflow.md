---
id: '000082'
status: done
created: 2026-06-04
updated: 2026-06-04
estimate_hours: 3.5
actual_hours: 3.5
---

# smoother cross-repo base-layer workflow

## Problem

ariadne is the **base layer**: derivative repos (nous, pair, …) symlink its
substrate (per `construct/setup.sh`), so a change to ariadne's working tree is
*live* in every derivative immediately. In this stage of development the base
layer is high-churn — it's still being extended — which the symlink model is
designed to support (iterate the base while developing on it).

The cost is **concurrency the operator can't easily reason about**. At the leaf,
conflicts are obvious ("am I editing the same thing in two pair sessions?"). At
the base, reasoning breaks down: it's easy to forget that work in nous *and* pair
will collide *if both touch ariadne*. The shared-live working tree means one
session's uncommitted/branched base state is invisibly read by every other
session. Two patterns make it worse:

1. **Base changes are driven from any derivative** (the AGENTS.md "peers" model —
   an agent in nous can just go make the ariadne change). Convenient, but the
   base work is initiated from a context that doesn't "own" ariadne, so it's easy
   to lose track of who's touching the base.
2. **Even *filing* a base issue (not working it) collides with sdlc gates.** A
   new `workshop/issues/*.md` is untracked working-tree residue; with #78/#80 it
   no longer blocks a merge by accident, but the broader friction remains — base
   issues are discovered mid/late in a *derivative* session, and capturing one
   shouldn't entangle with the base repo's working-tree state at all.

This issue makes the base-layer workflow smoother **without adding a gate to the
common path** — drafting and brainstorming stay free; the one new check fires
only where commitment happens.
