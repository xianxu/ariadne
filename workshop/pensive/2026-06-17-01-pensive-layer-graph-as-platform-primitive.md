---
type: pensive
date: 2026-06-17
topic: layer graph as a platform primitive
mode: eureka
description: The construct/deps layer graph is a platform primitive any subsystem can be DAG-aware against — not weave's private state. Pattern surfaced while designing DAG-merged dynamic skills.
references: [workshop/issues/000115-dag-merged-dynamic-skills-per-repo-datatype-enumeration-across-the-layer-graph.md, workshop/history/000111-make-datatype-discovery-better.md, atlas/workflow/weave.md]
---

# Pensive: Layer graph as a platform primitive

Working out how the datatype skill's noun-list should be *per-repo* (merged across
the dependency DAG, not a single owner-defined set) surfaced something bigger than
the bug. I'd been treating "DAG-awareness" as weave's job — as if anything that
needs to merge across layers has to be taught to weave. But the layer graph
(`construct/deps`) is just a public, on-disk fact about a repo. weave is its
*primary* consumer, not its *owner*. The datatype generator can read that graph
itself, at runtime, in whatever repo it's invoked — and weave never needs to learn
what a datatype is. That's the reframe: **the layer graph is a platform primitive,
and most things in ariadne should be DAG-aware in their own right.** This isn't
incidental — a big part of what ariadne is *for* is solving the sharing-and-
composition problem in the agentic world: what gets shared, what gets composed,
and how it resolves per consumer. If that's the mission, DAG-awareness shouldn't be
one module's privilege; it should be substrate.

The shape that keeps this from sprawling is a three-way split. **Mechanism is
shared and single** — "what is repo R's layer graph" lives in one library every
DAG-aware tool imports; that's the invariant, because the moment two tools walk the
graph differently you get subtle divergence bugs. **Policy is per-subsystem** — what
you *do* with the merged layers (datatype: union, local-wins-by-filename; prose:
the visibility math) belongs to each subsystem, sized to its own data. **Execution
is sequenced by weave but blind to content** — when a subsystem emits a compile
artifact, weave decides *when* it runs (the generate stage) and stays ignorant of
*what* it produced. weave owns when; the tool owns what; weave never owns content.

The payoff is open extension. The alternative — weave accreting a merge per
data-shape — makes it a god-module every new feature has to crack open. With the
graph as a primitive, a new DAG-sensitive subsystem is just "a build-in-owner tool
that reads the shared graph, owns its merge, and (if it produces an artifact)
carries an opaque marker weave sequences." weave stays a thin, generic compiler.
The design default falls out of it too: **alongside weave (a DAG-aware tool) unless
the thing genuinely needs weave's own composition/lowering.**

## Open questions

- How far does this generalize? What *else* wants to be DAG-aware — per-repo config,
  other cross-layer enumerations, generated indexes? Is there a second consumer that
  would validate the shared-walk library before I over-fit it to datatype?
- Where exactly is the line between "in weave" and "alongside weave"? "Needs weave's
  composition/lowering" is the rough cut, but I haven't pressure-tested it.
- Per-subsystem merge policies are freedom *and* fragmentation — when does *not*
  sharing weave's visibility math actually bite (a subsystem that quietly needs the
  `𝒜(R)` export/internal semantics and reinvents them wrong)?
- Maturity: this is still a pattern I'm feeling out. Does it earn a `target`
  (a defended invariant) once a second subsystem uses it — and is "the layer walk is
  one shared library" the invariant worth pinning?
