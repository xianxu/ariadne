---
id: 000560
status: codecomplete
deps: []
github_issue:
created: 2026-10-22
updated: 2026-10-22
estimate_hours: 1.61
collaboration-mode: collaboration
card_mirror: '7578fce03bf6e0381b2fd313fb8be81c7f3a01bc' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-22T15:10:08-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 0.61
---

# Suite queue: landing grants starve old waiters; bound grants by waiter age

## Problem

## Spec

Landing grants (`sdlc landing grant`, fleet-blocking priority) jump the machine-wide suite queue with no bound, so a
run of grants plus ariadne's weight-2 suite can starve small suites indefinitely. Evidence (ops-echo:0, 10-22): the two
oldest tickets (88, parley.nvim-slot3 #409, queued 11:58; 90, echo-slot1 echo#4, queued 12:01) waited about 2.5 h while
newer granted tickets went first (pair#773 twice, ariadne#552, ariadne#555), each granted for a good local reason.

Fix: **aging bounds grants.** A granted ticket jumps only waiters younger than an age limit (e.g. 60 minutes); any
waiter older than that is admitted first, in FIFO order, before further granted tickets. `sdlc queue status` shows
each waiter's age and when aging overrode a grant. The grant still beats everything younger, so it keeps its purpose.

## Done when

- With a ticket older than the age limit waiting, a newly granted ticket is admitted after it, not before (test).
- Granted tickets still jump younger waiters (test).
- Queue status names the aging override.

## Plan

- [x] Approve the durable plan: [bounded grant aging](../plans/000560-suite-queue-landing-grants-starve-old-waiters-bound-grants-by-waiter-age-plan.md).
- [x] Implement and verify pure protected-prefix ordering and immutable ticket age.
- [x] Add shared read-only queue inspection/status with actual admission and compatibility coverage.
- [ ] Complete conformance, recorded suite, review and landing.

## Log

### 2026-10-22
- 2026-10-22: closed — BR-1 actual-inspector provider-tree cancel/deadline regression RED7.063s, then fixed cancel/deadline/completion/provider-timeout and broad queue/suite/helper Go tests PASS77.260s; Python23 ownership/core/resource/evidence PASS0.689s, pause18 PASS4.222s, golint PASS1.694s. Clean c66fa30e12d1d12dd3897446f20266f71d444322 read-only live status and private weighted/renewal/watcher cohort PASS9.283s; raw evidence in workshop/plans/000560-supervised-inspection-evidence.md. Fresh recorded suite lease-3147671628 queued on pushed37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47, not claimed passing. Old lease1847011191 intentionally cancelled before execution. tl-worker permits close-review overlap; no-plan-check acknowledges pending suite/landing checklist, with recorded final-code pass still required before merge.; review verdict: SHIP

### 2026-10-22 — claimed, design checkpoint

#555's retained merge runner completed successfully before this item began; workspace returned to main-slot1. Claimed #560 and ran start-plan, now on its independent issue branch from main. Read current suite helper, grant provider, queue status and private helper/provider tests. Current effective_order is unconditional grant-first stable partition, which explains repeated-grant starvation; existing started is naive wall time from before identity lookup and is unsuitable for elapsed queue age. Keep fairness in Python admission, leaving Go grant validity and repository landing GrantHold unchanged (ARCH-DRY). A read-only subagent audited the same seams and recommended a protected prefix to avoid malformed-age inversions. No code changed. TL approved the policy/60-minute threshold; complete plan is now ready for approval, including shared status projection and bounded private conformance.

### 2026-10-22 — full plan approved; rollout on next helper start

TL approved the complete plan (delivery b7226f49-4802-4e84-a68e-dd94c93e60c5): one named setting/projection; text ages, limit and override; frozen JSON unchanged; additive monotonic member with unknown ages protected; no new durable history, hot reload or forced cutover. Running helpers pick up the new behavior at their next process start; the operator's planned full shutdown tonight restarts them. This issue does not initiate that shutdown. Enter full-flow change-code before implementation because the scheduler, bounded observer and status integration exceed the quick-flow code shell.

### 2026-10-22 — change-code admitted; red-first implementation

Plan-quality CLEAN disposed PQ-1/PQ-2; full-flow change-code exited0 after estimate-quality INFO. The estimate is optimistic about verification: implementation primitives include their targeted/private conformance runs and the review primitive includes close/landing attention; recorded-suite elapsed time remains an acknowledged unallocated estimate risk. No accounting overrides. Python policy/inspection and Go observer/rendering are isolated file ownership. Old grant-first regression failed with actual order[2,1] versus protected[1,2] at the60-minute boundary. New Go observer tests initially fail to compile because the observer API does not yet exist. Keep Go test runs after concurrent checkout edits stop, per the hermeticity guard.

### 2026-10-22 — targeted verification and ad-hoc review

Python and Go implementations are complete. New final-watcher lifecycle regression first failed because acquired slot metadata lost enqueue_monotonic; restoring the admitted record also preserves unknown nested members. The Go successful-parent descendant test failed under a deliberate cleanup-removal mutation (pid still running), then passed with cleanup restored. Main verification: Go queue/suite/helper targeted selection PASS101.305s; independent golint package PASS1.145s; Python age/resources/evidence22 tests PASS0.676s; pause18 tests PASS4.310s. The Go selection includes the real Python helper/production Go provider harness. Fresh-eyes review found no blocking production defect and asked for fixture joining plus successful-parent descendant coverage, both now added. Supplementary inspection rejects persisted nonfinite JSON explicitly rather than emitting invalid JSON; it does not rewrite that record. No machine queue or grant was altered. Next: clean built-binary read-only status and private cohort conformance, then pushed-head recorded suite and mandatory close review.

### 2026-10-22 — clean binary conformance passed; suite checkpoint

Built clean d7136fdce2100bff36eb271c9a4f00bd455c6232 (vcs.modified=false). Built-binary read-only queue status exited0, showed GRANT_MAX_WAIT_SECONDS=3600 (60min), sixteen live waiters with unknown legacy ages and their protected FIFO prefix through ticket130. That observation does not retrofit the old waiting helper processes. The private mixed-weight/renewed-grant cohort and final watcher metadata conformance both passed (2 tests,8.314s) using that clean binary as production provider. Raw build/status/completion evidence and source blob pins are committed in [suite aging evidence](../plans/000560-suite-aging-evidence.md). No live mutation, stage or forced cutover. Next: push this checkpoint and queue the recorded suite on its exact head; close/review after the suite, then enter merge on SHIP.

### 2026-10-22 — BR-1 real-inspector cancellation ownership

Close round1 returned REWORK on BR-1 (Critical, subprocess-cancellation-ownership). Independent review and a new production-path regression confirmed the provider escapes the inspector's Go-owned group through start_new_session=True. Real inspect cancellation/deadline left both provider and child running (RED7.063s); completion and provider-timeout controls passed. Supported cancel retired queued suite lease-1847011191 before execution; no passing suite is claimed and no same-head retry is attempted. TL approved supervised inspector ownership (delivery8b559327-691f-45b5-96c7-c1f7bc383c1e). Fix under way; fresh verification and suite required on fixed code.

### 2026-10-22 — BR-1 fixed and verified

The approved supervised mode now keeps provider descendants in the external Go-owned helper group; standalone queries still own detached groups, and the existing SIGTERM-unwind handler covers both standalone commands. The actual-inspector cancel/deadline/completion/provider-timeout regression and broader Go suite/queue/helper selection PASS77.260s. Python ownership/core/resource/evidence23 tests PASS0.689s; pause18 tests PASS4.222s. New review required to dispose BR-1; no close finalized yet. Old queued suite is fully retired, terminal63074 exited1 as intentional cancellation before execution. Next: clean build and conformance refresh, push fixed checkpoint, start one fresh recorded suite, then repeat close review with that suite queued.

### 2026-10-22 — fixed clean binary conformance refresh

Clean build c66fa30e12d1d12dd3897446f20266f71d444322 (modified=false) passed read-only live status again, now using explicit supervision. The same two private real helper/provider conformance cases passed9.283s. Current raw build/status and source pins are retained in [supervised inspection evidence](../plans/000560-supervised-inspection-evidence.md); this supersedes the earlier runtime capture after BR-1. golint PASS1.694s. Both short conformance terminals completed normally across the Couch restart. Next: push this fixed checkpoint, queue one fresh recorded suite, and rerun close review while it waits. Old suite lease1847011191 remains intentionally cancelled, not passing evidence.

### 2026-10-22 — SHIP; recorded suite outstanding

Close round2 returned SHIP with no findings; BR-1 addressed. Reviewer independently restored detached provider sessions in a disposable copy and reproduced surviving provider/child failures, then removed the copy. Close evidence commit9c821802139fd7259408a299fa49e6362208df39; card mirrord9096385cede775383ffff8b49248c18bf23edc1. Required recorded suite lease-3147671628 remains alive on exact pushed37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47 (retained terminal98827), last FIFO position16. Log: common Git dir sdlc/test-stats/inflight/f615ddb6dbd25ea386315b2508b44e3a/run.log. No suite pass claimed; close overlapped the queued check under tl-worker. Next: open PR for CI overlap, keep this queued runner, inspect its exact-head result when it finishes, then enter sdlc merge --yes after the pass. Merge owns integration and any fresh authoritative suite. Do not requeue a duplicate or restart while the runner remains queued/running.

### 2026-10-22 — shutdown checkpoint; existing merge retained

Branch suite lease-3147671628 completed PASS on37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47: 8 shards,2302 SDLC tests,76 other packages; wall858.5s after6558s queue wait. Saved log: common Git dir sdlc/local-checks/37a5bfda4a9c87c1ca1a6e722334f168c2c4fc47/suite.log (sha256 prefix11f820f89bf1). PR404 is open, SHIP and server CI green. Existing merge pid93795/retained terminal23091 is first waiter behind active#556 as of18:31PDT; it has not been restarted. No new merge/landing or item is started under the fleet shutdown notice (deliverya7e9b5cc-2445-454e-a42a-5368bfcdc7fa). An existing admitted/turn-holding runner may finish well before21:45; otherwise retain its queued state for the22:00–23:00 shutdown. No source changes after reviewed conformance.

Exact next step: inspect retained terminal23091 and sdlc landing status/PR404 before doing anything. If the existing runner is alive, retain it and let its gate own main integration and authoritative suite. If it lands, verify the merged PR and Git-reported merge SHA and report to the hub. At shutdown/restart, verify whether the runner/turn survived before any re-entry. Only after authorized recovery, re-enter sdlc merge --yes if the original runner is actually gone/refused and PR404 remains unmerged; never create a duplicate. Do not start new items. Current issue/plan/evidence are committed and pushed for recovery.

## Revisions

### 2026-10-22 — approved aging policy and60-minute threshold

TL approved immutable monotonic enqueue age under the allocation lock, FIFO prefix protected through the last at-bound/older ticket, and current grant ordering only in the younger suffix. Renewal never resets age; missing or invalid age protects its prefix. The threshold is60minutes, selected because the machine routinely runs an hour deep;30minutes would suppress nearly all useful priority. Preserve running holders and weighted all-or-none admission. The durable plan proposes one named Python policy setting shared by admission/inspection, with queue text status rendered from that observer; frozen --json remains unchanged. Full-plan approval precedes code.

### 2026-10-22 — plan-quality refinement

First change-code review accepted the design but blocked on PQ-1 (function-level test contract). Consolidated repeated case inventories into one strategy/independent guard per named risky function. Corrected PQ-2: existing holder reads are bounded, existing ticket/slot writers are not; this issue bounds inspection without expanding writer scope. No production code changed; rerun plan-quality.

### 2026-10-22 — approved supervised observation lifetime

BR-1 extends the already-required bounded observation lifecycle across the real provider seam. The Go-launched inspect command is explicitly supervised; its provider and descendants inherit the Go-owned helper process group, matching the existing drop-cost pattern. Python supervised cleanup kills/reaps the provider leader without targeting its shared group; Go unconditionally collects that whole group on return, cancellation or deadline. Standalone callers retain their detached provider cleanup. Add actual-inspector cancel, deadline, completion and provider-timeout regressions. No new scheduling policy, durable state or status JSON field. TL approved this refinement before implementation.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only. Calibration is flagged stale by estimate-source; values are provisional. Existing flock, provider and process-group seams cover this familiar stack; no new library or external API is needed. Map the bounded Python policy extension to smaller-go-module (the method has no separate Python primitive), the inspection protocol to greenfield-go-module, and CLI wiring to another smaller-go-module. Use upper implementation ranges for concurrency/process coverage, scaled once by0.40. Detailed approved design receives the0.2 design discount; familiarity1.0 and15% design buffer.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: issue-spec design=0.1 impl=0.08
item: smaller-go-module design=0.06 impl=0.2
item: greenfield-go-module design=0.2 impl=0.32
item: smaller-go-module design=0.06 impl=0.2
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0 impl=0.2
design-buffer: 0.15
total: 1.61
```

Design0.46×1.15=0.529h plus implementation1.08h gives1.609h, rounded1.61h.
