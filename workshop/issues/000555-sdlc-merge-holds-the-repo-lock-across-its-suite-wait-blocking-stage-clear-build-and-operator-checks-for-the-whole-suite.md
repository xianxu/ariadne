---
id: 000555
status: working
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b06bf6fd6c6339474650069c3bf84be0ba3c9cdb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T11:30:23-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# sdlc merge holds the repo lock across its suite wait, blocking stage --clear/--build (and operator checks) for the whole suite

## Problem

## Spec

10-22 ~11:30: pair:9's `sdlc stage --clear` on pair:0 (to free it for the operator's next staged check) waited on
pair's common sdlc repo lock, held for 30+ minutes by pid 17156, pair:6's `sdlc merge --yes` for pair#803, which
was sitting in its landing turn waiting for its integrated suite. So a landing waiting for a suite blocks every
stage operation in the repo, and with it the operator's live checks, for the whole suite duration.

Fix: a merge runner holds the common repo lock only around the short critical sections it needs (integrate, push,
merge, archive), not across the suite wait; or stage operations use a separate lock (the stage target's lock) that
doesn't conflict with a landing's suite wait. Keep the existing safety: a stage must still never race the landing's
integration or push.

## Done when

- With a merge runner waiting on its integrated suite, `sdlc stage --clear` and `--build` on the stage target proceed
  without waiting for the suite (test with a held fake suite).
- Stage operations still refuse or wait during the landing's actual integrate, push or merge sections (test).

## Plan

- [ ]

## Log

### 2026-10-22
