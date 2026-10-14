---
id: 000366
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '60aab92fae65d64866316e22c8b06d5cb67e79c0' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc local-check: a slot whose ariadne substrate predates the weighted lock fails with a raw usage error; detect the version skew and name 'weave refresh'

## Problem

After ariadne#332 landed (2026-10-14), the shared `sdlc` (`~/workspace/ariadne/bin/sdlc`) passes `--slots N` to the slot's `scripts/test-lock.py`, which comes from the slot's own ariadne substrate (`worktree/<repo>-slotN/ariadne`). Slots whose substrate predates ariadne#346 (pair:2, :4, :5, parley.nvim:2, :3 on 10-14) fail `sdlc local-check run --name suite` with a raw argparse usage error. pair:5 hit it re-closing pair#547. The fix is `weave refresh`, but nothing says so.

## Spec

Before invoking a substrate helper, `sdlc` checks the helper's capability (e.g. a version or capability line, or `--help` for `--slots`) and, on mismatch, refuses with one line naming the slot's substrate path and its commit, and the action `weave refresh` in the checkout. Same for other helpers sdlc calls with newer flags. No automatic refresh (it changes the slot's substrate under a running agent).

## Done when

- With a stale substrate fixture, `sdlc local-check run --name suite` refuses before queueing with the 'run weave refresh' message (test).
- A current substrate runs unchanged.

## Plan

- [ ]

## Log

### 2026-10-14
