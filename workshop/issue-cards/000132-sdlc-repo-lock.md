---
id: '000132'
status: done
started: 2026-06-27T11:24:51-07:00
created: 2026-06-26
updated: 2026-06-27
estimate_hours: 6.24
actual_hours: 1.02
---

# sdlc repo transaction lock

## Problem

`sdlc` mutating commands can race when multiple agents or tool calls run in the same repo. In pair#72, four concurrent `sdlc issue new` invocations contended on `.git/index.lock`, produced partial issue-sync commits, and raced on `origin/main` pushes.

Git does not provide a general "wait for index lock" semantic for `git add` / `git commit`, and waiting on `.git/index.lock` would still be too narrow: an SDLC transaction includes issue ID allocation, file writes, git add/commit, and push. The lock needs to cover the whole SDLC transaction, not just Git's internal index mutation.
