---
id: 000461
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T02:31:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.46
tracker:
    version: 1
    completion:
        token: close-8f94828d2be7
        repository: github.com/xianxu/ariadne
        reviewed_head: 70a7180bbdfc215ef63e281a128256d8a19c38d1
        evidence_commit: 4be0656531877f055ea326afaa6fc685c31f3a73
---

# dev-aliases in Couch slots run a stale slot-sibling ariadne build, not the :0 binary (OWNER misdetected)

## Problem

In a Couch slot, the `sdlc` command (and its siblings) is a shell function from `construct/dev-aliases.sh` (ariadne#57), captured in the agent's shell snapshot. It shadows `/Users/xianxu/workspace/ariadne/bin` on PATH. Its OWNER detection resolves to the slot's sibling checkout (e.g. `~/workspace/worktree/pair-slot2/ariadne`) instead of the operator's `:0` checkout (`~/workspace/ariadne`) that the dev-aliases header says it should use. Because `COUCH_TREE` is set, the #405 rule skips the rebuild, so the slot runs a binary built whenever that sibling was last built, forever. Found by pair:2 on 2026-10-19: its `sdlc` was built Oct 18 18:19 from 3a0012b4, before #394, #452, #454 and #453, and refused #394's `flake_tests` card field. Every slot is likely affected, which explains the night's old-runner card refusals, and means slots have not been running the gates and rules the TL rebuilt into ariadne:0.
