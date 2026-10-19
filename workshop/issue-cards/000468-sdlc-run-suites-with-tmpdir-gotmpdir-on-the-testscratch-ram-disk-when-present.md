---
id: 000468
status: done
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T09:32:22-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.32
tracker:
    version: 1
    completion:
        token: close-19b368fa7f23
        repository: github.com/xianxu/ariadne
        reviewed_head: 9e4c4f89b69e712e4ad6bd608f3416530827c64a
        evidence_commit: 3d92d990445f203f8b5bc09f9158a550279ee04e
        landed_commit: 546516f30d0af49dc3973e5baf54a716933d8046
---

# sdlc: run suites with TMPDIR/GOTMPDIR on the TestScratch RAM disk when present

## Problem

Test suites cause an fsync storm: a pair suite build minute showed 550,944 APFS container metadata writes in 64 s (median 8,600/s), 98% rewrites of the same ~30k blocks, i.e. thousands of forced transaction commits a second on the shared SSD (ops disk-writes analysis, 2026-10-19). Test temp files today land on the SSD: agents' TMPDIR on /Volumes/AgentScratch (an APFS volume in the same container) and suites such as ariadne's sharded run in the system temp folder (/var/folders/…). The operator approved a 16 GB RAM disk at /Volumes/TestScratch (ops teams/devinfra/tools/testscratch, a login item). 56 leftover go-build folders (~2 GB) had also accumulated in the scratch temp folder.
