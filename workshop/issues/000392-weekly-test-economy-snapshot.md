---
id: 000392
status: open
deps: []
github_issue:
created: 2026-10-15
updated: 2026-10-15
estimate_hours:
card_mirror: 'f1d5a7acac6ac47c94c4830a0bf81dfbb59f01c1' # card fields mirrored from issue-cards; edit via sdlc
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

- [ ] Design the evidence join and snapshot schema under #390's approved measurement revision.
- [ ] Build and test the producer, then save and link the first snapshot.

## Log

### 2026-10-15

Filed from #390's operator-approved test-economy design. This precedes #328 and any package-budget declaration.
