---
id: 000302
status: codecomplete
created: 2026-10-08
updated: 2026-10-12
estimate_hours:
github_issue:
started: 2026-10-12T10:29:38-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.53
tracker:
    version: 1
    handoff:
        token: move-438fe577c5af
        repository: github.com/xianxu/ariadne
        source_branch: refs/heads/main-slot2
        source_base: 096d6963b2023291f9e8bf615cfd231ee4c68b60
        source_head: 096d6963b2023291f9e8bf615cfd231ee4c68b60
        source_path: workshop/issues/000302-tempdir-cleanup-race.md
        source_blob: 8d8eaafd888dde0b3129017af7b1bcfde9f4d92b
        destination: workshop/issues/000302-tempdir-cleanup-race.md
        main_commit: ddc102977add38b0242f3bb5cc8b5bd5ec75b426
    completion:
        token: close-678d1de4a178
        repository: github.com/xianxu/ariadne
        reviewed_head: 6caa63c639fc0859603049140327749f441ee879
        evidence_commit: cbf53003cf92c265204cc94aa4db1d796a4d4cd2
---

# Test TempDir cleanup race: directory not empty under sharded make test

## Problem
