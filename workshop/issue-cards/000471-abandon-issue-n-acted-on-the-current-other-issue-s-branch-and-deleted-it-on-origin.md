---
id: 000471
status: codecomplete
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
github_issue:
started: 2026-10-19T13:17:35-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 3.16
tracker:
    version: 1
    completion:
        token: close-a477534abc52
        repository: github.com/xianxu/ariadne
        reviewed_head: 92211bc7cc8ab76e431e6e7830cbae082f5879fe
        evidence_commit: dc1903983646269f88abf7f06b3af1f5d2fedef1
---

# abandon: --issue N acted on the current (other issue's) branch and deleted it on origin

## Problem

2026-10-19, pair:1 (pair repo): `sdlc abandon --issue 694 --as punt --reason …` was run from the checkout of a *different*
issue's branch (#695). It acted on the current branch instead of #694's:
- committed #694's abandon log on top of #695's tip;
- saved that commit as `refs/ariadne/abandoned/000694` on origin;
- **deleted the #695 branch locally and on origin**;
- left #694's real branch untouched.

pair:1 repaired it by hand (recreated #695 at its tip, force-with-lease'd the archive ref to #694's real tip 45baebbbd, deleted
#694's branch). Nothing was lost only because the slot noticed. An abandon from the wrong checkout can delete another issue's
unmerged work on origin.
