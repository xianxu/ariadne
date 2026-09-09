---
id: 000218
status: working
deps: []
github_issue:
created: 2026-09-09
updated: 2026-09-09
estimate_hours: 1.55
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

Consequence for this issue: `TrunkFile`'s **behavior and test coverage** are
untouched. That is the precise claim — an earlier draft said "unchanged", which
forbade fixing the very references this removal falsifies.

**The rule:** behavior and coverage unchanged; names and prose that cite a
deleted feature get re-anchored. A dead concept left in identifiers and comments
is a trap for the next reader, and it is the class this issue is about.

**The enumeration**, swept in one round rather than discovered later:

| Site | Disposition |
|---|---|
| `atlas/workflow/sdlc-binary.md` TrunkFile section | Re-anchor on ariadne#207 — it opens "Built for `sdlc queue`". |
| `gitx/trunkfile.go` mode-preservation comment | Re-anchor. "Fine for a queue, wrong for a general primitive" loses its referent; the reason survives. |
| `gitx/trunkfile_test.go` fixture path `queue.md` (~20 sites) | Rename to a neutral fixture name. The filename is arbitrary to the test; leaving it names a feature that no longer exists. |
| `gitx/trunkfile_test.go:15,:83` comments | Re-anchor to neutral phrasing. |
| `workshop/lessons.md` test-double entry | KEEP, plus one clause noting the cited `fakeTrunk` / `TestQueueEdit_*` were removed here. A lesson legitimately cites history, but a reader must not hunt for code that is gone. |
| `workshop/issues/000207-...md:175` | Amend. It reasons from "#209's queue passes an intent-replaying transform" — an open issue citing a deleted consumer. The point (retry semantics belong to the caller's transform) survives the citation. |

**Coverage loss from the deleted tests: none.** A green suite cannot show this,
because a deleted test never fails, so it is asserted here instead.
`TestQueueCmd_WriteVerbsAreSpineGuarded` and `subcommandGuardSource` are the
tree's only source-inspecting guard-ordering tests, and they covered the queue
subcommands alone; the surviving spine verbs are covered by `repoguard_test.go`
off `processmanual.WorkflowVerbs()`, which never included queue.

## Done when

- `sdlc queue` no longer exists: absent from `sdlc --help`, and
  `cmd/sdlc/queue.go`, `queue_test.go`, `queue_e2e_test.go`,
  `cmd/sdlc/internal/queue/` and `cmd/sdlc/helptext/queue.md` are gone.
- `construct/datatype/queue.md` is gone and `datatype list` no longer offers
  `queue` — verified after regeneration, with `cmd/datatype/` and
  `construct/local/datatype/` clean. NOT verified by diffing
  `construct/generated/`, which is gitignored and would pass vacuously.
- `repoguard.go`'s spine-guard comment no longer claims a `queue` clause, so
  "exactly the lifecycle verbs" reads true again.
- `atlas/workflow/sdlc-binary.md` has no queue verb section and no queue row in
  the verb table; `atlas/index.md` no longer mentions it.
- `workshop/queue.md` is gone from `origin/main`.
- `gitx.TrunkFile`'s behavior and coverage are unchanged: no test is deleted from
  `trunkfile_test.go` and `go test ./cmd/sdlc/internal/gitx/` passes. (No test
  count is asserted — the previous draft claimed "37", which matches no countable
  set and could not be checked.)
- Every site in the re-anchor enumeration above is dispositioned, including the
  two outside this repo's control-flow: `lessons.md` and the OPEN ariadne#207.
- `gitx.FirstLine` still exists and `issueids.go` still consumes it.
- The `lessons.md` entry on test doubles survives — it is a general lesson, not
  queue-specific.
- **Producible anywhere** (these are the close-gate evidence):
  - `go build ./...` clean.
  - `go test ./cmd/sdlc/internal/... -count=1` green — every package this change
    can affect, and none of them take the repo lock.
  - `sdlc --help` lists no `queue`; `datatype list` offers no `queue`.
  - The identifier sweep returns **nothing**:

    ```
    grep -rn "sdlc queue\|internal/queue\|NewQueueCmd\|datatype/queue\|helptext/queue" \
      --include='*.go' --include='*.md' . \
      | grep -v "^workshop/history/" \
      | grep -v "^workshop/issues/00021[89]" \
      | grep -v "^construct/generated/"
    ```

    A bare `grep -rn queue` is NOT the check — measured at 758 raw hits, ~350
    after excluding deleted paths, of which ~330 are `workshop/history` archives
    and the rest are unrelated BFS-variable senses (`bootstrap.sh`,
    `list-peers.sh`, `projectstatus.go`, `layergraph/walk.go`,
    `fleetpolicy_test.go`). A clause whose command cannot produce a checkable
    result is not a verification. The three exclusions are principled, not
    convenient: history legitimately describes a feature that existed, this
    issue's own files describe the removal, and `construct/generated/` is
    gitignored and regenerated.

    **This sweep is necessary, not sufficient.** It catches *identifier*
    references only. Prose references — `lessons.md` citing deleted tests,
    ariadne#207 citing "#209's queue" — match none of those tokens, and the
    re-anchor table is the authority for them.
- **From a plain shell, outside an sdlc transaction:**
  `go test ./cmd/sdlc/ -count=1` green except the pre-existing ariadne#210
  failure (`fleet_plan_test.go:14`, a hardcoded path to an archived plan).

  The qualifier is not a hedge, it is the diagnosis. The plan-quality reviewer
  saw this command blow `go test`'s 600s default with a goroutine parked in
  syscall for nine minutes, while it passes here in 115-136s. The cause is
  neither slowness nor a hung child: `close_test.go:131` executes a real
  `NewCloseCmd()`, which takes the production `.git/sdlc.lock`, so a suite run
  launched from INSIDE a transaction — which is exactly where the gates run it —
  waits `DefaultWaitTimeout`, 30 minutes. `-timeout 1800s` cannot rescue it,
  being the same 30 minutes.

  That is **ariadne#219**, filed from this review. It is pre-existing, unrelated
  to the queue, and this change only deletes tests, so it strictly reduces
  runtime. #218 therefore does not carry a full-suite run as gate evidence — the
  producible list above stands in its place until #219 lands.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec              design=0.8  impl=0.08
item: cross-cutting-refactor  design=0.1  impl=0.14
item: atlas-docs              design=0.05 impl=0.06
item: milestone-review        design=0.0  impl=0.18
design-buffer: 0.15
total: 1.55
```

Σdesign 0.95 × 1.15 + Σimpl 0.46 × 1.0 = 1.55.

`issue-spec` is priced undiscounted and against the window `sdlc actual`
computes — open at `6060f035`, already reading 1.34h and attributed across #218
and #219. It covers the Problem section's four-part diagnosis, three
plan-quality rounds, and authoring ariadne#219 out of the third round's
diagnosis. The ×0.2 discount does not apply to the primitive whose deliverable
*is* the spec (the ariadne#215 lesson).

`cross-cutting-refactor` is the deletion plus the six re-anchor sites — the
right slug because the work is a multi-file sweep rather than a module, and its
design carries the ×0.2 discount since the enumeration table pre-resolves every
disposition. `atlas-docs` is the two atlas edits. One `milestone-review` row,
priced near the primitive ceiling: single-pass with no `Mx`, but ariadne#209
closed at eight review rounds on adjacent code, and a removal touching the base
layer is not obviously cheaper to review than an addition.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [x] Resolve the `gitx.TrunkFile` open decision with the operator — it stays;
      ariadne#207 consumes it next (see above).
- [ ] Delete the verb, the pure package, the helptext, the datatype prototype.
- [ ] Edit the four reference sites (`main.go`, `repoguard.go`, and the two atlas
      files); confirm `sdlc --help` no longer lists `queue`.
- [ ] Regenerate the datatype SKILL; confirm `queue` is gone from `datatype list`
      with no hand-edit to `cmd/datatype/` or `construct/local/datatype/`.
- [ ] Remove `workshop/queue.md` from origin/main.
- [ ] Sweep the re-anchor enumeration in the Spec — all six sites, including
      `lessons.md` and the open ariadne#207 — not just the atlas one.
- [ ] Verify the producible-anywhere list (build, `./cmd/sdlc/internal/...`,
      `sdlc --help`, `datatype list`, the grep sweep), plus a plain-shell
      `go test ./cmd/sdlc/` outside any transaction. Cause of the
      inside-transaction failure is ariadne#219, not this change.

## Log

### 2026-09-09

Withdrawn after a design conversation that started from the operator's second
thoughts on `workshop/queue.md` and ended somewhere more useful. The reasoning is
captured in `## Problem` because it generalizes past this feature; the succeeding
direction is a per-context scratchpad in pair's draft nvim pane, with an
inspection-scoped **punch list** in brain as the durable artifact — terminating
rather than eternal, which is what dissolves the rot problem #209 had to defer a
close-gate sweep to fight.
