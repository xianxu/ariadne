---
id: 000393
status: working
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T23:44:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# Serial per-repo landing admission with fleet-blocking priority

## Problem

Concurrent `sdlc merge` calls integrate against the same main; after one lands, the others reintegrate and rerun a roughly nine-minute suite plus queue wait. #376 required three integrations. A fleet-blocking fix (#391) also waited fourth for a suite with no priority lane. #390's operator-approved design chooses serial, per-repo landing admission before any batching.
