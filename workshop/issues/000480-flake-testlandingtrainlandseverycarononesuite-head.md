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
flow: {kind: quick, provenance: inferred, spec: "65ac5e84", done: "40d660fc"}
---

# flake: TestLandingTrainLandsEveryCarOnOneSuite/head

## Problem

Known flaky tests (#394), registered with same-head fail→pass evidence:

- github.com/xianxu/ariadne/cmd/sdlc.TestLandingTrainLandsEveryCarOnOneSuite/head: Same exact head 184cd3a42a421c9eea1309871204b7d05aeeb890: recorded suite attempt fc26a661e6e90915454776817ced8e68 completed all 12 shards and 72 other packages, with only TestLandingTrainLandsEveryCarOnOneSuite/head failing at landingtrain_test.go:410 (finished helper two-minute merge timeout; head variant 139.59s). Saved failed log /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/failed/fc26a661e6e90915454776817ced8e68.log, sha256 prefix fad0a6acdb5f. Without any source or HEAD change, focused go test ./cmd/sdlc -run ^TestLandingTrainLandsEveryCarOnOneSuite$ -count=1 -timeout=5m -v passed all variants: train 21.16s, null 20.01s, head 22.64s, total test 63.81s; package exit 0 in 64.342s. HEAD verified by git rev-parse before and after. Focused run is outside test-stats; this is explicit same-head fail-to-pass evidence. Register only /head, not its passing siblings.

## Spec

The hub approved folding #484 into #480 after the bounded lock investigation. Measure three complete suite runs each at 12, 8 and 6 shards on a quiet machine, recording wall time, peak one-minute load and kernel share. Select the shard default and machine-suite weight from the results. Preserve landing locks and timeout assertions; the cleanup-lock finding below is a candidate follow-up, not an implementation task here. Detailed execution: [suite concurrency plan](../plans/000480-suite-concurrency-plan.md).

Diagnose the load-coupled train completion failure before choosing a fix. Capture goroutines at about 60 seconds with at most four concurrent diagnostic test processes; timestamp repository-lock acquisition/release and identify the owner and production window. Compare normal and bounded concurrent runs. The hub reports the same failure at load 70+, and #484 separately investigates the suite's 12/8/6 shard operating envelope (single-suite load 150+). Do not assume a deadlock or enlarge the timeout without evidence.

ARCH-CONSTRAINTS/PURPOSE: cap diagnostic concurrency at four, record load and window duration, and distinguish CPU starvation from slow work unnecessarily protected by the lock. ARCH-ORDER: preserve transaction and checkout revalidation guarantees if a lock window changes. ARCH-DRY/PURE: reuse any existing suite-budget helper if the evidence favors its timeout policy. ARCH-MOCK: exercise real Git/locks through the existing stateful GitHub fixture. ARCH-SECURE/EVOLVE: diagnostics use fixture repositories; no persisted production format or authority changes are planned. ARCH-FUNERAL: temporary instrumentation and logs are bounded to this investigation and removed before the final diff.

## Done when

- The bounded lock investigation and four one-minute dumps are preserved on persistent storage, with cleanup-lock optimization recorded only as a candidate follow-up.
- Nine quiet, pinned-head suite attempts (three each at 12/8/6 shards) report complete results, wall time, peak one-minute load and kernel share; failures remain visible.
- The selected default and WF_TEST_WEIGHT have a measured rationale and regression coverage; no retry loop or longer landing-test timeout is introduced. A recorded final suite includes all original train variants.
- #484 is closed as folded into this issue; atlas is current and SDLC review/landing gates pass. Closing #480 unregisters its exact flaky subtest.

## Plan

- [x] Capture a timed goroutine dump and lock-window timings under bounded load; report the observed window and remaining uncertainty to the hub before choosing the fix.
- [ ] Execute the approved 12/8/6 matrix and implement the measured suite envelope with regression coverage (see durable plan).
- [ ] Verify focused/race checks and the recorded suite, close through fresh review, then land and report the outcome.

## Log

### 2026-10-19

### 2026-10-20

- Claimed after #470 landed; branch starts at main eff65960043040be38d0efb3dfb59c95ea1b8723. The fixed two-minute finished() helper bounds the whole concurrent merge, while each production Git command has its own two-minute ceiling. No suite-load scaling helper was found in the current test fixture or sharding runner. Initial load averages 3.60/22.38/28.70; capture actual windows rather than infer the owner from a lock-wait message.

- Diagnostic: temporary wrapper timestamped every real repository-lock wait/acquisition/release; a one-shot process timer dumped goroutines at 60 seconds. Baseline TestLandingTrainLandsEveryCarOnOneSuite passed all variants in 20.17s (6.70/6.72/6.75s). Its six final cleanup lock holds were 2.092–2.149s, totaling 12.741s. The other locked windows were much shorter.
- Started four test processes, each running the same parent three times. The existing targeted-test gate admitted three concurrently and queued the fourth; never exceeded four diagnostic processes. All 12 parent invocations / 36 variants passed, parent durations 22.70–35.74s. All four process dumps were captured at 60 seconds. Initial/final host load averages were 5.43/14.68/24.34 and 12.25/14.67/22.68; a mid-run sample was 16.36/15.64/23.39.
- The dominant window is runDurableMerge's final f.withLock at landing.go:652: integration confirmation, card completion, archive publication/confirmation, checkout/ref cleanup. Across 72 bounded-run cleanup holds it ranged 2.190–4.950s. Three 60s stacks show this window waiting for Git: archiveReadContext → archiveBlobContext → archiveSnapshotContext → confirmLandingArchive → landingArchiveComplete (landing.go:695), or resolveLandingTarget → revalidateLanding → returnLandingToRest (landing.go:704). No cyclic wait or single stuck Git command was observed. The fourth dump and all lock timings are retained locally with the others.
- Evidence is consistent with CPU/process starvation amplifying a serial Git-heavy cleanup window, but bounded load did not reproduce the original two-minute stall and does not prove its exact cause. Do not claim a deterministic lock bug or a definitive starvation diagnosis. Report this distinction and the #484 envelope alternative to the hub before selecting the correction. Local diagnostic directory: /Volumes/TestScratch/ariadne-480-a4f2_reb (bounded logs, script, instrumentation copy, compiled test); all temporary repository instrumentation was removed and the source diff is clean.

- Evidence preservation: copied all cited baseline/worker logs, four goroutine dumps, orchestration script and instrumentation source to /Users/xianxu/workspace/ariadne/.git/sdlc/test-stats/evidence/000480/lock-diagnostic. Eleven files, 695959 bytes, each source/destination SHA256 verified. Manifest SHA256 209e0bf229bb07ef62275bacb70c90158509c850442c191c810ede44f419837f. This persistent copy is authoritative; the RAM directory is scratch only.
- Hub approved folding #484 into this issue: nine quiet full-suite measurements at 12/8/6 shards (three each), then choose the default and WF_TEST_WEIGHT. The final cleanup window (landing.go:652, archive confirmation at 695 and checkout revalidation at 704) is a candidate follow-up; no lock restructuring is included here.
- Hub approved independent normal weight-2 suite-queue jobs for the matrix, with other jobs allowed between runs. After admission, require idle >=80% for five seconds and load1 <=6, stopping/reporting after ten minutes without quiet. No nine-run machine reservation; the hub will arrange an operator window if needed.

- Measurement preparation: detached clean checkout pinned to 0128b68f20debb7f29f58bb62ef60bdbae6d1f84 at /Volumes/TestScratch/ariadne-480-matrix.3wjrVt. Persistent harness, raw OS conformance and local usage SKILL.md are in Git-common-dir/sdlc/test-stats/evidence/000480/benchmark. Native Mach sampler versus overlapping three-second top interval differed by 4.07 percentage points system CPU and 2.67 points idle; both raw outputs and comparison are preserved. Pure CPU calculations and admission/quiet/cancellation state tests passed (22 cases before review fixes). Fresh harness review found two Important gaps: cleanup could miss a TERM-resistant descendant when make exits first, and HEAD alone did not reject dirty source. Both are being corrected before the first measurement. No matrix suite has started.

- Harness ready: cancellation regression now covers a TERM-resistant descendant after leader exit; clean HEAD/source checks run before/after measurement. Per-run private scratch prefers TestScratch and is cleaned; Go/hardware metadata and shared LPT timing-file hashes are recorded. All 27 focused Python tests pass; fresh code-quality review approved after fixes. Local SKILL usage baseline/retest covers run order, refusal, cancellation and reboot recovery. Harness digests are frozen in benchmark/harness-manifest.json; no production source changed. Starting the first independent weight-2 queue job (12 shards); a nonquiet deadline stops the matrix for hub scheduling.

- Matrix run 01 (12 shards), complete exit 0: wall 578.332s, peak load1 68.222, kernel/all 52.419%, kernel/busy 53.351%, idle 1.747%; quiet reached in 9.064s after 652s normal queue wait. All manifest file hashes verified. Evidence benchmark/results/run-01-12-a1; no failed Go events. Continuing approved order with an independent 8-shard queue job.

- Matrix run 02 (8 shards), complete exit 0: wall 583.718s, peak load1 29.414, kernel/all 42.176%, kernel/busy 53.881%, idle 21.724%; quiet reached in 109.765s after 583s normal queue wait. All manifest file hashes verified; no failed Go events. Evidence benchmark/results/run-02-8-a1. Read-only matrix summary utility and four synthetic tests also pass fresh review, with hashes in benchmark/summary-manifest.json. Continuing with an independent 6-shard queue job; no default selected from these two observations.

- Matrix run 03 (6 shards), complete exit 0: wall 747.421s, peak load1 24.426, kernel/all 43.458%, kernel/busy 52.563%, idle 17.322%; quiet reached in 83.555s after 1044s normal queue wait. All manifest file hashes verified; no failed Go events. Evidence benchmark/results/run-03-6-a1. First cycle shows a throughput/load tradeoff (6 shards about 29% slower than 12); no candidate chosen before all repetitions. Continuing approved second cycle 8/6/12 as separate queue jobs.

- Matrix run 04 (8 shards), complete exit 0: wall 599.859s, peak load1 33.847, kernel/all 49.223%, kernel/busy 54.679%, idle 9.978%; quiet reached in 158.078s after 557s normal queue wait. All manifest file hashes verified; no failed Go events. Evidence benchmark/results/run-04-8-a1. Continuing approved order with run 05 at 6 shards.

- Matrix run 05 (6 shards), complete exit 0: wall 819.640s, peak load1 39.812, kernel/all 39.778%, kernel/busy 51.517%, idle 22.787%; quiet reached in 213.532s after 1223s normal queue wait. All manifest file hashes verified; no failed Go events. Evidence benchmark/results/run-05-6-a1. Read-only fleet resource snapshot benchmark/run-05-fleet-snapshot.json has partial CPU attribution; these are whole-machine metrics under normal weight-2 suite isolation, not proof that every peak belongs exclusively to this suite. The full-suite gate excludes concurrent suites, while other fleet activity remains possible. Continuing approved order with run 06 at 12 shards.

- STOP / waiting for a scheduled window: run 06 acquired both machine slots after 3739s in the normal queue, then refused after 600.982s without the approved quiet baseline. It never launched a suite; its private scratch directory was removed and make exited. Longest qualifying interval was 4.031s (five required). During the last minute load1 peaked at only 4.977, but CPU idle averaged 67.713%, so the idle condition still failed. Preserve benchmark/results/run-06-12-a1 (manifest SHA256 43ac11b808b9cbee66e7a074c3bf0d5743180d97d0773de986eced0c362af11e), benchmark/run-06-quiet-analysis.json and benchmark/matrix-so-far.json. All cited payload hashes verify. Measurement checkout remains clean at Git-verified 0128b68f20debb7f29f58bb62ef60bdbae6d1f84.
- Per the hub's stop rule, do not launch another attempt until the hub schedules a suitable window. Resume matrix position 06 at 12 shards in a fresh run-06-12-a2 directory, then positions 07/08/09 at 6/12/8 shards. Keep all completed and refused attempts. If reboot removes the RAM checkout, recreate the same pinned revision using benchmark/SKILL.md. No shard default or weight decision has been made; production code remains unchanged.

Partial matrix (whole-machine measurements; all five complete suites passed 1792 cmd/sdlc tests plus 72 other packages and every original train variant):

| Run | Shards | Wall seconds | Peak load1 | Kernel / all CPU % | Kernel / busy CPU % |
|---:|---:|---:|---:|---:|---:|
| 01 | 12 | 578.332 | 68.222 | 52.419 | 53.351 |
| 02 | 8 | 583.718 | 29.414 | 42.176 | 53.881 |
| 03 | 6 | 747.421 | 24.426 | 43.458 | 52.563 |
| 04 | 8 | 599.859 | 33.847 | 49.223 | 54.679 |
| 05 | 6 | 819.640 | 39.812 | 39.778 | 51.517 |

Counts are 12:1, 8:2, 6:2; matrix incomplete. Four valid runs remain. The summary tool excludes the nonquiet attempt. These observations do not yet satisfy the nine-run decision contract.

## Revisions

- 2026-10-20 (hub-approved scope fold): original Done-when was “The root cause is fixed (never a retry loop or a longer timeout); closing this issue unregisters the tests.” Bounded diagnostics identify the costly cleanup window but do not reproduce the original two-minute stall. Replace the unqualified root-cause promise with measured suite-envelope correction and persistent diagnostic evidence; retain the no-retry/no-timeout-bump constraint. Fold #484's 12/8/6 measurement into #480; leave cleanup-lock optimization as a possible separate follow-up. The diagnostic plan row now distinguishes observed lock ownership from unresolved starvation attribution.
