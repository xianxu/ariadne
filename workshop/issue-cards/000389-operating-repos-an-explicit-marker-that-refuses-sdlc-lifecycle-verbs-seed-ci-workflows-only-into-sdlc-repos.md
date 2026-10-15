---
id: 000389
status: open
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
---

# Operating repos: an explicit marker that refuses sdlc lifecycle verbs; seed CI workflows only into sdlc repos

## Problem

ops (the TL's operating repo) layers on ariadne for skills and AGENTS, but it isn't an sdlc repo: the operator decided
on 2026-10-15 that ops runs no sdlc lifecycle (brain-like in that sense), only `sdlc project find/status/tick` reads
and project ticks. Two gaps follow:
1. Nothing marks a repo as "operating": the spine guard refuses lifecycle verbs only in brain (`.brain/config.md`) or
   when `workshop/issues/` is absent, with a generic message.
2. `construct/base.manifest` **seeds** `.github/workflows/merge-check.yml` and `local-gates-guard.yml` into every
   dependent repo, so `weave refresh` keeps regenerating them in ops, where merge-check would run on every direct push
   to main. The operator: ariadne shouldn't impose CI on dependents; only the workflow that protects sdlc's own gate
   settings (`local-gates-guard`, the server half of the local-gates trust model) belongs in repos that use sdlc's
   gates.
