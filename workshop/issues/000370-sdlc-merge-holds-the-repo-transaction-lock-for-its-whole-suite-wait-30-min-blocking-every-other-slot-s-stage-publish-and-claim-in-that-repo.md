---
id: 000370
status: open
deps: []
github_issue:
created: 2026-10-14
updated: 2026-10-14
estimate_hours:
card_mirror: '90deb7a813081c73efbaabb239a33d6849f465b0' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc merge holds the repo transaction lock for its whole suite wait (30+ min), blocking every other slot's stage, publish and claim in that repo

## Problem

Since ariadne#332 (sdlc merge integrates main and runs the full suite before landing), `sdlc merge` holds the repo's sdlc transaction lock for its whole run, including the suite queue wait and the suite itself, 30+ minutes on 2026-10-14, and a second run when main moved. Meanwhile every other slot in that repo is blocked from `sdlc stage` / `stage --clear`, `issue new` / `issue publish`, `claim`, `move` and other locked verbs: pair:6's stage clear waited on pair:3's merge (pid 3229); the TL's `issue new` for this very issue waited ~1h on ariadne:2's merge (pids 81291, 37511, which started a second suite); pair:6's P0 hotfix filing waited too. The operator made this a priority fix.

## Spec

Hold the transaction lock only around short git/tracker mutations: integrating main, recording state, the landing push, archive. Release it for replay review, GitHub registration polling, the CI wait and the suite (queue and run). On every reacquire, revalidate the exact branch HEAD and worktree and fetch main: HEAD drift refuses; main drift re-integrates and reruns the gates. Keep the integration state and the interrupted-merge recovery contract from #332 (crash between phases resumes or reports cleanly).

## Done when

- While a merge's suite is queued or running, another slot's `issue new`, `stage`, `claim` in the same repo proceed (test).
- HEAD drift during the suite refuses; main drift re-integrates and reruns; a crash between phases recovers (tests).

## Plan

- [ ]

## Log

### 2026-10-14
