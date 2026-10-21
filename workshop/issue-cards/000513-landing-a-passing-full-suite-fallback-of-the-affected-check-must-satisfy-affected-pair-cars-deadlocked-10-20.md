---
id: 000513
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T20:07:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.07
tracker:
    version: 1
    completion:
        token: close-8be9a86ab58b
        repository: github.com/xianxu/ariadne
        reviewed_head: e98939c6e3c7f9acd2be1d8785ffb98397d94830
        evidence_commit: f915c1e1e501e789e3ee239b84b0a6c53d140905
        landed_commit: b9c8a500c4971d117943c16d79ab96ae64cab911
---

# landing: a passing full-suite fallback of the affected check must satisfy 'affected' (pair cars deadlocked 10-20)

## Problem

On 10-20 pair's affected check fell back to the full suite on fresh checkouts (the runtime bundle isn't generated,
so the Go graph can't load; pair#738), and the fallback **passed** but posted only `sdlc-local/suite`. `sdlc merge`
then refuses readiness forever: "PR #489 needs recorded local checks on its head: affected". pair#732 (and pair#706)
deadlocked with all checks green; no pair landing for over two hours.
