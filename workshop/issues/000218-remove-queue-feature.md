---
id: 000218
status: working
deps: []
github_issue:
created: 2026-09-09
updated: 2026-09-09
estimate_hours:
started: 2026-09-09T10:48:12-07:00
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

## Spec

Remove the verb, its pure package, its helptext, its datatype prototype, and the
seeded trunk file. `gitx.TrunkFile` stays, re-anchored on ariadne#207.

**Delete:**

- `cmd/sdlc/queue.go`, `queue_test.go`, `queue_e2e_test.go`
- `cmd/sdlc/internal/queue/` (line/doc/intent + tests + fuzz corpus)
- `cmd/sdlc/helptext/queue.md`
- `construct/datatype/queue.md`
- `workshop/queue.md` **on origin/main** — the file the verb seeded, six commits
  of it. Deleting it is a normal commit; nothing needs the verb to remove it.

**Edit:**

- `cmd/sdlc/main.go` — drop the `add(NewQueueCmd(), ...)` registration.
- `cmd/sdlc/repoguard.go` — drop the `queue` clause from the spine-guard comment,
  restoring "exactly the lifecycle verbs" as a true claim.
- `atlas/workflow/sdlc-binary.md` — remove the queue verb section and its verb
  table row.
- `atlas/index.md` — remove the queue mention.

**Keep:**

- `gitx.FirstLine` — `issueids.go` consumes it (it had two definitions before
  #209 homed it here); unrelated to the queue.
- The `lessons.md` entry on test doubles that cannot produce the behavior their
  tests are named for. It is a general lesson, earned twice in #209, and no gate
  catches it.

**Regenerate, do not hand-edit:** `construct/generated/datatype/SKILL.md` drops
`queue` from its type list via `datatype --output construct/generated/datatype`.
It is gitignored, so verify against `cmd/datatype/` and `construct/local/datatype/`
staying clean rather than against a diff of the generated file.

## `gitx.TrunkFile` stays — decided

Removing the verb leaves it with zero consumers *today*, but ariadne#207 is next
and consumes it directly, so it is not orphaned for long.

The principle worth recording, because it is the exact inverse of the queue's:
**an issue id is a shared namespace with concurrent writers.** Two agents on two
branches both allocating one id collide, and ariadne#188 documents why repair is
expensive rather than merely annoying — the id leaks into the branch name, commit
subjects agents grep, `deps:` in sibling issues, and review sidecar filenames. A
namespace with multiple writers has exactly one correct home, and it is the
trunk. That is a correctness requirement.

The queue had no shared namespace. One operator's attention list, nothing to
collide, nothing expensive to repair — its "consistent across checkouts"
requirement was manufactured by the storage choice rather than demanded by the
content. Same mechanism, opposite justification: `TrunkFile` was
over-engineering for the queue and is load-bearing for issues.

Consequence for this issue: `TrunkFile` and its tests are untouched. Its atlas
section is **re-anchored**, not deleted — it currently opens "Built for
`sdlc queue`", which stops being true here.

## Plan

- [x] Resolve the `gitx.TrunkFile` open decision with the operator — it stays;
      ariadne#207 consumes it next (see above).
- [ ] Delete the verb, the pure package, the helptext, the datatype prototype.
- [ ] Edit the four reference sites (`main.go`, `repoguard.go`, and the two atlas
      files); confirm `sdlc --help` no longer lists `queue`.
- [ ] Regenerate the datatype SKILL; confirm `queue` is gone from `datatype list`
      with no hand-edit to `cmd/datatype/` or `construct/local/datatype/`.
- [ ] Remove `workshop/queue.md` from origin/main.
- [ ] Re-anchor TrunkFile's atlas section on ariadne#207 — it currently opens
      "Built for `sdlc queue`", which this change falsifies.
- [ ] Verify: `go test ./cmd/sdlc/...` green except the pre-existing ariadne#210
      failure; `sdlc queue` is gone; `datatype list` has no `queue`.

## Log` to
  point at the seam. That is a written consumer, not speculative generality.
- **Remove:** dead code in a base-layer binary propagates to every ariadne-styled
  repo. Git history makes it recoverable in minutes if #207 wants it, and #207
  would then build it against its own needs rather than the queue's.

Awaiting the operator. If removal is chosen, also amend #207's `## Log` so it
does not point at a seam that no longer exists — while keeping the correction it
records, that a content-preserving retry re-lands a colliding id and its
Done-when needs amending.

## Plan

- [ ]

## Log

### 2026-09-09

### 2026-09-09

Withdrawn after a design conversation that started from the operator's second
thoughts on `workshop/queue.md` and ended somewhere more useful. The reasoning is
captured in `## Problem` because it generalizes past this feature; the succeeding
direction is a per-context scratchpad in pair's draft nvim pane, with an
inspection-scoped **punch list** in brain as the durable artifact — terminating
rather than eternal, which is what dissolves the rot problem #209 had to defer a
close-gate sweep to fight.
