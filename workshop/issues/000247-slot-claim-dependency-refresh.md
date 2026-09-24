---
id: 000247
status: open
deps: []
github_issue:
created: 2026-09-23
updated: 2026-09-23
estimate_hours:
---

# Refresh private slot dependencies before claiming an issue

## Problem

Numbered slots own private dependency clones, but those clones have no separate
Couch thread and are easy to leave stale. In :0 the operator can directly manage
peer repositories. Starting a new issue in :N should refresh its private tools
and sources before planning or implementation relies on them.

## Spec

Make private dependency refresh a preparation step of `sdlc claim` in a verified
numbered slot, before publishing the issue reservation. Discover dependencies
from existing workspace declarations/bindings, not a Couch-maintained inventory.
Fetch and fast-forward each clean private dependency checkout on its expected
branch to its declared remote baseline. Dirty checkouts, unexpected branches,
and divergent local commits stop preparation with an actionable explanation;
never automatically stash, reset, discard work, or switch branches.

When dependencies change, rerun `weave compile` before completing the claim.
A failed fetch, update or compile leaves the claim uncompleted and can be retried
through the ordinary claim command. Keep retry behavior idempotent, without a
new phase registry or custom retry flag. Account for a retry after repositories
advanced but compilation failed; unchanged refs alone do not prove setup succeeded.

Primary :0 peer repositories remain operator-managed. The slot host's own branch
is not refreshed by claim; its refresh remains explicit. Preserve environment
isolation and existing machine-wide tool installation policy. This belongs to
SDLC and must work without Couch running.

Project: [couch-slots-v2](../../../pair/workshop/projects/couch-slots-v2.md).

## Done when

- Claim in :N refreshes declared private dependencies safely before reserving the issue, then compiles when needed.
- Dirty, unexpected-branch and divergent dependencies refuse with useful recovery guidance and preserve all local work.
- Fetch/compile failures do not complete the claim; ordinary retry converges, including compile failure after successful dependency updates.
- :0 peers, other environments and the host working branch remain unchanged.
- Real Git fixtures cover fast-forward, already-current, refusal, partial failure and retry; compilation has observable failure/retry coverage.
- Operator documentation explains the preparation step and explicit host refresh.

## Plan

## Log

### 2026-09-23

Filed from the operator-approved slot dependency refresh discussion. Scope is
capture only; implementation and engineering design have not started.
