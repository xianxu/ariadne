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

Adjacent, and fixed incidentally by the same change: `syncInPlace` pushes
without fetching first, so an ID allocated by `issue.NextID` from a stale local
`main` is only discovered when the push is rejected.

## Spec

**Build the commit in the object database and push it. Never touch a checkout.**

**The plumbing already exists — consume it, do not rebuild it.** `gitx.TrunkFile`
(ariadne#209 M1) is exactly this: `fetch` -> `read-tree` into a temp index ->
`hash-object -w --path` -> `update-index` -> `write-tree` -> `commit-tree` ->
`push <commit>:refs/heads/main` as a compare-and-swap, with a bounded retry, and
37 tests against a real bare origin. Every "detail that will bite" this issue
originally enumerated is handled there, two of them better than specced: the mode
is **preserved from the base tree** rather than hardcoded `100644`, and signing
reads `--type=bool` because git stores `commit.gpgsign` verbatim, so a repo
configured `yes` was silently getting unsigned commits.

`push <commit>:main` is the concurrency primitive: if `origin/main` moved since
the fetch, the push is rejected non-fast-forward. That is stronger than the check
it replaces, which can only see *local* divergence.

**This deletes machinery rather than adding a fallback.** The clean-main check and
the merge-base conflict detection exist only because the current route edits a
shared checkout.

### The retry must RE-ALLOCATE, not re-push — this is the correction

An earlier draft of this Spec said retry was trivial: *"the input is the current
content of this one file, not a diff, so on rejection: re-fetch, rebuild,
re-push."* **That is wrong, and wrong in the specific way ariadne#188 documents.**

A content-preserving retry re-pushes the same path with the same id. Two files
with different slugs at one id produce **no textual conflict**, so the duplicate
lands as a clean fast-forward. Replacing `pull --rebase` with a CAS push moves
#188's hole into a new mechanism rather than closing it.

`TrunkFile` does not decide this. `Update` re-reads the trunk and **re-calls the
transform** on a moved base, so mergeability is the caller's property. The
transform here must therefore consult the trunk's *id space*, not just this
file's bytes:

```go
// Runs AFTER Update's fetch, so the ls-tree sees current trunk state; on a CAS
// rejection Update re-fetches and re-runs this, so the check re-evaluates.
transform := func(old []byte) ([]byte, error) {
    if takenOnTrunk(id) {          // ls-tree over the issue dirs, any slug
        return nil, ErrIDTaken     // aborts Update; caller re-allocates
    }
    return content, nil
}
```

A transform error propagates immediately rather than retrying, so the caller
owns the re-allocation loop: allocate -> Update -> on `ErrIDTaken`, re-allocate,
rename, retry (bounded). **Allocate + commit + push is one retryable unit** —
ariadne#188's single surviving bullet, and the reason this issue can supersede it.

**Reserve on the FILENAME, not on frontmatter `id:`.** Measured 2026-09-09 while
claiming this very issue: a second file at `000207-*` existed with **no
frontmatter at all** — a grafted log fragment from another session.
`sdlc claim` still refused, because it matches on filename; a frontmatter-keyed
check would have missed it entirely. The id space the tooling actually collides
on is the filename prefix.

**Scope: replace `syncViaMainWorktree` only.** Leave `syncInPlace` alone — when
the caller is already on main, "here" and "main" coincide, and building
out-of-tree there would leave the caller's own branch behind the commit they just
pushed. (`syncInPlace`'s missing fetch stays a separate fix.)

frontmatter at all** — a grafted log fragment from another session.
`sdlc claim` still refused, because it matches on filename; a frontmatter-keyed
check would have missed it entirely. The id space the tooling actually collides
on is the filename prefix.

**Scope: replace `syncViaMainWorktree` only.** Leave `syncInPlace` alone — when
the caller is already on main, "here" and "main" coincide, and building
out-of-tree there would leave the caller's own branch behind the commit they just
pushed. (`syncInPlace`'s missing fetch stays a separate fix.)

## Done when

- `sdlc issue new` and `sdlc issue sync` publish from a feature branch with
  **no worktree on main anywhere** in the repo.
- They publish while another worktree on main is dirty, mid-rebase, or owned by
  another actor, without reading or writing that worktree.
- **A collision RE-ALLOCATES.** Two publishers race against one bare origin on
  the same id with different slugs; the loser lands at the *next* id and both
  files exist under **distinct** ids. The earlier wording — "asserts both issue
  files land" — is the bug for a collision, since two files landing at one id is
  precisely the defect; it passes on the thing it should catch.
- A push rejected by a concurrent publisher on an *unrelated* path retries and
  succeeds without re-allocating — the two cases are distinguished, not
  conflated.
- The reservation is keyed on the filename prefix, proven by a fixture whose
  colliding file has **no frontmatter**.
- `syncViaMainWorktree` and its clean-main and merge-base conflict checks are
  deleted, not left as a second path (`ARCH-DRY`) — a shadow sweep confirms no
  caller reaches them.
- `gitx.TrunkFile` is consumed, not reimplemented: no second copy of the
  plumbing, and its existing tests still pass unchanged (`ARCH-DRY`).
- ariadne#188 closes as superseded — its one surviving bullet
  (allocate+commit+push as a retryable unit that re-allocates) ships here.
- Tests run against a real throwaway repo with a local bare `origin`
  (`ARCH-MOCK`: git is the external binary, and a temp repo is its portable
  stateful fake — a function-call mock cannot exercise a non-fast-forward
  rejection).
- The published blob round-trips: checking out the pushed commit yields a file
  byte-identical to the local one, with attributes applied.

## Plan

- [ ] Write the collision test FIRST: two publishers, one bare origin, same id,
      different slugs -> distinct ids land. This is the assertion the old
      Done-when got backwards, and the shape that passes on its own defect, so it
      is written before the code it guards.
- [ ] `takenOnTrunk(id)` — `ls-tree` over the issue + history dirs on the
      tracking ref, matching the FILENAME prefix, any slug. Pure decision split
      from the git call (`ARCH-PURE`); a fixture case with no frontmatter.
- [ ] The publish path: allocate -> `TrunkFile.Update` with the id-checking
      transform -> on `ErrIDTaken`, re-allocate + rename + retry, bounded.
- [ ] Repoint the publish-from-elsewhere arm; delete `syncViaMainWorktree`,
      `mainHasUncommittedIssueChanges`, and the merge-base conflict detection.
      Shadow-sweep for callers.
- [ ] Verify no worktree on main is required, and that another worktree on main
      being dirty or mid-rebase is neither read nor written.
- [ ] Close ariadne#188 as superseded, recording which bullet shipped here.

`sdlc claim` still refused, because it matches on filename; a frontmatter-keyed
check would have missed it entirely. The id space the tooling actually collides
on is the filename prefix.

**Scope: replace `syncViaMainWorktree` only.** Leave `syncInPlace` alone — when
the caller is already on main, "here" and "main" coincide, and building
out-of-tree there would leave the caller's own branch behind the commit they just
pushed. (`syncInPlace`'s missing fetch stays a separate fix.)

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
