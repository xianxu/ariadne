---
id: 000206
status: done
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.31
started: 2026-09-02T11:36:01-07:00
actual_hours: 4.43
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
mid-planning.

**Operating envelope (`ARCH-CONSTRAINTS`).** The default path is one `git add`
and one `git commit` over a single issue file: sub-second, **zero network**, no
worktree hunt, no subprocess beyond git. That is the budget a verb the
constitution tells agents to run after every design move has to fit in, and it
is why the network lives behind `--push` rather than in the default. It also makes the verb usable from an in-place feature branch,
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
  `archiveCommitArgs`, which refuses an empty move list rather than silently
  emitting the whole-index form.
- `migrate.go`'s two commits (source + destination) carry a pathspec, and so do
  the `--no-commit` hints it prints. The source side is where nothing guards it.
- `sdlc change-code` commits the issue file **in the current worktree** across
  the whole cross-product it runs under: `resolveBranchName`'s three name modes
  × {on main, in-place feature branch, feature worktree} × {file untracked,
  tracked-and-edited}. It publishes only when already on `main`; from a branch
  the body must not reach `origin/main`. Run as a table, not as cells.
- `sdlc issue new` leaves the new issue committed even when it cannot be
  published — publication and durability are separable, and an untracked new
  issue is the hole this issue exists to close.
- A source-level guard fails the build when any `git commit` argv in
  `cmd/sdlc` is narrower in its add than in its commit. An exemption requires
  BOTH an allowlist entry with a reason AND a `git add -A` in the same function,
  so it cannot widen to a sibling commit.
- Every call site fixed here is pinned either by a test entering the production
  entry point, or — where the entry point is not in-process drivable — by a
  source-level wiring guard that fails when the call is deleted.
- `sdlc issue sync --issue N` commits that issue's files with a message naming
  the issue, from both `main` and a feature branch, and **does not push**: a
  test asserts the default leaves `origin/main` where it was, and that from a
  feature branch the commit lands on **that branch**, in that worktree.
- `sdlc issue sync --issue N --push` publishes — **including a body that is
  already committed**, which is the state a prior no-push sync (or a sync whose
  push failed) leaves behind, and the state both warning messages tell the
  operator to recover from.
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
- [x] Rework after the close review (BR-1..BR-4): derive change-code's issue id
      from the resolved path; sweep `migrate.go`; make `archiveCommitArgs`
      refuse empty moves; deliver the mid-planning trigger through
      `start-plan`'s output + `AGENTS.base.md` §2/§14.
- [x] Rework round 2 (BR-13, BR-14): a source-level pathspec guard over
      `cmd/sdlc/*.go` that no future site can escape; `syncIssue` commits where
      the branch is, proven by the full mode matrix; the same durability
      fallback in `issue new`.
- [x] Rework round 3 (BR-18, BR-19, BR-13 residual): publication no longer gated
      on working-tree changes; the guard's exemption needs its `add -A`;
      `TestVerbsWireTheirCommitHelpers` pins the call sites no behavioral test
      can reach.
- [x] Rework round 4 (BR-18 still true on the other arm): the publish arm routes
      the body instead of pushing an unchanged main; `syncInPlace` publishes only
      when main is ahead of origin; `TestIssueSync_PublishMatrix` covers the
      table that would have caught both.
- [x] Rework round 5 (BR-26): a body main already carries is dropped before
      conflict detection, so publishing twice is a no-op instead of a false
      conflict; the publish-existing intent is an explicit flag set only by
      `issue sync --push`, replacing the heuristic that made every clean-tree
      `claim` a wholesale push.

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

### 2026-09-02 — rework after the close review (BR-1..BR-4)

The boundary review returned REWORK on four blocking findings, all verified
against the code before acting.

**BR-1 (Critical) — the replacement covered fewer modes than the thing it
replaced.** `syncIssue` gated on `f.Issue > 0`, but `resolveBranchName` has
*three* name-resolution modes and only one sets `--issue`: `--name` derives the
path from the branch name, and the third auto-detects the single untracked issue
file. `commitUntrackedIssueFile` fired in all of them. Worse, in auto-detect mode
with `--worktree=yes` the created worktree ended up with **no issue file at all**
— `git worktree add` does not carry untracked files — so every later verb that
resolves the issue by file would fail there. Fixed by gating on the resolved
`issuePath` (which every mode produces) and deriving the id from it via
`issueIDFromPath`. The remaining early return is the genuine case: a `--name`
branch pointing at a file outside the `NNNNNN-` convention.

This is the second time in this issue that a "the class is swept" claim was
wrong in the same direction — asserting a property without checking every member
of the enumeration it quantifies over. Both were caught by a gate, not by me.

**BR-2 — the class had a sixth member.** `migrate.go` stages one explicit path
per side and then commits bare in both repos, and its source side has *no*
whole-repo guard: step (1) checks `status --porcelain -- relPath`, the migrated
file only. Its `--no-clean-check` bypass message even promises "staging is
explicit-path", the exact property the bare commit removed, and the
`--no-commit` hints handed the operator the defective form. All four fixed.

**BR-3 — the guard helper's own degenerate input.** `archiveCommitArgs(msg, nil)`
returned `commit -m msg --`, which git reads as no pathspec and commits the whole
index — the failure the helper exists to prevent. It now errors, matching
`syncPathspec`'s posture. Unreachable today (every caller guards `len(moves) >
0`), which is why it needed pinning rather than trusting.

**BR-4 — the trigger had no delivery point.** The verb's whole justification is
durability *during* planning, but the only place saying "run this whenever the
Spec/Plan/Log has moved" was `sdlc issue sync --help`, which an agent reaches
only by already knowing to look. That is shipping the mechanism and calling the
purpose done (`ARCH-PURPOSE`). Now delivered where the agent is already reading
about planning: a `syncPointer` line in `sdlc start-plan`'s output, and
`AGENTS.base.md` §2 (the `claim → start-plan → … → change-code` flow) and §14
(the context-checkpoint rule, whose "update its durable state" is precisely the
moment). `AGENTS.md`/`CLAUDE.md`/`GEMINI.md` are weave-generated entry files —
the edit belongs in `AGENTS.base.md`, and this is a base-layer export, so it
propagates to downstream repos on the next `sdlc propagate-base`.

Minors: the guard comment now describes the case it actually catches (M1);
`DryRun` is threaded into `syncIssue`'s `claimFlags` rather than relying on
`runChangeCode` returning first (M2); `--dry-run` covered on both the verb and
change-code (M4); the partial-commit-during-merge behavior change recorded in the
atlas (M5); `issueFilesForID` now single-sources the "files for issue N" glob
shared by `syncPathspec` and `resolveBranchName` (M7). M3 (no bespoke `merge.go`
regression) stands as the recorded choice already logged. **M6 not taken:** the
change is 9 files, over AGENTS.md §1's durable-plan threshold, but the design
lived in this issue's `## Spec` + `## Revisions` — which is what the plan-quality
gate actually read and blocked on twice. Back-filling `workshop/plans/` after the
fact would produce a copy, not a record.

### 2026-09-02 — rework round 2 (BR-13, BR-14): rules, not cells

Both findings arrived with the same instruction — *this is the Nth finding in
this family; do not fix the instance, state the rule* — and both were right that
the previous round had patched cells.

**BR-13 — the round-1 fixes were not actually pinned.** The reviewer reverted
each one in a scratch clone and ran the suite: `migrate.go`'s two pathspecs,
`push.go`'s two archive commits, `merge.go`'s, and the `start-plan` delivery
line all came back **green**. The cause was uniform and worth naming: my tests
hand-wrote the argv (`r.Git("commit", …, "--", moved)`) or hand-wired the helper
instead of driving `runMigrate` / `runPush` / `runStartPlan`. Real git, real
repos — and none of *our* code on the path. That is `ARCH-MOCK` inverted: a
real-git test that never reaches the production entry point is a mock of the
wiring.

The rule, and the fix that covers the class in one place:
`TestGitCommitsCarryTheirPathspec` parses `cmd/sdlc/*.go` and fails on any git
commit argv built without a `--` separator, with a two-entry allowlist carrying
its reason (`push.go:runPush`'s `commit -a`, `propagatebase.go:commitConsumption`
paired with `add -A` — both legitimately whole-tree, which is the real rule: *a
commit must be as narrow as its add*). Same shape as this tree's existing drift
guards. Verified by reverting four sites at once: four failures. The eighth site
someone writes next year fails immediately, without anyone remembering #206.
`sdlc issue sync` also joins `TestRunStartPlan_RendersAtPlanLens`'s want list —
a test whose own comment says it exists so a dropped line "can't ship silently",
and which had just let one ship silently.

**BR-14 — `change-code` from a feature worktree published WIP to main.** With
push on and a branch checked out, `syncIssue` took the publish route: copied the
in-progress body into the main worktree, committed it on `main`, pushed, and left
the branch's own copy dirty. So the "branch starts from a tracked state" property
was not achieved, `main` carried a divergent commit of a file the branch would
later commit its own version of, and a milestone re-run gained two network
round-trips.

The rule the reviewer asked for is now stated as an invariant with a table under
it: **the issue file ends up committed in THIS worktree, on the branch about to
carry the work** — across three name modes × three locations × two file states.
Publishing is conditioned on already being on `main`, not on the caller's intent;
from a branch, `pr`/`merge`/`close` are what publish. `TestChangeCodeSyncIssue_ModeMatrix`
runs all eighteen cells and goes red on the reverted fix.

**In-class, found by dogfooding rather than review.** Filing #207 from this
branch surfaced the same defect one verb over: `issue new`'s auto-sync could not
reach a worktree on `main`, warned, and left the new issue **untracked** — the
exact hole this issue exists to close, at its own front door. It now falls back
to a local commit and says the broadcast is pending. Fixing only `change-code`
would have been the cell again.

Minors also taken: `syncPathspec`'s doc no longer implies purity (it globs
through `issueFilesForID`); `migrateCommitArgs` single-sources the shape shared
by migrate's two commits and its two `--no-commit` hints, so the printed command
can't lose the `--` the executed one has. The pre-existing red
`TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory` is now **#207** —
it reads an archived plan by hardcoded path, is unrelated to this work, and a
permanently-red suite is what made the reviewer's revert measurements harder.

### 2026-09-02 — rework round 3 (BR-18, BR-19, BR-13 residual)

**BR-18 (Critical) — `--push` could not publish an already-committed body, and
both of this issue's own warnings named it as the recovery.** `syncInPlace`
computed `changedIssueFiles` and returned `[ok] No issue changes to sync.`
*before* the push. That early return was written when committing and publishing
were one act. This issue split them — which makes "committed locally, not yet
pushed" a state the code now deliberately creates, and `changedIssueFiles` is
empty in exactly that state. So `--push` short-circuited, reported success in
green, and left `origin/main` where it was. Both consumers hand the operator
that command: `change-code`'s warning for the common on-main case where the
commit landed and only the push failed, and `issue new`'s after the round-2
fallback deliberately committed locally. The reservation broadcast was
unrecoverable by the route its own message named.

The fix is the rule, not the line: **nothing to commit is not nothing to
publish.** Both arms now skip only the commit, never the publish. Two tests walk
the sequences that would have caught it —
`TestIssueSync_PushPublishesAnAlreadyCommittedBody` (sync local, then `--push`)
and `TestChangeCodeSyncIssue_RetryAfterAFailedPushPublishes` (commit lands, push
fails, clear the cause, run the command the warning printed).

**BR-13 residual — the guard's exemption was keyed by function, so it excused a
sibling.** `runPush` was allowlisted for its `commit -a`, which also excused the
archive commit twelve lines away — the very site BR-13 had named. The exemption
is now two-part: an argv is accepted if it carries `--`, or carries `-a` (which
says whole-tree in the argv itself and needs no entry at all), or sits in a
function that *demonstrably* stages `git add -A` **and** is allowlisted with its
reason. Requiring both halves is what stops an entry from widening. The
allowlist is down to one entry; reverting `push.go:204` now fails, as it should
have.

**BR-19 — the `runChangeCode → syncIssue` call site entered no test.** Deleting
the line left the suite green: every `TestChangeCodeSyncIssue_*` calls the helper
directly. Fourth finding in the same family, so the response is the rule rather
than one more test: *a fix at a call site is pinned only by a test entering the
production entry point; where the entry point is not in-process drivable, a
source-level guard must assert the wiring.* `TestVerbsWireTheirCommitHelpers`
asserts eight such edges (`runChangeCode→syncIssue`, `runPush`/`runMerge`/
`recoverInterruptedArchive`→`archiveCommitArgs`, `runMigrate`→`migrateCommitArgs`,
`runIssueNew`/`runIssueSync`→`syncIssuesToMain`, `runStartPlan`→`syncPointer`),
each with the reason it matters, and errors on a stale edge. It is weaker than
driving the verb — it proves the call exists, not that it runs — and that
weakness is stated where it lives. Landing #191 would let these become real
drives.

Minors: the atlas said "six sites" over an enumeration of seven; `syncPathspec`'s
"unreachable" claim rested on a reason that doesn't cover a *deleted* issue file
(non-empty `changedIssueFiles`, empty glob), so the claim is gone and the error
stands on its own; the mode matrix now says what its 12 duplicate cells buy.

### 2026-09-02 — rework round 4: BR-18 was still true, one arm over

Round 3 named the class correctly — *nothing to commit is not nothing to
publish* — measured the fix on `syncInPlace`, and swept **one of the class's two
members**. The other member shipped inverted, which is worse than the bug it
replaced: `syncViaMainWorktree`'s new empty-changed branch pushed `origin main`
from the main worktree *without carrying the body across*. Main has no new
commits in that state — the body is on the **branch** — so the verb printed
`[ok] Issues synced to main and pushed to origin.` and `synced` while
`origin/main` never moved. A false success on `issue new`'s own advertised
recovery path.

The correction, and the sentence worth keeping: **publication is the gap between
`origin/main` and this worktree's body, never the gap between the working tree
and HEAD.** The publish arm now re-seeds its file list from the *issue* when
nothing is dirty and routes the body through the normal copy → commit → push
flow, committing only if the copy staged something (a pathspec'd commit that
stages nothing is an error, not a no-op). `TestIssueSync_PublishMatrix` covers
{on main, feature worktree} × {body dirty, body already committed} and asserts
the published *content* — comparing SHAs would not catch a push that moved
nothing, which is precisely how this survived a round.

**Round 3's other over-correction:** making the no-changes path push
unconditionally turned an offline `sdlc claim` on a clean tree from `[ok] No
issue changes to sync.` (exit 0) into a fatal error — `claim` is the one caller
that die()s on a sync error, and it is idempotent by design. `syncInPlace` now
publishes only when local main is actually ahead of `origin/main`, so a clean,
already-published tree never reaches for the network.

Minors: `change-code --dry-run` said "Would sync + push" unconditionally while
`syncIssue` publishes only from main — both now read the same
`syncIssuePublishes()`; `recoverInterruptedArchive`'s dry-run hint routes through
`archiveCommitArgs` like migrate's does; the class guard now also matches
`gitx.RunGit`/`gitx.Capture`, the package's other live git seams, which an inline
`gitx.RunGit("commit", …)` would have escaped entirely.

### 2026-09-02 — rework round 5 (BR-26): idempotence

Round 4 was right about *what* to publish and fed it to the wrong question.
Re-seeding the file list from the issue puts a file main legitimately carries
(an earlier run of this verb put it there) in front of a conflict detector that
asks "did both sides touch this since merge-base" — true, and not a conflict. So
the workflow this issue's own docs recommend broke on its second lap:

    sdlc issue sync --issue N        # local checkpoint
    sdlc claim --issue N             # publishes
    sdlc claim --issue N             # Conflict detected! + exit 1

`sdlc claim` had been idempotent since it existed. The fix is one line of rule:
**a body the destination already carries byte-for-byte is nothing to route**, so
it is dropped from the list before the detector is ever asked. That stays inside
the Spec's "conflict detection is out of scope" — it does not resolve conflicts,
it declines to invent one.

**And round 4's own over-correction, undone properly.** `mainAheadOfOrigin` was a
heuristic standing in for an intent nobody had expressed: it made every
clean-tree `sdlc claim` push local main wholesale, publishing bodies a no-push
sync had deliberately kept local, and failing outright offline. Intent is now a
flag — `PublishExisting`, set by `sdlc issue sync --push` and nothing else — so
`claim` and `issue new` keep the no-op they have always had. Explicit intent
beat the inference, and the inference had been mine twice in a row.

**Five rounds, one pattern.** Every round found a real defect, and every defect
was the same shape: a property asserted over an enumeration after checking one
member of it — five commit sites, three name modes, three worktree locations, two
sync arms, one run of a verb that gets re-run. What finally holds isn't the
fixes; it's the two source guards and the three tables, which check the members
for me.

### 2026-09-02 — the publish-state enumeration, and what is left unswept

Round 5 fixed *one member* of an enumeration, and the close review named the
rest. Stated in full, the destination copy of the body can be:

| destination state | behavior |
| --- | --- |
| **absent** | routed and committed — the ordinary first publish |
| **identical** | dropped before conflict detection; publish only (round 5) |
| **older, because this branch published it, then the branch edited it** | **false `Conflict detected!`** — see below |
| **older, from elsewhere** | correctly detected as a conflict |
| **newer** | correctly detected as a conflict |

The third row is a real false positive: `syncOnBranch`'s detector asks "did both
sides touch this file since merge-base", and after a publish-then-edit both
sides have — though the only content on main is the copy this same verb put
there. **It is accepted, not swept**, for two reasons. It predates this issue
(measured at base `92bd1ad`: same detector, same question, reachable the same
way whenever a branch edits a file it previously synced to main), and resolving
it means asking whether main's version is an *ancestor* of the branch's — which
is conflict detection, explicitly out of scope in the Spec's final paragraph.
Round 5's filter stayed on the right side of that line by declining to invoke
the detector for byte-identical content; going further would cross it. The
resolution guide the detector already prints is the correct answer for that row
today.

### 2026-09-02 — the no-push guarantee is about the verb, not the repository

`helptext/issue.md` claimed a half-written Spec "cannot escape", and
`AGENTS.base.md` §2 implied the same. Both overstated it. A local commit is
still a commit: a later `sdlc claim`, `sdlc issue new` or `sdlc push` on main
publishes whatever main carries, this body included. The accurate claim — now
what both documents say — is that **this verb performs no network operation**.
That is the property agents can rely on, and it is the one the default actually
enforces. The consequence is recorded in the atlas beside the MERGE_HEAD
partial-commit note, as a behavior change rather than a footnote.

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
- 2026-09-02: closed — go test ./cmd/... ./pkg/... — green except the pre-existing unrelated TestFleetPlanHasAuthoritativeCorrectedCoreConceptInventory (archived plan path; filed as #207). BR-26 fixed as a rule: a body the destination already carries byte-for-byte is dropped from the route list before conflict detection is invoked (filesDifferingFrom), so the documented workflow survives a second lap. TestPublishIsIdempotent runs local-sync -> publish -> publish -> bare re-sync from both on-main and a feature worktree; verified it goes red on revert (the false conflict die()s the test process, exit status 1). Round 4 heuristic mainAheadOfOrigin removed in favor of explicit PublishExisting intent set only by `issue sync --push`; TestClaimStaysIdempotentOffline pins that a clean-tree claim with an unreachable origin exits 0 without moving main. Minor: change-code --dry-run renders the off-main caveat on its own line, derived from the same syncIssuePublishes() the code branches on.; review verdict: FIX-THEN-SHIP

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

**Close review round 1: REWORK.** Four blocking findings, all real; see the
Revisions entry. The reviewer reverted each claimed fix in a scratch copy to
confirm the regressions actually go red — worth noting because that is the check
that separates a pinned fix from a mirror of itself.

**Close review round 2: FIX-THEN-SHIP,** with two Important findings both
flagged as repeat families — the ledger's `family:` slug reporting, correctly,
that I was fixing instances. Round 1's fixes were measured green-on-revert at 5
of 7 sites. The response was the source-level guard rather than five more tests.
Filed #207 for the pre-existing red fleet-plan test.

**Close review round 3: REWORK,** on a Critical the diff introduced itself —
`--push` unable to finish a publish, named as the recovery by two warnings this
issue added. The pattern across all three rounds is one thing: I kept asserting
properties over enumerations without checking every member (five sites, then
three name modes, then three worktree locations, then "the guard covers the
eighth site"). Each round the gate found the member I hadn't checked. The
durable fix in each case was a rule the tree enforces, not another test.

**Close review round 4: REWORK,** BR-18 disposed `not-addressed` rather than
re-raised — the same claim was still true on the arm round 3 didn't measure. Four
rounds, one pattern, now stated plainly: I keep asserting a property over an
enumeration after checking one member of it. The gate has caught it every time,
and the durable answer each round was a table or a source guard rather than
another fix.

**Close review round 5: REWORK** on idempotence — publishing twice died on a
false conflict. Fixed as a rule (identical content is nothing to route) plus an
explicit `PublishExisting` intent replacing the heuristic round 4 introduced.

**Close review round 6: FIX-THEN-SHIP** — no blocking findings. Fixed under the
protocol before the close commit: every `runIssueSync` drive now goes through
`syncOK` (the natural `if err != nil` branch was dead code, so a refusal killed
the whole package binary with no attribution — the exact signal that made round
5's revert probe read as a bare "exit status 1"); the two "already carries this
body" messages now distinguish their states; `filesDifferingFrom` has a direct
unit; the doc claim and the publish-state enumeration are corrected in
`## Revisions`; the operating envelope is stated in the Spec.

**Estimate vs actual: 1.31 est / 4.43 actual (0.3×).** Six review rounds, five of
them finding a real defect. The estimate priced the design and the code, both of
which were close; it priced one `milestone-review` at 0.15h for a boundary that
took six. Worth carrying into the next estimate for multi-site class work: the
review cost scales with the number of *members in the class*, not the number of
files changed.

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
