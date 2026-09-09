---
id: 000218
status: codecomplete
deps: []
github_issue:
created: 2026-09-09
updated: 2026-09-09
estimate_hours: 2.75
started: 2026-09-09T10:48:12-07:00
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

**The enumeration**, swept in one round rather than discovered later. It is
**hand-typed and therefore not authoritative** — the diff is. Two sites in
`trunkfile_test.go` were re-anchored in the code before this table named them,
which is exactly how a hand-maintained list drifts from what was done. Read it as
the plan, and `git show` as the record (the "what a guard COMPUTES vs what it
takes on faith" lesson, applied to a table instead of a scan):

| Site | Disposition |
|---|---|
| `atlas/workflow/sdlc-binary.md` TrunkFile section | Re-anchor on ariadne#207 — it opens "Built for `sdlc queue`". |
| `gitx/trunkfile.go` mode-preservation comment | Re-anchor. "Fine for a queue, wrong for a general primitive" loses its referent; the reason survives. |
| `gitx/trunkfile_test.go` fixture path + commit subjects | Rename `queue.md` -> `note.md` and `queue: …` -> `trunk: …`. The names are arbitrary to the tests; leaving them names a feature that no longer exists. No count is asserted — an earlier draft said "~20", then "36", and neither matched a countable set. |
| `gitx/trunkfile_test.go` comments (fixture header, missing-path rationale, `PreservesFileMode`) | Re-anchor to neutral phrasing. |
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
- `workshop/queue.md` is **deleted on the branch** — `git ls-tree HEAD` finds it
  absent, and the branch is a clean descendant of main with no competing commits
  on that path, so the merge removes it from `origin/main`. The earlier wording
  claimed it was already gone from the trunk, which is not satisfiable at this
  boundary: the file is still at `a722b52` until #218 merges. Verify post-merge.
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
  - The identifier sweep returns **exactly one line** — the deliberate
    historical note in the TrunkFile atlas section
    (`atlas/workflow/sdlc-binary.md`, "It was built for `sdlc queue`, removed in
    #218"), which explains why the primitive outlived its first consumer:

    ```
    git grep -n -E 'sdlc queue|internal/queue|NewQueueCmd|datatype/queue|helptext/queue' \
      -- '*.go' '*.md' \
      ':!workshop/history' ':!*000218*' ':!*000219*'
    ```

    **`git grep`, not `grep -rn`, and that is the point.** The first draft of
    this clause used `grep -rn … .` and claimed one line, because it was measured
    in a shell where `grep` is aliased to `ugrep`, which emits repo-relative
    paths. Under a plain shell `grep -rn … .` prefixes every path with `./`, so
    the `^workshop/…` exclusions match nothing and the same command returns 103
    lines. A verification that depends on the operator's aliases is not a
    verification — the close reviewer ran it as written and got 99. `git grep`
    is deterministic: repo-relative paths everywhere, and gitignore-aware, so
    `construct/generated/` needs no exclusion at all.

    **The exclusions are by issue ID, not by directory, and that matters.** An
    earlier draft excluded `workshop/issues/000218*` and missed
    `workshop/plans/000218-*-close-review.md` — which the close *generates*, so it
    cannot exist when the command is tested beforehand. A check whose own gate
    produces the artifacts it must tolerate can only be validated after the fact;
    scoping by id rather than by path makes it stable across that.

    A bare `grep -rn queue` is NOT the check — measured at 758 raw hits, ~350
    after excluding deleted paths, of which ~330 are `workshop/history` archives
    and the rest are unrelated BFS-variable senses (`bootstrap.sh`,
    `list-peers.sh`, `projectstatus.go`, `layergraph/walk.go`,
    `fleetpolicy_test.go`). A clause whose command cannot produce a checkable
    result is not a verification. The exclusions are principled, not convenient:
    history legitimately describes a feature that existed, and this issue's own
    files describe the removal.

    Stating the expected residue rather than "nothing" is the same rule the
    round-3 advisory named: a clause whose command has known-benign hits must
    carry the filter *or* the residue. Rephrasing the atlas prose to make the
    sweep empty would have been contorting the artifact to fit the test.

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
item: issue-spec              design=1.2  impl=0.08
item: issue-spec              design=0.5  impl=0.06
item: cross-cutting-refactor  design=0.1  impl=0.24
item: atlas-docs              design=0.05 impl=0.06
item: milestone-review        design=0.0  impl=0.18
design-buffer: 0.15
total: 2.75
```

Σdesign 1.85 × 1.15 + Σimpl 0.62 × 1.0 = 2.75. Revised up from 1.55 after
estimate-quality; the first derivation was under-priced in four specific ways and
the correction is recorded rather than quietly applied.

**Two `issue-spec` rows, not one.** The first (1.2, near the 0.5-1.5 band top) is
this issue: the Problem section's four-part diagnosis, the disposition table, the
producible-evidence list, and **four** plan-quality rounds — not three, as an
earlier draft said. The second (0.5, band floor) is authoring ariadne#219, which
is its own spec deliverable and was previously folded into the first row for
free. The judge's precedent is exact and worth recording: ariadne#209 used the
identical `issue-spec design=0.8 impl=0.12` when its window read 0.90h and closed
4.49 against 3.06. This window read 1.30h at the same point and the row was
unchanged — the same under-pricing from a worse starting position.

**`cross-cutting-refactor impl` 0.14 -> 0.24.** The first value left verification
wall-clock invisible, and under v3.1 elapsed time *is* the unit: `go build`, the
internal-package suite, a plain-shell `go test ./cmd/sdlc/` measured at 115-136s,
the datatype regeneration, and the sweep are 0.08-0.10h on a first pass and
double if the close review forces a re-run. That has to coexist with deleting
~1,000 lines across seven files, plus six re-anchor sites — including **36**
`queue` occurrences in `trunkfile_test.go` (an earlier draft said "~20"), an edit
to the OPEN ariadne#207, and a `lessons.md` amendment.

**`design-buffer` stays 0.15.** v2.1 offers +30% absent a thorough plan doc, and
`workshop/plans/` holds no durable `000218-*-plan.md` — only the gate sidecar. The
+15% is claimed on the other criterion: the Spec's disposition table and
producible-evidence list are the plan, and for a deletion whose scope is an
enumeration, a separate plan document would restate them rather than add anything.
Worth naming as a judgement call — at +30% the total is 3.03.

**Known tension, for the ledger.** The buffer exists to cover design not yet done,
and here design is complete — four gate rounds closed it. Applying any buffer to
finished design is conservative by construction, so this number should be read as
an upper bound rather than a centre. Recording it because a ratio near 1.0 on this
row would be a coincidence of two errors, not accuracy.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [x] Resolve the `gitx.TrunkFile` open decision with the operator — it stays;
      ariadne#207 consumes it next (see above).
- [x] Delete the verb, the pure package, the helptext, the datatype prototype.
- [x] Edit the four reference sites (`main.go`, `repoguard.go`, and the two atlas
      files); confirm `sdlc --help` no longer lists `queue`.
- [x] Regenerate the datatype SKILL; confirm `queue` is gone from `datatype list`
      with no hand-edit to `cmd/datatype/` or `construct/local/datatype/`.
- [x] Delete `workshop/queue.md` on the branch (it leaves origin/main at merge,
      not at this boundary — the same premature-claim defect BR-7 fixed in
      Done-when and left here).
- [x] Sweep the re-anchor enumeration in the Spec — all six sites, including
      `lessons.md` and the open ariadne#207 — not just the atlas one.
- [x] Verify the producible-anywhere list (build, `./cmd/sdlc/internal/...`,
      `sdlc --help`, `datatype list`, the grep sweep), plus a plain-shell
      `go test ./cmd/sdlc/` outside any transaction. Cause of the
      inside-transaction failure is ariadne#219, not this change.

## Log

### 2026-09-09
- 2026-09-09: closed — BR-10 and BR-11 fixed, and both were the same root cause the ledger had been naming for five rounds. BR-10: my fix for BR-7 committed the defect it was fixing — a hand-typed enumeration of premature claims, wrong on arrival (5 live sites against the 3 it named, 2 of them introduced by this windows own re-anchor edits). Derived it instead via git grep ariadne#207 plus a grep for origin/main claims, dispositioned every hit in a ## Revisions entry INCLUDING the two that are NOT premature (trunkfile.go:22 is a capability claim true now, :325 a fact about #207s written spec), so the residue is dispositioned rather than silently skipped. Four real sites fixed: atlas:649 present-tense "its consumer IS #207", trunkfile.go:452 and trunkfile_test.go:721 present-tense "#207 POINTS this at" (both introduced by the re-anchor edit — fixing stale prose is when premature prose gets written), and the ## Plan origin/main tick BR-7 corrected in Done-when and left in the Plan. BR-11: the durable rule now in lessons.md, and tallying both issues made it ONE rule rather than eleven — across ~60 findings in 31 families, ~35 are a description written where the referent was available: a count instead of counting, a comment list instead of the grep that derives it, gits prose instead of gits exit code, a message asserting a cause instead of reporting an observation, a test name instead of a double that can produce the property, a hand-typed table instead of the diff. Recorded with both corollaries: "I verified it" is a property of my context not the artifact (the conclusion travels, the checking does not — hence verbatim, clean shell, at the readers state, and scope exclusions by issue id where the gate generates what the check must tolerate), and re-anchoring stale prose is when premature prose gets written. Removal evidence re-verified at HEAD: go build clean, gitx green, documented sweep run verbatim under bash returns exactly its stated one line.; review verdict: FIX-THEN-SHIP

Withdrawn after a design conversation that started from the operator's second
thoughts on `workshop/queue.md` and ended somewhere more useful. The reasoning is
captured in `## Problem` because it generalizes past this feature; the succeeding
direction is a per-context scratchpad in pair's draft nvim pane, with an
inspection-scoped **punch list** in brain as the durable artifact — terminating
rather than eternal, which is what dissolves the rot problem #209 had to defer a
close-gate sweep to fight.

### 2026-09-09 — premature-claim sweep, derived rather than hand-typed

BR-10 found my fix for BR-7 had the defect it was fixing: I hand-typed an
enumeration of premature claims, and it was already wrong — 5 live sites against
the 3 it named, 2 of them introduced by this window's own re-anchor edits.

Derived instead of listed. The commands, and their residue at the time of writing:

```
git grep -n 'ariadne#207' -- '*.go' '*.md' ':!workshop/history' ':!*000218*' ':!*000219*'
grep -n 'origin/main' workshop/issues/000218-remove-queue-feature.md
```

Dispositioned:

- `atlas/workflow/sdlc-binary.md:649` — "Its consumer **is** ariadne#207" stated a
  present fact about an unlanded issue. Now: no consumer in the tree today, #207
  is specced to consume it.
- `cmd/sdlc/internal/gitx/trunkfile.go:452` and `trunkfile_test.go:721` — "since
  #207 **points** this at arbitrary paths". Both introduced by this window's
  re-anchor, which is the tell: I was fixing stale prose and wrote premature
  prose. Now "is specced to point".
- `## Plan` — the `origin/main` tick, the same defect BR-7 corrected in Done-when
  and left in the Plan.
- `trunkfile.go:22` ("lets #207 consume this") and `:325` ("the same bound #207
  specs") — NOT premature. The first is a capability claim about the seam, true
  now; the second is a fact about #207's written spec, checkable now. Recorded so
  the residue is dispositioned rather than silently skipped.

### 2026-09-09 — the derivation's blind spot, and a rule for prose deletion

Two more rounds, both instructive rather than merely tedious.

**BR-10, round 5: the derived sweep had a blind spot for self-reference.** The
command `git grep -n 'ariadne#207' … ':!*000218*'` cannot reach
`workshop/issues/000207-...md:178`, because that file refers to itself as *"this
issue"* and never spells its own number. Deriving beat hand-typing — it found
four sites the hand list missed — but a derivation is only as complete as the
token it keys on, and self-reference is invisible to a search for the name.
Recorded because "derive it" is the rule this session landed on, and this is its
first known failure mode: **when a derivation keys on an identifier, the artifact
that owns that identifier is the one place it will not match.**

**BR-12/BR-13: a prose deletion is verified by re-reading the enclosing sentence,
not the deleted span.** A diff hunk shows only what left, so both defects were
invisible in review of the change itself: the `#207` parenthetical had swallowed
the sentence that followed it, and removing the trailing item from
`atlas/index.md:13` orphaned the conjunction and left `migrate` dangling after
the list's final comma. Both are the root-cause shape again — I verified the edit
I made rather than the result it produced.

**BR-9: a hand-restated count in a line I never touched.** `repoguard.go:23`
said "not 7 new per-verb flags" while the guard covers eight verbs. It was wrong
before #218 and stayed wrong through two rounds of fixing its immediate
neighbours, because I was reading the diff rather than the file. Now phrased
without a count at all.
