---
id: 000452
status: codecomplete
deps: []
github_issue:
created: 2026-10-18
updated: 2026-10-18
estimate_hours: 1.63
collaboration-mode: collaboration
card_mirror: '51977fe467d980dac94250f14d80ca4e9e7fab41' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-18T18:04:52-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:3
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot3/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
actual_hours: 1.05
---

# suite queue priority: a fleet-blocking landing grant also jumps the machine suite queue

## Problem

A fleet-blocking landing grant (sdlc landing grant) orders the merge line but not the machine-wide suite queue (WF_TEST_SUITES, two slots). On 2026-10-18 the urgent pair#672 and pair#676 each waited behind ordinary suites for a machine slot even with the grant; the weekly snapshot shows a median queue wait of 11 to 20 minutes and a 90th percentile over 30 minutes.

## Spec

A suite started for an issue holding an active fleet-blocking grant (in that repo's landing line) is admitted to the machine suite queue ahead of every waiting normal suite. It never preempts or kills a running suite: it takes the next free slot. Ordering among several priority suites is FIFO. The wait message names the priority and what it waits behind. Weight (#346) is still respected. The grant's expiry ends the priority.

## Done when

- With two slots busy and normal suites waiting, a granted issue's suite starts on the next free slot before the earlier waiters (regression with the real lock helper).
- Without a grant, ordering is unchanged FIFO.
- An expired grant gets no priority.

## Plan

Durable design: [implementation plan](../plans/000452-suite-queue-priority-a-fleet-blocking-landing-grant-also-jumps-the-machine-suite-queue-plan.md).

- [x] Pin priority, FIFO, expiry, weight and lifecycle behavior with the real lock helper.
- [x] Reuse the existing grant authority and reelect weighted admission under the existing mutex.
- [x] Verify private binary smoke and focused coverage, document rollout, and start the recorded suite before close (passing final-head suite remains required before landing).

## Log

### 2026-10-18
- 2026-10-18: closed — BR-1/BR-2 accepted in round 2; BR-3 plan table/file inventory now accurately distinguish modified runner from unchanged helpercapability.go. No runtime delta since cbd3a637. Real-helper regressions and focused family pass (63.600s / 0.069s); actual own-worktree binary smoke passes all 16 tests (32.322s). Final-code recorded suite 33012b468ed484ca9808f38f207b87eb is queued; its pass remains mandatory before landing. Prior full run passed 1657 cmd/sdlc tests and failed only existing #448 directory timeout; unchanged targeted diagnostic passed (58.567s). TL authorized private smoke instead of shared staging; atlas/help updated.; review verdict: SHIP

### 2026-10-18 — diagnosis and design checkpoint

Claimed before design; start-plan prepared the issue branch from main. Read-only inspection found the machine queue's first waiter becomes a sticky head holding .queue until it acquires capacity; sorting later tickets alone cannot satisfy this issue. The proposed admission keeps the existing slot/ticket lifetime locks and uses the existing allocation mutex for each election and all-or-none scan (ARCH-PURPOSE). The other key constraint is grant authority: the existing Go ValidGrants owns expiry, revocation and authorized-grantor policy. The plan proposes one read-only Go evidence protocol, memory-only cache and a pure Python stable partition, avoiding a second grant store or copied validator (ARCH-DRY, ARCH-PURE).

Simplest alternatives considered: a startup priority boolean/expiry flag goes stale after revocation or late grants, and Python-only grant validation duplicates the current authority. Moving the entire lock helper into Go is a broader rewrite than this issue needs. Full-flow design is saved; no production code, test code or configuration has changed. TL approval is requested for unavailable-provider degradation and the existing-old-helper drain boundary before change-code.

### Implementation and focused verification

Plan-quality accepted after temporal-cache and function-strategy findings were addressed. Estimate-quality INFO accepted; module estimates include their regression work, runner estimate includes both bindings, and the review allocation includes close/landing verification. Full-suite elapsed time is machine wait/verification, reported separately by the runner.

Red: python3 scripts/test/test-lock-priority.test.py failed with "older normal head overtook active granted issue" (1.484s) against the original real helper. Provider tests also failed for absent suiteSubject/suiteGrantSnapshot before production implementation. Green: the production Go provider plus real helper passes next-free admission and the expanded integration boundary. Final focused command: go test ./cmd/sdlc ./cmd/sdlc/internal/landingline -run 'Test(SuitePriority|SuiteQueuePriority|LocalCheckPriority|LocalCheckSignal|RepeatedLocalChecksRetainRealLockOwner|SuiteSlotCoversOnlyTheRun|.*Helper.*)' -count=1 -timeout 5m passed (42.857s / 0.127s). bash scripts/test/test-lock.test.sh passed every existing FIFO/weight/reentry/death/preflight/mixed-old-helper control. git diff --check passed.

The additional real-helper harness proves unchanged-log expiry lets the older normal suite proceed; live revocation/late grants, stopped priority FIFO, no partial weighted ownership/backfill, watched-owner death, missing-provider fallback, and direct make subject/provider selection. Counted tests verify per-repository source stats, a batch per cache miss, no provider on stable polls, temporal refresh, and bounded malformed/slow/oversized transport. Local-check binding tests prove the original common directory and tested issue head, with diagnostic/non-suite/different-head controls.

Optional environment metadata supplies provider/subject rather than adding mandatory argv: older helpers retain capacity compatibility and priority failure never becomes a capacity refusal. Existing helper capability requirements stay unchanged. No protected local-gates configuration changed.

### Own-worktree binary smoke

TL declined shared :0 staging and authorized this worktree's build instead (receipt aabb93a7-728d-4289-b2fe-5a3dcaec9128). make sdlc-build completed; go version -m confirms revision 4e9ba7e62802d49e11997a8150999bdda3c4d64e, modified=false, trimpath=true. SUITE_PRIORITY_TEST_PROVIDER pointing to this worktree's bin/sdlc with python3 scripts/test/test-lock-priority.test.py passed all 13 tests in 24.615s. All lock/grant paths were private fixtures. Shared tools and live queue ownership were untouched. TL handles :0 rebuild and subsequent natural fleet observation after landing. Recorded full branch suite is running through the normal machine queue.

### Provider cancellation follow-up

Before full-suite admission, a new cancellation transport regression failed: SIGTERM of the helper left its separate provider group alive. Root cause: Python's default SIGTERM termination skipped query_priority's finally cleanup. SIGTERM now unwinds through SystemExit, so existing provider-group kill/wait and ticket release execute. The regression passed in 0.420s. The full focused family then passed again (cmd/sdlc 60.882s; landingline 0.127s), including the 14-test harness and actual local-check signal cancellation. The first suite attempt was still queued and was cancelled using sdlc local-check cancel; it ran no tests. No live peer work was cancelled. Restart the recorded suite on the corrected commit. A PID-file synchronization race in the new test was corrected to wait for complete numeric contents, not file existence.

### Recorded suite and existing flake #448

Actual-binary smoke was repeated after the cancellation fix: all 14 tests passed in 28.388s; built revision e8b34c96c6be091b199fe68f6dda37e6eef6327d, modified=false, trimpath=true. The recorded branch suite on that same head completed all 12 cmd/sdlc shards with zero failures (1657 tests) and all 70 package runs, but scripts/test TestConstructScripts/plain exceeded its existing 2m directory-wide deadline. Suite wall 772.5s. Retained failed attempt b7e282640d6d444a91c06c0b1c46c866, log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/b7e282640d6d444a91c06c0b1c46c866.log, SHA256 prefix 838fec74cba2.

This is the established open flake [#448](000448-testconstructscripts-plain-hits-its-2m-cap-when-a-review-overlaps-the-suite.md), same whole-directory timeout under load. Targeted diagnostic go test ./scripts/test -run '^TestConstructScripts$' -count=1 -v -timeout 8m passed unchanged: all 18 script/environment combinations, 58.567s total (plain 23.15s, no_color_ci 18.47s, forced_color_terminal 16.81s). Under tl-worker's self-applied flake rule, one unchanged recorded rerun started on the exact same head; attempt cb7bc0e3512890ca1143b9e60156fd46, currently queued. No timeout/gate/config/code changes made to obtain a pass. Close review may overlap that queued run under the worker contract, but a passing final-code suite is still mandatory before landing. Both attempt outcomes will be reported.

### Boundary round 1 dispositions

Close returned REWORK with BR-1 (queue-ticket-lifecycle-serialization) and BR-2 (persisted-subject-validation), both reproduced independently with deterministic regressions. BR-1: an owner releases outside alloc, so admission must tolerate a pathname vanishing after live_tickets returns. Restored FileNotFoundError handling across the snapshot/open boundary, preserving mixed-version release and the existing lock design. live_tickets already tolerates disappearance; holder reads an open descriptor. No new lock or ownership relation added.

BR-2: shared pure valid_subject now validates both caller and persisted metadata before filesystem/protocol use; absolute canonical encodable paths, no NUL, bounded byte length, and nonnegative integer issue. Malformed persisted subjects warn and stay normal; deliberate issue-zero and legacy records remain normal without spurious warnings. stat's invalid-value errors also degrade visibly. Regressions sweep malformed path/type classes and prove no filesystem call occurs.

Both production failures were red before code; both pass after, 0.118s. Expanded focused family passes again: cmd/sdlc 63.600s, landingline 0.069s, with the 16-test real-helper/provider harness. The unchanged flake rerun cb7bc0e3512890ca1143b9e60156fd46 was still queued and cancelled through local-check cancel before any test ran, because final code changed for the review fixes. Its cancellation is not a failed suite or a passing retry. A new recorded full suite and close round will use the corrected code. The original failed full attempt and #448 diagnostic remain retained.

## Revisions

2026-10-18 — Boundary round 1 implementation revision: tolerate ticket release during snapshot/read (BR-1), and apply one shared persisted/caller subject validation rule before any path use (BR-2). No new mechanism or acceptance change; deterministic regressions and focused family pass. Final-code full suite still required before landing.

2026-10-18 — Round 2 accepted BR-1 and BR-2. BR-3 corrects the plan's Core concepts table and file inventory: helpercapability.go is reused unchanged, while localcheck_run.go carries the optional subject/provider environment binding. Prose only; no runtime or suite-candidate change. Actual-binary smoke passed all 16 tests in 32.322s. Final-code suite is queued as attempt 33012b468ed484ca9808f38f207b87eb on cbd3a637.

2026-10-18 — TL approved the durable plan. Missing/old/unreadable provider warns and falls back to normal FIFO for all suites. Old sticky waiters drain naturally. Explicit expiry regression proves a granted waiter loses priority on expiry and an older normal waiter proceeds. Approval receipt: 3b55ef84-675d-4f81-834d-80f67f178da1.

## Estimate

Produced via brain/data/life/42shots/velocity/estimate-logic-v3.1.md against baseline-v3.1.md. Method A only. Source marked stale; values provisional. After plan-quality accepted: two single-concern modules (read-only Go provider; Python admission/evidence), one existing-runner binding extension, documentation and one close review. Existing ValidGrants/JSONL/flock/subprocess libraries cover the mechanisms, no novel stack dependency. Thorough spec: design values apply the 0.2 discount; implementation values are 40% of v2 table upper bounds, familiarity 1.0; 15% design buffer. The Python module maps to the greenfield single-concern primitive.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.10 impl=0.32
item: greenfield-go-module design=0.20 impl=0.32
item: smaller-go-module design=0.06 impl=0.20
item: atlas-docs design=0.04 impl=0.08
item: milestone-review design=0.04 impl=0.20
design-buffer: 0.15
total: 1.63
```
