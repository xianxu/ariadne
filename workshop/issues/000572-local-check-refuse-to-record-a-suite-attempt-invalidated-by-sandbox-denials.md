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

- [ ] Implement the [host capability preflight plan](../plans/000572-suite-capability-preflight-plan.md), then focused/native conformance, recorded suite and one close review.

## Revisions

### 2026-10-23 — Independent capability evidence before execution

The original capture is mixed: besides EPERM failures it contains `invalid attempt record`. A majority classifier could hide a code defect, and the sharded reporter retains only the last 40 output events per failed test. ops:0 approved an independent preflight of the existing boot identity read and a private Unix listener, before queueing or launch (receipt `d01f12b8-070e-4d91-ab10-7056c270e6d4`). Use the existing could-not-run result with whole-call guidance on EPERM. No new verb and no post-hoc log classifier. Existing mixed failures remain failures.

Revised Done when: denied required host capabilities prevent execution and leave no attempt or status; repairing the environment permits a genuine same-head run. Healthy-host permission failures, all-denial text and mixed output retain normal failure recording. Cover ordinary suite/affected runs, baselines and rerun entry points without altering saved-pass repost behavior.

## Log

### 2026-10-23

Claimed and started planning from main `ee4dd7fa1e8488e3230dc462971be8796f317e41`. Read current local-run, startup-environment, reclassification, resume and flake-rerun boundaries. Authenticated ariadne:2's evidence receipt `e2e9eb41-c3bf-4269-8e92-c754340784b9`; independently read the raw failed log and hashed it and the same-head pass. Failed attempt `c5fff316216ec03a781b4796f770a3e5` on `94aa0864e648e752b90c7fbe354b19c1569a9126` remains untouched. No flake was registered. Bounded read-only exploration confirmed that ordinary log output cannot prove complete failure causation; chose independent capability preflight (ARCH-SECURE, ARCH-PURPOSE), shared at execution entry points (ARCH-DRY). Plan is awaiting hub approval; no production code changed.
