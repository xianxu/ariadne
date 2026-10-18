---
id: 000392
status: working
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-17
estimate_hours: 4.26
card_mirror: 'cab30c82d9025e928df1f7c1faca8276a238bf9d' # card fields mirrored from issue-cards; edit via sdlc
started: 2026-10-17T21:39:59-07:00
claimant:
    operator: Xian Xu
    machine: 4716879978a7b90f6b583da1716fd0e9
    machine_name: Xian’s MacBook Pro
    workspace: ariadne:1
    worktree: /Users/xianxu/workspace/worktree/ariadne-slot1/ariadne
    repository: github.com/xianxu/ariadne
flow: {kind: full, provenance: operator}
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

## Estimate

Produced via `brain/data/life/42shots/velocity/estimate-logic-v3.1.md` against `baseline-v3.1.md`. Method A only. The two greenfield modules are the weekly metric and exact attribution; the API integration is the bounded Git/tracker/landing evidence reader; the remaining module is the atomic brain writer and CLI. Existing Go libraries and local primitives cover the stack, so no new-library design allowance is needed. The source reports stale calibration, so these hours are provisional.

```estimate
model: estimate-logic-v3.1
familiarity: 1.0
item: greenfield-go-module design=0.5 impl=0.32
item: greenfield-go-module design=0.5 impl=0.32
item: api-integration design=0.6 impl=0.6
item: greenfield-go-module design=0.3 impl=0.32
item: atlas-docs design=0.1 impl=0.08
item: milestone-review design=0.1 impl=0.2
design-buffer: 0.15
total: 4.26
```

## Log

### 2026-10-15

Filed from #390's operator-approved test-economy design. This precedes #328 and any package-budget declaration.

2026-10-17: The hub directed this slot to build #392 before dependent #402. Existing `sdlc test-stats` and per-attempt records supply suite, queue, package and failure inputs; tracker completion bindings and the landing-line event log supply a confirmed landed issue and admission window. The initial snapshot will preserve coverage unknowns for pre-#376 legacy logs and any missing admission or Git ancestry instead of assigning them by date alone (ARCH-PURPOSE, ARCH-ORDER).

The hub approved the fixed UTC window, strict joined-evidence attribution, separate unassignable count, nullable ratio and versioned JSON schema; it will link the first live brain snapshot from the path reported at landing.

Plan-quality review found that the seven-day physical attempt population and landed-issue cohort were conflated, landing events needed semantic validation, and the test strategy was too procedural. The durable plan now separates the populations, validates the enqueue/landing/card/Git binding before attribution, and names adversarial strategies by function. Minor operating-envelope and lifecycle findings are addressed in the same revision (ARCH-PURPOSE, ARCH-SECURE, ARCH-CONSTRAINTS).

Change-code passed the full-flow plan gate in three rounds; its estimate-quality gate returned INFO about the tight writer allowance and operational verification time. The accepted provisional 4.26 h estimate follows the stale v3.1 calibration. TDD slices for the pure statistics, evidence reader, attribution and brain writer are implemented and focused tests pass. A read-only ariadne pilot for the week ending 2026-10-18 found 83 physical suites, 22 confirmed landings, 12 post-enqueue attributed attempts, 47 pre-enqueue attempts and 24 in-week attempts outside the landing cohort; the latter buckets explain why the two populations differ. The pilot exposed a missing package-only slow list, now added with coverage, and prompted a main-merge inventory check for missing landing events. This pilot wrote only to a temporary brain-shaped directory, not the durable brain repo.

The first full-fleet pilot found pre-landing-line merges in the seven-day UTC week, so ariadne and pair correctly have a null headline ratio with explicit missing-event coverage. It also exposed generic fleet discovery including unrelated sibling repositories; the command now selects only repositories with the vocabulary-declared issue home. The merge's GitHub PR and branch are checked against the landing event before attribution (ARCH-SECURE). The first saved snapshot will be an intentionally partial baseline; a later complete week can establish the numeric trend.

A second pilot revealed brain repos with a retained issue-home directory. The discovery predicate now honors brain's capture-only marker first. The recorded suite on the older pushed commit had not acquired the machine lock; it was cancelled without a test result so the corrected commit can receive the single authoritative run.

Adversarial attribution tests then exposed two partial-evidence paths: an earlier enqueue without its refusal and a legacy attempt on the landed branch. Both now null the ratio. The suite queued on the preceding head was likewise cancelled before it acquired the machine lock (ARCH-SECURE, ARCH-PURPOSE).

The snapshot CLI no longer accepts a repo filter: a filtered pilot could have occupied the immutable fleet-wide date path. The read-only test-stats command still offers per-repo inspection. This keeps one path equal to one denominator (ARCH-ORDER).

The hub approved the evidence-boundary revisions, then required the snapshot headline to show coverage alongside the ratio. A TDD check exposed that the pilot JSON had only a nullable ratio and dispersed warning strings. The repository headline now states covered/observed landed issues, unknown count and top reasons; incomplete source evidence remains a null ratio. The prior queued suite was cancelled before acquisition and will be re-run on the corrected head.

The recorded suite on the headline head completed 12 shards and 67 other packages in 9m19s. Only `TestVerbContextsReachTrackerReads` failed: `weeklyGit`'s nil-context fallback minted a new root context. All real callers pass the command context; the fallback now returns an explicit error before Git starts. The targeted guard, nil-context test and weekly tests pass. The failed attempt and its log remain recorded for the retry.
