---
id: 000266
status: open
deps: []
github_issue:
created: 2026-09-28
updated: 2026-09-28
estimate_hours:
card_mirror: '2b1f8da5113cec05740a08ea09b3c8e21a164de5' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc id lint cannot resolve publication target in CI checkouts

## Problem

Consumer merge-check CI fails on `40-duplicate-issue-id.sh` in every PR run
(seen in parley.nvim runs 36474779448, 36491104572 and 36511435873):

```
id lint COULD NOT RUN: /home/runner/work/parley.nvim/parley.nvim has a fetched
issue tracker but its publication target is unusable: configure main to track
one named remote/main: resolve publication target: git config: exit status 1
```

`actions/checkout` leaves a detached PR checkout with no local `main` tracking
a remote. The #252 lint wants a publication target it doesn't need just to read
IDs, and it exits 2 ("a check that did not look must not report clean"). So
consumer CI is red whatever the change is, which also blocked #241's "real
CI green" evidence (weave install and compile passed).

## Spec

A read-only lint should derive what it reads from the fetched tracker ref
without requiring publication configuration, or the seeded CI should configure
the tracking it needs. Pick one after reading the lint's resolution path.

## Done when

- parley.nvim merge-check passes `40-duplicate-issue-id.sh` on a PR run.
- A regression test covers a detached CI-style checkout with a fetched tracker.

## Plan

- [ ]

## Log

### 2026-09-28
