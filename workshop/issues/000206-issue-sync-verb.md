---
id: 000206
status: working
deps: []
github_issue:
created: 2026-09-02
updated: 2026-09-02
estimate_hours:
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
supplies a commit message and reuses the dispatch.

The commit message should name the issue the way the rest of the tree does:
`#206: issue: sync spec/plan` rather than the generic `issue-sync: update
issues`. Message shape is the only thing the new verb adds over the existing
helper, so keep it a parameter of the helper rather than a branch inside it.

**3. `change-code` calls it with `--push`.** `change-code` is already the gate
that sits exactly where planning ends — it runs plan-quality and then asks for
the estimate — which makes it precisely one of the milestones at which the
details *should* become external. Syncing there means the design lands durably
and is published at the natural boundary with no new ritual, while the bare verb
from (2) covers mid-planning checkpoints. One mechanism, two triggers, rather
than a parallel auto-sync that could drift from the verb.

**Out of scope:** fetch before `NextID`, conflict detection, merge strategy,
any locking of body edits.

## Done when

- `syncOnMain` commits with a pathspec; a test stages an unrelated file, runs a
  sync, and asserts the unrelated file is still staged and absent from the
  commit.
- `sdlc issue sync --issue N` commits that issue's files with a message naming
  the issue, from both `main` and a feature branch, and **does not push**.
- `sdlc issue sync --issue N --push` publishes; a test asserts the default does
  not reach `origin`.
- `sdlc issue sync` takes the repo transaction lock (asserted the way the other
  mutating verbs' lock coverage is).
- `sdlc change-code` leaves the issue file committed and pushed on success
  (the milestone case).
- No second sync implementation exists: `sdlc issue sync` and `change-code`
  both route through `syncIssuesToMain`.

## Plan

- [ ] Pathspec fix in `syncOnMain` + regression test for the swept-index case.
- [ ] Thread a commit-message parameter through `syncIssuesToMain` /
      `syncOnMain` / `syncOnBranch`; existing callers pass today's message.
- [ ] `sdlc issue sync --issue N` as a `markMutatingCommand` over that helper.
- [ ] Call it from `change-code` after the plan-quality gate passes.
- [ ] `sdlc issue --help` documents the verb and when to reach for it.

## Log

### 2026-09-02

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
