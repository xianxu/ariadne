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

- One command returns every non-terminal issue's claim state for a repository
  as versioned JSON, from a single tracker read; tested with a stateful
  tracker fixture covering this-machine, other-machine, unattributed and
  unreadable-owner cards.
- This machine's identity is printable and matches what `claim` records.
- Stale/unreachable tracker yields an explicit stale/unknown state, tested.
- The command appears in `sdlc help recovery` as read-only, with its proof.

## Plan

Durable plan: `workshop/plans/000288-bulk-read-only-claim-observation-plan.md`.
New read-only verb `sdlc issue claims [--json] [--repo]`, pure assembly in
`internal/observe/claims.go` reusing the per-issue observation's judgments.

- [ ] Pure `observe.AssembleClaims` + strict claims contract (relations, unreadable owner, unknown machine, stale/unknown tracker)
- [ ] Extract `localMachine`, `trackerReadInputs`, `localWorktrees` (shared with `claim` / `issue show`)
- [ ] `collectClaims` + `sdlc issue claims` with fixture tests (this/other machine, unattributed, terminal excluded, bounded git reads, stale, unknown identity)
- [ ] Recovery catalog entry (read-only, proofs) + atlas

## Log

### 2026-10-02

Filed from pair#367 at the operator's direction (the survey found no bulk
claim query). pair#367 depends on this issue.

Claimed and planned. Finding: a malformed claimant fails the whole tracker
snapshot parse (`validateCardScalar` → `parseClaimant`), so an unreadable owner
surfaces as `tracker: unknown` with `issues: []`, never as "no claims"; the
per-entry `unknown` path is kept and pure-tested.
