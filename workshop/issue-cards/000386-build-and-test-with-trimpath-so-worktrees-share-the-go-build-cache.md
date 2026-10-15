---
id: 000386
status: done
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T10:52:09-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.53
tracker:
    version: 1
    completion:
        token: close-60cd35fa71d4
        repository: github.com/xianxu/ariadne
        reviewed_head: 4e4c777f85600a57f01f49e1c4a9731f884ce104
        evidence_commit: 38b6cf977236644f18545990fb04229b829d2d7b
        landed_commit: 8b232cda89dc951ce81e3952a6056a3da19fe3f5
---

# Build and test with -trimpath so worktrees share the Go build cache

## Problem

On 2026-10-15 the shared Go build cache (`GOCACHE=/Volumes/AgentScratch/go-build`) filled the 186 GB scratch volume
(184 GB), failing every slot's suite. Main multiplier: builds embed each worktree's absolute source path, so the same
package is cached separately for each of ~20 slot worktrees; after a `go clean -cache`, ~30 GB came back within
minutes. The operator approved building with `-trimpath`.
