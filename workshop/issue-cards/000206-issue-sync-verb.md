---
id: '000206'
status: done
started: 2026-09-02T11:36:01-07:00
created: 2026-09-02
updated: 2026-09-02
estimate_hours: 1.31
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
