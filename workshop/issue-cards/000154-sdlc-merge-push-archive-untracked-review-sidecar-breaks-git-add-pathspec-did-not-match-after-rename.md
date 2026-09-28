---
id: '000154'
status: done
started: 2026-07-05T22:54:11-07:00
created: 2026-07-01
updated: 2026-07-06
estimate_hours: 0.53
actual_hours: 0.64
---

# sdlc merge/push archive: untracked review sidecar breaks git-add (pathspec did not match) after rename

## Problem

`sdlc merge`'s archive step fails its commit when the boundary-review **sidecar**
(`workshop/plans/NNNNNN-<slug>-close-review.md`) is **untracked** at merge time.
The PR itself merges (main advances), but the local archive-to-history commit
dies with:

```
==> Committing archived history in main...
Error: git -C <repo> add: exit status 128
fatal: pathspec 'workshop/plans/NNNNNN-<slug>-close-review.md' did not match any files
```

…leaving main in a half-archived state: the issue file is moved to
`workshop/history/` (untracked) and staged-deleted from `workshop/issues/`, the
sidecar is moved to `workshop/history/` (untracked), but the archive commit is
never made. Recovery is manual: `git add` the destination history paths (+ the
issue rename) and commit.

**Reproduced deterministically 3× this week** — parley.nvim #158, #159, #156.
(Earlier closes #152/#154/#155/#157 archived cleanly, so it is *state-dependent*,
not universal — see the tracked-vs-untracked analysis below for the likely
discriminator.)
