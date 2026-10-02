---
id: 000288
status: working
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours:
card_mirror: '79ff06cc9a7f98e9c6d6dfba3f40e22bb546cacb' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T12:14:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Bulk read-only claim observation

## Problem

pair#367 needs one row per local slot joining sdlc claims with Couch's slot
inventory. The only claim observation today is per issue: `sdlc issue show N
--json` (one tracker read per issue). `sdlc issue list` prints text, has no
`--json`, reads local `workshop/issues`, and carries no claimant. Nothing
answers "which working issues does this machine claim, and in which worktree"
in one call, and nothing prints this machine's own fingerprint, so a consumer
cannot even filter claimant fields it reads elsewhere.

## Spec

Project: pair/workshop/projects/cross-slot-work-scheduling.md. Captured for
operator review; no implementation is authorized by this issue creation.

Add a read-only bulk observation over the tracker's cards (versioned JSON, same
schema discipline as `issue show --json`): per non-terminal issue, card status
and revision, the claimant record, and `relation` judged against the local
machine rather than the cwd checkout (this-machine with its worktree path,
other-machine, unattributed, unknown). Answer from one tracker read (the
issue-tracker branch), with no per-issue probes on the request path; worktree
fate is optional and, if included, local-only. Expose this machine's identity
(fingerprint and name) so consumers can compare without reimplementing
`resolveClaimantIdentity`. A stale or unreachable tracker is reported as
stale/unknown, never as "no claims". Register it in the #280 recovery catalog
as read-only.

## Done when

- `sdlc fleet inventory --json` reports, per local worktree, the active claims
  this machine holds there (from one tracker read per repository), claims whose
  worktree is gone as `dangling_claims`, and this machine's identity; tested
  with the real tracker fixture (placed, dangling, other-machine omitted,
  unclaimed omitted).
- This machine's identity matches what `claim` records (one shared derivation).
- Stale/unreachable tracker, unreadable cards and unknown identity yield an
  explicit `claims_state` (stale/partial/unknown), never an empty "no claims";
  tested.
- One malformed tracker card no longer fails the whole tracker read: it is
  quarantined and reads as unknown; writes to it refuse naming the cause;
  others proceed; tested end to end.
- `fleet inventory` appears in `sdlc help recovery` as read-only, with proofs.

## Plan

Durable plan: `workshop/plans/000288-bulk-read-only-claim-observation-plan.md`.

- [ ] M1 — a malformed card is quarantined, not fatal (snapshot quarantine, `Require`, reader audit, transferguard fail-closed)
- [ ] M2 — fleet inventory reports this machine's claims per worktree (`localMachine`, pure `PlaceClaims`, contract, wiring, recovery catalog, atlas)

## Log

### 2026-10-02

Filed from pair#367 at the operator's direction (the survey found no bulk
claim query). pair#367 depends on this issue.

Claimed and planned. Finding: a malformed claimant fails the whole tracker
snapshot parse (`validateCardScalar` → `parseClaimant`), so an unreadable owner
surfaces as `tracker: unknown` with `issues: []`, never as "no claims"; the
per-entry `unknown` path is kept and pure-tested.

Operator review of the first plan: the question is local workspace state, so
the observation moves into `sdlc fleet inventory` (no new verb); only claimed
issues; orphaned claims are `dangling_claims`; a malformed card must not fail
the whole tracker read (folded in as M1).

## Revisions

- 2026-10-02 — operator review. Scope delta: home moves from a new
  `issue claims` verb to `sdlc fleet inventory`; unclaimed and other-machine
  issues are out; adds `dangling_claims`; adds per-card quarantine of malformed
  tracker cards (M1). `## Done when` rewritten to match; the original contract
  is in commit afe9485a.
