---
id: 000411
status: working
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
started: 2026-10-16T15:06:50-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
---

# start-plan's weave refresh fails the first run in a Claude sandbox (settings.json is write-denied); a stale settings file goes unnoticed

## Problem

In a Claude slot (parley.nvim:4, 2026-10-16), every first `sdlc start-plan` fails: its weave refresh tries to write
`.claude/settings.json`, which Claude Code's sandbox write-denies (`operation not permitted`; settings files are in the
sandbox's deny-within-allow list by design, so an agent can't rewrite its own permissions). Re-running start-plan
then goes through, so the slot's `.claude/settings.json` may lag the woven one without anyone noticing.
