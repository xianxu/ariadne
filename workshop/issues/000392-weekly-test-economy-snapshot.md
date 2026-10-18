---
id: 000392
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours:
card_mirror: '43f465e35c614cc4ba1092fa808abd432789a32d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:39:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
---

# Weekly test-economy snapshot from local checks and landings

## Problem

#390's test-economy design needs a trustworthy weekly trend before changing evidence tiers or declaring package budgets. `sdlc test-stats` reports recent local attempts, but no durable weekly snapshot joins them to landed issues; older records expire after 30 days. The headline metric, full suites per landed issue, has no stored baseline.

## Spec

Produce one dated, immutable weekly per-repo summary from #376's per-attempt local-check records and authoritative landing records. Show the exact seven-day window, source/coverage counts, suite and queue p50/p90, slowest packages with sample counts, confirmed/suspected/unknown flake classes, and full-suite attempts per landed issue with numerator, denominator, failed/repeated attempts and unknown attribution. Do not infer missing attempts or landings. Store the measurement snapshot in brain under its charter, outside the SDLC issue process; link it from the ops test-economy project. The command/process that builds it belongs in ariadne and becomes a default for every SDLC repo after the ariadne/pair pilot. Brain is the output store, not the implementation repository. Follow #390's approved design and measure before changing budgets.

## Done when

- [ ] A repeatable weekly command produces a dated snapshot in brain and links it from the ops project without duplicating the numbers there.
- [ ] The output separates wall and queue, gives p50/p90 and coverage for each repo, names slow packages and flake classes, and calculates full suites per landed issue from exact attempt and landing evidence.
- [ ] Tests cover failed/retried attempts, missing old fields, unassignable attempts, zero landings, and stable reruns of the same weekly window; unknown remains unknown.
- [ ] A first ariadne/pair snapshot is saved, and the default rollout and retention/backup responsibility are documented.

## Plan

- [ ] Design the evidence join and snapshot schema in `workshop/plans/000392-weekly-test-economy-snapshot-plan.md` under #390's approved measurement revision; obtain hub approval before code.
- [ ] Build the pure weekly model and exact attempt-to-landing join with fixture-backed red/green tests.
- [ ] Add the brain writer and CLI, test immutable reruns and partial evidence, then save the first ariadne/pair snapshot and give its link to the TL for the ops project.

## Log

### 2026-10-15

Filed from #390's operator-approved test-economy design. This precedes #328 and any package-budget declaration.

2026-10-17: The hub directed this slot to build #392 before dependent #402. Existing `sdlc test-stats` and per-attempt records supply suite, queue, package and failure inputs; tracker completion bindings and the landing-line event log supply a confirmed landed issue and admission window. The initial snapshot will preserve coverage unknowns for pre-#376 legacy logs and any missing admission or Git ancestry instead of assigning them by date alone (ARCH-PURPOSE, ARCH-ORDER).

The hub approved the fixed UTC window, strict joined-evidence attribution, separate unassignable count, nullable ratio and versioned JSON schema; it will link the first live brain snapshot from the path reported at landing.

Plan-quality review found that the seven-day physical attempt population and landed-issue cohort were conflated, landing events needed semantic validation, and the test strategy was too procedural. The durable plan now separates the populations, validates the enqueue/landing/card/Git binding before attribution, and names adversarial strategies by function. Minor operating-envelope and lifecycle findings are addressed in the same revision (ARCH-PURPOSE, ARCH-SECURE, ARCH-CONSTRAINTS).
