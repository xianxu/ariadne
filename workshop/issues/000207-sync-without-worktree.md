---
id: 000207
status: working
deps: [ariadne#206]
github_issue:
created: 2026-09-02
updated: 2026-09-09
estimate_hours:
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
is exactly this sequence, with a bounded CAS retry and 37 tests against a real
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
func (t *TrunkFile) UpdateMany(
    msg string,
    prepare func(trunk *TrunkView) (map[string][]byte, error),
) error
```

`TrunkView` gives read access to the just-fetched tracking ref — enough for
`refIDSpace` to run against the base this attempt will race.

Two contract points the single-path version settles differently, and which must
be decided rather than inherited:

- **`prepare` returns content; it does not receive old bytes.** `Update`'s
  `transform(old []byte)` is meaningful for one path. For N, "the old bytes" has
  no single referent, and issue publication does not transform prior content —
  it writes the local file. Callers needing the base read it through `TrunkView`.
- **The unchanged-content early return is WHOLE-SET, not per-file.** `Update`
  skips the commit when `bytes.Equal(old, next)`; here it skips when every
  prepared file already matches the trunk byte-for-byte. That preserves
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
3. Rename the LOCAL file too, so the working tree matches what was published.
   The file is untracked at `issue new` time, so this is `os.Rename`, not
   `git mv` — and that is worth asserting, because a tracked file would need
   staging and the two paths are easy to conflate.
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


## Plan

- [ ] Write the three collision tests FIRST. Each is an assertion an earlier
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
- [ ] `TrunkFile.UpdateMany` — N blobs into one temp index, one `write-tree`,
      one `commit-tree`, one CAS push. Test asserts the commit count.
- [ ] Collision decision as a pure function over `refIDSpace`'s map: given
      (id, my path, trunk id-space) -> publish | re-allocate | refuse. No git in
      its tests (`ARCH-PURE`); a case with a no-frontmatter colliding file.
- [ ] Wire `issue new` to the re-allocate arm (rename + `id:` rewrite + loud
      announcement) and `sync`/`claim` to the refuse arm.
- [ ] Repoint the publish-from-elsewhere arm; delete `syncViaMainWorktree`,
      `mainHasUncommittedIssueChanges`, and the merge-base conflict detection.
      Shadow-sweep for callers.
- [ ] Close ariadne#188 as superseded, recording which bullet shipped here.


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
ariadne#209 M1 with 37 tests against a real bare origin, and handles two details
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
