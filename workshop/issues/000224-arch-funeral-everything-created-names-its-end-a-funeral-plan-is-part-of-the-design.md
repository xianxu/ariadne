---
id: 000224
status: working
deps: []
github_issue:
created: 2026-09-12
updated: 2026-09-12
estimate_hours: 0.97
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
- The shared binary is rebuilt (`make sdlc-build` in ariadne) and `sdlc
  arch-principles` run from pair's cwd renders eight entries including
  `ARCH-FUNERAL`.
- pair#239 is designed against its `at-plan` text (that issue depends on
  this one).

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec           design=0.35 impl=0.05
item: smaller-go-module    design=0.05 impl=0.12
item: atlas-docs           design=0.05 impl=0.10
item: milestone-review     design=0.0  impl=0.18
design-buffer: 0.15
total: 0.97
```

Anchored on #215, the same change one entry earlier: est 1.49, **actual 0.61**.
Two rows are priced directly off that overrun.

`issue-spec` is the row that moved. #215 priced it 0.8 because its window
genuinely contained the original authoring — nothing pre-resolved that design, it
*was* the design. Here the spec and the draft clause text arrived already written
from a pair session, so this row covers only what this window adds: reading the
registry and its two confusable neighbours, the ARCH-CONSTRAINTS boundary
critique that replaced the draft's closing sentence, correcting the delivery step,
and two plan-quality rounds. That is a smaller job than authoring the principle,
and pricing it at 0.8 again would repeat exactly the error #215's own revision
diagnosed.

`smaller-go-module` impl stays at #215's corrected 0.12 — the `want` line is
typing, the four golden diffs are attention. `atlas-docs` goes 0.05 → 0.10
because this paragraph carries one thing #215's did not: the entry's own
retirement route (PQ-2). `milestone-review` holds at the 0.18 ceiling the #208
and #215 precedents both settled on.

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
      (#215 BR-1). Include the entry's own funeral: a registry entry costs
      ~50 lines in each of four gate prompts, and `architecture-deferred.md`
      is where one goes when it stops earning that (PQ-2)
- [ ] Verify delivery from the binary: `sdlc arch-principles` renders it, and
      `sdlc start-plan` / the gate prompts carry it via the existing embedding test
- [ ] Deliver by rebuilding the shared binary: `make sdlc-build` in ariadne,
      then `sdlc arch-principles` from pair's cwd shows eight entries
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

### 2026-09-12 — plan-quality round 1: the delivery step named a mechanism that carries nothing

**PQ-1 (Important), addressed.** The Plan and the fourth Done-when both said the
entry reaches pair via `sdlc propagate-base`. Verified against the tree rather
than taken on report: `construct/base.manifest` has 52 rows and none of them is
`cmd/sdlc/internal/judge/architecture.md`. The registry is `go:embed`'ed
(`architecture.go:47`) into one binary, and `~/.local/bin/sdlc` is a symlink to
`/Users/xianxu/workspace/ariadne/bin/sdlc` — so every repo on this machine runs
ariadne's build. Delivery is `make sdlc-build` (`Makefile.workflow:796`,
build-in-owner), and propagate-base would have committed into every dependent for
no effect.

Worth recording alongside it, because it makes the manifest look more relevant
than it is: the manifest DOES carry `symlink atlas/workflow` (row 219), so the
atlas page reaches dependents — but as a live symlink into ariadne, which means
it too needs no propagation step. Both halves of this issue's output are shared
by reference, one through a symlink and one through a binary. Nothing here is
copied, so nothing here is propagated.

This is the same shape as the family #207's review kept naming: a step written
from a plausible memory of how delivery works instead of from the manifest that
decides it.

**PQ-2 (Minor), folded in rather than carried.** An entry whose subject is "name
what removes this" should say what removes a registry entry. Each one costs
roughly fifty lines in four gate prompts, and `architecture-deferred.md` already
exists as the retirement route — one sentence in the atlas map paragraph, now a
Plan step.
