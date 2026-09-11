---
id: 000207
status: working
deps: [ariadne#206]
github_issue:
created: 2026-09-02
updated: 2026-09-09
estimate_hours: 3.64
started: 2026-09-09T18:26:20-07:00
---

# Publish issue files without a main worktree

## Problem

Publishing an issue file to `origin/main` from a feature branch requires that
*some worktree is checked out on main*. `syncViaMainWorktree` locates it via
`git worktree list --porcelain -z`, and when none exists it dead-ends:

```
could not find a worktree on branch 'main'. Is main checked out somewhere?
```

Hit for real on 2026-09-02: an agent ran `sdlc issue new` in `pair`, whose only
checkout had been moved to a feature branch by another actor. The issue file
was written, the ID was never reserved on `origin`, and the stub was left
untracked inside the other actor's working tree.

The route is elaborate because it drives someone else's checkout — find main,
assert it has no uncommitted issue changes, `pull --rebase` it, detect files
changed on both branches since the merge-base, copy the issue file across,
commit, push. Every one of those steps exists to make a *shared working
directory* safe, and each is a way to fail: main can be dirty, main can be
mid-rebase, main can be another actor's active tree, or main can simply not be
checked out anywhere.

Adjacent but **NOT** fixed here: `syncInPlace` pushes without fetching first, so
an id allocated from a stale local `main` is only discovered when the push is
rejected. An earlier draft claimed this change fixed it incidentally; it does
not, because the Scope below deliberately leaves `syncInPlace` alone. It stays a
separate one-line fix, and saying otherwise would have let it fall through the
gap between two sections of one issue.

## Spec

**Build the commit in the object database and push it. Never touch a checkout.**

**The plumbing already exists — consume it.** `gitx.TrunkFile` (ariadne#209 M1)
is exactly this sequence, with a bounded CAS retry and tests against a real
bare origin. It handles two details better than this Spec originally specced:
mode is **preserved from the base tree** rather than hardcoded `100644`, and
signing reads `--type=bool`, because git stores `commit.gpgsign` verbatim and a
repo configured `yes` was silently getting unsigned commits.

`push <commit>:main` is the concurrency primitive: if `origin/main` moved since
the fetch, the push is rejected non-fast-forward. Stronger than the check it
replaces, which can only see *local* divergence.

**This deletes machinery rather than adding a fallback.** The clean-main check
and the merge-base conflict detection exist only because the current route edits
a shared checkout.

### One commit, N files — and paths derived PER ATTEMPT

`syncViaMainWorktree` copies **every changed issue file** and commits them
together. `TrunkFile.Update` writes one path per commit, so a naive replacement
turns one commit into N and makes a partial publish reachable where it is not
today.

The signature is where the correctness lives, so it is written out rather than
elided:

```go
// UpdateMany publishes N files in ONE commit, DERIVING them on every attempt.
//
// prepare runs after each fetch and before the tree is built, and returns the
// complete set of paths to write. Paths are derived per attempt rather than
// passed in as a map, and that is the whole point: a caller whose path depends
// on trunk state — an issue id — must re-decide it against the base the CAS
// will actually race. A fixed map makes the retry re-push the same colliding
// id and land the duplicate as a clean fast-forward, which is ariadne#188's
// hole reappearing one layer up.
type TrunkWrite struct {
    Write  map[string][]byte // paths to create or replace
    Delete []string          // paths to REMOVE from the trunk
}

func (t *TrunkFile) UpdateMany(
    msg string,
    prepare func(trunk *TrunkView) (TrunkWrite, error),
) error
```

**Deletion is a separate field, not an absent key.** A bare
`map[string][]byte` cannot express "remove this path" — the path is simply
missing, which is indistinguishable from "unchanged". Today's arm fails LOUDLY on
a deleted issue file (`os.ReadFile` at `claim.go:459`); the whole-set early
return would instead report success while the file stayed on the trunk. That is a
silent-success regression, and it is the same shape this fleet has now shipped
several times: a state that matters was not representable, so it collapsed into
the benign one.

`TrunkView` gives read access to the just-fetched tracking ref — enough for
`refIDSpace` to run against the base this attempt will race.

Two contract points the single-path version settles differently, and which must
be decided rather than inherited:

- **`prepare` returns content; it does not receive old bytes.** `Update`'s
  `transform(old []byte)` is meaningful for one path. For N, "the old bytes" has
  no single referent, and issue publication does not transform prior content —
  it writes the local file. Callers needing the base read it through `TrunkView`.
- **The unchanged-content early return is WHOLE-SET, not per-file** — and it
  must account for `Delete`: skip the commit only when every `Write` already
  matches the trunk byte-for-byte AND every `Delete` path is already absent. That preserves
  `filesDifferingFrom`'s idempotence (`claim.go:398`) — "main already carries
  this body, publish only" must stay a no-op commit-wise, and a per-file rule
  would emit a commit whenever any one file differed.


### Re-allocation belongs to `issue new` ONLY

An earlier draft of this revision said the transform should check "is this id
taken on the trunk, any slug" and re-allocate. **That is wrong and would renumber
every existing issue**, because `syncIssuesToMain` serves republication too:
`sdlc issue sync` and `sdlc claim` push an issue whose id is *already* on the
trunk — its own. Any "id is taken" check fires on every one of them.

The asymmetry is not incidental, it is ariadne#188's whole point. Renumbering is
safe **only before anything references the id**. By `claim` time it has leaked
into the branch name, and after that into commit subjects agents grep, `deps:` in
sibling issues, and review sidecar filenames. So:

| Caller | On finding a *different slug* at this id on the trunk |
|---|---|
| `issue new` (first publication, nothing references it yet) | **Re-allocate** — see the step below. |
| `issue sync` / `claim` (republication) | **Refuse loudly.** The id has leaked; renumbering here is worse than the collision. Name both paths and point at a repair. |

Same path, same slug, is never a collision — that is our own prior publication.

**The re-allocate step, specified.** It is the riskiest thing this issue does, so
it is not left as "rename + retry":

1. Pick the next free id from `refIDSpace` over the base this attempt races —
   not from a value computed before the loop.
2. Rewrite **both** the filename and the `id:` frontmatter field. Either alone
   leaves an artifact whose name and body disagree, which every downstream
   consumer resolves differently (`sdlc claim` matches the filename;
   `vocabulary validate-instance` reads the frontmatter).
3. Bring the LOCAL file to the new path, ordered so a crash cannot strand the
   trunk: **write the new local path BEFORE the push, remove the old one only
   after the push succeeds.** The reverse order — push, then rename — leaves a
   window where the trunk holds the new id while the local file still carries the
   old one, and since `NextID` scans local files too, the next run would allocate
   around a name nothing published. With this order a crash leaves one harmless
   untracked duplicate and a consistent trunk. The file is untracked at
   `issue new` time, so this is `os.Rename`/`os.Remove`, not `git mv`.
4. Announce the new id **loudly** on stderr, naming the old id, the new one, and
   the colliding path. An operator who typed `issue new` and got a different
   number than the tool first reported must be told why.
5. Bounded at 3 attempts; on exhaustion refuse and name every colliding path
   seen, rather than reporting a generic contention error.

Nothing else is rewritten, because nothing else can reference the id yet — that
is exactly why this arm is safe here and refused for `sync`/`claim`.

**Consume `refIDSpace`, do not reimplement it.** `issueids.go:143` is already the
single-source trunk id-space reader, extracted in ariadne#213 because "the same
silent-degradation defect had to be found four separate times. One reader, one
failure policy." A `takenOnTrunk` helper would be the fifth (`ARCH-DRY`).

**Reserve on the FILENAME prefix, not frontmatter `id:`.** Measured 2026-09-09
while claiming this issue: a second file at `000207-*` existed with no
frontmatter at all — a grafted log fragment from another session — and `sdlc
claim` still refused, because it matches on filename. A frontmatter-keyed check
would have missed it. `refIDSpace` already keys on filename.

**Scope: replace `syncViaMainWorktree` only.** Leave `syncInPlace` alone — when
the caller is already on main, "here" and "main" coincide, and building
out-of-tree there would leave the caller's own branch behind the commit they just
pushed. (`syncInPlace`'s missing fetch stays a separate fix; the Problem section
above is corrected to say so rather than claiming it is fixed incidentally.)


## Done when

- `sdlc issue new` and `sdlc issue sync` publish from a feature branch with **no
  worktree on main anywhere**, and while another worktree on main is dirty or
  mid-rebase — without reading or writing that worktree.
- **N changed issue files land in ONE commit**, as they do today. A test asserts
  the commit count, not just the file contents; per-file commits would pass a
  contents-only assertion.
- **`issue new` re-allocates on a collision that appears MID-RETRY**, not only
  on one visible before the first attempt. The test seeds the peer inside the
  retry window and asserts `prepare` ran twice; a test that seeds it beforehand
  passes even when the decision sits outside the loop, which is the defect a
  fixed path map would ship. Both files end up under **distinct** ids, with
  filename and `id:` frontmatter rewritten together.
  (The original Done-when said "asserts both issue files land" — for a collision
  that is the bug, so it passed on exactly what it should have caught.)
- **`sync`/`claim` REFUSE on collision** rather than renumbering, and the refusal
  names both paths. A test asserts a republication of an already-published issue
  does *not* renumber — the regression the first draft of this Spec would have
  shipped.
- A push rejected on an *unrelated* path retries and succeeds without
  re-allocating; the two cases are distinguished, not conflated.
- The reservation is keyed on the filename prefix, proven by a fixture whose
  colliding file has **no frontmatter**.
- A **non-ASCII filename** publishes, through both the `diff` and the `ls-files`
  query — git quotes such paths by default, and the quoted form does not exist
  on disk.
- The trunk id space is read with `--full-tree`, so the collision guard is not
  blind when run from a subdirectory.
- Two files in ONE publish claiming one id are refused, naming both.
- **Publishing works from a SUBDIRECTORY**, through both the git-query path and
  the `PublishExisting` glob path — git resolves a pathspec against the process
  cwd, and a glob likewise.
- A retry does not treat its OWN rejected candidate as a taken id, so a
  re-allocation does not walk forward on every attempt.
- TWO re-allocations in one publish each land at a distinct id, with both
  originals removed and neither published file deleted.
- A FAILED publish keeps every original and removes the candidates it wrote; the
  write-before-push ordering is witnessed by a `pre-receive` hook that checks for
  the candidate while the push is in flight.
- `sdlc issue new` prints the path it ACTUALLY published, and its recovery advice
  after a taken-id failure does not point at a verb that refuses.
- The lost-update behaviour this replaces is **documented, not hidden**:
  `claim --help` and `atlas/workflow/issue-sync.md` both state it, and
  ariadne#222 carries the design.
- **A deleted issue file is removed from the trunk**, and a test asserts it —
  today's arm fails loudly on the missing source, so silently reporting success
  would be a regression the suite must catch, not a behaviour change.
- The re-allocate ordering is asserted: a failure injected between the local
  write and the push leaves the trunk unchanged and the working tree recoverable,
  never the trunk ahead of the local name.
- `refIDSpace` is the only trunk id-space reader; no second implementation
  (`ARCH-DRY`), confirmed by a shadow sweep.
- `gitx.TrunkFile` is consumed, not reimplemented, and its existing tests pass
  unchanged.
- `syncViaMainWorktree`, `mainHasUncommittedIssueChanges` and the merge-base
  conflict detection are deleted, not left as a second path; a shadow sweep
  confirms no caller reaches them.
- ariadne#188 closes as superseded — its one surviving bullet
  (allocate+commit+push as a retryable unit that re-allocates) ships here.
- Tests run against a real throwaway repo with a local bare `origin`
  (`ARCH-MOCK`); a function-call mock cannot produce a non-fast-forward
  rejection.
- The published blob round-trips: checking out the pushed commit yields a file
  byte-identical to the local one, with attributes applied.


## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec              design=1.2  impl=0.1
item: scope-pivot             design=0.5  impl=0.1
item: greenfield-go-module    design=0.1  impl=0.28
item: smaller-go-module       design=0.05 impl=0.2
item: cross-cutting-refactor  design=0.05 impl=0.14
item: atlas-docs              design=0.05 impl=0.1
item: milestone-review        design=0.0  impl=0.18
design-buffer: 0.30
total: 3.64
```

Σdesign 1.95 × 1.30 + Σimpl 1.10 × 1.0 = 3.64. Revised from 3.27 after
estimate-quality; four corrections, all in the same direction.

Priced against the window `sdlc actual` computes, which opens at `e014809d`
(2026-09-02) and attributes **2.12h to #207** with no code written — the number
the ledger will compare against.

**`design-buffer` 0.15 -> 0.30, by the book.** v2.1 grants +15% for a thorough
plan doc and there is none: `workshop/plans/000207-*-plan-gate.md` is a
plan-quality findings sidecar, not a design plan, and a later reader could
conflate them on filename. The anti-double-count escape hatch does not apply
either — the ×0.2 discount covers 4 of 7 rows but only 0.2 of the 1.9 design
subtotal, so there is almost nothing to prevent double-counting of.

**`UpdateMany` is `greenfield-go-module`, not `smaller-go-module`.** The previous
block kept `familiarity: 1.0` and borrowed ariadne#209's argument for it — that
the novelty premium rides inside a row's `impl`. That only works when the row has
band left, and both rows were already at the scaled ceiling (0.2 = 0.5 × 0.4),
so there was nowhere for it to go. Reclassifying is the honest fix: `TrunkWrite`
plus a per-attempt `prepare` with delete support and a whole-set early return is
a new concern, not a mirror-or-extend, and its band (0.3-0.8 -> 0.12-0.32 scaled)
has room at 0.28. A global ×1.5 would have inflated the four rows that are not
novel.

**`scope-pivot impl` 0.2 -> 0.1.** The table's pivot `impl` prices reworking
artifacts already built; this pivot landed pre-implementation and the
spec-rewriting labour is already carried by `issue-spec impl=0.1`. That was
~0.1h of double-count, in the one direction the rest of the block was under.

**`atlas-docs impl` 0.06 -> 0.1** to cover Plan step 6, closing ariadne#188 as
superseded — a real deliverable that no row named.

`issue-spec` (1.2, undiscounted per the ariadne#215 lesson) is the original
authoring plus the grafted `pair#195` field evidence. `scope-pivot` (0.5 design)
is today's redesign, kept separate because two Criticals reversed the
re-allocation model and PQ-8 replaced a parameter that could not put the
collision decision inside the CAS loop — that is a pivot, not spec-writing, and
folding it into `issue-spec` is what ariadne#218 was called on. `impl` on the
code rows sits high because the tests are real-git and the load-bearing one seeds
a peer *inside* the retry window.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*


## Plan

- [x] Write the three collision tests FIRST. Each is an assertion an earlier
      draft got backwards, and (b) is the one that decides whether the design is
      correct at all:
      **(a)** collision seeded BEFORE the first `prepare` — re-allocates on
      attempt 1. The easy path; passes even if the decision sits outside the
      retry loop, so it proves little on its own.
      **(b)** collision seeded DURING THE RETRY WINDOW — `prepare` pushes a peer
      file at our id on its *first* invocation, so attempt 1's push is rejected,
      `UpdateMany` re-fetches, and `prepare` runs again against a base that now
      holds the collision. Assert distinct ids land AND that `prepare` ran
      twice. **This fails if the collision decision is not inside the loop**, and
      it is the case a fixed `files` map re-pushes as a clean fast-forward.
      **(c)** `issue sync` of an already-published issue does **not** renumber.
- [x] `TrunkFile.UpdateMany` — `TrunkWrite{Write, Delete}` into one temp index,
      one `write-tree`, one `commit-tree`, one CAS push. Tests assert the commit
      count, that a `Delete` path leaves the trunk, and that the whole-set early
      return fires only when writes match AND deletes are already absent.
- [x] Collision decision as a pure function over `refIDSpace`'s map: given
      (id, my path, trunk id-space) -> publish | re-allocate | refuse. No git in
      its tests (`ARCH-PURE`); a case with a no-frontmatter colliding file.
- [x] Wire `issue new` to the re-allocate arm (rename + `id:` rewrite + loud
      announcement) and `sync`/`claim` to the refuse arm.
- [x] Repoint the publish-from-elsewhere arm; delete `syncViaMainWorktree`,
      `mainHasUncommittedIssueChanges`, and the merge-base conflict detection.
      Shadow-sweep for callers.
- [x] Record in ariadne#188 which bullet shipped here — its supersession table
      (`b7f08ec`). #188's CLOSE is #188's own lifecycle step, run after this
      one: an issue's plan should not depend on another issue's close gate.


## Log

### 2026-09-02

Blocked on `ariadne#206`, which is rewriting `syncIssuesToMain` right now —
this issue changes the same dispatch, so it lands after #206 merges rather than
racing it.

Grep confirms sdlc uses **no** git plumbing today: `hash-object`, `commit-tree`,
`write-tree`, `read-tree`, `update-index`, `mktree` and `GIT_INDEX_FILE` appear
nowhere in `cmd/` or `pkg/`. Both sync arms are porcelain over a checkout. So
this introduces the first plumbing use in the tree, which is worth weighing
against the machinery it removes — the argument for it is that every deleted
check exists solely to make a shared working directory safe.

Note sdlc *can* already create worktrees (`branchcreate.go:150,155` do
`worktree add` / `worktree add -b`; `merge.go:577` removes them), so "create a
temp worktree on main when none exists" is the available alternative. Rejected:
it checks out an entire repository to add one markdown file, needs
cleanup-on-failure, and races `git worktree prune` — while still leaving the
clean-main and conflict machinery in place. It solves the missing-worktree case
and none of the others.

### 2026-09-06 — field evidence from pair (was pair#195)

`pair#195` was filed against this same failure in the consumer repo before it
was noticed that `sdlc` owns it (`cmd/sdlc/claim.go:601`). It is now closed as a
duplicate of this issue; its evidence is grafted here because it is the
strongest case either issue carries — this is no longer a hypothetical race.

**Three ID collisions have actually occurred**, each costing a renumber plus
reference-chasing:

| ids | collided files | resolved |
| --- | --- | --- |
| `000172` | `clickable-status-bar` vs `parallelize-zellij-session-snapshot` | latter → `pair#191` |
| `000173` | `wire-actor-description` vs `disposition-six-production-symbols` | latter → `pair#192` |
| `000179` | `layout2-terminal-toggle` (open) vs `reattach-a-detached-thread` (**done, archived**) | former → `pair#194` |

The third is the worst shape and worth designing against specifically: one side
was already closed and archived, so `#179` named both a live issue and a shipped
one, and `sdlc claim --issue 179` refused outright with "multiple issue files
match" — the collision **blocked the issue from being worked at all**, and the
archive meant the duplicate was invisible to anyone reading `workshop/issues/`.

**Why this repo hits it structurally, not incidentally.** `sdlc change-code`
branches **in place** by default rather than creating a worktree, so a repo
being actively worked has no worktree on `main` — the reservation mechanism is
unavailable exactly when the workflow's own default mode is in use.
`git worktree list` in `pair` on 2026-09-06 confirms: the only checkout is on
`000172-clickable-status-bar`.

**Current exposure, measured 2026-09-06.** Every issue filed in `pair` across
these sessions printed the warning and is unreserved right now: **`pair#185`
through `pair#204`**, which includes nine filed in a single session that day
(`#196`–`#204`). Any concurrent session allocating "the next free ID" collides
with one of them and does not find out until someone runs `sdlc claim`.

Note the asymmetry that makes this easy to miss: `ariadne`'s own issues
(`#215`, `#216`, filed the same day) published to `origin/main` without
complaint, because this checkout *is* on `main`. The failure is invisible from
the repo that owns the code.

### 2026-09-07 — the primitive now exists (ariadne#209)

`gitx.TrunkFile` (`cmd/sdlc/internal/gitx/trunkfile.go`, shipped in #209 M1) is
the out-of-tree publish path this issue specs, already built and tested:
`fetch` → `read-tree` into a temp index → `hash-object -w --path` →
`update-index` → `write-tree` → `commit-tree` → `push <commit>:refs/heads/main`
as a compare-and-swap, with a bounded retry. Every plumbing detail this issue
names is handled — `--path` for `.gitattributes`, `-S` under `commit.gpgsign`
(read with `--type=bool`, because git stores the value verbatim and a repo
configured `yes` was silently getting unsigned commits), an absolute
`GIT_INDEX_FILE` in a `0700` temp dir, and file-mode preservation so an
executable path does not come back `100644`.

**Consume it rather than writing a second one** (`ARCH-DRY`). The seam is:

    Update(path, msg string, transform func(old []byte) (new []byte, err error)) error

**One correction to this issue's Spec, and it is the reason the seam is shaped
that way.** The Spec says retry is "trivial and bounded… the input is *the
current content of this one file*, not a diff, so on rejection: re-fetch,
rebuild, re-push." A content-preserving retry re-pushes the same path with the
same id — and since two files with different slugs at one id produce no textual
conflict, the duplicate lands as a clean fast-forward. That is precisely the hole
`ariadne#188` documents ("never rebase-and-retry holding the ID already
chosen"); replacing `pull --rebase` with a CAS push moves the hole into the new
mechanism rather than closing it.

`TrunkFile` does not decide this: `Update` re-reads and **re-calls the
transform** on a moved base, so mergeability is the caller's property. This
issue's transform would set content, which keeps last-writer-wins — correct for
an issue body, wrong for a colliding id.

*(ariadne#218 removed the queue verb that was `TrunkFile`'s first consumer, so
once this issue lands it will be the seam's only caller — today it has none.
Nothing about the seam changed: the queue's intent-replaying transform was the
worked example of a caller choosing merge semantics, and its removal does not
alter that the choice is the caller's.)*

So the Done-when clause *"a test drives two publishers against one bare origin
and asserts **both issue files land**"* needs amending: for two publishers
colliding on one id, both landing IS the bug. Assert that distinct ids
land, and that a collision re-allocates (the one #188 bullet #213 did not take).

### 2026-09-09 — a duplicate #207 file, consolidated (and it proves the issue)

The evidence above arrived as a SECOND FILE at this id —
`000207-publish-issue-files-without-a-main-worktree.md`, written by another
session (`a4a52c0`, 2026-09-06) that meant to append to this `## Log` and instead
created a headerless fragment with no frontmatter. `sdlc claim --issue 207`
refused today with "multiple issue files match", which is the exact failure mode
the grafted evidence calls the worst shape: **the collision blocked the issue
from being worked at all.**

So this issue's own file collided, on the id of the issue that exists to stop
collisions. Consolidated here; the orphan is deleted.

Two things worth carrying into the fix:

- **The #213 CI gate saw it and did not block.** `40-duplicate-issue-id.sh`
  reports pre-existing within-ref duplicates and exits 0 by design — blocking
  every merge until the backlog is renumbered would be worse than the bug. It
  currently reports three: `#000040` (two archived), `#000096` (one live, one
  archived — the worst shape), and this one. Working as specified; the residue is
  that a report nobody reads is indistinguishable from silence, and I merged #218
  straight past it by reading the summary line instead of the output.
- **The two mechanisms key differently.** `sdlc claim` matched on FILENAME
  (`000207-*`), which is why it caught this despite the fragment having no
  frontmatter at all. A check keyed on frontmatter `id:` would have missed a file
  that has none. Whatever #207 builds should reserve on the filename, since that
  is what the tooling collides on.

### 2026-09-09 — Spec revised before implementation

Three changes, none of them cosmetic.

**The plumbing block was replaced by "consume `gitx.TrunkFile`".** It shipped in
ariadne#209 M1 with tests against a real bare origin, and handles two details
better than this Spec had specced: file mode preserved from the base tree rather
than hardcoded `100644`, and signing read via `--type=bool` because git stores
`commit.gpgsign` verbatim.

**The retry semantics were wrong and are now the Spec's centrepiece.** "Re-fetch,
rebuild, re-push" re-lands a colliding id as a clean fast-forward, which is
ariadne#188's hole relocated into a new mechanism. The transform must consult the
trunk's id space and abort with `ErrIDTaken` so the caller re-allocates. That
makes allocate+commit+push one retryable unit — #188's single surviving bullet —
so this issue can supersede it.

**The Done-when asserted the defect.** "A test drives two publishers and asserts
**both issue files land**" passes when two files land at one id, which is exactly
the bug. Now: distinct ids, with the unrelated-path retry distinguished from the
collision retry rather than conflated. That assertion is Plan step 1 because it is
the shape that passes on its own defect.

Also folded in: reserve on the **filename prefix**, not frontmatter `id:` —
measured while claiming this issue, when a colliding `000207-*` file with no
frontmatter at all blocked `sdlc claim`. A frontmatter-keyed check would have
missed it.

### 2026-09-09 — plan-quality round 1: two Criticals in my own revision

The Spec revision I wrote hours earlier was wrong in three structural ways, all
caught before code.

**PQ-1 (Critical): the re-allocation check would have renumbered every existing
issue.** I specced "if the id is taken on the trunk, any slug, re-allocate" —
forgetting that `syncIssuesToMain` serves republication as well as creation.
`sdlc issue sync` and `sdlc claim` push an issue whose id is already on the
trunk (its own), so the check fires on every one of them. Re-allocation is only
safe before anything references the id; by `claim` time it has leaked into the
branch name. Split by caller: `issue new` re-allocates, `sync`/`claim` refuse.

**PQ-2 (Critical): `TrunkFile.Update` writes one path per commit; the arm it
replaces publishes N files in one.** A naive swap turns one commit into N and
makes a partial publish reachable. Needs `UpdateMany` — a small extension, since
the temp-index build already supports N blobs before one `write-tree`.

**PQ-3: I was about to write a fifth trunk id-space reader.** `refIDSpace`
(`issueids.go:143`) exists precisely because "the same silent-degradation defect
had to be found four separate times. One reader, one failure policy." Consume it.

All three share a shape worth naming: I designed against the consumer I had in
mind (`issue new`) and not against the seam's actual callers, then specced a
helper without checking whether the single source already existed. The
Done-when now asserts the *regression* too — that a republication does not
renumber — because that is the failure my own first draft would have shipped.

### 2026-09-09 — plan-quality round 2: the ellipsis was hiding the design

**PQ-8.** I wrote `UpdateMany(files map[string][]byte, msg string, transform …)`
with a literal ellipsis — at exactly the parameter that decides whether the CAS
retry re-lands a collision. Three things were hiding in it, and the third is
fatal:

- `files`-as-content contradicts a transform that receives old bytes
  (`trunkfile.go:369`); for N paths "the old bytes" has no referent.
- The unchanged-content early return (`trunkfile.go:377`) is per-file or
  whole-set, and that choice decides whether `filesDifferingFrom`'s idempotence
  survives (`claim.go:398`).
- **Nothing said the collision decision runs inside the retry loop.** With a
  fixed `files` map it cannot: a peer landing mid-window causes rejection ->
  retry -> re-push of the same id -> duplicate as a clean fast-forward.
  ariadne#188's hole, one layer up, in the issue that exists to close it.

Resolved by deriving paths **per attempt**: `prepare(trunk *TrunkView)` runs
after each fetch and returns the whole set, so a caller whose path depends on
trunk state re-decides it against the base the CAS will actually race.

**And my own step-1 test would have passed on that defect.** "Two publishers race
on one id" seeds the collision before the first check, so `prepare` sees it
immediately and re-allocates — without the retry ever running. The test now
seeds the peer *inside* the retry window and asserts `prepare` ran twice. That is
the third time this session a guard would have passed on the thing it was written
to catch, and the second where the reviewer found it rather than the suite.

**PQ-4.** "Re-allocate + rename + retry, bounded" is now five numbered steps,
because it is the riskiest operation here: id re-picked from the *current* base,
filename and `id:` frontmatter rewritten together (either alone leaves name and
body disagreeing, and downstream consumers resolve each differently), the local
file renamed with `os.Rename` since it is untracked at this point, a loud
announcement naming old id / new id / colliding path, and on exhaustion a refusal
naming every collision seen.

### 2026-09-09 — plan-quality round 3: cleared, with two residues taken

**PQ-5 (advisory, taken now rather than at close).** The Plan said nothing about
a DELETED issue file. `prepare` returning `map[string][]byte` cannot express
removal — an absent key is indistinguishable from unchanged — so the whole-set
early return would have reported success while the file stayed on the trunk.
Today's arm fails loudly (`os.ReadFile`, `claim.go:459`), so that is a
silent-success regression, not a behaviour change. `TrunkWrite{Write, Delete}`
makes removal representable; the early return now accounts for both halves.

Third time this fleet has shipped the same shape — a state that mattered was not
representable, so it collapsed into the benign one — after `cat-file -e`
conflating absent with unreadable, and `KindIssue` as the zero value hiding "not
mentioned".

**PQ-4 residual.** Step 3 implied the local rename happened after the push,
leaving an unnamed crash window: trunk holds the new id, local file keeps the
old, and since `NextID` scans local files the next run allocates around a name
nothing published. Reordered — write the new local path before the push, remove
the old only after it succeeds — so a crash leaves one harmless untracked
duplicate and a consistent trunk. Asserted in Done-when rather than left as
prose.

### 2026-09-11 — #188's boundary review, run over this issue's code

`sdlc close --issue 188` came back REWORK, and its window (`ff1b52b6..b7f08ecd`)
held this issue's implementation, so it served as an early review of #207.
Fourteen findings; every one of BR-1 to BR-13 is addressed, and BR-14 is the
pre-existing ariadne#210 failure.

**Three Criticals, and I first reported two.** I read only the tail of the
review output, which cut BR-1, and told the operator there were two. BR-1 was
still unfixed a day later.

- **BR-1:** `issue new` and `claim` pass `msg=""` meaning "the default subject",
  and the trunk arm handed `""` straight to `commit-tree` — an empty subject on
  every publish from a feature branch, which change-code makes the common case.
  Now one `defaultSyncSubject`, shared by both arms.
- **BR-2:** a failed publish ran `finish()` and deleted the original file.
- **BR-3:** `nextFreeID` saw only the trunk, so re-allocation could land on an
  id a local unpublished file held. I had cited ariadne#213's union rule in
  #188's supersession note hours before writing this code without it.

**BR-5's first fix was itself wrong.** It printed `synced` on every nil return,
including dry-run and the no-change exit — a machine-readable marker naming a
publication that never happened. `syncViaTrunkWithRealloc` now reports
`onTrunk`, true only when `UpdateMany` succeeded.

**BR-8 closed with real-git tests**, not fakes: a dirty, mid-rebase worktree on
main is neither read nor written; a declining `pre-receive` hook injects a
failure between the local write and the push, and the trunk stays put while the
original file survives; and a fresh clone's checkout is byte-identical to the
local file under `.gitattributes`, where `strings.Contains` had passed on
truncation and whitespace changes alike.

**Mutation-verified:** BR-1, BR-2 (by fake and by real git), BR-3, BR-4, BR-5,
BR-8b and BR-8c are each caught with the fix reverted. BR-8a has no mutation
target — the arm no longer touches a worktree — so that test guards against the
route being reintroduced.

**BR-13 was deleted, not fixed.** `CheckRunsOnEveryAttempt` asserted a counter
the fake increments unconditionally, so it could not fail; the property is held
by `ReallocatesOnMidRetryCollision`, which fails when the id-space read is
hoisted out of the loop.

Plan step 6 reworded: #207 records which of #188's bullets shipped here (done,
`b7f08ec`); closing #188 is #188's own lifecycle step, run after this one.

### 2026-09-11 — close review round 1: REWORK, 15 findings

Two Criticals and seven Importants, several of them produced by the PREVIOUS
round's fixes — the same compounding #209 showed.

- **BR-2 came from my BR-4 fix.** Resetting `rc` per attempt also dropped the
  record of files earlier attempts had written, so a rejected attempt's candidate
  survived as an orphan reserving an id nothing published. Split into
  `publishResult`: decisions reset per attempt, candidates persist across them.
- **BR-3**: one `rc` could not hold two re-allocations in one publish, so
  `finish()` paired the first file's original with the last file's new path and
  deleted a file that had just been published. Now a list.
- **BR-1 (Critical)**: `issue new` prints the created path on stdout as a
  machine contract, and after a re-allocation that file is the one `finish()`
  removed. It now prints what it published — and `runIssueNew` gained a
  publisher seam, because the path had no end-to-end test at all without one.
- **BR-9**: git QUOTES non-ASCII paths, the quoted form does not exist on disk,
  the read failed as not-exist, and the publisher classified it as a DELETION —
  then reported success having published nothing. Every query now uses `-z`, and
  an unexplained not-exist is an error rather than an assumed deletion.
- **BR-12** was pre-existing and wider than this issue: `ls-tree` resolves its
  pathspec against the CWD, so `refIDSpace` returned an EMPTY id space from any
  subdirectory — ariadne#213's allocation was blind the same way. `--full-tree`.
- **BR-6**: `unionIDSpace` swallowed a failed local scan reasoning that "the CAS
  will reject anyway". False — the CAS compares refs and knows nothing about an
  unpublished local id.
- **BR-8**: `Update` and `UpdateMany` carried duplicate retry loops; `Update` is
  now a thin adapter, which also removed `commitAndPush` entirely.
- **BR-5** is a real semantic loss, not a slip: same-issue concurrent edits are
  last-writer-wins now that the merge-base check is gone. The reviewer's remedy —
  document it and file a follow-up — is what shipped: `claim --help` and the
  atlas both state it, and **ariadne#222** carries the design, including why a
  naive merge-base check would refuse every ordinary republish.

**Twelve mutations run, twelve caught.** One took two tries: dropping `-z` from
the `diff` query was NOT caught, because the test's accented file was untracked
and so arrived via `ls-files`. The test now covers a committed-and-modified file
AND an untracked one, and both mutations are caught. Second time this session a
mutation was aimed at a line the test could not observe.

**Dogfooded for real**: this round's follow-up (ariadne#222) was filed with the
newly built binary from this feature branch, with no worktree on main — it landed
on `origin/main` under the default subject, which is BR-1's fix working in
production rather than in a fixture.

### 2026-09-11 — close review round 2: FIX-THEN-SHIP, three blockers

**BR-16 is the one that matters, and it is my own failure to sweep a class.**
Round 1's BR-12 said `ls-tree` resolves its pathspec against the process cwd; I
fixed `refIDSpace` and stopped. The same trap sat in `changedIssueFiles`'
pathspecs and `issueFilesForID`'s glob, so from any subdirectory `issue new`,
`claim` and `issue sync` published **nothing** and exited 0 — the id never
reserved, which is verbatim this issue's own Problem statement. The reviewer
reproduced it with the HEAD binary from `docs/sub`. Both paths are now pinned to
the repo root, and both mutations are caught. This is exactly what the ledger
means by "fix rules, not instances", and it took a second finding to land.

**BR-2 was still open for two distinct reasons.** My candidates-persist fix had
no test that ran `finish()` after a TWO-attempt re-allocation, so resetting
candidates would have left the suite green — the guard existed but nothing held
it. And underneath sat a real bug the reviewer reproduced: the retry counted its
OWN rejected candidate as a taken id and walked 000700 to 000702 with 000701
free, leaving an orphan reserving an id nothing published. Candidates are now
excluded from the id space they helped create, with both mutations caught.

**BR-7**: six further stale sites, including a sentence my own earlier edit had
truncated mid-clause in the atlas. Swept by derivation rather than by hand, and
the sweep separated the genuinely-live references — `merge` and the archive
still use `findMainWorktree` — from the ones describing a route that no longer
exists.

**Found while fixing, not reported**: `filepath.Rel` mixed a symlink-resolved
root with an unresolved path, yielding `../../..` escapes that git rejects as
"outside repository". Any repo reached through a symlink hits it — macOS `/tmp`,
a symlinked home — not just a fixture. `repoRel` resolves both sides.

**Reverted as out of scope**: `branchcreate.go` carries the same cwd-relative
shape, but its tests are stub-only with no repo, and pinning a root there needs
one injected through a different call chain. Left as a named sibling rather than
half-changed.
