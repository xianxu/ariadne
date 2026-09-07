---
id: 000209
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-07
estimate_hours:
started: 2026-09-07T16:11:22-07:00
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

Three deliverables: the datatype prototype, a `sdlc queue` verb backed by
`origin/main`, and ariadne's seeded queue.

The scope grew from "prototype only, no verb" — see `## Revisions`. The reason
is short: *always consistent across checkouts* cannot be delivered by prose. An
agent on a feature branch that runs `cat workshop/queue.md` reads that branch's
copy. Something has to fetch, and that something is a verb.

### Semantics: advisory, never binding

The queue indicates a section of work that will likely happen — it is not a
commitment. Execution still goes through the hard blockers: `deps:` is
authoritative and the queue never overrides it. If a line is really a block, it
belongs in `deps`, or there are two truths about blocking that will disagree.

### Storage: origin/main is the permanent base

`workshop/queue.md`, tracked normally on `main`. Tracked because the history is
the point — `git log workshop/queue.md` records how priority actually moved, and
the file shows up in PRs and on GitHub.

**Every operation is `fetch -> apply intent -> push`.** On a non-fast-forward
rejection: the same three steps again, bounded at 3 attempts. `origin/main` wins
at every step, which is what makes the rule uniform — there is no working copy
competing with the remote, so nothing merges and no three-way arises.

Accepted cost, stated rather than engineered around: the file *is* in the working
tree on a branch, and that copy can go stale. The datatype prose says the queue is
read through `sdlc queue`, and the verb always fetches. A stale `cat` of an
advisory file is a smaller problem than a second storage mechanism.

### Operations are intents, which is what makes replay total

```
sdlc queue                          # list, from origin/main
sdlc queue add <ref> "<why-now>"    # append
sdlc queue remove <ref>             # drop the line
sdlc queue move <ref> --before <ref> | --after <ref>
```

All four carry an *intent*, not file content. Replaying "add this line" or "move
X before Y" onto a base that moved is unambiguous, and the peer's concurrent edit
survives. `move` looked unmergeable only under the assumption that reordering
means handing over a whole rewritten file; expressed as an intent it is exactly as
replayable as `add`, so there is one rule and no exception (`ARCH-ORDER`).

### The interleaving space, enumerated

`origin/main` is durable state and peer pushes are events this process cannot
block, so the cells are written down rather than left to emerge:

| Event during `fetch -> apply -> push` | Policy |
|---|---|
| Peer added unrelated lines | Replay. Both survive. Fast-forward. |
| `remove X`, peer already removed `X` | Converge. No-op, succeed. |
| `add X`, peer already added `X` | Converge to one line; newer why-now wins; report which was kept. |
| `move X --before Y`, peer deleted `Y` | **Refuse.** The anchor is gone; do not guess a position. |
| Wholesale hand-edit of the file | No intent exists. Remote wins; report the drop. |
| 3 rejections in a row | Refuse, surface the last rejection. |

**A refusal is a handoff, not a dead end.** `sdlc`'s errors are next-action specs,
so a failed replay prints the current remote queue *and* the intent that could not
be applied — enough context for the operator or an agent to re-derive the
intended edit and re-run. That is why the unresolvable cells do not need
resolution logic: they need an informative failure.

Nondeterminism enters at exactly one place — the order in which peers' pushes
reach `origin/main` — and a failing ordering is reproduced by the two-publisher
test against a local bare origin (below), not by timing.

### Line format

One ref per line plus a few words of why-now, with an optional project tag:

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

### Relationship to siblings, to be stated in the datatype

- `deps:` — hard blocking. The queue is soft preference among the *unblocked*.
- project `status` (`paused`/`committed`/`executing`) and `roadmap` — these are
  where *project-level* prioritization already lives. The queue does not replace
  them.
- a project's `## Breakdown` — authoritative for ordering within that project.
  **The queue is for loose ends**, or the two orderings will disagree.

### Reuse: this builds the primitive ariadne#207 needs

The out-of-tree write path — `fetch`, `read-tree` into a temp index,
`hash-object -w --path`, `update-index`, `write-tree`, `commit-tree`, `push
<commit>:main` as a compare-and-swap — is precisely what #207 specs for
publishing issue files with no worktree on main. Build it once here behind the
existing `gitRunner` seam so #207 consumes it rather than writing a second one
(`ARCH-DRY`).

It also forecloses a defect in #207 as currently written: #207's retry re-pushes
"the current content of this one file", which for a colliding issue id lands the
duplicate as a clean fast-forward — the exact hole ariadne#188 documents. An
intent-replaying retry cannot express that bug. Note it on #207 rather than
fixing #207's prose here.

Plumbing details that bite if unnamed, inherited from #207's analysis:

- `hash-object -w --path <relpath>`, not bare `hash-object`, so `.gitattributes`
  filters and EOL normalization for that path apply. A blob written without them
  produces a commit whose checkout differs from the file.
- File mode `100644`.
- If the repo signs commits, `commit-tree` needs `-S`.
- The temp index must be a real temp file removed on every exit path, and must
  never be `$GIT_DIR/index` — writing that corrupts whatever checkout shares the
  git dir.

### Not in scope

- No close-gate removal of queue lines, and no cross-repo queue (both still
  deferred below, unchanged).
- Not changing `syncInPlace` or deleting `syncViaMainWorktree` — that is #207's
  job, and it lands after this provides the helper.

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

**Datatype**

- `construct/datatype/queue.md` exists and follows the sibling prototypes' shape
  (`type: type`, `name: queue`, a discovery `description`, then narrative prose).
- Its description triggers on the natural phrasings — "queue", "what's next",
  "pick the next thing", "plan the sequence".
- The prose states: advisory-not-binding and its relationship to `deps`; the
  division of labour with project `status`/`roadmap` and with a project's
  `## Breakdown`; the line format; the issue-line vs project-line distinction
  with its failure mode; and that the file is read through `sdlc queue` because
  a working-tree copy can be stale.
- The prose names the rot risk and points at the deferred close-gate removal, so
  a later reader finds a recorded decision rather than an apparent gap.
- `construct/generated/datatype/SKILL.md` picks `queue` up **without a
  hand-edit** — regeneration is the only step. A manual edit anywhere is a bug in
  the single-sourcing, not a task here.

**Verb**

- `sdlc queue` lists from `origin/main` and is correct from a feature branch, a
  detached HEAD, and a checkout with **no worktree on main anywhere**.
- `add` / `remove` / `move` publish to `origin/main` with no worktree involved and
  without reading or writing any other checkout — including while another
  worktree on main is dirty or mid-rebase.
- Every row of the interleaving table above has a test. In particular a
  **two-publisher test against a local bare origin** (`ARCH-MOCK`: git is the
  external binary, a temp repo is its portable stateful fake) asserts that a
  concurrent `add` replays and **both** lines land — a function-call mock cannot
  produce a real non-fast-forward rejection.
- `move --before Y` with `Y` deleted concurrently refuses, and the error prints
  the current remote queue plus the unapplied intent.
- Bounded at 3 attempts; the last rejection is surfaced, not swallowed.
- The published blob round-trips: checking out the pushed commit yields a file
  byte-identical to the intended content, with `.gitattributes` applied.
- The temp index is removed on every exit path, including the error paths, and is
  never `$GIT_DIR/index`.

**Seed**

- `ariadne/workshop/queue.md` carries the real current ordering, written through
  the verb rather than by hand — the cheapest test of whether the shape is right.

## Plan

Durable plan: `workshop/plans/000209-queue-datatype-plan.md` (authored via
`superpowers-writing-plans`) — 12 tasks with TDD steps, exact paths, and the
per-task commands. This section carries the two review boundaries only.

- [ ] M1 — `gitx.TrunkFile`: read + CAS-write one path on the trunk with no
      working tree, bounded retry that **re-runs the transform** on the moved
      base, gitattributes/signing round-trip, temp-index hygiene. Tasks 1-5.
- [ ] M2 — the queue: `Line`/`Doc`/`Intent` (pure, no git in their tests), the
      `sdlc queue` verb, refusal-as-handoff, `construct/datatype/queue.md`, and
      ariadne's queue seeded **through the verb**. Tasks 6-12.

Two milestones because M1 is a reusable primitive that stands alone — it is what
`ariadne#207` will consume — and M2 is the feature built on it. Each is worth its
own boundary review; neither is a one-shot that a plain checkbox would cover.

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

## Revisions

### 2026-09-07 — scope widened from prototype-only to prototype + verb

Reason: operator requirement — "this queue file should be operating against
remote main directly, so that it's always consistent across different checkouts."

The original Spec stated a non-goal: "Scope is the datatype prototype and nothing
else — no verb, no gate, no tooling", recorded in the Log as the operator's own
scoping ("it's just a datatype; how it's being used for now can be freeform").
That non-goal and the new requirement are incompatible, and the requirement wins:
consistency across checkouts is not expressible in prose, because an agent on a
feature branch reading `workshop/queue.md` reads that branch's copy. The Spec
above is rewritten accordingly rather than left standing with a changelog beside
it — one current version, per the ARCH-DRY lesson from ariadne#215 BR-1/PQ-1.

Still deferred, unchanged: close-gate removal of queue lines, and cross-repo
ordering.

**Design points settled during the brainstorm, with two corrections worth
keeping:**

- *Withdrawn:* that "remote wins" is total for `add`/`remove` but **undefined**
  for reorder. It is not undefined — remote winning gives a perfectly
  deterministic answer (the reorder is discarded). What is actually at stake is
  whether the operator's *intent* survives, which was wrongly dressed up as an
  ambiguity in the rule. The two policies proposed on that basis — "replay" and
  "refuse and re-run" — were also the same operation in different words.
- *Consequence, and the design's core move:* if `add`/`remove` survive a moved
  base because they are **intents** rather than file content, then reorder can be
  too. `move X --before Y` replays onto a changed base and preserves the peer's
  edit. Reorder looked special only because of an unexamined assumption that it
  meant handing over a whole rewritten file. So: one rule everywhere, no
  exception, and no need for the rejected alternative of a per-line rank key
  (which would have put a machine field into a format whose appeal is that a
  human reads it as a list).
- *Operator, on the unresolvable cells:* sdlc errors out, the operator knows it
  failed, and an agent can resolve the difference from the new remote state.
  That is why the residual cells need an **informative failure** rather than
  resolution logic — consistent with `sdlc --help`'s "its errors are next-action
  specs."
