---
id: 000362
status: open
deps: []
github_issue:
created: 2026-10-13
updated: 2026-10-13
estimate_hours:
card_mirror: '23b9f900634ddb06a95c7f1ed5149e5a1910a8c3' # card fields mirrored from issue-cards; edit via sdlc
---

# make test doesn't run construct/scripts/test/*.test.sh (dev-aliases, others), so their regressions are only caught by hand

## Problem

Shell tests under `construct/scripts/test/*.test.sh` (e.g. `dev-aliases.test.sh`, 48 cases added by #355) aren't run by `make test` or CI, so a regression in them is caught only if someone runs them by hand. #357 fixed the same gap for `test-lock.test.sh` by moving it under `scripts/test/`, which CI runs.

## Spec

Run every `construct/scripts/test/*.test.sh` from `make test` (and the CI merge check), the same way `scripts/test/` is run, under bash 3.2-compatible settings. If any are slow or need a live environment, mark them and leave those out explicitly rather than silently.

## Done when

- `make test` runs all `construct/scripts/test/*.test.sh` (or names the excluded ones and why); a deliberately broken test fails `make test`.

## Plan

- [ ]

## Log

### 2026-10-13
