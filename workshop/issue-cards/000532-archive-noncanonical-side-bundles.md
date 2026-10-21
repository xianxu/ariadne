---
id: 000532
status: codecomplete
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
github_issue:
started: 2026-10-21T16:36:06-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.25
tracker:
    version: 1
    completion:
        token: close-8db52632e43b
        repository: github.com/xianxu/ariadne
        reviewed_head: d6309d22f5cc7bf4911d0eba4f97abc0c7a1a6d8
        evidence_commit: eb95a774a2e110fdc66e413d9356ad6992507b24
---

# Archive preflight misses noncanonical side bundles

## Problem

pair#768 landed in PR521 (https://github.com/xianxu/pair/pull/521; reported merge 2026-10-21 14:37:41 PDT), but post-merge SDLC cleanup refuses with `active or conflicting archive artifact workshop/plans/000768-darwin-pty-final-output-plan.md.side/shared-owner-draft.md`. The card is done and the issue and parent plan are archived, while this tracked side document remains active. The documented retry `sdlc merge --branch 000768-darwin-pty-final-output --yes` repeats the refusal. Reported `sdlc issue recovery reconcile --issue 768` finds no unfinished operations and deleted the remote branch without archiving this leftover. Successful landing therefore leaves an incomplete artifact bundle and no successful documented cleanup retry.
