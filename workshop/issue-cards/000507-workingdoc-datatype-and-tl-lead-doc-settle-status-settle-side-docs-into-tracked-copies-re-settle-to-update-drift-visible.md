---
id: 000507
status: working
created: 2026-10-20
updated: 2026-10-21
estimate_hours: 2.93
github_issue:
started: 2026-10-21T12:19:30-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
---

# workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible

## Problem

Side docs (the operator–TL co-edited pages beside the session file) have become the main design medium, but they're
untracked, so settled decisions can be lost, and projects can't link them durably. Operator, 10-20: make a
`workingdoc` type; copy a settled side doc into a tracked place; projects link it as background. Concern raised:
once copied, the copy freezes while discussion continues in the side doc.
