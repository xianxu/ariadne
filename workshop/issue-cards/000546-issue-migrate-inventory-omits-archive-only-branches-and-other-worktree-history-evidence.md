---
id: 000546
status: done
created: 2026-10-21
updated: 2026-10-22
estimate_hours: 2.11
github_issue:
started: 2026-10-22T03:02:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.04
tracker:
    version: 1
    completion:
        token: close-b12f1952be1c
        repository: github.com/xianxu/ariadne
        reviewed_head: eaad89ade92b1faa464ae407615a6e803cdcda06
        evidence_commit: 692b1b01cfe0bc66034202b18813f463d4c950f6
        landed_commit: e6934a71ab74439a5149bac73b34ecd45db84600
---

# issue migrate inventory omits archive-only branches and other-worktree history evidence

## Problem

Pre-existing issue migrate inventory omits archived issue evidence outside pinned main. Discovered by #543 close review round 2 (BR-2, authoritative-inventory-completeness); TL explicitly moved importer completeness out of new-repo init's scope.

Reproductions against real Git/bare origin: (1) clone B creates and pushes a branch containing workshop/history/issues/000001-old.md while main has no issues; clone A migrationInventory scans branch edits only under workshop/issues and plans an empty tracker. (2) another linked live worktree contains uncommitted workshop/history/issues/000001-old.md; dirtyIssuePaths scans only workshop/issues, again allowing an empty plan. The same roots/scope rule must examine flat legacy history and detached/inherited branch/worktree evidence, rather than patching only these two sites.
