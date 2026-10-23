---
id: 000563
status: codecomplete
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
github_issue:
started: 2026-10-22T21:59:41-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.13
tracker:
    version: 1
    completion:
        token: close-f4bd46a3cba9
        repository: github.com/xianxu/ariadne
        reviewed_head: 6a0b35faea2b351b5b1ec19a310f6ea1d19f1948
        evidence_commit: a3460caca2e05f9aff2566c2cf6353c452d144fd
---

# Queue status must explain older supplier helper incompatibility

## Problem

After #560, central sdlc queue status invokes inspect --supervised on the helper selected by the caller Makefile.workflow supplier. Older Pair/Parley pinned substrates and pre-#560 Ariadne issue branches have an acquire-only helper, which argparse rejects with exit2. The status display hides stderr and suggests an ineffective retry.
