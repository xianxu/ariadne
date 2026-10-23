---
id: 000572
status: codecomplete
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: '12f87c09d35483b6d69bf0ddd72247433f37b694' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T04:39:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: inferred}
actual_hours: 0.84
---

# local-check: refuse to record a suite attempt invalidated by sandbox denials

## Problem

A failed local-check attempt has only two dispositions: register its tests as flakes (fail then pass is the evidence) or fix the code and push a new head (#394). Neither fits an attempt that its environment invalidated. In #561, `sdlc local-check run --name suite` was piped through `tail`. That broke the whole-call sandbox exemption, so the suite ran sandboxed and every failure was `operation not permitted` (resume-environment tests, TestQueueRealHelperPause, a golint socket test). The unpiped rerun on the same head passed, but sdlc refused to record it because the earlier failure was unexplained. Registering those tests as flakes would have falsified the flake record. The only honest way out was a new head and a second full suite: about 15 minutes of suite time plus queue wait. `local-check reclassify` covers startup-only failures, so it didn't apply.

## Spec

Detect a run invalidated by its environment and refuse to record it as a failure at all. When the failing tests' output is dominated by sandbox denials (`operation not permitted` / EPERM on file, socket or process operations), the run records neither a failure nor a pass. The checker posts no failing status and saves no failed attempt that blocks the head. It prints an error naming the cause: the run was sandboxed, so rerun it as the whole Bash call. Open questions for design:
- the classification rule: every failure carries a denial, or only a majority does;
- whether to add an explicit operator disposition for an existing attempt, such as `local-check invalidate --attempt ID --reason environment`, so attempts recorded before the detector existed can be cleared without a new head;
- how to keep a real permission bug in the code under test from being waved away. One idea is to require that the same tests pass on a rerun of the same head.

## Done when

- A suite run whose failures are all sandbox denials leaves no failed attempt and no failing `sdlc-local/*` status on the head, and says why.
- A run with any failure that is not a denial records as a failure exactly as today.
- Tests cover both cases plus a mixed run, using captured failure output.

## Plan

- [x] Implement the [host capability preflight plan](../plans/000572-suite-capability-preflight-plan.md) and focused/native conformance. Recorded-suite and review/landing gates remain separately tracked below.

## Revisions

### 2026-10-23 — Independent capability evidence before execution

The original capture is mixed: besides EPERM failures it contains `invalid attempt record`. A majority classifier could hide a code defect, and the sharded reporter retains only the last 40 output events per failed test. ops:0 approved an independent preflight of the existing boot identity read and a private Unix listener, before queueing or launch (receipt `d01f12b8-070e-4d91-ab10-7056c270e6d4`). Use the existing could-not-run result with whole-call guidance on EPERM. No new verb and no post-hoc log classifier. Existing mixed failures remain failures.

Revised Done when: denied required host capabilities prevent execution and leave no attempt or status; repairing the environment permits a genuine same-head run. Healthy-host permission failures, all-denial text and mixed output retain normal failure recording. Cover ordinary suite/affected runs, baselines and rerun entry points without altering saved-pass repost behavior.

## Log

### 2026-10-23
- 2026-10-23: closed — Recorded suite exposed deterministic overlong WF_TEST_SCRATCH probe path; failed attempt ed0499e973a9764a01b6040e40e59520 retained. Original scratch test and explicit long-root control reproduced red, then PASS with bounded canonical path selection before bind (103 bytes, at most two roots; no fallback on any permission or IO error). Expanded capability/scratch/environment/infrastructure checks PASS 11.099s; tracked golint PASS 0.799s; race capability/scratch controls PASS 5.412s. Native long-root boot/bind denials PASS 0.795s on 4086e5373, raw pins in workshop/plans/000572-native-conformance.md. New-source recorded suite retained on 37a4536e4, attempt 41c529571f7e1460e3c789680d4cf93c / lease-3102453792; permitted review overlap, pass required before landing. This fixes the approved short-path envelope, no scope change. (fixed-then-ship); review verdict: SHIP
- 2026-10-23: closed — BR-1 actual pure formatter/IO tables and affected entry corrected; BR-2 unused suppression removed, tracked golint red reproduced then PASS 1.072s; BR-3 README/help/atlas refusal guidance aligned. Capability controls including denied resume then repaired exact retry PASS 3.551s; earlier expanded checks PASS 67.441s and race PASS 5.221s. Final native boot/listener kernel-denial conformance PASS 1.011s on f65dc200e, raw source-pinned evidence workshop/plans/000572-native-conformance.md. Replacement final-source suite queued (ed0499e973a9764a01b6040e40e59520 / lease-238317380); tl-worker permits review overlap, recorded pass required before landing. Existing failure evidence unchanged.; review verdict: SHIP
- 2026-10-23: flow upgraded quick → full — 162 added lines in code files (limit 100); an earlier round of this close already ran the full review

Claimed and started planning from main `ee4dd7fa1e8488e3230dc462971be8796f317e41`. Read current local-run, startup-environment, reclassification, resume and flake-rerun boundaries. Authenticated ariadne:2's evidence receipt `e2e9eb41-c3bf-4269-8e92-c754340784b9`; independently read the raw failed log and hashed it and the same-head pass. Failed attempt `c5fff316216ec03a781b4796f770a3e5` on `94aa0864e648e752b90c7fbe354b19c1569a9126` remains untouched. No flake was registered. Bounded read-only exploration confirmed that ordinary log output cannot prove complete failure causation; chose independent capability preflight (ARCH-SECURE, ARCH-PURPOSE), shared at execution entry points (ARCH-DRY). Plan is awaiting hub approval; no production code changed.

ops:0 approved the plan (receipt `aee2106e-8cce-418b-b520-b23799ad991c`); `sdlc change-code --issue 572` admitted quick flow. Implemented one shared boot/listener preflight before ordinary suite/affected/baseline execution, automatic flake children and eligible resume children. Saved-pass repost skips execution checks. Error advice uses wrapped EPERM identity, never denial text; failures on a healthy host retain attribution. Copied the original mixed log unchanged as a hash-checked fixture. Resource tests cover cancellation and acquisition/close/remove failures. The affected graph-entry regression went red (`affected graph launched before preflight`) before the same helper was added there; this is the existing affected execution boundary, not a new mechanism.

Focused verification: capability controls PASS (3.582s); expanded environment, infrastructure, resume, affected and local-check controls PASS (67.441s); golint PASS (1.525s). Native Seatbelt boot-identity and listener denials PASS (1.072s): could-not-run, whole-call advice, queue/attempt/status unchanged; host listener succeeds before/after. Initial broad sysctl denial stopped Go's allocator before the test; narrowed the fixture to the actual `kern.boottime` read and confirmed both controls. Next: capture final pinned native evidence and race controls, then commit/push the final head and run its recorded suite. No full suite, close or landing has been started yet.

Implementation checkpoint `560df628f1d09836d86b0d3845f32a5097c7b52c`. Final native conformance on that source PASS (0.727s); [raw output and exact production/probe blob pins](../plans/000572-native-conformance.md) are committed text evidence so a contained reviewer can inspect the host-only check. Capability controls under `-race` PASS (5.221s). Whitespace check passes outside the byte-identical captured original log, whose four whitespace-only lines are intentionally preserved for hash verification. Next: recorded full suite on the pushed checkpoint, then close review and normal landing; no stage required.

Recorded suite retained on `c67660c1c`: attempt `91c394cba1738bdea7c3ddc0f4eaa204`, lease `lease-1362897573`, initially position 6 behind the existing FIFO holders. No second runner or priority request. Per the explicit tl-worker overlap rule, close review may run while this suite is queued; a final-head recorded PASS still precedes landing, and any code correction requires a fresh recorded suite. Outstanding: retained suite, close verdict, normal merge. No restart/continuation while the retained check is live.

The first close invocation stopped at the checklist before dispatch: its only row bundled completed implementation with the close review itself and the still-queued suite. Corrected that row to state completed implementation/conformance; the future gates remain explicitly outstanding in this Log and the durable plan. No bypass flag used. The actual 162-line production addition upgrades close to the full boundary review automatically.

Close round 1: REWORK, BR-1 pure-owner table mismatch, BR-2 unused lint suppression, BR-3 missing README refusal/retry guidance. Reproduced BR-2 after tracking the new file: golint failed on the unused suppression (1.204s); earlier pre-commit PASS omitted that then-untracked file. Corrected the concept table to the actual pure formatter and IO functions, added affected graph preflight to it, removed only the suppression directive, and updated README alongside existing help/atlas. Added the class lessons. No new runtime mechanism. Also covered the review's optional resume-denial -> repair -> exact retry sequence; its first fake pass omitted package elapsed proof and was rejected correctly, then corrected the fixture to valid Go events. Capability controls including real-store retry PASS (3.551s); tracked golint PASS after correction (1.072s).

Cancelled only this queued suite through `sdlc local-check cancel --run lease-1362897573` after the source correction; the retained command exited with context cancellation before acquisition. No suite ran or failure was recorded. Old attempt/lease are no longer live. Next: refresh the native source pin/output for the comment-only source correction, push, queue one final-head suite, and re-close with all three findings addressed.

Round-1 correction checkpoint `f65dc200e9fb967eaf02af9c38730eddfc6727d0`. Refreshed native conformance PASS (1.011s) on that HEAD; appended raw output and the changed probe-owner source blob to the existing evidence document. All three findings are addressed; no dispute or waiver requested. Outstanding: one new final-source recorded suite and round-2 close review, then normal landing.

Replacement suite retained on `372624840`: attempt `ed0499e973a9764a01b6040e40e59520`, lease `lease-238317380`, initially position 7 in normal FIFO. Native evidence Markdown formatting corrected outside its unchanged raw fences; plan progress checked through completed focused/native verification. Round-2 close review overlaps this retained suite under the worker contract; final-source recorded PASS remains mandatory before merge.

Close round 2 SHIP with all three findings disposed and none new; evidence `e7fea7115d276acf7afab6cdc6d08bb0ef283e34`, mirror `3577baebd7fbd43489763074341c92dc6a135462`. PR [#412](https://github.com/xianxu/ariadne/pull/412) opened; local-gates checks passed. Replacement suite acquired after 3165s and completed FAIL in 908.1s: `TestRunUnderSlotUsesScratch/scratch` failed because the new private listener used an overlong canonical path under explicit WF_TEST_SCRATCH. Other seven shards and all 76 packages passed (2382 SDLC tests total). This is a deterministic branch defect; failed attempt `ed0499e973a9764a01b6040e40e59520` stays recorded, with log SHA256 prefix `089bc070d221`. No flake registration or reclassification.

Reproduced the original scratch test and a new long-WF_TEST_SCRATCH control red with `bind: invalid argument`. Fixed the already-approved short-path envelope: choose the short system root only when the canonical candidate socket path exceeds 103 bytes, before bind, with two candidates maximum and owned rejected-directory cleanup. Permission and other errors never authorize another root. Focused capability/scratch/environment/infrastructure checks PASS (11.596s), tracked golint PASS (0.891s). Hub notified via submitted receipt `a9628746-ffba-4dc3-b2dc-060d97ac2ae2`. Added canonical symlink and native long-root denial coverage. Outstanding: final focused/native verification, updated source pins, a new recorded suite and re-close of this code correction, then landing.

Short-root correction committed as `4086e5373c5460ab9b967ecec1a51751b0dfb556`. Expanded focused controls PASS (11.099s), tracked golint PASS (0.799s), capability and original scratch controls under `-race` PASS (5.412s). Refreshed native host conformance PASS (0.795s), with the deliberately long scratch root and real boot/bind kernel denials; appended unchanged raw output and updated source/probe blobs to the evidence document. Next: new-source recorded suite, narrow `--fixed-then-ship` review of this bounded correction, then normal merge after both pass. The failed prior suite remains evidence of the corrected branch defect.

Narrow re-close round 3 SHIP, no new findings; evidence `4f0b071ed3c4d55f3f7a76681d55ee3f7fe31eb1`, mirror `0e4fd87387d597120e9d83c7730be32e28b3f288`. Reviewer independently removed the path fix in a disposable clone and observed the expected regression, then verified focused controls (23.209s) and tracked golint (1.242s); host-only native pins match. PR #412 updated. Final-source suite is retained on `37a4536e423b774dd71d6ca746783de445b5b3c4`: attempt `41c529571f7e1460e3c789680d4cf93c`, lease `lease-3102453792`, background terminal 92330, now position 4 in normal FIFO. No continuation/restart while it is queued or running. Exact next step: read that existing runner to completion; PASS -> record evidence and run `sdlc merge --yes`; a new failure -> diagnose on the exact head. No second runner, priority grant, stage, or failure waiver.

Final-source recorded suite PASS on `37a4536e423b774dd71d6ca746783de445b5b3c4`, attempt `41c529571f7e1460e3c789680d4cf93c`: all 8 shards, 2382 SDLC tests and 76 other packages, wall 1098.8s (18m19s), queue 3067s. Recorded `sdlc-local/suite` success; log `<common-git>/sdlc/local-checks/37a4536e423b774dd71d6ca746783de445b5b3c4/suite.log`, SHA256 prefix `1f4ba5c930d9`. Runner 92330 completed exit 0. The long-root scratch regression now passes in the full suite; earlier failed attempt remains intact. With SHIP and the corrected-source suite green, next step is normal `sdlc merge --yes`; its landing turn owns main integration and any required integrated-head suite.

Retained merge runner 93224 completed with a code conflict while integrating main, rather than being lost at the Couch restart. Following its recovery instruction, manually integrated `db77404e3257507f41c9d6fc3d819a8723bb36fa`. Resolved only the formatter overlap by using main's canonical `localCheckRunCommand` inside the preflight advice, preserving affected `--boarding`, explicit resume retry identity, EPERM whole-call guidance and absent-log formatting (ARCH-DRY). Updated the affected fixture for the new execution-mode parameter after the initial compile check exposed that interface drift. Focused capability, environment, infrastructure, canonical-command, scratch and affected controls PASS (80.378s); tracked golint PASS (0.882s). Resolved files pass whitespace checks; broad merge diff also contains inherited whitespace in main's archived #561 review artifacts, left unchanged. Next: push this integrated checkpoint, run its recorded full suite, then the required `sdlc close --issue 572 --fixed-then-ship --full-suite` and retry normal merge. No bypass, duplicate runner, flake registration or evidence reclassification.

Integrated-source recorded suite PASS on `e270803f0b0f04c560d4e9e49279bdfeda16266e`, attempt `ce5a5fb083cf6dfcd3374a94401985b1`, lease `lease-514948582`: all 8 shards, 2410 SDLC tests and 76 other packages, wall 902.4s, queue 2003s; exit 0, recorded `sdlc-local/suite`. Log `<common-git>/sdlc/local-checks/e270803f0b0f04c560d4e9e49279bdfeda16266e/suite.log`, SHA256 prefix `8e8a9c5d7072`. Retained terminal 34241 completed. Next: narrow fixed-then-ship re-close with this full-suite evidence for the formatter conflict resolution, then normal merge. Hub assigned #577 only after #572 lands; do not start its planning yet.
