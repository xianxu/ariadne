---
id: 000405
status: done
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
github_issue:
started: 2026-10-15T20:34:51-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.53
tracker:
    version: 1
    completion:
        token: close-f2911260c081
        repository: github.com/xianxu/ariadne
        reviewed_head: e772cdd5162cd8125de2bda2fe873679708226af
        evidence_commit: d83a96b914e030f5313bafbaa2ec87ee9d51e584
        landed_commit: ac59666058e556b97ffba4c373386a017a973443
---

# dev-aliases: skip the rebuild in Couch agent shells (COUCH_TREE) and when the binary is not stale

## Problem

Every call of a dev-aliases function rebuilds the binary in its owner's checkout, which is the :0 slot. That includes calls from agent shells in Couch slots. Agent sessions snapshot the operator's shell, so in each slot `couch`, `sdlc` and `weave` resolve to the dev-alias functions. Each agent call then writes `bin/X` into pair:0 or ariadne:0, the operator's staged checkouts that only the TL and the operator own. On 10-15 at 20:28, right after a Couch restart, ariadne:2's `couch --message-status` calls rebuilt `bin/couch` in pair:0. Agent sessions whose shell snapshot predates #355 also `rm -f` the binary before building, which is why the binaries briefly vanished during the restart. A sandboxed call can't write there, so its build fails, and the call fails with it.

Every call also pays a `go build`, even when nothing changed.
