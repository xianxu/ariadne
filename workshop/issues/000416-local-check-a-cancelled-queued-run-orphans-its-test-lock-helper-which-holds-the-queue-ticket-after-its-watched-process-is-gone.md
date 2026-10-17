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

- The Git/network preamble retains default SIGINT/SIGTERM termination; a production-handler blocked-fetch regression proves it.
- Queued cancellation still retires only the owned helper/holder/ticket; cleanup has an independent finite Git budget after cancellation.

## Plan

- [x] BR-1: scope managed signals to the owned queue/suite phase, restore default handling outside it, and give detached cleanup an independent bounded Git runner.
- [ ] Verify the unchanged production handler with blocked-fetch SIGINT/SIGTERM, queued cancellation, bounded cleanup and a fresh recorded suite before re-close; repeat the live check when the TL grants the stage.

- [x] Reproduce queued cancellation and watched-process SIGKILL with private lock files; show the successor is admitted and unrelated holders survive.
- [x] Reuse CLI signal-context ownership for local-check run, and bound cancellation of its owned helper process group; retain the existing per-run pipe holder.
- [x] Check the watched process before admission and while queued; leave through the existing ticket cleanup when it is gone.
- [x] Complete focused cancellation/helper checks, the recorded full suite and CI before close review.

## Log

### 2026-10-16

- Investigation: executeCLI only manages signals for review/calibration verbs, so local-check run receives a background context and exits before its defers unwind. acquire() checks no watched PID until after acquiring slots. #400 remains independent: this change only stops a child/group created by this run, never discovers or signals other runs.
- Design: ARCH-STATE/ARCH-FUNERAL — reuse existing per-run pipe holder, context and ticket-finally cleanup; no new durable state or cancellation selector. Cancelling during queue/acquisition unwinds only this run; sudden owner death is independently detected by the queued helper within its existing one-second poll. Expected cost: two small targeted process checks plus the required recorded suite; no extra hot-path subprocess probes.

- RED: private helper script regressions failed exactly for surviving helper exit, own ticket release, and an already-dead watched PID acquiring a free slot. Native CLI SIGINT/SIGTERM regression failed both modes with no managed cancellation (Python fix independently recovered tickets).
- GREEN: go test -trimpath ./cmd/sdlc -run '^Test(LocalCheckSignal|RunLocalCheck|CLISignal)' -count=1 -timeout 5m passed (67.045s), then the updated ownership/worktree/successor regression and all RunLocalCheck tests passed three times (44.983s). Full bash scripts/test/test-lock.test.sh passed, including queued SIGKILL, bounded helper exit, unrelated holder/successor preservation, successor admission, already-dead watch and no retained tickets.
- Implementation stays within quick flow: local-check run opts into the existing signal context; its helper uses the existing owned process-group cancellation and one-second pipe-drain bound. Interrupted acquisition preserves context-cancelled diagnostics. The helper checks liveness before allocation and each one-second queue poll, using the same liveness probe as its polling watcher. No stdout protocol state, extra durable artifact, global PID search or foreign-run signal is added.

- Live check: TL rebuilt staged ariadne:0 bin/sdlc; go version -m verified vcs.revision=005587765f51a3baf610fd3f70b1bed72f0b1a65, vcs.modified=false. Invoked that binary's real local-check command twice against a local fixture remote with a private flock and GitHub transport stub. SIGINT to the first owned queued runner returned nonzero with context-cancelled diagnostic in 1.014s, reaped its pipe holder, removed its detached worktree, and left only the queued successor's ticket while the unrelated slot remained locked. Releasing that slot admitted the successor, ran its command, and left zero tickets/detached worktrees. Evidence: /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne-416-live-x6esbb2g/{live.py,cancelled.log,successor.log}. Earlier fixture URL/setup attempt failed before queue allocation; no production failure. Stage's skipped binary sentinel is tracked independently as #417.
- Recorded suite launched on pushed 005587765f51a3baf610fd3f70b1bed72f0b1a65; run d5ecb880e188a182590c4815cf749537, detached sdlc-local-2369331256. Awaiting weighted FIFO admission; no gate bypass or peer cancellation.

- Recorded suite PASS on 005587765f51a3baf610fd3f70b1bed72f0b1a65: 12 shards / 1,524 cmd/sdlc tests plus 66 other packages, all exit 0, zero failures, wall 532.8s (8m53s). GitHub sdlc-local/suite recorded; saved log .git/sdlc/local-checks/005587765f51a3baf610fd3f70b1bed72f0b1a65/suite.log, SHA256 prefix 61836047642c. New queued-signal regression passed under suite load.
- Merged current origin/main (aa856d15a3b5037eab593919a089578df56f2a4f), including #414's independent tl-lead implementation, without conflicts. The required suite must run again for this integrated code generation; the prior pass is retained, not reused across code changes.

- Final integrated suite queued on 1b31e6fc1f288e278b1f1c69fc921726c03cd142, run 0bbae3af1b6112bb7cb13acbc505a1ca, detached sdlc-local-880418815; admitted queue position 9. The prior full pass remains valid evidence for #416's unchanged branch patch but is not presented as an integrated-head pass.

- Integrated recorded suite PASS on 1b31e6fc1f288e278b1f1c69fc921726c03cd142: 12 shards / 1,524 cmd/sdlc tests plus 66 other packages, all exit 0, zero failures, wall 635.6s (10m36s). GitHub sdlc-local/suite recorded; saved log .git/sdlc/local-checks/1b31e6fc1f288e278b1f1c69fc921726c03cd142/suite.log, SHA256 prefix 82bb97803186. CI on the current PR head also passed. All Done-when behaviors are covered by the permanent native signal/private-lock regressions and staged-binary live check; proceeding to the binary-owned close review.

- Close round 1 REWORK: BR-1 Critical reproduced that command-wide NotifyContext suppresses SIGTERM during context-free Git/network preamble. Initial queued tests replaced RunE and missed that family. TL approved the narrower resource-lifetime fix (receipt b6986b85-dc38-4def-a646-346c65ee6726): preserve default handling outside owned queue/suite effects, add the blocked-fetch case; do not generalize all subprocess seams in #416.
- Independent #420 filed unassigned for actual-time attribution. TL decision: queue waiting is not active work; exclude/subtract recorded queue_seconds because they are measured, not guessed. Both retained attempts contain 6,951.058874s queued in total; no replacement numeric actual was entered.

- BR-1 RED: unchanged production handler swallowed SIGINT/SIGTERM at blocked fetch and postrun trailer reads (four failures). After scoping signal ownership, all native signal cases passed; the added stalled-cleanup case then failed only with queued cancellation exceeding five seconds. GREEN: go test -trimpath ./cmd/sdlc -run ^Test(LocalCheckSignal|RunLocalCheck|CLISignal) -count=1 -timeout 5m passed (65.021s), including independent cleanup deadline and successor admission. Cleanup reuses boundedLandingGit with the existing recovery ceiling in all three production constructors (ARCH-DRY); no command-wide subprocess refactor.

- Integrated current main at e7a8db930a2b2bbffbb903cc4b4743ef693415e8 and pushed PR #274. Fresh recorded suite run ef42644f9ff6b93a2aa217acc8a0dbc2 is queued (private detached checkout sdlc-local-3069687343). Full bash scripts/test/test-lock.test.sh passed again, including queued watched-PID SIGKILL and successor preservation/admission.
- Staging error: I replaced ariadne:0's staged checkout without a grant while #418 owned it; #417 sentinel left bin/sdlc unchanged at 13d40cf0. TL receipt 08ff844e-92e2-4033-8ab5-00ea4936111b says leave :0 untouched, finish suite and close review, then await a grant after #417 for the repeated live check. Acknowledged; no further :0 mutation or stage clear. Prepared independent live fixture at /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne-416-br1-live-usblptdb/live.py; not run against a stale binary.

- TL granted ariadne:0 (receipt 26e429ca-504c-4f28-93e9-57008f06b226). Staged 789a2138e2ae37927e727bdd1102bea002922209, ran make sdlc-build explicitly after the known sentinel skip, verified exact vcs.revision and modified=false. Real staged CLI live check passed: queued SIGINT unwound in 1.018s, retired only its holder/tree, preserved foreign holder and queued successor, then admitted the successor with zero retained tickets/trees. Blocked-fetch SIGTERM exited with its default signal in 0.001s. Evidence: /var/folders/07/b9wcwwld4_v2w9r3hk525bm80000gn/T/ariadne-416-br1-live-usblptdb/{live.py,cancelled.log,successor.log,blocked-fetch.log}. Cleared stage, rebuilt main via make build plus make sdlc-build, verified all seven binaries at f5927dfedbfb2e6dbc7d8c91ac56085732181b7b modified=false and clean checkout; hub received free-stage report 3c21b52b-e224-4efa-8a89-cc094ffcc6cc (submitted).

- Revised suite ef42644f9ff6b93a2aa217acc8a0dbc2 was admitted after about 3,458s queued. Its queued-cancellation and bounded-cleanup native cases passed, but the new outside-phase SIGINT assertions failed (fetch and trailer); SIGTERM passed. This is a deterministic test precondition error, not an unrelated flake retry: the asynchronous suite inherits SIGINT=SIG_IGN, which Go intentionally preserves. Explicit ignored-SIGINT launch of the same compiled test reproduced both failures (9.33s); explicit SIG_DFL launch passed all four cases (2.98s). The owned fixture now sets SIGINT=SIG_DFL in an exec launcher before entering the unchanged production handler. Under an outer SIG_IGN, go test -trimpath ./cmd/sdlc -run ^TestLocalCheckSignal -count=1 -timeout 3m passed (8.359s). No production change; retain the failed suite and run a new recorded suite on the fixture fix.

- Full failed result retained: ef42644f9ff6b93a2aa217acc8a0dbc2 on e7a8db930a2b2bbffbb903cc4b4743ef693415e8, queue_seconds=3458.271406708, wall_seconds=496.335932, 12 shards / 1,525 cmd/sdlc tests plus 66 other packages. Only TestLocalCheckSignalPreservesOutsideOwnedPhase failed (the two SIGINT leaves); saved failed/ef42644f9ff6b93a2aa217acc8a0dbc2.log, SHA256 8825fc8a7eea1f45d493c0a6d7cd4d9ed939f8b475c741095d0360aeb49416f7. Fixture correction and docs-only main integration pushed at 6fe57a8f910e0f4d1b8cb8ca7f3eaa2c93b40145; fresh run c069e6e8e842f9cf532752c50aec50b7 queued at position ten, detached sdlc-local-457918715.

## Revisions

### 2026-10-16 — pre-close checklist wording

Clarify the final Plan checkbox as pre-close verification only. Close review and merge remain subsequent binary-owned gates, so checking the implementation plan does not falsely claim those future actions are already complete. No behavior or acceptance change.

### 2026-10-16 — BR-1: signal scope follows owned resources

The command-wide registration is superseded: create the signal context only after preparation, immediately around queue acquisition and suite execution; release the holder and restore signals before status/Git/network bookkeeping. Preserve the parent cancellation cause before stopping the signal context. Existing helper-group cancellation and watched-PID recovery stay. Detached worktree cleanup uses a separate bounded Git seam (the existing recovery ceiling, independent of the cancelled operation) in all production runner constructors. Outside this resource lifetime the command retains its former default signal behavior, avoiding the newly demonstrated swallowed termination across all preamble/postrun seams (ARCH-STATE/ARCH-PURPOSE).

Tests execute the original production command handler, rather than replacing RunE, and reproduce blocked fetch, queued cancellation and stalled cleanup with private fixtures. This replaces only the initial signal-ownership approach; no new protocol state, global cancellation selector or general Git/network refactor. Expected additional production code is below the quick-flow remainder; if it exceeds the shell, stop for the full workflow.
