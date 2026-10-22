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
state instead of building a stale tree.

## Done when

- `stage --clear` on a target whose resting branch is behind origin/main ends on origin/main's tip, with binaries
  built from it (test with a fake remote).
- A diverged resting branch makes `--clear` refuse before building (test).

## Plan

- [ ] Add production-path regressions using isolated Git slots and a bare remote: advance remote main while the target rests behind, then clear a stage and assert checkout plus built revision equal the freshly fetched tip. Repeat from an already-resting target (including retry after rebuild failure); cover diverged/ahead, dirty, offline and dry-run controls with no stale build.
- [ ] Route both clear paths through one pinned-main restoration/build sequence under the existing stage lock. Fetch once with gitx.TrunkFile.Snapshot, reject a resting head not contained by that pin before switching, preserve asset and untracked/ignored-file protections, fast-forward only, verify exact checkout and materialized assets, then rebuild against the pin. Fail closed on fetch/ancestry/checkout verification failure; no stale fallback.
- [ ] Update stage help, README and atlas; run targeted stage/asset controls and the recorded full suite, close, and publish. Ask the hub whether the isolated production-path proof suffices instead of touching a live :0.

Design: runUnstage is INTEGRATION glue, not a pure function. Reuse gitx's context-bound snapshot fetch and existing asset verification/build seams (ARCH-DRY, ARCH-PURE); do not turn fastForwardRest's warning-only policy into an ignored failure. The existing stage lock spans observation, restoration and build, so no new ownership resource or durable format is needed (ARCH-ORDER, ARCH-EVOLVE). This is an operator-started batch: one fetch and one fast-forward, no polling/retry loop or new hot path. The remote pin is a snapshot, not a promise main cannot advance afterwards.

Actions: clear, repair local Git state/network/build, retry clear.
| Internal state / failure mode | Visible state | Recovery action | Possible loss |
| --- | --- | --- | --- |
| Clean resting head contained by fetched main | Cleared and rebuilt at named pin | None | None; fast-forward preserves history |
| Dirty, ahead/diverged, or failed fetch/probe before checkout | Refused, no rebuild; target and lease retained | Repair named Git state/network, retry clear | None |
| Checkout/asset verification fails after a local effect | Refused with observed partial progress; no rebuild | Inspect/repair target, retry clear | No reset or rollback; retain existing work |
| Restored pin, rebuild fails | Resting checkout named, build unverified | Repair build, retry clear (fetches again) | Existing binary may remain stale until success |
| Dry run | Would fetch main, fast-forward and rebuild; freshness not verified | Run clear | None; no fetch, checkout, lease release or build |

Covered: both detached and already-resting clear paths, failed-build retries, refusal before stale build, existing stage ownership/assets. Not covered: unrelated move/merge restoration, automatic tool rollout, concurrent manual Git writers outside SDLC; existing stage lock only serializes cooperating operations. Tests use real private Git resources and the existing build-info seam; no live fleet mutation. The invariant depends on ancestry and effect ordering, so behavioral regressions defend it rather than a lexical lint rule that merely names a helper.

## Log

### 2026-10-21

- 2026-10-21: Claimed and started from fresh main while #534 remains on its separate retained issue branch awaiting the operator label. Inspected both clear paths: each rebuilt to.Head/to.RestHead without a remote fetch. The existing stage lock already spans clear/build; gitx.TrunkFile.Snapshot fetches into the exact remote-tracking ref and pins subsequent asset reads. Plan is quick-flow sized and stops for hub review before code, as dispatched.
