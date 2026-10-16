---
id: 000414
status: working
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
github_issue:
started: 2026-10-16T16:53:29-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead: section freshness stamps, a staleness check, and a derived Slots section

## Problem

The TL session file (`teams/<team>/tl-status-<operator>.md` in ops, written only through `tl-lead status`) goes stale:
on 2026-10-16 its Threads section was ~20 hours old while Needs you stayed current. Nothing tells the TL a section is
stale: updating Threads is a synthesis no single event asks for, and writes blocked by the operator's lock get queued
and forgotten. Operator, 10-16: move this into the hard shell; long-term, only Operator instructions (chat) and Needs
you need to be prose in the doc; the rest becomes command output, a "flexible cockpit", some of it eventually a pure
operations dashboard.
