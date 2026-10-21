---
id: 000511
status: open
created: 2026-10-20
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T03:30:03-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
---

# landing trains: don't restart CI when main advances only by issue-detail commits (filings collided with train 197 twice on 10-20)

## Problem

`sdlc issue new` and `sdlc issue publish` commit issue details straight to main. On 10-20, two such commits (pair#745
and #746 filings) landed while pair train 197 waited on CI; each time the conductor re-integrated main and restarted
CI, delaying a train that had already passed its full suite. Filing and landing collide.
