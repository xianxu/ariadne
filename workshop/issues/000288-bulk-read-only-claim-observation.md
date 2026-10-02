---
id: 000288
status: working
deps: []
github_issue:
created: 2026-10-02
updated: 2026-10-02
estimate_hours: 2.32
card_mirror: '2930208fffc4a9e3b657eb736e33aa8fc743e5a6' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-02T12:14:27-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
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

## Estimate

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: smaller-go-module        design=0.1  impl=0.16
item: cross-cutting-refactor   design=0.3  impl=0.2
item: greenfield-go-module     design=0.3  impl=0.24
item: smaller-go-module        design=0.05 impl=0.12
item: atlas-docs               design=0.05 impl=0.05
item: milestone-review         design=0.0  impl=0.14
item: milestone-review         design=0.0  impl=0.14
item: scope-pivot              design=0.3  impl=0.0
design-buffer: 0.15
total: 2.32
```

M1: snapshot quarantine (smaller module) + the reader/write-site audit
(cross-cutting). M2: pure `PlaceClaims` + contract (greenfield, thorough plan),
wiring + identity extraction (smaller). The scope pivot already happened in
design review.

*Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only.*

## Plan

Durable plan: `workshop/plans/000288-bulk-read-only-claim-observation-plan.md`.

- [x] M1 — a malformed card is quarantined, not fatal (snapshot quarantine, `Require`, reader audit, transferguard fail-closed)
- [x] M2 — fleet inventory reports this machine's claims per worktree (`localMachine`, pure `PlaceClaims`, contract, wiring, recovery catalog, atlas)

## Log

### 2026-10-02
- 2026-10-02: closed — Done-when met. (1) fleet inventory --json: per-worktree claims of this machine, dangling_claims, machine identity; one tracker read per repo — TestFleetInventoryPlacesClaims on the real tracker (placed; other-machine, ownerless, open omitted; removed slot dangling; one claims read across 3 worktrees), TestPlaceClaims, TestPlaceClaimsAcrossClones. (2) machine.fingerprint equals the card claimant.machine (same localMachine derivation; asserted). (3) stale/partial/unknown explicit — TestFleetInventoryStaleClaimsSaySo, TestRepoClaimsFrom, malformed fixture partial naming #42, TestFleetClaimsSkipUntrackedRemotes. (4) malformed card quarantined — TestParseSnapshotQuarantine, TestOneMalformedCardDoesNotBlockOthers across every reader. (5) sdlc help recovery lists fleet inventory [read-only] with proofs. make test green (903 cmd/sdlc tests; processgroup only fails in sandbox: /bin/ps blocked, passes unsandboxed). Live: go run ./cmd/sdlc fleet inventory --json shows #288 on this slot.; review verdict: SHIP
- 2026-10-02: closed M2 — make test green across 903 cmd/sdlc tests (processgroup fails only in sandbox). BR-11: TestFleetClaimsSkipUntrackedRemotes (untracked checkout with an unreachable publication remote -> absent, no error; verified it fails with the CutOver early return disabled); plan states the network budget (tracked repos only, one sequential fetch each). Earlier: BR-9 TestPlaceClaimsAcrossClones; BR-10 TestRepoClaimsFrom + partial in malformed fixture + other-machine/ownerless/open omitted on the real tracker.; review verdict: SHIP
- 2026-10-02: closed M1 — make test green across 901 cmd/sdlc tests (processgroup fails only in sandbox: /bin/ps blocked; passes unsandboxed). Every CardErr reader exercised by TestOneMalformedCardDoesNotBlockOthers incl. close milestone-mode die (expectDie, message names the cause); TestAssembleUnreadableCardIsUnknown; atlas notes landing deferral.; review verdict: SHIP

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

M1 implemented. Snapshot quarantine (`UnreadableCard`, `Require`), records
`CardErr`, every write-path lookup through `Require`, transferguard fail-closed,
reader audit (close/actual/push/projectstatus/issuefiles/state/issue show/observe/
fleet). E2E `TestOneMalformedCardDoesNotBlockOthers`. Side-quest:
`TestClose_MilestoneRefusesWithRedirect` ran `close` (repo lock + spine guard)
in the real checkout and died under sharded `make test` 3/6 runs with "not an
SDLC repo" — trigger not isolated (cwd/env were correct in 3 instrumented
passing runs); made hermetic (scratch SDLC repo). `make test` then green except
`processgroup` (sandbox blocks `/bin/ps`; passes unsandboxed).

M2 implemented: `localMachine` extracted (claim and fleet share it); pure
`fleet.PlaceClaims` + contract (claims, claims_state, claims_error per row;
machine and dangling_claims top-level; strict JSON); `LookupRepoClaims` over the
cached per-repository records; text rendering; catalog entry (read-only) with
proofs; help + atlas. Live run on this machine: this slot shows #288, slot2
#283, pair-slot1 #363/#367; brain repos `unknown` (remote unreachable from the
sandbox, so tracker presence is unconfirmed — reworded to say exactly that).

## Revisions

- 2026-10-02 — operator review. Scope delta: home moves from a new
  `issue claims` verb to `sdlc fleet inventory`; unclaimed and other-machine
  issues are out; adds `dangling_claims`; adds per-card quarantine of malformed
  tracker cards (M1). `## Done when` rewritten to match; the original contract
  is in commit afe9485a.
