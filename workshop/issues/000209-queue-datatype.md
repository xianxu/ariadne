---
id: 000209
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-07
estimate_hours: 2.75
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
| 3 rejections in a row | Refuse, surface the last rejection. |

There is deliberately **no row for a hand-edit of the working-tree copy**. An
earlier draft promised "remote wins; report the drop", which is undesignable:
nothing in this architecture ever reads the working-tree copy, so there is no
point at which a drop is detectable. The honest statement is that such an edit is
simply never read — and if the operator commits and pushes it through git in the
normal way, it *is* the trunk and there is no conflict at all. The datatype prose
carries the warning; the verb makes no promise it cannot keep.

**Reading is subject to the same offline policy the repo already settled.** Every
operation begins with a fetch. A failed fetch is **not** a refusal: degrade to the
stale `origin/main` tracking ref and announce it loudly, exactly as
`issueids.go:40-49,126-145` does for id allocation (`ARCH-DRY` — one policy, not
two). A repo with no `origin` remote at all reads the local ref and says so. A
write, by contrast, cannot degrade: with no reachable origin the CAS push has
nothing to compare against, so `add`/`remove`/`move` refuse and say why.

**Input validation, before any git call.** A malformed ref is rejected. So is a
why-now containing a newline or control character: a newline would silently become
a second queue entry and break `Doc`'s round-trip invariant. An embedded ` — ` or
`[...]` re-parses as a different why/tag split, so those are rejected too rather
than escaped — the format's value is that a human reads it, and escaping would
cost that.

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
  hand-edit** — regeneration is the only step. Verified by `grep queue` on the
  generated file plus a clean `git diff` of `cmd/datatype/SKILL.md.tmpl` and
  `construct/local/datatype/`. NOT by diffing `construct/generated/`, which is
  gitignored (`.gitignore:30`) and would pass vacuously.

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
- A failed fetch degrades a **read** to the stale tracking ref with a loud
  warning, and refuses a **write**; a repo with no `origin` says so rather than
  erroring obscurely.
- `Doc` round-trips arbitrary bytes, proven by a fuzz target rather than by
  chosen examples — it parses a human-editable file arriving from the trunk,
  which is input this process did not produce (`ARCH-SECURE`).
- A why-now carrying a newline, a control character, ` — `, or `[...]` is
  rejected before any git call.
- The published blob round-trips: checking out the pushed commit yields a file
  byte-identical to the intended content, with `.gitattributes` applied.
- The temp index is removed on every exit path, including the error paths, and is
  never `$GIT_DIR/index`.

**Seed**

- `ariadne/workshop/queue.md` carries the real current ordering, written through
  the verb rather than by hand — the cheapest test of whether the shape is right.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec             design=0.8  impl=0.12
item: greenfield-go-module   design=0.3  impl=0.24
item: greenfield-go-module   design=0.2  impl=0.2
item: smaller-go-module      design=0.05 impl=0.14
item: atlas-docs             design=0.05 impl=0.08
item: milestone-review       design=0.0  impl=0.18
item: milestone-review       design=0.0  impl=0.18
design-buffer: 0.15
total: 2.75
```

Σdesign 1.4 × 1.15 + Σimpl 1.14 × 1.0 = 2.75.

`issue-spec` is priced **undiscounted** and against the window `sdlc actual`
computes (open at `e6d6629a`, already reading 0.90h): a four-exchange brainstorm
that twice corrected the design, a full Spec rewrite widening the issue's scope,
a 256-line durable plan, and one plan-quality round that landed six findings. The
×0.2 spec-quality discount does not apply to the primitive whose deliverable *is*
the spec — the lesson recorded on ariadne#215.

The two `greenfield-go-module` rows are the new packages: `gitx.TrunkFile` (git
plumbing, CAS retry, offline policy, signing/attributes round-trip — `impl` above
mid because concurrency tests driving two publishers against a real bare origin
are fiddly to get deterministic) and `internal/queue` (`Line`/`Doc`/`Intent` plus
a fuzz target). `smaller-go-module` is the verb, which mirrors the established
cobra + helptext pattern. `atlas-docs` covers `construct/datatype/queue.md` and
the seed. Design on all four carries the ×0.2 discount — the plan gives the exact
plumbing sequence, the exact test bodies, and file:line for every seam.

**Two** `milestone-review` rows, not one: M1 and M2 each close separately, so each
buys its own boundary review. Priced near the primitive ceiling, per ariadne#208
(two rows at ceiling, still closed 1.72 against 1.13) and ariadne#215 (one row at
0.1 against a review that then raised BR-1 and needed a second round).

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000209-queue-datatype-plan.md` (authored via
`superpowers-writing-plans`) — 12 tasks with TDD steps, exact paths, and the
per-task commands. This section carries the two review boundaries only.

- [ ] M1 — `gitx.TrunkFile`: read + CAS-write one path on the trunk with no
      working tree, bounded retry that **re-runs the transform** on the moved
      base, gitattributes/signing round-trip, offline degrade-read/refuse-write, temp-index hygiene. Tasks 1-6.
- [ ] M2 — the queue: `Line`/`Doc`/`Intent` (pure, no git in their tests), the
      `sdlc queue` verb, refusal-as-handoff, `construct/datatype/queue.md`, and
      ariadne's queue seeded **through the verb**. Tasks 7-13.

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

### 2026-09-07 — plan-quality round 1: six findings, all taken

One Critical, three Important, two Minor. Every one verified against the code
before acting; all six were correct.

**PQ-1 (Critical) — the seam the plan named cannot exist.** The plan had
`gitx.NewTrunkFile(execGitRunner{}, ...)`, but `execGitRunner` is package `main`
(`runner.go:31`) and `main` already imports `gitx` (`actual.go:26`) — an import
cycle that would not compile. Separately, the plumbing needs `GIT_INDEX_FILE` on
`update-index` and `write-tree`, and neither `gitRunner.Git/GitInDir`
(`runner.go:36,40`) nor `gitx.run` (`window.go:32`) carries env; only `read-tree`
has an `--index-output` escape, so env is unavoidable.

Resolved *without* widening `gitRunner` — that interface is used across ~20 files
and this needs nothing from those callers. `TrunkFile` lives in `gitx` and uses
`gitx`'s own package-level `run` shim, plus a sibling `runEnv` shim for the three
index calls. `gitx` is a leaf package, so there is no cycle; the existing shim
pattern is reused rather than a second one invented (`ARCH-DRY`); and tests drive
real git against a bare origin, which `ARCH-MOCK` requires here anyway. Swept the
class: the verb consumes `TrunkFile` through an interface **declared in `main`**
(consumer-side, Go idiom), not one exported from `gitx`.

**PQ-3 (Important) — a Done-when clause promising something undesignable.** The
interleaving table's "wholesale hand-edit / remote wins, report the drop" row had
no task and no possible implementation site, because nothing here reads the
working-tree copy. Row deleted and the reasoning recorded in the Spec, rather than
inventing a worktree-comparison step to satisfy a row that should not have been
written.

**PQ-2 (Important) — no operating envelope on the per-invocation fetch.** Adopted
`issueids.go`'s settled offline policy verbatim in shape: reads degrade loudly to
the stale ref, writes refuse. Divergence would have meant two offline policies in
one binary.

**PQ-4 (Important)** — `Doc.Parse/Render` guarded by a fuzz target, not four
handcrafted inputs. **PQ-5 (Minor)** — the generated-artifact check pointed at
`construct/generated/`, which is gitignored and would pass vacuously; repointed at
`cmd/datatype/SKILL.md.tmpl` and `construct/local/datatype/`. **PQ-6 (Minor)** —
why-now is free text written straight into the line format; now validated before
any git call.
