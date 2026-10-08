---
id: 000303
status: open
deps: []
github_issue:
created: 2026-10-08
updated: 2026-10-08
estimate_hours:
card_mirror: '0e82656033603f789dde5c37cd26b6e62f9549c2' # card fields mirrored from issue-cards; edit via sdlc
---

# Catch the slot and its dependencies up to main before new work

## Problem

A slot (`:1+`) starts new work from whatever its checkout and its substrate dependencies were when last touched. After a merge, the slot's resting branch stays behind `main` ("workspace retained on unchanged main-slotN"). The slot's `sdlc` (built from the checkout) and its dependency repos (read live through `construct/deps`) are stale too. Observed after #284 landed: `issue new` on the slot still printed #283-era guidance, because the slot's resting branch, and the binary built from it, hadn't caught up. Today the operator remembers `weave refresh` per slot; nothing enforces it.

Operator idea (2026-10-08): the very start of new work should check that the slot's repo and its dependencies are clean and can fast-forward to their `origin/main`, and catch them up.

## Spec

To settle in design:
- **Where the check lives:** `claim` already fast-forwards a resting branch after claiming (#284). Starting work is `start-plan` (it creates the branch at fresh main). Candidates: at `claim`, at `start-plan`, or both. Plus a dependency sweep.
- **Dependencies:** walk `construct/deps` (start-plan's `substrateChain` already does, for its contention line). For each dependency's checkout:
  - clean and fast-forwardable → fast-forward;
  - dirty or diverged → refuse or warn with the next action.
  Decide refuse vs warn: a stale dependency silently changes what this slot builds against.
- **The binary:** the `sdlc` that runs is built from the checkout, so catching the checkout up changes the next invocation's binary. Decide whether start-plan re-execs or just reports "rebuilt; rerun".
- **Relationship to `weave refresh`:** reuse its refresh where it overlaps (ARCH-DRY) rather than a parallel catch-up.

## Done when

- Starting new work in a `:1+` slot whose resting branch or dependencies are behind `origin/main` catches them up (fast-forward) when clean, and otherwise stops with a next action naming what is dirty or diverged.
- A fixture with a stale slot and a stale dependency shows both caught up before the issue branch is created.

## Plan

- [ ] Design: where the check lives (claim / start-plan), refuse vs warn for dependencies, reuse of `weave refresh`.

## Log

### 2026-10-08
