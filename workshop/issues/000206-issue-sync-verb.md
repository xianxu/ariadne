---
id: 000206
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.31
started: 2026-09-02T11:36:01-07:00
---

# sdlc: commit planning output via issue sync

## Problem

`sdlc issue` is `new / set-status / list / show`. There is **no verb that
commits an issue's body**, so the entire planning phase — `## Spec`, `## Plan`,
`## Log`, the longest phase of an issue and the one that produces the design —
is a plain file write that nothing commits, pushes, or names.

**The early push is by design, and is not what this issue changes.** `sdlc
issue new` and `sdlc claim` publish the *reservation* — an ID and a name — which
is the whole external contract: peers need to know the issue exists and is
taken, not what is in it. Details stay private until a milestone. That is
deliberate and stays.

What is missing is the other half. The design that follows the claim stays
**uncommitted** in the working tree until some unrelated verb happens to sweep
it up, so:

- A compaction, a crash, or a closed terminal loses the design outright.
- When the edits are eventually swept, they land under
  `issue-sync: update issues` rather than a message naming what happened.

Durability and publication are separable, and only durability is missing. A
local commit costs nothing, publishes nothing, and is enough on a single host.

**Separately, a real bug on the shared path.** `syncOnMain` narrows `git add`
to a single issue when `--issue` is set (`cmd/sdlc/claim.go:167-176`) — both
`claim` and `issue new` pass it — but then runs a bare
`git commit -m "issue-sync: update issues"` (`cmd/sdlc/claim.go:179`) with no
pathspec, which commits the **whole index**. Anything a peer agent had staged
in that checkout is swept into a commit that misdescribes it. The repo
transaction lock serializes `sdlc` verbs against each other; it does nothing
about a peer running plain `git add`.

**Deliberately not a problem here.** This is a single-host, single-operator
fleet where the operator initiates every actor and concurrent actors are
overwhelmingly on *different* issues. Locking body edits, detecting content
conflicts, or fetching before `issue.NextID` (`cmd/sdlc/internal/issue/scaffold.go:31`)
would all be machinery for a collision the operator would have to cause on
purpose. The repo lock already lives on the git common dir, so linked worktrees
serialize. Scope is durability and honest commit messages, not mutual exclusion.

## Spec

Three pieces, smallest first.

**1. Pathspec on the shared commit.** `cmd/sdlc/claim.go:179` becomes
`git commit -m <msg> -- <the paths that were added>`. Git treats a pathspec as
implying `--only`, so the rest of the index is ignored. `changedIssueFiles`
already returns early when nothing changed, so the "pathspec matches nothing"
error is unreachable. This fixes every existing caller, not just the new verb.

**2. `sdlc issue sync --issue N` — commits, does not push.** Wrap, do not
author: agents edit markdown incrementally with their own tools, so a verb that
takes body content as an argument would fight how the work actually happens. The
verb takes the repo transaction lock (`markMutatingCommand`), stages only that
issue's files, and commits with a pathspec and a message naming the issue.

**Publishing stays with the milestone verbs.** A mid-planning sync is a cheap,
frequent, local act; pushing is an external contract and belongs at the
boundaries that already own it. `--push` exists for callers that *are* at a
milestone, and is off by default — so the common case cannot accidentally
publish a half-written Spec.

It is a **thin exposure of `syncIssuesToMain`** (`cmd/sdlc/claim.go:96`), which
already exists, is already branch-aware, and already dispatches between
`syncOnMain` and `syncOnBranch`. No second sync path (`ARCH-DRY`) — the verb
supplies a commit message and a publish choice, and reuses the dispatch.

**Which branch receives the durable commit.** The two existing arms are not
"main vs branch" — they are *commit here* vs *publish to origin/main from
elsewhere*. `syncOnBranch` finds the main worktree, refuses if it has
uncommitted issue changes, runs `git pull --rebase origin main` (a network call
that fails offline), copies the files across and commits **in the main
worktree**, leaving the invoking branch's copy dirty. Every one of those steps
exists to *publish*. So the arms get honest names — `syncInPlace` and
`syncViaMainWorktree` — and the dispatch gains its real discriminator:

    no-push  → syncInPlace          (commit in THIS worktree, on THIS branch)
    on main  → syncInPlace          (commit + push origin main)
    else     → syncViaMainWorktree  (today's publish-from-a-branch route)

A no-push sync is therefore local, offline-safe, and durable on the branch the
work is actually happening on — which is what makes it cheap enough to run
mid-planning. It also makes the verb usable from an in-place feature branch,
where `syncViaMainWorktree` cannot run at all because no worktree is on `main`.
`claim` and `issue new` never set no-push, so their dispatch is unchanged.

The publish parameter is spelled **`NoPush`**, not `Push`. `issue new`
constructs `claimFlags` as a struct literal (`cmd/sdlc/issue.go:311`); a `Push`
field would zero-value to false there and silently kill the reservation
broadcast that `#82 M1` added. `NoPush`'s zero value is today's behavior.

The commit message should name the issue the way the rest of the tree does:
`#206: issue-sync: spec/plan` rather than the generic `issue-sync: update
issues`. Message shape is the only thing the new verb adds over the existing
helper, so keep it a parameter of the helper rather than a branch inside it.

**The message anchors `#N`, so it must be declared bookkeeping.**
`cmd/sdlc/internal/gitx/window.go:131-136` states that issue-sync commits need
no `bookkeepingVerbs` entry *because* their subject carries no `#N`. A message
naming the issue breaks exactly that reasoning: `IsShippedWorkSubject` would
return true and count a tracker commit as shipped implementation, feeding
`detectDrift`'s close-off check (`state.go:130`), `branchStartByIssue`'s window
base (`milestoneclose.go:325`) and active-time's run claiming. A sync commit is
bookkeeping, not shipped work — the same category as `claim` and `close` — so
`issue-sync` joins `bookkeepingVerbs`, the stale parenthetical is corrected, and
a `window_test.go` table row pins the subject. The lead-in is `issue-sync`
rather than `issue` precisely so the whole-token match cannot swallow a real
implementation subject like `#206: issue sync verb: …` (`isWordByte` treats `-`
as a word byte, so `issue` alone would not have matched `issue-sync` either).

**3. `change-code` calls it with `--push`.** `change-code` is already the gate
that sits exactly where planning ends — it runs plan-quality and then asks for
the estimate — which makes it precisely one of the milestones at which the
details *should* become external. Syncing there means the design lands durably
and is published at the natural boundary with no new ritual, while the bare verb
from (2) covers mid-planning checkpoints. One mechanism, two triggers, rather
than a parallel auto-sync that could drift from the verb.

It replaces `commitUntrackedIssueFile` (`cmd/sdlc/branchcreate.go:117`), which
is `change-code`'s own private commit+push of the issue file — a third member of
the bare-commit class and a second implementation of what this verb now does.

**Posture: best-effort, warn and continue** — the same posture `issue new`
takes (`cmd/sdlc/issue.go:313`), and a deliberate choice rather than an
inherited one. `commitUntrackedIssueFile` warned on push failure and died on
commit failure; the replacement warns on both, because `change-code`'s job is to
*open implementation* and a tracker commit that could not land must not stand
between the operator and starting work. The warning names `sdlc issue sync
--issue N` as the retry. This also removes the only new way `change-code` could
fail: re-run on an in-place feature branch, `syncViaMainWorktree` finds no main
worktree and errors, where the old helper simply no-opped on an already-tracked
file.

The "branch starts from a tracked state" property the old helper existed for is
preserved in the case that has it: planning ends on `main`, so `syncInPlace`
commits the file in this worktree before the branch is cut. On a re-run from an
existing branch the file is already tracked and there is nothing to preserve.

**Out of scope:** fetch before `NextID`, conflict detection, merge strategy,
any locking of body edits.

## Done when

Every test below runs against a **real git repo** (`testfix.Repo`, the fixture
`hermeticrepo_test.go:14` wraps), never an argv-recording `gitRunner` fake: the
thing under test is git's `--only` pathspec *semantics*, and a fake would assert
the argv and pass even if those semantics were wrong (`ARCH-MOCK`).

- `syncInPlace` commits with a pathspec; a test stages an unrelated file, runs a
  sync, and asserts the unrelated file is still staged and absent from the
  commit.
- `syncViaMainWorktree` does too, proven on its own two-worktree fixture (bare
  origin + a main worktree + a feature worktree) — the arm has no coverage
  today. Its claim is narrower than the in-place arm's: see the Revisions
  correction below.
- Both archive commits — `push.go` (both call sites) and `merge.go` — commit
  with the same pathspec their `archiveAddArgs` staged, via a shared
  `archiveCommitArgs`. `merge`'s is a live gap, not a theoretical one.
- `sdlc issue sync --issue N` commits that issue's files with a message naming
  the issue, from both `main` and a feature branch, and **does not push**: a
  test asserts the default leaves `origin/main` where it was, and that from a
  feature branch the commit lands on **that branch**, in that worktree.
- `sdlc issue sync --issue N --push` publishes.
- `sdlc issue sync` takes the repo transaction lock (asserted the way the other
  mutating verbs' lock coverage is — `TestRepoLockCommandMetadata`).
- `gitx.IsShippedWorkSubject` classifies the new subject as bookkeeping, pinned
  by a `window_test.go` table row, and `window.go`'s stale "excluded for free"
  parenthetical no longer claims sync commits carry no `#N`.
- `sdlc change-code` leaves the issue file committed and pushed on success
  (the milestone case), and warns-and-continues rather than dying when the sync
  cannot complete.
- No second sync implementation exists: `sdlc issue sync` and `change-code`
  both route through `syncIssuesToMain`, and `commitUntrackedIssueFile` is
  gone.

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
design-buffer: 0.15
item: cross-cutting-refactor  design=0.10 impl=0.20
item: smaller-go-module       design=0.05 impl=0.12
item: smaller-go-module       design=0.05 impl=0.16
item: smaller-go-module       design=0.00 impl=0.08
item: smaller-go-module       design=0.05 impl=0.20
item: atlas-docs              design=0.05 impl=0.05
item: milestone-review        design=0.00 impl=0.15
total: 1.31
```

What each item covers, in Plan order:

1. `cross-cutting-refactor` — the pathspec sweep across all five commit sites
   (`claim.go` ×2, `push.go` ×2, `merge.go`) plus the `syncInPlace` /
   `syncViaMainWorktree` rename. Mechanical once the class is enumerated, which
   the Revisions section already did.
2. `smaller-go-module` — `sdlc issue sync`: cobra wiring, flags, `runIssueSync`,
   helptext. Mirrors `issue new`/`issue validate`, so the existing verbs are the
   spec.
3. `smaller-go-module` — `NoPush` + the commit-message parameter threaded
   through `syncIssuesToMain` and both arms, plus the `change-code` call site
   replacing `commitUntrackedIssueFile`.
4. `smaller-go-module` — `issue-sync` into `gitx.bookkeepingVerbs`, the stale
   comment, and the `window_test.go` table row. Pure + table-tested.
5. `smaller-go-module` — the test work that has no fixture today: the
   two-worktree (bare origin + main + feature) harness for
   `syncViaMainWorktree`, and the swept-index regressions on both arms.
6. `atlas-docs` — `helptext/issue.md` and the atlas entry.
7. `milestone-review` — one close-boundary review; this is single-pass work with
   one boundary, so one review, not one per milestone.

Design is `×0.2` spec-quality discounted throughout: the Spec resolves the
dispatch shape, the flag polarity, the message text, the bookkeeping-verb
decision and the fixture choice, so the remaining design is reading rather than
deciding. Buffer is `+15%` on that basis (v3.1 step 4). Familiarity stays at the
`1.0` default: the stack is familiar (Go + cobra, this repo), which is what Step
5 multiplies. The two genuinely first-time-here pieces — `--only` pathspec
semantics and the two-worktree fixture — are priced inside item 5's impl at the
scaled ceiling rather than smeared across every item by a global multiplier.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against
`baseline-v3.1.md`. Method A only.*

## Plan

- [x] Sweep the narrowed-add/bare-commit class in one pass: `syncOnMain` and
      `syncOnBranch` commit with the pathspec they staged, and
      `archiveCommitArgs(msg, moves)` joins `archiveAddArgs` so `push.go`'s two
      call sites and `merge.go`'s do the same. Regression test per sync arm
      (real repo; the branch arm gets its two-worktree fixture).
- [x] Thread a commit-message parameter through `syncIssuesToMain` /
      `syncOnMain` / `syncOnBranch`; existing callers pass the empty sentinel
      and keep today's per-path message verbatim.
- [x] Add `NoPush` to `claimFlags` (zero value = today's push) and rename the
      arms to `syncInPlace` / `syncViaMainWorktree`, with the dispatch
      discriminating on no-push first. `claim` / `issue new` unchanged.
- [x] `issue-sync` joins `gitx.bookkeepingVerbs`; correct the stale "excluded
      for free" comment; pin `#N: issue-sync: …` with a `window_test.go` row.
- [x] `sdlc issue sync --issue N [--push]` as a `markMutatingCommand` over that
      helper, registered in `TestRepoLockCommandMetadata`'s mutating list.
- [x] Call it from `change-code` with push on, *replacing*
      `commitUntrackedIssueFile` so the verb ends with exactly one issue-commit
      mechanism (ARCH-DRY); warn-and-continue on failure.
- [x] `sdlc issue --help` documents the verb and when to reach for it.

## Revisions

### 2026-09-02 — scope: sweep the class, not the site

Running the class lens (`ARCH-PURPOSE`) over the bare-commit bug turned the
Spec's single site into a five-site enumeration. Every member is "narrow the
`git add`, then commit the whole index".

| site | swept by |
| --- | --- |
| `claim.go:179` `syncOnMain` | the Spec's named site |
| `claim.go:311` `syncOnBranch` | identical defect, identical path. Its `mainHasUncommittedIssueChanges` precheck covers only *issue* files, so a peer's staged non-issue file in the main worktree is still swept |
| `branchcreate.go:117` `commitUntrackedIssueFile` | deleted outright — piece 3 subsumes it |
| `push.go:204`, `push.go:370` | `archiveAddArgs` stages precisely; the commit does not |
| `merge.go:550` | same, and the live one |

**The first cut of this section excluded the archive commits on a rationale the
code contradicts, and the plan-quality gate caught it (PQ-5).** For the record,
because a wrong exclusion closes a real gap: `sdlc push` refuses only on
*untracked* files (`push.go:99-106`) and then **auto-commits** tracked dirty
content with `git commit -a` (`push.go:109-119`) — it does not refuse on a
non-empty porcelain, as claimed. And `sdlc merge`'s `assessDirty` classifies
tracker paths as non-blocking by construction (`merge.go:155-166`) *and* checks
the branch worktree while the archive commit runs in the **main** worktree
(`merge.go:547-551`) — so `merge` has no precheck at all standing between a
foreign staged file and that commit. Both are in scope now.

`commitUntrackedIssueFile` is deleted rather than left beside the new sync: it
is `change-code`'s private commit+push of the issue file, so keeping it would
leave two implementations of one thing. The replacement also handles the
tracked-but-modified issue file the old helper ignored.

### 2026-09-02 — correction: the publish arm's sweep is a race, not a live bug

Implementation disproved this section's own claim that `claim.go:311` carries
"the identical defect". `syncViaMainWorktree` runs `git pull --rebase origin
main` in the main worktree (step 4) *before* it copies and commits, and git
refuses that outright when the index is dirty:

    error: cannot pull with rebase: Your index contains uncommitted changes.

So a peer's staged file makes the whole sync fail loudly, long before the bare
commit could sweep it. Reproduced by hand against a real two-worktree repo,
which is also how the first draft of the branch-arm regression test failed —
the test could not construct the state it was asserting about.

The pathspec on that arm stays. It is correct, it makes the two arms consistent,
and it closes the pull→commit window where a peer *could* stage something. But
it is defense-in-depth, not a live bug, and the tests say so: the deterministic
swept-index regressions are on `syncInPlace` and on `push`'s archive commit
(neither has a dirty-index guard), while the branch arm gets an end-to-end
real-git test plus an argv assertion that it passes a pathspec at all. Proving
git's `--only` semantics once against real git and then proving the wiring
per-site is the honest split; asserting argv everywhere would prove nothing
(`ARCH-MOCK`).

Same lesson as PQ-5, one level down: check whether a guard already stands
between the code and the failure before asserting the failure is reachable.

### 2026-09-02 — where the sync sits inside `change-code`

After the whole gate sequence, at the point `commitUntrackedIssueFile`
occupied — not between plan-quality and the estimate gates. The gate list is a
read-only refusal sequence with two guards riding on it
(`TestGateOrderPlanBeforeEstimate`, and `TestForceAckMatchesGateCatalog`, which
requires every gate name to have a force-bypass ACK pattern), so a mutating
member would have to pretend to be force-bypassable. The `## Estimate` block is
also still being written by the estimate gates, so an earlier sync would commit
a half-finished record and still need a second one. Mid-planning durability is
what `sdlc issue sync` is for — the Spec says so.

## Log

### 2026-09-02

**Gates.** plan-quality BLOCKED at round 1 with seven findings, all verified
real against the code before acting on them; round 2 CLEAN. PQ-5 was the
valuable one — the first Revisions draft excluded the archive commits on a
rationale (`sdlc push` refuses on a dirty tree) that `push.go:109-119`
contradicts, so the enumeration would have closed a live gap in `merge`.
estimate-quality returned INFO at 1.31h.

Two of its notes were folded in: items 2 and 3 swapped impl weights (the verb
mirrors existing verbs; the param-threading carries three Plan rows), and the
`familiarity: 1.0` rationale no longer argues against itself — the two novel
pieces are priced inside item 5 rather than by a global multiplier. Three notes
were read and left: `smaller-go-module` is the closest slug for a first-of-kind
test harness with no test-fixture primitive in the vocabulary (flagged so the
calibration row is not read as clean), the total assumes a clean close review,
and the ledger's under-estimate cluster for multi-site + heavy-real-git-test
work is a real risk the close will measure rather than something to pad now.

**Coverage for the archive sites.** `archiveCommitArgs` is pure, so it is
table-tested directly, and `push.go`'s in-place archive gets a real-repo
regression. `merge.go:550`'s `GitInDir(mainPath, …)` variant is covered by that
shared helper plus the `syncViaMainWorktree` two-worktree fixture exercising the
same "commit in another worktree with a pathspec" shape, rather than a third
bespoke harness.


Found while auditing whether concurrent agents can safely update issue files —
the question came from adopting an advisor/actor split where a brain session
designs and a per-repo actor implements, so two agents touch one repo's
`workshop/issues/` by arrangement rather than by accident.

**Boundary condition, recorded not designed for.** "Publish the reservation,
keep the details private until a milestone" is operator discipline, not an
enforced property, and it works because one operator initiates every actor and
therefore already knows what each stub is for. With more than one person it
degrades: an empty stub on `origin` cannot tell a peer whether it is abandoned,
actively being designed, or a placeholder, and two people can design the same
issue in parallel without noticing. The fix at that point is a richer *stub*
(intent line, owner, timestamp) rather than more locking — but building that now
is the over-generality trap that cost `pair` its couch overrun, so it is written
down and not built.

The audit's other findings were deliberately dropped, recorded here so they are
not rediscovered as gaps: `issue.NextID` scans local files with no fetch (moot
on one host — the lock is on the git common dir, so linked worktrees
serialize), and `syncOnMain` never rebases before pushing while `syncOnBranch`
does (`cmd/sdlc/claim.go:243`). The latter fails loudly at push rather than
corrupting anything, and on a single machine a stale local `main` is rare; it
is worth revisiting only if the fleet ever spans clones.
