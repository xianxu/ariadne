---
id: 000221
status: open
created: 2026-09-11
updated: 2026-10-09
estimate_hours:
github_issue:
started: 2026-10-09T18:40:42-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
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
