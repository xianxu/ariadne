---
id: 000343
status: codecomplete
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours:
card_mirror: '8511038a791bbaca83fe71a1e7f14ea19fec997e' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:56:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "1620eac1", done: "293f1398"}
actual_hours: 0.30
---

# weave: TestCompileRecoversAfterProcessDeathDuringGeneration flakes ~1 in 20 on main ('owned stage survived retry')

## Problem

`cmd/weave` `TestCompileRecoversAfterProcessDeathDuringGeneration` fails about 1 in 20 runs on main with 'owned stage survived retry' (ariadne:1, 2026-10-12, while closing #265).

## Spec

Unverified hypothesis (ariadne:1): the killed helper's generator grandchild is still alive when the retry reclaims the stage. Reproduce with a targeted loop (never under machine load), confirm the cause, and fix the root (wait for or kill the process group before reclaiming), not the test's timing.

## Done when

- The cause is confirmed and fixed; 100 targeted runs pass; the fix fails the test when reverted.

## Plan

Cause (confirmed 2026-10-13, below): the retry's `staging.Reclaim` judges a dead owner's
stage on one non-blocking `flock` attempt, while the owner's orphaned producer is still
exiting and holds the inherited lease. Fix: `Reclaim` waits a bounded settle for the lease,
as `finishProducers` already does on the normal path, through one shared
`staging.AwaitExclusive` (ARCH-DRY). No signalling: a live orphan keeps its stage
(documented policy); killing it would need recorded process groups plus per-producer leases
to stay safe from PID reuse, which this race does not need.

- [x] `staging.AwaitExclusive(stage, within)`; `finishProducers` and `Reclaim` use it.
- [x] Unit tests in `staging`: a dead owner's stage whose lease is released during the
  settle is reclaimed; one whose lease stays held is preserved after the bound. Both fail
  with the fix reverted.
- [x] Remove temporary diagnostics; 100 targeted runs of the original test pass.

## Log

### 2026-10-12

### 2026-10-13 — reproduced and confirmed (ariadne:3)
- 2026-10-13: closed — Delta since a6fb7490 is the two advisory minors only: test refactor (killOwnerMidGeneration + compileLeavesNoOwnedStage shared by both killed-owner tests; both pass 10/10; e2e still fails 3/3 with Reclaim reverted to one Exclusive) and a comment in staging.Reclaim + atlas sentence stating the up-to-Settle cost per stage held by a live orphan. go test ./cmd/weave/internal/staging green. Full suite green at c2585f05 (15:14). (fixed-then-ship); review verdict: SHIP
- 2026-10-13: closed — Cause confirmed: widened-window repro (generator kills parent, closes fd1/2, holds lease fd3 0.3s) failed 5/5 with the flake message; diagnostics each time: owner PID dead (ESRCH), Exclusive = ErrInUse. Fix: Reclaim waits staging.Settle via shared AwaitExclusive (also finishProducers). Revert check: TestReclaimWaitsForAnExitingProducer fails, TestCompileReclaimsStageOfAStillExitingGenerator fails 3/3. At load<40: original TestCompileRecoversAfterProcessDeathDuringGeneration 100/100, new e2e 50/50, go test ./cmd/weave/... green; full make test green 15:14 at c2585f05 (12 shards, 0 failed).; review verdict: SHIP

100 targeted runs at load under 40: 100 pass. The window is tiny and needs load to open
(the original failure came during #265's close, under suite load). Widening it on purpose
reproduces it 5/5 with the identical message: the marker kills its parent, then
`exec >&- 2>&-; sleep 0.3` closes its output (so the test's `CombinedOutput` returns) while
keeping fd 3, the inherited stage lease. Diagnostics on each failure: owner PID dead
(`kill 0` = ESRCH), `staging.Exclusive` = "stage still has producer leases". So `Reclaim`
preserved the stage because a dead owner's producer still held the lease. In the real flake
the producer is simply exiting: the kernel closes its descriptors in order, the output
pipe (fd 1/2) before the lease (fd 3), so the waiting test can resume and retry before the
lease is released.

### 2026-10-13 — fixed (ariadne:3)

`staging.AwaitExclusive(stage, staging.Settle)` (1s, the bound `finishProducers` already
used) now serves both `Reclaim` (dead owner) and `finishProducers` (cancellation). Tests:
`TestReclaimWaitsForAnExitingProducer` and `TestReclaimPreservesALiveOrphanAfterTheSettle`
(staging, in-process flock), plus `TestCompileReclaimsStageOfAStillExitingGenerator`, the
widened-window end-to-end with a 0.1s tail (10x margin under the settle). With `Reclaim`
reverted to one `Exclusive`: the exiting-producer unit test fails and the e2e fails 3/3.
At load under 40 (15:11): original test 100/100, new e2e 50/50, `go test ./cmd/weave/...`
green. Atlas: weave.md cleanup contract.

Close review (SHIP, round 1) left two advisory Minors, both fixed: the two killed-owner
tests share `killOwnerMidGeneration` and `compileLeavesNoOwnedStage` (the e2e still fails
3/3 with the fix reverted); the up-to-Settle cost per stage held by a live orphan is stated
in `Reclaim` and in atlas weave.md.
