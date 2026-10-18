---
id: 000437
status: codecomplete
created: 2026-10-18
updated: 2026-10-18
estimate_hours:
github_issue:
started: 2026-10-18T04:55:12-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 4.77
tracker:
    version: 1
    completion:
        token: close-cabbd749f45e
        repository: github.com/xianxu/ariadne
        reviewed_head: 67ca24ea443cd3d8628bf7a32aba4980f2bd11b6
        evidence_commit: 22887b2d5f23587489792eec982003d7da6dee43
---

# xx-universal-review edits through couch --session-sync apply; retire deletions: propose-strike

## Problem

Slice 4 of the universal-agent-input-interface project (design: `ops:teams/devinfra/projects/universal-agent-input-interface/design.side/10-s4-edit-path.md`, sections 3, 5 and 8). The shared review skill tells agents to edit documents with `couch --session-sync write` and to hand-write proposal markers under `deletions: propose-strike`. S4 gives edits one path, `apply`, with per-edit reasons, and one knob, `presentation: propose`, for proposals.
