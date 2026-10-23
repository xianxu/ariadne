---
id: 000572
status: working
deps: []
github_issue:
created: 2026-10-23
updated: 2026-10-23
estimate_hours:
collaboration-mode: collaboration
card_mirror: 'b272525ae2e819c2531aadde3853ae4243d20ca8' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-23T04:39:43-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:4
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot4/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: quick, provenance: inferred, spec: "2d30e3cd", done: "8c0d851f"}
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

Claimed and started planning from main `ee4dd7fa1e8488e3230dc462971be8796f317e41`. Read current local-run, startup-environment, reclassification, resume and flake-rerun boundaries. Authenticated ariadne:2's evidence receipt `e2e9eb41-c3bf-4269-8e92-c754340784b9`; independently read the raw failed log and hashed it and the same-head pass. Failed attempt `c5fff316216ec03a781b4796f770a3e5` on `94aa0864e648e752b90c7fbe354b19c1569a9126` remains untouched. No flake was registered. Bounded read-only exploration confirmed that ordinary log output cannot prove complete failure causation; chose independent capability preflight (ARCH-SECURE, ARCH-PURPOSE), shared at execution entry points (ARCH-DRY). Plan is awaiting hub approval; no production code changed.

ops:0 approved the plan (receipt `aee2106e-8cce-418b-b520-b23799ad991c`); `sdlc change-code --issue 572` admitted quick flow. Implemented one shared boot/listener preflight before ordinary suite/affected/baseline execution, automatic flake children and eligible resume children. Saved-pass repost skips execution checks. Error advice uses wrapped EPERM identity, never denial text; failures on a healthy host retain attribution. Copied the original mixed log unchanged as a hash-checked fixture. Resource tests cover cancellation and acquisition/close/remove failures. The affected graph-entry regression went red (`affected graph launched before preflight`) before the same helper was added there; this is the existing affected execution boundary, not a new mechanism.

Focused verification: capability controls PASS (3.582s); expanded environment, infrastructure, resume, affected and local-check controls PASS (67.441s); golint PASS (1.525s). Native Seatbelt boot-identity and listener denials PASS (1.072s): could-not-run, whole-call advice, queue/attempt/status unchanged; host listener succeeds before/after. Initial broad sysctl denial stopped Go's allocator before the test; narrowed the fixture to the actual `kern.boottime` read and confirmed both controls. Next: capture final pinned native evidence and race controls, then commit/push the final head and run its recorded suite. No full suite, close or landing has been started yet.

Implementation checkpoint `560df628f1d09836d86b0d3845f32a5097c7b52c`. Final native conformance on that source PASS (0.727s); [raw output and exact production/probe blob pins](../plans/000572-native-conformance.md) are committed text evidence so a contained reviewer can inspect the host-only check. Capability controls under `-race` PASS (5.221s). Whitespace check passes outside the byte-identical captured original log, whose four whitespace-only lines are intentionally preserved for hash verification. Next: recorded full suite on the pushed checkpoint, then close review and normal landing; no stage required.

Recorded suite retained on `c67660c1c`: attempt `91c394cba1738bdea7c3ddc0f4eaa204`, lease `lease-1362897573`, initially position 6 behind the existing FIFO holders. No second runner or priority request. Per the explicit tl-worker overlap rule, close review may run while this suite is queued; a final-head recorded PASS still precedes landing, and any code correction requires a fresh recorded suite. Outstanding: retained suite, close verdict, normal merge. No restart/continuation while the retained check is live.

The first close invocation stopped at the checklist before dispatch: its only row bundled completed implementation with the close review itself and the still-queued suite. Corrected that row to state completed implementation/conformance; the future gates remain explicitly outstanding in this Log and the durable plan. No bypass flag used. The actual 162-line production addition upgrades close to the full boundary review automatically.

Close round 1: REWORK, BR-1 pure-owner table mismatch, BR-2 unused lint suppression, BR-3 missing README refusal/retry guidance. Reproduced BR-2 after tracking the new file: golint failed on the unused suppression (1.204s); earlier pre-commit PASS omitted that then-untracked file. Corrected the concept table to the actual pure formatter and IO functions, added affected graph preflight to it, removed only the suppression directive, and updated README alongside existing help/atlas. Added the class lessons. No new runtime mechanism. Also covered the review's optional resume-denial -> repair -> exact retry sequence; its first fake pass omitted package elapsed proof and was rejected correctly, then corrected the fixture to valid Go events. Capability controls including real-store retry PASS (3.551s); tracked golint PASS after correction (1.072s).

Cancelled only this queued suite through `sdlc local-check cancel --run lease-1362897573` after the source correction; the retained command exited with context cancellation before acquisition. No suite ran or failure was recorded. Old attempt/lease are no longer live. Next: refresh the native source pin/output for the comment-only source correction, push, queue one final-head suite, and re-close with all three findings addressed.
