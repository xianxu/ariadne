---
id: 000224
status: working
deps: []
github_issue:
created: 2026-09-12
updated: 2026-09-12
estimate_hours:
started: 2026-09-12T18:26:36-07:00
---

# ARCH-FUNERAL: everything created names its end — a funeral plan is part of the design

## Problem

The registry has a principle for how a component behaves under load
(`ARCH-CONSTRAINTS`) and one for state carried between events
(`ARCH-ORDER`), but none for **what a program creates and never destroys**.
Agents are systematically weak here: the payoff of a retention rule is
months out, the cost of omitting one is invisible at review, and "append a
record" always passes.

Field evidence, pair, 2026-09-12:

- Pair's data store is 13 GB. One event log is 4.8 GB. 143 per-thread
  ledgers hold 73 MB, 99% of it launch snapshots that stop mattering the
  moment the launch is superseded. Nothing in the tree prunes any family
  (pair#239).
- The same day, a thread became unresumable because its append-only ledger
  crossed an 8 MiB read cap on its 14th relaunch — a cliff the design
  reached by construction, since the file was defined to grow and no one
  had said how far (pair#237, pair#238).
- The agent's own store next door is bounded: Claude Code sweeps its
  transcripts after 30 days. Pair records more about the same sessions and
  keeps it forever.

The gap is not "we forgot to add GC". It is that the artifact's **end** was
never part of its design, so nothing at plan time asked for it and nothing
at review time could flag its absence.

## Spec

Add `ARCH-FUNERAL` to `cmd/sdlc/internal/judge/architecture.md` in the
registry's shape (principle / at-plan / at-review), and to every site that
enumerates the marker set: the expected list in
`cmd/sdlc/internal/judge/judge_test.go` (~line 363) and the narrative in
`atlas/workflow/architecture-principles.md`. Draft:

> ## ARCH-FUNERAL — Everything created names its end
>
> - **principle:** A thing that is created has a lifecycle, and its end is
>   part of the design, not an operational afterthought. For every durable
>   artifact, record, cache, or handle a component produces, the design
>   says when it stops being needed and what removes it: a time-bound sweep,
>   a superseding generation, an owner whose death collects it, or an
>   explicit growth bound with the consequences of reaching it stated.
>   "Append-only" is a durability property, not an exemption — an append-only
>   store still names its bound and what happens at it. The lens is the
>   *end* of a thing's life where ARCH-CONSTRAINTS is its *rate* and
>   ARCH-ORDER its states in between.
> - **at-plan:** For each new artifact family, record, or growing structure,
>   write one line: who creates it, who is the last to need it, what
>   removes it, and how big it can get before that. If the answer is
>   "forever", say what forever costs per unit of use (per launch, per
>   session, per keystroke) and where it is measured. Target rather than
>   sweep: an in-memory value dies with its scope and needs no entry; a
>   file, a row, a persisted cache, or a background process does. `N/A` is
>   a claim that can be wrong: write it as "creates nothing durable because
>   X", never as a bare marker.
> - **at-review:** Flag any new file family, log, ledger row, cache, or
>   sidecar with no removal path in the diff or in an existing routine, and
>   any writer to an existing family whose growth per event is larger than
>   before (a row that embeds a snapshot; a log that gains a field). Flag a
>   size or age cap on the *reader* of a growing artifact with no matching
>   bound on its *writer* — that is a cliff, not a bound. Flag a cleanup
>   that only the operator can trigger when the growth is per-launch or
>   per-session.

Name: the operator's choice over ARCH-LIFECYCLE — "a funeral plan" is the
sentence an author remembers when they add a file family.

Also: the human narrative in `AGENTS.base.md` "Core Design Principles" lists
the two tactics that have no registry entry; this one *does* get an entry,
so nothing changes there beyond whatever the base-layer propagation carries.

## Done when

- `sdlc arch-principles` prints `ARCH-FUNERAL` between `ARCH-ORDER` and the
  end, in the registry's shape.
- The judge test's expected marker list includes it; the atlas narrative
  places it against ARCH-CONSTRAINTS and ARCH-ORDER as above.
- `sdlc start-plan` and the plan-quality / boundary-review prompts deliver
  it (they embed the same file — assert with the existing embedding test).
- Propagated to dependents via `sdlc propagate-base`; pair's `sdlc
  arch-principles` shows it.
- pair#239 is designed against its `at-plan` text (that issue depends on
  this one).

## Plan

`atlas/workflow/architecture-principles.md` "Adding an entry" says a new entry
touches FOUR things, and the goldens were the one this Plan originally missed —
the site that, before #208, could keep passing while covering nothing.

- [ ] Add the section to `cmd/sdlc/internal/judge/architecture.md`, with the
      ARCH-CONSTRAINTS boundary as revised below
- [ ] Extend `TestArchitectureMarkers`' hand-written `want` list — the one
      deliberate non-derived site, kept as the registry tripwire
- [ ] Re-capture the four goldens
      (`go test ./cmd/sdlc/internal/judge -run TestBuildPrompt_Golden -update-golden`)
      and read the diff: each must gain exactly this entry and nothing else
- [ ] Map-level paragraph in `atlas/workflow/architecture-principles.md` —
      boundaries, shaping choices, provenance. Never a copy of the clauses
      (#215 BR-1)
- [ ] Verify delivery from the binary: `sdlc arch-principles` renders it, and
      `sdlc start-plan` / the gate prompts carry it via the existing embedding test
- [ ] Propagate to dependents (`sdlc propagate-base`, run outside the sandbox);
      confirm from pair with `sdlc arch-principles`
- [ ] Close; pair#239 picks up the `at-plan` text

## Log

### 2026-09-12

- Filed from the pair session that fixed pair#237 and measured pair's
  13 GB store. The operator named it: ARCH-FUNERAL, "more sentimental" than
  ARCH-LIFECYCLE — and more memorable, which is the point of a marker.

## Revisions

### 2026-09-12 — the neighbour boundary, and the missing fourth site

**Boundary sentence replaced.** The draft closed the `principle:` clause with:

> The lens is the *end* of a thing's life where ARCH-CONSTRAINTS is its *rate*
> and ARCH-ORDER its states in between.

The triad is tidy but wrong about ARCH-ORDER, which governs state a **component**
carries between externally-arriving events — not the life stages of an artifact.
Writing a wrong neighbour boundary into the registry is the specific thing
`atlas/workflow/architecture-principles.md` tells future editors to watch for,
and ARCH-ORDER's own entry had to be fenced against ARCH-PURE and ARCH-SECURE for
the same reason.

It also leaves the ARCH-CONSTRAINTS boundary weaker than the field evidence
supports. "Rate" makes this sound like a member of that entry's list — which is
the test #215 used to reject folding ARCH-ORDER into ARCH-PURE ("that entry's IO
list enumerates *members*, and this is a difference in *kind*"). The difference
in kind is available and it is sharper: **ARCH-CONSTRAINTS budgets a component
while it is working; this one governs what is left behind once it has stopped.**
Pair's 13 GB store was not produced by load — it was produced by a year of
ordinary operation in which every individual write sat comfortably inside any
envelope anyone would have written down. That is a defect ARCH-CONSTRAINTS'
`at-review` clause cannot flag, which is the bar for a separate entry rather than
a new bullet.

pair#237 makes the same point from the other side: the 8 MiB read cap **was** the
declared constraint, and declaring it is what turned unbounded growth into a
cliff. ARCH-CONSTRAINTS has no vocabulary for "who is the last to need this",
which is the question that would have caught it.

So the clause now reads as a load/residue split, and drops ARCH-ORDER.

**Plan gained the goldens.** The Spec enumerated three sites; the atlas documents
four. `cmd/sdlc/internal/judge/testdata/golden/*.prompt` was the missing one —
and per the atlas it is precisely the site that used to pass while covering
nothing. Derived the real set from what #215 actually touched rather than from
the Spec's list.
