---
id: 000507
status: codecomplete
created: 2026-10-20
updated: 2026-10-22
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
actual_hours: 2.87
tracker:
    version: 1
    completion:
        token: close-3e25496f2b04
        repository: github.com/xianxu/ariadne
        reviewed_head: 0d41e87788e51ba8bd7af58c052bf11042500ebc
        evidence_commit: b37a90bb58cd866ef37b6d30bbe008f4aabf4fc7
---

# workingdoc datatype and tl-lead doc settle/status: settle side docs into tracked copies, re-settle to update, drift visible

## Problem

Side docs (the operator–TL co-edited pages beside the session file) have become the main design medium, but they're
untracked, so settled decisions can be lost, and projects can't link them durably. Operator, 10-20: make a
`workingdoc` type; copy a settled side doc into a tracked place; projects link it as background. Concern raised:
once copied, the copy freezes while discussion continues in the side doc.
