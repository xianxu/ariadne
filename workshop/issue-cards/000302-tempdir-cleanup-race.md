---
id: 000302
status: done
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
actual_hours: 1.54
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
        token: close-b72e9032701e
        repository: github.com/xianxu/ariadne
        reviewed_head: 47a4e4636218dcd3b3dba7a160a0773a03ea48f5
        evidence_commit: 88d07bdc3a03f19e9deb88a524419ba05b57191f
        landed_commit: f0ad603744af371c613193d7a0204763db765b73
---

# Test TempDir cleanup race: directory not empty under sharded make test

## Problem
