---
id: 000473
status: codecomplete
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.37
github_issue:
started: 2026-10-19T20:26:49-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.13
tracker:
    version: 1
    completion:
        token: close-01527b57c142
        repository: github.com/xianxu/ariadne
        reviewed_head: a3b1a8cb593559d6254cf0df09b0833e4fafcb46
        evidence_commit: bd94a28252cf61e9fadde9c7a3aff203b4f113a5
---

# test-economy: affected-suite as close evidence; the full suite runs once, in the train or at merge (option B')

## Problem

A change pays the full suite twice today: once on its branch (close evidence), and again at landing when main has moved
with code (the integrated suite), or in a train (the train suite). Measured 10-19 (`sdlc test-stats`, 7 days): pair suite
12.5 min median, slot wait 11 min median / 35 min p90, ready → landed 54 min median / 2.3 h p90; ariadne's suite takes both
slots. The machine is suite-bound. Discussion and the operator's decision: ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-11-landing-schedule.md` (option B′, approved 10-19).
