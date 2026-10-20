---
id: 000481
status: codecomplete
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
actual_hours: 3.14
tracker:
    version: 1
    completion:
        token: close-779bd455a346
        repository: github.com/xianxu/ariadne
        reviewed_head: 887279de149e943bdf664d9a1ad35d34ec2d91e2
        evidence_commit: a3e1bb123b6e27513ec8321046f8e6034eb6dc59
---

# sdlc queue pause/resume (+ --cancel-running) and the tl-worker drain rule, for couch --drain

## Problem

A fleet drain before a machine or Couch restart (pair counterpart: `couch --drain [--fast]`) needs the machine test
queue and the landing lines to stop admitting work, and workers need a rule for the drain notice. Operator item 15
(10-19), discussion: ops `teams/devinfra/tl-status-xian-xu.side/2026-10-19-15-fleet-drain.md`.
