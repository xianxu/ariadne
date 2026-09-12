---
id: 000221
status: open
deps: [ariadne#207]
github_issue:
created: 2026-09-11
updated: 2026-09-11
estimate_hours:
---

# sdlc publish and pr report success they did not achieve

## Problem

Three small, self-contained defects, split out of ariadne#220 so they can land
without waiting on that issue's design decision. All three were hit in one
parley.nvim#227 session on 2026-09-10. Each reports success for something that
did not happen, which is the shape that costs the most later: the operator
believes the state is published, and finds out at merge.

**(a) Trunk publishes carry an empty commit subject.** `sdlc issue new` from a
feature branch published three blank templates to parley `main` — `1ead2e2`,
`4905785`, `c955972` — each with no subject at all. `git log --oneline` shows a
bare sha, and `git log --grep` cannot find them. The trunk path passes `msg`
straight through (`claim.go:158` → `synctrunk.go:107`), while the on-main path
defaults it via `syncMessage` (`claim.go:313`). Line numbers are against
`000207-sync-without-worktree` at `671f0f6`.

**(b) A publish that wrote nothing reports "published".**
`sdlc claim --issue 234 --no-start` on a feature branch printed

```
  [ok] Issue changes published to the trunk.
synced
```

while nothing reached `main`: its first-parent line holds no commit between
`4905785` (18:15:53) and `c955972` (18:32:36). `syncViaTrunkWithRealloc` returns
`(nil, nil)` after printing "No issue changes to sync.", and `syncViaTrunk`
prints the success line and the `synced` marker unconditionally
(`synctrunk.go:59-62`). The agent moved on believing the issue was on the trunk;
it was not, and the PR merge later conflicted on that very file.

**(c) `sdlc pr` fails when the PR already exists.** On a branch whose PR is
open, it pushes (the useful half, and it succeeds), then `gh pr create` fails
with "a pull request for branch … already exists" and the verb exits nonzero
(`pr.go:107-115` never checks first). The documented recovery loop for `sdlc
merge` is "fix → commit → push → re-run", and `sdlc pr` is how the push
happens — so the loop's own second step ends in a red error every time.

**Not in scope, deliberately.** The fourth defect from that session — the trunk
path deciding what changed by diffing against `HEAD` rather than the trunk
(`claim.go:336-341`) — stays in ariadne#220. Fixing it alone would make the sync
publish a file the branch has also committed, which is exactly the dual-history
that produced #220's add/add conflicts. It needs that issue's single-writer
decision first.

**Ordering.** (a) and (b) are in #207's unmerged trunk path, hence `deps:`. (c)
is independent and already on `main`. The operator declined folding these into
#207 (2026-09-11): too much disruption to an issue that has already cleared
plan-quality and is mid-implementation.

## Spec

Each verb reports what it actually did.

- **(a)** The trunk path defaults its commit message the same way the on-main
  path does, and the default names the issue ids and the verb — e.g.
  `issue #234: created` or `issue-sync: #227, #234` — so `git log --grep` finds
  a publish by id.
- **(b)** A publish that writes nothing says so and does not print the success
  line. The `synced` marker is a machine-readable contract
  (`synctrunk.go:59-61`), so keep emitting it for callers that parse it, and
  make the human-facing line the truthful half. If a caller needs to tell the
  two apart, give the no-op its own marker rather than overloading `synced`.
- **(c)** `sdlc pr` looks for an open PR for `head→base` first. If one exists it
  prints that PR's URL and exits 0, having pushed. Creating and finding are the
  same outcome for the caller: the branch is published and a PR is open.

## Done when

- No `sdlc` publish can produce a commit with an empty subject; a test asserts
  the subject of the commit a trunk publish lands.
- A no-op publish prints no success line, still emits its machine marker, and a
  test covers both the no-op and the wrote-something paths.
- `sdlc pr` on a branch with an open PR exits 0, prints the existing URL, and
  has pushed; a test covers create, already-exists, and push-failure.
- `git log --grep '#<id>'` on the trunk finds that issue's publish commits.

## Plan

- [ ] (c) `sdlc pr`: find-or-create, exit 0 on an existing PR — independent of #207.
- [ ] (a) trunk-path commit message: default + name the ids.
- [ ] (b) truthful no-op reporting, machine marker preserved.
- [ ] Tests for all three, including the no-op path that (b) exposes.

## Log

### 2026-09-11

Split out of ariadne#220 at the operator's request, after they declined folding
these into #207. #220 keeps the structural half: the dual-write that makes issue
files conflict at merge, and the proposed single-writer design.
