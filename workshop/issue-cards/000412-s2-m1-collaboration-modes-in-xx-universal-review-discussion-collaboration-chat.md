---
id: 000412
status: done
created: 2026-10-16
updated: 2026-10-16
estimate_hours: 4.33
github_issue:
started: 2026-10-16T15:42:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.01
tracker:
    version: 1
    completion:
        token: close-cc0c70b1e9b5
        repository: github.com/xianxu/ariadne
        reviewed_head: 5fdc9f3521d89f638a5ac7bcca822c808f286500
        evidence_commit: 273b80b0d77603b8db686a99ab18dfa200de6df9
        landed_commit: 1db0ec0ed0b6e2affb426c2a311bebf5c24c70a2
---

# S2 M1: collaboration modes in xx-universal-review (discussion, collaboration, chat)

## Problem

One review mode set (`process-instruction:`) governs a whole document, and every 🤖 question is answered in place. The
operator wants each section to say where the answer to an ask goes. The final design is ops
`teams/devinfra/projects/universal-agent-input-interface/design.side/01-s2-section-modes.md` (ops db130ac0), with
its side docs 02 to 06; design owner parley.nvim:4. This is S2 milestone 1 on the ariadne side. M2 (the per-type
audit) is a later issue.
