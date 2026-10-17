---
id: 000416
status: working
deps: []
github_issue:
created: 2026-10-16
updated: 2026-10-16
estimate_hours:
card_mirror: '535418a37133202125da0a2aadbcdf6a757f13a3' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-16T19:36:21-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:5
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot5/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "b59c9a1c", done: "9912e973"}
---

# local-check: a cancelled queued run orphans its test-lock helper, which holds the queue ticket after its watched process is gone

## Problem

## Spec

Reported by pair:3 (pair#629), 2026-10-16. Cancelling `sdlc local-check run` while it was still queued for the machine test line left its `test-lock.py` helper orphaned:
- PID 42255, PPID 1, running `--watch 42240`, but process 42240 no longer existed;
- it kept queue ticket 9 for the cancelled run `sdlc-local-1087195770`, so later runs waited behind a ticket nobody would use.

pair:3 stopped only that helper and deleted no gate files.

Two gaps:
1. Cancelling a queued run does not stop its queue helper.
2. The helper's `--watch` does not notice that the watched process is gone, so the ticket is never released.

Related to, but not covered by, #400 (cancel must stop only your own run).

## Done when

- Cancelling a queued `local-check run` stops its helper and releases its ticket; a test shows the next run admitted.
- A helper whose watched PID is gone releases its ticket and exits within a bounded time; a test kills the watched process with SIGKILL and checks that the ticket is released.

## Plan

- [x] Reproduce queued cancellation and watched-process SIGKILL with private lock files; show the successor is admitted and unrelated holders survive.
- [x] Reuse CLI signal-context ownership for local-check run, and bound cancellation of its owned helper process group; retain the existing per-run pipe holder.
- [x] Check the watched process before admission and while queued; leave through the existing ticket cleanup when it is gone.
- [ ] Run focused cancellation/helper checks, the recorded full suite, close review and CI; merge on SHIP.

## Log

### 2026-10-16

- Investigation: executeCLI only manages signals for review/calibration verbs, so local-check run receives a background context and exits before its defers unwind. acquire() checks no watched PID until after acquiring slots. #400 remains independent: this change only stops a child/group created by this run, never discovers or signals other runs.
- Design: ARCH-STATE/ARCH-FUNERAL — reuse existing per-run pipe holder, context and ticket-finally cleanup; no new durable state or cancellation selector. Cancelling during queue/acquisition unwinds only this run; sudden owner death is independently detected by the queued helper within its existing one-second poll. Expected cost: two small targeted process checks plus the required recorded suite; no extra hot-path subprocess probes.

- RED: private helper script regressions failed exactly for surviving helper exit, own ticket release, and an already-dead watched PID acquiring a free slot. Native CLI SIGINT/SIGTERM regression failed both modes with no managed cancellation (Python fix independently recovered tickets).
- GREEN: go test -trimpath ./cmd/sdlc -run '^Test(LocalCheckSignal|RunLocalCheck|CLISignal)' -count=1 -timeout 5m passed (67.045s), then the updated ownership/worktree/successor regression and all RunLocalCheck tests passed three times (44.983s). Full bash scripts/test/test-lock.test.sh passed, including queued SIGKILL, bounded helper exit, unrelated holder/successor preservation, successor admission, already-dead watch and no retained tickets.
- Implementation stays within quick flow: local-check run opts into the existing signal context; its helper uses the existing owned process-group cancellation and one-second pipe-drain bound. Interrupted acquisition preserves context-cancelled diagnostics. The helper checks liveness before allocation and each one-second queue poll, using the same liveness probe as its polling watcher. No stdout protocol state, extra durable artifact, global PID search or foreign-run signal is added.

- Live check: TL rebuilt staged ariadne:0 bin/sdlc; go version -m verified vcs.revision=005587765f51a3baf610fd3f70b1bed72f0b1a65, vcs.modified=false. Invoked that binary's real local-check command twice against a local fixture remote with a private flock and GitHub transport stub. SIGINT to the first owned queued runner returned nonzero with context-cancelled diagnostic in 1.014s, reaped its pipe holder, removed its detached worktree, and left only the queued successor's ticket while the unrelated slot remained locked. Releasing that slot admitted the successor, ran its command, and left zero tickets/detached worktrees. Evidence: /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne-416-live-x6esbb2g/{live.py,cancelled.log,successor.log}. Earlier fixture URL/setup attempt failed before queue allocation; no production failure. Stage's skipped binary sentinel is tracked independently as #417.
- Recorded suite launched on pushed 005587765f51a3baf610fd3f70b1bed72f0b1a65; run d5ecb880e188a182590c4815cf749537, detached sdlc-local-2369331256. Awaiting weighted FIFO admission; no gate bypass or peer cancellation.
