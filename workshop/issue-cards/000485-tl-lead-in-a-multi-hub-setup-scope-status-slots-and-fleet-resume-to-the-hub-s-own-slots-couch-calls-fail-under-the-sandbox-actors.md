---
id: 000485
status: codecomplete
created: 2026-10-19
updated: 2026-10-20
estimate_hours: 2.02
github_issue:
started: 2026-10-20T11:08:14-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.69
tracker:
    version: 1
    completion:
        token: close-8c920715a6e4
        repository: github.com/xianxu/ariadne
        reviewed_head: e2d3af977ec4ae3d9e2758dc6f01cb59596a5b65
        evidence_commit: ac7b87f7bc7104e7ea7434f26692b5ca106d7d72
---

# tl-lead in a multi-hub setup: scope status slots and fleet resume to the hub's own slots; couch calls fail under the sandbox (actors '?')

## Problem

From introduce-team's live proof (10-19, two hubs: `couch ops ops-echo`):
1. **`tl-lead status slots` isn't scoped to the hub.** In ops-echo:0 it rendered the whole machine's fleet into echo's
   status file, not just the slots its `.hub` claims (echo).
2. **tl-lead's couch calls fail inside the agent sandbox.** `tl-lead status slots` shows `?` in every actor column,
   and `tl-lead fleet resume` lists every slot as "unknown: couch --actors unreadable … dial unix
   /tmp/pair-message-501/….sock: operation not permitted". The sandbox exemption covers a bare `couch` call only, so
   a couch call made from inside tl-lead is blocked. Run outside the sandbox, `fleet resume` classed correctly (one
   slot to resume, the rest running or idle).
3. `fleet resume`'s dry run prints dozens of "ambiguous … unclaimed" rows for dormant repos (robotics, zellij,
   xianxu.dev, …), which buries the one actionable row.
