---
id: 000563
status: open
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
---

# Queue status must explain older supplier helper incompatibility

## Problem

After #560, central sdlc queue status invokes inspect --supervised on the helper selected by the caller Makefile.workflow supplier. Older Pair/Parley pinned substrates and pre-#560 Ariadne issue branches have an acquire-only helper, which argparse rejects with exit2. The status display hides stderr and suggests an ineffective retry.
