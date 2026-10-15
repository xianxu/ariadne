---
id: 000374
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: 'bffe278743a80419676f76a97c18d1c56dd71641' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge's final cleanup fails inside the agent sandbox on git's shared config lock (kink 52)

## Problem

`sdlc merge`'s final cleanup (after landing) fails inside the agent's sandbox on git's shared config lock (`.git/config.lock`), seen 3+ times (TL kink 52); slots rerun the printed recovery command unsandboxed. Retro item (e); not covered by ariadne#370, which landed first.

## Spec

Find which cleanup step writes git's shared config (branch deletion with upstream config, worktree prune, `git config` edits) and either avoid the shared-config write in the sandboxed path (e.g. no branch config to remove, or defer it to the next resting-branch command), or make the step idempotent and resumable so the next sdlc run completes it without the operator. No sandbox widening.

## Done when

- A full `sdlc merge --yes` inside the default agent sandbox completes cleanup without a manual rerun (live check on a scratch issue), and an interrupted cleanup completes on the next run (test).

## Plan

- [ ]

## Log

### 2026-10-14
