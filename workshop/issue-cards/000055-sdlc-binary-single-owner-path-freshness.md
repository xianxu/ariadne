---
id: '000055'
status: wontfix
created: 2026-05-31
updated: 2026-06-03
estimate_hours: 3
---

# sdlc binary: single-owner build + PATH aggregation + freshness

## Problem

`sdlc` is the **only** binary in the ariadne ecosystem that gets duplicated
across repos. Every other tool is single-owner (nous owns `nous`/`gmail`/
`oneshot`; pair owns `pair`/`pair-wrap`/…; charon owns `charon`; gstack owns its
60+ scripts — one copy each, in the owning repo's `bin/`). But `sdlc` is built
*per derivative* via `construct/go.mod` (`replace => ../../ariadne`), so the same
ariadne source compiles into N independent `bin/sdlc` binaries, each with its own
staleness clock.

Snapshot taken 2026-05-31 — three different versions live on one machine:

```
ariadne      May 31 12:42   4842882   ← current (post-#51)
you-decide   May 31 14:28   4842882   ← current
nous         May 31 10:16   4826338   ← stale
pair         May 28 10:34   4826338   ← stale
parley.nvim  May 27 10:27   4773634   ← stale-r
```

This is the vestige of the abandoned "derivative is self-sustained / builds the
substrate itself" framing. We moved to peer-symlink + `./bootstrap.sh` (clone
base-layer deps as siblings); under that framing each base layer should **own**
the build of its own binaries and derivatives should not rebuild them. `sdlc`
just never got migrated to match.

Surfaced concretely by the #51 dogfood (#53 Phase B): you-decide's `bin/sdlc`
was a month stale (pre-#51) and died on the in-place merge with `find main
worktree: could not find a worktree on branch 'main'`. `pair` and `parley.nvim`
are pre-#51 *right now* and would hit the same failure.

The PATH-aggregation direction is already declared — `Makefile.workflow`'s
`sdlc-install` comment states the `~/bin` symlink approach was retired "so all
repo `bin/` dirs compose uniformly on PATH." This issue completes that decision
for `sdlc` and adds the freshness guarantee single-ownership still needs.
