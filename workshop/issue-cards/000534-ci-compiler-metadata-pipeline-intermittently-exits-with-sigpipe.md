---
id: 000534
status: codecomplete
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T17:14:04-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.37
tracker:
    version: 1
    completion:
        token: close-ca85ebee93c1
        repository: github.com/xianxu/ariadne
        reviewed_head: c267bbefd9e4e337aebf13a2a63392c2e0ea0e8f
        evidence_commit: f1102187a5a9556da4e841eee6b7e696cef3a986
---

# CI compiler metadata pipeline intermittently exits with SIGPIPE

## Problem

While verifying echo#1 with the unchanged upstream portable-ci.test.sh harness,
the actual workflow Build first pinned compiler step intermittently returned
141 even though its metadata had the selected revision and modified=false.
The template pipes printf into grep -Eq under set -euo pipefail. Early grep
success can close the pipe while printf is writing, producing SIGPIPE.
