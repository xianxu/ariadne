---
id: 000511
status: open
deps: []
github_issue:
created: 2026-10-20
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '9d9c2e525e0a83ff07bfe9203151fedcb7e140c2' # card fields mirrored from issue-cards; edit via sdlc
---

# landing trains: don't restart CI when main advances only by issue-detail commits (filings collided with train 197 twice on 10-20)

## Problem

`sdlc issue new` and `sdlc issue publish` commit issue details straight to main. On 10-20, two such commits (pair#745
and #746 filings) landed while pair train 197 waited on CI; each time the conductor re-integrated main and restarted
CI, delaying a train that had already passed its full suite. Filing and landing collide.

## Spec

Options (choose in the plan):
1. **Trains ignore detail-only commits:** when main advances only by commits that touch `workshop/issues/` (tracker
   mirrors, issue details, archive moves), the conductor fast-forwards past them without re-running CI or the
   suite, since they can't change behavior. The merge stays a true merge with main.
2. Batch detail publishes so they don't land while a train's CI is in flight.

Option 1 is preferred: a mechanical, verifiable rule ("the diff since the verified base touches only workshop/
issues/ paths").

## Done when

- A train whose main advanced only by issue-detail commits lands without restarting CI or the suite; any other
  commit still forces re-integration. Tests pin both.

## Plan

- [ ]

## Log

### 2026-10-20
