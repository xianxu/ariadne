---
id: 000638
status: open
deps: []
github_issue:
created: 2026-10-25
updated: 2026-10-25
estimate_hours:
card_mirror: 'af3ae464226fe5efc223f52ad3309a2bb19012d2' # card fields mirrored from issue-cards; edit via sdlc
---

# start-plan refuses 'main ahead or diverged' in an extra worktree while main equals origin/main

## Problem

## Spec

**`sdlc start-plan` misresolves main in an extra worktree (ops-echo:0, echo#45, 10-25).** Run from an isolated
planning worktree of echo:0, `sdlc start-plan` refused with "main has commits that are not on main (ahead or diverged);
reconcile them before planning", while `git` showed `main` and `origin/main` both at 2ae6d243aeb8. Nothing changed and
echo:0 wasn't blocked (its plan had started earlier), but the refusal is wrong: start-plan seems to resolve "main"
differently inside an additional worktree (a worktree-local ref, the slot's resting branch, or a stale remote ref).

## Done when

- start-plan in an additional worktree of a slot resolves main and origin/main the same way git does, with a test
  that creates an extra worktree at a clean main and expects no refusal; the refusal names both commits when it does
  fire.

## Plan

- [ ]

## Log

### 2026-10-25
