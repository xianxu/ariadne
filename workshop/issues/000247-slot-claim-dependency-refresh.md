---
id: 000247
status: open
deps: []
github_issue:
created: 2026-09-23
updated: 2026-09-23
estimate_hours:
---

# Explicitly refresh repositories and dependencies with Weave

## Problem

Numbered slots own private dependency clones, but those clones have no separate
Couch thread and are easy to leave stale. In :0 the operator can directly manage
peer repositories. Agents need an explicit way to refresh the host and its dependency sources
before planning or implementation relies on them, without coupling that action
to issue reservation.

## Spec

### Agreed contract — 2026-09-23

Add explicit `weave refresh`, run from a repository, to refresh that repository
and its declared Ariadne dependencies. Reuse Weave's dependency discovery and
local bindings; add explicit Git update behavior rather than treating discovery
or ordinary `weave compile` as a refresh. Existing compile preserves dependency
revisions. This command replaces the earlier `sdlc claim` preparation proposal:
claim remains a reservation operation, with no automatic refresh hook.

The agent/operator runs refresh on the workspace's resting branch before new
work when desired. The command operates on each checkout's current branch;
no slot-name knowledge or automatic branch switch is required. For example,
the host can be on `main-slot1` while its private dependencies are on `main`.
Explicit invocation in :0 also refreshes its declared shared sibling dependencies.
Other environments remain untouched; no Couch process or inventory is required.

**Preflight all repositories before advancing any working branch:**

1. Discover the host and dependency checkouts. Require every checkout to be
   clean and have no ongoing Git operation.
2. Fetch `origin/main` for each repository and capture its exact commit SHA.
   Also record each checkout's current branch and HEAD SHA.
3. By default, require each captured HEAD to be an ancestor of its captured
   `origin/main` SHA (equality is already current). Report all identified
   blockers and advance no working branch if preflight fails. Local commits
   ahead of or divergent from the target fail this default check.

**Apply only the validated snapshots:** before updating each repository,
revalidate its branch, HEAD, cleanliness and absence of an ongoing Git operation.
If its starting state changed, stop. Fast-forward the current branch to the
captured destination SHA, not a freshly resolved moving `origin/main` ref.
Do not run an initial `git pull`; fetch plus the explicit update owns the policy.
A later remote advance is intentionally left for a subsequent refresh.

`weave refresh --rebase` explicitly permits rebasing local commits onto the
captured `origin/main` SHA instead of requiring fast-forward-only eligibility.
It retains the all-repository clean/no-operation preflight and starting-state
revalidation. Rebase rewrites local commits and may conflict; stop for explicit
resolution with actionable guidance. Never automatically stash, reset, discard
work, or switch branches.

After all repository updates succeed, run `weave compile`, including when no
refs changed. This lets ordinary retry converge after an earlier compile failure
without a new completion marker, phase registry, or custom retry flag. Preserve
existing machine-wide tool installation policy.

Fetches may update remote-tracking refs during preflight, but working branches
remain unchanged until all checks pass. This is not an atomic multi-repository
transaction: a race, update failure or rebase conflict during application can
leave earlier repositories advanced. Report that progress and stop; do not roll
back completed updates. Retry re-observes current state and captures a new set
of targets after any active Git operation has been explicitly resolved.

Project: [couch-slots-v2](../../../pair/workshop/projects/couch-slots-v2.md).

## Done when

- `weave refresh` discovers the host and declared dependencies through existing Weave mechanisms and updates their current branches without slot-specific branch naming.
- Every checkout passes clean/no-operation and fast-forward preflight before any working branch advances; blockers preserve all local work and are reported clearly.
- Each update uses the captured destination SHA and revalidates the captured starting branch/HEAD and checkout readiness; moving remote refs cannot silently change the validated target.
- `--rebase` explicitly allows local commit replay onto the captured target, with conflict guidance and no automatic stash/reset/branch switch.
- Successful repository updates always lead to compile, including already-current retries after compile failure; partial progress is visible and never automatically rolled back.
- Claim behavior remains unchanged; explicit :0 refresh includes its declared peers, while unrelated environments and machine-wide tool installation policy remain unchanged.
- Real Git fixtures cover fast-forward, already-current, ahead/divergent refusal, dirty/active-operation refusal, all-repository preflight, target-ref movement, changed starting state, partial failure, rebase/conflict and compile retry.
- Operator documentation explains explicit resting-branch refresh, `--rebase`, captured targets and partial-progress recovery.

## Plan

## Log

### 2026-09-23

Filed from the operator-approved slot dependency refresh discussion. Scope is
capture only; implementation and engineering design have not started.

## Revisions

### 2026-09-23 — Explicit refresh replaces claim preparation

Reason: the operator chose a reusable Weave command rather than coupling source
updates to issue reservation. Delta: include the current repository as well as
its discovered dependencies; default to all-repository fast-forward preflight;
make rebase opt-in; capture target and starting SHAs and revalidate before each
update. Always compile after successful updates so retries need no extra state.
The Spec and Done when above supersede the original scope below. This records
the chat decision only; #247 remains open, with no implementation or estimate.
The existing filename is retained to preserve references.

#### Superseded original specification

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


#### Superseded original acceptance criteria

- Claim in :N refreshes declared private dependencies safely before reserving the issue, then compiles when needed.
- Dirty, unexpected-branch and divergent dependencies refuse with useful recovery guidance and preserve all local work.
- Fetch/compile failures do not complete the claim; ordinary retry converges, including compile failure after successful dependency updates.
- :0 peers, other environments and the host working branch remain unchanged.
- Real Git fixtures cover fast-forward, already-current, refusal, partial failure and retry; compilation has observable failure/retry coverage.
- Operator documentation explains the preparation step and explicit host refresh.
