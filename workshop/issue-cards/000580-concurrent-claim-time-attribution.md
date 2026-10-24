---
id: 000580
status: codecomplete
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
github_issue:
started: 2026-10-23T19:27:24-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 1.45
tracker:
    version: 1
    completion:
        token: close-01aae4f47e22
        repository: github.com/xianxu/ariadne
        reviewed_head: dd39e615a6561921075b0066ba0cd09a8e0ca22d
        evidence_commit: b8b2e136da4ff8c9f9f7e6a47313fd04c6d9ea20
---

# Active-time attribution cross-counts concurrent same-slot claims instead of active branches

## Problem

On2026-10-23, parley.nvim:3 held concurrent claims for parley.nvim#418 and parley.nvim#322. Actual work stayed on #418's issue branch until its landing at14:29 and hub dispatch to #322 at14:30. A fresh main-based #322 branch began then. At14:42, sdlc close --issue322 proposed4.59h measured actual, including a dominant203.5m/74% segment11:18–14:42. That segment includes #418's targeted flake classification, queued verification and landing, not #322 implementation.

Observed measurement text: window2026-10-23T00:22:39-07:00 through2026-10-23T14:42:03-07:00; provenance active-time-v3-engagement-v1;1 ownership interval;5 admitted sessions; dominant #322203.5m/74%11:18–14:42. Code candidate19dda8d9339507f2cfd8b368a8505a25d133e023; tracker5dcff7790ce1157ff18229a5c7d019573a3befcd. The reporter stopped the close review before finalization; tracker remained working with no completion and no actual adopted. TL35643b0f directed --no-actual, no typed hours, and this unclaimed filing.

The printed exclusions already removed some segments adopted by #418's earlier closes. Thus this particular capture directly proves wrong-issue attribution of #418's post-close tail, not that the same exact interval was persisted into both final cards. Concurrent claims must not cross-count one slot's work or assign an unadopted tail to the other claim merely because its ownership interval remains open.
