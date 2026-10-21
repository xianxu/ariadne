---
id: 000504
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours: 2.142
github_issue:
started: 2026-10-20T09:56:17-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 2.82
tracker:
    version: 1
    completion:
        token: close-6272c7a74a84
        repository: github.com/xianxu/ariadne
        reviewed_head: 4935e84497d52ad95e192fdb16f61ee974b38bfc
        evidence_commit: be564e2e52f08d205a36442809e0b0281f82babf
        landed_commit: 619ddaf81e4dd0d0cc33d885a59ca3ffebd09f3f
---

# tl-lead: the hub session file moves to session.md at the repo root (no team folder, no operator name), with a status move verb and the decisions-are-tracked rule

## Problem

The TL session file lives at `teams/<team>/tl-status-<operator>.md` (side docs in `tl-status-<operator>.side/`). Since
introduce-team settled on one team per ops repo, the team folder adds nothing (the repo is the team), and the
operator's name adds nothing either (the file is gitignored, so it's per machine and per operator already).
Operator, 10-20 (ops side doc `2026-10-20-03-session-file-everywhere.md`): move it to the repo root.

Also stated there, and worth writing into tl-lead: the session file and its side docs are the **ephemeral TL/human
interface**; durable decisions that workers can act on live in the issue system (tracked, managed by sdlc), so a
decision made in a side doc gets copied into the tracked artifact (issue, project Log, charter) when it's made.
