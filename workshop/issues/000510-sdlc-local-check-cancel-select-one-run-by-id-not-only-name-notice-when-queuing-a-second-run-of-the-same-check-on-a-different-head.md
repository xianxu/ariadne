---
id: 000510
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '1d22d7bec5e64c9540474d9cda854ac814bf6087' # card fields mirrored from issue-cards; edit via sdlc
---

# sdlc local-check cancel: select one run by id (not only --name); notice when queuing a second run of the same check on a different head

## Problem

`sdlc local-check cancel` selects runs only by `--name`. On 10-20, pair:6 had two queued `suite` runs from one
checkout (an original on an older head, and an affected-check fallback to the full suite on the current head, the
pair#738 case) and could not cancel just one: `--name suite` cancels both and drops the original's FIFO place, so it
had to cancel both and requeue.

## Spec

- `sdlc local-check cancel --run ID` (the leased run id that `local-check status` already shows), alongside `--name`;
  refuses an ID owned by another checkout.
- When `local-check run` would queue a second run of the same check from the same checkout on a **different** head,
  say so and offer to replace the stale one, instead of silently queueing both.

## Done when

- Tests pin: cancel by run ID leaves the sibling run queued in place; a foreign ID refuses; the duplicate-run notice.

## Plan

- [ ]

## Log

### 2026-10-20
