---
id: 000257
status: open
deps: [000255]
github_issue:
created: 2026-09-27
updated: 2026-09-27
estimate_hours:
card_mirror: 'f4ba1e05be1b424cd79af8663d99852002930c26' # card fields mirrored from issue-cards; edit via sdlc
---

# lint-ids reads the cutover marker from the checked commit, not the checkout

## Problem

`sdlc issue lint-ids --base B --head H` (merge-check `40-duplicate-issue-id.sh`)
decides the cutover state from the **checkout's** `workshop/issue-tracker.json`,
not from `H`. In a pre-push hook the pushed commit is not checked out, so the
two disagree. Found cutting over you-decide (#255, 2026-09-27): `migrate --apply`
had bootstrapped the tracker and was pushing the commit that adds the marker;
the hook's lint saw "tracker exists, checkout lacks the marker", exited 2, and
the publish gate blocked the push. The apply finished only with the hook
skipped once (operator-authorized). Any repository with the pre-push publish
gate hits this at cutover, and the lint is wrong in general: it must judge the
commit under check.

## Spec

With `--head`, lint-ids reads the cutover marker (and so the tracked/legacy
decision) from `H`'s tree; without `--head` it keeps reading the checkout. The
cutover mismatch guard still refuses a real mismatch at `H`.

## Done when

- A pre-push of the migration commit from a checkout without the marker passes
  the id lint (e2e: bootstrap tracker, lint `--head` the marker commit).
- A head without the marker in a tracked repository still refuses.

## Plan

- [ ] Read the marker at `--head` in lint-ids' mode decision
- [ ] Tests for both outcomes

## Log

### 2026-09-27
