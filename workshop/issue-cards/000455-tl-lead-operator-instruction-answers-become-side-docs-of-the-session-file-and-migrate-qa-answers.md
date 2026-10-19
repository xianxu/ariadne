---
id: 000455
status: open
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T21:13:34-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# tl-lead: operator-instruction answers become side docs of the session file (and migrate qa/ answers)

## Problem

The hub answers each Operator instructions bullet in a separate doc under `teams/<team>/qa/<operator>/`, linked from the bullet with `tl-lead status instructions link`. parley's attention finder lists an entry's children by scanning its `.side/` folder for `type: side-doc` files, so these answers never appear in the finder (operator, 2026-10-18; ops qa 2026-10-18-24-finder-answer-docs.md). The two conventions never met. The operator: "there might even be a task to migrate from qa/ to xxx.side/ folder for TL's session file?"; none existed.
