---
id: 000520
status: wontfix
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T02:45:18-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
tracker:
    version: 1
    abandoned:
        ref: refs/ariadne/abandoned/000520
        branch: 000520-tl-lead-one-confirmation-round-trip-before-acting-on-an-irreversible-request-submitted-remotely
        head: a6b6d9eec660062468538b545d0d59b660691a6e
---

# tl-lead: one confirmation round-trip before acting on an irreversible request submitted remotely

## Problem

remote-session (ops project, operator-approved 10-20): a remote instruction can make the TL do anything it could do
locally; least capability limits the API, not the agents.
