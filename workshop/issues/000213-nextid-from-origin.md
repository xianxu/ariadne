---
id: 000213
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-03
estimate_hours:
started: 2026-09-03T11:04:17-07:00
---

# Allocate issue IDs against origin/main

## Problem

`issue.NextID` scans **local files only** —
`cmd/sdlc/internal/issue/scaffold.go:31`, walking `issuesDir`, `historyDir` and
the archive subdir for the highest 6-digit prefix. A feature branch cut before
some issue landed on `main` has a `workshop/issues/` that never contained it, so
`sdlc issue new` on that branch reallocates ids that already exist.

**This is not a race.** Measured in `pair`, 2026-09-02:

```
12:10:10  branch 000170-… cut from main (88fe1de0)
12:42:03  #171, #172 land on origin/main
15:43:29  actor files its own #171 on the branch
17:07:08  actor files its own #172
```

Three hours after publication. Any `sdlc issue new` on that branch allocates
`171`, whether run then or next week — the branch simply does not contain the
newer files. The result is four files and two ids:

| id | on `origin/main` | on the branch |
|---|---|---|
| 171 | Always-on idle notification fallback | Reconcile stale incarnations left by a crashed couch |
| 172 | Clickable status bar switches to that actor | Parallelize the zellij session snapshot |

**The repo transaction lock is the wrong tool and demonstrably did not help.**
Both allocations ran in linked worktrees of the same repo, so they shared one
`.git/sdlc.lock` and *were* serialized. They still collided. The lock prevents
concurrent access to a shared view; it cannot reconcile two disjoint views.

**And nothing detects it, and never has.** Two colliding issues get different
slugs, so `000212-fleet-policy-json.md` and `000212-issue-status-dup.md` are
different *paths*. Git merges both cleanly. There is no point in the lifecycle at
which anything objects — not the push, not the merge, not `issue validate`.

An earlier draft of this section blamed `#206`'s `syncInPlace` for removing
push-rejection detection. **Measured, that is wrong**, and the correction matters
because it changes the age of the defect:

| repo | colliding ids |
| --- | --- |
| ariadne | #40, #96, #168, #212 |
| parley.nvim | #51, #66, #81, #90 |

`ariadne#40`'s two files were created 2026-05-28 and 2026-05-27; `#96`'s on
2026-06-14 — three months before `#206` landed (2026-09-02). Three of ariadne's
four are already merged and archived. A rejected push only ever caught a
*non-fast-forward race*, never an id collision, because the filenames differ.
This has been silently corrupting the id space since May.

**There is no double-locking anywhere.** `NextID` has one caller
(`cmd/sdlc/issue.go:261`) and does a plain `os.ReadDir`; no code path re-checks
the id at commit, push, or merge. The repo transaction lock is real but
orthogonal — it serializes access to one checkout and cannot reconcile two
disjoint views.

## Spec

**Allocate against the published id space, not the checkout's.**

- Before allocating, `git fetch origin main` (or the configured trunk), then
  compute the next id from the **union** of the local scan and the ids present
  in `origin/main`'s issue directories. Union, not replacement: unpushed local
  issues on this branch are real and must still be excluded.
- Read the remote ref's tree directly — `git ls-tree <ref> <dir>` for the three
  directories `NextID` already scans. Filenames carry the id, so no file
  contents are needed and no checkout is required. This is the same "treat the
  fetched ref as truth" move `ariadne#207` makes for publishing.
- **Offline must not block creation.** A failed fetch falls back to the local
  scan with a **loud** warning naming the risk ("origin unreachable; id
  allocated from local files only — may collide"). Creating an issue is not
  something to refuse when the network is down, but a silent fallback here
  recreates the bug it is meant to fix.
- Keep `NextID` pure. It takes the id sets; the fetch and `ls-tree` live in the
  caller's IO shell (`ARCH-PURE`), so the union logic stays unit-testable
  without git.

**Detection, because allocation alone cannot fix what already exists.** Two
colliding files exist in `pair` right now, and more may be sitting on other
branches. Add a duplicate-id check to the merge gate: refuse a merge that would
land an id already present on the trunk, naming both files. Cheap — the same
`ls-tree` — and it is the last point where the collision is still repairable.

### Enforcement must be in CI, not only in `sdlc merge`

A gate inside `sdlc merge` is **operator feedback, not enforcement**. It is
skipped by merging in the GitHub UI, by plain `gh pr merge`, by `--no-validate`,
and by any actor on a machine that has not pulled this fix. For an id space
shared across machines and repos that is not a guarantee.

`.github/workflows/merge-check.yml` already runs on every PR to main, computes
the merge-base range, and executes `scripts/merge-checks.d/*` through the
symlinked `run-merge-checks.sh` runner. A check there runs server-side, can be a
required status check, and **propagates to every derivative repo** — which
matters because half the measured collisions are in `parley.nvim`, not here.

So: `40-duplicate-issue-id.sh`, matching the existing check's contract
(`<check> <base> <head>`, exit 0 = pass, findings to stderr). It compares the ids
the PR range introduces against those on the base, and refuses a same-id
different-path pair.

### The gate must also see collisions already on the trunk

Comparing `HEAD` against `origin/main` cannot find the eight that already exist:
both sides are on the trunk, so the two trees agree and there is nothing to
object to. `issueFilesByID` keeps the first path seen for a within-ref duplicate,
which silently collapses exactly the state being hunted.

A **within-ref** scan is a different question — "does this single tree contain
two files claiming one id?" — and it is what surfaces the existing damage. It
reports rather than refuses on the base side (the collisions predate the gate and
blocking every merge until they are renumbered would be worse than the bug), and
refuses when the PR *introduces* one.

**Out of scope:** renumbering existing collisions (operator work, done by hand),
and any change to the repo transaction lock, which is working as designed.

## Done when

- `sdlc issue new` on a branch cut before a published issue allocates an id
  higher than that issue's — a test drives a bare origin, a branch cut before a
  push, and asserts no reuse.
- An unpushed local issue on the current branch is still counted, so two
  successive creations on one branch do not collide with each other.
- With origin unreachable, creation still succeeds and emits the warning; a
  test asserts the warning is present, not just that creation worked.
- `sdlc merge` refuses a branch introducing an id that already exists on the
  trunk, naming both paths.
- `scripts/merge-checks.d/40-duplicate-issue-id.sh` performs the same refusal in
  CI, so the guarantee survives a GitHub-UI merge, a bare `gh pr merge`,
  `--no-validate`, and an actor who has not pulled this fix. Exercised against a
  real repo, not only by reading the workflow.
- The check is a plain script under the existing runner contract, so it
  propagates to every derivative through the symlinked runner — `parley.nvim`
  carries four of the eight known collisions.
- A within-ref scan reports duplicate ids already present in a single tree, which
  is the only way the eight existing collisions are visible at all. It REPORTS on
  the base (they predate the gate; blocking every merge would be worse than the
  bug) and REFUSES when the PR introduces one.
- Tests run against a real throwaway repo with a local bare origin
  (`ARCH-MOCK`) — a function-call mock cannot express "a ref exists that the
  worktree does not contain", which is the entire bug.

## Plan

- [x] Extract the id-set union as a pure function; move the scan behind it.
- [x] `ls-tree` the trunk ref for the three issue directories; fetch first.
- [x] Offline fallback with the loud warning.
- [x] Duplicate-id refusal in the merge gate.
- [x] Bare-origin tests: branch-cut-before-push, two-on-one-branch, offline.
- [x] Within-ref duplicate scan (`DuplicateIDsInRef`), pure over an ls-tree
      listing, so the eight existing collisions are detectable at all.
- [x] `scripts/merge-checks.d/40-duplicate-issue-id.sh` — CI enforcement under
      the existing runner contract; refuses ids the PR introduces, reports ones
      already on the base.
- [x] Test the CI check by running it against a real repo with a planted
      collision, not by reading the workflow file.

## Log

### 2026-09-03 — SESSION HANDOFF (read this first)

**Where the work is: committed on `main`, NOT pushed.** Four commits:

```
d50a023  #213: within-ref id scan + CI enforcement
2084d1e  #213: CI enforcement for reused issue ids
6c2e535  #213: issue-sync: spec/plan
b92e13f  #213: allocate issue ids against the trunk, and refuse reused ids at merge
```

Working tree is clean. `origin/main` is at `8193cd8`, four behind.

**Workflow deviation, deliberate but unrecorded until now:** `sdlc change-code`
was never run for this issue, so no feature branch exists and there are no gate
ledgers in `workshop/plans/`. The work was committed straight to `main` because
the operator flagged the defect as pressing. That means **no plan-quality gate,
no estimate, and no boundary review have run** — `estimate_hours:` is still
empty. Whoever picks this up decides between:

- `sdlc push` from main (runs the pre-push judges + publish gate), or
- reset these four commits onto a branch and run the normal
  `change-code → close → pr → merge` arc.

The second is the honest path if the estimate/review evidence is wanted; the
first is faster and the code is already revert-verified. Not my call to make
silently, which is why it is written down instead of done.

**All 8 plan rows are ticked and the suite is green** except the pre-existing,
unrelated `TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` (that is
**#210**, an archived-plan path — not caused by this work).

### What landed, and where

| layer | file | what it does |
| --- | --- | --- |
| allocation | `cmd/sdlc/internal/issue/scaffold.go` | `NextID` is PURE (takes id sets); `ScanLocalIDs`, `IDDirs`, `IDFromFilename`, `IDsInTreeListing`, `DuplicateIDsInRef` |
| allocation IO | `cmd/sdlc/issueids.go` | `allocateIssueID` unions local + `ls-tree origin/main`; `repoRelativeIDDirs`; `refuseDuplicateIssueIDs` (the local merge gate) |
| detection verb | `cmd/sdlc/issuelintids.go` | `sdlc issue lint-ids [--base <sha>] [--head <ref>]` |
| CI enforcement | `scripts/merge-checks.d/40-duplicate-issue-id.sh` | adapter; runs on every PR via the existing `merge-check.yml` |
| local gate hook | `cmd/sdlc/merge.go` step 4.6 | behind `--no-validate` |
| tests | `cmd/sdlc/issueids_test.go` | real repo + real bare origin throughout |

### The three things a fresh session is most likely to get wrong

1. **`sdlc merge`'s gate is NOT enforcement.** It is bypassed by a GitHub-UI
   merge, a bare `gh pr merge`, `--no-validate`, or an actor who has not pulled
   this. The CI check is the enforceable one. Do not describe the local gate as
   if it guaranteed anything — I did, and it was wrong.
2. **Branch-vs-trunk comparison cannot see a collision already merged.** Both
   files are on the trunk, the trees agree, nothing is reported. Only the
   within-ref scan (`DuplicateIDsInRef`) finds those.
3. **`allocateIssueID` runs git against the process cwd** while the issue dirs
   are caller-supplied. `repoRelativeIDDirs` refuses dirs outside the current
   repo for that reason — without it, a `--issues-dir` elsewhere reads THIS
   repo's trunk as its own id space. Two fetch tests caught it by allocating a
   real ariadne id into a temp fixture.

### Immediate next actions, in order

1. Decide the publish path (see the deviation note above), then ship.
2. **Make the CI check a required status check** on `main` in GitHub branch
   protection. Without that it reports but does not block, and the whole point of
   this half was enforcement.
3. **Propagate to derivatives.** `parley.nvim` carries 4 of the 8 known
   collisions and consumes the symlinked runner; `sdlc propagate-base` is the
   mechanism. Not started.

### Still-open work this issue does NOT cover

- **The 8 existing collisions are untouched** — ariadne `#40 #96 #168 #212`,
  parley.nvim `#51 #66 #81 #90`. Scoped as operator work. `sdlc issue lint-ids`
  reports them; nothing renumbers them. ariadne's `#212` pair is live (both
  `open`) and one of the two is another session's issue.
- **Same-second races within one branch.** Allocation reads the trunk and then
  commits separately, so two actors on one branch in the same moment can still
  read identical state. Rare, not the cause of any of the 8, and unaddressed.
- `#210` (the red fleet-plan test) is unrelated but is the reason
  "the suite is green" always needs a qualifier in this repo.

### Loose end I created

`workshop/issues/000214-probe-the-allocator-against-origin.md` was a live probe
of the new allocator — it correctly allocated `000214`, proving the fix, and was
deleted in `b92e13f`. If a future `issue new` hands out `000214` again that is
correct, not a regression: nothing ever occupied that id.


### 2026-09-03 — CI enforcement

The local gate was operator feedback, not enforcement — skipped by a GitHub-UI
merge, a bare `gh pr merge`, `--no-validate`, or an actor who has not pulled the
fix. `scripts/merge-checks.d/40-duplicate-issue-id.sh` now runs the same check
server-side on every PR, and propagates to derivatives through the symlinked
runner.

The logic stayed in Go as `sdlc issue lint-ids` rather than being rewritten in
bash: filename parsing, the three id-bearing directories, and the
introduced-vs-pre-existing split are decided once, with tests. The script is a
four-line adapter.

**The within-ref scan is what makes the existing damage visible.** A
branch-vs-trunk comparison structurally cannot see a collision already merged —
both files are on the trunk, the two trees agree, nothing to report. Running
`sdlc issue lint-ids` on ariadne today prints all four: #40, #96, #168, #212. It
reports rather than refuses them; blocking every merge until they are renumbered
would be worse than the bug.

Verified by driving the real artifacts, not by reading them: the full
`run-merge-checks.sh` exits 0 on a clean range and 1 on a planted collision,
naming both paths, with both checks discovered.

**Two mistakes of mine worth recording.** A probe that planted a collision used
`git add -A` and then `git reset --hard`, which deleted the uncommitted
`lint-ids` verb, `DuplicateIDsInRef`, and the check script — I had to rewrite all
three. A destructive probe must stage only the file it plants. And the first cut
of the script created its temp dir *before* testing whether there was anything to
build, so a repo that cannot run the check died inside setup instead of taking
its documented skip path; every skip condition is now evaluated before any side
effect. The sandbox's restricted `mktemp` is what surfaced it — a real ordering
bug, found by an environment constraint rather than by review.

### 2026-09-03

Implemented. `NextID` is now pure — it takes id sets — and the IO lives in
`cmd/sdlc/issueids.go`: `ScanLocalIDs` for the checkout, `ls-tree` on
`origin/main` for the published space, unioned.

**Revert-verified**, which matters here because a mock cannot express this bug:
restoring the local-only allocator makes
`TestAllocateIssueID_BranchCutBeforePublish` fail with the defect in the message
— "reallocated the published id 000002" — and takes
`TestAllocateIssueID_ScansHistoryOnTheTrunk` with it. Every test runs a real repo
against a real bare origin, with the branch cut *before* a second clone publishes
the colliding id, so the branch's worktree genuinely never contains it.

**A design flaw of mine that the tests caught.** The first cut ran `git` relative
to the process cwd while the issue dirs are caller-supplied. With
`--issues-dir` pointing elsewhere — or in any test using a temp dir — it would
scan THIS repo's trunk for paths that mean something different there, reading a
stranger's id space as if it were ours. `repoRelativeIDDirs` now refuses any dir
outside the current repo and the caller falls back to the local scan with the
warning. Found because two unrelated fetch tests started allocating `000214`:
the real ariadne id space leaking into a temp-dir fixture.

**Fleet damage measured before starting**, since allocation cannot repair what
exists: ariadne #40, #96, #168, #212; parley.nvim #51, #66, #81, #90. `#40` and
`#96` date from May–June 2026, so the merge gate is not optional cleanup — three
of ariadne's four are already archived and beyond reach.

### 2026-09-02

Found while looking for a published `pair#171` that was invisible from a
feature branch holding a different `#171`.

Related but distinct: `ariadne#207` makes *publishing* work without a main
worktree by building the commit against a fetched `origin/main`. This issue
makes *allocation* read the same source. Same principle — the fetched trunk is
the truth, not this checkout — applied at the other end of the lifecycle. They
share no code and neither blocks the other.

Worth stating plainly, since it was assumed otherwise earlier today: "creating
an issue is safe under concurrency" holds only within one branch. Every
long-lived feature branch is a separate id space, silently.
