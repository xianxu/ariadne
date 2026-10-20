---
id: 000481
status: done
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 3.03
github_issue:
started: 2026-10-20T01:19:56-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.83
tracker:
    version: 1
    completion:
        token: close-7ef536b6b85f
        repository: github.com/xianxu/ariadne
        reviewed_head: 46b5c27226615ada30a2c22f1e7284fea2893c23
        evidence_commit: 3f4c99d2d3a81cd62fe0b9e9d6a44815f9706aa4
        landed_commit: 7836225efe51d963adb901c31181d52d0b76e415
---

# sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain

## Problem

A fleet drain before a machine or Couch restart (pair counterpart: `couch --drain [--fast]`) needs the machine test
queue and the landing lines to stop admitting work, and workers need a rule for the drain notice. Operator item 15
(10-19), discussion: ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-15-fleet-drain.md`.
