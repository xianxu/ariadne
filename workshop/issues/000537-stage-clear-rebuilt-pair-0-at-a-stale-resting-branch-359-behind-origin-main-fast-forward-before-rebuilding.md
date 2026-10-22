---
id: 000537
status: working
deps: []
github_issue:
created: 2026-10-21
updated: 2026-10-21
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b654ee88a92bce68140fe519dfc6d4b825fde4da' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-21T18:55:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:2
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot2/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "3f42d20a", done: "69e11376"}
---

# stage --clear rebuilt pair:0 at a stale resting branch (359 behind origin/main); fast-forward before rebuilding

## Problem

## Spec

10-21 18:40, pair: after #752 landed, the stage holder ran `sdlc stage --clear` on pair:0 (the operator's live
checkout, resting branch `main`). It returned pair:0 to its local `main`, which was 359 commits behind
`origin/main` (last fast-forwarded 02:13), and rebuilt and verified the Go binaries **at that stale commit**. If the
operator had restarted Couch then, they would have run a day-old build, losing everything that landed today. The
TL caught it by checking `git status` and fast-forwarded plus rebuilt by hand.

Fix: `stage --clear` fetches and fast-forwards the resting branch to `origin/main` before rebuilding (it's a resting
branch: fast-forward only, never committed on). If it can't fast-forward (diverged or dirty), refuse and name the
state instead of building a stale tree. Dry run fetches read-only and reports the resting branch's lag; only a failed fetch makes freshness unverified.

## Done when

- `stage --clear` on a target whose resting branch is behind origin/main ends on origin/main's tip, with binaries
  built from it (test with a fake remote).
- A diverged resting branch makes `--clear` refuse before building (test).
- Dry run fetches and reports behind/ahead counts without checkout, lease or build effects; failed fetch reports freshness unverified.

## Plan

- [x] Add production-path regressions using isolated Git slots and a bare remote: advance remote main while the target rests behind, then clear a stage and assert checkout plus built revision equal the freshly fetched tip. Repeat from an already-resting target (including retry after rebuild failure); cover diverged/ahead, dirty, offline and dry-run controls with no stale build.
- [x] Route both clear paths through one pinned-main restoration/build sequence under the existing stage lock. Fetch once with gitx.TrunkFile.Snapshot, reject a resting head not contained by that pin before switching, preserve asset and untracked/ignored-file protections, fast-forward only, verify exact checkout and materialized assets, then rebuild against the pin. Fail closed on fetch/ancestry/checkout verification failure; no stale fallback.
- [x] Update stage help, README and atlas; run targeted stage/asset controls and the recorded full suite, close, and publish. Ask the hub whether the isolated production-path proof suffices instead of touching a live :0.

Design: runUnstage is INTEGRATION glue, not a pure function. Reuse gitx's context-bound snapshot fetch and existing asset verification/build seams (ARCH-DRY, ARCH-PURE); do not turn fastForwardRest's warning-only policy into an ignored failure. The existing stage lock spans observation, restoration and build, so no new ownership resource or durable format is needed (ARCH-ORDER, ARCH-EVOLVE). This is an operator-started batch: one fetch and one fast-forward, no polling/retry loop or new hot path. The remote pin is a snapshot, not a promise main cannot advance afterwards.

Actions: clear, repair local Git state/network/build, retry clear.

| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Clean resting head contained by fetched main | Cleared and rebuilt at named pin | None | None; fast-forward preserves history |
| Dirty, ahead/diverged, or failed fetch/probe before checkout | Refused, no rebuild; target and lease retained | Repair named Git state/network, retry clear | None |
| Checkout/asset verification fails after a local effect | Refused with observed partial progress; no rebuild | Inspect/repair target, retry clear | No reset or rollback; retain existing work |
| Restored pin, rebuild fails | Resting checkout named, build unverified | Repair build, retry clear (fetches again) | Existing binary may remain stale until success |
| Dry run | Fetched behind/ahead counts and proposed clear, or freshness unverified if fetch fails | Run clear after any required repair | None; no checkout, lease release or build |

Covered: both detached and already-resting clear paths, failed-build retries, refusal before stale build, existing stage ownership/assets. Not covered: unrelated move/merge restoration, automatic tool rollout, concurrent manual Git writers outside SDLC; existing stage lock only serializes cooperating operations. Tests use real private Git resources and the existing build-info seam; no live fleet mutation. The invariant depends on ancestry and effect ordering, so behavioral regressions defend it rather than a lexical lint rule that merely names a helper.

## Log

### 2026-10-21
- 2026-10-21: closed — Bare-remote regressions failed before the fix and pass after: detached/resting clear and numbered rest build freshly fetched main; ahead/diverged, dirty and failed-fetch states refuse before build; dry-run reports fetched lag or unverified without checkout/lease/build changes; failed-build retry fetches newer main. Pin survives later remote advance; incoming/resting attributes and untracked/ignored files are protected. Stage/move/lease/asset, golint/shelllint and recovery controls pass. Recorded suite lease-2876290610 on pushed 4181f126: 2047 SDLC tests +75 packages, zero failures, 779.3s, log 92f3d39a9cc7; server checks green. Hub approved isolated production-path proof, no live :0 touch; landed-holder authorization filed separately as #538.; review verdict: SHIP

- 2026-10-21: Claimed and started from fresh main while #534 remains on its separate retained issue branch awaiting the operator label. Inspected both clear paths: each rebuilt to.Head/to.RestHead without a remote fetch. The existing stage lock already spans clear/build; gitx.TrunkFile.Snapshot fetches into the exact remote-tracking ref and pins subsequent asset reads. Plan is quick-flow sized and stops for hub review before code, as dispatched.

- 2026-10-21: Approved change-code entered quick flow. Production-path bare-remote regressions failed red on both stale builds, ahead/diverged and offline states accepted before build, and missing dry-run lag/unverified reporting. Dirty refusal already held. Shared runUnstage now snapshots fresh origin/main, reports ahead/behind, rejects local-only commits, validates asset transitions, restores rest and fast-forwards with no autostash/ignored overwrite, verifies exact pin/assets, then rebuilds at that pin. Stage lock spans the complete sequence; both detached and already-resting paths share it.
- Targeted proof: new freshness group passed (22.303s). Broader stage/move/lease/asset/rebuild controls passed (137.657s). Expanded stage/asset controls had only a new pinning-fixture setup failure (remote objects unavailable at the intended hook boundary); corrected object delivery in the private bare-remote fixture, not production behavior. The exact pinning test then passed (2.452s): post-checkout advanced remote main, checkout and built stamp remained on the captured pin. Numbered main-slot2, failed-build retry fetching newer main, and untracked/ignored collision controls passed in the expanded run. Golint and shelllint full packages passed (0.823s / 0.343s); diff check clean. Full suite still pending.
- Hub authorized separate filing for the landed-holder case if not small. Filed unclaimed #538: clearing from rest after branch deletion needs confirmed durable landing evidence and original-slot binding; missing branch alone must not authorize clearing. No stage-authority change in #537. Hub accepted isolated production-path proof for #537; no live :0 touched.

- Help proof: go run ./cmd/sdlc stage --help printed fresh-main clear, dry-run lag and recovery guidance (the expected non-fleet-binary warning is from running a temporary Go executable, no :0 mutation). Recovery catalog package passed (0.127s). PR #381 opened; initial server checks green on bdeab1657947efadadf18342577b471e469ddb6d.
- Asset-transition edge found before acceptance: incoming assets can be absent from the parked resting tree while its attributes still affect the intermediate switch/merge. The new real-Git regression failed because only staged/current and incoming attributes were checked; added the resting ref to incoming-asset preflight. Regression passed (3.849s), then all clear/asset/fast-forward controls passed (50.702s). No new authority or mechanism: preserves the existing materialization contract across the added fast-forward.
- Cancelled only queued lease-3908965038 on bdeab165 via local-check cancel --run after proving the code correction was needed; waiter exited with context-canceled before suite execution, not passing/failing test evidence. Preparing the corrected pushed head for a replacement recorded suite. No other run cancelled; no continuation restart. #534 remains waiting for its operator label.

- Recorded full suite PASS: lease-2876290610 on pushed 4181f12647ce2ae36991cce6ddc1051d5456a88c, retained through 1,287s FIFO wait; 2,047 SDLC tests and 75 other packages, zero failures, wall 779.3s (13m0s), log sha256 prefix 92f3d39a9cc7. Saved under .git/sdlc/local-checks/4181f12647ce2ae36991cce6ddc1051d5456a88c/suite.log. No flake retry/resume needed. Both GitHub merge-check and local-gates guard passed on that head.
- Continued after #534 landed as PR #380: returned to the unchanged tested #537 head and clean checkout. Implementation/test evidence is grouped under Log; the approved design delta remains in Revisions. No code changes after the recorded suite. Preparing the mandatory close review, with hub-approved isolated production-path proof and no live :0 mutation.

## Revisions

- 2026-10-21 18:58 PDT — Hub approved plan and isolated production-path proof, no live :0 touch. Dry-run policy revised from no fetch/unverified to a read-only fetch plus behind/ahead counts; retain unverified only on fetch failure. All other scope and recovery behavior approved unchanged.
