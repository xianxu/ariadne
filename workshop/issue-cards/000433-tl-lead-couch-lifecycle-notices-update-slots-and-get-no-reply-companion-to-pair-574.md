---
id: 000433
status: working
created: 2026-10-17
updated: 2026-10-17
estimate_hours:
github_issue:
started: 2026-10-17T19:05:11-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead: Couch lifecycle notices update Slots and get no reply (companion to pair#574)

## Problem

pair#574 makes Couch send the hub a system message for each slot lifecycle event (start, park, relaunch, exit, recovery, plus one summary after a Couch restart). tl-lead help does not say how the hub treats them.
