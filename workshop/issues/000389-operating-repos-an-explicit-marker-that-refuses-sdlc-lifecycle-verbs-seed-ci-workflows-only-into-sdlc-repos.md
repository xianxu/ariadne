---
id: 000389
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: '6732898557fbbfcc28d4f6f6bcb47f8e14045d6f' # card fields mirrored from issue-cards; edit via sdlc
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

## Spec

- **An explicit operating-repo marker** (e.g. `.sdlc/operating` or a field in the repo's ariadne config): sdlc's
  lifecycle verbs refuse there with a clear message ("<repo> is an operating repo: sdlc lifecycle verbs are not
  available; use the product repos"), while reads and `project find/status/tick` keep working.
- **Workflow seeding by repo kind:** sdlc repos (those with `workshop/issues/` and `.sdlc/`) get `local-gates-guard`
  (and merge-check only if it checks sdlc invariants such as issue-ID collisions and weave drift, not generic CI);
  operating repos and brain get none. Seeding never overwrites a repo's own workflows.
- Document the repo kinds (product/sdlc, operating, brain) in the constitution's peer-repo section.

## Done when

- In a repo with the operating marker, `sdlc claim` refuses with the operating-repo message; `sdlc project status`
  works. Test with a temp repo.
- `weave refresh` in an operating repo emits no workflows; in an sdlc repo it emits local-gates-guard. Tests over the
  manifest.
- ops gets the marker (TL adds it after this lands).

## Plan

- [ ]

## Log

### 2026-10-15
