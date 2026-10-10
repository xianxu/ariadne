---
id: 000313
status: open
deps: []
github_issue:
created: 2026-10-09
updated: 2026-10-09
estimate_hours:
card_mirror: '78f7c6acaa36f7361fe259cb9789998f26cb94a0' # card fields mirrored from issue-cards; edit via sdlc
---

# Recheck pair's git identity on 2026-11-09 (H1 recurrence watch)

## Problem

From 2026-06-21 to 2026-10-09, pair's shared `.git/config` carried `user.name=T` / `user.email=t@e.com`. That meant 1,597 commits on pair main authored as `T <t@e.com>`, and `operator: T` in claimant records (evidence H1 in `workshop/pensive/2026-10-09-01-pensive-sdlc-robustness-evidence.md`). On 2026-10-09 the operator unset both keys, and all 7 pair checkouts verified as `Xian Xu <xianxu@gmail.com>`.

The suspected writer is still in place: pair's shell tests (`tests/review-readiness-cli-test.sh`, `review-indicator-test.sh`, `review-*-restore-test.sh`, `review-observation-test.sh`). They run `git config user.name T` after `cd`/`git init` without stopping on failure, so a failed step writes into the real repo. The operator chose to watch for recurrence for a month rather than fix the tests now.

## Spec

On or after **2026-11-09**:
- run `git -C ~/workspace/pair config --local --get user.name` and `--get user.email`;
- run `git -C ~/workspace/pair log origin/main --since=2026-10-09 --author='t@e.com' --oneline`.

Outcomes:
- **Clean:** close this issue as done.
- **Recurred:** unset the keys again, and fix pair's tests to use `git -C "$REPO" config` (or `set -e` / `cd … || exit`). File the fix in pair, and close this issue pointing at it.

## Done when

- The check above has run on or after 2026-11-09, and its result is recorded in `## Log`.
- If it recurred, a pair fix is filed or landed.

## Plan

## Log

### 2026-10-09

Filed at the operator's request, as a recurrence check for the ariadne-robustness-1 housekeeping item. Not due before 2026-11-09.
