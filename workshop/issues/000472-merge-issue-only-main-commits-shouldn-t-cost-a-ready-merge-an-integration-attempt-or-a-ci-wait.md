---
id: 000472
status: open
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-19
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'fd0e89722d843029fbbdcb24015f0c030347a85e' # card fields mirrored from issue-cards; edit via sdlc
---

# merge: issue-only main commits shouldn't cost a ready merge an integration attempt or a CI wait

## Problem

2026-10-19, pair#705 (pair:7): its integrated suite passed and CI passed, then two **issue-only** commits reached main (a
`sdlc issue new` publish and an `abandon` archive, touching only `workshop/`). Each forced `sdlc merge` to integrate main
again and wait for fresh CI, using up attempt 3 of 3. On a busy repo, issue filing alone can starve a ready merge, and the
hub had to ask every slot to stop publishing.

## Spec

When main has moved since the PR head's last integration, and every new main commit touches only paths that cannot affect
the build or tests (the issue-tracker details paths: `workshop/issues/`, `workshop/history/`, `workshop/plans/`, …, from one
declared list), `sdlc merge`:
- does not count it as a new integration attempt;
- does not rerun the recorded suite (already true when code didn't change; confirm);
- integrates and merges without waiting for a fresh CI run when the repo's CI ignores those paths, or, if the repo requires
  an up-to-date branch with CI, waits for CI but doesn't spend an attempt.

Where possible, avoid integrating at all: merge the PR as is when GitHub allows it (branch protection not requiring
up-to-date), since issue-only main commits can't conflict with code.

## Done when

- Fixture: PR ready; main gains 3 issue-only commits during the wait; merge lands on its first attempt, with no suite rerun.
- Fixture: a main commit touching code still counts as an integration attempt (unchanged).

## Plan

- [ ]

## Log

### 2026-10-19
