---
id: 000555
status: open
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'c71ddeea6288c4d8bf0bf5fed3464fc4233a1cdd' # card fields mirrored from issue-cards; edit via sdlc
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
