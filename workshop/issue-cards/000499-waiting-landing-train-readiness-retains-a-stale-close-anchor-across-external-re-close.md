---
id: 000499
status: done
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
github_issue:
started: 2026-10-20T08:58:55-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
actual_hours: 0.98
tracker:
    version: 1
    completion:
        token: close-20b0a87b4429
        repository: github.com/xianxu/ariadne
        reviewed_head: 266a62dd1139bd222ea7594044201286fce91246
        evidence_commit: 87a28e178a5a4b49864f8fb7af10773f795676ff
        landed_commit: 2425f68c3cd2e68dbc0be635b4894018ab7f360c
---

# Waiting landing-train readiness retains a stale close anchor across external re-close

## Problem

A waiting landing-train car can remain permanently unready after its own branch is re-closed. During pair#711, merge pid 92126 started at 06:32 PDT on 2026-10-20, before a test-only correction and narrow re-close. It initially observed close evidence a64d76a2. The subsequent SHIP close bound c38a0d8a, and the branch advanced to 3a1cd887. The corrected recorded suite passed on its code ancestor 62319325 in 8m13s. Nevertheless landing status still reported the old a64d76a2 anchor and refused the two corrected test files. The runner remained alive, waiting, with no admitted landing.

Read-only source inspection: landingcar.go re-runs readiness every minute and reads the new local/remote PR head, but runLandingPublishGate composes tracker records through issuerecord.go. recordsScope caches root+issuesDir for the entire command; fresh records satisfy subsequent Fresh reads, and only writes by that command invalidate the cache. An external close therefore cannot update the cached completion anchor. The same class could hide an external reopen or changed completion, which should refuse rather than authorize stale evidence.

Recovery used: interrupt only the owning waiting merge runner (exit 130), then rerun central sdlc merge --yes. No gate waiver, shared queue mutation, status fabrication or source change.
