---
id: 000343
status: working
deps: []
github_issue:
created: 2026-10-12
updated: 2026-10-13
estimate_hours:
card_mirror: 'f729b3cd57a2f34b387532155c31f739e9e463a1' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-13T14:56:25-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
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

- [ ] `staging.AwaitExclusive(stage, within)`; `finishProducers` and `Reclaim` use it.
- [ ] Unit tests in `staging`: a dead owner's stage whose lease is released during the
  settle is reclaimed; one whose lease stays held is preserved after the bound. Both fail
  with the fix reverted.
- [ ] Remove temporary diagnostics; 100 targeted runs of the original test pass.

## Log

### 2026-10-12

### 2026-10-13 — reproduced and confirmed (ariadne:3)

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
