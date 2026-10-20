---
id: 000482
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T22:48:28-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.32
tracker:
    version: 1
    completion:
        token: close-211982711969
        repository: github.com/xianxu/ariadne
        reviewed_head: ce8324a5be05d9f778e62e4aee498824437e3229
        evidence_commit: 27416f59ac499544c6d3ccbb12a06a33392aaad4
---

# tl-lead: keep a just-landed slot's conversation (next same-area car), and a 'hold: soak until' option for risky changes

## Problem

When a landed change later breaks something, the slot whose conversation wrote it has the freshest context, but
tl-lead has no rule about keeping it: slots get prepared with a fresh conversation, rebooted, or dispatched into
unrelated code right after landing. Operator idea, 10-19 (ops side doc
`teams/devinfra/tl-status-xian-xu.side/2026-10-19-16-landing-queue-status.md`).
