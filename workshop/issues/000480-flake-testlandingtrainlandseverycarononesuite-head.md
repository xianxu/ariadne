---
id: 000480
status: working
deps: []
github_issue:
created: 2026-10-19
updated: 2026-10-20
estimate_hours:
collaboration-mode: collaboration
card_mirror: '3f88e2c9a03f28faf4acfbf8fbd44cbb5356ed98' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-20T00:35:07-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flake_tests: [github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head]
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/head

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head: Same exact head 184cd3a42a421c9eea1309871204b7d05aeeb890: recorded suite attempt fc26a661e6e90915454776817ced8e68 completed all 12 shards and 72 other packages, with only TestLandingTrainLandsEveryCarOnOneSuite/head failing at landingtrain_test.go:410 (finished helper two-minute merge timeout; head variant 139.59s). Saved failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fc26a661e6e90915454776817ced8e68.log, sha256 prefix fad0a6acdb5f. Without any source or HEAD change, focused go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 21.16s, null 20.01s, head 22.64s, total test 63.81s; package exit 0 in 64.342s. HEAD verified by git rev-parse before and after. Focused run is outside test-stats; this is explicit same-head fail-to-pass evidence. Register only /head, not its passing siblings.

## Spec

Diagnose the load-coupled train completion failure before choosing a fix. Capture goroutines at about 60 seconds with at most four concurrent diagnostic test processes; timestamp repository-lock acquisition/release and identify the owner and production window. Compare normal and bounded concurrent runs. The hub reports the same failure at load 70+, and #484 separately investigates the suite's 12/8/6 shard operating envelope (single-suite load 150+). Do not assume a deadlock or enlarge the timeout without evidence.

ARCH-CONSTRAINTS/PURPOSE: cap diagnostic concurrency at four, record load and window duration, and distinguish CPU starvation from slow work unnecessarily protected by the lock. ARCH-ORDER: preserve transaction and checkout revalidation guarantees if a lock window changes. ARCH-DRY/PURE: reuse any existing suite-budget helper if the evidence favors its timeout policy. ARCH-MOCK: exercise real Git/locks through the existing stateful GitHub fixture. ARCH-SECURE/EVOLVE: diagnostics use fixture repositories; no persisted production format or authority changes are planned. ARCH-FUNERAL: temporary instrumentation and logs are bounded to this investigation and removed before the final diff.

## Done when

- The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.

## Plan

- [x] Capture a timed goroutine dump and lock-window timings under bounded load; report the observed window and remaining uncertainty to the hub before choosing the fix.
- [ ] Implement the supported correction with regression coverage and an appropriate mechanical lint guard; revise acceptance if the approved timeout-budget alternative is selected.
- [ ] Verify focused/race checks and the recorded suite, close through fresh review, then land and report the outcome.

## Log

### 2026-10-19

### 2026-10-20

- Claimed after #470 landed; branch starts at main eff65960043040be38d0efb3dfb59c95ea1b8723. The fixed two-minute finished() helper bounds the whole concurrent merge, while each production Git command has its own two-minute ceiling. No suite-load scaling helper was found in the current test fixture or sharding runner. Initial load averages 3.60/22.38/28.70; capture actual windows rather than infer the owner from a lock-wait message.

- Diagnostic: temporary wrapper timestamped every real repository-lock wait/acquisition/release; a one-shot process timer dumped goroutines at 60 seconds. Baseline TestLandingTrainLandsEveryCarOnOneSuite passed all variants in 20.17s (6.70/6.72/6.75s). Its six final cleanup lock holds were 2.092–2.149s, totaling 12.741s. The other locked windows were much shorter.
- Started four test processes, each running the same parent three times. The existing targeted-test gate admitted three concurrently and queued the fourth; never exceeded four diagnostic processes. All 12 parent invocations / 36 variants passed, parent durations 22.70–35.74s. All four process dumps were captured at 60 seconds. Initial/final host load averages were 5.43/14.68/24.34 and 12.25/14.67/22.68; a mid-run sample was 16.36/15.64/23.39.
- The dominant window is runDurableMerge's final f.withLock at landing.go:652: integration confirmation, card completion, archive publication/confirmation, checkout/ref cleanup. Across 72 bounded-run cleanup holds it ranged 2.190–4.950s. Three 60s stacks show this window waiting for Git: archiveReadContext → archiveBlobContext → archiveSnapshotContext → confirmLandingArchive → landingArchiveComplete (landing.go:695), or resolveLandingTarget → revalidateLanding → returnLandingToRest (landing.go:704). No cyclic wait or single stuck Git command was observed. The fourth dump and all lock timings are retained locally with the others.
- Evidence is consistent with CPU/process starvation amplifying a serial Git-heavy cleanup window, but bounded load did not reproduce the original two-minute stall and does not prove its exact cause. Do not claim a deterministic lock bug or a definitive starvation diagnosis. Report this distinction and the #484 envelope alternative to the hub before selecting the correction. Local diagnostic directory: /Volumes/TestScratch/ariadne-480-a4f2_reb (bounded logs, script, instrumentation copy, compiled test); all temporary repository instrumentation was removed and the source diff is clean.
